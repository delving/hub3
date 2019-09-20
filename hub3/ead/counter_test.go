package ead

import (
	"reflect"
	"testing"
)

func TestDescriptionCounter_countForQuery(t *testing.T) {
	dc := NewDescriptionCounter()
	err := dc.AppendString(
		"Inventaris van het archief van de voc",
	)
	if err != nil {
		t.Errorf("unable to append bytes: %#v", err)
	}
	err = dc.AppendBytes(
		[]byte("Verenigde Oost-Indische Compagnie (VOC), 1602-1795 (1811)"),
	)
	if err != nil {
		t.Errorf("unable to append bytes: %#v", err)
	}
	//log.Printf("%#v", dc)
	type fields struct {
		counter map[string]int
	}
	type args struct {
		query string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			"one word count",
			fields{counter: dc.Counter},
			args{query: "voc"},
			2,
		},
		{
			"hyphenated count",
			fields{counter: dc.Counter},
			args{query: "1602-1795"},
			1,
		},
		{
			"hyphenated sub-count",
			fields{counter: dc.Counter},
			args{query: "1795"},
			1,
		},
		{
			"parathesis count",
			fields{counter: dc.Counter},
			args{query: "1811"},
			1,
		},
		{
			"parathesis query",
			fields{counter: dc.Counter},
			args{query: "(VOC)"},
			2,
		},
		{
			"multiword count",
			fields{counter: dc.Counter},
			args{query: "voc archief"},
			3,
		},
		{
			"wildcard count",
			fields{counter: dc.Counter},
			args{query: "v*"},
			5,
		},
		{
			"boolean count",
			fields{counter: dc.Counter},
			args{query: "1811 AND (VOC)"},
			3,
		},
		{
			"not boolean count",
			fields{counter: dc.Counter},
			args{query: "1811 -1795"},
			1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dc := DescriptionCounter{
				Counter: tt.fields.counter,
			}
			if got, _ := dc.CountForQuery(tt.args.query); got != tt.want {
				t.Errorf("DescriptionCounter.countForQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newQueryItem(t *testing.T) {
	type args struct {
		word string
	}
	tests := []struct {
		name  string
		args  args
		want  *queryItem
		want1 bool
	}{
		{
			"simple query",
			args{"amsterdam"},
			&queryItem{text: "amsterdam", wildcard: false},
			true,
		},
		{
			"wildcard query",
			args{"amster*"},
			&queryItem{text: "amster", wildcard: true},
			true,
		},
		{
			"bool query and",
			args{"AND"},
			nil,
			false,
		},
		{
			"exclusion query",
			args{"-monster"},
			nil,
			false,
		},
		{
			"diacritics query",
			args{"Geünieerde"},
			&queryItem{text: "geünieerde", wildcard: false, flat: "geunieerde"},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := newQueryItem(tt.args.word)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newQueryItem() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("newQueryItem() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_descriptionQuery_highlightQuery(t *testing.T) {
	type fields struct {
		query string
	}
	type args struct {
		text string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		want1   bool
		partial bool
	}{
		{
			"simple query",
			fields{query: "de"},
			args{text: "De grootste man is niet de oudste"},
			`<em class="dchl">De</em> grootste man is niet <em class="dchl">de</em> oudste`,
			true,
			false,
		},
		{
			"multiword query",
			fields{query: "de amsterdam"},
			args{text: "De boot in Amsterdam"},
			`<em class="dchl">De</em> boot in <em class="dchl">Amsterdam</em>`,
			true,
			false,
		},
		{
			"wildcard query",
			fields{query: "de amster*"},
			args{text: "De boot in Amsterdam"},
			`<em class="dchl">De</em> boot in <em class="dchl">Amsterdam</em>`,
			true,
			false,
		},
		{
			"negative query",
			fields{query: "-amsterdam"},
			args{text: "De boot in Amsterdam"},
			"De boot in Amsterdam",
			false,
			false,
		},
		{
			"no partial match",
			fields{query: "rol"},
			args{text: "zeemontsterrolen voor de rol"},
			"zeemontsterrolen voor de <em class=\"dchl\">rol</em>",
			true,
			false,
		},
		{
			"partial match",
			fields{query: "rol"},
			args{text: "zeemontsterrolen voor de rol"},
			"zeemontster<em class=\"dchl\">rol</em>en voor de <em class=\"dchl\">rol</em>",
			true,
			true,
		},
		{
			"multi word phrase",
			fields{query: "Henny van Schie"},
			args{text: "Henny van Schie, rol"},
			`<em class="dchl">Henny</em> <em class="dchl">van</em> <em class="dchl">Schie</em>, rol`,
			true,
			false,
		},
		{
			"diacritics: diacritic to diacritic",
			fields{query: "Geünieerde"},
			args{text: "de Geünieerde Provintiën"},
			`de <em class="dchl">Geünieerde</em> Provintiën`,
			true,
			false,
		},
		{
			"diacritics: diacritic to normalised",
			fields{query: "Geünieerde"},
			args{text: "de Geunieerde Provintiën"},
			`de <em class="dchl">Geunieerde</em> Provintiën`,
			true,
			false,
		},
		{
			"diacritics: normalised to diacritic",
			fields{query: "Geunieerde"},
			args{text: "de Geünieerde Provintiën"},
			`de <em class="dchl">Geünieerde</em> Provintiën`,
			true,
			false,
		},
		{
			"no match",
			fields{query: "leiden"},
			args{text: "De boot in Amsterdam"},
			"De boot in Amsterdam",
			false,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dq := NewDescriptionQuery(tt.fields.query)
			if tt.partial {
				dq.Partial = true
			}
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
			"de",
			true,
		},
		{
			"lowercase match",
			fields{query: "De"},
			args{word: "de"},
			"de",
			true,
		},
		{
			"wildcard match multi world",
			fields{query: "amster* de"},
			args{word: "Amsterdam"},
			"amsterdam",
			true,
		},
		{
			"no match",
			fields{query: "amster*"},
			args{word: "de"},
			"",
			false,
		},
		{
			"diacritics match ",
			fields{query: "geunieerde"},
			args{word: "Geünieerde"},
			"geünieerde",
			true,
		},
		{
			"inverted diacritics match ",
			fields{query: "geünieerde"},
			args{word: "Geunieerde"},
			"geunieerde",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dq := NewDescriptionQuery(tt.fields.query)
			gotWord, got := dq.match(tt.args.word)
			if got != tt.want {
				t.Errorf("descriptionQuery.match() %s = %v, want %v", tt.name, got, tt.want)
			}
			if gotWord != tt.wantWord {
				t.Errorf("descriptionQuery.match() %s = %v, want %v", tt.name, gotWord, tt.wantWord)
			}
		})
	}
}

func Test_queryItem_equal(t *testing.T) {
	type fields struct {
		text          string
		wildcard      bool
		partial       bool
		diacriticFold bool
		flat          string
		must          bool
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
			"no match",
			fields{
				text: "panorama",
			},
			args{word: "panoramas"},
			"",
			false,
		},
		{
			"direct match",
			fields{
				text: "panorama",
			},
			args{word: "panorama"},
			"panorama",
			true,
		},
		{
			"wildcard match",
			fields{
				text:     "pano",
				wildcard: true,
			},
			args{word: "panorama"},
			"panorama",
			true,
		},
		{
			"partial match",
			fields{
				text:    "oram",
				partial: true,
			},
			args{word: "panorama"},
			"oram",
			true,
		},
		{
			"diacritics on query",
			fields{
				text:          "geünieerde",
				diacriticFold: true,
				flat:          "geunieerde",
			},
			args{word: "geunieerde"},
			"geunieerde",
			true,
		},
		{
			"diacritics on input",
			fields{
				text: "geunieerde",
			},
			args{word: "geünieerde"},
			"geünieerde",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qi := &queryItem{
				text:          tt.fields.text,
				wildcard:      tt.fields.wildcard,
				partial:       tt.fields.partial,
				diacriticFold: tt.fields.diacriticFold,
				flat:          tt.fields.flat,
				must:          tt.fields.must,
			}
			gotWord, got := qi.equal(tt.args.word)
			if got != tt.want {
				t.Errorf("queryItem.equal() %s = got %v, want %v", tt.name, got, tt.want)
			}
			if gotWord != tt.wantWord {
				t.Errorf("queryItem.equal() %s = gotWord %v, wantWord %v", tt.name, gotWord, tt.wantWord)
			}
		})
	}
}
