//go:generate ../../../tools/readme_config_includer/generator
package influxdb

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
	"github.com/influxdata/telegraf/selfstat"
)

//go:embed sample.conf
var sampleConfig string

var ErrMissingURL = errors.New("missing URL")

type Client interface {
	Write(context.Context, []telegraf.Metric) error
	CreateDatabase(ctx context.Context, database string) error
	Database() string
	URL() string
	Close()
}

// InfluxDB struct is the primary data structure for the plugin
type InfluxDB struct {
	URLs                      []string            `toml:"urls"`
	LocalAddr                 string              `toml:"local_address"`
	Username                  config.Secret       `toml:"username"`
	Password                  config.Secret       `toml:"password"`
	Database                  string              `toml:"database"`
	DatabaseTag               string              `toml:"database_tag"`
	ExcludeDatabaseTag        bool                `toml:"exclude_database_tag"`
	RetentionPolicy           string              `toml:"retention_policy"`
	RetentionPolicyTag        string              `toml:"retention_policy_tag"`
	ExcludeRetentionPolicyTag bool                `toml:"exclude_retention_policy_tag"`
	UserAgent                 string              `toml:"user_agent"`
	WriteConsistency          string              `toml:"write_consistency"`
	Timeout                   config.Duration     `toml:"timeout"`
	UDPPayload                config.Size         `toml:"udp_payload"`
	HTTPProxy                 string              `toml:"http_proxy"`
	HTTPHeaders               map[string]string   `toml:"http_headers"`
	ContentEncoding           string              `toml:"content_encoding"`
	SkipDatabaseCreation      bool                `toml:"skip_database_creation"`
	InfluxUintSupport         bool                `toml:"influx_uint_support"`
	OmitTimestamp             bool                `toml:"influx_omit_timestamp"`
	Log                       telegraf.Logger     `toml:"-"`
	Statistics                *selfstat.Collector `toml:"-"`
	tls.ClientConfig

	clients    []Client
	serializer *influx.Serializer

	bytesWritten selfstat.Stat

	CreateHTTPClientF func(config *HTTPConfig) (Client, error)
	CreateUDPClientF  func(config *UDPConfig) (Client, error)
}

func (*InfluxDB) SampleConfig() string {
	return sampleConfig
}

func (i *InfluxDB) Init() error {
	// Set default values
	if i.ContentEncoding == "" {
		i.ContentEncoding = "gzip"
	}

	if len(i.URLs) == 0 {
		i.URLs = append(i.URLs, "http://localhost:8086")
	}

	// Setup serializer
	i.serializer = &influx.Serializer{
		UintSupport:   i.InfluxUintSupport,
		OmitTimestamp: i.OmitTimestamp,
	}
	if err := i.serializer.Init(); err != nil {
		return fmt.Errorf("initializing serializer failed: %w", err)
	}

	// Register internal metrics
	i.bytesWritten = i.Statistics.Register("write", "bytes_written", nil)

	return nil
}

func (i *InfluxDB) Connect() error {
	ctx := context.Background()
	i.clients = make([]Client, 0, len(i.URLs))

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

		var localIP *net.IPAddr
		var localPort int
		if i.LocalAddr != "" {
			var err error
			// Resolve the local address into IP address and the given port if any
			addr, sPort, err := net.SplitHostPort(i.LocalAddr)
			if err != nil {
				if !strings.Contains(err.Error(), "missing port") {
					return fmt.Errorf("invalid local address: %w", err)
				}
				addr = i.LocalAddr
			}
			localIP, err = net.ResolveIPAddr("ip", addr)
			if err != nil {
				return fmt.Errorf("cannot resolve local address: %w", err)
			}

			if sPort != "" {
				p, err := strconv.ParseUint(sPort, 10, 16)
				if err != nil {
					return fmt.Errorf("invalid port: %w", err)
				}
				localPort = int(p)
			}
		}

		switch parts.Scheme {
		case "udp", "udp4", "udp6":
			var c Client
			var err error
			if i.LocalAddr == "" {
				c, err = i.udpClient(parts, nil)
			} else {
				c, err = i.udpClient(parts, &net.UDPAddr{IP: localIP.IP, Port: localPort, Zone: localIP.Zone})
			}
			if err != nil {
				return err
			}

			i.clients = append(i.clients, c)
		case "http", "https", "unix":
			var c Client
			var err error
			if i.LocalAddr == "" {
				c, err = i.httpClient(ctx, parts, nil, proxy)
			} else {
				c, err = i.httpClient(ctx, parts, &net.TCPAddr{IP: localIP.IP, Port: localPort, Zone: localIP.Zone}, proxy)
			}
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
	i.clients = make([]Client, 0)

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

		var apiError *DatabaseNotFoundError
		if errors.As(err, &apiError) {
			if i.SkipDatabaseCreation {
				continue
			}
			// retry control
			// error so the write is retried
			if err := client.CreateDatabase(ctx, apiError.Database); err == nil {
				return errors.New("database created; retry write")
			}
			i.Log.Errorf("When writing to [%s]: database %q not found and failed to recreate", client.URL(), apiError.Database)
		} else {
			allErrorsAreDatabaseNotFoundErrors = false
		}
	}

	if allErrorsAreDatabaseNotFoundErrors {
		// return nil because we should not be retrying this
		return nil
	}
	return errors.New("could not write any address")
}

func (i *InfluxDB) udpClient(address *url.URL, localAddr *net.UDPAddr) (Client, error) {
	udpConfig := &UDPConfig{
		URL:            address,
		LocalAddr:      localAddr,
		MaxPayloadSize: int(i.UDPPayload),
		Serializer:     i.serializer,
		Log:            i.Log,
		BytesWritten:   i.bytesWritten,
	}

	c, err := i.CreateUDPClientF(udpConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating UDP client [%s]: %w", address, err)
	}

	return c, nil
}

func (i *InfluxDB) httpClient(ctx context.Context, address *url.URL, localAddr *net.TCPAddr, proxy *url.URL) (Client, error) {
	tlsConfig, err := i.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	serializer := &influx.Serializer{
		UintSupport:   i.InfluxUintSupport,
		OmitTimestamp: i.OmitTimestamp,
	}
	if err := serializer.Init(); err != nil {
		return nil, err
	}

	httpConfig := &HTTPConfig{
		URL:                       address,
		LocalAddr:                 localAddr,
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
		Serializer:                serializer,
		Log:                       i.Log,
		BytesWritten:              i.bytesWritten,
	}

	c, err := i.CreateHTTPClientF(httpConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP client [%s]: %w", address, err)
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
		}
	})
}
