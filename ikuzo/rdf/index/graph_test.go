package index

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/delving/hub3/ikuzo/rdf/formats/ntriples"
	"github.com/google/go-cmp/cmp"
	"github.com/matryer/is"
	"github.com/tidwall/sjson"
)

func TestGraph_Marshal(t *testing.T) {
	header := Header{
		OrgID:    "org1",
		Spec:     "spec1",
		HubID:    "hub1",
		EntryURI: "urn:123",
	}

	graph, err := NewGraph(header)
	if err != nil {
		t.Fatalf("failed to create new graph: %v", err)
	}

	resource1 := &Resource{
		ID:    "resource1",
		order: 1,
		Entries: []*Entry{
			{
				SearchLabel: "label1",
			},
		},
	}

	resource2 := &Resource{
		ID:    "resource2",
		order: 2,
		Entries: []*Entry{
			{
				SearchLabel: "label2",
			},
		},
	}

	resource3 := &Resource{
		ID:    "resource3",
		order: 3,
		Entries: []*Entry{
			{
				SearchLabel: "label3",
			},
		},
	}

	graph.Resources = append(graph.Resources, resource1, resource2, resource3)

	got, err := graph.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal graph: %v", err)
	}

	want, err := json.Marshal(graph)
	if err != nil {
		t.Fatalf("failed to marshal graph using json.Marshal: %v", err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Marshal() mismatch (-want +got):\n%s", diff)
	}
}

func getIndexGraph() (string, error) {
	f, err := os.Open("../formats/ntriples/testdata/rdf.nt")
	if err != nil {
		return "", err
	}
	defer f.Close()

	g1, err := ntriples.Parse(f, nil)
	if err != nil {
		return "", err
	}

	header := Header{
		OrgID:    "org1",
		Spec:     "spec1",
		HubID:    "hub1",
		EntryURI: "urn:123",
	}

	graph, err := NewGraph(header)
	if err != nil {
		return "", err
	}

	if err := graph.AddGraph(g1); err != nil {
		return "", err
	}

	b, err := graph.Marshal()
	if err != nil {
		return "", err
	}

	b, err = sjson.DeleteBytes(b, "meta.modified")
	if err != nil {
		return "", err
	}

	return formatJSON(b)
}

func formatJSON(data []byte) (string, error) {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, data, "", "    "); err != nil {
		return "", err
	}
	return prettyJSON.String(), nil
}

func TestGraph_MarshalFromGraph(t *testing.T) {
	is := is.New(t)

	want, err := getIndexGraph()
	is.NoErr(err)

	got, err := getIndexGraph()
	is.NoErr(err)

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Marshal() mismatch (-want +got):\n%s", diff)
	}
}

func TestGraphInline(t *testing.T) {
	is := is.New(t)

	_, graph, err := getGraphByFile("./testdata/rdf_brocade.rdf.xml", "rdfxml")
	is.NoErr(err)

	is.Equal(len(graph.roots), 0)

	err = graph.Inline()
	is.NoErr(err)
	is.Equal(len(graph.roots), 1)
	is.Equal(graph.roots[0].Types, []string{"https://www.iflastandards.info/ns/lrm/lrmoo#F3_Manifestation"})
}
