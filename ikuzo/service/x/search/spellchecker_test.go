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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/matryer/is"
)

func TestSpellChecker_SpellCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping spellcheck in short mode")
	}

	// nolint:misspell // old dutch
	text := `
	Teerste Deel Vande Spieghel der Zeevaerdt, vande navigatie der Westersche Zee,
	Innehoudende alle de Custen van Vranckrijck Spaignen ende t'principaelste deel
	van Engelandt, in diversche Zee Caerten begrepen, met den gebruijcke van dien,
	nu met grooter naersticheijt bij een vergadert ende ghepractizeert, Door Lucas
	Iansz Waghenaer Piloot ofte Stuijrman Residerende Inde vermaerde Zeestadt enchuijsen.
	Cum Privilegio ad decennium, Reg. Ma.is et Cancellarie Brabantie. 1583. Ghedruct tot
	Leyden by Christoffel Plantijn voor Lucas Janssz Waghenaer van Enckhuysen.
	Anno M.D.LXXXIIII.",
	`

	type pair struct {
		term  string
		count int
	}

	type fields struct {
		depth     int
		threshold int
		p         pair
	}

	type args struct {
		input string
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   []string
	}{
		{
			"simple correct",
			fields{
				depth:     5,
				threshold: 1,
				p: pair{
					term:  "plantijn",
					count: 10,
				},
			},
			args{
				input: "enkhuizen",
			},
			[]string{
				"enchuijsen",
				"enckhuysen",
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			s := NewSpellCheck(
				SetSuggestDepth(5),
				SetThreshold(1),
			)
			s.m = s.newModel()

			s.m.SetUseAutocomplete(false)

			tok := NewTokenizer()
			ts := tok.ParseString(text, 1)
			s.Train(ts)
			s.SetCount(tt.fields.p.term, tt.fields.p.count, true)
			s.SetCount("enchuijsen", 5, true)
			s.SetCount("enckhuysen", 3, true)

			got := s.SpellCheckSuggestions(tt.args.input, 5)

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("SpellCheck.SpellCheck() %s = mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}

// nolint:gocritic
func TestSpellChecker_NewSpellCheck(t *testing.T) {
	is := is.New(t)

	s := NewSpellCheck(
		SetSuggestDepth(5),
		SetThreshold(1),
	)

	is.Equal(s.m, nil)

	is.Equal(s.SpellCheckSuggestions("bom", 10), []string{})

	is.Equal(s.SpellCheck("bom"), "")

	s.SetCount("boom", 10, true)

	is.True(s.m != nil)

	is.Equal(s.SpellCheck("bom"), "boom")
}
