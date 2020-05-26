package parallel

import (
	"sync"

	"github.com/influxdata/telegraf"
)

type Parallel struct {
	sync.WaitGroup
	acc          telegraf.MetricStreamAccumulator
	ordered      bool
	workerCount  int
	orderedQueue chan futureMetric
}

func New(acc telegraf.MetricStreamAccumulator, workerCount int) *Parallel {
	p := &Parallel{
		acc:          acc,
		ordered:      true,
		workerCount:  10,
		orderedQueue: make(chan futureMetric, workerCount),
	}
	go p.background()
	return p
}

func (p *Parallel) Ordered() *Parallel {
	p.ordered = true
	return p
}

func (p *Parallel) Unordered() *Parallel {
	p.ordered = false
	return p
}

func (p *Parallel) Parallel(fn func(acc telegraf.MetricStreamAccumulator)) {
	p.Add(1)

	oa := orderedAccumulator{
		fn: fn,
	}
	p.orderedQueue <- oa.run()
}

func (p *Parallel) background() {
	for mCh := range p.orderedQueue {
		m := <-mCh
		if m != nil {
			p.acc.PassMetric(m)
		}
		p.Done()
	}
}

func (p *Parallel) Wait() {
	close(p.orderedQueue)
	p.WaitGroup.Wait()
}

// Wait from sync.WaitGroup
// func (p *Parallel) Wait()

type futureMetric chan telegraf.Metric

type orderedAccumulator struct {
	future futureMetric
	fn     func(acc telegraf.MetricStreamAccumulator)
}

func (o *orderedAccumulator) PassMetric(m telegraf.Metric) {
	o.future <- m
}
func (o *orderedAccumulator) DropMetric(m telegraf.Metric) {
	o.future <- nil
}

func (o *orderedAccumulator) run() futureMetric {
	o.future = make(futureMetric)
	go o.fn(o)
	return o.future
}
