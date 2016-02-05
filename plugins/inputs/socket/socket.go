package socket

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"

	"github.com/influxdata/telegraf/internal/encoding"
	"github.com/influxdata/telegraf/plugins/inputs"

	_ "github.com/influxdata/telegraf/internal/encoding/graphite"
	_ "github.com/influxdata/telegraf/internal/encoding/influx"
)

const (
	// DefaultBindAddress is the default binding interface if none is specified.
	DefaultBindAddress = ":2003"

	// DefaultProtocol is the default IP protocol used by the Graphite input.
	DefaultProtocol = "tcp"

	// DefaultUDPReadBuffer is the default buffer size for the UDP listener.
	// Sets the size of the operating system's receive buffer associated with
	// the UDP traffic. Keep in mind that the OS must be able
	// to handle the number set here or the UDP listener will error and exit.
	//
	// DefaultReadBuffer = 0 means to use the OS default, which is usually too
	// small for high UDP performance.
	//
	// Increasing OS buffer limits:
	//     Linux:      sudo sysctl -w net.core.rmem_max=<read-buffer>
	//     BSD/Darwin: sudo sysctl -w kern.ipc.maxsockbuf=<read-buffer>
	DefaultUdpReadBuffer = 0

	udpBufferSize = 65536
)

type tcpConnection struct {
	conn        net.Conn
	connectTime time.Time
}

func (c *tcpConnection) Close() {
	c.conn.Close()
}

