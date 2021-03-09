package index

import (
	"sync"
	"time"
)

type tx struct {
	orgID              string
	datasetID          string
	previousCountIndex uint64
	previousDateTime   time.Time
	seen               uint64
	endCountIndex      uint64
	endDateTime        time.Time
	revision           int // old revision key
	rw                 sync.RWMutex
	recordIDs          []string
}

func newTx(orgID, datasetID string) *tx {
	return nil
}

func getOrCreateTx(orgID, datasetID string) (*tx, error) {
	// increment revision

	return nil, nil
}
