package rethinkdb

import (
	"fmt"
	"net/url"
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"

	"gopkg.in/gorethink/gorethink.v3"
)

type RethinkDB struct {
	Servers []string
}

var sampleConfig = `
  ## An array of URI to gather stats about. Specify an ip or hostname
  ## with optional port add password. ie,
  ##   rethinkdb://user:auth_key@10.10.3.30:28105,
  ##   rethinkdb://10.10.3.33:18832,
  ##   10.0.0.1:10000, etc.
  servers = ["127.0.0.1:28015"]
  ##
  ## If you use actual rethinkdb of > 2.3.0 with username/password authorization,
  ## protocol have to be named "rethinkdb2" - it will use 1_0 H.
  # servers = ["rethinkdb2://username:password@127.0.0.1:28015"]
  ##
  ## If you use older versions of rethinkdb (<2.2) with auth_key, protocol
  ## have to be named "rethinkdb".
  # servers = ["rethinkdb://username:auth_key@127.0.0.1:28015"]
`

func (r *RethinkDB) SampleConfig() string {
	return sampleConfig
}

func (r *RethinkDB) Description() string {
	return "Read metrics from one or many RethinkDB servers"
}

var localhost = &Server{URL: &url.URL{Host: "127.0.0.1:28015"}}

// Reads stats from all configured servers accumulates stats.
// Returns one of the errors encountered while gather stats (if any).
func (r *RethinkDB) Gather(acc telegraf.Accumulator) error {
	if len(r.Servers) == 0 {
		return r.gatherServer(localhost, acc)
	}

	var wg sync.WaitGroup

	for _, serv := range r.Servers {
		u, err := url.Parse(serv)
		if err != nil {
			acc.AddError(fmt.Errorf("unable to parse to address '%s': %s", serv, err))
			continue
		} else if u.Scheme == "" {
			// fallback to simple string based address (i.e. "10.0.0.1:10000")
			u.Host = serv
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			acc.AddError(r.gatherServer(&Server{URL: u}, acc))
		}()
	}

	wg.Wait()

	return nil
}

func (r *RethinkDB) gatherServer(server *Server, acc telegraf.Accumulator) error {
	var err error
	connectOpts := gorethink.ConnectOpts{
		Address:       server.URL.Host,
		DiscoverHosts: false,
	}
	if server.URL.User != nil {
		pwd, set := server.URL.User.Password()
		if set && pwd != "" {
			connectOpts.AuthKey = pwd
			connectOpts.HandshakeVersion = gorethink.HandshakeV0_4
		}
	}
	if server.URL.Scheme == "rethinkdb2" && server.URL.User != nil {
		pwd, set := server.URL.User.Password()
		if set && pwd != "" {
			connectOpts.Username = server.URL.User.Username()
			connectOpts.Password = pwd
			connectOpts.HandshakeVersion = gorethink.HandshakeV1_0
		}
	}

	server.session, err = gorethink.Connect(connectOpts)
	if err != nil {
		return fmt.Errorf("unable to connect to RethinkDB, %s", err.Error())
	}
	defer server.session.Close()

	return server.gatherData(acc)
}

func init() {
	inputs.Add("rethinkdb", func() telegraf.Input {
		return &RethinkDB{}
	})
}
