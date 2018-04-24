// Copyright © 2017 Delving B.V. <info@delving.eu>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fragments

import (
	fmt "fmt"
	"io"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	c "github.com/delving/rapid-saas/config"
	r "github.com/kiivihal/rdf2go"
	//"github.com/olivere/elastic"
	elastic "gopkg.in/olivere/elastic.v5"

	"github.com/parnurzeal/gorequest"
)

var request *gorequest.SuperAgent

func init() {
	request = gorequest.New()
}

// IndexEntry holds info for earch triple in the V1 API
type IndexEntry struct {
	ID       string `json:"id,omitempty"`
	Value    string `json:"value,omitempty"`
	Language string `json:"language,omitempty"`
	Type     string `json:"@type,omitempty"`
	Raw      string `json:"raw,omitempty"`
}

// Legacy holds the legacy values
type Legacy struct {
	HubID            string `json:"delving_hubId,omitempty"`
	RecordType       string `json:"delving_recordType,omitempty"`
	Spec             string `json:"delving_spec,omitempty"`
	Owner            string `json:"delving_owner,omitempty"`
	OrgID            string `json:"delving_orgId,omitempty"`
	Collection       string `json:"delving_collection,omitempty"`
	Title            string `json:"delving_title,omitempty"`
	Creator          string `json:"delving_creator,omitempty"`
	Provider         string `json:"delving_provider,omitempty"`
	HasGeoHash       string `json:"delving_hasGeoHash"`
	HasDigitalObject string `json:"delving_hasDigitalObject"`
	HasLandingPage   string `json:"delving_hasLandingPage"`
	HasDeepZoom      string `json:"delving_hasDeepZoom"`
}

// NewLegacy returns a legacy struct with default values
func NewLegacy(indexDoc map[string]interface{}, fb *FragmentBuilder) *Legacy {
	l := &Legacy{
		HubID:      fb.fg.GetHubID(),
		RecordType: "mdr",
		Spec:       fb.fg.GetSpec(),
		OrgID:      fb.fg.GetOrgID(),
		Collection: fb.fg.GetSpec(),
	}
	var ok bool
	_, ok = indexDoc["nave_GeoHash"]
	l.HasGeoHash = strconv.FormatBool(ok)
	_, ok = indexDoc["edm_isShownBy"]
	l.HasDigitalObject = strconv.FormatBool(ok)
	_, ok = indexDoc["nave_DeepZoomUrl"]
	l.HasDeepZoom = strconv.FormatBool(ok)
	_, ok = indexDoc["edm_isShownAt"]
	l.HasLandingPage = strconv.FormatBool(ok)
	return l
}

// System holds system information for each IndexDoc
type System struct {
	Slug               string `json:"slug,omitempty"`
	Spec               string `json:"spec,omitempty"`
	Thumbnail          string `json:"thumbnail,omitempty"`
	Preview            string `json:"preview,omitempty"`
	Caption            string `json:"caption,omitempty"`
	AboutURI           string `json:"about_uri,omitempty"`
	SourceURI          string `json:"source_uri,omitempty"`
	GraphName          string `json:"graph_name,omitempty"`
	CreatedAt          string `json:"created_at,omitempty"`
	ModifiedAt         string `json:"modified_at,omitempty"`
	SourceGraph        string `json:"source_graph,omitempty"`
	ProxyResourceGraph string `json:"proxy_resource_graph,omitempty"`
	WebResourceGraph   string `json:"web_resource_graph,omitempty"`
	ContentHash        string `json:"content_hash,omitempty"`
	HasGeoHash         string `json:"hasGeoHash"`
	HasDigitalObject   string `json:"hasDigitalObject"`
	HasLandingPage     string `json:"hasLandingPage"`
	HasDeepZoom        string `json:"hasDeepZoom"`
}

