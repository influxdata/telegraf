package vsphere

import (
	"time"

	"github.com/influxdata/telegraf/selfstat"
)

// stopwatch is a simple helper for recording timing information, such as gather times and discovery times.
type stopwatch struct {
	stat  selfstat.Stat
	start time.Time
}

// newStopwatch creates a new StopWatch and starts measuring time its creation.
func newStopwatch(name, vCenter string) *stopwatch {
	return &stopwatch{
		stat:  selfstat.RegisterTiming("vsphere", name+"_ns", map[string]string{"vcenter": vCenter}),
		start: time.Now(),
	}
}

// newStopwatchWithTags creates a new StopWatch and starts measuring time its creation. Allows additional tags.
func newStopwatchWithTags(name, vCenter string, tags map[string]string) *stopwatch {
	tags["vcenter"] = vCenter
	return &stopwatch{
		stat:  selfstat.RegisterTiming("vsphere", name+"_ns", tags),
		start: time.Now(),
	}
}

// stop stops a stopwatch and records the time.
func (s *stopwatch) stop() {
	s.stat.Set(time.Since(s.start).Nanoseconds())
}

// sendInternalCounterWithTags is a convenience method for sending non-timing internal metrics. Allows additional tags
func sendInternalCounterWithTags(name, vCenter string, tags map[string]string, value int64) {
	tags["vcenter"] = vCenter
	s := selfstat.Register("vsphere", name, tags)
	s.Set(value)
}
