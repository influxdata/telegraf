package googlecoreiot

import (
	"compress/gzip"
	"crypto/subtle"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	tlsint "github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
	influx "github.com/influxdata/telegraf/plugins/parsers/influx"
)

// Attributes maps Default fields sent by Google Pub/Sub
type Attributes struct {
	DeviceID               string `json:"deviceId"`
	DeviceNumID            string `json:"deviceNumId"`
	DeviceRegistryID       string `json:"deviceRegistryId"`
	DeviceRegistryLocation string `json:"deviceRegistryLocation"`
	ProjectID              string `json:"projectId"`
	SubFolder              string `json:"subFolder"`
}

// Message Structure of a Google Pub/Sub message
type Message struct {
	Atts         map[string]interface{} `json:"attributes"`
	Data         string                 `json:"data"`
	MessageID    string                 `json:"messageId"`
	MessageID2   string                 `json:"message_id"`
	PublishTime  string                 `json:"publishTime"`
	PublishTime2 string                 `json:"publish_time"`
}

// Payload of Line-Protocol payload after Base64-decode
type Payload struct {
	Msg          Message `json:"message"`
	Subscription string  `json:"subscription"`
}

// JSONData structure of the Base-64 encoded payload
type JSONData struct {
	Name   string                 `json:"measurement"`
	Tags   map[string]string      `json:"tags"`
	Fields map[string]interface{} `json:"fields"`
	Time   int64                  `json:"time"`
}

// DEFAULT_MAX_BODY_SIZE is the default maximum request body size, in bytes.
// if the request body is over this size, we will return an HTTP 413 error.
// 500 MB
const defaultMaxBodySize = 500 * 1024 * 1024

type TimeFunc func() time.Time

type GoogleListener struct {
	ServiceAddress  string
	Path            string
	Methods         []string
	ReadTimeout     internal.Duration
	WriteTimeout    internal.Duration
	MaxBodySize     internal.Size
	Port            int
	Precision       string
	DataFormat      string
	MeasurementName string
	handler         *influx.MetricHandler

	tlsint.ServerConfig

	BasicUsername string
	BasicPassword string

	TimeFunc

	wg sync.WaitGroup

	listener net.Listener

	parsers.Parser
	acc telegraf.Accumulator
}

const sampleConfig = `
[[inputs.googlecoreiot]]
  ## Address and port to host HTTP listener on
  # service_address = ":9999"

  ## Path to serve
  ## default is /write
  # path = "/write"

  ## HTTP methods to accept.
  # methods = ["POST", "PUT"]

  ## maximum duration before timing out read of the request
  # read_timeout = "10s"
  ## maximum duration before timing out write of the response
  # write_timeout = "10s"

  # precision of the time stamps. can be one of the following:
  # second
  # millisecond
  # microsecond
  # nanosecond
  # Default is nanosecond
  
  # precision = "nanosecond"
  
  # Data Format is either influx or json
  # data_format="influx" 

  ## Set one or more allowed client CA certificate file names to 
  ## enable mutually authenticated TLS connections
  tls_allowed_cacerts = ["/etc/telegraf/clientca.pem"]

  ## Add service certificate and key
  tls_cert = "/etc/telegraf/cert.pem"
  tls_key = "/etc/telegraf/key.pem"
  ## Optional username and password to accept for HTTP basic authentication.
  ## You probably want to make sure you have TLS configured above for this.
  # basic_username = "foobar"
  # basic_password = "barfoo"
`

func (h *GoogleListener) SampleConfig() string {
	return sampleConfig
}

func (h *GoogleListener) Description() string {
	return "Influx Google Pub/Sub write listener"
}

func (h *GoogleListener) Gather(_ telegraf.Accumulator) error {
	return nil
}
func (h *GoogleListener) SetParser(parser parsers.Parser) {
	h.Parser = parser
}

