// Package rpc implements the Cap'n Proto RPC protocol.
package rpc // import "zombiezen.com/go/capnproto2/rpc"

import (
	"fmt"
	"io"
	"sync"

	"golang.org/x/net/context"
	"zombiezen.com/go/capnproto2"
	"zombiezen.com/go/capnproto2/rpc/internal/refcount"
	rpccapnp "zombiezen.com/go/capnproto2/std/capnp/rpc"
)

// A Conn is a connection to another Cap'n Proto vat.
// It is safe to use from multiple goroutines.
type Conn struct {
	transport  Transport
	log        Logger
	mainFunc   func(context.Context) (capnp.Client, error)
	mainCloser io.Closer

	out chan rpccapnp.Message

	bg       context.Context
	bgCancel context.CancelFunc
	workers  sync.WaitGroup

	// Mutable state protected by stateMu
	stateMu   sync.RWMutex
	stateCond sync.Cond // broadcasts when state changes
	state     connState
	closeErr  error

	// Mutable state protected by mu
	mu         chanMutex
	questions  []*question
	questionID idgen
	exports    []*export
	exportID   idgen
	embargoes  []chan<- struct{}
	embargoID  idgen
	answers    map[answerID]*answer
	imports    map[importID]*impent
}

type connParams struct {
	log            Logger
	mainFunc       func(context.Context) (capnp.Client, error)
	mainCloser     io.Closer
	sendBufferSize int
}

// A ConnOption is an option for opening a connection.
type ConnOption struct {
	f func(*connParams)
}

// MainInterface specifies that the connection should use client when
// receiving bootstrap messages.  By default, all bootstrap messages will
// fail.  The client will be closed when the connection is closed.
func MainInterface(client capnp.Client) ConnOption {
	rc, ref1 := refcount.New(client)
	ref2 := rc.Ref()
	return ConnOption{func(c *connParams) {
		c.mainFunc = func(ctx context.Context) (capnp.Client, error) {
			return ref1, nil
		}
		c.mainCloser = ref2
	}}
}

// BootstrapFunc specifies the function to call to create a capability
// for handling bootstrap messages.  This function should not make any
// RPCs or block.
func BootstrapFunc(f func(context.Context) (capnp.Client, error)) ConnOption {
	return ConnOption{func(c *connParams) {
		c.mainFunc = f
	}}
}

// SendBufferSize sets the number of outgoing messages to buffer on the
// connection.  This is in addition to whatever buffering the connection's
// transport performs.
func SendBufferSize(numMsgs int) ConnOption {
	return ConnOption{func(c *connParams) {
		c.sendBufferSize = numMsgs
	}}
}

// NewConn creates a new connection that communicates on c.
// Closing the connection will cause c to be closed.
func NewConn(t Transport, options ...ConnOption) *Conn {
	p := &connParams{
		log:            defaultLogger{},
		sendBufferSize: 4,
	}
	for _, o := range options {
		o.f(p)
	}

	conn := &Conn{
		transport:  t,
		out:        make(chan rpccapnp.Message, p.sendBufferSize),
		mainFunc:   p.mainFunc,
		mainCloser: p.mainCloser,
		log:        p.log,
		mu:         newChanMutex(),
	}
	conn.stateCond.L = conn.stateMu.RLocker()
	conn.bg, conn.bgCancel = context.WithCancel(context.Background())
	conn.workers.Add(2)
	go conn.dispatchRecv()
	go conn.dispatchSend()
	return conn
}

// Wait waits until the connection is closed or aborted by the remote vat.
// Wait will always return an error, usually ErrConnClosed or of type Abort.
func (c *Conn) Wait() error {
	c.stateMu.RLock()
	for c.state != connDead {
		c.stateCond.Wait()
	}
	err := c.closeErr
	c.stateMu.RUnlock()
	return err
}

// Close closes the connection and the underlying transport.
func (c *Conn) Close() error {
	c.stateMu.Lock()
	alive := c.state == connAlive
	if alive {
		c.bgCancel()
		c.closeErr = ErrConnClosed
		c.state = connDying
		c.stateCond.Broadcast()
	}
	c.stateMu.Unlock()
	if !alive {
		return ErrConnClosed
	}
	c.teardown(newAbortMessage(nil, errShutdown))
	c.stateMu.RLock()
	err := c.closeErr
	c.stateMu.RUnlock()
	if err != ErrConnClosed {
		return err
	}
	return nil
}

