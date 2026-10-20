package fragments

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	c "github.com/delving/hub3/config"
	elastic "github.com/olivere/elastic/v7"
)

// The tiebreak after the sort field is meta.hubID descending, as in nave
// (legacy.delving_hubId desc), so equal sort values come back in nave's order.
func TestSortTiebreakIsHubIDDescending(t *testing.T) {
	c.Config.ElasticSearch.OrgIDKey = "meta.orgID"

	for _, sortBy := range []string{"", "nave_notes", "_score"} {
		var body struct {
			Sort []map[string]json.RawMessage `json:"sort"`
		}

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(b, &body)
			io.WriteString(w, `{"hits":{"hits":[]}}`)
		}))

		ec, err := elastic.NewSimpleClient(elastic.SetURL(srv.URL))
		if err != nil {
			t.Fatal(err)
		}

		params := url.Values{}
		if sortBy != "" {
			params.Set("sortBy", sortBy)
		}

		sr, err := NewSearchRequest("test", params)
		if err != nil {
			t.Fatal(err)
		}

		s, _, err := sr.ElasticSearchService(ec)
		if err != nil {
			t.Fatal(err)
		}

		_, _ = s.Do(context.Background())
		srv.Close()

		if len(body.Sort) != 2 {
			t.Fatalf("sortBy %q: want field sort plus tiebreak, got %v", sortBy, body.Sort)
		}

		var tie struct {
			Order string `json:"order"`
		}
		if err := json.Unmarshal(body.Sort[1]["meta.hubID"], &tie); err != nil || tie.Order != "desc" {
			t.Errorf("sortBy %q: tiebreak %s, want meta.hubID desc", sortBy, body.Sort[1])
		}
	}
}
