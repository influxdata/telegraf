// Package pipetransport provides in-memory implementations of rpc.Transport for testing.
package pipetransport

import (
	"bytes"
	"errors"
	"sync"

	"golang.org/x/net/context"
	"zombiezen.com/go/capnproto2"
	"zombiezen.com/go/capnproto2/rpc"
	rpccapnp "zombiezen.com/go/capnproto2/std/capnp/rpc"
)

type pipeTransport struct {
	r        <-chan rpccapnp.Message
	w        chan<- rpccapnp.Message
	finish   chan struct{}
	otherFin chan struct{}

	rbuf bytes.Buffer

	mu       sync.Mutex
	inflight int
	done     bool
}

// New creates a synchronous in-memory pipe transport.
func New() (p, q rpc.Transport) {
	a, b := make(chan rpccapnp.Message), make(chan rpccapnp.Message)
	afin, bfin := make(chan struct{}), make(chan struct{})
	p = &pipeTransport{
		r:        a,
		w:        b,
		finish:   afin,
		otherFin: bfin,
	}
	q = &pipeTransport{
		r:        b,
		w:        a,
		finish:   bfin,
		otherFin: afin,
	}
	return
}

func (p *pipeTransport) SendMessage(ctx context.Context, msg rpccapnp.Message) error {
	if !p.startSend() {
		return errClosed
	}
	defer p.finishSend()

	buf, err := msg.Segment().Message().Marshal()
	if err != nil {
		return err
	}
	mm, err := capnp.Unmarshal(buf)
	if err != nil {
		return err
	}
	msg, err = rpccapnp.ReadRootMessage(mm)
	if err != nil {
		return err
	}

	select {
	case p.w <- msg:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-p.finish:
		return errClosed
	case <-p.otherFin:
		return errBrokenPipe
	}
}

func (p *pipeTransport) startSend() bool {
	p.mu.Lock()
	ok := !p.done
	if ok {
		p.inflight++
	}
	p.mu.Unlock()
	return ok
}

func (p *pipeTransport) finishSend() {
	p.mu.Lock()
	p.inflight--
	p.mu.Unlock()
}

func (p *pipeTransport) RecvMessage(ctx context.Context) (rpccapnp.Message, error) {
	// Scribble over shared buffer to test for race conditions.
	for b, i := p.rbuf.Bytes(), 0; i < len(b); i++ {
		b[i] = 0xff
	}
	p.rbuf.Reset()

	select {
	case msg, ok := <-p.r:
		if !ok {
			return rpccapnp.Message{}, errBrokenPipe
		}
		if err := capnp.NewEncoder(&p.rbuf).Encode(msg.Segment().Message()); err != nil {
			return rpccapnp.Message{}, err
		}
		m, err := capnp.Unmarshal(p.rbuf.Bytes())
		if err != nil {
			return rpccapnp.Message{}, err
		}
		return rpccapnp.ReadRootMessage(m)
	case <-ctx.Done():
		return rpccapnp.Message{}, ctx.Err()
	}
}

func (p *pipeTransport) Close() error {
	p.mu.Lock()
	done := p.done
	if !done {
		p.done = true
		close(p.finish)
		if p.inflight == 0 {
			close(p.w)
		}
	}
	p.mu.Unlock()
	if done {
		return errClosed
	}
	return nil
}

var (
	errBrokenPipe = errors.New("pipetransport: broken pipe")
	errClosed     = errors.New("pipetransport: write to broken pipe")
)
