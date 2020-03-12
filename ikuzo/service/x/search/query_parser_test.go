// nolint:gocritic
package search

import (
	"reflect"
	"strings"
	"testing"
	"text/scanner"

	"github.com/google/go-cmp/cmp"
	"github.com/matryer/is"
)

// nolint:funlen
func TestQueryParser_runParser(t *testing.T) {
	is := is.New(t)

	type args struct {
		input      string
		q          QueryTerm
		defaultAND bool
	}

	tests := []struct {
		name    string
		args    args
		want    QueryTerm
		wantErr bool
	}{
		{
			"single term",
			args{"word", QueryTerm{}, false},
			QueryTerm{
				shouldClauses: []*QueryTerm{
					{
						Value: "word",
					},
				},
			},
			false,
		},
		{
			"mixed identifier must be phrase",
			args{"\"1.04.02\"", QueryTerm{}, false},
			QueryTerm{
				shouldClauses: []*QueryTerm{
					{
						Value: "1.04.02", Phrase: true,
					},
				},
			},
			false,
		},
		{
			"phrase query",
			args{"\"multiple words\"", QueryTerm{}, false},
			QueryTerm{
				shouldClauses: []*QueryTerm{
					{Value: "multiple words", Phrase: true},
				},
			},
			false,
		},
		{
			"implicit OR multiword",
			args{"one two three", QueryTerm{}, false},
			QueryTerm{
				shouldClauses: []*QueryTerm{
					{Value: "one"},
					{Value: "two"},
					{Value: "three"},
				},
			},
			false,
		},
		{
			"implicit AND multiword",
			args{"one two three", QueryTerm{}, true},
			QueryTerm{
				mustClauses: []*QueryTerm{
					{Value: "one"},
					{Value: "two"},
					{Value: "three"},
				},
			},
			false,
		},
		{
			"explicit OR multiword",
			args{"one OR two OR three", QueryTerm{}, false},
			QueryTerm{
				shouldClauses: []*QueryTerm{
					{Value: "one"},
					{Value: "two"},
					{Value: "three"},
				},
			},
			false,
		},
		{
			"explicit AND multiword",
			args{"one AND two AND \"three words\"", QueryTerm{}, false},

			QueryTerm{
				mustClauses: []*QueryTerm{
					{Value: "one"},
					{Value: "two"},
					{Value: "three words", Phrase: true},
				},
			},
			false,
		},
		{
			"AND/OR group",
			args{"one AND two OR \"three words\"", QueryTerm{}, false},
			QueryTerm{
				mustClauses: []*QueryTerm{
					{Value: "one"},
				},
				shouldClauses: []*QueryTerm{
					{Value: "two"},
					{Value: "three words", Phrase: true},
				},
			},
			false,
		},
		{
			// greedy to left operator
			"OR/AND group (nested groups)",
			args{"one AND two AND \"three words\" OR \"no words\"", QueryTerm{}, false},
			QueryTerm{
				mustClauses: []*QueryTerm{
					{Value: "one"},
					{Value: "two"},
				},
				shouldClauses: []*QueryTerm{
					{Value: "three words", Phrase: true},
					{Value: "no words", Phrase: true},
				},
			},
			false,
		},
		{
			"explicit OR/AND group (nested groups)",
			args{"one AND (two OR three)", QueryTerm{}, false},
			QueryTerm{
				mustClauses: []*QueryTerm{
					{
						shouldClauses: []*QueryTerm{
							{Value: "two"},
							{Value: "three"},
						},
					},
					{Value: "one"},
				},
			},
			false,
		},
		{
			"nested explicit group",
			args{"(two OR three) AND one", QueryTerm{}, false},
			QueryTerm{
				mustClauses: []*QueryTerm{
					{
						shouldClauses: []*QueryTerm{
							{Value: "two"},
							{Value: "three"},
						},
					},
					{
						Value: "one",
					},
				},
			},
			false,
		},
		{
			"nested explicit group with error",
			args{"(two OR three~1a) AND one", QueryTerm{}, false},
			QueryTerm{},
			true,
		},
		{
			"leading implicite OR with nested explicit group",
			args{"three (two + \"four\") | one", QueryTerm{}, false},
			QueryTerm{
				shouldClauses: []*QueryTerm{
					{
						mustClauses: []*QueryTerm{
							{Value: "two"},
							{Value: "four", Phrase: true},
						},
					},
					{Value: "three"},
					{Value: "one"},
				},
			},
			false,
		},
		{
			"multi nested explicit group",
			args{"three | (two + (\"four\" | five)) | one", QueryTerm{}, false},
			QueryTerm{
				shouldClauses: []*QueryTerm{
					{
						mustClauses: []*QueryTerm{
							{
								shouldClauses: []*QueryTerm{{Value: "four", Phrase: true}, {Value: "five", Phrase: true}},
							},
							{Value: "two"},
						},
					},
					{Value: "three"},
					{Value: "one"},
				},
			},
			false,
		},
		{
			"NOT Query",
			args{"NOT one", QueryTerm{}, false},
			QueryTerm{
				mustNotClauses: []*QueryTerm{
					{Value: "one", Prohibited: true},
				},
			},
			false,
		},
		{
			"NOT shortcut with minus",
			args{"-one", QueryTerm{}, false},
			QueryTerm{
				mustNotClauses: []*QueryTerm{
					{Value: "one", Prohibited: true},
				},
			},
			false,
		},
		{
			"NOT shortcut with minus with should",
			args{"should -one", QueryTerm{}, false},
			QueryTerm{
				shouldClauses: []*QueryTerm{
					{Value: "should"},
				},
				mustNotClauses: []*QueryTerm{
					{Value: "one", Prohibited: true},
				},
			},
			false,
		},
		{
			"AND shortcut with plus",
			args{"+one", QueryTerm{}, false},
			QueryTerm{
				mustClauses: []*QueryTerm{
					{Value: "one"},
				},
			},
			false,
		},
		{
			"OR shortcut with pipe",
			args{"this | that", QueryTerm{}, true},
			QueryTerm{
				shouldClauses: []*QueryTerm{
					{Value: "this"},
					{Value: "that"},
				},
			},
			false,
		},
		{
			"AND shortcut with plus",
			args{"this + one", QueryTerm{}, false},
			QueryTerm{
				mustClauses: []*QueryTerm{
					{Value: "this"},
					{Value: "one"},
				},
			},
			false,
		},
		{
			"multiword NOT Query",
			args{"one NOT two", QueryTerm{}, false},
			QueryTerm{
				shouldClauses: []*QueryTerm{
					{Value: "one"},
				},
				mustNotClauses: []*QueryTerm{
					{Value: "two", Prohibited: true},
				},
			},
			false,
		},
		{
			"fielded query",
			args{"field:one", QueryTerm{}, false},
			QueryTerm{
				shouldClauses: []*QueryTerm{
					{Field: "field", Value: "one"},
				},
			},
			false,
		},
		{
			"fielded phrase query",
			args{"-field:\"one word\"", QueryTerm{}, false},
			QueryTerm{
				mustNotClauses: []*QueryTerm{
					{Field: "field", Value: "one word", Phrase: true, Prohibited: true},
				},
			},
			false,
		},
		{
			"query with boost",
			args{"word^2.5", QueryTerm{}, false},
			QueryTerm{
				shouldClauses: []*QueryTerm{
					{Value: "word", Boost: 2.5},
				},
			},
			false,
		},
		{
			"empty boost",
			args{"word^", QueryTerm{}, false},
			QueryTerm{
				shouldClauses: []*QueryTerm{
					{Value: "word", Boost: 0},
				},
			},
			false,
		},
		{
			"empty fuzzy operator",
			args{"word~", QueryTerm{}, false},
			QueryTerm{
				shouldClauses: []*QueryTerm{
					{Value: "word", Fuzzy: 2},
				},
			},
			false,
		},
		{
			"explicit fuzzy operator",
			args{"word~3", QueryTerm{}, false},
			QueryTerm{
				shouldClauses: []*QueryTerm{
					{Value: "word", Fuzzy: 3},
				},
			},
			false,
		},
		{
			"bad fuzzy operator",
			args{"word~1a", QueryTerm{}, false},
			QueryTerm{},
			true,
		},
		{
			"fuzzy operator for phrase is slop",
			args{"\"two words\"~3", QueryTerm{}, false},
			QueryTerm{
				shouldClauses: []*QueryTerm{
					{Value: "two words", Slop: 3, Phrase: true},
				},
			},
			false,
		},
		{
			"default phrase slop",
			args{"\"almost two words\"~", QueryTerm{}, false},
			QueryTerm{
				shouldClauses: []*QueryTerm{
					{Value: "almost two words", Slop: 3, Phrase: true},
				},
			},
			false,
		},
		{
			"prefix wildcard query",
			args{"prefix*", QueryTerm{}, false},
			QueryTerm{
				shouldClauses: []*QueryTerm{
					{Value: "prefix", PrefixWildcard: true},
				},
			},
			false,
		},
		{
			"suffix wildcard query",
			args{"*suffix", QueryTerm{}, false},
			QueryTerm{
				shouldClauses: []*QueryTerm{
					{Value: "suffix", SuffixWildcard: true},
				},
			},
			false,
		},
		{
			"analyzer for value",
			args{"övergångsställE", QueryTerm{}, false},
			QueryTerm{
				shouldClauses: []*QueryTerm{
					{Value: "overgangsstalle"},
				},
			},
			false,
		},
		{
			"analyzer not for field",
			args{"Title:övergångsställE", QueryTerm{}, false},
			QueryTerm{
				shouldClauses: []*QueryTerm{
					{Field: "Title", Value: "overgangsstalle"},
				},
			},
			false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			var err error

			qp, err := NewQueryParser()
			is.NoErr(err)
			qp.defaultAND = tt.args.defaultAND
			qp.s.Init(strings.NewReader(tt.args.input))

			if err = qp.runParser(&tt.args.q, "", nil); (err != nil) != tt.wantErr {
				t.Errorf("runParser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(tt.want, tt.args.q, cmp.AllowUnexported(QueryTerm{})); diff != "" {
				t.Errorf("runParser(); %s = mismatch (-want +got):\n%s", tt.name, diff)
			}

			if err == nil && qp.s.Scan() != scanner.EOF {
				t.Errorf("runParser(); %s = scanner should be empty", tt.name)
			}
		})
	}
}

