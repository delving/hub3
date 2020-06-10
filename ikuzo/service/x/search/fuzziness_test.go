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
	"fmt"
	"testing"
)

type distanceTest struct {
	name    string
	s1      string
	s2      string
	want    float64
	wantErr bool
}

var levenshteinTests = []distanceTest{
	{"Levenshtein", "boom", "boem", 1.0, false},
	{"insertion", "car", "cars", 1.0, false},
	{"substitution", "library", "librari", 1.0, false},
	{"deletion", "library", "librar", 1.0, false},
	{"one empty, left", "", "library", 7.0, false},
	{"one empty, right", "library", "", 7.0, false},
	{"transposition", "library", "librayr", 2, false},
	{"two empties", "", "", 0.0, false},
	{"unicode stuff!", "Schüßler", "Schübler", 1.0, false},
	{"unicode stuff!", "Schüßler", "Schußler", 1.0, false},
	{"unicode stuff!", "Schüßler", "Schüßler", 0.0, false},
	{"unicode stuff!", "Schßüler", "Schüßler", 2.0, false},
	{"unicode stuff!", "Schüßler", "Schüler", 1.0, false},
	{"unicode stuff!", "Schüßler", "Schüßlers", 1.0, false},
}

var damerauLevenshteinTests = []distanceTest{
	{"insertion", "car", "cars", 1.0, false},
	{"substitution", "library", "librari", 1.0, false},
	{"deletion", "library", "librar", 1.0, false},
	{"one empty, left", "", "library", 7.0, false},
	{"one empty, right", "library", "", 7.0, false},
	{"transposition", "library", "librayr", 1.0, false},
	{"two empties", "", "", 0.0, false},
	{"unicode stuff!", "Schüßler", "Schübler", 1.0, false},
	{"unicode stuff!", "Schüßler", "Schußler", 1.0, false},
	{"unicode stuff!", "Schüßler", "Schüßler", 0.0, false},
	{"unicode stuff!", "Schßüler", "Schüßler", 1.0, false},
	{"unicode stuff!", "Schüßler", "Schüler", 1.0, false},
	{"unicode stuff!", "Schüßler", "Schüßlers", 1.0, false},
	{"difference between DL and OSA. This is DL, so it should be 2.", "ca", "abc", 2.0, false},
}

var osaTests = []distanceTest{
	{"insertion", "car", "cars", 1.0, false},
	{"substitution", "library", "librari", 1.0, false},
	{"deletion", "library", "librar", 1.0, false},
	{"transposition", "library", "librayr", 1.0, false},
	{"one empty, left", "", "library", 7.0, false},
	{"one empty, right", "library", "", 7.0, false},
	{"two empties", "", "", 0.0, false},
	{"unicode stuff!", "Schüßler", "Schübler", 1.0, false},
	{"unicode stuff!", "Schüßler", "Schußler", 1.0, false},
	{"unicode stuff!", "Schüßler", "Schüßler", 0.0, false},
	{"unicode stuff!", "Schßüler", "Schüßler", 1.0, false},
	{"unicode stuff!", "Schüßler", "Schüler", 1.0, false},
	{"unicode stuff!", "Schüßler", "Schüßlers", 1.0, false},
	{"difference between DL and OSA. This is OSA, so it should be 3.", "ca", "abc", 3.0, false},
}

var smithWatermanTests = []distanceTest{
	{"insertion", "car", "cars", 3.0, false},
	{"substitution", "library", "librari", 6.0, false},
	{"deletion", "library", "librar", 6.0, false},
	{"one empty, left", "", "library", 7.0, false},
	{"one empty, right", "library", "", 7.0, false},
	{"transposition", "library", "librayr", 5.5, false},
	{"two empties", "", "", 0.0, false},
	{"unicode stuff!", "Schüßler", "Schübler", 6.0, false},
	{"unicode stuff!", "Schüßler", "Schußler", 6.0, false},
	{"unicode stuff!", "Schüßler", "Schüßler", 8.0, false},
	{"unicode stuff!", "Schßüler", "Schüßler", 6.0, false},
	{"unicode stuff!", "Schüßler", "Schüler", 6.5, false},
	{"unicode stuff!", "Schüßler", "Schüßlers", 8.0, false},
}

