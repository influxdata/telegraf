//go:generate ../../../tools/readme_config_includer/generator
package influxdb_v3_listener

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/subtle"
	"crypto/tls"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofrs/uuid/v5"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	common_tls "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/plugins/parsers/influx/influx_upstream"
	"github.com/influxdata/telegraf/selfstat"
)

//go:embed sample.conf
var sampleConfig string

const (
	// defaultMaxBodySize is the default maximum request body size, in bytes.
	// If the request body is over this size, we will return an HTTP 413 error.
	defaultMaxBodySize  = 32 * 1024 * 1024
	defaultReadTimeout  = 10 * time.Second
	defaultWriteTimeout = 10 * time.Second
)

type InfluxDBV3Listener struct {
	ServiceAddress        string          `toml:"service_address"`
	MaxUndeliveredMetrics int             `toml:"max_undelivered_metrics"`
	ReadTimeout           config.Duration `toml:"read_timeout"`
	WriteTimeout          config.Duration `toml:"write_timeout"`
	MaxBodySize           config.Size     `toml:"max_body_size"`
	Token                 config.Secret   `toml:"token"`
	DatabaseTag           string          `toml:"database_tag"`
	ParserType            string          `toml:"parser_type"`
	common_tls.ServerConfig

	Statistics *selfstat.Collector `toml:"-"`
	Log        telegraf.Logger     `toml:"-"`

	acc         telegraf.Accumulator
	trackingAcc telegraf.TrackingAccumulator

	ctx    context.Context
	cancel context.CancelFunc

	trackingMetricCount     map[telegraf.TrackingID]int64
	countLock               sync.Mutex
	totalUndeliveredMetrics atomic.Int64

	listener  net.Listener
	server    http.Server
	mux       http.ServeMux
	processID string
	timeFunc  func() time.Time

	bytesRecv       selfstat.Stat
	requestsRecv    selfstat.Stat
	requestsServed  selfstat.Stat
	writesServed    selfstat.Stat
	healthsServed   selfstat.Stat
	pingsServed     selfstat.Stat
	notFoundsServed selfstat.Stat
	authFailures    selfstat.Stat
}

func (*InfluxDBV3Listener) SampleConfig() string {
	return sampleConfig
}

func (h *InfluxDBV3Listener) Init() error {
	switch h.ParserType {
	case "", "internal", "upstream":
	default:
		return fmt.Errorf("invalid parser type %q", h.ParserType)
	}

	tags := map[string]string{"address": h.ServiceAddress}
	h.bytesRecv = h.Statistics.Register("influxdb_v3_listener", "bytes_received", tags)
	h.requestsRecv = h.Statistics.Register("influxdb_v3_listener", "requests_received", tags)
	h.requestsServed = h.Statistics.Register("influxdb_v3_listener", "requests_served", tags)
	h.writesServed = h.Statistics.Register("influxdb_v3_listener", "writes_served", tags)
	h.healthsServed = h.Statistics.Register("influxdb_v3_listener", "healths_served", tags)
	h.pingsServed = h.Statistics.Register("influxdb_v3_listener", "pings_served", tags)
	h.notFoundsServed = h.Statistics.Register("influxdb_v3_listener", "not_founds_served", tags)
	h.authFailures = h.Statistics.Register("influxdb_v3_listener", "auth_failures", tags)

	if h.MaxBodySize == 0 {
		h.MaxBodySize = config.Size(defaultMaxBodySize)
	}
	if h.ReadTimeout < config.Duration(time.Second) {
		h.ReadTimeout = config.Duration(defaultReadTimeout)
	}
	if h.WriteTimeout < config.Duration(time.Second) {
		h.WriteTimeout = config.Duration(defaultWriteTimeout)
	}

	processID, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("generating process identifier failed: %w", err)
	}
	h.processID = processID.String()

	return h.routes()
}

func (*InfluxDBV3Listener) Gather(telegraf.Accumulator) error {
	return nil
}

