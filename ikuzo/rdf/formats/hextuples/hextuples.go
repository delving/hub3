package hextuples

import (
	"encoding/json"
	"fmt"

	"github.com/OneOfOne/xxhash"
	"github.com/delving/hub3/ikuzo/rdf"
)

// hexPosition describes the position the data has in a HexTuple slice.
//
//go:generate stringer -type=HexPosition
type hexPosition int

const (
	subject hexPosition = iota
	predicate
	value
	dataType
	language
	graph
)

// MimeType of the hextuple ND-JSON stream
const MimeType string = "application/hex+x-ndjson"

// FileExtension for Hextuples in ND-JSON format
const FileExtension string = "hext"

// hexTupleLength is the length of a HexTuple array
const hexTupleLength = 6

const (
	namedNode = "globalId"
	blankNode = "localId"
)

type HexTuple struct {
	Subject   string `parquet:"name=subject, type=BYTE_ARRAY, convertedtype=UTF8, encoding=RLE_DICTIONARY"`
	Predicate string `parquet:"name=predicate, type=BYTE_ARRAY, convertedtype=UTF8, encoding=RLE_DICTIONARY"`
	Value     string `parquet:"name=value, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN"`
	DataType  string `parquet:"name=datatype, type=BYTE_ARRAY, convertedtype=UTF8, encoding=RLE_DICTIONARY"`
	Language  string `parquet:"name=language, type=BYTE_ARRAY, convertedtype=UTF8, encoding=RLE_DICTIONARY"`
	Graph     string `parquet:"name=graph, type=BYTE_ARRAY, convertedtype=UTF8, encoding=RLE_DICTIONARY"`
}

func (ht *HexTuple) AsTriple() (*rdf.Triple, error) {
	s, err := rdf.NewIRI(ht.Subject)
	if err != nil {
		return nil, err
	}

	p, err := rdf.NewIRI(ht.Predicate)
	if err != nil {
		return nil, err
	}

	var obj rdf.Object

	switch ht.DataType {
	case namedNode:
		obj, err = rdf.NewIRI(ht.Value)
		if err != nil {
			return nil, err
		}
	case blankNode:
		obj, err = rdf.NewBlankNode(ht.Value)
		if err != nil {
			return nil, err
		}

	default:
		if ht.DataType == "" && ht.Language == "" {
			obj, err = rdf.NewLiteral(ht.Value)
			if err != nil {
				return nil, err
			}
		}

		if ht.DataType != "" {
			dt, iriErr := rdf.NewIRI(ht.DataType)
			if iriErr != nil {
				return nil, iriErr
			}

			obj, err = rdf.NewLiteralWithType(ht.Value, dt)
			if err != nil {
				return nil, err
			}
		}

		if ht.Language != "" {
			obj, err = rdf.NewLiteralWithLang(ht.Value, ht.Language)
			if err != nil {
				return nil, err
			}
		}
	}

	return rdf.NewTriple(rdf.Subject(s), rdf.Predicate(p), obj), nil
}

func FromTriple(t *rdf.Triple, graph string) HexTuple {
	ht := HexTuple{
		Subject:   t.Subject.RawValue(),
		Predicate: t.Predicate.RawValue(),
		Value:     t.Object.RawValue(),
		Graph:     graph,
	}

	switch t.Object.Type() {
	case rdf.TermIRI:
		ht.DataType = namedNode
	case rdf.TermBlankNode:
		ht.DataType = blankNode
	case rdf.TermLiteral:
		obj := t.Object.(rdf.Literal)
		ht.DataType = obj.DataType.RawValue()
		ht.Language = obj.Lang()
	}

	return ht
}

func (ht *HexTuple) UnmarshalJSON(data []byte) error {
	var v []string
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	if len(v) != hexTupleLength {
		return fmt.Errorf("invalid length for hextuple array")
	}

	ht.Subject = v[subject]
	ht.Predicate = v[predicate]
	ht.Value = v[value]
	ht.DataType = v[dataType]
	ht.Language = v[language]
	ht.Graph = v[graph]

	return nil
}

// MarshalJSON marshals the HexTuple as a 6 length string array
func (ht *HexTuple) MarshalJSON() ([]byte, error) {
	hexList := [6]string{
		ht.Subject,
		ht.Predicate,
		ht.Value,
		ht.DataType,
		ht.Language,
		ht.Graph,
	}

	return json.Marshal(&hexList)
}

// New returns a HexTuple from a valid 6 item JSON string array
func New(b []byte) (HexTuple, error) {
	var ht HexTuple
	if err := json.Unmarshal(b, &ht); err != nil {
		return ht, fmt.Errorf("unable to marshal hextTuple; %w", err)
	}

	return ht, nil
}

func (ht *HexTuple) String() string {
	b, _ := ht.MarshalJSON()
	return string(b)
}

func (ht *HexTuple) Hash() string {
	b, _ := ht.MarshalJSON()
	hash := xxhash.Checksum64(b)

	return fmt.Sprintf("%d", hash)
}

func (ht *HexTuple) entry() Entry {
	return Entry{
		Value:    ht.Value,
		DataType: ht.DataType,
		Language: ht.Language,
	}
}
