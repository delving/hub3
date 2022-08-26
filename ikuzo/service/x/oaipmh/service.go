package oaipmh

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
	"github.com/teris-io/shortid"
)

var _ domain.Service = (*Service)(nil)

type Service struct {
	orgs                  domain.OrgConfigRetriever
	log                   zerolog.Logger
	store                 Store
	steps                 map[string]RequestConfig
	m                     sync.RWMutex
	sid                   *shortid.Shortid
	requireSetSpecForList bool
}

func NewService(options ...Option) (*Service, error) {
	s := &Service{
		steps:                 make(map[string]RequestConfig),
		requireSetSpecForList: true,
	}

	sid, err := shortid.New(1, shortid.DefaultABC, 2342)
	if err != nil {
		return nil, fmt.Errorf("unable to get seed generator: %w", err)
	}

	s.sid = sid

	// apply options
	for _, option := range options {
		if err := option(s); err != nil {
			return nil, err
		}
	}

	return s, nil
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router := chi.NewRouter()
	s.Routes("", router)
	router.ServeHTTP(w, r)
}

func (s *Service) Shutdown(ctx context.Context) error {
	return nil
}

func (s *Service) SetServiceBuilder(b *domain.ServiceBuilder) {
	s.log = b.Logger.With().Str("svc", "oai").Logger()
	s.orgs = b.Orgs
}
