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

type termVector struct {
	term    string
	vectors []testVector
}

type testVector struct {
	inPhrase bool
	DocID    int
	Location int
}

func (tv testVector) searchVector() search.Vector {
	return search.Vector{
		DocID:    tv.DocID,
		Location: tv.Location,
	}
}

func createMatches(vectors []termVector) *search.Matches {
	matches := search.NewMatches()

	for _, v := range vectors {
		tv := search.NewVectors()

		for _, vector := range v.vectors {
			if vector.inPhrase {
				tv.AddPhraseVector(vector.searchVector())
				continue
			}

			tv.AddVector(vector.searchVector())
		}

		matches.AppendTerm(v.term, tv)
	}

	return matches
}

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
		1, // TODO should these be split
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
			// "ten": 0,
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

			if diff := cmp.Diff(tt.want, got.TermFrequency(), cmp.AllowUnexported(search.Matches{}, search.Vectors{})); diff != "" {
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
		hits *search.Matches
	}

	tests := []struct {
		name     string
		fields   fields
		args     args
		matchFn  func(qt *search.QueryTerm, hits *search.Matches) bool
		want     bool
		wantHits *search.Matches
	}{
		{
			"simple match",
			fields{"zeeheldenkwartier held kwartier"},
			args{
				&search.QueryTerm{Value: "held"},
				search.NewMatches(),
			},
			ti.matchTerm,
			true,
			createMatches(
				[]termVector{
					{"held", []testVector{{DocID: 1, Location: 2}}},
				},
			),
		},
		{
			"no match",
			fields{"zeeheldenkwartier held kwartier"},
			args{
				&search.QueryTerm{Value: "zeestraat"},
				search.NewMatches(),
			},
			ti.matchTerm,
			false,
			createMatches([]termVector{}),
		},
		{
			"prohibited query",
			fields{"zeeheldenkwartier held kwartier"},
			args{
				&search.QueryTerm{Value: "held", Prohibited: true},
				search.NewMatches(),
			},
			ti.matchTerm,
			false,
			search.NewMatches(),
		},
		{
			"prefix query",
			fields{"zeeheldenkwartier held kwartier"},
			args{
				&search.QueryTerm{Value: "zee", PrefixWildcard: true},
				search.NewMatches(),
			},
			ti.matchWildcard,
			true,
			createMatches([]termVector{
				{"zeeheldenkwartier", []testVector{{DocID: 1, Location: 1}}},
			}),
		},
		{
			"suffix query",
			fields{"zeeheldenkwartier held kwartier"},
			args{
				&search.QueryTerm{Value: "kwartier", SuffixWildcard: true},
				search.NewMatches(),
			},
			ti.matchWildcard,
			true,
			createMatches([]termVector{
				{"zeeheldenkwartier", []testVector{{DocID: 1, Location: 1}}},
				{"kwartier", []testVector{{DocID: 1, Location: 3}}},
			}),
		},
		{
			"fuzzy query",
			fields{"batauia"},
			args{
				&search.QueryTerm{Value: "batavia", Fuzzy: 1},
				search.NewMatches(),
			},
			ti.matchFuzzy,
			true,
			createMatches([]termVector{
				{"batauia", []testVector{{DocID: 1, Location: 1}}},
			}),
		},
		{
			"fuzzy query (no match)",
			fields{"Betuwe"},
			args{
				&search.QueryTerm{Value: "batavia", Fuzzy: 1},
				search.NewMatches(),
			},
			ti.matchFuzzy,
			false,
			search.NewMatches(),
		},
		{
			"phrase query (no match)",
			fields{"ware helden van de zee, enteren de VOC. Ware helden zijn geen piraten."},
			args{
				&search.QueryTerm{Value: "ware held", Phrase: true},
				search.NewMatches(),
			},
			ti.matchPhrase,
			false,
			search.NewMatches(),
		},
		{
			"single word phrase query (match)",
			fields{"ware helden van de zee, enteren de VOC. Ware helden zijn geen helden maar piraten."},
			args{
				&search.QueryTerm{Value: "de", Phrase: true},
				search.NewMatches(),
			},
			ti.matchPhrase,
			true,
			createMatches([]termVector{
				{"de", []testVector{
					{DocID: 1, Location: 4},
					{DocID: 1, Location: 7},
				}},
			}),
		},
		{
			"phrase query (match)",
			fields{"ware helden van de zee, enteren de VOC. Ware helden zijn geen helden maar ware piraten."},
			args{
				&search.QueryTerm{Value: "ware helden", Phrase: true, Slop: 0},
				search.NewMatches(),
			},
			ti.matchPhrase,
			true,
			createMatches([]termVector{
				{"ware helden", []testVector{
					{DocID: 1, Location: 1, inPhrase: true},
					{DocID: 1, Location: 2},
					{DocID: 1, Location: 9, inPhrase: true},
					{DocID: 1, Location: 10},
				}},
			}),
		},
		{
			"phrase query (no match) without slop",
			fields{"ware helden van de zee, enteren de VOC. Ware helden zijn geen helden, ware piraten."},
			args{
				&search.QueryTerm{Value: "helden van zee", Phrase: true, Slop: 0},
				search.NewMatches(),
			},
			ti.matchPhrase,
			false,
			search.NewMatches(),
		},
		{
			"phrase query (match) with slop",
			fields{"ware helden van de zee, enteren de VOC. Ware helden zijn geen hélden, ware piraten."},
			args{
				&search.QueryTerm{Value: "helden ware", Phrase: true, Slop: 1},
				search.NewMatches(),
			},
			ti.matchPhrase,
			true,
			createMatches([]termVector{
				{"ware helden", []testVector{
					{DocID: 1, Location: 1, inPhrase: true},
					{DocID: 1, Location: 2},
					{DocID: 1, Location: 9, inPhrase: true},
					{DocID: 1, Location: 10},
				}},
				{"helden ware", []testVector{
					{DocID: 1, Location: 13, inPhrase: true},
					{DocID: 1, Location: 14},
				}},
			}),
		},
		{
			"phrase query (match) with punctuation",
			fields{"zijn zoon, mr. Joan Blaeu, door"},
			args{
				&search.QueryTerm{Value: "mr joan blaeu", Phrase: true},
				search.NewMatches(),
			},
			ti.matchPhrase,
			true,
			createMatches([]termVector{
				{"mr joan blaeu", []testVector{
					{DocID: 1, Location: 3, inPhrase: true},
					{DocID: 1, Location: 4, inPhrase: true},
					{DocID: 1, Location: 5},
				}},
			}),
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

			if diff := cmp.Diff(tt.wantHits, tt.args.hits, cmp.AllowUnexported(search.Matches{})); diff != "" {
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
		hits  *search.Matches
	}

	tests := []struct {
		name           string
		args           args
		wantHits       *search.Matches
		wantErr        bool
		wantNoMatchErr bool
	}{
		{
			"term query (match)",
			args{
				"-vrede",
				"vredespaleis",
				search.NewMatches(),
			},
			// TODO(kiivihal): determine if not matches should come back as 0. probably not.
			createMatches([]termVector{}),
			false,
			false,
		},
		{
			"term query (no match)",
			args{
				"-vredespaleis",
				"vredespaleis",
				search.NewMatches(),
			},
			search.NewMatches(),
			true,
			true,
		},
		{
			"boolean term query (no match)",
			args{
				"-word AND (-something AND word OR word3)",
				"word something ガンダムバルバトス word3",
				search.NewMatches(),
			},
			search.NewMatches(),
			true,
			true,
		},
		{
			"boolean term query (no match)",
			args{
				"word1 OR (word2 OR (word4 OR word5))",
				"word something ガンダムバルバトス word3",
				search.NewMatches(),
			},
			search.NewMatches(),
			true,
			true,
		},
		{
			"inverted boolean term query (no match)",
			args{
				"word AND NOT (word2 AND word3))",
				"word1 something ガンダムバルバトス word3",
				search.NewMatches(),
			},
			search.NewMatches(),
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

			if diff := cmp.Diff(tt.wantHits, tt.args.hits, cmp.AllowUnexported(search.Matches{})); diff != "" {
				t.Errorf("TextIndex.search() %s = mismatch (-want +got):\n%s", tt.name, diff)
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

	newTi, err := readFrom(&buf)
	is.NoErr(err)

	if diff := cmp.Diff(ti, newTi, cmp.AllowUnexported(TextIndex{}, search.Vector{})); diff != "" {
		t.Errorf("TextIndex serialization = mismatch (-want +got):\n%s", diff)
	}
}

func TestTextIndex_setDocID(t *testing.T) {
	type args struct {
		docID []int
	}

	tests := []struct {
		name string
		args args
		want int
	}{
		{
			"no docIDs",
			args{},
			1,
		},
		{
			"single docIDs",
			args{[]int{10}},
			10,
		},
		{
			"multiple docIDs",
			args{[]int{10, 11}},
			10,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			ti := NewTextIndex()
			got := ti.setDocID(tt.args.docID...)

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("TextIndex.setDocID() %s = mismatch (-want +got):\n%s", tt.name, diff)
			}

			if diff := cmp.Diff(ti.DocCount, got); diff != "" {
				t.Errorf("TextIndex.DocCount %s = mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}
