package search

import (
	"fmt"
	"strconv"
	"strings"
	"text/scanner"
	"unicode"
)

type Operator string

const (
	AndOperator      Operator = "AND"
	BoostOperator    Operator = "^"
	FieldOperator    Operator = ":"
	FuzzyOperator    Operator = "~"
	NilOperator      Operator = ""
	NotOperator      Operator = "NOT"
	OrOperator       Operator = "OR"
	WildCardOperator Operator = "*"

	fuzzinesDefault = 2
)

type QueryType int

const (
	BoolQuery QueryType = iota
	FuzzyQuery
	PhraseQuery
	TermQuery
	WildCardQuery
)

func (qt QueryType) String() string {
	return [...]string{
		"BoolQuery",
		"FuzzyQuery",
		"PhraseQuery",
		"TermQuery",
		"WildCardQuery",
	}[qt]
}

type QueryTerm struct {
	Field          string
	Value          string
	Prohibited     bool
	Phrase         bool
	SuffixWildcard bool
	PrefixWildcard bool
	Boost          float64
	Fuzzy          int // fuzzy is for words
	Slop           int // slop is for phrases
	mustClauses    []*QueryTerm
	mustNotClauses []*QueryTerm
	shouldClauses  []*QueryTerm
	nested         *QueryTerm
}

// Type returns the type of the Query.
func (qt *QueryTerm) Type() QueryType {
	switch {
	case qt.IsBoolQuery():
		return BoolQuery
	case qt.Phrase:
		return PhraseQuery
	case qt.PrefixWildcard, qt.SuffixWildcard:
		return WildCardQuery
	case qt.Fuzzy != 0:
		return FuzzyQuery
	}

	return TermQuery
}

// isBoolQuery returns true if the QueryTerm has a nested QueryTerm in a Boolean
// clause.
func (qt *QueryTerm) IsBoolQuery() bool {
	if len(qt.shouldClauses) > 0 {
		return true
	}

	if len(qt.mustClauses) > 0 {
		return true
	}

	if len(qt.mustNotClauses) > 0 {
		return true
	}

	return false
}

// Must returns a list of Required QueryTerms.
func (qt *QueryTerm) Must() []*QueryTerm {
	return qt.mustClauses
}

// MustNot returns a list of Prohibited QueryTerms.
func (qt *QueryTerm) MustNot() []*QueryTerm {
	return qt.mustNotClauses
}

// Should returns a list of Optional QueryTerms.
// One or more must match to satistify the Query.
func (qt *QueryTerm) Should() []*QueryTerm {
	return qt.shouldClauses
}

// copy creates a copy of the QueryTerm to solve pointer arithmatic issues
// when appending to boolean clause slice.
func (qt *QueryTerm) copy() *QueryTerm {
	c := *qt
	return &c
}

func (qt *QueryTerm) setBoost(qp *QueryParser) error {
	if unicode.IsDigit(qp.s.Peek()) {
		qp.s.Scan()
		text := qp.tokenText()

		boost, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return fmt.Errorf("unable to parse boost float from %s; %w", text, err)
		}

		qt.Boost = boost
	}

	return nil
}

func (qt *QueryTerm) setFuzziness(qp *QueryParser) error {
	r := qp.s.Peek()

	switch {
	case unicode.IsDigit(r):
		qp.s.Scan()
		text := qp.tokenText()

		fuzzy, err := strconv.Atoi(text)
		if err != nil {
			return fmt.Errorf("unable to parse fuzzy int from %s; %w", text, err)
		}

		qt.Fuzzy = fuzzy
	case unicode.IsSpace(r), r == scanner.EOF:
		// if phrase query default slop is the number of words
		if qt.Phrase {
			qt.Slop = len(strings.Fields(qt.Value))
			return nil
		}

		// default fuzzy is length of the term
		qt.Fuzzy = fuzzinesDefault
	}

	return nil
}

func (qt *QueryTerm) setWildcard(qp *QueryParser) (*QueryTerm, bool) {
	if qt != nil && qt.Value != "" {
		qt.PrefixWildcard = true
		return qt, true
	} else if qt == nil && unicode.IsLetter(qp.s.Peek()) {
		return &QueryTerm{SuffixWildcard: true}, true
	}

	return nil, false
}