func (h *InfluxDBV3Listener) Start(acc telegraf.Accumulator) error {
	h.acc = acc
	h.ctx, h.cancel = context.WithCancel(context.Background())
	if h.MaxUndeliveredMetrics > 0 {
		h.trackingAcc = h.acc.WithTracking(h.MaxUndeliveredMetrics)
		h.trackingMetricCount = make(map[telegraf.TrackingID]int64, h.MaxUndeliveredMetrics)
		go func() {
			for {
				select {
				case <-h.ctx.Done():
					return
				case info := <-h.trackingAcc.Delivered():
					h.countLock.Lock()
					if count, ok := h.trackingMetricCount[info.ID()]; ok {
						h.totalUndeliveredMetrics.Add(-count)
						delete(h.trackingMetricCount, info.ID())
					}
					h.countLock.Unlock()
				}
			}
		}()
	}

	tlsConf, err := h.ServerConfig.TLSConfig()
	if err != nil {
		return err
	}

	h.server = http.Server{
		Addr:         h.ServiceAddress,
		Handler:      h,
		TLSConfig:    tlsConf,
		ReadTimeout:  time.Duration(h.ReadTimeout),
		WriteTimeout: time.Duration(h.WriteTimeout),
	}

	if tlsConf != nil {
		h.listener, err = tls.Listen("tcp", h.ServiceAddress, tlsConf)
	} else {
		h.listener, err = net.Listen("tcp", h.ServiceAddress)
	}
	if err != nil {
		h.Stop()
		return err
	}

	go func() {
		if err := h.server.Serve(h.listener); !errors.Is(err, http.ErrServerClosed) {
			h.Log.Errorf("Serving HTTP on %s failed: %v", h.ServiceAddress, err)
		}
	}()

	h.Log.Infof("Started HTTP listener service on %s", h.ServiceAddress)

	return nil
}

func (h *InfluxDBV3Listener) Stop() {
	h.cancel()
	if err := h.server.Shutdown(context.Background()); err != nil {
		h.Log.Infof("Error shutting down HTTP server: %v", err)
	}
}

func (h *InfluxDBV3Listener) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	h.requestsRecv.Incr(1)
	h.mux.ServeHTTP(res, req)
	h.requestsServed.Incr(1)
}

func (h *InfluxDBV3Listener) routes() error {
	var token []byte
	if !h.Token.Empty() {
		secret, err := h.Token.Get()
		if err != nil {
			return fmt.Errorf("getting token failed: %w", err)
		}
		token = bytes.Clone(secret.Bytes())
		secret.Destroy()
	}
	auth := h.authenticate(token)

	h.mux.Handle("POST /api/v3/write_lp", auth(h.handleWrite()))
	h.mux.Handle("GET /health", h.handleHealth())
	h.mux.Handle("GET /api/v1/health", h.handleHealth())
	h.mux.Handle("GET /ping", h.handlePing())
	h.mux.Handle("POST /ping", h.handlePing())
	h.mux.Handle("/", auth(h.handleDefault()))

	return nil
}

// authenticate requires the configured token to be sent as one of the schemes
// InfluxDB 3 accepts, i.e. "Bearer <token>", "Token <token>" or as the password
// part of a "Basic" credential. It is a no-op if no token is configured.
func (h *InfluxDBV3Listener) authenticate(token []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			if len(token) > 0 {
				provided, err := requestToken(req.Header.Get("Authorization"))
				if err != nil || subtle.ConstantTimeCompare(provided, token) != 1 {
					h.authFailures.Incr(1)
					http.Error(res, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
					return
				}
			}
			next.ServeHTTP(res, req)
		})
	}
}

func (h *InfluxDBV3Listener) handleHealth() http.HandlerFunc {
	return func(res http.ResponseWriter, _ *http.Request) {
		defer h.healthsServed.Incr(1)

		res.Header().Set("Content-Type", "text/plain; charset=utf-8")

		pending := h.totalUndeliveredMetrics.Load()
		if h.MaxUndeliveredMetrics > 0 && pending >= int64(h.MaxUndeliveredMetrics) {
			res.WriteHeader(http.StatusServiceUnavailable)
			msg := fmt.Sprintf("pending undelivered metrics (%d) is above limit", pending)
			if _, err := res.Write([]byte(msg)); err != nil {
				h.Log.Debugf("Writing health response failed: %v", err)
			}
			return
		}

		res.WriteHeader(http.StatusOK)
		if _, err := res.Write([]byte("OK")); err != nil {
			h.Log.Debugf("Writing health response failed: %v", err)
		}
	}
}

