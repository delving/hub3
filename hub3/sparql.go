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
	"strconv"
	"time"

	. "bitbucket.org/delving/rapid/config"
	"bitbucket.org/delving/rapid/hub3/models"
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

# tag: countGraphPerSpec
SELECT (count(?subject) as ?count)
WHERE {
  ?subject <http://schemas.delving.eu/nave/terms/datasetSpec> "{{.Spec}}"
}
LIMIT 1

# tag: countRevisionsBySpec
SELECT ?revision (COUNT(?revision) as ?rCount)
WHERE
{
  ?subject <http://schemas.delving.eu/nave/terms/datasetSpec> "{{.Spec}}";
		<http://schemas.delving.eu/nave/terms/specRevision> ?revision .
}
GROUP BY ?revision

# tag: deleteAllGraphsBySpec
DELETE {
	GRAPH ?g {
	?s ?p ?o .
	}
}
WHERE {
	GRAPH ?g {
	?subject <http://schemas.delving.eu/nave/terms/datasetSpec> "{.Spec}".
	}
	GRAPH ?g {
	?s ?p ?o .
	}
};

# tag: deleteOrphanGraphsBySpec
DELETE {
	GRAPH ?g {
	?s ?p ?o .
	}
}
WHERE {
	GRAPH ?g {
	?subject <http://schemas.delving.eu/nave/terms/datasetSpec> "{.Spec}";
		<http://schemas.delving.eu/nave/terms/specRevision> {{.RevisionNumber}} .
	}
	GRAPH ?g {
	?s ?p ?o .
	}
};
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

func DeleteOrphansBySpec(spec string, revision int) (bool, error) {
	query, err := queryBank.Prepare("deleteOrphanGraphsBySpec", struct {
		Spec     string
		Revision int
	}{spec, revision})
	if err != nil {
		logger.WithField("spec", spec).Errorf("Unable to build deleteOrphanGraphsBySpec query: %s", err)
		return false, err
	}
	fmt.Println(query)

	return true, nil
}

//CountRevisionsBySpec counts each revision available in the spec
func CountRevisionsBySpec(spec string) ([]models.DataSetRevisions, error) {
	query, err := queryBank.Prepare("countRevisionsBySpec", struct{ Spec string }{spec})
	revisions := []models.DataSetRevisions{}
	if err != nil {
		logger.WithField("spec", spec).Errorf("Unable to build countRevisionsBySpec query: %s", err)
		return revisions, err
	}
	//fmt.Printf("%#v", query)
	res, err := SparqlRepo.Query(query)
	if err != nil {
		logger.WithField("query", query).Errorf("Unable query endpoint: %s", err)
		return revisions, err
	}
	//fmt.Printf("%#v", res)
	for _, v := range res.Solutions() {
		revision, err := strconv.Atoi(v["revision"].String())
		if err != nil {
			return revisions, fmt.Errorf("Unable to convert %#v to integer.", v["revision"])
		}
		revisionCount, err := strconv.Atoi(v["rCount"].String())
		if err != nil {
			return revisions, fmt.Errorf("Unable to convert %#v to integer.", v["rCount"])
		}
		revisions = append(revisions, models.DataSetRevisions{
			Number:      revision,
			RecordCount: revisionCount,
		})
	}
	return revisions, nil
}

// CountGraphsBySpec counts all the named graphs for a spec
func CountGraphsBySpec(spec string) (int, error) {
	query, err := queryBank.Prepare("countGraphPerSpec", struct{ Spec string }{spec})
	if err != nil {
		logger.WithField("spec", spec).Errorf("Unable to build CountGraphsBySpec query: %s", err)
		return 0, err
	}
	res, err := SparqlRepo.Query(query)
	if err != nil {
		logger.WithField("query", query).Errorf("Unable query endpoint: %s", err)
		return 0, err
	}
	countStr, ok := res.Bindings()["count"]
	if !ok {
		logger.WithField("bindings", res.Bindings()).Errorf("Unable to get count from results")
		return 0, fmt.Errorf("Unable to get count from result bindings: %#v", res.Bindings())
	}
	var count int
	count, err = strconv.Atoi(countStr[0].String())
	if err != nil {
		return 0, fmt.Errorf("Unable to convert %s to integer.", countStr)
	}
	return count, err
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
