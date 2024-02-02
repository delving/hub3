package sparql

import (
	"bytes"
	"context"
	fmt "fmt"
	"testing"

	"github.com/matryer/is"

	"github.com/delving/hub3/ikuzo/rdf"
)

func TestHarvestSubjects(t *testing.T) {
	is := is.New(t)

	cfg := HarvestConfig{
		URL: "https://eu.api.kleksi.com/apps/pqx31b/datasets/default/sparql",
		Queries: struct {
			NamespacePrefix        string
			WhereClause            string
			SubjectVar             string
			IncrementalWhereClause string
			GetGraphQuery          string
		}{
			NamespacePrefix: "PREFIX schema: <https://schema.org/> ",
			WhereClause:     "?s schema:identifier ?identifier .",
			SubjectVar:      "identifier",
			IncrementalWhereClause: `
				?s schema:identifier ?identifier ;
				schema:dateModified ?dateModified .
				FILTER(?dateModified > "~~DATE~~"^^xsd:date)
			`,
			GetGraphQuery: "",
		},
		GraphMimeType: "text/turtle",
		MaxSubjects:   750,
		PageSize:      500,
	}

	ids := make(chan string, 1000)

	err := HarvestSubjects(context.Background(), cfg, ids)
	is.NoErr(err)

	t.Logf("number of ids: %d", len(ids))
	is.Equal(len(ids), 750)
}

func TestHarvestGraphs(t *testing.T) {
	is := is.New(t)

	cfg := HarvestConfig{
		URL: "https://eu.api.kleksi.com/apps/pqx31b/datasets/default/sparql",
		Queries: struct {
			NamespacePrefix        string
			WhereClause            string
			SubjectVar             string
			IncrementalWhereClause string
			GetGraphQuery          string
		}{
			NamespacePrefix: "PREFIX schema: <https://schema.org/> ",
			WhereClause:     "?s schema:identifier ?identifier .",
			SubjectVar:      "identifier",
			IncrementalWhereClause: `
				?s schema:identifier ?identifier ;
				schema:dateModified ?dateModified .
				FILTER(?dateModified > "~~DATE~~"^^xsd:date)
			`,
			GetGraphQuery: "",
		},
		GraphMimeType: "text/turtle",
		MaxSubjects:   10,
		PageSize:      500,
	}

	var seen int
	var b bytes.Buffer
	fmt.Fprintln(&b, "<records>")

	cb := func(g *rdf.Graph) error {
		fmt.Fprintf(&b, "<record id=\"%s\">\n", g.Subject.RawValue())

		t.Logf("number of triples: %d", g.Len())

		// filterCfg := &mappingxml.FilterConfig{Subject: g.Subject}
		// _ = filterCfg
		// err := mappingxml.Serialize(g, &b, filterCfg)
		// if err != nil {
		// 	return err
		// }
		// ntriples.Serialize(g, &b)
		seen++
		fmt.Fprintln(&b, "</record>")
		return nil
	}
	fmt.Fprintln(&b, "</records>")

	err := HarvestGraphs(context.Background(), cfg, cb)
	is.NoErr(err)
	t.Logf("mappingxml records: %s", b.String())
	is.Equal(seen, 10)
}

func TestHarvest(t *testing.T) {
	is := is.New(t)

	cfg := RepoConfig{
		Host:      "https://api.kleksi.com",
		QueryPath: "/apps/thorn/datasets/default/sparql",
	}
	repo, err := NewRepo(cfg)
	is.NoErr(err)

	query := `PREFIX%20rdf%3A%20%3Chttp%3A%2F%2Fwww.w3.org%2F1999%2F02%2F22-rdf-syntax-ns%23%3E%0APREFIX%20rdfs%3A%20%3Chttp%3A%2F%2Fwww.w3.org%2F2000%2F01%2Frdf-schema%23%3E%0ASELECT%20*%20WHERE%20%7B%0A%20%20%3Fs%20a%20%20%3Chttps%3A%2F%2Fschema.org%2FCreativeWork%3E.%0A%7D%20LIMIT%2010`
	responses, err := Harvest(context.TODO(), repo, query)
	is.NoErr(err)
	_ = responses
}
