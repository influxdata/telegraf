//go:generate ../../../tools/readme_config_includer/generator
package influxdb_v2

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
)

//go:embed sample.conf
var sampleConfig string

var (
	defaultURL = "http://localhost:8086"

	ErrMissingURL = errors.New("missing URL")
)

type Client interface {
	Write(context.Context, []telegraf.Metric) error

	URL() string // for logging
	Close()
}

type InfluxDB struct {
	URLs             []string          `toml:"urls"`
	LocalAddr        string            `toml:"local_address"`
	Token            config.Secret     `toml:"token"`
	Organization     string            `toml:"organization"`
	Bucket           string            `toml:"bucket"`
	BucketTag        string            `toml:"bucket_tag"`
	ExcludeBucketTag bool              `toml:"exclude_bucket_tag"`
	Timeout          config.Duration   `toml:"timeout"`
	HTTPHeaders      map[string]string `toml:"http_headers"`
	HTTPProxy        string            `toml:"http_proxy"`
	UserAgent        string            `toml:"user_agent"`
	ContentEncoding  string            `toml:"content_encoding"`
	UintSupport      bool              `toml:"influx_uint_support"`
	OmitTimestamp    bool              `toml:"influx_omit_timestamp"`
	PingTimeout      config.Duration   `toml:"ping_timeout"`
	ReadIdleTimeout  config.Duration   `toml:"read_idle_timeout"`
	tls.ClientConfig

	Log telegraf.Logger `toml:"-"`

	clients []Client
}

func (*InfluxDB) SampleConfig() string {
	return sampleConfig
}

func (i *InfluxDB) Connect() error {
	if len(i.URLs) == 0 {
		i.URLs = append(i.URLs, defaultURL)
	}

	for _, u := range i.URLs {
		parts, err := url.Parse(u)
		if err != nil {
			return fmt.Errorf("error parsing url [%q]: %w", u, err)
		}

		var proxy *url.URL
		if len(i.HTTPProxy) > 0 {
			proxy, err = url.Parse(i.HTTPProxy)
			if err != nil {
				return fmt.Errorf("error parsing proxy_url [%s]: %w", i.HTTPProxy, err)
			}
		}

		var localAddr *net.TCPAddr
		if i.LocalAddr != "" {
			// Resolve the local address into IP address and the given port if any
			addr, sPort, err := net.SplitHostPort(i.LocalAddr)
			if err != nil {
				if !strings.Contains(err.Error(), "missing port") {
					return fmt.Errorf("invalid local address: %w", err)
				}
				addr = i.LocalAddr
			}
			local, err := net.ResolveIPAddr("ip", addr)
			if err != nil {
				return fmt.Errorf("cannot resolve local address: %w", err)
			}

			var port int
			if sPort != "" {
				p, err := strconv.ParseUint(sPort, 10, 16)
				if err != nil {
					return fmt.Errorf("invalid port: %w", err)
				}
				port = int(p)
			}

			localAddr = &net.TCPAddr{IP: local.IP, Port: port, Zone: local.Zone}
		}

		switch parts.Scheme {
		case "http", "https", "unix":
			c, err := i.getHTTPClient(parts, localAddr, proxy)
			if err != nil {
				return err
			}

			i.clients = append(i.clients, c)
		default:
			return fmt.Errorf("unsupported scheme [%q]: %q", u, parts.Scheme)
		}
	}

	return nil
}

func (i *InfluxDB) Close() error {
	for _, client := range i.clients {
		client.Close()
	}
	return nil
}

// Write sends metrics to one of the configured servers, logging each
// unsuccessful. If all servers fail, return an error.
func (i *InfluxDB) Write(metrics []telegraf.Metric) error {
	ctx := context.Background()

	var err error
	p := rand.Perm(len(i.clients))
	for _, n := range p {
		client := i.clients[n]
		err = client.Write(ctx, metrics)
		if err == nil {
			return nil
		}

		i.Log.Errorf("When writing to [%s]: %v", client.URL(), err)
	}

	return errors.New("failed to send metrics to any configured server(s)")
}

func (i *InfluxDB) getHTTPClient(address *url.URL, localAddr *net.TCPAddr, proxy *url.URL) (Client, error) {
	tlsConfig, err := i.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	serializer := &influx.Serializer{
		UintSupport:   i.UintSupport,
		OmitTimestamp: i.OmitTimestamp,
	}
	if err := serializer.Init(); err != nil {
		return nil, err
	}

	httpConfig := &HTTPConfig{
		URL:              address,
		LocalAddr:        localAddr,
		Token:            i.Token,
		Organization:     i.Organization,
		Bucket:           i.Bucket,
		BucketTag:        i.BucketTag,
		ExcludeBucketTag: i.ExcludeBucketTag,
		Timeout:          time.Duration(i.Timeout),
		Headers:          i.HTTPHeaders,
		Proxy:            proxy,
		UserAgent:        i.UserAgent,
		ContentEncoding:  i.ContentEncoding,
		TLSConfig:        tlsConfig,
		Serializer:       serializer,
		PingTimeout:      i.PingTimeout,
		ReadIdleTimeout:  i.ReadIdleTimeout,
		Log:              i.Log,
	}

	c, err := NewHTTPClient(httpConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP client [%s]: %w", address, err)
	}

	return c, nil
}

func init() {
	outputs.Add("influxdb_v2", func() telegraf.Output {
		return &InfluxDB{
			Timeout:         config.Duration(time.Second * 5),
			ContentEncoding: "gzip",
		}
	})
}
