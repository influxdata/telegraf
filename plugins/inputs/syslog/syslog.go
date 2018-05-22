package syslog

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/go-syslog/rfc5424"
	"github.com/influxdata/go-syslog/rfc5425"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const defaultReadTimeout = time.Millisecond * 500
const ipMaxPacketSize = 64 * 1024

// Syslog is a syslog plugin
type Syslog struct {
	Address            string `toml:"server"`
	Protocol           string
	Cacert             string `toml:"tls_cacert"`
	Cert               string `toml:"tls_cert"`
	Key                string `toml:"tls_key"`
	InsecureSkipVerify bool
	KeepAlivePeriod    *internal.Duration
	ReadTimeout        *internal.Duration
	MaxConnections     int
	BestEffort         bool

	now func() time.Time

	mu sync.Mutex
	wg sync.WaitGroup
	io.Closer

	isTCP         bool
	tcpListener   net.Listener
	tlsConfig     *tls.Config
	connections   map[string]net.Conn
	connectionsMu sync.Mutex

	udpListener net.PacketConn
}

var sampleConfig = `
    ## Specify an ip or hostname with port - eg., localhost:6514, 10.0.0.1:6514
    ## Address and port to host the syslog receiver.
    ## If no server is specified, then localhost is used as the host.
    ## If no port is specified, 6514 is used (RFC5425#section-4.1).
	server = ":6514"
	
	## Protocol (default = tcp)
	## Should be one of the following values:
	## tcp, tcp4, tcp6, unix, unixpacket, udp, udp4, udp6, ip, ip4, ip6, unixgram.
	## Otherwise forced to the default.
	# protocol = "tcp"

    ## TLS Config
    # tls_cacert = "/etc/telegraf/ca.pem"
    # tls_cert = "/etc/telegraf/cert.pem"
    # tls_key = "/etc/telegraf/key.pem"
    ## If false, skip chain & host verification
	# insecure_skip_verify = true
	
	## Period between keep alive probes.
	## 0 disables keep alive probes.
	## Defaults to the OS configuration.
	## Only applies to stream sockets (e.g. TCP).
	# keep_alive_period = "5m"

	## Maximum number of concurrent connections (default = 0).
	## 0 means unlimited.
	## Only applies to stream sockets (e.g. TCP).
	# max_connections = 1024

	## Read timeout (default = 500ms).
	## 0 means unlimited.
	## Only applies to stream sockets (e.g. TCP).
	read_timeout = 500ms

	## Whether to parse in best effort mode or not (default = false).
	## By default best effort parsing is off.
	# best_effort = false
`

// SampleConfig returns sample configuration message
func (s *Syslog) SampleConfig() string {
	return sampleConfig
}

// Description returns the plugin description
func (s *Syslog) Description() string {
	return "Influx syslog receiver as per RFC5425"
}

// Gather ...
func (s *Syslog) Gather(_ telegraf.Accumulator) error {
	return nil
}

// Start starts the service.
func (s *Syslog) Start(acc telegraf.Accumulator) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// tags := map[string]string{
	// 	"address": s.Address,
	// }

	switch s.Protocol {
	case "tcp", "tcp4", "tcp6", "unix", "unixpacket":
		s.isTCP = true
	case "udp", "udp4", "udp6", "ip", "ip4", "ip6", "unixgram":
		s.isTCP = false
	default:
		s.Protocol = "tcp"
		s.isTCP = true
	}

	if s.Protocol == "unix" || s.Protocol == "unixpacket" || s.Protocol == "unixgram" {
		os.Remove(s.Address)
	}

	if s.isTCP {
		l, err := net.Listen(s.Protocol, s.Address)
		if err != nil {
			return err
		}
		s.Closer = l
		s.tcpListener = l
		if tlsConfig, _ := internal.GetTLSConfig(s.Cert, s.Key, s.Cacert, s.InsecureSkipVerify); tlsConfig != nil {
			s.tlsConfig = tlsConfig
		}

		s.wg.Add(1)
		go s.listenStream(acc)
	} else {
		l, err := net.ListenPacket(s.Protocol, s.Address)
		if err != nil {
			return err
		}
		s.Closer = l
		s.udpListener = l

		s.wg.Add(1)
		go s.listenPacket(acc)
	}

	if s.Protocol == "unix" || s.Protocol == "unixpacket" || s.Protocol == "unixgram" {
		s.Closer = unixCloser{path: s.Address, closer: s.Closer}
	}

	log.Printf("I! Started syslog receiver at %s\n", s.Address)
	return nil
}

// Stop cleans up all resources
func (s *Syslog) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Closer != nil {
		s.Close()
	}
	s.wg.Wait()

	log.Printf("I! Stopped syslog receiver at %s\n", s.Address)
}

func (s *Syslog) listenPacket(acc telegraf.Accumulator) {
	defer s.wg.Done()
	b := make([]byte, ipMaxPacketSize)
	for {
		n, _, err := s.udpListener.ReadFrom(b)
		if err != nil {
			if !strings.HasSuffix(err.Error(), ": use of closed network connection") {
				log.Println(err)
				acc.AddError(err)
			}
			break
		}

		// if s.ReadTimeout != nil && s.ReadTimeout.Duration > 0 {
		// 	s.udpListener.SetReadDeadline(time.Now().Add(s.ReadTimeout.Duration))
		// }

		p := rfc5424.NewParser()
		mex, err := p.Parse(b[:n], &s.BestEffort)
		if mex != nil {
			acc.AddFields("syslog", fields(mex), tags(mex), s.now())
		}
		if err != nil {
			acc.AddError(err)
		}
	}
}

