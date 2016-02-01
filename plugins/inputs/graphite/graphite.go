package graphite

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"

	"github.com/influxdata/telegraf/plugins/inputs"
)

const (
	udpBufferSize = 65536
)

type tcpConnection struct {
	conn        net.Conn
	connectTime time.Time
}

func (c *tcpConnection) Close() {
	c.conn.Close()
}

// Graphite represents a Graphite service.
type Graphite struct {
	BindAddress   string
	Protocol      string
	UdpReadBuffer int
	Separator     string
	Tags          []string
	Templates     []string

	mu sync.Mutex

	parser *Parser
	logger *log.Logger
	config *Config

	tcpConnectionsMu sync.Mutex
	tcpConnections   map[string]*tcpConnection

	ln      net.Listener
	addr    net.Addr
	udpConn *net.UDPConn

	wg   sync.WaitGroup
	done chan struct{}

	// channel for all incoming parsed points
	metricC chan telegraf.Metric
}

var sampleConfig = `
  bind_address = ":2003" # the bind address
  protocol = "tcp" # or "udp" protocol to read via
  udp_read_buffer = 8388608 # (8*1024*1024) UDP read buffer size

  ### If matching multiple measurement files, this string will be used to join the matched values.
  separator = "."
  
  ### Default tags that will be added to all metrics.  These can be overridden at the template level
  ### or by tags extracted from metric
  tags = ["region=north-china", "zone=1c"]
  
  ### Each template line requires a template pattern.  It can have an optional
  ### filter before the template and separated by spaces.  It can also have optional extra
  ### tags following the template.  Multiple tags should be separated by commas and no spaces
  ### similar to the line protocol format.  The can be only one default template.
  ### Templates support below format:
  ### filter + template
  ### filter + template + extra tag
  ### filter + template with field key
  ### default template. Ignore the first graphite component "servers"
  templates = [
    "*.app env.service.resource.measurement",
    "stats.* .host.measurement* region=us-west,agent=sensu",
    "stats2.* .host.measurement.field",
    "measurement*"
 ]
`

func (g *Graphite) SampleConfig() string {
	return sampleConfig
}

func (g *Graphite) Description() string {
	return "Graphite read line-protocol metrics from tcp/udp socket"
}

// Open starts the Graphite input processing data.
func (g *Graphite) Start() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	c := &Config{
		BindAddress:   g.BindAddress,
		Protocol:      g.Protocol,
		UdpReadBuffer: g.UdpReadBuffer,
		Separator:     g.Separator,
		Tags:          g.Tags,
		Templates:     g.Templates,
	}
	c.WithDefaults()
	if err := c.Validate(); err != nil {
		return fmt.Errorf("Graphite input configuration is error! ", err.Error())
	}
	g.config = c

	parser, err := NewParserWithOptions(Options{
		Templates:   g.config.Templates,
		DefaultTags: g.config.DefaultTags(),
		Separator:   g.config.Separator})
	if err != nil {
		return fmt.Errorf("Graphite input parser config is error! ", err.Error())
	}
	g.parser = parser

	g.tcpConnections = make(map[string]*tcpConnection)
	g.done = make(chan struct{})
	g.metricC = make(chan telegraf.Metric, 10000)

	if strings.ToLower(g.config.Protocol) == "tcp" {
		g.addr, err = g.openTCPServer()
	} else if strings.ToLower(g.config.Protocol) == "udp" {
		g.addr, err = g.openUDPServer()
	} else {
		return fmt.Errorf("unrecognized Graphite input protocol %s", g.config.Protocol)
	}
	if err != nil {
		return err
	}

	g.logger.Printf("Listening on %s: %s", strings.ToUpper(g.config.Protocol), g.config.BindAddress)
	return nil
}

func (g *Graphite) closeAllConnections() {
	g.tcpConnectionsMu.Lock()
	defer g.tcpConnectionsMu.Unlock()
	for _, c := range g.tcpConnections {
		c.Close()
	}
}

