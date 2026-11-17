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
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestV2ESMapping(t *testing.T) {
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
			mapping := V2ESMapping(tt.args.shards, tt.args.replicas)

			if mapping == "" {
				t.Errorf("V2ESMapping() = mapping cannot be empty")
			}

			if !strings.Contains(mapping, fmt.Sprintf("\"number_of_replicas\": %d", tt.args.replicas)) {
				t.Errorf("V2ESMapping() = number_of_replicas not correct")
			}

			if !strings.Contains(mapping, fmt.Sprintf("\"number_of_shards\": %d", tt.args.shards)) {
				t.Errorf("V2ESMapping() = number_of_shards not correct")
			}
		})
	}
}

func TestV2Mapping_ContainsResolvedFields(t *testing.T) {
	mapping := V2ESMapping(1, 0)

	if !strings.Contains(mapping, `"resolvedFrom"`) {
		t.Error("V2 mapping should contain resolvedFrom field")
	}

	if !strings.Contains(mapping, `"resolvedLevel"`) {
		t.Error("V2 mapping should contain resolvedLevel field")
	}
}

// Both mappings are valid JSON, and the context fields use the names the v2
// document writes (camelCase since F22).
func TestV2MappingsValidAndCamelCaseContext(t *testing.T) {
	for name, m := range map[string]string{"create": V2ESMapping(1, 0), "update": V2MappingUpdate()} {
		var doc map[string]any
		if err := json.Unmarshal([]byte(m), &doc); err != nil {
			t.Fatalf("%s mapping is not valid JSON: %v", name, err)
		}
		props := doc["mappings"]
		if props == nil {
			props = doc
		}
		b, _ := json.Marshal(props)
		for _, f := range []string{`"subjectClass"`, `"searchLabel"`, `"objectID"`, `"inline"`, `"sortValue"`} {
			if !strings.Contains(string(b), f) {
				t.Errorf("%s mapping lacks %s", name, f)
			}
		}
		for _, f := range []string{`"SubjectClass"`, `"SearchLabel"`, `"ObjectID"`} {
			if strings.Contains(string(b), f) {
				t.Errorf("%s mapping still has %s", name, f)
			}
		}
	}
}
