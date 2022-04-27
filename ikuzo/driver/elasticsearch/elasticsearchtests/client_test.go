package elasticsearchtests

import (
	"fmt"
	"testing"

	"github.com/delving/hub3/ikuzo/driver/elasticsearch"
	"github.com/matryer/is"
)

func TestNewClient(t *testing.T) {
	//nolint: gocritic
	is := is.New(t)

	host := hostAndPort

	t.Run("empty config client", func(t *testing.T) {
		cfg := &elasticsearch.Config{}
		cfg.Urls = []string{fmt.Sprintf("http://%s", host)}
		cfg.DisableMetrics = true
		client, err := elasticsearch.NewClient(cfg)
		t.Logf("new client error: %s", err)
		is.NoErr(err)
		is.True(client != nil)
	})

	t.Run("new client with default config", func(t *testing.T) {
		cfg := elasticsearch.DefaultConfig()
		cfg.Urls = []string{fmt.Sprintf("http://%s", host)}
		cfg.DisableMetrics = true
		client, err := elasticsearch.NewClient(cfg)
		is.NoErr(err)

		is.True(client != nil) // client should not be empty

		// is.True(client.index != nil)
		// is.True(client.search != nil)
	})
}
