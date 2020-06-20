package http_listener_v2

import (
	"compress/gzip"
	"crypto/subtle"
	"crypto/tls"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
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

const (
	body  = "body"
	query = "query"
)

// TimeFunc provides a timestamp for the metrics
type TimeFunc func() time.Time

// HTTPListenerV2 is an input plugin that collects external metrics sent via HTTP
type HTTPListenerV2 struct {
	ServiceAddress string            `toml:"service_address"`
	Path           string            `toml:"path"`
	Methods        []string          `toml:"methods"`
	DataSource     string            `toml:"data_source"`
	ReadTimeout    internal.Duration `toml:"read_timeout"`
	WriteTimeout   internal.Duration `toml:"write_timeout"`
	MaxBodySize    internal.Size     `toml:"max_body_size"`
	Port           int               `toml:"port"`
	BasicUsername  string            `toml:"basic_username"`
	BasicPassword  string            `toml:"basic_password"`
	HTTPHeaderTags map[string]string `toml:"http_header_tags"`
	tlsint.ServerConfig

	TimeFunc
	Log telegraf.Logger

	wg sync.WaitGroup

	listener net.Listener

	parsers.Parser
	acc telegraf.Accumulator
}

const sampleConfig = `
  ## Address and port to host HTTP listener on
  service_address = ":8080"

  ## Path to listen to.
  # path = "/telegraf"

  ## HTTP methods to accept.
  # methods = ["POST", "PUT"]

  ## maximum duration before timing out read of the request
  # read_timeout = "10s"
  ## maximum duration before timing out write of the response
  # write_timeout = "10s"

  ## Maximum allowed http request body size in bytes.
  ## 0 means to use the default of 524,288,00 bytes (500 mebibytes)
  # max_body_size = "500MB"

  ## Part of the request to consume.  Available options are "body" and
  ## "query".
  # data_source = "body"

  ## Set one or more allowed client CA certificate file names to
  ## enable mutually authenticated TLS connections
  # tls_allowed_cacerts = ["/etc/telegraf/clientca.pem"]

  ## Add service certificate and key
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"

  ## Optional username and password to accept for HTTP basic authentication.
  ## You probably want to make sure you have TLS configured above for this.
  # basic_username = "foobar"
  # basic_password = "barfoo"

  ## Optional setting to map http headers into tags
  ## If the http header is not present on the request, no corresponding tag will be added
  ## If multiple instances of the http header are present, only the first value will be used
  # http_header_tags = {"HTTP_HEADER" = "TAG_NAME"}

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
`

func (h *HTTPListenerV2) SampleConfig() string {
	return sampleConfig
}

func (h *HTTPListenerV2) Description() string {
	return "Generic HTTP write listener"
}

func (h *HTTPListenerV2) Gather(_ telegraf.Accumulator) error {
	return nil
}

func (h *HTTPListenerV2) SetParser(parser parsers.Parser) {
	h.Parser = parser
}

// Start starts the http listener service.
func (h *HTTPListenerV2) Start(acc telegraf.Accumulator) error {
	if h.MaxBodySize.Size == 0 {
		h.MaxBodySize.Size = defaultMaxBodySize
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

	h.Log.Infof("Listening on %s", listener.Addr().String())

	return nil
}

// Stop cleans up all resources
func (h *HTTPListenerV2) Stop() {
	h.listener.Close()
	h.wg.Wait()
}

func (h *HTTPListenerV2) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	handler := h.serveWrite

	if req.URL.Path != h.Path {
		handler = http.NotFound
	}

	h.authenticateIfSet(handler, res, req)
}

func (h *HTTPListenerV2) serveWrite(res http.ResponseWriter, req *http.Request) {
	// Check that the content length is not too large for us to handle.
	if req.ContentLength > h.MaxBodySize.Size {
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

	var bytes []byte
	var ok bool

	switch strings.ToLower(h.DataSource) {
	case query:
		bytes, ok = h.collectQuery(res, req)
	default:
		bytes, ok = h.collectBody(res, req)
	}

	if !ok {
		return
	}

	metrics, err := h.Parse(bytes)
	if err != nil {
		h.Log.Debugf("Parse error: %s", err.Error())
		badRequest(res)
		return
	}

	for _, m := range metrics {
		for headerName, measurementName := range h.HTTPHeaderTags {
			headerValues, foundHeader := req.Header[headerName]
			if foundHeader && len(headerValues) > 0 {
				m.AddTag(measurementName, headerValues[0])
			}
		}

		h.acc.AddMetric(m)
	}

	res.WriteHeader(http.StatusNoContent)
}

func (h *HTTPListenerV2) collectBody(res http.ResponseWriter, req *http.Request) ([]byte, bool) {
	body := req.Body

	// Handle gzip request bodies
	if req.Header.Get("Content-Encoding") == "gzip" {
		var err error
		body, err = gzip.NewReader(req.Body)
		if err != nil {
			h.Log.Debug(err.Error())
			badRequest(res)
			return nil, false
		}
		defer body.Close()
	}

	body = http.MaxBytesReader(res, body, h.MaxBodySize.Size)
	bytes, err := ioutil.ReadAll(body)
	if err != nil {
		tooLarge(res)
		return nil, false
	}

	return bytes, true
}

func (h *HTTPListenerV2) collectQuery(res http.ResponseWriter, req *http.Request) ([]byte, bool) {
	rawQuery := req.URL.RawQuery

	query, err := url.QueryUnescape(rawQuery)
	if err != nil {
		h.Log.Debugf("Error parsing query: %s", err.Error())
		badRequest(res)
		return nil, false
	}

	return []byte(query), true
}

func tooLarge(res http.ResponseWriter) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusRequestEntityTooLarge)
	res.Write([]byte(`{"error":"http: request body too large"}`))
}

func methodNotAllowed(res http.ResponseWriter) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusMethodNotAllowed)
	res.Write([]byte(`{"error":"http: method not allowed"}`))
}

func internalServerError(res http.ResponseWriter) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusInternalServerError)
}

func badRequest(res http.ResponseWriter) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusBadRequest)
	res.Write([]byte(`{"error":"http: bad request"}`))
}

func (h *HTTPListenerV2) authenticateIfSet(handler http.HandlerFunc, res http.ResponseWriter, req *http.Request) {
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
	inputs.Add("http_listener_v2", func() telegraf.Input {
		return &HTTPListenerV2{
			ServiceAddress: ":8080",
			TimeFunc:       time.Now,
			Path:           "/telegraf",
			Methods:        []string{"POST", "PUT"},
			DataSource:     body,
		}
	})
}
