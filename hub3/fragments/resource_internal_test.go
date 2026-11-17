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

package fragments

import "testing"

func Test_padYears(t *testing.T) {
	type args struct {
		year  string
		start bool
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			"full-year",
			args{
				year:  "1990-05-12",
				start: true,
			},
			"1990-05-12",
			false,
		},
		{
			"year month (start)",
			args{
				year:  "1990-05",
				start: true,
			},
			"1990-05-01",
			false,
		},
		{
			"year month (end)",
			args{
				year:  "1990-05",
				start: false,
			},
			"1990-05-31",
			false,
		},
		{
			"year february (end)",
			args{
				year:  "1990-02",
				start: false,
			},
			"1990-02-28",
			false,
		},
		{
			"year april (end)",
			args{
				year:  "1990-04",
				start: false,
			},
			"1990-04-30",
			false,
		},
		{
			"year only (start)",
			args{
				year:  "1990",
				start: true,
			},
			"1990-01-01",
			false,
		},
		{
			"year only (end)",
			args{
				year:  "1990",
				start: false,
			},
			"1990-12-31",
			false,
		},
		{
			"unhyphenated date",
			args{
				year:  "19901011",
				start: false,
			},
			"1990-10-11",
			false,
		},
		{
			"unhyphenated year-month",
			args{
				year:  "199010",
				start: false,
			},
			"1990-10-31",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := padYears(tt.args.year, tt.args.start)
			if (err != nil) != tt.wantErr {
				t.Errorf("padYears() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("padYears() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_hyphenateDate(t *testing.T) {
	type args struct {
		date string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			"YYYYMMDD",
			args{date: "16880516"},
			"1688-05-16",
			false,
		},
		{
			"YYYYMM",
			args{date: "168805"},
			"1688-05",
			false,
		},
		{
			"YYYY",
			args{date: "1688"},
			"1688",
			false,
		},
		{
			"bad date string",
			args{date: "168"},
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := hyphenateDate(tt.args.date)
			if (err != nil) != tt.wantErr {
				t.Errorf("hyphenateDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("hyphenateDate() = %v, want %v", got, tt.want)
			}
		})
	}
}

// An unparseable typed number keeps its entry (without the number) and never
// drops the resource's other entries (F7).
func TestSetEntriesKeepsInvalidNumbers(t *testing.T) {
	const xsd = "http://www.w3.org/2001/XMLSchema#"
	fr := &FragmentResource{ID: "urn:x", predicates: map[string][]*FragmentEntry{
		"https://schema.org/width":  {{Value: "abc", DataType: xsd + "integer", EntryType: "Literal", Order: 1}},
		"https://schema.org/height": {{Value: "45", DataType: xsd + "integer", EntryType: "Literal", Order: 2}},
		"https://schema.org/depth":  {{Value: "x.y", DataType: xsd + "float", EntryType: "Literal", Order: 3}},
		"https://schema.org/name":   {{Value: "Molen", EntryType: "Literal", Order: 4}},
	}}

	for range 20 { // predicates is a map: try several iteration orders
		if err := fr.SetEntries(&ResourceMap{}); err != nil {
			t.Fatalf("SetEntries: %v", err)
		}
		if len(fr.Entries) != 4 {
			t.Fatalf("got %d entries, want 4", len(fr.Entries))
		}
		width, height, depth := fr.Entries[0], fr.Entries[1], fr.Entries[2]
		if width.Value != "abc" || width.Integer != 0 || height.Integer != 45 || depth.Value != "x.y" || depth.Float != 0 {
			t.Fatalf("entries %+v %+v %+v", width, height, depth)
		}
	}
}

// xsd:decimal and xsd:double fill Float like xsd:float (F20).
func TestNewResourceEntryDecimalAndDouble(t *testing.T) {
	const xsd = "http://www.w3.org/2001/XMLSchema#"
	for _, tc := range []struct {
		value, datatype string
		want            float64
	}{
		{"38.5", "decimal", 38.5},
		{"-0.125", "decimal", -0.125},
		{"1.5E2", "double", 150},
		{"1,5", "decimal", 0}, // invalid: kept without a number (F7)
		{"NaN", "double", 0}, // valid xsd:double, but JSON has no NaN/INF
		{"INF", "double", 0},
	} {
		fe := &FragmentEntry{Value: tc.value, DataType: xsd + tc.datatype, EntryType: "Literal"}
		re, err := fe.NewResourceEntry("https://schema.org/height", 1, &ResourceMap{})
		if err != nil {
			t.Fatalf("%s^^%s: %v", tc.value, tc.datatype, err)
		}
		if re.Float != tc.want || re.Value != tc.value {
			t.Errorf("%s^^%s: float %v value %q, want %v", tc.value, tc.datatype, re.Float, re.Value, tc.want)
		}
	}
}
