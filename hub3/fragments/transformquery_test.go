package fragments

import "testing"

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