// Close stops all data processing on the Graphite input.
func (g *Graphite) Stop() {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.closeAllConnections()

	if g.ln != nil {
		g.ln.Close()
	}
	if g.udpConn != nil {
		g.udpConn.Close()
	}

	close(g.done)
	g.wg.Wait()
	g.done = nil
}

// openTCPServer opens the Graphite input in TCP mode and starts processing data.
func (g *Graphite) openTCPServer() (net.Addr, error) {
	ln, err := net.Listen("tcp", g.config.BindAddress)
	if err != nil {
		return nil, err
	}
	g.ln = ln

	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		for {
			conn, err := g.ln.Accept()
			if opErr, ok := err.(*net.OpError); ok && !opErr.Temporary() {
				g.logger.Println("graphite TCP listener closed")
				return
			}
			if err != nil {
				g.logger.Println("error accepting TCP connection", err.Error())
				continue
			}

			g.wg.Add(1)
			go g.handleTCPConnection(conn)
		}
	}()
	return ln.Addr(), nil
}

// handleTCPConnection services an individual TCP connection for the Graphite input.
func (g *Graphite) handleTCPConnection(conn net.Conn) {
	defer g.wg.Done()
	defer conn.Close()
	defer g.untrackConnection(conn)

	g.trackConnection(conn)
	reader := bufio.NewReader(conn)

	for {
		// Read up to the next newline.
		buf, err := reader.ReadBytes('\n')
		if err != nil {
			return
		}

		// Trim the buffer, even though there should be no padding
		line := strings.TrimSpace(string(buf))
		g.handleLine(line)
	}
}

func (g *Graphite) trackConnection(c net.Conn) {
	g.tcpConnectionsMu.Lock()
	defer g.tcpConnectionsMu.Unlock()
	g.tcpConnections[c.RemoteAddr().String()] = &tcpConnection{
		conn:        c,
		connectTime: time.Now().UTC(),
	}
}
func (g *Graphite) untrackConnection(c net.Conn) {
	g.tcpConnectionsMu.Lock()
	defer g.tcpConnectionsMu.Unlock()
	delete(g.tcpConnections, c.RemoteAddr().String())
}

// openUDPServer opens the Graphite input in UDP mode and starts processing incoming data.
func (g *Graphite) openUDPServer() (net.Addr, error) {
	addr, err := net.ResolveUDPAddr("udp", g.config.BindAddress)
	if err != nil {
		return nil, err
	}

	g.udpConn, err = net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}

	if g.config.UdpReadBuffer != 0 {
		err = g.udpConn.SetReadBuffer(g.config.UdpReadBuffer)
		if err != nil {
			return nil, fmt.Errorf("unable to set UDP read buffer to %d: %s",
				g.config.UdpReadBuffer, err)
		}
	}

	buf := make([]byte, udpBufferSize)
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		for {
			n, _, err := g.udpConn.ReadFromUDP(buf)
			if err != nil {
				g.udpConn.Close()
				return
			}

			lines := strings.Split(string(buf[:n]), "\n")
			for _, line := range lines {
				g.handleLine(line)
			}
		}
	}()
	return g.udpConn.LocalAddr(), nil
}

func (g *Graphite) handleLine(line string) {
	if line == "" {
		return
	}

	// Parse it.
	metric, err := g.parser.Parse(line)
	if err != nil {
		switch err := err.(type) {
		case *UnsupposedValueError:
			// Graphite ignores NaN values with no error.
			if math.IsNaN(err.Value) {
				return
			}
		}
		g.logger.Printf("unable to parse line: %s: %s", line, err)
		return
	}

	g.metricC <- metric
}

func (g *Graphite) Gather(acc telegraf.Accumulator) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	npoints := len(g.metricC)
	for i := 0; i < npoints; i++ {
		metric := <-g.metricC
		acc.AddFields(metric.Name(), metric.Fields(), metric.Tags(), metric.Time())
	}
	return nil
}

func init() {
	inputs.Add("graphite", func() telegraf.Input {

		return &Graphite{logger: log.New(os.Stderr, "[graphite] ", log.LstdFlags)}
	})
}
