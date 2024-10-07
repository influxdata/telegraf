//go:generate ../../../tools/readme_config_includer/generator
package firehose

import (
	"compress/gzip"
	"crypto/tls"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
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

var once sync.Once

var allowedMethods = []string{http.MethodPost, http.MethodPut}
var statusCodeToMessage = map[int]string{
	http.StatusBadRequest:            "bad request",
	http.StatusMethodNotAllowed:      "method not allowed",
	http.StatusRequestEntityTooLarge: "request body too large",
	http.StatusUnauthorized:          "unauthorized",
	http.StatusOK:                    "",
}

// defaultMaxBodySize is the default maximum request body size, in bytes.
// if the request body is over this size, we will return an HTTP 413 error.
// 64 MB
const defaultMaxBodySize = 64 * 1024 * 1024

const (
	pathTag   = "firehose_http_path"
	logFormat = "RequestID:%s Message:%s"

	// request headers
	requestIDHeader        = "x-amz-firehose-request-id"
	accessKeyHeader        = "x-amz-firehose-access-key"
	commonAttributesHeader = "x-amz-firehose-common-attributes"
)

// TimeFunc provides a timestamp for the metrics
type TimeFunc func() time.Time

type firehoseRecord struct {
	EncodedData string `json:"data"`
}

type firehoseRequestBody struct {
	RequestID string           `json:"requestId"`
	Timestamp int64            `json:"timestamp"`
	Records   []firehoseRecord `json:"records"`
}

// https://docs.aws.amazon.com/firehose/latest/dev/httpdeliveryrequestresponse.html#requestformat
type firehoseRequest struct {
	req                *http.Request
	body               firehoseRequestBody
	responseStatusCode int
}

func (r *firehoseRequest) authenticate(expectedAccessKey string) bool {
	reqAccessKey := r.req.Header.Get(accessKeyHeader)
	if reqAccessKey != expectedAccessKey {
		r.responseStatusCode = http.StatusUnauthorized
		return false
	}
	return true
}

func (r *firehoseRequest) validate() error {
	// Check that the content length is not too large for us to handle.
	if r.req.ContentLength > int64(defaultMaxBodySize) {
		r.responseStatusCode = http.StatusRequestEntityTooLarge
		return errors.New("request body too large")
	}

	// Check if the requested HTTP method is allowed.
	isAcceptedMethod := false
	for _, method := range allowedMethods {
		if r.req.Method == method {
			isAcceptedMethod = true
			break
		}
	}
	if !isAcceptedMethod {
		r.responseStatusCode = http.StatusMethodNotAllowed
		return fmt.Errorf("%s method not allowed", r.req.Method)
	}

	contentType := r.req.Header.Get("content-type")
	if contentType != "application/json" {
		r.responseStatusCode = http.StatusBadRequest
		return fmt.Errorf("%s content type not accepted", contentType)
	}

	contentEncoding := r.req.Header.Get("content-encoding")
	if contentEncoding != "" && contentEncoding != "gzip" {
		r.responseStatusCode = http.StatusBadRequest
		return fmt.Errorf("%s content encoding not accepted", contentEncoding)
	}

	err := r.extractBody()
	if err != nil {
		return err
	}

	requestID := r.req.Header.Get(requestIDHeader)
	if requestID == "" {
		r.responseStatusCode = http.StatusBadRequest
		return fmt.Errorf("%s header is not set", requestIDHeader)
	}

	if requestID != r.body.RequestID {
		r.responseStatusCode = http.StatusBadRequest
		return fmt.Errorf("requestId in the body does not match the value of the request header, %s", requestIDHeader)
	}

	return nil
}

func (r *firehoseRequest) extractBody() error {
	encoding := r.req.Header.Get("content-encoding")
	switch encoding {
	case "gzip":
		g, err := gzip.NewReader(r.req.Body)
		if err != nil {
			r.responseStatusCode = http.StatusBadRequest
			return fmt.Errorf("unable to decode body - %s", err.Error())
		}
		defer g.Close()
		err = json.NewDecoder(g).Decode(&r.body)
		if err != nil {
			r.responseStatusCode = http.StatusBadRequest
			return err
		}
	default:
		defer r.req.Body.Close()
		err := json.NewDecoder(r.req.Body).Decode(&r.body)
		if err != nil {
			r.responseStatusCode = http.StatusBadRequest
			return err
		}
	}
	return nil
}

func (r *firehoseRequest) decodeData() ([][]byte, bool) {
	// decode base64-encoded data and return them as a slice of byte slices
	decodedData := make([][]byte, 0)
	for _, record := range r.body.Records {
		data, err := base64.StdEncoding.DecodeString(record.EncodedData)
		if err != nil {
			return nil, false
		}
		decodedData = append(decodedData, data)
	}
	return decodedData, true
}

func (r *firehoseRequest) sendResponse(res http.ResponseWriter) error {
	responseBody := struct {
		RequestID    string `json:"requestId"`
		Timestamp    int64  `json:"timestamp"`
		ErrorMessage string `json:"errorMessage,omitempty"`
	}{
		RequestID:    r.req.Header.Get(requestIDHeader),
		Timestamp:    time.Now().Unix(),
		ErrorMessage: statusCodeToMessage[r.responseStatusCode],
	}
	response, err := json.Marshal(responseBody)
	if err != nil {
		return err
	}
	res.Header().Set("content-type", "application/json")
	res.WriteHeader(r.responseStatusCode)
	_, err = res.Write(response)
	return err
}

// Firehose is an input plugin that collects external metrics sent via HTTP from AWS Data Firhose
type Firehose struct {
	ServiceAddress string          `toml:"service_address"`
	Paths          []string        `toml:"paths"`
	PathTag        bool            `toml:"path_tag"`
	ReadTimeout    config.Duration `toml:"read_timeout"`
	WriteTimeout   config.Duration `toml:"write_timeout"`
	AccessKey      string          `toml:"access_key"`
	ParameterTags  []string        `toml:"parameter_tags"`

	tlsint.ServerConfig
	tlsConf *tls.Config

	TimeFunc
	Log telegraf.Logger

	wg    sync.WaitGroup
	close chan struct{}

	listener net.Listener

	telegraf.Parser
	acc telegraf.Accumulator
}

func (*Firehose) SampleConfig() string {
	return sampleConfig
}

func (f *Firehose) Gather(_ telegraf.Accumulator) error {
	return nil
}

func (f *Firehose) SetParser(parser telegraf.Parser) {
	f.Parser = parser
}

// Start starts the http listener service.
func (f *Firehose) Start(acc telegraf.Accumulator) error {
	if f.ReadTimeout < config.Duration(time.Second) {
		f.ReadTimeout = config.Duration(time.Second * 10)
	}
	if f.WriteTimeout < config.Duration(time.Second) {
		f.WriteTimeout = config.Duration(time.Second * 10)
	}

	f.acc = acc

	server := f.createHTTPServer()

	f.wg.Add(1)
	go func() {
		defer f.wg.Done()
		if err := server.Serve(f.listener); err != nil {
			if !errors.Is(err, net.ErrClosed) {
				f.Log.Errorf("Serve failed - %s", err.Error())
			}
			close(f.close)
		}
	}()

	f.Log.Infof("Listening on %s", f.listener.Addr().String())

	return nil
}

func (f *Firehose) createHTTPServer() *http.Server {
	return &http.Server{
		Addr:         f.ServiceAddress,
		Handler:      f,
		ReadTimeout:  time.Duration(f.ReadTimeout),
		WriteTimeout: time.Duration(f.WriteTimeout),
		TLSConfig:    f.tlsConf,
	}
}

// Stop cleans up all resources
func (f *Firehose) Stop() {
	if f.listener != nil {
		f.listener.Close()
	}
	f.wg.Wait()
}

func (f *Firehose) Init() error {
	tlsConf, err := f.ServerConfig.TLSConfig()
	if err != nil {
		return err
	}

	var listener net.Listener
	if tlsConf != nil {
		listener, err = tls.Listen("tcp", f.ServiceAddress, tlsConf)
	} else {
		listener, err = net.Listen("tcp", f.ServiceAddress)
	}
	if err != nil {
		return err
	}
	f.tlsConf = tlsConf
	f.listener = listener

	return nil
}

func (f *Firehose) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	select {
	case <-f.close:
		res.WriteHeader(http.StatusGone)
		return
	default:
	}

	if !choice.Contains(req.URL.Path, f.Paths) {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	r := &firehoseRequest{req: req}

	if f.AccessKey != "" {
		if ok := r.authenticate(f.AccessKey); !ok {
			f.Log.Error(formatLog(req.Header.Get(requestIDHeader), "Unauthorized"))
			err := r.sendResponse(res)
			if err != nil {
				f.Log.Error(formatLog(req.Header.Get(requestIDHeader), err.Error()))
			}
			return
		}
	}

	if err := r.validate(); err != nil {
		f.Log.Error(formatLog(req.Header.Get(requestIDHeader), err.Error()))
		err := r.sendResponse(res)
		if err != nil {
			f.Log.Error(formatLog(req.Header.Get(requestIDHeader), err.Error()))
		}
		return
	}

	decodedBytesData, ok := r.decodeData()
	if !ok {
		f.Log.Error(formatLog(r.body.RequestID, "Failed to base64 decode record data"))
		err := r.sendResponse(res)
		if err != nil {
			f.Log.Error(formatLog(req.Header.Get(requestIDHeader), err.Error()))
		}
		return
	}

	metrics := make([]telegraf.Metric, 0)
	for _, bytes := range decodedBytesData {
		m, err := f.Parse(bytes)
		if err != nil {
			f.Log.Error(formatLog(r.body.RequestID, "Unable to parse data"))
			r.responseStatusCode = http.StatusBadRequest
			err := r.sendResponse(res)
			if err != nil {
				f.Log.Error(formatLog(req.Header.Get(requestIDHeader), err.Error()))
			}
			return
		}
		metrics = append(metrics, m...)
	}

	if len(metrics) == 0 {
		once.Do(func() {
			f.Log.Info(formatLog(r.body.RequestID, internal.NoMetricsCreatedMsg))
		})
	}

	commonAttributesHeaderValue := req.Header.Get(commonAttributesHeader)
	if len(commonAttributesHeaderValue) != 0 && len(f.ParameterTags) != 0 {
		parameters := make(map[string]interface{})
		err := json.Unmarshal([]byte(commonAttributesHeaderValue), &parameters)
		if err != nil {
			f.Log.Warn(formatLog(r.body.RequestID, commonAttributesHeader+" header's value is not a valid json"))
		}

		parameters, ok := parameters["commonAttributes"].(map[string]interface{})
		if !ok {
			f.Log.Warn(formatLog(r.body.RequestID, "Invalid value for header "+commonAttributesHeader))
		}

		for _, parameter := range f.ParameterTags {
			if value, ok := parameters[parameter]; ok {
				for _, m := range metrics {
					m.AddTag(parameter, value.(string))
				}
			}
		}
	}

	for _, m := range metrics {
		if f.PathTag {
			m.AddTag(pathTag, req.URL.Path)
		}
		f.acc.AddMetric(m)
	}

	r.responseStatusCode = http.StatusOK
	err := r.sendResponse(res)
	if err != nil {
		f.Log.Error(formatLog(req.Header.Get(requestIDHeader), err.Error()))
	}
}

func formatLog(requestID, message string) string {
	return fmt.Sprintf(logFormat, requestID, message)
}

func init() {
	inputs.Add("firehose", func() telegraf.Input {
		return &Firehose{
			ServiceAddress: ":8080",
			TimeFunc:       time.Now,
			Paths:          []string{"/telegraf"},
			close:          make(chan struct{}),
		}
	})
}
