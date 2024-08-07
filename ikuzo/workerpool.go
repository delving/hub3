// Copyright 2020 Delving B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ikuzo

import (
	"context"
	"sync"

	"github.com/delving/hub3/ikuzo/domain"
)

// WorkerService is the interface for background processes.
// The service needs to gracefully shutdown all goroutines when the context is
// canceled.
type WorkerService interface {
	Start(ctx context.Context, wg *sync.WaitGroup)
	domain.Shutdown
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
