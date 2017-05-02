package limiter

import (
	"sync"
	"time"
)

// NewRateLimiter returns a rate limiter that will will emit from the C
// channel only 'n' times every 'rate' seconds.
func NewRateLimiter(n int, rate time.Duration) *rateLimiter {
	r := &rateLimiter{
		C:        make(chan bool),
		rate:     rate,
		n:        n,
		shutdown: make(chan bool),
	}
	r.wg.Add(1)
	go r.limiter()
	return r
}

type rateLimiter struct {
	C    chan bool
	rate time.Duration
	n    int

	shutdown chan bool
	wg       sync.WaitGroup
}

func (r *rateLimiter) Stop() {
	close(r.shutdown)
	r.wg.Wait()
	close(r.C)
}

func (r *rateLimiter) limiter() {
	defer r.wg.Done()
	ticker := time.NewTicker(r.rate)
	defer ticker.Stop()
	counter := 0
	for {
		select {
		case <-r.shutdown:
			return
		case <-ticker.C:
			counter = 0
		default:
			if counter < r.n {
				select {
				case r.C <- true:
					counter++
				case <-r.shutdown:
					return
				}
			}
		}
	}
}
