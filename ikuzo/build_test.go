package ikuzo

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewBuildVersionInfo(t *testing.T) {
	type args struct {
		version    string
		commit     string
		buildagent string
		builddate  string
	}

	tests := []struct {
		name string
		args args
		want *BuildVersionInfo
	}{
		{
			"empty build info",
			args{},
			&BuildVersionInfo{
				Version: "devBuild",
			},
		},
		{
			"empty build info",
			args{
				version:    "1.2.3",
				commit:     "fb28",
				buildagent: "local",
				builddate:  "2020-04-20_02:01:12PM",
			},
			&BuildVersionInfo{
				Version:    "1.2.3",
				Commit:     "fb28",
				BuildAgent: "local",
				BuildDate:  "2020-04-20_02:01:12PM",
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			got := NewBuildVersionInfo(tt.args.version, tt.args.commit, tt.args.buildagent, tt.args.builddate)

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("NewBuildVersionInfo() %s = mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}
