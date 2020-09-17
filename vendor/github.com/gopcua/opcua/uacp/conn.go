// Copyright 2018-2020 opcua authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package uacp

import (
	"context"
	"io"
	"net"
	"sync/atomic"

	"github.com/gopcua/opcua/debug"
	"github.com/gopcua/opcua/errors"
	"github.com/gopcua/opcua/ua"
)

const (
	KB = 1024
	MB = 1024 * KB

	DefaultReceiveBufSize = 0xffff
	DefaultSendBufSize    = 0xffff
	DefaultMaxChunkCount  = 512
	DefaultMaxMessageSize = 2 * MB
)

// connid stores the current connection id. updated with atomic.AddUint32
var connid uint32

// nextid returns the next connection id
func nextid() uint32 {
	return atomic.AddUint32(&connid, 1)
}

func Dial(ctx context.Context, endpoint string) (*Conn, error) {
	debug.Printf("Connect to %s", endpoint)
	_, raddr, err := ResolveEndpoint(endpoint)
	if err != nil {
		return nil, err
	}
	var dialer net.Dialer
	c, err := dialer.DialContext(ctx, "tcp", raddr.String())
	if err != nil {
		return nil, err
	}

	conn := &Conn{
		TCPConn: c.(*net.TCPConn),
		id:      nextid(),
		ack: &Acknowledge{
			ReceiveBufSize: DefaultReceiveBufSize,
			SendBufSize:    DefaultSendBufSize,
			MaxChunkCount:  0, // use what the server wants
			MaxMessageSize: 0, // use what the server wants
		},
	}

	debug.Printf("conn %d: start HEL/ACK handshake", conn.id)
	if err := conn.handshake(endpoint); err != nil {
		debug.Printf("conn %d: HEL/ACK handshake failed: %s", conn.id, err)
		conn.Close()
		return nil, err
	}
	return conn, nil
}

// Listener is a OPC UA Connection Protocol network listener.
type Listener struct {
	l        *net.TCPListener
	ack      *Acknowledge
	endpoint string
}

// Listen acts like net.Listen for OPC UA Connection Protocol networks.
//
// Currently the endpoint can only be specified in "opc.tcp://<addr[:port]>/path" format.
//
// If the IP field of laddr is nil or an unspecified IP address, Listen listens
// on all available unicast and anycast IP addresses of the local system.
// If the Port field of laddr is 0, a port number is automatically chosen.
func Listen(endpoint string, ack *Acknowledge) (*Listener, error) {
	if ack == nil {
		ack = &Acknowledge{
			ReceiveBufSize: DefaultReceiveBufSize,
			SendBufSize:    DefaultSendBufSize,
			MaxChunkCount:  DefaultMaxChunkCount,
			MaxMessageSize: DefaultMaxMessageSize,
		}
	}

	network, laddr, err := ResolveEndpoint(endpoint)
	if err != nil {
		return nil, err
	}
	l, err := net.ListenTCP(network, laddr)
	if err != nil {
		return nil, err
	}
	return &Listener{
		l:        l,
		ack:      ack,
		endpoint: endpoint,
	}, nil
}

// Accept accepts the next incoming call and returns the new connection.
//
// The first param ctx is to be passed to monitor(), which monitors and handles
// incoming messages automatically in another goroutine.
func (l *Listener) Accept(ctx context.Context) (*Conn, error) {
	c, err := l.l.AcceptTCP()
	if err != nil {
		return nil, err
	}
	conn := &Conn{c, nextid(), l.ack}
	if err := conn.srvhandshake(l.endpoint); err != nil {
		c.Close()
		return nil, err
	}
	return conn, nil
}

// Close closes the Listener.
func (l *Listener) Close() error {
	return l.l.Close()
}

// Addr returns the listener's network address.
func (l *Listener) Addr() net.Addr {
	return l.l.Addr()
}

