// Copyright 2017 Delving B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
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
	"io"
	"log"
	"strings"

	c "github.com/delving/hub3/config"
	"github.com/delving/hub3/ikuzo/domain/domainpb"
	rdf "github.com/kiivihal/gon3"
	r "github.com/kiivihal/rdf2go"
)

// parseTurtleFile creates a graph from an uploaded file
func parseTurtleFile(r io.Reader) (*rdf.Graph, error) {
	parser := rdf.NewParser("")
	g, err := parser.Parse(r)
	return g, err
}

func rdf2term(term rdf.Term) r.Term {
	switch term := term.(type) {
	case *rdf.BlankNode:
		return r.NewBlankNode(term.RawValue())
	case *rdf.Literal:
		if len(term.LanguageTag) > 0 {
			return r.NewLiteralWithLanguage(term.LexicalForm, term.LanguageTag)
		}
		if term.DatatypeIRI != nil && len(term.DatatypeIRI.String()) > 0 {
			return r.NewLiteralWithDatatype(term.LexicalForm, r.NewResource(debrack(term.DatatypeIRI.String())))
		}
		return r.NewLiteral(term.RawValue())
	case *rdf.IRI:
		return r.NewResource(term.RawValue())
	}
	return nil
}

type RDFUploader struct {
	OrgID        string
	Spec         string
	SubjectClass string
	TypeClassURI string
	IDSplitter   string
	Revision     int32
	rm           *ResourceMap
	subjects     []string
}

func (upl *RDFUploader) createResourceMap(g *rdf.Graph) (*ResourceMap, error) {
	rm := NewEmptyResourceMap(upl.OrgID)
	idx := 0
	for t := range g.IterTriples() {
		idx++
		if t.Predicate.RawValue() == upl.TypeClassURI && t.Object.RawValue() == upl.SubjectClass {
			upl.subjects = append(upl.subjects, t.Subject.RawValue())
		}
		newTriple := r.NewTriple(rdf2term(t.Subject), rdf2term(t.Predicate), rdf2term(t.Object))
		err := rm.AppendOrderedTriple(newTriple, false, idx)
		if err != nil {
			return nil, err
		}
	}

	return rm, nil
}

func NewRDFUploader(orgID, spec, subjectClass, typePredicate, idSplitter string, revision int) *RDFUploader {
	return &RDFUploader{
		OrgID:        orgID,
		Spec:         spec,
		SubjectClass: subjectClass,
		TypeClassURI: typePredicate,
		IDSplitter:   idSplitter,
		Revision:     int32(revision),
	}
}

func (upl *RDFUploader) Parse(r io.Reader) (*ResourceMap, error) {
	g, err := parseTurtleFile(r)
	if err != nil {
		return nil, err
	}
	rm, err := upl.createResourceMap(g)
	if err != nil {
		return nil, err
	}
	upl.rm = rm
	log.Printf("number of subjects: %d", len(upl.subjects))
	return rm, nil
}

func (upl *RDFUploader) createFragmentGraph(subject string) (*FragmentGraph, error) {
	if !strings.Contains(subject, upl.IDSplitter) {
		return nil, fmt.Errorf("unable to find localID with splitter %s in %s", upl.IDSplitter, subject)
	}
	parts := strings.Split(subject, upl.IDSplitter)
	localID := parts[len(parts)-1]

	header := &Header{
		OrgID:         upl.OrgID,
		Spec:          upl.Spec,
		Revision:      upl.Revision,
		HubID:         fmt.Sprintf("%s_%s_%s", upl.OrgID, upl.Spec, localID),
		DocType:       FragmentGraphDocType,
		EntryURI:      subject,
		NamedGraphURI: fmt.Sprintf("%s/graph", subject),
		Modified:      NowInMillis(),
		Tags:          []string{"sourceUpload"},
	}

	fg := NewFragmentGraph()
	fg.Meta = header
	fg.SetResources(upl.rm)
	return fg, nil
}

func (upl *RDFUploader) SaveFragmentGraphs(bi BulkIndex) (int, error) {
	var seen int
	for _, s := range upl.subjects {
		seen++

		fg, err := upl.createFragmentGraph(s)
		if err != nil {
			return 0, err
		}

		fg.SetResources(upl.rm)

		m, err := fg.IndexMessage()
		if err != nil {
			return 0, err
		}

		err = bi.Publish(context.Background(), m)
		if err != nil {
			log.Printf("can't publish records: %v", err)
			return 0, err
		}

		fb := NewFragmentBuilder(fg)
		graph := r.NewGraph("")
		for _, rsc := range fg.Resources {
			for _, t := range rsc.GenerateTriples() {
				graph.Add(t)
			}
		}
		fb.Graph = graph

		// TODO(kiivihal): add support for parsing the graph
		fb.GetSortedWebResources(ctx)

		indexDoc, err := CreateV1IndexDoc(fb)
		if err != nil {
			log.Printf("Unable to create index doc: %s", err)
			return 0, err
		}

		b, err := json.Marshal(indexDoc)
		if err != nil {
			return 0, err
		}

		m = &domainpb.IndexMessage{
			OrganisationID: fg.Meta.OrgID,
			DatasetID:      fg.Meta.Spec,
			RecordID:       fg.Meta.HubID,
			IndexType:      domainpb.IndexType_V1,
			Source:         b,
		}

		if err := bi.Publish(ctx, m); err != nil {
			return 0, err
		}

		// TODO store sparql updates
	}

	return seen, nil
}

func (upl *RDFUploader) IndexFragments(bi BulkIndex) (int, error) {
	fg := NewFragmentGraph()
	fg.Meta = &Header{
		OrgID:    upl.OrgID,
		Revision: upl.Revision,
		DocType:  "sourceUpload",
		Spec:     upl.Spec,
		Tags:     []string{"sourceUpload"},
		Modified: NowInMillis(),
	}

	triplesProcessed := 0
	sparqlUpdates := []SparqlUpdate{}

	for k, fr := range upl.rm.Resources() {
		fg.Meta.EntryURI = k
		fg.Meta.NamedGraphURI = fmt.Sprintf("%s/graph", k)
		frags, err := fr.CreateFragments(fg)
		if err != nil {
			return 0, err
		}

		var triples bytes.Buffer

		for _, frag := range frags {
			if c.Config.RDF.RDFStoreEnabled {
				_, err = triples.WriteString(frag.Triple + "\n")
				if err != nil {
					return 0, err
				}
			}
			frag.Meta.AddTags("sourceUpload")
			err := frag.AddTo(bi)
			if err != nil {
				return 0, err
			}
			triplesProcessed++
		}
		revision := int(fg.Meta.Revision)
		if c.Config.RDF.RDFStoreEnabled {
			su := SparqlUpdate{
				Triples:       triples.String(),
				NamedGraphURI: fg.Meta.NamedGraphURI,
				Spec:          fg.Meta.Spec,
				SpecRevision:  revision,
			}
			triples.Reset()
			sparqlUpdates = append(sparqlUpdates, su)
			if len(sparqlUpdates) >= 250 {
				// insert the triples
				_, errs := RDFBulkInsert(upl.OrgID, sparqlUpdates)
				if len(errs) != 0 {
					return 0, errs[0]
				}
				sparqlUpdates = []SparqlUpdate{}
			}
		}
	}

	if len(sparqlUpdates) != 0 {
		_, errs := RDFBulkInsert(upl.OrgID, sparqlUpdates)
		if len(errs) != 0 {
			return 0, errs[0]
		}
	}
	return triplesProcessed, nil
}
