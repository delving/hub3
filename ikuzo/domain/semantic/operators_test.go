package semantic

import (
	"testing"
)

func TestOperator_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		operator Operator
		want     bool
	}{
		{"valid eq", OpEqual, true},
		{"valid in", OpIn, true},
		{"valid contains", OpContains, true},
		{"valid bbox", OpBBox, true},
		{"invalid operator", Operator("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.operator.IsValid(); got != tt.want {
				t.Errorf("Operator.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOperator_IsComparison(t *testing.T) {
	tests := []struct {
		name     string
		operator Operator
		want     bool
	}{
		{"eq is comparison", OpEqual, true},
		{"gt is comparison", OpGreaterThan, true},
		{"contains is not comparison", OpContains, false},
		{"bbox is not comparison", OpBBox, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.operator.IsComparison(); got != tt.want {
				t.Errorf("Operator.IsComparison() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOperator_IsGeospatial(t *testing.T) {
	tests := []struct {
		name     string
		operator Operator
		want     bool
	}{
		{"bbox is geospatial", OpBBox, true},
		{"within is geospatial", OpWithin, true},
		{"eq is not geospatial", OpEqual, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.operator.IsGeospatial(); got != tt.want {
				t.Errorf("Operator.IsGeospatial() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOperator_RequiresValue(t *testing.T) {
	tests := []struct {
		name     string
		operator Operator
		want     bool
	}{
		{"eq requires value", OpEqual, true},
		{"exists does not require value", OpExists, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.operator.RequiresValue(); got != tt.want {
				t.Errorf("Operator.RequiresValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOperator_Validate(t *testing.T) {
	tests := []struct {
		name     string
		operator Operator
		value    interface{}
		wantErr  bool
	}{
		{"eq with value is valid", OpEqual, "test", false},
		{"eq without value is invalid", OpEqual, nil, true},
		{"exists with value is invalid", OpExists, "test", true},
		{"exists without value is valid", OpExists, nil, false},
		{"invalid operator", Operator("invalid"), "test", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.operator.Validate(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Operator.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
