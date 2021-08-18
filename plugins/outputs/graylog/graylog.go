package graylog

import (
	"bytes"
	"compress/zlib"
	"crypto/rand"
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
	"github.com/influxdata/telegraf/plugins/outputs"
)

const (
	defaultGraylogEndpoint = "127.0.0.1:12201"
	defaultConnection      = "wan"
	defaultMaxChunkSizeWan = 1420
	defaultMaxChunkSizeLan = 8154
	defaultScheme          = "udp"
	defaultTimeout         = 5 * time.Second
)

type gelfConfig struct {
	GraylogEndpoint string
	Connection      string
	MaxChunkSizeWan int
	MaxChunkSizeLan int
}

type gelf interface {
	io.WriteCloser
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
}

func newGelfWriter(config gelfConfig, dialer *net.Dialer) gelf {
	if config.GraylogEndpoint == "" {
		config.GraylogEndpoint = defaultGraylogEndpoint
	}

	if config.Connection == "" {
		config.Connection = defaultConnection
	}

	if config.MaxChunkSizeWan == 0 {
		config.MaxChunkSizeWan = defaultMaxChunkSizeWan
	}

	if config.MaxChunkSizeLan == 0 {
		config.MaxChunkSizeLan = defaultMaxChunkSizeLan
	}

	scheme := defaultScheme
	parts := strings.SplitN(config.GraylogEndpoint, "://", 2)
	if len(parts) == 2 {
		scheme = strings.ToLower(parts[0])
		config.GraylogEndpoint = parts[1]
	}
	common := gelfCommon{
		gelfConfig: config,
		dialer:     dialer,
	}

	var g gelf
	switch scheme {
	case "tcp":
		g = &gelfTCP{gelfCommon: common}
	case "udp":
		fallthrough
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

	return
}

func (g *gelfUDP) Close() (err error) {
	if g.conn != nil {
		err = g.conn.Close()
		g.conn = nil
	}

	return
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

func (g *gelfUDP) send(b []byte) error {
	if g.conn == nil {
		conn, err := g.dialer.Dial("udp", g.gelfConfig.GraylogEndpoint)
		if err != nil {
			return err
		}
		g.conn = conn
	}

	_, err := g.conn.Write(b)
	if err != nil {
		g.conn.Close()
		g.conn = nil
	}

	return err
}

func (g *gelfTCP) Write(message []byte) (n int, err error) {
	err = g.write(message)
	if err != nil {
		return 0, err
	}

	n = len(message)

	return
}

func (g *gelfTCP) Close() (err error) {
	if g.conn != nil {
		err = g.conn.Close()
		g.conn = nil
	}

	return
}

func (g *gelfTCP) write(b []byte) error {
	if g.conn == nil {
		conn, err := g.dialer.Dial("tcp", g.gelfConfig.GraylogEndpoint)
		if err != nil {
			return err
		}
		g.conn = conn
	}

	_, err := g.conn.Write(b)
	if err != nil {
		g.conn.Close()
		g.conn = nil
	} else {
		_, err = g.conn.Write([]byte{0}) // message delimiter
		if err != nil {
			g.conn.Close()
			g.conn = nil
		}
	}

	return err
}

type Graylog struct {
	Servers           []string        `toml:"servers"`
	ShortMessageField string          `toml:"short_message_field"`
	Timeout           config.Duration `toml:"timeout"`

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
`

func (g *Graylog) Connect() error {
	writers := []io.Writer{}
	dialer := net.Dialer{Timeout: time.Duration(g.Timeout)}

	if len(g.Servers) == 0 {
		g.Servers = append(g.Servers, "localhost:12201")
	}

	for _, server := range g.Servers {
		w := newGelfWriter(gelfConfig{GraylogEndpoint: server}, &dialer)
		writers = append(writers, w)
		g.closers = append(g.closers, w)
	}

	g.writer = io.MultiWriter(writers...)
	return nil
}

func (g *Graylog) Close() error {
	for _, closer := range g.closers {
		closer.Close()
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
	out := []string{}

	m := make(map[string]interface{})
	m["version"] = "1.1"
	m["timestamp"] = float64(metric.Time().UnixNano()) / 1_000_000_000
	m["short_message"] = "telegraf"
	m["name"] = metric.Name()

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
		if tag.Key != "host" {
			m["_"+tag.Key] = tag.Value
		}
	}

	for _, field := range metric.FieldList() {
		if field.Key == g.ShortMessageField {
			m["short_message"] = field.Value
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

func init() {
	outputs.Add("graylog", func() telegraf.Output {
		return &Graylog{
			Timeout: config.Duration(defaultTimeout),
		}
	})
}
