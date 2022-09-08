package task

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/storage/x/redis"
	"github.com/go-chi/chi"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

var (
	_             domain.Service = (*Service)(nil)
	defaultQueues                = map[string]int{
		"critical": 6,
		"default":  3,
		"low":      1,
	}
)

const (
	defaultWorkers = 10
)

type ScheduleTask struct{}

type Service struct {
	orgs      domain.OrgConfigRetriever
	log       zerolog.Logger
	redisCfg  *redis.Config
	client    *asynq.Client
	server    *asynq.Server
	mux       *asynq.ServeMux
	nrWorkers int
	queues    map[string]int
	scheduler *asynq.Scheduler
}

func NewService(options ...Option) (*Service, error) {
	s := &Service{
		nrWorkers: defaultWorkers,
		queues:    defaultQueues,
	}

	// apply options
	for _, option := range options {
		if err := option(s); err != nil {
			return nil, err
		}
	}

	if s.redisCfg == nil {
		return nil, fmt.Errorf("redis.Config must be set to use task service")
	}

	s.client = s.asynqClient()
	s.scheduler = asynq.NewScheduler(
		s.redisClientOpt(),
		&asynq.SchedulerOpts{EnqueueErrorHandler: s.errorHandler},
	)
	s.server = s.asynqServer()
	s.mux = asynq.NewServeMux()

	// schedule health ping
	health := health{taskName: "health:ping"}
	if err := health.scheduleTask(s.scheduler); err != nil {
		return s, err
	}
	s.RegisterWorkerFunc(health.taskName, health.handleTask)

	return s, nil
}

func (s *Service) RegisterWorkerFunc(pattern string, handler func(context.Context, *asynq.Task) error) {
	s.mux.HandleFunc(pattern, handler)
}

func (s *Service) ScheduleTask(cronspec string, task *asynq.Task, opts ...asynq.Option) (entryID string, err error) {
	return s.scheduler.Register(cronspec, task, opts...)
}

func (s *Service) EnqueueTask(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	return s.client.Enqueue(task, opts...)
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router := chi.NewRouter()
	s.Routes("", router)
	router.ServeHTTP(w, r)
}

// StartWorkers is a blocking function that starts the background workers
// and scheduler.
func (s *Service) StartWorkers(ctx context.Context) error {
	g, _ := errgroup.WithContext(ctx)

	s.log.Info().Msg("starting asynq scheduler")
	g.Go(func() error { return s.scheduler.Run() })
	s.log.Info().Msg("starting asynq server")
	g.Go(func() error { return s.server.Run(s.mux) })

	s.log.Info().Msg("asynq workers are listening")

	if err := g.Wait(); err != nil {
		return fmt.Errorf("unable to start asynq workers; %w", err)
	}

	return nil
}

func (s *Service) Shutdown(ctx context.Context) error {
	s.log.Info().Msg("stopping asynq service")
	if s.client != nil {
		if err := s.client.Close(); err != nil {
			return err
		}
	}

	if s.server != nil {
		s.server.Stop()
		s.server.Shutdown()
	}

	if s.scheduler != nil {
		s.scheduler.Shutdown()
	}

	return nil
}

func (s *Service) SetServiceBuilder(b *domain.ServiceBuilder) {
	s.log = b.Logger.With().Str("svc", "task").Logger()
	s.orgs = b.Orgs
}

func (s *Service) errorHandler(task *asynq.Task, opts []asynq.Option, err error) {
	if errors.Is(err, asynq.ErrDuplicateTask) || errors.Is(err, asynq.ErrTaskIDConflict) {
		return
	}

	s.log.Warn().Msgf("unable to enqueue scheduled task: %#v", task)
}
