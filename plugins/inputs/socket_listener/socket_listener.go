package socket_listener

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
)

type setReadBufferer interface {
	SetReadBuffer(bytes int) error
}

type streamSocketListener struct {
	net.Listener
	*SocketListener

	connections    map[string]net.Conn
	connectionsMtx sync.Mutex
}

func (ssl *streamSocketListener) listen() {
	ssl.connections = map[string]net.Conn{}

	for {
		c, err := ssl.Accept()
		if err != nil {
			ssl.AddError(err)
			break
		}

		ssl.connectionsMtx.Lock()
		if ssl.MaxConnections > 0 && len(ssl.connections) >= ssl.MaxConnections {
			ssl.connectionsMtx.Unlock()
			c.Close()
			continue
		}
		ssl.connections[c.RemoteAddr().String()] = c
		ssl.connectionsMtx.Unlock()
		go ssl.read(c)
	}

	ssl.connectionsMtx.Lock()
	for _, c := range ssl.connections {
		c.Close()
	}
	ssl.connectionsMtx.Unlock()
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
	for scnr.Scan() {
		metrics, err := ssl.Parse(scnr.Bytes())
		if err != nil {
			ssl.AddError(fmt.Errorf("unable to parse incoming line"))
			//TODO rate limit
			continue
		}
		for _, m := range metrics {
			ssl.AddFields(m.Name(), m.Fields(), m.Tags(), m.Time())
		}
	}

	if err := scnr.Err(); err != nil {
		ssl.AddError(err)
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
			psl.AddError(err)
			break
		}

		metrics, err := psl.Parse(buf[:n])
		if err != nil {
			psl.AddError(fmt.Errorf("unable to parse incoming packet"))
			//TODO rate limit
			continue
		}
		for _, m := range metrics {
			psl.AddFields(m.Name(), m.Fields(), m.Tags(), m.Time())
		}
	}
}

type SocketListener struct {
	ServiceAddress string
	MaxConnections int
	ReadBufferSize int

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

  ## Maximum socket buffer size in bytes.
  ## For stream sockets, once the buffer fills up, the sender will start backing up.
  ## For datagram sockets, once the buffer fills up, metrics will start dropping.
  ## Defaults to the OS default.
  # read_buffer_size = 65535

  ## Data format to consume.
  ## Each data format has it's own unique set of configuration options, read
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

	switch spl[0] {
	case "tcp", "tcp4", "tcp6", "unix", "unixpacket":
		l, err := net.Listen(spl[0], spl[1])
		if err != nil {
			return err
		}

		if sl.ReadBufferSize > 0 {
			if srb, ok := l.(setReadBufferer); ok {
				srb.SetReadBuffer(sl.ReadBufferSize)
			} else {
				log.Printf("W! Unable to set read buffer on a %s socket", spl[0])
			}
		}

		ssl := &streamSocketListener{
			Listener:       l,
			SocketListener: sl,
		}

		sl.Closer = ssl
		go ssl.listen()
	case "udp", "udp4", "udp6", "ip", "ip4", "ip6", "unixgram":
		pc, err := net.ListenPacket(spl[0], spl[1])
		if err != nil {
			return err
		}

		if sl.ReadBufferSize > 0 {
			if srb, ok := pc.(setReadBufferer); ok {
				srb.SetReadBuffer(sl.ReadBufferSize)
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

func init() {
	inputs.Add("socket_listener", func() telegraf.Input { return newSocketListener() })
}
