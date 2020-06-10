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

package search

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAnalyzer_Transform(t *testing.T) {
	type args struct {
		text string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"lowercase word",
			args{text: "word"},
			"word",
		},
		{
			"uppercase word",
			args{text: "Word"},
			"word",
		},
		{
			"all uppercase word",
			args{text: "WORD"},
			"word",
		},
		{
			"ascii folding",
			args{text: "curaçao övergångsställe"},
			"curacao overgangsstalle",
		},
		{
			"uppercase ascii folding",
			args{text: "SKÄRGÅRDSÖ"},
			"skargardso",
		},
		{
			"trim unwanted punctuation characters",
			args{text: "[(word).,:;?]"},
			"word",
		},
		{
			"trim single quote",
			args{text: "'word'"},
			"word",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			a := &Analyzer{}

			if diff := cmp.Diff(tt.want, a.Transform(tt.args.text)); diff != "" {
				t.Errorf("Analyzer.Transform(); %s = mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}
