package semantic

import (
	"fmt"
	"time"
)

// ValidateGeoBounds validates geographic bounding box coordinates.
func ValidateGeoBounds(bounds *GeoBounds) error {
	if bounds == nil {
		return NewValidationError("bounds", "bounds cannot be nil", nil)
	}

	// Validate latitude range (-90 to 90)
	if bounds.North < -90 || bounds.North > 90 {
		return NewValidationError("north", fmt.Sprintf("north latitude must be between -90 and 90, got %f", bounds.North), bounds.North)
	}
	if bounds.South < -90 || bounds.South > 90 {
		return NewValidationError("south", fmt.Sprintf("south latitude must be between -90 and 90, got %f", bounds.South), bounds.South)
	}

	// Validate longitude range (-180 to 180)
	if bounds.West < -180 || bounds.West > 180 {
		return NewValidationError("west", fmt.Sprintf("west longitude must be between -180 and 180, got %f", bounds.West), bounds.West)
	}
	if bounds.East < -180 || bounds.East > 180 {
		return NewValidationError("east", fmt.Sprintf("east longitude must be between -180 and 180, got %f", bounds.East), bounds.East)
	}

	// Validate that bounds make sense
	if bounds.North < bounds.South {
		return NewValidationError("bounds", fmt.Sprintf("north (%f) must be greater than south (%f)", bounds.North, bounds.South), bounds)
	}

	return nil
}

// ValidateGeoPoint validates a geographic point.
func ValidateGeoPoint(point *GeoPoint) error {
	if point == nil {
		return NewValidationError("point", "point cannot be nil", nil)
	}

	// Validate latitude range (-90 to 90)
	if point.Lat < -90 || point.Lat > 90 {
		return NewValidationError("lat", fmt.Sprintf("latitude must be between -90 and 90, got %f", point.Lat), point.Lat)
	}

	// Validate longitude range (-180 to 180)
	if point.Lon < -180 || point.Lon > 180 {
		return NewValidationError("lon", fmt.Sprintf("longitude must be between -180 and 180, got %f", point.Lon), point.Lon)
	}

	return nil
}

// ValidateGeoDistance validates a distance value.
func ValidateGeoDistance(distance string) error {
	if distance == "" {
		return NewValidationError("distance", "distance cannot be empty", distance)
	}

	// Distance should end with a unit (km, mi, m)
	validUnits := []string{"km", "mi", "m", "km", "miles", "meters"}
	hasValidUnit := false
	for _, unit := range validUnits {
		if len(distance) > len(unit) && distance[len(distance)-len(unit):] == unit {
			hasValidUnit = true
			break
		}
	}

	if !hasValidUnit {
		return NewValidationError("distance", fmt.Sprintf("distance must include a unit (km, mi, m), got: %s", distance), distance)
	}

	return nil
}

// ValidateGeoPolygon validates a GeoJSON polygon.
func ValidateGeoPolygon(polygon *GeoPolygon) error {
	if polygon == nil {
		return NewValidationError("polygon", "polygon cannot be nil", nil)
	}

	if len(polygon.Coordinates) == 0 {
		return NewValidationError("coordinates", "polygon must have at least one ring", nil)
	}

	// Validate outer ring
	ring := polygon.Coordinates[0]
	if len(ring) < 4 {
		return NewValidationError("coordinates", "polygon ring must have at least 4 points (including closing point)", nil)
	}

	// Validate that first and last points are the same (closed ring)
	first := ring[0]
	last := ring[len(ring)-1]
	if len(first) >= 2 && len(last) >= 2 {
		if first[0] != last[0] || first[1] != last[1] {
			return NewValidationError("coordinates", "polygon ring must be closed (first and last points must be identical)", nil)
		}
	}

	// Validate each coordinate
	for i, coord := range ring {
		if len(coord) < 2 {
			return NewValidationError("coordinates", fmt.Sprintf("coordinate at index %d must have at least 2 values", i), coord)
		}

		lon := coord[0]
		lat := coord[1]

		// Validate ranges
		if lat < -90 || lat > 90 {
			return NewValidationError("coordinates", fmt.Sprintf("latitude at index %d must be between -90 and 90, got %f", i, lat), lat)
		}
		if lon < -180 || lon > 180 {
			return NewValidationError("coordinates", fmt.Sprintf("longitude at index %d must be between -180 and 180, got %f", i, lon), lon)
		}
	}

	return nil
}

