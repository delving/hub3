package search

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAutoComplete_FromStrings(t *testing.T) {
	type args struct {
		words []string
		text  string
	}

	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			"no words",
			args{
				words: []string{},
				text:  "",
			},
			[]byte(""),
		},
		{
			"single word",
			args{
				words: []string{"one"},
				text:  "one",
			},
			[]byte("\x00one\x00"),
		},
		{
			"multiple words",
			args{
				words: []string{"one", "two", "three"},
				text:  "one two. <lb/> three",
			},
			[]byte("\x00one\x00two\x00three\x00"),
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			ac := NewAutoComplete()
			ac.FromStrings(tt.args.words)

			if diff := cmp.Diff(tt.want, ac.data); diff != "" {
				t.Errorf("AutoComplete.FromStrings() %s = mismatch (-want +got):\n%s", tt.name, diff)
			}

			ac = NewAutoComplete()
			tok := NewTokenizer()
			ts := tok.ParseString(tt.args.text, 1)
			ac.FromTokenSteam(ts)

			if diff := cmp.Diff(tt.want, ac.data); diff != "" {
				t.Errorf("AutoComplete.FromTokenSteam() %s = mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}

func TestAutoComplete_getStringFromIndex(t *testing.T) {
	type fields struct {
		words []string
	}

	type args struct {
		index int
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			"match start",
			fields{[]string{"one", "two", "three"}},
			args{1},
			"one",
		},
		{
			"match middle",
			fields{[]string{"one", "two", "three"}},
			args{2},
			"one",
		},
		{
			"match end",
			fields{[]string{"one", "two", "three"}},
			args{3},
			"one",
		},
		{
			"match last word",
			fields{[]string{"one", "two", "three"}},
			args{13},
			"three",
		},
		{
			"out of range",
			fields{[]string{"one", "two", "three"}},
			args{20},
			"",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			ac := NewAutoComplete()
			ac.FromStrings(tt.fields.words)

			if diff := cmp.Diff(tt.want, ac.getStringFromIndex(tt.args.index)); diff != "" {
				t.Errorf("AutoComplete.getStringFromIndex() %s = mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}

func TestAutoComplete_Suggest(t *testing.T) {
	type fields struct {
		words []string
	}

	type args struct {
		input    string
		limit    int
		callback func(a Autos) Autos
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []Autos
		wantErr bool
	}{
		{
			"empty autocomplete",
			fields{},
			args{"bi", -1, nil},
			[]Autos{},
			true,
		},
		{
			"empty input",
			fields{[]string{"biathlon", "ambivalent"}},
			args{"", -1, nil},
			[]Autos{},
			true,
		},
		{
			"no matches",
			fields{[]string{"biathlon", "ambivalent"}},
			args{"u", -1, nil},
			[]Autos{},
			false,
		},
		{
			"matches",
			fields{[]string{"biathlon", "ambivalent", "biathlon"}},
			args{"bi", -1, nil},
			[]Autos{
				{Term: "biathlon", Count: 2},
				{Term: "ambivalent", Count: 1},
			},
			false,
		},
		{
			"partial matches",
			fields{[]string{"biathlon", "ambivalent"}},
			args{"bia", -1, nil},
			[]Autos{
				{Term: "biathlon", Count: 1},
			},
			false,
		},
		{
			"frequency sort",
			fields{[]string{
				"one",
				"two", "two",
				"three", "three", "three",
				"four", "four", "four", "four",
			}},
			args{"o", -1, nil},
			[]Autos{
				{Term: "four", Count: 4},
				{Term: "two", Count: 2},
				{Term: "one", Count: 1},
			},
			false,
		},
		{
			"frequency sort with limit",
			fields{[]string{
				"one",
				"two", "two",
				"three", "three", "three",
				"four", "four", "four", "four",
			}},
			args{"o", 2, nil},
			[]Autos{
				{Term: "four", Count: 4},
				{Term: "two", Count: 2},
			},
			false,
		},
		{
			"frequency sort with callback",
			fields{[]string{
				"one",
				"two", "two",
				"three", "three", "three",
				"four", "four", "four", "four",
			}},
			args{"o", 3, func(a Autos) Autos {
				hits := map[string]int{"one": 10, "four": 2}
				if count, ok := hits[a.Term]; ok {
					return Autos{
						Term:     a.Term,
						Count:    a.Count + count,
						Metadata: map[string][]string{"id": {"123"}},
					}
				}
				a.Count = 0
				return a
			}},
			[]Autos{
				{
					Term: "one", Count: 11,
					Metadata: map[string][]string{"id": {"123"}},
				},
				{Term: "four", Count: 6,
					Metadata: map[string][]string{"id": {"123"}},
				},
			},
			false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			ac := NewAutoComplete()
			ac.SuggestFn = tt.args.callback
			ac.FromStrings(tt.fields.words)

			got, err := ac.Suggest(tt.args.input, tt.args.limit)
			if (err != nil) != tt.wantErr {
				t.Errorf("AutoComplete.Suggest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("AutoComplete.Suggest() %s = mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}
