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
	"github.com/influxdata/telegraf/config"
	tlsint "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers/graphite"
)

//go:embed sample.conf
var sampleConfig string

var ErrNotConnected = errors.New("could not write to any server in cluster")

type connection struct {
	name      string
	conn      net.Conn
	connected bool
}

type Graphite struct {
	GraphiteTagSupport      bool   `toml:"graphite_tag_support"`
	GraphiteTagSanitizeMode string `toml:"graphite_tag_sanitize_mode"`
	GraphiteSeparator       string `toml:"graphite_separator"`
	GraphiteStrictRegex     string `toml:"graphite_strict_sanitize_regex"`
	// URL is only for backwards compatibility
	Servers   []string        `toml:"servers"`
	Prefix    string          `toml:"prefix"`
	Template  string          `toml:"template"`
	Templates []string        `toml:"templates"`
	Timeout   config.Duration `toml:"timeout"`
	Log       telegraf.Logger `toml:"-"`
	tlsint.ClientConfig

	connections []connection
	serializer  *graphite.GraphiteSerializer
}

func (*Graphite) SampleConfig() string {
	return sampleConfig
}

func (g *Graphite) Init() error {
	s := &graphite.GraphiteSerializer{
		Prefix:          g.Prefix,
		Template:        g.Template,
		StrictRegex:     g.GraphiteStrictRegex,
		TagSupport:      g.GraphiteTagSupport,
		TagSanitizeMode: g.GraphiteTagSanitizeMode,
		Separator:       g.GraphiteSeparator,
		Templates:       g.Templates,
	}
	if err := s.Init(); err != nil {
		return err
	}
	g.serializer = s

	// Set default values
	if len(g.Servers) == 0 {
		g.Servers = append(g.Servers, "localhost:2003")
	}

	// Fill in the connections from the server
	g.connections = make([]connection, 0, len(g.Servers))
	for _, server := range g.Servers {
		g.connections = append(g.connections, connection{
			name:      server,
			connected: false,
		})
	}

	return nil
}

func (g *Graphite) Connect() error {
	// Set tls config
	tlsConfig, err := g.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	// Find all non-connected servers and try to reconnect
	var newConnection bool
	var connectedServers int
	var failedServers []string
	for i, server := range g.connections {
		if server.connected {
			connectedServers++
			continue
		}
		newConnection = true

		// Dialer with timeout
		d := net.Dialer{Timeout: time.Duration(g.Timeout)}

		// Get secure connection if tls config is set
		var conn net.Conn
		if tlsConfig != nil {
			conn, err = tls.DialWithDialer(&d, "tcp", server.name, tlsConfig)
		} else {
			conn, err = d.Dial("tcp", server.name)
		}

		if err == nil {
			g.connections[i].conn = conn
			g.connections[i].connected = true
			connectedServers++
		} else {
			g.Log.Debugf("Failed to establish connection: %v", err)
			failedServers = append(failedServers, server.name)
		}
	}

	if newConnection {
		g.Log.Debugf("Successful connections: %d of %d", connectedServers, len(g.connections))
	}
	if len(failedServers) > 0 {
		g.Log.Debugf("Failed servers: %d", len(failedServers))
	}

	return nil
}

func (g *Graphite) Close() error {
	// Closing all connections
	for _, c := range g.connections {
		_ = c.conn.Close()
		c.connected = false
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
		g.Log.Debugf(
			"Couldn't set read deadline for connection due to error %v with remote address %s. closing conn explicitly",
			err,
			conn.RemoteAddr().String(),
		)
		err = conn.Close()
		g.Log.Debugf("Failed to close the connection: %v", err)
		return err
	}
	num, err := conn.Read(b)
	if errors.Is(err, io.EOF) {
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
	var netErr net.Error
	if !(errors.As(err, &netErr) && netErr.Timeout()) {
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
	for _, metric := range metrics {
		buf, err := g.serializer.Serialize(metric)
		if err != nil {
			g.Log.Errorf("Error serializing some metrics to graphite: %s", err.Error())
		}
		batch = append(batch, buf...)
	}

	// Try to connect to all servers not yet connected if any
	if err := g.Connect(); err != nil {
		return fmt.Errorf("failed to reconnect: %w", err)
	}

	// Return on success of if we encounter a non-retryable error
	if err := g.send(batch); err == nil || !errors.Is(err, ErrNotConnected) {
		return err
	}

	// Try to reconnect and resend
	failedServers := make([]string, 0, len(g.connections))
	for _, c := range g.connections {
		if !c.connected {
			failedServers = append(failedServers, c.name)
		}
	}
	if len(failedServers) > 0 {
		g.Log.Debugf("Reconnecting and retrying for the following servers: %s", strings.Join(failedServers, ","))
		if err := g.Connect(); err != nil {
			return fmt.Errorf("failed to reconnect: %w", err)
		}
	}

	return g.send(batch)
}

func (g *Graphite) send(batch []byte) error {
	// Try sending the data to a server. Try them in random order
	p := rand.Perm(len(g.connections))
	for i, n := range p {
		server := g.connections[n]

		// Skip unconnected servers
		if !server.connected {
			continue
		}

		if g.Timeout > 0 {
			deadline := time.Now().Add(time.Duration(g.Timeout))
			if err := server.conn.SetWriteDeadline(deadline); err != nil {
				g.Log.Warnf("failed to set write deadline for %q: %v", server.name, err)
				g.connections[n].connected = false
				continue
			}
		}

		// Check the connection state
		if err := g.checkEOF(server.conn); err != nil {
			// Mark server as failed so a new connection will be made
			g.connections[n].connected = false
			continue
		}
		_, err := server.conn.Write(batch)
		if err == nil {
			// Sending the data was successfully
			return nil
		}

		g.Log.Errorf("Writing to %q failed: %v", server.name, err)
		if i < len(p)-1 {
			g.Log.Info("Trying next server...")
		}
		// Mark server as failed so a new connection will be made
		if server.conn != nil {
			if err := server.conn.Close(); err != nil {
				g.Log.Debugf("Failed to close connection to %q: %v", server.name, err)
			}
		}
		g.connections[n].connected = false
	}

	// If we end here, none of the writes were successful
	return ErrNotConnected
}

func init() {
	outputs.Add("graphite", func() telegraf.Output {
		return &Graphite{Timeout: config.Duration(2 * time.Second)}
	})
}
