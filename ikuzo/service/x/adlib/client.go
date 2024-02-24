package adlib

import (
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

// Client is used to interact with the Adlib/Axiel API
type Client struct {
	url string
	c   *http.Client
}

func (c *Client) Timeout(duration time.Duration) {
	c.c.Timeout = duration
}

func New(url string) *Client {
	retryClient := retryablehttp.NewClient()
	retryClient.Logger = nil
	retryClient.RetryMax = 3

	return &Client{
		url: url,
		c:   retryClient.StandardClient(),
	}
}
