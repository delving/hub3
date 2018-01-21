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
	"net/url"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var nt = `<http://rapid.org/123> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://www.europeana.eu/schemas/edm/Place> .
<http://rapid.org/document/123> <http://www.europeana.eu/schemas/edm/type> "IMAGE" .`

var graphName = "http://rapid.org/123/graph"

var _ = Describe("Rdf", func() {

	Describe("Converting to nquads", func() {

		Context("from ntriples", func() {

			It("Should replace end markers with graph uri", func() {
				Expect(len(strings.Split(nt, "\n"))).To(Equal(2))
				Expect(nt).ToNot(ContainSubstring("/graph>"))
				Expect(nt).To(HaveSuffix("."))
				graphURI, _ := url.Parse(graphName)
				nquads := Ntriples2Nquads(nt, graphURI)
				Expect(nquads).ToNot(BeNil())
				Expect(nquads).To(ContainSubstring("/graph>"))
			})
		})
	})

})
