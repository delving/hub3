package rdf

import (
	"context"
	"errors"
)

var defaultResourceGetter ResourceGetter = resourceGetterNoOp{}

type ResourceGetter interface {
	GetResource(ctx context.Context, uri string) (*Resource, error)
	GetGraph(ctx context.Context, uri string) (*Graph, error)
}

func DefaultResourceGetter() ResourceGetter {
	return defaultResourceGetter
}

func SetDefault(rg ResourceGetter) {
	defaultResourceGetter = rg
}

type resourceGetterNoOp struct{}

func (resourceGetterNoOp) GetResource(ctx context.Context, uri string) (*Resource, error) {
	return nil, errors.New("no resource getter configured")
}

func (resourceGetterNoOp) GetGraph(ctx context.Context, uri string) (*Graph, error) {
	return nil, errors.New("no resource getter configured")
}
