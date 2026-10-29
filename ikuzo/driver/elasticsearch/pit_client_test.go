package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matryer/is"
)

func TestPITClientBothFlavors(t *testing.T) {
	tests := []struct {
		name         string
		version      string
		distribution string
		openPath     string
		openResp     string
		closePath    string
		closeBody    string
	}{
		{
			name:      "elasticsearch",
			version:   "8.15.0",
			openPath:  "/idx/_pit",
			openResp:  `{"id":"es-pit"}`,
			closePath: "/_pit",
			closeBody: `{"id":"es-pit"}`,
		},
		{
			name:         "opensearch",
			version:      "2.19.2",
			distribution: "opensearch",
			openPath:     "/idx/_search/point_in_time",
			openResp:     `{"pit_id":"os-pit","_shards":{},"creation_time":1}`,
			closePath:    "/_search/point_in_time",
			closeBody:    `{"pit_id":["os-pit"]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := is.New(t)

			var gotClose string

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.Method == http.MethodGet && r.URL.Path == "/":
					fmt.Fprintf(w, `{"version":{"number":%q,"distribution":%q}}`, tt.version, tt.distribution)
				case r.Method == http.MethodPost && r.URL.Path == tt.openPath:
					is.Equal(r.URL.Query().Get("keep_alive"), "1m")
					io.WriteString(w, tt.openResp)
				case r.Method == http.MethodDelete && r.URL.Path == tt.closePath:
					body, _ := io.ReadAll(r.Body)
					gotClose = string(body)
					io.WriteString(w, `{}`)
				default:
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
					http.Error(w, "no handler", http.StatusBadRequest)
				}
			}))
			defer srv.Close()

			c, err := NewPITClient([]string{srv.URL}, "", "")
			is.NoErr(err)

			ctx := context.Background()

			ok, err := c.SupportsPIT(ctx)
			is.NoErr(err)
			is.True(ok)

			id, err := c.CreatePIT(ctx, "idx", "1m")
			is.NoErr(err)
			is.Equal(id, map[string]string{"elasticsearch": "es-pit", "opensearch": "os-pit"}[tt.name])

			is.NoErr(c.ClosePIT(ctx, id))

			var want, got any
			is.NoErr(json.Unmarshal([]byte(tt.closeBody), &want))
			is.NoErr(json.Unmarshal([]byte(gotClose), &got))
			is.Equal(got, want)
		})
	}
}

func TestSupportsPITVersions(t *testing.T) {
	is := is.New(t)

	for _, tc := range []struct {
		version, distribution string
		want                  bool
	}{
		{"7.9.3", "", false},
		{"7.10.2", "", true},
		{"9.1.0", "", true},
		{"2.3.0", "opensearch", false},
		{"2.4.0", "opensearch", true},
		{"3.0.0", "opensearch", true},
	} {
		c := &PITClient{info: &VersionInfo{}}
		c.info.Version.Number = tc.version
		c.info.Version.Distribution = tc.distribution

		got, err := c.SupportsPIT(context.Background())
		is.NoErr(err)
		is.Equal(got, tc.want) // tc.version tc.distribution
	}
}
