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

var defaultSpecFields = []string{"version", "host", "short_message", "full_message", "timestamp", "level", "facility", "line", "file"}

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
	compressed, err := g.compress(message)
	if err != nil {
		return 0, err
	}

	chunksize := g.gelfConfig.MaxChunkSizeWan
	length := compressed.Len()

	if length > chunksize {
		chunkCountInt := int(math.Ceil(float64(length) / float64(chunksize)))

		id := make([]byte, 8)
		_, err = rand.Read(id)
		if err != nil {
			return 0, err
		}

		for i, index := 0, 0; i < length; i, index = i+chunksize, index+1 {
			packet, err := g.createChunkedMessage(index, chunkCountInt, id, &compressed)
			if err != nil {
				return 0, err
			}

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

func (g *gelfUDP) createChunkedMessage(index int, chunkCountInt int, id []byte, compressed *bytes.Buffer) (bytes.Buffer, error) {
	var packet bytes.Buffer

	chunksize := g.getChunksize()

	b, err := g.intToBytes(30)
	if err != nil {
		return packet, err
	}
	packet.Write(b) //nolint:revive // from buffer.go: "err is always nil"

	b, err = g.intToBytes(15)
	if err != nil {
		return packet, err
	}
	packet.Write(b) //nolint:revive // from buffer.go: "err is always nil"

	packet.Write(id) //nolint:revive // from buffer.go: "err is always nil"

	b, err = g.intToBytes(index)
	if err != nil {
		return packet, err
	}
	packet.Write(b) //nolint:revive // from buffer.go: "err is always nil"

	b, err = g.intToBytes(chunkCountInt)
	if err != nil {
		return packet, err
	}
	packet.Write(b) //nolint:revive // from buffer.go: "err is always nil"

	packet.Write(compressed.Next(chunksize)) //nolint:revive // from buffer.go: "err is always nil"

	return packet, nil
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

func (g *gelfUDP) intToBytes(i int) ([]byte, error) {
	buf := new(bytes.Buffer)

	err := binary.Write(buf, binary.LittleEndian, int8(i))
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func (g *gelfUDP) compress(b []byte) (bytes.Buffer, error) {
	var buf bytes.Buffer
	comp := zlib.NewWriter(&buf)

	if _, err := comp.Write(b); err != nil {
		return bytes.Buffer{}, err
	}

	if err := comp.Close(); err != nil {
		return bytes.Buffer{}, err
	}

	return buf, nil
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
	out := []string{}

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

	for _, tag := range metric.TagList() {
		if tag.Key == "host" {
			continue
		}

		if fieldInSpec(tag.Key) {
			m[tag.Key] = tag.Value
		} else {
			m["_"+tag.Key] = tag.Value
		}
	}

	for _, field := range metric.FieldList() {
		if field.Key == g.ShortMessageField {
			m["short_message"] = field.Value
		} else if fieldInSpec(field.Key) {
			m[field.Key] = field.Value
		} else {
			m["_"+field.Key] = field.Value
		}
	}

	serialized, err := ejson.Marshal(m)
	if err != nil {
		return []string{}, err
	}
	out = append(out, string(serialized))

	return out, nil
}

func fieldInSpec(field string) bool {
	for _, specField := range defaultSpecFields {
		if specField == field {
			return true
		}
	}

	return false
}

func init() {
	outputs.Add("graylog", func() telegraf.Output {
		return &Graylog{
			Timeout: config.Duration(defaultTimeout),
		}
	})
}