// NewSystem generates system info for the V1 doc
func NewSystem(indexDoc map[string]interface{}, fb *FragmentBuilder) *System {
	s := &System{}
	s.Slug = fb.fg.GetHubID()
	s.Spec = fb.fg.GetSpec()
	s.Preview = fmt.Sprintf("detail/foldout/void_edmrecord/%s", fb.fg.GetHubID())
	//s.Caption = ""
	s.AboutURI = fb.fg.GetAboutURI()
	s.SourceURI = fb.fg.GetAboutURI()
	s.GraphName = fb.fg.GetNamedGraphURI()
	now := time.Now()
	nowString := fmt.Sprintf(now.Format(time.RFC3339))
	s.CreatedAt = nowString
	s.ModifiedAt = nowString
	rdf, err := fb.GetRDF()
	if err == nil {
		s.SourceGraph = string(rdf)
	} else {
		log.Println("Unable to add RDF for %s", fb.fg.GetHubID())
	}
	// s.ProxyResourceGraph
	// s.WebResourceGraph
	// s.ContentHash
	thumbnails, ok := indexDoc["edm_object"]
	if ok {
		thumbs := thumbnails.([]*IndexEntry)
		if len(thumbs) > 0 {
			s.Thumbnail = thumbs[0].Value
		}
	}
	_, ok = indexDoc["nave_GeoHash"]
	s.HasGeoHash = strconv.FormatBool(ok)
	_, ok = indexDoc["edm_isShownBy"]
	s.HasDigitalObject = strconv.FormatBool(ok)
	_, ok = indexDoc["nave_DeepZoomUrl"]
	s.HasDeepZoom = strconv.FormatBool(ok)
	_, ok = indexDoc["edm_isShownAt"]
	s.HasLandingPage = strconv.FormatBool(ok)
	return s
}

//'slug': self.hub_id,
//'spec': self.get_spec_name(),
//'thumbnail': thumbnail if thumbnail else "",
//'preview': "detail/foldout/{}/{}".format(doc_type, self.hub_id),
//'caption': bindings.get_about_caption if bindings.get_about_caption else "",
//'about_uri': self.source_uri,
//'source_uri': self.source_uri,
//'graph_name': self.named_graph,
//'created_at': datetime.now().isoformat(),
//'modified_at': datetime.now().isoformat(),
//'source_graph': self.rdf_string(),
//'proxy_resource_graph': None,
//'web_resource_graph': None,
//'content_hash': content_hash,
//'hasGeoHash': "true" if bindings.has_geo() else ""false"",
//'hasDigitalObject': "true" if thumbnail else ""false"",
//'hasLandingePage': "true" if 'edm_isShownAt' in index_doc else ""false"",
//'hasDeepZoom': "true" if 'nave_deepZoom' in index_doc else ""false"",

// NewGraphFromTurtle creates a RDF graph from the 'text/turtle' format
func NewGraphFromTurtle(re io.Reader) (*r.Graph, error) {
	g := r.NewGraph("")
	err := g.Parse(re, "text/turtle")
	if err != nil {
		log.Println("Unable to parse the supplied turtle RDF.")
		return g, err
	}
	if g.Len() == 0 {
		//log.Println("No triples were added to the graph")
		return g, fmt.Errorf("no triples were added to the graph")
	}
	return g, nil
}

func getNSField(nsKey, label string) string {
	var nsURI string
	switch nsKey {
	case "edm":
		nsURI = "http://www.europeana.eu/schemas/edm/"
	case "nave":
		nsURI = "http://schemas.delving.eu/nave/terms/"
	case "dcterms":
		nsURI = "http://purl.org/dc/terms/"
	case "rdagr2":
		nsURI = "http://rdvocab.info/ElementsGr2/"
	case "dc":
		nsURI = "http://purl.org/dc/elements/1.1/"
	}
	if nsURI != "" {
		return fmt.Sprintf(
			"%s%s",
			nsURI,
			label,
		)
	}
	return ""
}

func getEDMField(s string) r.Term {
	return r.NewResource(getNSField("edm", s))
}

func getNaveField(s string) r.Term {
	return r.NewResource(getNSField("nave", s))
}

// GetUrns returs a list of WebResource urns
func (fb *FragmentBuilder) GetUrns() []string {
	var urns []string
	wrs := fb.Graph.All(nil, nil, getEDMField("WebResource"))
	for _, t := range wrs {
		s := strings.Trim(t.Subject.String(), "<>")
		if strings.HasPrefix(s, "urn:") {
			urns = append(urns, s)
		}
	}
	return urns
}

// ResolveWebResources retrieves RDF graph from remote MediaManager
// Only RDF Resources that start with 'urn:' are currently supported
func (fb *FragmentBuilder) ResolveWebResources() error {
	errChan := make(chan error)

	urns := fb.GetUrns()
	for _, urn := range urns {
		go fb.GetRemoteWebResource(urn, "", errChan)
	}
	totalUrns := len(urns)
	for i := 0; i < totalUrns; i++ {
		select {
		case err := <-errChan:
			log.Printf("Error resolving webresources for: %v: %v\n", urns, err)
			return err
		}
	}
	return nil
}

