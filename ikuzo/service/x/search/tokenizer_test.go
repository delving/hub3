package search

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/matryer/is"
)

type targs struct {
	text string
}

var tokenTests = []struct {
	name string
	args targs
	want *TokenStream
}{
	{
		"single word",
		targs{"word"},
		&TokenStream{
			[]Token{
				{Position: 1, WordPosition: 1, OffsetStart: 0, OffsetEnd: 3, RawText: "word", Normal: "word"},
			},
		},
	},
	{
		"normalized word",
		targs{"Word..."},
		&TokenStream{
			[]Token{
				{
					Position:     1,
					WordPosition: 1,
					OffsetStart:  0,
					OffsetEnd:    6,
					RawText:      "Word...",
					Normal:       "word",
				},
			},
		},
	},
	{
		"tag without attribute",
		targs{"<p>"},
		&TokenStream{
			[]Token{
				{
					Position:    1,
					OffsetStart: 0,
					OffsetEnd:   2,
					RawText:     "<p>",
					Normal:      "",
					Ignored:     true,
				},
			},
		},
	},
	{
		"tag with attribute",
		targs{"<p href=\"http://example.com\">"},
		&TokenStream{
			[]Token{
				{
					Position:    1,
					OffsetStart: 0,
					OffsetEnd:   28,
					RawText:     "<p href=\"http://example.com\">",
					Normal:      "",
					Ignored:     true,
				},
			},
		},
	},
	{
		"tagged words",
		targs{"<p href=\"http://example.com\">mr. Joan Blaeu</p>"},
		&TokenStream{
			[]Token{
				{
					Position:    1,
					OffsetStart: 0,
					OffsetEnd:   28,
					RawText:     "<p href=\"http://example.com\">",
					Normal:      "",
					Ignored:     true,
				},
				{Position: 2, WordPosition: 1, OffsetStart: 29, OffsetEnd: 31, RawText: "mr.", Normal: "mr", TrailingSpace: true},
				{Position: 3, WordPosition: 2, OffsetStart: 33, OffsetEnd: 36, RawText: "Joan", Normal: "joan", TrailingSpace: true},
				{Position: 4, WordPosition: 3, OffsetStart: 38, OffsetEnd: 42, RawText: "Blaeu", Normal: "blaeu"},
				{Position: 5, OffsetStart: 43, OffsetEnd: 46, Ignored: true, RawText: "</p>"},
			},
		},
	},
	{
		"sentence",
		targs{"Really, are you serious?"},
		&TokenStream{
			[]Token{
				{Position: 1, WordPosition: 1, OffsetEnd: 5, RawText: "Really", Normal: "really"},
				{
					Position:      2,
					OffsetStart:   6,
					OffsetEnd:     6,
					RawText:       ",",
					TrailingSpace: true,
					Punctuation:   true,
					Ignored:       true,
				},
				{
					Position:      3,
					WordPosition:  2,
					OffsetStart:   8,
					OffsetEnd:     10,
					RawText:       "are",
					Normal:        "are",
					TrailingSpace: true,
				},
				{
					Position:      4,
					WordPosition:  3,
					OffsetStart:   12,
					OffsetEnd:     14,
					RawText:       "you",
					Normal:        "you",
					TrailingSpace: true,
				},
				{
					Position:     5,
					WordPosition: 4,
					OffsetStart:  16,
					OffsetEnd:    22,
					RawText:      "serious",
					Normal:       "serious",
				},
				{Position: 6, OffsetStart: 23, OffsetEnd: 23, RawText: "?", Punctuation: true, Ignored: true},
			},
		},
	},
}

func TestTokenizer_parse(t *testing.T) {
	for _, tt := range tokenTests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			tok := NewTokenizer()

			got := tok.ParseString(tt.args.text)

			if diff := cmp.Diff(tt.want, got, cmp.AllowUnexported(TokenStream{})); diff != "" {
				t.Errorf("tokenizer.parse() %s = mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}

// nolint:gocritic
func TestTokenizer_Parsers(t *testing.T) {
	is := is.New(t)

	want := &TokenStream{[]Token{{Position: 1, WordPosition: 1, OffsetStart: 0, OffsetEnd: 3, RawText: "word", Normal: "word"}}}

	tok := NewTokenizer()

	got := tok.ParseBytes([]byte("word"))
	is.Equal(got, want)

	got = tok.ParseReader(strings.NewReader("word"))
	is.Equal(got, want)
}

func TestTokenStream_Highlight(t *testing.T) {
	type fields struct {
		text string
	}

	type args struct {
		positions map[int]bool
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			"no hits two words",
			fields{text: "two words"},
			args{},
			"two words",
		},
		{
			"two words first hit",
			fields{text: "two words"},
			args{map[int]bool{1: true}},
			"<em>two</em> words",
		},
		{
			"two words second hit",
			fields{text: "two words"},
			args{map[int]bool{2: true}},
			"two <em>words</em>",
		},
		{
			"phrase hit",
			fields{text: "two, words. no hit!"},
			args{map[int]bool{1: true, 2: true}},
			"<em>two, words.</em> no hit!",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			tok := NewTokenizer()
			tokens := tok.ParseString(tt.fields.text)

			got := tokens.Highlight(tt.args.positions)

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("TokenStream.Highlight() %s = mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}
