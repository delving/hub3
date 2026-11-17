package fragments

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	r "github.com/kiivihal/rdf2go"

	c "github.com/delving/hub3/config"
)

const (
	ex       = "http://example.org/"
	sName    = "https://schema.org/name"
	rdfTypeP = "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"
)

// levelsMap builds a ResourceMap from "s p o" lines (o is an IRI unless quoted)
// and runs SetContextLevels from ex+"root".
func levelsMap(t *testing.T, lines ...string) *ResourceMap {
	t.Helper()
	g := r.NewGraph("")
	for _, l := range lines {
		f := strings.SplitN(l, " ", 3)
		var o r.Term = r.NewResource(ex + f[2])
		if strings.HasPrefix(f[2], `"`) {
			o = r.NewLiteral(strings.Trim(f[2], `"`))
		}
		p := ex + f[1]
		switch f[1] {
		case "name":
			p = sName
		case "type":
			p, o = rdfTypeP, r.NewResource(ex+f[2])
		}
		g.AddTriple(r.NewResource(ex+f[0]), r.NewResource(p), o)
	}
	rm, err := NewResourceMap("test", g)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := rm.SetContextLevels(ex + "root"); err != nil {
		t.Fatal(err)
	}
	return rm
}

func level(t *testing.T, rm *ResourceMap, id string) int32 {
	t.Helper()
	fr, ok := rm.GetResource(ex + id)
	if !ok {
		t.Fatalf("no resource %s", id)
	}
	return fr.GetLevel()
}

// contextOf renders a resource's context as "subject-predicate-object@level".
func contextOf(t *testing.T, rm *ResourceMap, id string) []string {
	t.Helper()
	fr, _ := rm.GetResource(ex + id)
	out := []string{}
	for _, ctx := range fr.Context {
		out = append(out, fmt.Sprintf("%s-%s-%s@%d",
			strings.TrimPrefix(ctx.Subject, ex), strings.TrimPrefix(ctx.Predicate, ex),
			strings.TrimPrefix(ctx.ObjectID, ex), ctx.Level))
	}
	slices.Sort(out)
	return out
}

// No depth cap: a resource eight links down is level 9, not 1 (F12).
func TestContextLevelsDeepChain(t *testing.T) {
	lines := []string{}
	prev := "root"
	for i := 1; i <= 8; i++ {
		n := fmt.Sprintf("n%d", i)
		lines = append(lines, prev+" next "+n, n+` name "`+n+`"`)
		prev = n
	}
	rm := levelsMap(t, lines...)
	for i := 1; i <= 8; i++ {
		if got := level(t, rm, fmt.Sprintf("n%d", i)); got != int32(i+1) {
			t.Errorf("n%d level %d, want %d", i, got, i+1)
		}
	}
	if got := len(contextOf(t, rm, "n8")); got != 8 {
		t.Errorf("n8 has %d contexts, want 8", got)
	}
}

// A cycle stops at the first visit: A level 2, B level 3 (F9).
func TestContextLevelsCycle(t *testing.T) {
	rm := levelsMap(t, "root knows a", "a knows b", "b knows a", `a name "A"`, `b name "B"`)
	if a, b := level(t, rm, "a"), level(t, rm, "b"); a != 2 || b != 3 {
		t.Errorf("levels a %d b %d, want 2 and 3", a, b)
	}
	want := []string{"a-knows-b@2", "root-knows-a@1"}
	if got := contextOf(t, rm, "b"); !slices.Equal(got, want) {
		t.Errorf("b context %v, want %v", got, want)
	}
}

// Every shortest route is kept, a longer one is not (decision B).
func TestContextLevelsShortestRoutes(t *testing.T) {
	rm := levelsMap(t,
		"root creator p", "root about p", // two routes of length 1
		"root has x", "x mentions p", // a route of length 2: not a context of p
		"p birthPlace h", "x birthPlace h", // h: two shortest routes via p and x
		`p name "P"`, `x name "X"`, `h name "H"`)
	if got, want := contextOf(t, rm, "p"), []string{"root-about-p@1", "root-creator-p@1"}; !slices.Equal(got, want) {
		t.Errorf("p context %v, want %v", got, want)
	}
	want := []string{"p-birthPlace-h@2", "root-about-p@1", "root-creator-p@1", "root-has-x@1", "x-birthPlace-h@2"}
	if got := contextOf(t, rm, "h"); !slices.Equal(got, want) {
		t.Errorf("h context %v, want %v", got, want)
	}
	if got := level(t, rm, "h"); got != 3 {
		t.Errorf("h level %d, want 3", got)
	}
}

