package nde

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/delving/hub3/hub3/ead"
	"github.com/delving/hub3/ikuzo/domain"
	"github.com/go-chi/chi"
)

type Option func(*Service) error

type Service struct {
	defaultCfg *RegisterConfig
	cfgs       []*RegisterConfig
	lookUp     map[string]*RegisterConfig
}

func NewService(options ...Option) (*Service, error) {
	s := &Service{
		lookUp: map[string]*RegisterConfig{},
	}

	// apply options
	for _, option := range options {
		if err := option(s); err != nil {
			return nil, err
		}
	}

	for _, cfg := range s.cfgs {
		if cfg.URLPrefix == "default" {
			s.defaultCfg = cfg
		}

		s.lookUp[cfg.URLPrefix] = cfg
	}

	return s, nil
}

func SetConfig(cfgs []*RegisterConfig) Option {
	return func(s *Service) error {
		s.cfgs = cfgs
		return nil
	}
}

func (s *Service) HandleDataset(w http.ResponseWriter, r *http.Request) {
	spec := chi.URLParam(r, "spec")
	orgID := domain.GetOrganizationID(r)

	dataset, err := s.getDataset(orgID.String(), spec)
	if err != nil {
		if errors.Is(err, ead.ErrFileNotFound) {
			http.Error(w, "dataset not found", http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	s.renderJSONLD(w, r, dataset, http.StatusOK)
}

func (s *Service) HandleCatalog(w http.ResponseWriter, r *http.Request) {
	orgID := domain.GetOrganizationID(r)

	cfgName := chi.URLParam(r, "cfgName")

	cfg, ok := s.lookUp[cfgName]
	if !ok {
		http.Error(w, fmt.Errorf("unable to find config: %q", cfgName).Error(), http.StatusNotFound)
		return
	}

	catalog := cfg.newCatalog()

	if err := s.AddDatasets(orgID.String(), catalog, cfg.RecordTypeFilter); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.renderJSONLD(w, r, catalog, http.StatusOK)
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

func (s *Service) Shutdown(ctx context.Context) error {
	return nil
}

// JSON marshals 'v' to JSON, automatically escaping HTML and setting the
// Content-Type as application/json.
func (s *Service) renderJSONLD(w http.ResponseWriter, r *http.Request, v interface{}, status int) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(true)

	if err := enc.Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/ld+json; charset=utf-8")

	if status != 0 {
		w.WriteHeader(status)
	}

	w.Write(buf.Bytes())
}
