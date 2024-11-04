//go:generate ../../../tools/readme_config_includer/generator
package firehose

import (
	"context"
	"crypto/tls"
	_ "embed"
	"errors"
	"fmt"
	"net"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	common_tls "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

// Firehose is an input plugin that collects external metrics sent via HTTP from AWS Data Firhose
type Firehose struct {
	ServiceAddress string          `toml:"service_address"`
	Paths          []string        `toml:"paths"`
	ReadTimeout    config.Duration `toml:"read_timeout"`
	WriteTimeout   config.Duration `toml:"write_timeout"`
	AccessKey      config.Secret   `toml:"access_key"`
	ParameterTags  []string        `toml:"parameter_tags"`
	Log            telegraf.Logger `toml:"-"`

	common_tls.ServerConfig
	tlsConf *tls.Config

	once sync.Once

	listener net.Listener
	server   http.Server

	parser telegraf.Parser
	acc    telegraf.Accumulator
}

func (*Firehose) SampleConfig() string {
	return sampleConfig
}

func (*Firehose) Gather(telegraf.Accumulator) error {
	return nil
}

func (f *Firehose) SetParser(parser telegraf.Parser) {
	f.parser = parser
}

func (f *Firehose) Init() error {
	if f.ServiceAddress == "" {
		f.ServiceAddress = ":8080"
	}
	if len(f.Paths) == 0 {
		f.Paths = []string{"/telegraf"}
	}

	var err error
	f.tlsConf, err = f.ServerConfig.TLSConfig()
	return err
}

// Start starts the http listener service.
func (f *Firehose) Start(acc telegraf.Accumulator) error {
	f.acc = acc

	var err error
	if f.tlsConf != nil {
		f.listener, err = tls.Listen("tcp", f.ServiceAddress, f.tlsConf)
	} else {
		f.listener, err = net.Listen("tcp", f.ServiceAddress)
	}
	if err != nil {
		return fmt.Errorf("creating listener failed: %w", err)
	}

	f.server = http.Server{
		Addr:         f.ServiceAddress,
		Handler:      f,
		ReadTimeout:  time.Duration(f.ReadTimeout),
		WriteTimeout: time.Duration(f.WriteTimeout),
		TLSConfig:    f.tlsConf,
	}

	go func() {
		if err := f.server.Serve(f.listener); err != nil {
			if !errors.Is(err, net.ErrClosed) {
				f.Log.Errorf("Server failed: %v", err)
			}
		}
	}()

	f.Log.Infof("Listening on %s", f.listener.Addr().String())

	return nil
}

// Stop cleans up all resources
func (f *Firehose) Stop() {
	if err := f.server.Shutdown(context.Background()); err != nil {
		f.Log.Errorf("Shutting down server failed: %v", err)
	}
}

func (f *Firehose) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if !slices.Contains(f.Paths, req.URL.Path) {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	if err := f.handleRequest(req).sendResponse(res); err != nil {
		f.Log.Errorf("Sending response failed: %v", err)
	}
}

func (f *Firehose) handleRequest(req *http.Request) (r *request) {
	var err error
	if r, err = newFirehoseRequest(req); err != nil {
		f.Log.Errorf("Creating request object failed: %v", err)
		return r
	}

	records, paramTags, err := r.processRequest(f.AccessKey, f.ParameterTags)
	if err != nil {
		f.Log.Errorf("Processing request failed: %v", err)
		return r
	}

	var metrics []telegraf.Metric
	for _, record := range records {
		m, err := f.parser.Parse(record)
		if err != nil {
			// respond with bad request status code to inform firehose about the failure
			r.res.statusCode = http.StatusBadRequest
			f.Log.Errorf("Parse data from request %q failed: %v", r.body.RequestID, err)
			return r
		}
		metrics = append(metrics, m...)
	}

	if len(metrics) == 0 {
		f.once.Do(func() {
			f.Log.Info(internal.NoMetricsCreatedMsg)
		})
	}

	for _, m := range metrics {
		for k, v := range paramTags {
			m.AddTag(k, v)
		}
		m.AddTag("firehose_http_path", r.req.URL.Path)
		f.acc.AddMetric(m)
	}

	r.res.statusCode = http.StatusOK
	return r
}

func init() {
	inputs.Add("firehose", func() telegraf.Input {
		return &Firehose{
			ReadTimeout:  config.Duration(time.Second * 5),
			WriteTimeout: config.Duration(time.Second * 5),
		}
	})
}
