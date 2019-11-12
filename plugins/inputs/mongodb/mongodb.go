package mongodb

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	tlsint "github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"gopkg.in/mgo.v2"
)

type MongoDB struct {
	Servers          []string
	Ssl              Ssl
	mongos           map[string]*Server
	GatherPerdbStats bool
	GatherColStats   bool
	ColStatsDbs      []string
	tlsint.ClientConfig

	Log telegraf.Logger
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

  ## When true, collect per database stats
  # gather_perdb_stats = false

  ## When true, collect per collection stats
  # gather_col_stats = false

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

var localhost = &url.URL{Host: "mongodb://127.0.0.1:27017"}

// Reads stats from all configured servers accumulates stats.
// Returns one of the errors encountered while gather stats (if any).
func (m *MongoDB) Gather(acc telegraf.Accumulator) error {
	if len(m.Servers) == 0 {
		m.gatherServer(m.getMongoServer(localhost), acc)
		return nil
	}

	var wg sync.WaitGroup
	for i, serv := range m.Servers {
		if !strings.HasPrefix(serv, "mongodb://") {
			// Preserve backwards compatibility for hostnames without a
			// scheme, broken in go 1.8. Remove in Telegraf 2.0
			serv = "mongodb://" + serv
			m.Log.Warnf("Using %q as connection URL; please update your configuration to use an URL", serv)
			m.Servers[i] = serv
		}

		u, err := url.Parse(serv)
		if err != nil {
			m.Log.Errorf("Unable to parse address %q: %s", serv, err.Error())
			continue
		}
		if u.Host == "" {
			m.Log.Errorf("Unable to parse address %q", serv)
			continue
		}

		wg.Add(1)
		go func(srv *Server) {
			defer wg.Done()
			err := m.gatherServer(srv, acc)
			if err != nil {
				m.Log.Errorf("Error in plugin: %v", err)
			}
		}(m.getMongoServer(u))
	}

	wg.Wait()
	return nil
}

func (m *MongoDB) getMongoServer(url *url.URL) *Server {
	if _, ok := m.mongos[url.Host]; !ok {
		m.mongos[url.Host] = &Server{
			Log: m.Log,
			Url: url,
		}
	}
	return m.mongos[url.Host]
}

func (m *MongoDB) gatherServer(server *Server, acc telegraf.Accumulator) error {
	if server.Session == nil {
		var dialAddrs []string
		if server.Url.User != nil {
			dialAddrs = []string{server.Url.String()}
		} else {
			dialAddrs = []string{server.Url.Host}
		}
		dialInfo, err := mgo.ParseURL(dialAddrs[0])
		if err != nil {
			return fmt.Errorf("unable to parse URL %q: %s", dialAddrs[0], err.Error())
		}
		dialInfo.Direct = true
		dialInfo.Timeout = 5 * time.Second

		var tlsConfig *tls.Config

		if m.Ssl.Enabled {
			// Deprecated TLS config
			tlsConfig = &tls.Config{}
			if len(m.Ssl.CaCerts) > 0 {
				roots := x509.NewCertPool()
				for _, caCert := range m.Ssl.CaCerts {
					ok := roots.AppendCertsFromPEM([]byte(caCert))
					if !ok {
						return fmt.Errorf("failed to parse root certificate")
					}
				}
				tlsConfig.RootCAs = roots
			} else {
				tlsConfig.InsecureSkipVerify = true
			}
		} else {
			tlsConfig, err = m.ClientConfig.TLSConfig()
			if err != nil {
				return err
			}
		}

		// If configured to use TLS, add a dial function
		if tlsConfig != nil {
			dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
				conn, err := tls.Dial("tcp", addr.String(), tlsConfig)
				if err != nil {
					fmt.Printf("error in Dial, %s\n", err.Error())
				}
				return conn, err
			}
		}

		sess, err := mgo.DialWithInfo(dialInfo)
		if err != nil {
			return fmt.Errorf("unable to connect to MongoDB: %s", err.Error())
		}
		server.Session = sess
	}
	return server.gatherData(acc, m.GatherPerdbStats, m.GatherColStats, m.ColStatsDbs)
}

func init() {
	inputs.Add("mongodb", func() telegraf.Input {
		return &MongoDB{
			ColStatsDbs: []string{"local"},
			mongos:      make(map[string]*Server),
		}
	})
}
