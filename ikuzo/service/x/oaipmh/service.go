package oaipmh

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

type Service struct {
	cancel       context.CancelFunc
	group        *errgroup.Group
	workers      int
	rw           sync.Mutex
	defaultDelay int
	tasks        []HarvestTask
}

func NewService(options ...Option) (*Service, error) {
	s := &Service{}

	// apply options
	for _, option := range options {
		if err := option(s); err != nil {
			return nil, err
		}
	}

	if s.workers == 0 {
		s.workers = 1
	}

	if s.defaultDelay == 0 {
		s.defaultDelay = 1
	}

	return s, nil
}

func (s *Service) StartHarvestSync() error {
	// create errgroup and add cancel to service
	ctx, cancel := context.WithCancel(context.Background())
	g, gctx := errgroup.WithContext(ctx)
	_ = gctx

	s.cancel = cancel
	s.group = g

	ticker := time.NewTicker(time.Duration(s.defaultDelay) * time.Minute)

	for i := 0; i < s.workers; i++ {
		g.Go(func() error {
			for {
				select {
				case <-gctx.Done():
					return gctx.Err()
				case <-ticker.C:
					s.rw.Lock()
					task := s.findAvailableTask()
					if task == nil {
						s.rw.Unlock()
						continue
					}
					s.rw.Unlock()

					if err := s.runHarvest(gctx, task); err != nil {
						return err
					}
				}
			}
		})
	}

	return nil
}

func (s *Service) findAvailableTask() *HarvestTask {
	var task HarvestTask
	// query task that are
	// TODO(kiivihal): implement

	return &task
}

func (s *Service) runHarvest(ctx context.Context, task *HarvestTask) error {
	// handle the context cancellation on shutdown
	// TODO(kiivihal): implement me
	return nil
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

func (s *Service) Shutdown(ctx context.Context) error {
	s.cancel()

	if err := s.group.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	return nil
}
