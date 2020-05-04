package bulk

import (
	"context"
	"net/http"

	"github.com/delving/hub3/hub3"
	"github.com/delving/hub3/ikuzo/service/x/index"
	"github.com/go-chi/render"
)

type Option func(*Service) error

type Service struct {
	index *index.Service
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

func SetIndexService(is *index.Service) Option {
	return func(s *Service) error {
		s.index = is
		return nil
	}
}

// bulkApi receives bulkActions in JSON form (1 per line) and processes them in
// ingestion pipeline.
func (s *Service) Handle(w http.ResponseWriter, r *http.Request) {
	// TODO(kiivihal): decide what to do with workerpool
	response, err := hub3.ReadActions(r.Context(), r.Body, s.index, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, response)
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

func (s *Service) Shutdown(ctx context.Context) error {
	return nil
}