// shutdown cancels the background context and sets closeErr to e.
// No abort message will be sent on the transport.  After shutdown
// returns, the Conn will be in the dying or dead state.  Calling
// shutdown on a dying or dead Conn is a no-op.
func (c *Conn) shutdown(e error) {
	c.stateMu.Lock()
	if c.state == connAlive {
		c.bgCancel()
		c.closeErr = e
		c.state = connDying
		c.stateCond.Broadcast()
		go c.teardown(rpccapnp.Message{})
	}
	c.stateMu.Unlock()
}

// abort cancels the background context, sets closeErr to e, and queues
// an abort message to be sent on the transport before the Conn goes
// into the dead state.  After abort returns, the Conn will be in the
// dying or dead state.  Calling abort on a dying or dead Conn is a
// no-op.
func (c *Conn) abort(e error) {
	c.stateMu.Lock()
	if c.state == connAlive {
		c.bgCancel()
		c.closeErr = e
		c.state = connDying
		c.stateCond.Broadcast()
		go c.teardown(newAbortMessage(nil, e))
	}
	c.stateMu.Unlock()
}

// teardown moves the connection from the dying to the dead state.
func (c *Conn) teardown(abort rpccapnp.Message) {
	c.workers.Wait()

	c.mu.Lock()
	for _, q := range c.questions {
		if q != nil {
			q.cancel(ErrConnClosed)
		}
	}
	c.questions = nil
	exps := c.exports
	c.exports = nil
	c.embargoes = nil
	for _, a := range c.answers {
		a.cancel()
	}
	c.answers = nil
	c.imports = nil
	c.mainFunc = nil
	c.mu.Unlock()

	if c.mainCloser != nil {
		if err := c.mainCloser.Close(); err != nil {
			c.errorf("closing main interface: %v", err)
		}
		c.mainCloser = nil
	}
	// Closing an export may try to lock the Conn, so run it outside
	// critical section.
	for id, e := range exps {
		if e == nil {
			continue
		}
		if err := e.client.Close(); err != nil {
			c.errorf("export %v close: %v", id, err)
		}
	}
	exps = nil

	var werr error
	if abort.IsValid() {
		werr = c.transport.SendMessage(context.Background(), abort)
	}
	cerr := c.transport.Close()

	c.stateMu.Lock()
	if c.closeErr == ErrConnClosed {
		if cerr != nil {
			c.closeErr = cerr
		} else if werr != nil {
			c.closeErr = werr
		}
	}
	c.state = connDead
	c.stateCond.Broadcast()
	c.stateMu.Unlock()
}

// Bootstrap returns the receiver's main interface.
func (c *Conn) Bootstrap(ctx context.Context) capnp.Client {
	// TODO(light): Create a client that returns immediately.
	select {
	case <-c.mu:
		// Locked.
		defer c.mu.Unlock()
	case <-ctx.Done():
		return capnp.ErrorClient(ctx.Err())
	case <-c.bg.Done():
		return capnp.ErrorClient(ErrConnClosed)
	}

	q := c.newQuestion(ctx, nil /* method */)
	msg := newMessage(nil)
	boot, _ := msg.NewBootstrap()
	boot.SetQuestionId(uint32(q.id))
	// The mutex must be held while sending so that call order is preserved.
	// Worst case, this blocks until a message is sent on the transport.
	// Common case, this just adds to the channel queue.
	select {
	case c.out <- msg:
		q.start()
		return capnp.NewPipeline(q).Client()
	case <-ctx.Done():
		c.popQuestion(q.id)
		return capnp.ErrorClient(ctx.Err())
	case <-c.bg.Done():
		c.popQuestion(q.id)
		return capnp.ErrorClient(ErrConnClosed)
	}
}

