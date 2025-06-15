package fragments

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/delving/hub3/ikuzo/domain/domainpb"
	"github.com/delving/hub3/ikuzo/rdf"
)

// FragmentGraph is a container for all entries of an RDF Named Graph
type FragmentGraph struct {
	Meta         *Header                   `json:"meta,omitempty"`
	Tree         *Tree                     `json:"tree,omitempty"`
	Resources    []*FragmentResource       `json:"resources,omitempty"`
	Summary      *ResultSummary            `json:"summary,omitempty"`
	JSONLD       []map[string]any          `json:"jsonld,omitempty"`
	Fields       map[string][]string       `json:"fields,omitempty"`
	Highlights   []*ResourceEntryHighlight `json:"highlights,omitempty"`
	ProtoBuf     *ProtoBuf                 `json:"protoBuf,omitempty"`
	Item         *ItemV1                   `json:"item,omitempty"`
	MoreLikeThis *RelatedItems             `json:"relatedItems,omitempty"`
}

type RelatedItems struct {
	Items []*ItemV1 `json:"item"`
}

// ItemV1 represents a single search result item
type ItemV1 struct {
	DocID   string              `json:"doc_id"`
	DocType string              `json:"doc_type"`
	Fields  map[string][]string `json:"fields"`
}

func (fg *FragmentGraph) AddMoreLikeThis(item *ItemV1) {
	if fg.MoreLikeThis == nil {
		fg.MoreLikeThis = &RelatedItems{}
	}
	fg.MoreLikeThis.Items = append(fg.MoreLikeThis.Items, item)
}

func (fg *FragmentGraph) Graph() (*rdf.Graph, error) {
	if len(fg.Resources) == 0 {
		return nil, fmt.Errorf("unable to create *rdf.Graph because resources is empty")
	}
	g := rdf.NewGraph()
	for _, rsc := range fg.Resources {
		if err := rsc.AddTo(g); err != nil {
			return nil, err
		}
	}

	subj, err := rdf.NewIRI(fg.Meta.EntryURI)
	if err != nil {
		return nil, err
	}

	g.GraphName = fg.Meta.GetNamedGraphURI()

	g.Subject = rdf.Subject(subj)
	return g, nil
}

func (fg *FragmentGraph) Marshal() ([]byte, error) {
	return json.Marshal(fg)
}

func (fg *FragmentGraph) Reader() (io.Reader, error) {
	b, err := json.MarshalIndent(fg, "", "    ")
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(b), nil
}

func (fg *FragmentGraph) IndexMessage() (*domainpb.IndexMessage, error) {
	fg.GenerateFields()
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

// GenerateFields creates a map of searchLabel to object values
// for optimized indexing. This consolidates values from all resources
// in the graph into a flattened structure for easier search and aggregation.
// Values are deduplicated to ensure each unique value appears only once per searchLabel.
func (fg *FragmentGraph) GenerateFields() {
	if len(fg.Fields) > 0 {
		// Fields already generated
		return
	}

	// Initialize the fields map
	fg.Fields = make(map[string][]string)

	// Track values we've already seen for each searchLabel to avoid duplicates
	valuesSeen := make(map[string]map[string]struct{})

	// Process all resources
	for _, rsc := range fg.Resources {
		for _, entry := range rsc.Entries {
			// Skip non-literal entries
			if entry.EntryType != literal {
				continue
			}

			// Skip empty values
			if entry.Value == "" {
				continue
			}

			// Skip entries without a searchLabel
			if entry.SearchLabel == "" {
				continue
			}

			// Use searchLabel as the key
			if _, ok := fg.Fields[entry.SearchLabel]; !ok {
				fg.Fields[entry.SearchLabel] = []string{}
				valuesSeen[entry.SearchLabel] = make(map[string]struct{})
			}

			// Check if we've already seen this value for this searchLabel
			if _, seen := valuesSeen[entry.SearchLabel][entry.Value]; !seen {
				// Add the value to the fields map
				fg.Fields[entry.SearchLabel] = append(fg.Fields[entry.SearchLabel], entry.Value)
				// Mark this value as seen
				valuesSeen[entry.SearchLabel][entry.Value] = struct{}{}
			}
		}
	}
}
