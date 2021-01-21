package codec

import (
	"time"

	"github.com/influxdata/telegraf/plugins/inputs/zipkin/trace"
	"github.com/openzipkin/zipkin-go-opentracing/thrift/gen-go/zipkincore"
)

//now is a mockable time for now
var now = time.Now

// DefaultServiceName when the span does not have any serviceName
const DefaultServiceName = "unknown"

// Decoder decodes the bytes and returns a trace
type Decoder interface {
	Decode(octets []byte) ([]Span, error)
}

// Span are created by instrumentation in RPC clients or servers
type Span interface {
	Trace() (string, error)
	SpanID() (string, error)
	Parent() (string, error)
	Name() string
	Annotations() []Annotation
	BinaryAnnotations() ([]BinaryAnnotation, error)
	Timestamp() time.Time
	Duration() time.Duration
}

// Annotation represents an event that explains latency with a timestamp.
type Annotation interface {
	Timestamp() time.Time
	Value() string
	Host() Endpoint
}

// BinaryAnnotation represent tags applied to a Span to give it context
type BinaryAnnotation interface {
	Key() string
	Value() string
	Host() Endpoint
}

// Endpoint represents the network context of a service recording an annotation
type Endpoint interface {
	Host() string
	Name() string
}

// DefaultEndpoint is used if the annotations have no endpoints
type DefaultEndpoint struct{}

// Host returns 0.0.0.0; used when the host is unknown
func (d *DefaultEndpoint) Host() string { return "0.0.0.0" }

// Name returns "unknown" when an endpoint doesn't exist
func (d *DefaultEndpoint) Name() string { return DefaultServiceName }

// MicroToTime converts zipkin's native time of microseconds into time.Time
func MicroToTime(micro int64) time.Time {
	return time.Unix(0, micro*int64(time.Microsecond)).UTC()
}

// NewTrace converts a slice of []Span into a new Trace
func NewTrace(spans []Span) (trace.Trace, error) {
	tr := make(trace.Trace, len(spans))
	for i, span := range spans {
		bin, err := span.BinaryAnnotations()
		if err != nil {
			return nil, err
		}
		endpoint := serviceEndpoint(span.Annotations(), bin)
		id, err := span.SpanID()
		if err != nil {
			return nil, err
		}

		tid, err := span.Trace()
		if err != nil {
			return nil, err
		}

		pid, err := parentID(span)
		if err != nil {
			return nil, err
		}

		tr[i] = trace.Span{
			ID:                id,
			TraceID:           tid,
			Name:              span.Name(),
			Timestamp:         guessTimestamp(span),
			Duration:          convertDuration(span),
			ParentID:          pid,
			ServiceName:       endpoint.Name(),
			Annotations:       NewAnnotations(span.Annotations(), endpoint),
			BinaryAnnotations: NewBinaryAnnotations(bin, endpoint),
		}
	}
	return tr, nil
}

// NewAnnotations converts a slice of Annotation into a slice of new Annotations
func NewAnnotations(annotations []Annotation, endpoint Endpoint) []trace.Annotation {
	formatted := make([]trace.Annotation, len(annotations))
	for i, annotation := range annotations {
		formatted[i] = trace.Annotation{
			Host:        endpoint.Host(),
			ServiceName: endpoint.Name(),
			Timestamp:   annotation.Timestamp(),
			Value:       annotation.Value(),
		}
	}

	return formatted
}

// NewBinaryAnnotations is very similar to NewAnnotations, but it
// converts BinaryAnnotations instead of the normal Annotation
func NewBinaryAnnotations(annotations []BinaryAnnotation, endpoint Endpoint) []trace.BinaryAnnotation {
	formatted := make([]trace.BinaryAnnotation, len(annotations))
	for i, annotation := range annotations {
		formatted[i] = trace.BinaryAnnotation{
			Host:        endpoint.Host(),
			ServiceName: endpoint.Name(),
			Key:         annotation.Key(),
			Value:       annotation.Value(),
		}
	}
	return formatted
}

func minMax(span Span) (time.Time, time.Time) {
	min := now().UTC()
	max := time.Time{}.UTC()
	for _, annotation := range span.Annotations() {
		ts := annotation.Timestamp()
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

func guessTimestamp(span Span) time.Time {
	ts := span.Timestamp()
	if !ts.IsZero() {
		return ts
	}

	min, _ := minMax(span)
	return min
}

func convertDuration(span Span) time.Duration {
	duration := span.Duration()
	if duration != 0 {
		return duration
	}
	min, max := minMax(span)
	return max.Sub(min)
}

func parentID(span Span) (string, error) {
	// A parent ID of "" means that this is a parent span. In this case,
	// we set the parent ID of the span to be its own id, so it points to
	// itself.
	id, err := span.Parent()
	if err != nil {
		return "", err
	}

	if id != "" {
		return id, nil
	}
	return span.SpanID()
}

func serviceEndpoint(ann []Annotation, bann []BinaryAnnotation) Endpoint {
	for _, a := range ann {
		switch a.Value() {
		case zipkincore.SERVER_RECV, zipkincore.SERVER_SEND, zipkincore.CLIENT_RECV, zipkincore.CLIENT_SEND:
			if a.Host() != nil && a.Host().Name() != "" {
				return a.Host()
			}
		}
	}

	for _, a := range bann {
		if a.Key() == zipkincore.LOCAL_COMPONENT && a.Host() != nil && a.Host().Name() != "" {
			return a.Host()
		}
	}
	// Unable to find any "standard" endpoint host, so, use any that exist in the regular annotations
	for _, a := range ann {
		if a.Host() != nil && a.Host().Name() != "" {
			return a.Host()
		}
	}
	return &DefaultEndpoint{}
}
