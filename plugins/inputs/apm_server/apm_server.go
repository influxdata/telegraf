package apm_server

import (
	"context"
	"crypto/sha1"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	tlsint "github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"net"
	"net/http"
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
	s.mux.Handle("/config/v1/rum/agents", s.handleAgentConfiguration())
}

func (s *APMServer) handleServerInformation() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		res.Header().Set("Content-Type", "application/json")
		if req.URL.Path != "/" {
			res.WriteHeader(http.StatusNotFound)
			b, _ := json.Marshal(map[string]string{
				"error": "404 page not found",
			})
			_, _ = res.Write(b)
			return
		}

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

func init() {
	inputs.Add("apm_server", func() telegraf.Input {
		return &APMServer{
			ServiceAddress: ":8200",
		}
	})
}
