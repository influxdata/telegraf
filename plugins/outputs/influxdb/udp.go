package influxdb

import (
	"context"
	"fmt"
	"net"
	"net/url"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
)

const (
	// DefaultMaxPacketSize is the largest UDP packet that will be sent
	DefaultMaxPacketSize = 512
)

type Dialer interface {
	DialContext(ctx context.Context, network, address string) (Conn, error)
}

type Conn interface {
	Write(b []byte) (int, error)
	Close() error
}

type UDPConfig struct {
	MaxPacketSize int
	URL           *url.URL
	Serializer    serializers.Serializer
	Dialer        Dialer
}

func NewUDPClient(config *UDPConfig) (*udpClient, error) {
	if config.URL == nil {
		return nil, ErrMissingURL
	}

	size := config.MaxPacketSize
	if size == 0 {
		size = DefaultMaxPacketSize
	}

	serializer := config.Serializer
	if serializer == nil {
		s := influx.NewSerializer()
		s.SetMaxLineBytes(config.MaxPacketSize)
		serializer = s
	}

	dialer := config.Dialer
	if dialer == nil {
		dialer = &netDialer{net.Dialer{}}
	}

	client := &udpClient{
		url:        config.URL,
		serializer: serializer,
		dialer:     dialer,
	}
	return client, nil
}

type udpClient struct {
	conn       Conn
	dialer     Dialer
	serializer serializers.Serializer
	url        *url.URL
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
			return fmt.Errorf("could not serialize metric: %v", err)
		}

		_, err = c.conn.Write(octets)
		if err != nil {
			c.conn.Close()
			c.conn = nil
			return err
		}
	}

	return nil
}

func (c *udpClient) CreateDatabase(ctx context.Context) error {
	return nil
}

type netDialer struct {
	net.Dialer
}

func (d *netDialer) DialContext(ctx context.Context, network, address string) (Conn, error) {
	return d.Dialer.DialContext(ctx, network, address)
}
