//nolint
package influxdb

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
	CreateDatabase(ctx context.Context, database string) error
	Database() string
	URL() string
	Close()
}

// InfluxDB struct is the primary data structure for the plugin
type InfluxDB struct {
	URL                       string            `toml:"url" deprecated:"0.1.9;2.0.0;use 'urls' instead"`
	URLs                      []string          `toml:"urls"`
	Username                  string            `toml:"username"`
	Password                  string            `toml:"password"`
	Database                  string            `toml:"database"`
	DatabaseTag               string            `toml:"database_tag"`
	ExcludeDatabaseTag        bool              `toml:"exclude_database_tag"`
	RetentionPolicy           string            `toml:"retention_policy"`
	RetentionPolicyTag        string            `toml:"retention_policy_tag"`
	ExcludeRetentionPolicyTag bool              `toml:"exclude_retention_policy_tag"`
	UserAgent                 string            `toml:"user_agent"`
	WriteConsistency          string            `toml:"write_consistency"`
	Timeout                   config.Duration   `toml:"timeout"`
	UDPPayload                config.Size       `toml:"udp_payload"`
	HTTPProxy                 string            `toml:"http_proxy"`
	HTTPHeaders               map[string]string `toml:"http_headers"`
	ContentEncoding           string            `toml:"content_encoding"`
	SkipDatabaseCreation      bool              `toml:"skip_database_creation"`
	InfluxUintSupport         bool              `toml:"influx_uint_support"`
	tls.ClientConfig

	Precision string `toml:"precision" deprecated:"1.0.0;option is ignored"`

	clients []Client

	CreateHTTPClientF func(config *HTTPConfig) (Client, error)
	CreateUDPClientF  func(config *UDPConfig) (Client, error)

	Log telegraf.Logger
}

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
	for _, client := range i.clients {
		client.Close()
	}
	return nil
}

// Write sends metrics to one of the configured servers, logging each
// unsuccessful. If all servers fail, return an error.
func (i *InfluxDB) Write(metrics []telegraf.Metric) error {
	ctx := context.Background()

	allErrorsAreDatabaseNotFoundErrors := true
	var err error
	p := rand.Perm(len(i.clients))
	for _, n := range p {
		client := i.clients[n]
		err = client.Write(ctx, metrics)
		if err == nil {
			return nil
		}

		i.Log.Errorf("When writing to [%s]: %v", client.URL(), err)

		switch apiError := err.(type) {
		case *DatabaseNotFoundError:
			if i.SkipDatabaseCreation {
				continue
			}
			// retry control
			// error so the write is retried
			err := client.CreateDatabase(ctx, apiError.Database)
			if err != nil {
				i.Log.Errorf("When writing to [%s]: database %q not found and failed to recreate",
					client.URL(), apiError.Database)
			} else {
				return errors.New("database created; retry write")
			}
		default:
			allErrorsAreDatabaseNotFoundErrors = false
		}
	}

	if allErrorsAreDatabaseNotFoundErrors {
		// return nil because we should not be retrying this
		return nil
	}
	return errors.New("could not write any address")
}

func (i *InfluxDB) udpClient(url *url.URL) (Client, error) {
	config := &UDPConfig{
		URL:            url,
		MaxPayloadSize: int(i.UDPPayload),
		Serializer:     i.newSerializer(),
		Log:            i.Log,
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
		URL:                       url,
		Timeout:                   time.Duration(i.Timeout),
		TLSConfig:                 tlsConfig,
		UserAgent:                 i.UserAgent,
		Username:                  i.Username,
		Password:                  i.Password,
		Proxy:                     proxy,
		ContentEncoding:           i.ContentEncoding,
		Headers:                   i.HTTPHeaders,
		Database:                  i.Database,
		DatabaseTag:               i.DatabaseTag,
		ExcludeDatabaseTag:        i.ExcludeDatabaseTag,
		SkipDatabaseCreation:      i.SkipDatabaseCreation,
		RetentionPolicy:           i.RetentionPolicy,
		RetentionPolicyTag:        i.RetentionPolicyTag,
		ExcludeRetentionPolicyTag: i.ExcludeRetentionPolicyTag,
		Consistency:               i.WriteConsistency,
		Serializer:                i.newSerializer(),
		Log:                       i.Log,
	}

	c, err := i.CreateHTTPClientF(config)
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP client [%s]: %v", url, err)
	}

	if !i.SkipDatabaseCreation {
		err = c.CreateDatabase(ctx, c.Database())
		if err != nil {
			i.Log.Warnf("When writing to [%s]: database %q creation failed: %v",
				c.URL(), c.Database(), err)
		}
	}

	return c, nil
}

func (i *InfluxDB) newSerializer() *influx.Serializer {
	serializer := influx.NewSerializer()
	if i.InfluxUintSupport {
		serializer.SetFieldTypeSupport(influx.UintSupport)
	}

	return serializer
}

func init() {
	outputs.Add("influxdb", func() telegraf.Output {
		return &InfluxDB{
			Timeout: config.Duration(time.Second * 5),
			CreateHTTPClientF: func(config *HTTPConfig) (Client, error) {
				return NewHTTPClient(*config)
			},
			CreateUDPClientF: func(config *UDPConfig) (Client, error) {
				return NewUDPClient(*config)
			},
			ContentEncoding: "gzip",
		}
	})
}
