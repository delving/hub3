// Copyright 2020 Delving B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
