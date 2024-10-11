//go:generate ../../../tools/readme_config_includer/generator
package firehose

import (
	"crypto/tls"
	_ "embed"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/choice"
	tlsint "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

var allowedMethods = []string{http.MethodPost, http.MethodPut}
var statusCodeToMessage = map[int]string{
	http.StatusBadRequest:            "bad request",
	http.StatusMethodNotAllowed:      "method not allowed",
	http.StatusRequestEntityTooLarge: "request body too large",
	http.StatusUnauthorized:          "unauthorized",
	http.StatusOK:                    "",
}

// Firehose is an input plugin that collects external metrics sent via HTTP from AWS Data Firhose
type Firehose struct {
	ServiceAddress string          `toml:"service_address"`
	Paths          []string        `toml:"paths"`
	PathTag        bool            `toml:"path_tag"`
	ReadTimeout    config.Duration `toml:"read_timeout"`
	WriteTimeout   config.Duration `toml:"write_timeout"`
	AccessKey      config.Secret   `toml:"access_key"`
	ParameterTags  []string        `toml:"parameter_tags"`

	tlsint.ServerConfig
	tlsConf *tls.Config

	once sync.Once
	Log  telegraf.Logger

	wg    sync.WaitGroup
	close chan struct{}

	listener net.Listener

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
		f.ReadTimeout = config.Duration(time.Second * 10)
	}
	if f.WriteTimeout < config.Duration(time.Second) {
		f.WriteTimeout = config.Duration(time.Second * 10)
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

	server := &http.Server{
		Addr:         f.ServiceAddress,
		Handler:      f,
		ReadTimeout:  time.Duration(f.ReadTimeout),
		WriteTimeout: time.Duration(f.WriteTimeout),
		TLSConfig:    f.tlsConf,
	}

	f.wg.Add(1)
	go func() {
		defer f.wg.Done()
		if err := server.Serve(f.listener); err != nil {
			if !errors.Is(err, net.ErrClosed) {
				f.Log.Errorf("starting server failed: %v", err)
			}
			close(f.close)
		}
	}()

	f.Log.Infof("Listening on %s", f.listener.Addr().String())

	return nil
}

// Stop cleans up all resources
func (f *Firehose) Stop() {
	if f.listener != nil {
		f.listener.Close()
	}
	f.wg.Wait()
}

func (f *Firehose) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if !choice.Contains(req.URL.Path, f.Paths) {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	r := &firehoseRequest{req: req}
	requestID := req.Header.Get("x-amz-firehose-request-id")

	if err := r.authenticate(f.AccessKey); err != nil {
		f.Log.Error(err.Error())
		if err = r.sendResponse(res); err != nil {
			f.Log.Errorf("error sending response to request %s, %v", requestID, err.Error())
		}
		return
	}

	if err := r.validate(); err != nil {
		f.Log.Error(err.Error())
		if err = r.sendResponse(res); err != nil {
			f.Log.Errorf("error sending response to request %s, %v", requestID, err.Error())
		}
		return
	}

	decodedBytesData, ok := r.decodeData()
	if !ok {
		f.Log.Errorf("failed to base64 decode record data from request %s", requestID)
		if err := r.sendResponse(res); err != nil {
			f.Log.Errorf("error sending response to request %s, %v", requestID, err.Error())
		}
		return
	}

	var metrics []telegraf.Metric
	for _, bytes := range decodedBytesData {
		m, err := f.parser.Parse(bytes)
		if err != nil {
			f.Log.Errorf("unable to parse data from request %s", requestID)
			// respond with bad request status code to inform firehose about the failure
			r.responseStatusCode = http.StatusBadRequest
			if err = r.sendResponse(res); err != nil {
				f.Log.Errorf("error sending response to request %s, %v", requestID, err.Error())
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
			f.Log.Warnf("x-amz-firehose-common-attributes header's value is not a valid json in request %s", requestID)
		}

		parameters, ok := parameters["commonAttributes"].(map[string]interface{})
		if !ok {
			f.Log.Warnf("Invalid value for header x-amz-firehose-common-attributes in request %s", requestID)
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

	if f.PathTag {
		for _, m := range metrics {
			m.AddTag("firehose_http_path", req.URL.Path)
		}
	}

	for _, m := range metrics {
		f.acc.AddMetric(m)
	}

	r.responseStatusCode = http.StatusOK
	if err := r.sendResponse(res); err != nil {
		f.Log.Errorf("error sending response to request %s, %v", requestID, err.Error())
	}
}

func init() {
	inputs.Add("firehose", func() telegraf.Input {
		return &Firehose{
			ServiceAddress: ":8080",
			Paths:          []string{"/telegraf"},
			close:          make(chan struct{}),
		}
	})
}
