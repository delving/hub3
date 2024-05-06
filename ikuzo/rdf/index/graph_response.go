package index

// GraphResponse wraps the Graph with various custom output formats.
type GraphResponse struct {
	// embedded Graph
	Graph
	// Tree       *Tree                     `json:"tree,omitempty"`
	Summary *ResultSummary           `json:"summary,omitempty"`
	JSONLD  []map[string]interface{} `json:"jsonld,omitempty"`
	// Fields     map[string][]string       `json:"fields,omitempty"`
	// Highlights []*ResourceEntryHighlight `json:"highlights,omitempty"`
	// ProtoBuf   *ProtoBuf                 `json:"protobuf,omitempty"`
}

// Fields is a flat representation of triples where all resources are
// merged by their Predicate SearchLabel.
type Fields map[string][]string

// NewJSONLD creates a JSON-LD version of the FragmentGraph
func (gr *GraphResponse) NewJSONLD() []map[string]interface{} {
	gr.JSONLD = []map[string]interface{}{}
	ids := map[string]bool{}
	for _, rsc := range gr.Resources {
		if _, ok := ids[rsc.ID]; ok {
			continue
		}
		gr.JSONLD = append(gr.JSONLD, generateJSONLD(rsc))
		ids[rsc.ID] = true
	}
	return gr.JSONLD
}

// NewResultSummary creates a Summary from the FragmentGraph based on the
// RDFTag configuration.
func (gr *GraphResponse) NewResultSummary() *ResultSummary {
	gr.Summary = &ResultSummary{}
	for _, rsc := range gr.Resources {
		for _, entry := range rsc.Entries {
			gr.Summary.AddEntry(entry)
		}
	}
	return gr.Summary
}

type ldObject struct {
	ID       string `json:"@id,omitempty"`
	Value    string `json:"@value,omitempty"`
	Language string `json:"@language,omitempty"`
	Datatype string `json:"@type,omitempty"`
}

// asLdObject generates an rdf2go.LdObject for JSON-LD generation
func asLdObject(e *Entry) *ldObject {
	o := &ldObject{
		ID:       e.ID,
		Language: e.Language,
		Datatype: e.DataType,
	}
	if e.ID == "" {
		o.Value = e.Value
	}
	return o
}

// generateJSONLD converts a FragmenResource into a JSON-LD entry
func generateJSONLD(rsc *Resource) map[string]interface{} {
	m := map[string]interface{}{}
	m["@id"] = rsc.ID
	if len(rsc.Types) > 0 {
		m["@type"] = rsc.Types
	}
	grouped := map[string][]*Entry{}
	for _, e := range rsc.Entries {
		grouped[e.Predicate] = append(grouped[e.Predicate], e)
	}
	for predicate, entries := range grouped {
		objects := []*ldObject{}
		for _, e := range entries {
			objects = append(objects, asLdObject(e))
		}
		if len(objects) != 0 {
			m[predicate] = objects
		}
	}
	return m
}

// ResultSummary is a preview of an RDF Graph that connected based on predicate tags
//
// The goal of this summary is to provide an uniform way for search results to be presented
// via the API without the client needing to understand each individual rdf.Class that is indexed.
type ResultSummary struct {
	Title         []string `json:"title,omitempty"`
	Owner         string   `json:"owner,omitempty"`
	DatasetTitle  []string `json:"datasetTitle,omitempty"`
	Thumbnail     string   `json:"thumbnail,omitempty"`
	LandingPage   string   `json:"landingPage,omitempty"`
	LatLong       []string `json:"latLong,omitempty"`
	Date          []string `json:"date,omitempty"`
	Description   []string `json:"description,omitempty"`
	Subject       []string `json:"subject,omitempty"`
	Collection    []string `json:"collection,omitempty"`
	SubCollection []string `json:"subCollection,omitempty"`
	ObjectID      string   `json:"objectID,omitempty"`
	ObjectType    []string `json:"objectType,omitempty"`
	Creator       []string `json:"creator,omitempty"`
}

// AddEntry adds Summary fields based on the ResourceEntry tags
func (sum *ResultSummary) AddEntry(entry *Entry) {
	for _, tag := range entry.Tags {
		switch tag {
		case "title":
			sum.Title = append(sum.Title, entry.Value)
		case "thumbnail":
			// Always prefer edm:object for the thumbnail.
			// This also ensures that first webresource is used
			if entry.SearchLabel == "edm_object" {
				sum.Thumbnail = entry.Value
			}

			if len(sum.Thumbnail) == 0 {
				sum.Thumbnail = entry.Value
			}
		case "subject":
			sum.Subject = append(sum.Subject, entry.Value)
		case "creator":
			sum.Creator = append(sum.Creator, entry.Value)
		case "description":
			sum.Description = append(sum.Description, entry.Value)
		case "landingPage":
			if len(sum.LandingPage) == 0 {
				sum.LandingPage = entry.Value
			}
		case "collection":
			sum.Collection = append(sum.Collection, entry.Value)
		case "subCollection":
			sum.SubCollection = append(sum.SubCollection, entry.Value)
		case "objectType":
			sum.ObjectType = append(sum.ObjectType, entry.Value)
		case "objectID":
			if sum.ObjectID == "" {
				sum.ObjectID = entry.Value
			}
		case "owner":
			if sum.Owner == "" {
				sum.Owner = entry.Value
			}
		case "date":
			sum.Date = append(sum.Date, entry.Value)
		}
	}
}
