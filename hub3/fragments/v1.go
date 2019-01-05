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
	"bytes"
	"encoding/json"
	fmt "fmt"
	"io"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	c "github.com/delving/rapid-saas/config"
	r "github.com/kiivihal/rdf2go"

	"github.com/olivere/elastic"
	// TODO replace with dep injection later
	//elastic "gopkg.in/olivere/elastic.v5"

	"github.com/parnurzeal/gorequest"
)

var request *gorequest.SuperAgent

func init() {
	request = gorequest.New()
}

type SortedGraph struct {
	triples []*r.Triple
	lock    sync.Mutex
}

// AddTriple add triple to the list of triples in the sortedGraph.
// Note: there is not deduplication
func (sg *SortedGraph) Add(t *r.Triple) {
	sg.triples = append(sg.triples, t)
}

func (sg *SortedGraph) Triples() []*r.Triple {
	return sg.triples
}

// AddTriple is used to add a triple made of individual S, P, O objects
func (sg *SortedGraph) AddTriple(s r.Term, p r.Term, o r.Term) {
	sg.triples = append(sg.triples, r.NewTriple(s, p, o))
}

// Remove removes a triples from the SortedGraph
func (sg *SortedGraph) Remove(t *r.Triple) {
	triples := []*r.Triple{}
	for _, tt := range sg.triples {
		if t != tt {
			triples = append(triples, tt)
		}
	}
	sg.triples = triples
}

