package oaipmh

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// HarvestIdentifiers arvest the identifiers of a complete OAI set
// call the identifier callback function for each Header
func (request *Request) HarvestIdentifiers(callback func(*Header) error) error {
	request.Verb = "ListIdentifiers"
	return request.Harvest(func(resp *Response) error {
		if resp.ListIdentifiers == nil {
			return nil
		}

		headers := resp.ListIdentifiers.Headers
		for _, header := range headers {
			request.recordsSeen++
			if err := callback(&header); err != nil {
				return err
			}
		}
		return nil
	})
}

// HarvestRecords harvest the records of a complete OAI set
// call the record callback function for each Record
func (request *Request) HarvestRecords(callback func(*Record) error) error {
	request.Verb = "ListRecords"
	return request.Harvest(func(resp *Response) error {
		if resp.ListRecords == nil {
			return nil
		}

		records := resp.ListRecords.Records
		for _, record := range records {
			request.recordsSeen++
			if err := callback(&record); err != nil {
				return err
			}
		}
		return nil
	})
}

// Harvest perform a harvest of a complete OAI set, or simply one request
// call the batchCallback function argument with the OAI responses
func (request *Request) Harvest(batchCallback func(*Response) error) error {
	for {
		// Use Perform to get the OAI response
		oaiResponse, err := request.perform()
		if err != nil {
			slog.Info("unable to perform harvest; stopping now", "error", err)
			return err
		}

		if oaiResponse == nil {
			slog.Error("oai-pmh response is empty")
			return fmt.Errorf("unable to run batchCallback on an empty response")
		}

		// Execute the callback function with the response
		if batchErr := batchCallback(oaiResponse); batchErr != nil {
			return fmt.Errorf("unable to perform batch callback; %w ", batchErr)
		}

		// Check for a resumptionToken
		hasResumptionToken, resumptionToken, completeListSize := oaiResponse.GetResumptionToken()

		// Break the loop if there is no resumption token
		if !hasResumptionToken {
			break
		}

		// Prepare the request for the next iteration
		request.Set = ""
		request.MetadataPrefix = ""
		request.From = ""
		request.ResumptionToken = resumptionToken
		request.CompleteListSize = completeListSize
	}
	return nil
}

func (request *Request) writeDebug(b []byte, triesRemaining int) error {
	if request.DebugOut == "" {
		return nil
	}
	if err := os.MkdirAll(request.DebugOut, os.ModePerm); err != nil {
		slog.Error("unable to write debug directory", "error", err)
		return err
	}

	token := request.ResumptionToken
	if token == "" {
		token = "first_page"
	}
	if strings.Contains(token, "/") {
		token = strings.ReplaceAll(token, "/", "-")
	}

	fname := filepath.Join(request.DebugOut, fmt.Sprintf("%06d_%s_%d.xml", request.pagesSeen, token, triesRemaining))

	return os.WriteFile(fname, b, os.ModePerm)
}

// perform an HTTP GET request using the OAI Requests fields
// and return an OAI Response reference
func (request *Request) perform() (oaiResponse *Response, err error) {
	if request.client == nil {
		if request.Timeout == 0 {
			request.Timeout = time.Duration(60 * time.Second)
		}
		request.client = &http.Client{
			Timeout: request.Timeout,
		}
	}

	request.pagesSeen++

	err = retry(40, time.Second, func(triesRemaining int) error {
		req, requestErr := http.NewRequest(http.MethodGet, request.GetFullURL(), nil)
		if requestErr != nil {
			return requestErr
		}

		if request.UserName != "" && request.Password != "" {
			req.SetBasicAuth(request.UserName, request.Password)
		}

		resp, requestErr := request.client.Do(req)
		if requestErr != nil {
			return requestErr
		}

		// Make sure the response body object will be closed after
		// reading all the content body's data
		defer resp.Body.Close()

		data, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			slog.Error("unable to read body", "err", readErr, "url", request.GetFullURL())
			return stop{readErr}
		}

		if writeErr := request.writeDebug(data, triesRemaining); writeErr != nil {
			slog.Error("unable to write debug file", "error", writeErr)
			return stop{writeErr}
		}

		l := slog.With(
			"status_code", resp.StatusCode, "url", request.GetFullURL(),
			"body", string(data),
		)

		s := resp.StatusCode
		switch {
		case s >= 500:
			// Retry
			l.Warn("server error")
			return fmt.Errorf("server error: %v", s)
		case s == 408:
			// Retry
			l.Warn("timeout")
			return fmt.Errorf("timeout error: %v", s)
		case s >= 400:
			// Don't retry, it was client's fault
			l.Warn("client error; stopping retry")
			return stop{fmt.Errorf("client error: %v", s)}
		default:
			marshallErr := xml.Unmarshal(data, &oaiResponse)
			if marshallErr != nil {
				if bytes.Contains(data, []byte("<OAI-PMH>")) {
					l.Error("unable to unmarshal OAI-PMH response; no retry", "error", marshallErr, "response", string(data))
					return stop{marshallErr}
				}
				l.Error("unable to unmarshal OAI-PMH response; retrying", "error", marshallErr, "response", string(data))
				return fmt.Errorf("bad oai-pmh response, so retrying")
			}

			var listSize int
			var isEmptyResponse bool

			switch {
			case oaiResponse.ListRecords != nil:
				listSize = len(oaiResponse.ListRecords.Records)
				isEmptyResponse = listSize == 0
			case oaiResponse.ListIdentifiers != nil:
				listSize = len(oaiResponse.ListIdentifiers.Headers)
				isEmptyResponse = listSize == 0
			default:
				// If we're here, we have neither ListRecords nor ListIdentifiers
				// This might be a valid empty response if it's the final page
				isEmptyResponse = true
			}

			hasToken, _, tokenListSize := oaiResponse.GetResumptionToken()

			// Check if we've received all expected records
			isComplete := request.CompleteListSize > 0 && (request.recordsSeen+listSize) >= request.CompleteListSize

			// Check for an invalid state: no token but harvest is incomplete
			if !hasToken && !isComplete && request.CompleteListSize > 0 && !isEmptyResponse {
				// We're expecting more records but got no resumption token
				l.Error("missing resumption token in incomplete harvest; retrying", "response", string(data))
				slog.Info("retrying due to missing resumption token",
					"expected_completeListSize", request.CompleteListSize,
					"records_returned", listSize,
					"records_seen", request.recordsSeen,
					"url", request.GetFullURL(),
					"tokenListSize", tokenListSize,
				)
				return fmt.Errorf("missing resumption token in incomplete harvest; retrying")
			}

			// If we get here, either:
			// 1. We have a resumption token (harvest continues)
			// 2. We have no token but harvest is complete (normal completion)
			// 3. We have no token, but don't know the complete list size (assume complete)
			// All these cases are valid
			return nil
		}
	})
	if err != nil {
		slog.Error("unable to finish oai-pmh harvest", "error", err, "url", request.GetFullURL())
		return nil, err
	}

	return oaiResponse, nil
}
