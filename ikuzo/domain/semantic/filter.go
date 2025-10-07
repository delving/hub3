package semantic

import (
	"encoding/json"
	"fmt"
)

// Filter represents a query filter.
// Filters are used to narrow down search results based on field values.
type Filter interface {
	// Type returns the @type of the filter for JSON-LD
	Type() string

	// Field returns the field name this filter applies to
	Field() string

	// Operator returns the operator used by this filter
	Operator() Operator

	// Validate checks if the filter is valid
	Validate() error

	// MarshalJSON enables JSON serialization
	MarshalJSON() ([]byte, error)
}

// PropertyFilter filters by property value with a specific operator.
type PropertyFilter struct {
	TypeValue    string      `json:"@type"`
	FieldName    string      `json:"field"`
	OperatorType Operator    `json:"operator"`
	Value        interface{} `json:"value"`
}

// Type returns the @type.
func (pf *PropertyFilter) Type() string {
	if pf.TypeValue == "" {
		return "hub3:PropertyFilter"
	}
	return pf.TypeValue
}

// Field returns the field name.
func (pf *PropertyFilter) Field() string {
	return pf.FieldName
}

// Operator returns the operator.
func (pf *PropertyFilter) Operator() Operator {
	return pf.OperatorType
}

// Validate checks if the filter is valid.
func (pf *PropertyFilter) Validate() error {
	if pf.FieldName == "" {
		return fmt.Errorf("filter field cannot be empty")
	}

	if !pf.OperatorType.IsValid() {
		return fmt.Errorf("invalid operator: %s", pf.OperatorType)
	}

	if err := pf.OperatorType.Validate(pf.Value); err != nil {
		return fmt.Errorf("filter validation failed: %w", err)
	}

	return nil
}

// MarshalJSON implements custom JSON marshaling.
func (pf *PropertyFilter) MarshalJSON() ([]byte, error) {
	type Alias PropertyFilter
	return json.Marshal(&struct {
		Type string `json:"@type"`
		*Alias
	}{
		Type:  pf.Type(),
		Alias: (*Alias)(pf),
	})
}

// RangeFilter filters by numeric or date ranges.
type RangeFilter struct {
	TypeValue    string      `json:"@type,omitempty"`
	FieldName    string      `json:"field"`
	OperatorType Operator    `json:"operator"` // gt, gte, lt, lte, or "between"
	Min          interface{} `json:"min,omitempty"`
	Max          interface{} `json:"max,omitempty"`
}

// Type returns the @type.
func (rf *RangeFilter) Type() string {
	if rf.TypeValue == "" {
		return "hub3:RangeFilter"
	}
	return rf.TypeValue
}

// Field returns the field name.
func (rf *RangeFilter) Field() string {
	return rf.FieldName
}

// Operator returns the operator.
func (rf *RangeFilter) Operator() Operator {
	return rf.OperatorType
}

// Validate checks if the filter is valid.
func (rf *RangeFilter) Validate() error {
	if rf.FieldName == "" {
		return fmt.Errorf("filter field cannot be empty")
	}

	if !rf.OperatorType.IsValid() && rf.OperatorType != "between" {
		return fmt.Errorf("invalid range operator: %s", rf.OperatorType)
	}

	// For "between" operator, both min and max are required
	if rf.OperatorType == "between" {
		if rf.Min == nil || rf.Max == nil {
			return fmt.Errorf("between operator requires both min and max values")
		}
	}

	return nil
}

// MarshalJSON implements custom JSON marshaling.
func (rf *RangeFilter) MarshalJSON() ([]byte, error) {
	type Alias RangeFilter
	return json.Marshal(&struct {
		Type string `json:"@type"`
		*Alias
	}{
		Type:  rf.Type(),
		Alias: (*Alias)(rf),
	})
}

// ExistsFilter checks if a field has any value.
type ExistsFilter struct {
	TypeValue string `json:"@type,omitempty"`
	FieldName string `json:"field"`
}

// Type returns the @type.
func (ef *ExistsFilter) Type() string {
	if ef.TypeValue == "" {
		return "hub3:ExistsFilter"
	}
	return ef.TypeValue
}

// Field returns the field name.
func (ef *ExistsFilter) Field() string {
	return ef.FieldName
}

// Operator returns the exists operator.
func (ef *ExistsFilter) Operator() Operator {
	return OpExists
}

// Validate checks if the filter is valid.
func (ef *ExistsFilter) Validate() error {
	if ef.FieldName == "" {
		return fmt.Errorf("filter field cannot be empty")
	}
	return nil
}

// MarshalJSON implements custom JSON marshaling.
func (ef *ExistsFilter) MarshalJSON() ([]byte, error) {
	type Alias ExistsFilter
	return json.Marshal(&struct {
		Type string `json:"@type"`
		*Alias
	}{
		Type:  ef.Type(),
		Alias: (*Alias)(ef),
	})
}

// GeoPoint represents a geographic point.
type GeoPoint struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// GeoBounds represents a bounding box.
type GeoBounds struct {
	West  float64 `json:"west"`
	South float64 `json:"south"`
	East  float64 `json:"east"`
	North float64 `json:"north"`
}

// Validate checks if the bounds are valid.
func (gb *GeoBounds) Validate() error {
	if gb.West < -180 || gb.West > 180 {
		return fmt.Errorf("invalid west longitude: %f (must be -180 to 180)", gb.West)
	}
	if gb.East < -180 || gb.East > 180 {
		return fmt.Errorf("invalid east longitude: %f (must be -180 to 180)", gb.East)
	}
	if gb.South < -90 || gb.South > 90 {
		return fmt.Errorf("invalid south latitude: %f (must be -90 to 90)", gb.South)
	}
	if gb.North < -90 || gb.North > 90 {
		return fmt.Errorf("invalid north latitude: %f (must be -90 to 90)", gb.North)
	}
	if gb.West >= gb.East {
		return fmt.Errorf("west (%f) must be less than east (%f)", gb.West, gb.East)
	}
	if gb.South >= gb.North {
		return fmt.Errorf("south (%f) must be less than north (%f)", gb.South, gb.North)
	}
	return nil
}

