package mappingxml

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/matryer/is"

	"github.com/delving/hub3/ikuzo/rdf"
	"github.com/delving/hub3/ikuzo/rdf/formats/ntriples"
)

func TestSerialize(t *testing.T) {
	var g *rdf.Graph
	t.Run("test flat", func(t *testing.T) {
		is := is.New(t)
		g = rdf.NewGraph()
		g.UseIndex = true
		g.UseResource = true

		is.Equal(g.Len(), 0)
		f, err := os.Open("testdata/rdf.nt")
		is.NoErr(err)
		defer f.Close()
		_, err = ntriples.Parse(f, g)
		is.NoErr(err)

		iri, err := rdf.NewIRI("https://schema.org/CreativeWork")
		is.NoErr(err)

		cfg := FilterConfig{RDFType: iri}
		var buf bytes.Buffer
		err = Serialize(g, &buf, &cfg)
		is.NoErr(err)

		// os.WriteFile("/tmp/data.xml", buf.Bytes(), os.ModePerm)

		b, err := os.ReadFile("./testdata/rdf.golden.xml")
		is.NoErr(err)
		if diff := cmp.Diff(string(b), buf.String()+"\n"); diff != "" {
			t.Errorf("mapping xml = mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("filterResources", func(t *testing.T) {
		is := is.New(t)

		is.Equal(g.Len(), 28)
		is.Equal(len(g.Resources()), 2)

		iri, err := rdf.NewIRI("https://schema.org/CreativeWork")
		is.NoErr(err)

		cfg := FilterConfig{RDFType: iri}
		filtered := filterResources(g.Resources(), &cfg)
		is.Equal(len(filtered), 1)

		s, err := rdf.NewIRI("https://klek.si/pqx31b/01339n")
		is.NoErr(err)
		cfg = FilterConfig{Subject: s}
		filtered = filterResources(g.Resources(), &cfg)
		is.Equal(len(filtered), 1)
	})
}

func TestExcludePrefixes(t *testing.T) {
	is := is.New(t)
	
	// Create a simple graph for testing
	g := rdf.NewGraph()
	g.UseIndex = true
	g.UseResource = true
	
	// Add some test triples with different prefixes
	subj1, _ := rdf.NewIRI("https://example.com/resource1")
	subj2, _ := rdf.NewIRI("http://schemas.delving.eu/nave/terms/resource2")
	pred1, _ := rdf.NewIRI("https://example.com/predicate1")
	pred2, _ := rdf.NewIRI("http://schemas.delving.eu/nave/terms/predicate2")
	obj, _ := rdf.NewLiteral("test value")
	
	g.Add(rdf.NewTriple(subj1, pred1, obj))
	g.Add(rdf.NewTriple(subj2, pred1, obj))
	g.Add(rdf.NewTriple(subj1, pred2, obj))
	
	// Test filtering subjects with excluded prefix
	cfg := &FilterConfig{ExcludePrefixes: []string{"http://schemas.delving.eu/nave/terms/"}}
	filtered := filterResources(g.Resources(), cfg)
	
	// Should only have the resource with subject1 (not the one with delving prefix)
	is.Equal(len(filtered), 1)
	is.Equal(filtered[0].Subject().RawValue(), "https://example.com/resource1")
}

func TestExcludePrefixesSerialization(t *testing.T) {
	is := is.New(t)
	
	// Load existing test data which already has proper namespaces
	g := rdf.NewGraph()
	g.UseIndex = true
	g.UseResource = true
	
	f, err := os.Open("testdata/rdf.nt")
	is.NoErr(err)
	defer f.Close()
	_, err = ntriples.Parse(f, g)
	is.NoErr(err)
	
	// Test serialization with subject filtering (filter out klek.si resources)
	cfg := &FilterConfig{ExcludePrefixes: []string{"https://klek.si/"}}
	
	var buf bytes.Buffer
	err = Serialize(g, &buf, cfg)
	is.NoErr(err)
	
	xmlOutput := buf.String()
	
	// Should NOT contain the excluded subject
	is.True(!strings.Contains(xmlOutput, "https://klek.si/"))
	// Should contain the other resource
	is.True(strings.Contains(xmlOutput, "https://eu.cdn.kleksi.com/"))
}

func TestExcludeTypePrefixes(t *testing.T) {
	is := is.New(t)
	
	// Create a simple graph for testing
	g := rdf.NewGraph()
	g.UseIndex = true
	g.UseResource = true
	
	// Add some test triples with different types
	subj1, _ := rdf.NewIRI("https://example.com/resource1")
	subj2, _ := rdf.NewIRI("https://example.com/resource2")
	type1, _ := rdf.NewIRI("https://schema.org/CreativeWork")
	type2, _ := rdf.NewIRI("http://schemas.delving.eu/nave/terms/SomeType")
	pred1, _ := rdf.NewIRI("https://schema.org/name")
	obj, _ := rdf.NewLiteral("test value")
	
	// Add rdf:type triples
	g.Add(rdf.NewTriple(subj1, rdf.IsA, type1))
	g.Add(rdf.NewTriple(subj2, rdf.IsA, type2))
	// Add other predicates
	g.Add(rdf.NewTriple(subj1, pred1, obj))
	g.Add(rdf.NewTriple(subj2, pred1, obj))
	
	// Test filtering resources with excluded type prefixes
	cfg := &FilterConfig{ExcludeTypePrefixes: []string{"http://schemas.delving.eu/nave/terms/"}}
	filtered := filterResources(g.Resources(), cfg)
	
	// Should only have the resource with type1 (not the one with delving type)
	is.Equal(len(filtered), 1)
	is.Equal(filtered[0].Subject().RawValue(), "https://example.com/resource1")
	
	// Verify the remaining resource has the correct type
	types := filtered[0].Types()
	is.Equal(len(types), 1)
	is.Equal(types[0].RawValue(), "https://schema.org/CreativeWork")
}

func TestCombinedPrefixFilters(t *testing.T) {
	is := is.New(t)
	
	// Create a graph with resources that should be filtered by different criteria
	g := rdf.NewGraph()
	g.UseIndex = true
	g.UseResource = true
	
	// Resource 1: valid subject, valid type, valid predicate
	subj1, _ := rdf.NewIRI("https://example.com/resource1")
	type1, _ := rdf.NewIRI("https://schema.org/CreativeWork")
	pred1, _ := rdf.NewIRI("https://schema.org/name")
	
	// Resource 2: has excluded subject prefix
	subj2, _ := rdf.NewIRI("http://schemas.delving.eu/nave/terms/resource2")
	type2, _ := rdf.NewIRI("https://schema.org/CreativeWork")
	
	// Resource 3: has excluded type prefix
	subj3, _ := rdf.NewIRI("https://example.com/resource3")
	type3, _ := rdf.NewIRI("http://schemas.delving.eu/nave/terms/SomeType")
	
	obj, _ := rdf.NewLiteral("test value")
	
	// Add triples
	g.Add(rdf.NewTriple(subj1, rdf.IsA, type1))
	g.Add(rdf.NewTriple(subj1, pred1, obj))
	
	g.Add(rdf.NewTriple(subj2, rdf.IsA, type2))
	g.Add(rdf.NewTriple(subj2, pred1, obj))
	
	g.Add(rdf.NewTriple(subj3, rdf.IsA, type3))
	g.Add(rdf.NewTriple(subj3, pred1, obj))
	
	// Test filtering with both ExcludePrefixes and ExcludeTypePrefixes
	cfg := &FilterConfig{
		ExcludePrefixes:     []string{"http://schemas.delving.eu/nave/terms/"},
		ExcludeTypePrefixes: []string{"http://schemas.delving.eu/nave/terms/"},
	}
	filtered := filterResources(g.Resources(), cfg)
	
	// Should only have resource1 (resource2 excluded by subject, resource3 excluded by type)
	is.Equal(len(filtered), 1)
	is.Equal(filtered[0].Subject().RawValue(), "https://example.com/resource1")
}