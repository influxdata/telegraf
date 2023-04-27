//go:generate ../../../tools/readme_config_includer/generator
package graylog

import (
	"bytes"
	"compress/zlib"
	"crypto/rand"
	"crypto/tls"
	_ "embed"
	"encoding/binary"
	ejson "encoding/json"
	"fmt"
	"io"
	"math"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	tlsint "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
)

//go:embed sample.conf
var sampleConfig string

const (
	defaultEndpoint         = "127.0.0.1:12201"
	defaultConnection       = "wan"
	defaultMaxChunkSizeWan  = 1420
	defaultMaxChunkSizeLan  = 8154
	defaultScheme           = "udp"
	defaultTimeout          = 5 * time.Second
	defaultReconnectionTime = 15 * time.Second
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
	packet.Write(b)

	b, err = g.intToBytes(15)
	if err != nil {
		return packet, err
	}
	packet.Write(b)

	packet.Write(id)

	b, err = g.intToBytes(index)
	if err != nil {
		return packet, err
	}
	packet.Write(b)

	b, err = g.intToBytes(chunkCountInt)
	if err != nil {
		return packet, err
	}
	packet.Write(b)

	packet.Write(compressed.Next(chunksize))

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
	Reconnection      bool            `toml:"connection_retry"`
	ReconnectionTime  config.Duration `toml:"connection_retry_wait_time"`
	Log               telegraf.Logger `toml:"-"`
	tlsint.ClientConfig

	writer      io.Writer
	closers     []io.WriteCloser
	unconnected []string
	stopRetry   bool
	wg          sync.WaitGroup

	sync.Mutex
}

func (*Graylog) SampleConfig() string {
	return sampleConfig
}

func (g *Graylog) Connect() error {
	if len(g.Servers) == 0 {
		g.Servers = append(g.Servers, "localhost:12201")
	}

	tlsCfg, err := g.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	if g.Reconnection {
		go g.connectRetry(tlsCfg)
		return nil
	}

	unconnected, gelfs := g.connectEndpoints(g.Servers, tlsCfg)
	if len(unconnected) > 0 {
		servers := strings.Join(unconnected, ",")
		return fmt.Errorf("connect: connection failed for %s", servers)
	}
	writers := make([]io.Writer, 0, len(gelfs))
	closers := make([]io.WriteCloser, 0, len(gelfs))
	for _, w := range gelfs {
		writers = append(writers, w)
		closers = append(closers, w)
	}
	g.Lock()
	defer g.Unlock()
	g.writer = io.MultiWriter(writers...)
	g.closers = closers

	return nil
}

func (g *Graylog) connectRetry(tlsCfg *tls.Config) {
	var writers []io.Writer
	var closers []io.WriteCloser
	var attempt int64

	g.wg.Add(1)

	unconnected := append([]string{}, g.Servers...)
	for {
		unconnected, gelfs := g.connectEndpoints(unconnected, tlsCfg)
		for _, w := range gelfs {
			writers = append(writers, w)
			closers = append(closers, w)
		}
		g.Lock()
		g.unconnected = unconnected
		stopRetry := g.stopRetry
		g.Unlock()
		if stopRetry {
			g.Log.Info("Stopping connection retries...")
			break
		}
		if len(unconnected) == 0 {
			break
		}
		attempt++
		servers := strings.Join(unconnected, ",")
		g.Log.Infof("Not connected to endpoints %s after attempt #%d...", servers, attempt)
		time.Sleep(time.Duration(g.ReconnectionTime))
	}
	g.Log.Info("Connected!")

	g.Lock()
	g.writer = io.MultiWriter(writers...)
	g.closers = closers
	g.Unlock()

	g.wg.Done()
}

func (g *Graylog) connectEndpoints(servers []string, tlsCfg *tls.Config) ([]string, []gelf) {
	writers := make([]gelf, 0, len(servers))
	unconnected := make([]string, 0, len(servers))
	dialer := &net.Dialer{Timeout: time.Duration(g.Timeout)}
	for _, server := range servers {
		w := newGelfWriter(gelfConfig{Endpoint: server}, dialer, tlsCfg)
		if err := w.Connect(); err != nil {
			g.Log.Warnf("failed to connect to server [%s]: %v", server, err)
			unconnected = append(unconnected, server)
			continue
		}
		writers = append(writers, w)
	}
	return unconnected, writers
}

func (g *Graylog) Close() error {
	g.Lock()
	g.stopRetry = true
	g.Unlock()
	g.wg.Wait()

	for _, closer := range g.closers {
		_ = closer.Close()
	}
	return nil
}

func (g *Graylog) Write(metrics []telegraf.Metric) error {
	g.Lock()
	writer := g.writer
	g.Unlock()

	if writer == nil {
		g.Lock()
		unconnected := strings.Join(g.unconnected, ",")
		g.Unlock()

		return fmt.Errorf("not connected to %s", unconnected)
	}
	for _, metric := range metrics {
		values, err := g.serialize(metric)
		if err != nil {
			return err
		}

		for _, value := range values {
			_, err = writer.Write([]byte(value))
			if err != nil {
				return fmt.Errorf("error writing message: %q: %w", value, err)
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
			Timeout:          config.Duration(defaultTimeout),
			ReconnectionTime: config.Duration(defaultReconnectionTime),
		}
	})
}
