package search

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestVector(t *testing.T) {
	type locations struct {
		doc int
		pos int
	}

	type fields struct {
		locations []locations
	}

	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			"empty vector",
			fields{locations: []locations{}},
			0,
		},
		{
			"non-empty vector",
			fields{[]locations{
				{0, 1},
			}},
			1,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			tv := NewVectors()
			for _, loc := range tt.fields.locations {
				tv.Add(loc.doc, loc.pos)
			}

			if got := tv.Size(); got != tt.want {
				t.Errorf("Vector.TermCount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVectors_HasVector(t *testing.T) {
	type fields struct {
		vectors []Vector
	}

	type args struct {
		vector Vector
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			"empty vectors",
			fields{[]Vector{}},
			args{Vector{DocID: 10, Location: 3}},
			false,
		},
		{
			"non-empty vectors",
			fields{[]Vector{{DocID: 10, Location: 3}}},
			args{Vector{DocID: 10, Location: 3}},
			true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			tv := NewVectors()

			for _, v := range tt.fields.vectors {
				tv.AddVector(v)
			}

			if got := tv.HasVector(tt.args.vector); got != tt.want {
				t.Errorf("Vectors.HasVector() = %v, want %v", got, tt.want)
			}

			if got := tv.HasDoc(tt.args.vector.DocID); got != tt.want {
				t.Errorf("Vectors.HasDoc() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVectors_Merge(t *testing.T) {
	type fields struct {
		vectors []Vector
	}

	type args struct {
		vectors []Vector
	}

	tests := []struct {
		name         string
		fields       fields
		args         args
		want         *Vectors
		wantDocCount int
		wantSize     int
	}{
		{
			"empty source",
			fields{},
			args{[]Vector{{DocID: 1, Location: 1}}},
			&Vectors{
				Locations: map[Vector]bool{{1, 1}: true},
				Docs:      map[int]bool{1: true},
			},
			1,
			1,
		},
		{
			"empty target",
			fields{[]Vector{{DocID: 1, Location: 1}}},
			args{},
			&Vectors{
				Locations: map[Vector]bool{{1, 1}: true},
				Docs:      map[int]bool{1: true},
			},
			1,
			1,
		},
		{
			"empty target",
			fields{[]Vector{
				{DocID: 1, Location: 1},
				{DocID: 3, Location: 2},
			}},
			args{[]Vector{
				{DocID: 1, Location: 1},
				{DocID: 1, Location: 2},
				{DocID: 2, Location: 1},
			}},
			&Vectors{
				Locations: map[Vector]bool{
					{1, 1}: true,
					{1, 2}: true,
					{2, 1}: true,
					{3, 2}: true,
				},
				Docs: map[int]bool{
					1: true,
					2: true,
					3: true,
				},
			},
			3,
			4,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			source := NewVectors()
			for _, v := range tt.fields.vectors {
				source.AddVector(v)
			}

			target := NewVectors()
			for _, v := range tt.args.vectors {
				target.AddVector(v)
			}

			source.Merge(target)

			if diff := cmp.Diff(tt.want, source); diff != "" {
				t.Errorf("Vectors.Merge() %s = mismatch (-want +got):\n%s", tt.name, diff)
			}

			if got := source.Size(); got != tt.wantSize {
				t.Errorf("Vectors.Size() = %v, want %v", got, tt.wantSize)
			}

			if got := source.DocCount(); got != tt.wantDocCount {
				t.Errorf("Vectors.HasDocCount() = %v, want %v", got, tt.wantDocCount)
			}
		})
	}
}
