package index

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
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

func TestEntry_Fingerprint(t *testing.T) {
	type fields struct {
		ID                string
		Predicate         string
		SearchLabel       string
		Value             string
		Language          string
		DataType          string
		EntryType         EntryType
		Level             int32
		Order             int
		Tags              []string
		TypeIndexField    TypeIndexField
		CustomFilterField CustomFilterField
		Inline            *Resource
		fingerprint       string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// source="&{ID: Predicate:https://data.antwerp.be/def/dlod/isConnectedTo SearchLabel:dlod_isConnectedTo Value:au::5099 Language: DataType: EntryType:Literal Level:0 Order:12 Tags:[] TypeIndexField:{Date:[] DateRange:<nil> Integer:0 Float:0 IntRange:<nil> LatLong:} CustomFilterField:{FilterIDs:[] Type: Role:} Inline:<nil> fingerprint:12991542146013111616}" target="&{ID: Predicate:https://data.antwerp.be/def/dlod/isRelatedTo SearchLabel:dlod_isRelatedTo Value:au::5099 Language: DataType: EntryType:Literal Level:0 Order:16 Tags:[] TypeIndexField:{Date:[] DateRange:<nil> Integer:0 Float:0 IntRange:<nil> LatLong:} CustomFilterField:{FilterIDs:[] Type: Role:} Inline:<nil> fingerprint:12991542146013111616}"
		{
			name: "",
			fields: fields{
				Predicate:   "https://data.antwerp.be/def/dlod/isConnectedTo",
				SearchLabel: "dlod_isConnectedTo",
				Value:       "au::5099",
				EntryType:   "Literal",
				Order:       12,
			},
			want: "1799388965257883955",
		},
		{
			name: "",
			fields: fields{
				Predicate:   "https://data.antwerp.be/def/dlod/isRelatedTo",
				SearchLabel: "dlod_isRelatedTo",
				Value:       "au::5099",
				EntryType:   "Literal",
				Order:       16,
			},
			want: "10987807126339939380",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Entry{
				ID:                tt.fields.ID,
				Predicate:         tt.fields.Predicate,
				SearchLabel:       tt.fields.SearchLabel,
				Value:             tt.fields.Value,
				Language:          tt.fields.Language,
				DataType:          tt.fields.DataType,
				EntryType:         tt.fields.EntryType,
				Level:             tt.fields.Level,
				Order:             tt.fields.Order,
				Tags:              tt.fields.Tags,
				TypeIndexField:    tt.fields.TypeIndexField,
				CustomFilterField: tt.fields.CustomFilterField,
				Inline:            tt.fields.Inline,
				fingerprint:       tt.fields.fingerprint,
			}
			if got := e.Fingerprint(); got != tt.want {
				t.Errorf("Entry.Fingerprint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEntry_MarshalJSON(t *testing.T) {
	tests := []struct {
		name       string
		entry      *Entry
		wantFields []string
		wantErr    bool
	}{
		{
			name: "literal with search label",
			entry: &Entry{
				ID:          "",
				Predicate:   "http://purl.org/dc/elements/1.1/title",
				SearchLabel: "dc_title",
				Value:       "Test Title",
				EntryType:   Literal,
			},
			wantFields: []string{"searchLabel", "@value"},
			wantErr:    false,
		},
		{
			name: "resource with search label",
			entry: &Entry{
				ID:          "http://example.org/resource/1",
				Predicate:   "http://purl.org/dc/elements/1.1/relation",
				SearchLabel: "dc_relation",
				EntryType:   ResourceType,
			},
			wantFields: []string{"searchLabel", "@id"},
			wantErr:    false,
		},
		{
			name: "literal without search label",
			entry: &Entry{
				Predicate: "http://purl.org/dc/elements/1.1/description",
				Value:     "Test Description",
				EntryType: Literal,
			},
			wantFields: []string{"@value"},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotData, err := json.Marshal(tt.entry)
			if (err != nil) != tt.wantErr {
				t.Errorf("Entry.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Convert JSON to string for easier inspection
			gotStr := string(gotData)

			// Check that all expected fields exist in the JSON
			for _, field := range tt.wantFields {
				if !strings.Contains(gotStr, field) {
					t.Errorf("Entry.MarshalJSON() = %v, missing expected field %q", gotStr, field)
				}
			}

			// Dynamic fields should NOT be included (as that functionality has been removed)
			if tt.entry.EntryType == Literal && tt.entry.SearchLabel != "" && tt.entry.Value != "" {
				unexpectedDynamicField := fmt.Sprintf(`"%s":"%s"`, tt.entry.SearchLabel, tt.entry.Value)
				if strings.Contains(gotStr, unexpectedDynamicField) {
					t.Errorf("Entry.MarshalJSON() = %v, contains unexpected dynamic field %q that should have been removed", gotStr, unexpectedDynamicField)
				}
			}

			// Test unmarshaling back
			var unmarshaled Entry
			err = json.Unmarshal(gotData, &unmarshaled)
			if err != nil {
				t.Errorf("Failed to unmarshal JSON back to Entry: %v", err)
				return
			}

			// Verify essential fields were preserved
			if unmarshaled.SearchLabel != tt.entry.SearchLabel {
				t.Errorf("Unmarshaled entry searchLabel = %v, want %v", unmarshaled.SearchLabel, tt.entry.SearchLabel)
			}
			if unmarshaled.Value != tt.entry.Value {
				t.Errorf("Unmarshaled entry value = %v, want %v", unmarshaled.Value, tt.entry.Value)
			}
			if unmarshaled.EntryType != tt.entry.EntryType {
				t.Errorf("Unmarshaled entry entryType = %v, want %v", unmarshaled.EntryType, tt.entry.EntryType)
			}
		})
	}
}

func TestEntry_ResolvedFields(t *testing.T) {
	entry := &Entry{
		ID:            "http://data.rkd.nl/artists/66219",
		Predicate:     "http://purl.org/dc/elements/1.1/creator",
		SearchLabel:   "dc_creator",
		Value:         "Rembrandt van Rijn",
		EntryType:     ResourceType,
		ResolvedFrom:  "http://xmlns.com/foaf/0.1/name",
		ResolvedLevel: 1,
	}

	if entry.ResolvedFrom != "http://xmlns.com/foaf/0.1/name" {
		t.Errorf("ResolvedFrom = %v, want foaf:name URI", entry.ResolvedFrom)
	}

	if entry.ResolvedLevel != 1 {
		t.Errorf("ResolvedLevel = %d, want 1", entry.ResolvedLevel)
	}
}

func TestEntry_ResolvedFields_JSON(t *testing.T) {
	entry := &Entry{
		ID:            "http://example.org/agent/1",
		Predicate:     "http://purl.org/dc/elements/1.1/creator",
		SearchLabel:   "dc_creator",
		Value:         "Test Agent",
		EntryType:     ResourceType,
		ResolvedFrom:  "http://www.w3.org/2004/02/skos/core#prefLabel",
		ResolvedLevel: 1,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("failed to marshal entry: %v", err)
	}

	jsonStr := string(data)

	if !strings.Contains(jsonStr, `"resolvedFrom"`) {
		t.Error("expected resolvedFrom in JSON output")
	}

	if !strings.Contains(jsonStr, `"resolvedLevel"`) {
		t.Error("expected resolvedLevel in JSON output")
	}
}

func TestEntry_ResolvedFields_OmitEmpty(t *testing.T) {
	// When not set (zero values), fields should be omitted from JSON
	entry := &Entry{
		Predicate:   "http://purl.org/dc/elements/1.1/title",
		SearchLabel: "dc_title",
		Value:       "Test Title",
		EntryType:   Literal,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("failed to marshal entry: %v", err)
	}

	jsonStr := string(data)

	if strings.Contains(jsonStr, "resolvedFrom") {
		t.Error("resolvedFrom should be omitted when empty")
	}

	if strings.Contains(jsonStr, "resolvedLevel") {
		t.Error("resolvedLevel should be omitted when zero")
	}
}