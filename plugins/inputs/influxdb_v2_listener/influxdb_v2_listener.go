package influxdb_v2_listener

import (
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	tlsint "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/selfstat"
)

const (
	// defaultMaxBodySize is the default maximum request body size, in bytes.
	// if the request body is over this size, we will return an HTTP 413 error.
	defaultMaxBodySize = 32 * 1024 * 1024
)

type InfluxDBV2Listener struct {
	ServiceAddress string `toml:"service_address"`
	port           int
	tlsint.ServerConfig

	ReadTimeout        internal.Duration `toml:"read_timeout"`
	WriteTimeout       internal.Duration `toml:"write_timeout"`
	MaxBodySize        internal.Size     `toml:"max_body_size"`
	MaxLineSize        internal.Size     `toml:"max_line_size"` // deprecated in 1.14; ignored
	Token              string            `toml:"token"`
	DatabaseTag        string            `toml:"database_tag"`
	RetentionPolicyTag string            `toml:"retention_policy_tag"`

	timeFunc influx.TimeFunc

	listener net.Listener
	server   http.Server

	acc telegraf.Accumulator

	bytesRecv       selfstat.Stat
	requestsServed  selfstat.Stat
	writesServed    selfstat.Stat
	queriesServed   selfstat.Stat
	readysServed    selfstat.Stat
	requestsRecv    selfstat.Stat
	notFoundsServed selfstat.Stat
	buffersCreated  selfstat.Stat
	authFailures    selfstat.Stat

	startTime time.Time

	Log telegraf.Logger `toml:"-"`

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
  ## 0 means to use the default of 32MiB.
  max_body_size = "32MiB"

  ## Optional tag name used to store the database.
  ## If the write has a database in the query string then it will be kept in this tag name.
  ## This tag can be used in downstream outputs.
  ## The default value of nothing means it will be off and the database will not be recorded.
  # database_tag = ""

  ## If set the retention policy specified in the write query will be added as
  ## the value of this tag name.
  # retention_policy_tag = ""

  ## Set one or more allowed client CA certificate file names to
  ## enable mutually authenticated TLS connections
  tls_allowed_cacerts = ["/etc/telegraf/clientca.pem"]

  ## Add service certificate and key
  tls_cert = "/etc/telegraf/cert.pem"
  tls_key = "/etc/telegraf/key.pem"

  ## Optional token to accept for HTTP authentication.
  ## You probably want to make sure you have TLS configured above for this.
  # token = "foobar"
`

func (h *InfluxDBV2Listener) SampleConfig() string {
	return sampleConfig
}

func (h *InfluxDBV2Listener) Description() string {
	return "Accept metrics over InfluxDB 1.x HTTP API"
}

func (h *InfluxDBV2Listener) Gather(_ telegraf.Accumulator) error {
	return nil
}

func (h *InfluxDBV2Listener) routes() {
	authHandler := internal.TokenAuthHandler(h.Token,
		func(_ http.ResponseWriter) {
			h.authFailures.Incr(1)
		},
	)

	h.mux.Handle("/api/v2/write", authHandler(h.handleWrite()))
	h.mux.Handle("/api/v2/query", authHandler(h.handleQuery()))
	h.mux.Handle("/api/v2/ready", h.handleReady())
	h.mux.Handle("/", authHandler(h.handleDefault()))
}

func (h *InfluxDBV2Listener) Init() error {
	tags := map[string]string{
		"address": h.ServiceAddress,
	}
	h.bytesRecv = selfstat.Register("influxdb_v2_listener", "bytes_received", tags)
	h.requestsServed = selfstat.Register("influxdb_v2_listener", "requests_served", tags)
	h.writesServed = selfstat.Register("influxdb_v2_listener", "writes_served", tags)
	h.queriesServed = selfstat.Register("influxdb_v2_listener", "queries_served", tags)
	h.readysServed = selfstat.Register("influxdb_v2_listener", "readys_served", tags)
	h.requestsRecv = selfstat.Register("influxdb_v2_listener", "requests_received", tags)
	h.notFoundsServed = selfstat.Register("influxdb_v2_listener", "not_founds_served", tags)
	h.buffersCreated = selfstat.Register("influxdb_v2_listener", "buffers_created", tags)
	h.authFailures = selfstat.Register("influxdb_v2_listener", "auth_failures", tags)
	h.routes()

	if h.MaxBodySize.Size == 0 {
		h.MaxBodySize.Size = defaultMaxBodySize
	}

	if h.MaxLineSize.Size != 0 {
		h.Log.Warnf("Use of deprecated configuration: 'max_line_size'; parser now handles lines of unlimited length and option is ignored")
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
func (h *InfluxDBV2Listener) Start(acc telegraf.Accumulator) error {
	h.acc = acc

	tlsConf, err := h.ServerConfig.TLSConfig()
	if err != nil {
		return err
	}

	h.server = http.Server{
		Addr:         h.ServiceAddress,
		Handler:      h,
		ReadTimeout:  h.ReadTimeout.Duration,
		WriteTimeout: h.WriteTimeout.Duration,
		TLSConfig:    tlsConf,
	}

	var listener net.Listener
	if tlsConf != nil {
		listener, err = tls.Listen("tcp", h.ServiceAddress, tlsConf)
		if err != nil {
			return err
		}
	} else {
		listener, err = net.Listen("tcp", h.ServiceAddress)
		if err != nil {
			return err
		}
	}
	h.listener = listener
	h.port = listener.Addr().(*net.TCPAddr).Port

	go func() {
		err = h.server.Serve(h.listener)
		if err != http.ErrServerClosed {
			h.Log.Infof("Error serving HTTP on %s", h.ServiceAddress)
		}
	}()

	h.startTime = time.Now()

	h.Log.Infof("Started HTTP listener service on %s", h.ServiceAddress)

	return nil
}

// Stop cleans up all resources
func (h *InfluxDBV2Listener) Stop() {
	err := h.server.Shutdown(context.Background())
	if err != nil {
		h.Log.Infof("Error shutting down HTTP server: %v", err.Error())
	}
}

func (h *InfluxDBV2Listener) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	h.requestsRecv.Incr(1)
	h.mux.ServeHTTP(res, req)
	h.requestsServed.Incr(1)
}

func (h *InfluxDBV2Listener) handleQuery() http.HandlerFunc {
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

func (h *InfluxDBV2Listener) handleReady() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		defer h.readysServed.Incr(1)

		// respond to ready requests
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
		b, _ := json.Marshal(map[string]string{
			// "2019-03-13T10:09:33.891196-04:00"
			"started": h.startTime.Format(time.RFC3339Nano),
			"status":  "ready",
			"up":      string(h.startTime.Sub(time.Now()))}) // based on header set above
		res.Write(b)
	}
}

func (h *InfluxDBV2Listener) handleDefault() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		defer h.notFoundsServed.Incr(1)
		http.NotFound(res, req)
	}
}

func (h *InfluxDBV2Listener) handleWrite() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		defer h.writesServed.Incr(1)
		// Check that the content length is not too large for us to handle.
		if req.ContentLength > h.MaxBodySize.Size {
			tooLarge(res)
			return
		}

		db := req.URL.Query().Get("db")
		rp := req.URL.Query().Get("rp")

		body := req.Body
		body = http.MaxBytesReader(res, body, h.MaxBodySize.Size)
		// Handle gzip request bodies
		if req.Header.Get("Content-Encoding") == "gzip" {
			var err error
			body, err = gzip.NewReader(body)
			if err != nil {
				h.Log.Debugf("Error decompressing request body: %v", err.Error())
				badRequest(res, err.Error())
				return
			}
			defer body.Close()
		}

		parser := influx.NewStreamParser(body)
		parser.SetTimeFunc(h.timeFunc)

		precisionStr := req.URL.Query().Get("precision")
		if precisionStr != "" {
			precision := getPrecisionMultiplier(precisionStr)
			parser.SetTimePrecision(precision)
		}

		var m telegraf.Metric
		var err error
		var parseErrorCount int
		var lastPos int = 0
		var firstParseErrorStr string
		for {
			select {
			case <-req.Context().Done():
				// Shutting down before parsing is finished.
				res.WriteHeader(http.StatusServiceUnavailable)
				return
			default:
			}

			m, err = parser.Next()
			pos := parser.Position()
			h.bytesRecv.Incr(int64(pos - lastPos))
			lastPos = pos

			// Continue parsing metrics even if some are malformed
			if parseErr, ok := err.(*influx.ParseError); ok {
				parseErrorCount += 1
				errStr := parseErr.Error()
				if firstParseErrorStr == "" {
					firstParseErrorStr = errStr
				}
				continue
			} else if err != nil {
				// Either we're exiting cleanly (err ==
				// influx.EOF) or there's an unexpected error
				break
			}

			if h.DatabaseTag != "" && db != "" {
				m.AddTag(h.DatabaseTag, db)
			}

			if h.RetentionPolicyTag != "" && rp != "" {
				m.AddTag(h.RetentionPolicyTag, rp)
			}

			h.acc.AddMetric(m)

		}
		if err != influx.EOF {
			h.Log.Debugf("Error parsing the request body: %v", err.Error())
			badRequest(res, err.Error())
			return
		}
		if parseErrorCount > 0 {
			var partialErrorString string
			switch parseErrorCount {
			case 1:
				partialErrorString = fmt.Sprintf("%s", firstParseErrorStr)
			case 2:
				partialErrorString = fmt.Sprintf("%s (and 1 other parse error)", firstParseErrorStr)
			default:
				partialErrorString = fmt.Sprintf("%s (and %d other parse errors)", firstParseErrorStr, parseErrorCount-1)
			}
			partialWrite(res, partialErrorString)
			return
		}

		// http request success
		res.WriteHeader(http.StatusNoContent)
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

func getPrecisionMultiplier(precision string) time.Duration {
	// Influxdb defaults silently to nanoseconds if precision isn't
	// one of the following:
	var d time.Duration
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
		d = time.Nanosecond
	}
	return d
}

func init() {
	inputs.Add("influxdb_v2_listener", func() telegraf.Input {
		return &InfluxDBV2Listener{
			ServiceAddress: ":8186",
			timeFunc:       time.Now,
		}
	})
}