// GetSortedWebResources returns a list of subjects sorted by nave:resourceSortOrder.
// WebResources without a sortKey will appended in order they are found to the end of the list.
func (fb *FragmentBuilder) GetSortedWebResources() []ResourceSortOrder {
	resources := make(map[string]int)
	cleanGraph := r.NewGraph("")

	hasUrns := len(fb.GetUrns()) > 0

	graphType := fb.Graph.One(
		nil,
		r.NewResource("http://www.w3.org/1999/02/22-rdf-syntax-ns#type"),
		r.NewResource("http://www.openarchives.org/ore/terms/Aggregation"),
	)
	subj := r.NewResource("")
	if graphType != nil {
		subj = graphType.Subject
	}
	for triple := range fb.Graph.IterTriples() {
		s := triple.Subject.String()
		p := triple.Predicate.String()
		switch p {
		case getNaveField("resourceSortOrder").String():
			rawInt := triple.Object.(*r.Literal).RawValue()
			i, err := strconv.Atoi(rawInt)
			if err != nil {
				resources[s] = 1000
			} else {
				resources[s] = i
			}
			cleanGraph.Add(triple)
		case r.NewResource("http://www.w3.org/1999/02/22-rdf-syntax-ns#type").String():
			if triple.Object.String() == getEDMField("WebResource").String() {
				if !strings.HasSuffix(s, "__>") {
					_, ok := resources[s]
					if !ok {
						resources[s] = 1000
					}
					cleanGraph.Add(triple)
				}
			} else {
				cleanGraph.Add(triple)
			}
		case getEDMField("hasView").String():
			break
		case getEDMField("isShownBy").String(), getEDMField("object").String():
			if hasUrns {
				break
			}
			fallthrough
		//g.Remove(triple)
		//case getNaveField("thumbSmall").String(), getNaveField("thumbnail").String(), getNaveField("thumbLarge").String():
		//fb.Graph.Remove(triple)
		//case getNaveField("deepZoomUrl").String():
		//fb.Graph.Remove(triple)
		default:
			cleanGraph.Add(triple)
		}

	}
	var ss []ResourceSortOrder
	for k, v := range resources {
		ss = append(ss, ResourceSortOrder{k, v})
	}

	// sort by key
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value < ss[j].Value
	})

	// replace 1000 with incremental number
	for i, s := range ss {
		if s.Value == 1000 {
			ss[i].Value = i + 1
		}
		if len(subj.String()) > 0 {
			if i == 0 && hasUrns {
				fb.AddDefaults(r.NewResource(s.CleanKey()), subj, cleanGraph)
			}
			hasView := r.NewTriple(
				subj,
				getEDMField("hasView"),
				r.NewResource(s.CleanKey()),
			)
			cleanGraph.Add(hasView)
		}
	}

	fb.Graph = cleanGraph
	return ss
}

// ResourceSortOrder holds the sort keys
type ResourceSortOrder struct {
	Key   string
	Value int
}

// CleanKey strips leading and trailing "<>" from the key.
func (rso ResourceSortOrder) CleanKey() string {
	return strings.Trim(rso.Key, "<>")
}

func (fb *FragmentBuilder) GetObject(s r.Term, p r.Term) r.Term {
	t := fb.Graph.One(s, p, nil)
	if t != nil {
		return t.Object
	}
	return nil
}

func (fb *FragmentBuilder) AddDefaults(wr r.Term, s r.Term, g *r.Graph) {
	isShownBy := fb.GetObject(wr, getNaveField("thumbLarge"))
	if isShownBy != nil {
		g.AddTriple(s, getEDMField("isShownBy"), isShownBy)
	}
	object := fb.GetObject(wr, getNaveField("thumbSmall"))
	if object != nil {
		g.AddTriple(s, getEDMField("object"), isShownBy)
	}
}

// GetRemoteWebResource retrieves a remote Graph from the MediaManare and
// inserts it into the Graph
func (fb *FragmentBuilder) GetRemoteWebResource(urn string, orgID string, errChan chan error) {
	if strings.HasPrefix(urn, "urn:") {
		url := fb.MediaManagerURL(urn, orgID)
		request := gorequest.New()
		resp, _, errs := request.Get(url).End()
		if errs != nil {
			errChan <- errs[0]
			return
		}
		err := fb.Graph.Parse(resp.Body, "text/turtle")
		errChan <- err
		return
	}
	return
}

