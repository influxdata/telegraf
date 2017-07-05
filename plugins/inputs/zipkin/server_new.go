package zipkin

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/gorilla/mux"
	"github.com/openzipkin/zipkin-go-opentracing/_thrift/gen-go/zipkincore"
)

// Server is an implementation of tracer which is a helper for running an
// http server which accepts zipkin requests
type Server struct {
	Path      string
	tracer    Tracer
	waitGroup *sync.WaitGroup
}

// Register allows server to implement the Service interface. Server's register metod
// registers its handler on mux, and sets the servers tracer with tracer
func (s *Server) Register(router *mux.Router, tracer Tracer) error {
	// TODO: potentially move router into Server if appropriate
	router.HandleFunc(s.Path, s.SpanHandler).Methods("POST")
	s.tracer = tracer
	return nil
}

// SpanHandler is the handler Server uses for handling zipkin POST requests
func (s *Server) SpanHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request from: %s", r.URL.String())
	log.Printf("Raw request data is: %#+v", r)
	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		e := fmt.Errorf("Encountered error: %s", err)
		log.Println(e)
		s.tracer.Error(e)
		//TODO: Change http status that is sent back to client
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

//NewServer returns a new server instance given path to handle
func NewServer(path string) *Server {
	return &Server{
		Path: path,
	}
}
