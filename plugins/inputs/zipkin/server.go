package zipkin

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	//TODO: must add below deps to Godeps file
	"github.com/apache/thrift/lib/go/thrift"
	"github.com/openzipkin/zipkin-go-opentracing/_thrift/gen-go/zipkincore"
)

const version = "1.0"
const port = "9411"

// Zipkin writes to this route by default
const DefaultRoute = "/api/v1/spans"

// Provides a shutdown timout in order to gracefully shutdown our http server
const DefaultShutdownTimeout = 5 * time.Second

type ZipkinServer struct {
	errorChan chan error
	dataChan  chan SpanData
	Port      string
	//Done       chan struct{}
	//listener   net.Listener
	HTTPServer *http.Server
	logger     *log.Logger
}

// SpanData is an alias for a slice of references to zipkincore.Span
// created for better abstraction of internal package types
type SpanData []*zipkincore.Span

// NewHTTPServer creates a new Zipkin http server given a port and a set of
// channels
func NewHTTPServer(port int, e chan error, d chan SpanData) *ZipkinServer {
	logger := log.New(os.Stdout, "", 0)
	return &ZipkinServer{
		errorChan: e,
		dataChan:  d,
		Port:      strconv.Itoa(port),
		logger:    logger,
	}
}

// Version adds a version header to response
// Delete later
func Version(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("X-Proxy-Version", version)
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// Logger is middleware that logs the request
// delete later, re-implement in a better way
// inspired by the httptrace package
func Logger(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received request from url %s\n", r.URL.String())
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// MainHandler returns a closure with access to a *ZipkinServer pointer
// for use as an http ZipkinServer handler
func (s *ZipkinServer) MainHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		s.logger.Printf("Received request from: %s", r.URL.String())
		s.logger.Printf("Raw request data is: %#+v", r)
		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			e := fmt.Errorf("Encoutered error: %s", err)
			s.logger.Println(e)
			s.errorChan <- e
			w.WriteHeader(http.StatusNoContent)
			return
		}
		buffer := thrift.NewTMemoryBuffer()
		if _, err = buffer.Write(body); err != nil {
			s.logger.Println(err)
			s.errorChan <- err
			w.WriteHeader(http.StatusNoContent)
			return
		}
		transport := thrift.NewTBinaryProtocolTransport(buffer)
		_, size, err := transport.ReadListBegin()
		if err != nil {
			s.logger.Printf("%s", err)
			s.errorChan <- err
			w.WriteHeader(http.StatusNoContent)
			return
		}
		var spans []*zipkincore.Span
		for i := 0; i < size; i++ {
			zs := &zipkincore.Span{}
			if err = zs.Read(transport); err != nil {
				s.logger.Printf("%s", err)
				s.errorChan <- err
				w.WriteHeader(http.StatusNoContent)
				return
			}
			spans = append(spans, zs)
		}
		err = transport.ReadListEnd()
		if err != nil {
			s.logger.Printf("%s", err)
			s.errorChan <- err
			w.WriteHeader(http.StatusNoContent)
			return
		}
		out, _ := json.MarshalIndent(spans, "", "    ")
		s.logger.Println(string(out))
		s.dataChan <- SpanData(spans)
		w.WriteHeader(http.StatusNoContent)
	}
	return http.HandlerFunc(fn)
}

// Serve creates an internal http.Server and calls its ListenAndServe method to
// start serving. It uses the MainHandler as the route handler.
func (s *ZipkinServer) Serve() {
	mux := http.NewServeMux()
	mux.Handle(DefaultRoute, s.MainHandler())

	s.HTTPServer = &http.Server{
		Addr:    ":" + s.Port,
		Handler: mux,
	}

	s.logger.Printf("Starting zipkin HTTP server on %s\n", s.Port)
	s.logger.Fatal(s.HTTPServer.ListenAndServe())
}

// Shutdown gracefully shuts down the internal http server
func (s *ZipkinServer) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultShutdownTimeout)
	defer cancel()

	s.HTTPServer.Shutdown(ctx)
}

// HandleZipkinRequests starts a zipkin http server on the port specified
// wthin the *Server it is called on. It receives data from zipkin, and sends
// it back to the caller over the channels the caller constructed the *Server
// with

/*func (s *Server) HandleZipkinRequests() {
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
	//TODO: put a sync group around this ListenForStop()
	// create wait group; add to wait group (wg.Add(1))
	// pass in to ListenForStop()
	//wg.Add()
	//go func(){
	go s.ListenForStop()
	//wg.Done()
	//}()

	// TODO: don't need to use graceful anymore in go 1.8 (there is graceful Server
	// shutdown)
	httpServer := &graceful.Server{Server: new(http.Server)}
	httpServer.SetKeepAlivesEnabled(true)
	httpServer.TCPKeepAlive = 5 * time.Second
	httpServer.Handler = Version(Logger(mux))

	log.Fatal(httpServer.Serve(listener))

}*/

/*func (s *Server) addListener(l net.Listener) {
	s.listener = l
}*/

// ListenForStop selects over the Server.Done channel, and stops the
// server's internal net.Listener when a singnal is received.
/*func (s *Server) ListenForStop() {
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
}*/

// CloseAllChannels closes the Server's communication channels on the server's end.
func (s *ZipkinServer) CloseAllChannels() {
	log.Printf("Closing all communication channels...\n")
	close(s.dataChan)
	close(s.errorChan)
	//close(s.Done)
}
