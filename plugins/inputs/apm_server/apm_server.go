package apm_server

import (
	"compress/gzip"
	"compress/zlib"
	"context"
	"crypto/sha1"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	tlsint "github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/inputs"
	jsonparser "github.com/influxdata/telegraf/plugins/parsers/json"
	"io"
	"log"
	"net"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// APM Server is a input plugin that listens for requests sent by Elastic APM Agents.
type APMServer struct {
	ServiceAddress string            `toml:"service_address"`
	ReadTimeout    internal.Duration `toml:"read_timeout"`
	WriteTimeout   internal.Duration `toml:"write_timeout"`
	tlsint.ServerConfig

	port     int
	listener net.Listener
	server   http.Server

	buildDate time.Time
	buildSHA  string

	acc telegraf.Accumulator

	Log telegraf.Logger

	mux http.ServeMux
}

func (s *APMServer) Description() string {
	return "APM Server is a input plugin that listens for requests sent by Elastic APM Agents."
}

func (s *APMServer) SampleConfig() string {
	return `
  ## Address and port to list APM Agents
  service_address = ":8200"

  ## maximum duration before timing out read of the request
  # read_timeout = "10s"
  ## maximum duration before timing out write of the response
  # write_timeout = "10s"
`
}

func (s *APMServer) Init() error {

	s.routes()

	if s.ReadTimeout.Duration < time.Second {
		s.ReadTimeout.Duration = time.Second * 10
	}
	if s.WriteTimeout.Duration < time.Second {
		s.WriteTimeout.Duration = time.Second * 10
	}

	// prepare build_sha and build_date for ServerInformation endpoint
	h := sha1.New()
	h.Write([]byte(s.SampleConfig()))
	s.buildSHA = hex.EncodeToString(h.Sum(nil))
	s.buildDate = time.Now()

	return nil
}

func (s *APMServer) Gather(acc telegraf.Accumulator) error {
	acc.AddFields("apm_server", map[string]interface{}{"service_address": s.ServiceAddress}, nil)

	return nil
}

// Start starts the http listener service.
func (s *APMServer) Start(acc telegraf.Accumulator) error {
	s.acc = acc

	tlsConf, err := s.ServerConfig.TLSConfig()
	if err != nil {
		return err
	}

	s.server = http.Server{
		Addr:         s.ServiceAddress,
		Handler:      s,
		ReadTimeout:  s.ReadTimeout.Duration,
		WriteTimeout: s.WriteTimeout.Duration,
		TLSConfig:    tlsConf,
	}

	var listener net.Listener
	if tlsConf != nil {
		listener, err = tls.Listen("tcp", s.ServiceAddress, tlsConf)
		if err != nil {
			return err
		}
	} else {
		listener, err = net.Listen("tcp", s.ServiceAddress)
		if err != nil {
			return err
		}
	}
	s.listener = listener
	s.port = listener.Addr().(*net.TCPAddr).Port

	go func() {
		err = s.server.Serve(s.listener)
		if err != http.ErrServerClosed {
			s.Log.Infof("Error start APM Server on %s", s.ServiceAddress)
		}
	}()

	s.Log.Infof("Started APM Server on %s", s.ServiceAddress)

	return nil
}

// Stop cleans up all resources
func (s *APMServer) Stop() {
	err := s.server.Shutdown(context.Background())
	if err != nil {
		s.Log.Infof("Error shutting down HTTP server: %v", err.Error())
	}
}

func (s *APMServer) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	s.mux.ServeHTTP(res, req)
}

func (s *APMServer) routes() {
	s.mux.Handle("/", s.handleServerInformation())
	s.mux.Handle("/config/v1/agents", s.handleAgentConfiguration())
	s.mux.Handle("/config/v1/rum/agents", s.handleRUM(s.handleAgentConfiguration(), "GET, OPTIONS"))
	s.mux.Handle("/assets/v1/sourcemaps", s.handleSourceMap())
	s.mux.Handle("/intake/v2/events", s.handleEventsIntake())
}

func (s *APMServer) handleServerInformation() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		if req.URL.Path != "/" {
			s.errorResponse(res, http.StatusNotFound, "404 page not found")
			return
		}

		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
		b, _ := json.Marshal(map[string]string{
			"build_date": s.buildDate.Format(time.RFC3339),
			"build_sha":  s.buildSHA,
			"version":    internal.Version(),
		})
		_, _ = res.Write(b)
	}
}

func (s *APMServer) handleAgentConfiguration() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusForbidden)
	}
}

func (s *APMServer) handleSourceMap() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusAccepted)
	}
}

