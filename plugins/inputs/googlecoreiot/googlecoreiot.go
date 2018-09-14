package googlecoreiot

import (
	"compress/gzip"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	tlsint "github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	influx "github.com/influxdata/telegraf/plugins/parsers/influx"

	"github.com/influxdata/telegraf/selfstat"
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

const (
	// DEFAULT_MAX_BODY_SIZE is the default maximum request body size, in bytes.
	// if the request body is over this size, we will return an HTTP 413 error.
	// 500 MB
	DEFAULT_MAX_BODY_SIZE = 500 * 1024 * 1024

	// DEFAULT_MAX_LINE_SIZE is the maximum size, in bytes, that can be allocated for
	// a single InfluxDB point.
	// 64 KB
	DEFAULT_MAX_LINE_SIZE = 64 * 1024
)

type TimeFunc func() time.Time

type HTTPListener struct {
	ServiceAddress  string
	ReadTimeout     internal.Duration
	WriteTimeout    internal.Duration
	MaxBodySize     int64
	MaxLineSize     int
	Port            int
	MeasurementName string
	Precision       string
	Protocol        string

	tlsint.ServerConfig

	TimeFunc

	mu sync.Mutex
	wg sync.WaitGroup

	listener net.Listener

	handler *influx.MetricHandler
	acc     telegraf.Accumulator
	pool    *pool

	BytesRecv       selfstat.Stat
	RequestsServed  selfstat.Stat
	WritesServed    selfstat.Stat
	QueriesServed   selfstat.Stat
	PingsServed     selfstat.Stat
	RequestsRecv    selfstat.Stat
	WritesRecv      selfstat.Stat
	QueriesRecv     selfstat.Stat
	PingsRecv       selfstat.Stat
	NotFoundsServed selfstat.Stat
	BuffersCreated  selfstat.Stat
	AuthFailures    selfstat.Stat
}

const sampleConfig = `
  ## Address and port to host HTTP listener on
  service_address = ":9999"

  ## maximum duration before timing out read of the request
  read_timeout = "10s"
  ## maximum duration before timing out write of the response
  write_timeout = "10s"

  # precision of the time stamps. can be one of the following:
  # second
  # millisecond
  # microsecond
  # nanosecond
  # Default is nanosecond
  
  precision = "nanosecond"
  
  # Data Format is either line protocol or json
  protocol="line protocol" 

  ## Set one or more allowed client CA certificate file names to 
  ## enable mutually authenticated TLS connections
  tls_allowed_cacerts = ["/etc/telegraf/clientca.pem"]

  ## Add service certificate and key
  tls_cert = "/etc/telegraf/cert.pem"
  tls_key = "/etc/telegraf/key.pem"

`

func (h *HTTPListener) SampleConfig() string {
	return sampleConfig
}

func (h *HTTPListener) Description() string {
	return "Influx HTTP write listener"
}

func (h *HTTPListener) Gather(_ telegraf.Accumulator) error {
	h.BuffersCreated.Set(h.pool.ncreated())
	return nil
}

// Start starts the http listener service.
func (h *HTTPListener) Start(acc telegraf.Accumulator) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	tags := map[string]string{
		"address": h.ServiceAddress,
	}
	h.BytesRecv = selfstat.Register("http_listener", "bytes_received", tags)
	h.RequestsServed = selfstat.Register("http_listener", "requests_served", tags)
	h.WritesServed = selfstat.Register("http_listener", "writes_served", tags)
	h.QueriesServed = selfstat.Register("http_listener", "queries_served", tags)
	h.PingsServed = selfstat.Register("http_listener", "pings_served", tags)
	h.RequestsRecv = selfstat.Register("http_listener", "requests_received", tags)
	h.WritesRecv = selfstat.Register("http_listener", "writes_received", tags)
	h.QueriesRecv = selfstat.Register("http_listener", "queries_received", tags)
	h.PingsRecv = selfstat.Register("http_listener", "pings_received", tags)
	h.NotFoundsServed = selfstat.Register("http_listener", "not_founds_served", tags)
	h.BuffersCreated = selfstat.Register("http_listener", "buffers_created", tags)
	h.AuthFailures = selfstat.Register("http_listener", "auth_failures", tags)

	if h.MaxBodySize == 0 {
		h.MaxBodySize = DEFAULT_MAX_BODY_SIZE
	}
	if h.MaxLineSize == 0 {
		h.MaxLineSize = DEFAULT_MAX_LINE_SIZE
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
	if h.Protocol == "" {
		h.Protocol = "line protocol"
	}

	h.acc = acc
	h.pool = NewPool(200, h.MaxLineSize)

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

	h.handler = influx.NewMetricHandler()

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

	log.Printf("I! Started HTTP listener service on %s\n", h.ServiceAddress)
	return nil
}

// Stop cleans up all resources
func (h *HTTPListener) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.listener.Close()
	h.wg.Wait()

	log.Println("I! Stopped HTTP listener service on ", h.ServiceAddress)
}

