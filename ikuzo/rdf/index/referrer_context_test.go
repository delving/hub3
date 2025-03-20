package index

import (
	"os"

	"github.com/delving/hub3/ikuzo/rdf"
	"github.com/delving/hub3/ikuzo/rdf/formats/ntriples"
	"github.com/delving/hub3/ikuzo/rdf/formats/rdfxml"
	"github.com/tidwall/sjson"
)

func getGraphByFile(path, format string) (string, *Graph, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", nil, err
	}
	defer f.Close()

	var g1 *rdf.Graph

	switch format {
	case "ntriples":
		g1, err = ntriples.Parse(f, nil)
		if err != nil {
			return "", nil, err
		}
	default:
		g1, err = rdfxml.Parse(f, nil, "")
		if err != nil {
			return "", nil, err
		}
	}

	header := Header{
		OrgID:    "org1",
		Spec:     "spec1",
		HubID:    "hub1",
		EntryURI: "urn:123",
	}

	graph, err := NewGraph(header)
	if err != nil {
		return "", nil, err
	}

	if err := graph.AddGraph(g1); err != nil {
		return "", nil, err
	}

	b, err := graph.Marshal()
	if err != nil {
		return "", nil, err
	}

	b, err = sjson.DeleteBytes(b, "meta.modified")
	if err != nil {
		return "", nil, err
	}

	data, err := formatJSON(b)
	if err != nil {
		return "", nil, err
	}

	return data, graph, nil
}
