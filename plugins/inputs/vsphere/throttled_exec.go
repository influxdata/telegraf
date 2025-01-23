package vsphere

import (
	"context"
	"sync"
)

// throttledExecutor provides a simple mechanism for running jobs in separate
// goroutines while limit the number of concurrent jobs running at any given time.
type throttledExecutor struct {
	limiter chan struct{}
	wg      sync.WaitGroup
}

// newThrottledExecutor creates a new ThrottlesExecutor with a specified maximum
// number of concurrent jobs
func newThrottledExecutor(limit int) *throttledExecutor {
	if limit == 0 {
		panic("Limit must be > 0")
	}
	return &throttledExecutor{limiter: make(chan struct{}, limit)}
}

// run schedules a job for execution as soon as possible while respecting the maximum concurrency limit.
func (t *throttledExecutor) run(ctx context.Context, job func()) {
	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		select {
		case t.limiter <- struct{}{}:
			defer func() {
				<-t.limiter
			}()
			job()
		case <-ctx.Done():
			return
		}
	}()
}

// wait blocks until all scheduled jobs have finished
func (t *throttledExecutor) wait() {
	t.wg.Wait()
}
