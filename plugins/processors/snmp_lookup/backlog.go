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
		b.log.Debugf("Adding unresolved metric %v", entry.metric)
		b.acc.AddMetric(entry.metric)

		b.elements.Remove(e)
	}

	return count
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
	b.Lock()
	defer b.Unlock()

	var outOfOrder bool
	var forRemoval []*list.Element
	e := b.elements.Front()
	for e != nil {
		entry := e.Value.(backlogEntry)

		// Check if we can resolve the element
		if entry.agent == agent {
			tags, found := tm.rows[entry.index]
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
		if entry.resolved && (!b.ordered || !outOfOrder) {
			b.acc.AddMetric(entry.metric)
			forRemoval = append(forRemoval, e)
		}
		e.Value = entry
		e = e.Next()
	}

	// We need to remove the elements in a separate loop to not interfere with
	// the list iteration above.
	for _, e := range forRemoval {
		b.elements.Remove(e)
	}
}
