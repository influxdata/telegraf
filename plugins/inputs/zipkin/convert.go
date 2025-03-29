package zipkin

import (
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs/zipkin/trace"
)

// lineProtocolConverter implements the recorder interface;
// it is a type meant to encapsulate the storage of zipkin tracing data in telegraf as line protocol.
type lineProtocolConverter struct {
	acc telegraf.Accumulator
}

// newLineProtocolConverter returns an instance of lineProtocolConverter that will add to the given telegraf.Accumulator
func newLineProtocolConverter(acc telegraf.Accumulator) *lineProtocolConverter {
	return &lineProtocolConverter{
		acc: acc,
	}
}

// record is lineProtocolConverter's implementation of the record method of the recorder interface;
// it takes a trace as input, and adds it to an internal telegraf.Accumulator.
func (l *lineProtocolConverter) record(t trace.Trace) error {
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

func (l *lineProtocolConverter) error(err error) {
	l.acc.AddError(err)
}

// formatName formats name and service name Zipkin forces span and service names to be lowercase:
// https://github.com/openzipkin/zipkin/pull/805
func formatName(name string) string {
	return strings.ToLower(name)
}
