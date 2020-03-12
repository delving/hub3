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
		pos  int
		slop int
	}

	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			"no slop",
			args{pos: 1, slop: 0},
			[]int{1, 2},
		},
		{
			"slop 1",
			args{pos: 1, slop: 1},
			[]int{0, 1, 2},
		},
		{
			"slop 3",
			args{pos: 1, slop: 3},
			[]int{0, 1, 2, 3, 4},
		},
		{
			"slop 3; start 10",
			args{pos: 10, slop: 3},
			[]int{7, 8, 9, 10, 11, 12, 13},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			got := ValidPhrasePosition(tt.args.pos, tt.args.slop)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ValidPhrasePosition() %s = mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}
