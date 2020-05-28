package parallel

import (
	"sync"

	"github.com/influxdata/telegraf"
)

type Unordered struct {
	wg    sync.WaitGroup
	acc   telegraf.MetricStreamAccumulator
	queue chan telegraf.Metric
	pool  WorkerPool
}

func NewUnordered(acc telegraf.MetricStreamAccumulator, workerCount int) *Unordered {
	queue := make(chan telegraf.Metric, 10)
	p := &Unordered{
		acc:   acc,
		queue: queue,
		pool:  NewWorkerPool(workerCount),
	}
	go p.readQueue()
	return p
}

func (p *Unordered) Do(fn func(acc telegraf.MetricStreamAccumulator)) {
	p.wg.Add(1)
	p.pool.Checkout()
	go func() {
		fn(p)
		p.pool.Checkin()
		p.wg.Done()
	}()
}

func (p *Unordered) readQueue() {
	for m := range p.queue {
		if m != nil {
			p.acc.PassMetric(m)
		}
		p.wg.Done()
	}
}

func (p *Unordered) Wait() {
	close(p.queue)
	p.wg.Wait()
}

// match the accumulator interface so we can pose as it to track the metric count
func (p *Unordered) PassMetric(m telegraf.Metric) {
	p.wg.Add(1)
	p.queue <- m
}
