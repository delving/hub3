// Copyright 2017 Delving B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fragments_test

import (
	"os"

	"github.com/kiivihal/rdf2go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/delving/hub3/hub3/fragments"
)

var _ = Describe("CSV", func() {
	Describe("converting to RDF", func() {
		Context("when parsing a form", func() {
			It("should initialize valid", func() {
				conv := NewCSVConvertor("test")
				Expect(conv.SubjectColumn).To(BeEmpty())
			})

			It("should get the subject column from a list of headers", func() {
				in, err := os.Open("testdata/UUIDsMemorixNaarHub3_new.csv")
				Expect(err).ToNot(HaveOccurred())
				conv := CSVConvertor{
					SubjectColumn: "handle-uuid",
					Separator:     ";",
					InputFile:     in,
				}
				records, err := conv.GetReader()
				Expect(records).ToNot(HaveLen(0))
				Expect(err).ToNot(HaveOccurred())
				header := records[0]
				Expect(conv.GetSubjectColumn(header, conv.SubjectColumn)).To(Equal(9))
			})

			It("should create a header map", func() {
				in, err := os.Open("testdata/UUIDsMemorixNaarHub3_new.csv")
				Expect(err).ToNot(HaveOccurred())
				conv := NewCSVConvertor("test")
				conv.InputFile = in
				conv.Separator = ";"
				conv.SubjectColumn = "handle-uuid"
				conv.PredicateURIBase = "http=//data.hub3.nl/def/"

				records, err := conv.GetReader()
				Expect(err).ToNot(HaveOccurred())
				conv.CreateHeader(records[0])
				hMap := conv.HeaderMap()
				Expect(hMap).ToNot(BeEmpty())
				Expect(hMap[0].String()).To(HaveSuffix(">"))
				Expect(hMap[0].String()).To(ContainSubstring("data.hub3.nl/def/"))
			})

			It("should create a subject uri", func() {
				conv := CSVConvertor{
					// InputFile:     in,
					Separator:      ";",
					SubjectColumn:  "handle-uuid",
					SubjectClass:   "http://www.europeana.eu/schemas/edm/WebResource",
					SubjectURIBase: "http://data.hub3.nl/resource/",
				}

				uri, typeTriple := conv.CreateSubjectResource("1234")
				Expect(uri.String()).To(Equal("<http://data.hub3.nl/resource/1234>"))
				Expect(typeTriple.Object.String()).To(Equal(rdf2go.NewResource(conv.SubjectClass).String()))
			})

			It("should create a triple for non-empty values", func() {
				conv := NewCSVConvertor("test")
				conv.Separator = ";"
				conv.SubjectColumn = "handle-uuid"
				conv.SubjectClass = "http://www.europeana.eu/schemas/edm/WebResource"
				conv.SubjectURIBase = "http://data.hub3.nl/resource/"

				t := conv.CreateTriple(rdf2go.NewResource("urn:s"), 0, "not empty")
				Expect(t).ToNot(BeNil())

				t = conv.CreateTriple(rdf2go.NewResource("urn:s"), 0, "")
				Expect(t).To(BeNil())

				t = conv.CreateTriple(rdf2go.NewResource("urn:s"), 0, " ")
				Expect(t).To(BeNil())
			})

			It("should parse a file", func() {
				in, err := os.Open("testdata/UUIDsMemorixNaarHub3_new.csv")
				Expect(err).ToNot(HaveOccurred())
				conv := NewCSVConvertor("test")
				conv.InputFile = in
				conv.Separator = ";"
				conv.SubjectColumn = "handle-uuid"
				Expect(conv.InputFile).ToNot(BeNil())
				triples, totalRows, err := conv.CreateTriples()
				Expect(err).ToNot(HaveOccurred())
				Expect(triples).ToNot(BeEmpty())
				Expect(totalRows).To(Equal(2944))
			})
		})
	})
})
