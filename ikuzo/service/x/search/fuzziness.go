package search

import (
	"fmt"

	"github.com/antzucaro/matchr"
)

type DistanceCalculator int

const (
	// Levenshtein computes the Levenshtein distance between two
	// strings. The returned value - distance - is the number of insertions,
	// deletions, and substitutions it takes to transform one
	// string (s1) into another (s2). Each step in the transformation "costs"
	// one distance point.
	Levenshtein DistanceCalculator = iota

	// DamerauLevenshtein computes the Damerau-Levenshtein distance between two
	// strings. The returned value - distance - is the number of insertions,
	// deletions, substitutions, and transpositions it takes to transform one
	// string (s1) into another (s2). Each step in the transformation "costs"
	// one distance point. It is similar to the Optimal String Alignment,
	// algorithm, but is more complex because it allows multiple edits on
	// substrings.
	DamerauLevenshtein

	// Hamming computes the Hamming distance between two equal-length strings.
	// This is the number of times the two strings differ between characters at
	// the same index. This implementation is based off of the algorithm
	// description found at http://en.wikipedia.org/wiki/Hamming_distance.
	Hamming

	// Jaro computes the Jaro edit distance between two strings. It represents
	// this with a float64 between 0 and 1 inclusive, with 0 indicating the two
	// strings are not at all similar and 1 indicating the two strings are exact
	// matches.
	//
	// See http://en.wikipedia.org/wiki/Jaro%E2%80%93Winkler_distance for a
	// full description.
	Jaro

	// JaroWinkler computes the Jaro-Winkler edit distance between two strings.
	// This is a modification of the Jaro algorithm that gives additional weight
	// to prefix matches.
	JaroWinkler

	// OSA computes the Optimal String Alignment distance between two
	// strings. The returned value - distance - is the number of insertions,
	// deletions, substitutions, and transpositions it takes to transform one
	// string (s1) into another (s2). Each step in the transformation "costs"
	// one distance point. It is similar to Damerau-Levenshtein, but is simpler
	// because it does not allow multiple edits on any substring.
	Osa

	// SmithWaterman computes the Smith-Waterman local sequence alignment for the
	// two input strings. This was originally designed to find similar regions in
	// strings representing DNA or protein sequences.
	SmithWaterman
)

func (dc DistanceCalculator) String() string {
	return [...]string{
		"Levenshtein",
		"DamerauLevenshtein",
		"Hamming",
		"Jaro",
		"JaroWinkler",
		"Osa",
		"SmithWaterman",
	}[dc]
}

// Distance calculates the similarity between two strings.
// The DistanceCalculator determines which algorithm is used.
//
// This function is a wrapper for the matchr library. The documentation of the
// DistanceCalculator constants and parts of the testdata are adopted from there.
// For more information about the implementation, see http://github.com/antzucaro/matchr.
//
// There are two groups of algorithms:
//
// Edit distance: Levenshstein, Damerau-Levenshtein, Hamming, Jaro-Winkler, SmithWaterman
//
// Sound similarity: Metaphone, Nysiis, Osa, Phonex, Soundex
func Distance(s1, s2 string, c DistanceCalculator) (float64, error) {
	var distance float64

	switch c {
	case Levenshtein:
		distance = float64(matchr.Levenshtein(s1, s2))
	case DamerauLevenshtein:
		distance = float64(matchr.DamerauLevenshtein(s1, s2))
	case Osa:
		distance = float64(matchr.OSA(s1, s2))
	case SmithWaterman:
		distance = matchr.SmithWaterman(s1, s2)
	default:
		return 0.0, fmt.Errorf("unknown DistanceCalculator: %s", c)
	}

	return distance, nil
}

type PhoneticPreprocessor int

const (
	// DoubleMetaphone computes the Double-Metaphone value of the input string.
	// This value is a phonetic representation of how the string sounds, with
	// affordances for many different language dialects. It was originally
	// developed by Lawrence Phillips in the 1990s.
	//
	// More information about this algorithm can be found on Wikipedia at
	// http://en.wikipedia.org/wiki/Metaphone.
	DoubleMetaphone PhoneticPreprocessor = iota

	// NYSIIS computes the NYSIIS phonetic encoding of the input string. It is a
	// modification of the traditional Soundex algorithm.
	Nysiis

	// Phonex computes the Phonex phonetic encoding of the input string. Phonex is
	// a modification of the venerable Soundex algorithm. It accounts for a few
	// more letter combinations to improve accuracy on some data sets.
	//
	// This implementation is based off of the original C implementation by the
	// creator - A. J. Lait - as found in his research paper entitled "An
	// Assessment of Name Matching Algorithms."
	Phonex

	// Soundex computes the Soundex phonetic representation of the input string. It
	// attempts to encode homophones with the same characters. More information can
	// be found at http://en.wikipedia.org/wiki/Soundex.
	Soundex
)

func (pp PhoneticPreprocessor) String() string {
	return [...]string{
		"DoubleMetaphone",
		"Nysiis",
		"Phonex",
		"Soundex",
	}[pp]
}

func Transform(s1 string, pp PhoneticPreprocessor) (string, error) {
	var output string

	switch pp {
	case DoubleMetaphone:
		output, _ = matchr.DoubleMetaphone(s1)
	case Nysiis:
		output = matchr.NYSIIS(s1)
	case Soundex:
		output = matchr.Soundex(s1)
	case Phonex:
		output = matchr.Phonex(s1)
	default:
		return "", fmt.Errorf("unknown phonetic preprocessor; %s", pp)
	}

	return output, nil
}

// IsFuzzyMatch determines if two strings are similar enough within the specified fuzziness.
func IsFuzzyMatch(s1, s2 string, fuzziness float64, c DistanceCalculator) (bool, error) {
	distance, err := Distance(s1, s2, c)
	if err != nil {
		return false, fmt.Errorf("unable to make fuzzy match; %w", err)
	}

	return fuzziness >= distance, nil
}
