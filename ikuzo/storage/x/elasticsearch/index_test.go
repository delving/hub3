package elasticsearch

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/delving/hub3/ikuzo/storage/x/elasticsearch/mapping"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/matryer/is"
)

// nolint:gocritic
func TestIndex(t *testing.T) {
	is := is.New(t)

	es, err := elasticsearch.NewDefaultClient()
	is.NoErr(err)

	res, err := es.Info()
	is.NoErr(err)

	is.Equal(res.StatusCode, http.StatusOK)

	// check what happens when index does not exist
	err = IndexExists(es, "hub3test")
	is.True(errors.Is(err, ErrIndexNotFound))

	// check what happens when you delete an index that does not exist
	err = IndexDelete(es, "hub3test")
	is.True(errors.Is(err, ErrIndexNotFound))

	// create index that does not exist
	name, err := IndexCreate(es, "hub3test", mapping.V1ESMapping(0, 0), true)
	is.NoErr(err)

	// delete the test index
	defer func() {
		err = IndexDelete(es, name)
		is.True(errors.Is(err, ErrIndexNotFound))
	}()

	// check on index name
	is.True(strings.HasPrefix(name, "hub3test_"))

	// index must now exist
	err = IndexExists(es, name)
	is.NoErr(err)

	// the alias must also exist
	createIndexName, err := AliasGet(es, "hub3test")
	is.NoErr(err)
	is.Equal(name, createIndexName)

	// can't recreate an index with same alias
	newName, err := IndexCreate(es, "hub3test", "", true)
	is.True(errors.Is(err, ErrIndexAlreadyCreated))
	is.Equal(newName, name)

	secondIndexName, err := IndexCreate(es, "hub3test", mapping.V2ESMapping(0, 0), false)
	is.NoErr(err)

	// check on index name
	is.True(strings.HasPrefix(secondIndexName, "hub3test_"))

	// delete the second test index
	defer func() {
		err = IndexDelete(es, secondIndexName)
		is.NoErr(err)
	}()

	err = IndexExists(es, secondIndexName)
	is.NoErr(err)

	// switch the alias and delete the first index
	oldIndexName, err := IndexSwitch(es, "hub3test", secondIndexName, true)
	is.NoErr(err)
	is.Equal(name, oldIndexName)

	// second index should now be linked to the alias
	aliasIndex, err := AliasGet(es, "hub3test")
	is.NoErr(err)
	is.Equal(secondIndexName, aliasIndex)

	// first index should no longer exist
	err = IndexExists(es, name)
	is.True(errors.Is(err, ErrIndexNotFound))
}
