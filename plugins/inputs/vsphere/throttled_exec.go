package vsphere

import "sync"

// ThrottledExecutor provides a simple mechanism for running jobs in separate
// goroutines while limit the number of concurrent jobs running at any given time.
type ThrottledExecutor struct {
	limiter chan struct{}
	wg      sync.WaitGroup
}

// NewThrottledExecutor creates a new ThrottlesExecutor with a specified maximum
// number of concurrent jobs
func NewThrottledExecutor(limit int) *ThrottledExecutor {
	if limit == 0 {
		panic("Limit must be > 0")
	}
	return &ThrottledExecutor{limiter: make(chan struct{}, limit)}
}

// Run schedules a job for execution as soon as possible while respecting the
// maximum concurrency limit.
func (t *ThrottledExecutor) Run(job func()) {
	t.wg.Add(1)
	t.limiter <- struct{}{}
	go func() {
		defer func() {
			<-t.limiter
		}()
		defer t.wg.Done()
		job()
	}()
}

// Wait blocks until all scheduled jobs have finished
func (t *ThrottledExecutor) Wait() {
	t.wg.Wait()
}
