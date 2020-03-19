// nolint:gocritic
package memory

import (
	"testing"

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
		name   string
		fields fields
		args   args
		want   string
		want1  bool
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

			got, got1 := tq.Highlight(tt.args.text)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("TextQuery.Highlight() %s = mismatch (-want +got):\n%s", tt.name, diff)
			}

			if got1 != tt.want1 {
				t.Errorf("TextQuery.Highlight() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_hightlightWithVectors(t *testing.T) {
	type args struct {
		text      string
		positions map[int]bool
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"no hightlights",
			args{
				text:      "hello world",
				positions: map[int]bool{},
			},
			"hello world",
		},
		{
			"one word highlight",
			args{
				text:      "hello world",
				positions: map[int]bool{2: true},
			},
			"hello <em class=\"dchl\">world</em>",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			tq := NewTextQuery(nil)

			got := tq.hightlightWithVectors(tt.args.text, tt.args.positions)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("hightlightWithVectors() %s = mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}
