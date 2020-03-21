package ead

import (
	"testing"
)

func Test_getCursor(t *testing.T) {
	type args struct {
		rows int
		page int
	}

	tests := []struct {
		name string
		args args
		want int
	}{
		{"zero page", args{10, 0}, 0},
		{"10, page 1", args{10, 1}, 0},
		{"10, page 2", args{10, 2}, 10},
		{"10, page 3", args{10, 3}, 20},
		{"10, page 5", args{10, 5}, 40},
		{"10, page 10", args{10, 10}, 90},

		{"16, page 1", args{16, 1}, 0},
		{"16, page 2", args{16, 2}, 16},

		{"5, page 1", args{5, 1}, 0},
		{"5, page 2", args{5, 2}, 5},
		{"5, page 3", args{5, 3}, 10},

		{"3, page 1", args{3, 1}, 0},
		{"3, page 2", args{3, 2}, 3},
		{"3, page 3", args{3, 3}, 6},
		{"3, page 4", args{3, 4}, 9},
		{"3, page 5", args{3, 5}, 12},

		{"2, page 1", args{2, 1}, 0},
		{"2, page 2", args{2, 2}, 2},
		{"2, page 3", args{2, 3}, 4},
		{"2, page 4", args{2, 4}, 6},

		{"1, page 1", args{1, 1}, 0},
		{"1, page 2", args{1, 2}, 1},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := getCursor(tt.args.rows, tt.args.page); got != tt.want {
				t.Errorf("getCursor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isAdvancedSearch(t *testing.T) {
	type args struct {
		query string
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{"simple query", args{query: "one word"}, false},
		{"AND query", args{query: "this AND that"}, true},
		{"lower case and", args{query: "this and that"}, false},
		{"OR query", args{query: "this OR that"}, true},
		{"NOT query", args{query: "this NOT that"}, true},
		{"lower case or", args{query: "this or that"}, false},
		{"exclude query", args{query: "this -that"}, true},
		{"include query", args{query: "this +that"}, true},
		{"phrase query", args{query: "this \"two words\""}, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := isAdvancedSearch(tt.args.query); got != tt.want {
				t.Errorf("IsAdvancedSearch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getPageCount(t *testing.T) {
	type args struct {
		archives int
		rows     int
	}

	tests := []struct {
		name string
		args args
		want int
	}{
		{"no results", args{archives: 0, rows: 10}, 0},
		{"no rows", args{archives: 10, rows: 0}, 0},
		{"1 page", args{archives: 5, rows: 10}, 1},
		{"1 page equal", args{archives: 10, rows: 10}, 1},
		{"2 pages", args{archives: 15, rows: 10}, 2},
		{"3 pages", args{archives: 27, rows: 10}, 3},
		{"2 pages", args{archives: 31, rows: 10}, 4},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := getPageCount(tt.args.archives, tt.args.rows); got != tt.want {
				t.Errorf("getPageCount() = %v, want %v", got, tt.want)
			}
		})
	}
}
