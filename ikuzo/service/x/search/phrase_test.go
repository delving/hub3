package search

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestIsPhraseMatch(t *testing.T) {
	type args struct {
		pos1 int
		pos2 int
		slop int
	}

	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			"no match",
			args{5, 10, 0},
			false,
			false,
		},
		{
			"match no slop",
			args{5, 6, 0},
			true,
			false,
		},
		{
			"reverse match no slop",
			args{6, 5, 0},
			false,
			false,
		},
		{
			"match 1 slop",
			args{5, 6, 1},
			true,
			false,
		},
		{
			"reverse match 1 slop",
			args{7, 5, 1},
			false,
			false,
		},
		{
			"match 5 slop",
			args{4, 9, 5},
			true,
			false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			got, err := IsPhraseMatch(tt.args.pos1, tt.args.pos2, tt.args.slop)
			if (err != nil) != tt.wantErr {
				t.Errorf("PhraseDistanceMatch() %s; error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("PhraseDistanceMatch() %s = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestValidPhrasePosition(t *testing.T) {
	type args struct {
		vector Vector
		slop   int
	}

	tests := []struct {
		name string
		args args
		want []Vector
	}{
		{
			"no slop",
			args{vector: Vector{DocID: 1, Location: 1}, slop: 0},
			[]Vector{
				{DocID: 1, Location: 1},
				{DocID: 1, Location: 2},
			},
		},
		{
			"slop 1",
			args{vector: Vector{DocID: 1, Location: 1}, slop: 1},
			[]Vector{
				{DocID: 1, Location: 1},
				{DocID: 1, Location: 2},
			},
		},
		{
			"slop 1; start 5",
			args{vector: Vector{DocID: 1, Location: 5}, slop: 1},
			[]Vector{
				{DocID: 1, Location: 4},
				{DocID: 1, Location: 5},
				{DocID: 1, Location: 6},
			},
		},
		{
			"slop 3",
			args{vector: Vector{DocID: 1, Location: 1}, slop: 3},
			[]Vector{
				{DocID: 1, Location: 1},
				{DocID: 1, Location: 2},
				{DocID: 1, Location: 3},
				{DocID: 1, Location: 4},
			},
		},
		{
			"slop 3; start 10",
			args{vector: Vector{DocID: 1, Location: 10}, slop: 3},
			[]Vector{
				{DocID: 1, Location: 7},
				{DocID: 1, Location: 8},
				{DocID: 1, Location: 9},
				{DocID: 1, Location: 10},
				{DocID: 1, Location: 11},
				{DocID: 1, Location: 12},
				{DocID: 1, Location: 13},
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			got := ValidPhrasePosition(tt.args.vector, tt.args.slop)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ValidPhrasePosition() %s = mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}
