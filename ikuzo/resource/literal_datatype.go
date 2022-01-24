package resource

import "time"

// DateFormat defines the string representation of xsd:DateTime values. You can override
// it if you need another layout.
var DateFormat = time.RFC3339

// The XML schema built-in datatypes (xsd):
// https://dvcs.w3.org/hg/rdf/raw-file/default/rdf-concepts/index.html#xsd-datatypes
var (
	// Core types:                                                    // Corresponding Go datatype:

	xsdString  = &IRI{str: "http://www.w3.org/2001/XMLSchema#string"}  // string
	xsdBoolean = &IRI{str: "http://www.w3.org/2001/XMLSchema#boolean"} // bool
	xsdDecimal = &IRI{str: "http://www.w3.org/2001/XMLSchema#decimal"} // float64
	xsdInteger = &IRI{str: "http://www.w3.org/2001/XMLSchema#integer"} // int

	// IEEE floating-point numbers:

	xsdDouble = &IRI{str: "http://www.w3.org/2001/XMLSchema#double"} // float64
	xsdFloat  = &IRI{str: "http://www.w3.org/2001/XMLSchema#float"}  // float64

	// Time and date:

	// xsdDate = IRI{str: "http://www.w3.org/2001/XMLSchema#date"}
	// xsdTime          = IRI{str: "http://www.w3.org/2001/XMLSchema#time"}
	xsdDateTime = &IRI{str: "http://www.w3.org/2001/XMLSchema#dateTime"} // time.Time
	// xsdDateTimeStamp = IRI{str: "http://www.w3.org/2001/XMLSchema#dateTimeStamp"}

	// Recurring and partial dates:

	// xsdYear              = IRI{str: "http://www.w3.org/2001/XMLSchema#gYear"}
	// xsdMonth             = IRI{str: "http://www.w3.org/2001/XMLSchema#gMonth"}
	// xsdDay               = IRI{str: "http://www.w3.org/2001/XMLSchema#gDay"}
	// xsdYearMonth         = IRI{str: "http://www.w3.org/2001/XMLSchema#gYearMonth"}
	// xsdDuration          = IRI{str: "http://www.w3.org/2001/XMLSchema#Duration"}
	// xsdYearMonthDuration = IRI{str: "http://www.w3.org/2001/XMLSchema#yearMonthDuration"}
	// xsdDayTimeDuration   = IRI{str: "http://www.w3.org/2001/XMLSchema#dayTimeDuration"}

	// Limited-range integer numbers

	xsdByte = &IRI{str: "http://www.w3.org/2001/XMLSchema#byte"} // []byte
	// xsdShort = IRI{str: "http://www.w3.org/2001/XMLSchema#short"} // int16
	xsdInt = &IRI{str: "http://www.w3.org/2001/XMLSchema#int"} // int32
	// xsdLong  = IRI{str: "http://www.w3.org/2001/XMLSchema#long"}  // int64

	// Various

	rdfLangString = &IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#langString"} // string
	xmlLiteral    = &IRI{str: "http://www.w3.org/1999/02/22-rdf-syntax-ns#XMLLiteral"} // string

	// SupportDataTypes contains a list of valid DataType IRIs.
	// You can extend or set this list with additional XSD IRIs.
	SupportDataTypes = []*IRI{
		// core types
		xsdString, xsdBoolean, xsdDecimal, xsdInteger,
		// IEEE floating-point numbers:
		xsdDouble, xsdFloat,
		// Time and date:
		xsdDateTime,
		// Limited-range integer numbers
		xsdByte, xsdInt,
		// Various
		rdfLangString, xmlLiteral,
	}
)
