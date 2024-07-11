package oaipmh

import (
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// HarvestIdentifiers arvest the identifiers of a complete OAI set
// call the identifier callback function for each Header
func (request *Request) HarvestIdentifiers(callback func(*Header)) {
	request.Verb = "ListIdentifiers"
	request.Harvest(func(resp *Response) {
		headers := resp.ListIdentifiers.Headers
		for _, header := range headers {
			callback(&header)
		}
	})
}

// HarvestRecords harvest the identifiers of a complete OAI set
// call the identifier callback function for each Header
func (request *Request) HarvestRecords(callback func(*Record)) {
	request.Verb = "ListRecords"
	request.Harvest(func(resp *Response) {
		records := resp.ListRecords.Records
		for _, record := range records {
			callback(&record)
		}
	})
}

// Harvest perform a harvest of a complete OAI set, or simply one request
// call the batchCallback function argument with the OAI responses
func (request *Request) Harvest(batchCallback func(*Response)) {
	// Use Perform to get the OAI response
	oaiResponse := request.Perform()

	// Execute the callback function with the response
	batchCallback(oaiResponse)

	// Check for a resumptionToken
	hasResumptionToken, resumptionToken, completeListSize := oaiResponse.GetResumptionToken()

	// Harvest further if there is a resumption token
	if hasResumptionToken {
		request.Set = ""
		request.MetadataPrefix = ""
		request.From = ""
		request.ResumptionToken = resumptionToken
		request.CompleteListSize = completeListSize
		request.Harvest(batchCallback)
	}
}

// Perform an HTTP GET request using the OAI Requests fields
// and return an OAI Response reference
func (request *Request) Perform() (oaiResponse *Response) {
	timeout := time.Duration(60 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	err := retry(40, time.Second, func() error {
		req, err := http.NewRequest(http.MethodGet, request.GetFullURL(), nil)
		if err != nil {
			return err
		}

		if request.UserName != "" && request.Password != "" {
			req.SetBasicAuth(request.UserName, request.Password)
		}

		resp, err := client.Do(req)
		if err != nil {
			return err
		}

		// Make sure the response body object will be closed after
		// reading all the content body's data
		defer resp.Body.Close()

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			slog.Error("unable to read body", "err", err, "url", request.GetFullURL())
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
			err = xml.Unmarshal(data, &oaiResponse)
			if err != nil {
				l.Error("unable to unmarshal OAI-PMH response", "error", err, "response", string(data))
				return stop{err}
			}

			return nil
		}
	})
	if err != nil {
		// unable to harvest panic for now
		slog.Error("unable to finish oai-pmh harvest", "error", err, "url", request.GetFullURL())
		panic(err)
	}
	return
}
