package health

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	tlsint "github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
)

const (
	defaultServiceAddress = "tcp://:8080"
	defaultReadTimeout    = 5 * time.Second
	defaultWriteTimeout   = 5 * time.Second
)

var sampleConfig = `
  ## Address and port to listen on.
  ##   ex: service_address = "tcp://localhost:8080"
  ##       service_address = "unix:///var/run/telegraf-health.sock"
  # service_address = "tcp://:8080"

  ## The maximum duration for reading the entire request.
  # read_timeout = "5s"
  ## The maximum duration for writing the entire response.
  # write_timeout = "5s"

  ## Username and password to accept for HTTP basic authentication.
  # basic_username = "user1"
  # basic_password = "secret"

  ## Allowed CA certificates for client certificates.
  # tls_allowed_cacerts = ["/etc/telegraf/clientca.pem"]

  ## TLS server certificate and private key.
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"

  ## One or more check sub-tables should be defined, it is also recommended to
  ## use metric filtering to limit the metrics that flow into this output.
  ##
  ## When using the default buffer sizes, this example will fail when the
  ## metric buffer is half full.
  ##
  ## namepass = ["internal_write"]
  ## tagpass = { output = ["influxdb"] }
  ##
  ## [[outputs.health.compares]]
  ##   field = "buffer_size"
  ##   lt = 5000.0
  ##
  ## [[outputs.health.contains]]
  ##   field = "buffer_size"
`

type Checker interface {
	// Check returns true if the metrics meet its criteria.
	Check(metrics []telegraf.Metric) bool
}

type Health struct {
	ServiceAddress string            `toml:"service_address"`
	ReadTimeout    internal.Duration `toml:"read_timeout"`
	WriteTimeout   internal.Duration `toml:"write_timeout"`
	BasicUsername  string            `toml:"basic_username"`
	BasicPassword  string            `toml:"basic_password"`
	tlsint.ServerConfig

	Compares []*Compares `toml:"compares"`
	Contains []*Contains `toml:"contains"`
	checkers []Checker

	wg     sync.WaitGroup
	server *http.Server
	origin string

	mu      sync.Mutex
	healthy bool
}

func (h *Health) SampleConfig() string {
	return sampleConfig
}

func (h *Health) Description() string {
	return "Configurable HTTP health check resource based on metrics"
}

// Connect starts the HTTP server.
func (h *Health) Connect() error {
	h.checkers = make([]Checker, 0)
	for i := range h.Compares {
		h.checkers = append(h.checkers, h.Compares[i])
	}
	for i := range h.Contains {
		h.checkers = append(h.checkers, h.Contains[i])
	}

	tlsConf, err := h.ServerConfig.TLSConfig()
	if err != nil {
		return err
	}

	authHandler := internal.AuthHandler(h.BasicUsername, h.BasicPassword, onAuthError)

	h.server = &http.Server{
		Addr:         h.ServiceAddress,
		Handler:      authHandler(h),
		ReadTimeout:  h.ReadTimeout.Duration,
		WriteTimeout: h.WriteTimeout.Duration,
		TLSConfig:    tlsConf,
	}

	listener, err := h.listen(tlsConf)
	if err != nil {
		return err
	}

	h.origin = h.getOrigin(listener, tlsConf)

	log.Printf("I! [outputs.health] Listening on %s", h.origin)

	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		err := h.server.Serve(listener)
		if err != http.ErrServerClosed {
			log.Printf("E! [outputs.health] Serve error on %s: %v", h.origin, err)
		}
		h.origin = ""
	}()

	return nil
}

func onAuthError(rw http.ResponseWriter, code int) {
	http.Error(rw, http.StatusText(code), code)
}

func (h *Health) listen(tlsConf *tls.Config) (net.Listener, error) {
	u, err := url.Parse(h.ServiceAddress)
	if err != nil {
		return nil, err
	}

	network := "tcp"
	address := u.Host
	if u.Host == "" {
		network = "unix"
		address = u.Path
	}

	if tlsConf != nil {
		return tls.Listen(network, address, tlsConf)
	} else {
		return net.Listen(network, address)
	}

}

func (h *Health) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	var code = http.StatusOK
	if !h.isHealthy() {
		code = http.StatusServiceUnavailable
	}

	rw.Header().Set("Server", internal.ProductToken())
	http.Error(rw, http.StatusText(code), code)
}

// Write runs all checks over the metric batch and adjust health state.
func (h *Health) Write(metrics []telegraf.Metric) error {
	healthy := true
	for _, checker := range h.checkers {
		success := checker.Check(metrics)
		if !success {
			healthy = false
		}
	}

	h.setHealthy(healthy)
	return nil
}

// Close shuts down the HTTP server.
func (h *Health) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	h.server.Shutdown(ctx)
	h.wg.Wait()
	return nil
}

// Origin returns the URL of the HTTP server.
func (h *Health) Origin() string {
	return h.origin
}

func (h *Health) getOrigin(listener net.Listener, tlsConf *tls.Config) string {
	switch listener.Addr().Network() {
	case "tcp":
		scheme := "http"
		if tlsConf != nil {
			scheme = "https"
		}
		origin := &url.URL{
			Scheme: scheme,
			Host:   listener.Addr().String(),
		}
		return origin.String()
	case "unix":
		return listener.Addr().String()
	default:
		return ""
	}
}

func (h *Health) setHealthy(healthy bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.healthy = healthy
}

func (h *Health) isHealthy() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.healthy
}

func NewHealth() *Health {
	return &Health{
		ServiceAddress: defaultServiceAddress,
		ReadTimeout:    internal.Duration{Duration: defaultReadTimeout},
		WriteTimeout:   internal.Duration{Duration: defaultWriteTimeout},
		healthy:        true,
	}
}

func init() {
	outputs.Add("health", func() telegraf.Output {
		return NewHealth()
	})
}
