package zipkin

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const (
	// DefaultPort is the default port zipkin listens on, which zipkin implementations
	// expect.
	DefaultPort = 9411

	// DefaultRoute is the default route zipkin uses, and zipkin implementations
	// expect.
	DefaultRoute = "/api/v1/spans"

	// DefaultShutdownTimeout is the max amount of time telegraf will wait
	// for the plugin to shutdown
	DefaultShutdownTimeout = 5
)

// Recorder represents a type which can record zipkin trace data as well as
// any accompanying errors, and process that data.
type Recorder interface {
	Record(Trace) error
	Error(error)
}

// Handler represents a type which can register itself with a router for
// http routing, and a Recorder for trace data collection.
type Handler interface {
	Register(router *mux.Router, recorder Recorder) error
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

// Trace is an array (or a series) of spans
type Trace []Span

const sampleConfig = `
  # path = "/api/v1/spans" # URL path for span data
  # port = 9411            # Port on which Telegraf listens
`

// Zipkin is a telegraf configuration structure for the zipkin input plugin,
// but it also contains fields for the management of a separate, concurrent
// zipkin http server
type Zipkin struct {
	ServiceAddress string
	Port           int
	Path           string

	address   string
	handler   Handler
	server    *http.Server
	waitGroup *sync.WaitGroup
}

// Description is a necessary method implementation from telegraf.ServiceInput
func (z Zipkin) Description() string {
	return "This plugin implements the Zipkin http server to gather trace and timing data needed to troubleshoot latency problems in microservice architectures."
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
	z.handler = NewSpanHandler(z.Path)

	var wg sync.WaitGroup
	z.waitGroup = &wg

	router := mux.NewRouter()
	converter := NewLineProtocolConverter(acc)
	z.handler.Register(router, converter)

	z.server = &http.Server{
		Handler: router,
	}

	addr := ":" + strconv.Itoa(z.Port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	z.address = ln.Addr().String()
	log.Printf("I! Started the zipkin listener on %s", z.address)

	go func() {
		wg.Add(1)
		defer wg.Done()

		z.Listen(ln, acc)
	}()

	return nil
}

// Stop shuts the internal http server down with via context.Context
func (z *Zipkin) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultShutdownTimeout)

	defer z.waitGroup.Wait()
	defer cancel()

	z.server.Shutdown(ctx)
}

// Listen creates an http server on the zipkin instance it is called with, and
// serves http until it is stopped by Zipkin's (*Zipkin).Stop()  method.
func (z *Zipkin) Listen(ln net.Listener, acc telegraf.Accumulator) {
	if err := z.server.Serve(ln); err != nil {
		// Because of the clean shutdown in `(*Zipkin).Stop()`
		// We're expecting a server closed error at some point
		// So we don't want to display it as an error.
		// This interferes with telegraf's internal data collection,
		// by making it appear as if a serious error occurred.
		if err != http.ErrServerClosed {
			acc.AddError(fmt.Errorf("E! Error listening: %v", err))
		}
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
