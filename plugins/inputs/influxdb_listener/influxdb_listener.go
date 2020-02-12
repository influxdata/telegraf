package influxdb_listener

import (
	"compress/gzip"
	"crypto/subtle"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	tlsint "github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/selfstat"
)

const (
	// DEFAULT_MAX_BODY_SIZE is the default maximum request body size, in bytes.
	// if the request body is over this size, we will return an HTTP 413 error.
	// 500 MB
	DEFAULT_MAX_BODY_SIZE = 500 * 1024 * 1024
)

type InfluxDBListener struct {
	ServiceAddress string `toml:"service_address"`
	// Port gets pulled out of ServiceAddress
	Port int
	tlsint.ServerConfig

	ReadTimeout   internal.Duration `toml:"read_timeout"`
	WriteTimeout  internal.Duration `toml:"write_timeout"`
	MaxBodySize   internal.Size     `toml:"max_body_size"`
	BasicUsername string            `toml:"basic_username"`
	BasicPassword string            `toml:"basic_password"`
	DatabaseTag   string            `toml:"database_tag"`

	TimeFunc influx.TimeFunc

	mu sync.Mutex
	wg sync.WaitGroup

	listener net.Listener

	acc telegraf.Accumulator

	bytesRecv       selfstat.Stat
	requestsServed  selfstat.Stat
	writesServed    selfstat.Stat
	queriesServed   selfstat.Stat
	pingsServed     selfstat.Stat
	requestsRecv    selfstat.Stat
	notFoundsServed selfstat.Stat
	buffersCreated  selfstat.Stat
	authFailures    selfstat.Stat

	Log telegraf.Logger

	mux http.ServeMux
}

const sampleConfig = `
  ## Address and port to host InfluxDB listener on
  service_address = ":8186"

  ## maximum duration before timing out read of the request
  read_timeout = "10s"
  ## maximum duration before timing out write of the response
  write_timeout = "10s"

  ## Maximum allowed HTTP request body size in bytes.
  ## 0 means to use the default of 524,288,000 bytes (500 mebibytes)
  max_body_size = "500MiB"


  ## Optional tag name used to store the database. 
  ## If the write has a database in the query string then it will be kept in this tag name.
  ## This tag can be used in downstream outputs.
  ## The default value of nothing means it will be off and the database will not be recorded.
  # database_tag = ""

  ## Set one or more allowed client CA certificate file names to
  ## enable mutually authenticated TLS connections
  tls_allowed_cacerts = ["/etc/telegraf/clientca.pem"]

  ## Add service certificate and key
  tls_cert = "/etc/telegraf/cert.pem"
  tls_key = "/etc/telegraf/key.pem"

  ## Optional username and password to accept for HTTP basic authentication.
  ## You probably want to make sure you have TLS configured above for this.
  # basic_username = "foobar"
  # basic_password = "barfoo"
`

func (h *InfluxDBListener) SampleConfig() string {
	return sampleConfig
}

func (h *InfluxDBListener) Description() string {
	return "InfluxDB listener"
}

func (h *InfluxDBListener) Gather(_ telegraf.Accumulator) error {
	return nil
}

func (h *InfluxDBListener) routes() {
	h.mux.HandleFunc("/write", h.handleAuth(h.handleWrite()))
	h.mux.HandleFunc("/query", h.handleAuth(h.handleQuery()))
	h.mux.HandleFunc("/ping", h.handlePing())
	h.mux.HandleFunc("/", h.handleAuth(h.handleDefault()))
}

func (h *InfluxDBListener) Init() error {
	tags := map[string]string{
		"address": h.ServiceAddress,
	}
	h.bytesRecv = selfstat.Register("influxdb_listener", "bytes_received", tags)
	h.requestsServed = selfstat.Register("influxdb_listener", "requests_served", tags)
	h.writesServed = selfstat.Register("influxdb_listener", "writes_served", tags)
	h.queriesServed = selfstat.Register("influxdb_listener", "queries_served", tags)
	h.pingsServed = selfstat.Register("influxdb_listener", "pings_served", tags)
	h.requestsRecv = selfstat.Register("influxdb_listener", "requests_received", tags)
	h.notFoundsServed = selfstat.Register("influxdb_listener", "not_founds_served", tags)
	h.buffersCreated = selfstat.Register("influxdb_listener", "buffers_created", tags)
	h.authFailures = selfstat.Register("influxdb_listener", "auth_failures", tags)
	h.routes()

	if h.MaxBodySize.Size == 0 {
		h.MaxBodySize.Size = DEFAULT_MAX_BODY_SIZE
	}

	if h.ReadTimeout.Duration < time.Second {
		h.ReadTimeout.Duration = time.Second * 10
	}
	if h.WriteTimeout.Duration < time.Second {
		h.WriteTimeout.Duration = time.Second * 10
	}

	return nil
}

// Start starts the InfluxDB listener service.
func (h *InfluxDBListener) Start(acc telegraf.Accumulator) error {
	h.mu.Lock()
	defer h.mu.Unlock()

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

	h.Log.Infof("Started HTTP listener service on %s", h.ServiceAddress)

	return nil
}

// Stop cleans up all resources
func (h *InfluxDBListener) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.listener.Close()
	h.wg.Wait()

	h.Log.Infof("Stopped HTTP listener service on %s", h.ServiceAddress)
}

func (h *InfluxDBListener) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	h.requestsRecv.Incr(1)
	h.mux.ServeHTTP(res, req)
	h.requestsServed.Incr(1)
}

