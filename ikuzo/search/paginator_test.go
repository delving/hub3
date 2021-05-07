package search

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewPaginator(t *testing.T) {
	type args struct {
		total       int
		pageSize    int
		currentPage int
		cursor      int
	}

	tests := []struct {
		name    string
		args    args
		want    *Paginator
		wantErr bool
	}{
		{
			"empty paginator",
			args{total: 0, pageSize: 16, currentPage: 0, cursor: 0},
			&Paginator{
				Start:       0,
				Rows:        16,
				NumFound:    0,
				FirstPage:   0,
				LastPage:    0,
				CurrentPage: 1,
				HasNext:     false,
			},
			false,
		},
		{
			"full paginator",
			args{total: 1316227, pageSize: 16, currentPage: 0, cursor: 0},
			&Paginator{
				Start:              1,
				Rows:               16,
				NumFound:           1316227,
				FirstPage:          1,
				LastPage:           82265,
				CurrentPage:        1,
				HasNext:            true,
				HasPrevious:        false,
				NextPageNumber:     2,
				PreviousPageNumber: 0,
				NextPage:           17,
				PreviousPage:       0,
			},
			false,
		},
		{
			"page 3 paginator",
			args{total: 1316227, pageSize: 16, currentPage: 3},
			&Paginator{
				Start:              33,
				Rows:               16,
				NumFound:           1316227,
				FirstPage:          1,
				LastPage:           82265,
				CurrentPage:        3,
				HasNext:            true,
				HasPrevious:        true,
				NextPageNumber:     4,
				PreviousPageNumber: 2,
				NextPage:           49,
				PreviousPage:       17,
			},
			false,
		},
		{
			"last page paginator with cursor",
			args{total: 48, pageSize: 16, currentPage: 0, cursor: 32},
			&Paginator{
				Start:              33,
				Rows:               16,
				NumFound:           48,
				FirstPage:          1,
				LastPage:           3,
				CurrentPage:        3,
				HasNext:            false,
				HasPrevious:        true,
				NextPageNumber:     0,
				PreviousPageNumber: 2,
				NextPage:           0,
				PreviousPage:       17,
			},
			false,
		},
		{
			"last page paginator",
			args{total: 48, pageSize: 16, currentPage: 3},
			&Paginator{
				Start:              33,
				Rows:               16,
				NumFound:           48,
				FirstPage:          1,
				LastPage:           3,
				CurrentPage:        3,
				HasNext:            false,
				HasPrevious:        true,
				NextPageNumber:     0,
				PreviousPageNumber: 2,
				NextPage:           0,
				PreviousPage:       17,
			},
			false,
		},
		{
			"before last page paginator",
			args{total: 49, pageSize: 16, currentPage: 3},
			&Paginator{
				Start:              33,
				Rows:               16,
				NumFound:           49,
				FirstPage:          1,
				LastPage:           4,
				CurrentPage:        3,
				HasNext:            true,
				HasPrevious:        true,
				NextPageNumber:     4,
				PreviousPageNumber: 2,
				NextPage:           49,
				PreviousPage:       17,
			},
			false,
		},
		{
			"last page paginator",
			args{total: 48, pageSize: 16, currentPage: 3},
			&Paginator{
				Start:              33,
				Rows:               16,
				NumFound:           48,
				FirstPage:          1,
				LastPage:           3,
				CurrentPage:        3,
				HasNext:            false,
				HasPrevious:        true,
				NextPageNumber:     0,
				PreviousPageNumber: 2,
				NextPage:           0,
				PreviousPage:       17,
			},
			false,
		},
		{
			"last page paginator",
			args{total: 48, pageSize: 16, currentPage: 3},
			&Paginator{
				Start:              33,
				Rows:               16,
				NumFound:           48,
				FirstPage:          1,
				LastPage:           3,
				CurrentPage:        3,
				HasNext:            false,
				HasPrevious:        true,
				NextPageNumber:     0,
				PreviousPageNumber: 2,
				NextPage:           0,
				PreviousPage:       17,
			},
			false,
		},
		{
			"last page paginator",
			args{total: 48, pageSize: 16, currentPage: 3},
			&Paginator{
				Start:              33,
				Rows:               16,
				NumFound:           48,
				FirstPage:          1,
				LastPage:           3,
				CurrentPage:        3,
				HasNext:            false,
				HasPrevious:        true,
				NextPageNumber:     0,
				PreviousPageNumber: 2,
				NextPage:           0,
				PreviousPage:       17,
			},
			false,
		},
		{
			"invalid page",
			args{total: 48, pageSize: 16, currentPage: 56},
			nil,
			true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPaginator(tt.args.total, tt.args.pageSize, tt.args.currentPage, tt.args.cursor)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPaginator() error %s = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("NewPaginator() %s mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}

func TestPaginator_getPageLinks(t *testing.T) {
	type args struct {
		total       int
		pageSize    int
		currentPage int
		cursor      int
	}

	tests := []struct {
		name    string
		args    args
		want    []PageLink
		wantErr bool
	}{
		{
			"empty result",
			args{0, 16, 1, 0},
			[]PageLink{{Start: 1, IsLinked: false, PageNumber: 1}},
			false,
		},
		{
			"two pages",
			args{21, 16, 1, 0},
			[]PageLink{
				{Start: 1, IsLinked: false, PageNumber: 1},
				{Start: 17, IsLinked: true, PageNumber: 2},
			},
			false,
		},
		{
			"full paging window",
			args{1000, 16, 1, 0},
			[]PageLink{
				{Start: 1, IsLinked: false, PageNumber: 1},
				{Start: 17, IsLinked: true, PageNumber: 2},
				{Start: 33, IsLinked: true, PageNumber: 3},
				{Start: 49, IsLinked: true, PageNumber: 4},
				{Start: 65, IsLinked: true, PageNumber: 5},
				{Start: 81, IsLinked: true, PageNumber: 6},
				{Start: 97, IsLinked: true, PageNumber: 7},
				{Start: 113, IsLinked: true, PageNumber: 8},
				{Start: 129, IsLinked: true, PageNumber: 9},
				{Start: 145, IsLinked: true, PageNumber: 10},
			},
			false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			p, err := NewPaginator(tt.args.total, tt.args.pageSize, tt.args.currentPage, tt.args.cursor)
			if (err != nil) != tt.wantErr {
				t.Errorf("Paginator.getPageLinks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			got, err := p.getPageLinks()
			if (err != nil) != tt.wantErr {
				t.Errorf("Paginator.getPageLinks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Paginator.getPageLinks() %s mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}

func TestPaginator_getPageNumber(t *testing.T) {
	type fields struct {
		Cursor      int
		Rows        int
		NumFound    int
		CurrentPage int
	}

	tests := []struct {
		name    string
		fields  fields
		want    int
		wantErr bool
	}{
		{
			"empty result; first page",
			fields{
				Cursor:      0,
				Rows:        16,
				NumFound:    0,
				CurrentPage: 0,
			},
			1,
			false,
		},
		{
			"first page; with results",
			fields{
				Cursor:      1,
				Rows:        16,
				NumFound:    100,
				CurrentPage: 0,
			},
			1,
			false,
		},
		{
			"second page; with results",
			fields{
				Cursor:      17,
				Rows:        16,
				NumFound:    100,
				CurrentPage: 0,
			},
			2,
			false,
		},
		{
			"third page; with results",
			fields{
				Cursor:      33,
				Rows:        16,
				NumFound:    100,
				CurrentPage: 0,
			},
			3,
			false,
		},
		{
			"third page; with page",
			fields{
				Cursor:      0,
				Rows:        16,
				NumFound:    100,
				CurrentPage: 3,
			},
			3,
			false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			p := &Paginator{
				Start:       tt.fields.Cursor,
				Rows:        tt.fields.Rows,
				NumFound:    tt.fields.NumFound,
				CurrentPage: tt.fields.CurrentPage,
			}

			got, err := p.getPageNumber()
			if (err != nil) != tt.wantErr {
				t.Errorf("Paginator.getPageNumber() %s error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("Paginator.getPageNumber() %s = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
