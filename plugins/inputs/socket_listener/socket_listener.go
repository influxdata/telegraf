package socket_listener

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
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

	wg := sync.WaitGroup{}

	for {
		c, err := ssl.Accept()
		if err != nil {
			if !strings.HasSuffix(err.Error(), ": use of closed network connection") {
				ssl.Log.Error(err.Error())
			}
			break
		}

		if ssl.ReadBufferSize.Size > 0 {
			if srb, ok := c.(setReadBufferer); ok {
				srb.SetReadBuffer(int(ssl.ReadBufferSize.Size))
			} else {
				ssl.Log.Warnf("Unable to set read buffer on a %s socket", ssl.sockType)
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
			ssl.Log.Errorf("Unable to configure keep alive %q: %s", ssl.ServiceAddress, err.Error())
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			ssl.read(c)
		}()
	}

	ssl.connectionsMtx.Lock()
	for _, c := range ssl.connections {
		c.Close()
	}
	ssl.connectionsMtx.Unlock()

	wg.Wait()
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

		body, err := ssl.decoder.Decode(scnr.Bytes())
		if err != nil {
			ssl.Log.Errorf("Unable to decode incoming line: %s", err.Error())
			continue
		}

		metrics, err := ssl.Parse(body)
		if err != nil {
			ssl.Log.Errorf("Unable to parse incoming line: %s", err.Error())
			// TODO rate limit
			continue
		}
		for _, m := range metrics {
			ssl.AddMetric(m)
		}
	}

	if err := scnr.Err(); err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			ssl.Log.Debugf("Timeout in plugin: %s", err.Error())
		} else if netErr != nil && !strings.HasSuffix(err.Error(), ": use of closed network connection") {
			ssl.Log.Error(err.Error())
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
				psl.Log.Error(err.Error())
			}
			break
		}

		body, err := psl.decoder.Decode(buf[:n])
		if err != nil {
			psl.Log.Errorf("Unable to decode incoming packet: %s", err.Error())
		}

		metrics, err := psl.Parse(body)
		if err != nil {
			psl.Log.Errorf("Unable to parse incoming packet: %s", err.Error())
			// TODO rate limit
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
	SocketMode      string             `toml:"socket_mode"`
	ContentEncoding string             `toml:"content_encoding"`
	tlsint.ServerConfig

	wg sync.WaitGroup

	Log telegraf.Logger

	parsers.Parser
	telegraf.Accumulator
	io.Closer
	decoder internal.ContentDecoder
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

  ## Change the file mode bits on unix sockets.  These permissions may not be
  ## respected by some platforms, to safely restrict write permissions it is best
  ## to place the socket into a directory that has previously been created
  ## with the desired permissions.
  ##   ex: socket_mode = "777"
  # socket_mode = ""

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

  ## Content encoding for message payloads, can be set to "gzip" to or
  ## "identity" to apply no encoding.
  # content_encoding = "identity"
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

	protocol := spl[0]
	addr := spl[1]

	var err error
	sl.decoder, err = internal.NewContentDecoder(sl.ContentEncoding)
	if err != nil {
		return err
	}

	if protocol == "unix" || protocol == "unixpacket" || protocol == "unixgram" {
		// no good way of testing for "file does not exist".
		// Instead just ignore error and blow up when we try to listen, which will
		// indicate "address already in use" if file existed and we couldn't remove.
		os.Remove(addr)
	}

	switch protocol {
	case "tcp", "tcp4", "tcp6", "unix", "unixpacket":
		var (
			err error
			l   net.Listener
		)

		tlsCfg, err := sl.ServerConfig.TLSConfig()
		if err != nil {
			return err
		}

		if tlsCfg == nil {
			l, err = net.Listen(protocol, addr)
		} else {
			l, err = tls.Listen(protocol, addr, tlsCfg)
		}
		if err != nil {
			return err
		}

		sl.Log.Infof("Listening on %s://%s", protocol, l.Addr())

		// Set permissions on socket
		if (spl[0] == "unix" || spl[0] == "unixpacket") && sl.SocketMode != "" {
			// Convert from octal in string to int
			i, err := strconv.ParseUint(sl.SocketMode, 8, 32)
			if err != nil {
				return err
			}

			os.Chmod(spl[1], os.FileMode(uint32(i)))
		}

		ssl := &streamSocketListener{
			Listener:       l,
			SocketListener: sl,
			sockType:       spl[0],
		}

		sl.Closer = ssl
		sl.wg = sync.WaitGroup{}
		sl.wg.Add(1)
		go func() {
			defer sl.wg.Done()
			ssl.listen()
		}()
	case "udp", "udp4", "udp6", "ip", "ip4", "ip6", "unixgram":
		pc, err := udpListen(protocol, addr)
		if err != nil {
			return err
		}

		// Set permissions on socket
		if spl[0] == "unixgram" && sl.SocketMode != "" {
			// Convert from octal in string to int
			i, err := strconv.ParseUint(sl.SocketMode, 8, 32)
			if err != nil {
				return err
			}

			os.Chmod(spl[1], os.FileMode(uint32(i)))
		}

		if sl.ReadBufferSize.Size > 0 {
			if srb, ok := pc.(setReadBufferer); ok {
				srb.SetReadBuffer(int(sl.ReadBufferSize.Size))
			} else {
				sl.Log.Warnf("Unable to set read buffer on a %s socket", protocol)
			}
		}

		sl.Log.Infof("Listening on %s://%s", protocol, pc.LocalAddr())

		psl := &packetSocketListener{
			PacketConn:     pc,
			SocketListener: sl,
		}

		sl.Closer = psl
		sl.wg = sync.WaitGroup{}
		sl.wg.Add(1)
		go func() {
			defer sl.wg.Done()
			psl.listen()
		}()
	default:
		return fmt.Errorf("unknown protocol '%s' in '%s'", protocol, sl.ServiceAddress)
	}

	if protocol == "unix" || protocol == "unixpacket" || protocol == "unixgram" {
		sl.Closer = unixCloser{path: spl[1], closer: sl.Closer}
	}

	return nil
}

func udpListen(network string, address string) (net.PacketConn, error) {
	switch network {
	case "udp", "udp4", "udp6":
		var addr *net.UDPAddr
		var err error
		var ifi *net.Interface
		if spl := strings.SplitN(address, "%", 2); len(spl) == 2 {
			address = spl[0]
			ifi, err = net.InterfaceByName(spl[1])
			if err != nil {
				return nil, err
			}
		}
		addr, err = net.ResolveUDPAddr(network, address)
		if err != nil {
			return nil, err
		}
		if addr.IP.IsMulticast() {
			return net.ListenMulticastUDP(network, ifi, addr)
		}
		return net.ListenUDP(network, addr)
	}
	return net.ListenPacket(network, address)
}

func (sl *SocketListener) Stop() {
	if sl.Closer != nil {
		sl.Close()
		sl.Closer = nil
	}
	sl.wg.Wait()
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
