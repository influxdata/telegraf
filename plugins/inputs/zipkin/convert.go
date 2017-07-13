package zipkin

import (
	"strconv"
	"time"

	"github.com/openzipkin/zipkin-go-opentracing/_thrift/gen-go/zipkincore"
)

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
		if span.GetTimestamp() == 0 {
			s.Timestamp = time.Now()
		} else {
			s.Timestamp = time.Unix(0, span.GetTimestamp()*int64(time.Microsecond))
		}

		duration := time.Duration(span.GetDuration())
		//	fmt.Println("Duration: ", duration)
		s.Duration = duration * time.Microsecond

		parentID := span.GetParentID()
		//	fmt.Println("Parent ID: ", parentID)

		// A parent ID of 0 means that this is a parent span. In this case,
		// we set the parent ID of the span to be its own id, so it points to
		// itself.

		if parentID == 0 {
			s.ParentID = s.ID
		} else {
			s.ParentID = strconv.FormatInt(parentID, 10)
		}

		//	fmt.Println("ID:", s.ID)
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
			a.Host = strconv.Itoa(int(endpoint.GetIpv4())) + ":" + strconv.Itoa(int(endpoint.GetPort()))
			a.ServiceName = endpoint.GetServiceName()
		} else {
			a.Host, a.ServiceName = "", ""
		}

		a.Timestamp = time.Unix(annotation.GetTimestamp(), 0)
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