// Start starts the http listener service.
func (h *GoogleListener) Start(acc telegraf.Accumulator) error {
	if h.MaxBodySize.Size == 0 {
		h.MaxBodySize.Size = defaultMaxBodySize
	}

	if h.ReadTimeout.Duration < time.Second {
		h.ReadTimeout.Duration = time.Second * 10
	}
	if h.WriteTimeout.Duration < time.Second {
		h.WriteTimeout.Duration = time.Second * 10
	}
	if h.MeasurementName == "" {
		h.MeasurementName = "Core_IoT"
	}
	if h.DataFormat == "" {
		h.DataFormat = "influx"
	}
	if h.Path == "" {
		h.Path = "/write"
	}

	h.acc = acc

	tlsConf, err := h.ServerConfig.TLSConfig()
	if err != nil {
		return err
	}

	server := &http.Server{
		Addr:         h.ServiceAddress,
		Handler:      h,
		ReadTimeout:  h.ReadTimeout.Duration,
		WriteTimeout: h.WriteTimeout.Duration,
		TLSConfig:    tlsConf,
	}

	var listener net.Listener
	if tlsConf != nil {
		listener, err = tls.Listen("tcp", h.ServiceAddress, tlsConf)
	} else {
		listener, err = net.Listen("tcp", h.ServiceAddress)
	}
	if err != nil {
		return err
	}
	h.listener = listener
	h.Port = listener.Addr().(*net.TCPAddr).Port
	switch h.Precision {

	case "second":
		h.handler.SetTimePrecision(time.Second)
		break
	case "microsecond":
		h.handler.SetTimePrecision(time.Microsecond)
		break
	case "nanosecond":
		h.handler.SetTimePrecision(time.Nanosecond)
		break
	default:
		h.handler.SetTimePrecision(time.Millisecond)
		break
	}
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		server.Serve(h.listener)
	}()

	log.Printf("I! Started Google Core IoT service on %s\n", h.ServiceAddress)

	return nil
}

/// Stop cleans up all resources
func (h *GoogleListener) Stop() {
	h.listener.Close()
	h.wg.Wait()

	log.Println("I! Stopped Google Core IoT service on ", h.ServiceAddress)
}

func (h *GoogleListener) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if req.URL.Path == h.Path {
		h.AuthenticateIfSet(h.serveWrite, res, req)
	} else {
		h.AuthenticateIfSet(http.NotFound, res, req)
	}
}

func (h *GoogleListener) decodeLineProtocol(payload []byte, obj Payload) ([]telegraf.Metric, error) {
	var metrics []telegraf.Metric
	var config = parsers.Config{}
	config.DataFormat = h.DataFormat
	parser, err := parsers.NewParser(&config)
	if err != nil {
		log.Println("E! Parser ", err)
		return metrics, err

	}
	metrics, err = parser.Parse(payload)
	if err != nil {
		log.Println("E! Parser ", err)
		return metrics, err
	}

	for key, value := range obj.Msg.Atts {
		for _, m := range metrics {
			m.AddTag(key, value.(string))
			m.AddTag("message_id", obj.Msg.MessageID)
			m.AddTag("message_id_2", obj.Msg.MessageID2)
			m.AddTag("subscription", obj.Subscription)
		}
	}
	return metrics, nil
}

func (h *GoogleListener) decodeJSON(payload []byte, obj Payload) (telegraf.Metric, error) {

	e := JSONData{}
	err := json.Unmarshal(payload, &e)
	if err != nil {
		log.Println("E! JSON Unmarshall ", err)
		return h.handler.Metric()
	}
	b := []byte(strconv.FormatInt(e.Time, 10))
	h.handler.SetTimestamp(b)
	metrics, err := h.handler.Metric()
	if err != nil {
		log.Println("E! Metrics ", err)
		return h.handler.Metric()
	}
	metrics.SetName(e.Name)
	for key, value := range e.Fields {
		metrics.AddField(key, value)
	}

	for key, value := range obj.Msg.Atts {

		metrics.AddTag(key, value.(string))

	}
	metrics.AddTag("message_id", obj.Msg.MessageID)
	metrics.AddTag("message_id_2", obj.Msg.MessageID2)
	metrics.AddTag("subscription", obj.Subscription)
	return metrics, nil

}

