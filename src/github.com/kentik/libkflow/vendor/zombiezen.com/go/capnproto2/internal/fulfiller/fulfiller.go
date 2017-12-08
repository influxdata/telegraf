// Package fulfiller provides a type that implements capnp.Answer that
// resolves by calling setter methods.
package fulfiller

import (
	"errors"
	"sync"

	"zombiezen.com/go/capnproto2"
	"zombiezen.com/go/capnproto2/internal/queue"
)

// callQueueSize is the maximum number of pending calls.
const callQueueSize = 64

// Fulfiller is a promise for a Struct.  The zero value is an unresolved
// answer.  A Fulfiller is considered to be resolved once Fulfill or
// Reject is called.  Calls to the Fulfiller will queue up until it is
// resolved.  A Fulfiller is safe to use from multiple goroutines.
type Fulfiller struct {
	once     sync.Once
	resolved chan struct{} // initialized by init()

	// Protected by mu
	mu     sync.RWMutex
	answer capnp.Answer
	queue  []pcall // initialized by init()
}

// init initializes the Fulfiller.  It is idempotent.
// Should be called for each method on Fulfiller.
func (f *Fulfiller) init() {
	f.once.Do(func() {
		f.resolved = make(chan struct{})
		f.queue = make([]pcall, 0, callQueueSize)
	})
}

// Fulfill sets the fulfiller's answer to s.  If there are queued
// pipeline calls, the capabilities on the struct will be embargoed
// until the queued calls finish.  Fulfill will panic if the fulfiller
// has already been resolved.
func (f *Fulfiller) Fulfill(s capnp.Struct) {
	f.init()
	f.mu.Lock()
	if f.answer != nil {
		f.mu.Unlock()
		panic("Fulfiller.Fulfill called more than once")
	}
	f.answer = capnp.ImmediateAnswer(s)
	queues := f.emptyQueue(s)
	ctab := s.Segment().Message().CapTable
	for capIdx, q := range queues {
		ctab[capIdx] = newEmbargoClient(ctab[capIdx], q)
	}
	close(f.resolved)
	f.mu.Unlock()
}

// emptyQueue splits the queue by which capability it targets and
// drops any invalid calls.  Once this function returns, f.queue will
// be nil.
func (f *Fulfiller) emptyQueue(s capnp.Struct) map[capnp.CapabilityID][]ecall {
	qs := make(map[capnp.CapabilityID][]ecall, len(f.queue))
	for i, pc := range f.queue {
		c, err := capnp.TransformPtr(s.ToPtr(), pc.transform)
		if err != nil {
			pc.f.Reject(err)
			continue
		}
		in := c.Interface()
		if !in.IsValid() {
			pc.f.Reject(capnp.ErrNullClient)
			continue
		}
		cn := in.Capability()
		if qs[cn] == nil {
			qs[cn] = make([]ecall, 0, len(f.queue)-i)
		}
		qs[cn] = append(qs[cn], pc.ecall)
	}
	f.queue = nil
	return qs
}

// Reject sets the fulfiller's answer to err.  If there are queued
// pipeline calls, they will all return errors.  Reject will panic if
// the error is nil or the fulfiller has already been resolved.
func (f *Fulfiller) Reject(err error) {
	if err == nil {
		panic("Fulfiller.Reject called with nil")
	}
	f.init()
	f.mu.Lock()
	if f.answer != nil {
		f.mu.Unlock()
		panic("Fulfiller.Reject called more than once")
	}
	f.answer = capnp.ErrorAnswer(err)
	for i := range f.queue {
		f.queue[i].f.Reject(err)
		f.queue[i] = pcall{}
	}
	close(f.resolved)
	f.mu.Unlock()
}

// Done returns a channel that is closed once f is resolved.
func (f *Fulfiller) Done() <-chan struct{} {
	f.init()
	return f.resolved
}

// Peek returns f's resolved answer or nil if f has not been resolved.
// The Struct method of an answer returned from Peek returns immediately.
func (f *Fulfiller) Peek() capnp.Answer {
	f.init()
	f.mu.RLock()
	a := f.answer
	f.mu.RUnlock()
	return a
}

// Struct waits until f is resolved and returns its struct if fulfilled
// or an error if rejected.
func (f *Fulfiller) Struct() (capnp.Struct, error) {
	<-f.Done()
	return f.Peek().Struct()
}

// PipelineCall calls PipelineCall on the fulfilled answer or queues the
// call if f has not been fulfilled.
func (f *Fulfiller) PipelineCall(transform []capnp.PipelineOp, call *capnp.Call) capnp.Answer {
	f.init()

	// Fast path: pass-through after fulfilled.
	if a := f.Peek(); a != nil {
		return a.PipelineCall(transform, call)
	}

	f.mu.Lock()
	// Make sure that f wasn't fulfilled.
	if a := f.answer; a != nil {
		f.mu.Unlock()
		return a.PipelineCall(transform, call)
	}
	if len(f.queue) == cap(f.queue) {
		f.mu.Unlock()
		return capnp.ErrorAnswer(errCallQueueFull)
	}
	cc, err := call.Copy(nil)
	if err != nil {
		f.mu.Unlock()
		return capnp.ErrorAnswer(err)
	}
	g := new(Fulfiller)
	f.queue = append(f.queue, pcall{
		transform: transform,
		ecall: ecall{
			call: cc,
			f:    g,
		},
	})
	f.mu.Unlock()
	return g
}