// MediaManagerURL returns the URL for the Remote WebResource call.
func (fb *FragmentBuilder) MediaManagerURL(urn string, orgID string) string {
	if orgID == "" {
		orgID = c.Config.OrgID
	}
	return fmt.Sprintf(
		"%s/api/webresource/%s/%s",
		c.Config.WebResource.MediaManagerHost,
		orgID,
		urn,
	)
}

func (fb *FragmentBuilder) GetResourceLabel(t *r.Triple) (string, bool) {
	switch t.Object.(type) {
	case *r.Resource:
		id := r.GetResourceID(t.Object)
		label, ok := fb.ResourceLabels[id]
		return label, ok
	}
	return "", false
}

// SetResourceLabels extracts resource labels from the graph which is used for
// presenting labels for Triple.Object instances that are resources.
func (fb *FragmentBuilder) SetResourceLabels() error {
	prefLabel := r.NewResource("http://www.w3.org/2004/02/skos/core#prefLabel")
	for _, t := range fb.Graph.All(nil, prefLabel, nil) {
		fb.ResourceLabels[t.GetSubjectID()] = t.Object.(*r.Literal).RawValue()
	}
	return nil
}

// CreateV1IndexDoc creates a map that can me marshaled to json
func CreateV1IndexDoc(fb *FragmentBuilder) (map[string]interface{}, error) {
	indexDoc := make(map[string]interface{})

	// set the resourceLabels
	err := fb.SetResourceLabels()
	if err != nil {
		return indexDoc, err
	}

	for t := range fb.Graph.IterTriples() {
		searchLabel, err := GetFieldKey(t)
		if err != nil {
			return indexDoc, err
		}
		var fields []*IndexEntry
		var ok bool
		fields, ok = indexDoc[searchLabel].([]*IndexEntry)
		if !ok {
			fields = []*IndexEntry{}
		}
		entry, err := fb.CreateV1IndexEntry(t)
		if err != nil {
			return indexDoc, err
		}
		indexDoc[searchLabel] = append(fields, entry)
	}
	indexDoc["delving_spec"] = IndexEntry{
		Type:  "Literal",
		Value: fb.fg.GetSpec(),
		Raw:   fb.fg.GetSpec(),
	}
	indexDoc["spec"] = fb.fg.GetSpec()
	indexDoc["entryURI"] = fb.fg.GetAboutURI()
	indexDoc["revision"] = fb.fg.GetRevision()
	indexDoc["hubID"] = fb.fg.GetHubID()
	indexDoc["system"] = NewSystem(indexDoc, fb)
	indexDoc["legacy"] = NewLegacy(indexDoc, fb)
	return indexDoc, nil
}

// GetFieldKey returns the namespaced version of the Predicate of the Triple
func GetFieldKey(t *r.Triple) (string, error) {
	return c.Config.NameSpaceMap.GetSearchLabel(t.Predicate.RawValue())
	return "", nil
}

// CreateV1IndexEntry creates an IndexEntry from a r.Triple
func (fb *FragmentBuilder) CreateV1IndexEntry(t *r.Triple) (*IndexEntry, error) {
	ie := &IndexEntry{}

	switch t.Object.(type) {
	case *r.Resource:
		ie.Type = "URIRef"
		ie.ID = t.Object.RawValue()
		label, ok := fb.GetResourceLabel(t)
		if ok {
			ie.Value = label
			ie.Raw = label
		} else {
			ie.Value = t.Object.RawValue()
			ie.Raw = t.Object.RawValue()
		}
	case *r.Literal:
		ie.Type = "Literal"
		value := t.Object.RawValue()
		if len(value) > 32765 {
			value = value[:32000]
		}
		ie.Value = value
		ie.Raw = value
		if len(ie.Raw) > 256 {
			ie.Raw = value[:256]
		}
		// replace double quotes a single quote
		ie.Raw = strings.Replace(ie.Raw, "\"", "'", -1)
		l := t.Object.(*r.Literal)
		ie.Language = l.Language
	case *r.BlankNode:
		ie.Type = "Bnode"
		ie.ID = t.Object.RawValue()
		ie.Value = t.Object.RawValue()
		ie.Raw = t.Object.RawValue()
	default:

		return ie, fmt.Errorf("unknown object type: %#v", t.Object)
	}
	return ie, nil
}

