package cloud_pubsub_push

import (
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	tlsint "github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
)

// defaultMaxBodySize is the default maximum request body size, in bytes.
// if the request body is over this size, we will return an HTTP 413 error.
// 500 MB
const defaultMaxBodySize = 500 * 1024 * 1024
const defaultMaxUndeliveredMessages = 1000

type PubSubPush struct {
	ServiceAddress string
	Path           string
	ReadTimeout    internal.Duration
	WriteTimeout   internal.Duration
	MaxBodySize    internal.Size

	MaxUndeliveredMessages int `toml:"max_undelivered_messages"`

	tlsint.ServerConfig
	parsers.Parser

	listener net.Listener
	acc      telegraf.TrackingAccumulator
	ctx      context.Context
	cancel   context.CancelFunc
	wg       *sync.WaitGroup
	mu       sync.Mutex

	undelivered map[telegraf.TrackingID]*http.ResponseWriter
	sem         chan struct{}
	rsem        chan struct{}
}

// Message defines the structure of a Google Pub/Sub message.
type Message struct {
	Atts map[string]string `json:"attributes"`
	Data string            `json:"data"` // Data is base64 encoded data
}

// Payload is the received Google Pub/Sub data. (https://cloud.google.com/pubsub/docs/push)
type Payload struct {
	Msg          Message `json:"message"`
	Subscription string  `json:"subscription"`
}

const sampleConfig = `
  ## Address and port to host HTTP listener on
  service_address = ":8080"

  ## Path to listen to.
  # path = "/"

  ## maximum duration before timing out read of the request
  # read_timeout = "10s"
  ## maximum duration before timing out write of the response
  # write_timeout = "10s"

  ## Maximum allowed http request body size in bytes.
  ## 0 means to use the default of 524,288,00 bytes (500 mebibytes)
  # max_body_size = "500MB"

  ## Optional. Maximum messages to read from PubSub that have not been written
  ## to an output. Defaults to 1000.
  ## For best throughput set based on the number of metrics within
  ## each message and the size of the output's metric_batch_size.
  ##
  ## For example, if each message contains 10 metrics and the output
  ## metric_batch_size is 1000, setting this to 100 will ensure that a
  ## full batch is collected and the write is triggered immediately without
  ## waiting until the next flush_interval.
	# max_undelivered_messages = 1000

  ## Set one or more allowed client CA certificate file names to
  ## enable mutually authenticated TLS connections
  # tls_allowed_cacerts = ["/etc/telegraf/clientca.pem"]

  ## Add service certificate and key
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
`

func (p *PubSubPush) SampleConfig() string {
	return sampleConfig
}

func (p *PubSubPush) Description() string {
	return "Google Cloud Pub/Sub Push HTTP listener"
}

func (p *PubSubPush) Gather(_ telegraf.Accumulator) error {
	return nil
}

func (p *PubSubPush) SetParser(parser parsers.Parser) {
	p.Parser = parser
}

// Start starts the http listener service.
func (p *PubSubPush) Start(acc telegraf.Accumulator) error {
	if p.MaxBodySize.Size == 0 {
		p.MaxBodySize.Size = defaultMaxBodySize
	}

	if p.ReadTimeout.Duration < time.Second {
		p.ReadTimeout.Duration = time.Second * 10
	}
	if p.WriteTimeout.Duration < time.Second {
		p.WriteTimeout.Duration = time.Second * 10
	}

	tlsConf, err := p.ServerConfig.TLSConfig()
	if err != nil {
		return err
	}

	server := &http.Server{
		Addr:         p.ServiceAddress,
		Handler:      p,
		ReadTimeout:  p.ReadTimeout.Duration,
		WriteTimeout: p.WriteTimeout.Duration,
		TLSConfig:    tlsConf,
	}

	if tlsConf != nil {
		p.listener, err = tls.Listen("tcp", p.ServiceAddress, tlsConf)
	} else {
		p.listener, err = net.Listen("tcp", p.ServiceAddress)
	}
	if err != nil {
		return err
	}

	p.ctx, p.cancel = context.WithCancel(context.Background())
	p.wg = &sync.WaitGroup{}
	p.acc = acc.WithTracking(p.MaxUndeliveredMessages)
	p.rsem = make(chan struct{}, p.MaxUndeliveredMessages)
	p.sem = make(chan struct{}, p.MaxUndeliveredMessages)

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		server.Serve(p.listener)
	}()

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		p.receiveDelivered()
	}()

	return nil
}

