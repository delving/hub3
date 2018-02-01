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
	"net/url"
	"strconv"
	"strings"

	c "bitbucket.org/delving/rapid/config"
	"github.com/OneOfOne/xxhash"
	r "github.com/deiu/rdf2go"
	elastic "gopkg.in/olivere/elastic.v5"
)

// DOCTYPE is the ElasticSearch doctype for the Fragment Struct
const DOCTYPE = "lodfragment"

// FragmentGraph holds all the information to build and store Fragments
type FragmentGraph struct {
	OrgID         string   `json:"orgID"`
	Spec          string   `json:"spec"`
	HubID         string   `json:"hubID"`
	Revision      int32    `json:"revision"`
	NamedGraphURI string   `json:"namedGraphURI"`
	Tags          []string `json:"tags"`
	Graph         *r.Graph `json:"graph"`
	MimeType      string   `json:"mimeType"`
}

// NewFragmentGraph creates a new instance of FragmentGraph
func NewFragmentGraph() *FragmentGraph {
	return &FragmentGraph{
		Graph: r.NewGraph(""),
	}
}

// ParseGraph creates a RDF2Go Graph
func (fg *FragmentGraph) ParseGraph(rdf io.Reader, mimeType string) error {
	var err error
	switch mimeType {
	case "text/turtle":
		err = fg.Graph.Parse(rdf, mimeType)
	case "application/ld+json":
		err = fg.Graph.Parse(rdf, mimeType)
	default:
		return fmt.Errorf(
			"Unsupported RDF mimeType %s. Currently, only 'text/turtle' and 'application/ld+json' are supported",
			mimeType,
		)
	}
	if err != nil {
		log.Printf("Unable to parse RDF string into graph: %v\n%s\n", err, rdf)
		return err
	}
	fg.MimeType = mimeType
	return nil
}

// SaveFragments creates and stores all the fragments
func (fg *FragmentGraph) SaveFragments(p *elastic.BulkProcessor) error {
	if fg.Graph.Len() == 0 {
		return fmt.Errorf("cannot store fragments from empty graph")
	}
	for t := range fg.Graph.IterTriples() {
		frag, err := fg.CreateFragment(t)
		if err != nil {
			log.Printf("Unable to create fragment due to %v.", err)
			return err
		}
		err = frag.AddTo(p)
		if err != nil {
			log.Printf("Unable to save fragment due to %v.", err)
			return err
		}
	}
	return nil
}

// CreateFragment creates a fragment from a triple
func (fg *FragmentGraph) CreateFragment(triple *r.Triple) (*Fragment, error) {
	f := &Fragment{
		Spec:          fg.Spec,
		Revision:      fg.Revision,
		NamedGraphURI: fg.NamedGraphURI,
		OrgID:         fg.OrgID,
		HubID:         fg.HubID,
	}
	f.Subject = triple.Subject.RawValue()
	f.Predicate = triple.Predicate.RawValue()
	f.Object = triple.Object.RawValue()
	f.Triple = triple.String()
	switch triple.Object.(type) {
	case *r.Literal:
		f.ObjectType = ObjectType_LITERAL
		f.ObjectTypeRaw = "literal"
		l := triple.Object.(*r.Literal)
		f.Language = l.Language
		// Set default datatypes
		f.DataType = ObjectXSDType_STRING
		f.XsdRaw, _ = f.GetDataType().GetLabel()
		if l.Datatype != nil {
			xsdType, err := GetObjectXSDType(l.Datatype.String())
			if err != nil {
				log.Printf("Unable to get xsdType for %s", l.Datatype.String())
				break
			}
			prefixLabel, err := xsdType.GetPrefixLabel()
			if err != nil {
				log.Printf("Unable to get xsdType prefix label for %s", l.Datatype.String())
				break
			}
			f.XsdRaw = prefixLabel
			f.DataType = xsdType
		}
	case *r.Resource:
		f.ObjectType = ObjectType_RESOURCE
		f.ObjectTypeRaw = "resource"
		f.TypeLink = f.IsTypeLink()
		if fg.Graph.Len() == 0 {
			log.Printf("Warn: Graph is empty can't do linking checks\n")
			break
		}
		f.GraphExternalLink = fg.IsGraphExternal(triple.Object)
		isDomainExternal, err := fg.IsDomainExternal(f.Object)
		if err != nil {
			log.Printf("Unable to parse object domain: %#v", err)
			break
		}
		f.DomainExternalLink = isDomainExternal
	default:
		return f, fmt.Errorf("unknown object type: %#v", triple.Object)
	}
	return f, nil
}

// IsDomainExternal checks if the object link points to another domain
func (fg *FragmentGraph) IsDomainExternal(obj string) (bool, error) {
	u, err := url.Parse(obj)
	if err != nil {
		return false, err
	}
	return !strings.Contains(c.Config.RDF.BaseURL, u.Host), nil
}

// IsGraphExternal checks if the object link points outside the current graph
func (fg *FragmentGraph) IsGraphExternal(obj r.Term) bool {
	found := fg.Graph.One(obj, nil, nil)
	return found == nil
}

// IsTypeLink checks if the Predicate is a RDF type link
func (f Fragment) IsTypeLink() bool {
	return f.Predicate == "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"
}

// NewFragmentRequest creates a finder for Fragments
// Use the funcs to setup filters and search properties
// then call Find to execute.
func NewFragmentRequest() *FragmentRequest {
	fr := &FragmentRequest{}
	fr.Page = int32(1)
	return fr
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
		Type(DOCTYPE).Id(f.ID()).
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

// Find executes the search and returns a response
func (fr FragmentRequest) Find(client *elastic.Client) (FragmentResponse, error) {
	var resp FragmentResponse
	// TODO: implement the search
	return resp, nil
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
		return "<`1`>", err
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
