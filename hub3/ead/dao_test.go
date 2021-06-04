package ead

import (
	"testing"

	"github.com/delving/hub3/hub3/ead/eadpb"
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
