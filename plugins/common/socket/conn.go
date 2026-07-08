package socket

import (
	"bytes"
	"net"
	"os"
	"sync"
	"time"
)

var maxPayload = maxPacketSize - 64 // some overhead for headers

var _ net.Conn = (*conn)(nil)

// conn is a byte stream over the packets exchanged with a single
// remote peer on a listener's net.PacketConn. Received packet payloads
// are concatenated into a buffer, and writes are sent as one or more packets to the peer.
type conn struct {
	pc    net.PacketConn        // shared with the listener; not closed by conn
	raddr net.Addr              // peer address: the packet source and write destination
	ln    *PacketStreamListener // used to deregister on Close

	mu   sync.Mutex
	buf  bytes.Buffer
	rerr error // returned by Read once buf drains (e.g. after the socket fails)

	data      chan struct{} // buffered(1); pinged when new bytes are buffered
	done      chan struct{} // closed by Close to broadcast shutdown to all waiters
	closeOnce sync.Once

	readDeadline  pipeDeadline
	writeDeadline pipeDeadline
}

func newConn(ln *PacketStreamListener, raddr net.Addr) *conn {
	return &conn{
		pc:            ln.pc,
		raddr:         raddr,
		ln:            ln,
		data:          make(chan struct{}, 1),
		done:          make(chan struct{}),
		readDeadline:  makePipeDeadline(),
		writeDeadline: makePipeDeadline(),
	}
}

// deliver appends one received packet's payload to the stream buffer.
func (c *conn) deliver(payload []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.rerr != nil || isClosedChan(c.done) {
		return
	}
	c.buf.Write(payload)
	c.signal()
}

// fail records a terminal read error (e.g. the socket closing) and wakes any
// blocked reader
func (c *conn) fail(err error) {
	c.mu.Lock()
	if c.rerr == nil {
		c.rerr = err
	}
	c.mu.Unlock()
	c.signal()
}

// signal performs a non-blocking wake of a reader parked on c.data.
func (c *conn) signal() {
	select {
	case c.data <- struct{}{}:
	default:
	}
}

// Read implements the net.Conn interface, reading from the stream buffer.
func (c *conn) Read(b []byte) (int, error) {
	for {
		c.mu.Lock()
		if c.buf.Len() > 0 {
			n, _ := c.buf.Read(b)
			c.mu.Unlock()
			return n, nil
		}
		rerr := c.rerr
		c.mu.Unlock()

		if rerr != nil {
			return 0, rerr
		}
		if isClosedChan(c.done) {
			return 0, net.ErrClosed
		}
		if isClosedChan(c.readDeadline.wait()) {
			return 0, os.ErrDeadlineExceeded
		}

		select {
		case <-c.data:
			// New bytes may be buffered; loop to re-check under the lock.
		case <-c.done:
			return 0, net.ErrClosed
		case <-c.readDeadline.wait():
			return 0, os.ErrDeadlineExceeded
		}
	}
}

// Write implements the net.Conn interface, sending the given bytes as one or more packets to the peer.
func (c *conn) Write(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}

	total := 0
	for len(b) > 0 {
		if isClosedChan(c.done) {
			return total, net.ErrClosed
		}
		if isClosedChan(c.writeDeadline.wait()) {
			return total, os.ErrDeadlineExceeded
		}

		n := min(len(b), maxPayload)
		if _, err := c.pc.WriteTo(b[:n], c.raddr); err != nil {
			return total, err
		}
		total += n
		b = b[n:]
	}
	return total, nil
}

// Close implements the net.Conn interface, closing the stream and removing it from the listener.
func (c *conn) Close() error {
	c.closeOnce.Do(func() {
		// close all the pending readers and writers
		close(c.done)
		// deregister from the listener so a later packet from the same peer starts a fresh stream
		c.ln.remove(c)
	})
	return nil
}

func (c *conn) LocalAddr() net.Addr  { return c.pc.LocalAddr() }
func (c *conn) RemoteAddr() net.Addr { return c.raddr }

func (c *conn) SetDeadline(t time.Time) error {
	err := c.SetReadDeadline(t)
	if err != nil {
		return err
	}
	err = c.SetWriteDeadline(t)
	return err
}

func (c *conn) SetReadDeadline(t time.Time) error {
	c.readDeadline.set(t)
	c.signal() // wake a parked reader so it re-evaluates the deadline
	return nil
}

func (c *conn) SetWriteDeadline(t time.Time) error {
	c.writeDeadline.set(t)
	return nil
}