// handleMessage is run from the receive goroutine to process a single
// message.  m cannot be held onto past the return of handleMessage, and
// c.mu is not held at the start of handleMessage.
func (c *Conn) handleMessage(m rpccapnp.Message) {
	switch m.Which() {
	case rpccapnp.Message_Which_unimplemented:
		// no-op for now to avoid feedback loop
	case rpccapnp.Message_Which_abort:
		a, err := copyAbort(m)
		if err != nil {
			c.errorf("decode abort: %v", err)
			// Keep going, since we're trying to abort anyway.
		}
		c.infof("abort: %v", a)
		c.shutdown(a)
	case rpccapnp.Message_Which_return:
		m = copyRPCMessage(m)
		c.mu.Lock()
		err := c.handleReturnMessage(m)
		c.mu.Unlock()

		if err != nil {
			c.errorf("handle return: %v", err)
		}
	case rpccapnp.Message_Which_finish:
		mfin, err := m.Finish()
		if err != nil {
			c.errorf("decode finish: %v", err)
			return
		}
		id := answerID(mfin.QuestionId())

		c.mu.Lock()
		a := c.popAnswer(id)
		if a == nil {
			c.mu.Unlock()
			c.errorf("finish called for unknown answer %d", id)
			return
		}
		a.cancel()
		if mfin.ReleaseResultCaps() {
			for _, id := range a.resultCaps {
				c.releaseExport(id, 1)
			}
		}
		c.mu.Unlock()
	case rpccapnp.Message_Which_bootstrap:
		boot, err := m.Bootstrap()
		if err != nil {
			c.errorf("decode bootstrap: %v", err)
			return
		}
		id := answerID(boot.QuestionId())

		c.mu.Lock()
		err = c.handleBootstrapMessage(id)
		c.mu.Unlock()

		if err != nil {
			c.errorf("handle bootstrap: %v", err)
		}
	case rpccapnp.Message_Which_call:
		m = copyRPCMessage(m)
		c.mu.Lock()
		err := c.handleCallMessage(m)
		c.mu.Unlock()

		if err != nil {
			c.errorf("handle call: %v", err)
		}
	case rpccapnp.Message_Which_release:
		rel, err := m.Release()
		if err != nil {
			c.errorf("decode release: %v", err)
			return
		}
		id := exportID(rel.Id())
		refs := int(rel.ReferenceCount())

		c.mu.Lock()
		c.releaseExport(id, refs)
		c.mu.Unlock()
	case rpccapnp.Message_Which_disembargo:
		m = copyRPCMessage(m)
		c.mu.Lock()
		err := c.handleDisembargoMessage(m)
		c.mu.Unlock()

		if err != nil {
			// Any failure in a disembargo is a protocol violation.
			c.abort(err)
		}
	default:
		c.infof("received unimplemented message, which = %v", m.Which())
		um := newUnimplementedMessage(nil, m)
		c.sendMessage(um)
	}
}

func newUnimplementedMessage(buf []byte, m rpccapnp.Message) rpccapnp.Message {
	n := newMessage(buf)
	n.SetUnimplemented(m)
	return n
}

func (c *Conn) fillParams(payload rpccapnp.Payload, cl *capnp.Call) error {
	params, err := cl.PlaceParams(payload.Segment())
	if err != nil {
		return err
	}
	if err := payload.SetContent(params); err != nil {
		return err
	}
	ctab, err := c.makeCapTable(payload.Segment())
	if err != nil {
		return err
	}
	if err := payload.SetCapTable(ctab); err != nil {
		return err
	}
	return nil
}

func transformToPromisedAnswer(s *capnp.Segment, answer rpccapnp.PromisedAnswer, transform []capnp.PipelineOp) error {
	opList, err := rpccapnp.NewPromisedAnswer_Op_List(s, int32(len(transform)))
	if err != nil {
		return err
	}
	for i, op := range transform {
		opList.At(i).SetGetPointerField(uint16(op.Field))
	}
	err = answer.SetTransform(opList)
	return err
}

// handleReturnMessage is to handle a received return message.
// The caller is holding onto c.mu.
func (c *Conn) handleReturnMessage(m rpccapnp.Message) error {
	ret, err := m.Return()
	if err != nil {
		return err
	}
	id := questionID(ret.AnswerId())
	q := c.popQuestion(id)
	if q == nil {
		return fmt.Errorf("received return for unknown question id=%d", id)
	}
	if ret.ReleaseParamCaps() {
		for _, id := range q.paramCaps {
			c.releaseExport(id, 1)
		}
	}
	q.mu.RLock()
	qstate := q.state
	q.mu.RUnlock()
	if qstate == questionCanceled {
		// We already sent the finish message.
		return nil
	}
	releaseResultCaps := true
	switch ret.Which() {
	case rpccapnp.Return_Which_results:
		releaseResultCaps = false
		results, err := ret.Results()
		if err != nil {
			return err
		}
		if err := c.populateMessageCapTable(results); err == errUnimplemented {
			um := newUnimplementedMessage(nil, m)
			c.sendMessage(um)
			return errUnimplemented
		} else if err != nil {
			c.abort(err)
			return err
		}
		content, err := results.ContentPtr()
		if err != nil {
			return err
		}
		q.fulfill(content)
	case rpccapnp.Return_Which_exception:
		exc, err := ret.Exception()
		if err != nil {
			return err
		}
		e := error(Exception{exc})
		if q.method != nil {
			e = &capnp.MethodError{
				Method: q.method,
				Err:    e,
			}
		} else {
			e = bootstrapError{e}
		}
		q.reject(e)
	case rpccapnp.Return_Which_canceled:
		err := &questionError{
			id:     id,
			method: q.method,
			err:    fmt.Errorf("receiver reported canceled"),
		}
		c.errorf("%v", err)
		q.reject(err)
		return nil
	default:
		um := newUnimplementedMessage(nil, m)
		c.sendMessage(um)
		return errUnimplemented
	}
	fin := newFinishMessage(nil, id, releaseResultCaps)
	c.sendMessage(fin)
	return nil
}

