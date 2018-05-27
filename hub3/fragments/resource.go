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

	c "github.com/delving/rapid-saas/config"
	r "github.com/kiivihal/rdf2go"
)

// FragmentReferrerContext holds the referrer in formation for creating new fragments
type FragmentReferrerContext struct {
	Subject      string   `json:"subject"`
	SubjectClass []string `json:"subjectClass"`
	Predicate    string   `json:"predicate"`
	SearchLabel  string   `json:"searchLabel"`
	Level        int      `json:"level"`
	ObjectID     string   `json:"objectID"`
	// todo: decide if the sortKey belongs here
	//SortKey         int      `json:"sortKey"`
}

// NewContext returns the context for the current fragmentresource
func (fr *FragmentResource) NewContext(predicate, objectID string) *FragmentReferrerContext {
	label, err := c.Config.NameSpaceMap.GetSearchLabel(predicate)
	if err != nil {
		log.Printf("Unable to create search label for %s  due to %s\n", predicate, err)
		label = ""
	}

	return &FragmentReferrerContext{
		Subject:      fr.ID,
		SubjectClass: fr.Types,
		Predicate:    predicate,
		Level:        fr.GetLevel(),
		ObjectID:     objectID,
		SearchLabel:  label,
	}
}

// ResourceMap is a convenience structure to hold the resourceMap data and functions
type ResourceMap struct {
	resources map[string]*FragmentResource `json:"resources"`
}

// FragmentResource holds all the conttext information for a resource
// It works together with the FragmentBuilder to create the linked fragments
type FragmentResource struct {
	ID                   string                      `json:"id"`
	Types                []string                    `json:"types"`
	GraphExternalContext []*FragmentReferrerContext  `json:"graphExternalContext"`
	Context              []*FragmentReferrerContext  `json:"context"`
	Predicates           map[string][]*FragmentEntry `json:""`
	ObjectIDs            []*FragmentReferrerContext  `json:"objectIDs"`
}

func (rm *ResourceMap) GetSubject(uri string) (*FragmentResource, bool) {
	subject, ok := rm.Get(uri)
	return subject, ok
}

/*

workflow:

 - Add ReferrerContext during the first run
 - get the subject  FragmentResource
 - loop over the ObjectIDs (todo needs to have a better descriptive name)
 - get from fragment resource map
 - insert ReferrerContext into FragmentResource context block
 - Set level of the ReferrerContext (or better set it at current level plus 1)
 - recurse into ObjectIDs until you reach level 3 (break at level 4; this should also not be part of the grap)
	- break should result in moving onto next object Id on level 2 or 1
 - When done create fragments.


 TODO: restructure fragments into blocks with header, geoblock, context flock (maybe nested)

*/

// FragmentEntry holds all the information for the object of a rdf2go.Triple
type FragmentEntry struct {
	ID        string            `json:"@id,omitempty"`
	Value     string            `json:"@value,omitempty"`
	Language  string            `json:"@language,omitempty"`
	Datatype  string            `json:"@type,omitempty"`
	Entrytype string            `json:"entrytype"`
	Inline    *FragmentResource `json:"inline"`
}

// NewResourceMap creates a map for all the resources in the rdf2go.Graph
func NewResourceMap(g *r.Graph) (*ResourceMap, error) {
	rm := &ResourceMap{make(map[string]*FragmentResource)}

	if g.Len() == 0 {
		return rm, fmt.Errorf("The graph cannot be empty")
	}

	for t := range g.IterTriples() {
		err := AppendTriple(rm.resources, t)
		if err != nil {
			return rm, err
		}
	}
	return rm, nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func containsContext(s []*FragmentReferrerContext, e *FragmentReferrerContext) bool {
	for _, a := range s {
		if a.ObjectID == e.ObjectID && a.Predicate == e.Predicate {
			return true
		}
	}
	return false
}

func debrack(s string) string {
	if len(s) < 2 {
		return s
	}
	if s[0] != '<' {
		return s
	}
	if s[len(s)-1] != '>' {
		return s
	}
	return s[1 : len(s)-1]
}

// CreateFragmentEntry creates a FragmentEntry from a triple
func CreateFragmentEntry(t *r.Triple) (*FragmentEntry, string) {
	entry := &FragmentEntry{}
	switch o := t.Object.(type) {
	case *r.Resource:
		id := r.GetResourceID(o)
		entry.ID = r.GetResourceID(o)
		entry.Entrytype = "Resource"
		return entry, id
	case *r.BlankNode:
		id := r.GetResourceID(o)
		entry.ID = r.GetResourceID(o)
		entry.Entrytype = "Bnode"
		return entry, id
	case *r.Literal:
		entry.Value = o.Value
		entry.Entrytype = "Literal"
		if o.Datatype != nil && len(o.Datatype.String()) > 0 {
			if o.Datatype.String() != "<http://www.w3.org/2001/XMLSchema#string>" {
				entry.Datatype = debrack(o.Datatype.String())
			}
		}
		if len(o.Language) > 0 {
			entry.Language = o.Language
		}
	}
	return entry, ""
}

// AppendTriple appends a triple to a subject map
func AppendTriple(resources map[string]*FragmentResource, t *r.Triple) error {
	id := t.GetSubjectID()
	fr, ok := resources[id]
	if !ok {
		fr = &FragmentResource{}
		fr.ID = id
		resources[id] = fr
		fr.Predicates = make(map[string][]*FragmentEntry)
	}

	ttype, ok := t.GetRDFType()
	if ok {
		if !contains(fr.Types, ttype) {
			fr.Types = append(fr.Types, ttype)
		}
		return nil
	}

	p := r.GetResourceID(t.Predicate)
	predicates, ok := fr.Predicates[p]
	if !ok {
		predicates = []*FragmentEntry{}
	}
	entry, fragID := CreateFragmentEntry(t)
	if fragID != "" {
		if fragID != id {
			ctx := fr.NewContext(p, fragID)
			if !containsContext(fr.ObjectIDs, ctx) {
				fr.ObjectIDs = append(fr.ObjectIDs, ctx)
			}
		}
	}
	// TODO check for duplicates
	fr.Predicates[p] = append(predicates, entry)

	return nil
}

// Resources returns the map
func (rm *ResourceMap) Resources() map[string]*FragmentResource {
	return rm.resources
}

// Get returns a Fragment resource from the ResourceMap
func (rm *ResourceMap) Get(subject string) (*FragmentResource, bool) {
	fr, ok := rm.resources[subject]
	return fr, ok
}

// GetLevel returns the relative level that this resource has from the root
// or parent resource
func (fr *FragmentResource) GetLevel() int {
	return len(fr.Context) + 1
}

// NewContext returns the context for the current fragmentresource

// AddContext Adds the referrer context to the FragmentResource

// AddGraphExternalContext

// InlineFragmentResources