// CreateESAction creates bulkAPIRequest from map[string]interface{}
func CreateESAction(indexDoc map[string]interface{}, id string) (*elastic.BulkIndexRequest, error) {
	v1Index := fmt.Sprintf("%s", c.Config.ElasticSearch.IndexName)
	r := elastic.NewBulkIndexRequest().
		Index(v1Index).
		Type("void_edmrecord").
		RetryOnConflict(3).
		Id(id).
		Doc(indexDoc)
	return r, nil
}

// V1ESMapping has the legacy mapping for V1 indexes. It should only be used when indexV1 is enabled in the configuration.
var V1ESMapping = `
{
    "settings": {
		"number_of_shards":3,
		"number_of_replicas":2,
        "analysis": {
            "filter": {
                "dutch_stop": {
                    "type":       "stop",
                    "stopwords":  "_dutch_"
                },
                "dutch_stemmer": {
                    "type":       "stemmer",
                    "language":   "dutch"
                },
                "dutch_override": {
                    "type":       "stemmer_override",
                    "rules": [
                        "fiets=>fiets",
                        "bromfiets=>bromfiets",
                        "ei=>eier",
                        "kind=>kinder"
                    ]
                }
            },
            "analyzer": {
                "dutch": {
                    "tokenizer":  "standard",
                    "filter": [
                        "lowercase",
                        "dutch_stop",
                        "dutch_override",
                        "dutch_stemmer"
                    ]
                }
            }
        }
    },
    "mappings": {
        "_default_":
            {
                "_all": {
                    "enabled": "false"
                },
                "date_detection": "false",
                "properties": {
                    "id": {"type": "integer"},
                    "absolute_url": {"type": "keyword"},
                    "point": { "type": "geo_point" },
                    "delving_geohash": { "type": "geo_point" },
                    "delving_geoHash": { "type": "geo_point" },
                    "system": {
                        "properties": {
							"about_uri": {"fields": {"raw": { "type": "keyword"}}, "type": "text"},
							"caption": {"fields": {"raw": { "type": "keyword"}}, "type": "text"},
							"preview": {"fields": {"raw": { "type": "keyword"}}, "type": "text"},
                            "created_at": {"format": "dateOptionalTime", "type": "date"},
							"graph_name": {"fields": {"raw": { "type": "keyword"}}, "type": "text"},
                            "modified_at": {"format": "dateOptionalTime", "type": "date"},
							"slug": {"fields": {"raw": { "type": "keyword"}}, "type": "text"},
                            "geohash": { "type": "geo_point" },
                            "source_graph": { "index": "false", "type": "text", "doc_values": "false" },
							"source_uri": {"fields": {"raw": { "type": "keyword"}}, "type": "text"},
							"spec": {"fields": {"raw": { "type": "keyword"}}, "type": "text"},
							"thumbnail": {"fields": {"raw": { "type": "keyword"}}, "type": "text"}
                    }
                }},
                "dynamic_templates": [
                    {"legacy": { "path_match": "legacy.*",
                        "mapping": { "type": "keyword",
                            "fields": { "raw": { "type": "keyword"}, "value": { "type": "text" } }
                        }
                    }},
                    {"dates": { "match": "*_at", "mapping": { "type": "date" } }},
                    {"rdf": {
                        "path_match": "rdf.*",
                        "mapping": {
                            "type": "text",
                            "fields": {
                                "raw": {
                                    "type": "keyword"
                                },
                                "value": {
                                    "type": "text"
                                }
                            }
                        }
                    }},
                    {"uri": { "match": "id", "mapping": { "type": "keyword" } }},
                    {"point": { "match": "point", "mapping": { "type": "geo_point" }}},
                    {"geo_hash": { "match": "delving_geohash", "mapping": { "type": "geo_point" } }},
                    {"value": { "match": "value", "mapping": { "type": "text" } }},
                    {"raw": {
						"match": "raw",
						"mapping": {"type": "keyword", "ignore_above": 1024}
					}},
                    {"id": { "match": "id", "mapping": { "type": "keyword" } }},
                    {"graphs": { "match": "*_graph", "mapping": { "type": "text", "index": "false" } }},
                    {"inline": { "match": "inline", "mapping": { "type": "object", "include_in_parent": "true" } }},
                    {"strings": {
                        "match_mapping_type": "string",
                        "mapping": {"type": "text", "fields": {"raw": {"type": "keyword", "ignore_above": 1024 }}}
                    }}
                ]
            }
    }}
`
