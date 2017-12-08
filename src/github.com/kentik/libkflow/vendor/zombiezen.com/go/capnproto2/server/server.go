// Package server provides runtime support for implementing Cap'n Proto
// interfaces locally.
package server // import "zombiezen.com/go/capnproto2/server"

import (
	"errors"
	"sort"
	"sync"

	"golang.org/x/net/context"
	"zombiezen.com/go/capnproto2"
	"zombiezen.com/go/capnproto2/internal/fulfiller"
)

// A Method describes a single method on a server object.
type Method struct {
	capnp.Method
	Impl        Func
	ResultsSize capnp.ObjectSize
}

// A Func is a function that implements a single method.
type Func func(ctx context.Context, options capnp.CallOptions, params, results capnp.Struct) error

// Closer is the interface that wraps the Close method.
type Closer interface {
	Close() error
}

// A server is a locally implemented interface.
type server struct {
	methods sortedMethods
	closer  Closer
	queue   chan *call
	stop    chan struct{}
	done    chan struct{}
}

// New returns a client that makes calls to a set of methods.
// If closer is nil then the client's Close is a no-op.  The server
// guarantees message delivery order by blocking each call on the
// return or acknowledgment of the previous call.  See the Ack function
// for more details.
func New(methods []Method, closer Closer) capnp.Client {
	s := &server{
		methods: make(sortedMethods, len(methods)),
		closer:  closer,
		queue:   make(chan *call),
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
	}
	copy(s.methods, methods)
	sort.Sort(s.methods)
	go s.dispatch()
	return s
}

// dispatch runs in its own goroutine.
func (s *server) dispatch() {
	defer close(s.done)
	for {
		select {
		case cl := <-s.queue:
			err := s.startCall(cl)
			if err != nil {
				cl.ans.Reject(err)
			}
		case <-s.stop:
			return
		}
	}
}

// startCall runs in the dispatch goroutine to start a call.
func (s *server) startCall(cl *call) error {
	_, out, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return err
	}
	results, err := capnp.NewRootStruct(out, cl.method.ResultsSize)
	if err != nil {
		return err
	}
	acksig := newAckSignal()
	opts := cl.Options.With([]capnp.CallOption{capnp.SetOptionValue(ackSignalKey, acksig)})
	go func() {
		err := cl.method.Impl(cl.Ctx, opts, cl.Params, results)
		if err == nil {
			cl.ans.Fulfill(results)
		} else {
			cl.ans.Reject(err)
		}
	}()
	select {
	case <-acksig.c:
	case <-cl.ans.Done():
		// Implementation functions may not call Ack, which is fine for
		// smaller functions.
	case <-cl.Ctx.Done():
		// Ideally, this would reject the answer immediately, but then you
		// would race with the implementation function.
	}
	return nil
}

func (s *server) Call(cl *capnp.Call) capnp.Answer {
	sm := s.methods.find(&cl.Method)
	if sm == nil {
		return capnp.ErrorAnswer(&capnp.MethodError{
			Method: &cl.Method,
			Err:    capnp.ErrUnimplemented,
		})
	}
	cl, err := cl.Copy(nil)
	if err != nil {
		return capnp.ErrorAnswer(err)
	}
	scall := newCall(cl, sm)
	select {
	case s.queue <- scall:
		return &scall.ans
	case <-s.stop:
		return capnp.ErrorAnswer(errClosed)
	case <-cl.Ctx.Done():
		return capnp.ErrorAnswer(cl.Ctx.Err())
	}
}

func (s *server) Close() error {
	close(s.stop)
	<-s.done
	if s.closer == nil {
		return nil
	}
	return s.closer.Close()
}

// Ack acknowledges delivery of a server call, allowing other methods
// to be called on the server.  It is intended to be used inside the
// implementation of a server function.  Calling Ack on options that
// aren't from a server method implementation is a no-op.
//
// Example:
//
//	func (my *myServer) MyMethod(call schema.MyServer_myMethod) error {
//		server.Ack(call.Options)
//		// ... do long-running operation ...
//		return nil
//	}
//
// Ack need not be the first call in a function nor is it required.
// Since the function's return is also an acknowledgment of delivery,
// short functions can return without calling Ack.  However, since
// clients will not return an Answer until the delivery is acknowledged,
// it is advisable to ack early.
func Ack(opts capnp.CallOptions) {
	if ack, _ := opts.Value(ackSignalKey).(*ackSignal); ack != nil {
		ack.signal()
	}
}

type call struct {
	*capnp.Call
	ans    fulfiller.Fulfiller
	method *Method
}

func newCall(cl *capnp.Call, sm *Method) *call {
	return &call{Call: cl, method: sm}
}

type sortedMethods []Method

// find returns the method with the given ID or nil.
func (sm sortedMethods) find(id *capnp.Method) *Method {
	i := sort.Search(len(sm), func(i int) bool {
		m := &sm[i]
		if m.InterfaceID != id.InterfaceID {
			return m.InterfaceID >= id.InterfaceID
		}
		return m.MethodID >= id.MethodID
	})
	if i == len(sm) {
		return nil
	}
	m := &sm[i]
	if m.InterfaceID != id.InterfaceID || m.MethodID != id.MethodID {
		return nil
	}
	return m
}

func (sm sortedMethods) Len() int {
	return len(sm)
}

func (sm sortedMethods) Less(i, j int) bool {
	if id1, id2 := sm[i].InterfaceID, sm[j].InterfaceID; id1 != id2 {
		return id1 < id2
	}
	return sm[i].MethodID < sm[j].MethodID
}

func (sm sortedMethods) Swap(i, j int) {
	sm[i], sm[j] = sm[j], sm[i]
}

type ackSignal struct {
	c    chan struct{}
	once sync.Once
}

func newAckSignal() *ackSignal {
	return &ackSignal{c: make(chan struct{})}
}

func (ack *ackSignal) signal() {
	ack.once.Do(func() {
		close(ack.c)
	})
}

// callOptionKey is the unexported key type for predefined options.
type callOptionKey int

// Predefined call options
const (
	ackSignalKey callOptionKey = iota + 1
)

var errClosed = errors.New("capnp: server closed")
