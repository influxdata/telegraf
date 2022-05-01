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
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	tlsint "github.com/influxdata/telegraf/plugins/common/tls"
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

		if ssl.ReadBufferSize > 0 {
			if srb, ok := c.(setReadBufferer); ok {
				if err := srb.SetReadBuffer(int(ssl.ReadBufferSize)); err != nil {
					ssl.Log.Error(err.Error())
					break
				}
			} else {
				ssl.Log.Warnf("Unable to set read buffer on a %s socket", ssl.sockType)
			}
		}

		ssl.connectionsMtx.Lock()
		if ssl.MaxConnections > 0 && len(ssl.connections) >= ssl.MaxConnections {
			ssl.connectionsMtx.Unlock()
			// Ignore the returned error as we cannot do anything about it anyway
			//nolint:errcheck,revive
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
		// Ignore the returned error as we cannot do anything about it anyway
		//nolint:errcheck,revive
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
	if *ssl.KeepAlivePeriod == 0 {
		return tcpc.SetKeepAlive(false)
	}
	if err := tcpc.SetKeepAlive(true); err != nil {
		return err
	}
	return tcpc.SetKeepAlivePeriod(time.Duration(*ssl.KeepAlivePeriod))
}

func (ssl *streamSocketListener) removeConnection(c net.Conn) {
	ssl.connectionsMtx.Lock()
	delete(ssl.connections, c.RemoteAddr().String())
	ssl.connectionsMtx.Unlock()
}

func (ssl *streamSocketListener) read(c net.Conn) {
	defer ssl.removeConnection(c)
	defer c.Close()

	decoder, err := internal.NewStreamContentDecoder(ssl.ContentEncoding, c)
	if err != nil {
		ssl.Log.Error("Read error: %v", err)
		return
	}

	scnr := bufio.NewScanner(decoder)
	for {
		if ssl.ReadTimeout != nil && *ssl.ReadTimeout > 0 {
			if err := c.SetReadDeadline(time.Now().Add(time.Duration(*ssl.ReadTimeout))); err != nil {
				ssl.Log.Error("setting read deadline failed: %v", err)
				return
			}
		}
		if !scnr.Scan() {
			break
		}

		body := scnr.Bytes()

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
	decoder internal.ContentDecoder
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
	ServiceAddress  string           `toml:"service_address"`
	MaxConnections  int              `toml:"max_connections"`
	ReadBufferSize  config.Size      `toml:"read_buffer_size"`
	ReadTimeout     *config.Duration `toml:"read_timeout"`
	KeepAlivePeriod *config.Duration `toml:"keep_alive_period"`
	SocketMode      string           `toml:"socket_mode"`
	ContentEncoding string           `toml:"content_encoding"`
	tlsint.ServerConfig

	wg sync.WaitGroup

	Log telegraf.Logger

	parsers.Parser
	telegraf.Accumulator
	io.Closer
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

	if protocol == "unix" || protocol == "unixpacket" || protocol == "unixgram" {
		// no good way of testing for "file does not exist".
		// Instead just ignore error and blow up when we try to listen, which will
		// indicate "address already in use" if file existed and we couldn't remove.
		//nolint:errcheck,revive
		os.Remove(addr)
	}

	switch protocol {
	case "tcp", "tcp4", "tcp6", "unix", "unixpacket":
		tlsCfg, err := sl.ServerConfig.TLSConfig()
		if err != nil {
			return err
		}

		var l net.Listener
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

			if err := os.Chmod(spl[1], os.FileMode(uint32(i))); err != nil {
				return err
			}
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
		decoder, err := internal.NewContentDecoder(sl.ContentEncoding)
		if err != nil {
			return err
		}

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

			if err := os.Chmod(spl[1], os.FileMode(uint32(i))); err != nil {
				return err
			}
		}

		if sl.ReadBufferSize > 0 {
			if srb, ok := pc.(setReadBufferer); ok {
				if err := srb.SetReadBuffer(int(sl.ReadBufferSize)); err != nil {
					sl.Log.Warnf("Setting read buffer on a %s socket failed: %v", protocol, err)
				}
			} else {
				sl.Log.Warnf("Unable to set read buffer on a %s socket", protocol)
			}
		}

		sl.Log.Infof("Listening on %s://%s", protocol, pc.LocalAddr())

		psl := &packetSocketListener{
			PacketConn:     pc,
			SocketListener: sl,
			decoder:        decoder,
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
		// Ignore the returned error as we cannot do anything about it anyway
		//nolint:errcheck,revive
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
	// Ignore the error if e.g. the file does not exist
	//nolint:errcheck,revive
	os.Remove(uc.path)
	return err
}

func init() {
	inputs.Add("socket_listener", func() telegraf.Input { return newSocketListener() })
}
