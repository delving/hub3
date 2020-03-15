// nolint:gocritic
package memory

import (
	"bytes"
	"errors"
	"testing"

	"github.com/delving/hub3/ikuzo/service/x/search"
	"github.com/google/go-cmp/cmp"
	"github.com/matryer/is"
)

var appendTests = []struct {
	name    string
	text    []byte
	want    int
	wantErr bool
}{
	{
		"single word",
		[]byte("word"),
		1,
		false,
	},
	{
		"two words",
		[]byte("two words"),
		2,
		false,
	},
	{
		"one word with hyphen",
		[]byte("two-words"),
		3,
		false,
	},
	{
		"empty string after normalisation will not throw error",
		[]byte(".,;:"),
		0,
		false,
	},
}

func TestTextIndex_appendBytes(t *testing.T) {
	is := is.New(t)

	for _, tt := range appendTests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			ti := NewTextIndex()

			if err := ti.AppendBytes(tt.text); (err != nil) != tt.wantErr {
				t.Errorf("TextIndex.appendBytes() %s; error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}

			is.Equal(ti.size(), tt.want)
		})
	}
}

func TestTextIndex_appendString(t *testing.T) {
	is := is.New(t)

	for _, tt := range appendTests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			ti := NewTextIndex()

			if err := ti.AppendString(string(tt.text)); (err != nil) != tt.wantErr {
				t.Errorf("TextIndex.appendBytes() %s; error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}

			is.Equal(ti.size(), tt.want)
		})
	}
}

func TestTextIndex_reset(t *testing.T) {
	is := is.New(t)

	ti := NewTextIndex()

	is.Equal(ti.size(), 0)

	err := ti.AppendString("some words")
	is.NoErr(err)

	is.Equal(ti.size(), 2)

	ti.reset()

	is.Equal(ti.size(), 0)
}

type searchArgs struct {
	query string
	text  string
}

var searchTests = []struct {
	name           string
	args           searchArgs
	want           map[string]int
	wantErr        bool
	wantNoMatchErr bool
}{
	{
		"query not found",
		searchArgs{
			query: "ten",
			text:  "zero to nine",
		},
		map[string]int{},
		true,
		true,
	},
	{
		"must query not found",
		searchArgs{
			query: "+ten",
			text:  "zero to nine",
		},
		map[string]int{},
		true,
		true,
	},
	{
		"query excluded with match",
		searchArgs{
			query: "-zero",
			text:  "zero to nine",
		},
		map[string]int{},
		true,
		true,
	},
	{
		"query excluded with no match",
		searchArgs{
			query: "-ten",
			text:  "zero to nine",
		},
		map[string]int{
			"ten": 0,
		},
		false,
		false,
	},
	{
		"single must match",
		searchArgs{
			query: "+zero",
			text:  "zero to nine",
		},
		map[string]int{
			"zero": 1,
		},
		false,
		false,
	},
	{
		"single should match",
		searchArgs{
			query: "zero",
			text:  "zero to nine",
		},
		map[string]int{
			"zero": 1,
		},
		false,
		false,
	},
	{
		"multi-word should match",
		searchArgs{
			query: "zero to none",
			text:  "zero to nine to zeros",
		},
		map[string]int{
			"zero": 1,
			"to":   2,
		},
		false,
		false,
	},
	{
		"multi-word should match",
		searchArgs{
			query: "zero to none",
			text:  "zero to nine to zeros",
		},
		map[string]int{
			"zero": 1,
			"to":   2,
		},
		false,
		false,
	},
	{
		"prefix query",
		searchArgs{
			query: "zer* to none",
			text:  "zero to nine to zeros",
		},
		map[string]int{
			"zero":  1,
			"zeros": 1,
			"to":    2,
		},
		false,
		false,
	},
	{
		"suffix query",
		searchArgs{
			query: "*o none",
			text:  "zero to nine to zeros",
		},
		map[string]int{
			"zero": 1,
			"to":   2,
		},
		false,
		false,
	},
	{
		"mixed query",
		searchArgs{
			query: "zero AND nine",
			text:  "zero to nine to zeros",
		},
		map[string]int{
			"zero": 1,
			"nine": 1,
		},
		false,
		false,
	},
	{
		"mixed nested query",
		searchArgs{
			query: "(something OR zero) AND to",
			text:  "zero to nine to zeros",
		},
		map[string]int{
			"to":   2,
			"zero": 1,
		},
		false,
		false,
	},
	{
		"mixed NOT query",
		searchArgs{
			query: "(something OR zero) AND -to",
			text:  "zero to nine to zeros",
		},
		map[string]int{},
		true,
		true,
	},
	{
		"mixed NOT query",
		searchArgs{
			query: "NOT (something OR zero) AND to",
			text:  "zero to nine to zeros",
		},
		map[string]int{
			"to":   2,
			"zero": 1,
		},
		false,
		false,
	},
	{
		"fuzzy search",
		searchArgs{
			query: "zer~",
			text:  "zero to nine to zeros",
		},
		map[string]int{
			"zero":  1,
			"zeros": 1,
		},
		false,
		false,
	},
	{
		"phrase search",
		searchArgs{
			query: "\"to nine\"",
			text:  "zero to nine to zeros",
		},
		map[string]int{
			"to nine": 1,
		},
		false,
		false,
	},
}

