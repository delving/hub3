package search

import "testing"

func TestTermVector_Size(t *testing.T) {

	type fields struct {
		Positions map[int]bool
	}

	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			"empty vector",
			fields{map[int]bool{}},
			0,
		},
		{
			"empty vector",
			fields{map[int]bool{1: true}},
			1,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			tv := NewTermVector()
			tv.Positions = tt.fields.Positions

			if got := tv.Size(); got != tt.want {
				t.Errorf("TermVector.Size() = %v, want %v", got, tt.want)
			}
		})
	}
}