func newFinishMessage(buf []byte, questionID questionID, release bool) rpccapnp.Message {
	m := newMessage(buf)
	f, _ := m.NewFinish()
	f.SetQuestionId(uint32(questionID))
	f.SetReleaseResultCaps(release)
	return m
}

// populateMessageCapTable converts the descriptors in the payload into
// clients and sets it on the message the payload is a part of.
func (c *Conn) populateMessageCapTable(payload rpccapnp.Payload) error {
	msg := payload.Segment().Message()
	ctab, err := payload.CapTable()
	if err != nil {
		return err
	}
	for i, n := 0, ctab.Len(); i < n; i++ {
		desc := ctab.At(i)
		switch desc.Which() {
		case rpccapnp.CapDescriptor_Which_none:
			msg.AddCap(nil)
		case rpccapnp.CapDescriptor_Which_senderHosted:
			id := importID(desc.SenderHosted())
			client := c.addImport(id)
			msg.AddCap(client)
		case rpccapnp.CapDescriptor_Which_senderPromise:
			// We do the same thing as senderHosted, above. @kentonv suggested this on
			// issue #2; this let's messages be delivered properly, although it's a bit
			// of a hack, and as Kenton describes, it has some disadvantages:
			//
			// > * Apps sometimes want to wait for promise resolution, and to find out if
			// >   it resolved to an exception. You won't be able to provide that API. But,
			// >   usually, it isn't needed.
			// > * If the promise resolves to a capability hosted on the receiver,
			// >   messages sent to it will uselessly round-trip over the network
			// >   rather than being delivered locally.
			id := importID(desc.SenderPromise())
			client := c.addImport(id)
			msg.AddCap(client)
		case rpccapnp.CapDescriptor_Which_receiverHosted:
			id := exportID(desc.ReceiverHosted())
			e := c.findExport(id)
			if e == nil {
				return fmt.Errorf("rpc: capability table references unknown export ID %d", id)
			}
			msg.AddCap(e.rc.Ref())
		case rpccapnp.CapDescriptor_Which_receiverAnswer:
			recvAns, err := desc.ReceiverAnswer()
			if err != nil {
				return err
			}
			id := answerID(recvAns.QuestionId())
			a := c.answers[id]
			if a == nil {
				return fmt.Errorf("rpc: capability table references unknown answer ID %d", id)
			}
			recvTransform, err := recvAns.Transform()
			if err != nil {
				return err
			}
			transform := promisedAnswerOpsToTransform(recvTransform)
			msg.AddCap(a.pipelineClient(transform))
		default:
			c.errorf("unknown capability type %v", desc.Which())
			return errUnimplemented
		}
	}
	return nil
}

// makeCapTable converts the clients in the segment's message into capability descriptors.
func (c *Conn) makeCapTable(s *capnp.Segment) (rpccapnp.CapDescriptor_List, error) {
	msgtab := s.Message().CapTable
	t, err := rpccapnp.NewCapDescriptor_List(s, int32(len(msgtab)))
	if err != nil {
		return rpccapnp.CapDescriptor_List{}, nil
	}
	for i, client := range msgtab {
		desc := t.At(i)
		if client == nil {
			desc.SetNone()
			continue
		}
		c.descriptorForClient(desc, client)
	}
	return t, nil
}

