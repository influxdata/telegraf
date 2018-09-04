package pgproto3_test

import (
	"testing"

	"github.com/jackc/pgx/pgproto3"
)

func TestBackendReceiveInterrupted(t *testing.T) {
	t.Parallel()

	server := &interruptReader{}
	server.push([]byte{'Q', 0, 0, 0, 6})

	backend, err := pgproto3.NewBackend(server, nil)
	if err != nil {
		t.Fatal(err)
	}

	msg, err := backend.Receive()
	if err == nil {
		t.Fatal("expected err")
	}
	if msg != nil {
		t.Fatalf("did not expect msg, but %v", msg)
	}

	server.push([]byte{'I', 0})

	msg, err = backend.Receive()
	if err != nil {
		t.Fatal(err)
	}
	if msg, ok := msg.(*pgproto3.Query); !ok || msg.String != "I" {
		t.Fatalf("unexpected msg: %v", msg)
	}
}
