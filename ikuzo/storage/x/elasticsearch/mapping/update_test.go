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

package mapping

import (
	"testing"
)

func TestValidate(t *testing.T) {
	type args struct {
		keys map[string]string
	}

	tests := []struct {
		name        string
		args        args
		wantOld     string
		wantCurrent string
		wantOk      bool
	}{
		{
			"mismatched",
			args{
				map[string]string{
					"123": "hello world",
				},
			},
			"123",
			"45ab6734b21e6968",
			false,
		},
		{
			"matched",
			args{
				map[string]string{
					"45ab6734b21e6968": "hello world",
				},
			},
			"",
			"",
			true,
		},
		{
			"v2 mapping",
			args{map[string]string{v2MappingSha: v2Mapping}},
			"",
			"",
			true,
		},
		{
			"v2 update mapping",
			args{map[string]string{v2UpdateMappingSha: v2MappingUpdate}},
			"",
			"",
			true,
		},
		{
			"fragment mapping",
			args{map[string]string{fragmentMappingSha: fragmentMapping}},
			"",
			"",
			true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			gotOld, gotCurrent, gotOk := validate(tt.args.keys)
			if gotOld != tt.wantOld {
				t.Errorf("validate() %s gotOld = %v, want %v", tt.name, gotOld, tt.wantOld)
			}
			if gotCurrent != tt.wantCurrent {
				t.Errorf("validate() %s gotCurrent  = %v, want %v", tt.name, gotCurrent, tt.wantCurrent)
			}
			if gotOk != tt.wantOk {
				t.Errorf("validate() %s gotOk = %v, want %v", tt.name, gotOk, tt.wantOk)
			}
		})
	}
}

func TestValidateMappings(t *testing.T) {
	tests := []struct {
		name        string
		wantOld     string
		wantCurrent string
		wantOk      bool
	}{
		{
			"check current mappings",
			"",
			"",
			true,
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			gotOld, gotCurrent, gotOk := validateMappings()
			if gotOld != tt.wantOld {
				t.Errorf("ValidateMappings() gotOld = %v, want %v", gotOld, tt.wantOld)
			}

			if gotCurrent != tt.wantCurrent {
				t.Errorf("ValidateMappings() gotCurrent = %v, want %v", gotCurrent, tt.wantCurrent)
			}

			if gotOk != tt.wantOk {
				t.Errorf("ValidateMappings() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}
