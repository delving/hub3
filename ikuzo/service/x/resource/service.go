package resource

import (
	"context"
	"net/http"
)

type Option func(*Service) error

type Service struct {
}

func NewService(options ...Option) (*Service, error) {
	s := &Service{}

	// apply options
	for _, option := range options {
		if err := option(s); err != nil {
			return nil, err
		}
	}

	return s, nil
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

func (s *Service) Shutdown(ctx context.Context) error {
	return nil
}

func (s *Service) Publish() error {
	// TODO(kiivihal): implement publish method
	return nil
}

func (s *Service) DropResources() error {
	return nil
}

func (s *Service) DropOrphans() error {
	return nil
}