func (s *Syslog) listenStream(acc telegraf.Accumulator) {
	defer s.wg.Done()

	s.connections = map[string]net.Conn{}

	for {
		conn, err := s.tcpListener.Accept()
		if err != nil {
			if !strings.HasSuffix(err.Error(), ": use of closed network connection") {
				log.Println(err)
				acc.AddError(err)
			}
			break
		}
		var tcpConn, _ = conn.(*net.TCPConn)
		if s.tlsConfig != nil {
			conn = tls.Server(conn, s.tlsConfig)
		}

		s.connectionsMu.Lock()
		if s.MaxConnections > 0 && len(s.connections) >= s.MaxConnections {
			s.connectionsMu.Unlock()
			conn.Close()
			continue
		}
		s.connections[conn.RemoteAddr().String()] = conn
		s.connectionsMu.Unlock()

		if err := s.setKeepAlive(tcpConn); err != nil {
			acc.AddError(fmt.Errorf("unable to configure keep alive (%s): %s", s.Address, err))
		}

		go s.handle(conn, acc)
	}

	s.connectionsMu.Lock()
	for _, c := range s.connections {
		c.Close()
	}
	s.connectionsMu.Unlock()
}

func (s *Syslog) removeConnection(c net.Conn) {
	s.connectionsMu.Lock()
	delete(s.connections, c.RemoteAddr().String())
	s.connectionsMu.Unlock()
}

func (s *Syslog) handle(conn net.Conn, acc telegraf.Accumulator) {
	defer s.removeConnection(conn)
	defer conn.Close()

	if s.ReadTimeout != nil && s.ReadTimeout.Duration > 0 {
		conn.SetReadDeadline(time.Now().Add(s.ReadTimeout.Duration))
	}

	var p *rfc5425.Parser
	if s.BestEffort {
		p = rfc5425.NewParser(conn, rfc5425.WithBestEffort())
	} else {
		p = rfc5425.NewParser(conn)
	}

	p.ParseExecuting(func(r *rfc5425.Result) {
		s.store(*r, acc)
	})
}

func (s *Syslog) setKeepAlive(c *net.TCPConn) error {
	if s.KeepAlivePeriod == nil {
		return nil
	}

	if s.KeepAlivePeriod.Duration == 0 {
		return c.SetKeepAlive(false)
	}
	if err := c.SetKeepAlive(true); err != nil {
		return err
	}
	return c.SetKeepAlivePeriod(s.KeepAlivePeriod.Duration)
}

func (s *Syslog) store(res rfc5425.Result, acc telegraf.Accumulator) {
	if res.Error != nil {
		acc.AddError(res.Error)
	}
	if res.MessageError != nil {
		acc.AddError(res.MessageError)
	}
	if res.Message != nil {
		acc.AddFields("syslog", fields(res.Message), tags(res.Message), s.now())
	}
}

func tags(msg *rfc5424.SyslogMessage) map[string]string {
	ts := map[string]string{}
	if lvl := msg.SeverityLevel(); lvl != nil {
		ts["severity"] = strconv.Itoa(int(*msg.Severity()))
		ts["severity_level"] = *lvl
	}

	if f := msg.FacilityMessage(); f != nil {
		ts["facility"] = strconv.Itoa(int(*msg.Facility()))
		ts["facility_message"] = *f
	}

	if msg.Hostname() != nil {
		ts["hostname"] = *msg.Hostname()
	}

	if msg.Appname() != nil {
		ts["appname"] = *msg.Appname()
	}

	return ts
}

func fields(msg *rfc5424.SyslogMessage) map[string]interface{} {
	flds := map[string]interface{}{
		"version": msg.Version(),
	}

	if msg.Timestamp() != nil {
		flds["timestamp"] = *msg.Timestamp()
	}

	if msg.ProcID() != nil {
		flds["procid"] = *msg.ProcID()
	}

	if msg.MsgID() != nil {
		flds["msgid"] = *msg.MsgID()
	}

	if msg.Message() != nil {
		flds["message"] = *msg.Message()
	}

	if msg.StructuredData() != nil {
		for sdid, sdparams := range *msg.StructuredData() {
			if len(sdparams) == 0 {
				// When SD-ID does not have params we indicate its presence with a bool
				flds[sdid] = true
				continue
			}
			for name, value := range sdparams {
				// Using whitespace as separator since it is not allowed by the grammar within SDID
				flds[sdid+" "+name] = value
			}
		}
	}

	return flds
}

type unixCloser struct {
	path   string
	closer io.Closer
}

func (uc unixCloser) Close() error {
	err := uc.closer.Close()
	os.Remove(uc.path) // ignore error
	return err
}

func init() {
	receiver := &Syslog{
		Address: ":6514",
		now:     time.Now,
		ReadTimeout: &internal.Duration{
			Duration: defaultReadTimeout,
		},
	}

	inputs.Add("syslog", func() telegraf.Input { return receiver })
}
