package sflow_a10

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/selfstat"
)

const sampleConfig = `
  ## Address to listen for sFlow packets.
  ##   example: service_address = "udp://:6343"
  ##            service_address = "udp4://:6343"
  ##            service_address = "udp6://:6343"
  service_address = "udp://:6343"

  ## Set the size of the operating system's receive buffer.
  ##   example: read_buffer_size = "64KiB"
  # read_buffer_size = ""

  # XML file containing counter definitions, according to A10 specification
  a10_definitions_file = "/path/to/xml_file.xml"
`

const (
	parserGoRoutines           = 8
	defaultAllowPendingMessage = 100000
	// UDP_MAX_PACKET_SIZE is the UDP packet limit, see
	// https://en.wikipedia.org/wiki/User_Datagram_Protocol#Packet_structure
	UDP_MAX_PACKET_SIZE int = 64 * 1024
)

type SFlow_A10 struct {
	ServiceAddress     string        `toml:"service_address"`
	ReadBufferSize     config.Size   `toml:"read_buffer_size"`
	A10DefinitionsFile string        `toml:"a10_definitions_file"`

	sync.Mutex

	Log telegraf.Logger `toml:"-"`

	addr    net.Addr
	decoder *PacketDecoder
	cancel  context.CancelFunc
	wg      sync.WaitGroup

	acc telegraf.Accumulator

	// Channel for all incoming sflow packets
	in   chan input
	done chan struct{}

	UDPlistener *net.UDPConn

	// Number of messages allowed to queue up in between calls to Gather. If this
	// fills up, packets will get dropped until the next Gather interval is ran.
	AllowedPendingMessages int

	SflowUDPPacketsRecv selfstat.Stat
	SflowUDPPacketsDrop selfstat.Stat
	SflowUDPBytesRecv   selfstat.Stat
	ParseTimeNS         selfstat.Stat

	// A pool of byte slices to handle parsing
	bufPool sync.Pool

	// drops tracks the number of dropped metrics.
	drops int
}

type input struct {
	*bytes.Buffer
	time.Time
	Addr string
}

// Description answers a description of this input plugin
func (s *SFlow_A10) Description() string {
	return "SFlow_A10 V5 Protocol Listener"
}

// SampleConfig answers a sample configuration
func (s *SFlow_A10) SampleConfig() string {
	return sampleConfig
}

func (s *SFlow_A10) Init() error {
	if s.A10DefinitionsFile == "" {
		return errors.New("XML DefinitionsFile cannot be empty")
	}
	data, err := ioutil.ReadFile(s.A10DefinitionsFile)
	if err != nil {
		return err
	}

	return s.initInternal(data)
}

func (s *SFlow_A10) initInternal(xmlData []byte) error {
	s.decoder = NewDecoder()
	s.decoder.Log = s.Log
	counterBlocks, err := s.readA10XMLData(xmlData)
	if err != nil {
		return err
	}
	s.decoder.CounterBlocks = counterBlocks

	return nil
}

// Start starts this sFlow_A10 listener listening on the configured network for sFlow packets
func (s *SFlow_A10) Start(acc telegraf.Accumulator) error {
	s.decoder.OnPacket(func(p *V5Format) {
		metrics, err := makeMetricsForCounters(p, s.decoder)
		if err != nil {
			s.Log.Errorf("Failed to make metric from packet: %s", err)
			return
		}
		for _, m := range metrics {
			acc.AddMetric(m)
		}
	})

	s.acc = acc

	s.Lock()
	defer s.Unlock()

	tags := map[string]string{
		"address": s.ServiceAddress,
	}
	s.SflowUDPPacketsRecv = selfstat.Register("sflow_a10", "sflow_udp_packets_received", tags)
	s.SflowUDPPacketsDrop = selfstat.Register("sflow_a10", "sflow_udp_packets_dropped", tags)
	s.SflowUDPBytesRecv = selfstat.Register("sflow_a10", "sflow_udp_bytes_received", tags)
	s.ParseTimeNS = selfstat.Register("sflow_a10", "sflow_parse_time_ns", tags)

	s.in = make(chan input, s.AllowedPendingMessages)
	s.done = make(chan struct{})
	s.bufPool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}

	u, err := url.Parse(s.ServiceAddress)
	if err != nil {
		return err
	}

	conn, err := listenUDP(u.Scheme, u.Host)
	if err != nil {
		return err
	}
	s.addr = conn.LocalAddr()
	s.UDPlistener = conn

	if s.ReadBufferSize > 0 {
		if err := conn.SetReadBuffer(int(s.ReadBufferSize)); err != nil {
			return err
		}
	}

	s.Log.Infof("Listening on %s://%s", s.addr.Network(), s.addr.String())

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.udpListen(conn)
	}()

	for i := 1; i <= parserGoRoutines; i++ {
		// Start the packet parser
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.parse()
		}()
	}

	s.Log.Info("Started the sflow_a10 service on ", s.ServiceAddress)

	return nil
}

