package channel

import (
	"sync"

	"github.com/influxdata/telegraf"
)

// Relay metrics from one channel to another. Connects a source channel to a destination channel
// Think of it as connecting two pipes together.
// dst is closed when src is closed and the last message has been written out to dst.
type Relay struct {
	sync.Mutex
	src <-chan telegraf.Metric
	dst chan<- telegraf.Metric
}

func NewRelay(src <-chan telegraf.Metric, dst chan<- telegraf.Metric) *Relay {
	return &Relay{
		src: src,
		dst: dst,
	}
}

func (r *Relay) Start() {
	go func() {
		for m := range r.src {
			r.GetDest() <- m
		}
		close(r.GetDest())
	}()
}

// SetDest changes the destination channel. this is to make channels hot-swappable, kind of like adding or removing items in a linked list.
// Should not be called after the source channel closes.
func (r *Relay) SetDest(dst chan<- telegraf.Metric) {
	r.Lock()
	defer r.Unlock()
	r.dst = dst
}

// GetDest is the current dst channel
func (r *Relay) GetDest() chan<- telegraf.Metric {
	r.Lock()
	defer r.Unlock()
	if r.dst == nil {
		panic("dst channel should never be nil")
	}
	return r.dst
}
