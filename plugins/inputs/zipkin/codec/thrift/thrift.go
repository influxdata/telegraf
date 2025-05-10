package thrift

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/apache/thrift/lib/go/thrift"

	"github.com/influxdata/telegraf/plugins/inputs/zipkin/codec"
	"github.com/influxdata/telegraf/plugins/inputs/zipkin/codec/thrift/gen-go/zipkincore"
)

// Thrift decodes binary data to create a Trace
type Thrift struct{}

// Decode unmarshals and validates bytes in thrift format
func (*Thrift) Decode(octets []byte) ([]codec.Span, error) {
	spans, err := unmarshalThrift(octets)
	if err != nil {
		return nil, err
	}

	res := make([]codec.Span, 0, len(spans))
	for _, s := range spans {
		res = append(res, &span{s})
	}
	return res, nil
}

// unmarshalThrift converts raw bytes in thrift format to a slice of spans
func unmarshalThrift(body []byte) ([]*zipkincore.Span, error) {
	buffer := thrift.NewTMemoryBuffer()
	buffer.Write(body)

	transport := thrift.NewTBinaryProtocolConf(buffer, nil)
	_, size, err := transport.ReadListBegin(context.Background())
	if err != nil {
		return nil, err
	}

	spans := make([]*zipkincore.Span, 0, size)
	for i := 0; i < size; i++ {
		zs := &zipkincore.Span{}
		if err := zs.Read(context.Background(), transport); err != nil {
			return nil, err
		}
		spans = append(spans, zs)
	}

	if err := transport.ReadListEnd(context.Background()); err != nil {
		return nil, err
	}
	return spans, nil
}

var _ codec.Endpoint = &endpoint{}

type endpoint struct {
	*zipkincore.Endpoint
}

// Host returns the host address of the endpoint as a string.
func (e *endpoint) Host() string {
	ipv4 := func(addr int32) string {
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, uint32(addr))
		return net.IP(buf).String()
	}

	if e.Endpoint == nil {
		return ipv4(int32(0))
	}
	if e.Endpoint.GetPort() == 0 {
		return ipv4(e.Endpoint.GetIpv4())
	}
	// Zipkin uses a signed int16 for the port, but, warns us that they actually treat it
	// as an unsigned int16. So, we convert from int16 to int32 followed by taking & 0xffff
	// to convert from signed to unsigned
	// https://github.com/openzipkin/zipkin/blob/57dc2ec9c65fe6144e401c0c933b4400463a69df/zipkin/src/main/java/zipkin/Endpoint.java#L44
	return ipv4(e.Endpoint.GetIpv4()) + ":" + strconv.FormatInt(int64(int(e.Endpoint.GetPort())&0xffff), 10)
}

// Name returns the name of the service associated with the endpoint as a string.
func (e *endpoint) Name() string {
	if e.Endpoint == nil {
		return codec.DefaultServiceName
	}
	return e.Endpoint.GetServiceName()
}

var _ codec.BinaryAnnotation = &binaryAnnotation{}

type binaryAnnotation struct {
	*zipkincore.BinaryAnnotation
}

// Key returns the key of the binary annotation as a string.
func (b *binaryAnnotation) Key() string {
	return b.BinaryAnnotation.GetKey()
}

// Value returns the value of the binary annotation as a string.
func (b *binaryAnnotation) Value() string {
	return string(b.BinaryAnnotation.GetValue())
}

// Host returns the endpoint associated with the binary annotation as a codec.Endpoint.
func (b *binaryAnnotation) Host() codec.Endpoint {
	if b.BinaryAnnotation.Host == nil {
		return nil
	}
	return &endpoint{b.BinaryAnnotation.Host}
}

var _ codec.Annotation = &annotation{}

type annotation struct {
	*zipkincore.Annotation
}

// Timestamp returns the timestamp of the annotation as a time.Time object.
func (a *annotation) Timestamp() time.Time {
	ts := a.Annotation.GetTimestamp()
	if ts == 0 {
		return time.Time{}
	}
	return codec.MicroToTime(ts)
}

// Value returns the value of the annotation as a string.
func (a *annotation) Value() string {
	return a.Annotation.GetValue()
}

// Host returns the endpoint associated with the annotation as a codec.Endpoint.
func (a *annotation) Host() codec.Endpoint {
	if a.Annotation.Host == nil {
		return nil
	}
	return &endpoint{a.Annotation.Host}
}

var _ codec.Span = &span{}

type span struct {
	*zipkincore.Span
}

// Trace returns the trace ID of the span and an error if the trace ID is invalid.
func (s *span) Trace() (string, error) {
	if s.Span.GetTraceIDHigh() == 0 && s.Span.GetTraceID() == 0 {
		return "", errors.New("span does not have a trace ID")
	}

	if s.Span.GetTraceIDHigh() == 0 {
		return fmt.Sprintf("%x", s.Span.GetTraceID()), nil
	}
	return fmt.Sprintf("%x%016x", s.Span.GetTraceIDHigh(), s.Span.GetTraceID()), nil
}

// SpanID returns the span ID of the span and an error if the span ID is invalid.
func (s *span) SpanID() (string, error) {
	return formatID(s.Span.GetID()), nil
}

// Parent returns the parent span ID of the span and an error if the parent ID is invalid.
func (s *span) Parent() (string, error) {
	id := s.Span.GetParentID()
	if id != 0 {
		return formatID(id), nil
	}
	return "", nil
}

// Name returns the name of the span.
func (s *span) Name() string {
	return s.Span.GetName()
}

// Annotations returns the annotations of the span as a slice of codec.Annotation.
func (s *span) Annotations() []codec.Annotation {
	res := make([]codec.Annotation, 0, len(s.Span.Annotations))
	for _, ann := range s.Span.Annotations {
		res = append(res, &annotation{ann})
	}
	return res
}

// BinaryAnnotations returns the binary annotations of the span as a slice of codec.BinaryAnnotation and an error if the binary annotations cannot be retrieved.
func (s *span) BinaryAnnotations() ([]codec.BinaryAnnotation, error) {
	res := make([]codec.BinaryAnnotation, 0, len(s.Span.BinaryAnnotations))
	for _, ann := range s.Span.BinaryAnnotations {
		res = append(res, &binaryAnnotation{ann})
	}
	return res, nil
}

// Timestamp returns the timestamp of the span as a time.Time object.
func (s *span) Timestamp() time.Time {
	ts := s.Span.GetTimestamp()
	if ts == 0 {
		return time.Time{}
	}
	return codec.MicroToTime(ts)
}

// Duration returns the duration of the span as a time.Duration object.
func (s *span) Duration() time.Duration {
	return time.Duration(s.Span.GetDuration()) * time.Microsecond
}

// formatID formats the given ID as a hexadecimal string.
func formatID(id int64) string {
	return strconv.FormatInt(id, 16)
}
