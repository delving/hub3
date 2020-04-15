package revision

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/matryer/is"
)

// nolint:gocritic
func TestRepo(t *testing.T) {
	is := is.New(t)
	dir, err := ioutil.TempDir("", "revision")
	is.NoErr(err)

	defer os.RemoveAll(dir)

	s, err := NewService(dir)
	is.NoErr(err)
	is.True(s != nil)

	org := "demo-org"
	ds := "demo-spec"

	is.True(strings.HasSuffix(s.repoPath(org, ds), "demo-org/demo-spec"))

	repo, err := s.NewRepo(org, ds)
	is.NoErr(err)

	is.True(strings.HasSuffix(repo.path, "demo-org/demo-spec"))

	_, err = os.Stat(repo.path)
	is.True(os.IsNotExist(err))

	err = repo.Create()
	is.NoErr(err)

	w, err := repo.r.Worktree()
	is.NoErr(err)

	fname := "example-git-file"
	fpath := filepath.Join(repo.path, fname)

	t.Logf("example filename: %s", fpath)
	err = ioutil.WriteFile(fpath, []byte("hello world!"), 0644)
	is.NoErr(err)

	hash, err := w.Add(fname)
	is.NoErr(err)
	is.True(hash.String() != "")

	status, err := w.Status()
	is.NoErr(err)

	t.Logf("hash string: %q", status)
	is.True(!status.IsClean())

	// is.True(false)
}