// Levels, contexts and their labels do not depend on triple order (F11, F24).
func TestContextLevelsOrderIndependent(t *testing.T) {
	prev := c.Config.RDFTag.Label
	c.Config.RDFTag.Label = []string{sName}
	t.Cleanup(func() { c.Config.RDFTag.Label = prev })

	lines := []string{"root a b", "root c d", "b e d", "d f g", `root name "Root"`, `d name "D"`, `g name "G"`, "root type Work"}
	fwd := levelsMap(t, lines...)
	reversed := slices.Clone(lines)
	slices.Reverse(reversed)
	rev := levelsMap(t, reversed...)
	for _, id := range []string{"b", "d", "g"} {
		if level(t, fwd, id) != level(t, rev, id) || !slices.Equal(contextOf(t, fwd, id), contextOf(t, rev, id)) {
			t.Errorf("%s: forward level %d %v, reversed level %d %v", id,
				level(t, fwd, id), contextOf(t, fwd, id), level(t, rev, id), contextOf(t, rev, id))
		}
	}
	for _, rm := range []*ResourceMap{fwd, rev} {
		d, _ := rm.GetResource(ex + "d")
		for _, ctx := range d.Context {
			if ctx.Subject == ex+"root" && (ctx.Label != "Root" || !slices.Equal(ctx.SubjectClass, []string{ex + "Work"})) {
				t.Errorf("context %s: label %q class %v, want Root and Work", ctx.Predicate, ctx.Label, ctx.SubjectClass)
			}
		}
	}
}

// GetLabel picks by RDFTag.LabelLanguages, not by triple order (F26).
func TestGetLabelLanguagePreference(t *testing.T) {
	prevLabel, prevLangs := c.Config.RDFTag.Label, c.Config.RDFTag.LabelLanguages
	t.Cleanup(func() { c.Config.RDFTag.Label, c.Config.RDFTag.LabelLanguages = prevLabel, prevLangs })
	c.Config.RDFTag.Label = []string{sName, "http://www.w3.org/2004/02/skos/core#prefLabel"}

	names := func(vals ...[2]string) *FragmentResource {
		fr := &FragmentResource{ID: ex + "x", predicates: map[string][]*FragmentEntry{}}
		for _, v := range vals {
			fr.predicates[sName] = append(fr.predicates[sName], &FragmentEntry{Value: v[0], Language: v[1]})
		}
		return fr
	}
	en, nl, nlNL, none, de := [2]string{"oak panel", "en"}, [2]string{"eikenhouten paneel", "nl"},
		[2]string{"paneel", "nl-NL"}, [2]string{"panel", ""}, [2]string{"Eichentafel", "de"}

	for _, tc := range []struct {
		langs []string
		vals  [][2]string
		want  string
	}{
		{nil, [][2]string{en, nl}, "oak panel"}, // default: en, untagged, then the rest
		{nil, [][2]string{nl, en}, "oak panel"},
		{nil, [][2]string{none, nl}, "panel"},
		{nil, [][2]string{de, en}, "oak panel"},
		{nil, [][2]string{de}, "Eichentafel"},
		{nil, [][2]string{nlNL, nl}, "eikenhouten paneel"}, // both unranked: smallest value
		{[]string{"nl"}, [][2]string{nlNL, en}, "paneel"},  // primary subtag
		{[]string{"nl", "", "en"}, [][2]string{en, nl}, "eikenhouten paneel"},
	} {
		c.Config.RDFTag.LabelLanguages = tc.langs
		if got, _ := names(tc.vals...).GetLabel(); got != tc.want {
			t.Errorf("langs %v values %v: got %q, want %q", tc.langs, tc.vals, got, tc.want)
		}
	}

	// predicate order still comes first: a schema:name in any language beats a skos:prefLabel@nl
	c.Config.RDFTag.LabelLanguages = nil
	fr := names(en)
	fr.predicates["http://www.w3.org/2004/02/skos/core#prefLabel"] = []*FragmentEntry{{Value: "voorkeur", Language: "nl"}}
	if got, _ := fr.GetLabel(); got != "oak panel" {
		t.Errorf("got %q, want the schema:name", got)
	}
}
