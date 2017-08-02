package zipkin

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/openzipkin/zipkin-go-opentracing/_thrift/gen-go/zipkincore"
)

// DefaultServiceName when the span does not have any serviceName
const DefaultServiceName = "unknown"

//now is a moackable time for now
var now = time.Now

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
		fields := map[string]interface{}{
			"duration_ns": s.Duration.Nanoseconds(),
		}
		tags := map[string]string{
			"id":           s.ID,
			"parent_id":    s.ParentID,
			"trace_id":     s.TraceID,
			"name":         s.Name,
			"service_name": s.ServiceName,
		}
		l.acc.AddFields("zipkin", fields, tags, s.Timestamp)

		for _, a := range s.Annotations {
			tags := map[string]string{
				"id":            s.ID,
				"parent_id":     s.ParentID,
				"trace_id":      s.TraceID,
				"name":          s.Name,
				"service_name":  a.ServiceName,
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
				"name":           s.Name,
				"service_name":   b.ServiceName,
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

// NewTrace converts a slice of []*zipkincore.Spans into a new Trace
func NewTrace(spans []*zipkincore.Span) Trace {
	trace := make(Trace, len(spans))
	for i, span := range spans {
		endpoint := serviceEndpoint(span.GetAnnotations(), span.GetBinaryAnnotations())
		trace[i] = Span{
			ID:                formatID(span.GetID()),
			TraceID:           formatTraceID(span.GetTraceIDHigh(), span.GetTraceID()),
			Name:              span.GetName(),
			Timestamp:         guessTimestamp(span),
			Duration:          convertDuration(span),
			ParentID:          parentID(span),
			ServiceName:       serviceName(endpoint),
			Annotations:       NewAnnotations(span.GetAnnotations(), endpoint),
			BinaryAnnotations: NewBinaryAnnotations(span.GetBinaryAnnotations(), endpoint),
		}
	}
	return trace
}

// NewAnnotations converts a slice of *zipkincore.Annotation into a slice
// of new Annotations
func NewAnnotations(annotations []*zipkincore.Annotation, endpoint *zipkincore.Endpoint) []Annotation {
	formatted := make([]Annotation, len(annotations))
	for i, annotation := range annotations {
		formatted[i] = Annotation{
			Host:        host(endpoint),
			ServiceName: serviceName(endpoint),
			Timestamp:   microToTime(annotation.GetTimestamp()),
			Value:       annotation.GetValue(),
		}
	}

	return formatted
}

// NewBinaryAnnotations is very similar to NewAnnotations, but it
// converts zipkincore.BinaryAnnotations instead of the normal zipkincore.Annotation
func NewBinaryAnnotations(annotations []*zipkincore.BinaryAnnotation, endpoint *zipkincore.Endpoint) []BinaryAnnotation {
	formatted := make([]BinaryAnnotation, len(annotations))
	for i, annotation := range annotations {
		formatted[i] = BinaryAnnotation{
			Host:        host(endpoint),
			ServiceName: serviceName(endpoint),
			Key:         annotation.GetKey(),
			Value:       string(annotation.GetValue()),
			Type:        annotation.GetAnnotationType().String(),
		}
	}
	return formatted
}

func microToTime(micro int64) time.Time {
	return time.Unix(0, micro*int64(time.Microsecond)).UTC()
}

func formatID(id int64) string {
	return strconv.FormatInt(id, 10)
}

func formatTraceID(high, low int64) string {
	if high == 0 {
		return fmt.Sprintf("%x", low)
	}
	return fmt.Sprintf("%x%016x", high, low)
}

func minMax(span *zipkincore.Span) (time.Time, time.Time) {
	min := now().UTC()
	max := time.Time{}.UTC()
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

func ipv4(addr int32) string {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(addr))
	return net.IP(buf).String()
}

func host(h *zipkincore.Endpoint) string {
	if h == nil {
		return ipv4(int32(0))
	}
	if h.GetPort() == 0 {
		return ipv4(h.GetIpv4())
	}
	// Zipkin uses a signed int16 for the port, but, warns us that they actually treat it
	// as an unsigned int16. So, we convert from int16 to int32 followed by taking & 0xffff
	// to convert from signed to unsigned
	// https://github.com/openzipkin/zipkin/blob/57dc2ec9c65fe6144e401c0c933b4400463a69df/zipkin/src/main/java/zipkin/Endpoint.java#L44
	return ipv4(h.GetIpv4()) + ":" + strconv.FormatInt(int64(int(h.GetPort())&0xffff), 10)
}

func serviceName(h *zipkincore.Endpoint) string {
	if h == nil {
		return DefaultServiceName
	}
	return h.GetServiceName()
}

func serviceEndpoint(ann []*zipkincore.Annotation, bann []*zipkincore.BinaryAnnotation) *zipkincore.Endpoint {
	for _, a := range ann {
		switch a.Value {
		case zipkincore.SERVER_RECV, zipkincore.SERVER_SEND, zipkincore.CLIENT_RECV, zipkincore.CLIENT_SEND:
			if a.Host != nil && a.Host.ServiceName != "" {
				return a.Host
			}
		}
	}

	for _, a := range bann {
		if a.Key == zipkincore.LOCAL_COMPONENT && a.Host != nil && a.Host.ServiceName != "" {
			return a.Host
		}
	}
	// Unable to find any "standard" endpoint host, so, use any that exist in the regular annotations
	for _, a := range ann {
		if a.Host != nil && a.Host.ServiceName != "" {
			return a.Host
		}
	}
	return nil
}
