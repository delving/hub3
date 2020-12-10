package ead

import "testing"

func Test_extractSpecs(t *testing.T) {
	type args struct {
		specs []string
	}

	tests := []struct {
		name            string
		args            args
		wantArchiveID   string
		wantInventoryID string
	}{
		{
			"empty",
			args{},
			"",
			"",
		},
		{
			"all fields",
			args{specs: []string{"TOE:9.01.001", "INV:5ED"}},
			"9.01.001",
			"5ED",
		},
		{
			"only archiveID",
			args{specs: []string{"TOE:9.01.001"}},
			"9.01.001",
			"",
		},
		{
			"only inventoryID",
			args{specs: []string{"INV:5ED"}},
			"",
			"5ED",
		},
		{
			"wrong prefix",
			args{specs: []string{"5ED", "TOE1:1.04.02"}},
			"",
			"",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			gotArchiveID, gotInventoryID := extractSpecs(tt.args.specs)
			if gotArchiveID != tt.wantArchiveID {
				t.Errorf("extractSpecs() %s -> gotArchiveID = %v, want %v", tt.name, gotArchiveID, tt.wantArchiveID)
			}

			if gotInventoryID != tt.wantInventoryID {
				t.Errorf("extractSpecs() %s -> gotUuid = %v, want %v", tt.name, gotInventoryID, tt.wantInventoryID)
			}
		})
	}
}
