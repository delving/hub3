package elasticsearch

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/go-chi/chi/v5"
	"github.com/mailgun/groupcache"
	"github.com/matryer/is"
	"github.com/rs/zerolog"
)

func TestProxyPassesElasticsearchAnswer(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
	}{
		{"result", http.StatusOK, `{"hits":{"total":{"value":0},"hits":[]}}`},
		// An error must reach the client with its status and body, not as an
		// empty 200 (which was then cached).
		{"error", http.StatusBadRequest, `{"error":{"type":"search_phase_execution_exception"},"status":400}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := is.New(t)

			es := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Elastic-Product", "Elasticsearch")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.status)
				io.WriteString(w, tt.body)
			}))
			defer es.Close()

			client, err := elasticsearch.NewClient(elasticsearch.Config{Addresses: []string{es.URL}})
			is.NoErr(err)

			p := &Proxy{es: client, log: zerolog.Nop()}
			p.group = groupcache.NewGroup("esRemote-test-"+tt.name, 1<<20, groupcache.GetterFunc(p.retrieveFromElasticSearch))

			router := chi.NewRouter()
			router.Handle("/{index}/_search", p)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/idx/_search", strings.NewReader(`{"size":0}`)))

			is.Equal(w.Code, tt.status)
			is.Equal(w.Body.String(), tt.body)
			is.Equal(w.Header().Get("X-Elastic-Product"), "Elasticsearch")
		})
	}
}