// handleBootstrapMessage handles a received bootstrap message.
// The caller holds onto c.mu.
func (c *Conn) handleBootstrapMessage(id answerID) error {
	ctx, cancel := c.newContext()
	defer cancel()
	a := c.insertAnswer(id, cancel)
	if a == nil {
		// Question ID reused, error out.
		retmsg := newReturnMessage(nil, id)
		r, _ := retmsg.Return()
		setReturnException(r, errQuestionReused)
		return c.sendMessage(retmsg)
	}
	if c.mainFunc == nil {
		return a.reject(errNoMainInterface)
	}
	main, err := c.mainFunc(ctx)
	if err != nil {
		return a.reject(errNoMainInterface)
	}
	m := &capnp.Message{
		Arena:    capnp.SingleSegment(make([]byte, 0)),
		CapTable: []capnp.Client{main},
	}
	s, _ := m.Segment(0)
	in := capnp.NewInterface(s, 0)
	return a.fulfill(in.ToPtr())
}

// handleCallMessage handles a received call message.  It mutates the
// capability table of its parameter.  The caller holds onto c.mu.
func (c *Conn) handleCallMessage(m rpccapnp.Message) error {
	mcall, err := m.Call()
	if err != nil {
		return err
	}
	mt, err := mcall.Target()
	if err != nil {
		return err
	}
	if mt.Which() != rpccapnp.MessageTarget_Which_importedCap && mt.Which() != rpccapnp.MessageTarget_Which_promisedAnswer {
		um := newUnimplementedMessage(nil, m)
		return c.sendMessage(um)
	}
	mparams, err := mcall.Params()
	if err != nil {
		return err
	}
	if err := c.populateMessageCapTable(mparams); err == errUnimplemented {
		um := newUnimplementedMessage(nil, m)
		return c.sendMessage(um)
	} else if err != nil {
		c.abort(err)
		return err
	}
	ctx, cancel := c.newContext()
	id := answerID(mcall.QuestionId())
	a := c.insertAnswer(id, cancel)
	if a == nil {
		// Question ID reused, error out.
		c.abort(errQuestionReused)
		return errQuestionReused
	}
	meth := capnp.Method{
		InterfaceID: mcall.InterfaceId(),
		MethodID:    mcall.MethodId(),
	}
	paramContent, err := mparams.ContentPtr()
	if err != nil {
		return err
	}
	cl := &capnp.Call{
		Ctx:    ctx,
		Method: meth,
		Params: paramContent.Struct(),
	}
	if err := c.routeCallMessage(a, mt, cl); err != nil {
		return a.reject(err)
	}
	return nil
}

func (c *Conn) routeCallMessage(result *answer, mt rpccapnp.MessageTarget, cl *capnp.Call) error {
	switch mt.Which() {
	case rpccapnp.MessageTarget_Which_importedCap:
		id := exportID(mt.ImportedCap())
		e := c.findExport(id)
		if e == nil {
			return errBadTarget
		}
		answer := c.lockedCall(e.client, cl)
		go joinAnswer(result, answer)
	case rpccapnp.MessageTarget_Which_promisedAnswer:
		mpromise, err := mt.PromisedAnswer()
		if err != nil {
			return err
		}
		id := answerID(mpromise.QuestionId())
		if id == result.id {
			// Grandfather paradox.
			return errBadTarget
		}
		pa := c.answers[id]
		if pa == nil {
			return errBadTarget
		}
		mtrans, err := mpromise.Transform()
		if err != nil {
			return err
		}
		transform := promisedAnswerOpsToTransform(mtrans)
		pa.mu.Lock()
		if pa.done {
			obj, err := pa.obj, pa.err
			pa.mu.Unlock()
			client := clientFromResolution(transform, obj, err)
			answer := c.lockedCall(client, cl)
			go joinAnswer(result, answer)
		} else {
			err = pa.queueCallLocked(cl, pcall{transform: transform, qcall: qcall{a: result}})
			pa.mu.Unlock()
		}
		return err
	default:
		panic("unreachable")
	}
	return nil
}

