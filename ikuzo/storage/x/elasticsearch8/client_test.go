package elasticsearch8

import (
	"testing"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/rs/zerolog"
)

func TestNewClient(t *testing.T) {
	t.Run("with defaults", func(t *testing.T) {
		c, err := NewClient(Config{
			Logger: zerolog.Nop(),
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if c.es == nil {
			t.Fatal("expected non-nil elasticsearch client")
		}

		if c.indexPrefix != "" {
			t.Errorf("expected empty index prefix, got %q", c.indexPrefix)
		}
	})

	t.Run("with index prefix", func(t *testing.T) {
		c, err := NewClient(Config{
			IndexPrefix: "test",
			Logger:      zerolog.Nop(),
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if c.indexPrefix != "test" {
			t.Errorf("expected index prefix %q, got %q", "test", c.indexPrefix)
		}
	})
}

func TestIndexName(t *testing.T) {
	tests := []struct {
		name        string
		indexPrefix string
		orgID       string
		want        string
	}{
		{
			name:  "lowercase orgID without prefix",
			orgID: "museum",
			want:  "museumv2",
		},
		{
			name:  "mixed-case orgID without prefix",
			orgID: "Museum",
			want:  "museumv2",
		},
		{
			name:        "with prefix",
			indexPrefix: "test",
			orgID:       "museum",
			want:        "test_museumv2",
		},
		{
			name:        "with prefix and mixed-case",
			indexPrefix: "test",
			orgID:       "Museum",
			want:        "test_museumv2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{indexPrefix: tt.indexPrefix}

			got := c.IndexName(tt.orgID)
			if got != tt.want {
				t.Errorf("IndexName(%q) = %q, want %q", tt.orgID, got, tt.want)
			}
		})
	}
}

func TestNewClientFromExisting(t *testing.T) {
	es, err := elasticsearch.NewDefaultClient()
	if err != nil {
		t.Fatalf("unexpected error creating default client: %v", err)
	}

	c := NewClientFromExisting(es, zerolog.Nop())

	if c.es != es {
		t.Error("expected wrapped client to be the same instance")
	}

	if c.ES() != es {
		t.Error("ES() should return the underlying client")
	}
}
