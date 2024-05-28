package index

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"unicode"

	"github.com/OneOfOne/xxhash"

	"github.com/delving/hub3/ikuzo/rdf"
)

type EntryType string

const (
	Literal      EntryType = "Literal"
	ResourceType EntryType = "Resource"
	Bnode        EntryType = "Bnode"
)

// Entry is an explicit conversion of an rdf.Triple for indexing in nested fields.
type Entry struct {
	// ID is the URI/IRI of the resource/blank-node of the object of the triple
	ID string `json:"@id,omitempty"`

	// Predicate is the URI/IRI of the predicate of the triple
	Predicate string `json:"predicate,omitempty"`

	// SearchLabel is the namespaced form of the predicate
	SearchLabel string `json:"searchLabel,omitempty"`

	// Value is the literal value of the triple. When the object of the triple
	// is a Resource or Bnode the Value can also be the inlined label of that resource.
	Value string `json:"@value,omitempty"`

	// Language is the language code of the triple
	Language string `json:"@language,omitempty"`

	// DataType is the rdf.DataType of the triple
	DataType string `json:"@type,omitempty"`

	// EntryType is object type of the triple, e.g. bnode, literal, resource
	EntryType EntryType `json:"entrytype,omitempty"`

	// Level is the depth of the resource. The higher the Level the lower it should be ranked
	Level int32 `json:"level,omitempty"`

	// Order the position the source triple had in the rdf.Graph.
	// This allows us to keep the entries sorted across serialization
	Order int `json:"order,omitempty"`

	// Tags can be queried and can trigger indexing in custom TypeIndexField
	Tags []string `json:"tags,omitempty"`

	// TypeIndexField are fields that trigger custom functionality in the index
	TypeIndexField

	// CustomFilterField are fields that can be used to create facets across different SearchLabels
	CustomFilterField

	// Inline is used to Inline resources for a grouped view. These are never indexed
	Inline *Resource `json:"inline,omitempty"`

	// fingerprint is the unique id of the triple values. It is used for deduplication.
	// It is set when fingerprint() is called.
	fingerprint string
}

// AddTags adds a tag string to the tags array of the Header
func (e *Entry) AddTags(tags ...string) {
	e.Tags = appendUnique(e.Tags, tags...)
}

// AsTriple converts an Entry to an rdf.Triple
func (e *Entry) AsTriple(subject rdf.Subject) (*rdf.Triple, error) {
	var err error
	predicate, err := rdf.NewIRI(e.Predicate)
	if err != nil {
		return nil, err
	}

	var object rdf.Object

	switch t := e.EntryType; t {
	case Bnode:
		object, err = rdf.NewBlankNode(e.ID)
	case ResourceType:
		object, err = rdf.NewIRI(e.ID)
	case Literal:
		switch {
		case e.Language != "":
			object, err = rdf.NewLiteralWithLang(e.Value, e.Language)
		case e.DataType != "":
			dt, iriErr := rdf.NewIRI(e.DataType)
			if iriErr != nil {
				return nil, iriErr
			}
			object, err = rdf.NewLiteralWithType(e.Value, dt)
		default:
			object, err = rdf.NewLiteral(e.Value)
		}
	default:
		slog.Warn("bad datatype", "entry", e, "type", t)
	}

	if err != nil {
		return nil, err
	}

	return rdf.NewTriple(
		subject,
		predicate,
		object,
	), nil
}

// Fingerprint calculates the unique content hash of the embedded triple.
// This value can be used for deduplication.
func (e *Entry) Fingerprint() string {
	if e.fingerprint != "" {
		return e.fingerprint
	}

	d := xxhash.New64()
	d.WriteString(e.ID)
	d.WriteString(e.Predicate)
	d.WriteString(e.Value)
	d.WriteString(e.Language)
	d.WriteString(e.DataType)

	e.fingerprint = fmt.Sprintf("%d", d.Sum64())

	return e.fingerprint
}

