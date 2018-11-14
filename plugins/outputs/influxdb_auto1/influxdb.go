package influxdb_auto1

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"time"

	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
)

var (
	defaultURL = "http://localhost:8086"

	defaultAuto1MonitoringKey = "MonitoringDomain"
	defaultAuto1MonitoringKeySplitter = ","

	ErrMissingURL = errors.New("missing URL")
)

type Client interface {
	Write(context.Context, []telegraf.Metric) error
	CreateDatabase(ctx context.Context) error

	URL() string
	Database() string
}

// InfluxDB struct is the primary data structure for the plugin
type InfluxDB struct {
	URL                  string   // url deprecated in 0.1.9; use urls
	URLs                 []string `toml:"urls"`
	Username             string
	Password             string
	Database             string
	UserAgent            string
	RetentionPolicy      string
	WriteConsistency     string
	Timeout              internal.Duration
	UDPPayload           internal.Size     `toml:"udp_payload"`
	HTTPProxy            string            `toml:"http_proxy"`
	HTTPHeaders          map[string]string `toml:"http_headers"`
	ContentEncoding      string            `toml:"content_encoding"`
	SkipDatabaseCreation bool              `toml:"skip_database_creation"`
	InfluxUintSupport    bool              `toml:"influx_uint_support"`
	tls.ClientConfig

	Precision string // precision deprecated in 1.0; value is ignored

	clients []Client

	CreateHTTPClientF func(config *HTTPConfig) (Client, error)
	CreateUDPClientF  func(config *UDPConfig) (Client, error)

	serializer *influx.Serializer

	Auto1MonitoringDomainKey  string            `toml:"Auto1MonitoringDomainKey"`
	Auto1MonitoringDomainKeySplitter  string    `toml:"Auto1MonitoringDomainKeySplitter"`
}

var sampleConfig = `
  ## The full HTTP or UDP URL for your InfluxDB instance.
  ##
  ## Multiple URLs can be specified for a single cluster, only ONE of the
  ## urls will be written to each interval.
  # urls = ["unix:///var/run/influxdb.sock"]
  # urls = ["udp://127.0.0.1:8089"]
  # urls = ["http://127.0.0.1:8086"]

  ## The target database for metrics; will be created as needed.
  ## For UDP url endpoint database needs to be configured on server side.
  # database = "telegraf"

  ## If true, no CREATE DATABASE queries will be sent.  Set to true when using
  ## Telegraf with a user without permissions to create databases or when the
  ## database already exists.
  # skip_database_creation = false

  ## Name of existing retention policy to write to.  Empty string writes to
  ## the default retention policy.  Only takes effect when using HTTP.
  # retention_policy = ""

  ## Write consistency (clusters only), can be: "any", "one", "quorum", "all".
  ## Only takes effect when using HTTP.
  # write_consistency = "any"

  ## Timeout for HTTP messages.
  # timeout = "5s"

  ## HTTP Basic Auth
  # username = "telegraf"
  # password = "metricsmetricsmetricsmetrics"

  ## HTTP User-Agent
  # user_agent = "telegraf"

  ## UDP payload size is the maximum packet size to send.
  # udp_payload = "512B"

  ## Optional TLS Config for use on HTTP connections.
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## HTTP Proxy override, if unset values the standard proxy environment
  ## variables are consulted to determine which proxy, if any, should be used.
  # http_proxy = "http://corporate.proxy:3128"

  ## Additional HTTP headers
  # http_headers = {"X-Special-Header" = "Special-Value"}

  ## HTTP Content-Encoding for write request body, can be set to "gzip" to
  ## compress body or "identity" to apply no encoding.
  # content_encoding = "identity"

  ## When true, Telegraf will output unsigned integers as unsigned values,
  ## i.e.: "42u".  You will need a version of InfluxDB supporting unsigned
  ## integer values.  Enabling this option will result in field type errors if
  ## existing data has been written.
  # influx_uint_support = false

  # defaults to 
  Auto1MonitoringDomainKey = "MonitoringDomain"
  Auto1MonitoringDomainKeySplitter = ","
`

