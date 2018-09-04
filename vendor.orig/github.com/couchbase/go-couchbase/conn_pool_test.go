package couchbase

import (
	"errors"
	"io"
	"testing"
	"time"

	"github.com/couchbase/gomemcached"
	"github.com/couchbase/gomemcached/client"
)

type testT struct {
	closed bool
}

func (t testT) Read([]byte) (int, error) {
	return 0, io.EOF
}

func (t testT) Write([]byte) (int, error) {
	return 0, io.EOF
}

var errAlreadyClosed = errors.New("already closed")

func (t *testT) Close() error {
	if t.closed {
		return errAlreadyClosed
	}
	t.closed = true
	return nil
}

func testMkConn(h string, ah AuthHandler) (*memcached.Client, error) {
	return memcached.Wrap(&testT{})
}

func TestConnPool(t *testing.T) {
	cp := newConnectionPool("h", &basicAuth{}, 3, 6)
	cp.mkConn = testMkConn

	seenClients := map[*memcached.Client]bool{}

	// build some connections

	for i := 0; i < 5; i++ {
		sc, err := cp.Get()
		if err != nil {
			t.Fatalf("Error getting connection from pool: %v", err)
		}
		seenClients[sc] = true
	}

	if len(cp.connections) != 0 {
		t.Errorf("Expected 0 connections after gets, got %v",
			len(cp.connections))
	}

	// return them
	for k := range seenClients {
		cp.Return(k)
	}

	if len(cp.connections) != 3 {
		t.Errorf("Expected 3 connections after returning them, got %v",
			len(cp.connections))
	}

	// Try again.
	matched := 0
	grabbed := []*memcached.Client{}
	for i := 0; i < 5; i++ {
		sc, err := cp.Get()
		if err != nil {
			t.Fatalf("Error getting connection from pool: %v", err)
		}
		if seenClients[sc] {
			matched++
		}
		grabbed = append(grabbed, sc)
	}

	if matched != 3 {
		t.Errorf("Expected to match 3 conns, matched %v", matched)
	}

	for _, c := range grabbed {
		cp.Return(c)
	}

	// Connect write error.
	sc, err := cp.Get()
	if err != nil {
		t.Fatalf("Error getting a connection: %v", err)
	}
	err = sc.Transmit(&gomemcached.MCRequest{})
	if err == nil {
		t.Fatalf("Expected error sending a request")
	}
	if sc.IsHealthy() {
		t.Fatalf("Expected unhealthy connection")
	}
	cp.Return(sc)

	if len(cp.connections) != 2 {
		t.Errorf("Expected to have 2 conns, have %v", len(cp.connections))
	}

	err = cp.Close()
	if err != nil {
		t.Errorf("Expected clean close, got %v", err)
	}

	err = cp.Close()
	if err == nil {
		t.Errorf("Expected error on second pool close")
	}
}

func TestConnPoolSoonAvailable(t *testing.T) {
	defer func(d time.Duration) { ConnPoolAvailWaitTime = d }(ConnPoolAvailWaitTime)
	defer func() { ConnPoolCallback = nil }()

	m := map[string]int{}
	timings := []time.Duration{}
	ConnPoolCallback = func(host string, source string, start time.Time, err error) {
		m[source] = m[source] + 1
		timings = append(timings, time.Since(start))
	}

	cp := newConnectionPool("h", &basicAuth{}, 3, 4)
	cp.mkConn = testMkConn

	seenClients := map[*memcached.Client]bool{}

	// build some connections

	var aClient *memcached.Client
	for {
		sc, err := cp.GetWithTimeout(time.Millisecond)
		if err == ErrTimeout {
			break
		}
		if err != nil {
			t.Fatalf("Error getting connection from pool: %v", err)
		}
		aClient = sc
		seenClients[sc] = true
	}

	time.AfterFunc(time.Millisecond, func() { cp.Return(aClient) })

	ConnPoolAvailWaitTime = time.Second

	sc, err := cp.Get()
	if err != nil || sc != aClient {
		t.Errorf("Expected a successful connection, got %v/%v", sc, err)
	}

	// Try again, but let's close it while we're stuck in secondary wait
	time.AfterFunc(time.Millisecond, func() { cp.Close() })

	sc, err = cp.Get()
	if err != errClosedPool {
		t.Errorf("Expected a closed pool, got %v/%v", sc, err)
	}

	t.Logf("Callback report: %v, timings: %v", m, timings)
}

func TestConnPoolClosedFull(t *testing.T) {
	cp := newConnectionPool("h", &basicAuth{}, 3, 4)
	cp.mkConn = testMkConn

	seenClients := map[*memcached.Client]bool{}

	// build some connections

	for {
		sc, err := cp.GetWithTimeout(time.Millisecond)
		if err == ErrTimeout {
			break
		}
		if err != nil {
			t.Fatalf("Error getting connection from pool: %v", err)
		}
		seenClients[sc] = true
	}

	time.AfterFunc(2*time.Millisecond, func() { cp.Close() })

	sc, err := cp.Get()
	if err != errClosedPool {
		t.Errorf("Expected closed pool error after closed, got %v/%v", sc, err)
	}
}

