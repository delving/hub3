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
	want []Token
}{
	{
		"single word",
		targs{"word"},
		[]Token{
			{Position: 1, WordPosition: 1, OffsetStart: 0, OffsetEnd: 3, RawText: "word", Normal: "word"},
		},
	},
	{
		"normalized word",
		targs{"Word..."},
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
	{
		"tag without attribute",
		targs{"<p>"},
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
	{
		"tag with attribute",
		targs{"<p href=\"http://example.com\">"},
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
	{
		"tagged words",
		targs{"<p href=\"http://example.com\">mr. Joan Blaeu</p>"},
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
	{
		"sentence",
		targs{"Really, are you serious?"},
		[]Token{
			{Position: 1, WordPosition: 1, OffsetEnd: 5, RawText: "Really", Normal: "really"},
			{
				Position:      2,
				OffsetStart:   6,
				OffsetEnd:     6,
				RawText:       ",",
				TrailingSpace: true,
				Punctuation:   true,
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
			{Position: 6, OffsetStart: 23, OffsetEnd: 23, RawText: "?", Punctuation: true},
		},
	},
}

func TestTokenizer_parse(t *testing.T) {
	for _, tt := range tokenTests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			tok := NewTokenizer()

			got := tok.ParseString(tt.args.text)

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("tokenizer.parse() %s = mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}

// nolint:gocritic
func TestTokenizer_Parsers(t *testing.T) {
	is := is.New(t)

	want := []Token{{Position: 1, WordPosition: 1, OffsetStart: 0, OffsetEnd: 3, RawText: "word", Normal: "word"}}

	tok := NewTokenizer()

	got := tok.ParseBytes([]byte("word"))
	is.Equal(got, want)

	got = tok.ParseReader(strings.NewReader("word"))
	is.Equal(got, want)
}
