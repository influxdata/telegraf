package socket_listener

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	tlsint "github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
)

type setReadBufferer interface {
	SetReadBuffer(bytes int) error
}

type streamSocketListener struct {
	net.Listener
	*SocketListener

	sockType string

	connections    map[string]net.Conn
	connectionsMtx sync.Mutex
}

func (ssl *streamSocketListener) listen() {
	ssl.connections = map[string]net.Conn{}

	for {
		c, err := ssl.Accept()
		if err != nil {
			if !strings.HasSuffix(err.Error(), ": use of closed network connection") {
				ssl.AddError(err)
			}
			break
		}

		if ssl.ReadBufferSize.Size > 0 {
			if srb, ok := c.(setReadBufferer); ok {
				srb.SetReadBuffer(int(ssl.ReadBufferSize.Size))
			} else {
				log.Printf("W! Unable to set read buffer on a %s socket", ssl.sockType)
			}
		}

		ssl.connectionsMtx.Lock()
		if ssl.MaxConnections > 0 && len(ssl.connections) >= ssl.MaxConnections {
			ssl.connectionsMtx.Unlock()
			c.Close()
			continue
		}
		ssl.connections[c.RemoteAddr().String()] = c
		ssl.connectionsMtx.Unlock()

		if err := ssl.setKeepAlive(c); err != nil {
			ssl.AddError(fmt.Errorf("unable to configure keep alive (%s): %s", ssl.ServiceAddress, err))
		}

		go ssl.read(c)
	}

	ssl.connectionsMtx.Lock()
	for _, c := range ssl.connections {
		c.Close()
	}
	ssl.connectionsMtx.Unlock()
}

func (ssl *streamSocketListener) setKeepAlive(c net.Conn) error {
	if ssl.KeepAlivePeriod == nil {
		return nil
	}
	tcpc, ok := c.(*net.TCPConn)
	if !ok {
		return fmt.Errorf("cannot set keep alive on a %s socket", strings.SplitN(ssl.ServiceAddress, "://", 2)[0])
	}
	if ssl.KeepAlivePeriod.Duration == 0 {
		return tcpc.SetKeepAlive(false)
	}
	if err := tcpc.SetKeepAlive(true); err != nil {
		return err
	}
	return tcpc.SetKeepAlivePeriod(ssl.KeepAlivePeriod.Duration)
}

func (ssl *streamSocketListener) removeConnection(c net.Conn) {
	ssl.connectionsMtx.Lock()
	delete(ssl.connections, c.RemoteAddr().String())
	ssl.connectionsMtx.Unlock()
}

func (ssl *streamSocketListener) read(c net.Conn) {
	defer ssl.removeConnection(c)
	defer c.Close()

	scnr := bufio.NewScanner(c)
	for {
		if ssl.ReadTimeout != nil && ssl.ReadTimeout.Duration > 0 {
			c.SetReadDeadline(time.Now().Add(ssl.ReadTimeout.Duration))
		}
		if !scnr.Scan() {
			break
		}
		metrics, err := ssl.Parse(scnr.Bytes())
		if err != nil {
			ssl.AddError(fmt.Errorf("unable to parse incoming line: %s", err))
			//TODO rate limit
			continue
		}
		for _, m := range metrics {
			ssl.AddMetric(m)
		}
	}

	if err := scnr.Err(); err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			log.Printf("D! Timeout in plugin [input.socket_listener]: %s", err)
		} else if netErr != nil && !strings.HasSuffix(err.Error(), ": use of closed network connection") {
			ssl.AddError(err)
		}
	}
}

type packetSocketListener struct {
	net.PacketConn
	*SocketListener
}

func (psl *packetSocketListener) listen() {
	buf := make([]byte, 64*1024) // 64kb - maximum size of IP packet
	for {
		n, _, err := psl.ReadFrom(buf)
		if err != nil {
			if !strings.HasSuffix(err.Error(), ": use of closed network connection") {
				psl.AddError(err)
			}
			break
		}

		metrics, err := psl.Parse(buf[:n])
		if err != nil {
			psl.AddError(fmt.Errorf("unable to parse incoming packet: %s", err))
			//TODO rate limit
			continue
		}
		for _, m := range metrics {
			psl.AddMetric(m)
		}
	}
}

type SocketListener struct {
	ServiceAddress  string             `toml:"service_address"`
	MaxConnections  int                `toml:"max_connections"`
	ReadBufferSize  internal.Size      `toml:"read_buffer_size"`
	ReadTimeout     *internal.Duration `toml:"read_timeout"`
	KeepAlivePeriod *internal.Duration `toml:"keep_alive_period"`
	tlsint.ServerConfig

	parsers.Parser
	telegraf.Accumulator
	io.Closer
}

