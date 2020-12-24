package oaipmh

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/render"
	"github.com/kiivihal/goharvest/oai"
	"github.com/rs/zerolog/log"
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
	var task *HarvestTask
	for _, t := range s.tasks {
		if t.GetLastCheck().Add(t.CheckEvery).Before(time.Now()) {
			task = &t
		}
	}

	return task
}

func (s *Service) runHarvest(ctx context.Context, task *HarvestTask) error {
	g, gctx := errgroup.WithContext(ctx)
	_ = gctx
	g.Go(func() error {
		defer func() {
			if r := recover(); r != nil {
				log.Error().
					Str("identifier", task.Name).
					Err(fmt.Errorf("panic message : %v on url %s", r, task.Request.GetFullURL())).
					Msg("unable to run harvest task")
			}
		}()
		if task.GetLastCheck().IsZero() {
			task.SetUnixStartFrom()
		}
		task.Request.Harvest(func(response *oai.Response) {
			recordTime := task.CallbackFn(response)
			task.SetLastCheck(recordTime)
		})
		return nil
	})

	if err := g.Wait(); errors.Is(err, context.Canceled) {
		return err
	}

	return nil
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}

// HarvestNow starts all harvest task from the beginning.
func (s *Service) HarvestNow(w http.ResponseWriter, r *http.Request) {
	errs := make([]string, 0)
	taskNames := make([]string, 0)
	for _, task := range s.tasks {
		t := time.Time{}
		task.SetLastCheck(t)
		err := s.runHarvest(r.Context(), &task)
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}
		taskNames = append(taskNames, task.Name)
	}
	if len(errs) > 0 {
		log.Error().
			Err(fmt.Errorf("%s", strings.Join(errs, "\n"))).
			Msg("harvest-now could complete properly")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	msg := fmt.Sprintf("Tasks processed: %s", strings.Join(taskNames, ", "))
	render.Status(r, http.StatusAccepted)
	render.PlainText(w, r, msg)
}

func (s *Service) Shutdown(ctx context.Context) error {
	s.cancel()

	if err := s.group.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	return nil
}