// ValidateDateString validates a date string in RFC3339 format.
func ValidateDateString(dateStr string) error {
	if dateStr == "" {
		return NewValidationError("date", "date cannot be empty", nil)
	}

	// Try parsing as RFC3339
	_, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		// Try ISO8601 without timezone
		_, err = time.Parse("2006-01-02T15:04:05", dateStr)
		if err != nil {
			// Try date only
			_, err = time.Parse("2006-01-02", dateStr)
			if err != nil {
				return NewValidationError("date", fmt.Sprintf("invalid date format, expected RFC3339 or ISO8601: %s", dateStr), dateStr)
			}
		}
	}

	return nil
}

// ValidatePagination validates pagination parameters.
func ValidatePagination(pagination *Pagination, maxSize int) error {
	if pagination == nil {
		return nil
	}

	if pagination.Page < 1 {
		return NewValidationError("page", fmt.Sprintf("page must be >= 1, got %d", pagination.Page), pagination.Page)
	}

	if pagination.Size < 1 {
		return NewValidationError("size", fmt.Sprintf("size must be >= 1, got %d", pagination.Size), pagination.Size)
	}

	if maxSize > 0 && pagination.Size > maxSize {
		return NewValidationError("size", fmt.Sprintf("size cannot exceed %d, got %d", maxSize, pagination.Size), pagination.Size)
	}

	return nil
}

// ValidateSortField validates a sort field name.
func ValidateSortField(field string, allowedFields []SortConfig) error {
	if field == "" {
		return NewValidationError("sort", "sort field cannot be empty", nil)
	}

	// Check if field is in allowed list
	for _, allowed := range allowedFields {
		if allowed.Field == field {
			return nil
		}
	}

	// Build list of allowed fields for error message
	fieldNames := make([]string, len(allowedFields))
	for i, f := range allowedFields {
		fieldNames[i] = f.Field
	}

	return NewValidationError("sort", fmt.Sprintf("sort field '%s' is not allowed, valid fields: %v", field, fieldNames), field)
}

// ValidateTextQuery validates a text query.
func ValidateTextQuery(query *TextQuery, maxLength int) error {
	if query == nil {
		return nil
	}

	if query.Value == "" {
		return NewValidationError("query", "query value cannot be empty", nil)
	}

	if maxLength > 0 && len(query.Value) > maxLength {
		return NewValidationError("query", fmt.Sprintf("query length cannot exceed %d characters, got %d", maxLength, len(query.Value)), query.Value)
	}

	return nil
}

// ValidateFacetRequest validates a facet request.
func ValidateFacetRequest(facet FacetRequest, config *FacetConfig) error {
	if facet.Field == "" {
		return NewValidationError("facet.field", "facet field cannot be empty", nil)
	}

	if facet.Limit < 0 {
		return NewValidationError("facet.limit", fmt.Sprintf("facet limit must be >= 0, got %d", facet.Limit), facet.Limit)
	}

	// Validate sort field if specified
	if facet.Sort != "" {
		validSorts := []string{"count", "value", "index"}
		valid := false
		for _, vs := range validSorts {
			if facet.Sort == vs {
				valid = true
				break
			}
		}
		if !valid {
			return NewValidationError("facet.sort", fmt.Sprintf("invalid sort value: %s, valid values: %v", facet.Sort, validSorts), facet.Sort)
		}
	}

	return nil
}

// ValidateFilterCount validates the number of filters in a query.
func ValidateFilterCount(filters []Filter, maxFilters int) error {
	if maxFilters > 0 && len(filters) > maxFilters {
		return NewValidationError("filters", fmt.Sprintf("cannot exceed %d filters, got %d", maxFilters, len(filters)), len(filters))
	}
	return nil
}

// ValidateFacetCount validates the number of facets in a query.
func ValidateFacetCount(facets []FacetRequest, maxFacets int) error {
	if maxFacets > 0 && len(facets) > maxFacets {
		return NewValidationError("facets", fmt.Sprintf("cannot exceed %d facets, got %d", maxFacets, len(facets)), len(facets))
	}
	return nil
}
