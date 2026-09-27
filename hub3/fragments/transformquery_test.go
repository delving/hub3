package fragments

import (
	"net/url"
	"testing"
)

// The cases come from the parity corpus -- real queries out of five years of
// access log -- plus the shapes that make a naive rewrite dangerous: a colon
// inside a quoted value, a URI, and a name that is already qualified.
func TestTransformQuery(t *testing.T) {
	for _, tt := range []struct {
		name string
		in   string
		want string
	}{
		{
			"bare field gets qualified",
			`dc_title:"Albert Neuhuys"`,
			`fields.dc_title:"Albert Neuhuys"`,
		},
		{
			"the _text suffix keeps its own rule",
			`dc_title_text:"Albert Neuhuys"`,
			`fields.dc_title:"Albert Neuhuys"`,
		},
		{
			"delving_spec still becomes meta.spec, not fields.delving_spec",
			`delving_spec:museum-helmond-vervaardigers`,
			`meta.spec:museum-helmond-vervaardigers`,
		},
		{
			"the corpus case that returned nothing",
			`delving_spec:museum-helmond-vervaardigers AND dc_title:"Albert Neuhuys"`,
			`meta.spec:museum-helmond-vervaardigers AND fields.dc_title:"Albert Neuhuys"`,
		},
		{
			"an already qualified field is left alone",
			`fields.dc_title:"Albert Neuhuys"`,
			`fields.dc_title:"Albert Neuhuys"`,
		},
		{
			"meta is left alone",
			`meta.spec:bhic`,
			`meta.spec:bhic`,
		},
		{
			"a colon inside a value is not a field",
			`dc_identifier:"inv: 1998-042"`,
			`fields.dc_identifier:"inv: 1998-042"`,
		},
		{
			"a URI is not a field",
			`dc_identifier:http://example.org/thing`,
			`fields.dc_identifier:http://example.org/thing`,
		},
		{
			"a plain term is untouched",
			`kerk`,
			`kerk`,
		},
		{
			"several fields in one query",
			`dc_creator:Neuhuys OR dc_subject:portret`,
			`fields.dc_creator:Neuhuys OR fields.dc_subject:portret`,
		},
		{
			"a field inside parentheses",
			`(dc_title:kerk OR dc_title:kapel)`,
			`(fields.dc_title:kerk OR fields.dc_title:kapel)`,
		},
		{
			"an unbalanced quote is left as it stands",
			`dc_title:"onafgemaakt`,
			`fields.dc_title:"onafgemaakt`,
		},
		{
			"empty stays empty",
			``,
			``,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := transformQuery(tt.in); got != tt.want {
				t.Errorf("transformQuery(%q)\n  got  %q\n  want %q", tt.in, got, tt.want)
			}
		})
	}
}

// Sort direction. Nothing in five years of access log asks for a direction, so
// what a bare sortBy does is what every real request gets.
func TestSortDirection(t *testing.T) {
	for _, tt := range []struct {
		name    string
		params  url.Values
		wantAsc bool
	}{
		{"a named field sorts ascending, like Django",
			url.Values{"sortBy": {"tib_notes"}}, true},
		{"sortOrder=desc still gets descending",
			url.Values{"sortBy": {"tib_notes"}, "sortOrder": {"desc"}}, false},
		{"sortOrder=asc is the default said out loud",
			url.Values{"sortBy": {"tib_notes"}, "sortOrder": {"asc"}}, true},
		{"the caret keeps working",
			url.Values{"sortBy": {"^tib_notes"}}, true},
		{"_score stays descending: the best match comes first",
			url.Values{"sortBy": {"_score"}}, false},
		{"random stays descending",
			url.Values{"sortBy": {"random_180"}}, false},
		{"no sortBy leaves the flag alone",
			url.Values{}, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			sr, err := NewSearchRequest("brabantcloud", tt.params)
			if err != nil {
				t.Fatalf("NewSearchRequest: %v", err)
			}
			if sr.SortAsc != tt.wantAsc {
				t.Errorf("sortBy=%q sortOrder=%q: SortAsc = %v, want %v",
					tt.params.Get("sortBy"), tt.params.Get("sortOrder"), sr.SortAsc, tt.wantAsc)
			}
		})
	}
}
