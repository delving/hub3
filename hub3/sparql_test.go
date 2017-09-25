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

		_, errs := deleteNameSpace(Config.OrgID)
		if errs != nil {
			//log.Error(errs)
			fmt.Println(errs)
		}
		_, errs = createNameSpace(Config.OrgID)
		if errs != nil {
			//log.Error(errs)
			fmt.Println(errs)
		}
		Context("Ask", func() {

			It("should return a boolean", func() {
				ask := AskSPARQL("ASK {<urn:123> ?p ?o}")
				Expect(ask).To(Equal(false))
			})
		})

		Context("Describe", func() {
			It("should return sparql bindings", func() {
				describe := DescribeSPARQL("urn:123")
				log.Debug(describe)
				//fmt.Println(describe)
				Expect(describe).ToNot(BeEmpty())
			})
		})
	})
})
