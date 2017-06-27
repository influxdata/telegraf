package zipkin

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/openzipkin/zipkin-go-opentracing/_thrift/gen-go/zipkincore"
	"github.com/tylerb/graceful"
)

const version = "1.0"
const port = "9411"

type Server struct {
	errorChan chan error
	dataChan  chan SpanData
	Port      string
	Done      chan struct{}
	listener  net.Listener
}

type SpanData []*zipkincore.Span

func NewHTTPServer(port int, e chan error, d chan SpanData, f chan struct{}) *Server {
	return &Server{
		errorChan: e,
		dataChan:  d,
		Port:      strconv.Itoa(port),
		Done:      f,
	}
}

// Version adds a version header to response
func Version(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("X-Proxy-Version", version)
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// Logger is middleware that logs the request
func Logger(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received request from url %s\n", r.URL.String())
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// MainHandler returns a closure with access to a *Server pointer
// for use as an http server handler
func (s *Server) MainHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		log.Printf("Received request from: %s", r.URL.String())
		log.Printf("Raw request data is: %#+v", r)
		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			e := fmt.Errorf("Encoutered error: %s", err)
			log.Println(e)
			s.errorChan <- e
		}
		buffer := thrift.NewTMemoryBuffer()
		if _, err = buffer.Write(body); err != nil {
			log.Println(err)
			s.errorChan <- err
		}
		transport := thrift.NewTBinaryProtocolTransport(buffer)
		_, size, err := transport.ReadListBegin()
		if err != nil {
			log.Printf("%s", err)
			s.errorChan <- err
			return
		}
		var spans []*zipkincore.Span
		for i := 0; i < size; i++ {
			zs := &zipkincore.Span{}
			if err = zs.Read(transport); err != nil {
				log.Printf("%s", err)
				s.errorChan <- err
				return
			}
			spans = append(spans, zs)
		}
		err = transport.ReadListEnd()
		if err != nil {
			log.Printf("%s", err)
			s.errorChan <- err
			return
		}
		out, _ := json.MarshalIndent(spans, "", "    ")
		log.Println(string(out))
		s.dataChan <- SpanData(spans)
		w.WriteHeader(http.StatusNoContent)
	}
	return http.HandlerFunc(fn)
}

func (s *Server) HandleZipkinRequests() {
	log.Printf("Starting zipkin HTTP server on %s\n", s.Port)
	mux := http.NewServeMux()
	// The func MainHandler returns has been closure-ified
	mux.Handle("/api/v1/spans", s.MainHandler())

	listener, err := net.Listen("tcp", ":"+s.Port)
	if err != nil {
		//Print error out for debugging purposes
		e := fmt.Errorf("Error listening on port 9411 %v", err)
		//Send error through channel to caller
		s.errorChan <- e
		return
	}

	s.addListener(listener)
	go s.ListenForStop()

	httpServer := &graceful.Server{Server: new(http.Server)}
	httpServer.SetKeepAlivesEnabled(true)
	httpServer.TCPKeepAlive = 5 * time.Second
	httpServer.Handler = Version(Logger(mux))
	log.Fatal(httpServer.Serve(listener))
}

func (s *Server) addListener(l net.Listener) {
	s.listener = l
}

func (s *Server) ListenForStop() {
	if s.listener == nil {
		log.Fatal("Listen called without listener instance")
		return
	}
	select {
	case <-s.Done:
		log.Println("closing up listener...")
		s.listener.Close()
		return
	}
}

func (s *Server) CloseAllChannels() {
	log.Printf("Closing all communication channels...\n")
	close(s.dataChan)
	close(s.errorChan)
	close(s.Done)
}

/*func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		server()
	}()
	wg.Wait()
}*/
