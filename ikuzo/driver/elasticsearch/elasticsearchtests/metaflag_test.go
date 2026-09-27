package elasticsearchtests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/ikuzo/driver/elasticsearch/internal/mapping"
)

// End to end for meta.hasDigitalObject (#3052), against a real Elasticsearch.
//
// Three things are asserted that a unit test cannot reach:
//
//   - the update mapping is accepted by the _mapping endpoint. That is the path
//     existing indices take, and the index is strict on dynamic fields, so
//     without the field there a document carrying it is rejected rather than
//     stored without it.
//   - a document built by the real IndexMessage path carries the flag.
//   - the filter finds it. The filter sends the string "true" against a boolean
//     field; Elasticsearch coerces that, but it is an assumption worth a test
//     rather than a belief, especially as the v1 field is a string in the index
//     and this one is not.
func TestMetaHasDigitalObject(t *testing.T) {
	base := fmt.Sprintf("http://%s", hostAndPort)
	index := "hub3test-metaflag"

	do := func(method, path, body string) (int, string) {
		t.Helper()

		var r io.Reader
		if body != "" {
			r = strings.NewReader(body)
		}

		req, err := http.NewRequest(method, base+path, r)
		if err != nil {
			t.Fatalf("%s %s: %v", method, path, err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("%s %s: %v", method, path, err)
		}
		defer resp.Body.Close()

		b, _ := io.ReadAll(resp.Body)

		return resp.StatusCode, string(b)
	}

	_, _ = do(http.MethodDelete, "/"+index, "")

	if code, body := do(http.MethodPut, "/"+index, mapping.V2ESMapping(1, 0)); code != http.StatusOK {
		t.Fatalf("creating the index: %d %s", code, body)
	}

	t.Cleanup(func() { _, _ = do(http.MethodDelete, "/"+index, "") })

	// The path every existing index takes. Before the field was added here,
	// this left the index unable to accept a record that has it.
	if code, body := do(http.MethodPut, "/"+index+"/_mapping", mapping.V2MappingUpdate()); code != http.StatusOK {
		t.Fatalf("applying the update mapping: %d %s", code, body)
	}

	// On the resources, not on Fields: edm_isShownBy is a URI and the flattened
	// Fields map holds literals only, which is what made the first version of
	// this call every record media-less.
	for _, rec := range []struct {
		hubID  string
		labels []string
	}{
		{"test_spec_with", []string{"dc_title", "edm_isShownBy"}},
		{"test_spec_without", []string{"dc_title"}},
	} {
		entries := make([]*fragments.ResourceEntry, 0, len(rec.labels))
		for _, l := range rec.labels {
			entries = append(entries, &fragments.ResourceEntry{SearchLabel: l, Value: "v"})
		}

		fg := &fragments.FragmentGraph{
			Meta: &fragments.Header{
				OrgID:   "test",
				Spec:    "spec",
				HubID:   rec.hubID,
				DocType: fragments.FragmentGraphDocType,
			},
			Resources: []*fragments.FragmentResource{{Entries: entries}},
		}

		msg, err := fg.IndexMessage()
		if err != nil {
			t.Fatalf("IndexMessage for %s: %v", rec.hubID, err)
		}

		code, body := do(http.MethodPut, "/"+index+"/_doc/"+rec.hubID, string(msg.Source))
		if code != http.StatusCreated && code != http.StatusOK {
			t.Fatalf("indexing %s: %d %s", rec.hubID, code, body)
		}
	}

	if code, body := do(http.MethodPost, "/"+index+"/_refresh", ""); code != http.StatusOK {
		t.Fatalf("refresh: %d %s", code, body)
	}

	// The string "true" against a boolean field, exactly as the filter builder
	// produces it.
	for _, tt := range []struct {
		value string
		want  int
	}{
		{"true", 1},
		{"false", 1},
	} {
		query := fmt.Sprintf(
			`{"query":{"term":{"meta.hasDigitalObject":%q}}}`, tt.value)

		code, body := do(http.MethodPost, "/"+index+"/_search", query)
		if code != http.StatusOK {
			t.Fatalf("searching for %s: %d %s", tt.value, code, body)
		}

		var res struct {
			Hits struct {
				Total struct {
					Value int `json:"value"`
				} `json:"total"`
				Hits []struct {
					ID string `json:"_id"`
				} `json:"hits"`
			} `json:"hits"`
		}
		if err := json.NewDecoder(bytes.NewReader([]byte(body))).Decode(&res); err != nil {
			t.Fatalf("decoding the answer: %v", err)
		}

		if got := res.Hits.Total.Value; got != tt.want {
			t.Errorf("meta.hasDigitalObject:%s matched %d records, want %d\n%s",
				tt.value, got, tt.want, body)
		}

		if len(res.Hits.Hits) == 1 {
			wantID := "test_spec_with"
			if tt.value == "false" {
				wantID = "test_spec_without"
			}

			if got := res.Hits.Hits[0].ID; got != wantID {
				t.Errorf("meta.hasDigitalObject:%s matched %q, want %q", tt.value, got, wantID)
			}
		}
	}
}
