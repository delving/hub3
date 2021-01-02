package index

import (
	"go.etcd.io/bbolt"
)

type store struct {
	db *bbolt.DB
}

func newStore() (*store, error) {
	db, err := bbolt.Open("index.db", 0600, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists([]byte("index-refs"))
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
		b := tx.Bucket([]byte("index-refs"))
		return b.Delete([]byte(hubID))
	})
}

func (s *store) Get(hubID string) (sha string, err error) {
	var v []byte

	err = s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("index-refs"))
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

func (s *store) Batch(hubID, sha string) error {
	return s.db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("index-refs"))
		return b.Put([]byte(hubID), []byte(sha))
	})
}

func (s *store) Put(hubID, sha string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("index-refs"))
		return b.Put([]byte(hubID), []byte(sha))
	})
}
