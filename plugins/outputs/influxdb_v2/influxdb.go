package influxdb_v2

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
)

var (
	defaultURL = "http://localhost:8086"

	ErrMissingURL = errors.New("missing URL")
)

var sampleConfig = `
  ## The URLs of the InfluxDB cluster nodes.
  ##
  ## Multiple URLs can be specified for a single cluster, only ONE of the
  ## urls will be written to each interval.
  ##   ex: urls = ["https://us-west-2-1.aws.cloud2.influxdata.com"]
  urls = ["http://127.0.0.1:8086"]

  ## Token for authentication.
  token = ""

  ## Organization is the name of the organization you wish to write to; must exist.
  organization = ""

  ## Destination bucket to write into.
  bucket = ""

  ## The value of this tag will be used to determine the bucket.  If this
  ## tag is not set the 'bucket' option is used as the default.
  # bucket_tag = ""

  ## If true, the bucket tag will not be added to the metric.
  # exclude_bucket_tag = false

  ## Timeout for HTTP messages.
  # timeout = "5s"

  ## Additional HTTP headers
  # http_headers = {"X-Special-Header" = "Special-Value"}

  ## HTTP Proxy override, if unset values the standard proxy environment
  ## variables are consulted to determine which proxy, if any, should be used.
  # http_proxy = "http://corporate.proxy:3128"

  ## HTTP User-Agent
  # user_agent = "telegraf"

  ## Content-Encoding for write request body, can be set to "gzip" to
  ## compress body or "identity" to apply no encoding.
  # content_encoding = "gzip"

  ## Enable or disable uint support for writing uints influxdb 2.0.
  # influx_uint_support = false

  ## Optional TLS Config for use on HTTP connections.
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
`

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
	Timeout          internal.Duration `toml:"timeout"`
	HTTPHeaders      map[string]string `toml:"http_headers"`
	HTTPProxy        string            `toml:"http_proxy"`
	UserAgent        string            `toml:"user_agent"`
	ContentEncoding  string            `toml:"content_encoding"`
	UintSupport      bool              `toml:"influx_uint_support"`
	tls.ClientConfig

	clients []Client
}

func (i *InfluxDB) Connect() error {
	ctx := context.Background()

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
			c, err := i.getHTTPClient(ctx, parts, proxy)
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

func (i *InfluxDB) Description() string {
	return "Configuration for sending metrics to InfluxDB"
}

func (i *InfluxDB) SampleConfig() string {
	return sampleConfig
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

		log.Printf("E! [outputs.influxdb_v2] when writing to [%s]: %v", client.URL(), err)
	}

	return err
}

func (i *InfluxDB) getHTTPClient(ctx context.Context, url *url.URL, proxy *url.URL) (Client, error) {
	tlsConfig, err := i.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	config := &HTTPConfig{
		URL:              url,
		Token:            i.Token,
		Organization:     i.Organization,
		Bucket:           i.Bucket,
		BucketTag:        i.BucketTag,
		ExcludeBucketTag: i.ExcludeBucketTag,
		Timeout:          i.Timeout.Duration,
		Headers:          i.HTTPHeaders,
		Proxy:            proxy,
		UserAgent:        i.UserAgent,
		ContentEncoding:  i.ContentEncoding,
		TLSConfig:        tlsConfig,
		Serializer:       i.newSerializer(),
	}

	c, err := NewHTTPClient(config)
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP client [%s]: %v", url, err)
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
			Timeout:         internal.Duration{Duration: time.Second * 5},
			ContentEncoding: "gzip",
		}
	})
}