// Socket represents a Socket listening service.
type Socket struct {
	BindAddress   string
	Protocol      string
	UdpReadBuffer int

	DataFormat string

	Separator string
	Templates []string

	mu sync.Mutex

	encodingParser encoding.Parser

	logger *log.Logger

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

  # Data format to consume. This can be "influx" or "graphite" (line-protocol)
  # NOTE json only reads numerical measurements, strings and booleans are ignored.
  data_format = "graphite"

  ### If matching multiple measurement files, this string will be used to join the matched values.
  separator = "."

  ### Each template line requires a template pattern.  It can have an optional
  ### filter before the template and separated by spaces.  It can also have optional extra
  ### tags following the template.  Multiple tags should be separated by commas and no spaces
  ### similar to the line protocol format.  The can be only one default template.
  ### Templates support below format:
  ### 1. filter + template
  ### 2. filter + template + extra tag
  ### 3. filter + template with field key
  ### 4. default template
  templates = [
    "*.app env.service.resource.measurement",
    "stats.* .host.measurement* region=us-west,agent=sensu",
    "stats2.* .host.measurement.field",
    "measurement*"
 ]
`

func (s *Socket) SampleConfig() string {
	return sampleConfig
}

func (s *Socket) Description() string {
	return "Socket read influx or graphite line-protocol metrics from tcp/udp socket"
}

// Open starts the Socket input processing data.
func (s *Socket) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.BindAddress == "" {
		s.BindAddress = DefaultBindAddress
	}
	if s.Protocol == "" {
		s.Protocol = DefaultProtocol
	}
	if s.UdpReadBuffer < 0 {
		s.UdpReadBuffer = DefaultUdpReadBuffer
	}

	configs := make(map[string]interface{})
	configs["Separator"] = s.Separator
	configs["Templates"] = s.Templates

	var err error
	s.encodingParser, err = encoding.NewParser(s.DataFormat, configs)

	if err != nil {
		return fmt.Errorf("Socket input configuration is error: %s ", err.Error())
	}

	s.tcpConnections = make(map[string]*tcpConnection)
	s.done = make(chan struct{})
	s.metricC = make(chan telegraf.Metric, 50000)

	if strings.ToLower(s.Protocol) == "tcp" {
		s.addr, err = s.openTCPServer()
	} else if strings.ToLower(s.Protocol) == "udp" {
		s.addr, err = s.openUDPServer()
	} else {
		return fmt.Errorf("unrecognized Socket input protocol %s", s.Protocol)
	}
	if err != nil {
		return err
	}

	s.logger.Printf("Socket Plugin Listening on %s: %s", strings.ToUpper(s.Protocol), s.BindAddress)
	return nil
}

func (s *Socket) closeAllConnections() {
	s.tcpConnectionsMu.Lock()
	defer s.tcpConnectionsMu.Unlock()
	for _, c := range s.tcpConnections {
		c.Close()
	}
}

// Close stops all data processing on the Socket input.
func (s *Socket) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.closeAllConnections()

	if s.ln != nil {
		s.ln.Close()
	}
	if s.udpConn != nil {
		s.udpConn.Close()
	}

	close(s.done)
	s.wg.Wait()
	s.done = nil
}

// openTCPServer opens the Socket input in TCP mode and starts processing data.
func (s *Socket) openTCPServer() (net.Addr, error) {
	ln, err := net.Listen("tcp", s.BindAddress)
	if err != nil {
		return nil, err
	}
	s.ln = ln

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			conn, err := s.ln.Accept()
			if opErr, ok := err.(*net.OpError); ok && !opErr.Temporary() {
				s.logger.Println("Socket TCP listener closed")
				return
			}
			if err != nil {
				s.logger.Println("error accepting TCP connection", err.Error())
				continue
			}

			s.wg.Add(1)
			go s.handleTCPConnection(conn)
		}
	}()
	return ln.Addr(), nil
}

// handleTCPConnection services an individual TCP connection for the Socket input.
func (s *Socket) handleTCPConnection(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()
	defer s.untrackConnection(conn)

	s.trackConnection(conn)
	reader := bufio.NewReader(conn)

	for {
		// Read up to the next newline.
		buf, err := reader.ReadBytes('\n')
		if err != nil {
			return
		}

		// Trim the buffer, even though there should be no padding
		// line := strings.TrimSpace(string(buf))
		s.handleLines(buf)
	}
}

func (s *Socket) trackConnection(c net.Conn) {
	s.tcpConnectionsMu.Lock()
	defer s.tcpConnectionsMu.Unlock()
	s.tcpConnections[c.RemoteAddr().String()] = &tcpConnection{
		conn:        c,
		connectTime: time.Now().UTC(),
	}
}
func (s *Socket) untrackConnection(c net.Conn) {
	s.tcpConnectionsMu.Lock()
	defer s.tcpConnectionsMu.Unlock()
	delete(s.tcpConnections, c.RemoteAddr().String())
}

// openUDPServer opens the Socket input in UDP mode and starts processing incoming data.
func (s *Socket) openUDPServer() (net.Addr, error) {
	addr, err := net.ResolveUDPAddr("udp", s.BindAddress)
	if err != nil {
		return nil, err
	}

	s.udpConn, err = net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}

	if s.UdpReadBuffer != 0 {
		err = s.udpConn.SetReadBuffer(s.UdpReadBuffer)
		if err != nil {
			return nil, fmt.Errorf("unable to set UDP read buffer to %d: %s",
				s.UdpReadBuffer, err)
		}
	}

	buf := make([]byte, udpBufferSize)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			n, _, err := s.udpConn.ReadFromUDP(buf)
			if err != nil {
				s.udpConn.Close()
				return
			}

			s.handleLines(buf[:n])
		}
	}()
	return s.udpConn.LocalAddr(), nil
}

func (s *Socket) handleLines(buf []byte) {
	if buf == nil || len(buf) < 1 {
		return
	}

	// Parse it.
	metrics, err := s.encodingParser.Parse(buf)
	if err != nil {
		s.logger.Printf("unable to parse lines: %s: %s", buf, err)
		return
	}

	for _, metric := range metrics {
		s.metricC <- metric
	}

}

func (s *Socket) Gather(acc telegraf.Accumulator) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	npoints := len(s.metricC)
	for i := 0; i < npoints; i++ {
		metric := <-s.metricC
		acc.AddFields(metric.Name(), metric.Fields(), metric.Tags(), metric.Time())
	}
	return nil
}

func init() {
	inputs.Add("socket", func() telegraf.Input {
		return &Socket{logger: log.New(os.Stderr, "[socket] ", log.LstdFlags)}
	})
}
