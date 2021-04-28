package ead

import (
	"github.com/delving/hub3/hub3/ead/eadpb"
	"testing"
)

func TestFindDuplicatesInFilenames(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
		length  int
		items   []*eadpb.File
	}{
		{
			"Should throw error because of a duplicate filename",
			true,
			1,
			[]*eadpb.File{
				{Filename: "file_a"},
				{Filename: "file_a"},
			},
		},
		{
			"Should not throw error because of there are no duplicate filenames",
			false,
			1,
			[]*eadpb.File{
				{Filename: "file_a"},
				{Filename: "file_b"},
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			if err := assertUniqueFilenames(tt.items); (err != nil) != tt.wantErr {
				t.Errorf("assertUniqueFilenames = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReturnMapOfUniqueFilenamesWithTheirSortKey(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
		items   []*eadpb.File
	}{
		{
			"Should not throw error because of happy flow",
			false,
			[]*eadpb.File{
				{Filename: "file_a", SortKey: 1},
				{Filename: "file_a", SortKey: 2},
			},
		},
		{
			"Should throw error because the first sortKey is not as expected",
			true,
			[]*eadpb.File{
				{Filename: "file_a", SortKey: 2},
				{Filename: "file_a", SortKey: 3},
			},
		},
		{
			"Should throw error because the sortKeys are not in succeeding order",
			true,
			[]*eadpb.File{
				{Filename: "file_a", SortKey: 1},
				{Filename: "file_a", SortKey: 3},
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			if err := assertSortKeysAreOrdered(tt.items, 1); (err != nil) != tt.wantErr {
				t.Errorf("assertSortKeysAreOrdered = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
