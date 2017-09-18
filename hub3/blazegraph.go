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

// blazegraph contains all functionality to query and update data in the underlying BlazeGraph triple-store.
package hub3

import (
	"fmt"

	. "bitbucket.org/delving/rapid/config"
	"github.com/parnurzeal/gorequest"
	"github.com/sirupsen/logrus"
)

var request = gorequest.New()

var blazegraphNamespaceProperties = `
com.bigdata.rdf.store.AbstractTripleStore.textIndex=true
com.bigdata.rdf.store.AbstractTripleStore.axiomsClass=com.bigdata.rdf.axioms.NoAxioms
com.bigdata.rdf.sail.isolatableIndices=false
com.bigdata.rdf.sail.truthMaintenance=false
com.bigdata.rdf.store.AbstractTripleStore.justify=false
com.bigdata.rdf.sail.namespace=%s
com.bigdata.namespace.%s.spo.com.bigdata.btree.BTree.branchingFactor=1024
com.bigdata.rdf.store.AbstractTripleStore.quads=false
com.bigdata.namespace.%s.lex.com.bigdata.btree.BTree.branchingFactor=400
com.bigdata.journal.Journal.groupCommit=false
com.bigdata.rdf.store.AbstractTripleStore.geoSpatial=true
com.bigdata.rdf.store.AbstractTripleStore.statementIdentifiers=false
`

// format namespaceBaseURI
func namespaceBaseURI() string {
	return fmt.Sprintf("%s/blazegraph/namespace", Config.RDF.SparqlHost)
}

// format the namespace URI for querying the blazegraph end-point
func blazegraphEndpoint(name string) string {
	if name == "" {
		name = Config.OrgID
	}
	return fmt.Sprintf("%s/%s", namespaceBaseURI(), name)
}

// namespaceExists checks if the namespace is available in blazegraph
func namespaceExist(name string) (bool, []error) {
	resp, _, errs := request.Post(blazegraphEndpoint(name)).
		//Set("Content-Type", "text/plain").
		End()
	if len(errs) != 0 {
		log.Error(errs)
		return false, errs
	}
	exists := resp.StatusCode != 404
	return exists, errs

}

// createNameSpaceProperties creates the blazegraph properties
func createNameSpaceProperties(name string) string {
	if name == "" {
		name = Config.OrgID
	}
	return fmt.Sprintf(
		blazegraphNamespaceProperties,
		name,
		name,
		name,
	)
}

// createNameSpace creates a namespace in blazegraph
func createNameSpace(name string) (bool, []error) {
	properties := createNameSpaceProperties(name)
	endPoint := namespaceBaseURI()
	log.WithFields(logrus.Fields{"endpoint": endPoint}).Debugf("Blazegraph properties: %s", properties)
	resp, _, errs := request.Post(endPoint).
		Set("Content-Type", "text/plain").
		Send(properties).
		End()
	if len(errs) != 0 {
		log.Error(errs)
		return false, errs
	}
	created := resp.StatusCode == 201
	return created, errs
}

// deleteNamespace deletes a namespace in blazegraph
func deleteNameSpace(name string) (bool, []error) {
	resp, _, errs := request.Delete(blazegraphEndpoint(name)).End()
	if len(errs) != 0 {
		log.Error(errs)
		return false, errs
	}
	deleted := resp.StatusCode == 200
	return deleted, errs
}
