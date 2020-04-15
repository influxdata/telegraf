package ping

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

const (
	// protocolIPv4ICMP defines ICMP for IPv4.
	protocolIPv4ICMP = 1
	// protocolIPv6ICMP defines ICMP for IPv6.
	protocolIPv6ICMP = 58
)

// A Request represents an icmp echo request to be sent by a client.
type Request struct {
	sentAt time.Time // time at which echo request was sent
	dst    net.Addr  // useable destination address

	ID   int    // ID is the ICMP ID. It is an identifier to aid in matching echos and replies when using privileged datagrams, may be zero.
	Seq  int    // Seq is the ICMP sequence number.
	Data []byte // Data is generally an arbitrary byte string of size 56. It is used to set icmp.Echo.Data.
	Dst  net.IP // The address of the host to which the message should be sent.
	Src  net.IP // The address of the host that composes the ICMP message.
}

// Response represents an icmp echo response received by a client.
type Response struct {
	rcvdAt time.Time // time at which echo response was received

	TotalLength int           // Length of internet header and data of the echo response in octets.
	RTT         time.Duration // RTT is the round-trip time it took to ping.
	TTL         int           // Time to live in seconds; as this field is decremented at each machine in which the datagram is processed, the value in this field should be at least as great as the number of gateways which this datagram will traverse. Maximum possible value of this field is 255.

	ID   int    // ID is the ICMP ID. It is an identifier to aid in matching echos and replies when using privileged datagrams, may be zero.
	Seq  uint   // Seq is the ICMP sequence number.
	Data []byte // Data is the body of the ICMP response.
	Dst  net.IP // The local address of the host that composed the echo request.
	Src  net.IP // The address of the host to which the message was received from.

	Req *Request // Req is the request that elicited this response.
}

// Client is a ping client.
type Client struct{}

// DefaultClient is the default client used by Do.
var DefaultClient = &Client{}

// NewRequest resolves dst as an IPv4 address and returns a pointer to a request
// using that as the destination.
func NewRequest(dst string) (*Request, error) {
	host, err := net.ResolveIPAddr("ip4", dst)
	if err != nil {
		return nil, fmt.Errorf("can't resolve host: %s", err.Error())
	}

	return &Request{Dst: net.ParseIP(host.String())}, nil
}

// NewRequest6 resolves dst as an IPv6 address and returns a pointer to a request
// using that as the destination.
func NewRequest6(dst string) (*Request, error) {
	host, err := net.ResolveIPAddr("ip6", dst)
	if err != nil {
		return nil, fmt.Errorf("can't resolve host: %s", err.Error())
	}

	return &Request{Dst: net.ParseIP(host.String())}, nil
}

// Do sends a ping request using the default client and returns a ping response.
func Do(ctx context.Context, req *Request) (*Response, error) {
	return DefaultClient.Do(ctx, req)
}

// Do sends a ping request and returns a ping response.
func (c *Client) Do(ctx context.Context, req *Request) (*Response, error) {
	conn, network, err := c.listen(req)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var addr net.Addr
	if req.isIPv6() {
		addr, err = net.ResolveIPAddr("ip6", req.Dst.String())
	} else {
		addr, err = net.ResolveIPAddr("ip4", req.Dst.String())
	}
	if err != nil {
		return nil, err
	}
	req.dst = addr

	switch network {
	case "udp4", "udp6":
		if a, ok := addr.(*net.IPAddr); ok {
			req.dst = &net.UDPAddr{IP: a.IP, Zone: a.Zone}
		}
	}

	sentAt, err := send(ctx, conn, req)
	if err != nil {
		return nil, err
	}

	resp, err := read(ctx, conn, req)
	if err != nil {
		return nil, err
	}

	resp.RTT = resp.rcvdAt.Sub(sentAt)
	req.sentAt = sentAt
	resp.Req = req

	return resp, nil
}

// IPv4 resolves dst as an IPv4 address and pings using DefaultClient, returning
// the response and error.
func IPv4(ctx context.Context, dst string) (*Response, error) {
	req, err := NewRequest(dst)
	if err != nil {
		return nil, err
	}
	return Do(ctx, req)
}

// IPv6 resolves dst as an IPv6 address and pings using DefaultClient, returning
// the response and error.
func IPv6(ctx context.Context, dst string) (*Response, error) {
	req, err := NewRequest6(dst)
	if err != nil {
		return nil, err
	}
	return Do(ctx, req)
}

// listen tries first to create a privileged datagram-oriented ICMP endpoint then
// attempts to create a non-privileged one. If both fail, it returns an error.
func (c *Client) listen(req *Request) (*icmp.PacketConn, string, error) {
	network := "ip4:icmp"

	if req.isIPv6() {
		network = "ip6:ipv6-icmp"
	}

	srcIP := req.Src.String()
	if srcIP == "<nil>" {
		srcIP = ""
	}

	conn, err := icmp.ListenPacket(network, srcIP)
	if err != nil {
		network = "udp4"
		if req.isIPv6() {
			network = "udp6"
		}

		var err2 error
		conn, err2 = icmp.ListenPacket(network, srcIP)
		if err2 != nil {
			return nil, "", fmt.Errorf("error listening for ICMP packets: %s: %s", err.Error(), err2.Error())
		}
	}

	return conn, network, nil
}

func (req *Request) isIPv6() bool {
	if p4 := req.Dst.To4(); len(p4) == net.IPv4len {
		return false
	}
	return true
}

func (req *Request) proto() int {
	if req.isIPv6() {
		return protocolIPv6ICMP
	}
	return protocolIPv4ICMP
}

