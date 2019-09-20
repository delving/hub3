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
