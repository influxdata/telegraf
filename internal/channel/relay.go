package channel

import (
	"sync"

	"github.com/influxdata/telegraf"
)

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

func (r *Relay) SetDest(dst chan<- telegraf.Metric) {
	r.Lock()
	defer r.Unlock()
	r.dst = dst
}

func (r *Relay) GetDest() chan<- telegraf.Metric {
	r.Lock()
	defer r.Unlock()
	if r.dst == nil {
		panic("dst channel should never be nil")
	}
	return r.dst
}
