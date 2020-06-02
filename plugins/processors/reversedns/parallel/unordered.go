package parallel

import (
	"context"
	"sync"

	"github.com/influxdata/telegraf"
	"golang.org/x/sync/semaphore"
)

type Unordered struct {
	wg    sync.WaitGroup
	acc   telegraf.MetricStreamAccumulator
	queue chan telegraf.Metric
	sem   *semaphore.Weighted
}

func NewUnordered(acc telegraf.MetricStreamAccumulator, workerCount int64) *Unordered {
	queue := make(chan telegraf.Metric, 10)
	p := &Unordered{
		acc:   acc,
		queue: queue,
		sem:   semaphore.NewWeighted(workerCount),
	}

	p.wg.Add(1)
	go func() {
		p.readQueue()
		p.wg.Done()
	}()
	return p
}

func (p *Unordered) Do(fn func(acc telegraf.MetricStreamAccumulator)) {
	p.sem.Acquire(context.TODO(), 1)
	go func() {
		fn(p)
		p.sem.Release(1)
	}()
}

func (p *Unordered) readQueue() {
	for m := range p.queue {
		if m != nil {
			p.acc.PassMetric(m)
		}
	}
}

func (p *Unordered) Stop() {
	close(p.queue)
	p.wg.Wait()
}

// match the accumulator interface so we can pose as it to track the metric count
func (p *Unordered) PassMetric(m telegraf.Metric) {
	p.queue <- m
}
