package mongodb

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	tlsint "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoDB struct {
	Servers             []string
	Ssl                 Ssl
	GatherClusterStatus bool
	GatherPerdbStats    bool
	GatherColStats      bool
	GatherTopStat       bool
	ColStatsDbs         []string
	tlsint.ClientConfig

	Log telegraf.Logger `toml:"-"`

	clients []*Server
}

type Ssl struct {
	Enabled bool
	CaCerts []string `toml:"cacerts"`
}

var sampleConfig = `
  ## An array of URLs of the form:
  ##   "mongodb://" [user ":" pass "@"] host [ ":" port]
  ## For example:
  ##   mongodb://user:auth_key@10.10.3.30:27017,
  ##   mongodb://10.10.3.33:18832,
  servers = ["mongodb://127.0.0.1:27017"]

  ## When true, collect cluster status
  ## Note that the query that counts jumbo chunks triggers a COLLSCAN, which
  ## may have an impact on performance.
  # gather_cluster_status = true

  ## When true, collect per database stats
  # gather_perdb_stats = false

  ## When true, collect per collection stats
  # gather_col_stats = false

  ## When true, collect usage statistics for each collection
  ## (insert, update, queries, remove, getmore, commands etc...).
  # gather_top_stat = false

  ## List of db where collections stats are collected
  ## If empty, all db are concerned
  # col_stats_dbs = ["local"]

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
`

func (m *MongoDB) SampleConfig() string {
	return sampleConfig
}

func (*MongoDB) Description() string {
	return "Read metrics from one or many MongoDB servers"
}

func (m *MongoDB) Init() error {
	var tlsConfig *tls.Config
	if m.Ssl.Enabled {
		// Deprecated TLS config
		tlsConfig = &tls.Config{}
		if len(m.Ssl.CaCerts) > 0 {
			roots := x509.NewCertPool()
			for _, caCert := range m.Ssl.CaCerts {
				if ok := roots.AppendCertsFromPEM([]byte(caCert)); !ok {
					return fmt.Errorf("failed to parse root certificate")
				}
			}
			tlsConfig.RootCAs = roots
		} else {
			tlsConfig.InsecureSkipVerify = true
		}
	} else {
		var err error
		tlsConfig, err = m.ClientConfig.TLSConfig()
		if err != nil {
			return err
		}
	}

	if len(m.Servers) == 0 {
		m.Servers = []string{"mongodb://127.0.0.1:27017"}
	}

	for _, connUrl := range m.Servers {
		u, err := url.Parse(connUrl)
		if err != nil {
			return fmt.Errorf("unable to parse connection URL: %q", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		opts := options.Client().ApplyURI(connUrl)
		if tlsConfig != nil {
			opts.TLSConfig = tlsConfig
		}
		if opts.ReadPreference == nil {
			opts.ReadPreference = readpref.Nearest()
		}

		client, err := mongo.Connect(ctx, opts)
		if err != nil {
			cancel()
			return fmt.Errorf("unable to connect to MongoDB: %q", err)
		}

		err = client.Ping(ctx, opts.ReadPreference)
		if err != nil {
			cancel()
			return fmt.Errorf("unable to connect to MongoDB: %s", err)
		}

		server := &Server{
			client:   client,
			hostname: u.Host,
			Log:      m.Log,
		}
		m.clients = append(m.clients, server)

		cancel()
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
