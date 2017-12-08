package capnp

import "sync/atomic"

// A ReadLimiter tracks the number of bytes read from a message in order
// to avoid amplification attacks as detailed in
// https://capnproto.org/encoding.html#amplification-attack.
// It is safe to use from multiple goroutines.
type ReadLimiter struct {
	limit uint64
}

// canRead reports whether the amount of bytes can be stored safely.
func (rl *ReadLimiter) canRead(sz Size) bool {
	for {
		curr := atomic.LoadUint64(&rl.limit)
		ok := curr >= uint64(sz)
		var new uint64
		if ok {
			new = curr - uint64(sz)
		} else {
			new = 0
		}
		if atomic.CompareAndSwapUint64(&rl.limit, curr, new) {
			return ok
		}
	}
}

// Reset sets the number of bytes allowed to be read.
func (rl *ReadLimiter) Reset(limit uint64) {
	atomic.StoreUint64(&rl.limit, limit)
}

// Unread increases the limit by sz.
func (rl *ReadLimiter) Unread(sz Size) {
	atomic.AddUint64(&rl.limit, uint64(sz))
}