// GeoPolygon represents a geographic polygon using GeoJSON format.
type GeoPolygon struct {
	Type        string        `json:"type"` // "Polygon"
	Coordinates [][][]float64 `json:"coordinates"`
}

// GeoBBoxFilter filters by bounding box.
type GeoBBoxFilter struct {
	TypeValue string     `json:"@type,omitempty"`
	FieldName string     `json:"field"`
	Bounds    *GeoBounds `json:"bounds"`
}

// Type returns the @type.
func (gbf *GeoBBoxFilter) Type() string {
	if gbf.TypeValue == "" {
		return "hub3:GeoBBoxFilter"
	}
	return gbf.TypeValue
}

// Field returns the field name.
func (gbf *GeoBBoxFilter) Field() string {
	return gbf.FieldName
}

// Operator returns the bbox operator.
func (gbf *GeoBBoxFilter) Operator() Operator {
	return OpBBox
}

// Validate checks if the filter is valid.
func (gbf *GeoBBoxFilter) Validate() error {
	if gbf.FieldName == "" {
		return fmt.Errorf("filter field cannot be empty")
	}
	if gbf.Bounds == nil {
		return fmt.Errorf("bbox filter requires bounds")
	}
	return gbf.Bounds.Validate()
}

// MarshalJSON implements custom JSON marshaling.
func (gbf *GeoBBoxFilter) MarshalJSON() ([]byte, error) {
	type Alias GeoBBoxFilter
	return json.Marshal(&struct {
		Type string `json:"@type"`
		*Alias
	}{
		Type:  gbf.Type(),
		Alias: (*Alias)(gbf),
	})
}

// GeoDistanceFilter filters by distance from a point.
type GeoDistanceFilter struct {
	TypeValue string    `json:"@type,omitempty"`
	FieldName string    `json:"field"`
	Point     *GeoPoint `json:"point"`
	Distance  string    `json:"distance"` // e.g., "10km", "5mi"
}

// Type returns the @type.
func (gdf *GeoDistanceFilter) Type() string {
	if gdf.TypeValue == "" {
		return "hub3:GeoDistanceFilter"
	}
	return gdf.TypeValue
}

// Field returns the field name.
func (gdf *GeoDistanceFilter) Field() string {
	return gdf.FieldName
}

// Operator returns the within operator.
func (gdf *GeoDistanceFilter) Operator() Operator {
	return OpWithin
}

// Validate checks if the filter is valid.
func (gdf *GeoDistanceFilter) Validate() error {
	if gdf.FieldName == "" {
		return fmt.Errorf("filter field cannot be empty")
	}
	if gdf.Point == nil {
		return fmt.Errorf("distance filter requires a point")
	}
	if gdf.Point.Lat < -90 || gdf.Point.Lat > 90 {
		return fmt.Errorf("invalid latitude: %f (must be -90 to 90)", gdf.Point.Lat)
	}
	if gdf.Point.Lon < -180 || gdf.Point.Lon > 180 {
		return fmt.Errorf("invalid longitude: %f (must be -180 to 180)", gdf.Point.Lon)
	}
	if gdf.Distance == "" {
		return fmt.Errorf("distance filter requires a distance value")
	}
	return nil
}

// MarshalJSON implements custom JSON marshaling.
func (gdf *GeoDistanceFilter) MarshalJSON() ([]byte, error) {
	type Alias GeoDistanceFilter
	return json.Marshal(&struct {
		Type string `json:"@type"`
		*Alias
	}{
		Type:  gdf.Type(),
		Alias: (*Alias)(gdf),
	})
}

// GeoPolygonFilter filters by polygon.
type GeoPolygonFilter struct {
	TypeValue string      `json:"@type,omitempty"`
	FieldName string      `json:"field"`
	Polygon   *GeoPolygon `json:"polygon"`
}

// Type returns the @type.
func (gpf *GeoPolygonFilter) Type() string {
	if gpf.TypeValue == "" {
		return "hub3:GeoPolygonFilter"
	}
	return gpf.TypeValue
}

// Field returns the field name.
func (gpf *GeoPolygonFilter) Field() string {
	return gpf.FieldName
}

// Operator returns the polygon operator.
func (gpf *GeoPolygonFilter) Operator() Operator {
	return OpPolygon
}

// Validate checks if the filter is valid.
func (gpf *GeoPolygonFilter) Validate() error {
	if gpf.FieldName == "" {
		return fmt.Errorf("filter field cannot be empty")
	}
	if gpf.Polygon == nil {
		return fmt.Errorf("polygon filter requires a polygon")
	}
	if gpf.Polygon.Type != "Polygon" {
		return fmt.Errorf("invalid polygon type: %s (must be 'Polygon')", gpf.Polygon.Type)
	}
	if len(gpf.Polygon.Coordinates) == 0 {
		return fmt.Errorf("polygon must have at least one ring")
	}
	return nil
}

// MarshalJSON implements custom JSON marshaling.
func (gpf *GeoPolygonFilter) MarshalJSON() ([]byte, error) {
	type Alias GeoPolygonFilter
	return json.Marshal(&struct {
		Type string `json:"@type"`
		*Alias
	}{
		Type:  gpf.Type(),
		Alias: (*Alias)(gpf),
	})
}
