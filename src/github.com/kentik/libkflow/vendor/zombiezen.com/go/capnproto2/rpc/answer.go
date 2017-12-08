package rpc

import (
	"errors"
	"sync"

	"golang.org/x/net/context"
	"zombiezen.com/go/capnproto2"
	"zombiezen.com/go/capnproto2/internal/fulfiller"
	"zombiezen.com/go/capnproto2/internal/queue"
	rpccapnp "zombiezen.com/go/capnproto2/std/capnp/rpc"
)

// callQueueSize is the maximum number of calls that can be queued per answer or client.
// TODO(light): make this a ConnOption
const callQueueSize = 64

// insertAnswer creates a new answer with the given ID, returning nil
// if the ID is already in use.
func (c *Conn) insertAnswer(id answerID, cancel context.CancelFunc) *answer {
	if c.answers == nil {
		c.answers = make(map[answerID]*answer)
	} else if _, exists := c.answers[id]; exists {
		return nil
	}
	a := &answer{
		id:       id,
		cancel:   cancel,
		conn:     c,
		resolved: make(chan struct{}),
		queue:    make([]pcall, 0, callQueueSize),
	}
	c.answers[id] = a
	return a
}

func (c *Conn) popAnswer(id answerID) *answer {
	if c.answers == nil {
		return nil
	}
	a := c.answers[id]
	delete(c.answers, id)
	return a
}

type answer struct {
	id         answerID
	cancel     context.CancelFunc
	resultCaps []exportID
	conn       *Conn
	resolved   chan struct{}

	mu    sync.RWMutex
	obj   capnp.Ptr
	err   error
	done  bool
	queue []pcall
}

// fulfill is called to resolve an answer successfully.  It returns an
// error if its connection is shut down while sending messages.  The
// caller must be holding onto a.conn.mu.
func (a *answer) fulfill(obj capnp.Ptr) error {
	a.mu.Lock()
	if a.done {
		panic("answer.fulfill called more than once")
	}
	a.obj, a.done = obj, true
	// TODO(light): populate resultCaps

	retmsg := newReturnMessage(nil, a.id)
	ret, _ := retmsg.Return()
	payload, _ := ret.NewResults()
	payload.SetContentPtr(obj)
	var firstErr error
	if payloadTab, err := a.conn.makeCapTable(ret.Segment()); err == nil {
		payload.SetCapTable(payloadTab)
		if err := a.conn.sendMessage(retmsg); err != nil {
			firstErr = err
		}
	} else {
		firstErr = err
	}

	queues, err := a.emptyQueue(obj)
	if err != nil && firstErr == nil {
		firstErr = err
	}
	ctab := obj.Segment().Message().CapTable
	for capIdx, q := range queues {
		ctab[capIdx] = newQueueClient(a.conn, ctab[capIdx], q)
	}
	close(a.resolved)
	a.mu.Unlock()
	return firstErr
}

