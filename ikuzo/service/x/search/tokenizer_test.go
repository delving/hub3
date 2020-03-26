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
				{Vector: 1, TermVector: 1, OffsetStart: 0, OffsetEnd: 3, RawText: "word", Normal: "word", DocID: 1},
			},
		},
	},
	{
		"error with single quote",
		targs{"op 's"},
		&TokenStream{
			[]Token{
				{Vector: 1, TermVector: 1, OffsetStart: 0, OffsetEnd: 1, RawText: "op", Normal: "op", DocID: 1, TrailingSpace: true},
				{Vector: 2, TermVector: 2, OffsetStart: 3, OffsetEnd: 4, RawText: "'s", Normal: "s", DocID: 1},
			},
		},
	},
	{
		"normalized word",
		targs{"Word..."},
		&TokenStream{
			[]Token{
				{
					Vector:      1,
					TermVector:  1,
					OffsetStart: 0,
					OffsetEnd:   6,
					RawText:     "Word...",
					Normal:      "word",
					DocID:       1,
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
					Vector:      1,
					OffsetStart: 0,
					OffsetEnd:   2,
					RawText:     "<p>",
					Normal:      "",
					Ignored:     true,
					DocID:       1,
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
					Vector:      1,
					OffsetStart: 0,
					OffsetEnd:   28,
					RawText:     "<p href=\"http://example.com\">",
					Normal:      "",
					Ignored:     true,
					DocID:       1,
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
					Vector:      1,
					OffsetStart: 0,
					OffsetEnd:   28,
					RawText:     "<p href=\"http://example.com\">",
					Normal:      "",
					Ignored:     true,
					DocID:       1,
				},
				{Vector: 2, TermVector: 1, OffsetStart: 29, OffsetEnd: 31,
					RawText: "mr.", Normal: "mr", TrailingSpace: true, DocID: 1},
				{Vector: 3, TermVector: 2, OffsetStart: 33, OffsetEnd: 36,
					RawText: "Joan", Normal: "joan", TrailingSpace: true, DocID: 1},
				{Vector: 4, TermVector: 3, OffsetStart: 38, OffsetEnd: 42,
					RawText: "Blaeu", Normal: "blaeu", DocID: 1},
				{Vector: 5, OffsetStart: 43, OffsetEnd: 46, Ignored: true,
					RawText: "</p>", DocID: 1},
			},
		},
	},
	{
		"mixed tags",
		targs{
			"<p><persname>Christoffel Plantijn</persname>, <unitdate calendar=\"gregorian\" era=\"ce\" normal=\"1584\"> 1584</unitdate> </p>",
		},
		&TokenStream{
			[]Token{
				{Vector: 1, OffsetEnd: 2, Ignored: true, RawText: "<p>", DocID: 1},
				{Vector: 2, OffsetStart: 3, OffsetEnd: 12, Ignored: true, RawText: "<persname>", DocID: 1},
				{
					Vector:        3,
					TermVector:    1,
					OffsetStart:   13,
					OffsetEnd:     23,
					RawText:       "Christoffel",
					Normal:        "christoffel",
					TrailingSpace: true,
					DocID:         1},
				{
					Vector:      4,
					TermVector:  2,
					OffsetStart: 25,
					OffsetEnd:   32,
					RawText:     "Plantijn",
					Normal:      "plantijn",
					DocID:       1},
				{
					Vector:      5,
					OffsetStart: 33,
					OffsetEnd:   43,
					Ignored:     true,
					RawText:     "</persname>",
					DocID:       1},
				{
					Vector:        6,
					OffsetStart:   44,
					OffsetEnd:     44,
					Ignored:       true,
					RawText:       ",",
					TrailingSpace: true,
					Punctuation:   true,
					DocID:         1},
				{
					Vector:        7,
					OffsetStart:   46,
					OffsetEnd:     100,
					Ignored:       true,
					RawText:       `<unitdate calendar="gregorian" era="ce" normal="1584"> `,
					TrailingSpace: true,
					DocID:         1},
				{
					Vector:      8,
					TermVector:  3,
					OffsetStart: 101,
					OffsetEnd:   104,
					RawText:     "1584",
					Normal:      "1584",
					DocID:       1},
				{
					Vector:        9,
					OffsetStart:   105,
					OffsetEnd:     116,
					Ignored:       true,
					RawText:       "</unitdate> ",
					TrailingSpace: true,
					DocID:         1},
				{Vector: 10, OffsetStart: 117, OffsetEnd: 120, Ignored: true, RawText: "</p>", DocID: 1},
			},
		},
	},
	{
		"sentence",
		targs{"Really, are you serious?"},
		&TokenStream{
			[]Token{
				{Vector: 1, TermVector: 1, OffsetEnd: 5, RawText: "Really", Normal: "really", DocID: 1},
				{
					Vector:        2,
					OffsetStart:   6,
					OffsetEnd:     6,
					RawText:       ",",
					TrailingSpace: true,
					Punctuation:   true,
					Ignored:       true,
					DocID:         1},
				{
					Vector:        3,
					TermVector:    2,
					OffsetStart:   8,
					OffsetEnd:     10,
					RawText:       "are",
					Normal:        "are",
					TrailingSpace: true,
					DocID:         1},
				{
					Vector:        4,
					TermVector:    3,
					OffsetStart:   12,
					OffsetEnd:     14,
					RawText:       "you",
					Normal:        "you",
					TrailingSpace: true,
					DocID:         1},
				{
					Vector:      5,
					TermVector:  4,
					OffsetStart: 16,
					OffsetEnd:   22,
					RawText:     "serious",
					Normal:      "serious",
					DocID:       1},
				{Vector: 6, OffsetStart: 23, OffsetEnd: 23, RawText: "?", Punctuation: true, Ignored: true, DocID: 1},
			},
		},
	},
}

