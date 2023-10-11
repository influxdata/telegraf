package prometheus_http

import "sync/atomic"

// A Counter is a thread-safe counter implementation
type Counter int64

// Incr method increments the counter by some value
func (c *Counter) Incr(val int64) {
	atomic.AddInt64((*int64)(c), val)
}

// Reset method resets the counter's value to zero
func (c *Counter) Reset() {
	atomic.StoreInt64((*int64)(c), 0)
}

// Value method returns the counter's current value
func (c *Counter) Value() int64 {
	return atomic.LoadInt64((*int64)(c))
}
