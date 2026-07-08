package socket

import (
	"net"
	"sync"
)

// amount of pending streams to accept before dropping packets from new peers
const acceptBacklog = 16

const maxPacketSize = 65536 // maximum UDP packet size, including headers

var _ net.Listener = (*PacketStreamListener)(nil)

type PacketStreamListener struct {
	pc net.PacketConn

	mu       sync.Mutex
	conns    map[string]*conn
	acceptCh chan *conn
	closed   chan struct{}
	once     sync.Once
}

// Listen builds a packet connection via newPacketConn and wraps it in a
// stream net.Listener. Each remote address is treated as a separate stream.
func Listen(newPacketConn func() (net.PacketConn, error)) (*PacketStreamListener, error) {
	pc, err := newPacketConn()
	if err != nil {
		return nil, err
	}
	ln := &PacketStreamListener{
		pc:       pc,
		conns:    make(map[string]*conn),
		acceptCh: make(chan *conn, acceptBacklog),
		closed:   make(chan struct{}),
	}
	go ln.readLoop()
	return ln, nil
}

// readLoop reads packets from the shared connection and routes each one to the
// stream for its source address, creating (and offering to Accept) a new stream
// the first time a peer is seen.
func (ln *PacketStreamListener) readLoop() {
	buf := make([]byte, maxPacketSize)
	for {
		n, raddr, err := ln.pc.ReadFrom(buf)
		if err != nil {
			return
		}

		key := raddr.String()
		ln.mu.Lock()
		c := ln.conns[key]
		if c == nil {
			c = newConn(ln, raddr)
			ln.conns[key] = c
			ln.mu.Unlock()

			select {
			case ln.acceptCh <- c:
			case <-ln.closed:
				ln.mu.Lock()
				delete(ln.conns, key)
				ln.mu.Unlock()
				continue
			}
		} else {
			ln.mu.Unlock()
		}

		c.deliver(buf[:n])
	}
}

func (l *PacketStreamListener) Accept() (net.Conn, error) {
	select {
	case c := <-l.acceptCh:
		return c, nil
	case <-l.closed:
		return nil, net.ErrClosed
	}
}

func (l *PacketStreamListener) Addr() net.Addr {
	return l.pc.LocalAddr()
}

func (l *PacketStreamListener) Close() error {
	l.once.Do(func() {
		close(l.closed)
		l.pc.Close()

		l.mu.Lock()
		for _, c := range l.conns {
			c.fail(net.ErrClosed)
		}
		l.mu.Unlock()
	})
	return nil
}

// remove deregisters an accepted stream so a later packet from the same peer
// starts a fresh stream.
func (ln *PacketStreamListener) remove(c *conn) {
	ln.mu.Lock()
	if ln.conns[c.raddr.String()] == c {
		delete(ln.conns, c.raddr.String())
	}
	ln.mu.Unlock()
}
