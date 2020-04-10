package iotdevice

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/amenzhinsky/iothub/common"
)

// once wraps a function that can return an error and
// executes it only once, all sequential calls return nils.
func once(on *sync.Once, fn func() error) error {
	var err error
	on.Do(func() {
		err = fn()
	})
	return err
}

func newEventsMux() *eventsMux {
	return &eventsMux{done: make(chan struct{})}
}

type eventsMux struct {
	on   sync.Once
	mu   sync.RWMutex
	subs []*EventSub
	done chan struct{}
}

func (m *eventsMux) once(fn func() error) error {
	return once(&m.on, fn)
}

func (m *eventsMux) Dispatch(msg *common.Message) {
	m.mu.RLock()
	for _, s := range m.subs {
		//go func() {
		select {
		case <-s.done:
		case <-m.done:
		case s.ch <- msg:
		}
		//}()
	}
	m.mu.RUnlock()
}

func (m *eventsMux) sub() *EventSub {
	s := newEventSub()
	m.mu.Lock()
	m.subs = append(m.subs, s)
	m.mu.Unlock()
	return s
}

func (m *eventsMux) unsub(s *EventSub) {
	m.mu.Lock()
	for i, ss := range m.subs {
		if ss == s {
			s.close(nil)
			m.subs = append(m.subs[:i], m.subs[i+1:]...)
			break
		}
	}
	m.mu.Unlock()
}

func (m *eventsMux) close(err error) {
	m.mu.Lock()
	select {
	case <-m.done:
		panic("already closed")
	default:
	}
	close(m.done)
	for _, s := range m.subs {
		s.close(ErrClosed)
	}
	m.subs = m.subs[0:0]
	m.mu.Unlock()
}

func newEventSub() *EventSub {
	return &EventSub{
		ch:   make(chan *common.Message, 10), // TODO: configurable value
		done: make(chan struct{}),
	}
}

type EventSub struct {
	ch   chan *common.Message
	err  error
	done chan struct{}
}

func (s *EventSub) C() <-chan *common.Message {
	return s.ch
}

func (s *EventSub) Err() error {
	return s.err
}

func (s *EventSub) close(err error) {
	s.err = err
	close(s.done)
	close(s.ch)
}

func newTwinStateMux() *twinStateMux {
	return &twinStateMux{done: make(chan struct{})}
}

type twinStateMux struct {
	on   sync.Once
	mu   sync.RWMutex
	subs []*TwinStateSub
	done chan struct{}
}

func (m *twinStateMux) once(fn func() error) error {
	return once(&m.on, fn)
}

func (m *twinStateMux) Dispatch(b []byte) {
	var v TwinState
	if err := json.Unmarshal(b, &v); err != nil {
		log.Printf("unmarshal error: %s", err) // TODO
		return
	}

	m.mu.RLock()
	select {
	case <-m.done:
		panic("already closed")
	default:
	}
	for _, sub := range m.subs {
		go func(sub *TwinStateSub) {
			select {
			case sub.ch <- v:
			case <-m.done:
			}
		}(sub)
	}
	m.mu.RUnlock()
}

func (m *twinStateMux) sub() *TwinStateSub {
	s := &TwinStateSub{ch: make(chan TwinState, 10)}
	m.mu.Lock()
	m.subs = append(m.subs, s)
	m.mu.Unlock()
	return s
}

func (m *twinStateMux) unsub(s *TwinStateSub) {
	m.mu.Lock()
	for i, ss := range m.subs {
		if ss == s {
			m.subs = append(m.subs[:i], m.subs[i+1:]...)
			break
		}
	}
	m.mu.Unlock()
}

func (m *twinStateMux) close(err error) {
	m.mu.Lock()
	for _, s := range m.subs {
		s.err = ErrClosed
		close(s.ch)
	}
	m.subs = m.subs[0:0]
	m.mu.Unlock()
}

type TwinStateSub struct {
	ch  chan TwinState
	err error
}

func (s *TwinStateSub) C() <-chan TwinState {
	return s.ch
}

func (s *TwinStateSub) Err() error {
	return s.err
}

func newMethodMux() *methodMux {
	return &methodMux{}
}

// methodMux is direct-methods dispatcher.
type methodMux struct {
	on sync.Once
	mu sync.RWMutex
	m  map[string]DirectMethodHandler
}

func (m *methodMux) once(fn func() error) error {
	return once(&m.on, fn)
}

// handle registers the given direct-method handler.
func (m *methodMux) handle(method string, fn DirectMethodHandler) error {
	if fn == nil {
		panic("fn is nil")
	}
	m.mu.Lock()
	if m.m == nil {
		m.m = map[string]DirectMethodHandler{}
	}
	if _, ok := m.m[method]; ok {
		m.mu.Unlock()
		return fmt.Errorf("method %q is already registered", method)
	}
	m.m[method] = fn
	m.mu.Unlock()
	return nil
}

// remove deregisters the named method.
func (m *methodMux) remove(method string) {
	m.mu.Lock()
	if m.m != nil {
		delete(m.m, method)
	}
	m.mu.Unlock()
}

// Dispatch dispatches the named method, error is not nil only when dispatching fails.
func (m *methodMux) Dispatch(method string, b []byte) (int, []byte, error) {
	m.mu.RLock()
	f, ok := m.m[method]
	m.mu.RUnlock()
	if !ok {
		return 0, nil, fmt.Errorf("method %q is not registered", method)
	}

	var v map[string]interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return jsonErr(err)
	}
	v, err := f(v)
	if err != nil {
		return jsonErr(err)
	}
	if v == nil {
		v = map[string]interface{}{}
	}
	b, err = json.Marshal(v)
	if err != nil {
		return jsonErr(err)
	}
	return 200, b, nil
}

func jsonErr(err error) (int, []byte, error) {
	return 500, []byte(fmt.Sprintf(`{"error":%q}`, err.Error())), nil
}
