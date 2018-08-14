package graphite

import (
	"crypto/tls"
	"errors"
	"io"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/influxdata/telegraf"
	tlsint "github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

type Graphite struct {
	GraphiteTagSupport bool
	// URL is only for backwards compatibility
	Servers  []string
	Prefix   string
	Template string
	Timeout  int
	conns    []net.Conn
	tlsint.ClientConfig
}

var sampleConfig = `
  ## TCP endpoint for your graphite instance.
  ## If multiple endpoints are configured, output will be load balanced.
  ## Only one of the endpoints will be written to with each iteration.
  servers = ["localhost:2003"]
  ## Prefix metrics name
  prefix = ""
  ## Graphite output template
  ## see https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  template = "host.tags.measurement.field"

  ## Enable Graphite tags support
  # graphite_tag_support = false

  ## timeout in seconds for the write connection to graphite
  timeout = 2

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
`

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
		conn.Close()
	}
	return nil
}

func (g *Graphite) SampleConfig() string {
	return sampleConfig
}

func (g *Graphite) Description() string {
	return "Configuration for Graphite server to send metrics to"
}

// We need check eof as we can write to nothing without noticing anything is wrong
// the connection stays in a close_wait
// We can detect that by finding an eof
// if not for this, we can happily write and flush without getting errors (in Go) but getting RST tcp packets back (!)
// props to Tv via the authors of carbon-relay-ng` for this trick.
func checkEOF(conn net.Conn) {
	b := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
	num, err := conn.Read(b)
	if err == io.EOF {
		log.Printf("E! Conn %s is closed. closing conn explicitly", conn)
		conn.Close()
		return
	}
	// just in case i misunderstand something or the remote behaves badly
	if num != 0 {
		log.Printf("I! conn %s .conn.Read data? did not expect that.  data: %s\n", conn, b[:num])
	}
	// Log non-timeout errors or close.
	if e, ok := err.(net.Error); !(ok && e.Timeout()) {
		log.Printf("E! conn %s checkEOF .conn.Read returned err != EOF, which is unexpected.  closing conn. error: %s\n", conn, err)
		conn.Close()
	}
}

// Choose a random server in the cluster to write to until a successful write
// occurs, logging each unsuccessful. If all servers fail, return error.
func (g *Graphite) Write(metrics []telegraf.Metric) error {
	// Prepare data
	var batch []byte
	s, err := serializers.NewGraphiteSerializer(g.Prefix, g.Template, g.GraphiteTagSupport)
	if err != nil {
		return err
	}

	for _, metric := range metrics {
		buf, err := s.Serialize(metric)
		if err != nil {
			log.Printf("E! Error serializing some metrics to graphite: %s", err.Error())
		}
		batch = append(batch, buf...)
	}

	err = g.send(batch)

	// try to reconnect and retry to send
	if err != nil {
		log.Println("E! Graphite: Reconnecting and retrying: ")
		g.Connect()
		err = g.send(batch)
	}

	return err
}

func (g *Graphite) send(batch []byte) error {
	// This will get set to nil if a successful write occurs
	err := errors.New("Could not write to any Graphite server in cluster\n")

	// Send data to a random server
	p := rand.Perm(len(g.conns))
	for _, n := range p {
		if g.Timeout > 0 {
			g.conns[n].SetWriteDeadline(time.Now().Add(time.Duration(g.Timeout) * time.Second))
		}
		checkEOF(g.conns[n])
		if _, e := g.conns[n].Write(batch); e != nil {
			// Error
			log.Println("E! Graphite Error: " + e.Error())
			// Close explicitely
			g.conns[n].Close()
			// Let's try the next one
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
