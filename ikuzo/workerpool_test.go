// nolint:gocritic
package ikuzo

import (
	"context"
	"testing"

	"github.com/matryer/is"
)

func Test_newWorkerPool(t *testing.T) {
	is := is.New(t)
	wp := newWorkerPool(context.TODO())
	is.True(wp.wg != nil)
	is.True(wp.ctx.Err() == nil)
}
