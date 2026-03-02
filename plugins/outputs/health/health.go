//go:generate ../../../tools/readme_config_includer/generator
package health

import (
	"context"
	"crypto/tls"
	_ "embed"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	common_tls "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
)

//go:embed sample.conf
var sampleConfig string

const (
	defaultServiceAddress = "tcp://:8080"
	defaultReadTimeout    = 5 * time.Second
	defaultWriteTimeout   = 5 * time.Second
)

type Checker interface {
	// Check returns true if the metrics meet its criteria.
	Check(metrics []telegraf.Metric) bool
}

type Health struct {
	ServiceAddress        string          `toml:"service_address"`
	ReadTimeout           config.Duration `toml:"read_timeout"`
	WriteTimeout          config.Duration `toml:"write_timeout"`
	BasicUsername         string          `toml:"basic_username"`
	BasicPassword         string          `toml:"basic_password"`
	Compares              []*Compares     `toml:"compares"`
	Contains              []*Contains     `toml:"contains"`
	MaxTimeBetweenMetrics config.Duration `toml:"max_time_between_metrics"`
	DefaultStatus         int             `toml:"default_status"`
	Log                   telegraf.Logger `toml:"-"`
	common_tls.ServerConfig

	checkers []Checker

	wg      sync.WaitGroup
	server  *http.Server
	origin  string
	network string
	address string
	tlsConf *tls.Config

	mu               sync.Mutex
	lastMetricTime   time.Time
	healthy          bool
	metricsAvailable bool
}

func (*Health) SampleConfig() string {
	return sampleConfig
}

func (h *Health) Init() error {
	u, err := url.Parse(h.ServiceAddress)
	if err != nil {
		return err
	}

	switch u.Scheme {
	case "http", "https":
		h.network = "tcp"
		h.address = u.Host
	case "unix":
		h.network = u.Scheme
		h.address = u.Path
	case "tcp4", "tcp6", "tcp":
		h.network = u.Scheme
		h.address = u.Host
	default:
		return errors.New("service_address contains invalid scheme")
	}

	if h.DefaultStatus == 0 {
		h.DefaultStatus = http.StatusOK
	} else if h.DefaultStatus < 0 || h.DefaultStatus > 599 {
		return fmt.Errorf("invalid default HTTP status code %d", h.DefaultStatus)
	}

	h.tlsConf, err = h.ServerConfig.TLSConfig()
	if err != nil {
		return err
	}

	h.checkers = make([]Checker, 0)
	for i := range h.Compares {
		h.checkers = append(h.checkers, h.Compares[i])
	}
	for i := range h.Contains {
		h.checkers = append(h.checkers, h.Contains[i])
	}

	return nil
}

// Connect starts the HTTP server.
func (h *Health) Connect() error {
	authHandler := internal.BasicAuthHandler(h.BasicUsername, h.BasicPassword, "health", onAuthError)

	h.server = &http.Server{
		Addr:         h.ServiceAddress,
		Handler:      authHandler(h),
		ReadTimeout:  time.Duration(h.ReadTimeout),
		WriteTimeout: time.Duration(h.WriteTimeout),
		TLSConfig:    h.tlsConf,
	}

	listener, err := h.listen()
	if err != nil {
		return err
	}

	h.origin = h.getOrigin(listener)

	h.Log.Infof("Listening on %s", h.origin)

	// Initialize lastMetricTime here to fail if no metrics are received
	// before the configured max timeout.
	h.lastMetricTime = time.Now()
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		err := h.server.Serve(listener)
		if err != http.ErrServerClosed {
			h.Log.Errorf("Serve error on %s: %v", h.origin, err)
		}
		h.origin = ""
	}()

	return nil
}

func onAuthError(_ http.ResponseWriter) {}

func (h *Health) listen() (net.Listener, error) {
	if h.tlsConf != nil {
		return tls.Listen(h.network, h.address, h.tlsConf)
	}
	return net.Listen(h.network, h.address)
}

func (h *Health) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()

	w.Header().Set("Server", internal.ProductToken())

	// Check the timeout independent of the available metrics
	if h.MaxTimeBetweenMetrics > 0 && time.Since(h.lastMetricTime) >= time.Duration(h.MaxTimeBetweenMetrics) {
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}

	// Return the default status if we have no metrics to check
	if !h.metricsAvailable {
		http.Error(w, http.StatusText(h.DefaultStatus), h.DefaultStatus)
		return
	}

	// Check the health conditions and return 503 - Service Unavailabe for
	// unhealthy states
	if !h.healthy {
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}

	http.Error(w, http.StatusText(http.StatusOK), http.StatusOK)
}

// Write runs all checks over the metric batch and adjust health state.
func (h *Health) Write(metrics []telegraf.Metric) error {
	ts := time.Now()
	healthy := true
	for _, checker := range h.checkers {
		success := checker.Check(metrics)
		if !success {
			healthy = false
			break
		}
	}

	// healthy only represents the result of the configured checkers and not
	// the MaxTimeBetweenMetrics validation. The timeout check is done when
	// serving the HTTP response.
	h.mu.Lock()
	defer h.mu.Unlock()
	h.lastMetricTime = ts
	h.healthy = healthy
	h.metricsAvailable = true

	return nil
}

// Close shuts down the HTTP server.
func (h *Health) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := h.server.Shutdown(ctx)
	h.wg.Wait()
	return err
}

// Origin returns the URL of the HTTP server.
func (h *Health) Origin() string {
	return h.origin
}

func (h *Health) getOrigin(listener net.Listener) string {
	scheme := "http"
	if h.tlsConf != nil {
		scheme = "https"
	}
	if h.network == "unix" {
		scheme = "unix"
	}

	switch h.network {
	case "unix":
		origin := &url.URL{
			Scheme: scheme,
			Path:   listener.Addr().String(),
		}
		return origin.String()
	default:
		origin := &url.URL{
			Scheme: scheme,
			Host:   listener.Addr().String(),
		}
		return origin.String()
	}
}

func NewHealth() *Health {
	return &Health{
		ServiceAddress: defaultServiceAddress,
		ReadTimeout:    config.Duration(defaultReadTimeout),
		WriteTimeout:   config.Duration(defaultWriteTimeout),
	}
}

func init() {
	outputs.Add("health", func() telegraf.Output {
		return NewHealth()
	})
}
