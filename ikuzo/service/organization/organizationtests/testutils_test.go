package organizationtests_test

import (
	"context"
	"testing"

	"github.com/delving/hub3/ikuzo/service/organization/organizationtests"
	"github.com/matryer/is"
)

func TestNewTestOrganizationService(t *testing.T) {
	//nolint:gocritic
	is := is.New(t)

	svc := organizationtests.NewTestOrganizationService()
	is.True(svc != nil)

	orgs, err := svc.Filter(context.TODO())
	is.NoErr(err)
	is.Equal(len(orgs), 2)
}
