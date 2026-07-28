package pid

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

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

	mu          sync.Mutex
	orms        map[string]Store
	failedUntil map[string]time.Time
}

func NewService(options ...Option) (*Service, error) {
	s := &Service{
		orms:        make(map[string]Store),
		failedUntil: make(map[string]time.Time),
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

// storeFailureBackoff is how long a failed store open is remembered before a
// retry is allowed. Without it a persistent open failure (e.g. a missing
// dataPath directory) was retried for EVERY record during bulk indexing —
// ~200 attempts/second — and each failed modernc/sqlite open leaks
// allocations outside the Go heap, ballooning RSS until the OOM killer
// fires (observed live: 10GB RSS with a 170MB Go heap).
const storeFailureBackoff = time.Minute

func (s *Service) GetStore(orgID string) (store Store, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	store, ok := s.orms[orgID]
	if !ok {
		if until, failed := s.failedUntil[orgID]; failed && time.Now().Before(until) {
			return nil, fmt.Errorf("pid store for %s unavailable (previous open failed; retry after backoff)", orgID)
		}
		store, err = s.newStore(orgID)
		if err != nil {
			s.failedUntil[orgID] = time.Now().Add(storeFailureBackoff)
			return nil, fmt.Errorf("unable to create new store: %w", err)
		}
		delete(s.failedUntil, orgID)
	}

	return store, nil
}

// newStore must be called with s.mu held.
func (s *Service) newStore(orgID string) (store Store, err error) {
	_, ok := s.orgs.RetrieveConfig(orgID)
	if !ok {
		return nil, domain.ErrOrgNotFound
	}

	// A missing data directory surfaces from the sqlite driver as the
	// misleading "unable to open database file: out of memory (14)"
	// (SQLITE_CANTOPEN); create it up front instead.
	if err := os.MkdirAll(s.dataPath, 0o755); err != nil {
		return nil, fmt.Errorf("unable to create pid data path %s: %w", s.dataPath, err)
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
