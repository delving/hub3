package rdf

import (
	"strings"
	"testing"

	"github.com/matryer/is"
)

func TestNestedRDFAboutAsSubject(t *testing.T) {
	is := is.New(t)

	// The RDF-XML document with nested resources
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<rdf:RDF xmlns:crm="http://www.cidoc-crm.org/cidoc-crm/" xmlns:lrm="https://www.iflastandards.info/ns/lrm/lrmoo#" xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <lrm:F3_Manifestation rdf:about="https://data.antwerp.be/id/manifestation/brocade-catalog/c:lvd:538499">
    <crm:P2_has_type>
      <crm:E55_Type rdf:about="https://data.antwerp.be/id/term/brocade-authorities/a::pt.42:1">
        <crm:P2_has_type rdf:resource="https://data.antwerp.be/id/term/brocade-authorities/ty/pt"/>
      </crm:E55_Type>
    </crm:P2_has_type>
  </lrm:F3_Manifestation>
</rdf:RDF>`

	// Create a decoder with a StringReader
	decoder := NewRDFXMLDecoder(strings.NewReader(xmlData))

	// Parse all triples
	triples, err := decoder.DecodeAll()
	is.NoErr(err)

	// The target IRI we're looking for as a subject
	targetIRI := "https://data.antwerp.be/id/term/brocade-authorities/a::pt.42:1"

	// Check if there's at least one triple with our target as subject
	foundAsSubject := false
	for _, triple := range triples {
		if iri, ok := triple.Subj.(IRI); ok {
			if iri.str == targetIRI {
				foundAsSubject = true
				break
			}
		}
	}

	// Log all parsed triples for debugging
	t.Logf("Parsed %d triples:", len(triples))
	for i, triple := range triples {
		t.Logf("Triple %d: %v", i, triple)
	}

	// Assert that we found the target IRI as a subject
	is.True(foundAsSubject) // The target IRI should be a subject in at least one triple
}

// TestNestedTypedNodesWithoutParseType tests deeply nested typed node elements
// without rdf:parseType="Resource". This is valid RDF/XML using striping rules:
// typed node elements (e.g. <crm:E35_Title>) inside property elements create
// blank nodes with rdf:type automatically. This is the output format produced
// by the SIP-Creator for CIDOC-CRM mappings.
func TestNestedTypedNodesWithoutParseType(t *testing.T) {
	is := is.New(t)

	// SIP-Creator output for a CIDOC-CRM mapping - uses typed node elements
	// without rdf:parseType="Resource"
	xmlData := `<?xml version='1.0' encoding='UTF-8'?>
<rdf:RDF xmlns:crm="http://www.cidoc-crm.org/cidoc-crm/" xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#" xmlns:rdfs="http://www.w3.org/2000/01/rdf-schema#" xmlns:xml="http://www.w3.org/XML/1998/namespace">
    <crm:E22_Human-Made_Object>
        <crm:P102_has_title>
            <crm:E35_Title>
                <crm:P190_has_symbolic_content xml:lang="nl-NL">Aquarel: Schepen nabij een kust.</crm:P190_has_symbolic_content>
                <crm:P2_has_type>
                    <crm:E55_Type>
                        <rdfs:label>werktitel</rdfs:label>
                    </crm:E55_Type>
                </crm:P2_has_type>
            </crm:E35_Title>
        </crm:P102_has_title>
        <crm:P2_has_type>
            <crm:E55_Type>
                <rdfs:label xml:lang="en-US">watercolors (paintings)</rdfs:label>
                <rdfs:label xml:lang="nl-NL">aquarel</rdfs:label>
            </crm:E55_Type>
        </crm:P2_has_type>
        <crm:P1_is_identified_by>
            <crm:E42_Identifier>
                <crm:P190_has_symbolic_content>AS.1963.030.144</crm:P190_has_symbolic_content>
            </crm:E42_Identifier>
        </crm:P1_is_identified_by>
    </crm:E22_Human-Made_Object>
</rdf:RDF>`

	decoder := NewRDFXMLDecoder(strings.NewReader(xmlData))

	triples, err := decoder.DecodeAll()
	is.NoErr(err) // RDF/XML with nested typed nodes (no parseType) should parse without error

	t.Logf("Parsed %d triples:", len(triples))
	for i, triple := range triples {
		t.Logf("Triple %d: %v", i, triple)
	}

	// Expected triples from this document:
	// 1. _:b0 rdf:type crm:E22_Human-Made_Object
	// 2. _:b0 crm:P102_has_title _:b1
	// 3. _:b1 rdf:type crm:E35_Title
	// 4. _:b1 crm:P190_has_symbolic_content "Aquarel: Schepen nabij een kust."@nl-NL
	// 5. _:b1 crm:P2_has_type _:b2
	// 6. _:b2 rdf:type crm:E55_Type
	// 7. _:b2 rdfs:label "werktitel"
	// 8. _:b0 crm:P2_has_type _:b3
	// 9. _:b3 rdf:type crm:E55_Type
	// 10. _:b3 rdfs:label "watercolors (paintings)"@en-US
	// 11. _:b3 rdfs:label "aquarel"@nl-NL
	// 12. _:b0 crm:P1_is_identified_by _:b4
	// 13. _:b4 rdf:type crm:E42_Identifier
	// 14. _:b4 crm:P190_has_symbolic_content "AS.1963.030.144"
	is.Equal(len(triples), 14) // Should produce exactly 14 triples

	// Verify key triples exist
	rdfType := "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"
	crmNS := "http://www.cidoc-crm.org/cidoc-crm/"
	rdfsNS := "http://www.w3.org/2000/01/rdf-schema#"

	// Count rdf:type triples - should be 5 (E22, E35, E55, E55, E42)
	typeCount := 0
	for _, triple := range triples {
		if pred, ok := triple.Pred.(IRI); ok && pred.str == rdfType {
			typeCount++
		}
	}
	is.Equal(typeCount, 5) // 5 typed node elements should produce 5 rdf:type triples

	// Verify the title literal with language tag exists
	foundTitle := false
	for _, triple := range triples {
		if pred, ok := triple.Pred.(IRI); ok && pred.str == crmNS+"P190_has_symbolic_content" {
			if lit, ok := triple.Obj.(Literal); ok && lit.str == "Aquarel: Schepen nabij een kust." {
				is.Equal(lit.lang, "nl-NL")
				foundTitle = true
			}
		}
	}
	is.True(foundTitle) // Should find the title with nl-NL language tag

	// Verify rdfs:label "werktitel" exists (deeply nested, 4 levels deep)
	foundLabel := false
	for _, triple := range triples {
		if pred, ok := triple.Pred.(IRI); ok && pred.str == rdfsNS+"label" {
			if lit, ok := triple.Obj.(Literal); ok && lit.str == "werktitel" {
				foundLabel = true
			}
		}
	}
	is.True(foundLabel) // Should find deeply nested rdfs:label "werktitel"
}

// TestFullDocumentNestedRDFAbout tests the complete document with all its nested elements
func TestFullDocumentNestedRDFAbout(t *testing.T) {
	is := is.New(t)

	// This is the complete RDF-XML document to test
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<rdf:RDF xmlns:crm="http://www.cidoc-crm.org/cidoc-crm/" xmlns:lrm="https://www.iflastandards.info/ns/lrm/lrmoo#" xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <lrm:F3_Manifestation rdf:about="https://data.antwerp.be/id/manifestation/brocade-catalog/c:lvd:538499">
    <crm:P43_has_dimension>
      <crm:E54_Dimension>
        <crm:P129i_is_subject_of>
          <crm:E33_Linguistic_Object>
            <crm:P190_has_symbolic_content>1 v.</crm:P190_has_symbolic_content>
            <crm:P2_has_type rdf:resource="https://data.antwerp.be/id/term/brocade-catalog/co-pg"/>
          </crm:E33_Linguistic_Object>
        </crm:P129i_is_subject_of>
      </crm:E54_Dimension>
    </crm:P43_has_dimension>
    <crm:P102_has_title>
      <crm:E35_Title>
        <crm:P72_has_language rdf:resource="https://data.antwerp.be/id/term/brocade-catalog/lg/dut"/>
        <crm:P190_has_symbolic_content>Keur van proza- en dichtstukken</crm:P190_has_symbolic_content>
        <crm:P2_has_type rdf:resource="https://data.antwerp.be/id/term/brocade-catalog/ti-ty/h"/>
      </crm:E35_Title>
    </crm:P102_has_title>
    <crm:P2_has_type>
      <crm:E55_Type rdf:about="https://data.antwerp.be/id/term/brocade-catalog/dr/paper">
        <crm:P2_has_type rdf:resource="https://data.antwerp.be/id/term/brocade-catalog/dr"/>
      </crm:E55_Type>
    </crm:P2_has_type>
    <crm:P2_has_type>
      <crm:E55_Type rdf:about="https://data.antwerp.be/id/term/brocade-authorities/a::pt.42:1">
        <crm:P2_has_type rdf:resource="https://data.antwerp.be/id/term/brocade-authorities/ty/pt"/>
      </crm:E55_Type>
    </crm:P2_has_type>
  </lrm:F3_Manifestation>
</rdf:RDF>`

	// Create a decoder with a StringReader
	decoder := NewRDFXMLDecoder(strings.NewReader(xmlData))

	// Parse all triples
	triples, err := decoder.DecodeAll()
	is.NoErr(err)

	// The target IRI we're looking for as a subject
	targetIRI := "https://data.antwerp.be/id/term/brocade-authorities/a::pt.42:1"

	// Check if there's at least one triple with our target as subject
	foundAsSubject := false
	var subjectTriple Triple

	for _, triple := range triples {
		if iri, ok := triple.Subj.(IRI); ok {
			if iri.str == targetIRI {
				foundAsSubject = true
				subjectTriple = triple
				break
			}
		}
	}

	// Assert that we found the target IRI as a subject
	is.True(foundAsSubject) // The target IRI should be a subject in at least one triple

	// If found, verify the predicate and object of the triple with our target as subject
	if foundAsSubject {
		expectedPredicate := "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"
		expectedObject := "http://www.cidoc-crm.org/cidoc-crm/E55_Type"

		predIRI, ok := subjectTriple.Pred.(IRI)
		is.True(ok)                              // Predicate should be an IRI
		is.Equal(predIRI.str, expectedPredicate) // Check predicate value

		objIRI, ok := subjectTriple.Obj.(IRI)
		is.True(ok)                          // Object should be an IRI
		is.Equal(objIRI.str, expectedObject) // Check object value
	}
}