func TestTokenizer_parse(t *testing.T) {
	for _, tt := range tokenTests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			tok := NewTokenizer()

			got := tok.Parse(strings.NewReader(tt.args.text), 0)

			if diff := cmp.Diff(tt.want, got, cmp.AllowUnexported(TokenStream{})); diff != "" {
				t.Errorf("tokenizer.parse() %s = mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}

// nolint:gocritic
func TestTokenizer_Parsers(t *testing.T) {
	want := &TokenStream{[]Token{{Vector: 1, TermVector: 1, OffsetStart: 0, OffsetEnd: 3,
		RawText: "word", Normal: "word", DocID: 1}}}

	tok := NewTokenizer()

	got := tok.ParseBytes([]byte("word"), 1)
	if diff := cmp.Diff(want, got, cmp.AllowUnexported(TokenStream{})); diff != "" {
		t.Errorf("tokenizer.parse() %s = mismatch (-want +got):\n%s", "parseBytes", diff)
	}

	got = tok.ParseString("word", 1)
	if diff := cmp.Diff(want, got, cmp.AllowUnexported(TokenStream{})); diff != "" {
		t.Errorf("tokenizer.parse() %s = mismatch (-want +got):\n%s", "parseString", diff)
	}

	if diff := cmp.Diff(got.tokens, got.Tokens(), cmp.AllowUnexported(Token{})); diff != "" {
		t.Errorf("tokenizer.Tokens() %s = mismatch (-want +got):\n%s", "Tokens()", diff)
	}
}

// nolint:gocritic
func TestParseError(t *testing.T) {
	is := is.New(t)

	tok := NewTokenizer(SetPhraseAware())

	is.Equal(len(tok.errors), 0)

	_ = tok.ParseString("\"word", 1)

	is.Equal(len(tok.errors), tok.s.ErrorCount)
	is.Equal(len(tok.errors), 1)
}

func TestTokenStream_Highlight(t *testing.T) {
	type fields struct {
		text string
	}

	type args struct {
		positions []int
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
			args{[]int{1}},
			"<em>two</em> words",
		},
		{
			"two words second hit",
			fields{text: "two words"},
			args{[]int{2}},
			"two <em>words</em>",
		},
		{
			"phrase hit",
			fields{text: "two, words. no hit!"},
			args{[]int{1, 2}},
			"<em>two, words.</em> no hit!",
		},
		{
			"phrase hit",
			fields{text: "<p>two, words.</p> no hit!"},
			args{[]int{1, 2}},
			"<p><em>two, words.</p> </em> no hit!",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			tok := NewTokenizer()
			tokens := tok.Parse(strings.NewReader(tt.fields.text), 0)

			terms := NewVectors()

			for _, v := range tt.args.positions {
				terms.Add(1, v)
			}

			got := tokens.Highlight(terms, "", "")

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("TokenStream.Highlight() %s = mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}
