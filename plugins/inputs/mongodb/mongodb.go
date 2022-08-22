//go:generate ../../../tools/readme_config_includer/generator
package mongodb

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	_ "embed"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/influxdata/telegraf"
	tlsint "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// DO NOT REMOVE THE NEXT TWO LINES! This is required to embed the sampleConfig data.
//
//go:embed sample.conf
var sampleConfig string

type MongoDB struct {
	Servers                []string
	Ssl                    Ssl
	GatherClusterStatus    bool
	GatherPerdbStats       bool
	GatherColStats         bool
	GatherTopStat          bool
	IgnoreUnreachableHosts bool
	ColStatsDbs            []string
	tlsint.ClientConfig

	Log telegraf.Logger `toml:"-"`

	clients   []*Server
	tlsConfig *tls.Config
}

type Ssl struct {
	Enabled bool     `toml:"ssl_enabled" deprecated:"1.3.0;use 'tls_*' options instead"`
	CaCerts []string `toml:"cacerts" deprecated:"1.3.0;use 'tls_ca' instead"`
}

func (*MongoDB) SampleConfig() string {
	return sampleConfig
}

func (m *MongoDB) Init() error {
	if m.Ssl.Enabled {
		// Deprecated TLS config
		m.tlsConfig = &tls.Config{
			InsecureSkipVerify: m.ClientConfig.InsecureSkipVerify,
		}
		if len(m.Ssl.CaCerts) == 0 {
			return fmt.Errorf("you must explicitly set insecure_skip_verify to skip cerificate validation")
		}

		roots := x509.NewCertPool()
		for _, caCert := range m.Ssl.CaCerts {
			if ok := roots.AppendCertsFromPEM([]byte(caCert)); !ok {
				return fmt.Errorf("failed to parse root certificate")
			}
		}
		m.tlsConfig.RootCAs = roots
	} else {
		var err error
		m.tlsConfig, err = m.ClientConfig.TLSConfig()
		if err != nil {
			return err
		}
	}

	if len(m.Servers) == 0 {
		m.Servers = []string{"mongodb://127.0.0.1:27017"}
	}

	return nil
}

// Start runs after init and setup mongodb connections
func (m *MongoDB) Start() error {
	for _, connURL := range m.Servers {
		if !strings.HasPrefix(connURL, "mongodb://") && !strings.HasPrefix(connURL, "mongodb+srv://") {
			// Preserve backwards compatibility for hostnames without a
			// scheme, broken in go 1.8. Remove in Telegraf 2.0
			connURL = "mongodb://" + connURL
			m.Log.Warnf("Using %q as connection URL; please update your configuration to use an URL", connURL)
		}

		u, err := url.Parse(connURL)
		if err != nil {
			return fmt.Errorf("unable to parse connection URL: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel() //nolint:revive

		opts := options.Client().ApplyURI(connURL)
		if m.tlsConfig != nil {
			opts.TLSConfig = m.tlsConfig
		}
		if opts.ReadPreference == nil {
			opts.ReadPreference = readpref.Nearest()
		}

		client, err := mongo.Connect(ctx, opts)
		if err != nil {
			return fmt.Errorf("unable to connect to MongoDB: %w", err)
		}

		err = client.Ping(ctx, opts.ReadPreference)
		if err != nil {
			if !m.IgnoreUnreachableHosts {
				return fmt.Errorf("unable to connect to MongoDB: %w", err)
			}

			m.Log.Errorf("unable to connect to MongoDB: %w", err)
		}

		server := &Server{
			client:   client,
			hostname: u.Host,
			Log:      m.Log,
		}
		m.clients = append(m.clients, server)
	}

	return nil
}

// Reads stats from all configured servers accumulates stats.
// Returns one of the errors encountered while gather stats (if any).
func (m *MongoDB) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup
	for _, client := range m.clients {
		wg.Add(1)
		go func(srv *Server) {
			defer wg.Done()
			if m.IgnoreUnreachableHosts {
				if err := srv.ping(); err != nil {
					return
				}
			}

			err := srv.gatherData(acc, m.GatherClusterStatus, m.GatherPerdbStats, m.GatherColStats, m.GatherTopStat, m.ColStatsDbs)
			if err != nil {
				m.Log.Errorf("failed to gather data: %q", err)
			}
		}(client)
	}

	wg.Wait()
	return nil
}

func init() {
	inputs.Add("mongodb", func() telegraf.Input {
		return &MongoDB{
			GatherClusterStatus: true,
			GatherPerdbStats:    false,
			GatherColStats:      false,
			GatherTopStat:       false,
			ColStatsDbs:         []string{"local"},
		}
	})
}
