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
	"bytes"
	"context"
	"encoding/json"
	fmt "fmt"
	"log"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/OneOfOne/xxhash"
	c "github.com/delving/rapid-saas/config"
	r "github.com/kiivihal/rdf2go"
	//elastic "github.com/olivere/elastic"
	elastic "gopkg.in/olivere/elastic.v5"
)

// FragmentDocType is the ElasticSearch doctype for the Fragment
const FragmentDocType = "fragment"

// FragmentGraphDocType is the ElasticSearch doctype for the FragmentGraph
const FragmentGraphDocType = "graph"

// DocType is the default doctype since elasticsearch deprecated mapping types
const DocType = "doc"

// SIZE of the fragments returned
const SIZE = 100

// RDFType is the URI for RDF:type
const RDFType = "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"

// GetAboutURI returns the subject of the FragmentGraph
func (fg *FragmentGraph) GetAboutURI() string {
	return strings.TrimSuffix(fg.GetNamedGraphURI(), "/graph")
}

// AddTags adds a tag to the fragment tag list
func (f *Fragment) AddTags(tag ...string) {
	for _, t := range tag {
		f.Meta.Tags = append(f.Meta.Tags, t)
	}
}

// CreateLodKey returns the path including the # fragments from the subject URL
// This is used for the Linked Open Data resolving
func (f *Fragment) CreateLodKey() (string, error) {
	u, err := url.Parse(f.GetSubject())
	if err != nil {
		return "", err
	}
	lodKey := u.Path
	if c.Config.LOD.SingleEndpoint == "" {
		lodResourcePrefix := fmt.Sprintf("/%s", c.Config.LOD.Resource)
		if !strings.HasPrefix(u.Path, lodResourcePrefix) {
			return "", nil
		}
		lodKey = strings.TrimPrefix(u.Path, lodResourcePrefix)
	}
	if u.Fragment != "" {
		lodKey = fmt.Sprintf("%s#%s", lodKey, u.Fragment)
	}
	return lodKey, nil
}

// TODO remove later
// AddHeader adds header information for stand-alone fragments.
// When Fragments are embedded inside a FragmentGraph this information is
// redundant.
//func (f *Fragment) AddHeader(fb *FragmentBuilder) error {
//f.DocType = FragmentDocType
//f.Spec = fb.fg.GetSpec()
//f.Revision = fb.fg.GetRevision()
//f.NamedGraphURI = fb.fg.GetNamedGraphURI()
//f.OrgID = fb.fg.GetOrgID()
//f.HubID = fb.fg.GetHubID()
//lodKey, err := f.CreateLodKey()
//if err != nil {
//return err
//}
//if lodKey != "" {
//f.LodKey = lodKey
//}
//return nil

//}

// IsTypeLink checks if the Predicate is a RDF type link
func (f Fragment) IsTypeLink() bool {
	return f.Predicate == RDFType
}

// NewFragmentRequest creates a finder for Fragments
// Use the funcs to setup filters and search properties
// then call Find to execute.
func NewFragmentRequest() *FragmentRequest {
	fr := &FragmentRequest{}
	fr.Page = int32(1)
	return fr
}

// AssignObject cleans the object string and sets the language when applicable
func (fr *FragmentRequest) AssignObject(o string) {
	if strings.Contains(o, "@") {
		parts := strings.Split(o, "@")
		o = parts[0]
		if len(parts[1]) > 0 {
			fr.Language = parts[1]
		}
	}
	if len(o) > 0 && o[0] == '"' {
		o = o[1:]
	}
	if len(o) > 0 && o[len(o)-1] == '"' {
		o = o[:len(o)-1]
	}
	fr.Object = o
}