func (qt *QueryTerm) validate() {
	// clean up phrase query
	if strings.HasPrefix(qt.Value, "\"") && strings.HasSuffix(qt.Value, "\"") {
		qt.Phrase = true
		qt.Value = strings.TrimFunc(qt.Value, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsNumber(r)
		})
	}

	if qt.Fuzzy > 0 && qt.Phrase {
		qt.Slop = qt.Fuzzy
		qt.Fuzzy = 0
	}
}

type QueryOption func(*QueryParser) error

// QueryParser is used to parse human readable query syntax.
//
// The main idea behind this parser is that a person should be able to type whatever they want to represent a query,
// and this parser will do its best to interpret what to search for no matter how poorly composed the request may be.
// Tokens are considered to be any of a term, phrase, or subquery for the operations described below.
// Whitespace including ' ' '\n' '\r' and '\t' and certain operators may be used to delimit tokens ( ) + | " .

// Any errors in query syntax will be ignored and the parser will attempt to decipher what it can;
// however, this may mean odd or unexpected results.

// Query Operators

// '+' specifies AND operation: token1+token2
// '|' specifies OR operation: token1|token2
// '-' negates a single token: -token0
// '"' creates phrases of terms: "term1 term2 ..."
// '*' at the end of terms specifies prefix query: term*
// '~N' at the end of terms specifies fuzzy query: term~1
// '~N' at the end of phrases specifies near query: "term1 term2"~5
// '^N' at the end of terms specifies boost query: term~1.5
// '^N' at the end of phrases specifies a boost query: "term1 term2"~2.4
// '(' and ')' specifies precedence: token1 + (token2 | token3)
// ':' in the middle of terms specifies the end of a query field

//
// The default operator is OR if no other operator is specified. For example, the following will OR token1 and token2
// together: token1 token2
//
// Normal operator precedence will be simple order from right to left. For example, the following will evaluate token1
// OR token2 first, then AND with token3:
//
// token1 | token2 + token3
//
// Escaping
// An individual term may contain any possible character with certain characters requiring escaping using a '\'. The
// following characters will need to be escaped in terms and phrases: + | " ( ) ' \

// The '-' operator is a special case. On individual terms (not phrases) the first character of a term that is - must
// be escaped; however, any '-' characters beyond the first character do not need to be escaped. For example:

// -term1 -- Specifies NOT operation against term1
// \-term1 -- Searches for the term -term1.
// term-1 -- Searches for the term term-1.
// term\-1 -- Searches for the term term-1.
// The '*' operator is a special case. On individual terms (not phrases) the last character of a term that is '*' must
// be escaped; however, any '*' characters before the last character do not need to be escaped:

// term1* -- Searches for the prefix term1
// term1\* -- Searches for the term term1*
// term*1 -- Searches for the term term*1
// term\*1 -- Searches for the term term*1
// Note that above examples consider the terms before text processing.
//
// The specification and documentation is adopted from the Lucene documentation for the SimpleQueryParser:
// https://lucene.apache.org/core/6_6_1/queryparser/org/apache/lucene/queryparser/simple/SimpleQueryParser.html
type QueryParser struct {
	defaultAND bool
	s          *scanner.Scanner
	a          Analyzer
	fields     []string
}

// NewQueryParser returns a QueryParser that can be used to parse user queries.
func NewQueryParser(options ...QueryOption) (*QueryParser, error) {
	var s scanner.Scanner
	// set custom tokenization rune
	s.IsIdentRune = isIdentRune

	qp := &QueryParser{
		s: &s,
	}

	// apply options
	for _, option := range options {
		if err := option(qp); err != nil {
			return nil, err
		}
	}

	return qp, nil
}

// SetDefaultOperator sets the default boolean search operator for the query
func SetDefaultOperator(op Operator) QueryOption {
	return func(qp *QueryParser) error {
		switch op {
		case AndOperator:
			qp.defaultAND = true
		case OrOperator:
			qp.defaultAND = false
		case NotOperator, NilOperator:
			return fmt.Errorf("NOT and NIL operators are not valid values")
		}

		return nil
	}
}

