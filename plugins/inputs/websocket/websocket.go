package websocket

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"nhooyr.io/websocket"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/common/proxy"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
)

var sampleConfig = `
  ## URL to read the metrics from (mandatory)
  url = "ws://localhost:8080"

  ## Message to send to the websocket in order to initialize the connection
	## If set to empty (default), nothing will be sent.
  # handshake_body = ""

	## Message to send to the websocket in order to trigger sending of a metric
	## If set to empty (default), this plugin will wait for the server to send
	## messages in an event-based fashion. Otherwise, the content of this option
	## will be sent in each gather interval actively triggering a metric.
  # trigger_body = ""

	## Amount of time allowed to complete a request
  # timeout = "5s"

	## Interval to ping the server
	## This interval has to be at least one second.
  # ping_interval = "1s"

  ## HTTP Proxy support
  # http_proxy_url = ""

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  # data_format = "influx"
`

type Websocket struct {
	URL           string            `toml:"url"`
	HandshakeBody string            `toml:"handshake_body"`
	TriggerBody   string            `toml:"trigger_body"`
	PingInterval  internal.Duration `toml:"ping_interval"`
	Timeout       internal.Duration `toml:"timeout"`
	Log           telegraf.Logger   `toml:"-"`
	tls.ClientConfig
	proxy.HTTPProxy

	client            *http.Client
	connection        *websocket.Conn
	listenCancel      context.CancelFunc
	watchdogReconnect chan bool
	connected         bool
	acc               telegraf.Accumulator

	parser parsers.Parser
}

func (w *Websocket) SampleConfig() string {
	return sampleConfig
}

func (w *Websocket) Description() string {
	return "Read formatted metrics from one or more Websocket endpoints"
}

func (w *Websocket) SetParser(parser parsers.Parser) {
	w.parser = parser
}

func (w *Websocket) Init() error {
	tlsCfg, err := w.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	proxy, err := w.HTTPProxy.Proxy()
	if err != nil {
		return err
	}

	transport := &http.Transport{
		TLSClientConfig: tlsCfg,
		Proxy:           proxy,
	}

	w.client = &http.Client{
		Transport: transport,
	}

	return nil
}

func (w *Websocket) Start(acc telegraf.Accumulator) error {
	w.acc = acc

	if w.PingInterval.Duration < 1*time.Second {
		w.PingInterval = internal.Duration{Duration: 1 * time.Second}
	}

	if err := w.connect(); err != nil {
		return err
	}

	// Start the watchdog
	w.watchdogReconnect = make(chan bool)
	go w.watchdog()

	return nil
}

func (w *Websocket) Stop() {
	w.Log.Debugf("Stopping watchdog...")
	w.disconnect()
}

func (w *Websocket) Gather(acc telegraf.Accumulator) error {
	// Gather is not required if we expect event based readings
	if w.TriggerBody == "" {
		return nil
	}

	// In case we are not connected attempting to connect
	if !w.connected {
		if err := w.connect(); err != nil {
			return err
		}
	}

	ctx := context.Background()
	if w.Timeout.Duration > 0 {
		c, cancel := context.WithTimeout(ctx, w.Timeout.Duration)
		defer cancel()
		ctx = c
	}

	// Trigger the reading and read the metrics
	if err := w.connection.Write(ctx, websocket.MessageText, []byte(w.TriggerBody)); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || websocket.CloseStatus(err) != -1 {
			// The error is related to the connection or a timeout occured. Trigger reconnect.
			w.disconnect()
		}
		return err
	}

	if err := w.read(ctx, acc); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || websocket.CloseStatus(err) != -1 {
			// The error is related to the connection or a timeout occured. Trigger reconnect.
			w.disconnect()
		}
		return err
	}

	return nil
}

