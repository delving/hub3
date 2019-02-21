package fragments

import (
	fmt "fmt"
	"log"
	"mime/multipart"
	"strings"

	rdf "github.com/deiu/gon3"
	c "github.com/delving/rapid-saas/config"
	r "github.com/kiivihal/rdf2go"
	elastic "github.com/olivere/elastic"
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

func (upl *RDFUploader) createResourceMap(g *rdf.Graph) (*ResourceMap, error) {
	rm := NewEmptyResourceMap()
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

func (upl *RDFUploader) Parse(f multipart.File) (*ResourceMap, error) {
	g, err := parseTurtleFile(f)
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

func (upl *RDFUploader) SaveFragmentGraphs(p *elastic.BulkProcessor) (int, error) {
	var seen int
	// TODO store sparql updates
	for _, s := range upl.subjects {
		seen++
		fg, err := upl.createFragmentGraph(s)
		if err != nil {
			return 0, err
		}
		r := elastic.NewBulkIndexRequest().
			Index(c.Config.ElasticSearch.IndexName).
			Type(DocType).
			RetryOnConflict(3).
			Id(fg.Meta.HubID).
			Doc(fg)
		p.Add(r)
	}
	return seen, nil
}

func (upl *RDFUploader) IndexFragments(p *elastic.BulkProcessor) (int, error) {

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
