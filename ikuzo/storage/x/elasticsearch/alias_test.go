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

package elasticsearch

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/matryer/is"
)

// nolint:gocritic
func TestAlias(t *testing.T) {
	is := is.New(t)

	es, err := elasticsearch.NewDefaultClient()
	is.NoErr(err)

	res, err := es.Info()
	is.NoErr(err)
	is.True(res.IsError() == false)

	indexName, err := AliasGet(es, "unknownalias")
	is.True(errors.Is(err, ErrAliasNotFound))
	is.Equal(indexName, "")

	err = AliasDelete(es, "", "hub3test-alias")
	is.True(errors.Is(err, ErrIndexNotFound))

	err = AliasDelete(es, "", "hub3test-alias")
	is.True(errors.Is(err, ErrIndexNotFound))

	// alias not created because unknown index
	err = AliasCreate(es, "hub3-alias", "hub3test")
	is.True(errors.Is(err, ErrIndexNotFound))

	firstIndexName, err := IndexCreate(es, "hub3test", "", false)
	is.NoErr(err)
	is.True(strings.HasPrefix(firstIndexName, "hub3test_"))

	secondIndexName, err := IndexCreate(es, "hub3test2", "", false)
	is.NoErr(err)
	is.True(strings.HasPrefix(secondIndexName, "hub3test2_"))

	defer func() {
		err = IndexDelete(es, firstIndexName)
		is.NoErr(err)
		err = IndexDelete(es, secondIndexName)
		is.NoErr(err)
	}()

	err = AliasDelete(es, firstIndexName, "hub3test-alias")
	is.True(errors.Is(err, ErrAliasNotFound))

	err = AliasCreate(es, "hub3test-alias", firstIndexName)
	is.NoErr(err)

	indexName, err = AliasGet(es, "hub3test-alias")
	is.NoErr(err)
	is.Equal(firstIndexName, indexName)

	oldIndexName, err := AliasUpdate(es, "hub3test-alias", secondIndexName)
	is.NoErr(err)
	is.Equal(firstIndexName, oldIndexName)

	indexName, err = AliasGet(es, "hub3test-alias")
	is.NoErr(err)
	is.Equal(secondIndexName, indexName)

	err = AliasDelete(es, firstIndexName, "hub3test-alias")
	is.True(errors.Is(err, ErrAliasNotFound))

	err = AliasDelete(es, secondIndexName, "hub3test-alias")
	is.NoErr(err)

	_, err = AliasGet(es, "hub3test-alias")
	is.True(errors.Is(err, ErrAliasNotFound))
}

func Test_getIndexNameFromAlias(t *testing.T) {
	type args struct {
		r io.Reader
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"sample",
			args{strings.NewReader(
				`{
					"logs_20302801" : {
						"aliases" : {
						"2030" : {}
					}
				}`,
			)},
			"logs_20302801",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			if got := getIndexNameFromAlias(tt.args.r); got != tt.want {
				t.Errorf("getIndexNameFromAlias() = %q, want %v", got, tt.want)
			}
		})
	}
}
