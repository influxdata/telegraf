package graphite

import (
	"crypto/tls"
	"errors"
	"io"
	"math/rand"
	"net"
	"time"

	"github.com/influxdata/telegraf"
	tlsint "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

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

	// Get Connections
	var conns []net.Conn
	for _, server := range g.Servers {
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
		}
	}
	g.conns = conns
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
func (g *Graphite) checkEOF(conn net.Conn) {
	b := make([]byte, 1024)

	if err := conn.SetReadDeadline(time.Now().Add(10 * time.Millisecond)); err != nil {
		g.Log.Errorf("Couldn't set read deadline for connection %s. closing conn explicitly", conn)
		_ = conn.Close()
		return
	}
	num, err := conn.Read(b)
	if err == io.EOF {
		g.Log.Errorf("Conn %s is closed. closing conn explicitly", conn)
		_ = conn.Close()
		return
	}
	// just in case i misunderstand something or the remote behaves badly
	if num != 0 {
		g.Log.Infof("conn %s .conn.Read data? did not expect that. data: %s", conn, b[:num])
	}
	// Log non-timeout errors or close.
	if e, ok := err.(net.Error); !(ok && e.Timeout()) {
		g.Log.Errorf("conn %s checkEOF .conn.Read returned err != EOF, which is unexpected.  closing conn. error: %s", conn, err)
		_ = conn.Close()
	}
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

	// try to reconnect and retry to send
	if err != nil {
		g.Log.Error("Graphite: Reconnecting and retrying...")
		_ = g.Connect()
		err = g.send(batch)
	}

	return err
}

func (g *Graphite) send(batch []byte) error {
	// This will get set to nil if a successful write occurs
	err := errors.New("could not write to any Graphite server in cluster")

	// Send data to a random server
	p := rand.Perm(len(g.conns))
	for _, n := range p {
		if g.Timeout > 0 {
			_ = g.conns[n].SetWriteDeadline(time.Now().Add(time.Duration(g.Timeout) * time.Second))
		}
		g.checkEOF(g.conns[n])
		if _, e := g.conns[n].Write(batch); e != nil {
			// Error
			g.Log.Errorf("Graphite Error: " + e.Error())
			// Close explicitly and let's try the next one
			_ = g.conns[n].Close()
		} else {
			// Success
			err = nil
			break
		}
	}

	return err
}

func init() {
	outputs.Add("graphite", func() telegraf.Output {
		return &Graphite{}
	})
}
