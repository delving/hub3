package search

import (
	"bytes"
	"io"
	"strings"
	"text/scanner"
	"unicode"
)

type Token struct {
	Position      int
	WordPosition  int
	OffsetStart   int
	OffsetEnd     int
	Ignored       bool
	RawText       string
	Normal        string
	TrailingSpace bool
	Punctuation   bool
}

func (t Token) incWordPos() bool {
	return !t.Punctuation && !t.Ignored
}

type Tokenizer struct {
	s        *scanner.Scanner
	a        Analyzer
	wordPos  int
	tokenPos int
}

func NewTokenizer() *Tokenizer {
	return &Tokenizer{}
}

func (t *Tokenizer) ParseString(text string) *TokenStream {
	return t.parse(strings.NewReader(text))
}

func (t *Tokenizer) ParseBytes(b []byte) *TokenStream {
	return t.parse(bytes.NewReader(b))
}

func (t *Tokenizer) ParseReader(r io.Reader) *TokenStream {
	return t.parse(r)
}

func (t *Tokenizer) resetScanner() {
	var s scanner.Scanner
	t.s = &s
	t.s.IsIdentRune = isIdentRune
	t.wordPos = 0
	t.tokenPos = 0
}

func (t *Tokenizer) parse(r io.Reader) *TokenStream {
	t.resetScanner()
	t.s.Init(r)

	tokens := []Token{}
	for tok := t.s.Scan(); tok != scanner.EOF; tok = t.s.Scan() {
		tokens = append(tokens, t.runParser(tok))
	}

	return &TokenStream{tokens: tokens}
}

func (t *Tokenizer) takeTag(text string) string {
	var str strings.Builder

	str.WriteString(text)

	for ttok := t.s.Scan(); ttok != scanner.EOF; ttok = t.s.Scan() {
		chars := t.s.TokenText()
		str.WriteString(chars)

		if unicode.IsSpace(t.s.Peek()) {
			str.WriteString(" ")
		}

		if chars == ">" {
			break
		}
	}

	return str.String()
}

func (t *Tokenizer) runParser(tok rune) Token {
	t.tokenPos++
	text := t.s.TokenText()

	var ignored bool

	pos := t.s.Position

	if text == "<" {
		ignored = true

		text = t.takeTag(text)
	}

	token := Token{
		RawText:     text,
		Position:    t.tokenPos,
		OffsetStart: pos.Offset,
		OffsetEnd:   pos.Offset + (len(text) - 1),
		Ignored:     ignored,
	}

	if unicode.IsSpace(t.s.Peek()) {
		token.TrailingSpace = true
	}

	if len(text) == 1 && unicode.IsPunct(tok) {
		token.Punctuation = true
		token.Ignored = true
	}

	if !token.Ignored {
		token.Normal = t.a.Transform(token.RawText)
	}

	if token.incWordPos() {
		t.wordPos++
		token.WordPosition = t.wordPos
	}

	return token
}

type TokenStream struct {
	tokens []Token
}

func (ts *TokenStream) Tokens() []Token {
	return ts.tokens
}

func (ts *TokenStream) String() string {
	var str strings.Builder

	for _, token := range ts.tokens {
		str.WriteString(token.RawText)

		if token.TrailingSpace {
			str.WriteString(" ")
		}
	}

	return str.String()
}

func setDefaultTags(startTag, endTag string) (start, end string) {
	if startTag == "" {
		startTag = "<em>"
	}

	if endTag == "" {
		endTag = "</em>"
	}

	return startTag, endTag
}

func (ts *TokenStream) Highlight(positions map[int]bool, startTag, endTag string) string {
	if len(positions) == 0 {
		return ts.String()
	}

	startTag, endTag = setDefaultTags(startTag, endTag)

	var str strings.Builder

	var inHighlight bool

	var insertSpace bool

	for _, token := range ts.tokens {
		_, ok := positions[token.WordPosition]
		if !ok && inHighlight && !token.Ignored {
			str.WriteString(endTag)
		}

		if insertSpace {
			insertSpace = false

			str.WriteString(" ")
		}

		if ok && !inHighlight {
			inHighlight = true

			str.WriteString(startTag)
		}

		str.WriteString(token.RawText)

		if token.TrailingSpace {
			insertSpace = true
		}

		if !ok && inHighlight && !token.Ignored {
			inHighlight = false
		}
	}

	if inHighlight {
		str.WriteString(endTag)
	}

	return str.String()
}