func TestTextIndex_search(t *testing.T) {
	is := is.New(t)

	// test prohibited
	t.Run("mustNot should always have search.QueryTerm.Prohibited == true", func(t *testing.T) {
		ti := NewTextIndex()
		err := ti.AppendString("something")
		is.NoErr(err)

		queryParser, err := search.NewQueryParser()
		is.NoErr(err)

		q, err := queryParser.Parse("-this")
		is.NoErr(err)

		for _, q := range q.MustNot() {
			q.Prohibited = false
		}

		_, err = ti.Search(q)
		if err == nil {
			t.Errorf("TextIndex.Search() QueryTerm in mustNotClauses without Prohibited == true is not allowed")
		}
	})

	for _, tt := range searchTests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			ti := NewTextIndex()

			queryParser, err := search.NewQueryParser()
			is.NoErr(err)

			err = ti.AppendString(tt.args.text)
			is.NoErr(err)

			query, err := queryParser.Parse(tt.args.query)
			is.NoErr(err)

			got, err := ti.Search(query)
			if (err != nil) != tt.wantErr {
				t.Errorf("TextIndex.search() %s error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}

			if errors.Is(err, ErrSearchNoMatch) != tt.wantNoMatchErr {
				t.Errorf("TextIndex.search() %s error = %v, wantNoMatchErr %v", tt.name, err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got.hits); diff != "" {
				t.Errorf("TextIndex.search() %s = mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}

// nolint:funlen
func TestTextIndex_matchCustom(t *testing.T) {
	is := is.New(t)
	ti := NewTextIndex()

	type fields struct {
		text string
	}

	type args struct {
		qt   *search.QueryTerm
		hits *SearchHits
	}

	tests := []struct {
		name     string
		fields   fields
		args     args
		matchFn  func(qt *search.QueryTerm, hits *SearchHits) bool
		want     bool
		wantHits *SearchHits
	}{
		{
			"simple match",
			fields{"zeeheldenkwartier held kwartier"},
			args{
				&search.QueryTerm{Value: "held"},
				newSearchHits(),
			},
			ti.matchTerm,
			true,
			&SearchHits{
				hits:           map[string]int{"held": 1},
				matchPositions: map[int]bool{1: true},
			},
		},
		{
			"no match",
			fields{"zeeheldenkwartier held kwartier"},
			args{
				&search.QueryTerm{Value: "zeestraat"},
				newSearchHits(),
			},
			ti.matchTerm,
			false,
			&SearchHits{
				hits:           map[string]int{},
				matchPositions: map[int]bool{},
			},
		},
		{
			"prohibited query",
			fields{"zeeheldenkwartier held kwartier"},
			args{
				&search.QueryTerm{Value: "held", Prohibited: true},
				newSearchHits(),
			},
			ti.matchTerm,
			false,
			&SearchHits{
				hits:           map[string]int{},
				matchPositions: map[int]bool{},
			},
		},
		{
			"prefix query",
			fields{"zeeheldenkwartier held kwartier"},
			args{
				&search.QueryTerm{Value: "zee", PrefixWildcard: true},
				newSearchHits(),
			},
			ti.matchWildcard,
			true,
			&SearchHits{
				hits:           map[string]int{"zeeheldenkwartier": 1},
				matchPositions: map[int]bool{0: true},
			},
		},
		{
			"suffix query",
			fields{"zeeheldenkwartier held kwartier"},
			args{
				&search.QueryTerm{Value: "kwartier", SuffixWildcard: true},
				newSearchHits(),
			},
			ti.matchWildcard,
			true,
			&SearchHits{
				hits:           map[string]int{"zeeheldenkwartier": 1, "kwartier": 1},
				matchPositions: map[int]bool{0: true, 2: true},
			},
		},
		{
			"fuzzy query",
			fields{"batauia"},
			args{
				&search.QueryTerm{Value: "batavia", Fuzzy: 1},
				newSearchHits(),
			},
			ti.matchFuzzy,
			true,
			&SearchHits{
				hits:           map[string]int{"batauia": 1},
				matchPositions: map[int]bool{0: true},
			},
		},
		{
			"fuzzy query (no match)",
			fields{"Betuwe"},
			args{
				&search.QueryTerm{Value: "batavia", Fuzzy: 1},
				newSearchHits(),
			},
			ti.matchFuzzy,
			false,
			&SearchHits{
				hits:           map[string]int{},
				matchPositions: map[int]bool{},
			},
		},
		{
			"phrase query (no match)",
			fields{"ware helden van de zee, enteren de VOC. Ware helden zijn geen piraten."},
			args{
				&search.QueryTerm{Value: "ware held", Phrase: true},
				newSearchHits(),
			},
			ti.matchPhrase,
			false,
			&SearchHits{
				hits:           map[string]int{},
				matchPositions: map[int]bool{},
			},
		},
		{
			"single word phrase query (match)",
			fields{"ware helden van de zee, enteren de VOC. Ware helden zijn geen helden maar piraten."},
			args{
				&search.QueryTerm{Value: "de", Phrase: true},
				newSearchHits(),
			},
			ti.matchPhrase,
			true,
			&SearchHits{
				hits:           map[string]int{"de": 2},
				matchPositions: map[int]bool{3: true, 6: true},
			},
		},
		{
			"phrase query (match)",
			fields{"ware helden van de zee, enteren de VOC. Ware helden zijn geen helden maar ware piraten."},
			args{
				&search.QueryTerm{Value: "ware helden", Phrase: true, Slop: 0},
				newSearchHits(),
			},
			ti.matchPhrase,
			true,
			&SearchHits{
				hits:           map[string]int{"ware helden": 2},
				matchPositions: map[int]bool{0: true, 1: true, 8: true, 9: true},
			},
		},
		{
			"phrase query (no match) without slop",
			fields{"ware helden van de zee, enteren de VOC. Ware helden zijn geen helden, ware piraten."},
			args{
				&search.QueryTerm{Value: "helden van zee", Phrase: true, Slop: 0},
				newSearchHits(),
			},
			ti.matchPhrase,
			false,
			&SearchHits{
				hits:           map[string]int{},
				matchPositions: map[int]bool{},
			},
		},
		{
			"phrase query (match) with slop",
			fields{"ware helden van de zee, enteren de VOC. Ware helden zijn geen hélden, ware piraten."},
			args{
				&search.QueryTerm{Value: "helden ware", Phrase: true, Slop: 1},
				newSearchHits(),
			},
			ti.matchPhrase,
			true,
			&SearchHits{
				hits:           map[string]int{"ware helden": 2, "helden ware": 1},
				matchPositions: map[int]bool{0: true, 1: true, 8: true, 9: true, 12: true, 13: true},
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			ti.reset()

			err := ti.AppendString(tt.fields.text)
			is.NoErr(err)

			if got := tt.matchFn(tt.args.qt, tt.args.hits); got != tt.want {
				t.Errorf("TextIndex.match custom %s = %v, want %v", tt.name, got, tt.want)
			}

			if diff := cmp.Diff(tt.wantHits, tt.args.hits, cmp.AllowUnexported(SearchHits{})); diff != "" {
				t.Errorf("TextIndex.match custom %s = mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}

func TestTextIndex_searchMustNot(t *testing.T) {
	is := is.New(t)

	type args struct {
		query string
		text  string
		hits  *SearchHits
	}

	tests := []struct {
		name           string
		args           args
		wantHits       *SearchHits
		wantErr        bool
		wantNoMatchErr bool
	}{
		{
			"term query (match)",
			args{
				"-vrede",
				"vredespaleis",
				newSearchHits(),
			},
			&SearchHits{
				hits:           map[string]int{"vrede": 0},
				matchPositions: map[int]bool{},
			},
			false,
			false,
		},
		{
			"term query (no match)",
			args{
				"-vredespaleis",
				"vredespaleis",
				newSearchHits(),
			},
			&SearchHits{
				hits:           map[string]int{},
				matchPositions: map[int]bool{},
			},
			true,
			true,
		},
		{
			"boolean term query (no match)",
			args{
				"-word AND (-something AND word OR word3)",
				"word something ガンダムバルバトス word3",
				newSearchHits(),
			},
			&SearchHits{
				hits:           map[string]int{},
				matchPositions: map[int]bool{},
			},
			true,
			true,
		},
		{
			"boolean term query (no match)",
			args{
				"word1 OR (word2 OR (word4 OR word5))",
				"word something ガンダムバルバトス word3",
				newSearchHits(),
			},
			&SearchHits{
				hits:           map[string]int{},
				matchPositions: map[int]bool{},
			},
			true,
			true,
		},
		{
			"inverted boolean term query (no match)",
			args{
				"word AND NOT (word2 AND word3))",
				"word1 something ガンダムバルバトス word3",
				newSearchHits(),
			},
			&SearchHits{
				hits:           map[string]int{},
				matchPositions: map[int]bool{},
			},
			true,
			true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			ti := NewTextIndex()

			queryParser, err := search.NewQueryParser()
			is.NoErr(err)

			err = ti.AppendString(tt.args.text)
			is.NoErr(err)

			query, err := queryParser.Parse(tt.args.query)
			is.NoErr(err)

			if err := ti.search(query, tt.args.hits); (err != nil) != tt.wantErr {
				t.Errorf("TextIndex.search() %s; error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.wantHits, tt.args.hits, cmp.AllowUnexported(SearchHits{})); diff != "" {
				t.Errorf("TextIndex.search() %s = mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}

func TestSearchHits_Total(t *testing.T) {
	type fields struct {
		hits map[string]int
	}

	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			"no results",
			fields{hits: map[string]int{}},
			0,
		},
		{
			"no results",
			fields{hits: map[string]int{
				"one":       1,
				"two times": 2,
				"many":      10,
			}},
			13,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			sh := &SearchHits{
				hits: tt.fields.hits,
			}
			if got := sh.Total(); got != tt.want {
				t.Errorf("SearchHits.Total() = %v, want %v", got, tt.want)
			}

			if diff := cmp.Diff(sh.hits, sh.Hits(), cmp.AllowUnexported(SearchHits{})); diff != "" {
				t.Errorf("SearchHits.Total() %s = mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}

func TestSearchHits_Merge(t *testing.T) {
	type fields struct {
		hits map[string]int
	}

	type args struct {
		hits map[string]int
	}

	tests := []struct {
		name     string
		fields   fields
		args     args
		wantHits map[string]int
	}{
		{
			"source only merge",
			fields{hits: map[string]int{"word": 1}},
			args{hits: map[string]int{}},
			map[string]int{"word": 1},
		},
		{
			"partial merge",
			fields{hits: map[string]int{"word": 1}},
			args{hits: map[string]int{"word": 1, "words": 2}},
			map[string]int{"word": 2, "words": 2},
		},
		{
			"target only merge",
			fields{hits: map[string]int{}},
			args{hits: map[string]int{"word": 1, "words": 2}},
			map[string]int{"word": 1, "words": 2},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			src := &SearchHits{
				hits: tt.fields.hits,
			}

			target := newSearchHits()
			target.hits = tt.args.hits

			src.Merge(target)

			if diff := cmp.Diff(tt.wantHits, src.Hits(), cmp.AllowUnexported(SearchHits{})); diff != "" {
				t.Errorf("SearchHits.Merge() %s = mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}

func TestTextIndexSerialization(t *testing.T) {
	is := is.New(t)

	ti := NewTextIndex()
	err := ti.AppendString("One two three. One two. One two three")
	is.NoErr(err)

	var buf bytes.Buffer

	err = ti.writeTo(&buf)
	is.NoErr(err)

	newTi := NewTextIndex()
	err = newTi.readFrom(&buf)
	is.NoErr(err)

	if diff := cmp.Diff(ti, newTi, cmp.AllowUnexported(TextIndex{}, TermVector{})); diff != "" {
		t.Errorf("TextIndex serialization = mismatch (-want +got):\n%s", diff)
	}
}
