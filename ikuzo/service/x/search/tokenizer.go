// Copyright 2020 Delving B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package search

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"
	"text/scanner"
	"unicode"
)

type Token struct {
	Vector        int
	TermVector    int
	OffsetStart   int
	OffsetEnd     int
	Ignored       bool
	RawText       string
	Normal        string
	TrailingSpace bool
	Punctuation   bool
	DocID         int
}

func (t *Token) isTermVector() bool {
	return !t.Punctuation && !t.Ignored
}

func (t *Token) GetTermVector() Vector {
	return Vector{
		Location: t.TermVector,
		DocID:    t.DocID,
	}
}

type Tokenizer struct {
	s           *scanner.Scanner
	a           Analyzer
	termVector  int
	tokenVector int
	docID       int
	errors      []string
	phraseAware bool
}

type TokenOption func(tok *Tokenizer)

func NewTokenizer(options ...TokenOption) *Tokenizer {
	tok := &Tokenizer{}

	for _, option := range options {
		option(tok)
	}

	return tok
}

func SetPhraseAware() TokenOption {
	return func(tok *Tokenizer) {
		tok.phraseAware = true
	}
}

func (t *Tokenizer) ParseString(text string, docID int) *TokenStream {
	return t.Parse(strings.NewReader(text), docID)
}

func (t *Tokenizer) ParseBytes(b []byte, docID int) *TokenStream {
	return t.Parse(bytes.NewReader(b), docID)
}

func (t *Tokenizer) resetScanner() {
	var s scanner.Scanner
	t.s = &s
	t.termVector = 0
	t.tokenVector = 0

	if t.phraseAware {
		t.s.IsIdentRune = isPhraseIdentRune
		return
	}

	t.s.IsIdentRune = isIdentRune
}

func (t *Tokenizer) parseError(docID int) func(s *scanner.Scanner, msg string) {
	return func(s *scanner.Scanner, msg string) {
		if t.errors == nil {
			t.errors = []string{}
		}

		pos := s.Position

		t.errors = append(
			t.errors,
			fmt.Sprintf(
				"error: %s; %d:%d; docID: %d; tokenText: %s",
				msg,
				pos.Line,
				pos.Column,
				docID,
				s.TokenText(),
			),
		)
	}
}

// Parse creates a stream of tokens from an io.Reader.
// Each time Parse is called the document count is auto-incremented if a document
// identifier of 0 is given. Otherwise each call to Parse would effectively create
// the same vectors as the previous runs.
func (t *Tokenizer) Parse(r io.Reader, docID int) *TokenStream {
	t.resetScanner()
	t.s.Init(r)
	t.s.Error = t.parseError(docID)

	if docID == 0 {
		t.docID++
	} else {
		t.docID = docID
	}

	tokens := []Token{}
	for tok := t.s.Scan(); tok != scanner.EOF; tok = t.s.Scan() {
		tokens = append(tokens, t.runParser(tok))
	}

	if t.s.ErrorCount != 0 {
		for _, parseErr := range t.errors {
			log.Printf("parse error: %s", parseErr)
		}
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
	t.tokenVector++
	text := t.s.TokenText()

	var ignored bool

	pos := t.s.Position

	if text == "<" {
		ignored = true

		text = t.takeTag(text)
	}

	token := Token{
		RawText:     text,
		Vector:      t.tokenVector,
		OffsetStart: pos.Offset,
		OffsetEnd:   pos.Offset + (len(text) - 1),
		Ignored:     ignored,
		DocID:       t.docID,
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

	if token.isTermVector() {
		t.termVector++
		token.TermVector = t.termVector
	}

	return token
}

type TokenStream struct {
	tokens []Token
}

func (ts *TokenStream) next(idx int) (Token, bool) {
	if idx > len(ts.tokens) {
		return Token{}, false
	}

	return ts.tokens[idx], true
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

// TODO(kiivihal): refactor to reduce cyclo complexity
func (ts *TokenStream) Highlight(vectors *Vectors, tagLabel, emClass string) string {
	if vectors != nil && vectors.Size() == 0 {
		return ts.String()
	}

	startTag, startStyled, endTag := setDefaultTags(tagLabel, emClass)

	var str strings.Builder

	var inHighlight bool

	var insertSpace bool

	var splitHightlight bool

	for idx, token := range ts.tokens {
		next, hasNext := ts.next(idx)
		if hasNext && next.Ignored && !next.Punctuation && inHighlight {
			str.WriteString(endTag)

			inHighlight = false
			insertSpace = false
			splitHightlight = true
		}

		ok := vectors.HasVector(token.GetTermVector())
		if !ok && inHighlight && !token.Ignored {
			str.WriteString(endTag)
		}

		if insertSpace {
			insertSpace = false

			str.WriteString(" ")
		}

		if ok && !inHighlight {
			inHighlight = true

			if !splitHightlight {
				str.WriteString(startStyled)
			} else {
				str.WriteString(startTag)
			}
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

func setDefaultTags(tag, cssClass string) (startTag, startStyled, endTag string) {
	if tag == "" {
		tag = "em"
	}

	endTag = fmt.Sprintf("</%s>", tag)

	startTag = fmt.Sprintf("<%s>", tag)
	startStyled = startTag

	if cssClass != "" {
		startStyled = fmt.Sprintf("<%s class=\"%s\">", tag, cssClass)
	}

	return startTag, startStyled, endTag
}
