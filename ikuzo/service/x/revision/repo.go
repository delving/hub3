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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"code.gitea.io/gitea/modules/git"
	"github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/ikuzo/domain/domainpb"
	gitgo "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

const (
	WorkingVersion = "working-version"
	headVersion    = "HEAD"
)

type Repository struct {
	OrgID     string
	DatasetID string
	path      string
	gr        *git.Repository
	w         *gitgo.Worktree
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
		return plumbing.ZeroHash, fmt.Errorf("unable to commit; %w", commitErr)
	}

	return commit, nil
}

// Write writes the content of io.Reader to a file at path.
// When the file does not exist a new file is created.
//
// You must repo.Add with the path before it can be committed.
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

	defer f.Close()

	_, err = io.Copy(f, r)
	if err != nil {
		return fmt.Errorf("unable to write to file; %w", err)
	}

	return nil
}

// Read returns a Reader for the given path for a specific revision.
// When the revision is empty the HEAD version in returned.
func (repo *Repository) Read(path, revision string) (io.ReadCloser, error) {
	if revision == "" || strings.EqualFold(revision, headVersion) {
		revision = headVersion
	}

	if revision == WorkingVersion {
		fPath := filepath.Join(repo.path, path)

		f, err := os.Open(fPath)
		if err != nil {
			return nil, err
		}

		return f, nil
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

func (repo *Repository) Publish(msg string) (PublisherStats, error) {
	stats := PublisherStats{}
	// if at any state you fail return and reset
	// add resources
	// check status, return if nothing has changed with HEAD
	// commit repo and get current sha
	// get dataset repo
	// get dataset file
	// update repo hash in dataset file
	// update dataset repo files
	// commit dataset
	// add dataset repo sha to stats
	return stats, nil
}

func (repo *Repository) Commit(msg string, options *gitgo.CommitOptions) (plumbing.Hash, error) {
	w, err := repo.workTree()
	if err != nil {
		return plumbing.ZeroHash, err
	}

	if repo.IsClean() {
		h, err := repo.HEAD()
		if err != nil {
			return plumbing.ZeroHash, err
		}

		return h, nil
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

	cmd := git.NewCommand(context.Background(), "add")
	err := cmd.AddArguments("--").AddArguments(path).Run(
		&git.RunOpts{Dir: repo.path},
	)
	if err != nil {
		return fmt.Errorf("add error: %w", err)
	}

	return nil
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

// ResetPath removes all entries for a directory path.
//
// This functionality allows full reingests to mark deleted entries.
func (repo *Repository) ResetPath(path string) error {
	fPath := filepath.Join(repo.path, path)
	if removeErr := os.RemoveAll(fPath); removeErr != nil {
		return fmt.Errorf("unable to remove trs path; %w", removeErr)
	}

	if createErr := os.MkdirAll(fPath, os.ModePerm); createErr != nil {
		return fmt.Errorf("unable to create trs path; %w", createErr)
	}

	return nil
}

func (repo *Repository) HEAD() (plumbing.Hash, error) {
	cmd := git.NewCommand(context.Background(), "rev-parse")

	sha, _, err := cmd.AddArguments(headVersion).RunStdString(
		&git.RunOpts{Dir: repo.path},
	)
	if err != nil {
		return plumbing.ZeroHash, err
	}

	return plumbing.NewHash(sha), nil
}

func (repo *Repository) IsClean() bool {
	cmd := git.NewCommand(context.Background(), "status")

	status, _, err := cmd.AddArguments("--porcelain").RunStdString(
		&git.RunOpts{Dir: repo.path},
	)
	if err != nil {
		// TODO(kiivihal): add logger for error
		return false
	}

	return status == ""
}

func (repo *Repository) diff(path, from, until string) (string, error) {
	if from == "" {
		revisions, err := repo.revisions()
		if err != nil {
			return "", err
		}

		if revisions > 1 {
			from = headVersion + "^"
		}
	}

	if until == "" {
		until = headVersion
	}

	cmd := git.NewCommand(context.Background(), "log").
		AddArguments("--reverse").
		AddArguments("--name-status").
		AddArguments("--no-renames").
		AddArguments("--no-decorate").
		AddArguments("--no-merges").
		AddArguments("--pretty=format:\"%H %ci\"")

	if from != "" {
		until = fmt.Sprintf("%s...%s", from, until)
	}

	cmd = cmd.AddArguments(until)

	if path != "" {
		cmd = cmd.AddArguments("--").AddArguments(path)
	}

	// TODO(kiivihal): remove this statement
	log.Printf("diff command: %s", cmd.String())

	resp, _, err := cmd.RunStdString(&git.RunOpts{Dir: repo.path})

	return resp, err
}

// revisions returns the number of committed revisions on the current branch
func (repo *Repository) revisions() (int, error) {
	resp, _, err := git.NewCommand(context.Background(), "rev-list").
		AddArguments("--count").
		AddArguments(headVersion).
		RunStdString(&git.RunOpts{Dir: repo.path})
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(strings.TrimSpace(resp))
}

func (repo *Repository) Changes(path, from, until string) (chan DiffFile, error) {
	c := make(chan DiffFile, 2500)

	output, err := repo.diff(path, from, until)
	if err != nil {
		return c, err
	}

	go func(lines string) {
		defer close(c)

		p := newLogParser()
		if err := p.generate(lines, c); err != nil {
			log.Printf("unable to parse diff files; %v", err)
		}
	}(output)

	return c, nil
}

func (repo *Repository) Exists(path string) bool {
	fPath := filepath.Join(repo.path, path)
	_, err := os.Stat(fPath)

	return !os.IsNotExist(err)
}

type Publisher interface {
	Publish(ctx context.Context, messages ...*domainpb.IndexMessage) error
}

type PublishStats struct {
	Deleted int
	Updated int
}

func (repo *Repository) PublishChanged(from, until string, p ...Publisher) (PublishStats, error) {
	resourcePath := "rsc"
	stats := PublishStats{}

	files, err := repo.Changes(resourcePath, from, until)
	if err != nil {
		return stats, err
	}

	for f := range files {
		f := f

		hubID := strings.TrimSuffix(strings.TrimPrefix(f.Path, resourcePath+"/"), ".json")

		m := &domainpb.IndexMessage{
			OrganisationID: repo.OrgID,
			DatasetID:      repo.DatasetID,
			RecordID:       hubID,
			IndexName:      config.Config.ElasticSearch.GetIndexName(repo.OrgID),
		}
		if f.State == StatusDeleted {
			m.Deleted = true
			stats.Deleted++
		}

		if !m.Deleted {
			r, err := repo.Read(f.Path, f.CommitID)
			if err != nil {
				return stats, err
			}

			var fg fragments.FragmentGraph
			if decodeErr := json.NewDecoder(r).Decode(&fg); decodeErr != nil {
				return stats, decodeErr
			}

			fg.Meta.Modified = fragments.NowInMillis()
			fg.Meta.SourceID = f.CommitID

			b, err := fg.Marshal()
			if err != nil {
				return stats, err
			}

			m.Source = b

			stats.Updated++
		}

		for _, publisher := range p {
			if err := publisher.Publish(context.Background(), m); err != nil {
				return stats, err
			}
		}
	}

	return stats, nil
}
