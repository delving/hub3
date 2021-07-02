package oaipmh

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

type ServerOption func(*Server) error

type RepoStore interface {
	ListRepos(ctx context.Context) ([]Repo, error)
	GetRepo(ctx context.Context, repo string) (Repo, error)
}

type Server struct {
	repos       RepoStore
	routePrefix string
}

type store struct {
	path string
}

func NewFsRepoStore(path string) (RepoStore, error) {
	s := store{path: path}
	if err := os.MkdirAll(s.path, os.ModePerm); err != nil {
		return s, fmt.Errorf("unable to create OAI-PMH fs store; %w", err)
	}

	return s, nil
}

func (s store) GetRepo(ctx context.Context, repo string) (Repo, error) {
	return Repo{}, fmt.Errorf("implement me")
}

func (s store) ListRepos(ctx context.Context) ([]Repo, error) {
	var repos []Repo

	files, err := ioutil.ReadDir(s.path)
	if err != nil {
		return repos, err
	}

	for _, f := range files {
		if _, err := os.Stat(filepath.Join(s.path, f.Name(), "identify.json")); os.IsNotExist(err) {
			continue
		}

		repos = append(repos, Repo{Name: f.Name()})
	}

	return repos, nil
}

type Repo struct {
	Name  string
	Links map[string]string
}

func NewServer(options ...ServerOption) (*Server, error) {
	s := &Server{}

	// apply options
	for _, option := range options {
		if err := option(s); err != nil {
			return nil, err
		}
	}

	if s.repos == nil {
		return s, fmt.Errorf("empty ServerStore is not allowed")
	}

	return s, nil
}

func (s *Server) handleListRepos() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		basePath := r.URL.Path

		repos, err := s.repos.ListRepos(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp := make([]Repo, len(repos))

		for idx, repo := range repos {
			repo.Links = make(map[string]string)
			repo.Links["oai-pmh"] = filepath.Join(basePath, repo.Name)
			resp[idx] = repo
		}

		render.JSON(w, r, resp)
	}
}

func (s *Server) handleOaiPmhRequest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// repo := chi.URLParam(r, "repo")

		req := NewRequest(r)
		render.JSON(w, r, req)
	}
}

func (s *Server) Routes(routePrefix string) func(router chi.Router) {
	if routePrefix == "" {
		routePrefix = "/"
	}

	return func(router chi.Router) {
		router.Get(routePrefix, s.handleListRepos())

		if !strings.HasSuffix(routePrefix, "/") {
			routePrefix += "/"
		}

		router.Get(routePrefix+"{repo}", s.handleOaiPmhRequest())
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) http.HandlerFunc {
	// closure to setup chi.Router
	router := chi.NewRouter()
	s.Routes("")(router)

	return func(w http.ResponseWriter, r *http.Request) {
		router.ServeHTTP(w, r)
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	return nil
}

func SetServerStore(store RepoStore) ServerOption {
	return func(s *Server) error {
		s.repos = store
		return nil
	}
}