func (i *InfluxDB) Connect() error {
	ctx := context.Background()

	urls := make([]string, 0, len(i.URLs))
	urls = append(urls, i.URLs...)
	if i.URL != "" {
		urls = append(urls, i.URL)
	}

	if len(urls) == 0 {
		urls = append(urls, defaultURL)
	}

	if i.Auto1MonitoringDomainKey == "" {
		i.Auto1MonitoringDomainKey = defaultAuto1MonitoringKey
	}

	if i.Auto1MonitoringDomainKeySplitter == "" {
		i.Auto1MonitoringDomainKeySplitter = defaultAuto1MonitoringKeySplitter
	}

	i.serializer = influx.NewSerializer()
	if i.InfluxUintSupport {
		i.serializer.SetFieldTypeSupport(influx.UintSupport)
	}

	for _, u := range urls {
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
		case "udp", "udp4", "udp6":
			c, err := i.udpClient(parts)
			if err != nil {
				return err
			}

			i.clients = append(i.clients, c)
		case "http", "https", "unix":
			c, err := i.httpClient(ctx, parts, proxy)
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
	return nil
}

func (i *InfluxDB) Description() string {
	return "Configuration for sending metrics to InfluxDB"
}

func (i *InfluxDB) SampleConfig() string {
	return sampleConfig
}

func timeTrack(start time.Time, name string) {
    elapsed := time.Since(start)
    log.Printf("I! [outputs.influxdb_auto1] %s took %s", name, time.Duration(elapsed))
}

// augmented a composed `Auto1MonitoringDomainKey` delimited by `Auto1MonitoringDomainKeySplitter`
// and creates aditional data points. it does removed original metric with non parsed 
// Auto1MonitoringDomainKey.
func (i *InfluxDB) augment_data(metrics []telegraf.Metric) []telegraf.Metric {
	//  // log execution time o thi method call
	// defer timeTrack(time.Now(), "augment_data")
	custom_metrics := metrics[:0]
	for _, point := range metrics {
		value, ok := point.GetTag(i.Auto1MonitoringDomainKey)
		if ok {
			domains := strings.Split(value, i.Auto1MonitoringDomainKeySplitter)
			if len(domains) > 1 {
				// contruct current + extra ones
				for _, domain := range domains {
					metric_data := point.Copy()
					metric_data.RemoveTag(i.Auto1MonitoringDomainKey)
					metric_data.AddTag(i.Auto1MonitoringDomainKey, strings.TrimSpace(domain))
					custom_metrics = append(custom_metrics, metric_data)
				}
			}
		}
	}
	return custom_metrics
}

// Write sends metrics to one of the configured servers, logging each
// unsuccessful. If all servers fail, return an error.
func (i *InfluxDB) Write(metrics []telegraf.Metric) error {

	// // log execution time o thi method call
	// defer timeTrack(time.Now(), "WriteMetrics")

	ctx := context.Background()

	var err error
	p := rand.Perm(len(i.clients))
	for _, n := range p {
		client := i.clients[n]

		metrics = i.augment_data(metrics)

		err = client.Write(ctx, metrics)
		if err == nil {
			return nil
		}

		switch apiError := err.(type) {
		case *APIError:
			if !i.SkipDatabaseCreation {
				if apiError.Type == DatabaseNotFound {
					err := client.CreateDatabase(ctx)
					if err != nil {
						log.Printf("E! [outputs.influxdb_auto1] when writing to [%s]: database %q not found and failed to recreate",
							client.URL(), client.Database())
					}
				}
			}
		}

		log.Printf("E! [outputs.influxdb_auto1] when writing to [%s]: %v", client.URL(), err)
	}

	return errors.New("could not write any address")
}

func (i *InfluxDB) udpClient(url *url.URL) (Client, error) {
	config := &UDPConfig{
		URL:            url,
		MaxPayloadSize: int(i.UDPPayload.Size),
		Serializer:     i.serializer,
	}

	c, err := i.CreateUDPClientF(config)
	if err != nil {
		return nil, fmt.Errorf("error creating UDP client [%s]: %v", url, err)
	}

	return c, nil
}

func (i *InfluxDB) httpClient(ctx context.Context, url *url.URL, proxy *url.URL) (Client, error) {
	tlsConfig, err := i.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	config := &HTTPConfig{
		URL:             url,
		Timeout:         i.Timeout.Duration,
		TLSConfig:       tlsConfig,
		UserAgent:       i.UserAgent,
		Username:        i.Username,
		Password:        i.Password,
		Proxy:           proxy,
		ContentEncoding: i.ContentEncoding,
		Headers:         i.HTTPHeaders,
		Database:        i.Database,
		RetentionPolicy: i.RetentionPolicy,
		Consistency:     i.WriteConsistency,
		Serializer:      i.serializer,
	}

	c, err := i.CreateHTTPClientF(config)
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP client [%s]: %v", url, err)
	}

	if !i.SkipDatabaseCreation {
		err = c.CreateDatabase(ctx)
		if err != nil {
			log.Printf("W! [outputs.influxdb_auto1] when writing to [%s]: database %q creation failed: %v",
				c.URL(), c.Database(), err)
		}
	}

	return c, nil
}

func init() {
	outputs.Add("influxdb_auto1", func() telegraf.Output {
		return &InfluxDB{
			Timeout: internal.Duration{Duration: time.Second * 5},
			CreateHTTPClientF: func(config *HTTPConfig) (Client, error) {
				return NewHTTPClient(config)
			},
			CreateUDPClientF: func(config *UDPConfig) (Client, error) {
				return NewUDPClient(config)
			},
		}
	})
}
