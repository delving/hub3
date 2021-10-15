package resource

import "errors"

var (
	// ErrEmptyIRI is returned when an IRI is empty
	ErrEmptyIRI = errors.New("empty IRI is not allowed")

	// ErrInvalidNamespaceLabel is returned when invalid characters are
	// part of the namespace label.
	ErrInvalidNamespaceLabel = errors.New("invalid namespace label")

	// ErrDisallowedCharacterInIRI is returned when an IRI contains
	// any of the disallowed characters: [\x00-\x20<>"{}|^`\]
	// This error is ofter wrapped in another error showing the offending character
	ErrDisallowedCharacterInIRI = errors.New("disallowed character")

	// ErrEmptyBlankNode is returned when a BlankNode is empty
	ErrEmptyBlankNode = errors.New("empty BlankNode is not allowed")

	// ErrInvalidLanguageTag is returned when a RDF Literal.lang tag is invalid
	ErrInvalidLanguageTag = errors.New("invalid language tag")

	// ErrInvalidDataType is returned when the RDF Literal.DataType IRI is malformed
	ErrInvalidDataType = errors.New("invalid Literal.DataType IRI")

	// ErrUnsupportedDataType is returned when the RDF Literal.DataType is not part
	// of SupportDataTypes.
	ErrUnsupportedDataType = errors.New("unsupported Literal.DataType IRI")

	// ErrInvalidLiteral is returned when the RDF Literal value is invalid
	ErrInvalidLiteral = errors.New("invalid literal value")
)
