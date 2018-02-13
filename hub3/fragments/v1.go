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
	"time"

	c "bitbucket.org/delving/rapid/config"
	r "github.com/deiu/rdf2go"
	"github.com/olivere/elastic"
)

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
	HasGeoHash       bool   `json:"delving_hasGeoHash"`
	HasDigitalObject bool   `json:"delving_hasDigitalObject"`
	HasLandingPage   bool   `json:"delving_hasLandingPage"`
	HasDeepZoom      bool   `json:"delving_hasDeepZoom"`
}

// NewLegacy returns a legacy struct with default values
func NewLegacy(indexDoc map[string]interface{}, fb *FragmentBuilder) *Legacy {
	l := &Legacy{
		HubID:      fb.fg.GetHubID(),
		RecordType: "void_edmrecord",
		Spec:       fb.fg.GetSpec(),
		OrgID:      fb.fg.GetOrgID(),
		Collection: fb.fg.GetSpec(),
	}
	var ok bool
	if _, ok = indexDoc["nave_GeoHash"]; ok {
		l.HasGeoHash = ok
	}
	if _, ok = indexDoc["edm_isShownBy"]; ok {
		l.HasDigitalObject = ok
	}
	if _, ok = indexDoc["nave_DeepZoomUrl"]; ok {
		l.HasDeepZoom = ok
	}
	if _, ok = indexDoc["edm_isShownAt"]; ok {
		l.HasLandingPage = ok
	}
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
	HasGeoHash         bool   `json:"hasGeoHash"`
	HasDigitalObject   bool   `json:"hasDigitalObject"`
	HasLandingPage     bool   `json:"hasLandingPage"`
	HasDeepZoom        bool   `json:"hasDeepZoom"`
}

// NewSystem generates system info for the V1 doc
func NewSystem(indexDoc map[string]interface{}, fb *FragmentBuilder) *System {
	s := &System{}
	s.Slug = fb.fg.GetHubID()
	s.Spec = fb.fg.GetSpec()
	s.Preview = fmt.Sprintf("detail/foldout/void_edmrecord/%s", fb.fg.GetHubID())
	s.SourceGraph = string(fb.fg.GetRDF())
	now := time.Now()
	nowString := fmt.Sprintf(now.Format("2018-02-13T12:44:31.432985"))
	s.CreatedAt = nowString
	s.ModifiedAt = nowString
	var ok bool
	if _, ok = indexDoc["nave_GeoHash"]; ok {
		s.HasGeoHash = ok
	}
	if _, ok = indexDoc["edm_isShownBy"]; ok {
		s.HasDigitalObject = ok
	}
	if _, ok = indexDoc["nave_DeepZoomUrl"]; ok {
		s.HasDeepZoom = ok
	}
	if _, ok = indexDoc["edm_isShownAt"]; ok {
		s.HasLandingPage = ok
	}
	// todo
	//s.Caption = ""
	//s.Thumbnail = ""
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
//'hasGeoHash': "true" if bindings.has_geo() else "false",
//'hasDigitalObject': "true" if thumbnail else "false",
//'hasLandingePage': "true" if 'edm_isShownAt' in index_doc else "false",
//'hasDeepZoom': "true" if 'nave_deepZoom' in index_doc else "false",

// NewGraphFromTurtle creates a RDF graph from the 'text/turtle' format
func NewGraphFromTurtle(re io.Reader) (*r.Graph, error) {
	g := r.NewGraph("")
	err := g.Parse(re, "text/turtle")
	if err != nil {
		log.Println("Unable to parse the supplied turtle RDF.")
		return g, err
	}
	if g.Len() == 0 {
		log.Println("No triples were added to the graph")
		return g, fmt.Errorf("no triples were added to the graph")
	}
	return g, nil
}

// CreateV1IndexDoc creates a map that can me marshaled to json
func CreateV1IndexDoc(fb *FragmentBuilder) (map[string]interface{}, error) {
	indexDoc := make(map[string]interface{})
	// todo create NS from predicate map
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
		entry, err := CreateV1IndexEntry(t)
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
func CreateV1IndexEntry(t *r.Triple) (*IndexEntry, error) {
	ie := &IndexEntry{}

	switch t.Object.(type) {
	case *r.Resource:
		ie.Type = "URIRef"
		ie.ID = t.Object.RawValue()
		// todo replace with getLabel
		ie.Value = t.Object.RawValue()
		ie.Raw = t.Object.RawValue()
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
		Id(id).
		Doc(indexDoc)
	return r, nil
}
