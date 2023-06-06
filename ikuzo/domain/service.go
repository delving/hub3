package domain

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hibiken/asynq"

	"github.com/delving/hub3/ikuzo/logger"
)

var ErrServiceNotEnabled = fmt.Errorf("service not enabled for this organization")

// Service defines minimal API service of an ikuzo.Service
type Service interface {
	// Metrics() interface{}
	http.Handler
	Router
	SetServiceBuilder(b *ServiceBuilder)
	Shutdown
}

// Shutdown must be implement by each service that uses background services or connections.
type Shutdown interface {
	Shutdown(ctx context.Context) error
}

// Router implements a callback to register routes to a chi.Router
// If pattern is non-empty this mount point will be used, instead of the
// default specified by the domain.Service implementation
type Router interface {
	Routes(pattern string, router chi.Router)
}

type ServiceBuilder struct {
	Logger      *logger.CustomLogger
	Orgs        OrgConfigRetriever
	TaskService TaskService
}

type TaskService interface {
	RegisterWorkerFunc(pattern string, handler func(context.Context, *asynq.Task) error)
	ScheduleTask(cronspec string, task *asynq.Task, opts ...asynq.Option) (entryID string, err error)
	EnqueueTask(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error)
}
