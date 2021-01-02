package index

import (
	"os"
	"testing"

	"github.com/matryer/is"
)

// nolint:gocritic
func TestIndexStore(t *testing.T) {
	defer func() {
		e := os.Remove("index.db")
		if e != nil {
			t.Fatal(e)
		}
	}()

	t.Run("store crud", func(t *testing.T) {
		is := is.New(t)
		s, err := newStore()
		is.NoErr(err)

		hubID, sha := "org_spec_123", "0c5e290190fef0b2933"
		err = s.Put(hubID, sha)
		is.NoErr(err)

		storedSha, err := s.Get("org_spec_123")
		is.NoErr(err)
		is.Equal(sha, storedSha) // sha should not be empty

		ok, err := s.HashIsEqual(hubID, sha)
		is.NoErr(err)
		is.True(ok)

		ok, err = s.HashIsEqual(hubID, "unequal sha")
		is.NoErr(err)
		is.True(!ok)

		err = s.Delete(hubID)
		is.NoErr(err)

		missingSha, err := s.Get("org_spec_123")
		is.NoErr(err)
		is.Equal(missingSha, "") // sha should be empty

		s.db.Close()
	})
}
