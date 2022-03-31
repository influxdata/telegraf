package cloud_pubsub_push

import (
	"context"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	tlsint "github.com/influxdata/telegraf/plugins/common/tls"
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
	Token          string
	Path           string
	ReadTimeout    config.Duration
	WriteTimeout   config.Duration
	MaxBodySize    config.Size
	AddMeta        bool
	Log            telegraf.Logger

	MaxUndeliveredMessages int `toml:"max_undelivered_messages"`

	tlsint.ServerConfig
	parsers.Parser

	server *http.Server
	acc    telegraf.TrackingAccumulator
	ctx    context.Context
	cancel context.CancelFunc
	wg     *sync.WaitGroup
	mu     *sync.Mutex

	undelivered map[telegraf.TrackingID]chan bool
	sem         chan struct{}
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

func (p *PubSubPush) Gather(_ telegraf.Accumulator) error {
	return nil
}

func (p *PubSubPush) SetParser(parser parsers.Parser) {
	p.Parser = parser
}

// Start starts the http listener service.
func (p *PubSubPush) Start(acc telegraf.Accumulator) error {
	if p.MaxBodySize == 0 {
		p.MaxBodySize = config.Size(defaultMaxBodySize)
	}

	if p.ReadTimeout < config.Duration(time.Second) {
		p.ReadTimeout = config.Duration(time.Second * 10)
	}
	if p.WriteTimeout < config.Duration(time.Second) {
		p.WriteTimeout = config.Duration(time.Second * 10)
	}

	tlsConf, err := p.ServerConfig.TLSConfig()
	if err != nil {
		return err
	}

	p.server = &http.Server{
		Addr:        p.ServiceAddress,
		Handler:     http.TimeoutHandler(p, time.Duration(p.WriteTimeout), "timed out processing metric"),
		ReadTimeout: time.Duration(p.ReadTimeout),
		TLSConfig:   tlsConf,
	}

	p.ctx, p.cancel = context.WithCancel(context.Background())
	p.wg = &sync.WaitGroup{}
	p.acc = acc.WithTracking(p.MaxUndeliveredMessages)
	p.sem = make(chan struct{}, p.MaxUndeliveredMessages)
	p.undelivered = make(map[telegraf.TrackingID]chan bool)
	p.mu = &sync.Mutex{}

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		p.receiveDelivered()
	}()

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		if tlsConf != nil {
			if err := p.server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
				p.Log.Errorf("listening and serving TLS failed: %v", err)
			}
		} else {
			if err := p.server.ListenAndServe(); err != nil {
				p.Log.Errorf("listening and serving TLS failed: %v", err)
			}
		}
	}()

	return nil
}

// Stop cleans up all resources
func (p *PubSubPush) Stop() {
	p.cancel()
	//nolint:errcheck,revive // we cannot do anything if the shutdown fails
	p.server.Shutdown(p.ctx)
	p.wg.Wait()
}

func (p *PubSubPush) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if req.URL.Path == p.Path {
		p.AuthenticateIfSet(p.serveWrite, res, req)
	} else {
		p.AuthenticateIfSet(http.NotFound, res, req)
	}
}

func (p *PubSubPush) serveWrite(res http.ResponseWriter, req *http.Request) {
	select {
	case <-req.Context().Done():
		res.WriteHeader(http.StatusServiceUnavailable)
		return
	case <-p.ctx.Done():
		res.WriteHeader(http.StatusServiceUnavailable)
		return
	case p.sem <- struct{}{}:
		break
	}

	// Check that the content length is not too large for us to handle.
	if req.ContentLength > int64(p.MaxBodySize) {
		res.WriteHeader(http.StatusRequestEntityTooLarge)
		return
	}

	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body := http.MaxBytesReader(res, req.Body, int64(p.MaxBodySize))
	bytes, err := io.ReadAll(body)
	if err != nil {
		res.WriteHeader(http.StatusRequestEntityTooLarge)
		return
	}

	var payload Payload
	if err = json.Unmarshal(bytes, &payload); err != nil {
		p.Log.Errorf("Error decoding payload %s", err.Error())
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	sDec, err := base64.StdEncoding.DecodeString(payload.Msg.Data)
	if err != nil {
		p.Log.Errorf("Base64-decode failed %s", err.Error())
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	metrics, err := p.Parse(sDec)
	if err != nil {
		p.Log.Debug(err.Error())
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	if p.AddMeta {
		for i := range metrics {
			for k, v := range payload.Msg.Atts {
				metrics[i].AddTag(k, v)
			}
			metrics[i].AddTag("subscription", payload.Subscription)
		}
	}

	ch := make(chan bool, 1)
	p.mu.Lock()
	p.undelivered[p.acc.AddTrackingMetricGroup(metrics)] = ch
	p.mu.Unlock()

	select {
	case <-req.Context().Done():
		res.WriteHeader(http.StatusServiceUnavailable)
		return
	case success := <-ch:
		if success {
			res.WriteHeader(http.StatusNoContent)
		} else {
			res.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func (p *PubSubPush) receiveDelivered() {
	for {
		select {
		case <-p.ctx.Done():
			return
		case info := <-p.acc.Delivered():
			<-p.sem

			p.mu.Lock()
			ch, ok := p.undelivered[info.ID()]
			if !ok {
				p.mu.Unlock()
				continue
			}

			delete(p.undelivered, info.ID())
			p.mu.Unlock()

			if info.Delivered() {
				ch <- true
			} else {
				ch <- false
				p.Log.Debug("Metric group failed to process")
			}
		}
	}
}

func (p *PubSubPush) AuthenticateIfSet(handler http.HandlerFunc, res http.ResponseWriter, req *http.Request) {
	if p.Token != "" {
		if subtle.ConstantTimeCompare([]byte(req.FormValue("token")), []byte(p.Token)) != 1 {
			http.Error(res, "Unauthorized.", http.StatusUnauthorized)
			return
		}
	}

	handler(res, req)
}

func init() {
	inputs.Add("cloud_pubsub_push", func() telegraf.Input {
		return &PubSubPush{
			ServiceAddress:         ":8080",
			Path:                   "/",
			MaxUndeliveredMessages: defaultMaxUndeliveredMessages,
		}
	})
}
