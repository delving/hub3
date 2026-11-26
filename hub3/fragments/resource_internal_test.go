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

import (
	"math/rand/v2"
	"slices"
	"sort"
	"testing"

	c "github.com/delving/hub3/config"
)

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
		{"NaN", "double", 0},  // valid xsd:double, but JSON has no NaN/INF
		{"INF", "double", 0},
		{"1e300", "double", 0}, // beyond the mapping's 32-bit float: OpenSearch refused the record
		{"-1e39", "decimal", 0},
		{"3.4e38", "double", 3.4e38}, // still a float32
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

// Typed date literals get isoDate and dateRange whatever their predicate; values
// the mapping cannot take are left out instead of failing the record (F19).
func TestNewResourceEntryDatesFromDatatype(t *testing.T) {
	prev := c.Config.RDFTagMap
	c.Config.RDFTagMap = nil
	t.Cleanup(func() { c.Config.RDFTagMap = prev })

	const xsd = "http://www.w3.org/2001/XMLSchema#"
	for _, tc := range []struct {
		value, datatype string
		iso, from, to   string // "" = no date fields
	}{
		{"1660-01-01", "date", "1660-01-01", "1660-01-01", "1660-01-01"},
		{"1660-01-01+01:00", "date", "1660-01-01", "1660-01-01", "1660-01-01"},
		{"1660-01-01T00:00:00", "dateTime", "1660-01-01", "1660-01-01", "1660-01-01"},
		{"1660-01-01T12:30:00Z", "dateTime", "1660-01-01", "1660-01-01", "1660-01-01"},
		{"1660", "gYear", "1660-01-01", "1660-01-01", "1660-12-31"},
		{"1660-02", "gYearMonth", "1660-02-01", "1660-02-01", "1660-02-28"},
		{"ca. 1660", "gYear", "", "", ""},
		{"-0500", "gYear", "", "", ""},
		{"1660-13", "gYearMonth", "", "", ""},
		{"1660-02-30", "date", "", "", ""},
	} {
		fe := &FragmentEntry{Value: tc.value, DataType: xsd + tc.datatype, EntryType: "Literal"}
		re, err := fe.NewResourceEntry("http://www.cidoc-crm.org/cidoc-crm/P82a_begin_of_the_begin", 1, &ResourceMap{})
		if err != nil {
			t.Fatalf("%s^^%s: %v", tc.value, tc.datatype, err)
		}
		var gotFrom, gotTo string
		if re.DateRange != nil {
			gotFrom, gotTo = re.DateRange.Greater, re.DateRange.Less
		}
		wantDate := []string{}
		if tc.iso != "" {
			wantDate = []string{tc.iso}
		}
		if !slices.Equal(append([]string{}, re.Date...), wantDate) || gotFrom != tc.from || gotTo != tc.to {
			t.Errorf("%s^^%s: isoDate %v range %q..%q, want %v %q..%q", tc.value, tc.datatype, re.Date, gotFrom, gotTo, wantDate, tc.from, tc.to)
		}
	}
}

// A predicate configured as isoDate/dateRange whose value is not a date the
// mapping accepts gets no date fields; OpenSearch refused the whole record (F19).
func TestNewResourceEntryConfiguredDatesValidated(t *testing.T) {
	prev := c.Config.RDFTagMap
	c.Config.RDFTagMap = c.NewRDFTagMap(&c.RawConfig{RDFTag: c.RDFTag{
		IsoDate:   []string{"https://schema.org/dateCreated"},
		DateRange: []string{"https://schema.org/dateCreated"},
	}})
	t.Cleanup(func() { c.Config.RDFTagMap = prev })

	for _, tc := range []struct {
		value     string
		wantDates []string
	}{
		{"1660-05-01", []string{"1660-05-01", "1660-05-01", "1660-05-01"}},
		{"zomer 1660", []string{"1660-01-01", "1660-12-31"}}, // raw text out, the range's year stays
		{"1650-1700", []string{"1650-01-01", "1700-12-31"}}, // years joined by a hyphen: a period (the raw text is no isoDate)
		{"1650-13", nil}, // year-month with no such month: no valid date
	} {
		fe := &FragmentEntry{Value: tc.value, EntryType: "Literal"}
		re, err := fe.NewResourceEntry("https://schema.org/dateCreated", 1, &ResourceMap{})
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(re.Date, tc.wantDates) {
			t.Errorf("%q: isoDate %v, want %v", tc.value, re.Date, tc.wantDates)
		}
		if tc.wantDates == nil && re.DateRange != nil {
			t.Errorf("%q: dateRange %+v, want none", tc.value, re.DateRange)
		}
	}
}

// BySortOrder is a valid ordering: resources without entries first, by where
// they first appear in the source, then by first entry order (F15).
func TestBySortOrderIsDeterministic(t *testing.T) {
	mk := func(id string, first int, entryOrder ...int) *FragmentResource {
		fr := &FragmentResource{ID: id, firstOrder: first}
		for _, o := range entryOrder {
			fr.Entries = append(fr.Entries, &ResourceEntry{Order: o})
		}
		return fr
	}
	want := []string{"typeB", "typeA", "typeD", "typeC", "d", "c"}
	rnd := rand.New(rand.NewPCG(1, 2))
	for i := range 200 {
		rs := []*FragmentResource{
			mk("typeA", 5), mk("c", 3, 3), mk("typeB", 2), mk("d", 1, 1), mk("typeC", 9), mk("typeD", 7),
		}
		rnd.Shuffle(len(rs), func(a, b int) { rs[a], rs[b] = rs[b], rs[a] })
		sort.Sort(BySortOrder(rs))
		got := []string{}
		for _, r := range rs {
			got = append(got, r.ID)
		}
		if !slices.Equal(got, want) {
			t.Fatalf("round %d: got %v, want %v", i, got, want)
		}
	}
}

// An xsd:integer outside the mapping's 32-bit integer keeps its entry without
// the number; OpenSearch refused the whole record.
func TestNewResourceEntryIntegerRange(t *testing.T) {
	for _, tc := range []struct {
		value string
		want  int
	}{
		{"2147483647", 2147483647},
		{"-2147483648", -2147483648},
		{"2147483648", 0},
		{"12345678901", 0},
	} {
		fe := &FragmentEntry{Value: tc.value, DataType: "http://www.w3.org/2001/XMLSchema#integer", EntryType: "Literal"}
		re, err := fe.NewResourceEntry("https://schema.org/identifier", 1, &ResourceMap{})
		if err != nil {
			t.Fatal(err)
		}
		if re.Integer != tc.want || re.Value != tc.value {
			t.Errorf("%s: integer %d value %q, want %d", tc.value, re.Integer, re.Value, tc.want)
		}
	}
}

// isIndexableDate accepts only what the mapping's formats parse: seconds with
// no fraction or exactly three digits.
func TestIsIndexableDate(t *testing.T) {
	for v, want := range map[string]bool{
		"2024-01-01":                  true,
		"2024":                        true,
		"01-02-2024":                  true,
		"2024-01-01 10:00:00":         true,
		"2024-01-01T10:00:00Z":        true,
		"2024-01-01T10:00:00+01:00":   true,
		"2024-01-01T10:00:00.123Z":    true,
		"2024-01-01T10:00:00.5Z":      false,
		"2024-01-01T10:00:00.123456Z": false,
		"2024-01-01T10:00:00":         false,
		"zomer 1660":                  false,
	} {
		if got := isIndexableDate(v); got != want {
			t.Errorf("isIndexableDate(%q) = %v, want %v", v, got, want)
		}
	}
}
