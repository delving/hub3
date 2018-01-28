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
	"log"

	r "github.com/deiu/rdf2go"
	elastic "gopkg.in/olivere/elastic.v5"
)

// FragmentGraph holds all the information to build and store Fragments
type FragmentGraph struct {
	Spec          string
	Revision      int32
	NamedGraphURI string
}

// CreateFragment creates a fragment from a triple
func (fg *FragmentGraph) CreateFragment(triple *r.Triple) (*Fragment, error) {
	f := &Fragment{
		Spec:          fg.Spec,
		Revision:      fg.Revision,
		NamedGraphURI: fg.NamedGraphURI,
	}
	f.Subject = triple.Subject.RawValue()
	f.Predicate = triple.Predicate.RawValue()
	f.Object = triple.Object.RawValue()
	switch triple.Object.(type) {
	case *r.Literal:
		f.ObjectType = ObjectType_LITERAL
		l := triple.Object.(*r.Literal)
		log.Printf("lang: %s\n", l.Language)
		f.Language = l.Language
		//f.DataType = l.Datatype.String()
	case *r.Resource:
		f.ObjectType = ObjectType_RESOURCE
	default:
		return f, fmt.Errorf("unknown object type: %#v", triple.Object)
	}
	//fmt.Printf("%#v", triple.Object.Language)
	return f, nil
}

// NewFragmentRequest creates a finder for Fragments
// Use the funcs to setup filters and search properties
// then call Find to execute.
func NewFragmentRequest() *FragmentRequest {
	return &FragmentRequest{}
}

// Find executes the search and returns a response
func (fr FragmentRequest) Find(client *elastic.Client) (FragmentResponse, error) {
	var resp FragmentResponse
	// TODO: implement the search
	return resp, nil
}
