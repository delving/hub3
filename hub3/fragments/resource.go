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

	r "github.com/kiivihal/rdf2go"
)

// FragmentReferrerContext holds the referrer in formation for creating new fragments
type FragmentReferrerContext struct {
	Subject         string   `json:"subject"`
	SubjectClass    []string `json:"subjectClass"`
	Predicate       string   `json:"predicate"`
	SearchLabel     string   `json:"searchLabel"`
	Level           int      `json:"level"`
	FragmentSubject string   `json:"fragmentSubject"`
	SortKey         int      `json:"sortKey"`
}

// FragmentResource holds all the conttext information for a resource
// It works together with the FragmentBuilder to create the linked fragments
type FragmentResource struct {
	ID         string                      `json:"id"`
	Types      []string                    `json:"types"`
	Context    []FragmentReferrerContext   `json:"context"`
	Predicates map[string][]*FragmentEntry `json:""`
	ObjectIDs  []string                    `json:"objectIDs"`
}

// FragmentEntry holds all the information for the object of a rdf2go.Triple
type FragmentEntry struct {
	ID        string `json:"@id,omitempty"`
	Value     string `json:"@value,omitempty"`
	Language  string `json:"@language,omitempty"`
	Datatype  string `json:"@type,omitempty"`
	Entrytype string `json:"entrytype"`
}

// CreateResourceMap creates a map for all the resources in the rdf2go.Graph
func CreateResourceMap(g *r.Graph) (map[string]*FragmentResource, error) {
	rm := make(map[string]*FragmentResource)

	if g.Len() == 0 {
		return rm, fmt.Errorf("The graph cannot be empty")
	}

	for t := range g.IterTriples() {
		err := AppendTriple(rm, t)
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
		if !contains(fr.ObjectIDs, fragID) && fragID != id {
			fr.ObjectIDs = append(fr.ObjectIDs, fragID)
		}

	}
	// TODO check for duplicates
	fr.Predicates[p] = append(predicates, entry)

	return nil
}
