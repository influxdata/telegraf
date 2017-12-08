package rpc

import (
	"sync"

	"golang.org/x/net/context"
	"zombiezen.com/go/capnproto2"
	"zombiezen.com/go/capnproto2/internal/fulfiller"
	"zombiezen.com/go/capnproto2/internal/queue"
	rpccapnp "zombiezen.com/go/capnproto2/std/capnp/rpc"
)

// newQuestion creates a new question with an unassigned ID.
func (c *Conn) newQuestion(ctx context.Context, method *capnp.Method) *question {
	id := questionID(c.questionID.next())
	q := &question{
		ctx:      ctx,
		conn:     c,
		method:   method,
		resolved: make(chan struct{}),
		id:       id,
	}
	// TODO(light): populate paramCaps
	if int(id) == len(c.questions) {
		c.questions = append(c.questions, q)
	} else {
		c.questions[id] = q
	}
	return q
}

func (c *Conn) findQuestion(id questionID) *question {
	if int(id) >= len(c.questions) {
		return nil
	}
	return c.questions[id]
}

func (c *Conn) popQuestion(id questionID) *question {
	q := c.findQuestion(id)
	if q == nil {
		return nil
	}
	c.questions[id] = nil
	c.questionID.remove(uint32(id))
	return q
}

type question struct {
	id        questionID
	ctx       context.Context
	conn      *Conn
	method    *capnp.Method // nil if this is bootstrap
	paramCaps []exportID
	resolved  chan struct{}

	// Protected by conn.mu
	derived [][]capnp.PipelineOp

	// Fields below are protected by mu.
	mu    sync.RWMutex
	obj   capnp.Ptr
	err   error
	state questionState
}

type questionState uint8

// Question states
const (
	questionInProgress questionState = iota
	questionResolved
	questionCanceled
)

// start signals that the question has been sent.
func (q *question) start() {
	go func() {
		select {
		case <-q.resolved:
			// Resolved naturally, nothing to do.
		case <-q.ctx.Done():
			select {
			case <-q.conn.mu:
				if q.cancel(q.ctx.Err()) {
					q.conn.sendMessage(newFinishMessage(nil, q.id, true /* release */))
				}
				q.conn.mu.Unlock()
			case <-q.resolved:
			case <-q.conn.bg.Done():
			}
		case <-q.conn.bg.Done():
			// TODO(light): connection should reject all questions on shutdown.
		}
	}()
}

// fulfill is called to resolve a question successfully.
// The caller must be holding onto q.conn.mu.
func (q *question) fulfill(obj capnp.Ptr) {
	ctab := obj.Segment().Message().CapTable
	visited := make([]bool, len(ctab))
	for _, d := range q.derived {
		tgt, err := capnp.TransformPtr(obj, d)
		if err != nil {
			continue
		}
		in := tgt.Interface()
		if !in.IsValid() {
			continue
		}
		if ic := isImport(in.Client()); ic != nil && ic.conn == q.conn {
			// Imported from remote vat.  Don't need to disembargo.
			continue
		}
		cn := in.Capability()
		if visited[cn] {
			continue
		}
		visited[cn] = true
		id, e := q.conn.newEmbargo()
		ctab[cn] = newEmbargoClient(ctab[cn], e, q.conn.bg.Done())
		m := newDisembargoMessage(nil, rpccapnp.Disembargo_context_Which_senderLoopback, id)
		dis, _ := m.Disembargo()
		mt, _ := dis.NewTarget()
		pa, _ := mt.NewPromisedAnswer()
		pa.SetQuestionId(uint32(q.id))
		transformToPromisedAnswer(m.Segment(), pa, d)
		mt.SetPromisedAnswer(pa)

		select {
		case q.conn.out <- m:
		case <-q.conn.bg.Done():
			// TODO(soon): perhaps just drop all embargoes in this case?
		}
	}

	q.mu.Lock()
	if q.state != questionInProgress {
		panic("question.fulfill called more than once")
	}
	q.obj, q.state = obj, questionResolved
	close(q.resolved)
	q.mu.Unlock()
}

