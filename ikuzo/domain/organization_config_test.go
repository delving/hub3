package domain

import "testing"

// TestSitemapConfigURL covers both URL modes: legacy last-path-segment
// formatting and UseHubID full-id formatting for DIW deep links.
func TestSitemapConfigURL(t *testing.T) {
	tests := []struct {
		name string
		cfg  SitemapConfig
		id   string
		want string
	}{
		{
			name: "legacy: last path segment into RelPathFmt",
			cfg:  SitemapConfig{BaseURL: "https://example.org", RelPathFmt: "/doc/%s"},
			id:   "https://data.example.org/resource/aggregation/spec/local-1",
			want: "https://example.org/doc/local-1",
		},
		{
			name: "legacy: empty RelPathFmt returns id verbatim",
			cfg:  SitemapConfig{BaseURL: "https://example.org"},
			id:   "https://data.example.org/resource/x",
			want: "https://data.example.org/resource/x",
		},
		{
			name: "useHubID: full id query-escaped into RelPathFmt",
			cfg: SitemapConfig{
				BaseURL:    "https://verhaalvanutrecht.nl/collecties/",
				RelPathFmt: "?id=%s",
				UseHubID:   true,
			},
			id:   "leu/collection/local-1",
			want: "https://verhaalvanutrecht.nl/collecties/?id=leu%2Fcollection%2Flocal-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.URL(tt.id); got != tt.want {
				t.Errorf("URL(%q) = %q, want %q", tt.id, got, tt.want)
			}
		})
	}
}

// TestSitemapConfigIndexBase covers the sub-sitemap host override.
func TestSitemapConfigIndexBase(t *testing.T) {
	cfg := SitemapConfig{BaseURL: "https://customer.example"}
	if got := cfg.IndexBase(); got != "https://customer.example" {
		t.Errorf("IndexBase() = %q, want BaseURL fallback", got)
	}
	cfg.IndexBaseURL = "https://hub.example"
	if got := cfg.IndexBase(); got != "https://hub.example" {
		t.Errorf("IndexBase() = %q, want IndexBaseURL", got)
	}
}
