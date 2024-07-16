package oaipmh

import (
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

		if err := request.writeDebug(data, triesRemaining); err != nil {
			slog.Error("unable to write debug file", "error", err)
			return stop{err}
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
				l.Error("unable to unmarshal OAI-PMH response", "error", marshallErr, "response", string(data))
				return stop{marshallErr}
			}

			return nil
		}
	})
	if err != nil {
		slog.Error("unable to finish oai-pmh harvest", "error", err, "url", request.GetFullURL())
		return nil, err
	}

	return oaiResponse, nil
}
