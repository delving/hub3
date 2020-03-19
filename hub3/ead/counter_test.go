package ead

import (
	"testing"
)

// nolint:funlen
func Test_descriptionQuery_highlightQuery(t *testing.T) {
	type fields struct {
		query string
	}

	type args struct {
		text string
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
		want1  bool
	}{
		{
			"simple query",
			fields{query: "de"},
			args{text: "De grootste man is niet de oudste"},
			`<em class="dchl">De</em> grootste man is niet <em class="dchl">de</em> oudste`,
			true,
		},
		{
			"multiword query",
			fields{query: "de amsterdam"},
			args{text: "De boot in Amsterdam"},
			`<em class="dchl">De</em> boot in <em class="dchl">Amsterdam</em>`,
			true,
		},
		{
			"wildcard query",
			fields{query: "de amster*"},
			args{text: "De boot in Amsterdam"},
			`<em class="dchl">De</em> boot in <em class="dchl">Amsterdam</em>`,
			true,
		},
		{
			"negative query",
			fields{query: "-amsterdam"},
			args{text: "De boot in Amsterdam"},
			"De boot in Amsterdam",
			false,
		},
		{
			"multi word phrase",
			fields{query: "Henny van Schie"},
			args{text: "Henny van Schie, rol"},
			`<em class="dchl">Henny van Schie,</em> rol`,
			true,
		},
		{
			"multi word match phrase",
			fields{query: "\"Henny van Schie\""},
			args{text: "Henny van Schie, rol"},
			`<em class="dchl">Henny van Schie,</em> rol`,
			true,
		},
		{
			"diacritics: diacritic to diacritic",
			fields{query: "Geünieerde"},
			args{text: "de Geünieerde Provintiën"},
			`de <em class="dchl">Geünieerde</em> Provintiën`,
			true,
		},
		{
			"diacritics: diacritic to normalised",
			fields{query: "Geünieerde"},
			args{text: "de Geunieerde Provintiën"},
			`de <em class="dchl">Geunieerde</em> Provintiën`,
			true,
		},
		{
			"diacritics: normalised to diacritic",
			fields{query: "Geunieerde"},
			args{text: "de Geünieerde Provintiën"},
			`de <em class="dchl">Geünieerde</em> Provintiën`,
			true,
		},
		{
			"no match",
			fields{query: "leiden"},
			args{text: "De boot in Amsterdam"},
			"De boot in Amsterdam",
			false,
		},
		{
			"multiword match with period",
			fields{query: "mr. Joan Blaeu"},
			args{text: "zijn zoon, mr. Joan Blaeu, "},
			"zijn zoon, <em class=\"dchl\">mr. Joan Blaeu,</em>",
			true,
		},
		{
			"phrase match with period",
			fields{query: "\"mr. Joan Blaeu\""},
			args{text: "zijn zoon, mr. Joan Blaeu, "},
			"zijn zoon, <em class=\"dchl\">mr. Joan Blaeu,</em>",
			true,
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			dq := NewDescriptionQuery(tt.fields.query)

			got, got1 := dq.highlightQuery(tt.args.text)
			if got != tt.want {
				t.Errorf("descriptionQuery.highlightQuery() %s; got = %v, want %v", tt.name, got, tt.want)
			}

			if got1 != tt.want1 {
				t.Errorf("descriptionQuery.highlightQuery() %s; got1 = %v, want %v", tt.name, got1, tt.want1)
			}
		})
	}
}

func Test_descriptionQuery_match(t *testing.T) {
	type fields struct {
		query string
	}

	type args struct {
		word string
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantWord string
		want     bool
	}{
		{
			"simple query",
			fields{query: "de"},
			args{word: "De"},
			"<em class=\"dchl\">De</em>",
			true,
		},
		{
			"lowercase match",
			fields{query: "De"},
			args{word: "de"},
			"<em class=\"dchl\">de</em>",
			true,
		},
		{
			"wildcard match multi world",
			fields{query: "amster* de"},
			args{word: "Amsterdam"},
			"<em class=\"dchl\">Amsterdam</em>",
			true,
		},
		{
			"no match",
			fields{query: "amster*"},
			args{word: "de"},
			"de",
			false,
		},
		{
			"diacritics match ",
			fields{query: "geunieerde"},
			args{word: "Geünieerde"},
			"<em class=\"dchl\">Geünieerde</em>",
			true,
		},
		{
			"inverted diacritics match ",
			fields{query: "geünieerde"},
			args{word: "Geunieerde"},
			"<em class=\"dchl\">Geunieerde</em>",
			true,
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			dq := NewDescriptionQuery(tt.fields.query)
			gotWord, got := dq.highlightQuery(tt.args.word)
			dq.tq.EmStartTag = ""
			dq.tq.EmEndTag = ""

			if got != tt.want {
				t.Errorf("descriptionQuery.match() %s = %v, want %v", tt.name, got, tt.want)
			}

			if gotWord != tt.wantWord {
				t.Errorf("descriptionQuery.match() %s = %v, want %v", tt.name, gotWord, tt.wantWord)
			}
		})
	}
}
