package ead

import "testing"

func TestDescriptionCounter_countForQuery(t *testing.T) {
	dc := newDescriptionCounter([]byte("Inventaris van het archief van de voc Verenigde Oost-Indische Compagnie (VOC), 1602-1795 (1811)"))
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
			fields{counter: dc.counter},
			args{query: "voc"},
			2,
		},
		{
			"hyphenated count",
			fields{counter: dc.counter},
			args{query: "1602-1795"},
			1,
		},
		{
			"hyphenated sub-count",
			fields{counter: dc.counter},
			args{query: "1795"},
			1,
		},
		{
			"parathesis count",
			fields{counter: dc.counter},
			args{query: "1811"},
			1,
		},
		{
			"parathesis query",
			fields{counter: dc.counter},
			args{query: "(VOC)"},
			2,
		},
		{
			"multiword count",
			fields{counter: dc.counter},
			args{query: "voc archief"},
			3,
		},
		{
			"wildcard count",
			fields{counter: dc.counter},
			args{query: "v*"},
			5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dc := DescriptionCounter{
				counter: tt.fields.counter,
			}
			if got, _ := dc.countForQuery(tt.args.query); got != tt.want {
				t.Errorf("DescriptionCounter.countForQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}
