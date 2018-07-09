package fragments

import (
	"fmt"
	"strings"

	rdf "github.com/deiu/gon3"
	r "github.com/kiivihal/rdf2go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var ntriples = `<http://sws.geonames.org/2759059> <http://schemas.delving.eu/nave/terms/province> "Gemeente Sint-Michielsgestel" .
<http://sws.geonames.org/2759059> <http://schemas.delving.eu/nave/terms/municipality> "Gemeente Sint-Michielsgestel" .
_:b0-af19c481183c3d84 <http://schemas.delving.eu/nave/terms/place> "Berlicum" .`

var turtle = `@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .
@prefix rdfs: <http://www.w3.org/2000/01/rdf-schema#> .
@prefix skos: <http://www.w3.org/2004/02/skos/core#> .
@prefix skosxl: <http://www.w3.org/2008/05/skos-xl#> .
@prefix owl: <http://www.w3.org/2002/07/owl#> .
@prefix dc: <http://purl.org/dc/elements/1.1/> .
@prefix dcterms: <http://purl.org/dc/terms/> .
@prefix xsd: <http://www.w3.org/2001/XMLSchema#> .
@prefix tags: <http://www.holygoat.co.uk/owl/redwood/0.1/tags/> .
@prefix foaf: <http://xmlns.com/foaf/0.1/> .
@prefix cycAnnot: <http://sw.cyc.com/CycAnnotations_v1#> .
@prefix csw: <http://semantic-web.at/ontologies/csw.owl#> .
@prefix dbpedia: <http://dbpedia.org/resource/> .
@prefix freebase: <http://rdf.freebase.com/ns/> .
@prefix opencyc: <http://sw.opencyc.org/concept/> .
@prefix cyc: <http://sw.cyc.com/concept/> .
@prefix ctag: <http://commontag.org/ns#> .

<https://data.cultureelerfgoed.nl/term/id/cht/00238397-f6da-4444-b4f4-a2c5a3698c70> dcterms:modified "2017-03-17T09:09:33Z"^^xsd:dateTime ;
	<http://schema.semantic-web.at/ppt/inSubtree> <https://data.cultureelerfgoed.nl/term/id/cht/bece25a6-eb64-46e8-85a8-2a7991f02a2c> ;
	a skos:Concept , <https://data.cultureelerfgoed.nl/vocab/id/cht#CHTconcept> ;
	skos:altLabel "ansjovisjol"@nl ;
	skos:broader <https://data.cultureelerfgoed.nl/term/id/cht/4bb94bab-587b-4095-a431-1fc71814cff1> ;
	skos:inScheme <https://data.cultureelerfgoed.nl/term/id/cht/b532325c-dc08-49db-b4f1-15e53b037ec3> ;
	skos:prefLabel "ansjovisjollen"@nl ;
	skos:scopeNote "Fries vissersvaartuig uit Stavoren en Molkwerum, ook wel herfst- of fuikenjol genoemd; zij is een variant van de Staverse jol en werd i.h.b. gebruikt voor de ansjovisvisserij waarvoor lichte, fijne netten nodig waren; deze jollen waren rond 1900 in de vaart; de oudste typen werden gesleept, maar later werden ze getuigd met een sprietzeil en een fok die op de steven vast zat."@nl ;
	<https://data.cultureelerfgoed.nl/vocab/id/rce#hasConceptStatus> <https://data.cultureelerfgoed.nl/term/id/cht/c58475d5-0795-4623-b4be-ea1524f4b4fb> .`

var _ = Describe("Rdfstream", func() {

	Describe("when parsing a stream", func() {

		Context("and given a reader", func() {

			It("should produces a list of triples", func() {
				parser := rdf.NewParser("")
				reader := strings.NewReader(turtle)
				g, err := parser.Parse(reader)
				Expect(err).ToNot(HaveOccurred())
				for t := range g.IterTriples() {
					triple := r.NewTriple(rdf2term(t.Subject), rdf2term(t.Predicate), rdf2term(t.Object))
					fmt.Println(triple)
				}
				//Expect(g.IterTriples()).ToNot(BeEmpty())
			})
		})
	})

})
