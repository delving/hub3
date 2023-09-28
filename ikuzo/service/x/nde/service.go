package nde

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"

	"github.com/delving/hub3/hub3/ead"
	"github.com/delving/hub3/hub3/models"
	"github.com/delving/hub3/ikuzo/domain"
)

type Option func(*Service) error

type Service struct {
	defaultCfg       *RegisterConfig
	cfgs             []*RegisterConfig
	lookUp           map[string]*RegisterConfig
	recordTypeLookup map[string]*RegisterConfig
	orgs             domain.OrgConfigRetriever
	log              zerolog.Logger
	ctx              context.Context
	cancel           context.CancelFunc
	orgID            string
}

func NewService(options ...Option) (*Service, error) {
	s := &Service{
		lookUp:           map[string]*RegisterConfig{},
		recordTypeLookup: map[string]*RegisterConfig{},
	}

	// apply options
	for _, option := range options {
		if err := option(s); err != nil {
			return nil, err
		}
	}

	for _, cfg := range s.cfgs {
		if cfg.Default {
			s.defaultCfg = cfg
		}

		s.lookUp[cfg.URLPrefix] = cfg
		s.recordTypeLookup[cfg.RecordTypeFilter] = cfg
	}

	s.ctx, s.cancel = context.WithCancel(context.Background())

	go func() {
		if err := s.scheduleNarthexUpdate(); err != nil {
			s.log.Err(err).Msg("unable to run scheduled narthex update")
		}
	}()

	return s, nil
}

func (s *Service) HandleDataset(w http.ResponseWriter, r *http.Request) {
	spec := chi.URLParam(r, "spec")
	orgID := domain.GetOrganizationID(r)
	if s.orgID == "" {
		s.orgID = orgID.String()
		go func() {
			if err := s.updateTitles(s.ctx, s.orgID); err != nil {
				s.log.Error().Err(err).Msg("unable to run narthex update titles")
			}
		}()
	}

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

func (s *Service) enabledConfig() []string {
	var cfgs []string
	for prefix := range s.lookUp {
		cfgs = append(cfgs, prefix)
	}

	return cfgs
}

func (s *Service) HandleCatalog(w http.ResponseWriter, r *http.Request) {
	orgID := domain.GetOrganizationID(r)
	if s.orgID == "" {
		s.orgID = orgID.String()
		go func() {
			if err := s.updateTitles(s.ctx, s.orgID); err != nil {
				s.log.Error().Err(err).Msg("unable to run narthex update titles")
			}
		}()
	}

	cfgName := chi.URLParam(r, "cfgName")

	cfg, ok := s.lookUp[cfgName]
	if !ok {
		http.Error(
			w,
			fmt.Errorf("unable to find config: %q \n allowed config: %#v", cfgName, s.enabledConfig()).Error(),
			http.StatusNotFound,
		)

		return
	}

	catalog := cfg.newCatalog()

	total, err := s.getDatasetsCount(orgID.String(), cfg.RecordTypeFilter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := catalog.addHydraView(r.URL.Query().Get("page"), total); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := s.AddDatasets(orgID.String(), catalog, cfg.RecordTypeFilter); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.renderJSONLD(w, r, catalog, http.StatusOK)
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

func (s *Service) Shutdown(ctx context.Context) error {
	s.cancel()
	return nil
}

func (s *Service) SetServiceBuilder(b *domain.ServiceBuilder) {
	s.log = b.Logger.With().Str("svc", "sitemap").Logger()
	s.orgs = b.Orgs
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

func (s *Service) scheduleNarthexUpdate() error {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	eg, _ := errgroup.WithContext(s.ctx)

	for {
		select {
		case <-s.ctx.Done():
			if err := eg.Wait(); err != nil {
				return fmt.Errorf("error while waiting for go-routines to shut down: %w", err)
			}
			return s.ctx.Err()
		case <-ticker.C:
			if s.orgID == "" { // only run when the orgID is set
				continue
			}
			if err := s.updateTitles(s.ctx, s.orgID); err != nil {
				// handle err
				return err
			}
		}
	}
}

func (s *Service) updateTitles(ctx context.Context, orgID string) error {
	titles, err := models.GetNDETitles(orgID)
	if err != nil {
		return err
	}

	for _, title := range titles {
		if title.Spec == "" {
			continue
		}

		if s.ctx.Err() != nil {
			return s.ctx.Err()
		}

		ds, err := models.GetDataSet(orgID, title.Spec)
		if err != nil {
			s.log.Warn().Err(err).Msgf("unable to find a dataset with spec %q", title.Spec)
			continue
		}

		ds.Description = title.Description
		ds.Rights = title.Rights
		if err := ds.Save(); err != nil {
			s.log.Error().Err(err).Msgf("unable to save dataset info for: %q", title.Spec)
		}
	}

	return nil
}
