package search

import (
	"github.com/sajari/fuzzy"
)

type SpellCheckOption func(*SpellChecker)

type SpellChecker struct {
	m         *fuzzy.Model
	depth     int
	threshold int
	a         *Analyzer
}

func NewSpellCheck(options ...SpellCheckOption) *SpellChecker {
	s := &SpellChecker{
		depth:     2,
		threshold: 5,
		a:         &Analyzer{},
	}

	for _, option := range options {
		option(s)
	}

	return s
}

func (s *SpellChecker) newModel() *fuzzy.Model {
	m := fuzzy.NewModel()
	m.SetThreshold(s.threshold)
	m.SetDepth(s.depth)

	return m
}

func (s *SpellChecker) Train(stream *TokenStream) {
	if s.m == nil {
		s.m = s.newModel()
	}

	for _, token := range stream.Tokens() {
		if !token.Ignored && token.Normal != "" {
			s.m.TrainWord(token.Normal)
		}
	}
}

func (s *SpellChecker) SetCount(term string, count int, suggest bool) {
	if s.m == nil {
		s.m = s.newModel()
	}

	s.m.SetCount(term, count, suggest)
}

// Return the most likely correction for the input termgg
func (s *SpellChecker) SpellCheck(input string) string {
	if s.m == nil {
		return ""
	}

	return s.m.SpellCheck(s.a.Transform(input))
}

// Return the most likely corrections in order from best to worst
func (s *SpellChecker) SpellCheckSuggestions(input string, n int) []string {
	if s.m == nil {
		return []string{}
	}

	return s.m.SpellCheckSuggestions(s.a.Transform(input), n)
}

func SetSuggestDepth(depth int) SpellCheckOption {
	return func(c *SpellChecker) {
		c.depth = depth
	}
}

func SetThreshold(threshold int) SpellCheckOption {
	return func(c *SpellChecker) {
		c.threshold = threshold
	}
}
