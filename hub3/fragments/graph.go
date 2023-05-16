package fragments

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/delving/hub3/ikuzo/domain/domainpb"
	"github.com/delving/hub3/ikuzo/rdf"
)

// FragmentGraph is a container for all entries of an RDF Named Graph
type FragmentGraph struct {
	Meta       *Header                   `json:"meta,omitempty"`
	Tree       *Tree                     `json:"tree,omitempty"`
	Resources  []*FragmentResource       `json:"resources,omitempty"`
	Summary    *ResultSummary            `json:"summary,omitempty"`
	JSONLD     []map[string]interface{}  `json:"jsonld,omitempty"`
	Fields     map[string][]string       `json:"fields,omitempty"`
	Highlights []*ResourceEntryHighlight `json:"highlights,omitempty"`
	ProtoBuf   *ProtoBuf                 `json:"protobuf,omitempty"`
}

func (fg *FragmentGraph) Graph() (*rdf.Graph, error) {
	g := rdf.NewGraph()
	for _, rsc := range fg.Resources {
		if err := rsc.AddTo(g); err != nil {
			return nil, err
		}
	}

	return g, nil
}

func (fg *FragmentGraph) Marshal() ([]byte, error) {
	return json.Marshal(fg)
}

func (fg *FragmentGraph) Reader() (int, io.Reader, error) {
	// TODO: idempotency is an issue
	fg.Meta.Modified = 0
	fg.Meta.Revision = 0
	b, err := json.MarshalIndent(fg, "", "    ")
	if err != nil {
		return 0, nil, err
	}

	return len(b), bytes.NewReader(b), nil
}

func (fg *FragmentGraph) IndexMessage() (*domainpb.IndexMessage, error) {
	b, err := fg.Marshal()
	if err != nil {
		return nil, err
	}

	return &domainpb.IndexMessage{
		OrganisationID: fg.Meta.OrgID,
		DatasetID:      fg.Meta.Spec,
		RecordID:       fg.Meta.HubID,
		IndexType:      domainpb.IndexType_V2,
		Source:         b,
	}, nil
}