// PipelineClose waits until f is resolved and then calls PipelineClose
// on the fulfilled answer.
func (f *Fulfiller) PipelineClose(transform []capnp.PipelineOp) error {
	<-f.Done()
	return f.Peek().PipelineClose(transform)
}

// pcall is a queued pipeline call.
type pcall struct {
	transform []capnp.PipelineOp
	ecall
}

// EmbargoClient is a client that flushes a queue of calls.
// Fulfiller will create these automatically when pipelined calls are
// made on unresolved answers.  EmbargoClient is exported so that rpc
// can avoid making calls on its own Conn.
type EmbargoClient struct {
	client capnp.Client

	mu    sync.RWMutex
	q     queue.Queue
	calls ecallList
}

func newEmbargoClient(client capnp.Client, queue []ecall) capnp.Client {
	ec := &EmbargoClient{
		client: client,
		calls:  make(ecallList, callQueueSize),
	}
	ec.q.Init(ec.calls, copy(ec.calls, queue))
	go ec.flushQueue()
	return ec
}

func (ec *EmbargoClient) push(cl *capnp.Call) capnp.Answer {
	f := new(Fulfiller)
	cl, err := cl.Copy(nil)
	if err != nil {
		return capnp.ErrorAnswer(err)
	}
	i := ec.q.Push()
	if i == -1 {
		return capnp.ErrorAnswer(errCallQueueFull)
	}
	ec.calls[i] = ecall{cl, f}
	return f
}

// flushQueue is run in its own goroutine.
func (ec *EmbargoClient) flushQueue() {
	var c ecall
	ec.mu.Lock()
	if i := ec.q.Front(); i != -1 {
		c = ec.calls[i]
	}
	ec.mu.Unlock()
	for c.call != nil {
		ans := ec.client.Call(c.call)
		go func(f *Fulfiller, ans capnp.Answer) {
			s, err := ans.Struct()
			if err == nil {
				f.Fulfill(s)
			} else {
				f.Reject(err)
			}
		}(c.f, ans)
		ec.mu.Lock()
		ec.q.Pop()
		if i := ec.q.Front(); i != -1 {
			c = ec.calls[i]
		} else {
			c = ecall{}
		}
		ec.mu.Unlock()
	}
}

// Client returns the underlying client if the embargo has been lifted
// and nil otherwise.
func (ec *EmbargoClient) Client() capnp.Client {
	ec.mu.RLock()
	ok := ec.isPassthrough()
	ec.mu.RUnlock()
	if !ok {
		return nil
	}
	return ec.client
}

func (ec *EmbargoClient) isPassthrough() bool {
	return ec.q.Len() == 0
}

// Call either queues a call to the underlying client or starts a call
// if the embargo has been lifted.
func (ec *EmbargoClient) Call(cl *capnp.Call) capnp.Answer {
	// Fast path: queue is flushed.
	ec.mu.RLock()
	ok := ec.isPassthrough()
	ec.mu.RUnlock()
	if ok {
		return ec.client.Call(cl)
	}

	// Add to queue.
	ec.mu.Lock()
	// Since we released the lock, check that the queue hasn't been flushed.
	if ec.isPassthrough() {
		ec.mu.Unlock()
		return ec.client.Call(cl)
	}
	ans := ec.push(cl)
	ec.mu.Unlock()
	return ans
}

// TryQueue will attempt to queue a call or return nil if the embargo
// has been lifted.
func (ec *EmbargoClient) TryQueue(cl *capnp.Call) capnp.Answer {
	ec.mu.Lock()
	if ec.isPassthrough() {
		ec.mu.Unlock()
		return nil
	}
	ans := ec.push(cl)
	ec.mu.Unlock()
	return ans
}

// Close closes the underlying client, rejecting any queued calls.
func (ec *EmbargoClient) Close() error {
	ec.mu.Lock()
	// reject all queued calls
	for ec.q.Len() > 0 {
		ec.calls[ec.q.Front()].f.Reject(errQueueCallCancel)
		ec.q.Pop()
	}
	ec.mu.Unlock()
	return ec.client.Close()
}

// ecall is an queued embargoed call.
type ecall struct {
	call *capnp.Call
	f    *Fulfiller
}

type ecallList []ecall

func (el ecallList) Len() int {
	return len(el)
}

func (el ecallList) Clear(i int) {
	el[i] = ecall{}
}

var (
	errCallQueueFull   = errors.New("capnp: promised answer call queue full")
	errQueueCallCancel = errors.New("capnp: queued call canceled")
)
