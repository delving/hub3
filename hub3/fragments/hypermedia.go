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
	fmt "fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"text/template"

	"github.com/delving/hub3/config"
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
	if url.Scheme == "" {
		url.Scheme = "http"
	}
	if config.Config.HTTP.ProxyTLS || r.TLS != nil {
		url.Scheme = "https"
	}
	if r.Host == "" {
		r.Host = "localhost:3000"
	}

	basePage := fmt.Sprintf("%s://%s%s", url.Scheme, r.Host, url.EscapedPath())

	pageNumber := fr.GetPage()
	nextPage := pageNumber + int32(1)
	previousPage := pageNumber - int32(1)

	pagerURI := basePage + "?"
	var cleanPagerURI string
	if r.URL.RawQuery != "" {
		pagerURI += r.URL.RawQuery
		regString := "page=[0-9]*[&]{0,1}"
		re := regexp.MustCompile(regString)
		cleanPagerURI = re.ReplaceAllString(pagerURI, "")
	}

	pagerURI = strings.TrimSuffix(pagerURI, "?")
	if cleanPagerURI == "" {
		cleanPagerURI = pagerURI
	}

	cleanPagerURI = strings.Trim(strings.TrimSuffix(cleanPagerURI, "?"), "&")

	sep := "?"
	if strings.Contains(cleanPagerURI, "?") {
		sep = "&"
	}

	return &HyperMediaDataSet{
		PagerURI:     pagerURI,
		DataSetURI:   basePage,
		TotalItems:   totalHits,
		FirstPage:    fmt.Sprintf("%s%spage=1", cleanPagerURI, sep),
		NextPage:     fmt.Sprintf("%s%spage=%d", cleanPagerURI, sep, nextPage),
		PreviousPage: fmt.Sprintf("%s%spage=%d", cleanPagerURI, sep, previousPage),
		ItemsPerPage: int64(FRAGMENT_SIZE),
		CurrentPage:  pageNumber,
	}
}

// CreateControls creates an byte array of the RDF media controls
func (hmd HyperMediaDataSet) CreateControls() ([]byte, error) {
	var buf bytes.Buffer
	err := t.Execute(&buf, hmd)
	if err != nil {
		return nil, fmt.Errorf("unable to execute the hyperMedia template: %v", err)
	}
	return buf.Bytes(), nil
}

// HasNext returns if the dataset has a next page
func (hmd HyperMediaDataSet) HasNext() bool {
	return ((hmd.CurrentPage + 1) * FRAGMENT_SIZE) < int32(hmd.TotalItems)
}

// HasPrevious returns if the dataset has a previous page
func (hmd HyperMediaDataSet) HasPrevious() bool {
	return (hmd.CurrentPage - 1) > 0
}
