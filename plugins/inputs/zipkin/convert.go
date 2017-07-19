package zipkin

import (
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/openzipkin/zipkin-go-opentracing/_thrift/gen-go/zipkincore"
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
func (l *LineProtocolConverter) Record(t Trace) error {
	for _, s := range t {
		for _, a := range s.Annotations {
			fields := map[string]interface{}{
				// TODO: Maybe we don't need "annotation_timestamp"?
				"annotation_timestamp": a.Timestamp.Unix(),
				"duration":             s.Duration,
			}

			tags := map[string]string{
				"id":               s.ID,
				"parent_id":        s.ParentID,
				"trace_id":         s.TraceID,
				"name":             s.Name,
				"service_name":     a.ServiceName,
				"annotation_value": a.Value,
				"endpoint_host":    a.Host,
			}
			l.acc.AddFields("zipkin", fields, tags, s.Timestamp)
		}

		for _, b := range s.BinaryAnnotations {
			fields := map[string]interface{}{
				"duration": s.Duration,
			}

			tags := map[string]string{
				"id":               s.ID,
				"parent_id":        s.ParentID,
				"trace_id":         s.TraceID,
				"name":             s.Name,
				"service_name":     b.ServiceName,
				"annotation_value": b.Value,
				"endpoint_host":    b.Host,
				"key":              b.Key,
				"type":             b.Type,
			}
			l.acc.AddFields("zipkin", fields, tags, s.Timestamp)
		}
	}

	return nil
}

func (l *LineProtocolConverter) Error(err error) {
	l.acc.AddError(err)
}

// NewTrace converts a slice of []*zipkincore.Spans into a new Trace
func NewTrace(spans []*zipkincore.Span) Trace {
	trace := make(Trace, len(spans))
	for i, span := range spans {
		trace[i] = Span{
			ID:                formatID(span.GetID()),
			TraceID:           formatTraceID(span.GetTraceIDHigh(), span.GetTraceID()),
			Name:              span.GetName(),
			Timestamp:         guessTimestamp(span),
			Duration:          convertDuration(span),
			ParentID:          parentID(span),
			Annotations:       NewAnnotations(span.GetAnnotations()),
			BinaryAnnotations: NewBinaryAnnotations(span.GetBinaryAnnotations()),
		}
	}
	return trace
}

// NewAnnotations converts a slice of *zipkincore.Annotation into a slice
// of new Annotations
func NewAnnotations(annotations []*zipkincore.Annotation) []Annotation {
	formatted := make([]Annotation, len(annotations))
	for i, annotation := range annotations {
		formatted[i] = Annotation{
			Host:        host(annotation.GetHost()),
			ServiceName: serviceName(annotation.GetHost()),
			Timestamp:   microToTime(annotation.GetTimestamp()),
			Value:       annotation.GetValue(),
		}
	}

	return formatted
}

// NewBinaryAnnotations is very similar to NewAnnotations, but it
// converts zipkincore.BinaryAnnotations instead of the normal zipkincore.Annotation
func NewBinaryAnnotations(annotations []*zipkincore.BinaryAnnotation) []BinaryAnnotation {
	formatted := make([]BinaryAnnotation, len(annotations))
	for i, annotation := range annotations {
		formatted[i] = BinaryAnnotation{
			Host:        host(annotation.GetHost()),
			ServiceName: serviceName(annotation.GetHost()),
			Key:         annotation.GetKey(),
			Value:       string(annotation.GetValue()),
			Type:        annotation.GetAnnotationType().String(),
		}
	}
	return formatted
}

func microToTime(micro int64) time.Time {
	return time.Unix(0, micro*int64(time.Microsecond))
}

func formatID(id int64) string {
	return strconv.FormatInt(id, 10)
}

func formatTraceID(high, low int64) string {
	return formatID(high) + ":" + formatID(low)
}

func minMax(span *zipkincore.Span) (time.Time, time.Time) {
	min := time.Now()
	max := time.Unix(0, 0)
	for _, annotation := range span.Annotations {
		ts := microToTime(annotation.GetTimestamp())
		if !ts.IsZero() && ts.Before(min) {
			min = ts
		}
		if !ts.IsZero() && ts.After(max) {
			max = ts
		}
	}
	if max.IsZero() {
		max = min
	}
	return min, max
}

func guessTimestamp(span *zipkincore.Span) time.Time {
	if span.GetTimestamp() != 0 {
		return microToTime(span.GetTimestamp())
	}
	min, _ := minMax(span)
	return min
}

func convertDuration(span *zipkincore.Span) time.Duration {
	duration := time.Duration(span.GetDuration()) * time.Microsecond
	if duration != 0 {
		return duration
	}
	min, max := minMax(span)
	return max.Sub(min)
}

func parentID(span *zipkincore.Span) string {
	// A parent ID of 0 means that this is a parent span. In this case,
	// we set the parent ID of the span to be its own id, so it points to
	// itself.
	id := span.GetParentID()
	if id != 0 {
		return formatID(id)
	}
	return formatID(span.ID)
}

func host(h *zipkincore.Endpoint) string {
	if h == nil {
		return ""
	}
	return strconv.Itoa(int(h.GetIpv4())) + ":" + strconv.Itoa(int(h.GetPort()))
}

func serviceName(h *zipkincore.Endpoint) string {
	if h == nil {
		return ""
	}
	return h.GetServiceName()
}
