package influxdb_v2

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
)

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
	Token            string            `toml:"token"`
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
	tls.ClientConfig

	Log telegraf.Logger `toml:"-"`

	clients []Client
}

func (i *InfluxDB) Connect() error {
	if len(i.URLs) == 0 {
		i.URLs = append(i.URLs, defaultURL)
	}

	for _, u := range i.URLs {
		parts, err := url.Parse(u)
		if err != nil {
			return fmt.Errorf("error parsing url [%q]: %v", u, err)
		}

		var proxy *url.URL
		if len(i.HTTPProxy) > 0 {
			proxy, err = url.Parse(i.HTTPProxy)
			if err != nil {
				return fmt.Errorf("error parsing proxy_url [%s]: %v", i.HTTPProxy, err)
			}
		}

		switch parts.Scheme {
		case "http", "https", "unix":
			c, err := i.getHTTPClient(parts, proxy)
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

	return fmt.Errorf("failed to send metrics to any configured server(s)")
}

func (i *InfluxDB) getHTTPClient(address *url.URL, proxy *url.URL) (Client, error) {
	tlsConfig, err := i.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	httpConfig := &HTTPConfig{
		URL:              address,
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
		Serializer:       i.newSerializer(),
		Log:              i.Log,
	}

	c, err := NewHTTPClient(httpConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP client [%s]: %v", address, err)
	}

	return c, nil
}

func (i *InfluxDB) newSerializer() *influx.Serializer {
	serializer := influx.NewSerializer()
	if i.UintSupport {
		serializer.SetFieldTypeSupport(influx.UintSupport)
	}

	return serializer
}

func init() {
	outputs.Add("influxdb_v2", func() telegraf.Output {
		return &InfluxDB{
			Timeout:         config.Duration(time.Second * 5),
			ContentEncoding: "gzip",
		}
	})
}
