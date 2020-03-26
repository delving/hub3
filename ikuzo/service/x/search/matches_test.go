package search

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

type testVector struct {
	term    string
	vectors []Vector
}

func createMatches(vectors []testVector) *Matches {
	matches := NewMatches()

	for _, v := range vectors {
		tv := NewVectors()

		for _, vector := range v.vectors {
			tv.AddVector(vector)
		}

		matches.AppendTerm(v.term, tv)
	}

	return matches
}

// nolint:funlen
func TestMatches_Merge(t *testing.T) {
	type fields struct {
		vectors []testVector
	}

	type args struct {
		vectors []testVector
	}

	tests := []struct {
		name          string
		fields        fields
		args          args
		want          []testVector
		wantTotal     int
		wantTermCount int
		wantDocCount  int
	}{
		{
			"source only merge",
			fields{
				[]testVector{{"word", []Vector{{1, 1}}}},
			},
			args{
				[]testVector{},
			},
			[]testVector{{"word", []Vector{{1, 1}}}},
			1,
			1,
			1,
		},
		{
			"source only merge empty vector",
			fields{
				[]testVector{{"word", []Vector{{1, 1}}}},
			},
			args{
				[]testVector{{"word", []Vector{}}},
			},
			[]testVector{{"word", []Vector{{1, 1}}}},
			1,
			1,
			1,
		},
		{
			"partial merge",
			fields{
				[]testVector{
					{"word", []Vector{{1, 1}}},
				},
			},
			args{
				[]testVector{
					{"word", []Vector{{2, 1}}},
					{"words", []Vector{{3, 1}, {3, 2}}},
				},
			},
			[]testVector{
				{"word", []Vector{{1, 1}, {2, 1}}},
				{"words", []Vector{{3, 1}, {3, 2}}},
			},
			4,
			2,
			3,
		},
		{
			"target only merge",
			fields{
				[]testVector{},
			},
			args{
				[]testVector{
					{"word", []Vector{{2, 1}}},
					{"words", []Vector{{3, 1}, {3, 2}}},
				},
			},
			[]testVector{
				{"word", []Vector{{2, 1}}},
				{"words", []Vector{{3, 1}, {3, 2}}},
			},
			3,
			2,
			2,
		},
		{
			"idempotent merge merge",
			fields{
				[]testVector{{"word", []Vector{{1, 1}}}},
			},
			args{
				[]testVector{{"words", []Vector{{2, 1}}}},
			},
			[]testVector{
				{"word", []Vector{{1, 1}}},
				{"words", []Vector{{2, 1}}},
			},
			2,
			2,
			2,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			source := createMatches(tt.fields.vectors)

			target := createMatches(tt.args.vectors)

			source.Merge(target)

			want := createMatches(tt.want)

			if diff := cmp.Diff(want, source, cmp.AllowUnexported(Matches{}, Vectors{})); diff != "" {
				t.Errorf("Matches.Merge() %s = mismatch (-want +got):\n%s", tt.name, diff)
			}

			if got := source.Total(); got != tt.wantTotal {
				t.Errorf("Matches.HasTotal() %s = %v, want %v", tt.name, got, tt.wantTotal)
			}

			if got := source.TermCount(); got != tt.wantTermCount {
				t.Errorf("Matches.TermCount() %s = %v, want %v", tt.name, got, tt.wantTermCount)
			}

			if got := source.DocCount(); got != tt.wantDocCount {
				t.Errorf("Matches.DocCount() %s = %v, want %v", tt.name, got, tt.wantDocCount)
			}

			if source.Vectors().Size() != source.termVectors.Size() {
				t.Errorf("Matches.Vectors() length = %d, want %d", source.Vectors().Size(), source.termVectors.Size())
			}
		})
	}
}

func TestMatches_Total(t *testing.T) {
	type fields struct {
		hits map[string]int
	}

	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			"no results",
			fields{hits: map[string]int{}},
			0,
		},
		{
			"some results",
			fields{hits: map[string]int{
				"one":       1,
				"two times": 2,
				"many":      10,
			}},
			13,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			sh := &Matches{
				termFrequency: tt.fields.hits,
			}
			if got := sh.Total(); got != tt.want {
				t.Errorf("Matches.Total() = %v, want %v", got, tt.want)
			}

			if diff := cmp.Diff(sh.termFrequency, sh.TermFrequency(), cmp.AllowUnexported(Matches{})); diff != "" {
				t.Errorf("Matches.Total() %s = mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}

func TestMatches_HasDocID(t *testing.T) {
	type fields struct {
		vectors []testVector
	}

	type args struct {
		docID int
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			"no match",
			fields{vectors: []testVector{
				{"word", []Vector{{1, 1}}}}},
			args{docID: 10},
			false,
		},
		{
			"match",
			fields{vectors: []testVector{
				{"word", []Vector{{1, 1}}}}},
			args{docID: 1},
			true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			m := createMatches(tt.fields.vectors)

			if got := m.HasDocID(tt.args.docID); got != tt.want {
				t.Errorf("Matches.HasDocID() %s = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
