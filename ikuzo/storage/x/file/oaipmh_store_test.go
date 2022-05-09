package file

import (
	"context"
	"testing"

	"github.com/delving/hub3/ikuzo/service/x/oaipmh"
	"github.com/matryer/is"
)

func TestOAIPMHStore(t *testing.T) {
	var store *OAIPMHStore
	ctx := context.TODO()

	t.Run("newStore", func(t *testing.T) {
		is := is.New(t)
		var err error

		store, err = NewOAIPMHStore()
		is.NoErr(err)

		store.Path = "./testdata/oaipmh"
	})

	t.Run("filtersets", func(t *testing.T) {
		is := is.New(t)

		sets, err := store.filterSets("orgid", "", "")
		is.NoErr(err)
		is.Equal(len(sets), 4)

		sets, err = store.filterSets("orgid", "", "raw")
		is.NoErr(err)
		is.Equal(len(sets), 1)

		sets, err = store.filterSets("orgid", "", "edm")
		is.NoErr(err)
		is.Equal(len(sets), 3)

		sets, err = store.filterSets("orgid", "a", "edm")
		is.NoErr(err)
		is.Equal(len(sets), 2)
	})

	t.Run("listsets", func(t *testing.T) {
		is := is.New(t)

		sets, errors, err := store.ListSets(ctx, &oaipmh.QueryConfig{OrgID: "orgid"})
		is.NoErr(err)
		is.Equal(len(errors), 0)
		is.Equal(len(sets), 4)
	})

	t.Run("listidentifiers", func(t *testing.T) {
		is := is.New(t)
		q := &oaipmh.QueryConfig{OrgID: "orgid", DatasetID: "1", MetadataPrefix: "raw"}
		headers, errors, err := store.ListIdentifiers(ctx, q)
		is.NoErr(err)
		is.Equal(len(errors), 0)
		is.Equal(q.TotalSize, 3)
		is.Equal(len(headers), 3)
	})
}
