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
	"fmt"
	"strings"
	"testing"
)

func TestSetDefaults(t *testing.T) {
	type args struct {
		shards   int
		replicas int
	}

	tests := []struct {
		name  string
		args  args
		want  int
		want1 int
	}{
		{
			"set defaults",
			args{
				shards:   0,
				replicas: 0,
			},
			defaultShards,
			defaultReplicas,
		},
		{
			"set default replicas",
			args{
				shards:   3,
				replicas: 0,
			},
			3,
			defaultReplicas,
		},
		{
			"set default shards",
			args{
				shards:   0,
				replicas: 3,
			},
			defaultShards,
			3,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			got, got1 := setDefaults(tt.args.shards, tt.args.replicas)
			if got != tt.want {
				t.Errorf("setDefaults() got = %v, want %v", got, tt.want)
			}

			if got1 != tt.want1 {
				t.Errorf("setDefaults() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestV1ESMapping(t *testing.T) {
	type args struct {
		shards   int
		replicas int
	}

	tests := []struct {
		name string
		args args
	}{
		{
			"with defaults",
			args{
				shards:   3,
				replicas: 1,
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			mapping := V1ESMapping(tt.args.shards, tt.args.replicas)

			if mapping == "" {
				t.Errorf("V1ESMapping() = mapping cannot be empty")
			}

			if !strings.Contains(mapping, fmt.Sprintf("\"number_of_replicas\": %d", tt.args.replicas)) {
				t.Errorf("V1ESMapping() = number_of_replicas not correct")
			}

			if !strings.Contains(mapping, fmt.Sprintf("\"number_of_shards\": %d", tt.args.shards)) {
				t.Errorf("V1ESMapping() = number_of_shards not correct")
			}
		})
	}
}
