package search

import (
	"fmt"
	"strconv"
	"strings"
)

// FacetField configures aggregrations for fields in the search response.
type FacetField struct {
	Field           string
	sortAsc         bool
	path            string
	nestedField     string
	aggregationType string
	size            int
	orderByKey      bool
}

// newFacetField parses a string and returns a *FacetField.
//
// The input field is a shorthand representation of the FacetField options,
// designed to be used in configuration or URL Query parameters.
//
// The shorthand is build up of modifiers and field-prefixes. The modifiers
// change the behaviour of the facet. The field-prefixes determine which
// facet-type and which field value is returned.
//
// The following modifiers are supported:
//
// ^ prefix reverse the sort-order of the facet. The default sort-order is
//   descending.
//   Example: ^dc_title.
//
// @ suffix sorts the facets on the key. The default is to sort on the count,
//   i.e. the number of records in which the facet value is found.
//   Example: dc_title@
//
// ~ sets the number of facets that are returned. If not set the default value
//   from the search.Service is used.
//   Example: dc_title~10
//
// In the fields, the '.' is used as a field-separator to define section and
// subfield. The following sections are supported:
//
// meta: is the header section that included with every search record
// tree: is the section to support search in hierachical structures such as
//       EAD and SKOS.
//
// The default section is 'resources.entries', which does not need to be explicitly
// added. When a '.' separated prefix is absent 'resources.entries' is added.
//
// Field-prefixes define which resource field should be searched. The default
// field is '@value', which is the object value of a triple. The namespaced field
// that is separated by an underscore '_' is called the SearchLabel. This determines
// which RDF predicate is used for the facet.
// TODO add link to namespace package and how SearchLabel are created from triples.
//
// The default field-prefix is term-aggregation. This returns a type-frequency
// list and does not have to be specified. It uses the @value field for the term.
//
// The following field-prefixes are supported that modify the default behaviour.
//
// datehistogram: uses the date field and returns a complete list of years
// and their frequency.
//
// dateminmax: uses the date field to return the earliest and latest date
// in the result set.
//
// tag: uses the tags field and returns a type-frequency list.
//
// id: used the 'id' field instead of the '@value' field. This means that it is
// aggregation the RDF resource URI instead of the literal value.
//
// Empty values are not allowed and will return an error.
func newFacetField(field string) (*FacetField, error) {
	if field == "" {
		return nil, fmt.Errorf("empty input is not allowed: %s", field)
	}

	ff := FacetField{
		path: nestedPath,
	}

	// ^ prefix means that the facet has ascending sort-order.
	if strings.HasPrefix(field, "^") {
		ff.sortAsc = true
		field = strings.TrimPrefix(field, "^")
	}

	// fields with @ suffix are sorted by its key instead of their frequency.
	if strings.HasSuffix(field, "@") {
		ff.orderByKey = true
		field = strings.TrimSuffix(field, "@")
	}

	// ~ is followed by the number of facet entries returned
	if strings.Contains(field, "~") {
		parts := strings.Split(field, "~")
		field = parts[0]
		if len(parts) == 2 && parts[1] != "" {
			size, err := strconv.Atoi(parts[1])
			if err != nil {
				// not a valid integer returning an error
				return nil, err
			}
			ff.size = size
		}
	}

	ff.Field = field

	// set internal query paths and aggregationType
	switch {
	case strings.HasPrefix(ff.Field, "meta."):
		ff.path = ff.Field
	case strings.HasPrefix(ff.Field, "tree."):
		ff.path = ff.Field
	case strings.HasPrefix(ff.Field, "id."):
		ff.nestedField = resourceField
		ff.Field = strings.TrimPrefix(ff.Field, "id.")
	case strings.HasPrefix(ff.Field, "datehistogram."):
		ff.nestedField = dateField
		ff.Field = strings.TrimPrefix(ff.Field, "datehistogram.")
		ff.aggregationType = "datehistogram"
	case strings.HasPrefix(ff.Field, "dateminmax."):
		ff.nestedField = dateField
		ff.Field = strings.TrimPrefix(ff.Field, "dateminmax.")
		ff.aggregationType = "dateminmax"
	case strings.HasPrefix(ff.Field, "tag."):
		ff.nestedField = tagField
		ff.Field = strings.TrimPrefix(ff.Field, "tag.")
	case ff.Field == tagField:
		// special case that should not use searchLabel query in aggregation
		ff.nestedField = tagField
	default:
		ff.path = nestedPath
		ff.nestedField = literalField
	}

	if ff.Field == "" {
		return nil, fmt.Errorf("empty input is not allowed: %s", field)
	}

	return &ff, nil
}
