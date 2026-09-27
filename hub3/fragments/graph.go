package fragments

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

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
	Semantic     *SemanticView             `json:"semantic,omitempty"`
}

type RelatedItems struct {
	Items         []*ItemV1       `json:"item,omitempty"`
	SemanticItems []*FragmentGraph `json:"items,omitempty"`
}

// ItemV1 represents a single search result item
type ItemV1 struct {
	DocID   string              `json:"doc_id"`
	DocType string              `json:"doc_type"`
	Fields  map[string][]string `json:"fields"`
}

func (fg *FragmentGraph) ItemType(path string) string {
	// only the first one is returned
	for _, rsc := range fg.Resources {
		for _, entry := range rsc.Entries {
			if entry.SearchLabel == path {
				if entry.Value != "" {
					return strings.ToLower(entry.Value)
				}
			}
		}
	}

	// nothing found return default
	return "default"
}

func (fg *FragmentGraph) AddMoreLikeThis(item *ItemV1) {
	if fg.MoreLikeThis == nil {
		fg.MoreLikeThis = &RelatedItems{}
	}
	fg.MoreLikeThis.Items = append(fg.MoreLikeThis.Items, item)
}

func (fg *FragmentGraph) AddSemanticMoreLikeThis(related *FragmentGraph) {
	if fg.MoreLikeThis == nil {
		fg.MoreLikeThis = &RelatedItems{}
	}
	fg.MoreLikeThis.SemanticItems = append(fg.MoreLikeThis.SemanticItems, related)
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

// setDerivedMeta fills the parts of the meta block that are our conclusion
// about a record rather than something the source supplied.
func (fg *FragmentGraph) setDerivedMeta() {
	if fg.Meta == nil {
		return
	}

	// A record has a digital object when it says where that object is shown.
	//
	// Read from the resources, not from fg.Fields. GenerateFields keeps only
	// literal entries, and edm_isShownBy is a URI -- so it is never in that map
	// however many images the record has. Reading it there marked all 56 records
	// of a beeldmateriaal dataset as having no media while both v1 APIs said
	// they had, which is how this was found: on real data after a resend, not by
	// reading the code.
	//
	// The condition itself is the one the v1 legacy block uses for
	// delving_hasDigitalObject (see NewLegacy), deliberately: both are served
	// for as long as the Django v1 has consumers, and a flag that means two
	// different things depending on which API you ask is worse than one that is
	// imperfect in a known way.
	//
	// The known imperfection: edm_object, edm_hasView and the thumbnails on the
	// WebResource are not consulted, so a record exposing its object only
	// through those counts as having none. Broadening that is a decision for
	// both sides at once, not something to slip in here.
	hasObject := fg.hasSearchLabel("edm_isShownBy")
	fg.Meta.HasDigitalObject = &hasObject
}

// hasSearchLabel reports whether any resource carries an entry with this search
// label, whatever its entry type.
//
// The search labels come from the same namespace manager that v1's indexer
// consults through GetFieldKey when it walks the triples, so this asks the same
// question of the same table -- only of the parsed resources rather than by
// walking the graph again.
func (fg *FragmentGraph) hasSearchLabel(label string) bool {
	for _, rsc := range fg.Resources {
		if len(rsc.FilterEntries(label)) > 0 {
			return true
		}
	}

	return false
}

func (fg *FragmentGraph) IndexMessage() (*domainpb.IndexMessage, error) {
	fg.GenerateFields()
	fg.setDerivedMeta()

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
