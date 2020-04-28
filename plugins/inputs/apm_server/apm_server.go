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
	"github.com/influxdata/telegraf/filter"
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
	ServiceAddress    string            `toml:"service_address"`
	IdleTimeout       internal.Duration `toml:"idle_timeout"`
	ReadTimeout       internal.Duration `toml:"read_timeout"`
	WriteTimeout      internal.Duration `toml:"write_timeout"`
	EventTypeSeparate bool              `toml:"event_type_separate"`

	ExcludeEventTypes []string `toml:"exclude_events"`
	//customize json -> line protocol mapping
	ExcludedFields []string `toml:"exclude_fields"`
	TagKeys        []string `toml:"tag_keys"`

	tlsint.ServerConfig

	port     int
	listener net.Listener
	server   http.Server

	buildDate time.Time
	buildSHA  string

	acc telegraf.Accumulator

	Log telegraf.Logger

	mux http.ServeMux

	eventTypeFilter      filter.Filter
	excludedFieldsFilter filter.Filter
	tagFilter            filter.Filter
}

func (s *APMServer) Description() string {
	return "APM Server is a input plugin that listens for requests sent by Elastic APM Agents."
}

func (s *APMServer) SampleConfig() string {
	return `
  ## Address and port to list APM Agents
  service_address = ":8200"

  ## maximum amount of time to wait for the next request when keep-alives are enabled	
  # idle_timeout =	"45s"	
  ## maximum duration before timing out read of the request
  # read_timeout =	"30s"
  ## maximum duration before timing out write of the response
  # write_timeout = "30s"
  ## exclude event types
  exclude_events = ["span"]	
  ## exclude fields matching following patterns
  exclude = ["exception_stacktrace_*", "log_stacktrace_*"]
  ## store selected fields as tags 
  # tag_keys =["my_tag_1", "my_tag_2" ]
`
}

func (s *APMServer) Init() error {

	s.routes()

	if s.IdleTimeout.Duration < time.Second {
		s.IdleTimeout.Duration = time.Second * 10
	}
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

	excludedFilter, err := filter.Compile(s.ExcludedFields)
	if err != nil {
		return err
	}
	s.excludedFieldsFilter = excludedFilter

	tagFilter, err := filter.Compile(s.TagKeys)
	if err != nil {
		return err
	}
	s.tagFilter = tagFilter

	eventTypeFilter, err := filter.Compile(s.ExcludeEventTypes)
	if err != nil {
		return err
	}
	s.eventTypeFilter = eventTypeFilter

	return nil
}

func (s *APMServer) Gather(_ telegraf.Accumulator) error {
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
		IdleTimeout:  s.IdleTimeout.Duration,
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
	s.mux.Handle("/config/v1/rum/agents", s.handleRUM(s.handleAgentConfiguration()))
	s.mux.Handle("/assets/v1/sourcemaps", s.handleSourceMap())
	s.mux.Handle("/intake/v2/events", s.handleEventsIntake())
	s.mux.Handle("/intake/v2/rum/events", s.handleRUM(s.handleEventsIntake()))
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
			"version":    "7.6.0",
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

			if s.eventTypeFilter != nil && s.eventTypeFilter.Match(eventType) {
				continue
			}

			f := jsonparser.JSONFlattener{}
			if err := f.FullFlattenJSON("", metadata.(map[string]interface{})["metadata"], true, true); err != nil {
				s.errorResponse(res, http.StatusBadRequest, err.Error())
				return
			}

			// Tags
			tags := make(map[string]string, len(f.Fields))
			tags["apm_event_type"] = eventType
			for k := range f.Fields {

				//skip if excluded
				if s.excludedFieldsFilter != nil && s.excludedFieldsFilter.Match(k) {
					continue
				}

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
			if err := f.FullFlattenJSON("", event.(map[string]interface{})[eventType], true, true); err != nil {
				s.errorResponse(res, http.StatusBadRequest, err.Error())
				return
			}

			for k := range f.Fields {

				//remove _value suffix
				if strings.HasSuffix(k, "_value") {
					var val = f.Fields[k]
					delete(f.Fields, k)
					k = k[0 : len(k)-len("_value")]
					f.Fields[k] = val
				}

				// Exclude fields filter
				if s.excludedFieldsFilter != nil && s.excludedFieldsFilter.Match(k) {
					delete(f.Fields, k)
				}

				// Store fields as tags
				if s.tagFilter != nil && s.tagFilter.Match(k) {
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
					delete(f.Fields, k)
				}
			}
			delete(f.Fields, "timestamp")

			// Timestamp
			timestamp, err := parseTimestamp(event, eventType)
			if err != nil {
				s.errorResponse(res, http.StatusBadRequest, err.Error())
				return
			}

			var measurementName = "apm_server"
			if s.EventTypeSeparate {
				measurementName = "apm_" + eventType
			}
			if m, err := metric.New(measurementName, tags, f.Fields, timestamp); err != nil {
				s.errorResponse(res, http.StatusBadRequest, err.Error())
				return
			} else {
				s.acc.AddMetric(m)
			}
		}

		res.WriteHeader(http.StatusAccepted)
	}
}

func (s *APMServer) handleRUM(handler http.HandlerFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		// Handle CORS
		//
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS#The_HTTP_response_headers
		//
		origin := req.Header.Get("Origin")
		res.Header().Set("Access-Control-Allow-Origin", origin)

		if req.Method == "OPTIONS" {
			res.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
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
