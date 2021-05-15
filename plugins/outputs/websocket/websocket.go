package websocket

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf/testutil"

	ws "github.com/gorilla/websocket"
)

var sampleConfig = `
  ## URL is the address to send metrics to. Make sure ws or wss scheme is used.
  url = "ws://127.0.0.1:8080/telegraf"

  ## Timeouts (make sure read_timeout is larger than server ping interval or set to zero).
  # connect_timeout = "30s"
  # write_timeout = "30s"
  # read_timeout = "30s"

  ## Optionally turn on using text data frames (binary by default).
  # use_text_frames = false

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Data format to output.
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  # data_format = "influx"

  ## Additional HTTP Upgrade headers
  # [outputs.websocket.headers]
  #   Authorization = "Bearer <TOKEN>"
`

const (
	defaultConnectTimeout = 30 * time.Second
	defaultWriteTimeout   = 30 * time.Second
	defaultReadTimeout    = 30 * time.Second
)

// WebSocket can output to WebSocket endpoint.
type WebSocket struct {
	URL            string            `toml:"url"`
	ConnectTimeout config.Duration   `toml:"connect_timeout"`
	WriteTimeout   config.Duration   `toml:"write_timeout"`
	ReadTimeout    config.Duration   `toml:"read_timeout"`
	Headers        map[string]string `toml:"headers"`
	UseTextFrames  bool              `toml:"use_text_frames"`
	Logger         telegraf.Logger   `toml:"-"`
	tls.ClientConfig

	conn       *ws.Conn
	serializer serializers.Serializer
}

// SetSerializer implements serializers.SerializerOutput.
func (w *WebSocket) SetSerializer(serializer serializers.Serializer) {
	w.serializer = serializer
}

// Description of plugin.
func (w *WebSocket) Description() string {
	return "Generic WebSocket output writer."
}

// SampleConfig returns plugin config sample.
func (w *WebSocket) SampleConfig() string {
	return sampleConfig
}

var errInvalidURL = errors.New("invalid websocket URL")

// Init the output plugin.
func (w *WebSocket) Init() error {
	if parsedURL, err := url.Parse(w.URL); err != nil || (parsedURL.Scheme != "ws" && parsedURL.Scheme != "wss") {
		return fmt.Errorf("%w: \"%s\"", errInvalidURL, w.URL)
	}
	return nil
}

// Connect to the output endpoint.
func (w *WebSocket) Connect() error {
	tlsCfg, err := w.ClientConfig.TLSConfig()
	if err != nil {
		return fmt.Errorf("error creating TLS config: %v", err)
	}

	dialer := &ws.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: time.Duration(w.ConnectTimeout),
		TLSClientConfig:  tlsCfg,
	}

	headers := http.Header{}
	for k, v := range w.Headers {
		headers.Set(k, v)
	}

	conn, resp, err := dialer.Dial(w.URL, headers)
	if err != nil {
		return fmt.Errorf("error dial: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusSwitchingProtocols {
		return fmt.Errorf("wrong status code while connecting to server: %d", resp.StatusCode)
	}

	w.conn = conn
	go w.read(conn)

	return nil
}

func (w *WebSocket) read(conn *ws.Conn) {
	defer func() { _ = conn.Close() }()
	if w.ReadTimeout > 0 {
		if err := conn.SetReadDeadline(time.Now().Add(time.Duration(w.ReadTimeout))); err != nil {
			return
		}
		conn.SetPingHandler(func(string) error {
			err := conn.SetReadDeadline(time.Now().Add(time.Duration(w.ReadTimeout)))
			if err != nil {
				return err
			}
			return conn.WriteControl(ws.PongMessage, nil, time.Now().Add(time.Duration(w.WriteTimeout)))
		})
	}
	for {
		// Need to read a connection (to properly process pings from a server).
		_, _, err := conn.ReadMessage()
		if err != nil {
			if ws.IsUnexpectedCloseError(err, ws.CloseGoingAway, ws.CloseAbnormalClosure) {
				w.Logger.Errorf("error reading websocket connection: %v", err)
			}
			return
		}
		if w.ReadTimeout > 0 {
			if err := conn.SetReadDeadline(time.Now().Add(time.Duration(w.ReadTimeout))); err != nil {
				return
			}
		}
	}
}

// Write writes the given metrics to the destination. Not thread-safe.
func (w *WebSocket) Write(metrics []telegraf.Metric) error {
	if w.conn == nil {
		// Previous write failed with error and ws conn was closed.
		if err := w.Connect(); err != nil {
			return err
		}
	}

	messageData, err := w.serializer.SerializeBatch(metrics)
	if err != nil {
		return err
	}

	messageType := ws.BinaryMessage
	if w.UseTextFrames {
		messageType = ws.TextMessage
	}

	if w.WriteTimeout > 0 {
		if err := w.conn.SetWriteDeadline(time.Now().Add(time.Duration(w.WriteTimeout))); err != nil {
			return fmt.Errorf("error setting write deadline: %v", err)
		}
	}
	err = w.conn.WriteMessage(messageType, messageData)
	if err != nil {
		_ = w.conn.Close()
		w.conn = nil
		return fmt.Errorf("error writing to connection: %v", err)
	}
	return nil
}

// Close closes the connection. Noop if already closed.
func (w *WebSocket) Close() error {
	if w.conn == nil {
		return nil
	}
	err := w.conn.Close()
	w.conn = nil
	return err
}

func newWebSocket() *WebSocket {
	return &WebSocket{
		ConnectTimeout: config.Duration(defaultConnectTimeout),
		WriteTimeout:   config.Duration(defaultWriteTimeout),
		ReadTimeout:    config.Duration(defaultReadTimeout),
		Logger:         testutil.Logger{},
	}
}

func init() {
	outputs.Add("websocket", func() telegraf.Output {
		return newWebSocket()
	})
}