func TestDistance(t *testing.T) {
	tests := []struct {
		name       string
		calculator DistanceCalculator
		examples   []distanceTest
	}{
		{
			"Levenshtein",
			Levenshtein,
			levenshteinTests,
		},
		{
			"DamerauLevenshtein",
			DamerauLevenshtein,
			damerauLevenshteinTests,
		},
		{
			"OSA",
			Osa,
			osaTests,
		},
		{
			"SmithWaterman",
			SmithWaterman,
			smithWatermanTests,
		},
		{
			"unknown DistanceCalculator",
			100,
			[]distanceTest{
				{"error", "car", "cars", 0, true},
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		for _, nt := range tt.examples {
			nt := nt
			testName := fmt.Sprintf("%s: %s", tt.name, nt.name)

			t.Run(testName, func(t *testing.T) {
				got, err := Distance(nt.s1, nt.s2, tt.calculator)

				if (err != nil) != nt.wantErr {
					t.Errorf("Distance() error = %v, wantErr %v", err, nt.wantErr)
				}

				if got != nt.want {
					t.Errorf("Distance() %s = %v, want %v", testName, got, nt.want)
				}
			})
		}
	}
}

// test cases from http://rosettacode.org/wiki/Soundex#F.23
type soundTests struct {
	s1      string
	want    string
	wantErr bool
}

var soundexTests = []soundTests{
	{"Ashcraft", "A261", false},
	{"Ashhhcraft", "A261", false},
	{"Ashcroft", "A261", false},
	{"Burroughs", "B620", false},
	{"Burrows", "B620", false},
	{"Ekzampul", "E251", false},
	{"Example", "E251", false},
	{"Ellery", "E460", false},
	{"Euler", "E460", false},
	{"Ghosh", "G200", false},
	{"Gauss", "G200", false},
	{"Gutierrez", "G362", false},
	{"Heilbronn", "H416", false},
	{"Hilbert", "H416", false},
	{"Jackson", "J250", false},
	{"Kant", "K530", false},
	{"Knuth", "K530", false},
	{"Lee", "L000", false},
	{"Lukasiewicz", "L222", false},
	{"Lissajous", "L222", false},
	{"Ladd", "L300", false},
	{"Lloyd", "L300", false},
	{"Moses", "M220", false},
	{"O'Hara", "O600", false},
	{"Pfister", "P236", false},
	{"Rubin", "R150", false},
	{"Robert", "R163", false},
	{"Rupert", "R163", false},
	{"Soundex", "S532", false},
	{"Sownteks", "S532", false},
	{"Tymczak", "T522", false},
	{"VanDeusen", "V532", false},
	{"Washington", "W252", false},
	{"Wheaton", "W350", false},
	// transcription test
	{"batavia", "B310", false},
	{"datavia", "D310", false},
	{"patavia", "P310", false},
	{"bataria", "B360", false},
	{"kasteel", "K234", false},
	{"casteel", "C234", false},
	{"casteels", "C234", false},
}

var phonexTests = []soundTests{
	{"123 testsss", "T230", false},
	{"24/7 test", "T230", false},
	{"A", "A000", false},
	{"Lee", "L000", false},
	{"Kuhne", "C500", false},
	{"Meyer-Lansky", "M452", false},
	{"Oepping", "A150", false},
	{"Daley", "D400", false},
	{"Dalitz", "D432", false},
	{"Duhlitz", "D432", false},
	{"Dull", "D400", false},
	{"De Ledes", "D430", false},
	{"Sandemann", "S500", false},
	{"Schüßler", "S460", false},
	{"Schmidt", "S530", false},
	{"Sinatra", "S536", false},
	{"Heinrich", "A562", false},
	{"Hammerschlag", "A524", false},
	{"Williams", "W450", false},
	{"Wilms", "W500", false},
	{"Wilson", "W250", false},
	{"Worms", "W500", false},
	{"Zedlitz", "S343", false},
	{"Zotteldecke", "S320", false},
	{"ZYX test", "S232", false},
	{"Scherman", "S500", false},
	{"Schurman", "S500", false},
	{"Sherman", "S500", false},
	{"Shermansss", "S500", false},
	{"Shireman", "S650", false},
	{"Shurman", "S500", false},
	{"Euler", "A460", false},
	{"Ellery", "A460", false},
	{"Hilbert", "A130", false},
	{"Heilbronn", "A165", false},
	{"Gauss", "G000", false},
	{"Ghosh", "G200", false},
	{"Knuth", "N300", false},
	{"Kant", "C530", false},
	{"Lloyd", "L430", false},
	{"Ladd", "L300", false},
	{"Lukasiewicz", "L200", false},
	{"Lissajous", "L200", false},
	{"Ashcraft", "A261", false},
	{"Philip", "F410", false},
	{"Fripp", "F610", false},
	{"Czarkowska", "C200", false},
	{"Hornblower", "A514", false},
	{"Looser", "L260", false},
	{"Wright", "R230", false},
	{"Phonic", "F520", false},
	{"Quickening", "C250", false},
	{"Kuickening", "C250", false},
	{"Joben", "G150", false},
	{"Zelda", "S300", false},
	{"S", "0000", false},
	{"H", "0000", false},
	{"", "0000", false},
}

var nysiisTests = []soundTests{
	{"knight", "NAGT", false},
	{"mitchell", "MATCAL", false},
	{"o'daniel", "ODANAL", false},
	{"brown sr", "BRANSR", false},
	{"browne III", "BRAN", false},
	{"browne IV", "BRANAV", false},
	{"O'Banion", "OBANAN", false},
	{"Mclaughlin", "MCLAGL", false},
	{"McCormack", "MCARNA", false},
	{"Chapman", "CAPNAN", false},
	{"Silva", "SALV", false},
	{"McDonald", "MCDANA", false},
	{"Lawson", "LASAN", false},
	{"Jacobs", "JACAB", false},
	{"Greene", "GRAN", false},
	{"O'Brien", "OBRAN", false},
	{"Morrison", "MARASA", false},
	{"Larson", "LARSAN", false},
	{"Willis", "WAL", false},
	{"Mackenzie", "MCANSY", false},
	{"Carr", "CAR", false},
	{"Lawrence", "LARANC", false},
	{"Matthews", "MAT", false},
	{"Richards", "RACARD", false},
	{"Bishop", "BASAP", false},
	{"Franklin", "FRANCL", false},
	{"McDaniel", "MCDANA", false},
	{"Harper", "HARPAR", false},
	{"Lynch", "LYNC", false},
	{"Watkins", "WATCAN", false},
	{"Carlson", "CARLSA", false},
	{"Wheeler", "WALAR", false},
	{"Louis XVI", "LASXV", false},
	{"2002", "", false},
	{"1/2", "", false},
	{"", "", false},
}

var metaphoneTests = []soundTests{
	{"harper", "HRPR", false},
}

func TestTransform(t *testing.T) {
	tests := []struct {
		name         string
		preprocessor PhoneticPreprocessor
		examples     []soundTests
	}{
		{
			"double metaphone",
			DoubleMetaphone,
			metaphoneTests,
		},
		{
			"soundex",
			Soundex,
			soundexTests,
		},
		{
			"phonex",
			Phonex,
			phonexTests,
		},
		{
			"nysiss",
			Nysiis,
			nysiisTests,
		},
		{
			"unknown PhoneticPreprocessor",
			100,
			[]soundTests{
				{"car", "", true},
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		for _, nt := range tt.examples {
			nt := nt

			t.Run(tt.name, func(t *testing.T) {
				got, err := Transform(nt.s1, tt.preprocessor)

				if (err != nil) != nt.wantErr {
					t.Errorf("Transform() error = %v, wantErr %v", err, nt.wantErr)
				}

				if err == nil && got != nt.want {
					t.Errorf("Transform() = %v, want %v", got, nt.want)
				}
			})
		}
	}
}

func TestIsFuzzyMatch(t *testing.T) {
	type args struct {
		s1        string
		s2        string
		fuzziness float64
		c         DistanceCalculator
	}

	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			"same word no fuzziness",
			args{"word", "word", 0, Levenshtein},
			true,
			false,
		},
		{
			"same word with fuzziness",
			args{"word", "word", 1, Levenshtein},
			true,
			false,
		},
		{
			"uppercase no fuzziness",
			args{"Word", "word", 0, Levenshtein},
			false,
			false,
		},
		{
			"uppercase with fuzziness",
			args{"Word", "word", 1, Levenshtein},
			true,
			false,
		},
		{
			"different words too low fuzziness",
			args{"woorden", "woord", 1, Levenshtein},
			false,
			false,
		},
		{
			"different words high fuzziness",
			args{"woorden", "woord", 2, Levenshtein},
			true,
			false,
		},
		{
			"unknown calculator",
			args{"Word", "word", 0, 1000},
			false,
			true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			got, err := IsFuzzyMatch(tt.args.s1, tt.args.s2, tt.args.fuzziness, tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsFuzzyMatch() %s; error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("IsFuzzyMatch() %s = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
