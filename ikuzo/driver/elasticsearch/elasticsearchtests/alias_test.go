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

package elasticsearchtests

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/delving/hub3/ikuzo/driver/elasticsearch"
	"github.com/matryer/is"
)

// nolint:gocritic
func TestAlias(t *testing.T) {
	is := is.New(t)

	host := hostAndPort

	cfg := elasticsearch.DefaultConfig()
	cfg.Urls = []string{fmt.Sprintf("http://%s", host)}
	cfg.DisableMetrics = true

	es, err := elasticsearch.NewClient(cfg)
	is.NoErr(err)

	res, err := es.Ping()
	t.Logf("elasticsearch status: %s", res.Status())
	is.NoErr(err)
	is.True(res.IsError() == false)

	alias := es.Alias()

	indexName, err := alias.Get("unknownalias")
	is.True(errors.Is(err, elasticsearch.ErrAliasNotFound))
	is.Equal(indexName, "")

	err = alias.Delete("hub3test-alias", "")
	is.True(errors.Is(err, elasticsearch.ErrIndexNotFound))

	err = alias.Delete("hub3test-alias", "")
	is.True(errors.Is(err, elasticsearch.ErrIndexNotFound))

	// alias not created because unknown index
	err = alias.Create("hub3-alias", "hub3test")
	is.True(errors.Is(err, elasticsearch.ErrIndexNotFound))

	indices := es.Indices()

	firstIndexName, err := indices.Create("hub3test", "", false)
	is.NoErr(err)
	is.True(strings.HasPrefix(firstIndexName, "hub3test_"))

	secondIndexName, err := indices.Create("hub3test2", "", false)
	is.NoErr(err)
	is.True(strings.HasPrefix(secondIndexName, "hub3test2_"))

	defer func() {
		err = indices.Delete(firstIndexName)
		is.NoErr(err)
		err = indices.Delete(secondIndexName)
		is.NoErr(err)
	}()

	err = alias.Delete("hub3test-alias", firstIndexName)
	is.True(errors.Is(err, elasticsearch.ErrAliasNotFound))

	err = alias.Create("hub3test-alias", firstIndexName)
	is.NoErr(err)

	indexName, err = alias.Get("hub3test-alias")
	is.NoErr(err)
	is.Equal(firstIndexName, indexName)

	oldIndexName, err := alias.Update("hub3test-alias", secondIndexName)
	is.NoErr(err)
	is.Equal(firstIndexName, oldIndexName)

	indexName, err = alias.Get("hub3test-alias")
	is.NoErr(err)
	is.Equal(secondIndexName, indexName)

	err = alias.Delete("hub3test-alias", firstIndexName)
	is.True(errors.Is(err, elasticsearch.ErrAliasNotFound))

	err = alias.Delete("hub3test-alias", secondIndexName)
	is.NoErr(err)

	_, err = alias.Get("hub3test-alias")
	is.True(errors.Is(err, elasticsearch.ErrAliasNotFound))
}
