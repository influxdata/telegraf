package snmp_lookup

import (
	"container/list"
	"sync"

	"github.com/influxdata/telegraf"
)

type backlogEntry struct {
	metric   telegraf.Metric
	agent    string
	index    string
	resolved bool
}

type backlog struct {
	elements *list.List
	ordered  bool

	acc telegraf.Accumulator
	log telegraf.Logger

	sync.Mutex
}

func newBacklog(acc telegraf.Accumulator, log telegraf.Logger, ordered bool) *backlog {
	return &backlog{
		elements: list.New(),
		ordered:  ordered,
		acc:      acc,
		log:      log,
	}
}

func (b *backlog) destroy() int {
	b.Lock()
	defer b.Unlock()

	count := b.elements.Len()
	for {
		e := b.elements.Front()
		if e == nil {
			break
		}
		entry := e.Value.(backlogEntry)
		b.acc.AddMetric(entry.metric)

		b.elements.Remove(e)
	}

	return count
}

func (b *backlog) isEmpty() bool {
	b.Lock()
	defer b.Unlock()
	return b.elements.Len() == 0
}

func (b *backlog) push(agent, index string, m telegraf.Metric) {
	e := backlogEntry{
		metric: m,
		agent:  agent,
		index:  index,
	}
	b.Lock()
	defer b.Unlock()
	_ = b.elements.PushBack(e)
}

func (b *backlog) resolve(agent string, tm *tagMap) {
	b.log.Debugf("resolving agent %q", agent)
	b.Lock()
	defer b.Unlock()

	var i int
	var outOfOrder bool
	var forRemoval []*list.Element
	e := b.elements.Front()
	for e != nil {
		entry := e.Value.(backlogEntry)
		b.log.Debugf("  * entry %d: %v", i, entry.metric)
		i++

		// Check if we can resolve the element
		if entry.agent == agent {
			tags, found := tm.rows[entry.index]
			b.log.Debugf("  - agent match, index %s found %v", entry.index, found)
			if found {
				for k, v := range tags {
					entry.metric.AddTag(k, v)
				}
			} else {
				b.log.Warnf("Cannot resolve metrics because index %q not found for agent %q!", entry.index, agent)
			}
			entry.resolved = true
		}

		// Check if we can release the metric in ordered mode...
		outOfOrder = outOfOrder || !entry.resolved
		b.log.Debugf("  - out-of-order: %v    entry resolved: %v", outOfOrder, entry.resolved)
		if entry.resolved && (!b.ordered || !outOfOrder) {
			b.log.Debugf("releasing metric %v", entry.metric)
			b.acc.AddMetric(entry.metric)
			forRemoval = append(forRemoval, e)
		}
		e.Value = entry
		e = e.Next()
	}

	for _, e := range forRemoval {
		b.log.Debugf("  - removing %v", e.Value)
		b.elements.Remove(e)
	}
}
