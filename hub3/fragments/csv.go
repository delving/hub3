package fragments

import (
	"bytes"
	"encoding/csv"
	fmt "fmt"
	"io"
	"strings"

	c "github.com/delving/hub3/config"
	r "github.com/kiivihal/rdf2go"
	elastic "github.com/olivere/elastic/v7"
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
	ObjectIntegerColumns  []string  `json:"objectIntegerColumns"`
	ThumbnailURIBase      string    `json:"thumbnailURIBase"`
	ThumbnailColumn       string    `json:"thumbnailColumn"`
	ManifestURIBase       string    `json:"manifestURIBase"`
	ManifestColumn        string    `json:"manifestColumn"`
	ManifestLocale        string    `json:"manifestLocale"`
	DefaultSpec           string    `json:"defaultSpec"`
	InputFile             io.Reader `json:"-"`
	RowsProcessed         int       `json:"rowsProcessed"`
	TriplesCreated        int       `json:"triplesCreated"`
	integerMap            map[int]bool
	resourceMap           map[int]bool
	headerMap             map[int]r.Term
	storeRDF              bool
}

// NewCSVConvertor creates a CSV convertor from an net/http Form
func NewCSVConvertor() *CSVConvertor {
	return &CSVConvertor{
		headerMap:   make(map[int]r.Term),
		integerMap:  make(map[int]bool),
		resourceMap: make(map[int]bool),
		storeRDF:    c.Config.RDF.RDFStoreEnabled,
	}
}

// HeaderMap gives access to a read-only version of the header map
func (con CSVConvertor) HeaderMap() map[int]r.Term {
	return con.headerMap
}

// IndexFragments stores the fragments generated from the CSV into ElasticSearch
func (con *CSVConvertor) IndexFragments(p *elastic.BulkProcessor, revision int) (int, int, error) {

	fg := NewFragmentGraph()
	fg.Meta = &Header{
		OrgID:    c.Config.OrgID,
		Revision: int32(revision),
		DocType:  "csvUpload",
		Spec:     con.DefaultSpec,
		Tags:     []string{"csvUpload"},
		Modified: NowInMillis(),
	}

	rm, rowsProcessed, err := con.Convert()

	if err != nil {
		return 0, 0, err
	}

	sparqlUpdates := []SparqlUpdate{}
	var triples bytes.Buffer

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
			if con.storeRDF {
				_, err = triples.WriteString(frag.Triple + "\n")
				if err != nil {
					return 0, 0, err
				}
			}
			triplesProcessed = triplesProcessed + 1
		}
		if con.storeRDF {
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
				_, errs := RDFBulkInsert(sparqlUpdates)
				if len(errs) != 0 {
					return 0, 0, errs[0]
				}
				sparqlUpdates = []SparqlUpdate{}
			}
		}
	}

	if len(sparqlUpdates) != 0 {
		_, errs := RDFBulkInsert(sparqlUpdates)
		if len(errs) != 0 {
			return 0, 0, errs[0]
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
	var subjectColumnIdx int
	var thumbnailColumnIdx int
	var manifestColumnIdx int

	triples := []*r.Triple{}

	for idx, row := range records {
		if idx == 0 {
			header = row
			con.CreateHeader(header)
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
			if con.ManifestColumn != "" {
				manifestColumnIdx, err = con.GetSubjectColumn(
					header,
					con.ManifestColumn,
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
			if len(strings.TrimSpace(column)) == 0 {
				continue
			}

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
				manifest := r.NewTriple(
					s,
					r.NewResource(
						fmt.Sprintf("%s/manifest", con.PredicateURIBase),
					),
					r.NewLiteral(
						fmt.Sprintf(
							"%s/%s/hubID/",
							strings.TrimSuffix(con.ManifestURIBase, "s"),
							strings.TrimSpace(column),
						),
					),
				)
				triples = append(triples, thumbnail, manifest)
			}
			if con.ManifestColumn != "" && idx == manifestColumnIdx {
				manifest := r.NewTriple(
					s,
					r.NewResource(
						fmt.Sprintf("%s/manifests", con.PredicateURIBase),
					),
					r.NewLiteral(
						fmt.Sprintf(
							"%s/%s/hubID/%s",
							con.ManifestURIBase,
							strings.TrimSpace(column),
							con.ManifestLocale,
						),
					),
				)
				triples = append(triples, manifest)
			}
			t := con.CreateTriple(s, idx, column)
			if t != nil {
				triples = append(triples, t)
			}
			if err != nil {
				return nil, 0, err
			}
		}

	}

	return triples, len(records), nil
}

// CreateHeader creates a map based on column id for the predicates
func (con *CSVConvertor) CreateHeader(row []string) {
	for idx, column := range row {
		con.headerMap[idx] = r.NewResource(
			fmt.Sprintf("%s/%s", strings.TrimSuffix(con.PredicateURIBase, "/"), strings.ToLower(column)),
		)
		if stringInSlice(column, con.ObjectResourceColumns) {
			con.resourceMap[idx] = true
		}
		if stringInSlice(column, con.ObjectIntegerColumns) {
			con.integerMap[idx] = true
		}
	}
	return
}

// CreateTriple creates a rdf2go.Triple from the CSV column
func (con *CSVConvertor) CreateTriple(subject r.Term, idx int, column string) *r.Triple {
	c := strings.TrimSpace(column)
	predicate := con.headerMap[idx]
	if len(c) == 0 {
		return nil
	}
	if con.integerMap[idx] {
		return r.NewTriple(
			subject,
			predicate,
			r.NewLiteralWithDatatype(c, r.NewResource("http://www.w3.org/2001/XMLSchema#integer")),
		)
	}
	if con.ObjectURIFormat != "" && con.resourceMap[idx] {
		return r.NewTriple(
			subject,
			predicate,
			r.NewResource(fmt.Sprintf("%s%s", con.ObjectURIFormat, column)),
		)
	}
	return r.NewTriple(
		subject,
		predicate,
		r.NewLiteral(c),
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
	if con.Separator == "" {
		return nil, fmt.Errorf("Separator cannot be empty")
	}
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
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