// SetFields sets the default search fields for the query
func SetFields(field ...string) QueryOption {
	return func(qp *QueryParser) error {
		qp.fields = field
		return nil
	}
}

// Fields returns the default search fields for the query
func (qp *QueryParser) Fields() []string {
	return qp.fields
}

func (qp *QueryParser) appendQuery(parent *QueryTerm, op Operator, qt *QueryTerm) {
	if op == "" {
		if qp.defaultAND {
			op = AndOperator
		} else {
			op = OrOperator
		}
	}

	if qt.nested != nil {
		qp.appendQuery(parent, op, qt.nested.copy())
		qt.nested = nil
	}

	qt.Value = qp.a.TransformPhrase(qt.Value)

	switch op {
	case AndOperator:
		parent.mustClauses = append(parent.mustClauses, qt.copy())
	case OrOperator:
		parent.shouldClauses = append(parent.shouldClauses, qt.copy())
	case NotOperator:
		qt.Prohibited = true
		parent.mustNotClauses = append(parent.mustNotClauses, qt.copy())
	}
}

func (qp *QueryParser) tokenText() string {
	return qp.s.TokenText()
}

func (qp *QueryParser) Parse(query string) (*QueryTerm, error) {
	qp.s.Init(strings.NewReader(query))

	q := &QueryTerm{}

	err := qp.runParser(q, NilOperator, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to parse query input %s, due to; %w", query, err)
	}

	return q, nil
}

func (qp *QueryParser) processOperators(q *QueryTerm, op Operator, qt *QueryTerm, token string) (bool, error) {
	switch token {
	case string(OrOperator), "|":
		return true, qp.runParser(q, OrOperator, qt)
	case string(AndOperator), "+":
		return true, qp.runParser(q, AndOperator, qt)
	case string(NotOperator), "-":
		if qt != nil {
			qp.appendQuery(q, op, qt)
			qt = nil
		}

		return true, qp.runParser(q, NotOperator, qt)
	case string(BoostOperator):
		if err := qt.setBoost(qp); err != nil {
			return false, err
		}

		return true, qp.runParser(q, op, qt)
	case string(FuzzyOperator):
		if err := qt.setFuzziness(qp); err != nil {
			return false, err
		}

		qt.validate()

		return true, qp.runParser(q, op, qt)
	case string(WildCardOperator):
		if wildcard, ok := qt.setWildcard(qp); ok {
			return true, qp.runParser(q, op, wildcard)
		}
	}

	return false, nil
}

// recursive function to call parser.
// It is greedy towards the left operand. So it will look for the right operator to determine the boolean relation.
func (qp *QueryParser) runParser(q *QueryTerm, op Operator, qt *QueryTerm) error {
	tok := qp.s.Scan()

	text := qp.tokenText()

	ok, err := qp.processOperators(q, op, qt, text)
	if ok || err != nil {
		return err
	}

	// process delimiters
	switch text {
	case string(FieldOperator):
		qt.Field = qt.Value
		qt.Value = ""
		tok = qp.s.Scan()
		text = qp.tokenText()
	case "(":
		// start now bool
		nestedBoolQuery := &QueryTerm{}
		nestedQueryTerm := &QueryTerm{}

		err := qp.runParser(nestedBoolQuery, op, nestedQueryTerm)
		if err != nil {
			return err
		}

		if qt == nil {
			qt = &QueryTerm{}
		}

		qt.nested = nestedBoolQuery

		return qp.runParser(q, op, qt)
	}

	if qt != nil && qt.Value != "" {
		qp.appendQuery(q, op, qt)
	}

	// end of the group so return so the nested bool can be closed
	if text == ")" {
		return nil
	}

	// finish up when end of scanner is reached.
	if tok == scanner.EOF {
		return nil
	}

	if qt == nil {
		qt = &QueryTerm{}
	}

	qt.Value = text

	qt.validate()

	return qp.runParser(q, op, qt)
}

func isIdentRune(ch rune, i int) bool {
	return ch == '_' || unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '.' && i > 0 || ch == '-' && i > 0 || ch == '/'
}
