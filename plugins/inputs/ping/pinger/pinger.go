// Package ping is an ICMP ping library seeking to emulate the unix "ping"
// command.
//
// Here is a very simple example that sends & receives 3 packets:
//
//	pinger, err := ping.NewPinger("www.google.com")
//	if err != nil {
//		panic(err)
//	}
//
//	pinger.Count = 3
//	pinger.Run() // blocks until finished
//	stats := pinger.Statistics() // get send/receive/rtt stats
//
// Here is an example that emulates the unix ping command:
//
//	pinger, err := ping.NewPinger("www.google.com")
//	if err != nil {
//		fmt.Printf("ERROR: %s\n", err.Error())
//		return
//	}
//
//	pinger.OnRecv = func(pkt *ping.Packet) {
//		fmt.Printf("%d bytes from %s: icmp_seq=%d time=%v\n",
//			pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Rtt)
//	}
//	pinger.OnFinish = func(stats *ping.Statistics) {
//		fmt.Printf("\n--- %s ping statistics ---\n", stats.Addr)
//		fmt.Printf("%d packets transmitted, %d packets received, %v%% packet loss\n",
//			stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss)
//		fmt.Printf("round-trip min/avg/max/stddev = %v/%v/%v/%v\n",
//			stats.MinRtt, stats.AvgRtt, stats.MaxRtt, stats.StdDevRtt)
//	}
//
//	fmt.Printf("PING %s (%s):\n", pinger.Addr(), pinger.IPAddr())
//	pinger.Run()
//
// It sends ICMP packet(s) and waits for a response. If it receives a response,
// it calls the "receive" callback. When it's finished, it calls the "finish"
// callback.
//
// For a full ping example, see "cmd/ping/ping.go".
//
package pinger

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

const (
	timeSliceLength  = 8
	protocolICMP     = 1
	protocolIPv6ICMP = 58
)

var (
	ipv4Proto = map[string]string{"ip": "ip4:icmp", "udp": "udp4"}
	ipv6Proto = map[string]string{"ip": "ip6:ipv6-icmp", "udp": "udp6"}
)

// NewPinger returns a new Pinger struct pointer
func NewPinger(privileged bool) (*Pinger, error) {
	p := &Pinger{
		Interval: time.Second,
		Timeout:  time.Second * 100000,
		Count:    -1,
		network:  "udp",
		Size:     timeSliceLength,
		done:     make(chan bool),
	}

	sends := &SendEntries{
		entries: make(map[string]*SendEntry),
	}
	p.sends = sends

	if privileged {
		p.network = "ip"
	} else {
		p.network = "udp"
	}

	p.ipv4 = true
	if p.conn = p.listen(ipv4Proto[p.network], p.source); p.conn == nil {
		return nil, errors.New("Could not create ipv4 connection")
	}

	go p.Recv()

	return p, nil

}

type SendEntries struct {
	entries map[string]*SendEntry
	mutex   sync.RWMutex
}

type SendEntry struct {
	rc      chan *Packet
	id      int
	timeout time.Duration
	time    time.Time
}

// Pinger represents ICMP packet sender/receiver
type Pinger struct {
	// Interval is the wait time between each packet send. Default is 1s.
	Interval time.Duration

	// Timeout specifies a timeout before ping exits, regardless of how many
	// packets have been received.
	Timeout time.Duration

	// Count tells pinger to stop after sending (and receiving) Count echo
	// packets. If this option is not specified, pinger will operate until
	// interrupted.
	Count int

	// Debug runs in debug mode
	Debug bool

	// Number of packets sent
	PacketsSent int

	// Number of packets received
	PacketsRecv int

	// rtts is all of the Rtts
	rtt time.Duration

	// Size of packet being sent
	Size int

	// stop chan bool
	done chan bool

	addr string
	conn *icmp.PacketConn

	sends *SendEntries

	ipv4     bool
	source   string
	size     int
	id       int
	sequence int
	network  string
}

type packet struct {
	bytes  []byte
	nbytes int
}

// Packet represents a received and processed ICMP echo packet.
type Packet struct {
	// Rtt is the round-trip time it took to ping.
	Rtt time.Duration

	// IPAddr is the address of the host being pinged.
	IPAddr *net.IPAddr

	// NBytes is the number of bytes in the message.
	Nbytes int

	// Seq is the ICMP sequence number.
	Seq int
}

// Statistics represent the stats of a currently running or finished
// pinger operation.
type Statistics struct {
	// PacketsRecv is the number of packets received.
	PacketsRecv int

	// PacketsSent is the number of packets sent.
	PacketsSent int

	// PacketLoss is the percentage of packets lost.
	PacketLoss float64

	// IPAddr is the address of the host being pinged.
	IPAddr *net.IPAddr

	// Addr is the string address of the host being pinged.
	Addr string

	// Rtts is all of the round-trip times sent via this pinger.
	Rtt time.Duration
}

