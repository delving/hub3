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

// nolint:gocritic
package elasticsearch

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/delving/hub3/ikuzo/service/x/search"
	"github.com/google/go-cmp/cmp"
	"github.com/matryer/is"
)

func TestNewBoolQuery(t *testing.T) {
	is := is.New(t)

	type fields struct {
		queryFields []QueryField
	}

	type args struct {
		q string
	}

	tests := []struct {
		name       string
		fields     fields
		args       args
		defaultAnd bool
		want       string
	}{
		{
			"empty query is match all",
			fields{},
			args{q: ""},
			false,
			"{\"match_all\":{}}",
		},
		{
			"single word OR query",
			fields{[]QueryField{{Field: "full_text"}}},
			args{q: "one"},
			false,
			`{"bool":{"should":{"match":{"full_text":{"query":"one"}}}}}`,
		},
		{
			"multi word OR query",
			fields{[]QueryField{{Field: "full_text"}}},
			args{q: "one two"},
			false,
			`{"bool":{"should":[` +
				`{"match":{"full_text":{"query":"one"}}},` +
				`{"match":{"full_text":{"query":"two"}}}` +
				`]}}`,
		},
		{
			"AND query",
			fields{[]QueryField{{Field: "full_text"}}},
			args{q: "one AND two"},
			false,
			`{"bool":{"must":[{"match":{"full_text":{"query":"one"}}},{"match":{"full_text":{"query":"two"}}}]}}`,
		},
		{
			"NOT query",
			fields{[]QueryField{{Field: "full_text"}}},
			args{q: "-one"},
			false,
			`{"bool":{"must_not":{"match":{"full_text":{"query":"one"}}}}}`,
		},
		{
			"mix query",
			fields{[]QueryField{{Field: "full_text"}}},
			args{q: `"hello there" AND -one`},
			false,
			`{"bool":{` +
				`"must":{"match_phrase":{"full_text":{"query":"hello there"}}},` +
				`"must_not":{"match":{"full_text":{"query":"one"}}}` +
				`}}`,
		},
	}

	for _, tt := range tests {
		tt := tt

		qp, err := search.NewQueryParser()
		is.NoErr(err)

		qt, err := qp.Parse(tt.args.q)
		is.NoErr(err)

		for _, must := range qt.Must() {
			fmt.Printf("query: %+v", must.Type())
		}

		qb := NewQueryBuilder(tt.fields.queryFields...)

		bq := qb.NewElasticQuery(qt)
		is.True(bq != nil)

		bqSource, err := bq.Source()
		is.NoErr(err)

		got, err := json.Marshal(bqSource)
		is.NoErr(err)
		t.Run(tt.name, func(t *testing.T) {
			if diff := cmp.Diff(tt.want, string(got)); diff != "" {
				t.Errorf("NewBoolQuery(); %s = mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}

// nolint:gocritic
func TestNewElasticQuery(t *testing.T) {
	is := is.New(t)

	type fields struct {
		fields []QueryField
	}

	type args struct {
		q *search.QueryTerm
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			"empty query",
			fields{[]QueryField{{Field: "full_text"}}},
			args{&search.QueryTerm{
				Value: "",
			}},
			`{"match_all":{}}`,
		},
		{
			"single word",
			fields{[]QueryField{{Field: "full_text"}}},
			args{&search.QueryTerm{
				Value: "word",
			}},
			`{"match":{"full_text":{"query":"word"}}}`,
		},
		{
			"fuzzy single word",
			fields{[]QueryField{{Field: "full_text"}}},
			args{&search.QueryTerm{
				Value: "word",
				Fuzzy: 2,
				Boost: 2,
			}},
			`{"match":{"full_text":{"boost":2,"fuzziness":"2","query":"word"}}}`,
		},
		{
			"phrase query",
			fields{[]QueryField{{Field: "full_text"}}},
			args{&search.QueryTerm{
				Value:  "two words",
				Phrase: true,
			}},
			`{"match_phrase":{"full_text":{"query":"two words"}}}`,
		},
		{
			"phrase query with field boost",
			fields{[]QueryField{{Field: "full_text", Boost: 1}}},
			args{&search.QueryTerm{
				Value:  "two words",
				Phrase: true,
			}},
			`{"match_phrase":{"full_text":{"boost":1,"query":"two words"}}}`,
		},
		{
			"phrase query with slop",
			fields{[]QueryField{{Field: "full_text"}}},
			args{&search.QueryTerm{
				Value:  "two words",
				Phrase: true,
				Fuzzy:  2,
				Boost:  2,
			}},
			`{"match_phrase":{"full_text":{"boost":2,"query":"two words","slop":2}}}`,
		},
		{
			"multi field boost",
			fields{[]QueryField{{Field: "title", Boost: 3}, {Field: "subject"}}},
			args{&search.QueryTerm{
				Value: "word",
			}},
			`{"dis_max":{"queries":[` +
				`{"match":{"title":{"boost":3,"query":"word"}}},` +
				`{"match":{"subject":{"query":"word"}}}` +
				`]}}`,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			qb := NewQueryBuilder(tt.fields.fields...)
			bq := qb.NewElasticQuery(tt.args.q)
			is.True(bq != nil)

			bqSource, err := bq.Source()
			is.NoErr(err)

			got, err := json.Marshal(bqSource)
			is.NoErr(err)

			if diff := cmp.Diff(tt.want, string(got)); diff != "" {
				t.Errorf("NewElasticQuery(); %s = mismatch (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}