// processTags processes the tag array and sets values in TypeIndexField
func (e *Entry) processTags() error {
	if e.Value != "" {
		for _, tag := range e.Tags {
			switch tag {
			case "isoDate":
				e.Date = append(e.Date, e.Value)
			case "dateRange":
				indexRange, err := createDateRange(e.Value)
				if err != nil {
					return fmt.Errorf("unable to create dateRange for: %#v; %w", e.Value, err)
				}
				e.DateRange = &indexRange
				if indexRange.Greater != "" {
					e.Date = append(e.Date, indexRange.Greater)
				}
				if indexRange.Less != "" {
					e.Date = append(e.Date, indexRange.Less)
				}
			case "latLong":
				e.LatLong = e.Value
			case "integer":
				i, err := strconv.Atoi(e.Value)
				if err != nil {
					slog.Info("Unable to create integer", "value", e.Value, "error", err)
					continue
				}
				e.Integer = i
			}
		}
	}
	return nil
}

// TypeIndexField groups custom indexing fields together
type TypeIndexField struct {
	Date      []string    `json:"isoDate,omitempty"`
	DateRange *IndexRange `json:"dateRange,omitempty"`
	Integer   int         `json:"integer,omitempty"`
	Float     float64     `json:"float,omitempty"`
	IntRange  *IndexRange `json:"intRange,omitempty"`
	LatLong   string      `json:"latLong,omitempty"`
}

// CustomFilterField groups fields that can be used to create meta aggregations
// for facetting.
type CustomFilterField struct {
	FilterIDs []string `json:"mFilterID,omitempty"`
	Type      string   `json:"mType,omitempty"`
	Role      string   `json:"mRole,omitempty"`
}

// IndexRange is used for indexing ranges.
type IndexRange struct {
	Greater string `json:"gte"`
	Less    string `json:"lte"`
}

// Valid checks if Less is smaller than Greater.
func (ir IndexRange) Valid() error {
	if ir.Greater > ir.Less {
		return fmt.Errorf("%s should not be greater than %s", ir.Less, ir.Greater)
	}
	return nil
}

// CreateDateRange creates a date indexRange
func createDateRange(period string) (IndexRange, error) {
	ir := IndexRange{}
	parts := strings.FieldsFunc(strings.TrimSpace(period), splitPeriod)
	switch len(parts) {
	case 1:
		// start and end year
		ir.Greater, _ = padYears(parts[0], true)
		ir.Less, _ = padYears(parts[0], false)
	case 2:
		ir.Greater, _ = padYears(parts[0], true)
		ir.Less, _ = padYears(parts[1], false)
	default:
		return ir, fmt.Errorf("unable to create data range for: %#v", parts)
	}

	if err := ir.Valid(); err != nil {
		return ir, err
	}

	return ir, nil
}

// hyphenateDate converts a string of date string into the hyphenated form.
// Only YYYYMMDD and YYYYMM are supported.
func hyphenateDate(date string) (string, error) {
	switch len(date) {
	case 4:
		return date, nil
	case 6:
		return fmt.Sprintf("%s-%s", date[:4], date[4:]), nil
	case 8:
		return fmt.Sprintf("%s-%s-%s", date[:4], date[4:6], date[6:]), nil
	}
	return "", fmt.Errorf("unable to hyphenate date string: %#v", date)
}

func padYears(year string, start bool) (string, error) {
	var parts []string
	for _, p := range strings.Split(year, "-") {
		if strings.TrimSpace(p) != "" {
			parts = append(parts, strings.TrimSpace(p))
		}
	}

	switch len(parts) {
	case 3:
		return year, nil
	case 2:
		year := parts[0]
		month := parts[1]
		switch start {
		case true:
			return fmt.Sprintf("%s-%s-01", year, month), nil
		case false:
			switch parts[1] {
			case "01", "03", "05", "07", "08", "10", "12":
				return fmt.Sprintf("%s-%s-31", year, month), nil
			case "02":
				return fmt.Sprintf("%s-%s-28", year, month), nil
			default:
				return fmt.Sprintf("%s-%s-30", year, month), nil
			}
		}
	case 1:
		year := parts[0]
		switch len(year) {
		case 4:
			switch start {
			case true:
				return fmt.Sprintf("%s-01-01", year), nil
			case false:
				return fmt.Sprintf("%s-12-31", year), nil
			}
		default:
			// try to hyphenate the date
			date, err := hyphenateDate(year)
			if err != nil {
				return "", err
			}
			return padYears(date, start)
		}
	}
	return "", fmt.Errorf("unsupported case for padding: %s", year)
}

func splitPeriod(c rune) bool {
	return !unicode.IsNumber(c) && c != '-'
}