// ParseQueryString sets the FragmentRequest values from url.Values
func (fr *FragmentRequest) ParseQueryString(v url.Values) error {
	for k, v := range v {
		switch k {
		case "subject":
			fr.Subject = v[0]
		case "predicate":
			fr.Predicate = v[0]
		case "object":
			fr.Object = v[0]
		case "language":
			fr.Language = v[0]
		case "graph":
			fr.Graph = v[0]
		case "page":
			page, err := strconv.ParseInt(v[0], 10, 32)
			if err != nil {
				return fmt.Errorf("Unable to convert page %s into an int32", v[0])
			}
			fr.Page = int32(page)
		default:
			return fmt.Errorf("unknown ")
		}
	}
	return nil
}

func buildQueryClause(q *elastic.BoolQuery, fieldName string, fieldValue string) *elastic.BoolQuery {
	searchField := fmt.Sprintf("%s.keyword", fieldName)
	if len(fieldValue) == 0 {
		return q
	}
	if strings.HasPrefix("-", fieldValue) {
		fieldValue = strings.TrimPrefix(fieldValue, "-")
		return q.MustNot(elastic.NewTermQuery(searchField, fieldValue))
	}
	return q.Must(elastic.NewTermQuery(searchField, fieldValue))
}

// GetESPage returns the 0 based page for Elastic Search
func (fr FragmentRequest) GetESPage() int {
	if fr.GetPage() < 2 {
		return 0
	}
	return int((fr.GetPage() * SIZE) - 1)
}

// Find returns a list of matching LodFragments
func (fr FragmentRequest) Find(ctx context.Context, client *elastic.Client) (*r.Graph, error) {
	q := elastic.NewBoolQuery()
	buildQueryClause(q, "subject", fr.GetSubject())
	buildQueryClause(q, "predicate", fr.GetPredicate())
	buildQueryClause(q, "object", fr.GetObject())
	q = q.Must(elastic.NewTermQuery("docType", FragmentDocType))
	if len(fr.GetSpec()) != 0 {
		q = q.Must(elastic.NewTermQuery("spec", fr.GetSpec()))
	}
	if c.Config.DevMode {
		src, err := q.Source()
		if err != nil {
			log.Fatal("Unable get query source")
			return &r.Graph{}, err
		}
		data, err := json.Marshal(src)
		if err != nil {
			log.Fatal("Unable get query source")
			return &r.Graph{}, err
		}
		fmt.Println(string(data))
	}
	res, err := client.Search().
		Index(c.Config.ElasticSearch.IndexName).
		Query(q).
		Size(SIZE).
		From(fr.GetESPage()).
		Do(ctx)
	if err != nil {
		return &r.Graph{}, err
	}
	var buffer bytes.Buffer
	if res == nil {
		log.Printf("expected response != nil; got: %v", res)
		return &r.Graph{}, fmt.Errorf("expected response != nil")
	}
	if res.Hits.TotalHits == 0 {
		log.Println("Nothing found for this query.")
		return &r.Graph{}, nil
	}
	var frtyp Fragment
	for _, item := range res.Each(reflect.TypeOf(frtyp)) {
		frag := item.(Fragment)
		buffer.WriteString(fmt.Sprintln(frag.Triple))
		//triples = append(triples, frag.Triple)
	}
	//g := CreateHyperMediaControlGraph(fr.GetSpec(), res.Hits.TotalHits, 1)
	g := r.NewGraph("")
	err = g.Parse(&buffer, "text/turtle")
	if err != nil {
		log.Printf("unable to parse triples from result: %s", err)
		return g, err
	}
	return g, nil
}