func containsString(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// GenerateJSONLD creates a interfaggce based model of the RDF Graph.
// This can be used to create various JSON-LD output formats, e.g.
// expand, flatten, compacted, etc.
func (sg *SortedGraph) GenerateJSONLD() ([]map[string]interface{}, error) {
	m := map[string]*r.LdEntry{}
	entries := []map[string]interface{}{}
	orderedSubjects := []string{}

	for _, t := range sg.triples {
		s := t.GetSubjectID()
		if !containsString(orderedSubjects, s) {
			orderedSubjects = append(orderedSubjects, s)
		}
		err := r.AppendTriple(m, t)
		if err != nil {
			return entries, err
		}
	}

	//log.Printf("subjects: %#v", orderedSubjects)
	// this most be sorted
	for _, v := range orderedSubjects {
		//log.Printf("v range: %s", v)
		ldEntry, ok := m[v]
		if ok {
			//log.Printf("ldentry: %#v", ldEntry.AsEntry())
			entries = append(entries, ldEntry.AsEntry())
		}
	}

	//log.Printf("graph: \n%#v", entries)

	return entries, nil
}

func (g *SortedGraph) SerializeFlatJSONLD(w io.Writer) error {
	entries, err := g.GenerateJSONLD()
	if err != nil {
		return err
	}
	bytes, err := json.Marshal(entries)
	if err != nil {
		return err
	}
	fmt.Fprint(w, string(bytes))
	return nil
}

func (sg *SortedGraph) GetRDF() ([]byte, error) {
	var b bytes.Buffer
	err := sg.SerializeFlatJSONLD(&b)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
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
		HubID:      fb.fg.Meta.GetHubID(),
		RecordType: "mdr",
		Spec:       fb.fg.Meta.GetSpec(),
		OrgID:      fb.fg.Meta.GetOrgID(),
		Collection: fb.fg.Meta.GetSpec(),
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
	s.Slug = fb.fg.Meta.GetHubID()
	s.Spec = fb.fg.Meta.GetSpec()
	s.Preview = fmt.Sprintf("detail/foldout/void_edmrecord/%s", fb.fg.Meta.GetHubID())
	//s.Caption = ""
	s.AboutURI = fb.fg.GetAboutURI()
	s.SourceURI = fb.fg.GetAboutURI()
	s.GraphName = fb.fg.Meta.NamedGraphURI
	now := time.Now()
	nowString := fmt.Sprintf(now.Format(time.RFC3339))
	s.CreatedAt = nowString
	s.ModifiedAt = nowString
	rdf, err := fb.SortedGraph.GetRDF()
	if err == nil {
		s.SourceGraph = string(rdf)
	} else {
		log.Printf("Unable to add RDF for %s\n", fb.fg.Meta.GetHubID())
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

// GetNSField get as namespace field. It is a utility function
func GetNSField(nsKey, label string) string {
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

// GetEDMField returns a rdf2go.Resource for a field
func GetEDMField(s string) r.Term {
	return r.NewResource(GetNSField("edm", s))
}

// GetNaveField returns a rdf2go.Resource for a field
func GetNaveField(s string) r.Term {
	return r.NewResource(GetNSField("nave", s))
}

// GetUrns returs a list of WebResource urns
func (fb *FragmentBuilder) GetUrns() []string {
	var urns []string
	wrs := fb.Graph.All(nil, nil, GetEDMField("WebResource"))
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
	// replace with err group
	errChan := make(chan error)

	urns := fb.GetUrns()
	for _, urn := range urns {
		go fb.GetRemoteWebResource(urn, "", errChan)
	}
	totalUrns := len(urns)
	for i := 0; i < totalUrns; i++ {
		select {
		case err := <-errChan:
			if err != nil {
				log.Printf("Error resolving webresources for: %v: %#v\n", urns, err)
				return err
			}
		}
	}
	return nil
}

type WebTriples struct {
	triples map[string][]*r.Triple
}

func NewWebTriples() *WebTriples {
	return &WebTriples{triples: make(map[string][]*r.Triple)}
}

func (wt *WebTriples) Append(s string, t *r.Triple) {
	wto, ok := wt.triples[s]
	if !ok {
		wt.triples[s] = []*r.Triple{t}
		return
	}
	wt.triples[s] = append(wto, t)
	return
}

// CleanWebResourceGraph remove mapped webresources when urns are used for WebResource Subjects
func (fb *FragmentBuilder) CleanWebResourceGraph(hasUrns bool) (*SortedGraph, map[string]ResourceSortOrder, []r.Term, *WebTriples) {
	resources := make(map[string]ResourceSortOrder)
	webTriples := NewWebTriples()
	aggregates := []r.Term{}

	cleanGraph := &SortedGraph{}

	seen := 0
	for triple := range fb.Graph.IterTriples() {
		seen++
		s := triple.Subject.String()
		p := triple.Predicate.String()
		subjectIsBNode := false

		switch triple.Subject.(type) {
		case *r.BlankNode:
			subjectIsBNode = true
		}
		if hasUrns && strings.HasSuffix(s, "__>") {
			continue
		}
		switch p {
		case GetNaveField("resourceSortOrder").String():
			rawInt := triple.Object.(*r.Literal).RawValue()
			i, err := strconv.Atoi(rawInt)
			order := 1000
			if err == nil {
				order = i
			}
			resources[s] = ResourceSortOrder{
				Key:   s,
				Value: order,
			}
			continue
		case r.NewResource("http://www.w3.org/1999/02/22-rdf-syntax-ns#type").String():
			o := triple.Object.String()
			if o == GetEDMField("WebResource").String() {
				if !strings.HasSuffix(s, "__>") {
					webTriples.Append(s, triple)
					_, ok := resources[s]
					if !ok {
						resources[s] = ResourceSortOrder{
							Key:   s,
							Value: 1000,
						}
					}
				}
			} else if strings.HasPrefix(o, "<http://schemas.delving.eu/nave/terms/") {
				aggregates = append(aggregates, triple.Subject)
				cleanGraph.Add(triple)
			} else {
				cleanGraph.Add(triple)
			}
		case GetEDMField("hasView").String():
			continue
		case GetEDMField("isShownBy").String(), GetEDMField("object").String():
			continue
		case GetNaveField("thumbnail").String(), GetNaveField("smallThumbnail").String(), GetNaveField("largeThumbnail").String():
			if hasUrns && subjectIsBNode {
				continue
			}
			webTriples.Append(s, triple)
		case GetNaveField("thumbSmall").String(), GetNaveField("thumbnail").String(), GetNaveField("thumbLarge").String():
			if hasUrns && subjectIsBNode {
				continue
			}
			webTriples.Append(s, triple)
		case GetNaveField("deepZoomUrl").String():
			if hasUrns && subjectIsBNode {
				continue
			}
			webTriples.Append(s, triple)
		default:
			cleanGraph.Add(triple)
		}

	}

	return cleanGraph, resources, aggregates, webTriples
}

// GetSortedWebResources returns a list of subjects sorted by nave:resourceSortOrder.
// WebResources without a sortKey will appended in order they are found to the end of the list.
func (fb *FragmentBuilder) GetSortedWebResources() []ResourceSortOrder {
	hasUrns := len(fb.GetUrns()) > 0

	subj := r.NewResource(fb.fg.Meta.GetEntryURI())

	// get remote webresources
	if c.Config.WebResource.ResolveRemoteWebResources {
		err := fb.ResolveWebResources()
		if err != nil {
			log.Printf("err: %#v", err)
			//return err
		}
	}

	cleanGraph, resources, aggregates, webTriples := fb.CleanWebResourceGraph(hasUrns)

	var ss []ResourceSortOrder
	for _, v := range resources {
		ss = append(ss, v)
	}

	keySort := ss
	sort.Slice(keySort, func(i, j int) bool {
		return ss[i].Key < ss[j].Key
	})
	lexSort := make(map[string]int)
	for i, key := range keySort {
		lexSort[key.Key] = i + 1
	}

	for i, s := range ss {
		if s.Value == 1000 {
			ss[i].Value = lexSort[s.Key]
		}

	}
	// sort by Value
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value < ss[j].Value
	})

	for _, s := range ss {
		if len(subj.String()) > 0 {
			if s.Value == 1 && hasUrns {
				fb.AddDefaults(r.NewResource(s.CleanKey()), subj, cleanGraph)
			}
			hasView := r.NewTriple(
				subj,
				GetEDMField("hasView"),
				r.NewResource(s.CleanKey()),
			)
			cleanGraph.Add(hasView)

			sortOrder := r.NewTriple(
				r.NewResource(s.CleanKey()),
				GetNaveField("resourceSortOrder"),
				r.NewLiteral(fmt.Sprintf("%d", s.Value)),
			)
			wt, ok := webTriples.triples[s.Key]
			if ok {
				for _, t := range wt {
					cleanGraph.Add(t)
				}
			}

			//log.Printf("sortOrder: %#v", s)
			//log.Printf("sort triple: %#v", sortOrder.String())
			// add resourceSortOrder back
			cleanGraph.Add(sortOrder)
		} else {
			log.Printf("subject not found: %s", subj)
		}
	}
	// add ore:aggregates
	for _, t := range aggregates {
		cleanGraph.AddTriple(
			subj,
			r.NewResource("http://www.openarchives.org/ore/terms/aggregates"),
			t,
		)
	}

	fb.SortedGraph = cleanGraph

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

// GetObject returns a single object from the rdf2go.Graph
func (fb *FragmentBuilder) GetObject(s r.Term, p r.Term) r.Term {
	t := fb.Graph.One(s, p, nil)
	if t != nil {
		return t.Object
	}
	return nil
}

// AddDefaults add default thumbnail fields to a edm:WebResource
func (fb *FragmentBuilder) AddDefaults(wr r.Term, s r.Term, g *SortedGraph) {
	isShownBy := fb.GetObject(wr, GetNaveField("thumbLarge"))
	if isShownBy == nil {
		//ld, _ := g.GenerateJSONLD()
		log.Printf("should find thumbLarge: %s, %s \n %s", wr.String(), s.String(), "")
		time.Sleep(100 * time.Second)
	}
	if isShownBy != nil {
		g.AddTriple(s, GetEDMField("isShownBy"), isShownBy)
	}
	object := fb.GetObject(wr, GetNaveField("thumbSmall"))
	if object != nil {
		g.AddTriple(s, GetEDMField("object"), isShownBy)
	}
}

// GetRemoteWebResource retrieves a remote Graph from the MediaManare and
// inserts it into the Graph
func (fb *FragmentBuilder) GetRemoteWebResource(urn string, orgID string, errChan chan error) {
	if strings.HasPrefix(urn, "urn:") {
		url := fb.MediaManagerURL(urn, orgID)
		request := gorequest.New()
		//log.Printf("webresource url: %s", url)
		resp, _, errs := request.Get(url).End()
		if errs != nil {
			for err := range errs {
				log.Printf("err: %#v", err)
			}
			errChan <- errs[0]
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			log.Printf("urn not found: %s?format=plain", url)
			errChan <- nil
			return
		}
		defer resp.Body.Close()
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
		strings.Replace(urn, "%", "%%", -1),
	)
}

// GetResourceLabel returns the label for a resource
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
	// TODO add support for additionalLabels from the configuration
	labels := []r.Term{
		r.NewResource("http://www.w3.org/2004/02/skos/core#prefLabel"),
		r.NewResource("http://xmlns.com/foaf/0.1/name"),
	}
	for _, label := range labels {
		for _, t := range fb.Graph.All(nil, label, nil) {
			subjectID := t.GetSubjectID()
			_, ok := fb.ResourceLabels[subjectID]
			if ok {
				break
			}
			fb.ResourceLabels[subjectID] = t.Object.(*r.Literal).RawValue()
		}
	}
	return nil
}

func fieldsContains(s []*IndexEntry, e *IndexEntry) bool {
	for _, a := range s {
		if a.Raw == e.Raw {
			return true
		}
	}
	return false
}

// CreateV1IndexDoc creates a map that can me marshaled to json
func CreateV1IndexDoc(fb *FragmentBuilder) (map[string]interface{}, error) {
	indexDoc := make(map[string]interface{})

	// set the resourceLabels
	err := fb.SetResourceLabels()
	if err != nil {
		return indexDoc, err
	}

	var triples []*r.Triple
	for _, t := range fb.SortedGraph.Triples() {
		triples = append(triples, t)
	}

	for _, t := range triples {
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
		if fieldsContains(fields, entry) {
			fb.SortedGraph.Remove(t)
			continue
		}
		indexDoc[searchLabel] = append(fields, entry)
	}
	indexDoc["delving_spec"] = IndexEntry{
		Type:  "Literal",
		Value: fb.fg.Meta.GetSpec(),
		Raw:   fb.fg.Meta.GetSpec(),
	}
	indexDoc["spec"] = fb.fg.Meta.GetSpec()
	indexDoc["orgID"] = fb.fg.Meta.GetOrgID()
	indexDoc["entryURI"] = fb.fg.GetAboutURI()
	indexDoc["revision"] = fb.fg.Meta.GetRevision()
	indexDoc["hubID"] = fb.fg.Meta.GetHubID()
	indexDoc["system"] = NewSystem(indexDoc, fb)
	indexDoc["legacy"] = NewLegacy(indexDoc, fb)

	//_, ok := indexDoc["edm_isShownBy"]
	//if !ok {
	//log.Printf("indexdoc: %#v", indexDoc["system"])
	//time.Sleep(10 * time.Second)
	//}
	return indexDoc, nil
}

// GetFieldKey returns the namespaced version of the Predicate of the Triple
func GetFieldKey(t *r.Triple) (string, error) {
	return c.Config.NameSpaceMap.GetSearchLabel(t.Predicate.RawValue())
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

// V1ESMapping has the legacy mapping for V1 indexes. It should only be used when indexV1 is enabled in the
// configuration.
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
                    "enabled": "true"
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