func TestConnPoolWaitFull(t *testing.T) {
	cp := newConnectionPool("h", &basicAuth{}, 3, 4)
	cp.mkConn = testMkConn

	seenClients := map[*memcached.Client]bool{}

	// build some connections

	var aClient *memcached.Client
	for {
		sc, err := cp.GetWithTimeout(time.Millisecond)
		if err == ErrTimeout {
			break
		}
		if err != nil {
			t.Fatalf("Error getting connection from pool: %v", err)
		}
		aClient = sc
		seenClients[sc] = true
	}

	time.AfterFunc(2*time.Millisecond, func() { cp.Return(aClient) })

	sc, err := cp.Get()
	if err != nil || sc != aClient {
		t.Errorf("Expected a successful connection, got %v/%v", sc, err)
	}
}

func TestConnPoolWaitFailFull(t *testing.T) {
	cp := newConnectionPool("h", &basicAuth{}, 3, 4)
	cp.mkConn = testMkConn

	seenClients := map[*memcached.Client]bool{}

	// build some connections

	var aClient *memcached.Client
	for {
		sc, err := cp.GetWithTimeout(time.Millisecond)
		if err == ErrTimeout {
			break
		}
		if err != nil {
			t.Fatalf("Error getting connection from pool: %v", err)
		}
		aClient = sc
		seenClients[sc] = true
	}

	// causes failure
	aClient.Transmit(&gomemcached.MCRequest{})
	time.AfterFunc(2*time.Millisecond, func() { cp.Return(aClient) })

	sc, err := cp.Get()
	if err != nil || sc == aClient {
		t.Errorf("Expected a new successful connection, got %v/%v", sc, err)
	}
}

func TestConnPoolWaitDoubleFailFull(t *testing.T) {
	cp := newConnectionPool("h", &basicAuth{}, 3, 4)
	cp.mkConn = testMkConn

	seenClients := map[*memcached.Client]bool{}

	// build some connections

	var aClient *memcached.Client
	for {
		sc, err := cp.GetWithTimeout(time.Millisecond)
		if err == ErrTimeout {
			break
		}
		if err != nil {
			t.Fatalf("Error getting connection from pool: %v", err)
		}
		aClient = sc
		seenClients[sc] = true
	}

	cp.mkConn = func(h string, ah AuthHandler) (*memcached.Client, error) {
		return nil, io.EOF
	}

	// causes failure
	aClient.Transmit(&gomemcached.MCRequest{})
	time.AfterFunc(2*time.Millisecond, func() { cp.Return(aClient) })

	sc, err := cp.Get()
	if err != io.EOF {
		t.Errorf("Expected to fail getting a new connection, got %v/%v", sc, err)
	}
}

func TestConnPoolNil(t *testing.T) {
	var cp *connectionPool
	c, err := cp.Get()
	if err == nil {
		t.Errorf("Expected an error getting from nil, got %v", c)
	}

	// This just shouldn't error.
	cp.Return(c)
}

func TestConnPoolClosed(t *testing.T) {
	cp := newConnectionPool("h", &basicAuth{}, 3, 6)
	cp.mkConn = testMkConn
	c, err := cp.Get()
	if err != nil {
		t.Fatal(err)
	}
	cp.Close()

	// This should cause the connection to be closed
	cp.Return(c)
	if err = c.Close(); err != errAlreadyClosed {
		t.Errorf("Expected to close connection, wasn't closed (%v)", err)
	}

	sc, err := cp.Get()
	if err != errClosedPool {
		t.Errorf("Expected closed pool error after closed, got %v/%v", sc, err)
	}
}

func TestConnPoolCloseWrongPool(t *testing.T) {
	cp := newConnectionPool("h", &basicAuth{}, 3, 6)
	cp.mkConn = testMkConn
	c, err := cp.Get()
	if err != nil {
		t.Fatal(err)
	}
	cp.Close()

	// Return to a different pool.  Should still be OK.
	cp = newConnectionPool("h", &basicAuth{}, 3, 6)
	cp.mkConn = testMkConn
	c, err = cp.Get()
	if err != nil {
		t.Fatal(err)
	}
	cp.Close()

	cp.Return(c)
	if err = c.Close(); err != errAlreadyClosed {
		t.Errorf("Expected to close connection, wasn't closed (%v)", err)
	}
}

func TestConnPoolCloseNil(t *testing.T) {
	cp := newConnectionPool("h", &basicAuth{}, 3, 6)
	cp.mkConn = testMkConn
	c, err := cp.Get()
	if err != nil {
		t.Fatal(err)
	}
	cp.Close()

	cp = nil
	cp.Return(c)
	if err = c.Close(); err != errAlreadyClosed {
		t.Errorf("Expected to close connection, wasn't closed (%v)", err)
	}
}

func TestConnPoolStartTapFeed(t *testing.T) {
	var cp *connectionPool
	args := memcached.DefaultTapArguments()
	tf, err := cp.StartTapFeed(&args)
	if err != errNoPool {
		t.Errorf("Expected no pool error with no pool, got %v/%v", tf, err)
	}

	cp = newConnectionPool("h", &basicAuth{}, 3, 6)
	cp.mkConn = testMkConn

	tf, err = cp.StartTapFeed(&args)
	if err != io.EOF {
		t.Errorf("Expected to fail a tap feed with EOF, got %v/%v", tf, err)
	}

	cp.Close()
	tf, err = cp.StartTapFeed(&args)
	if err != errClosedPool {
		t.Errorf("Expected a closed pool, got %v/%v", tf, err)
	}
}

func BenchmarkBestCaseCPGet(b *testing.B) {
	cp := newConnectionPool("h", &basicAuth{}, 3, 6)
	cp.mkConn = testMkConn

	for i := 0; i < b.N; i++ {
		c, err := cp.Get()
		if err != nil {
			b.Fatalf("Error getting from pool: %v", err)
		}
		cp.Return(c)
	}
}
