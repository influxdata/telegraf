//go:generate ../../../tools/readme_config_includer/generator
package influxdb_v3

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	common_http "github.com/influxdata/telegraf/plugins/common/http"
	"github.com/influxdata/telegraf/plugins/common/ratelimiter"
	"github.com/influxdata/telegraf/plugins/outputs"
)

//go:embed sample.conf
var sampleConfig string

type clientConfig struct {
	LocalAddr          string          `toml:"local_address"`
	Token              config.Secret   `toml:"token"`
	Database           string          `toml:"database"`
	DatabaseTag        string          `toml:"database_tag"`
	ExcludeDatabaseTag bool            `toml:"exclude_database_tag"`
	Sync               *bool           `toml:"sync"`
	Timeout            config.Duration `toml:"timeout"`
	UserAgent          string          `toml:"user_agent"`
	ContentEncoding    string          `toml:"content_encoding"`
	UintSupport        bool            `toml:"influx_uint_support"`
	OmitTimestamp      bool            `toml:"influx_omit_timestamp"`
	common_http.HTTPClientConfig
	ratelimiter.RateLimitConfig
}

type InfluxDB struct {
	URLs []string        `toml:"urls"`
	Log  telegraf.Logger `toml:"-"`
	clientConfig

	clients []*client
}

func (*InfluxDB) SampleConfig() string {
	return sampleConfig
}

func (i *InfluxDB) Init() error {
	// Set defaults
	if len(i.URLs) == 0 {
		i.URLs = append(i.URLs, "http://localhost:8181")
	}

	if i.UserAgent == "" {
		i.UserAgent = internal.ProductToken()
	}

	// Check input values
	switch i.ContentEncoding {
	case "":
		i.ContentEncoding = "gzip"
	case "none", "identity", "gzip":
	default:
		return fmt.Errorf("invalid content encoding %q", i.ContentEncoding)
	}

	// Create the clients
	i.clients = make([]*client, 0, len(i.URLs))
	for _, addr := range i.URLs {
		c, err := newClient(addr, &i.clientConfig, i.Log)
		if err != nil {
			return fmt.Errorf("creating client for %q failed: %w", addr, err)
		}
		i.clients = append(i.clients, c)
	}

	return nil
}

func (i *InfluxDB) Connect() error {
	for idx, c := range i.clients {
		if err := c.connect(); err != nil {
			return fmt.Errorf("connecting to %q failed: %w", i.URLs[idx], err)
		}
	}

	return nil
}

func (i *InfluxDB) Close() error {
	for _, c := range i.clients {
		c.close()
	}
	return nil
}

// Write sends metrics to one of the configured servers, logging each
// unsuccessful. If all servers fail, return an error.
func (i *InfluxDB) Write(metrics []telegraf.Metric) error {
	ctx := context.Background()

	for _, n := range rand.Perm(len(i.clients)) {
		client := i.clients[n]
		if err := client.write(ctx, metrics); err != nil {
			addr := i.URLs[n]
			i.Log.Errorf("Writing to %q failed: %v", addr, err)
			var werr *internal.PartialWriteError
			if errors.As(err, &werr) || errors.Is(err, internal.ErrSizeLimitReached) {
				return err
			}
			continue
		}
		return nil
	}

	return errors.New("failed to send metrics to any configured server(s)")
}

func init() {
	outputs.Add("influxdb_v3", func() telegraf.Output {
		return &InfluxDB{
			clientConfig: clientConfig{
				UintSupport: true,
				Timeout:     config.Duration(time.Second * 5),
			},
		}
	})
}
