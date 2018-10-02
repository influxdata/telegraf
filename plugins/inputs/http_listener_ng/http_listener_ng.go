package http_listener_ng

import (
	"compress/gzip"
	"crypto/subtle"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	tlsint "github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
)

// defaultMaxBodySize is the default maximum request body size, in bytes.
// if the request body is over this size, we will return an HTTP 413 error.
// 500 MB
const defaultMaxBodySize = 500 * 1024 * 1024

type TimeFunc func() time.Time

type HTTPListenerNG struct {
	ServiceAddress string
	Path           string
	Methods        []string
	ReadTimeout    internal.Duration
	WriteTimeout   internal.Duration
	MaxBodySize    int64
	Port           int

	tlsint.ServerConfig

	BasicUsername string
	BasicPassword string

	TimeFunc

	mu sync.Mutex
	wg sync.WaitGroup

	listener net.Listener

	parsers.Parser
	acc telegraf.Accumulator
}

const sampleConfig = `
  ## Address and port to host HTTP listener on
  service_address = ":8186"

  ## Path to listen to.
  path = "/telegraf"

  ## HTTP methods to accept.
  methods = ["POST", "PUT"]

  ## maximum duration before timing out read of the request
  read_timeout = "10s"
  ## maximum duration before timing out write of the response
  write_timeout = "10s"

  ## Maximum allowed http request body size in bytes.
  ## 0 means to use the default of 536,870,912 bytes (500 mebibytes)
  max_body_size = 0

  ## Set one or more allowed client CA certificate file names to 
  ## enable mutually authenticated TLS connections
  tls_allowed_cacerts = ["/etc/telegraf/clientca.pem"]

  ## Add service certificate and key
  tls_cert = "/etc/telegraf/cert.pem"
  tls_key = "/etc/telegraf/key.pem"

  ## Optional username and password to accept for HTTP basic authentication.
  ## You probably want to make sure you have TLS configured above for this.
  basic_username = "foobar"
  basic_password = "barfoo"

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
`

func (h *HTTPListenerNG) SampleConfig() string {
	return sampleConfig
}

func (h *HTTPListenerNG) Description() string {
	return "Generic HTTP write listener"
}

func (h *HTTPListenerNG) Gather(_ telegraf.Accumulator) error {
	return nil
}

func (h *HTTPListenerNG) SetParser(parser parsers.Parser) {
	h.Parser = parser
}

// Start starts the http listener service.
func (h *HTTPListenerNG) Start(acc telegraf.Accumulator) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.MaxBodySize == 0 {
		h.MaxBodySize = defaultMaxBodySize
	}

	if h.ReadTimeout.Duration < time.Second {
		h.ReadTimeout.Duration = time.Second * 10
	}
	if h.WriteTimeout.Duration < time.Second {
		h.WriteTimeout.Duration = time.Second * 10
	}

	h.acc = acc

	tlsConf, err := h.ServerConfig.TLSConfig()
	if err != nil {
		return err
	}

	server := &http.Server{
		Addr:         h.ServiceAddress,
		Handler:      h,
		ReadTimeout:  h.ReadTimeout.Duration,
		WriteTimeout: h.WriteTimeout.Duration,
		TLSConfig:    tlsConf,
	}

	var listener net.Listener
	if tlsConf != nil {
		listener, err = tls.Listen("tcp", h.ServiceAddress, tlsConf)
	} else {
		listener, err = net.Listen("tcp", h.ServiceAddress)
	}
	if err != nil {
		return err
	}
	h.listener = listener
	h.Port = listener.Addr().(*net.TCPAddr).Port

	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		server.Serve(h.listener)
	}()

	log.Printf("I! Started HTTP listener NG service on %s\n", h.ServiceAddress)

	return nil
}

// Stop cleans up all resources
func (h *HTTPListenerNG) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.listener.Close()
	h.wg.Wait()

	log.Println("I! Stopped HTTP listener NG service on ", h.ServiceAddress)
}

func (h *HTTPListenerNG) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if req.URL.Path == h.Path {
		h.AuthenticateIfSet(h.serveWrite, res, req)
	} else {
		h.AuthenticateIfSet(http.NotFound, res, req)
	}
}

func (h *HTTPListenerNG) serveWrite(res http.ResponseWriter, req *http.Request) {
	// Check that the content length is not too large for us to handle.
	if req.ContentLength > h.MaxBodySize {
		tooLarge(res)
		return
	}

	// Check if the requested HTTP method was specified in config.
	isAcceptedMethod := false
	for _, method := range h.Methods {
		if req.Method == method {
			isAcceptedMethod = true
			break
		}
	}
	if !isAcceptedMethod {
		methodNotAllowed(res)
		return
	}

	// Handle gzip request bodies
	body := req.Body
	if req.Header.Get("Content-Encoding") == "gzip" {
		var err error
		body, err = gzip.NewReader(req.Body)
		defer body.Close()
		if err != nil {
			log.Println("D! " + err.Error())
			badRequest(res, err.Error())
			return
		}
	}

	// Add +1 for EOF
	buf := make([]byte, h.MaxBodySize+1)
	n, err := io.ReadFull(body, buf)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		log.Println("D! " + err.Error())
		// problem reading the request body
		badRequest(res, err.Error())
	} else {
		// finished reading the request body
		if err := h.parse(buf[:n]); err != nil {
			log.Println("D! " + err.Error())
			badRequest(res, err.Error())
		} else {
			res.WriteHeader(http.StatusNoContent)
		}
	}
}

func (h *HTTPListenerNG) parse(b []byte) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	metrics, err := h.Parse(b)
	if err != nil {
		return fmt.Errorf("unable to parse: %s", err.Error())
	}

	for _, m := range metrics {
		h.acc.AddFields(m.Name(), m.Fields(), m.Tags(), m.Time())
	}

	return nil
}

func tooLarge(res http.ResponseWriter) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusRequestEntityTooLarge)
	res.Write([]byte(`{"error":"http: request body too large"}`))
}

func badRequest(res http.ResponseWriter, errString string) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusBadRequest)
	if errString != "" {
		res.Write([]byte(fmt.Sprintf(`{"error":%q}`, errString)))
	} else {
		res.Write([]byte(`{"error":"http: bad request"}`))
	}
}

func methodNotAllowed(res http.ResponseWriter) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusMethodNotAllowed)
	res.Write([]byte(`{"error":"http: method not allowed"}`))
}

func (h *HTTPListenerNG) AuthenticateIfSet(handler http.HandlerFunc, res http.ResponseWriter, req *http.Request) {
	if h.BasicUsername != "" && h.BasicPassword != "" {
		reqUsername, reqPassword, ok := req.BasicAuth()
		if !ok ||
			subtle.ConstantTimeCompare([]byte(reqUsername), []byte(h.BasicUsername)) != 1 ||
			subtle.ConstantTimeCompare([]byte(reqPassword), []byte(h.BasicPassword)) != 1 {

			http.Error(res, "Unauthorized.", http.StatusUnauthorized)
			return
		}
		handler(res, req)
	} else {
		handler(res, req)
	}
}

func init() {
	parser, _ := parsers.NewInfluxParser()

	inputs.Add("http_listener_ng", func() telegraf.Input {
		return &HTTPListenerNG{
			ServiceAddress: ":8186",
			TimeFunc:       time.Now,
			Parser:         parser,
			Path:           "/telegraf",
			Methods:        []string{"POST", "PUT"},
		}
	})
}
