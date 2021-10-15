package resource

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/delving/hub3/ikuzo/validator"
)

// Literal represents a RDF literal; a value with a datatype and
// (optionally) an associated language tag for strings.
type Literal struct {
	// The literal is always stored as a string, regardless of datatype.
	str string

	// Val represents the typed value of a RDF Literal, boxed in an empty interface.
	// A type assertion is needed to get the value in the corresponding Go type.
	val interface{}

	// lang, if not empty, represents the language tag of a string.
	// A language tagged string has the datatype: rdf:langString.
	lang string

	// The datatype of the Literal.
	DataType *IRI
}

// Equal returns whether this literal is equivalent to another.
func (l Literal) Equal(other Term) bool {
	var spec Literal

	switch t := other.(type) {
	case *Literal:
		spec = *t
	case Literal:
		spec = t
	default:
		// unsupported type
		return false
	}

	if l.str != spec.str {
		return false
	}

	if l.lang != spec.lang {
		return false
	}

	if l.DataType != nil {
		if !l.DataType.Equal(spec.DataType) {
			return false
		}
	}

	return true
}

// Lang returns the language of a language-tagged string.
func (l Literal) Lang() string {
	return l.lang
}

func atLang(lang string) string {
	if lang != "" {
		if strings.HasPrefix(lang, "@") {
			return lang
		}

		return "@" + lang
	}

	return ""
}

// String returns the NTriples representation of this literal string.
func (l Literal) String() string {
	str := l.str
	str = strings.Replace(str, "\\", "\\\\", -1)
	str = strings.Replace(str, "\"", "\\\"", -1)
	str = strings.Replace(str, "\n", "\\n", -1)
	str = strings.Replace(str, "\r", "\\r", -1)
	str = strings.Replace(str, "\t", "\\t", -1)

	str = fmt.Sprintf("\"%s\"", str)

	str += atLang(l.lang)

	// xsdString is implied
	if l.DataType != nil && (l.DataType != xsdString && l.DataType != rdfLangString) {
		str += "^^" + l.DataType.RawValue()
	}

	return str
}

func (l Literal) RawValue() string {
	return l.str
}

// Type returns the TermType of a Literal.
func (l Literal) Type() TermType {
	return TermLiteral
}

// Typed tries to parse the Literal's value into a Go type, according to the
// the DataType.
func (l Literal) Typed() (interface{}, error) {
	if l.val == nil {
		switch l.DataType.str {
		case xsdInteger.str, xsdInt.str:
			i, err := strconv.Atoi(l.str)
			if err != nil {
				return nil, err
			}

			// l.val = i

			return i, nil
		case xsdDouble.str, xsdDecimal.str:
			f, err := strconv.ParseFloat(l.str, 64)
			if err != nil {
				return nil, err
			}

			// l.val = f

			return f, nil
		case xsdBoolean.str:
			b, err := strconv.ParseBool(l.str)
			if err != nil {
				return nil, err
			}

			// l.val = b

			return b, nil
		case xsdByte.str:
			return []byte(l.str), nil
			// TODO xsdDateTime etc
		default:
			return l.str, nil
		}
	}

	return l.val, nil
}

func isValidDataType(dt *IRI) bool {
	if dt == nil {
		return false
	}

	for _, other := range SupportDataTypes {
		if dt.Equal(other) {
			return true
		}
	}

	return false
}

func (l Literal) Validate() *validator.Validator {
	v := validator.New()

	v.Check(l.str != "", "literal", ErrInvalidLiteral, "cannot be empty")

	if l.lang != "" {
		validateLanguageTag(v, l.lang)
	}

	v.Check(l.DataType != nil, "dataType", ErrInvalidDataType, "cannot be nil")

	if l.DataType != nil {
		v.Check(isValidDataType(l.DataType), "dataType", ErrUnsupportedDataType, l.DataType.RawValue())
	}

	// TODO(kiivihal): implement remainder validator

	return v
}

// validAsObject denotes that a Literal is valid as a Triple's Object.
func (l Literal) validAsObject() {}

// NewLiteral creates a RDF literal, it fails if the value string is not
// not well-formed.
//
// The literal will have the datatype IRI xsd:String.
func NewLiteral(str string) (Literal, error) {
	l := Literal{str: str, DataType: rdfLangString}

	v := l.Validate()
	if !v.Valid() {
		return Literal{}, v.ErrorOrNil()
	}

	return l, nil
}

// NewLiteralInferred returns a new Literal, or an error on invalid input. It tries
// to map the given Go values to a corresponding xsd datatype.
func NewLiteralInferred(v interface{}) (Literal, error) {
	switch t := v.(type) {
	case bool:
		return Literal{val: t, str: fmt.Sprintf("%v", t), DataType: xsdBoolean}, nil
	case int, int32, int64:
		return Literal{val: t, str: fmt.Sprintf("%v", t), DataType: xsdInteger}, nil
	case string:
		if strings.TrimSpace(t) == "" {
			return Literal{}, fmt.Errorf("%w: cannot be empty", ErrInvalidLiteral)
		}

		return Literal{str: t, DataType: xsdString}, nil
	case float32, float64:
		return Literal{val: t, str: fmt.Sprintf("%v", t), DataType: xsdDouble}, nil
	case time.Time:
		return Literal{val: t, str: t.Format(DateFormat), DataType: xsdDateTime}, nil
	case []byte:
		return Literal{val: t, str: string(t), DataType: xsdByte}, nil
	default:
		return Literal{}, fmt.Errorf("cannot infer XSD datatype from %#v", t)
	}
}

func validateLanguageTag(v *validator.Validator, lang string) {
	v.Check(!strings.HasPrefix(lang, "-"), "languageTag", ErrInvalidLanguageTag, "must start with a letter")
	v.Check(!strings.HasSuffix(lang, "-"), "languageTag", ErrInvalidLanguageTag, "trailing '-' disallowed")

	var afterDash bool

	for _, r := range lang {
		switch {
		case (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z'):
			continue
		case r == '-':
			if afterDash {
				v.AddError("languageTag", ErrInvalidLanguageTag, "only one '-' allowed")
			}

			afterDash = true
		case r >= '0' && r <= '9':
			if afterDash {
				continue
			}

			fallthrough
		default:
			v.AddError("languageTag", ErrInvalidLanguageTag, fmt.Sprintf("unexpected character: %q", r))
		}
	}
}

// NewLiteralWithLang creates a RDF literal with a given language tag, or fails
// if the language tag is not well-formed.
//
// The literal will have the datatype IRI xsd:String.
func NewLiteralWithLang(str, lang string) (Literal, error) {
	l := Literal{str: str, lang: lang, DataType: rdfLangString}

	v := l.Validate()
	if !v.Valid() {
		return Literal{}, v.ErrorOrNil()
	}

	return l, nil
}

// NewLiteralWithType returns a literal with the given datatype, or fails
// if the DataType IRI is malformed.
func NewLiteralWithType(str string, dt *IRI) (Literal, error) {
	l := Literal{str: str, DataType: dt}

	v := l.Validate()
	if !v.Valid() {
		return Literal{}, v.ErrorOrNil()
	}

	return l, nil
}
