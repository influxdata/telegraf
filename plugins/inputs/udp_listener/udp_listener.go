package udp_listener

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/selfstat"
)

// UDPListener main struct for the collector
type UDPListener struct {
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

	Log telegraf.Logger
}

// UDPMaxPacketSize is packet limit, see
// https://en.wikipedia.org/wiki/User_Datagram_Protocol#Packet_structure
const UDPMaxPacketSize int = 64 * 1024

var dropwarn = "udp_listener message queue full. " +
	"We have dropped %d messages so far. " +
	"You may want to increase allowed_pending_messages in the config"

var malformedwarn = "udp_listener has received %d malformed packets" +
	" thus far."

// All the work is done in the Start() function, so this is just a dummy
// function.
func (u *UDPListener) Gather(_ telegraf.Accumulator) error {
	return nil
}

func (u *UDPListener) SetParser(parser parsers.Parser) {
	u.parser = parser
}

func (u *UDPListener) Start(acc telegraf.Accumulator) error {
	u.Lock()
	defer u.Unlock()

	u.Log.Warn("DEPRECATED: the UDP listener plugin has been deprecated " +
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

	if err := u.udpListen(); err != nil {
		return err
	}

	u.wg.Add(1)
	go u.udpParser()

	u.Log.Infof("Started service on %q (ReadBuffer: %d)", u.ServiceAddress, u.UDPBufferSize)
	return nil
}

func (u *UDPListener) Stop() {
	u.Lock()
	defer u.Unlock()
	close(u.done)
	u.wg.Wait()
	// Ignore the returned error as we cannot do anything about it anyway
	//nolint:errcheck,revive
	u.listener.Close()
	close(u.in)
	u.Log.Infof("Stopped service on %q", u.ServiceAddress)
}

func (u *UDPListener) udpListen() error {
	var err error

	address, _ := net.ResolveUDPAddr("udp", u.ServiceAddress)
	u.listener, err = net.ListenUDP("udp", address)

	if err != nil {
		return err
	}

	u.Log.Infof("Server listening on %q", u.listener.LocalAddr().String())

	if u.UDPBufferSize > 0 {
		err = u.listener.SetReadBuffer(u.UDPBufferSize) // if we want to move away from OS default
		if err != nil {
			return fmt.Errorf("failed to set UDP read buffer to %d: %s", u.UDPBufferSize, err)
		}
	}

	u.wg.Add(1)
	go u.udpListenLoop()
	return nil
}

func (u *UDPListener) udpListenLoop() {
	defer u.wg.Done()

	buf := make([]byte, UDPMaxPacketSize)
	for {
		select {
		case <-u.done:
			return
		default:
			if err := u.listener.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
				u.Log.Error("setting read-deadline failed: " + err.Error())
			}

			n, _, err := u.listener.ReadFromUDP(buf)
			if err != nil {
				if err, ok := err.(net.Error); !ok || !err.Timeout() {
					u.Log.Error(err.Error())
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
					u.Log.Errorf(dropwarn, u.drops)
				}
			}
		}
	}
}

func (u *UDPListener) udpParser() {
	defer u.wg.Done()

	var packet []byte
	var metrics []telegraf.Metric
	var err error
	for {
		select {
		case <-u.done:
			if len(u.in) == 0 {
				return
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
					u.Log.Errorf(malformedwarn, u.malformed)
				}
			}
		}
	}
}

func init() {
	inputs.Add("udp_listener", func() telegraf.Input {
		return &UDPListener{
			ServiceAddress:         ":8092",
			AllowedPendingMessages: 10000,
		}
	})
}
