package zipkin

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/openzipkin/zipkin-go-opentracing/_thrift/gen-go/zipkincore"
)

const (
	// DefaultPort is the default port zipkin listens on, which zipkin implementations
	// expect.
	DefaultPort = 9411

	// DefaultRoute is the default route zipkin uses, and zipkin implementations
	// expect.
	DefaultRoute = "/api/v1/spans"

	DefaultShutdownTimeout = 5
)

// TODO: connect methods lexically; method implementations should go right under
// struct definition. Maybe change order of structs, organize where structs are
// declared based on when their type is used

// Tracer represents a type which can record zipkin trace data as well as
// any accompanying errors, and process that data.
type Tracer interface {
	Record(Trace) error
	Error(error)
}

// Service represents a type which can register itself with a router for
// http routing, and a Tracer for trace data collection.
type Service interface {
	Register(router *mux.Router, tracer Tracer) error
}

// BinaryAnnotation represents a zipkin binary annotation. It contains
// all of the same fields as might be found in its zipkin counterpart.
type BinaryAnnotation struct {
	Key         string
	Value       string
	Host        string // annotation.endpoint.ipv4 + ":" + annotation.endpoint.port
	ServiceName string
	Type        string
}

func (b BinaryAnnotation) ToMeta() MetaAnnotation {
	return MetaAnnotation{
		Key:         b.Key,
		Value:       b.Value,
		Host:        b.Host,
		ServiceName: b.ServiceName,
		Type:        b.Type,
	}
}

// Annotation represents an ordinary zipkin annotation. It contains the data fields
// which will become fields/tags in influxdb
type Annotation struct {
	Timestamp   time.Time
	Value       string
	Host        string // annotation.endpoint.ipv4 + ":" + annotation.endpoint.port
	ServiceName string
}

func (a Annotation) ToMeta() MetaAnnotation {
	return MetaAnnotation{
		Key:         a.Value,
		Value:       "NONE",
		Type:        "NONE",
		Timestamp:   a.Timestamp,
		Host:        a.Host,
		ServiceName: a.ServiceName,
	}
}

type MetaAnnotation struct {
	Key         string
	Value       string
	Type        string
	Timestamp   time.Time
	Host        string
	HostIPV6    string
	ServiceName string
}

//Span represents a specific zipkin span. It holds the majority of the same
// data as a zipkin span sent via the thrift protocol, but is presented in a
// format which is more straightforward for storage purposes.
type Span struct {
	ID                string
	TraceID           string // zipkin traceid high concat with traceid
	Name              string
	ParentID          string
	Timestamp         time.Time // If zipkin input is nil then time.Now()
	Duration          time.Duration
	Annotations       []Annotation
	BinaryAnnotations []BinaryAnnotation
}

// Trace is an array (or a series) of spans
type Trace []Span