// reject is called to resolve a question with failure.
// The caller must be holding onto q.conn.mu.
func (q *question) reject(err error) {
	if err == nil {
		panic("question.reject called with nil")
	}
	q.mu.Lock()
	if q.state != questionInProgress {
		panic("question.reject called more than once")
	}
	q.err = err
	q.state = questionResolved
	close(q.resolved)
	q.mu.Unlock()
}

// cancel is called to resolve a question with cancellation.
// The caller must be holding onto q.conn.mu.
func (q *question) cancel(err error) bool {
	if err == nil {
		panic("question.cancel called with nil")
	}
	q.mu.Lock()
	canceled := q.state == questionInProgress
	if canceled {
		q.err = err
		q.state = questionCanceled
		close(q.resolved)
	}
	q.mu.Unlock()
	return canceled
}

// addPromise records a returned capability as being used for a call.
// This is needed for determining embargoes upon resolution.  The
// caller must be holding onto q.conn.mu.
func (q *question) addPromise(transform []capnp.PipelineOp) {
	for _, d := range q.derived {
		if transformsEqual(transform, d) {
			return
		}
	}
	q.derived = append(q.derived, transform)
}

func transformsEqual(t, u []capnp.PipelineOp) bool {
	if len(t) != len(u) {
		return false
	}
	for i := range t {
		if t[i].Field != u[i].Field {
			return false
		}
	}
	return true
}

func (q *question) Struct() (capnp.Struct, error) {
	select {
	case <-q.resolved:
	case <-q.conn.bg.Done():
		return capnp.Struct{}, ErrConnClosed
	}
	q.mu.RLock()
	s, err := q.obj.Struct(), q.err
	q.mu.RUnlock()
	return s, err
}

func (q *question) PipelineCall(transform []capnp.PipelineOp, ccall *capnp.Call) capnp.Answer {
	select {
	case <-q.conn.mu:
	case <-ccall.Ctx.Done():
		return capnp.ErrorAnswer(ccall.Ctx.Err())
	case <-q.conn.bg.Done():
		return capnp.ErrorAnswer(ErrConnClosed)
	}
	ans := q.lockedPipelineCall(transform, ccall)
	q.conn.mu.Unlock()
	return ans
}

// lockedPipelineCall is equivalent to PipelineCall but assumes that the
// caller is already holding onto q.conn.mu.
func (q *question) lockedPipelineCall(transform []capnp.PipelineOp, ccall *capnp.Call) capnp.Answer {
	if q.conn.findQuestion(q.id) != q {
		// Question has been finished.  The call should happen as if it is
		// back in application code.
		q.mu.RLock()
		obj, err, state := q.obj, q.err, q.state
		q.mu.RUnlock()
		if state == questionInProgress {
			panic("question popped but not done")
		}
		client := clientFromResolution(transform, obj, err)
		return q.conn.lockedCall(client, ccall)
	}

	pipeq := q.conn.newQuestion(ccall.Ctx, &ccall.Method)
	msg := newMessage(nil)
	msgCall, _ := msg.NewCall()
	msgCall.SetQuestionId(uint32(pipeq.id))
	msgCall.SetInterfaceId(ccall.Method.InterfaceID)
	msgCall.SetMethodId(ccall.Method.MethodID)
	target, _ := msgCall.NewTarget()
	a, _ := target.NewPromisedAnswer()
	a.SetQuestionId(uint32(q.id))
	err := transformToPromisedAnswer(a.Segment(), a, transform)
	if err != nil {
		q.conn.popQuestion(pipeq.id)
		return capnp.ErrorAnswer(err)
	}
	payload, _ := msgCall.NewParams()
	if err := q.conn.fillParams(payload, ccall); err != nil {
		q.conn.popQuestion(q.id)
		return capnp.ErrorAnswer(err)
	}

	select {
	case q.conn.out <- msg:
	case <-ccall.Ctx.Done():
		q.conn.popQuestion(pipeq.id)
		return capnp.ErrorAnswer(ccall.Ctx.Err())
	case <-q.conn.bg.Done():
		q.conn.popQuestion(pipeq.id)
		return capnp.ErrorAnswer(ErrConnClosed)
	}
	q.addPromise(transform)
	pipeq.start()
	return pipeq
}