// Endpoint returns the listener's EndpointURL.
func (l *Listener) Endpoint() string {
	return l.endpoint
}

type Conn struct {
	*net.TCPConn
	id  uint32
	ack *Acknowledge
}

func (c *Conn) ID() uint32 {
	return c.id
}

func (c *Conn) ReceiveBufSize() uint32 {
	return c.ack.ReceiveBufSize
}

func (c *Conn) SendBufSize() uint32 {
	return c.ack.SendBufSize
}

func (c *Conn) MaxMessageSize() uint32 {
	return c.ack.MaxMessageSize
}

func (c *Conn) MaxChunkCount() uint32 {
	return c.ack.MaxChunkCount
}

func (c *Conn) Close() error {
	debug.Printf("conn %d: close", c.id)
	return c.TCPConn.Close()
}

func (c *Conn) handshake(endpoint string) error {
	hel := &Hello{
		Version:        c.ack.Version,
		ReceiveBufSize: c.ack.ReceiveBufSize,
		SendBufSize:    c.ack.SendBufSize,
		MaxMessageSize: c.ack.MaxMessageSize,
		MaxChunkCount:  c.ack.MaxChunkCount,
		EndpointURL:    endpoint,
	}

	if err := c.Send("HELF", hel); err != nil {
		return err
	}

	b, err := c.Receive()
	if err != nil {
		return err
	}

	msgtyp := string(b[:4])
	switch msgtyp {
	case "ACKF":
		ack := new(Acknowledge)
		if _, err := ack.Decode(b[hdrlen:]); err != nil {
			return errors.Errorf("uacp: decode ACK failed: %s", err)
		}
		if ack.Version != 0 {
			return errors.Errorf("uacp: invalid version %d", ack.Version)
		}
		if ack.MaxChunkCount == 0 {
			ack.MaxChunkCount = DefaultMaxChunkCount
			debug.Printf("conn %d: server has no chunk limit. Using %d", c.id, ack.MaxChunkCount)
		}
		if ack.MaxMessageSize == 0 {
			ack.MaxMessageSize = DefaultMaxMessageSize
			debug.Printf("conn %d: server has no message size limit. Using %d", c.id, ack.MaxMessageSize)
		}
		c.ack = ack
		debug.Printf("conn %d: recv %#v", c.id, ack)
		return nil

	case "ERRF":
		errf := new(Error)
		if _, err := errf.Decode(b[hdrlen:]); err != nil {
			return errors.Errorf("uacp: decode ERR failed: %s", err)
		}
		debug.Printf("conn %d: recv %#v", c.id, errf)
		return errf

	default:
		c.SendError(ua.StatusBadTCPInternalError)
		return errors.Errorf("invalid handshake packet %q", msgtyp)
	}
}

func (c *Conn) srvhandshake(endpoint string) error {
	b, err := c.Receive()
	if err != nil {
		c.SendError(ua.StatusBadTCPInternalError)
		return err
	}

	// HEL or RHE?
	msgtyp := string(b[:4])
	msg := b[hdrlen:]
	switch msgtyp {
	case "HELF":
		hel := new(Hello)
		if _, err := hel.Decode(msg); err != nil {
			c.SendError(ua.StatusBadTCPInternalError)
			return err
		}
		if hel.EndpointURL != endpoint {
			c.SendError(ua.StatusBadTCPEndpointURLInvalid)
			return errors.Errorf("uacp: invalid endpoint url %s", hel.EndpointURL)
		}
		if err := c.Send("ACKF", c.ack); err != nil {
			c.SendError(ua.StatusBadTCPInternalError)
			return err
		}
		debug.Printf("conn %d: recv %#v", c.id, hel)
		return nil

	case "RHEF":
		rhe := new(ReverseHello)
		if _, err := rhe.Decode(msg); err != nil {
			c.SendError(ua.StatusBadTCPInternalError)
			return err
		}
		if rhe.EndpointURL != endpoint {
			c.SendError(ua.StatusBadTCPEndpointURLInvalid)
			return errors.Errorf("uacp: invalid endpoint url %s", rhe.EndpointURL)
		}
		debug.Printf("conn %d: connecting to %s", c.id, rhe.ServerURI)
		c.Close()
		var dialer net.Dialer
		c2, err := dialer.DialContext(context.Background(), "tcp", rhe.ServerURI)
		if err != nil {
			return err
		}
		c.TCPConn = c2.(*net.TCPConn)
		debug.Printf("conn %d: recv %#v", c.id, rhe)
		return nil

	case "ERRF":
		errf := new(Error)
		if _, err := errf.Decode(b[hdrlen:]); err != nil {
			return errors.Errorf("uacp: decode ERR failed: %s", err)
		}
		debug.Printf("conn %d: recv %#v", c.id, errf)
		return errf

	default:
		c.SendError(ua.StatusBadTCPInternalError)
		return errors.Errorf("invalid handshake packet %q", msgtyp)
	}
}

