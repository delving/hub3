package fragments

import (
	"encoding/csv"
	fmt "fmt"
	"io"
	"strings"

	"github.com/delving/rapid-saas/config"
	r "github.com/kiivihal/rdf2go"
	elastic "gopkg.in/olivere/elastic.v5"
)

// CSVConvertor holds all values to convert a CSV to RDF
type CSVConvertor struct {
	SubjectColumn         string    `json:"subjectColumn"`
	Separator             string    `json:"separator"`
	PredicateURIBase      string    `json:"predicateURIBase"`
	SubjectClass          string    `json:"subjectClass"`
	SubjectURIBase        string    `json:"subjectURIBase"`
	ObjectURIFormat       string    `json:"objectURIFormat"`
	ObjectResourceColumns []string  `json:"objectResourceColumns"`
	ThumbnailURIBase      string    `json:"thumbnailURIBase"`
	ThumbnailColumn       string    `json:"thumbnailColumn"`
	DefaultSpec           string    `json:"defaultSpec"`
	InputFile             io.Reader `json:"inputFile"`
	RowsProcessed         int       `json:"rowsProcessed"`
	TriplesCreated        int       `json:"triplesCreated"`
}

// NewCSVConvertor creates a CSV convertor from an net/http Form
func NewCSVConvertor() *CSVConvertor {
	return &CSVConvertor{}
}

// IndexFragments stores the fragments generated from the CSV into ElasticSearch
func (con *CSVConvertor) IndexFragments(p *elastic.BulkProcessor, revision int) (int, int, error) {

	fg := NewFragmentGraph()
	fg.Meta = &Header{
		OrgID:    config.Config.OrgID,
		Revision: int32(revision),
		DocType:  "csvUpload",
		Spec:     con.DefaultSpec,
		Tags:     []string{"csvUpload"},
	}

	rm, rowsProcessed, err := con.Convert()

	if err != nil {
		return 0, 0, err
	}

	triplesProcessed := 0
	for k, fr := range rm.Resources() {
		fg.Meta.EntryURI = k
		fg.Meta.NamedGraphURI = fmt.Sprintf("%s/graph", k)
		frags, err := fr.CreateFragments(fg)
		if err != nil {
			return 0, 0, err
		}

		for _, frag := range frags {
			frag.Meta.AddTags("csvUpload")
			err := frag.AddTo(p)
			if err != nil {
				return 0, 0, err
			}
			triplesProcessed = triplesProcessed + 1
		}
	}
	return triplesProcessed, rowsProcessed, nil
}

//Convert converts the CSV InputFile to an RDF ResourceMap
func (con *CSVConvertor) Convert() (*ResourceMap, int, error) {
	rm := &ResourceMap{make(map[string]*FragmentResource)}

	triples, rowsProcessed, err := con.CreateTriples()
	if err != nil {
		return rm, 0, err
	}

	if len(triples) == 0 {
		return rm, 0, fmt.Errorf("the list of triples cannot be empty")
	}

	for _, t := range triples {
		err := rm.AppendTriple(t, false)
		if err != nil {
			return rm, 0, err
		}
	}

	return rm, rowsProcessed, nil

}

// CreateTriples converts a csv file to a list of Triples
func (con *CSVConvertor) CreateTriples() ([]*r.Triple, int, error) {

	records, err := con.GetReader()
	if err != nil {
		return nil, 0, err
	}

	var header []string
	var headerMap map[int]r.Term
	var subjectColumnIdx int
	var thumbnailColumnIdx int

	triples := []*r.Triple{}

	for idx, row := range records {
		if idx == 0 {
			header = row
			headerMap = con.CreateHeader(header)
			subjectColumnIdx, err = con.GetSubjectColumn(
				header,
				con.SubjectColumn,
			)
			if err != nil {
				return nil, 0, err
			}
			if con.ThumbnailColumn != "" {
				thumbnailColumnIdx, err = con.GetSubjectColumn(
					header,
					con.ThumbnailColumn,
				)
				if err != nil {
					return nil, 0, err
				}
			}
			continue
		}

		s, sType := con.CreateSubjectResource(row[subjectColumnIdx])
		triples = append(triples, sType)

		for idx, column := range row {

			if con.ThumbnailColumn != "" && idx == thumbnailColumnIdx {
				thumbnail := r.NewTriple(
					s,
					r.NewResource(
						fmt.Sprintf("%s/thumbnail", con.PredicateURIBase),
					),
					r.NewLiteral(
						fmt.Sprintf("%s/%s", con.ThumbnailURIBase, column),
					),
				)
				triples = append(triples, thumbnail)
				continue
			}
			p := headerMap[idx]
			triples = append(triples, con.CreateTriple(s, p, column))
			if err != nil {
				return nil, 0, err
			}
		}

	}

	return triples, len(records), nil
}

// CreateHeader creates a map based on column id for the predicates
func (con *CSVConvertor) CreateHeader(row []string) map[int]r.Term {
	m := make(map[int]r.Term)
	for idx, column := range row {
		m[idx] = r.NewResource(
			fmt.Sprintf("%s/%s", strings.TrimSuffix(con.PredicateURIBase, "/"), strings.ToLower(column)),
		)
	}
	return m
}

// CreateTriple creates a rdf2go.Triple from the CSV column
func (con *CSVConvertor) CreateTriple(subject r.Term, predicate r.Term, column string) *r.Triple {
	return r.NewTriple(
		subject,
		predicate,
		r.NewLiteral(column),
	)
}

// CreateSubjectResource creates the Subject  URI and type triple for the subject column
func (con *CSVConvertor) CreateSubjectResource(subjectID string) (r.Term, *r.Triple) {
	cleanID := strings.Replace(subjectID, "-", "", 0)
	sep := "/"
	if strings.HasSuffix(con.SubjectURIBase, ":") {
		sep = ""
	}
	s := r.NewResource(fmt.Sprintf("%s%s%s", strings.TrimSuffix(con.SubjectURIBase, "/"), sep, cleanID))
	t := r.NewTriple(
		s,
		r.NewResource("http://www.w3.org/1999/02/22-rdf-syntax-ns#type"),
		r.NewResource(con.SubjectClass),
	)
	return s, t
}

// GetReader returns a nested array of strings
func (con *CSVConvertor) GetReader() ([][]string, error) {
	r := csv.NewReader(con.InputFile)
	r.Comma = []rune(con.Separator)[0]
	r.Comment = '#'

	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	return records, nil
}

// GetSubjectColumn returns the index of the subject column
func (con *CSVConvertor) GetSubjectColumn(headers []string, columnLabel string) (int, error) {
	for idx, column := range headers {
		if column == columnLabel {
			return idx, nil
		}
	}

	return 0, fmt.Errorf("subjectColumn %s not found in header", con.SubjectColumn)
}

// func Valid bool
// todo add curl example
