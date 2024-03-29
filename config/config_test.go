// Copyright 2017 Delving B.V.
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
package config_test

import (
	"fmt"

	. "github.com/delving/hub3/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	Describe("after initialisation", func() {
		Context("when calling initConfig", func() {
			It("should be initialised with defaults", func() {
				InitConfig()
				Expect(Config.HTTP.Port).To(Equal(3001))
			})
		})

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
			endpoint := Config.GetSparqlEndpoint("test", "")

			It("should not be empty", func() {
				Expect(endpoint).ToNot(BeEmpty())
			})

			It("should use the sparqlhost setting from the configuration", func() {
				Expect(Config.RDF.SparqlHost).ToNot(BeEmpty())
				Expect(endpoint).To(HavePrefix(Config.RDF.SparqlHost))
			})

			It("should use the sparql path from the configuration", func() {
				Expect(Config.RDF.SparqlPath).To(ContainSubstring("%s"))
				Expect(endpoint).To(ContainSubstring("/test/sparql"))
			})
		})

		Context("when a dbname is specified", func() {
			endpoint := Config.GetSparqlEndpoint("test", "hub32")
			It("should use dbName to inject into the sparql path", func() {
				orgID := "test"
				Expect(endpoint).To(ContainSubstring("/hub32/"))
				Expect(endpoint).ToNot(ContainSubstring(fmt.Sprintf("/%s/", orgID)))
			})
		})
	})
})