func (h *InfluxDBListener) handleQuery() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		defer h.queriesServed.Incr(1)
		// Deliver a dummy response to the query endpoint, as some InfluxDB
		// clients test endpoint availability with a query
		res.Header().Set("Content-Type", "application/json")
		res.Header().Set("X-Influxdb-Version", "1.0")
		res.WriteHeader(http.StatusOK)
		res.Write([]byte("{\"results\":[]}"))
	}
}

func (h *InfluxDBListener) handlePing() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		defer h.pingsServed.Incr(1)
		verbose := req.URL.Query().Get("verbose")

		// respond to ping requests
		if verbose != "" && verbose != "0" && verbose != "false" {
			res.WriteHeader(http.StatusOK)
			b, _ := json.Marshal(map[string]string{"version": "1.0"}) // based on header set above
			res.Write(b)
		} else {
			res.WriteHeader(http.StatusNoContent)
		}
	}
}

func (h *InfluxDBListener) handleDefault() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		defer h.notFoundsServed.Incr(1)
		http.NotFound(res, req)
	}
}

func (h *InfluxDBListener) handleWrite() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		defer h.writesServed.Incr(1)
		// Check that the content length is not too large for us to handle.
		if req.ContentLength > h.MaxBodySize.Size {
			tooLarge(res)
			return
		}

		db := req.URL.Query().Get("db")

		// Handle gzip request bodies
		body := req.Body
		if req.Header.Get("Content-Encoding") == "gzip" {
			var err error
			body, err = gzip.NewReader(req.Body)
			if err != nil {
				h.Log.Debug(err.Error())
				badRequest(res, err.Error())
				return
			}
			defer body.Close()
		}
		body = http.MaxBytesReader(res, body, h.MaxBodySize.Size)

		parser := influx.NewStreamParser(body)
		parser.SetTimeFunc(h.TimeFunc)

		var err error
		precisionStr := req.URL.Query().Get("precision")
		if precisionStr != "" {
			var precision time.Duration
			precision, err = getPrecisionMultiplier(precisionStr)
			if err != nil {
				h.Log.Debug(err.Error())
				badRequest(res, err.Error())
			}
			parser.SetTimePrecision(precision)
		}

		var m telegraf.Metric
		var partial bool
		for {
			m, err = parser.Next()
			h.bytesRecv.Incr(int64(parser.Position())) //todo incr by position difference

			//continue parsing metrics even if some are malformed
			if _, ok := err.(*influx.ParseError); ok {
				//todo save error for X-Influxdb-Error header
				partial = true
				continue
			} else if err != nil {
				break
			}

			if h.DatabaseTag != "" && db != "" {
				m.AddTag(h.DatabaseTag, db)
			}
			h.acc.AddMetric(m)
		}
		if err.Error() != "EOF" {
			h.Log.Debug(err.Error())
			// problem reading the request body
			badRequest(res, err.Error())
			return
		}
		if partial {
			res.WriteHeader(http.StatusBadRequest)
		} else {
			res.WriteHeader(http.StatusNoContent)
		}
	}
}

func tooLarge(res http.ResponseWriter) {
	res.Header().Set("Content-Type", "application/json")
	res.Header().Set("X-Influxdb-Version", "1.0")
	res.Header().Set("X-Influxdb-Error", "http: request body too large")
	res.WriteHeader(http.StatusRequestEntityTooLarge)
	res.Write([]byte(`{"error":"http: request body too large"}`))
}

func badRequest(res http.ResponseWriter, errString string) {
	res.Header().Set("Content-Type", "application/json")
	res.Header().Set("X-Influxdb-Version", "1.0")
	if errString == "" {
		errString = "http: bad request"
	}
	res.Header().Set("X-Influxdb-Error", errString)
	res.WriteHeader(http.StatusBadRequest)
	res.Write([]byte(fmt.Sprintf(`{"error":%q}`, errString)))
}

func partialWrite(res http.ResponseWriter, errString string) {
	res.Header().Set("Content-Type", "application/json")
	res.Header().Set("X-Influxdb-Version", "1.0")
	res.Header().Set("X-Influxdb-Error", errString)
	res.WriteHeader(http.StatusBadRequest)
	res.Write([]byte(fmt.Sprintf(`{"error":%q}`, errString)))
}

func (h *InfluxDBListener) handleAuth(f http.HandlerFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("WWW-Authenticate", "Basic realm=\"Restricted\"")

		if h.BasicUsername != "" && h.BasicPassword != "" {
			reqUsername, reqPassword, ok := req.BasicAuth()
			if !ok ||
				subtle.ConstantTimeCompare([]byte(reqUsername), []byte(h.BasicUsername)) != 1 ||
				subtle.ConstantTimeCompare([]byte(reqPassword), []byte(h.BasicPassword)) != 1 {

				h.authFailures.Incr(1)
				http.Error(res, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}

		f(res, req)
	}
}

func getPrecisionMultiplier(precision string) (time.Duration, error) {
	d := time.Nanosecond
	switch precision {
	case "u":
		d = time.Microsecond
	case "ms":
		d = time.Millisecond
	case "s":
		d = time.Second
	case "m":
		d = time.Minute
	case "h":
		d = time.Hour
	default:
		return d, errors.New("unknown precision")
	}
	return d, nil
}

func init() {
	// http_listener deprecated in 1.9
	inputs.Add("http_listener", func() telegraf.Input {
		return &InfluxDBListener{
			ServiceAddress: ":8186",
			TimeFunc:       time.Now,
		}
	})
	inputs.Add("influxdb_listener", func() telegraf.Input {
		return &InfluxDBListener{
			ServiceAddress: ":8186",
			TimeFunc:       time.Now,
		}
	})
}
