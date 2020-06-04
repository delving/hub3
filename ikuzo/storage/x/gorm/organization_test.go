package gorm

import (
	"context"
	"errors"
	"testing"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/matryer/is"
)

// nolint:gocritic
func TestOrganizationStore(t *testing.T) {
	t.Skip("removed sqlite so skipping this test for now")
	is := is.New(t)

	db, err := NewDB("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("unable to create database: %#v", err)
	}

	defer db.Close()

	// drop table if exist
	db.DropTableIfExists(domain.Organization{})

	o, err := NewOrganizationStore(db)
	is.NoErr(err)

	orgs, err := o.Filter(context.TODO())
	is.NoErr(err)
	is.True(len(orgs) == 0)

	count, err := o.Count(context.TODO())
	is.NoErr(err)
	is.True(count == 0)

	err = o.Put(context.TODO(), domain.Organization{ID: "demo"})
	is.NoErr(err)

	count, err = o.Count(context.TODO())
	is.NoErr(err)
	is.True(count == 1)

	org, err := o.Get(context.TODO(), "demo")
	is.NoErr(err)
	is.Equal(org.ID, domain.OrganizationID("demo"))

	orgs, err = o.Filter(context.TODO(), domain.OrganizationFilter{
		OffSet: 5,
		Limit:  10,
		Org: domain.Organization{
			ID: domain.OrganizationID("demo"),
		},
	})
	is.NoErr(err)
	is.True(len(orgs) == 0)

	// not found
	org, err = o.Get(context.TODO(), "unknown")
	is.True(errors.Is(err, domain.ErrOrgNotFound))

	// delete organization
	err = o.Delete(context.TODO(), "demo")
	is.NoErr(err)

	count, err = o.Count(context.TODO())
	is.NoErr(err)
	is.True(count == 0)

	orgs, err = o.Filter(context.TODO())
	is.NoErr(err)
	is.True(len(orgs) == 0)

	// multiple close calls
	is.NoErr(o.Shutdown(context.TODO()))
	is.NoErr(o.Shutdown(context.TODO()))
}
