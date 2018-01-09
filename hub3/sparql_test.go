package hub3

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

	Describe("Excecuting a SPARQL query", func() {

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
