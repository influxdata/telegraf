package zipkin

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/gorilla/mux"
	"github.com/influxdata/telegraf"
	"github.com/openzipkin/zipkin-go-opentracing/_thrift/gen-go/zipkincore"
)

type BinaryAnnotation struct {
	//stuff
}

type Annotation struct {
	Timestamp   time.Time
	Value       string
	Host        string // annotation.endpoint.ipv4 + ":" + annotation.endpoint.port
	ServiceName string
}

type Span struct {
	ID          string // zipkin traceid high concat with traceid
	Name        string
	ParentID    *int64
	Timestamp   time.Time // If zipkin input is nil then time.Now()
	Duration    *int64
	TraceIDHigh *int64
	Annotations []Annotation
}

type Trace []Span

type Tracer interface {
	Record(Trace) error
	Error(error)
}

type Service interface {
	Register(router *mux.Router, tracer Tracer)
}

type Zipkin struct {
	ServiceAddress string
	Path           string
	tracing        Service
	server         *http.Server
}

type Server struct {
	Path      string
	Port      string
	tracer    Tracer
	waitGroup *sync.WaitGroup
}

func NewServer(path string, port int) *Server {
	return &Server{
		Path: path,
		Port: strconv.Itoa(port),
	}
}

func (s *Server) Register(router *mux.Router, tracer Tracer) error {
	router.HandleFunc(s.Path, s.SpanHandler).Methods("POST")
	s.tracer = tracer
	return nil
}

func UnmarshalZipkinResponse(spans []*zipkincore.Span) (Trace, error) {
	var trace Trace
	for _, span := range spans {

		s := &Span{}
		s.ID = strconv.FormatInt(span.GetID(), 10)
		s.Annotations = UnmarshalAnnotations(span.GetAnnotations())
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
		if duration != 0 {
			s.Duration = &duration
		}

		parentID := span.GetParentID()
		if parentID != 0 {
			s.ID += strconv.FormatInt(parentID, 10)
			s.ParentID = &parentID
		}

		traceIDHIGH := span.GetTraceIDHigh()
		if traceIDHIGH != 0 {
			s.TraceIDHigh = &traceIDHIGH
		}
	}

	return trace, nil
}

func UnmarshalAnnotations(annotations []*zipkincore.Annotation) []Annotation {
	var formatted []Annotation
	for _, annotation := range annotations {
		a := Annotation{}
		endpoint := annotation.GetHost()
		if endpoint != nil {
			a.Host = strconv.Itoa(int(endpoint.GetIpv4()))
			a.ServiceName = endpoint.GetServiceName()
		} else {
			a.Host, a.ServiceName = "", ""
		}

		a.Timestamp = time.Unix(annotation.GetTimestamp(), 0)
		a.Value = annotation.GetValue()
		formatted = append(formatted, a)
	}
	return formatted
}

func (s *Server) SpanHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request from: %s", r.URL.String())
	log.Printf("Raw request data is: %#+v", r)
	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		e := fmt.Errorf("Encoutered error: %s", err)
		log.Println(e)
		s.tracer.Error(e)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	buffer := thrift.NewTMemoryBuffer()
	if _, err = buffer.Write(body); err != nil {
		log.Println(err)
		s.tracer.Error(err)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	transport := thrift.NewTBinaryProtocolTransport(buffer)
	_, size, err := transport.ReadListBegin()
	if err != nil {
		log.Printf("%s", err)
		s.tracer.Error(err)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var spans []*zipkincore.Span
	for i := 0; i < size; i++ {
		zs := &zipkincore.Span{}
		if err = zs.Read(transport); err != nil {
			log.Printf("%s", err)
			s.tracer.Error(err)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		spans = append(spans, zs)
	}

	err = transport.ReadListEnd()
	if err != nil {
		log.Printf("%s", err)
		s.tracer.Error(err)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	//marshal json for debugging purposes
	out, _ := json.MarshalIndent(spans, "", "    ")
	log.Println(string(out))

	trace, err := UnmarshalZipkinResponse(spans)
	if err != nil {
		log.Println("Error: ", err)
		s.tracer.Error(err)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err = s.tracer.Record(trace); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

const sampleConfig = `
  ##
  # field = value
`

func (z Zipkin) Description() string {
	return "Allows for the collection of zipkin tracing spans for storage in influxdb"
}

func (z Zipkin) SampleConfig() string {
	return sampleConfig
}

func (z *Zipkin) Gather(acc telegraf.Accumulator) {

}

func (z *Zipkin) Start(acc telegraf.Accumulator) {
	//start collecting data
}

func (z *Zipkin) Stop() {
	//clean up any channels, etc.
}
