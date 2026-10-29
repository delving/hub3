package models

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/knakk/sparql"
)

func TestGetNDEInfo(t *testing.T) {
	// a full first page forces a second request; the short second page ends the paging
	var queries []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		queries = append(queries, r.Form.Get("query"))

		n := 2
		if len(queries) == 1 {
			n = ndeInfoPageSize
		}

		rows := make([]string, n)
		for i := range rows {
			rows[i] = fmt.Sprintf(`{"spec":{"type":"literal","value":"spec-%d-%d"},"description":{"type":"literal","value":"Description %d"}}`, len(queries), i, i)
		}

		w.Header().Set("Content-Type", "application/sparql-results+json")
		fmt.Fprintf(w, `{"head":{"vars":["spec","description"]},"results":{"bindings":[%s]}}`, strings.Join(rows, ","))
	}))
	defer srv.Close()

	repo, err := sparql.NewRepo(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	infos, err := getNDEInfo(repo)
	if err != nil {
		t.Fatal(err)
	}

	if len(queries) != 2 || !strings.Contains(queries[1], fmt.Sprintf("OFFSET %d", ndeInfoPageSize)) {
		t.Fatalf("expected two paged queries, got %q", queries)
	}

	if len(infos) != ndeInfoPageSize+2 {
		t.Fatalf("got %d infos, want %d", len(infos), ndeInfoPageSize+2)
	}

	if got := infos[ndeInfoPageSize]; got.Spec != "spec-2-0" || got.Description != "Description 0" {
		t.Errorf("unexpected info %+v", got)
	}
}
