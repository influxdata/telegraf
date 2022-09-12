//go:generate ../../../tools/readme_config_includer/generator
package graphite

import (
	"crypto/tls"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	tlsint "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

// DO NOT REMOVE THE NEXT TWO LINES! This is required to embed the sampleConfig data.
//
//go:embed sample.conf
var sampleConfig string

type Graphite struct {
	GraphiteTagSupport      bool   `toml:"graphite_tag_support"`
	GraphiteTagSanitizeMode string `toml:"graphite_tag_sanitize_mode"`
	GraphiteSeparator       string `toml:"graphite_separator"`
	// URL is only for backwards compatibility
	Servers   []string        `toml:"servers"`
	Prefix    string          `toml:"prefix"`
	Template  string          `toml:"template"`
	Templates []string        `toml:"templates"`
	Timeout   int             `toml:"timeout"`
	Log       telegraf.Logger `toml:"-"`

	conns []net.Conn
	tlsint.ClientConfig
	failedServers []string
}

func (*Graphite) SampleConfig() string {
	return sampleConfig
}

func (g *Graphite) Connect() error {
	// Set default values
	if g.Timeout <= 0 {
		g.Timeout = 2
	}
	if len(g.Servers) == 0 {
		g.Servers = append(g.Servers, "localhost:2003")
	}

	// Set tls config
	tlsConfig, err := g.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	// Only retry the failed servers
	servers := g.Servers
	if len(g.failedServers) > 0 {
		servers = g.failedServers
		// Remove failed server from exisiting connections
		var workingConns []net.Conn
		for _, conn := range g.conns {
			var found bool
			for _, server := range servers {
				if conn.RemoteAddr().String() == server {
					found = true
					break
				}
			}
			if !found {
				workingConns = append(workingConns, conn)
			}
		}
		g.conns = workingConns
	}

	// Get Connections
	var conns []net.Conn
	var failedServers []string
	for _, server := range servers {
		// Dialer with timeout
		d := net.Dialer{Timeout: time.Duration(g.Timeout) * time.Second}

		// Get secure connection if tls config is set
		var conn net.Conn
		if tlsConfig != nil {
			conn, err = tls.DialWithDialer(&d, "tcp", server, tlsConfig)
		} else {
			conn, err = d.Dial("tcp", server)
		}

		if err == nil {
			conns = append(conns, conn)
		} else {
			g.Log.Debugf("Failed to establish connection: %v", err)
			failedServers = append(failedServers, server)
		}
	}

	if len(g.failedServers) > 0 {
		g.conns = append(g.conns, conns...)
		g.failedServers = failedServers
	} else {
		g.conns = conns
	}

	return nil
}

func (g *Graphite) Close() error {
	// Closing all connections
	for _, conn := range g.conns {
		_ = conn.Close()
	}
	return nil
}

// We need check eof as we can write to nothing without noticing anything is wrong
// the connection stays in a close_wait
// We can detect that by finding an eof
// if not for this, we can happily write and flush without getting errors (in Go) but getting RST tcp packets back (!)
// props to Tv via the authors of carbon-relay-ng` for this trick.
func (g *Graphite) checkEOF(conn net.Conn) error {
	b := make([]byte, 1024)

	if err := conn.SetReadDeadline(time.Now().Add(10 * time.Millisecond)); err != nil {
		g.Log.Debugf("Couldn't set read deadline for connection due to error %v with remote address %s. closing conn explicitly", err, conn.RemoteAddr().String())
		err = conn.Close()
		g.Log.Debugf("Failed to close the connection: %v", err)
		return err
	}
	num, err := conn.Read(b)
	if err == io.EOF {
		g.Log.Debugf("Conn %s is closed. closing conn explicitly", conn.RemoteAddr().String())
		err = conn.Close()
		g.Log.Debugf("Failed to close the connection: %v", err)
		return err
	}
	// just in case i misunderstand something or the remote behaves badly
	if num != 0 {
		g.Log.Infof("conn %s .conn.Read data? did not expect that. data: %s", conn, b[:num])
	}
	// Log non-timeout errors and close.
	if e, ok := err.(net.Error); !(ok && e.Timeout()) {
		g.Log.Debugf("conn %s checkEOF .conn.Read returned err != EOF, which is unexpected.  closing conn. error: %s", conn, err)
		err = conn.Close()
		g.Log.Debugf("Failed to close the connection: %v", err)
		return err
	}

	return nil
}

// Choose a random server in the cluster to write to until a successful write
// occurs, logging each unsuccessful. If all servers fail, return error.
func (g *Graphite) Write(metrics []telegraf.Metric) error {
	// Prepare data
	var batch []byte
	s, err := serializers.NewGraphiteSerializer(g.Prefix, g.Template, g.GraphiteTagSupport, g.GraphiteTagSanitizeMode, g.GraphiteSeparator, g.Templates)
	if err != nil {
		return err
	}

	for _, metric := range metrics {
		buf, err := s.Serialize(metric)
		if err != nil {
			g.Log.Errorf("Error serializing some metrics to graphite: %s", err.Error())
		}
		batch = append(batch, buf...)
	}

	err = g.send(batch)

	// If a send failed for a server, try to reconnect to that server
	if len(g.failedServers) > 0 {
		g.Log.Debugf("Reconnecting and retrying for the following servers: %s", strings.Join(g.failedServers, ","))
		err = g.Connect()
		if err != nil {
			return fmt.Errorf("Failed to reconnect: %v", err)
		}
		err = g.send(batch)
	}

	return err
}

func (g *Graphite) send(batch []byte) error {
	// This will get set to nil if a successful write occurs
	globalErr := errors.New("could not write to any Graphite server in cluster")

	// Send data to a random server
	p := rand.Perm(len(g.conns))
	for _, n := range p {
		if g.Timeout > 0 {
			err := g.conns[n].SetWriteDeadline(time.Now().Add(time.Duration(g.Timeout) * time.Second))
			if err != nil {
				g.Log.Errorf("failed to set write deadline for %s: %v", g.conns[n].RemoteAddr().String(), err)
				// Mark server as failed so a new connection will be made
				g.failedServers = append(g.failedServers, g.conns[n].RemoteAddr().String())
			}
		}
		err := g.checkEOF(g.conns[n])
		if err != nil {
			// Mark server as failed so a new connection will be made
			g.failedServers = append(g.failedServers, g.conns[n].RemoteAddr().String())
			break
		}
		if _, e := g.conns[n].Write(batch); e != nil {
			// Error
			g.Log.Debugf("Graphite Error: " + e.Error())
			// Close explicitly and let's try the next one
			err := g.conns[n].Close()
			g.Log.Debugf("Failed to close the connection: %v", err)
			// Mark server as failed so a new connection will be made
			g.failedServers = append(g.failedServers, g.conns[n].RemoteAddr().String())
		} else {
			globalErr = nil
			break
		}
	}

	return globalErr
}

func init() {
	outputs.Add("graphite", func() telegraf.Output {
		return &Graphite{}
	})
}
