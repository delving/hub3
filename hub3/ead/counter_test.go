package ead

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
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

func TestDescriptionCounter_add(t *testing.T) {
	dc := NewDescriptionCounter()
	type args struct {
		item *DataItem
	}
	tests := []struct {
		name         string
		args         args
		wordCount    int
		maxItemCount int
		wantErr      bool
	}{
		{
			"first simple add",
			args{item: &DataItem{Order: uint64(1), Text: "first data item."}},
			3,
			1,
			false,
		},
		{
			"second add",
			args{item: &DataItem{Order: uint64(2), Text: "second data item"}},
			4,
			2,
			false,
		},
		{
			"re-add second add with case change",
			args{item: &DataItem{Order: uint64(2), Text: "Second Data Item"}},
			4,
			2,
			false,
		},
		{
			"third add",
			args{item: &DataItem{Order: uint64(2), Text: "Third data item"}},
			5,
			3,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := dc.add(tt.args.item); (err != nil) != tt.wantErr {
				t.Errorf("DescriptionCounter.add() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(dc.Counter) != tt.wordCount {
				t.Errorf("DescriptionCounter.add() word count; got = %v, want %v", len(dc.Counter), tt.wordCount)
			}
			if len(dc.DataItemIdx) == 0 {
				t.Errorf("DescriptionCounter.add() dataItemIdx cannot be empty, got: %#v", dc.DataItemIdx)
			}
			for _, v := range dc.DataItemIdx {
				if len(v) > tt.maxItemCount && len(v) != 0 {
					t.Errorf("DescriptionCounter.add() max item count; got = %v, want %v", len(v), tt.maxItemCount)
				}
			}
		})
	}
}

func TestDescriptionCounter_CountForQuery(t *testing.T) {
	type fields struct {
		dataItems []*DataItem
	}
	type args struct {
		query string
	}
	dataItems := []*DataItem{
		&DataItem{Order: uint64(1), Text: "first data item."},
		&DataItem{Order: uint64(2), Text: "second data item amster"},
		&DataItem{Order: uint64(3), Text: "third data item amsterdam."},
	}

	tests := []struct {
		name     string
		fields   fields
		args     args
		want     int
		want1    map[string]int
		idxCount []uint64
	}{
		{
			"simple search",
			fields{dataItems: dataItems},
			args{query: "item"},
			3,
			map[string]int{
				"item": 3,
			},
			[]uint64{uint64(1), uint64(2), uint64(3)},
		},
		{
			"full word search",
			fields{dataItems: dataItems},
			args{query: "amster"},
			1,
			map[string]int{
				"amster": 1,
			},
			[]uint64{uint64(2)},
		},
		{
			"wildcard search",
			fields{dataItems: dataItems},
			args{query: "amster*"},
			2,
			map[string]int{
				"amster":    1,
				"amsterdam": 1,
			},
			[]uint64{uint64(2), uint64(3)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dc := NewDescriptionCounter()
			for _, di := range tt.fields.dataItems {
				err := dc.add(di)
				if err != nil {
					t.Errorf("DescriptionCounter.CountForQuery() adding dataItems should not throw an error; %#v", err)
				}
			}
			got, got1 := dc.CountForQuery(tt.args.query)
			if got != tt.want {
				t.Errorf("DescriptionCounter.CountForQuery() got = %v, want %v", got, tt.want)
			}
			if !cmp.Equal(got1, tt.want1) {
				t.Errorf("DescriptionCounter.CountForQuery() got1 = %v, want %v", got1, tt.want1)
			}
			gotIdxCount := dc.GetDataItemIdx(got1)
			if !cmp.Equal(gotIdxCount, tt.idxCount) {
				t.Errorf("DescriptionCounter.CountForQuery() gotIdxCount = %v, want %v", gotIdxCount, tt.idxCount)
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
