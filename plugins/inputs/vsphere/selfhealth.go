package vsphere

import (
	"time"

	"github.com/influxdata/telegraf/selfstat"
)

// Stopwatch is a simple helper for recording timing information,
// such as gather times and discovery times.
type Stopwatch struct {
	stat  selfstat.Stat
	start time.Time
}

// NewStopwatch creates a new StopWatch and starts measuring time
// its creation.
func NewStopwatch(name, vCenter string) *Stopwatch {
	return &Stopwatch{
		stat:  selfstat.RegisterTiming("vsphere", name+"_ns", map[string]string{"vcenter": vCenter}),
		start: time.Now(),
	}
}

// NewStopwatchWithTags creates a new StopWatch and starts measuring time
// its creation. Allows additional tags.
func NewStopwatchWithTags(name, vCenter string, tags map[string]string) *Stopwatch {
	tags["vcenter"] = vCenter
	return &Stopwatch{
		stat:  selfstat.RegisterTiming("vsphere", name+"_ns", tags),
		start: time.Now(),
	}
}

// Stop stops a Stopwatch and records the time.
func (s *Stopwatch) Stop() {
	s.stat.Set(time.Since(s.start).Nanoseconds())
}

// SendInternalCounter is a convenience method for sending
// non-timing internal metrics.
func SendInternalCounter(name, vCenter string, value int64) {
	s := selfstat.Register("vsphere", name, map[string]string{"vcenter": vCenter})
	s.Set(value)
}

// SendInternalCounterWithTags is a convenience method for sending
// non-timing internal metrics. Allows additional tags
func SendInternalCounterWithTags(name, vCenter string, tags map[string]string, value int64) {
	tags["vcenter"] = vCenter
	s := selfstat.Register("vsphere", name, tags)
	s.Set(value)
}
