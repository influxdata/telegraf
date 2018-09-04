package memcached

import (
	"testing"

	"github.com/couchbase/gomemcached"
)

func TestMakeTapEvent(t *testing.T) {
	e := makeTapEvent(gomemcached.MCRequest{
		Opcode: gomemcached.TAP_MUTATION,
		Key:    []byte("hi"),
		Body:   []byte("world"),
		Cas:    0x4321000012340000,
	})
	if e.Cas != 0x4321000012340000 {
		t.Fatalf("Expected Cas to match")
	}
	e = makeTapEvent(gomemcached.MCRequest{
		Opcode: gomemcached.TAP_DELETE,
		Key:    []byte("hi"),
		Body:   []byte("world"),
		Cas:    0x9321000012340000,
	})
	if e.Cas != 0x9321000012340000 {
		t.Fatalf("Expected Cas to match")
	}
}
