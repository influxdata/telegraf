package http_listener

import (
	"sync/atomic"
)

type pool struct {
	buffers chan []byte
	size    int

	created int64
}

// NewPool returns a new pool object.
// n is the number of buffers
// bufSize is the size (in bytes) of each buffer
func NewPool(n, bufSize int) *pool {
	return &pool{
		buffers: make(chan []byte, n),
		size:    bufSize,
	}
}

func (p *pool) get() []byte {
	select {
	case b := <-p.buffers:
		return b
	default:
		atomic.AddInt64(&p.created, 1)
		return make([]byte, p.size)
	}
}

func (p *pool) put(b []byte) {
	select {
	case p.buffers <- b:
	default:
	}
}

func (p *pool) ncreated() int64 {
	return atomic.LoadInt64(&p.created)
}