func (c *Conn) handleDisembargoMessage(msg rpccapnp.Message) error {
	d, err := msg.Disembargo()
	if err != nil {
		return err
	}
	dtarget, err := d.Target()
	if err != nil {
		return err
	}
	switch d.Context().Which() {
	case rpccapnp.Disembargo_context_Which_senderLoopback:
		id := embargoID(d.Context().SenderLoopback())
		if dtarget.Which() != rpccapnp.MessageTarget_Which_promisedAnswer {
			return errDisembargoNonImport
		}
		dpa, err := dtarget.PromisedAnswer()
		if err != nil {
			return err
		}
		aid := answerID(dpa.QuestionId())
		a := c.answers[aid]
		if a == nil {
			return errDisembargoMissingAnswer
		}
		dtrans, err := dpa.Transform()
		if err != nil {
			return err
		}
		transform := promisedAnswerOpsToTransform(dtrans)
		queued, err := a.queueDisembargo(transform, id, dtarget)
		if err != nil {
			return err
		}
		if !queued {
			// There's nothing to embargo; everything's been delivered.
			resp := newDisembargoMessage(nil, rpccapnp.Disembargo_context_Which_receiverLoopback, id)
			rd, _ := resp.Disembargo()
			if err := rd.SetTarget(dtarget); err != nil {
				return err
			}
			c.sendMessage(resp)
		}
	case rpccapnp.Disembargo_context_Which_receiverLoopback:
		id := embargoID(d.Context().ReceiverLoopback())
		c.disembargo(id)
	default:
		um := newUnimplementedMessage(nil, msg)
		c.sendMessage(um)
	}
	return nil
}

// newDisembargoMessage creates a disembargo message.  Its target will be left blank.
func newDisembargoMessage(buf []byte, which rpccapnp.Disembargo_context_Which, id embargoID) rpccapnp.Message {
	msg := newMessage(buf)
	d, _ := msg.NewDisembargo()
	switch which {
	case rpccapnp.Disembargo_context_Which_senderLoopback:
		d.Context().SetSenderLoopback(uint32(id))
	case rpccapnp.Disembargo_context_Which_receiverLoopback:
		d.Context().SetReceiverLoopback(uint32(id))
	default:
		panic("unreachable")
	}
	return msg
}

// newContext creates a new context for a local call.
func (c *Conn) newContext() (context.Context, context.CancelFunc) {
	return context.WithCancel(c.bg)
}

func promisedAnswerOpsToTransform(list rpccapnp.PromisedAnswer_Op_List) []capnp.PipelineOp {
	n := list.Len()
	transform := make([]capnp.PipelineOp, 0, n)
	for i := 0; i < n; i++ {
		op := list.At(i)
		switch op.Which() {
		case rpccapnp.PromisedAnswer_Op_Which_getPointerField:
			transform = append(transform, capnp.PipelineOp{
				Field: op.GetPointerField(),
			})
		case rpccapnp.PromisedAnswer_Op_Which_noop:
			// no-op
		}
	}
	return transform
}

func newAbortMessage(buf []byte, err error) rpccapnp.Message {
	n := newMessage(buf)
	e, _ := n.NewAbort()
	toException(e, err)
	return n
}

func newReturnMessage(buf []byte, id answerID) rpccapnp.Message {
	retmsg := newMessage(buf)
	ret, _ := retmsg.NewReturn()
	ret.SetAnswerId(uint32(id))
	ret.SetReleaseParamCaps(false)
	return retmsg
}

func setReturnException(ret rpccapnp.Return, err error) rpccapnp.Exception {
	e, _ := rpccapnp.NewException(ret.Segment())
	toException(e, err)
	ret.SetException(e)
	return e
}

// clientFromResolution retrieves a client from a resolved question or
// answer by applying a transform.
func clientFromResolution(transform []capnp.PipelineOp, obj capnp.Ptr, err error) capnp.Client {
	if err != nil {
		return capnp.ErrorClient(err)
	}
	out, err := capnp.TransformPtr(obj, transform)
	if err != nil {
		return capnp.ErrorClient(err)
	}
	c := out.Interface().Client()
	if c == nil {
		return capnp.ErrorClient(capnp.ErrNullClient)
	}
	return c
}

func newMessage(buf []byte) rpccapnp.Message {
	_, s, err := capnp.NewMessage(capnp.SingleSegment(buf))
	if err != nil {
		panic(err)
	}
	m, err := rpccapnp.NewRootMessage(s)
	if err != nil {
		panic(err)
	}
	return m
}

// chanMutex is a mutex backed by a channel so that it can be used in a select.
// A receive is a lock and a send is an unlock.
type chanMutex chan struct{}

type connState int

const (
	connAlive connState = iota
	connDying
	connDead
)

func newChanMutex() chanMutex {
	mu := make(chanMutex, 1)
	mu <- struct{}{}
	return mu
}

func (mu chanMutex) Lock() {
	<-mu
}

func (mu chanMutex) TryLock(ctx context.Context) error {
	select {
	case <-mu:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (mu chanMutex) Unlock() {
	mu <- struct{}{}
}
