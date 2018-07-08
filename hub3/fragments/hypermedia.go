package fragments

import (
	"bytes"
	fmt "fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"text/template"
)

const hyperTmpl = `<{{.DataSetURI}}> <http://rdfs.org/ns/void#subset> <{{.PagerURI}}> .
<{{.DataSetURI}}> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://rdfs.org/ns/void#Dataset> .
<{{.DataSetURI}}> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/ns/hydra/core#Collection> .
<{{.DataSetURI}}> <http://www.w3.org/ns/hydra/core#itemsPerPage> "100"^^<http://www.w3.org/2001/XMLSchema#long> .
<{{.DataSetURI}}> <http://www.w3.org/ns/hydra/core#search> <{{.PagerURI}}#triplePattern> .
<{{.PagerURI}}> <http://rdfs.org/ns/void#triples> "{{.TotalItems}}"^^<http://www.w3.org/2001/XMLSchema#integer> .
<{{.PagerURI}}> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/ns/hydra/core#Collection> .
<{{.PagerURI}}> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.w3.org/ns/hydra/core#PagedCollection> .
<{{.PagerURI}}> <http://www.w3.org/ns/hydra/core#firstPage> <{{.FirstPage}}> .
<{{.PagerURI}}> <http://www.w3.org/ns/hydra/core#itemsPerPage> "{{.ItemsPerPage}}"^^<http://www.w3.org/2001/XMLSchema#integer> .
{{ if .HasNext -}}<{{.PagerURI}}> <http://www.w3.org/ns/hydra/core#nextPage> <{{.NextPage}}> .  {{- end }}
{{ if .HasPrevious -}}<{{.PagerURI}}> <http://www.w3.org/ns/hydra/core#previousPage> <{{.PreviousPage}}> . {{- end }}
<{{.PagerURI}}> <http://www.w3.org/ns/hydra/core#totalItems> "{{.TotalItems}}"^^<http://www.w3.org/2001/XMLSchema#integer> .
<{{.PagerURI}}#subject> <http://www.w3.org/ns/hydra/core#property> <http://www.w3.org/1999/02/22-rdf-syntax-ns#subject> .
<{{.PagerURI}}#subject> <http://www.w3.org/ns/hydra/core#variable> "subject" .
<{{.PagerURI}}#predicate> <http://www.w3.org/ns/hydra/core#property> <http://www.w3.org/1999/02/22-rdf-syntax-ns#predicate> .
<{{.PagerURI}}#predicate> <http://www.w3.org/ns/hydra/core#variable> "predicate" .
<{{.PagerURI}}#object> <http://www.w3.org/ns/hydra/core#property> <http://www.w3.org/1999/02/22-rdf-syntax-ns#object> .
<{{.PagerURI}}#object> <http://www.w3.org/ns/hydra/core#variable> "object" .
<{{.PagerURI}}#triplePattern> <http://www.w3.org/ns/hydra/core#mapping> <{{.PagerURI}}#subject> .
<{{.PagerURI}}#triplePattern> <http://www.w3.org/ns/hydra/core#mapping> <{{.PagerURI}}#predicate> .
<{{.PagerURI}}#triplePattern> <http://www.w3.org/ns/hydra/core#mapping> <{{.PagerURI}}#object> .
<{{.PagerURI}}#triplePattern> <http://www.w3.org/ns/hydra/core#template> "{{.DataSetURI}}{?subject,predicate,object}" .
`

var t *template.Template

func init() {
	var err error
	t, err = template.New("hypermedia").Parse(hyperTmpl)
	if err != nil {
		log.Fatal(err)
	}

}

// HyperMediaDataSet holds all the configuration information to generate the
// HyperMediaControls RDF
type HyperMediaDataSet struct {
	DataSetURI   string
	PagerURI     string
	TotalItems   int64
	ItemsPerPage int64
	FirstPage    string
	PreviousPage string
	NextPage     string
	CurrentPage  int32
}

// NewHyperMediaDataSet creates the basis to generate triple-pattern-fragment controls
func NewHyperMediaDataSet(r *http.Request, totalHits int64, fr *FragmentRequest) *HyperMediaDataSet {
	url := r.URL
	currentPage := url.Query().Get("page")
	if url.Scheme == "" {
		url.Scheme = "http"
		if r.TLS != nil {
			url.Scheme = "https"
		}
	}
	if url.Host == "" {
		url.Host = "localhost:3000"
	}

	regString := fmt.Sprintf("[?|&]page=%s", currentPage)
	var re = regexp.MustCompile(regString)
	basePage := re.ReplaceAllString(url.String(), "")
	pageNumber := fr.GetPage()
	nextPage := pageNumber + int32(1)
	previousPage := pageNumber - int32(1)
	sep := "&"
	if !strings.Contains(basePage, "?") {
		sep = "?"
	}

	log.Printf("current %d, next %d, previous %d", pageNumber, nextPage, previousPage)

	return &HyperMediaDataSet{
		DataSetURI:   fmt.Sprintf("%s://%s%s", url.Scheme, url.Host, url.EscapedPath()),
		PagerURI:     url.String(),
		TotalItems:   totalHits,
		FirstPage:    fmt.Sprintf("%s%spage=1", basePage, sep),
		NextPage:     fmt.Sprintf("%s%spage=%d", basePage, sep, nextPage),
		PreviousPage: fmt.Sprintf("%s%spage=%d", basePage, sep, previousPage),
		ItemsPerPage: int64(FRAGMENT_SIZE),
		CurrentPage:  pageNumber,
	}
}

// CreateControls creates an byte array of the RDF media controls
func (hmd HyperMediaDataSet) CreateControls() ([]byte, error) {
	var buf bytes.Buffer
	err := t.Execute(&buf, hmd)
	if err != nil {
		return nil, fmt.Errorf("Unable to execute the hyperMedia template: %v", err)
	}
	return buf.Bytes(), nil
}

// HasNext returns if the dataset has a next page
func (hmd HyperMediaDataSet) HasNext() bool {
	fmt.Println(hmd.CurrentPage, FRAGMENT_SIZE, hmd.TotalItems)
	return ((hmd.CurrentPage + 1) * FRAGMENT_SIZE) < int32(hmd.TotalItems)
}

// HasPrevious returns if the dataset has a previous page
func (hmd HyperMediaDataSet) HasPrevious() bool {
	return (hmd.CurrentPage - 1) > 0
}
