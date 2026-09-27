package fragments

import "testing"

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
