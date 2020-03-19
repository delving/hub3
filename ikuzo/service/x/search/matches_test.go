package search

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMatches_Merge(t *testing.T) {
	type fields struct {
		hits map[string]int
	}

	type args struct {
		hits map[string]int
	}

	tests := []struct {
		name     string
		fields   fields
		args     args
		wantHits map[string]int
	}{
		{
			"source only merge",
			fields{hits: map[string]int{"word": 1}},
			args{hits: map[string]int{}},
			map[string]int{"word": 1},
		},
		{
			"partial merge",
			fields{hits: map[string]int{"word": 1}},
			args{hits: map[string]int{"word": 1, "words": 2}},
			map[string]int{"word": 2, "words": 2},
		},
		{
			"target only merge",
			fields{hits: map[string]int{}},
			args{hits: map[string]int{"word": 1, "words": 2}},
			map[string]int{"word": 1, "words": 2},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			src := &Matches{
				termFrequency: tt.fields.hits,
			}

			target := NewMatches()
			target.termFrequency = tt.args.hits

			src.Merge(target)

			if diff := cmp.Diff(tt.wantHits, src.TermFrequency(), cmp.AllowUnexported(Matches{})); diff != "" {
				t.Errorf("Matches.Merge() %s = mismatch (-want +got):\n%s", tt.name, diff)
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
			"no results",
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