func (s *SendEntries) Get(item string) *SendEntry {
	s.mutex.RLock()
	ret := s.entries[item]
	s.mutex.RUnlock()

	return ret
}

func (s *SendEntries) GetTimeouts() []*SendEntry {
	var ret []*SendEntry
	s.mutex.RLock()
	for _, entry := range s.entries {
		if time.Since(entry.time) > entry.timeout {
			ret = append(ret, entry)
		}
	}
	s.mutex.RUnlock()

	return ret
}

func (s *SendEntries) FindById(id int) *SendEntry {
	var ret *SendEntry
	s.mutex.RLock()
	for _, entry := range s.entries {
		if entry.id == id {
			ret = entry
			break
		}
	}
	s.mutex.RUnlock()

	return ret
}

func (s *SendEntries) Set(item string, entry *SendEntry) {
	s.mutex.Lock()
	s.entries[item] = entry
	s.mutex.Unlock()
}

func (s *SendEntries) Add(addr string, timeout int64) *SendEntry {
	s.mutex.Lock()
	var newEntry *SendEntry
	if _, ok := s.entries[addr]; !ok {
		c := make(chan *Packet)
		id := rand.Intn(0xffff)
		newEntry = &SendEntry{
			rc:      c,
			id:      id,
			timeout: time.Duration(timeout) * time.Millisecond,
			time:    time.Now(),
		}
	}
	s.entries[addr] = newEntry
	s.mutex.Unlock()

	return newEntry
}

func (s *SendEntries) Del(item string) {
	s.mutex.Lock()
	delete(s.entries, item)
	s.mutex.Unlock()
}

// Addr returns the string ip address of the target host.
func (p *Pinger) Addr() string {
	return p.addr
}

// SetPrivileged sets the type of ping pinger will send.
// false means pinger will send an "unprivileged" UDP ping.
// true means pinger will send a "privileged" raw ICMP ping.
// NOTE: setting to true requires that it be run with super-user privileges.
func (p *Pinger) SetPrivileged(privileged bool) {
	if privileged {
		p.network = "ip"
	} else {
		p.network = "udp"
	}
}

// Privileged returns whether pinger is running in privileged mode.
func (p *Pinger) Privileged() bool {
	return p.network == "ip"
}

func (p *Pinger) Recv() {
	recv := make(chan *packet, 1000)
	go p.recvICMP(p.conn, recv)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)

	for {
		select {
		case <-c:
			close(p.done)
		case <-p.done:
			return
		case r := <-recv:
			err := p.processPacket(r)
			if err != nil {
				log.Println("ERROR [ping.processPacket]", err)
			}
		default:
			time.Sleep(1 * time.Millisecond)
		}

		for _, entry := range p.sends.GetTimeouts() {
			outPkt := &Packet{Nbytes: 0}
			outPkt.Rtt = entry.timeout
			entry.rc <- outPkt
		}
	}
}

func (p *Pinger) Send(addr string, timeout int64) (*Statistics, error) {
	entry := p.sends.Add(addr, timeout)
	recv := entry.rc
	err := p.sendICMP(p.conn, addr)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	for {
		select {
		case r := <-recv:
			p.sends.Del(addr)
			rttMs := int64(r.Rtt / 1000000)
			ret := &Statistics{
				Rtt:         r.Rtt,
				PacketsSent: 1,
				PacketsRecv: 1,
				PacketLoss:  0,
			}
			if rttMs == timeout {
				ret.PacketsRecv = 0
				ret.PacketLoss = 100
				return ret, errors.New("timed out")
			} else {
				return ret, nil
			}
		default:
			time.Sleep(1 * time.Millisecond)
		}
	}
}

// Statistics returns the statistics of the pinger. This can be run while the
// pinger is running or after it is finished. OnFinish calls this function to
// get it's finished statistics.
func (p *Pinger) Statistics() *Statistics {
	loss := float64(p.PacketsSent-p.PacketsRecv) / float64(p.PacketsSent) * 100

	s := Statistics{
		PacketsSent: p.PacketsSent,
		PacketsRecv: p.PacketsRecv,
		PacketLoss:  loss,
		Rtt:         p.rtt,
		Addr:        p.addr,
	}

	return &s
}