//UnmarshalZipkinResponse is a helper method for unmarhsalling a slice of []*zipkincore.Spans
// into a Trace (technically a []Span)
func UnmarshalZipkinResponse(spans []*zipkincore.Span) (Trace, error) {
	var trace Trace
	for _, span := range spans {

		s := Span{}
		s.ID = strconv.FormatInt(span.GetID(), 10)
		s.TraceID = strconv.FormatInt(span.GetTraceID(), 10)
		if span.GetTraceIDHigh() != 0 {
			s.TraceID = strconv.FormatInt(span.GetTraceIDHigh(), 10) + s.TraceID
		}

		s.Annotations = UnmarshalAnnotations(span.GetAnnotations())

		var err error
		s.BinaryAnnotations, err = UnmarshalBinaryAnnotations(span.GetBinaryAnnotations())
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

// UnmarshalAnnotations is a helper method for unmarshalling a slice of
// *zipkincore.Annotation into a slice of Annotations
func UnmarshalAnnotations(annotations []*zipkincore.Annotation) []Annotation {
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

func (s Span) MergeAnnotations() {
}

// UnmarshalBinaryAnnotations is very similar to UnmarshalAnnotations, but it
// Unmarshalls zipkincore.BinaryAnnotations instead of the normal zipkincore.Annotation
func UnmarshalBinaryAnnotations(annotations []*zipkincore.BinaryAnnotation) ([]BinaryAnnotation, error) {
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

		val := annotation.GetValue()
		b.Key = annotation.GetKey()
		b.Value = string(val)
		b.Type = annotation.GetAnnotationType().String()
		formatted = append(formatted, b)
	}

	return formatted, nil
}

type LineProtocolConverter struct {
	acc telegraf.Accumulator
}

func (l *LineProtocolConverter) Record(t Trace) error {
	log.Printf("received trace: %#+v\n", t)
	//log.Printf("...But converter implementation is not yet done. Here's some example data")
	log.Printf("Writing to telegraf...\n")
	for _, s := range t {
		for _, a := range s.Annotations {
			fields := map[string]interface{}{
				// Maybe we don't need "annotation_timestamp"?
				"annotation_timestamp": a.Timestamp.Unix(),
				"duration":             s.Duration,
			}

			log.Printf("Duration is: %d", s.Duration)

			tags := map[string]string{
				"id":               s.ID,
				"parent_id":        s.ParentID,
				"trace_id":         s.TraceID,
				"name":             s.Name,
				"service_name":     a.ServiceName,
				"annotation_value": a.Value,
				"endpoint_host":    a.Host,
			}
			log.Println("adding data")
			l.acc.AddFields("zipkin", fields, tags, s.Timestamp)
		}

		for _, b := range s.BinaryAnnotations {
			fields := map[string]interface{}{
				"duration": s.Duration,
			}

			log.Printf("Duration is: %d", s.Duration)

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
			log.Printf("adding data")
			l.acc.AddFields("zipkin", fields, tags, s.Timestamp)
		}
	}

	return nil

}

func (l *LineProtocolConverter) Error(err error) {
	l.acc.AddError(err)
}

func NewLineProtocolConverter(acc telegraf.Accumulator) *LineProtocolConverter {
	return &LineProtocolConverter{
		acc: acc,
	}
}

const sampleConfig = `
  ##
  # path = /path/your/zipkin/impl/posts/to
  # port = <port_your_zipkin_impl_uses>
`

// Zipkin is a telegraf configuration structure for the zipkin input plugin,
// but it also contains fields for the management of a separate, concurrent
// zipkin http server
type Zipkin struct {
	ServiceAddress string
	Port           int
	Path           string
	tracing        Service
	server         *http.Server
	waitGroup      *sync.WaitGroup
}

// Description is a necessary method implementation from telegraf.ServiceInput
func (z Zipkin) Description() string {
	return "Allows for the collection of zipkin tracing spans for storage in InfluxDB"
}

// SampleConfig is a  necessary  method implementation from telegraf.ServiceInput
func (z Zipkin) SampleConfig() string {
	return sampleConfig
}

// Gather is empty for the zipkin plugin; all gathering is done through
// the separate goroutine launched in (*Zipkin).Start()
func (z *Zipkin) Gather(acc telegraf.Accumulator) error { return nil }

// Start launches a separate goroutine for collecting zipkin client http requests,
// passing in a telegraf.Accumulator such that data can be collected.
func (z *Zipkin) Start(acc telegraf.Accumulator) error {
	log.Println("starting...")
	if z.tracing == nil {
		t := NewServer(z.Path)
		z.tracing = t
	}

	var wg sync.WaitGroup
	z.waitGroup = &wg

	go func() {
		wg.Add(1)
		defer wg.Done()

		z.Listen(acc)
	}()

	return nil
}

// Stop shuts the internal http server down with via context.Context
func (z *Zipkin) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultShutdownTimeout)
	defer cancel()

	defer z.waitGroup.Wait()
	z.server.Shutdown(ctx)
}

// Listen creates an http server on the zipkin instance it is called with, and
// serves http until it is stopped by Zipkin's (*Zipkin).Stop()  method.
func (z *Zipkin) Listen(acc telegraf.Accumulator) {
	r := mux.NewRouter()
	converter := NewLineProtocolConverter(acc)
	z.tracing.Register(r, converter)

	if z.server == nil {
		z.server = &http.Server{
			Addr:    ":" + strconv.Itoa(z.Port),
			Handler: r,
		}
	}
	if err := z.server.ListenAndServe(); err != nil {
		acc.AddError(fmt.Errorf("E! Error listening: %v", err))
	}
}

func init() {
	inputs.Add("zipkin", func() telegraf.Input {
		return &Zipkin{
			Path: DefaultRoute,
			Port: DefaultPort,
		}
	})
}
