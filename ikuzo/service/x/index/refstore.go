package index

import (
	"go.etcd.io/bbolt"
)

var refBucket = []byte("index-refs")

type store struct {
	db *bbolt.DB
}

type shaRef struct {
	HubID string
	Sha   string
}

func newStore() (*store, error) {
	db, err := bbolt.Open("index.db", 0600, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists(refBucket)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &store{db: db}, nil
}

func (s *store) Delete(hubID string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(refBucket)
		return b.Delete([]byte(hubID))
	})
}

func (s *store) Get(hubID string) (sha string, err error) {
	var v []byte

	err = s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(refBucket)
		v = b.Get([]byte(hubID))
		return nil
	})
	if err != nil {
		return "", err
	}

	if v == nil {
		return "", nil
	}

	return string(v), nil
}

func (s *store) HashIsEqual(hubID, sha string) (bool, error) {
	storedSha, err := s.Get(hubID)
	if err != nil {
		return false, err
	}

	if storedSha != sha {
		return false, nil
	}

	return true, nil
}

func (s *store) Put(refs ...shaRef) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(refBucket)

		for _, ref := range refs {
			// TODO(kiivihal): implement retry later
			if err := b.Put([]byte(ref.HubID), []byte(ref.Sha)); err != nil {
				return err
			}
		}

		return nil
	})
}
