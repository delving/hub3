package domain

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewHubID(t *testing.T) {
	type args struct {
		input string
	}

	tests := []struct {
		name    string
		args    args
		want    HubID
		wantErr bool
	}{
		{
			"valid",
			args{input: "org_spec_123"},
			HubID{OrgID: "org", DatasetID: "spec", LocalID: "123"},
			false,
		},
		{
			"invalid empty",
			args{input: ""},
			HubID{OrgID: "", DatasetID: "", LocalID: ""},
			true,
		},
		{
			"invalid missing id",
			args{input: "org_spec"},
			HubID{OrgID: "", DatasetID: "", LocalID: ""},
			true,
		},
		{
			"invalid missing id",
			args{input: "org_spec_"},
			HubID{OrgID: "", DatasetID: "", LocalID: ""},
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewHubID(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewHubID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("NewHubID() %s mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}
