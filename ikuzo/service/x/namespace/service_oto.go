package namespace

import (
	"context"

	"github.com/delving/hub3/ikuzo/definitions/generated"
	"github.com/pacedotdev/oto/otohttp"
)

type otoService struct {
	s *Service
}

func (o *otoService) GetNamespace(ctx context.Context, r generated.GetNamespaceRequest) (*generated.GetNamespaceResponse, error) {
	return nil, nil
}

func (s *Service) RegisterOtoService(server *otohttp.Server) error {
	generated.RegisterNamespaceService(server, &otoService{s: s})
	return nil
}
