package oaipmh

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/delving/hub3/ikuzo/domain"
)

// Request is indicating the protocol request that generated this response.
//
// The rules for generating the request element are as follows:
// 1. The content of the request element must always be the base URL of the protocol request;
// 2. The only valid attributes for the request element are the keys of the key=value pairs of protocol request. The attribute values must be the corresponding values of those key=value pairs;
// 3. In cases where the request that generated this response did not result in an error or exception condition, the attributes and attribute values of the request element must match the key=value pairs of the protocol request;
// 4. In cases where the request that generated this response resulted in a badVerb or badArgument error condition, the repository must return the base URL of the protocol request only. Attributes must not be provided in these cases.
//
// http://www.openarchives.org/OAI/openarchivesprotocol.html#XMLResponse
type Request struct {
	Verb             string `xml:"verb,attr,omitempty"`
	Identifier       string `xml:"identifier,attr,omitempty"`
	MetadataPrefix   string `xml:"metadataPrefix,attr,omitempty"`
	From             string `xml:"from,attr,omitempty"`
	Until            string `xml:"until,attr,omitempty"`
	Set              string `xml:"set,attr,omitempty"`
	ResumptionToken  string `xml:"resumptionToken,attr,omitempty"`
	BaseURL          string `xml:",chardata"`
	completeListSize int
	orgConfig        *domain.OrganizationConfig
	// TODO(kiivihal): determine if these can be removed
	cursor int
	limit  int
}

// NewRequest builds a Request from the query paramers of the http.Request
//
// This is used to build a server-side request. The client needs to build the
// Request directly using the struct.
func NewRequest(r *http.Request) Request {
	if r.URL.Scheme == "" {
		r.URL.Scheme = "http"
	}

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

func (request *Request) RequestConfig() RequestConfig {
	return RequestConfig{
		ID:           "",
		FirstRequest: request,
		OrgID:        request.orgConfig.OrgID(),
		DatasetID:    request.Set,
		TotalSize:    0,
		Finished:     false,
	}
}

func (request *Request) rawToken() (RawToken, error) {
	if request.ResumptionToken == "" {
		return RawToken{}, nil
	}

	return parseToken(request.ResumptionToken)
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

	URL := request.BaseURL + "?" + strings.Join(array, "&")

	return URL
}
