package hub3

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Indexer", func() {

	Describe("when initialised", func() {

		It("should have a search client", func() {
			c := ESClient()
			Expect(c).ToNot(BeNil())
			Expect(client).ToNot(BeNil())
		})

		It("should have a bulk-indexer", func() {
			bps := IndexingProcessor()
			Expect(bps).ToNot(BeNil())
			Expect(processor).ToNot(BeNil())
		})

	})
})

var testTurtleRecord string = `
@prefix abc: <http://www.ab-c.nl/> .
@prefix abm: <http://purl.org/abm/sen> .
@prefix aff: <http://schemas.delving.eu/aff/> .
@prefix cc: <http://creativecommons.org/ns#> .
@prefix custom: <http://www.delving.eu/namespaces/custom/> .
@prefix dbpedia-owl: <http://dbpedia.org/ontology/> .
@prefix dc: <http://purl.org/dc/elements/1.1/> .
@prefix dcterms: <http://purl.org/dc/terms/> .
@prefix delving: <http://schemas.delving.eu/> .
@prefix devmode: <http://localhost:8000/resource/> .
@prefix drup: <http://www.itin.nl/drupal> .
@prefix edm: <http://www.europeana.eu/schemas/edm/> .
@prefix europeana: <http://www.europeana.eu/schemas/ese/> .
@prefix foaf: <http://xmlns.com/foaf/0.1/> .
@prefix gn: <http://www.geonames.org/ontology#> .
@prefix icn: <http://www.icn.nl/schemas/icn/> .
@prefix itin: <http://www.itin.nl/namespace> .
@prefix mip: <http://data.brabantcloud.nl/resource/ns/mip/> .
@prefix musip: <http://www.musip.nl/> .
@prefix narthex: <http://schemas.delving.eu/narthex/terms/> .
@prefix nave: <http://schemas.delving.eu/nave/terms/> .
@prefix ns1: <urn:ebu:metadata-schema:> .
@prefix ore: <http://www.openarchives.org/ore/terms/> .
@prefix owl: <http://www.w3.org/2002/07/owl#> .
@prefix raw: <http://delving.eu/namespaces/raw> .
@prefix rda: <http://rdvocab.info/ElementsGr2/> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .
@prefix rdfs: <http://www.w3.org/2000/01/rdf-schema#> .
@prefix skos: <http://www.w3.org/2004/02/skos/core#> .
@prefix tib: <http://schemas.delving.eu/resource/ns/tib/> .
@prefix wgs84_pos: <http://www.w3.org/2003/01/geo/wgs84_pos#> .
@prefix xml: <http://www.w3.org/XML/1998/namespace> .
@prefix xsd: <http://www.w3.org/2001/XMLSchema#> .

<http://data.brabantcloud.nl/resource/aggregation/ton-smits-huis/C06097/about> a foaf:Document ;
    cc:attributionName "ton-smits-huis" ;
    narthex:belongsTo <http://data.brabantcloud.nl/resource/dataset/ton-smits-huis> ;
    narthex:contentHash "2938fd50b5d9df105b4a28211ac222e170a73ea2" ;
    narthex:saveTime "2017-09-08T13:45:15Z" ;
    narthex:synced false ;
    foaf:primaryTopic <http://data.brabantcloud.nl/resource/aggregation/ton-smits-huis/C06097> .

<http://data.brabantcloud.nl/resource/agent/ton-smits-huis/Ton%20Smits> a edm:Agent ;
    rda:professionOrOccupation "cartoonist" ;
    skos:altLabel "Smits, Ton" ;
    skos:prefLabel "Ton Smits" .

<http://data.brabantcloud.nl/resource/document/ton-smits-huis/C06097> a edm:ProvidedCHO ;
    dc:creator <http://data.brabantcloud.nl/resource/agent/ton-smits-huis/Ton%20Smits> ;
    dc:description "Een man zit aan een tafeltje in een restaurant. Hij zet grote ogen op als er een violist aan komt lopen. De violist -in zigeuner-kleding- komt spelend naast de tafel van de man staan. Glimlachend luistert de man naar de vioolmuziek. De muziek ontroert hem. Op de derde tekening biggelt er een traan over de rechterwang van de man. Op de vierde tekening huilt hij. Op de vijfde tekening brult de man het uit van verdriet. De tranen spoelen over zijn gezicht. De violist is al die tijd onverstoorbaar glimlachend doorgegaan met musiceren. Nu schrikt hij als de man hem boos iets toeroept. De violist deinst achteruit. De tranen lopen de man nog steeds over de wangen. Op de zevende en laatste tekening rent de violist geschrokken weg. De man blijft boos acher aan het tafeltje." ;
    dc:identifier "C06097" ;
    dc:rights "© L. Smits-Zoetmulder, info@tonsmitshuis.nl" ;
    dc:subject "angst",
        "luisteren",
        "mannen",
        "musiceren",
        "musici",
        "rennen",
        "restaurants",
        "schreeuwen",
        "tafels (dragend meubilair)",
        "tranen (menselijk)",
        "verdriet",
        "violen",
        "violisten",
        "vluchten",
        "vreugde",
        "woede",
        "zigeuners" ;
    dc:title "strip bestaande uit 7 tekeningen" ;
    dc:type "cartoon",
        "strip" ;
    dcterms:medium "papier",
        "viltstift" ;
    edm:type "IMAGE" .

<http://data.brabantcloud.nl/resource/aggregation/ton-smits-huis/C06097> a narthex:Record,
        ore:Aggregation ;
    edm:aggregatedCHO <http://data.brabantcloud.nl/resource/document/ton-smits-huis/C06097> ;
    edm:dataProvider "Ton Smits Huis" ;
    edm:isShownAt <http://data.brabantcloud.nl/resource/aggregation/ton-smits-huis/C06097> ;
    edm:isShownBy <http://media.delving.org/thumbnail/brabantcloud/ton-smits-huis/C06097agS/500> ;
    edm:object <http://media.delving.org/thumbnail/brabantcloud/ton-smits-huis/C06097agS/220> ;
    edm:provider "Erfgoed Brabant" ;
    edm:rights <http://www.europeana.eu/rights/rr-r/> ;
    ore:aggregates [ a nave:BrabantCloudResource ;
            nave:collection "Ton Smits Huis" ;
            nave:collectionPart "Cartoons" ;
            nave:material "papier",
                "viltstift" ;
            nave:objectNumber "C06097" ;
            nave:objectSoort "cartoon",
                "strip" ;
            nave:place "Eindhoven" ;
            nave:technique "getekend" ;
            nave:thumbLarge "http://media.delving.org/thumbnail/brabantcloud/ton-smits-huis/C06097agS/500",
                "http://media.delving.org/thumbnail/brabantcloud/ton-smits-huis/C06097bS/500",
                "http://media.delving.org/thumbnail/brabantcloud/ton-smits-huis/C06097cS/500",
                "http://media.delving.org/thumbnail/brabantcloud/ton-smits-huis/C06097dS/500",
                "http://media.delving.org/thumbnail/brabantcloud/ton-smits-huis/C06097eS/500",
                "http://media.delving.org/thumbnail/brabantcloud/ton-smits-huis/C06097fS/500",
                "http://media.delving.org/thumbnail/brabantcloud/ton-smits-huis/C06097gS/500" ;
            nave:thumbSmall "http://media.delving.org/thumbnail/brabantcloud/ton-smits-huis/C06097agS/220",
                "http://media.delving.org/thumbnail/brabantcloud/ton-smits-huis/C06097bS/220",
                "http://media.delving.org/thumbnail/brabantcloud/ton-smits-huis/C06097cS/220",
                "http://media.delving.org/thumbnail/brabantcloud/ton-smits-huis/C06097dS/220",
                "http://media.delving.org/thumbnail/brabantcloud/ton-smits-huis/C06097eS/220",
                "http://media.delving.org/thumbnail/brabantcloud/ton-smits-huis/C06097fS/220",
                "http://media.delving.org/thumbnail/brabantcloud/ton-smits-huis/C06097gS/220" ],
        [ a nave:DelvingResource ;
            nave:allowDeepZoom "true" ;
            nave:allowLinkedOpenData "true" ;
            nave:allowSourceDownload "false" ;
            nave:deepZoomUrl "http://media.delving.org/deepzoom/brabantcloud/ton-smits-huis/C06097agS.tif.dzi",
                "http://media.delving.org/deepzoom/brabantcloud/ton-smits-huis/C06097bS.tif.dzi",
                "http://media.delving.org/deepzoom/brabantcloud/ton-smits-huis/C06097cS.tif.dzi",
                "http://media.delving.org/deepzoom/brabantcloud/ton-smits-huis/C06097dS.tif.dzi",
                "http://media.delving.org/deepzoom/brabantcloud/ton-smits-huis/C06097eS.tif.dzi",
                "http://media.delving.org/deepzoom/brabantcloud/ton-smits-huis/C06097fS.tif.dzi",
                "http://media.delving.org/deepzoom/brabantcloud/ton-smits-huis/C06097gS.tif.dzi" ;
            nave:featured "false" ;
            nave:public "true" ] .
`