func (s *APMServer) handleEventsIntake() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		if !strings.Contains(req.Header.Get("Content-Type"), "application/x-ndjson") {
			message := fmt.Sprintf("invalid content type: '%s'", req.Header.Get("Content-Type"))
			s.errorResponse(res, http.StatusBadRequest, message)
			return
		}

		var metadata interface{}
		reader, err := serverRequestBody(req)
		if err != nil {
			s.errorResponse(res, http.StatusBadRequest, err.Error())
			return
		}

		d := json.NewDecoder(reader)
		for {
			var event interface{}
			if err := d.Decode(&event); err != nil {
				if err != io.EOF {
					s.errorResponse(res, http.StatusBadRequest, err.Error())
					break
				}
				// EOF => end
				break
			}

			eventType := reflect.ValueOf(event.(map[string]interface{})).MapKeys()[0].String()
			if eventType == "metadata" {
				metadata = event
				continue
			}

			f := jsonparser.JSONFlattener{FieldsSeparator: "."}
			if err := f.FullFlattenJSON("", metadata, true, true); err != nil {
				s.errorResponse(res, http.StatusBadRequest, err.Error())
				return
			}

			// Tags
			tags := make(map[string]string, len(f.Fields))
			tags["type"] = eventType
			for k := range f.Fields {
				switch value := f.Fields[k].(type) {
				case string:
					tags[k] = value
				case bool:
					tags[k] = strconv.FormatBool(value)
				case float64:
					tags[k] = strconv.FormatFloat(value, 'f', -1, 64)
				default:
					log.Printf("E! [handleEventsIntake] Unrecognized tag type %T", value)
				}
			}

			// Fields
			f.Fields = make(map[string]interface{})
			if err := f.FullFlattenJSON("", event, true, true); err != nil {
				s.errorResponse(res, http.StatusBadRequest, err.Error())
				return
			}
			delete(f.Fields, eventType+".timestamp")

			// Timestamp
			timestamp, err := parseTimestamp(event, eventType)
			if err != nil {
				s.errorResponse(res, http.StatusBadRequest, err.Error())
				return
			}

			if m, err := metric.New("apm_server", tags, f.Fields, timestamp); err != nil {
				s.errorResponse(res, http.StatusBadRequest, err.Error())
				return
			} else {
				s.acc.AddMetric(m)
			}
		}

		res.WriteHeader(http.StatusAccepted)
	}
}

func (s *APMServer) handleRUM(handler http.HandlerFunc, allowedMethods string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		// Handle CORS
		//
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS#The_HTTP_response_headers
		//
		origin := req.Header.Get("Origin")
		res.Header().Set("Access-Control-Allow-Origin", origin)

		if req.Method == "OPTIONS" {
			res.Header().Set("Access-Control-Allow-Methods", allowedMethods)
			res.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Encoding, Accept")
			res.Header().Set("Access-Control-Expose-Headers", "Etag")
			res.Header().Set("Access-Control-Max-Age", "86400")
			res.Header().Set("Vary", "Origin")
			res.WriteHeader(http.StatusNoContent)
		} else {
			handler.ServeHTTP(res, req)
		}
	}
}

func parseTimestamp(event interface{}, eventType string) (time.Time, error) {
	value := event.(map[string]interface{})[eventType].(map[string]interface{})["timestamp"]
	if value == nil {
		return time.Now().UTC(), nil
	}
	if valueFloat, ok := value.(float64); ok {
		microseconds := int64(valueFloat)
		secPart := microseconds / 1000000
		microPart := microseconds - (secPart * 1000000)
		return time.Unix(secPart, microPart*1000).UTC(), nil
	}
	return time.Now().UTC(), errors.New(fmt.Sprintf("cannot parse timestamp: '%s'", value))
}

func serverRequestBody(req *http.Request) (io.ReadCloser, error) {
	reader := req.Body
	switch req.Header.Get("Content-Encoding") {

	case "gzip":
		var err error
		if reader, err = gzip.NewReader(reader); err != nil {
			return nil, err
		}
	case "deflate":
		var err error
		if reader, err = zlib.NewReader(reader); err != nil {
			return nil, err
		}
	}

	return reader, nil
}

func (s *APMServer) errorResponse(res http.ResponseWriter, statusCode int, message string) {
	s.Log.Error(message)
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(statusCode)
	b, _ := json.Marshal(map[string]string{
		"error": message,
	})
	_, _ = res.Write(b)
}

func init() {
	inputs.Add("apm_server", func() telegraf.Input {
		return &APMServer{
			ServiceAddress: ":8200",
		}
	})
}
