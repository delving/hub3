package pid

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/render"
)

var _ domain.Service = (*Service)(nil)

type Service struct {
	orgs  domain.OrgConfigRetriever
	log   zerolog.Logger
	store Store
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

func (s *Service) Update(pid PID) error {
	if pid.Type == 0 {
		if strings.HasPrefix(pid.ID, "ark:") {
			pid.Type = Ark
		}
	}

	storedPid, err := s.store.Get(pid.ID)
	if err != nil && !errors.Is(err, ErrTombstone) {
		return err
	}

	if err := pid.Merge(storedPid); err != nil {
		return err
	}

	return s.store.Save(pid)
}