// CreateHyperMediaControlGraph creates a graph based on the triple-pattern-fragment spec
// see http://www.hydra-cg.com/spec/latest/triple-pattern-fragments/#controls
func CreateHyperMediaControlGraph(spec string, total int64, page int) *r.Graph {
	g := r.NewGraph("")
	hits := strconv.Itoa(int(total))
	subject := r.NewResource(fmt.Sprintf("%s/fragments/%s", c.Config.RDF.BaseURL, spec))
	g.AddTriple(subject, r.NewResource("http://rdfs.org/ns/void#subset"), subject)
	g.AddTriple(
		subject,
		r.NewResource("http://www.w3.org/1999/02/22-rdf-syntax-ns#type"),
		r.NewResource("http://www.w3.org/ns/hydra/core#Collection"),
	)
	g.AddTriple(
		subject,
		r.NewResource("http://www.w3.org/1999/02/22-rdf-syntax-ns#type"),
		r.NewResource("http://www.w3.org/ns/hydra/core#PagedCollection"),
	)

	g.AddTriple(
		subject,
		r.NewResource("http://purl.org/dc/terms/title"),
		r.NewLiteralWithLanguage(fmt.Sprintf("Linked Data Fragment of %s", spec), "en"),
	)
	g.AddTriple(
		subject,
		r.NewResource("http://purl.org/dc/terms/description"),
		r.NewLiteralWithLanguage(
			fmt.Sprintf(
				"Triple Pattern Fragment of the '%s' dataset containing triples matching the pattern { ?s ?p ?o  }.",
				spec,
			),
			"en"),
	)
	g.AddTriple(
		subject,
		r.NewResource("http://purl.org/dc/term/source"),
		subject,
	)
	g.AddTriple(
		subject,
		r.NewResource("http://www.w3.org/ns/hydra/core#totalItems"),
		r.NewLiteralWithDatatype(hits, r.NewResource("http://www.w3.org/2001/XMLSchema#integer")),
	)
	g.AddTriple(
		subject,
		r.NewResource("http://rdfs.org/ns/void#triples"),
		r.NewLiteralWithDatatype(hits, r.NewResource("http://www.w3.org/2001/XMLSchema#integer")),
	)
	g.AddTriple(
		subject,
		r.NewResource("http://rdfs.org/ns/void#triples"),
		r.NewLiteralWithDatatype(hits, r.NewResource("http://www.w3.org/2001/XMLSchema#integer")),
	)
	g.AddTriple(
		subject,
		r.NewResource("http://www.w3.org/ns/hydra/core#itemsPerPage"),
		r.NewLiteralWithDatatype("100", r.NewResource("http://www.w3.org/2001/XMLSchema#integer")),
	)
	g.AddTriple(
		subject,
		r.NewResource("http://www.w3.org/ns/hydra/core#firstPage"),
		r.NewLiteral("1"),
	)
	return g
}

// CreateHash creates an xxhash-based hash of a string
func CreateHash(input string) string {
	hash := xxhash.Checksum64([]byte(input))
	return fmt.Sprintf("%016x", hash)
}

// Quad returns a RDF Quad from the Fragment
func (f Fragment) Quad() string {
	// remove trailing period
	cleanTriple := strings.TrimSuffix(f.GetTriple(), " .")
	return fmt.Sprintf("%s <%s> .", cleanTriple, f.GetNamedGraphURI())
}

// ID is the hashed identifier of the Fragment Quad field.
// This is used as identifier by the storage layer.
func (f Fragment) ID() string {
	return CreateHash(f.Quad())
}

// CreateBulkIndexRequest converts the fragment into a request that can be
// submitted to the ElasticSearch BulkIndexService
func (f Fragment) CreateBulkIndexRequest() (*elastic.BulkIndexRequest, error) {
	r := elastic.NewBulkIndexRequest().
		Index(c.Config.ElasticSearch.IndexName).
		Type(DocType).
		Id(f.ID()).
		Doc(f)
	return r, nil
}

// AddTo adds the BulkableRequest to the Storage interface where it is flushed periodically.
func (f Fragment) AddTo(p *elastic.BulkProcessor) error {
	cbr, err := f.CreateBulkIndexRequest()
	if err != nil {
		return err
	}
	p.Add(cbr)
	return nil
}

