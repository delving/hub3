package namespace

import (
	"context"

	"github.com/delving/hub3/ikuzo/definitions/generated"
	"github.com/delving/hub3/ikuzo/domain"
	"github.com/pacedotdev/oto/otohttp"
	"github.com/rs/zerolog/log"
)

type otoService struct {
	s *Service
}

func (o *otoService) GetNamespace(ctx context.Context, r generated.GetNamespaceRequest) (*generated.GetNamespaceResponse, error) {
	ns, err := o.s.GetWithPrefix(r.Prefix)
	if err != nil {
		return nil, err
	}

	return &generated.GetNamespaceResponse{Namespaces: []*domain.Namespace{ns}}, nil
}

func (o *otoService) ListNamespace(ctx context.Context, r generated.ListNamespaceRequest) (*generated.ListNamespaceResponse, error) {
	if r.Prefix != "" {
		ns, err := o.s.GetWithPrefix(r.Prefix)
		if err != nil {
			return nil, err
		}

		return &generated.ListNamespaceResponse{Namespaces: []*domain.Namespace{ns}}, nil
	}

	ns, err := o.s.List()
	if err != nil {
		log.Error().Err(err).Msg("list error")
		return nil, err
	}

	return &generated.ListNamespaceResponse{Namespaces: ns}, nil
}

func (s *Service) RegisterOtoService(server *otohttp.Server) error {
	generated.RegisterNamespaceService(server, &otoService{s: s})
	return nil
}
