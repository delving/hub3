package esproxy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matryer/is"
)

func TestHandleRoot(t *testing.T) {
	is := is.New(t)
	s := &Service{}

	// An Elasticsearch client asks for JSON and must see Elasticsearch 7.14+.
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()
	s.handleRoot(w, r)

	is.Equal(w.Code, http.StatusOK)
	is.Equal(w.Header().Get("X-Elastic-Product"), "Elasticsearch")

	var info struct {
		Version struct {
			Number string `json:"number"`
		} `json:"version"`
	}
	is.NoErr(json.Unmarshal(w.Body.Bytes(), &info))
	is.Equal(info.Version.Number, "7.17.0")

	// A browser keeps the redirect to /version.
	r = httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Accept", "text/html,application/xhtml+xml")
	w = httptest.NewRecorder()
	s.handleRoot(w, r)

	is.Equal(w.Code, http.StatusFound)
	is.Equal(w.Header().Get("Location"), "/version")
}
