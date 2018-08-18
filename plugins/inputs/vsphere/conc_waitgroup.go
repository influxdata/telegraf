package vsphere

import (
	"sync"
)

// ConcurrentWaitGroup is a WaitGroup with special semantics. While a standard wait group
// requires Add() and Wait() to be called from the same goroutine, a ConcurrentWaitGroup
// allows Add() and Wait() to be called concurrently.
// By allowing this, we can no longer tell whether the jobs we're monitoring are truly
// finished or whether the job count temporarily dropped to zero. To prevent this, we
// add the requirement that no more jobs can be started once the counter has reached zero.
// This is done by returning a flag from Add() that, when set to false, means that
// a Wait() has become unblocked and that the caller must abort its attempt to run a job.
type ConcurrentWaitGroup struct {
	mux  sync.Mutex
	cond *sync.Cond
	jobs int
	done bool
}

// NewConcurrentWaitGroup returns a new NewConcurrentWaitGroup.
func NewConcurrentWaitGroup() *ConcurrentWaitGroup {
	c := &ConcurrentWaitGroup{}
	c.cond = sync.NewCond(&c.mux)
	return c
}

// Add signals the beginning of one or more jobs. The function returns false
// if a Wait() has already been unblocked and callers should not run the job.
func (c *ConcurrentWaitGroup) Add(inc int) bool {
	c.mux.Lock()
	defer c.mux.Unlock()

	if c.done {
		return false
	}
	c.jobs += inc
	if c.jobs == 0 {
		c.cond.Broadcast()
	}
	return true
}

// Done signals that a job is done. Once the number of running jobs reaches
// zero, any blocked calls to Wait() will unblock.
func (c *ConcurrentWaitGroup) Done() {
	c.Add(-1)
}

// Wait blocks until the number of running jobs reaches zero.
func (c *ConcurrentWaitGroup) Wait() {
	c.mux.Lock()
	defer c.mux.Unlock()
	for c.jobs != 0 {
		c.cond.Wait()
	}
	c.done = true
}
