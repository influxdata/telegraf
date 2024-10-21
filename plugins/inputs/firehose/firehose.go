//go:generate ../../../tools/readme_config_includer/generator
package firehose

import (
	"context"
	"crypto/tls"
	_ "embed"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	tlsint "github.com/influxdata/telegraf/plugins/common/tls"
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

	tlsint.ServerConfig
	tlsConf *tls.Config

	once sync.Once
	Log  telegraf.Logger

	listener net.Listener
	server   http.Server

	parser telegraf.Parser
	acc    telegraf.Accumulator
}

func (*Firehose) SampleConfig() string {
	return sampleConfig
}

func (f *Firehose) Gather(_ telegraf.Accumulator) error {
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
	if f.ReadTimeout < config.Duration(time.Second) {
		f.ReadTimeout = config.Duration(time.Second * 5)
	}
	if f.WriteTimeout < config.Duration(time.Second) {
		f.WriteTimeout = config.Duration(time.Second * 5)
	}

	var err error
	f.tlsConf, err = f.ServerConfig.TLSConfig()
	if err != nil {
		return err
	}

	return nil
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
		return err
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
				f.Log.Errorf("starting server failed: %v", err)
			}
		}
	}()

	f.Log.Infof("listening on %s", f.listener.Addr().String())

	return nil
}

// Stop cleans up all resources
func (f *Firehose) Stop() {
	err := f.server.Shutdown(context.Background())
	if err != nil {
		f.Log.Infof("shutting down server failed: %v", err)
	}
}

func (f *Firehose) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if !slices.Contains(f.Paths, req.URL.Path) {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	requestID := req.Header.Get("x-amz-firehose-request-id")
	r := &request{req: req}

	if err := r.authenticate(f.AccessKey); err != nil {
		f.Log.Errorf("authentication failed: %v", err)
		if err = r.sendResponse(res); err != nil {
			f.Log.Errorf("sending response to request %q failed: %v", requestID, err)
		}
		return
	}

	if err := r.validate(); err != nil {
		f.Log.Errorf("validation failed: %v", err)
		if err = r.sendResponse(res); err != nil {
			f.Log.Errorf("sending response to request %q failed: %v", requestID, err)
		}
		return
	}

	data, err := r.decodeData()
	if err != nil {
		f.Log.Errorf("decode base64 data from request %q failed: %v", requestID, err)
		if err = r.sendResponse(res); err != nil {
			f.Log.Errorf("sending response to request %q failed: %v", requestID, err)
		}
		return
	}

	var metrics []telegraf.Metric
	for _, bytes := range data {
		m, err := f.parser.Parse(bytes)
		if err != nil {
			f.Log.Errorf("parse data from request %q failed: %v", requestID, err)
			// respond with bad request status code to inform firehose about the failure
			r.res.statusCode = http.StatusBadRequest
			if err = r.sendResponse(res); err != nil {
				f.Log.Errorf("sending response to request %q failed: %v", requestID, err)
			}
			return
		}
		metrics = append(metrics, m...)
	}

	if len(metrics) == 0 {
		f.once.Do(func() {
			f.Log.Info(internal.NoMetricsCreatedMsg)
		})
		return
	}

	attributesHeader := req.Header.Get("x-amz-firehose-common-attributes")
	if len(attributesHeader) != 0 && len(f.ParameterTags) != 0 {
		var parameters map[string]interface{}
		if err := json.Unmarshal([]byte(attributesHeader), &parameters); err != nil {
			f.Log.Warnf("x-amz-firehose-common-attributes header's value is not a valid json in request %q", requestID)
		}

		parameters, ok := parameters["commonAttributes"].(map[string]interface{})
		if !ok {
			f.Log.Warnf("invalid value for header x-amz-firehose-common-attributes in request %q", requestID)
		} else {
			for _, parameter := range f.ParameterTags {
				if value, ok := parameters[parameter]; ok {
					for _, m := range metrics {
						m.AddTag(parameter, value.(string))
					}
				}
			}
		}
	}

	for _, m := range metrics {
		m.AddTag("firehose_http_path", req.URL.Path)
	}

	for _, m := range metrics {
		f.acc.AddMetric(m)
	}

	r.res.statusCode = http.StatusOK
	if err := r.sendResponse(res); err != nil {
		f.Log.Errorf("sending response to request %q failed: %v", requestID, err)
	}
}

func init() {
	inputs.Add("firehose", func() telegraf.Input {
		return &Firehose{
			ServiceAddress: ":8080",
			Paths:          []string{"/telegraf"},
		}
	})
}
