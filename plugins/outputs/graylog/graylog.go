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

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

const (
	defaultGraylogEndpoint = "127.0.0.1:12201"
	defaultConnection      = "wan"
	defaultMaxChunkSizeWan = 1420
	defaultMaxChunkSizeLan = 8154
)

type GelfConfig struct {
	GraylogEndpoint string
	Connection      string
	MaxChunkSizeWan int
	MaxChunkSizeLan int
}

type Gelf struct {
	GelfConfig
}

func NewGelfWriter(config GelfConfig) *Gelf {
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

	g := &Gelf{GelfConfig: config}

	return g
}

func (g *Gelf) Write(message []byte) (n int, err error) {
	compressed := g.compress(message)

	chunksize := g.GelfConfig.MaxChunkSizeWan
	length := compressed.Len()

	if length > chunksize {

		chunkCountInt := int(math.Ceil(float64(length) / float64(chunksize)))

		id := make([]byte, 8)
		rand.Read(id)

		for i, index := 0, 0; i < length; i, index = i+chunksize, index+1 {
			packet := g.createChunkedMessage(index, chunkCountInt, id, &compressed)
			_, err = g.send(packet.Bytes())
			if err != nil {
				return 0, err
			}
		}
	} else {
		_, err = g.send(compressed.Bytes())
		if err != nil {
			return 0, err
		}
	}

	n = len(message)

	return
}

func (g *Gelf) createChunkedMessage(index int, chunkCountInt int, id []byte, compressed *bytes.Buffer) bytes.Buffer {
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

func (g *Gelf) getChunksize() int {
	if g.GelfConfig.Connection == "wan" {
		return g.GelfConfig.MaxChunkSizeWan
	}

	if g.GelfConfig.Connection == "lan" {
		return g.GelfConfig.MaxChunkSizeLan
	}

	return g.GelfConfig.MaxChunkSizeWan
}

func (g *Gelf) intToBytes(i int) []byte {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, int8(i))
	return buf.Bytes()
}

func (g *Gelf) compress(b []byte) bytes.Buffer {
	var buf bytes.Buffer
	comp := zlib.NewWriter(&buf)

	comp.Write(b)
	comp.Close()

	return buf
}

func (g *Gelf) send(b []byte) (n int, err error) {
	udpAddr, err := net.ResolveUDPAddr("udp", g.GelfConfig.GraylogEndpoint)
	if err != nil {
		return
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return
	}

	n, err = conn.Write(b)
	return
}

type Graylog struct {
	Servers           []string `toml:"servers"`
	ShortMessageField string   `toml:"short_message_field"`
	writer            io.Writer
}

var sampleConfig = `
  ## UDP endpoint for your graylog instance.
  servers = ["127.0.0.1:12201"]

  ## The field to use as the GELF short_message, if unset the static string
  ## "telegraf" will be used.
  ##   example: short_message_field = "message"
  # short_message_field = ""
`

func (g *Graylog) Connect() error {
	writers := []io.Writer{}

	if len(g.Servers) == 0 {
		g.Servers = append(g.Servers, "localhost:12201")
	}

	for _, server := range g.Servers {
		w := NewGelfWriter(GelfConfig{GraylogEndpoint: server})
		writers = append(writers, w)
	}

	g.writer = io.MultiWriter(writers...)
	return nil
}

func (g *Graylog) Close() error {
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
	m["timestamp"] = metric.Time().UnixNano() / 1000000000
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
		return &Graylog{}
	})
}
