package trace

import (
	"time"
)

// Trace is an array (or a series) of spans
type Trace []Span

// Span represents a specific zipkin span. It holds the majority of the same
// data as a zipkin span sent via the thrift protocol, but is presented in a
// format which is more straightforward for storage purposes.
type Span struct {
	ID                string
	TraceID           string // zipkin traceid high concat with traceid
	Name              string
	ParentID          string
	ServiceName       string
	Timestamp         time.Time // If zipkin input is nil then time.Now()
	Duration          time.Duration
	Annotations       []Annotation
	BinaryAnnotations []BinaryAnnotation
}

// BinaryAnnotation represents a zipkin binary annotation. It contains
// all of the same fields as might be found in its zipkin counterpart.
type BinaryAnnotation struct {
	Key         string
	Value       string
	Host        string // annotation.endpoint.ipv4 + ":" + annotation.endpoint.port
	ServiceName string
}

// Annotation represents an ordinary zipkin annotation. It contains the data fields
// which will become fields/tags in influxdb
type Annotation struct {
	Timestamp   time.Time
	Value       string
	Host        string // annotation.endpoint.ipv4 + ":" + annotation.endpoint.port
	ServiceName string
}
