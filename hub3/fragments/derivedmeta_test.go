package fragments

import (
	"encoding/json"
	"strings"
	"testing"
)

// The flag is our conclusion about a record, so it has to be set from what the
// record actually carries -- and it has to keep meaning the same thing as the
// v1 legacy field for as long as both are served.
func TestSetDerivedMetaHasDigitalObject(t *testing.T) {
	for _, tt := range []struct {
		name   string
		fields map[string][]string
		want   bool
	}{
		{"isShownBy present", map[string][]string{"edm_isShownBy": {"http://x/1.jpg"}}, true},
		{"no media fields at all", map[string][]string{"dc_title": {"Kerk"}}, false},
		{"empty graph", map[string][]string{}, false},
		{
			// Documents the known imperfection rather than asserting it is
			// right: v1 does the same, and changing it has to change both.
			"object without isShownBy still counts as none",
			map[string][]string{"edm_object": {"http://x/1.jpg"}},
			false,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			fg := &FragmentGraph{Meta: &Header{}, Fields: tt.fields}
			fg.setDerivedMeta()
			if got := fg.Meta.GetHasDigitalObject(); got != tt.want {
				t.Errorf("HasDigitalObject = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetDerivedMetaWithoutMeta(t *testing.T) {
	fg := &FragmentGraph{Fields: map[string][]string{"edm_isShownBy": {"x"}}}
	fg.setDerivedMeta() // must not panic on a graph without a header
}

// The flag is only useful if a filter on it reaches it. delving_hasDigitalObject
// failed here: the filter went looking through the nested resources.entries,
// where a field we derive has never been, and returned nothing. Under meta. the
// builder takes a flat term query instead, which is the whole reason for
// putting the flag there.
func TestMetaFilterIsFlat(t *testing.T) {
	for _, tt := range []struct {
		name   string
		filter string
		want   string
	}{
		{
			"a meta filter is a flat term query",
			"meta.hasDigitalObject:true",
			`{"bool":{"must":{"term":{"meta.hasDigitalObject":"true"}}}}`,
		},
		{
			"and so is false",
			"meta.hasDigitalObject:false",
			`{"bool":{"must":{"term":{"meta.hasDigitalObject":"false"}}}}`,
		},
		{
			// The contrast that explains the change: this one is why the
			// v1 field could not be filtered on from here.
			"an ordinary field still goes through the nested entries",
			"dc_title:Kerk",
			"resources.entries",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			qf, err := NewQueryFilter(tt.filter)
			if err != nil {
				t.Fatalf("NewQueryFilter(%q): %v", tt.filter, err)
			}

			q, err := qf.ElasticFilter()
			if err != nil {
				t.Fatalf("ElasticFilter(): %v", err)
			}

			src, err := q.Source()
			if err != nil {
				t.Fatalf("Source(): %v", err)
			}

			b, err := json.Marshal(src)
			if err != nil {
				t.Fatalf("Marshal(): %v", err)
			}

			got := string(b)
			if strings.HasPrefix(tt.want, "{") {
				if got != tt.want {
					t.Errorf("filter %q\n  got  %s\n  want %s", tt.filter, got, tt.want)
				}
				return
			}
			if !strings.Contains(got, tt.want) {
				t.Errorf("filter %q: expected %q somewhere in\n  %s", tt.filter, tt.want, got)
			}
		})
	}
}

// The flag has to survive marshalling, including when it is false.
//
// A plain proto3 bool would not: the generated json tag carries omitempty, so
// false dropped out of the document entirely and a filter on
// meta.hasDigitalObject:false matched nothing -- half of what #3052 asks for,
// failing silently. The field is declared optional for that reason, and this
// guards it without needing a live Elasticsearch to notice.
func TestHasDigitalObjectSurvivesMarshalling(t *testing.T) {
	for _, tt := range []struct {
		name   string
		fields map[string][]string
		want   string
	}{
		{"true is written", map[string][]string{"edm_isShownBy": {"x"}}, `"hasDigitalObject":true`},
		{"false is written too", map[string][]string{"dc_title": {"y"}}, `"hasDigitalObject":false`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			fg := &FragmentGraph{Meta: &Header{HubID: "h"}, Fields: tt.fields}

			msg, err := fg.IndexMessage()
			if err != nil {
				t.Fatalf("IndexMessage: %v", err)
			}

			if got := string(msg.Source); !strings.Contains(got, tt.want) {
				t.Errorf("expected %s in the document, got\n  %s", tt.want, got)
			}
		})
	}
}
