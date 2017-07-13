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
	router.HandleFunc(s.Path, s.SpanHandler).Methods("POST")
	s.tracer = tracer
	return nil
}

func unmarshalThrift(body []byte) ([]*zipkincore.Span, error) {
	buffer := thrift.NewTMemoryBuffer()
	if _, err := buffer.Write(body); err != nil {
		log.Println("Error in buffer write: ", err)
		return nil, err
	}

	transport := thrift.NewTBinaryProtocolTransport(buffer)
	_, size, err := transport.ReadListBegin()
	if err != nil {
		log.Printf("Error in ReadListBegin: %s", err)
		return nil, err
	}
	var spans []*zipkincore.Span
	for i := 0; i < size; i++ {
		zs := &zipkincore.Span{}
		if err = zs.Read(transport); err != nil {
			log.Printf("Error reading into zipkin struct: %s", err)
			return nil, err
		}
		spans = append(spans, zs)
	}

	err = transport.ReadListEnd()
	if err != nil {
		log.Printf("%s", err)
		return nil, err
	}
	return spans, nil
}

// SpanHandler is the handler Server uses for handling zipkin POST requests
func (s *Server) SpanHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request from: %s", r.URL.String())
	log.Printf("Raw request data is: %#+v", r)
	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)

	/*f, err := os.OpenFile("plugins/inputs/zipkin/testdata/file.dat", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
	defer f.Close()
	_, err = f.Write(body)
	if err != nil {
		log.Printf("Could not write to data file")
	}*/

	if err != nil {
		e := fmt.Errorf("Encountered error reading: %s", err)
		log.Println(e)
		s.tracer.Error(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	spans, err := unmarshalThrift(body)
	if err != nil {
		log.Println("Error: ", err)
		s.tracer.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//marshal json for debugging purposes
	out, _ := json.MarshalIndent(spans, "", "    ")
	log.Println(string(out))

	/*	f, err = os.OpenFile("plugins/inputs/zipkin/testdata/json/file.json", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
		defer f.Close()
		_, err = f.Write(out)
		if err != nil {
			log.Printf("Could not write to data file")
		}*/

	trace, err := UnmarshalZipkinResponse(spans)
	if err != nil {
		log.Println("Error: ", err)
		s.tracer.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
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
