package internal

import (
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

func NewClient(retry, timeout int) *retryablehttp.Client {
	c := retryablehttp.NewClient()
	c.RetryMax = retry

	if timeout == 0 {
		timeout = 15
	}

	c.HTTPClient.Timeout = time.Duration(timeout) * time.Second

	return c
}
