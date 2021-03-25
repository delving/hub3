package revision

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func Test_logParser_parseLine(t *testing.T) {
	type fields struct {
		commitID string
	}

	type args struct {
		input string
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    DiffFile
		wantErr bool
	}{
		{
			"empty line",
			fields{},
			args{},
			DiffFile{},
			true,
		},
		{
			"parse commit",
			fields{},
			args{input: "acb3c03233ceec7809f892e9efe939821755ff0b 2020-04-01 14:23:22 +0200"},
			DiffFile{},
			false,
		},
		{
			"parse commit with double quote",
			fields{},
			args{input: "\"acb3c03233ceec7809f892e9efe939821755ff0b 2020-04-01 14:23:22 +0200\""},
			DiffFile{},
			false,
		},
		{
			"only modified",
			fields{commitID: "acb3c03233ceec7809f892e9efe939821755ff0b"},
			args{input: `M	rsc/NL-HaNA_4.OSK_2~2.105.json`},
			DiffFile{
				State:    StatusModified,
				Path:     "rsc/NL-HaNA_4.OSK_2~2.105.json",
				CommitID: "acb3c03233ceec7809f892e9efe939821755ff0b",
			},
			false,
		},
		{
			"only Deleted",
			fields{commitID: "acb3c03233ceec7809f892e9efe939821755ff0b"},
			args{input: `D	rsc/NL-HaNA_4.OSK_2~2.106.json`},
			DiffFile{
				State:    StatusDeleted,
				Path:     "rsc/NL-HaNA_4.OSK_2~2.106.json",
				CommitID: "acb3c03233ceec7809f892e9efe939821755ff0b",
			},
			false,
		},
		{
			"only Added",
			fields{commitID: "acb3c03233ceec7809f892e9efe939821755ff0b"},
			args{input: `A	rsc/NL-HaNA_4.OSK_2~2.107.json`},
			DiffFile{
				State:    StatusAdded,
				Path:     "rsc/NL-HaNA_4.OSK_2~2.107.json",
				CommitID: "acb3c03233ceec7809f892e9efe939821755ff0b",
			},
			false,
		},
		{
			"with whitespace",
			fields{commitID: "acb3c03233ceec7809f892e9efe939821755ff0b"},
			args{input: `A	rsc/NL-HaNA_4.OSK_2~2.107 123.json`},
			DiffFile{
				State:    StatusAdded,
				Path:     "rsc/NL-HaNA_4.OSK_2~2.107 123.json",
				CommitID: "acb3c03233ceec7809f892e9efe939821755ff0b",
			},
			false,
		},
		{
			"unknown state",
			fields{},
			args{input: `!	rsc/NL-HaNA_4.OSK_2~2.108.json`},
			DiffFile{},
			true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			p := &logParser{
				commitID: tt.fields.commitID,
			}

			got, err := p.parseLine(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("logParser.parseLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("logParser.parseLine() %s = mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}

func Test_logParser_parse(t *testing.T) {
	type fields struct {
		commitID string
		files    map[string]DiffFile
	}

	type args struct {
		input string
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]DiffFile
		wantErr bool
	}{
		{
			"empty lines",
			fields{},
			args{input: ""},
			nil,
			false,
		},
		{
			"multiline",
			fields{files: map[string]DiffFile{}},
			args{input: `acb3c03233ceec7809f892e9efe939821755ff0b 2020-04-01 14:23:22 +0000
A	rsc/1.json
A	rsc/2.json
A	rsc/3.json
A	rsc/4.json
bcb3c03233ceec7809f892e9efe939821755ff0b 2020-04-01 14:24:22 +0000
D	rsc/1.json
M	rsc/2.json
D	rsc/3.json
ccb3c03233ceec7809f892e9efe939821755ff0b 2020-04-01 14:25:22 +0000
A	rsc/1.json`},
			map[string]DiffFile{
				"rsc/1.json": {
					State:      "A",
					Path:       "rsc/1.json",
					CommitID:   "ccb3c03233ceec7809f892e9efe939821755ff0b",
					CommitDate: time.Date(2020, 04, 01, 14, 25, 22, 00, time.UTC),
				},

				"rsc/2.json": {
					State:      "M",
					Path:       "rsc/2.json",
					CommitID:   "bcb3c03233ceec7809f892e9efe939821755ff0b",
					CommitDate: time.Date(2020, 04, 01, 14, 24, 22, 00, time.UTC),
				},
				"rsc/3.json": {
					State:      "D",
					Path:       "rsc/3.json",
					CommitID:   "bcb3c03233ceec7809f892e9efe939821755ff0b",
					CommitDate: time.Date(2020, 04, 01, 14, 24, 22, 00, time.UTC),
				},
				"rsc/4.json": {
					State:      "A",
					Path:       "rsc/4.json",
					CommitID:   "acb3c03233ceec7809f892e9efe939821755ff0b",
					CommitDate: time.Date(2020, 04, 01, 14, 23, 22, 00, time.UTC),
				},
			},
			false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			p := &logParser{
				commitID: tt.fields.commitID,
				files:    tt.fields.files,
			}
			if err := p.parse(tt.args.input); (err != nil) != tt.wantErr {
				t.Errorf("logParser.parse() %s error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, p.files); diff != "" {
				t.Errorf("logParser.parse() %s = mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}

func Test_parseGitDate(t *testing.T) {
	type args struct {
		text string
	}

	tests := []struct {
		name    string
		args    args
		want    time.Time
		wantErr bool
	}{
		{
			"empty string",
			args{text: ""},
			time.Time{},
			true,
		},
		{
			"correct date",
			args{text: "2020-04-01 14:23:22 +0000"},
			time.Date(2020, 04, 01, 14, 23, 22, 00, time.UTC),
			false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			got, err := parseGitDate(tt.args.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseGitDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("parseGitDate() %s = mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}
