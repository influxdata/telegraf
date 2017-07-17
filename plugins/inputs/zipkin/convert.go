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
func NewTrace(spans []*zipkincore.Span) (Trace, error) {
	var trace Trace
	for _, span := range spans {
		s := Span{}
		s.ID = strconv.FormatInt(span.GetID(), 10)
		s.TraceID = strconv.FormatInt(span.GetTraceID(), 10)
		if span.GetTraceIDHigh() != 0 {
			s.TraceID = strconv.FormatInt(span.GetTraceIDHigh(), 10) + s.TraceID
		}

		s.Annotations = NewAnnotations(span.GetAnnotations())

		var err error
		s.BinaryAnnotations, err = NewBinaryAnnotations(span.GetBinaryAnnotations())
		if err != nil {
			return nil, err
		}
		s.Name = span.GetName()
		//TODO: find out what zipkin does with a timestamp of zero
		if span.GetTimestamp() == 0 {
			s.Timestamp = time.Now()
		} else {
			s.Timestamp = microToTime(span.GetTimestamp())
		}

		duration := time.Duration(span.GetDuration())
		s.Duration = duration * time.Microsecond

		parentID := span.GetParentID()

		// A parent ID of 0 means that this is a parent span. In this case,
		// we set the parent ID of the span to be its own id, so it points to
		// itself.
		if parentID == 0 {
			s.ParentID = s.ID
		} else {
			s.ParentID = strconv.FormatInt(parentID, 10)
		}

		trace = append(trace, s)
	}

	return trace, nil
}

// NewAnnotations converts a slice of *zipkincore.Annotation into a slice
// of new Annotations
func NewAnnotations(annotations []*zipkincore.Annotation) []Annotation {
	var formatted []Annotation
	for _, annotation := range annotations {
		a := Annotation{}
		endpoint := annotation.GetHost()
		if endpoint != nil {
			//TODO: Fix Ipv4 hostname to bit shifted
			a.Host = strconv.Itoa(int(endpoint.GetIpv4())) + ":" + strconv.Itoa(int(endpoint.GetPort()))
			a.ServiceName = endpoint.GetServiceName()
		} else {
			a.Host, a.ServiceName = "", ""
		}

		a.Timestamp = microToTime(annotation.GetTimestamp())
		a.Value = annotation.GetValue()
		formatted = append(formatted, a)
	}
	//fmt.Println("formatted annotations: ", formatted)
	return formatted
}

// NewBinaryAnnotations is very similar to NewAnnotations, but it
// converts zipkincore.BinaryAnnotations instead of the normal zipkincore.Annotation
func NewBinaryAnnotations(annotations []*zipkincore.BinaryAnnotation) ([]BinaryAnnotation, error) {
	var formatted []BinaryAnnotation
	for _, annotation := range annotations {
		b := BinaryAnnotation{}
		endpoint := annotation.GetHost()
		if endpoint != nil {
			b.Host = strconv.Itoa(int(endpoint.GetIpv4())) + ":" + strconv.Itoa(int(endpoint.GetPort()))
			b.ServiceName = endpoint.GetServiceName()
		} else {
			b.Host, b.ServiceName = "", ""
		}

		b.Key = annotation.GetKey()
		b.Value = string(annotation.GetValue())
		b.Type = annotation.GetAnnotationType().String()
		formatted = append(formatted, b)
	}

	return formatted, nil
}

func microToTime(micro int64) time.Time {
	return time.Unix(0, micro*int64(time.Microsecond))
}
