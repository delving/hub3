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
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"code.gitea.io/gitea/modules/git"
	gitgo "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type Repository struct {
	organization string
	dataset      string
	path         string
	gr           *git.Repository
	w            *gitgo.Worktree
}

// SingleFlight writes io.Reader to path and creates a commit with commitMessage.
func (repo *Repository) SingleFlight(path string, r io.Reader, commitMessage string) (plumbing.Hash, error) {
	if err := repo.Write(path, r); err != nil {
		return plumbing.ZeroHash, err
	}

	if addErr := repo.Add(path); addErr != nil {
		return plumbing.ZeroHash, fmt.Errorf("unable to add file to staging; %w", addErr)
	}

	commit, commitErr := repo.Commit(commitMessage, nil)
	if commitErr != nil {
		return plumbing.ZeroHash, commitErr
	}

	return commit, nil
}

// Write writes the content of io.Reader to a file at path.
// When the file does not exist a new file is created.
//
// An error is only returned when creating or write to the file fails.
func (repo *Repository) Write(path string, r io.Reader) error {
	fPath := filepath.Join(repo.path, path)
	if err := os.MkdirAll(filepath.Dir(fPath), os.ModePerm); err != nil {
		return fmt.Errorf("unable to create directories; %w", err)
	}

	f, err := os.Create(fPath)
	if err != nil {
		return fmt.Errorf("unable to create file; %w", err)
	}

	_, err = io.Copy(f, r)
	if err != nil {
		return fmt.Errorf("unable to write to file; %w", err)
	}

	return nil
}

// Read returns a Reader for the given path for a specific revision.
// When the revision is empty the HEAD version in returned.
func (repo *Repository) Read(path, revision string) (io.ReadCloser, error) {
	if revision == "" || strings.EqualFold(revision, "head") {
		revision = "HEAD"
	}

	tree, err := repo.gr.GetTree(revision)
	if err != nil {
		return nil, err
	}

	blob, err := tree.GetBlobByPath(path)
	if err != nil {
		return nil, err
	}

	return blob.DataAsync()
}

func (repo *Repository) Commit(msg string, options *gitgo.CommitOptions) (plumbing.Hash, error) {
	w, err := repo.workTree()
	if err != nil {
		return plumbing.ZeroHash, err
	}

	if options == nil {
		options = &gitgo.CommitOptions{
			Author: &object.Signature{
				Name: "hub3",
				When: time.Now(),
			},
		}
	}

	return w.Commit(msg, options)
}

// Add adds all files with path to the staging area.
func (repo *Repository) Add(path string) error {
	if path == "" {
		path = "."
	}

	return git.AddChanges(repo.path, false, path)
}

func (repo *Repository) workTree() (*gitgo.Worktree, error) {
	if repo.w == nil {
		r, err := gitgo.PlainOpen(repo.path)
		if err != nil {
			return nil, err
		}

		repo.w, err = r.Worktree()
		if err != nil {
			return nil, err
		}
	}

	return repo.w, nil
}

func (repo *Repository) Status() (gitgo.Status, error) {
	w, err := repo.workTree()
	if err != nil {
		return nil, err
	}

	return w.Status()
}