// GetLabel retrieves the XSD label of the ObjectXSDType
func (t ObjectXSDType) GetLabel() (string, error) {
	label, ok := objectXSDType2XSDLabel[int32(t)]
	if !ok {
		return "", fmt.Errorf("%s has no xsd label", t.String())
	}
	return label, nil
}

// GetPrefixLabel retrieves the XSD label of the ObjectXSDType with xsd: prefix.
func (t ObjectXSDType) GetPrefixLabel() (string, error) {
	label, err := t.GetLabel()
	if err != nil {
		return "", err
	}
	return strings.Replace(label, "http://www.w3.org/2001/XMLSchema#", "xsd:", 1), nil
}

// GetObjectXSDType returns the ObjectXSDType from a valid XSD label
func GetObjectXSDType(label string) (ObjectXSDType, error) {
	if len(xsdLabel2ObjectXSDType) == 0 {
		for k, v := range objectXSDType2XSDLabel {
			xsdLabel2ObjectXSDType[v] = k
		}
	}
	if strings.HasPrefix(label, "<") || strings.HasSuffix(label, ">") {
		label = strings.TrimPrefix(label, "<")
		label = strings.TrimSuffix(label, ">")
	}
	typeInt, ok := xsdLabel2ObjectXSDType[label]
	if !ok {
		return ObjectXSDType_STRING, fmt.Errorf("xsd:label %s has no ObjectXSDType", label)
	}
	t, ok := int2ObjectXSDType[typeInt]
	if !ok {
		return ObjectXSDType_STRING, fmt.Errorf("xsd:label %s has no ObjectXSDType", label)
	}
	return t, nil
}

// SaveDataSet creates a fragment entry for a Dataset
func SaveDataSet(spec string, p *elastic.BulkProcessor) error {
	fg := NewFragmentGraph()
	fg.Spec = "datasets"
	fb := NewFragmentBuilder(fg)
	subject := r.NewResource(fmt.Sprintf("%s/fragments/%s", c.Config.RDF.BaseURL, spec))
	fb.Graph.AddTriple(
		subject,
		r.NewResource("http://www.w3.org/1999/02/22-rdf-syntax-ns#type"),
		r.NewResource("http://rdfs.org/ns/void#Dataset"),
	)
	fb.Graph.AddTriple(subject, r.NewResource("http://www.w3.org/2000/01/rdf-schema#label"), r.NewLiteral(spec))
	fb.Graph.AddTriple(subject, r.NewResource("http://purl.org/dc/terms/title"), r.NewLiteral(spec))
	// TODO add new fragment builder here
	//return fb.CreateFragments(p, false, true)
	return nil
}

// ESMapping is the default mapping for the RDF records enabled by rapid
// TODO update the mapping to new resource structure with nesting
var ESMapping = `{
	"settings":{
		"number_of_shards":3,
		"number_of_replicas":2
	},
	"mappings":{
		"doc": {
			"properties": {
				"spec": {"type": "keyword"},
				"orgID": {"type": "keyword"},
				"objectNumber": {"type": "keyword"},
				"hubID": {"type": "text", "fields": {"keyword": {"type": "keyword"}}},
				"revision": {"type": "long"},
				"entryURI": {"type": "keyword"},
				"namedGraphURI": {"type": "keyword"},
				"RDF": {"type": "binary", "index": "false", "store": "false"},
				"rdfMimeType": {"type": "text", "fields": {"keyword": {"type": "keyword"}}},
				"tags": {"type": "keyword"},
				"LastModified": {"type": "date"},
				"docType": {"type": "keyword"},
				"level": {"type": "long"},
				"fragments": {
					"type": "nested",
					"properties": {
						"object": {"type": "text", "fields": {"keyword": {"type": "keyword", "ignore_above": 256}}}
					}
				},
				"object": {"type": "text", "fields": {"keyword": {"type": "keyword"}}},
				"stats": {
					"type": "object"
				}
			}
		}
}}`