func TestSetDefaultOperator(t *testing.T) {
	is := is.New(t)

	qt, err := NewQueryParser()
	is.NoErr(err)
	is.Equal(qt.defaultAND, false)

	tests := []struct {
		name     string
		operator Operator
		want     bool
		wantErr  bool
	}{
		{
			"or should be the default",
			OrOperator,
			false,
			false,
		},
		{
			"NilOperator is not allowed as default",
			NilOperator,
			false,
			true,
		},
		{
			"NotOperator is not allowed as default",
			NotOperator,
			false,
			true,
		},
		{
			"AndOperator should be set",
			AndOperator,
			true,
			false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			qt, err := NewQueryParser(SetDefaultOperator(tt.operator))

			if (err != nil) != tt.wantErr {
				t.Errorf("QueryParser.SetDefaultOperator() %s error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}

			if qt != nil && qt.defaultAND != tt.want {
				t.Errorf("QueryParser.SetDefaultOperator() %s = %v, want %v", tt.name, qt.defaultAND, tt.want)
			}
		})
	}
}

func TestQueryParser_appendQuery(t *testing.T) {
	type fields struct {
		defaultAND bool
	}

	type args struct {
		op Operator
		qt *QueryTerm
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   *QueryTerm
	}{
		{
			"default case",
			fields{defaultAND: false},
			args{
				NilOperator,
				&QueryTerm{Value: "word"},
			},
			&QueryTerm{
				shouldClauses: []*QueryTerm{{Value: "word"}},
			},
		},
		{
			"or operator case",
			fields{defaultAND: false},
			args{
				OrOperator,
				&QueryTerm{Value: "word"},
			},
			&QueryTerm{
				shouldClauses: []*QueryTerm{{Value: "word"}},
			},
		},
		{
			"and operator case",
			fields{defaultAND: false},
			args{
				AndOperator,
				&QueryTerm{Value: "word"},
			},
			&QueryTerm{
				mustClauses: []*QueryTerm{{Value: "word"}},
			},
		},
		{
			"default and case",
			fields{defaultAND: true},
			args{
				NilOperator,
				&QueryTerm{Value: "word"},
			},
			&QueryTerm{
				mustClauses: []*QueryTerm{{Value: "word"}},
			},
		},
		{
			"not operator case",
			fields{defaultAND: false},
			args{
				NotOperator,
				&QueryTerm{Value: "word"},
			},
			&QueryTerm{
				mustNotClauses: []*QueryTerm{{Value: "word", Prohibited: true}},
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			q := &QueryTerm{}

			qp := &QueryParser{
				defaultAND: tt.fields.defaultAND,
			}
			qp.appendQuery(q, tt.args.op, tt.args.qt)

			if diff := cmp.Diff(tt.want, q, cmp.AllowUnexported(QueryTerm{})); diff != "" {
				t.Errorf("QueryParser.appendQuery() %s = mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}

func TestQueryParser_tokenText(t *testing.T) {
	is := is.New(t)

	type fields struct {
		query string
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			"simple word",
			fields{"word"},
			"word",
		},
		{
			"word starts with digit",
			fields{"1word"},
			"1word",
		},
		{
			"identifier",
			fields{"1.04.02"},
			"1.04.02",
		},
		{
			"slash is allowed",
			fields{"4.VEL/123"},
			"4.VEL/123",
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			qp, err := NewQueryParser()
			is.NoErr(err)

			qp.s.Init(strings.NewReader(tt.fields.query))

			// call scan to forward to the first word
			qp.s.Scan()

			if diff := cmp.Diff(tt.want, qp.tokenText()); diff != "" {
				t.Errorf("QueryPares.tokenText(); %s = mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}

func TestQueryParser_Parse(t *testing.T) {
	is := is.New(t)

	type args struct {
		query string
	}

	tests := []struct {
		name    string
		args    args
		want    *QueryTerm
		wantErr bool
	}{
		{
			"correct full query",
			args{"more words"},
			&QueryTerm{
				shouldClauses: []*QueryTerm{
					{Value: "more"},
					{Value: "words"},
				},
			},
			false,
		},
		{
			"bad fuzzy specifier",
			args{"more~1a"},
			nil,
			true,
		},
		{
			"bad boost specifier",
			args{"more^1a"},
			nil,
			true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			qp, err := NewQueryParser()
			is.NoErr(err)

			got, err := qp.Parse(tt.args.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("QueryParser.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("QueryParser.Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isIdentRune(t *testing.T) {
	type args struct {
		ch rune
		i  int
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"unicode letter at start",
			args{rune('a'), 0},
			true,
		},
		{
			"unicode special letter at start",
			args{rune('ç'), 0},
			true,
		},
		{
			"alpha not at start",
			args{rune('z'), 1},
			true,
		},
		{
			"unicode special letter at start",
			args{rune('é'), 10},
			true,
		},
		{
			"digit not at the start",
			args{rune('9'), 1},
			true,
		},
		{
			"digit at the start",
			args{rune('9'), 0},
			true,
		},
		{
			"period at the start",
			args{rune('.'), 0},
			false,
		},
		{
			"period not at the start",
			args{rune('.'), 1},
			true,
		},
		{
			"hyphen not at the start",
			args{rune('-'), 1},
			true,
		},
		{
			"hyphen at the start",
			args{rune('-'), 0},
			false,
		},
		{
			"forward slash at start",
			args{rune('/'), 0},
			true,
		},
		{
			"forward slash not at start",
			args{rune('/'), 0},
			true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			if got := isIdentRune(tt.args.ch, tt.args.i); got != tt.want {
				t.Errorf("isIdentRune() %s = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestQueryTerm_setWildcard(t *testing.T) {
	is := is.New(t)

	type fields struct {
		Value string
	}

	type args struct {
		query string
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   *QueryTerm
		want1  bool
	}{
		{
			"prefix query",
			fields{Value: "prefix"},
			args{query: "prefix*"},
			&QueryTerm{Value: "prefix", PrefixWildcard: true},
			true,
		},
		{
			"suffix query",
			fields{Value: ""},
			args{query: "*suffix"},
			&QueryTerm{Value: "", SuffixWildcard: true},
			true,
		},
		{
			"single wildcard",
			fields{Value: ""},
			args{query: "*"},
			nil,
			false,
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			qt := &QueryTerm{
				Value: tt.fields.Value,
			}

			if qt.Value == "" {
				qt = nil
			}

			qp, err := NewQueryParser()
			is.NoErr(err)

			qp.s.Init(strings.NewReader(tt.args.query))

			qp.s.Scan()

			got, got1 := qt.setWildcard(qp)
			if diff := cmp.Diff(tt.want, got, cmp.AllowUnexported(QueryTerm{})); diff != "" {
				t.Errorf("QueryTerm.setWildcard(); %s = mismatch (-want +got):\n%s", tt.name, diff)
			}

			if got1 != tt.want1 {
				t.Errorf("QueryTerm.setWildcard() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestQueryTerm_isBoolQuery(t *testing.T) {
	is := is.New(t)

	// term with empty clauses
	t.Run("only term query", func(t *testing.T) {
		qt := &QueryTerm{Field: "title", Value: "value"}
		if qt.isBoolQuery() == true {
			t.Errorf("QueryTerm.HasClauses() only term query, should have no clauses")
		}
	})

	tests := []struct {
		name  string
		query string
		want  bool
	}{
		{
			"simple should query",
			"one",
			true,
		},
		{
			"multi word OR query",
			"one | two",
			true,
		},
		{
			"multi word AND query",
			"one + two",
			true,
		},
		{
			"multi word NOT query",
			"-one -two",
			true,
		},
		{
			"single word NOT query",
			"-one",
			true,
		},
	}

	for _, tt := range tests {
		tt := tt

		parser, err := NewQueryParser()
		is.NoErr(err)

		qt, err := parser.Parse(tt.query)
		is.NoErr(err)

		t.Run(tt.name, func(t *testing.T) {
			if got := qt.isBoolQuery(); got != tt.want {
				t.Errorf("QueryTerm.HasClauses() %s = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestQueryTerm_ClauseGetters(t *testing.T) {
	is := is.New(t)

	parser, err := NewQueryParser()
	is.NoErr(err)

	qt, err := parser.Parse("-one OR (two AND three) AND title:four")
	is.NoErr(err)

	is.Equal(len(qt.shouldClauses), len(qt.Should()))

	is.Equal(len(qt.mustClauses), len(qt.Must()))

	is.Equal(len(qt.mustNotClauses), len(qt.MustNot()))
}

func TestQueryType_String(t *testing.T) {
	tests := []struct {
		name string
		qt   QueryType
		want string
	}{
		{"bool query", BoolQuery, "BoolQuery"},
		{"fuzzy query", FuzzyQuery, "FuzzyQuery"},
		{"phrase query", PhraseQuery, "PhraseQuery"},
		{"term query", TermQuery, "TermQuery"},
		{"wildcard query", WildCardQuery, "WildCardQuery"},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			if got := tt.qt.String(); got != tt.want {
				t.Errorf("QueryType.String() %s = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestQueryTerm_Type(t *testing.T) {
	type fields struct {
		Field          string
		Value          string
		Prohibited     bool
		Phrase         bool
		SuffixWildcard bool
		PrefixWildcard bool
		Boost          float64
		Fuzzy          int
		Slop           int
		mustClauses    []*QueryTerm
		mustNotClauses []*QueryTerm
		shouldClauses  []*QueryTerm
		nested         *QueryTerm
	}

	tests := []struct {
		name   string
		fields fields
		want   QueryType
	}{
		{
			"bool query",
			fields{mustNotClauses: []*QueryTerm{{Value: "word"}}},
			BoolQuery,
		},
		{
			"term query",
			fields{Value: "word"},
			TermQuery,
		},
		{
			"prohibited term query",
			fields{Value: "word", Prohibited: true},
			TermQuery,
		},
		{
			"phrase query",
			fields{Value: "two words", Phrase: true},
			PhraseQuery,
		},
		{
			"prefix wildcard query",
			fields{Value: "words", PrefixWildcard: true},
			WildCardQuery,
		},
		{
			"suffix wildcard query",
			fields{Value: "words", SuffixWildcard: true},
			WildCardQuery,
		},
		{
			"fuzzy query",
			fields{Value: "words", Fuzzy: 2},
			FuzzyQuery,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			qt := &QueryTerm{
				Field:          tt.fields.Field,
				Value:          tt.fields.Value,
				Prohibited:     tt.fields.Prohibited,
				Phrase:         tt.fields.Phrase,
				SuffixWildcard: tt.fields.SuffixWildcard,
				PrefixWildcard: tt.fields.PrefixWildcard,
				Boost:          tt.fields.Boost,
				Fuzzy:          tt.fields.Fuzzy,
				Slop:           tt.fields.Slop,
				mustClauses:    tt.fields.mustClauses,
				mustNotClauses: tt.fields.mustNotClauses,
				shouldClauses:  tt.fields.shouldClauses,
				nested:         tt.fields.nested,
			}
			if got := qt.Type(); got != tt.want {
				t.Errorf("QueryTerm.Type() = %v, want %v", got, tt.want)
			}
		})
	}
}
