package rpc

import (
	"bytes"
	"io"
	"time"

	"golang.org/x/net/context"
	"zombiezen.com/go/capnproto2"
	rpccapnp "zombiezen.com/go/capnproto2/std/capnp/rpc"
)

// Transport is the interface that abstracts sending and receiving
// individual messages of the Cap'n Proto RPC protocol.
type Transport interface {
	// SendMessage sends msg.
	SendMessage(ctx context.Context, msg rpccapnp.Message) error

	// RecvMessage waits to receive a message and returns it.
	// Implementations may re-use buffers between calls, so the message is
	// only valid until the next call to RecvMessage.
	RecvMessage(ctx context.Context) (rpccapnp.Message, error)

	// Close releases any resources associated with the transport.
	Close() error
}

type streamTransport struct {
	rwc      io.ReadWriteCloser
	deadline writeDeadlineSetter

	enc  *capnp.Encoder
	dec  *capnp.Decoder
	wbuf bytes.Buffer
}

// StreamTransport creates a transport that sends and receives messages
// by serializing and deserializing unpacked Cap'n Proto messages.
// Closing the transport will close the underlying ReadWriteCloser.
func StreamTransport(rwc io.ReadWriteCloser) Transport {
	d, _ := rwc.(writeDeadlineSetter)
	s := &streamTransport{
		rwc:      rwc,
		deadline: d,
		dec:      capnp.NewDecoder(rwc),
	}
	s.wbuf.Grow(4096)
	s.enc = capnp.NewEncoder(&s.wbuf)
	return s
}

func (s *streamTransport) SendMessage(ctx context.Context, msg rpccapnp.Message) error {
	s.wbuf.Reset()
	if err := s.enc.Encode(msg.Segment().Message()); err != nil {
		return err
	}
	if s.deadline != nil {
		// TODO(light): log errors
		if d, ok := ctx.Deadline(); ok {
			s.deadline.SetWriteDeadline(d)
		} else {
			s.deadline.SetWriteDeadline(time.Time{})
		}
	}
	_, err := s.rwc.Write(s.wbuf.Bytes())
	return err
}

func (s *streamTransport) RecvMessage(ctx context.Context) (rpccapnp.Message, error) {
	var (
		msg *capnp.Message
		err error
	)
	read := make(chan struct{})
	go func() {
		msg, err = s.dec.Decode()
		close(read)
	}()
	select {
	case <-read:
	case <-ctx.Done():
		return rpccapnp.Message{}, ctx.Err()
	}
	if err != nil {
		return rpccapnp.Message{}, err
	}
	return rpccapnp.ReadRootMessage(msg)
}

func (s *streamTransport) Close() error {
	return s.rwc.Close()
}

type writeDeadlineSetter interface {
	SetWriteDeadline(t time.Time) error
}

// dispatchSend runs in its own goroutine and sends messages on a transport.
func (c *Conn) dispatchSend() {
	defer c.workers.Done()
	for {
		select {
		case msg := <-c.out:
			err := c.transport.SendMessage(c.bg, msg)
			if err != nil {
				c.errorf("writing %v: %v", msg.Which(), err)
			}
		case <-c.bg.Done():
			return
		}
	}
}

// sendMessage enqueues a message to be sent or returns an error if the
// connection is shut down before the message is queued.  It is safe to
// call from multiple goroutines and does not require holding c.mu.
func (c *Conn) sendMessage(msg rpccapnp.Message) error {
	select {
	case c.out <- msg:
		return nil
	case <-c.bg.Done():
		return ErrConnClosed
	}
}

// dispatchRecv runs in its own goroutine and receives messages from a transport.
func (c *Conn) dispatchRecv() {
	defer c.workers.Done()
	for {
		msg, err := c.transport.RecvMessage(c.bg)
		if err == nil {
			c.handleMessage(msg)
		} else if isTemporaryError(err) {
			c.errorf("read temporary error: %v", err)
		} else {
			c.shutdown(err)
			return
		}
	}
}

// copyMessage clones a Cap'n Proto buffer.
func copyMessage(msg *capnp.Message) *capnp.Message {
	n := msg.NumSegments()
	segments := make([][]byte, n)
	for i := range segments {
		s, err := msg.Segment(capnp.SegmentID(i))
		if err != nil {
			panic(err)
		}
		segments[i] = make([]byte, len(s.Data()))
		copy(segments[i], s.Data())
	}
	return &capnp.Message{Arena: capnp.MultiSegment(segments)}
}

// copyRPCMessage clones an RPC packet.
func copyRPCMessage(m rpccapnp.Message) rpccapnp.Message {
	mm := copyMessage(m.Segment().Message())
	rpcMsg, err := rpccapnp.ReadRootMessage(mm)
	if err != nil {
		panic(err)
	}
	return rpcMsg
}

// isTemporaryError reports whether e has a Temporary() method that
// returns true.
func isTemporaryError(e error) bool {
	type temp interface {
		Temporary() bool
	}
	t, ok := e.(temp)
	return ok && t.Temporary()
}
