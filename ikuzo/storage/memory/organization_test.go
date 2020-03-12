// nolint:gocritic
package memory

import (
	"context"
	"errors"
	"testing"

	"github.com/matryer/is"
	"github.com/delving/hub3/ikuzo/domain"
)

func TestMemoryStore(t *testing.T) {
	is := is.New(t)
	ctx := context.TODO()

	store := NewOrganizationStore()
	orgs, err := store.List(ctx)
	is.NoErr(err)
	is.Equal(len(orgs), 0)

	// test put
	orgID, err := domain.NewOrganizationID("demo")
	is.NoErr(err)

	err = store.Put(ctx, domain.Organization{ID: orgID})
	is.NoErr(err)

	// should have one org
	orgs, err = store.List(ctx)
	is.NoErr(err)
	is.Equal(len(orgs), 1)

	// get an org
	getOrgID, err := store.Get(ctx, orgID)
	is.NoErr(err)
	is.Equal(orgID, getOrgID.ID)

	// delete an org
	err = store.Delete(ctx, orgID)
	is.NoErr(err)
	orgs, err = store.List(ctx)
	is.NoErr(err)
	is.Equal(len(orgs), 0)

	// org not found
	getOrgID, err = store.Get(ctx, orgID)
	is.True(errors.Is(err, domain.ErrOrgNotFound))
}

func TestService_Shutdown(t *testing.T) {
	is := is.New(t)

	ts := NewOrganizationStore()

	is.True(!ts.shutdownCalled)

	err := ts.Shutdown(context.TODO())
	is.NoErr(err)

	is.True(ts.shutdownCalled)
}
