package fragments

import (
	fmt "fmt"
	"mime/multipart"

	rdf "github.com/deiu/gon3"
	r "github.com/kiivihal/rdf2go"
	elastic "gopkg.in/olivere/elastic.v5"
)

// parseTurtleFile creates a graph from an uploaded file
func parseTurtleFile(f multipart.File) (*rdf.Graph, error) {
	parser := rdf.NewParser("")
	g, err := parser.Parse(f)
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

func createResourceMap(g *rdf.Graph) (*ResourceMap, error) {
	rm := NewEmptyResourceMap()
	idx := 0
	for t := range g.IterTriples() {
		idx++
		newTriple := r.NewTriple(rdf2term(t.Subject), rdf2term(t.Predicate), rdf2term(t.Object))
		err := rm.AppendOrderedTriple(newTriple, false, idx)
		if err != nil {
			return nil, err
		}
	}
	return rm, nil
}

type RDFUploader struct {
	OrgID        string
	Spec         string
	SubjectClass string
	rm           *ResourceMap
}

func NewRDFUploader(orgID, spec, subjectClass string) *RDFUploader {
	return &RDFUploader{OrgID: orgID, Spec: spec, SubjectClass: subjectClass}
}

func (upl *RDFUploader) Parse(f multipart.File) (*ResourceMap, error) {
	g, err := parseTurtleFile(f)
	if err != nil {
		return nil, err
	}
	rm, err := createResourceMap(g)
	if err != nil {
		return nil, err
	}
	upl.rm = rm
	return rm, nil
}

func (upl *RDFUploader) SaveFragmentGraphs(p *elastic.BulkProcessor) (int, error) {
	// todo implement full saving with a scanner
	return 0, nil
}

func (upl *RDFUploader) IndexFragments(p *elastic.BulkProcessor, revision int) (int, error) {

	fg := NewFragmentGraph()
	fg.Meta = &Header{
		OrgID:    upl.OrgID,
		Revision: int32(revision),
		DocType:  "sourceUpload",
		Spec:     upl.Spec,
		Tags:     []string{"sourceUpload"},
		Modified: NowInMillis(),
	}

	triplesProcessed := 0
	for k, fr := range upl.rm.Resources() {
		fg.Meta.EntryURI = k
		fg.Meta.NamedGraphURI = fmt.Sprintf("%s/graph", k)
		frags, err := fr.CreateFragments(fg)
		if err != nil {
			return 0, err
		}

		for _, frag := range frags {
			frag.Meta.AddTags("sourceUpload")
			err := frag.AddTo(p)
			if err != nil {
				return 0, err
			}
			triplesProcessed++
		}
	}
	return triplesProcessed, nil
}