func (sl *SocketListener) Description() string {
	return "Generic socket listener capable of handling multiple socket types."
}

func (sl *SocketListener) SampleConfig() string {
	return `
  ## URL to listen on
  # service_address = "tcp://:8094"
  # service_address = "tcp://127.0.0.1:http"
  # service_address = "tcp4://:8094"
  # service_address = "tcp6://:8094"
  # service_address = "tcp6://[2001:db8::1]:8094"
  # service_address = "udp://:8094"
  # service_address = "udp4://:8094"
  # service_address = "udp6://:8094"
  # service_address = "unix:///tmp/telegraf.sock"
  # service_address = "unixgram:///tmp/telegraf.sock"

  ## Maximum number of concurrent connections.
  ## Only applies to stream sockets (e.g. TCP).
  ## 0 (default) is unlimited.
  # max_connections = 1024

  ## Read timeout.
  ## Only applies to stream sockets (e.g. TCP).
  ## 0 (default) is unlimited.
  # read_timeout = "30s"

  ## Optional TLS configuration.
  ## Only applies to stream sockets (e.g. TCP).
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key  = "/etc/telegraf/key.pem"
  ## Enables client authentication if set.
  # tls_allowed_cacerts = ["/etc/telegraf/clientca.pem"]

  ## Maximum socket buffer size (in bytes when no unit specified).
  ## For stream sockets, once the buffer fills up, the sender will start backing up.
  ## For datagram sockets, once the buffer fills up, metrics will start dropping.
  ## Defaults to the OS default.
  # read_buffer_size = "64KiB"

  ## Period between keep alive probes.
  ## Only applies to TCP sockets.
  ## 0 disables keep alive probes.
  ## Defaults to the OS configuration.
  # keep_alive_period = "5m"

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  # data_format = "influx"
`
}

func (sl *SocketListener) Gather(_ telegraf.Accumulator) error {
	return nil
}

func (sl *SocketListener) SetParser(parser parsers.Parser) {
	sl.Parser = parser
}

func (sl *SocketListener) Start(acc telegraf.Accumulator) error {
	sl.Accumulator = acc
	spl := strings.SplitN(sl.ServiceAddress, "://", 2)
	if len(spl) != 2 {
		return fmt.Errorf("invalid service address: %s", sl.ServiceAddress)
	}

	if spl[0] == "unix" || spl[0] == "unixpacket" || spl[0] == "unixgram" {
		// no good way of testing for "file does not exist".
		// Instead just ignore error and blow up when we try to listen, which will
		// indicate "address already in use" if file existed and we couldn't remove.
		os.Remove(spl[1])
	}

	switch spl[0] {
	case "tcp", "tcp4", "tcp6", "unix", "unixpacket":
		var (
			err error
			l   net.Listener
		)

		tlsCfg, err := sl.ServerConfig.TLSConfig()
		if err != nil {
			return nil
		}

		if tlsCfg == nil {
			l, err = net.Listen(spl[0], spl[1])
		} else {
			l, err = tls.Listen(spl[0], spl[1], tlsCfg)
		}
		if err != nil {
			return err
		}

		ssl := &streamSocketListener{
			Listener:       l,
			SocketListener: sl,
			sockType:       spl[0],
		}

		sl.Closer = ssl
		go ssl.listen()
	case "udp", "udp4", "udp6", "ip", "ip4", "ip6", "unixgram":
		pc, err := net.ListenPacket(spl[0], spl[1])
		if err != nil {
			return err
		}

		if sl.ReadBufferSize.Size > 0 {
			if srb, ok := pc.(setReadBufferer); ok {
				srb.SetReadBuffer(int(sl.ReadBufferSize.Size))
			} else {
				log.Printf("W! Unable to set read buffer on a %s socket", spl[0])
			}
		}

		psl := &packetSocketListener{
			PacketConn:     pc,
			SocketListener: sl,
		}

		sl.Closer = psl
		go psl.listen()
	default:
		return fmt.Errorf("unknown protocol '%s' in '%s'", spl[0], sl.ServiceAddress)
	}

	if spl[0] == "unix" || spl[0] == "unixpacket" || spl[0] == "unixgram" {
		sl.Closer = unixCloser{path: spl[1], closer: sl.Closer}
	}

	return nil
}

func (sl *SocketListener) Stop() {
	if sl.Closer != nil {
		sl.Close()
		sl.Closer = nil
	}
}

func newSocketListener() *SocketListener {
	parser, _ := parsers.NewInfluxParser()

	return &SocketListener{
		Parser: parser,
	}
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
	inputs.Add("socket_listener", func() telegraf.Input { return newSocketListener() })
}