func (h *InfluxDBV3Listener) handlePing() http.HandlerFunc {
	return func(res http.ResponseWriter, _ *http.Request) {
		defer h.pingsServed.Incr(1)

		res.Header().Set("Content-Type", "application/json")
		res.Header().Set("X-Influxdb-Build", "telegraf")
		res.Header().Set("X-Influxdb-Version", internal.Version)
		res.WriteHeader(http.StatusOK)

		b, err := json.Marshal(map[string]string{
			"product_name": "telegraf",
			"version":      internal.Version,
			"revision":     internal.Commit,
			"process_id":   h.processID,
		})
		if err != nil {
			h.Log.Debugf("Marshalling ping response failed: %v", err)
			return
		}
		if _, err := res.Write(b); err != nil {
			h.Log.Debugf("Writing ping response failed: %v", err)
		}
	}
}

func (h *InfluxDBV3Listener) handleDefault() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		defer h.notFoundsServed.Incr(1)
		http.NotFound(res, req)
	}
}

func (h *InfluxDBV3Listener) handleWrite() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		defer h.writesServed.Incr(1)

		// Check that the content length is not too large for us to handle
		if req.ContentLength > int64(h.MaxBodySize) {
			h.tooLarge(res)
			return
		}

		// InfluxDB requires the database, the remaining parameters are optional
		if req.URL.RawQuery == "" {
			h.badRequest(res, "missing query parameter 'db'")
			return
		}
		query := req.URL.Query()
		database := query.Get("db")
		if database == "" {
			h.badRequest(res, "db name cannot be empty")
			return
		}
		precision, err := parsePrecision(query.Get("precision"))
		if err != nil {
			h.badRequest(res, err.Error())
			return
		}
		acceptPartial, err := parseBool(query.Get("accept_partial"), true)
		if err != nil {
			h.badRequest(res, err.Error())
			return
		}
		// The 'no_sync' parameter controls acknowledgement of the write-ahead
		// log which does not exist here, so it is only validated
		if _, err := parseBool(query.Get("no_sync"), false); err != nil {
			h.badRequest(res, err.Error())
			return
		}

		body, err := h.readBody(res, req)
		if err != nil {
			// Bodies sent without a content-length only exceed the limit while reading them
			if maxBytes, ok := errors.AsType[*http.MaxBytesError](err); ok {
				h.Log.Debugf("Rejecting write request: body exceeds the limit of %d bytes", maxBytes.Limit)
				h.tooLarge(res)
				return
			}
			h.badRequest(res, err.Error())
			return
		}
		h.bytesRecv.Incr(int64(len(body)))

		parser, err := h.newParser(bytes.NewReader(body), precision)
		if err != nil {
			h.Log.Errorf("Creating parser failed: %v", err)
			http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// Parse the body line by line, keeping the valid metrics and collecting
		// the errors of the invalid lines. The original lines are only needed
		// for reporting, so they are split off the body on the first error.
		var metrics []telegraf.Metric
		var invalid []lineError
		var lines []string
		for {
			select {
			case <-req.Context().Done():
				// Shutting down before parsing is finished
				res.WriteHeader(http.StatusServiceUnavailable)
				return
			default:
			}

			m, err := parser.Next()
			if err != nil {
				lineNumber, ok := parseErrorLine(err)
				if !ok {
					if !errors.Is(err, influx.EOF) && !errors.Is(err, io.EOF) {
						h.badRequest(res, err.Error())
						return
					}
					break
				}

				if lines == nil {
					lines = strings.Split(string(body), "\n")
				}
				e := lineError{
					OriginalLine: originalLine(lines, lineNumber),
					LineNumber:   lineNumber,
					ErrorMessage: err.Error(),
				}
				if !acceptPartial {
					h.rejectedWrite(res, e)
					return
				}
				invalid = append(invalid, e)
				continue
			}

			if h.DatabaseTag != "" {
				m.AddTag(h.DatabaseTag, database)
			}
			if precision == 0 {
				m.SetTime(guessPrecision(m.Time()))
			}
			metrics = append(metrics, m)
		}

		if h.MaxUndeliveredMetrics > 0 {
			if !h.writeWithTracking(res, metrics) {
				return
			}
		} else {
			for _, m := range metrics {
				h.acc.AddMetric(m)
			}
		}

		if len(invalid) > 0 {
			h.partialWrite(res, invalid)
			return
		}
		res.WriteHeader(http.StatusNoContent)
	}
}

