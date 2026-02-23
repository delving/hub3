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
