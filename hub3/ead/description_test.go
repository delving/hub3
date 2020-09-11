// Copyright 2017 Delving B.V.
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

package ead

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_itemBuilder_parse(t *testing.T) {
	type args struct {
		b []byte
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
		length  int
		items   []*DataItem
	}{
		{
			"no lb",
			args{[]byte(`<item>bertillonnage, 1896-1922, drankwet, 1883-1905 opiumwet, vanaf 1928</item>`)},
			false,
			1,
			[]*DataItem{
				{Type: 5, Text: "bertillonnage, 1896-1922, drankwet, 1883-1905 opiumwet, vanaf 1928", Depth: 1, ParentIDS: "", Tag: "item", Order: 1},
			},
		},
		{
			"emph",
			args{[]byte(`<item>bertillonnage, <emph>1896-1922,</emph> drankwet</item>`)},
			false,
			1,
			[]*DataItem{
				{Type: 5, Text: "bertillonnage, <em>1896-1922,</em> drankwet", Depth: 1, ParentIDS: "", Tag: "item", Order: 1},
			},
		},
		{
			"double lb",
			args{[]byte(`<item>bertillonnage, 1896-1922,<lb></lb> drankwet</item>`)},
			false,
			1,
			[]*DataItem{
				{Type: 5, Text: "bertillonnage, 1896-1922,<lb/> drankwet", Depth: 1, ParentIDS: "", Tag: "item", Order: 1},
			},
		},
		{
			"single lb",
			args{[]byte(`<item>bertillonnage, 1896-1922,<lb/> drankwet</item>`)},
			false,
			1,
			[]*DataItem{
				{Type: 5, Text: "bertillonnage, 1896-1922,<lb/> drankwet", Depth: 1, ParentIDS: "", Tag: "item", Order: 1},
			},
		},
		{
			"<lb/>",
			args{[]byte(`<defitem>
		<label>Strafrecht en strafvordering: verzamelterm voor:</label>
		<item>bertillonnage, 1896-1922<lb></lb>drankwet, 1883-1905</item>
		</defitem>`)},
			false,
			3,
			[]*DataItem{
				{Type: 6, Text: "", Depth: 1, ParentIDS: "", Tag: "defitem", Order: 1},
				{Type: 7, Text: "Strafrecht en strafvordering: verzamelterm voor:", Depth: 2, ParentIDS: "1", Tag: "label", Order: 2},
				{Type: 5, Text: "bertillonnage, 1896-1922<lb/> drankwet, 1883-1905", Depth: 2, ParentIDS: "1", Tag: "item", Order: 3},
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			ib := newItemBuilder(context.TODO())
			if err := ib.parse(tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("itemBuilder.parse() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.length != len(ib.items) {
				t.Errorf("itemBuilder.parse() got items = %d, wantErr %d", len(ib.items), tt.length)
			}

			if diff := cmp.Diff(tt.items, ib.items); diff != "" {
				t.Errorf("itemBuilder.parse() mismatch (-want +got):\n%s", diff)
			}

			t.Logf("item builder: %#v", ib.items[len(ib.items)-1])
			// t.FailNow()
		})
	}
}
