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

package api_test

import (
	"strings"

	"bitbucket.org/delving/rapid/hub3/api"

	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("V1", func() {

	Describe("Should be able to parse RDF", func() {

		Context("When given RDF as an io.Reader", func() {

			It("Should create a graph", func() {
				turtle, err := os.Open("test_data/test2.ttl")
				Expect(err).ToNot(HaveOccurred())
				Expect(turtle).ToNot(BeNil())
				g, err := api.NewGraphFromTurtle(turtle)
				Expect(err).ToNot(HaveOccurred())
				Expect(g).ToNot(BeNil())
				Expect(g.Len()).To(Equal(59))
				//triples := g.IterTriples()
				//fmt.Printf("%#v", triples)
				//for triple := range triples {
				//log.Println(triple.String())
				//}
			})

			It("Should throw an error when receiving invalid RDF", func() {
				badRDF := strings.NewReader("")
				g, err := api.NewGraphFromTurtle(badRDF)
				Expect(err).To(HaveOccurred())
				Expect(g.Len()).To(Equal(0))
			})
		})

	})

})
