package hextuples

import (
	"encoding/json"
	"fmt"
)

// HexPosition describes the position the data has in a HexTuple slice.
//
//go:generate stringer -type=HexPosition
type HexPosition int

const (
	Subject HexPosition = iota
	Predicate
	Value
	DataType
	Language
	Graph
)

// MimeType of the hextuple ND-JSON stream
const MimeType string = "application/hex+x-ndjson"

// FileExtension for Hextuples in ND-JSON format
const FileExtension string = "hext"

type HexTuple struct {
	hextuple [6]string
}

// New returns a HexTuple from a valid 6 item JSON string array
func New(b []byte) (HexTuple, error) {
	var ht HexTuple
	if err := json.Unmarshal(b, &ht.hextuple); err != nil {
		return ht, fmt.Errorf("unable to marshal hextTuple; %w", err)
	}

	return ht, nil
}

// Subject returns the Subject of the HexTuple
func (ht *HexTuple) Subject() string {
	return ht.hextuple[Subject]
}

// Predicate returns the Predicate of the HexTuple
func (ht *HexTuple) Predicate() string {
	return ht.hextuple[Predicate]
}

// Value returns the Value of the HexTuple
func (ht *HexTuple) Value() string {
	return ht.hextuple[Value]
}

// DataType returns the DataType of the HexTuple
func (ht *HexTuple) DataType() string {
	return ht.hextuple[DataType]
}

// Language returns the Language of the HexTuple
func (ht *HexTuple) Language() string {
	return ht.hextuple[Language]
}

// Graph returns the Graph of the HexTuple
func (ht *HexTuple) Graph() string {
	return ht.hextuple[Graph]
}

// Valid returns nil when the HexTuple is valid
func (ht *HexTuple) Valid() error {
	return nil
}
