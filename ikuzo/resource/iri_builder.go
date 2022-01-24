package resource

import (
	"fmt"
	"strings"

	"github.com/delving/hub3/ikuzo/validator"
)

// Shortcuts for the default support namespaces
var (
	DC      = &IRIBuilder{baseIRI: IRI{str: "http://purl.org/dc/elements/1.1/"}}
	DCAT    = &IRIBuilder{baseIRI: IRI{str: "http://www.w3.org/ns/dcat#"}}
	DCTERMS = &IRIBuilder{baseIRI: IRI{str: "http://purl.org/dc/terms/"}}
	EDM     = &IRIBuilder{baseIRI: IRI{str: "http://www.europeana.eu/schemas/edm/"}}
	FOAF    = &IRIBuilder{baseIRI: IRI{str: "http://xmlns.com/foaf/0.1/"}}
	IIIF    = &IRIBuilder{baseIRI: IRI{str: "http://iiif.io/api/image/2#"}}
	NAVE    = &IRIBuilder{baseIRI: IRI{str: "http://schemas.delving.eu/nave/terms/"}}
	ODRL    = &IRIBuilder{baseIRI: IRI{str: "http://www.w3.org/ns/odrl/2/"}}
	ORE     = &IRIBuilder{baseIRI: IRI{str: "http://www.openarchives.org/ore/terms/"}}
	OWL     = &IRIBuilder{baseIRI: IRI{str: "http://www.w3.org/2002/07/owl#"}}
	RDF     = &IRIBuilder{baseIRI: IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#"}}
	RDFS    = &IRIBuilder{baseIRI: IRI{str: "http://www.w3.org/2000/01/rdf-schema#"}}
	SCHEMA  = &IRIBuilder{baseIRI: IRI{str: "http://schema.org/"}}
	SKOS    = &IRIBuilder{baseIRI: IRI{str: "http://www.w3.org/2004/02/skos/core#"}}
	XML     = &IRIBuilder{baseIRI: IRI{str: "http://www.w3.org/XML/1998/namespace"}}
	XSD     = &IRIBuilder{baseIRI: IRI{str: "http://www.w3.org/2001/XMLSchema#"}}
)

// IRIBuilder is used to easily create namespaces IRIs.
type IRIBuilder struct {
	baseIRI IRI
}

// IRI returns a namespaces IRI suffixed with the given label.
//
// An error is returned when the label is invalid. Although this is extra error
// handling for each new IRI, we prefer the guarantees of correctness of the returned IRI.
func (b *IRIBuilder) IRI(label string) (IRI, error) {
	v := b.validate(label)
	if !v.Valid() {
		return IRI{}, v.ErrorOrNil()
	}

	iri, err := NewIRI(b.baseIRI.RawValue() + label)
	if err != nil {
		return iri, err
	}

	return iri, nil
}

func (b *IRIBuilder) validate(label string) *validator.Validator {
	v := validator.New()
	v.Check(!strings.Contains(label, "/"), "namespace label", ErrInvalidNamespaceLabel, fmt.Sprintf("'/' is not allowed in label '%s'", label))
	v.Check(!strings.Contains(label, "#"), "namespace label", ErrInvalidNamespaceLabel, fmt.Sprintf("'#' is not allowed in label '%s'", label))

	return v
}

// NewIRIBuilder returns an IRIBuilder.
//
// This base IRI is expected to be valid.
func NewIRIBuilder(base IRI) *IRIBuilder {
	return &IRIBuilder{baseIRI: base}
}
