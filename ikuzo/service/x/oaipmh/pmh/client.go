package pmh

import (
	"net/http"
	"net/url"
	"time"
)

// Client
type Client struct {

	// baseURL is the target URL for the OAI-PMH repository.
	// All query params are ignored and reset
	baseURL *url.URL

	// HTTPClient is the default http.Client with 60 second timeout.
	// The timeout can be increased by providing your own http.Client
	HTTPClient http.Client
}

func NewClient(endpoint string) (Client, error) {
	client := http.Client{
		Timeout: 60 * time.Second,
	}

	u, err := url.ParseRequestURI(endpoint)
	if err != nil {
		return Client{}, err
	}

	return Client{
		baseURL:    u,
		HTTPClient: client,
	}, nil
}

// func (c Client) Identify() (Identify, error) {

// }
