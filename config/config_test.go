package config_test

import (
	"fmt"

	. "bitbucket.org/delving/rapid/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {

	Describe("after initialisation", func() {

		Context("when calling initConfig", func() {

			It("should be initialised with defaults", func() {
				InitConfig()
				Expect(Config.HTTP.Port).To(Equal(3001))
				Expect(Config.OrgID).ToNot(BeEmpty())
			})

		})

		//Context("without calling initConfig", func() {
		//It("should not be initialised", func() {
		//Expect(Config.OrgID).To(BeEmpty())
		//Expect(Config.HTTP.Port).To(Equal(0))
		//})
		//})

		Context("when setting a config value", func() {

			It("should be available in the global scope", func() {
				Expect(Config.Logging.SentryDSN).To(BeEmpty())
				Config.Logging.SentryDSN = "test"
				Expect(Config.Logging.SentryDSN).ToNot(BeEmpty())
			})
		})

	})

	Describe("building the SPARQL endpoint", func() {

		Context("when constructed", func() {
			InitConfig()
			endpoint := Config.GetSparqlEndpoint("")

			It("should not be empty", func() {
				Expect(endpoint).ToNot(BeEmpty())
			})

			It("should use the sparqlhost setting from the configuration", func() {
				Expect(Config.RDF.SparqlHost).ToNot(BeEmpty())
				Expect(endpoint).To(HavePrefix(Config.RDF.SparqlHost))
			})

			It("should use the sparql path from the configuration", func() {
				Expect(Config.RDF.SparqlPath).To(ContainSubstring("%s"))
				Expect(endpoint).To(ContainSubstring("/rapid/sparql"))
			})

			It("should should inject the orgId from the configuration when dbName is empty", func() {
				Expect(endpoint).To(ContainSubstring(Config.OrgID))
			})

		})

		Context("when a dbname is specified", func() {

			endpoint := Config.GetSparqlEndpoint("rapid2")
			It("should use dbName to inject into the sparql path", func() {
				orgId := Config.OrgID
				Expect(endpoint).To(ContainSubstring("/rapid2/"))
				Expect(endpoint).ToNot(ContainSubstring(fmt.Sprintf("/%s/", orgId)))
			})
		})
	})
})
