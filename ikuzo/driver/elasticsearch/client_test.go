package elasticsearch

import (
	"testing"

	"github.com/matryer/is"
)

func TestNewClient(t *testing.T) {
	//nolint: gocritic
	is := is.New(t)

	t.Run("empty config client", func(t *testing.T) {
		client, err := NewClient(&Config{})
		is.True(err != nil)
		is.True(client == nil)
	})

	t.Run("new client with default config", func(t *testing.T) {
		client, err := NewClient(DefaultConfig())
		is.NoErr(err)
		is.True(client != nil) // client should not be empty

		is.True(client.index != nil)
		is.True(client.search != nil)
	})
}
