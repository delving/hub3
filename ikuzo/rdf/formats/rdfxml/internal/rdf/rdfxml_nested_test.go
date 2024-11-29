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
