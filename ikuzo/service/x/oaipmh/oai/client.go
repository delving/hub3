package oaipmh

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Request represents a request URL and query string to an OAI-PMH service
type Request struct {
	BaseURL          string
	Verb             string
	MetadataPrefix   string
	Set              string
	From             string
	Until            string
	ResumptionToken  string
	Identifier       string
	completeListSize int
}

// NewRequest builds a Request from the query paramers of the http.Request
//
// This is used to build a server-side request. The client needs to build the
// Request directly using the struct.
func NewRequest(r *http.Request) Request {
	baseURL := fmt.Sprintf("%s://%s%s", r.URL.Scheme, r.Host, r.URL.Path)
	q := r.URL.Query()
	req := Request{
		Verb:            q.Get("verb"),
		MetadataPrefix:  q.Get("metadataPrefix"),
		Set:             q.Get("set"),
		From:            q.Get("from"),
		Until:           q.Get("until"),
		Identifier:      q.Get("identifier"),
		ResumptionToken: q.Get("resumptionToken"),
		BaseURL:         baseURL,
	}

	return req
}

// GetFullURL represents the OAI Request in a string format
func (request *Request) GetFullURL() string {
	array := []string{}

	add := func(name, value string) {
		if value != "" {
			array = append(array, name+"="+value)
		}
	}

	add("verb", request.Verb)
	add("set", request.Set)
	add("metadataPrefix", request.MetadataPrefix)
	add("resumptionToken", url.QueryEscape(request.ResumptionToken))
	add("identifier", request.Identifier)
	add("from", request.From)
	add("until", request.Until)

	URL := strings.Join([]string{request.BaseURL, "?", strings.Join(array, "&")}, "")

	return URL
}

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
	if hasResumptionToken == true {
		request.Set = ""
		request.MetadataPrefix = ""
		request.From = ""
		request.ResumptionToken = resumptionToken
		request.completeListSize = completeListSize
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

	err := retry(10, time.Second, func() error {

		resp, err := client.Get(request.GetFullURL())
		if err != nil {
			return err
		}

		// Make sure the response body object will be closed after
		// reading all the content body's data
		defer resp.Body.Close()

		s := resp.StatusCode
		switch {
		case s >= 500:
			// Retry
			return fmt.Errorf("server error: %v", s)
		case s == 408:
			// Retry
			return fmt.Errorf("Timeout error: %v", s)
		case s >= 400:
			// Don't retry, it was client's fault
			return stop{fmt.Errorf("client error: %v", s)}
		default:
			// Happy
			// Read all the data
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return stop{err}
			}

			// Unmarshall all the data
			err = xml.Unmarshal(body, &oaiResponse)
			if err != nil {
				return stop{err}
			}

			return nil
		}

	})
	if err != nil {
		// unable to harvest panic for now
		log.Printf("problem url: %s", request.GetFullURL())
		panic(err)
	}
	return
}
