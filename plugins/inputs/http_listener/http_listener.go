package http_listener

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"crypto/x509"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/selfstat"
)

const (
	// DEFAULT_MAX_BODY_SIZE is the default maximum request body size, in bytes.
	// if the request body is over this size, we will return an HTTP 413 error.
	// 500 MB
	DEFAULT_MAX_BODY_SIZE = 500 * 1024 * 1024

	// MAX_LINE_SIZE is the maximum size, in bytes, that can be allocated for
	// a single InfluxDB point.
	// 64 KB
	DEFAULT_MAX_LINE_SIZE = 64 * 1024
)

type HTTPListener struct {
	ServiceAddress string
	ReadTimeout    internal.Duration
	WriteTimeout   internal.Duration
	MaxBodySize    int64
	MaxLineSize    int
	Port           int

	TlsAllowedCacerts []string
	TlsCert           string
	TlsKey            string

	mu sync.Mutex
	wg sync.WaitGroup

	listener net.Listener

	parser parsers.Parser
	acc    telegraf.Accumulator
	pool   *pool

	BytesRecv       selfstat.Stat
	RequestsServed  selfstat.Stat
	WritesServed    selfstat.Stat
	QueriesServed   selfstat.Stat
	PingsServed     selfstat.Stat
	RequestsRecv    selfstat.Stat
	WritesRecv      selfstat.Stat
	QueriesRecv     selfstat.Stat
	PingsRecv       selfstat.Stat
	NotFoundsServed selfstat.Stat
	BuffersCreated  selfstat.Stat
}

const sampleConfig = `
  ## Address and port to host HTTP listener on
  service_address = ":8186"

  ## maximum duration before timing out read of the request
  read_timeout = "10s"
  ## maximum duration before timing out write of the response
  write_timeout = "10s"

  ## Maximum allowed http request body size in bytes.
  ## 0 means to use the default of 536,870,912 bytes (500 mebibytes)
  max_body_size = 0

  ## Maximum line size allowed to be sent in bytes.
  ## 0 means to use the default of 65536 bytes (64 kibibytes)
  max_line_size = 0

  ## Set one or more allowed client CA certificate file names to 
  ## enable mutually authenticated TLS connections
  tls_allowed_cacerts = ["/etc/telegraf/clientca.pem"]

  ## Add service certificate and key
  tls_cert = "/etc/telegraf/cert.pem"
  tls_key = "/etc/telegraf/key.pem"

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
`

func (h *HTTPListener) SampleConfig() string {
	return sampleConfig
}

func (h *HTTPListener) Description() string {
	return "Influx HTTP write listener"
}

func (h *HTTPListener) Gather(_ telegraf.Accumulator) error {
	h.BuffersCreated.Set(h.pool.ncreated())
	return nil
}

// Start starts the http listener service.
func (h *HTTPListener) Start(acc telegraf.Accumulator) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	tags := map[string]string{
		"address": h.ServiceAddress,
	}
	h.BytesRecv = selfstat.Register("http_listener", "bytes_received", tags)
	h.RequestsServed = selfstat.Register("http_listener", "requests_served", tags)
	h.WritesServed = selfstat.Register("http_listener", "writes_served", tags)
	h.QueriesServed = selfstat.Register("http_listener", "queries_served", tags)
	h.PingsServed = selfstat.Register("http_listener", "pings_served", tags)
	h.RequestsRecv = selfstat.Register("http_listener", "requests_received", tags)
	h.WritesRecv = selfstat.Register("http_listener", "writes_received", tags)
	h.QueriesRecv = selfstat.Register("http_listener", "queries_received", tags)
	h.PingsRecv = selfstat.Register("http_listener", "pings_received", tags)
	h.NotFoundsServed = selfstat.Register("http_listener", "not_founds_served", tags)
	h.BuffersCreated = selfstat.Register("http_listener", "buffers_created", tags)

	if h.MaxBodySize == 0 {
		h.MaxBodySize = DEFAULT_MAX_BODY_SIZE
	}
	if h.MaxLineSize == 0 {
		h.MaxLineSize = DEFAULT_MAX_LINE_SIZE
	}

	if h.ReadTimeout.Duration < time.Second {
		h.ReadTimeout.Duration = time.Second * 10
	}
	if h.WriteTimeout.Duration < time.Second {
		h.WriteTimeout.Duration = time.Second * 10
	}

	h.acc = acc
	h.pool = NewPool(200, h.MaxLineSize)

	tlsConf := h.getTLSConfig()

	server := &http.Server{
		Addr:         h.ServiceAddress,
		Handler:      h,
		ReadTimeout:  h.ReadTimeout.Duration,
		WriteTimeout: h.WriteTimeout.Duration,
		TLSConfig:    tlsConf,
	}

	var err error
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

	log.Printf("I! Started HTTP listener service on %s\n", h.ServiceAddress)

	return nil
}

// Stop cleans up all resources
func (h *HTTPListener) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.listener.Close()
	h.wg.Wait()

	log.Println("I! Stopped HTTP listener service on ", h.ServiceAddress)
}

func (h *HTTPListener) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	h.RequestsRecv.Incr(1)
	defer h.RequestsServed.Incr(1)
	switch req.URL.Path {
	case "/write":
		h.WritesRecv.Incr(1)
		defer h.WritesServed.Incr(1)
		h.serveWrite(res, req)
	case "/query":
		h.QueriesRecv.Incr(1)
		defer h.QueriesServed.Incr(1)
		// Deliver a dummy response to the query endpoint, as some InfluxDB
		// clients test endpoint availability with a query
		res.Header().Set("Content-Type", "application/json")
		res.Header().Set("X-Influxdb-Version", "1.0")
		res.WriteHeader(http.StatusOK)
		res.Write([]byte("{\"results\":[]}"))
	case "/ping":
		h.PingsRecv.Incr(1)
		defer h.PingsServed.Incr(1)
		// respond to ping requests
		res.WriteHeader(http.StatusNoContent)
	default:
		defer h.NotFoundsServed.Incr(1)
		// Don't know how to respond to calls to other endpoints
		http.NotFound(res, req)
	}
}

