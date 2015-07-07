package mongodb

import (
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/influxdb/telegraf/plugins"
	"gopkg.in/mgo.v2"
)

type MongoDB struct {
	Servers []string
	mongos  map[string]*Server
}

var sampleConfig = `
# An array of URI to gather stats about. Specify an ip or hostname
# with optional port add password. ie mongodb://user:auth_key@10.10.3.30:27017,
# mongodb://10.10.3.33:18832, 10.0.0.1:10000, etc.
#
# If no servers are specified, then 127.0.0.1 is used as the host and 27107 as the port.
servers = ["127.0.0.1:27017"]`

func (m *MongoDB) SampleConfig() string {
	return sampleConfig
}

func (*MongoDB) Description() string {
	return "Read metrics from one or many MongoDB servers"
}

var localhost = &url.URL{Host: "127.0.0.1:27017"}

// Reads stats from all configured servers accumulates stats.
// Returns one of the errors encountered while gather stats (if any).
func (m *MongoDB) Gather(acc plugins.Accumulator) error {
	if len(m.Servers) == 0 {
		m.gatherServer(m.getMongoServer(localhost), acc)
		return nil
	}

	var wg sync.WaitGroup

	var outerr error

	for _, serv := range m.Servers {
		u, err := url.Parse(serv)
		if err != nil {
			return fmt.Errorf("Unable to parse to address '%s': %s", serv, err)
		} else if u.Scheme == "" {
			u.Scheme = "mongodb"
			// fallback to simple string based address (i.e. "10.0.0.1:10000")
			u.Host = serv
			if u.Path == u.Host {
				u.Path = ""
			}
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			outerr = m.gatherServer(m.getMongoServer(u), acc)
		}()
	}

	wg.Wait()

	return outerr
}

func (m *MongoDB) getMongoServer(url *url.URL) *Server {
	if _, ok := m.mongos[url.Host]; !ok {
		m.mongos[url.Host] = &Server{
			Url: url,
		}
	}
	return m.mongos[url.Host]
}

func (m *MongoDB) gatherServer(server *Server, acc plugins.Accumulator) error {
	if server.Session == nil {
		var dialAddrs []string
		if server.Url.User != nil {
			dialAddrs = []string{server.Url.String()}
		} else {
			dialAddrs = []string{server.Url.Host}
		}
		dialInfo, err := mgo.ParseURL(dialAddrs[0])
		if err != nil {
			return fmt.Errorf("Unable to parse URL (%s), %s\n", dialAddrs[0], err.Error())
		}
		dialInfo.Direct = true
		dialInfo.Timeout = time.Duration(10) * time.Second
		sess, err := mgo.DialWithInfo(dialInfo)
		if err != nil {
			return fmt.Errorf("Unable to connect to MongoDB, %s\n", err.Error())
		}
		server.Session = sess
	}
	return server.gatherData(acc)
}

func init() {
	plugins.Add("mongodb", func() plugins.Plugin {
		return &MongoDB{
			mongos: make(map[string]*Server),
		}
	})
}
