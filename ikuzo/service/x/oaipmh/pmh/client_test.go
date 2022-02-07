package pmh

import (
	"testing"
	"time"

	"github.com/matryer/is"
)

// nolint:gocritic
func TestClient(t *testing.T) {
	testURL := "http://localhost:8000/api/oai-pmh?"

	t.Run("setting up client", func(t *testing.T) {
		is := is.New(t)

		c, err := NewClient(testURL)
		is.NoErr(err)
		is.Equal(c.baseURL.Path, "/api/oai-pmh")

		is.True(c.HTTPClient.Timeout == 60*time.Second) // default timeout should be sixty seconds
	})

	// t.Run("test identify", func(t *testing.T) {
	// is := is.New(t)
	// c := NewClient(testURL)

	// resp, err := c.Identify()
	// is.NoErr(err)

	// })
}
