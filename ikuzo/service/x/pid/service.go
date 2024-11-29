package pid

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"gorm.io/gorm"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/render"
)

var _ domain.Service = (*Service)(nil)

type Service struct {
	orgs  domain.OrgConfigRetriever
	log   zerolog.Logger
	ts    domain.TaskService
	store Store
	orm   *gorm.DB
}

func NewService(options ...Option) (*Service, error) {
	s := &Service{}

	// apply options
	for _, option := range options {
		if err := option(s); err != nil {
			return nil, err
		}
	}

	if err := s.registerTaskHandlers(); err != nil {
		return nil, err
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
