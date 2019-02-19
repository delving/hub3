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

package models

import (
	"bytes"
	"fmt"
	"strconv"
	"time"

	"github.com/delving/rapid-saas/config"
	"github.com/delving/rapid-saas/hub3/fragments"
	"github.com/knakk/rdf"
	"github.com/knakk/sparql"
	"github.com/sirupsen/logrus"
)

const queries = `
# SPARQL queries that are loaded as a QueryBank

# ask returns a boolean
# tag: ask_subject
ASK { <{{ .URI }}> ?p ?o }

# tag: ask_predicate
ASK { ?s <{{ .URI }}> ?o }

# tag: ask_object
ASK { ?s <{{ .URI }}> ?o }

# tag: ask_query
ASK { {{ .Query }} }

# The DESCRIBE form returns a single result RDF graph containing RDF data about resources.
# tag: describe
DESCRIBE <{{.URI}}>

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
	?subject <http://schemas.delving.eu/nave/terms/datasetSpec> "{{.Spec}}".
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
	?subject <http://schemas.delving.eu/nave/terms/datasetSpec> "{{.Spec}}";
		<http://schemas.delving.eu/nave/terms/specRevision> ?revision .
		FILTER (?revision != {{.RevisionNumber}}).
	}
	GRAPH ?g {
	?s ?p ?o .
	}
};
`

var queryBank sparql.Bank

// SparqlQueryURL is the fully qualified URI to the SPARQL endpoint
var SparqlQueryURL string

// SparqlUpdateURL is the fully qualified URI to the SPARQL Update endpoint
var SparqlUpdateURL string

// SparqlRepo is the repository used for querying
var SparqlRepo *sparql.Repo

// SparqlUpdateRepo is the repository used for updating the TripleStore
var SparqlUpdateRepo *sparql.Repo

func init() {
	SparqlQueryURL = config.Config.GetSparqlEndpoint("")
	SparqlUpdateURL = config.Config.GetSparqlUpdateEndpoint("")
	f := bytes.NewBufferString(queries)
	queryBank = sparql.LoadBank(f)
	SparqlRepo = buildRepo(SparqlQueryURL)
	SparqlUpdateRepo = buildRepo(SparqlUpdateURL)
}

// buildRepo builds the query repository
func buildRepo(endPoint string) *sparql.Repo {
	if endPoint == "" {
		endPoint = config.Config.GetSparqlEndpoint("")
	}
	repo, err := sparql.NewRepo(endPoint,
		sparql.Timeout(time.Millisecond*1500),
	)
	if err != nil {
		logger.Fatal(err)
	}
	return repo
}

// DeleteAllGraphsBySpec issues an SPARQL Update query to delete all graphs for a DataSet from the triple store
func DeleteAllGraphsBySpec(spec string) (bool, error) {
	query, err := queryBank.Prepare("deleteAllGraphsBySpec", struct{ Spec string }{spec})
	if err != nil {
		logger.WithField("spec", spec).Errorf("Unable to build deleteAllGraphsBySpec query: %s", err)
		return false, err
	}
	logger.Info(query)
	errs := fragments.UpdateViaSparql(query)
	if errs != nil {
		logger.WithField("query", query).Errorf("Unable query endpoint: %s", errs)
		return false, errs[0]
	}
	return true, nil
}

// DeleteGraphsOrphansBySpec issues an SPARQL Update query to delete all orphaned graphs
// for a DataSet from the triple store.
func DeleteGraphsOrphansBySpec(spec string, revision int) (bool, error) {
	query, err := queryBank.Prepare("deleteOrphanGraphsBySpec", struct {
		Spec           string
		RevisionNumber int
	}{spec, revision})
	if err != nil {
		logger.WithField("spec", spec).Errorf("Unable to build deleteOrphanGraphsBySpec query: %s", err)
		return false, err
	}
	logger.Info(query)
	errs := fragments.UpdateViaSparql(query)
	if errs != nil {
		logger.WithField("query", query).Errorf("Unable query endpoint: %s", errs)
		return false, errs[0]
	}
	return true, nil
}

//CountRevisionsBySpec counts each revision available in the spec
func CountRevisionsBySpec(spec string) ([]DataSetRevisions, error) {
	query, err := queryBank.Prepare("countRevisionsBySpec", struct{ Spec string }{spec})
	revisions := []DataSetRevisions{}
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
	//fmt.Printf("%#v", res.Solutions())
	for _, v := range res.Solutions() {
		revisionTerm, ok := v["revision"]
		if !ok {
			logger.Infof("No revisions found for spec %s", spec)
			return revisions, nil
		}
		revision, err := strconv.Atoi(revisionTerm.String())
		if err != nil {
			return revisions, fmt.Errorf("unable to convert %#v to integer", v["revision"])
		}
		revisionCount, err := strconv.Atoi(v["rCount"].String())
		if err != nil {
			return revisions, fmt.Errorf("unable to convert %#v to integer", v["rCount"])
		}
		revisions = append(revisions, DataSetRevisions{
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
		return 0, fmt.Errorf("unable to convert %s to integer", countStr)
	}
	return count, err
}

// PrepareAsk takes an a string and returns a valid SPARQL ASK query
func PrepareAsk(uri string) (string, error) {
	q, err := queryBank.Prepare("ask_subject", struct{ URI string }{uri})
	if err != nil {
		fields := logrus.Fields{"err": err, "uri": uri}
		logger.WithFields(fields).Error("Unable to build ask query")
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

// DescribeSPARQL creates a describe query for a given URI.
func DescribeSPARQL(uri string) (map[string][]rdf.Term, error) {
	query, err := queryBank.Prepare("describe", struct{ URI string }{uri})
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