// readBody returns the request body, limited to the maximum body size and
// decompressed if the client sent it compressed.
func (h *InfluxDBV3Listener) readBody(res http.ResponseWriter, req *http.Request) ([]byte, error) {
	body := http.MaxBytesReader(res, req.Body, int64(h.MaxBodySize))
	if req.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(body)
		if err != nil {
			return nil, fmt.Errorf("decompressing request body failed: %w", err)
		}
		defer reader.Close()

		return io.ReadAll(reader)
	}

	return io.ReadAll(body)
}

func (h *InfluxDBV3Listener) newParser(r io.Reader, precision time.Duration) (streamParser, error) {
	if h.ParserType == "upstream" {
		parser := influx_upstream.NewStreamParser(r)
		parser.SetTimeFunc(influx_upstream.TimeFunc(h.timeFunc))
		if precision > 0 {
			if err := parser.SetTimePrecision(precision); err != nil {
				return nil, fmt.Errorf("setting time precision failed: %w", err)
			}
		}
		return parser, nil
	}

	parser := influx.NewStreamParser(r)
	parser.SetTimeFunc(h.timeFunc)
	if precision > 0 {
		parser.SetTimePrecision(precision)
	}
	return parser, nil
}

// writeWithTracking delivers the metrics through the tracking accumulator and
// reports whether they were accepted. The response is written by the caller if
// and only if they were.
func (h *InfluxDBV3Listener) writeWithTracking(res http.ResponseWriter, metrics []telegraf.Metric) bool {
	if len(metrics) == 0 {
		return true
	}

	if len(metrics) > h.MaxUndeliveredMetrics {
		res.WriteHeader(http.StatusRequestEntityTooLarge)
		h.Log.Debugf("status %d, always rejecting batch of %d metrics: larger than max_undelivered_metrics %d",
			http.StatusRequestEntityTooLarge, len(metrics), h.MaxUndeliveredMetrics)
		return false
	}

	remaining := int64(h.MaxUndeliveredMetrics) - h.totalUndeliveredMetrics.Load()
	if int64(len(metrics)) > remaining {
		res.WriteHeader(http.StatusTooManyRequests)
		h.Log.Debugf("status %d, rejecting batch of %d metrics: larger than remaining undelivered metrics %d",
			http.StatusTooManyRequests, len(metrics), remaining)
		return false
	}

	h.countLock.Lock()
	trackingID := h.trackingAcc.AddTrackingMetricGroup(metrics)
	h.trackingMetricCount[trackingID] = int64(len(metrics))
	h.totalUndeliveredMetrics.Add(int64(len(metrics)))
	h.countLock.Unlock()

	return true
}

func (h *InfluxDBV3Listener) partialWrite(res http.ResponseWriter, invalid []lineError) {
	h.writeError(res, http.StatusBadRequest, "partial write of line protocol occurred", invalid)
}

func (h *InfluxDBV3Listener) rejectedWrite(res http.ResponseWriter, invalid lineError) {
	h.writeError(res, http.StatusBadRequest, "line protocol parsing error", invalid)
}

func (h *InfluxDBV3Listener) writeError(res http.ResponseWriter, code int, msg string, data any) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(code)

	b, err := json.Marshal(errorMessage{Error: msg, Data: data})
	if err != nil {
		h.Log.Debugf("Marshalling error response failed: %v", err)
		return
	}
	if _, err := res.Write(b); err != nil {
		h.Log.Debugf("Writing error response failed: %v", err)
	}
}

// badRequest reports a request that was rejected before any line protocol was
// parsed. Those errors are plain text in InfluxDB, only write failures use JSON.
func (h *InfluxDBV3Listener) badRequest(res http.ResponseWriter, msg string) {
	h.Log.Debugf("Rejecting write request: %s", msg)
	http.Error(res, msg, http.StatusBadRequest)
}

func (h *InfluxDBV3Listener) tooLarge(res http.ResponseWriter) {
	res.Header().Set("X-Influxdb-Error", "http: request body too large")
	msg := "the request body exceeded the maximum size of " + strconv.FormatInt(int64(h.MaxBodySize), 10) + " bytes"
	http.Error(res, msg, http.StatusRequestEntityTooLarge)
}