// reject is called to resolve an answer with failure.  It returns an
// error if its connection is shut down while sending messages.  The
// caller must be holding onto a.conn.mu.
func (a *answer) reject(err error) error {
	if err == nil {
		panic("answer.reject called with nil")
	}
	a.mu.Lock()
	if a.done {
		panic("answer.reject called more than once")
	}
	a.err, a.done = err, true
	m := newReturnMessage(nil, a.id)
	mret, _ := m.Return()
	setReturnException(mret, err)
	var firstErr error
	if err := a.conn.sendMessage(m); err != nil {
		firstErr = err
	}
	for i := range a.queue {
		if err := a.queue[i].a.reject(err); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	a.queue = nil
	close(a.resolved)
	a.mu.Unlock()
	return firstErr
}

// emptyQueue splits the queue by which capability it targets
// and drops any invalid calls.  Once this function returns, a.queue
// will be nil.
func (a *answer) emptyQueue(obj capnp.Ptr) (map[capnp.CapabilityID][]qcall, error) {
	var firstErr error
	qs := make(map[capnp.CapabilityID][]qcall, len(a.queue))
	for i, pc := range a.queue {
		c, err := capnp.TransformPtr(obj, pc.transform)
		if err != nil {
			if err := pc.a.reject(err); err != nil && firstErr == nil {
				firstErr = err
			}
			continue
		}
		ci := c.Interface()
		if !ci.IsValid() {
			if err := pc.a.reject(capnp.ErrNullClient); err != nil && firstErr == nil {
				firstErr = err
			}
			continue
		}
		cn := ci.Capability()
		if qs[cn] == nil {
			qs[cn] = make([]qcall, 0, len(a.queue)-i)
		}
		qs[cn] = append(qs[cn], pc.qcall)
	}
	a.queue = nil
	return qs, firstErr
}

// queueCallLocked enqueues a call to be made after the answer has been
// resolved.  The answer must not be resolved yet.  pc should have
// transform and one of pc.a or pc.f to be set.  The caller must be
// holding onto a.mu.
func (a *answer) queueCallLocked(call *capnp.Call, pc pcall) error {
	if len(a.queue) == cap(a.queue) {
		return errQueueFull
	}
	var err error
	pc.call, err = call.Copy(nil)
	if err != nil {
		return err
	}
	a.queue = append(a.queue, pc)
	return nil
}

// queueDisembargo enqueues a disembargo message.
func (a *answer) queueDisembargo(transform []capnp.PipelineOp, id embargoID, target rpccapnp.MessageTarget) (queued bool, err error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.done {
		return false, errDisembargoOngoingAnswer
	}
	if a.err != nil {
		return false, errDisembargoNonImport
	}
	targetPtr, err := capnp.TransformPtr(a.obj, transform)
	if err != nil {
		return false, err
	}
	client := targetPtr.Interface().Client()
	qc, ok := client.(*queueClient)
	if !ok {
		// No need to embargo, disembargo immediately.
		return false, nil
	}
	if ic := isImport(qc.client); ic == nil || a.conn != ic.conn {
		return false, errDisembargoNonImport
	}
	qc.mu.Lock()
	if !qc.isPassthrough() {
		err = qc.pushEmbargoLocked(id, target)
		if err == nil {
			queued = true
		}
	}
	qc.mu.Unlock()
	return queued, err
}

func (a *answer) pipelineClient(transform []capnp.PipelineOp) capnp.Client {
	return &localAnswerClient{a: a, transform: transform}
}

// joinAnswer resolves an RPC answer by waiting on a generic answer.
// The caller must not be holding onto a.conn.mu.
func joinAnswer(a *answer, ca capnp.Answer) {
	s, err := ca.Struct()
	select {
	case <-a.conn.mu:
		// Locked.
	case <-a.conn.bg.Done():
		return
	}
	if err == nil {
		a.fulfill(s.ToPtr())
	} else {
		a.reject(err)
	}
	a.conn.mu.Unlock()
}

// joinFulfiller resolves a fulfiller by waiting on a generic answer.
func joinFulfiller(f *fulfiller.Fulfiller, ca capnp.Answer) {
	s, err := ca.Struct()
	if err != nil {
		f.Reject(err)
	} else {
		f.Fulfill(s)
	}
}

type queueClient struct {
	client capnp.Client
	conn   *Conn

	mu    sync.RWMutex
	q     queue.Queue
	calls qcallList
}

func newQueueClient(c *Conn, client capnp.Client, queue []qcall) *queueClient {
	qc := &queueClient{
		client: client,
		conn:   c,
		calls:  make(qcallList, callQueueSize),
	}
	qc.q.Init(qc.calls, copy(qc.calls, queue))
	go qc.flushQueue()
	return qc
}

func (qc *queueClient) pushCallLocked(cl *capnp.Call) capnp.Answer {
	f := new(fulfiller.Fulfiller)
	cl, err := cl.Copy(nil)
	if err != nil {
		return capnp.ErrorAnswer(err)
	}
	i := qc.q.Push()
	if i == -1 {
		return capnp.ErrorAnswer(errQueueFull)
	}
	qc.calls[i] = qcall{call: cl, f: f}
	return f
}

func (qc *queueClient) pushEmbargoLocked(id embargoID, tgt rpccapnp.MessageTarget) error {
	i := qc.q.Push()
	if i == -1 {
		return errQueueFull
	}
	qc.calls[i] = qcall{embargoID: id, embargoTarget: tgt}
	return nil
}

// flushQueue is run in its own goroutine.
func (qc *queueClient) flushQueue() {
	var c qcall
	qc.mu.RLock()
	if i := qc.q.Front(); i != -1 {
		c = qc.calls[i]
	}
	qc.mu.RUnlock()
	for c.which() != qcallInvalid {
		qc.handle(&c)

		qc.mu.Lock()
		qc.q.Pop()
		if i := qc.q.Front(); i != -1 {
			c = qc.calls[i]
		} else {
			c = qcall{}
		}
		qc.mu.Unlock()
	}
}

func (qc *queueClient) handle(c *qcall) {
	switch c.which() {
	case qcallRemoteCall:
		answer := qc.client.Call(c.call)
		go joinAnswer(c.a, answer)
	case qcallLocalCall:
		answer := qc.client.Call(c.call)
		go joinFulfiller(c.f, answer)
	case qcallDisembargo:
		msg := newDisembargoMessage(nil, rpccapnp.Disembargo_context_Which_receiverLoopback, c.embargoID)
		d, _ := msg.Disembargo()
		d.SetTarget(c.embargoTarget)
		qc.conn.sendMessage(msg)
	}
}

func (qc *queueClient) isPassthrough() bool {
	return qc.q.Len() == 0
}

func (qc *queueClient) Call(cl *capnp.Call) capnp.Answer {
	// Fast path: queue is flushed.
	qc.mu.RLock()
	ok := qc.isPassthrough()
	qc.mu.RUnlock()
	if ok {
		return qc.client.Call(cl)
	}

	// Add to queue.
	qc.mu.Lock()
	// Since we released the lock, check that the queue hasn't been flushed.
	if qc.isPassthrough() {
		qc.mu.Unlock()
		return qc.client.Call(cl)
	}
	ans := qc.pushCallLocked(cl)
	qc.mu.Unlock()
	return ans
}

func (qc *queueClient) tryQueue(cl *capnp.Call) capnp.Answer {
	qc.mu.Lock()
	if qc.isPassthrough() {
		qc.mu.Unlock()
		return nil
	}
	ans := qc.pushCallLocked(cl)
	qc.mu.Unlock()
	return ans
}

func (qc *queueClient) Close() error {
	select {
	case <-qc.conn.mu:
		// Locked.
	case <-qc.conn.bg.Done():
		return ErrConnClosed
	}
	rejErr := qc.rejectQueue()
	qc.conn.mu.Unlock()
	if err := qc.client.Close(); err != nil {
		return err
	}
	return rejErr
}

// rejectQueue drains the client's queue.  It returns an error if the
// connection was shut down while messages are sent.  The caller must be
// holding onto qc.conn.mu.
func (qc *queueClient) rejectQueue() error {
	var firstErr error
	qc.mu.Lock()
	for ; qc.q.Len() > 0; qc.q.Pop() {
		c := qc.calls[qc.q.Front()]
		switch c.which() {
		case qcallRemoteCall:
			if err := c.a.reject(errQueueCallCancel); err != nil && firstErr == nil {
				firstErr = err
			}
		case qcallLocalCall:
			c.f.Reject(errQueueCallCancel)
		case qcallDisembargo:
			m := newDisembargoMessage(nil, rpccapnp.Disembargo_context_Which_receiverLoopback, c.embargoID)
			d, _ := m.Disembargo()
			d.SetTarget(c.embargoTarget)
			if err := qc.conn.sendMessage(m); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}
	qc.mu.Unlock()
	return firstErr
}

// pcall is a queued pipeline call.
type pcall struct {
	transform []capnp.PipelineOp
	qcall
}

// qcall is a queued call.
type qcall struct {
	// Calls
	a    *answer              // non-nil if remote call
	f    *fulfiller.Fulfiller // non-nil if local call
	call *capnp.Call

	// Disembargo
	embargoID     embargoID
	embargoTarget rpccapnp.MessageTarget
}

// Queued call types.
const (
	qcallInvalid = iota
	qcallRemoteCall
	qcallLocalCall
	qcallDisembargo
)

func (c *qcall) which() int {
	switch {
	case c.a != nil:
		return qcallRemoteCall
	case c.f != nil:
		return qcallLocalCall
	case c.embargoTarget.IsValid():
		return qcallDisembargo
	default:
		return qcallInvalid
	}
}

type qcallList []qcall

func (ql qcallList) Len() int {
	return len(ql)
}

func (ql qcallList) Clear(i int) {
	ql[i] = qcall{}
}

// A localAnswerClient is used to provide a pipelined client of an answer.
type localAnswerClient struct {
	a         *answer
	transform []capnp.PipelineOp
}

func (lac *localAnswerClient) Call(call *capnp.Call) capnp.Answer {
	lac.a.mu.Lock()
	if lac.a.done {
		obj, err := lac.a.obj, lac.a.err
		lac.a.mu.Unlock()
		return clientFromResolution(lac.transform, obj, err).Call(call)
	}
	f := new(fulfiller.Fulfiller)
	err := lac.a.queueCallLocked(call, pcall{
		transform: lac.transform,
		qcall:     qcall{f: f},
	})
	lac.a.mu.Unlock()
	if err != nil {
		return capnp.ErrorAnswer(errQueueFull)
	}
	return f
}

func (lac *localAnswerClient) Close() error {
	lac.a.mu.RLock()
	obj, err, done := lac.a.obj, lac.a.err, lac.a.done
	lac.a.mu.RUnlock()
	if !done {
		return nil
	}
	client := clientFromResolution(lac.transform, obj, err)
	return client.Close()
}

var (
	errQueueFull       = errors.New("rpc: pipeline queue full")
	errQueueCallCancel = errors.New("rpc: queued call canceled")

	errDisembargoOngoingAnswer = errors.New("rpc: disembargo attempted on in-progress answer")
	errDisembargoNonImport     = errors.New("rpc: disembargo attempted on non-import capability")
	errDisembargoMissingAnswer = errors.New("rpc: disembargo attempted on missing answer (finished too early?)")
)
