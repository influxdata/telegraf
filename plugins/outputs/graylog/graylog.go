package graylog

import (
	"bytes"
	"compress/zlib"
	"crypto/rand"
	"crypto/tls"
	"encoding/binary"
	ejson "encoding/json"
	"fmt"
	"io"
	"math"
	"net"
	"os"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	tlsint "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
)

const (
	defaultEndpoint        = "127.0.0.1:12201"
	defaultConnection      = "wan"
	defaultMaxChunkSizeWan = 1420
	defaultMaxChunkSizeLan = 8154
	defaultScheme          = "udp"
	defaultTimeout         = 5 * time.Second
)

type gelfConfig struct {
	Endpoint        string
	Connection      string
	MaxChunkSizeWan int
	MaxChunkSizeLan int
}

type gelf interface {
	io.WriteCloser
	Connect() error
}

type gelfCommon struct {
	gelfConfig
	dialer *net.Dialer
	conn   net.Conn
}

type gelfUDP struct {
	gelfCommon
}

type gelfTCP struct {
	gelfCommon
	tlsConfig *tls.Config
}

func newGelfWriter(cfg gelfConfig, dialer *net.Dialer, tlsConfig *tls.Config) gelf {
	if cfg.Endpoint == "" {
		cfg.Endpoint = defaultEndpoint
	}

	if cfg.Connection == "" {
		cfg.Connection = defaultConnection
	}

	if cfg.MaxChunkSizeWan == 0 {
		cfg.MaxChunkSizeWan = defaultMaxChunkSizeWan
	}

	if cfg.MaxChunkSizeLan == 0 {
		cfg.MaxChunkSizeLan = defaultMaxChunkSizeLan
	}

	scheme := defaultScheme
	parts := strings.SplitN(cfg.Endpoint, "://", 2)
	if len(parts) == 2 {
		scheme = strings.ToLower(parts[0])
		cfg.Endpoint = parts[1]
	}
	common := gelfCommon{
		gelfConfig: cfg,
		dialer:     dialer,
	}

	var g gelf
	switch scheme {
	case "tcp":
		g = &gelfTCP{gelfCommon: common, tlsConfig: tlsConfig}
	default:
		g = &gelfUDP{gelfCommon: common}
	}

	return g
}

func (g *gelfUDP) Write(message []byte) (n int, err error) {
	compressed := g.compress(message)

	chunksize := g.gelfConfig.MaxChunkSizeWan
	length := compressed.Len()

	if length > chunksize {
		chunkCountInt := int(math.Ceil(float64(length) / float64(chunksize)))

		id := make([]byte, 8)
		rand.Read(id)

		for i, index := 0, 0; i < length; i, index = i+chunksize, index+1 {
			packet := g.createChunkedMessage(index, chunkCountInt, id, &compressed)
			err = g.send(packet.Bytes())
			if err != nil {
				return 0, err
			}
		}
	} else {
		err = g.send(compressed.Bytes())
		if err != nil {
			return 0, err
		}
	}

	n = len(message)

	return n, nil
}

func (g *gelfUDP) Close() (err error) {
	if g.conn != nil {
		err = g.conn.Close()
		g.conn = nil
	}

	return err
}

func (g *gelfUDP) createChunkedMessage(index int, chunkCountInt int, id []byte, compressed *bytes.Buffer) bytes.Buffer {
	var packet bytes.Buffer

	chunksize := g.getChunksize()

	packet.Write(g.intToBytes(30))
	packet.Write(g.intToBytes(15))
	packet.Write(id)

	packet.Write(g.intToBytes(index))
	packet.Write(g.intToBytes(chunkCountInt))

	packet.Write(compressed.Next(chunksize))

	return packet
}

func (g *gelfUDP) getChunksize() int {
	if g.gelfConfig.Connection == "wan" {
		return g.gelfConfig.MaxChunkSizeWan
	}

	if g.gelfConfig.Connection == "lan" {
		return g.gelfConfig.MaxChunkSizeLan
	}

	return g.gelfConfig.MaxChunkSizeWan
}

func (g *gelfUDP) intToBytes(i int) []byte {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, int8(i))
	return buf.Bytes()
}

func (g *gelfUDP) compress(b []byte) bytes.Buffer {
	var buf bytes.Buffer
	comp := zlib.NewWriter(&buf)

	comp.Write(b)
	comp.Close()

	return buf
}

func (g *gelfUDP) Connect() error {
	conn, err := g.dialer.Dial("udp", g.gelfConfig.Endpoint)
	if err != nil {
		return err
	}
	g.conn = conn
	return nil
}

func (g *gelfUDP) send(b []byte) error {
	if g.conn == nil {
		err := g.Connect()
		if err != nil {
			return err
		}
	}

	_, err := g.conn.Write(b)
	if err != nil {
		_ = g.conn.Close()
		g.conn = nil
	}

	return err
}