// streamParser is the common part of the two line protocol stream parsers, both
// of which can resume after a parse error to report the next line.
type streamParser interface {
	Next() (telegraf.Metric, error)
}

// errorMessage is the InfluxDB 3 error response. The data is a list of line
// errors for a partial write and a single line error for a rejected one.
type errorMessage struct {
	Error string `json:"error"`
	Data  any    `json:"data,omitempty"`
}

type lineError struct {
	OriginalLine string `json:"original_line"`
	LineNumber   int    `json:"line_number"`
	ErrorMessage string `json:"error_message"`
}

// requestToken extracts the token from an authorization header using one of the
// schemes InfluxDB 3 supports.
func requestToken(header string) ([]byte, error) {
	scheme, credentials, found := strings.Cut(header, " ")
	if !found || strings.Contains(credentials, " ") {
		return nil, errors.New("authorization header is not in the form of 'Authorization: <auth-scheme> <token>'")
	}

	switch scheme {
	case "Bearer", "Token":
		return []byte(credentials), nil
	case "Basic":
		decoded, err := base64.StdEncoding.DecodeString(credentials)
		if err != nil {
			return nil, fmt.Errorf("decoding basic credentials failed: %w", err)
		}
		// The token is the password part, the username is ignored
		_, token, found := strings.Cut(string(decoded), ":")
		if !found || strings.Contains(token, ":") {
			return nil, errors.New("basic credentials are not in the form of '<username>:<token>'")
		}
		return []byte(token), nil
	}

	return nil, fmt.Errorf("unsupported auth-scheme %q, supported auth-schemes are Bearer, Token and Basic", scheme)
}

// parsePrecision maps the precision parameter to the multiplier of the
// timestamps in the body. A zero duration denotes the "auto" precision which
// has to be determined per timestamp.
func parsePrecision(precision string) (time.Duration, error) {
	switch precision {
	case "", "auto":
		return 0, nil
	case "s", "second":
		return time.Second, nil
	case "ms", "millisecond":
		return time.Millisecond, nil
	case "u", "us", "microsecond":
		return time.Microsecond, nil
	case "n", "ns", "nanosecond":
		return time.Nanosecond, nil
	}

	return 0, fmt.Errorf("unknown precision %q, expected one of auto, second, millisecond, microsecond or nanosecond", precision)
}

func parseBool(value string, defaultValue bool) (bool, error) {
	if value == "" {
		return defaultValue, nil
	}

	// InfluxDB only accepts the lower-case spelling of the boolean values
	switch value {
	case "true":
		return true, nil
	case "false":
		return false, nil
	}

	return false, fmt.Errorf("provided string was not `true` or `false`, got %q", value)
}

// guessPrecision determines the unit of a timestamp parsed at nanosecond
// precision by its magnitude, the way InfluxDB does for the "auto" precision.
// Timestamps without a unit reaching into the nanosecond range, which includes
// the server time filled in for lines without a timestamp, are left untouched.
func guessPrecision(t time.Time) time.Time {
	raw := t.UnixNano()

	magnitude := raw / int64(time.Second)
	if magnitude < 0 {
		magnitude = -magnitude
	}

	switch {
	case magnitude < 5:
		return time.Unix(raw, 0)
	case magnitude < 5_000:
		return time.Unix(0, raw*int64(time.Millisecond))
	case magnitude < 5_000_000:
		return time.Unix(0, raw*int64(time.Microsecond))
	}

	return t
}

// parseErrorLine reports the one-based number of the line that failed to parse.
func parseErrorLine(err error) (int, bool) {
	if internalErr, ok := errors.AsType[*influx.ParseError](err); ok {
		return internalErr.LineNumber, true
	}

	if upstreamErr, ok := errors.AsType[*influx_upstream.ParseError](err); ok {
		return int(upstreamErr.Line), true
	}

	return 0, false
}

// originalLine returns the line the parser reported an error for. The parsers
// truncate the line at the error column, so it is taken from the body instead.
func originalLine(lines []string, lineNumber int) string {
	if lineNumber < 1 || lineNumber > len(lines) {
		return ""
	}

	return strings.TrimSuffix(lines[lineNumber-1], "\r")
}

func init() {
	inputs.Add("influxdb_v3_listener", func() telegraf.Input {
		return &InfluxDBV3Listener{
			ServiceAddress: ":8181",
			timeFunc:       time.Now,
		}
	})
}
