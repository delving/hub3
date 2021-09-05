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
	"net/http"
	"strings"
	"testing"

	"github.com/delving/hub3/ikuzo/driver/elasticsearch"
	"github.com/delving/hub3/ikuzo/driver/elasticsearch/internal/mapping"
	"github.com/matryer/is"
)

// nolint:gocritic
func TestIndex(t *testing.T) {
	is := is.New(t)

	host := hostAndPort

	cfg := elasticsearch.DefaultConfig()
	cfg.Urls = []string{fmt.Sprintf("http://%s", host)}
	cfg.DisableMetrics = true

	es, err := elasticsearch.NewClient(cfg)
	is.NoErr(err)

	res, err := es.Ping()
	is.NoErr(err)

	is.Equal(res.StatusCode, http.StatusOK)

	indices := es.Indices()

	// check what happens when index does not exist
	err = indices.Exists("hub3test")
	is.True(errors.Is(err, elasticsearch.ErrIndexNotFound))

	// check what happens when you delete an index that does not exist
	err = indices.Delete("hub3test")
	is.True(errors.Is(err, elasticsearch.ErrIndexNotFound))

	// create index that does not exist
	name, err := indices.Create("hub3test", mapping.V1ESMapping(0, 0), true)
	is.NoErr(err)

	// delete the test index
	defer func() {
		err = indices.Delete(name)
		is.True(errors.Is(err, elasticsearch.ErrIndexNotFound))
	}()

	// check on index name
	is.True(strings.HasPrefix(name, "hub3test_"))

	// index must now exist
	err = indices.Exists(name)
	is.NoErr(err)

	alias := es.Alias()

	// the alias must also exist
	createIndexName, err := alias.Get("hub3test")
	is.NoErr(err)
	is.Equal(name, createIndexName)

	// can't recreate an index with same alias
	newName, err := indices.Create("hub3test", "", true)
	is.True(errors.Is(err, elasticsearch.ErrIndexAlreadyCreated))
	is.Equal(newName, name)

	secondIndexName, err := indices.Create("hub3test", mapping.V2ESMapping(0, 0), false)
	is.NoErr(err)

	// check on index name
	is.True(strings.HasPrefix(secondIndexName, "hub3test_"))

	// delete the second test index
	defer func() {
		err = indices.Delete(secondIndexName)
		is.NoErr(err)
	}()

	err = indices.Exists(secondIndexName)
	is.NoErr(err)

	// switch the alias and delete the first index
	oldIndexName, err := indices.Switch("hub3test", secondIndexName, true)
	is.NoErr(err)
	is.Equal(name, oldIndexName)

	// second index should now be linked to the alias
	aliasIndex, err := alias.Get("hub3test")
	is.NoErr(err)
	is.Equal(secondIndexName, aliasIndex)

	// first index should no longer exist
	err = indices.Exists(name)
	is.True(errors.Is(err, elasticsearch.ErrIndexNotFound))
}