func (g *gelfTCP) Write(message []byte) (n int, err error) {
	err = g.send(message)
	if err != nil {
		return 0, err
	}

	n = len(message)

	return n, nil
}

func (g *gelfTCP) Close() (err error) {
	if g.conn != nil {
		err = g.conn.Close()
		g.conn = nil
	}

	return err
}

func (g *gelfTCP) Connect() error {
	var err error
	var conn net.Conn
	if g.tlsConfig == nil {
		conn, err = g.dialer.Dial("tcp", g.gelfConfig.Endpoint)
	} else {
		conn, err = tls.DialWithDialer(g.dialer, "tcp", g.gelfConfig.Endpoint, g.tlsConfig)
	}
	if err != nil {
		return err
	}
	g.conn = conn
	return nil
}

func (g *gelfTCP) send(b []byte) error {
	if g.conn == nil {
		err := g.Connect()
		if err != nil {
			return err
		}
	}

	_, err := g.conn.Write(b)
	if err != nil {
		_ = g.conn.Close()
		g.conn = nil
	} else {
		_, err = g.conn.Write([]byte{0}) // message delimiter
		if err != nil {
			_ = g.conn.Close()
			g.conn = nil
		}
	}

	return err
}

type Graylog struct {
	Servers           []string        `toml:"servers"`
	ShortMessageField string          `toml:"short_message_field"`
	NameFieldNoPrefix bool            `toml:"name_field_noprefix"`
	Timeout           config.Duration `toml:"timeout"`
	tlsint.ClientConfig

	writer  io.Writer
	closers []io.WriteCloser
}

var sampleConfig = `
  ## Endpoints for your graylog instances.
  servers = ["udp://127.0.0.1:12201"]

  ## Connection timeout.
  # timeout = "5s"

  ## The field to use as the GELF short_message, if unset the static string
  ## "telegraf" will be used.
  ##   example: short_message_field = "message"
  # short_message_field = ""

  ## According to GELF payload specification, additional fields names must be prefixed
  ## with an underscore. Previous versions did not prefix custom field 'name' with underscore.
  ## Set to true for backward compatibility.
  # name_field_no_prefix = false

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
`

func (g *Graylog) Connect() error {
	var writers []io.Writer
	dialer := &net.Dialer{Timeout: time.Duration(g.Timeout)}

	if len(g.Servers) == 0 {
		g.Servers = append(g.Servers, "localhost:12201")
	}

	tlsCfg, err := g.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	for _, server := range g.Servers {
		w := newGelfWriter(gelfConfig{Endpoint: server}, dialer, tlsCfg)
		err := w.Connect()
		if err != nil {
			return fmt.Errorf("failed to connect to server [%s]: %v", server, err)
		}
		writers = append(writers, w)
		g.closers = append(g.closers, w)
	}

	g.writer = io.MultiWriter(writers...)
	return nil
}

func (g *Graylog) Close() error {
	for _, closer := range g.closers {
		_ = closer.Close()
	}
	return nil
}

func (g *Graylog) SampleConfig() string {
	return sampleConfig
}

func (g *Graylog) Description() string {
	return "Send telegraf metrics to graylog"
}

func (g *Graylog) Write(metrics []telegraf.Metric) error {
	for _, metric := range metrics {
		values, err := g.serialize(metric)
		if err != nil {
			return err
		}

		for _, value := range values {
			_, err := g.writer.Write([]byte(value))
			if err != nil {
				return fmt.Errorf("error writing message: %q, %v", value, err)
			}
		}
	}
	return nil
}

func (g *Graylog) serialize(metric telegraf.Metric) ([]string, error) {
	var out []string

	m := make(map[string]interface{})
	m["version"] = "1.1"
	m["timestamp"] = float64(metric.Time().UnixNano()) / 1_000_000_000
	m["short_message"] = "telegraf"
	if g.NameFieldNoPrefix {
		m["name"] = metric.Name()
	} else {
		m["_name"] = metric.Name()
	}

	if host, ok := metric.GetTag("host"); ok {
		m["host"] = host
	} else {
		host, err := os.Hostname()
		if err != nil {
			return []string{}, err
		}
		m["host"] = host
	}

	add := func(key string, value interface{}) {
		switch key {
		case "short_message", "full_message":
			m[key] = value
		default:
			m["_"+key] = value
		}
	}

	for _, tag := range metric.TagList() {
		if tag.Key != "host" {
			add(tag.Key, tag.Value)
		}
	}

	for _, field := range metric.FieldList() {
		if field.Key == g.ShortMessageField {
			m["short_message"] = field.Value
		} else {
			add(field.Key, field.Value)
		}
	}

	serialized, err := ejson.Marshal(m)
	if err != nil {
		return []string{}, err
	}
	out = append(out, string(serialized))

	return out, nil
}

func init() {
	outputs.Add("graylog", func() telegraf.Output {
		return &Graylog{
			Timeout: config.Duration(defaultTimeout),
		}
	})
}
