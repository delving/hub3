package pid

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/driver/gorm"
	"github.com/delving/hub3/ikuzo/rdf"
	"github.com/delving/hub3/ikuzo/render"
)

var _ domain.Service = (*Service)(nil)

type Service struct {
	orgs     domain.OrgConfigRetriever
	log      zerolog.Logger
	ts       domain.TaskService
	dataPath string
	orms     map[string]Store
}

func NewService(options ...Option) (*Service, error) {
	s := &Service{
		orms: make(map[string]Store),
	}

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
	s.log = b.Logger.With().Str("svc", "pid").Logger()
	s.orgs = b.Orgs
	s.ts = b.TaskService

	if s.ts != nil {
		if err := s.registerTaskHandlers(); err != nil {
			s.log.Error().Err(err).Msg("unable to register task handlers, continuing without task support")
		}
	} else {
		s.log.Warn().Msg("task service not available, skipping task handler registration")
	}

	rdf.SetDefault(s)
}

func (s *Service) renderError(w http.ResponseWriter, r *http.Request, err error, statusCode ...int) {
	s.log.Error().Err(err).Msg("unable to handle request")
	if len(statusCode) != 0 {
		code := statusCode[0]
		w.WriteHeader(code)
	}
	render.JSON(w, r, map[string]string{
		"error": err.Error(),
	})
}

func (s *Service) GetStore(orgID string) (store Store, err error) {
	store, ok := s.orms[orgID]
	if !ok {
		store, err = s.newStore(orgID)
		if err != nil {
			return nil, fmt.Errorf("unable to create new store: %w", err)
		}
	}

	return store, nil
}

func (s *Service) newStore(orgID string) (store Store, err error) {
	_, ok := s.orgs.RetrieveConfig(orgID)
	if !ok {
		return nil, domain.ErrOrgNotFound
	}

	db, err := gorm.OpenSqliteDB(filepath.Join(s.dataPath, fmt.Sprintf("pid_%s.db", orgID)), nil)
	if err != nil {
		return nil, err
	}

	store, err = NewGormStore(db)
	if err != nil {
		return nil, err
	}

	s.orms[orgID] = store

	return store, nil
}