// Stop cleans up all resources
func (p *PubSubPush) Stop() {
	p.listener.Close()
	p.cancel()
	p.wg.Wait()
}

func (p *PubSubPush) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if req.URL.Path == p.Path {
		p.serveWrite(res, req)
		<-p.rsem
	} else {
		http.NotFound(res, req)
	}
	res.WriteHeader(http.StatusInternalServerError)
}

func (p *PubSubPush) serveWrite(res http.ResponseWriter, req *http.Request) {
	select {
	case <-p.ctx.Done():
		log.Printf("E! [inputs.cloud_pubsub_push] %s", p.ctx.Err())
		p.rsem <- struct{}{}
		return
	case p.sem <- struct{}{}:
		break
	}

	// Check that the content length is not too large for us to handle.
	if req.ContentLength > p.MaxBodySize.Size {
		res.WriteHeader(http.StatusRequestEntityTooLarge)
		p.rsem <- struct{}{}
		return
	}

	if req.Method != "POST" {
		res.WriteHeader(http.StatusMethodNotAllowed)
		p.rsem <- struct{}{}
		return
	}

	// todo: may not be necessary
	// Handle gzip request bodies
	body := req.Body
	if req.Header.Get("Content-Encoding") == "gzip" {
		var err error
		body, err = gzip.NewReader(req.Body)
		if err != nil {
			log.Println("D! [inputs.cloud_pubsub_push] " + err.Error())
			res.WriteHeader(http.StatusBadRequest)
			p.rsem <- struct{}{}
			return
		}
		defer body.Close()
	}

	body = http.MaxBytesReader(res, body, p.MaxBodySize.Size)
	bytes, err := ioutil.ReadAll(body)
	if err != nil {
		res.WriteHeader(http.StatusRequestEntityTooLarge)
		p.rsem <- struct{}{}
		return
	}

	var payload Payload
	if err = json.Unmarshal(bytes, &payload); err != nil {
		log.Printf("E! [inputs.cloud_pubsub_push] Error decoding payload %s", err.Error())
		res.WriteHeader(http.StatusBadRequest)
		p.rsem <- struct{}{}
		return
	}

	sDec, err := base64.StdEncoding.DecodeString(payload.Msg.Data)
	if err != nil {
		log.Printf("E! [inputs.cloud_pubsub_push] Base64-Decode Failed %s", err.Error())
		res.WriteHeader(http.StatusBadRequest)
		p.rsem <- struct{}{}
		return
	}

	metrics, err := p.Parse(sDec)
	if err != nil {
		log.Println("D! [inputs.cloud_pubsub_push] " + err.Error())
		res.WriteHeader(http.StatusBadRequest)
		p.rsem <- struct{}{}
		return
	}

	for i := range metrics {
		for k, v := range payload.Msg.Atts {
			metrics[i].AddTag(k, v)
		}
		metrics[i].AddTag("subscription", payload.Subscription)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	id := p.acc.AddTrackingMetricGroup(metrics)
	if p.undelivered == nil {
		p.undelivered = make(map[telegraf.TrackingID]*http.ResponseWriter)
	}
	p.undelivered[id] = &res
}

func (p *PubSubPush) receiveDelivered() {
	for {
		select {
		case <-p.ctx.Done():
			return
		case info := <-p.acc.Delivered():
			<-p.sem
			res := p.removeDelivered(info.ID())

			if res != nil {
				if info.Delivered() {
					(*res).WriteHeader(http.StatusNoContent)
				} else {
					log.Println("D! [inputs.cloud_pubsub_push] Metric group failed to process")
				}
			}
			p.rsem <- struct{}{}
		}
	}
}

func (p *PubSubPush) removeDelivered(id telegraf.TrackingID) *http.ResponseWriter {
	p.mu.Lock()
	defer p.mu.Unlock()

	msg, ok := p.undelivered[id]
	if !ok {
		return nil
	}

	delete(p.undelivered, id)
	return msg
}

func init() {
	inputs.Add("cloud_pubsub_push", func() telegraf.Input {
		return &PubSubPush{
			ServiceAddress: ":8080",
			Path:           "/",
			MaxUndeliveredMessages: defaultMaxUndeliveredMessages,
		}
	})
}