func (h *HTTPListener) serveWrite(res http.ResponseWriter, req *http.Request) {
	// Check that the content length is not too large for us to handle.
	if req.ContentLength > h.MaxBodySize {
		tooLarge(res)
		return
	}
	now := time.Now()

	precision := req.URL.Query().Get("precision")

	// Handle gzip request bodies
	body := req.Body
	if req.Header.Get("Content-Encoding") == "gzip" {
		var err error
		body, err = gzip.NewReader(req.Body)
		defer body.Close()
		if err != nil {
			log.Println("E! " + err.Error())
			badRequest(res)
			return
		}
	}
	body = http.MaxBytesReader(res, body, h.MaxBodySize)

	var return400 bool
	var hangingBytes bool
	buf := h.pool.get()
	defer h.pool.put(buf)
	bufStart := 0
	for {
		n, err := io.ReadFull(body, buf[bufStart:])
		if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
			log.Println("E! " + err.Error())
			// problem reading the request body
			badRequest(res)
			return
		}
		h.BytesRecv.Incr(int64(n))

		if err == io.EOF {
			if return400 {
				badRequest(res)
			} else {
				res.WriteHeader(http.StatusNoContent)
			}
			return
		}

		if hangingBytes {
			i := bytes.IndexByte(buf, '\n')
			if i == -1 {
				// still didn't find a newline, keep scanning
				continue
			}
			// rotate the bit remaining after the first newline to the front of the buffer
			i++ // start copying after the newline
			bufStart = len(buf) - i
			if bufStart > 0 {
				copy(buf, buf[i:])
			}
			hangingBytes = false
			continue
		}

		if err == io.ErrUnexpectedEOF {
			// finished reading the request body
			if err := h.parse(buf[:n+bufStart], now, precision); err != nil {
				log.Println("E! " + err.Error())
				return400 = true
			}
			if return400 {
				badRequest(res)
			} else {
				res.WriteHeader(http.StatusNoContent)
			}
			return
		}

		// if we got down here it means that we filled our buffer, and there
		// are still bytes remaining to be read. So we will parse up until the
		// final newline, then push the rest of the bytes into the next buffer.
		i := bytes.LastIndexByte(buf, '\n')
		if i == -1 {
			// drop any line longer than the max buffer size
			log.Printf("E! http_listener received a single line longer than the maximum of %d bytes",
				len(buf))
			hangingBytes = true
			return400 = true
			bufStart = 0
			continue
		}
		if err := h.parse(buf[:i+1], now, precision); err != nil {
			log.Println("E! " + err.Error())
			return400 = true
		}
		// rotate the bit remaining after the last newline to the front of the buffer
		i++ // start copying after the newline
		bufStart = len(buf) - i
		if bufStart > 0 {
			copy(buf, buf[i:])
		}
	}
}

func (h *HTTPListener) parse(b []byte, t time.Time, precision string) error {
	var metrics []telegraf.Metric
	var err error

	switch h.parser.(type) {
	case *influx.InfluxParser:
		metrics, err = h.parser.(*influx.InfluxParser).ParseWithDefaultTimePrecision(b, t, precision)
	default:
		metrics, err = h.parser.Parse(b)
	}

	for _, m := range metrics {
		h.acc.AddFields(m.Name(), m.Fields(), m.Tags(), m.Time())
	}

	return err
}

func (h *HTTPListener) SetParser(parser parsers.Parser) {
	h.parser = parser
}

func tooLarge(res http.ResponseWriter) {
	res.Header().Set("Content-Type", "application/json")
	res.Header().Set("X-Influxdb-Version", "1.0")
	res.WriteHeader(http.StatusRequestEntityTooLarge)
	res.Write([]byte(`{"error":"http: request body too large"}`))
}

func badRequest(res http.ResponseWriter) {
	res.Header().Set("Content-Type", "application/json")
	res.Header().Set("X-Influxdb-Version", "1.0")
	res.WriteHeader(http.StatusBadRequest)
	res.Write([]byte(`{"error":"http: bad request"}`))
}

func (h *HTTPListener) getTLSConfig() *tls.Config {
	tlsConf := &tls.Config{
		InsecureSkipVerify: false,
		Renegotiation:      tls.RenegotiateNever,
	}

	if len(h.TlsCert) == 0 || len(h.TlsKey) == 0 {
		return nil
	}

	cert, err := tls.LoadX509KeyPair(h.TlsCert, h.TlsKey)
	if err != nil {
		return nil
	}
	tlsConf.Certificates = []tls.Certificate{cert}

	if h.TlsAllowedCacerts != nil {
		tlsConf.ClientAuth = tls.RequireAndVerifyClientCert
		clientPool := x509.NewCertPool()
		for _, ca := range h.TlsAllowedCacerts {
			c, err := ioutil.ReadFile(ca)
			if err != nil {
				continue
			}
			clientPool.AppendCertsFromPEM(c)
		}
		tlsConf.ClientCAs = clientPool
	}

	return tlsConf
}

func init() {
	inputs.Add("http_listener", func() telegraf.Input {
		return &HTTPListener{
			ServiceAddress: ":8186",
		}
	})
}
