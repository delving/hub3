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

package hub3

import (
	"bytes"
	"fmt"
	"time"

	. "bitbucket.org/delving/rapid/config"
	"github.com/knakk/rdf"
	"github.com/knakk/sparql"
	"github.com/sirupsen/logrus"
)

const queries = `
# SPARQL queries that are loaded as a QueryBank

# ask returns a boolean
# tag: ask_subject
ASK { <{{ .Uri }}> ?p ?o }

# tag: ask_predicate
ASK { ?s <{{ .Uri }}> ?o }

# tag: ask_object
ASK { ?s <{{ .Uri }}> ?o }

# tag: ask_query
ASK { {{ .Query }} }

# The DESCRIBE form returns a single result RDF graph containing RDF data about resources.
# tag: describe
DESCRIBE <{{.Uri}}>
`

var queryBank sparql.Bank

// SparqlQueryURL is the fully qualified URI to the SPARQL endpoint
var SparqlQueryURL string

// SparqlRepo is the repository used for querying
var SparqlRepo *sparql.Repo

func init() {
	SparqlQueryURL = Config.GetSparqlEndpoint("")
	f := bytes.NewBufferString(queries)
	queryBank = sparql.LoadBank(f)
	SparqlRepo = buildRepo(SparqlQueryURL)
}

// buildRepo builds the query repository
func buildRepo(endPoint string) *sparql.Repo {
	if endPoint == "" {
		endPoint = Config.GetSparqlEndpoint("")
	}
	repo, err := sparql.NewRepo(endPoint,
		sparql.Timeout(time.Millisecond*1500),
	)
	if err != nil {
		logger.Fatal(err)
	}
	return repo
}

// PrepareAsk takes an a string and returns a valid SPARQL ASK query
func PrepareAsk(uri string) (string, error) {
	q, err := queryBank.Prepare("ask_subject", struct{ Uri string }{uri})

	if err != nil {
		logger.WithFields(logrus.Fields{"err": err, "uri": uri}).Error("Unable to build ask query")
		return "", err
	}
	return q, err
}

// AskSPARQL performs a SPARQL ASK query
func AskSPARQL(query string) (bool, error) {
	res, err := SparqlRepo.Query(query)
	if err != nil {
		logger.WithField("sparql", "ask").Fatal(err)
		return false, err
	}
	bindings := res.Results.Bindings
	logger.Debug(bindings)
	fmt.Println(bindings)
	return false, nil
}

func DescribeSPARQL(uri string) (map[string][]rdf.Term, error) {
	query, err := queryBank.Prepare("describe", struct{ Uri string }{uri})
	if err != nil {
		logger.WithField("uri", uri).Errorf("Unable to build describe query.")
		return nil, err
	}
	res, err := SparqlRepo.Query(query)
	if err != nil {
		logger.WithField("query", query).Errorf("Unable query endpoint: %s", err)
		return nil, err
	}
	return res.Bindings(), nil
}
