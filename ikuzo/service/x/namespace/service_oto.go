package namespace

import (
	"context"

	"github.com/delving/hub3/ikuzo/definitions/generated"
	"github.com/delving/hub3/ikuzo/domain"
	"github.com/pacedotdev/oto/otohttp"
)

type otoService struct {
	s *Service
}

func (o *otoService) GetNamespace(ctx context.Context, r generated.GetNamespaceRequest) (*generated.GetNamespaceResponse, error) {
	ns, err := o.s.Get(r.ID)
	if err != nil {
		return nil, err
	}

	return &generated.GetNamespaceResponse{Namespaces: []*domain.Namespace{ns}}, nil
}

func (o *otoService) DeleteNamespace(ctx context.Context, r generated.DeleteNamespaceRequest) (*generated.DeleteNamespaceResponse, error) {
	return &generated.DeleteNamespaceResponse{}, o.s.Delete(r.ID)
}

func (o *otoService) PutNamespace(ctx context.Context, r generated.PutNamespaceRequest) (*generated.PutNamespaceResponse, error) {
	_, err := o.s.Put(r.Namespace.Prefix, r.Namespace.Base)
	return &generated.PutNamespaceResponse{}, err
}

func (o *otoService) Search(ctx context.Context, r generated.SearchNamespaceRequest) (*generated.SearchNamespaceResponse, error) {
	if r.Prefix != "" {
		ns, err := o.s.GetWithPrefix(r.Prefix)
		if err != nil {
			return &generated.SearchNamespaceResponse{}, err
		}

		return &generated.SearchNamespaceResponse{Hits: []*domain.Namespace{ns}}, nil
	}

	if r.BaseURI != "" {
		ns, err := o.s.GetWithBase(r.Prefix)
		if err != nil {
			return &generated.SearchNamespaceResponse{}, err
		}

		return &generated.SearchNamespaceResponse{Hits: []*domain.Namespace{ns}}, nil
	}

	ns, err := o.s.List()
	if err != nil {
		return &generated.SearchNamespaceResponse{}, err
	}

	return &generated.SearchNamespaceResponse{Hits: ns}, nil
}

func (s *Service) RegisterOtoService(server *otohttp.Server) error {
	generated.RegisterNamespaceService(server, &otoService{s: s})
	return nil
}
