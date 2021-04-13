package ead

import (
	"github.com/delving/hub3/hub3/ead/eadpb"
	"testing"
)

func Test_find_duplicates_in_filenames(t *testing.T) {
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
			if _, err := mapToUniqueFilenamesWithSortKey(tt.items); (err != nil) != tt.wantErr {
				t.Errorf("mapToUniqueFilenamesWithSortKey = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_return_map_of_unique_filenames_with_their_sort_key(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
		items   *map[string]int32
	}{
		{
			"Should not throw error because of happy flow",
			false,
			&map[string]int32{
				"file_a": 1,
				"file_b": 2,
			},
		},
		{
			"Should throw error because the first sortKey is not as expected",
			true,
			&map[string]int32{
				"file_a": 2,
				"file_b": 3,
			},
		},
		{
			"Should throw error because the sortKeys are not in succeeding order",
			true,
			&map[string]int32{
				"file_a": 1,
				"file_b": 3,
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			if err := validateSortKeysAreOrdered(tt.items, 1); (err != nil) != tt.wantErr {
				t.Errorf("validateSortKeysAreOrdered = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
