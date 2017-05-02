package udp_listener

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/selfstat"
)

// UdpListener main struct for the collector
type UdpListener struct {
	ServiceAddress string

	// UDPBufferSize should only be set if you want/need the telegraf UDP socket to
	// differ from the system setting. In cases where you set the rmem_default to a lower
	// value at the host level, but need a larger buffer for UDP bursty traffic, this
	// setting enables you to configure that value ONLY for telegraf UDP sockets on this listener
	// Set this to 0 (or comment out) to take system default
	//
	// NOTE: You should ensure that your rmem_max is >= to this setting to work properly!
	// (e.g. sysctl -w net.core.rmem_max=N)
	UDPBufferSize          int `toml:"udp_buffer_size"`
	AllowedPendingMessages int

	// UDPPacketSize is deprecated, it's only here for legacy support
	// we now always create 1 max size buffer and then copy only what we need
	// into the in channel
	// see https://github.com/influxdata/telegraf/pull/992
	UDPPacketSize int `toml:"udp_packet_size"`

	sync.Mutex
	wg sync.WaitGroup

	in   chan []byte
	done chan struct{}
	// drops tracks the number of dropped metrics.
	drops int
	// malformed tracks the number of malformed packets
	malformed int

	parser parsers.Parser

	// Keep the accumulator in this struct
	acc telegraf.Accumulator

	listener *net.UDPConn

	PacketsRecv selfstat.Stat
	BytesRecv   selfstat.Stat
}

// UDP_MAX_PACKET_SIZE is packet limit, see
// https://en.wikipedia.org/wiki/User_Datagram_Protocol#Packet_structure
const UDP_MAX_PACKET_SIZE int = 64 * 1024

var dropwarn = "E! Error: udp_listener message queue full. " +
	"We have dropped %d messages so far. " +
	"You may want to increase allowed_pending_messages in the config\n"

var malformedwarn = "E! udp_listener has received %d malformed packets" +
	" thus far."

const sampleConfig = `
  # DEPRECATED: the TCP listener plugin has been deprecated in favor of the
  # socket_listener plugin
  # see https://github.com/influxdata/telegraf/tree/master/plugins/inputs/socket_listener
`

func (u *UdpListener) SampleConfig() string {
	return sampleConfig
}

func (u *UdpListener) Description() string {
	return "Generic UDP listener"
}

// All the work is done in the Start() function, so this is just a dummy
// function.
func (u *UdpListener) Gather(_ telegraf.Accumulator) error {
	return nil
}

func (u *UdpListener) SetParser(parser parsers.Parser) {
	u.parser = parser
}

func (u *UdpListener) Start(acc telegraf.Accumulator) error {
	u.Lock()
	defer u.Unlock()

	log.Println("W! DEPRECATED: the UDP listener plugin has been deprecated " +
		"in favor of the socket_listener plugin " +
		"(https://github.com/influxdata/telegraf/tree/master/plugins/inputs/socket_listener)")

	tags := map[string]string{
		"address": u.ServiceAddress,
	}
	u.PacketsRecv = selfstat.Register("udp_listener", "packets_received", tags)
	u.BytesRecv = selfstat.Register("udp_listener", "bytes_received", tags)

	u.acc = acc
	u.in = make(chan []byte, u.AllowedPendingMessages)
	u.done = make(chan struct{})

	u.udpListen()

	u.wg.Add(1)
	go u.udpParser()

	log.Printf("I! Started UDP listener service on %s (ReadBuffer: %d)\n", u.ServiceAddress, u.UDPBufferSize)
	return nil
}

func (u *UdpListener) Stop() {
	u.Lock()
	defer u.Unlock()
	close(u.done)
	u.wg.Wait()
	u.listener.Close()
	close(u.in)
	log.Println("I! Stopped UDP listener service on ", u.ServiceAddress)
}

func (u *UdpListener) udpListen() error {
	var err error

	address, _ := net.ResolveUDPAddr("udp", u.ServiceAddress)
	u.listener, err = net.ListenUDP("udp", address)

	if err != nil {
		return fmt.Errorf("E! Error: ListenUDP - %s", err)
	}

	log.Println("I! UDP server listening on: ", u.listener.LocalAddr().String())

	if u.UDPBufferSize > 0 {
		err = u.listener.SetReadBuffer(u.UDPBufferSize) // if we want to move away from OS default
		if err != nil {
			return fmt.Errorf("E! Failed to set UDP read buffer to %d: %s", u.UDPBufferSize, err)
		}
	}

	u.wg.Add(1)
	go u.udpListenLoop()
	return nil
}

func (u *UdpListener) udpListenLoop() {
	defer u.wg.Done()

	buf := make([]byte, UDP_MAX_PACKET_SIZE)
	for {
		select {
		case <-u.done:
			return
		default:
			u.listener.SetReadDeadline(time.Now().Add(time.Second))

			n, _, err := u.listener.ReadFromUDP(buf)
			if err != nil {
				if err, ok := err.(net.Error); ok && err.Timeout() {
				} else {
					log.Printf("E! Error: %s\n", err.Error())
				}
				continue
			}
			u.BytesRecv.Incr(int64(n))
			u.PacketsRecv.Incr(1)
			bufCopy := make([]byte, n)
			copy(bufCopy, buf[:n])

			select {
			case u.in <- bufCopy:
			default:
				u.drops++
				if u.drops == 1 || u.drops%u.AllowedPendingMessages == 0 {
					log.Printf(dropwarn, u.drops)
				}
			}
		}
	}
}

func (u *UdpListener) udpParser() error {
	defer u.wg.Done()

	var packet []byte
	var metrics []telegraf.Metric
	var err error
	for {
		select {
		case <-u.done:
			if len(u.in) == 0 {
				return nil
			}
		case packet = <-u.in:
			metrics, err = u.parser.Parse(packet)
			if err == nil {
				for _, m := range metrics {
					u.acc.AddFields(m.Name(), m.Fields(), m.Tags(), m.Time())
				}
			} else {
				u.malformed++
				if u.malformed == 1 || u.malformed%1000 == 0 {
					log.Printf(malformedwarn, u.malformed)
				}
			}
		}
	}
}

func init() {
	inputs.Add("udp_listener", func() telegraf.Input {
		return &UdpListener{
			ServiceAddress:         ":8092",
			AllowedPendingMessages: 10000,
		}
	})
}
