package multicast_listener

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
)

type setReadBufferer interface {
	SetReadBuffer(bytes int) error
}

type v4PacketSocketListener struct {
	*ipv4.PacketConn
	*SocketListener
}

func (psl *v4PacketSocketListener) listen() {
	buf := make([]byte, 64*1024) // 64kb - maximum size of IP packet
	for {
		n, _, _, err := psl.ReadFrom(buf)
		if err != nil {
			if !strings.HasSuffix(err.Error(), ": use of closed network connection") {
				psl.AddError(err)
			}
			break
		}

		metrics, err := psl.Parse(buf[:n])
		if err != nil {
			psl.AddError(fmt.Errorf("unable to parse incoming packet: %s", err))
			// TODO rate limit
			continue
		}
		for _, m := range metrics {
			psl.AddMetric(m)
		}
	}
}

type v6PacketSocketListener struct {
	*ipv6.PacketConn
	*SocketListener
}

func (psl *v6PacketSocketListener) listen() {
	buf := make([]byte, 64*1024) // 64kb - maximum size of IP packet
	for {
		n, _, _, err := psl.ReadFrom(buf)
		if err != nil {
			if !strings.HasSuffix(err.Error(), ": use of closed network connection") {
				psl.AddError(err)
			}
			break
		}

		metrics, err := psl.Parse(buf[:n])
		if err != nil {
			psl.AddError(fmt.Errorf("unable to parse incoming packet: %s", err))
			// TODO rate limit
			continue
		}
		for _, m := range metrics {
			psl.AddMetric(m)
		}
	}
}

type SourceSpecificGroup struct {
	GroupAddress  string `toml:"group_address"`
	SourceAddress string `toml:"source_address"`
}

type SocketListener struct {
	ServiceAddress       string                `toml:"service_address"`
	ReadBufferSize       internal.Size         `toml:"read_buffer_size"`
	ReadTimeout          *internal.Duration    `toml:"read_timeout"`
	Groups               []string              `toml:"groups"`
	SourceSpecificGroups []SourceSpecificGroup `toml:"source_specific_groups"`

	parsers.Parser
	telegraf.Accumulator
	io.Closer
}

func (sl *SocketListener) Description() string {
	return "Multicast socket listener."
}

func (sl *SocketListener) SampleConfig() string {
	return `
  ## URL to listen on
  # service_address = "udp4://:8094"
  # service_address = "udp6://:8094"

  ## Multicast groups to join
  # [[inputs.multicast_listener.groups]]
  #   address = "239.200.4.108"

  ## Multicast source-specific groups to join
  # [[inputs.multicast_listener.source_specific_groups]]
  #   group_address = "239.200.4.108"
  #   source_address = "192.168.1.10"

  ## Read timeout.
  ## 0 (default) is unlimited.
  # read_timeout = "30s"

  ## Maximum socket buffer size (in bytes when no unit specified).
  ## For stream sockets, once the buffer fills up, the sender will start backing up.
  ## For datagram sockets, once the buffer fills up, metrics will start dropping.
  ## Defaults to the OS default.
  # read_buffer_size = "64KiB"

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

	protocol := spl[0]
	addr := spl[1]

	switch protocol {
	case "udp4", "udp6":
		pc, err := net.ListenPacket(protocol, addr)
		if err != nil {
			return err
		}

		if sl.ReadBufferSize.Size > 0 {
			if srb, ok := pc.(setReadBufferer); ok {
				srb.SetReadBuffer(int(sl.ReadBufferSize.Size))
			} else {
				log.Printf("W! [inputs.multicast_listener] Unable to set read buffer on a %s socket", protocol)
			}
		}

		log.Printf("I! [inputs.multicast_listener] Listening on %s://%s", protocol, pc.LocalAddr())

		switch protocol {
		case "udp4":
			psl := &v4PacketSocketListener{
				PacketConn:     ipv4.NewPacketConn(pc),
				SocketListener: sl,
			}

			sl.Closer = psl
			go psl.listen()
		case "udp6":
			psl := &v6PacketSocketListener{
				PacketConn:     ipv6.NewPacketConn(pc),
				SocketListener: sl,
			}

			sl.Closer = psl
			go psl.listen()
		}
	default:
		return fmt.Errorf("unknown protocol '%s' in '%s'", protocol, sl.ServiceAddress)
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
	inputs.Add("multicast_listener", func() telegraf.Input { return newSocketListener() })
}
