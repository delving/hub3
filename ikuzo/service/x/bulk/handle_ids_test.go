package bulk

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/olivere/elastic/v7"
)

func idHit(hubID string) *elastic.SearchHit {
	return &elastic.SearchHit{
		Source: json.RawMessage(fmt.Sprintf(`{"meta":{"hubID":%q}}`, hubID)),
		Sort:   []interface{}{hubID},
	}
}

func TestWriteIndexIDs(t *testing.T) {
	pages := [][]*elastic.SearchHit{
		{idHit("a"), idHit("b")},
		{{Source: json.RawMessage(`{"meta":{}}`), Sort: []interface{}{"b2"}}, idHit("c")},
		{},
	}

	var calls [][]interface{}

	w := httptest.NewRecorder()

	seen, err := writeIndexIDs(w, func(searchAfter []interface{}) ([]*elastic.SearchHit, error) {
		calls = append(calls, searchAfter)
		page := pages[len(calls)-1]

		return page, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if seen != 3 {
		t.Errorf("seen = %d, want 3", seen)
	}

	if got, want := w.Body.String(), "a\nb\nc\n"; got != want {
		t.Errorf("body = %q, want %q", got, want)
	}

	if !w.Flushed {
		t.Error("expected the response to be flushed")
	}

	wantAfter := []interface{}{nil, "b", "c"}
	if len(calls) != len(wantAfter) {
		t.Fatalf("fetch called %d times, want %d", len(calls), len(wantAfter))
	}

	for i, want := range wantAfter {
		if want == nil {
			if calls[i] != nil {
				t.Errorf("call %d searchAfter = %v, want nil", i, calls[i])
			}

			continue
		}

		if len(calls[i]) != 1 || calls[i][0] != want {
			t.Errorf("call %d searchAfter = %v, want [%v]", i, calls[i], want)
		}
	}
}

func TestWriteIndexIDsErrorAfterFirstPage(t *testing.T) {
	boom := errors.New("boom")
	call := 0

	w := httptest.NewRecorder()

	seen, err := writeIndexIDs(w, func(_ []interface{}) ([]*elastic.SearchHit, error) {
		call++
		if call == 1 {
			return []*elastic.SearchHit{idHit("a")}, nil
		}

		return nil, boom
	})
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}

	if seen != 1 {
		t.Errorf("seen = %d, want 1", seen)
	}

	if got := w.Body.String(); got != "a\n" {
		t.Errorf("body = %q, want %q", got, "a\n")
	}
}
