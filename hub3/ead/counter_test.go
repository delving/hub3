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
				t.Errorf("descriptionQuery.highlightQuery() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("descriptionQuery.highlightQuery() got1 = %v, want %v", got1, tt.want1)
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
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			"simple query",
			fields{query: "de"},
			args{word: "De"},
			true,
		},
		{
			"lowercase match",
			fields{query: "De"},
			args{word: "de"},
			true,
		},
		{
			"wildcard match multi world",
			fields{query: "amster* de"},
			args{word: "Amsterdam"},
			true,
		},
		{
			"no match",
			fields{query: "amster*"},
			args{word: "de"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dq := NewDescriptionQuery(tt.fields.query)
			_, got := dq.match(tt.args.word)
			if got != tt.want {
				t.Errorf("descriptionQuery.match() = %v, want %v", got, tt.want)
			}
		})
	}
}
