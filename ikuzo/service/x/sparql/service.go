package sparql

import (
	"bytes"
	"context"
	"net/http"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/rdf"
	"github.com/delving/hub3/ikuzo/service/x/lod"
	"github.com/go-chi/chi"
	"github.com/knakk/sparql"
	"github.com/rs/zerolog"
)

var _ domain.Service = (*Service)(nil)

type Service struct {
	bank sparql.Bank
	// store      TripleStore
	orgs       domain.OrgConfigRetriever
	log        zerolog.Logger
	repos      map[domain.OrganizationID]*Repo
	retry      int
	timeout    int
	queryLimit int
}

func NewService(options ...Option) (*Service, error) {
	s := &Service{
		retry:      1,
		timeout:    5,
		queryLimit: 50,
	}

	f := bytes.NewBufferString(queries)
	s.bank = sparql.LoadBank(f)

	// apply options
	for _, option := range options {
		if err := option(s); err != nil {
			return nil, err
		}
	}

	return s, nil
}

// mergeBank overrides queries on the service query bank when they have the same key.
func (s *Service) mergeBank(customBank sparql.Bank) error {
	for k, v := range customBank {
		s.bank[k] = v
	}

	return nil
}

// implement sparql proxy implementation
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router := chi.NewRouter()
	s.Routes("", router)
	router.ServeHTTP(w, r)
}

// should connections be shutdown
func (s *Service) Shutdown(ctx context.Context) error {
	return nil
}

func (s *Service) SetServiceBuilder(b *domain.ServiceBuilder) {
	s.log = b.Logger.With().Str("svc", "lod").Logger()
	s.orgs = b.Orgs
}

var _ lod.Resolver = (*Service)(nil)

func (s *Service) Resolve(ctx context.Context, orgID domain.OrganizationID, subj rdf.Subject) (g *rdf.Graph, err error) {
	cfg, ok := s.orgs.RetrieveConfig(orgID.String())
	if !ok {
		return nil, domain.ErrOrgNotFound
	}

	_ = cfg.Config

	return nil, nil
}

func (s *Service) GetRepo(orgID domain.OrganizationID) (*Repo, error) {
	repo, ok := s.repos[orgID]
	if !ok {
		orgCfg, known := s.orgs.RetrieveConfig(orgID.String())
		if !known {
			return nil, domain.ErrOrgNotFound
		}

		if !orgCfg.SPARQL.Enabled {
			return nil, domain.ErrServiceNotEnabled
		}

		cfg := RepoConfig{
			Host:           orgCfg.SPARQL.Host,
			QueryPath:      orgCfg.SPARQL.QueryPath,
			UpdatePath:     orgCfg.SPARQL.UpdatePath,
			GraphStorePath: orgCfg.SPARQL.GraphStorePath,
			Bank:           &s.bank,
			Transport: struct {
				Retry    int
				Timeout  int
				UserName string "json:\"userName\""
				Password string "json:\"password\""
			}{
				Retry:    1,
				Timeout:  10,
				UserName: orgCfg.SPARQL.UserName,
				Password: orgCfg.SPARQL.Password,
			},
		}

		var err error
		repo, err = NewRepo(cfg)
		if err != nil {
			return nil, err
		}

		s.repos[orgID] = repo
	}

	return repo, nil
}
