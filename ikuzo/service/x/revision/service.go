package revision

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/sosedoff/gitkit"
)

type Service struct {
	base   string
	server *gitkit.Server
}

func NewService(path string) (*Service, error) {
	s := &Service{base: path}
	err := s.setupGitKit()

	return s, err
}

func (s *Service) NewRepo(organization, dataset string) (*Repo, error) {
	repo := &Repo{
		path: s.repoPath(organization, dataset),
	}

	return repo, nil
}

func (s *Service) repoPath(organization, dataset string) string {
	return filepath.Join(s.base, organization, dataset)
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.server.ServeHTTP(w, r)
}

type Repo struct {
	organization string
	dataset      string
	path         string
	r            *git.Repository
}

func (r *Repo) Create() error {
	repository, err := git.PlainInit(r.path, false)
	if err != nil {
		return fmt.Errorf("unable to create plain repo; %w", err)
	}

	r.r = repository

	return nil
}

func (s *Service) setupGitKit() error {
	service := gitkit.New(gitkit.Config{
		Dir: s.base,
		// AutoCreate: true,
		// Auth:       false,
		// AutoHooks:  true,
		// Hooks:      hooks,
	})

	// Configure git server. Will create git repos path if it does not exist.
	// If hooks are set, it will also update all repos with new version of hook scripts.
	if err := service.Setup(); err != nil {
		return err
	}

	s.server = service

	return nil
}

// ReadDir
// Open(fname, sha1)
// ReadFile()
// Write(io.Reader)
// GetRepo(organization, dataset string)
