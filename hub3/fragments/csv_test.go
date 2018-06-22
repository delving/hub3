package fragments_test

import (
	"os"

	"github.com/deiu/rdf2go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/delving/rapid-saas/hub3/fragments"
)

var _ = Describe("CSV", func() {

	Describe("converting to RDF", func() {

		Context("when parsing a form", func() {

			It("should initialize valid", func() {

				conv := NewCSVConvertor()
				Expect(conv.SubjectColumn).To(BeEmpty())

			})

			It("should get the subject column from a list of headers", func() {
				in, err := os.Open("test_data/UUIDsMemorixNaarHub3_new.csv")
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
				Expect(conv.GetSubjectColumn(header)).To(Equal(9))

			})

			It("should create a header map", func() {
				in, err := os.Open("test_data/UUIDsMemorixNaarHub3_new.csv")
				Expect(err).ToNot(HaveOccurred())
				conv := CSVConvertor{
					InputFile:        in,
					Separator:        ";",
					SubjectColumn:    "handle-uuid",
					PredicateURIBase: "http://data.rapid.nl/def/",
				}
				records, err := conv.GetReader()
				Expect(err).ToNot(HaveOccurred())
				hMap := conv.CreateHeader(records[0])
				Expect(hMap).ToNot(BeEmpty())
				Expect(hMap[0].String()).To(HaveSuffix(">"))
				Expect(hMap[0].String()).To(ContainSubstring("data.rapid.nl/def/"))

			})

			It("should create a subject uri", func() {
				conv := CSVConvertor{
					//InputFile:     in,
					Separator:      ";",
					SubjectColumn:  "handle-uuid",
					SubjectClass:   "http://www.europeana.eu/schemas/edm/WebResource",
					SubjectURIBase: "http://data.rapid.nl/resource/",
				}

				uri, typeTriple := conv.CreateSubjectResource("1234")
				Expect(uri.String()).To(Equal("<http://data.rapid.nl/resource/1234>"))
				Expect(typeTriple.Object.String()).To(Equal(rdf2go.NewResource(conv.SubjectClass).String()))
			})

			It("should parse a file", func() {
				in, err := os.Open("test_data/UUIDsMemorixNaarHub3_new.csv")
				Expect(err).ToNot(HaveOccurred())
				conv := CSVConvertor{
					InputFile:     in,
					Separator:     ";",
					SubjectColumn: "handle-uuid",
				}
				Expect(conv.InputFile).ToNot(BeNil())
				triples, err := conv.CreateTriples()
				Expect(err).ToNot(HaveOccurred())
				// TODO fix later
				Expect(triples).ToNot(BeEmpty())
			})
		})
	})

})