// Gather is a NOOP for sFlow as it receives, asynchronously, sFlow network packets
func (s *SFlow_A10) Gather(_ telegraf.Accumulator) error {
	return nil
}

func (s *SFlow_A10) Stop() {
	s.Log.Infof("Stopping the sflow_a10 service")
	close(s.done)
	s.Lock()
	if s.UDPlistener != nil {
		s.UDPlistener.Close()
	}
	s.Unlock()
	s.wg.Wait()

	s.Lock()
	close(s.in)
	s.Log.Infof("Stopped listener service on %q", s.ServiceAddress)
	s.Unlock()
}

func (s *SFlow_A10) Address() net.Addr {
	return s.addr
}

func (s *SFlow_A10) udpListen(conn *net.UDPConn) {
	buf := make([]byte, UDP_MAX_PACKET_SIZE)
	for {
		select {
		case <-s.done:
			return
		default:
			// accept connection
			n, addr, err := conn.ReadFromUDP(buf)
			if err != nil {
				if !strings.HasSuffix(err.Error(), ": use of closed network connection") {
					s.acc.AddError(err)
				}
				break
			}
			s.SflowUDPPacketsRecv.Incr(1)
			s.SflowUDPBytesRecv.Incr(int64(n))
			b := s.bufPool.Get().(*bytes.Buffer)
			b.Reset()
			b.Write(buf[:n])
			select {
			case s.in <- input{
				Buffer: b,
				Time:   time.Now(),
				Addr:   addr.IP.String()}:
			default:
				s.SflowUDPPacketsDrop.Incr(1)
				s.drops++
				if s.drops == 1 || s.AllowedPendingMessages == 0 || s.drops%s.AllowedPendingMessages == 0 {
					s.Log.Errorf("Sflow message queue full. "+
						"We have dropped %d messages so far. "+
						"You may want to increase allowed_pending_messages in the config", s.drops)
				}
			}
		}

	}
}

func (s *SFlow_A10) parse() error {
	for {
		select {
		case <-s.done:
			return nil
		case in := <-s.in:
			start := time.Now()
			if err := s.decoder.Decode(in.Buffer); err != nil {
				s.acc.AddError(fmt.Errorf("unable to parse incoming packet: %s", err))
			}
			elapsed := time.Since(start)
			s.ParseTimeNS.Set(elapsed.Nanoseconds())
		}
	}
}

func listenUDP(network string, address string) (*net.UDPConn, error) {
	switch network {
	case "udp", "udp4", "udp6":
		addr, err := net.ResolveUDPAddr(network, address)
		if err != nil {
			return nil, err
		}
		return net.ListenUDP(network, addr)
	default:
		return nil, fmt.Errorf("unsupported network type: %s", network)
	}
}

// init registers this SFlow_A10 input plug in with the Telegraf framework
func init() {
	inputs.Add("sflow_a10", func() telegraf.Input {
		return &SFlow_A10{
			AllowedPendingMessages: defaultAllowPendingMessage,
		}
	})
}
