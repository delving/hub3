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

package sparql

import (
	"bufio"
	"bytes"
	"context"
	fmt "fmt"
	"io"
	"strings"
	"text/template"

	"github.com/delving/hub3/ikuzo/rdf"
	"github.com/delving/hub3/ikuzo/rdf/formats/ntriples"
)

const sparqlUpdateTemplate = `
	GRAPH <{{.NamedGraphURI}}> {
		<{{.NamedGraphURI}}> <http://schemas.delving.eu/nave/terms/datasetSpec> "{{.Spec}}" .
		<{{.NamedGraphURI}}> <http://schemas.delving.eu/nave/terms/specRevision> "{{.SpecRevision}}"^^<http://www.w3.org/2001/XMLSchema#integer> .
		{{ .Triples }}
	}
`

type BulkMetrics struct {
	Graphs  int
	Triples int
}

// BulkRequest can send multiple namedgraphs to an repo with a single request.
// Do(ctx) sends the actual request.
type BulkRequest struct {
	repo      *Repo
	datasetID string
	revision  int
	tmpl      *template.Template
	updates   bytes.Buffer
	drops     bytes.Buffer
	m         *BulkMetrics
}

// NewBulkRequest retutrns a new BulkRequest
func NewBulkRequest(repo *Repo, datasetID string, revision int) *BulkRequest {
	return &BulkRequest{
		repo:      repo,
		datasetID: datasetID,
		revision:  revision,
		tmpl:      repo.updateTmpl,
	}
}

// AddString adds triples to a namedgraph.
// A namedgraph cannot be added multiple times. All triples will be merged in
// the last entry.
func (br *BulkRequest) AddString(namedGraphUri, triples string) error {
	br.m.Graphs++
	update := sparqlUpdate{
		Triples:       triples,
		NamedGraphURI: namedGraphUri,
		Spec:          br.datasetID,
		SpecRevision:  br.revision,
	}

	_, err := br.drops.WriteString(fmt.Sprintf("DROP GRAPH <%s>;\n", update.NamedGraphURI))
	if err != nil {
		return err
	}

	return br.tmpl.Execute(&br.updates, update)
}

// Add adds all the triples from the rdf.Graph
func (br *BulkRequest) Add(namedGraphUri string, g *rdf.Graph) error {
	var buf bytes.Buffer
	if err := ntriples.Serialize(g, &buf); err != nil {
		return err
	}

	return br.AddString(namedGraphUri, buf.String())
}

// UpdateQuery returns the SPARQL update query that is send with the Do function
func (br *BulkRequest) UpdateQuery() string {
	return fmt.Sprintf(
		"%s \n INSERT DATA {%s}\n",
		br.drops.String(),
		br.updates.String(),
	)
}

// Do executes the SPARQL update for all the added namedgraphs.
func (br *BulkRequest) Do(ctx context.Context) error {
	if err := br.repo.Update(br.UpdateQuery()); err != nil {
		return fmt.Errorf("unable to submit bulk request")
	}

	return nil
}

// sparqlUpdate contains the elements to perform a SPARQL update query
type sparqlUpdate struct {
	Triples       string `json:"triples"`
	NamedGraphURI string `json:"graphUri"`
	Spec          string `json:"datasetSpec"`
	SpecRevision  int    `json:"specRevision"`
}

// TripleCount counts the number of Ntriples in a string
func (su sparqlUpdate) TripleCount() (int, error) {
	r := strings.NewReader(su.Triples)
	return lineCounter(r)
}

func lineCounter(r io.Reader) (int, error) {
	scanner := bufio.NewScanner(r)
	lineCount := 0
	for scanner.Scan() {
		lineCount++
	}
	return lineCount, nil
}