// hdrlen is the size of the uacp header
const hdrlen = 8

// Receive reads a full UACP message from the underlying connection.
// The size of b must be at least ReceiveBufSize. Otherwise,
// the function returns an error.
func (c *Conn) Receive() ([]byte, error) {
	b := make([]byte, c.ack.ReceiveBufSize)

	if _, err := io.ReadFull(c, b[:hdrlen]); err != nil {
		// todo(fs): do not wrap this error since it hides io.EOF
		// todo(fs): use golang.org/x/xerrors
		return nil, err
	}

	var h Header
	if _, err := h.Decode(b[:hdrlen]); err != nil {
		return nil, errors.Errorf("uacp: header decode failed: %s", err)
	}

	if h.MessageSize > c.ack.ReceiveBufSize {
		return nil, errors.Errorf("uacp: message too large: %d > %d bytes", h.MessageSize, c.ack.ReceiveBufSize)
	}

	if _, err := io.ReadFull(c, b[hdrlen:h.MessageSize]); err != nil {
		// todo(fs): do not wrap this error since it hides io.EOF
		// todo(fs): use golang.org/x/xerrors
		return nil, err
	}

	debug.Printf("conn %d: recv %s%c with %d bytes", c.id, h.MessageType, h.ChunkType, h.MessageSize)

	if h.MessageType == "ERR" {
		errf := new(Error)
		if _, err := errf.Decode(b[hdrlen:h.MessageSize]); err != nil {
			return nil, errors.Errorf("uacp: failed to decode ERRF message: %s", err)
		}
		return nil, errf
	}
	return b[:h.MessageSize], nil
}

func (c *Conn) Send(typ string, msg interface{}) error {
	if len(typ) != 4 {
		return errors.Errorf("invalid msg type: %s", typ)
	}

	body, err := ua.Encode(msg)
	if err != nil {
		return errors.Errorf("encode msg failed: %s", err)
	}

	h := Header{
		MessageType: typ[:3],
		ChunkType:   typ[3],
		MessageSize: uint32(len(body) + hdrlen),
	}

	if h.MessageSize > c.ack.SendBufSize {
		return errors.Errorf("send packet too large: %d > %d bytes", h.MessageSize, c.ack.SendBufSize)
	}

	hdr, err := h.Encode()
	if err != nil {
		return errors.Errorf("encode hdr failed: %s", err)
	}

	b := append(hdr, body...)
	if _, err := c.Write(b); err != nil {
		return errors.Errorf("write failed: %s", err)
	}
	debug.Printf("conn %d: sent %s with %d bytes", c.id, typ, len(b))

	return nil
}

func (c *Conn) SendError(code ua.StatusCode) {
	// we swallow the error to silence complaints from the linter
	// since sending an error will close the connection and we
	// want to bubble a different error up.
	_ = c.Send("ERRF", &Error{ErrorCode: uint32(code)})
}
