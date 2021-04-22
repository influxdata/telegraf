package websocket

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"

	"github.com/gorilla/websocket"
)

var sampleConfig = `
  ## URL is the address to send metrics to. Make sure ws or wss scheme is used.
  url = "ws://127.0.0.1:8080/telegraf"

  ## Timeouts (make sure read_timeout is larger than server ping interval or set to zero).
  # connect_timeout = "5s"
  # write_timeout = "5s"
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
  # [outputs.http.headers]
  #   # Should be set manually to "application/json" for json data_format
  #   Authorization = "Bearer <TOKEN>"
`

const (
	defaultConnectTimeout = 5 * time.Second
	defaultWriteTimeout   = 5 * time.Second
	defaultReadTimeout    = 30 * time.Second
)

type WebSocket struct {
	URL            string            `toml:"url"`
	ConnectTimeout config.Duration   `toml:"connect_timeout"`
	WriteTimeout   config.Duration   `toml:"write_timeout"`
	ReadTimeout    config.Duration   `toml:"read_timeout"`
	Headers        map[string]string `toml:"headers"`
	UseTextFrames  bool              `toml:"use_text_frames"`
	tls.ClientConfig

	conn       *websocket.Conn
	serializer serializers.Serializer
}

func (w *WebSocket) SetSerializer(serializer serializers.Serializer) {
	w.serializer = serializer
}

func (w *WebSocket) Description() string {
	return "Generic WebSocket output writer."
}

func (w *WebSocket) SampleConfig() string {
	return sampleConfig
}

func (w *WebSocket) Connect() error {
	if parsedURL, err := url.Parse(w.URL); err != nil || (parsedURL.Scheme != "ws" && parsedURL.Scheme != "wss") {
		return fmt.Errorf("valid websocket URL expected, got \"%s\"", w.URL)
	}

	tlsCfg, err := w.ClientConfig.TLSConfig()
	if err != nil {
		return fmt.Errorf("error creating TLS config: %v", err)
	}

	dialer := &websocket.Dialer{
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
	if resp.StatusCode != http.StatusSwitchingProtocols {
		return fmt.Errorf("wrong status code while connecting to server: %d", resp.StatusCode)
	}

	w.conn = conn
	go w.read(conn)
	return nil
}

func (w *WebSocket) read(conn *websocket.Conn) {
	if w.ReadTimeout > 0 {
		_ = conn.SetReadDeadline(time.Now().Add(time.Duration(w.ReadTimeout)))
		conn.SetPingHandler(func(string) error {
			_ = conn.SetReadDeadline(time.Now().Add(time.Duration(w.ReadTimeout)))
			_ = conn.WriteControl(websocket.PongMessage, nil, time.Now().Add(time.Duration(w.WriteTimeout)))
			return nil
		})
	}
	for {
		// Just read connection (required to process pings from a server).
		_, _, err := conn.ReadMessage()
		if err != nil {
			// Return from read. If this was caused by read deadline then
			// eventually connection may be closed by a server since client
			// stops answering on pings.
			return
		}
	}
}

// Write writes the given metrics to the destination. Not thread-safe.
func (w *WebSocket) Write(metrics []telegraf.Metric) error {
	if w.conn == nil {
		// Previous write failed with permanent error and ws conn was closed.
		if err := w.Connect(); err != nil {
			return err
		}
	}

	messageData, err := w.serializer.SerializeBatch(metrics)
	if err != nil {
		return err
	}

	messageType := websocket.BinaryMessage
	if w.UseTextFrames {
		messageType = websocket.TextMessage
	}

	if w.WriteTimeout > 0 {
		_ = w.conn.SetWriteDeadline(time.Now().Add(time.Duration(w.WriteTimeout)))
	}
	err = w.conn.WriteMessage(messageType, messageData)
	if err != nil {
		_ = w.conn.Close()
		w.conn = nil
		return fmt.Errorf("error writing to connection: %v", err)
	}
	if w.WriteTimeout > 0 {
		_ = w.conn.SetWriteDeadline(time.Time{})
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

func init() {
	outputs.Add("websocket", func() telegraf.Output {
		return &WebSocket{
			ConnectTimeout: config.Duration(defaultConnectTimeout),
			WriteTimeout:   config.Duration(defaultWriteTimeout),
			ReadTimeout:    config.Duration(defaultReadTimeout),
		}
	})
}