func (h *HTTPListener) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	h.RequestsRecv.Incr(1)
	defer h.RequestsServed.Incr(1)
	switch req.URL.Path {
	case "/write":
		h.WritesRecv.Incr(1)
		defer h.WritesServed.Incr(1)
	default:
		defer h.NotFoundsServed.Incr(1)
	}
}

func (h *HTTPListener) decodeLineProtocol(payload []byte, obj Payload) ([]telegraf.Metric, error) {
	parser := influx.NewParser(h.handler)
	metrics, err := parser.Parse(payload)
	if err != nil {
		log.Println("E! ", err)
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

func (h *HTTPListener) decodeJSON(payload []byte, obj Payload) (telegraf.Metric, error) {

	e := JSONData{}
	err := json.Unmarshal(payload, &e)
	if err != nil {
		log.Println("E! ", err)
		return h.handler.Metric()
	}
	b := []byte(strconv.FormatInt(e.Time, 10))
	h.handler.SetTimestamp(b)
	metrics, err := h.handler.Metric()
	if err != nil {
		log.Println("E! ", err)
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
func (h *HTTPListener) serveWrite(res http.ResponseWriter, req *http.Request) {
	// Check that the content length is not too large for us to handle.
	if req.ContentLength > h.MaxBodySize {
		tooLarge(res)
		return
	}
	//now := h.TimeFunc()

	// Handle gzip request bodies
	body := req.Body
	if req.Header.Get("Content-Encoding") == "gzip" {
		var err error
		body, err = gzip.NewReader(req.Body)
		defer body.Close()
		if err != nil {
			log.Println("E! " + err.Error())
			badRequest(res)
			return
		}
	}
	body = http.MaxBytesReader(res, body, h.MaxBodySize)

	buf := h.pool.get()
	defer h.pool.put(buf)
	decoder := json.NewDecoder(req.Body)

	var t Payload
	err := decoder.Decode(&t)
	if err != nil {
		log.Println("E! ", err)
	}
	sDec, err := base64.StdEncoding.DecodeString(t.Msg.Data)
	if err != nil {
		log.Println("E! " + err.Error())
		badRequest(res)
		return
	}
	if h.Protocol == "line protocol" {
		metrics, err := h.decodeLineProtocol(sDec, t)
		if err != nil {
			log.Println("E! " + err.Error())
			badRequest(res)
			return
		}
		res.WriteHeader(http.StatusNoContent)
		for _, m := range metrics {
			h.acc.AddFields(m.Name(), m.Fields(), m.Tags(), m.Time())
		}
	}
	if h.Protocol == "json" {
		metrics, err := h.decodeJSON(sDec, t)
		if err != nil {
			log.Println("E! " + err.Error())
			badRequest(res)
			return
		}
		res.WriteHeader(http.StatusNoContent)
		h.acc.AddFields(metrics.Name(), metrics.Fields(), metrics.Tags(), metrics.Time())
	}

}

func (h *HTTPListener) parse(b []byte, t time.Time, precision string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.handler.SetTimeFunc(func() time.Time { return t })
	return nil
}

func tooLarge(res http.ResponseWriter) {
	res.Header().Set("Content-Type", "application/json")
	res.Header().Set("X-Influxdb-Version", "1.0")
	res.WriteHeader(http.StatusRequestEntityTooLarge)
	res.Write([]byte(`{"error":"http: request body too large"}`))
}

func badRequest(res http.ResponseWriter) {
	res.Header().Set("Content-Type", "application/json")
	res.Header().Set("X-Influxdb-Version", "1.0")
	res.WriteHeader(http.StatusBadRequest)
	res.Write([]byte(`{"error":"http: bad request"}`))
}

func init() {
	inputs.Add("google_core_iot", func() telegraf.Input {
		return &HTTPListener{
			ServiceAddress: ":9999",
			TimeFunc:       time.Now,
		}
	})
}
