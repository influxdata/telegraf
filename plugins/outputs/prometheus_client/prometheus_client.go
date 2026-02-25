//go:generate ../../../tools/readme_config_includer/generator
package prometheus_client

import (
	"context"
	"crypto/tls"
	_ "embed"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mdlayher/vsock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	common_tls "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/outputs/prometheus_client/v1"
	"github.com/influxdata/telegraf/plugins/outputs/prometheus_client/v2"
	serializers_prometheus "github.com/influxdata/telegraf/plugins/serializers/prometheus"
)

//go:embed sample.conf
var sampleConfig string

const (
	defaultListen             = ":9273"
	defaultPath               = "/metrics"
	defaultExpirationInterval = config.Duration(60 * time.Second)
	defaultReadTimeout        = 10 * time.Second
	defaultWriteTimeout       = 10 * time.Second
	defaultNameSanitization   = "legacy"
)

type Collector interface {
	Describe(ch chan<- *prometheus.Desc)
	Collect(ch chan<- prometheus.Metric)
	Add(metrics []telegraf.Metric) error
}

type PrometheusClient struct {
	Listen             string                             `toml:"listen"`
	ReadTimeout        config.Duration                    `toml:"read_timeout"`
	WriteTimeout       config.Duration                    `toml:"write_timeout"`
	MetricVersion      int                                `toml:"metric_version"`
	BasicUsername      string                             `toml:"basic_username"`
	BasicPassword      config.Secret                      `toml:"basic_password"`
	IPRange            []string                           `toml:"ip_range"`
	ExpirationInterval config.Duration                    `toml:"expiration_interval"`
	Path               string                             `toml:"path"`
	CollectorsExclude  []string                           `toml:"collectors_exclude"`
	StringAsLabel      bool                               `toml:"string_as_label"`
	ExportTimestamp    bool                               `toml:"export_timestamp"`
	TypeMappings       serializers_prometheus.MetricTypes `toml:"metric_types"`
	NameSanitization   string                             `toml:"name_sanitization"`
	HTTPHeaders        map[string]*config.Secret          `toml:"http_headers"`
	Log                telegraf.Logger                    `toml:"-"`

	common_tls.ServerConfig

	server    *http.Server
	url       *url.URL
	collector Collector
	wg        sync.WaitGroup
}

func (*PrometheusClient) SampleConfig() string {
	return sampleConfig
}

func (p *PrometheusClient) Init() error {
	defaultCollectors := map[string]bool{
		"gocollector": true,
		"process":     true,
	}
	for _, collector := range p.CollectorsExclude {
		delete(defaultCollectors, collector)
	}

	registry := prometheus.NewRegistry()
	for collector := range defaultCollectors {
		switch collector {
		case "gocollector":
			err := registry.Register(collectors.NewGoCollector())
			if err != nil {
				return err
			}
		case "process":
			err := registry.Register(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("unrecognized collector %s", collector)
		}
	}

	switch p.NameSanitization {
	case "":
		p.NameSanitization = defaultNameSanitization
	case "legacy", "utf8":
		// Valid sanitization modes.
	default:
		return fmt.Errorf("invalid name_sanitization %q: must be \"legacy\" or \"utf8\"", p.NameSanitization)
	}

	if err := p.TypeMappings.Init(); err != nil {
		return err
	}

	switch p.MetricVersion {
	default:
		fallthrough
	case 1:
		p.collector = v1.NewCollector(
			time.Duration(p.ExpirationInterval),
			p.StringAsLabel,
			p.ExportTimestamp,
			p.TypeMappings,
			p.Log,
			p.NameSanitization,
		)
		err := registry.Register(p.collector)
		if err != nil {
			return err
		}
	case 2:
		p.collector = v2.NewCollector(
			time.Duration(p.ExpirationInterval),
			p.StringAsLabel,
			p.ExportTimestamp,
			p.TypeMappings,
			p.NameSanitization,
		)
		err := registry.Register(p.collector)
		if err != nil {
			return err
		}
	}

	ipRange := make([]*net.IPNet, 0, len(p.IPRange))
	for _, cidr := range p.IPRange {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			return fmt.Errorf("error parsing ip_range: %w", err)
		}

		ipRange = append(ipRange, ipNet)
	}

	psecret, err := p.BasicPassword.Get()
	if err != nil {
		return err
	}
	password := psecret.String()
	psecret.Destroy()

	authHandler := internal.BasicAuthHandler(p.BasicUsername, password, "prometheus", onAuthError)
	rangeHandler := internal.IPRangeHandler(ipRange, onError)
	promHandler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError})
	landingPageHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte("Telegraf Output Plugin: Prometheus Client "))
		if err != nil {
			p.Log.Errorf("Error occurred when writing HTTP reply: %v", err)
		}
	})

	mux := http.NewServeMux()
	if p.Path == "" {
		p.Path = "/metrics"
	}
	mux.Handle(p.Path, p.headerHandler(authHandler(rangeHandler(promHandler))))
	mux.Handle("/", p.headerHandler(authHandler(rangeHandler(landingPageHandler))))

	tlsConfig, err := p.TLSConfig()
	if err != nil {
		return err
	}

	if p.ReadTimeout < config.Duration(time.Second) {
		p.ReadTimeout = config.Duration(defaultReadTimeout)
	}
	if p.WriteTimeout < config.Duration(time.Second) {
		p.WriteTimeout = config.Duration(defaultWriteTimeout)
	}

	p.server = &http.Server{
		Addr:         p.Listen,
		Handler:      mux,
		TLSConfig:    tlsConfig,
		ReadTimeout:  time.Duration(p.ReadTimeout),
		WriteTimeout: time.Duration(p.WriteTimeout),
	}

	return nil
}