func (w *Websocket) connect() error {
	w.Log.Debugf("Connecting to %q", w.URL)

	options := &websocket.DialOptions{
		HTTPClient: w.client,
	}

	dialctx := context.Background()
	if w.Timeout.Duration > 0 {
		c, cancel := context.WithTimeout(dialctx, w.Timeout.Duration)
		defer cancel()
		dialctx = c
	}
	conn, _, err := websocket.Dial(dialctx, w.URL, options)
	if err != nil {
		return err
	}
	w.connection = conn

	if w.TriggerBody == "" {
		w.Log.Debugf("Listening for events...")
		var listenctx context.Context
		listenctx, w.listenCancel = context.WithCancel(context.Background())

		go w.ping(listenctx)
		go w.listen(listenctx, w.acc)
	} else {
		w.Log.Debugf("Actively gathering metrics...")
		w.listenCancel = nil
	}

	if w.HandshakeBody != "" {
		handshakectx := context.Background()
		if w.Timeout.Duration > 0 {
			c, cancel := context.WithTimeout(handshakectx, w.Timeout.Duration)
			defer cancel()
			handshakectx = c
		}
		if err := w.handshake(handshakectx); err != nil {
			return err
		}
	}
	w.connected = true
	w.Log.Infof("Connected to %q", w.URL)

	return nil
}

func (w *Websocket) disconnect() {
	w.Log.Debugf("Disconnecting...")
	w.connected = false

	if w.listenCancel != nil {
		w.listenCancel()
	}

	w.connection.Close(websocket.StatusNormalClosure, "shutdown")
	w.client.CloseIdleConnections()
}

func (w *Websocket) watchdog() {
	for range w.watchdogReconnect {
		w.Log.Infof("Connection problem. Trying to reconnect...")
		w.disconnect()

		for i := 1; ; i++ {
			w.Log.Debugf("Attempting reconnect #%d...", i)
			if err := w.connect(); err == nil {
				break
			}
			retryWait := time.Duration(i) * 100 * time.Millisecond
			if retryWait > w.Timeout.Duration {
				retryWait = w.Timeout.Duration
			}
			time.Sleep(retryWait)
		}
	}
}

func (w *Websocket) ping(ctx context.Context) {
	ticker := time.NewTicker(w.PingInterval.Duration)
	defer ticker.Stop()

	for range ticker.C {
		if err := w.connection.Ping(ctx); err != nil {
			if errors.Is(err, context.Canceled) {
				// We are requested to stop reading
				w.Log.Debug("Stop pinging, got cancel request.")
				return
			}
			if status := websocket.CloseStatus(err); status != -1 {
				w.Log.Infof("Connection closed by peer with status %q", status.String())
			} else {
				w.Log.Errorf("ping failed: %v", err)
			}
			w.watchdogReconnect <- true
			return
		}
	}
}

func (w *Websocket) handshake(ctx context.Context) error {
	return w.connection.Write(ctx, websocket.MessageText, []byte(w.HandshakeBody))
}

func (w *Websocket) listen(ctx context.Context, acc telegraf.Accumulator) {
	for {
		err := w.read(ctx, acc)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				// We are requested to stop reading
				return
			}
			if websocket.CloseStatus(err) != -1 {
				// The error is related to the connection
				return
			}
			acc.AddError(err)
			continue
		}
	}
}

func (w *Websocket) read(ctx context.Context, acc telegraf.Accumulator) error {
	_, buf, err := w.connection.Read(ctx)
	if err != nil {
		return err
	}
	metrics, err := w.parser.Parse(buf)
	if err != nil {
		return fmt.Errorf("parsing failed: %v", err)
	}

	for _, metric := range metrics {
		if !metric.HasTag("url") {
			metric.AddTag("url", w.URL)
		}
		acc.AddMetric(metric)
	}

	return nil
}

func init() {
	inputs.Add("http", func() telegraf.Input {
		return &Websocket{
			Timeout: internal.Duration{Duration: 5 * time.Second},
		}
	})
}