func (p *Pinger) recvICMP(
	conn *icmp.PacketConn,
	recv chan<- *packet,
) {
	for {
		select {
		case <-p.done:
			return
		default:
			bytes := make([]byte, 512)
			conn.SetReadDeadline(time.Now().Add(time.Millisecond * 100))
			n, _, err := conn.ReadFrom(bytes)
			if err != nil {
				if neterr, ok := err.(*net.OpError); ok {
					if neterr.Timeout() {
						// Read timeout
						continue
					} else {
						close(p.done)
						return
					}
				}
			}

			recv <- &packet{bytes: bytes, nbytes: n}
		}
	}
}

func (p *Pinger) processPacket(recv *packet) error {
	var bytes []byte
	var proto int
	if p.ipv4 {
		if p.network == "ip" {
			bytes = ipv4Payload(recv.bytes)
		} else {
			bytes = recv.bytes
		}
		proto = protocolICMP
	} else {
		bytes = recv.bytes
		proto = protocolIPv6ICMP
	}

	var m *icmp.Message
	var err error
	if m, err = icmp.ParseMessage(proto, bytes[:recv.nbytes]); err != nil {
		return fmt.Errorf("Error parsing icmp message")
	}

	if m.Type != ipv4.ICMPTypeEchoReply && m.Type != ipv6.ICMPTypeEchoReply {
		// Not an echo reply, ignore it
		return nil
	}

	// Check if reply from same ID
	body := m.Body.(*icmp.Echo)
	entry := p.sends.FindById(body.ID)
	if entry != nil {
		outPkt := &Packet{
			Nbytes: recv.nbytes,
		}

		switch pkt := m.Body.(type) {
		case *icmp.Echo:
			outPkt.Rtt = time.Since(bytesToTime(pkt.Data[:timeSliceLength]))
			outPkt.Seq = pkt.Seq
			entry.rc <- outPkt
			break
		default:
			// Very bad, not sure how this can happen
			err = fmt.Errorf("Error, invalid ICMP echo reply. Body type: %T, %s",
				pkt, pkt)
		}
	}

	return err
}

func (p *Pinger) sendICMP(conn *icmp.PacketConn, addr string) error {
	var typ icmp.Type
	if p.ipv4 {
		typ = ipv4.ICMPTypeEcho
	} else {
		typ = ipv6.ICMPTypeEchoRequest
	}

	ipaddr, err := net.ResolveIPAddr("ip", addr)
	if err != nil {
		return err
	}

	var dst net.Addr = ipaddr
	if p.network == "udp" {
		dst = &net.UDPAddr{IP: ipaddr.IP, Zone: ipaddr.Zone}
	}

	t := timeToBytes(time.Now())
	if p.Size-timeSliceLength != 0 {
		t = append(t, byteSliceOfSize(p.Size-timeSliceLength)...)
	}

	id := p.sends.Get(addr).id
	bytes, err := (&icmp.Message{
		Type: typ, Code: 0,
		Body: &icmp.Echo{
			ID:   id,
			Seq:  p.sequence,
			Data: t,
		},
	}).Marshal(nil)
	if err != nil {
		return err
	}

	for {
		if _, err := conn.WriteTo(bytes, dst); err != nil {
			if neterr, ok := err.(*net.OpError); ok {
				if neterr.Err == syscall.ENOBUFS {
					continue
				}
			}
		}
		p.PacketsSent += 1
		p.sequence += 1
		break
	}
	return nil
}

func (p *Pinger) listen(netProto string, source string) *icmp.PacketConn {
	conn, err := icmp.ListenPacket(netProto, source)
	if err != nil {
		fmt.Printf("Error listening for ICMP packets: %s\n", err.Error())
		close(p.done)
		return nil
	}
	return conn
}

func byteSliceOfSize(n int) []byte {
	b := make([]byte, n)
	for i := 0; i < len(b); i++ {
		b[i] = 1
	}

	return b
}

func ipv4Payload(b []byte) []byte {
	if len(b) < ipv4.HeaderLen {
		return b
	}
	hdrlen := int(b[0]&0x0f) << 2
	return b[hdrlen:]
}

func bytesToTime(b []byte) time.Time {
	var nsec int64
	for i := uint8(0); i < 8; i++ {
		nsec += int64(b[i]) << ((7 - i) * 8)
	}
	return time.Unix(nsec/1000000000, nsec%1000000000)
}

func isIPv4(ip net.IP) bool {
	return len(ip.To4()) == net.IPv4len
}

func isIPv6(ip net.IP) bool {
	return len(ip) == net.IPv6len
}

func timeToBytes(t time.Time) []byte {
	nsec := t.UnixNano()
	b := make([]byte, 8)
	for i := uint8(0); i < 8; i++ {
		b[i] = byte((nsec >> ((7 - i) * 8)) & 0xff)
	}
	return b
}
