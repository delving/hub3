package revision

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"code.gitea.io/gitea/modules/git"
	"github.com/matryer/is"
)

// nolint:gocritic
func testNewRepo(s *Service, org, ds string) func(t *testing.T) {
	return func(t *testing.T) {
		is := is.New(t)
		tests := []struct {
			name      string
			repoExist bool
		}{
			{name: "first access", repoExist: false},
			{name: "second access", repoExist: true},
		}

		for _, tt := range tests {
			tt := tt

			t.Run(tt.name, func(t *testing.T) {
				// repo should not exist
				_, err := os.Stat(filepath.Join(s.base, org, ds))
				is.Equal(errors.Is(err, os.ErrNotExist), !tt.repoExist)

				// create repo if it does not exist
				repo, err := s.OpenRepository(org, ds)
				is.Equal(errors.Is(err, ErrRepositoryNotExists), !tt.repoExist)
				if !tt.repoExist {
					repo, err = s.InitRepository(org, ds)
					is.NoErr(err)
				}

				is.True(strings.HasSuffix(repo.path, filepath.Join(org, ds)))

				// .git repo should also be created
				gitInfo, err := os.Stat(filepath.Join(s.base, org, ds, ".git"))
				is.NoErr(err)
				is.True(gitInfo.IsDir())

				// repo should have a git repository
				is.True(repo.gr != nil)
			})
		}
	}
}

// nolint:gocritic
func TestService(t *testing.T) {
	is := is.New(t)
	dir, err := ioutil.TempDir("", "revision")
	is.NoErr(err)

	t.Logf("tmpDir: %s", dir)

	defer os.RemoveAll(dir)

	s, err := NewService(dir)
	is.NoErr(err)
	is.True(s != nil)

	var org, ds = "demo-org", "demo-spec"

	t.Run("NewRepo", testNewRepo(s, org, ds))

	repo, err := s.OpenRepository(org, ds)
	is.NoErr(err)

	status, err := repo.Status()
	is.NoErr(err)

	// repo should be empty when created
	empty, err := repo.gr.IsEmpty()
	is.NoErr(err)
	is.True(empty)
	is.Equal(status.IsClean(), true)

	//  add single flight
	_, err = repo.SingleFlight(".keep", strings.NewReader("hub3"), "add .keep file")
	is.NoErr(err)

	// read a committed file
	r, err := repo.Read(".keep", "")
	is.NoErr(err)

	content, err := ioutil.ReadAll(r)
	is.NoErr(err)
	is.Equal(content, []byte("hub3"))

	// commit a new file
	err = repo.Write("first.txt", strings.NewReader("first file"))
	is.NoErr(err)

	status, err = repo.Status()
	is.NoErr(err)
	is.True(!status.IsClean())
	t.Logf("status: %#v", status)
	is.True(status.IsUntracked("first.txt"))

	err = repo.Add("first.txt")
	is.NoErr(err)

	status, err = repo.Status()
	is.NoErr(err)
	is.True(!status.IsClean())
	is.True(!status.IsUntracked("first.txt"))

	commitHash, err := repo.Commit("added new file", nil)
	is.NoErr(err)
	is.True(!commitHash.IsZero())

	status, err = repo.Status()
	is.NoErr(err)
	is.True(status.IsClean())
	is.True(!status.IsUntracked("first.txt"))

	cfs, err := git.GetCommitFileStatus(repo.path, "HEAD")
	is.NoErr(err)
	is.True(len(cfs.Added) == 1)

	pastCommit, err := repo.gr.GetCommitByPath(".keep")
	is.NoErr(err)

	head, err := repo.gr.GetCommit("HEAD")
	is.NoErr(err)

	files, err := head.GetFilesChangedSinceCommit(pastCommit.ID.String())
	is.NoErr(err)

	commits, err := repo.gr.FileCommitsCount("HEAD", ".keep")
	is.NoErr(err)
	is.True(commits == 1)

	t.Logf("files: %#v", files)

	is.True(len(files) == 2)
}
