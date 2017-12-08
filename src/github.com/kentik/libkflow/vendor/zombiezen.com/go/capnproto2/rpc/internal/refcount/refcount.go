// Package refcount implements a reference-counting client.
package refcount

import (
	"errors"
	"runtime"
	"sync"

	"zombiezen.com/go/capnproto2"
)

// A RefCount will close its underlying client once all its references are closed.
type RefCount struct {
	Client capnp.Client

	mu   sync.Mutex
	refs int
}

// New creates a reference counter and the first client reference.
func New(c capnp.Client) (rc *RefCount, ref1 capnp.Client) {
	if rr, ok := c.(*Ref); ok {
		return rr.rc, rr.rc.Ref()
	}
	rc = &RefCount{Client: c, refs: 1}
	ref1 = rc.newRef()
	return
}

// Ref makes a new client reference.
func (rc *RefCount) Ref() capnp.Client {
	rc.mu.Lock()
	if rc.refs <= 0 {
		rc.mu.Unlock()
		return capnp.ErrorClient(errZeroRef)
	}
	rc.refs++
	rc.mu.Unlock()
	return rc.newRef()
}

func (rc *RefCount) newRef() *Ref {
	r := &Ref{rc: rc}
	runtime.SetFinalizer(r, (*Ref).Close)
	return r
}

func (rc *RefCount) call(cl *capnp.Call) capnp.Answer {
	// We lock here so that we can prevent the client from being closed
	// while we start the call.
	rc.mu.Lock()
	if rc.refs <= 0 {
		rc.mu.Unlock()
		return capnp.ErrorAnswer(errClosed)
	}
	ans := rc.Client.Call(cl)
	rc.mu.Unlock()
	return ans
}

// decref decreases the reference count by one, closing the Client if it reaches zero.
func (rc *RefCount) decref() error {
	shouldClose := false

	rc.mu.Lock()
	if rc.refs <= 0 {
		rc.mu.Unlock()
		return errClosed
	}
	rc.refs--
	if rc.refs == 0 {
		shouldClose = true
	}
	rc.mu.Unlock()

	if shouldClose {
		return rc.Client.Close()
	}
	return nil
}

var (
	errZeroRef = errors.New("rpc: Ref() called on zeroed refcount")
	errClosed  = errors.New("rpc: Close() called on closed client")
)

// A Ref is a single reference to a client wrapped by RefCount.
type Ref struct {
	rc   *RefCount
	once sync.Once
}

// Call makes a call on the underlying client.
func (r *Ref) Call(cl *capnp.Call) capnp.Answer {
	return r.rc.call(cl)
}

// Client returns the underlying client.
func (r *Ref) Client() capnp.Client {
	return r.rc.Client
}

// Close decrements the reference count.  Close will be called on
// finalization (i.e. garbage collection).
func (r *Ref) Close() error {
	var err error
	closed := false
	r.once.Do(func() {
		err = r.rc.decref()
		closed = true
	})
	if !closed {
		return errClosed
	}
	return err
}
