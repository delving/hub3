package ikuzo

import (
	"context"
	"sync"
)

// WorkerService is the interface for background processes.
// The service needs to gracefully shutdown all goroutines when the context is
// canceled.
type WorkerService interface {
	Start(ctx context.Context, wg *sync.WaitGroup)
	ServiceCancellation
}

type workerPool struct {
	// ctx must be a cancelable context
	ctx context.Context
	// wg tracks the number of goroutines
	wg *sync.WaitGroup
}

func newWorkerPool(ctx context.Context) *workerPool {
	return &workerPool{
		ctx: ctx,
		wg:  &sync.WaitGroup{},
	}
}
