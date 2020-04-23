package search

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_newFacetField(t *testing.T) {
	type args struct {
		field string
	}
	tests := []struct {
		name    string
		args    args
		want    *FacetField
		wantErr bool
	}{
		{
			"simple field",
			args{field: "dc_title"},
			&FacetField{
				Field:       "dc_title",
				path:        nestedPath,
				nestedField: literalField,
			},
			false,
		},
		{
			"empty input should throw error",
			args{field: ""},
			nil,
			true,
		},
		{
			"meta field should not have nested path",
			args{field: "meta.spec"},
			&FacetField{
				Field: "meta.spec",
				path:  "meta.spec",
			},
			false,
		},
		{
			"tree field should not have nested path",
			args{field: "tree.depth"},
			&FacetField{
				Field: "tree.depth",
				path:  "tree.depth",
			},
			false,
		},
		{
			"when field is prefixed with ^ it should sort Ascending",
			args{field: "^tree.depth"},
			&FacetField{
				Field:   "tree.depth",
				path:    "tree.depth",
				sortAsc: true,
			},
			false,
		},
		{
			"when field is prefixed with id. it should use the @id nested field",
			args{field: "id.dc_subject"},
			&FacetField{
				Field:       "dc_subject",
				path:        "resources.entries",
				nestedField: resourceField,
			},
			false,
		},
		{
			"when field is prefixed with datehistogram. it should use the 'date' nested field",
			args{field: "datehistogram.dc_date"},
			&FacetField{
				Field:           "dc_date",
				path:            "resources.entries",
				nestedField:     dateField,
				aggregationType: "datehistogram",
			},
			false,
		},
		{
			"when field is prefixed with dateminmax. it should use the 'date' nested field",
			args{field: "dateminmax.dc_date"},
			&FacetField{
				Field:           "dc_date",
				path:            "resources.entries",
				nestedField:     dateField,
				aggregationType: "dateminmax",
			},
			false,
		},
		{
			"when field is prefixed with tags it should use the 'tag' nested field",
			args{field: "tag.dc_date"},
			&FacetField{
				Field:       "dc_date",
				path:        "resources.entries",
				nestedField: tagField,
			},
			false,
		},
		{
			"when field is identical to 'tags' this is the nested field",
			args{field: "tags"},
			&FacetField{
				Field:       "tags",
				path:        "resources.entries",
				nestedField: tagField,
			},
			false,
		},

		{
			"when field contains a tilde `~` it is followed by the bucket size",
			args{field: "dc_date~10"},
			&FacetField{
				Field:       "dc_date",
				path:        "resources.entries",
				nestedField: literalField,
				size:        10,
			},
			false,
		},
		{
			"empty value after `~` should be ignored",
			args{field: "dc_date~"},
			&FacetField{
				Field:       "dc_date",
				path:        "resources.entries",
				nestedField: literalField,
				size:        0,
			},
			false,
		},
		{
			"non integer value after `~` should be throw an error",
			args{field: "dc_date~abc"},
			nil,
			true,
		},
		{
			"combining multiple qualifiers",
			args{field: "^id.dc_date~10"},
			&FacetField{
				Field:       "dc_date",
				path:        "resources.entries",
				nestedField: resourceField,
				size:        10,
				sortAsc:     true,
			},
			false,
		},
		{
			"when suffixed by `@` sort by Value instead of frequency count",
			args{field: "^id.dc_date~10@"},
			&FacetField{
				Field:       "dc_date",
				path:        "resources.entries",
				nestedField: resourceField,
				size:        10,
				sortAsc:     true,
				orderByKey:  true,
			},
			false,
		},
		{
			"^@ is not a valid facet field",
			args{field: "^@"},
			nil,
			true,
		},
		{
			"empty Field is not allowed",
			args{field: "id."},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newFacetField(tt.args.field)
			if (err != nil) != tt.wantErr {
				t.Errorf("newFacetField() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			opt := cmp.AllowUnexported(FacetField{})
			if diff := cmp.Diff(tt.want, got, opt); diff != "" {
				t.Errorf("newFacetField() %s mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}