func (p *PrometheusClient) listenTCP(host string) (net.Listener, error) {
	if p.server.TLSConfig != nil {
		return tls.Listen("tcp", host, p.server.TLSConfig)
	}
	return net.Listen("tcp", host)
}

func listenVsock(host string) (net.Listener, error) {
	_, portStr, err := net.SplitHostPort(host)
	if err != nil {
		return nil, err
	}
	port, err := strconv.ParseUint(portStr, 10, 32)
	if err != nil {
		return nil, err
	}
	return vsock.Listen(uint32(port), nil)
}

func (p *PrometheusClient) listen() (net.Listener, error) {
	u, err := url.ParseRequestURI(p.Listen)
	// fallback to legacy way
	if err != nil {
		return p.listenTCP(p.Listen)
	}
	switch strings.ToLower(u.Scheme) {
	case "", "tcp", "http":
		return p.listenTCP(u.Host)
	case "vsock":
		return listenVsock(u.Host)
	default:
		return p.listenTCP(u.Host)
	}
}

func (p *PrometheusClient) Connect() error {
	listener, err := p.listen()
	if err != nil {
		return err
	}

	scheme := "http"
	if p.server.TLSConfig != nil {
		scheme = "https"
	}

	p.url = &url.URL{
		Scheme: scheme,
		Host:   listener.Addr().String(),
		Path:   p.Path,
	}

	p.Log.Infof("Listening on %s", p.URL())

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		err := p.server.Serve(listener)
		if err != nil && err != http.ErrServerClosed {
			p.Log.Errorf("Server error: %v", err)
		}
	}()

	return nil
}

func (p *PrometheusClient) headerHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for key, secret := range p.HTTPHeaders {
			value, err := secret.Get()
			if err == nil {
				w.Header().Set(key, value.String())
			}
		}
		next.ServeHTTP(w, r)
	})
}

func onAuthError(_ http.ResponseWriter) {
}

func onError(rw http.ResponseWriter, code int) {
	http.Error(rw, http.StatusText(code), code)
}

// URL returns the address the plugin is listening on.  If not listening
// an empty string is returned.
func (p *PrometheusClient) URL() string {
	if p.url != nil {
		return p.url.String()
	}
	return ""
}

func (p *PrometheusClient) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := p.server.Shutdown(ctx)
	p.wg.Wait()
	p.url = nil
	prometheus.Unregister(p.collector)
	return err
}

func (p *PrometheusClient) Write(metrics []telegraf.Metric) error {
	return p.collector.Add(metrics)
}

func init() {
	outputs.Add("prometheus_client", func() telegraf.Output {
		return &PrometheusClient{
			Listen:             defaultListen,
			Path:               defaultPath,
			ExpirationInterval: defaultExpirationInterval,
			StringAsLabel:      true,
			NameSanitization:   defaultNameSanitization,
		}
	})
}
