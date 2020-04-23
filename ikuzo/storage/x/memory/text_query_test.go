// nolint:gocritic
package memory

import (
	"testing"

	"github.com/delving/hub3/ikuzo/service/x/search"
	"github.com/google/go-cmp/cmp"
	"github.com/matryer/is"
)

func TestTextQuery_Highlight(t *testing.T) {
	is := is.New(t)

	type fields struct {
		q    string
		hits map[string]int
	}

	type args struct {
		text string
	}

	tests := []struct {
		name     string
		fields   fields
		args     args
		want     string
		wantHits bool
	}{
		{
			"no hits",
			fields{"one", map[string]int{}},
			args{text: "not 1"},
			"not 1",
			false,
		},
		{
			"index error",
			fields{"one", map[string]int{}},
			args{text: ""},
			"",
			false,
		},
		{
			"one hit",
			fields{"one", map[string]int{}},
			args{text: "only one"},
			"only <em class=\"dchl\">one</em>",
			true,
		},
		{
			"asciifolding hit",
			fields{"prive", map[string]int{}},
			args{text: "very privé"},
			"very <em class=\"dchl\">privé</em>",
			true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			tq, err := NewTextQueryFromString(tt.fields.q)
			is.NoErr(err)
			id := tq.ti.setDocID()

			err = tq.AppendString(tt.args.text, id)
			is.NoErr(err)

			_, err = tq.PerformSearch()
			is.NoErr(err)

			got, got1 := tq.Highlight(tt.args.text, id)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("TextQuery.Highlight() %s = mismatch (-want +got):\n%s", tt.name, diff)
			}

			if got1 != tt.wantHits {
				t.Errorf("TextQuery.Highlight() %s = got1 %v, want %v", tt.name, got1, tt.wantHits)
			}
		})
	}
}

func Test_hightlightWithVectors(t *testing.T) {
	type args struct {
		text    string
		docID   int
		vectors []termVector
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"no hightlights",
			args{
				text:    "hello world",
				docID:   0,
				vectors: nil,
			},
			"hello world",
		},
		{
			"one word highlight",
			args{
				text:  "hello world",
				docID: 1,
				vectors: []termVector{
					{"wold", []testVector{{DocID: 1, Location: 2}}},
				},
			},
			"hello <em class=\"dchl\">world</em>",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			tq := NewTextQuery(nil)
			tq.ti.setDocID()

			tv := search.NewVectors()
			for _, v := range tt.args.vectors {
				for _, vector := range v.vectors {
					tv.AddVector(vector.searchVector())
				}
			}

			got := tq.hightlightWithVectors(tt.args.text, 1, tv)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("hightlightWithVectors() %s = mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}
