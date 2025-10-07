package semantic

import (
	"encoding/json"
	"testing"
)

func TestPropertyFilter_Validate(t *testing.T) {
	tests := []struct {
		name    string
		filter  *PropertyFilter
		wantErr bool
	}{
		{
			name: "valid filter",
			filter: &PropertyFilter{
				FieldName:    "title",
				OperatorType: OpEqual,
				Value:        "test",
			},
			wantErr: false,
		},
		{
			name: "missing field",
			filter: &PropertyFilter{
				FieldName:    "",
				OperatorType: OpEqual,
				Value:        "test",
			},
			wantErr: true,
		},
		{
			name: "invalid operator",
			filter: &PropertyFilter{
				FieldName:    "title",
				OperatorType: Operator("invalid"),
				Value:        "test",
			},
			wantErr: true,
		},
		{
			name: "eq without value",
			filter: &PropertyFilter{
				FieldName:    "title",
				OperatorType: OpEqual,
				Value:        nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.filter.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("PropertyFilter.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPropertyFilter_JSON(t *testing.T) {
	filter := &PropertyFilter{
		FieldName:    "title",
		OperatorType: OpEqual,
		Value:        "test",
	}

	data, err := json.Marshal(filter)
	if err != nil {
		t.Fatalf("failed to marshal filter: %v", err)
	}

	var decoded PropertyFilter
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal filter: %v", err)
	}

	if decoded.FieldName != filter.FieldName {
		t.Errorf("FieldName = %v, want %v", decoded.FieldName, filter.FieldName)
	}
	if decoded.OperatorType != filter.OperatorType {
		t.Errorf("OperatorType = %v, want %v", decoded.OperatorType, filter.OperatorType)
	}
}

func TestGeoBounds_Validate(t *testing.T) {
	tests := []struct {
		name    string
		bounds  *GeoBounds
		wantErr bool
	}{
		{
			name: "valid bounds",
			bounds: &GeoBounds{
				West:  4.8,
				South: 52.3,
				East:  4.9,
				North: 52.4,
			},
			wantErr: false,
		},
		{
			name: "west >= east",
			bounds: &GeoBounds{
				West:  4.9,
				South: 52.3,
				East:  4.8,
				North: 52.4,
			},
			wantErr: true,
		},
		{
			name: "south >= north",
			bounds: &GeoBounds{
				West:  4.8,
				South: 52.4,
				East:  4.9,
				North: 52.3,
			},
			wantErr: true,
		},
		{
			name: "invalid longitude",
			bounds: &GeoBounds{
				West:  -181,
				South: 52.3,
				East:  4.9,
				North: 52.4,
			},
			wantErr: true,
		},
		{
			name: "invalid latitude",
			bounds: &GeoBounds{
				West:  4.8,
				South: -91,
				East:  4.9,
				North: 52.4,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.bounds.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("GeoBounds.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGeoDistanceFilter_Validate(t *testing.T) {
	tests := []struct {
		name    string
		filter  *GeoDistanceFilter
		wantErr bool
	}{
		{
			name: "valid filter",
			filter: &GeoDistanceFilter{
				FieldName: "spatialCoverage",
				Point:     &GeoPoint{Lat: 52.37, Lon: 4.89},
				Distance:  "10km",
			},
			wantErr: false,
		},
		{
			name: "missing point",
			filter: &GeoDistanceFilter{
				FieldName: "spatialCoverage",
				Point:     nil,
				Distance:  "10km",
			},
			wantErr: true,
		},
		{
			name: "missing distance",
			filter: &GeoDistanceFilter{
				FieldName: "spatialCoverage",
				Point:     &GeoPoint{Lat: 52.37, Lon: 4.89},
				Distance:  "",
			},
			wantErr: true,
		},
		{
			name: "invalid latitude",
			filter: &GeoDistanceFilter{
				FieldName: "spatialCoverage",
				Point:     &GeoPoint{Lat: 91, Lon: 4.89},
				Distance:  "10km",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.filter.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("GeoDistanceFilter.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
