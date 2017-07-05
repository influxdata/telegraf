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

// Annotation represents an ordinary zipkin annotation. It contains the data fields
// which will become fields/tags in influxdb
type Annotation struct {
	Timestamp   time.Time
	Value       string
	Host        string // annotation.endpoint.ipv4 + ":" + annotation.endpoint.port
	ServiceName string
}

//Span represents a specific zipkin span. It holds the majority of the same
// data as a zipkin span sent via the thrift protocol, but is presented in a
// format which is more straightforward for storage purposes.
type Span struct {
	ID                string // zipkin traceid high concat with traceid
	Name              string
	ParentID          *int64
	Timestamp         time.Time // If zipkin input is nil then time.Now()
	Duration          *int64
	TraceIDHigh       *int64
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
			s.Timestamp = time.Unix(span.GetTimestamp(), 0)
		}

		//duration, parent id, and trace id high are all optional fields.
		// below, we check to see if any of these fields are non-zero, and if they are,
		// we set the repsective fields in our Span structure to the address of
		// these values. Otherwise, we just leave them as nil

		duration := span.GetDuration()
		fmt.Println("Duration: ", duration)
		if duration != 0 {
			s.Duration = &duration
		}

		parentID := span.GetParentID()
		fmt.Println("Parent ID: ", parentID)
		if parentID != 0 {
			s.ParentID = &parentID
		}

		traceIDHIGH := span.GetTraceIDHigh()
		fmt.Println("Trace id high: ", traceIDHIGH)
		if traceIDHIGH != 0 {
			s.ID += strconv.FormatInt(traceIDHIGH, 10)
			s.TraceIDHigh = &traceIDHIGH
		}
		fmt.Println("ID:", s.ID)
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
	fmt.Println("formatted annotations: ", formatted)
	return formatted
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

			fmt.Println("Binary Annotation host and service name: ", b.Host, b.ServiceName)
		} else {
			b.Host, b.ServiceName = "", ""
		}

		val := annotation.GetValue()
		/*log.Println("Value: ", string(val))
		dst := make([]byte, base64.StdEncoding.DecodedLen(len(val)))
		_, err := base64.StdEncoding.Decode(dst, annotation.GetValue())
		if err != nil {
			return nil, err
		}*/

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
	log.Printf("...But converter implementation is not yet done. Here's some example data")

	fields := map[string]interface{}{
		"Duration":           "1060",
		"Timestamp":          time.Unix(1498852876, 0),
		"Annotations":        []string{"An annotation"},
		"binary_annotations": []string{"A binary annotation"},
	}

	tags := map[string]string{
		"host": "http://hostname.com",
		"port": "5555",
	}

	l.acc.AddFields("zipkin", fields, tags)
	return nil
}

func (l *LineProtocolConverter) Error(err error) {

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

// Stop shuts the internal http server down with via context.Context
func (z *Zipkin) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultShutdownTimeout)
	defer cancel()

	defer z.waitGroup.Wait()
	z.server.Shutdown(ctx)
}

func init() {
	inputs.Add("zipkin", func() telegraf.Input {
		return &Zipkin{
			Path: DefaultRoute,
			Port: DefaultPort,
		}
	})
}
