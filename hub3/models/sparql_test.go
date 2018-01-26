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
	"fmt"

	"github.com/labstack/gommon/log"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "bitbucket.org/delving/rapid/config"
)

var _ = Describe("Sparql", func() {

	InitConfig()

	Describe("Initialisation", func() {

		Context("when called", func() {

			It("should have initialised the sparql query bank", func() {
				Expect(queryBank).ToNot(BeNil())
			})

			It("should have initialised the SparqlEndpoint", func() {
				Expect(SparqlQueryURL).ToNot(BeNil())
			})
		})
	})

	Describe("Building SPARQL queries", func() {

		Context("when ASK", func() {
			It("should include the uri in the query", func() {
				uri := "urn:123"
				askQuery, err := PrepareAsk(uri)
				Expect(err).To(BeNil())
				Expect(askQuery).To(HavePrefix("ASK "))
				Expect(askQuery).To(ContainSubstring(uri))
			})
		})

	})

	Describe("Executing a SPARQL query", func() {

		Context("Ask", func() {

			It("should return a boolean", func() {
				ask, err := AskSPARQL("ASK {<urn:123> ?p ?o}")
				Expect(err).To(BeNil())
				Expect(ask).To(Equal(false))
			})

			It("should return a result", func() {

				res, err := SparqlRepo.Query("ASK { ?s ?p ?o } LIMIT 1")
				if err != nil {
					log.Fatal(err)
				}
				Expect(err).To(BeNil())
				Expect(res).ToNot(BeNil())
				fmt.Printf("results: %#v", res.Results)
				fmt.Println(res.Results.Bindings)
				for i, m := range res.Results.Bindings {
					s := m["s"]
					fmt.Printf("%d %s %s\n", i, s.Type, s.Value)
				}
				fmt.Println(res.Results.Bindings)
			})
		})

		Context("Describe", func() {
			It("should return sparql bindings", func() {
				Skip("Not supported for now")
				describe, err := DescribeSPARQL("http://data.collectienederland.nl/resource/document/coda-rce/SZ94373")
				Expect(err).To(BeNil())
				log.Debug(describe)
				//fmt.Println(describe)
				Expect(describe).ToNot(BeEmpty())
			})
		})
	})
})