func (q *question) PipelineClose(transform []capnp.PipelineOp) error {
	<-q.resolved
	q.mu.RLock()
	obj, err := q.obj, q.err
	q.mu.RUnlock()
	if err != nil {
		return err
	}
	x, err := capnp.TransformPtr(obj, transform)
	if err != nil {
		return err
	}
	c := x.Interface().Client()
	if c == nil {
		return capnp.ErrNullClient
	}
	return c.Close()
}

// embargoClient is a client that waits until an embargo signal is
// received to deliver calls.
type embargoClient struct {
	cancel  <-chan struct{}
	client  capnp.Client
	embargo embargo

	mu    sync.RWMutex
	q     queue.Queue
	calls ecallList
}

func newEmbargoClient(client capnp.Client, e embargo, cancel <-chan struct{}) *embargoClient {
	ec := &embargoClient{
		client:  client,
		embargo: e,
		cancel:  cancel,
		calls:   make(ecallList, callQueueSize),
	}
	ec.q.Init(ec.calls, 0)
	go ec.flushQueue()
	return ec
}

func (ec *embargoClient) push(cl *capnp.Call) capnp.Answer {
	f := new(fulfiller.Fulfiller)
	cl, err := cl.Copy(nil)
	if err != nil {
		return capnp.ErrorAnswer(err)
	}
	i := ec.q.Push()
	if i == -1 {
		return capnp.ErrorAnswer(errQueueFull)
	}
	ec.calls[i] = ecall{cl, f}
	return f
}

func (ec *embargoClient) Call(cl *capnp.Call) capnp.Answer {
	// Fast path: queue is flushed.
	ec.mu.RLock()
	ok := ec.isPassthrough()
	ec.mu.RUnlock()
	if ok {
		return ec.client.Call(cl)
	}

	ec.mu.Lock()
	if ec.isPassthrough() {
		ec.mu.Unlock()
		return ec.client.Call(cl)
	}
	ans := ec.push(cl)
	ec.mu.Unlock()
	return ans
}

func (ec *embargoClient) tryQueue(cl *capnp.Call) capnp.Answer {
	ec.mu.Lock()
	if ec.isPassthrough() {
		ec.mu.Unlock()
		return nil
	}
	ans := ec.push(cl)
	ec.mu.Unlock()
	return ans
}

func (ec *embargoClient) isPassthrough() bool {
	select {
	case <-ec.embargo:
	default:
		return false
	}
	return ec.q.Len() == 0
}

func (ec *embargoClient) Close() error {
	ec.mu.Lock()
	for ; ec.q.Len() > 0; ec.q.Pop() {
		c := ec.calls[ec.q.Front()]
		c.f.Reject(errQueueCallCancel)
	}
	ec.mu.Unlock()
	return ec.client.Close()
}

// flushQueue is run in its own goroutine.
func (ec *embargoClient) flushQueue() {
	select {
	case <-ec.embargo:
	case <-ec.cancel:
		ec.mu.Lock()
		for ec.q.Len() > 0 {
			ec.q.Pop()
		}
		ec.mu.Unlock()
		return
	}
	var c ecall
	ec.mu.RLock()
	if i := ec.q.Front(); i != -1 {
		c = ec.calls[i]
	}
	ec.mu.RUnlock()
	for c.call != nil {
		ans := ec.client.Call(c.call)
		go joinFulfiller(c.f, ans)

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

type ecall struct {
	call *capnp.Call
	f    *fulfiller.Fulfiller
}

type ecallList []ecall

func (el ecallList) Len() int {
	return len(el)
}

func (el ecallList) Clear(i int) {
	el[i] = ecall{}
}
