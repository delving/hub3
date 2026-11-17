package fragments

import (
	"encoding/json"
	"reflect"
	"testing"

	elastic "github.com/olivere/elastic/v7"
)

const (
	role   = "https://schema.org/Role"
	person = "https://schema.org/Person"
	org    = "https://schema.org/Organization"
)

func TestParseResourcePath(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want ResourcePath
	}{
		{"[]schema:creator->[schema:Role]schema:creator->[schema:Person|schema:Organization]schema:name", ResourcePath{
			Hops:     []PathHop{{SearchLabel: "schema_creator"}, {"schema_creator", []string{role}}},
			Classes:  []string{person, org},
			Property: "schema_name",
		}},
		{"[schema:VisualArtwork]schema:creator->[]schema:name", ResourcePath{
			Hops:     []PathHop{{"schema_creator", []string{"https://schema.org/VisualArtwork"}}},
			Property: "schema_name",
		}},
		{"[]schema:creator->[schema:Person]", ResourcePath{Hops: []PathHop{{SearchLabel: "schema_creator"}}, Classes: []string{person}}},
		{"schema:name", ResourcePath{Property: "schema_name"}},
		{"[" + person + "]schema:name", ResourcePath{Classes: []string{person}, Property: "schema_name"}},
	} {
		got, err := ParseResourcePath(tc.in)
		if err != nil {
			t.Errorf("%s: %v", tc.in, err)
			continue
		}
		if !reflect.DeepEqual(*got, tc.want) {
			t.Errorf("%s:\n got %+v\nwant %+v", tc.in, *got, tc.want)
		}
	}
	for _, bad := range []string{"", "[schema:Person", "[]->[schema:Person]schema:name", "[nope:X]schema:name",
		"[schema:Person||schema:Organization]schema:name", "[]nope:creator"} {
		if _, err := ParseResourcePath(bad); err == nil {
			t.Errorf("%q parsed", bad)
		}
	}
}

func sourceJSON(t *testing.T, q elastic.Query) string {
	t.Helper()
	src, err := q.Source()
	if err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(src)
	return string(b)
}

func asJSON(t *testing.T, v any) string {
	t.Helper()
	b, _ := json.Marshal(v)
	return string(b)
}

// nested is a nested bool query as olivere writes it: one must clause as
// an object, more as a list.
func nested(path string, must ...any) map[string]any {
	var m any = must
	if len(must) == 1 {
		m = must[0]
	}
	return map[string]any{"nested": map[string]any{"path": path, "query": map[string]any{"bool": map[string]any{"must": m}}}}
}

func term(field string, v any) map[string]any {
	return map[string]any{"term": map[string]any{field: v}}
}

func terms(field string, vs ...any) map[string]any {
	return map[string]any{"terms": map[string]any{field: vs}}
}

// The resource at the end of a path is one nested resources element: each hop
// is a context condition at its level, the class after a hop is the
// SubjectClass of the next level's context or, at the end, the resource's
// own types, and the value is an entry one level below the resource.
func TestResourcePathQueries(t *testing.T) {
	path, err := ParseResourcePath("[]schema:creator->[schema:Role]schema:creator->[schema:Person|schema:Organization]schema:name")
	if err != nil {
		t.Fatal(err)
	}
	hops := []any{
		nested("resources.context", term("resources.context.level", 1), term("resources.context.searchLabel", "schema_creator")),
		nested("resources.context", term("resources.context.level", 2), term("resources.context.searchLabel", "schema_creator"),
			terms("resources.context.subjectClass", role)),
		terms("resources.types", person, org),
	}
	entry := func(values ...any) any {
		must := []any{term("resources.entries.searchLabel", "schema_name"), term("resources.entries.level", 3)}
		if len(values) > 0 {
			must = append(must, terms("resources.entries.@value.keyword", values...))
		}
		return nested("resources.entries", must...)
	}

	for name, tc := range map[string]struct {
		q    elastic.Query
		want any
	}{
		"value":     {path.ValueQuery("Sphinx"), nested("resources", append(append([]any{}, hops...), entry("Sphinx"))...)},
		"one of":    {path.ValueQuery("Sphinx", "Bram"), nested("resources", append(append([]any{}, hops...), entry("Sphinx", "Bram"))...)},
		"has value": {path.ValueQuery(), nested("resources", append(append([]any{}, hops...), entry())...)},
	} {
		if got, want := sourceJSON(t, tc.q), asJSON(t, tc.want); got != want {
			t.Errorf("%s:\n got %s\nwant %s", name, got, want)
		}
	}

	res := mustParse(t, "[]schema:creator->[schema:Person]")
	hop := nested("resources.context", term("resources.context.level", 1), term("resources.context.searchLabel", "schema_creator"))
	for name, tc := range map[string]struct {
		q    elastic.Query
		want any
	}{
		"is a Person":       {res.ValueQuery(), nested("resources", hop, terms("resources.types", person))},
		"is that Person":    {res.ValueQuery("http://x/p1"), nested("resources", hop, terms("resources.types", person), terms("resources.id", "http://x/p1"))},
		"same resource":     {res.ResourceQuery(EntryQuery("schema_name", 2, "Sphinx")), nested("resources", hop, terms("resources.types", person), nested("resources.entries", term("resources.entries.searchLabel", "schema_name"), term("resources.entries.level", 2), terms("resources.entries.@value.keyword", "Sphinx")))},
		"the object itself": {mustParse(t, "schema:name").ValueQuery("Molen"), nested("resources", nested("resources.entries", term("resources.entries.searchLabel", "schema_name"), term("resources.entries.level", 1), terms("resources.entries.@value.keyword", "Molen")))},
		"root class": {mustParse(t, "[schema:VisualArtwork]schema:creator->[]schema:name").ValueQuery("Sphinx"), nested("resources",
			nested("resources.context", term("resources.context.level", 1), term("resources.context.searchLabel", "schema_creator"),
				terms("resources.context.subjectClass", "https://schema.org/VisualArtwork")),
			nested("resources.entries", term("resources.entries.searchLabel", "schema_name"), term("resources.entries.level", 2), terms("resources.entries.@value.keyword", "Sphinx")))},
	} {
		if got, want := sourceJSON(t, tc.q), asJSON(t, tc.want); got != want {
			t.Errorf("%s:\n got %s\nwant %s", name, got, want)
		}
	}
}

func mustParse(t *testing.T, s string) *ResourcePath {
	t.Helper()
	p, err := ParseResourcePath(s)
	if err != nil {
		t.Fatal(err)
	}
	return p
}
