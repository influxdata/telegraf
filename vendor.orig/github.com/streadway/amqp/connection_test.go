// Copyright (c) 2016, Sean Treadway, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/streadway/amqp

// +build integration

package amqp

import (
	"crypto/tls"
	"net"
	"sync"
	"testing"
	"time"
)

func TestRequiredServerLocale(t *testing.T) {
	conn := integrationConnection(t, "AMQP 0-9-1 required server locale")
	requiredServerLocale := defaultLocale

	for _, locale := range conn.Locales {
		if locale == requiredServerLocale {
			return
		}
	}

	t.Fatalf("AMQP 0-9-1 server must support at least the %s locale, server sent the following locales: %#v", requiredServerLocale, conn.Locales)
}

func TestDefaultConnectionLocale(t *testing.T) {
	conn := integrationConnection(t, "client default locale")

	if conn.Config.Locale != defaultLocale {
		t.Fatalf("Expected default connection locale to be %s, is was: %s", defaultLocale, conn.Config.Locale)
	}
}

func TestChannelOpenOnAClosedConnectionFails(t *testing.T) {
	conn := integrationConnection(t, "channel on close")

	conn.Close()

	if _, err := conn.Channel(); err != ErrClosed {
		t.Fatalf("channel.open on a closed connection %#v is expected to fail", conn)
	}
}

// TestChannelOpenOnAClosedConnectionFails_ReleasesAllocatedChannel ensures the
// channel allocated is released if opening the channel fails.
func TestChannelOpenOnAClosedConnectionFails_ReleasesAllocatedChannel(t *testing.T) {
	conn := integrationConnection(t, "releases channel allocation")
	conn.Close()

	before := len(conn.channels)

	if _, err := conn.Channel(); err != ErrClosed {
		t.Fatalf("channel.open on a closed connection %#v is expected to fail", conn)
	}

	if len(conn.channels) != before {
		t.Fatalf("channel.open failed, but the allocated channel was not released")
	}
}

// TestRaceBetweenChannelAndConnectionClose ensures allocating a new channel
// does not race with shutting the connection down.
//
// See https://github.com/streadway/amqp/issues/251 - thanks to jmalloc for the
// test case.
func TestRaceBetweenChannelAndConnectionClose(t *testing.T) {
	defer time.AfterFunc(10*time.Second, func() { panic("Close deadlock") }).Stop()

	conn := integrationConnection(t, "allocation/shutdown race")

	go conn.Close()
	for i := 0; i < 10; i++ {
		go func() {
			ch, err := conn.Channel()
			if err == nil {
				ch.Close()
			}
		}()
	}
}

// TestRaceBetweenChannelShutdownAndSend ensures closing a channel
// (channel.shutdown) does not race with calling channel.send() from any other
// goroutines.
//
// See https://github.com/streadway/amqp/pull/253#issuecomment-292464811 for
// more details - thanks to jmalloc again.
func TestRaceBetweenChannelShutdownAndSend(t *testing.T) {
	defer time.AfterFunc(10*time.Second, func() { panic("Close deadlock") }).Stop()

	conn := integrationConnection(t, "channel close/send race")
	defer conn.Close()

	ch, _ := conn.Channel()

	go ch.Close()
	for i := 0; i < 10; i++ {
		go func() {
			// ch.Ack calls ch.send() internally.
			ch.Ack(42, false)
		}()
	}
}

func TestQueueDeclareOnAClosedConnectionFails(t *testing.T) {
	conn := integrationConnection(t, "queue declare on close")
	ch, _ := conn.Channel()

	conn.Close()

	if _, err := ch.QueueDeclare("an example", false, false, false, false, nil); err != ErrClosed {
		t.Fatalf("queue.declare on a closed connection %#v is expected to return ErrClosed, returned: %#v", conn, err)
	}
}

func TestConcurrentClose(t *testing.T) {
	const concurrency = 32

	conn := integrationConnection(t, "concurrent close")
	defer conn.Close()

	wg := sync.WaitGroup{}
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()

			err := conn.Close()

			if err == nil {
				t.Log("first concurrent close was successful")
				return
			}

			if err == ErrClosed {
				t.Log("later concurrent close were successful and returned ErrClosed")
				return
			}

			// BUG(st) is this really acceptable? we got a net.OpError before the
			// connection was marked as closed means a race condition between the
			// network connection and handshake state. It should be a package error
			// returned.
			if _, neterr := err.(*net.OpError); neterr {
				t.Logf("unknown net.OpError during close, ignoring: %+v", err)
				return
			}

			// A different/protocol error occurred indicating a race or missed condition
			if _, other := err.(*Error); other {
				t.Fatalf("Expected no error, or ErrClosed, or a net.OpError from conn.Close(), got %#v (%s) of type %T", err, err, err)
			}
		}()
	}
	wg.Wait()
}

// TestPlaintextDialTLS esnures amqp:// connections succeed when using DialTLS.
func TestPlaintextDialTLS(t *testing.T) {
	uri, err := ParseURI(integrationURLFromEnv())
	if err != nil {
		t.Fatalf("parse URI error: %s", err)
	}

	// We can only test when we have a plaintext listener
	if uri.Scheme != "amqp" {
		t.Skip("requires server listening for plaintext connections")
	}

	conn, err := DialTLS(uri.String(), &tls.Config{MinVersion: tls.VersionTLS12})
	if err != nil {
		t.Fatalf("unexpected dial error, got %v", err)
	}
	conn.Close()
}
