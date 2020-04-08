package iotdevice

import (
	"bytes"
	"testing"

	"github.com/amenzhinsky/iothub/common"
)

func TestEventsMuxSub(t *testing.T) {
	mux := newEventsMux()
	sub := mux.sub()
	mux.Dispatch(&common.Message{
		Payload: []byte("hello"),
	})
	msg := <-sub.C()
	if !bytes.Equal(msg.Payload, []byte("hello")) {
		t.Fatalf("invalid payload = %v, want %v", msg.Payload, []byte("hello"))
	}
	mux.unsub(sub)
	mux.Dispatch(&common.Message{
		Payload: []byte("hello"),
	})
	if !isClosed(sub.C()) {
		t.Fatal("C is not closed after unsub")
	}
	if err := sub.Err(); err != nil {
		t.Fatal(err)
	}
}

func isClosed(ch <-chan *common.Message) bool {
	select {
	case _, ok := <-ch:
		return !ok
	default:
		return false
	}
}

func TestEventsMuxClose(t *testing.T) {
	mux := newEventsMux()
	sub := mux.sub()
	mux.close(ErrClosed)
	if err := sub.Err(); err != ErrClosed {
		t.Fatalf("closed mux sub err = %v, want %v", err, ErrClosed)
	}
}

func TestMethodMux(t *testing.T) {
	m := methodMux{}
	if err := m.handle("add", func(v map[string]interface{}) (map[string]interface{}, error) {
		v["b"] = 2
		return v, nil
	}); err != nil {
		t.Fatal(err)
	}
	defer m.remove("add")

	rc, data, err := m.Dispatch("add", []byte(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	if rc != 200 {
		t.Errorf("rc = %d, want %d", rc, 200)
	}
	w := []byte(`{"a":1,"b":2}`)
	if !bytes.Equal(data, w) {
		t.Errorf("data = %q, want %q", data, w)
	}
}
