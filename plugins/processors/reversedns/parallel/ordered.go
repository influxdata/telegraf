package parallel

import (
	"sync"

	"github.com/influxdata/telegraf"
)

type Ordered struct {
	wg    sync.WaitGroup
	acc   telegraf.MetricStreamAccumulator
	queue chan futureMetric
}

func NewOrdered(acc telegraf.MetricStreamAccumulator, workerCount int) *Ordered {
	p := &Ordered{
		acc:   acc,
		queue: make(chan futureMetric, workerCount),
	}
	go p.readQueue()
	return p
}

func (p *Ordered) Do(fn func(acc telegraf.MetricStreamAccumulator)) {
	p.wg.Add(1)

	oa := orderedAccumulator{
		fn: fn,
	}
	oa.run(p.queue)
}

func (p *Ordered) readQueue() {
	// wait for the response from each worker in order
	for mCh := range p.queue {
		// allow each worker to write out multiple metrics
		for m := range mCh {
			if m != nil {
				p.acc.PassMetric(m)
			}
		}
		p.wg.Done()
	}
}

func (p *Ordered) Wait() {
	close(p.queue)
	p.wg.Wait()
}

type futureMetric chan telegraf.Metric

type orderedAccumulator struct {
	future futureMetric
	fn     func(acc telegraf.MetricStreamAccumulator)
}

func (o *orderedAccumulator) PassMetric(m telegraf.Metric) {
	o.future <- m
}

func (o *orderedAccumulator) run(queue chan futureMetric) {
	o.future = make(futureMetric)
	queue <- o.future // must write future chan to queue before launching goroutine

	go func() {
		o.fn(o)
		close(o.future)
	}()
}
