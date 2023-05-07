package harvest

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

var _ domain.Service = (*Service)(nil)

type Service struct {
	ctx          context.Context
	cancel       context.CancelFunc
	group        *errgroup.Group
	workers      int
	rw           sync.Mutex
	defaultDelay int
	tasks        []*HarvestTask
	log          zerolog.Logger
	orgs         domain.OrgConfigRetriever
}

func NewService(options ...Option) (*Service, error) {
	s := &Service{
		tasks: []*HarvestTask{},
	}

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

func (s *Service) AddTask(task ...*HarvestTask) error {
	s.rw.Lock()
	defer s.rw.Unlock()

	if len(task) > 0 {
		s.tasks = append(s.tasks, task...)
	}

	return nil
}

func (s *Service) StartHarvestSync() error {
	// create errgroup and add cancel to service
	ctx, cancel := context.WithCancel(context.Background())
	g, gctx := errgroup.WithContext(ctx)

	s.cancel = cancel
	s.group = g
	s.ctx = ctx

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
	for _, t := range s.tasks {
		if t.running {
			continue
		}

		if t.GetLastCheck().Add(t.CheckEvery).Before(time.Now()) {
			return t
		}
	}

	return nil
}

func (s *Service) runHarvest(ctx context.Context, task *HarvestTask) error {
	g, _ := errgroup.WithContext(ctx)

	g.Go(func() error {
		defer func() {
			if r := recover(); r != nil {
				log.Error().
					Str("identifier", task.Name).
					Err(fmt.Errorf("panic message : %v on url %s", r, task.Request.GetFullURL())).
					Msg("unable to run harvest task")
			}
		}()

		err := task.Harvest(ctx)
		if err != nil {
			return err
		}

		return nil
	})

	if err := g.Wait(); errors.Is(err, context.Canceled) {
		return err
	}

	return nil
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router := chi.NewRouter()
	s.Routes("", router)
	router.ServeHTTP(w, r)
}

func (s *Service) SetServiceBuilder(b *domain.ServiceBuilder) {
	s.log = b.Logger.With().Str("svc", "sitemap").Logger()
	s.orgs = b.Orgs
}

// HarvestNow starts all harvest task from the beginning.
func (s *Service) HarvestNow(w http.ResponseWriter, r *http.Request) {
	resync := r.URL.Query().Get("resync")
	errs := make([]string, 0)
	taskNames := make([]string, 0)

	for _, task := range s.tasks {
		if strings.EqualFold(resync, "true") {
			err := task.getOrCreateHarvestInfo()
			if err != nil {
				errs = append(errs, err.Error())
				continue
			}

			task.HarvestInfo.LastModified = time.Time{}
			if err := task.writeHarvestInfo(); err != nil {
				errs = append(errs, err.Error())
				continue
			}
		}

		err := s.runHarvest(s.ctx, task)
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

	render.Status(r, http.StatusAccepted)

	msg := fmt.Sprintf("Tasks processed: %s", strings.Join(taskNames, ", "))
	render.PlainText(w, r, msg)
}

func (s *Service) Shutdown(ctx context.Context) error {
	s.cancel()

	if err := s.group.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	return nil
}
