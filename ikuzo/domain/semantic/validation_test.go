package semantic

import (
	"testing"
)

func TestValidateGeoBounds(t *testing.T) {
	tests := []struct {
		name    string
		bounds  *GeoBounds
		wantErr bool
	}{
		{
			name: "valid bounds",
			bounds: &GeoBounds{
				North: 52.5,
				South: 52.3,
				West:  4.8,
				East:  4.9,
			},
			wantErr: false,
		},
		{
			name:    "nil bounds",
			bounds:  nil,
			wantErr: true,
		},
		{
			name: "north out of range",
			bounds: &GeoBounds{
				North: 91.0,
				South: 52.3,
				West:  4.8,
				East:  4.9,
			},
			wantErr: true,
		},
		{
			name: "south out of range",
			bounds: &GeoBounds{
				North: 52.5,
				South: -91.0,
				West:  4.8,
				East:  4.9,
			},
			wantErr: true,
		},
		{
			name: "west out of range",
			bounds: &GeoBounds{
				North: 52.5,
				South: 52.3,
				West:  -181.0,
				East:  4.9,
			},
			wantErr: true,
		},
		{
			name: "east out of range",
			bounds: &GeoBounds{
				North: 52.5,
				South: 52.3,
				West:  4.8,
				East:  181.0,
			},
			wantErr: true,
		},
		{
			name: "north less than south",
			bounds: &GeoBounds{
				North: 52.3,
				South: 52.5,
				West:  4.8,
				East:  4.9,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGeoBounds(tt.bounds)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGeoBounds() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateGeoPoint(t *testing.T) {
	tests := []struct {
		name    string
		point   *GeoPoint
		wantErr bool
	}{
		{
			name:    "valid point",
			point:   &GeoPoint{Lat: 52.5, Lon: 4.8},
			wantErr: false,
		},
		{
			name:    "nil point",
			point:   nil,
			wantErr: true,
		},
		{
			name:    "lat out of range high",
			point:   &GeoPoint{Lat: 91.0, Lon: 4.8},
			wantErr: true,
		},
		{
			name:    "lat out of range low",
			point:   &GeoPoint{Lat: -91.0, Lon: 4.8},
			wantErr: true,
		},
		{
			name:    "lon out of range high",
			point:   &GeoPoint{Lat: 52.5, Lon: 181.0},
			wantErr: true,
		},
		{
			name:    "lon out of range low",
			point:   &GeoPoint{Lat: 52.5, Lon: -181.0},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGeoPoint(tt.point)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGeoPoint() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateGeoDistance(t *testing.T) {
	tests := []struct {
		name     string
		distance string
		wantErr  bool
	}{
		{
			name:     "valid km",
			distance: "10km",
			wantErr:  false,
		},
		{
			name:     "valid miles",
			distance: "5mi",
			wantErr:  false,
		},
		{
			name:     "valid meters",
			distance: "100m",
			wantErr:  false,
		},
		{
			name:     "empty string",
			distance: "",
			wantErr:  true,
		},
		{
			name:     "no unit",
			distance: "10",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGeoDistance(tt.distance)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGeoDistance() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateGeoPolygon(t *testing.T) {
	tests := []struct {
		name    string
		polygon *GeoPolygon
		wantErr bool
	}{
		{
			name: "valid polygon",
			polygon: &GeoPolygon{
				Type: "Polygon",
				Coordinates: [][][]float64{
					{
						{4.8, 52.3},
						{4.9, 52.3},
						{4.9, 52.4},
						{4.8, 52.4},
						{4.8, 52.3}, // Closing point
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "nil polygon",
			polygon: nil,
			wantErr: true,
		},
		{
			name: "no coordinates",
			polygon: &GeoPolygon{
				Type:        "Polygon",
				Coordinates: [][][]float64{},
			},
			wantErr: true,
		},
		{
			name: "too few points",
			polygon: &GeoPolygon{
				Type: "Polygon",
				Coordinates: [][][]float64{
					{
						{4.8, 52.3},
						{4.9, 52.3},
						{4.8, 52.3},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "not closed",
			polygon: &GeoPolygon{
				Type: "Polygon",
				Coordinates: [][][]float64{
					{
						{4.8, 52.3},
						{4.9, 52.3},
						{4.9, 52.4},
						{4.8, 52.4},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "lat out of range",
			polygon: &GeoPolygon{
				Type: "Polygon",
				Coordinates: [][][]float64{
					{
						{4.8, 91.0},
						{4.9, 52.3},
						{4.9, 52.4},
						{4.8, 52.4},
						{4.8, 91.0},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "lon out of range",
			polygon: &GeoPolygon{
				Type: "Polygon",
				Coordinates: [][][]float64{
					{
						{181.0, 52.3},
						{4.9, 52.3},
						{4.9, 52.4},
						{4.8, 52.4},
						{181.0, 52.3},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGeoPolygon(tt.polygon)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGeoPolygon() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateDateString(t *testing.T) {
	tests := []struct {
		name    string
		date    string
		wantErr bool
	}{
		{
			name:    "valid RFC3339",
			date:    "2024-01-15T10:30:00Z",
			wantErr: false,
		},
		{
			name:    "valid ISO8601",
			date:    "2024-01-15T10:30:00",
			wantErr: false,
		},
		{
			name:    "valid date only",
			date:    "2024-01-15",
			wantErr: false,
		},
		{
			name:    "empty string",
			date:    "",
			wantErr: true,
		},
		{
			name:    "invalid format",
			date:    "15/01/2024",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDateString(tt.date)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDateString() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePagination(t *testing.T) {
	tests := []struct {
		name       string
		pagination *Pagination
		maxSize    int
		wantErr    bool
	}{
		{
			name:       "valid pagination",
			pagination: &Pagination{Page: 1, Size: 20},
			maxSize:    100,
			wantErr:    false,
		},
		{
			name:       "nil pagination",
			pagination: nil,
			maxSize:    100,
			wantErr:    false,
		},
		{
			name:       "page less than 1",
			pagination: &Pagination{Page: 0, Size: 20},
			maxSize:    100,
			wantErr:    true,
		},
		{
			name:       "size less than 1",
			pagination: &Pagination{Page: 1, Size: 0},
			maxSize:    100,
			wantErr:    true,
		},
		{
			name:       "size exceeds max",
			pagination: &Pagination{Page: 1, Size: 200},
			maxSize:    100,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePagination(tt.pagination, tt.maxSize)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePagination() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateSortField(t *testing.T) {
	allowedFields := []SortConfig{
		{Field: "title", PropertyURI: "http://schema.org/name"},
		{Field: "dateCreated", PropertyURI: "http://schema.org/dateCreated"},
	}

	tests := []struct {
		name    string
		field   string
		wantErr bool
	}{
		{
			name:    "valid field",
			field:   "title",
			wantErr: false,
		},
		{
			name:    "valid field 2",
			field:   "dateCreated",
			wantErr: false,
		},
		{
			name:    "invalid field",
			field:   "unknown",
			wantErr: true,
		},
		{
			name:    "empty field",
			field:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSortField(tt.field, allowedFields)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSortField() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateTextQuery(t *testing.T) {
	tests := []struct {
		name      string
		query     *TextQuery
		maxLength int
		wantErr   bool
	}{
		{
			name:      "valid query",
			query:     &TextQuery{Value: "test"},
			maxLength: 100,
			wantErr:   false,
		},
		{
			name:      "nil query",
			query:     nil,
			maxLength: 100,
			wantErr:   false,
		},
		{
			name:      "empty query",
			query:     &TextQuery{Value: ""},
			maxLength: 100,
			wantErr:   true,
		},
		{
			name:      "query too long",
			query:     &TextQuery{Value: "this is a very long query"},
			maxLength: 10,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTextQuery(tt.query, tt.maxLength)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTextQuery() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateFacetRequest(t *testing.T) {
	config := &FacetConfig{
		Field:       "dateCreated",
		FacetType:   "date",
		PropertyURI: "http://schema.org/dateCreated",
	}

	tests := []struct {
		name    string
		facet   FacetRequest
		wantErr bool
	}{
		{
			name:    "valid facet",
			facet:   FacetRequest{Field: "dateCreated", Limit: 10},
			wantErr: false,
		},
		{
			name:    "empty field",
			facet:   FacetRequest{Field: "", Limit: 10},
			wantErr: true,
		},
		{
			name:    "negative limit",
			facet:   FacetRequest{Field: "dateCreated", Limit: -1},
			wantErr: true,
		},
		{
			name:    "valid sort",
			facet:   FacetRequest{Field: "dateCreated", Limit: 10, Sort: "count"},
			wantErr: false,
		},
		{
			name:    "invalid sort",
			facet:   FacetRequest{Field: "dateCreated", Limit: 10, Sort: "invalid"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFacetRequest(tt.facet, config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFacetRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateFilterCount(t *testing.T) {
	filters := []Filter{
		&PropertyFilter{FieldName: "creator", OperatorType: OpEqual, Value: "test"},
		&PropertyFilter{FieldName: "type", OperatorType: OpEqual, Value: "test"},
	}

	tests := []struct {
		name       string
		filters    []Filter
		maxFilters int
		wantErr    bool
	}{
		{
			name:       "within limit",
			filters:    filters,
			maxFilters: 10,
			wantErr:    false,
		},
		{
			name:       "no limit",
			filters:    filters,
			maxFilters: 0,
			wantErr:    false,
		},
		{
			name:       "exceeds limit",
			filters:    filters,
			maxFilters: 1,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFilterCount(tt.filters, tt.maxFilters)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFilterCount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateFacetCount(t *testing.T) {
	facets := []FacetRequest{
		{Field: "type", Limit: 10},
		{Field: "creator", Limit: 10},
	}

	tests := []struct {
		name      string
		facets    []FacetRequest
		maxFacets int
		wantErr   bool
	}{
		{
			name:      "within limit",
			facets:    facets,
			maxFacets: 10,
			wantErr:   false,
		},
		{
			name:      "no limit",
			facets:    facets,
			maxFacets: 0,
			wantErr:   false,
		},
		{
			name:      "exceeds limit",
			facets:    facets,
			maxFacets: 1,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFacetCount(tt.facets, tt.maxFacets)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFacetCount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
