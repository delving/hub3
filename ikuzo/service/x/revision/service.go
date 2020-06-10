// Copyright 2020 Delving B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package revision

import (
	"errors"
	"net/http"
	"path/filepath"
	"strings"

	"code.gitea.io/gitea/modules/git"
	gitgo "github.com/go-git/go-git/v5"
	"github.com/sosedoff/gitkit"
)

var (
	ErrRepositoryNotExists = errors.New("repository does not exist")
)

type Service struct {
	base     string
	server   *gitkit.Server
	BareRepo bool
}

func NewService(path string) (*Service, error) {
	s := &Service{base: path}
	if strings.HasSuffix(s.base, "/") {
		s.base = strings.TrimSuffix(s.base, "/")
	}

	err := s.setupGitKit()

	return s, err
}

// InitRepository initializes a Repository and returns it.
//
// An error is only returned if there are underlying FS errors.
func (s *Service) InitRepository(organization, dataset string) (*Repository, error) {
	if err := git.InitRepository(s.repoPath(organization, dataset), false); err != nil {
		return nil, err
	}

	return s.OpenRepository(organization, dataset)
}

// OpenRepository returns an *Repository. When the Repository is not initialized
// or does not exist a ErrRepositoryNotExists is returned.
//
// To create a repository you need to call InitRepository.
func (s *Service) OpenRepository(organization, dataset string) (*Repository, error) {
	repo := &Repository{
		path:         s.repoPath(organization, dataset),
		organization: organization,
		dataset:      dataset,
	}

	gr, err := git.OpenRepository(repo.path)
	if err != nil {
		if err.Error() == "no such file or directory" || errors.Is(err, gitgo.ErrRepositoryNotExists) {
			return nil, ErrRepositoryNotExists
		}

		return nil, err
	}

	repo.gr = gr

	return repo, nil
}

func (s *Service) repoPath(organization, dataset string) string {
	return filepath.Join(s.base, organization, dataset)
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.server.ServeHTTP(w, r)
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
