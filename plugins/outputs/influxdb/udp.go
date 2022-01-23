package influxdb

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"net/url"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
)

const (
	// DefaultMaxPayloadSize is the maximum length of the UDP data payload
	DefaultMaxPayloadSize = 512
)

type Dialer interface {
	DialContext(ctx context.Context, network, address string) (Conn, error)
}

type Conn interface {
	Write(b []byte) (int, error)
	Close() error
}

type UDPConfig struct {
	MaxPayloadSize int
	URL            *url.URL
	Serializer     *influx.Serializer
	Dialer         Dialer
	Log            telegraf.Logger
}

func NewUDPClient(config UDPConfig) (*udpClient, error) {
	if config.URL == nil {
		return nil, ErrMissingURL
	}

	size := config.MaxPayloadSize
	if size == 0 {
		size = DefaultMaxPayloadSize
	}

	serializer := config.Serializer
	if serializer == nil {
		s := influx.NewSerializer()
		serializer = s
	}
	serializer.SetMaxLineBytes(size)

	dialer := config.Dialer
	if dialer == nil {
		dialer = &netDialer{net.Dialer{}}
	}

	client := &udpClient{
		url:        config.URL,
		serializer: serializer,
		dialer:     dialer,
		log:        config.Log,
	}
	return client, nil
}

type udpClient struct {
	conn       Conn
	dialer     Dialer
	serializer *influx.Serializer
	url        *url.URL
	log        telegraf.Logger
}

func (c *udpClient) URL() string {
	return c.url.String()
}

func (c *udpClient) Database() string {
	return ""
}

func (c *udpClient) Write(ctx context.Context, metrics []telegraf.Metric) error {
	if c.conn == nil {
		conn, err := c.dialer.DialContext(ctx, c.url.Scheme, c.url.Host)
		if err != nil {
			return fmt.Errorf("error dialing address [%s]: %s", c.url, err)
		}
		c.conn = conn
	}

	for _, metric := range metrics {
		octets, err := c.serializer.Serialize(metric)
		if err != nil {
			// Since we are serializing multiple metrics, don't fail the
			// entire batch just because of one unserializable metric.
			c.log.Errorf("When writing to [%s] could not serialize metric: %v",
				c.URL(), err)
			continue
		}

		scanner := bufio.NewScanner(bytes.NewReader(octets))
		scanner.Split(scanLines)
		for scanner.Scan() {
			_, err = c.conn.Write(scanner.Bytes())
		}
		if err != nil {
			_ = c.conn.Close()
			c.conn = nil
			return err
		}
	}

	return nil
}

func (c *udpClient) CreateDatabase(_ context.Context, _ string) error {
	return nil
}

type netDialer struct {
	net.Dialer
}

func (d *netDialer) DialContext(ctx context.Context, network, address string) (Conn, error) {
	return d.Dialer.DialContext(ctx, network, address)
}

func scanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, data[0 : i+1], nil
	}
	return 0, nil, nil
}

func (c *udpClient) Close() {
}