func (h *GoogleListener) serveNotFound(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")
	res.Header().Set("X-Influxdb-Version", "1.0")
	res.WriteHeader(http.StatusNotFound)
	res.Write([]byte(`{"error":"http: not found"}`))
	return
}

func (h *GoogleListener) serveWrite(res http.ResponseWriter, req *http.Request) {
	// Check that the content length is not too large for us to handle.
	if req.ContentLength > h.MaxBodySize.Size {
		tooLarge(res)
		return
	}

	// Check if the requested HTTP method was specified in config.
	isAcceptedMethod := false
	for _, method := range h.Methods {
		if req.Method == method {
			isAcceptedMethod = true
			break
		}
	}
	if !isAcceptedMethod {
		methodNotAllowed(res)
		return
	}

	// Handle gzip request bodies
	body := req.Body
	if req.Header.Get("Content-Encoding") == "gzip" {
		var err error
		body, err = gzip.NewReader(req.Body)
		if err != nil {
			log.Println("D! " + err.Error())
			badRequest(res)
			return
		}
		defer body.Close()
	}

	body = http.MaxBytesReader(res, body, h.MaxBodySize.Size)

	decoder := json.NewDecoder(body)
	var t Payload
	err := decoder.Decode(&t)
	if err != nil {
		log.Println("E! Failed to decode Payload", err)
		badRequest(res)
		return
	}
	sDec, err := base64.StdEncoding.DecodeString(t.Msg.Data)
	if err != nil {
		log.Println("E! Base64-Decode Failed" + err.Error())
		badRequest(res)
		return
	}
	input := strings.Split(string(sDec), "\n")
	x := 0
	for x < len(input) {
		if h.DataFormat == "influx" {
			metrics, err := h.decodeLineProtocol([]byte(input[x]), t)
			if err != nil {
				log.Println("E! Line Protocol Decode failed " + err.Error())
				badRequest(res)
				return
			}

			for _, m := range metrics {
				h.acc.AddFields(m.Name(), m.Fields(), m.Tags(), m.Time())
			}
		}
		if h.DataFormat == "json" {
			metrics, err := h.decodeJSON([]byte(input[x]), t)
			if err != nil {
				log.Println("E! JSON Decode Failed " + err.Error())
				badRequest(res)
				return
			}
			h.acc.AddFields(metrics.Name(), metrics.Fields(), metrics.Tags(), metrics.Time())
		}
		x++
	}

	res.WriteHeader(http.StatusNoContent)
}

func tooLarge(res http.ResponseWriter) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusRequestEntityTooLarge)
	res.Write([]byte(`{"error":"http: request body too large"}`))
}

func methodNotAllowed(res http.ResponseWriter) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusMethodNotAllowed)
	res.Write([]byte(`{"error":"http: method not allowed"}`))
}

func internalServerError(res http.ResponseWriter) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusInternalServerError)
}

func badRequest(res http.ResponseWriter) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusBadRequest)
	res.Write([]byte(`{"error":"http: bad request"}`))
}

func (h *GoogleListener) AuthenticateIfSet(handler http.HandlerFunc, res http.ResponseWriter, req *http.Request) {
	if h.BasicUsername != "" && h.BasicPassword != "" {
		reqUsername, reqPassword, ok := req.BasicAuth()
		if !ok ||
			subtle.ConstantTimeCompare([]byte(reqUsername), []byte(h.BasicUsername)) != 1 ||
			subtle.ConstantTimeCompare([]byte(reqPassword), []byte(h.BasicPassword)) != 1 {

			http.Error(res, "Unauthorized.", http.StatusUnauthorized)
			return
		}
		handler(res, req)
	} else {
		handler(res, req)
	}
}

func init() {
	inputs.Add("googlecoreiot", func() telegraf.Input {
		return &GoogleListener{
			ServiceAddress: ":9999",
			TimeFunc:       time.Now,
			Path:           "/write",
			Methods:        []string{"POST", "PUT"},
			handler:        influx.NewMetricHandler(),
		}
	})
}
