package zipkin

import (
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs/zipkin/trace"
)

// LineProtocolConverter implements the Recorder interface; it is a
// type meant to encapsulate the storage of zipkin tracing data in
// telegraf as line protocol.
type LineProtocolConverter struct {
	acc telegraf.Accumulator
}

// NewLineProtocolConverter returns an instance of LineProtocolConverter that
// will add to the given telegraf.Accumulator
func NewLineProtocolConverter(acc telegraf.Accumulator) *LineProtocolConverter {
	return &LineProtocolConverter{
		acc: acc,
	}
}

// Record is LineProtocolConverter's implementation of the Record method of
// the Recorder interface; it takes a trace as input, and adds it to an internal
// telegraf.Accumulator.
func (l *LineProtocolConverter) Record(t trace.Trace) error {
	for _, s := range t {
		fields := map[string]interface{}{
			"duration_ns": s.Duration.Nanoseconds(),
		}

		tags := map[string]string{
			"id":           s.ID,
			"parent_id":    s.ParentID,
			"trace_id":     s.TraceID,
			"name":         formatName(s.Name),
			"service_name": formatName(s.ServiceName),
		}
		l.acc.AddFields("zipkin", fields, tags, s.Timestamp)

		for _, a := range s.Annotations {
			tags := map[string]string{
				"id":            s.ID,
				"parent_id":     s.ParentID,
				"trace_id":      s.TraceID,
				"name":          formatName(s.Name),
				"service_name":  formatName(a.ServiceName),
				"annotation":    a.Value,
				"endpoint_host": a.Host,
			}
			l.acc.AddFields("zipkin", fields, tags, s.Timestamp)
		}

		for _, b := range s.BinaryAnnotations {
			tags := map[string]string{
				"id":             s.ID,
				"parent_id":      s.ParentID,
				"trace_id":       s.TraceID,
				"name":           formatName(s.Name),
				"service_name":   formatName(b.ServiceName),
				"annotation":     b.Value,
				"endpoint_host":  b.Host,
				"annotation_key": b.Key,
			}
			l.acc.AddFields("zipkin", fields, tags, s.Timestamp)
		}
	}

	return nil
}

func (l *LineProtocolConverter) Error(err error) {
	l.acc.AddError(err)
}

// formatName formats name and service name
// Zipkin forces span and service names to be lowercase:
// https://github.com/openzipkin/zipkin/pull/805
func formatName(name string) string {
	return strings.ToLower(name)
}
