package index

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
)

type shaRef struct {
	HubID string
	Sha   string
}

type batchWriter struct {
	refs      []shaRef
	s         *Service
	batchSize int
	ticker    *time.Ticker
	rw        sync.Mutex
}

func newBatchWriter(s *Service, batchSize int) batchWriter {
	ticker := time.NewTicker(5 * time.Second)
	return batchWriter{
		refs:      []shaRef{},
		batchSize: batchSize,
		s:         s,
		ticker:    ticker,
	}
}

func (b *batchWriter) flush() error {
	b.rw.Lock()
	defer b.rw.Unlock()

	if len(b.refs) == 0 {
		// nothing to do
		return nil
	}

	atomic.AddUint64(&b.s.m.Index.StoreRef, uint64(len(b.refs)))
	atomic.AddUint64(&b.s.m.Index.FlushRef, 1)

	if err := b.s.store.Put(b.refs...); err != nil {
		log.Error().Err(err).Msg("unable to store shaRefs")
	}

	b.refs = []shaRef{}

	return nil
}

func (b *batchWriter) run(ctx context.Context) error {
	for ref := range b.s.queue {
		if len(b.refs) >= b.batchSize {
			if err := b.flush(); err != nil {
				log.Error().Err(err).Msg("unable to flush shaRefs")
			}
		}

		b.rw.Lock()
		b.refs = append(b.refs, ref)
		b.rw.Unlock()

		select {
		case <-ctx.Done():
			if err := b.flush(); err != nil {
				log.Error().Err(err).Msg("unable to flush shaRefs")
			}

			return ctx.Err()
		default:
		}
	}

	return nil
}
