package index

import (
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
