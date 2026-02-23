package semantic

import (
	"testing"
)

func TestCollapseOptions_Validate(t *testing.T) {
	tests := []struct {
		name    string
		opts    *CollapseOptions
		wantErr bool
	}{
		{
			name:    "nil is valid (no collapse)",
			opts:    nil,
			wantErr: false,
		},
		{
			name:    "valid collapse",
			opts:    &CollapseOptions{Field: "edm:dataProvider"},
			wantErr: false,
		},
		{
			name:    "valid with size",
			opts:    &CollapseOptions{Field: "edm:dataProvider", Size: 3},
			wantErr: false,
		},
		{
			name:    "empty field is invalid",
			opts:    &CollapseOptions{Field: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("CollapseOptions.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestQueryOptions_HasCollapse(t *testing.T) {
	opts := &QueryOptions{}
	if opts.Collapse != nil {
		t.Error("expected nil Collapse on new QueryOptions")
	}

	opts.Collapse = &CollapseOptions{Field: "edm:dataProvider"}
	if opts.Collapse == nil {
		t.Error("expected non-nil Collapse")
	}
	if opts.Collapse.Field != "edm:dataProvider" {
		t.Errorf("got field %q, want %q", opts.Collapse.Field, "edm:dataProvider")
	}
}

func TestFacetBoolType_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		fbt   FacetBoolType
		valid bool
	}{
		{"or", FacetBoolOr, true},
		{"and", FacetBoolAnd, true},
		{"empty defaults to or", "", true},
		{"invalid", FacetBoolType("xor"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fbt.IsValid(); got != tt.valid {
				t.Errorf("FacetBoolType(%q).IsValid() = %v, want %v", tt.fbt, got, tt.valid)
			}
		})
	}
}

func TestQueryOptions_NewFields(t *testing.T) {
	opts := &QueryOptions{
		FacetBoolType: FacetBoolAnd,
		Peek:          true,
		Debug:         "query",
	}

	if opts.FacetBoolType != FacetBoolAnd {
		t.Errorf("FacetBoolType = %q, want %q", opts.FacetBoolType, FacetBoolAnd)
	}
	if !opts.Peek {
		t.Error("Peek should be true")
	}
	if opts.Debug != "query" {
		t.Errorf("Debug = %q, want %q", opts.Debug, "query")
	}
}

func TestFacetRequest_Cursor(t *testing.T) {
	fr := FacetRequest{
		Field:  "dc:creator",
		Limit:  50,
		Cursor: "abc123",
	}
	if fr.Cursor != "abc123" {
		t.Errorf("FacetRequest.Cursor = %q, want %q", fr.Cursor, "abc123")
	}
}

func TestPropertyFilter_Hidden(t *testing.T) {
	pf := &PropertyFilter{
		FieldName:    "orgID",
		OperatorType: OpEqual,
		Value:        "museum-x",
		Hidden:       true,
	}
	if !pf.Hidden {
		t.Error("PropertyFilter.Hidden should be true")
	}
}