func read(ctx context.Context, conn *icmp.PacketConn, req *Request) (*Response, error) {
	if c4 := conn.IPv4PacketConn(); c4 != nil {
		return read4(ctx, c4, req)
	}
	if c6 := conn.IPv6PacketConn(); c6 != nil {
		return read6(ctx, c6, req)
	}
	return nil, errors.New("bad icmp connection type")
}

func read4(ctx context.Context, conn *ipv4.PacketConn, req *Request) (*Response, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			conn.SetReadDeadline(time.Now().Add(time.Millisecond * 100))
			bytesReceived := make([]byte, 1500)

			n, cm, src, err := conn.ReadFrom(bytesReceived)
			rcv := time.Now()
			if err != nil {
				if neterr, ok := err.(*net.OpError); ok {
					if neterr.Timeout() {
						continue
					} else {
						return nil, err
					}
				}
				return nil, err
			}

			// If there was no data read or the packet didn't originate from the host
			// assumed, skip processing.
			if n <= 0 || (cm != nil && cm.Src.String() != req.Dst.String()) {
				continue
			}

			// Process the data as an icmp message.
			m, err := icmp.ParseMessage(protocolIPv4ICMP, bytesReceived[:n])
			if err != nil {
				return nil, err
			}

			// Likely an `ICMPTypeDestinationUnreachable`, ignore it.
			if m.Type != ipv4.ICMPTypeEchoReply {
				continue
			}

			// Verify the sequence numbers match our expectations for correct rtt.
			// If using `ip4:icmp`, the ID can be verified as well (preferred).
			b, ok := m.Body.(*icmp.Echo)
			if !ok || b.Seq != req.Seq {
				continue
			}

			srcHost, _, _ := net.SplitHostPort(src.String())
			dstHost, _, _ := net.SplitHostPort(conn.LocalAddr().String())
			resp := &Response{
				ID:          b.ID,
				Seq:         uint(b.Seq),
				Data:        b.Data,
				TotalLength: n,
				Src:         net.ParseIP(srcHost),
				Dst:         net.ParseIP(dstHost),
				rcvdAt:      rcv,
			}
			if cm != nil {
				resp.TTL = cm.TTL
			}
			return resp, nil
		}
	}
}

func read6(ctx context.Context, conn *ipv6.PacketConn, req *Request) (*Response, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			conn.SetReadDeadline(time.Now().Add(time.Millisecond * 100))
			bytesReceived := make([]byte, 1500)

			n, cm, src, err := conn.ReadFrom(bytesReceived)
			rcv := time.Now()
			if err != nil {
				if neterr, ok := err.(*net.OpError); ok {
					if neterr.Timeout() {
						continue
					} else {
						return nil, err
					}
				}
				return nil, err
			}

			// If there was no data read or the packet didn't originate from the host
			// assumed, skip processing.
			if n <= 0 || cm.Src.String() != req.Dst.String() {
				continue
			}

			// Process the data as an icmp message.
			m, err := icmp.ParseMessage(protocolIPv6ICMP, bytesReceived[:n])
			if err != nil {
				return nil, err
			}

			// Likely an `ICMPTypeDestinationUnreachable`, ignore it.
			if m.Type != ipv6.ICMPTypeEchoReply {
				continue
			}

			// Verify the sequence numbers match our expectations for correct rtt.
			// If using `ip4:icmp`, the ID can be verified as well (preferred).
			b, ok := m.Body.(*icmp.Echo)
			if !ok || b.Seq != req.Seq {
				continue
			}

			srcHost, _, _ := net.SplitHostPort(src.String())
			dstHost, _, _ := net.SplitHostPort(conn.LocalAddr().String())
			return &Response{
				ID:          b.ID,
				Seq:         uint(b.Seq),
				Data:        bytesReceived[:n],
				TotalLength: n,
				Src:         net.ParseIP(srcHost),
				Dst:         net.ParseIP(dstHost),
				TTL:         cm.HopLimit,
				rcvdAt:      rcv,
			}, nil
		}
	}
}

func (req *Request) data() []byte {
	if len(req.Data) == 0 {
		return bytes.Repeat([]byte{1}, 56)
	}
	return req.Data
}

func send(ctx context.Context, conn *icmp.PacketConn, req *Request) (time.Time, error) {
	sentAt := time.Time{}
	select {
	case <-ctx.Done():
		return sentAt, nil
	default:
		body := &icmp.Echo{
			ID:   req.ID,
			Seq:  req.Seq,
			Data: req.data(),
		}

		msg := &icmp.Message{
			Type: ipv6.ICMPTypeEchoRequest,
			Code: 0,
			Body: body,
		}

		if req.proto() == protocolIPv4ICMP {
			msg.Type = ipv4.ICMPTypeEcho
			conn.IPv4PacketConn().SetControlMessage(ipv4.FlagTTL|ipv4.FlagSrc|ipv4.FlagDst, true)
		} else {
			conn.IPv6PacketConn().SetControlMessage(ipv6.FlagHopLimit|ipv6.FlagSrc|ipv6.FlagDst, true)
		}

		msgBytes, err := msg.Marshal(nil)
		if err != nil {
			return sentAt, err
		}

		if timeout, ok := ctx.Deadline(); ok {
			if err := conn.SetWriteDeadline(timeout); err != nil {
				return sentAt, err
			}
		}
		if _, err := conn.WriteTo(msgBytes, req.dst); err != nil {
			return sentAt, err
		}
		sentAt = time.Now()
	}

	return sentAt, nil
}
