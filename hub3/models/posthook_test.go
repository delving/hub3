package models

import (
	"bytes"
	"io/ioutil"
	"strings"

	r "github.com/kiivihal/rdf2go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func getRDFString(path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

var testJSON = `[{"http://purl.org/dc/terms/extent": [{"@value": "Hoogte: 186 mm, diameter: 148-165 mm"}], "http://purl.org/dc/terms/createdEnd": [{"@value": "-0221-01-01T00:00:01"}], "urn:ebu:metadata-schema:ebuCore_2014/hasMimeType":[{"@value":"image/jpeg","@type":"http://www.w3.org/2001/XMLSchema#string"}]}, "http://purl.org/dc/terms/created": [{"@value": "-481 t/m -221"}], "@id": "http://data.brabantcloud.nl/resource/document/museum-klok-en-peel/2458", "http://purl.org/dc/elements/1.1/title": [{"@value": "Terracottabel type bo [Periode van de Strijdende Staten]"}], "http://purl.org/dc/terms/spatial": [{"@value": "China, Azie"}], "@type": ["http://www.europeana.eu/schemas/edm/ProvidedCHO"], "http://purl.org/dc/elements/1.1/identifier": [{"@value": "2458"}], "http://purl.org/dc/elements/1.1/description": [{"@value": "Bellen uit terracotta zoals deze waren grafgiften. Zij werden gemaakt ter vervanging van het originele object dat in voorgaande perioden de dode meegegeven werd. Het bekendste voorbeeld van dit gebruik is het terracotta-leger van de eerste keizer van China."}], "http://purl.org/dc/elements/1.1/date": [{"@value": "-481 t/m -221"}], "http://www.europeana.eu/schemas/edm/type": [{"@value": "IMAGE"}], "http://purl.org/dc/terms/medium": [{"@value": "keramiek"}], "http://purl.org/dc/terms/created": [{"@value": "-0481-01-01T00:00:01"}]}, {"@type": ["http://xmlns.com/foaf/0.1/Document"], "http://schemas.delving.eu/narthex/terms/saveTime": [{"@value": "2018-02-12T18:36:30Z"}], "http://schemas.delving.eu/narthex/terms/belongsTo": [{"@id": "http://data.brabantcloud.nl/resource/dataset/museum-klok-en-peel"}], "http://schemas.delving.eu/narthex/terms/synced": [{"@value": false}], "http://schemas.delving.eu/narthex/terms/contentHash": [{"@value": "de8bc9366bacd77ed1d3060f0ba2b73e124c74f0"}], "@id": "http://data.brabantcloud.nl/resource/aggregation/museum-klok-en-peel/2458/about", "http://creativecommons.org/ns#attributionName": [{"@value": "museum-klok-en-peel"}], "http://xmlns.com/foaf/0.1/primaryTopic": [{"@id": "http://data.brabantcloud.nl/resource/aggregation/museum-klok-en-peel/2458"}]}, {"http://schemas.delving.eu/nave/terms/allowSourceDownload": [{"@value": "false"}], "@type": ["http://schemas.delving.eu/nave/terms/DelvingResource"], "http://schemas.delving.eu/nave/terms/allowLinkedOpenData": [{"@value": "true"}], "http://schemas.delving.eu/nave/terms/featured": [{"@value": "false"}], "http://schemas.delving.eu/nave/terms/allowDeepZoom": [{"@value": "true"}], "@id": "_:Nd1aca3dce3c7451ab5ce6d0c0f7a3009", "http://schemas.delving.eu/nave/terms/public": [{"@value": "true"}], "http://schemas.delving.eu/nave/terms/deepZoomUrl": [{"@value": "https://media.delving.org/iip/deepzoom/mnt/tib/tiles/brabantcloud/museum-klok-en-peel/2458-Bel_type_bo_terracotta_China_strijdende_staten_voorkant.tif.dzi"}]}, {"http://www.europeana.eu/schemas/edm/isShownBy": [{"@id": "https://media.delving.org/thumbnail/brabantcloud/museum-klok-en-peel/2458-Bel_type_bo_terracotta_China_strijdende_staten_voorkant/500"}], "@type": ["http://www.openarchives.org/ore/terms/Aggregation", "http://schemas.delving.eu/narthex/terms/Record"], "http://www.europeana.eu/schemas/edm/rights": [{"@id": "http://creativecommons.org/publicdomain/zero/1.0/"}], "http://www.europeana.eu/schemas/edm/object": [{"@id": "https://media.delving.org/thumbnail/brabantcloud/museum-klok-en-peel/2458-Bel_type_bo_terracotta_China_strijdende_staten_voorkant/220"}], "http://www.europeana.eu/schemas/edm/provider": [{"@value": "Erfgoed Brabant"}], "http://www.europeana.eu/schemas/edm/dataProvider": [{"@value": "Museum Klok & Peel"}], "http://www.europeana.eu/schemas/edm/aggregatedCHO": [{"@id": "http://data.brabantcloud.nl/resource/document/museum-klok-en-peel/2458"}], "@id": "http://data.brabantcloud.nl/resource/aggregation/museum-klok-en-peel/2458", "http://www.openarchives.org/ore/terms/aggregates": [{"@id": "_:Nc7d29843d06541eca36bea1cf446e648"}, {"@id": "_:Nd1aca3dce3c7451ab5ce6d0c0f7a3009"}], "http://www.europeana.eu/schemas/edm/isShownAt": [{"@id": "http://data.brabantcloud.nl/resource/aggregation/museum-klok-en-peel/2458"}]}, {"http://schemas.delving.eu/nave/terms/creatorRole": [{"@value": "gieter"}], "@type": ["http://schemas.delving.eu/nave/terms/BrabantCloudResource"], "http://schemas.delving.eu/nave/terms/collection": [{"@value": "Museum Klok & Peel"}], "http://schemas.delving.eu/nave/terms/thumbLarge": [{"@value": "https://media.delving.org/thumbnail/brabantcloud/museum-klok-en-peel/2458-Bel_type_bo_terracotta_China_strijdende_staten_voorkant/500"}], "@id": "_:Nc7d29843d06541eca36bea1cf446e648", "http://schemas.delving.eu/nave/terms/material": [{"@value": "keramiek"}], "http://schemas.delving.eu/nave/terms/collectionPart": [{"@value": "opgravingen"}], "http://schemas.delving.eu/nave/terms/collectionType": [{"@value": "Algemeen"}], "http://schemas.delving.eu/nave/terms/thumbSmall": [{"@value": "https://media.delving.org/thumbnail/brabantcloud/museum-klok-en-peel/2458-Bel_type_bo_terracotta_China_strijdende_staten_voorkant/220"}], "http://schemas.delving.eu/nave/terms/dimension": [{"@value": "Hoogte: 186 mm, diameter: 148-165 mm"}], "http://schemas.delving.eu/nave/terms/objectNumber": [{"@value": "2458"}]}]`

var subject = "http://data.brabantcloud.nl/resource/aggregation/enb-83-beeldmateriaal/enb-83.beeldmateriaal-620b3fa2-a2d8-796c-eae1-b8b9ca6947b7-14b9d8fd-a7f5-c901-2e2d-ae6d0966bd25"

var _ = Describe("Posthook", func() {

	Describe("when creating", func() {

		Context("from an RDF string", func() {

			It("should populate a graph", func() {
				content, err := getRDFString("test_data/enb_test_1.nt")
				Expect(err).ToNot(HaveOccurred())
				g := r.NewGraph(subject)
				err = g.Parse(strings.NewReader(content), "text/turtle")
				Expect(err).ToNot(HaveOccurred())
				posthook := NewPostHookJob(g, "enb-83-beeldmateriaal", false, subject)
				Expect(posthook).ToNot(BeNil())
				Expect(posthook.Graph.Len()).ToNot(Equal(0))
				jsonld, err := posthook.String()
				Expect(err).ToNot(HaveOccurred())
				Expect(jsonld).To(ContainSubstring("brabant"))
				Expect(jsonld).To(ContainSubstring("{\"@id\":"))
			})

			It("should update triple for ebuCore uris", func() {
				g := r.NewGraph("")
				t := r.NewTriple(
					r.NewResource(subject),
					r.NewResource("urn:ebu:metadata-schema:ebuCore_2014/hasMimeType"),
					r.NewLiteral("image/jpeg"),
				)
				Expect(g.Len()).To(Equal(0))
				ok := cleanEbuCore(g, t)
				Expect(ok).To(BeTrue())
				Expect(g.Len()).To(Equal(1))

				var b bytes.Buffer
				Expect(b.Len()).To(Equal(0))
				err := g.Serialize(&b, "application/ld+json")
				Expect(err).ToNot(HaveOccurred())
				Expect(b.Len()).ToNot(Equal(0))
				Expect(b.String()).ToNot(ContainSubstring("ebuCore_2014"))
				Expect(b.String()).To(ContainSubstring("ebucore#"))
			})

			It("should update triple for date uris", func() {
				g := r.NewGraph("")
				t := r.NewTriple(
					r.NewResource(subject),
					r.NewResource("http://purl.org/dc/terms/created"),
					r.NewLiteral("1984"),
				)
				Expect(g.Len()).To(Equal(0))
				ok := cleanDates(g, t)
				Expect(ok).To(BeTrue())
				Expect(g.Len()).To(Equal(1))

				var b bytes.Buffer
				Expect(b.Len()).To(Equal(0))
				err := g.Serialize(&b, "application/ld+json")
				Expect(err).ToNot(HaveOccurred())
				Expect(b.Len()).ToNot(Equal(0))
				Expect(b.String()).To(ContainSubstring("createdRaw"))
			})

			It("should rewrite ebuCore predicates", func() {
				content, err := getRDFString("test_data/enb_test_1.nt")
				Expect(err).ToNot(HaveOccurred())
				g := r.NewGraph(subject)
				err = g.Parse(strings.NewReader(content), "text/turtle")
				Expect(err).ToNot(HaveOccurred())
				Expect(g.Len()).ToNot(Equal(0))
				posthook := NewPostHookJob(g, "enb-83-beeldmateriaal", false, subject)
				Expect(posthook).ToNot(BeNil())
				Expect(posthook.Graph.Len()).ToNot(Equal(0))
				//posthook.cleanPostHookGraph()
				jsonld, err := posthook.String()
				Expect(err).ToNot(HaveOccurred())
				Expect(jsonld).To(ContainSubstring("brabant"))
				Expect(jsonld).To(ContainSubstring("{\"@id\":"))
				Expect(jsonld).ToNot(ContainSubstring("ebuCore_2014"))

			})

		})

	})

	//Describe("when converting to json-ld", func() {

	//Context("given a json-ld as string", func() {

	//It("should parse it correctly", func() {
	//g := r.NewGraph("")
	//err := g.Parse(strings.NewReader(testJSON), "application/ld+json")
	//Expect(err).ToNot(HaveOccurred())
	//Expect(g).ToNot(BeNil())

	//var b bytes.Buffer
	//Expect(b.Len()).To(Equal(0))
	//err = g.Serialize(&b, "application/ld+json")
	//Expect(err).ToNot(HaveOccurred())
	//Expect(b.Len()).ToNot(Equal(0))

	//////dl := nld.NewDefaultDocumentLoader(nil)

	////rdfString := b.String()
	////Expect(rdfString).ToNot(BeEmpty())

	////Expect(testJSON).To(Equal(rdfString))

	//////rdoc, err := dl.LoadDocument(jsonString)
	//////Expect(err).ToNot(HaveOccurred())
	//////Expect(rdoc).ToNot(BeNil())

	////proc := nld.NewJsonLdProcessor()
	////options := nld.NewJsonLdOptions("")
	////options.Format = "text/turtle"

	////doc, err := proc.FromRDF(rdfString, options)
	////Expect(err).ToNot(HaveOccurred())
	////Expect(doc).ToNot(BeEmpty())

	//})

	//It("should serialize it back to the correct string", func() {
	//r := strings.NewReader(testJSON)

	//var document interface{}
	//dec := json.NewDecoder(r)

	//err := dec.Decode(&document)
	//Expect(err).ToNot(HaveOccurred())

	//proc := nld.NewJsonLdProcessor()
	//options := nld.NewJsonLdOptions("")

	//expanded, err := proc.Expand(document, options)
	//Expect(err).ToNot(HaveOccurred())

	//Expect(expanded).To(Equal(document))
	////
	//})

	//It("should still be valid when applying the date cleaning", func() {
	////
	//})

	//It("should still be correct after applying the resource sorting", func() {
	////
	//})
	//})
	//})

})
