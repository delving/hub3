package lod

import (
	"context"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/rdf"
)

type Resolver interface {
	Resolve(ctx context.Context, orgID domain.OrganizationID, s rdf.Subject) (g *rdf.Graph, err error)
}
