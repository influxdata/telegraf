// Copyright (c) 2012, Sean Treadway, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/streadway/amqp

package amqp

import (
	"bytes"
	"io"
	"reflect"
	"testing"
	"time"
)

type server struct {
	*testing.T
	r reader             // framer <- client
	w writer             // framer -> client
	S io.ReadWriteCloser // Server IO
	C io.ReadWriteCloser // Client IO

	// captured client frames
	start connectionStartOk
	tune  connectionTuneOk
}

func defaultConfig() Config {
	return Config{
		SASL:   []Authentication{&PlainAuth{"guest", "guest"}},
		Vhost:  "/",
		Locale: defaultLocale,
	}
}

func newSession(t *testing.T) (io.ReadWriteCloser, *server) {
	rs, wc := io.Pipe()
	rc, ws := io.Pipe()

	rws := &logIO{t, "server", pipe{rs, ws}}
	rwc := &logIO{t, "client", pipe{rc, wc}}

	server := server{
		T: t,
		r: reader{rws},
		w: writer{rws},
		S: rws,
		C: rwc,
	}

	return rwc, &server
}

func (t *server) expectBytes(b []byte) {
	in := make([]byte, len(b))
	if _, err := io.ReadFull(t.S, in); err != nil {
		t.Fatalf("io error expecting bytes: %v", err)
	}

	if bytes.Compare(b, in) != 0 {
		t.Fatalf("failed bytes: expected: %s got: %s", string(b), string(in))
	}
}

func (t *server) send(channel int, m message) {
	defer time.AfterFunc(time.Second, func() { panic("send deadlock") }).Stop()

	if msg, ok := m.(messageWithContent); ok {
		props, body := msg.getContent()
		class, _ := msg.id()
		t.w.WriteFrame(&methodFrame{
			ChannelId: uint16(channel),
			Method:    msg,
		})
		t.w.WriteFrame(&headerFrame{
			ChannelId:  uint16(channel),
			ClassId:    class,
			Size:       uint64(len(body)),
			Properties: props,
		})
		t.w.WriteFrame(&bodyFrame{
			ChannelId: uint16(channel),
			Body:      body,
		})
	} else {
		t.w.WriteFrame(&methodFrame{
			ChannelId: uint16(channel),
			Method:    m,
		})
	}
}

// drops all but method frames expected on the given channel
func (t *server) recv(channel int, m message) message {
	defer time.AfterFunc(time.Second, func() { panic("recv deadlock") }).Stop()

	var remaining int
	var header *headerFrame
	var body []byte

	for {
		frame, err := t.r.ReadFrame()
		if err != nil {
			t.Fatalf("frame err, read: %s", err)
		}

		if frame.channel() != uint16(channel) {
			t.Fatalf("expected frame on channel %d, got channel %d", channel, frame.channel())
		}

		switch f := frame.(type) {
		case *heartbeatFrame:
			// drop

		case *headerFrame:
			// start content state
			header = f
			remaining = int(header.Size)
			if remaining == 0 {
				m.(messageWithContent).setContent(header.Properties, nil)
				return m
			}

		case *bodyFrame:
			// continue until terminated
			body = append(body, f.Body...)
			remaining -= len(f.Body)
			if remaining <= 0 {
				m.(messageWithContent).setContent(header.Properties, body)
				return m
			}

		case *methodFrame:
			if reflect.TypeOf(m) == reflect.TypeOf(f.Method) {
				wantv := reflect.ValueOf(m).Elem()
				havev := reflect.ValueOf(f.Method).Elem()
				wantv.Set(havev)
				if _, ok := m.(messageWithContent); !ok {
					return m
				}
			} else {
				t.Fatalf("expected method type: %T, got: %T", m, f.Method)
			}

		default:
			t.Fatalf("unexpected frame: %+v", f)
		}
	}
}

func (t *server) expectAMQP() {
	t.expectBytes([]byte{'A', 'M', 'Q', 'P', 0, 0, 9, 1})
}

func (t *server) connectionStart() {
	t.send(0, &connectionStart{
		VersionMajor: 0,
		VersionMinor: 9,
		Mechanisms:   "PLAIN",
		Locales:      "en_US",
	})

	t.recv(0, &t.start)
}

func (t *server) connectionTune() {
	t.send(0, &connectionTune{
		ChannelMax: 11,
		FrameMax:   20000,
		Heartbeat:  10,
	})

	t.recv(0, &t.tune)
}

func (t *server) connectionOpen() {
	t.expectAMQP()
	t.connectionStart()
	t.connectionTune()

	t.recv(0, &connectionOpen{})
	t.send(0, &connectionOpenOk{})
}

func (t *server) connectionClose() {
	t.recv(0, &connectionClose{})
	t.send(0, &connectionCloseOk{})
}

func (t *server) channelOpen(id int) {
	t.recv(id, &channelOpen{})
	t.send(id, &channelOpenOk{})
}

func TestDefaultClientProperties(t *testing.T) {
	rwc, srv := newSession(t)

	go func() {
		srv.connectionOpen()
		rwc.Close()
	}()

	if c, err := Open(rwc, defaultConfig()); err != nil {
		t.Fatalf("could not create connection: %v (%s)", c, err)
	}

	if want, got := defaultProduct, srv.start.ClientProperties["product"]; want != got {
		t.Errorf("expected product %s got: %s", want, got)
	}

	if want, got := defaultVersion, srv.start.ClientProperties["version"]; want != got {
		t.Errorf("expected version %s got: %s", want, got)
	}

	if want, got := defaultLocale, srv.start.Locale; want != got {
		t.Errorf("expected locale %s got: %s", want, got)
	}
}

func TestCustomClientProperties(t *testing.T) {
	rwc, srv := newSession(t)

	config := defaultConfig()
	config.Properties = Table{
		"product": "foo",
		"version": "1.0",
	}

	go func() {
		srv.connectionOpen()
		rwc.Close()
	}()

	if c, err := Open(rwc, config); err != nil {
		t.Fatalf("could not create connection: %v (%s)", c, err)
	}

	if want, got := config.Properties["product"], srv.start.ClientProperties["product"]; want != got {
		t.Errorf("expected product %s got: %s", want, got)
	}

	if want, got := config.Properties["version"], srv.start.ClientProperties["version"]; want != got {
		t.Errorf("expected version %s got: %s", want, got)
	}
}

func TestOpen(t *testing.T) {
	rwc, srv := newSession(t)
	go func() {
		srv.connectionOpen()
		rwc.Close()
	}()

	if c, err := Open(rwc, defaultConfig()); err != nil {
		t.Fatalf("could not create connection: %v (%s)", c, err)
	}
}

func TestChannelOpen(t *testing.T) {
	rwc, srv := newSession(t)

	go func() {
		srv.connectionOpen()
		srv.channelOpen(1)

		rwc.Close()
	}()

	c, err := Open(rwc, defaultConfig())
	if err != nil {
		t.Fatalf("could not create connection: %v (%s)", c, err)
	}

	ch, err := c.Channel()
	if err != nil {
		t.Fatalf("could not open channel: %v (%s)", ch, err)
	}
}

func TestOpenFailedSASLUnsupportedMechanisms(t *testing.T) {
	rwc, srv := newSession(t)

	go func() {
		srv.expectAMQP()
		srv.send(0, &connectionStart{
			VersionMajor: 0,
			VersionMinor: 9,
			Mechanisms:   "KERBEROS NTLM",
			Locales:      "en_US",
		})
	}()

	c, err := Open(rwc, defaultConfig())
	if err != ErrSASL {
		t.Fatalf("expected ErrSASL got: %+v on %+v", err, c)
	}
}

func TestOpenFailedCredentials(t *testing.T) {
	rwc, srv := newSession(t)

	go func() {
		srv.expectAMQP()
		srv.connectionStart()
		// Now kill/timeout the connection indicating bad auth
		rwc.Close()
	}()

	c, err := Open(rwc, defaultConfig())
	if err != ErrCredentials {
		t.Fatalf("expected ErrCredentials got: %+v on %+v", err, c)
	}
}

func TestOpenFailedVhost(t *testing.T) {
	rwc, srv := newSession(t)

	go func() {
		srv.expectAMQP()
		srv.connectionStart()
		srv.connectionTune()
		srv.recv(0, &connectionOpen{})

		// Now kill/timeout the connection on bad Vhost
		rwc.Close()
	}()

	c, err := Open(rwc, defaultConfig())
	if err != ErrVhost {
		t.Fatalf("expected ErrVhost got: %+v on %+v", err, c)
	}
}

func TestConfirmMultipleOrdersDeliveryTags(t *testing.T) {
	rwc, srv := newSession(t)
	defer rwc.Close()

	go func() {
		srv.connectionOpen()
		srv.channelOpen(1)

		srv.recv(1, &confirmSelect{})
		srv.send(1, &confirmSelectOk{})

		srv.recv(1, &basicPublish{})
		srv.recv(1, &basicPublish{})
		srv.recv(1, &basicPublish{})
		srv.recv(1, &basicPublish{})

		// Single tag, plus multiple, should produce
		// 2, 1, 3, 4
		srv.send(1, &basicAck{DeliveryTag: 2})
		srv.send(1, &basicAck{DeliveryTag: 1})
		srv.send(1, &basicAck{DeliveryTag: 4, Multiple: true})

		srv.recv(1, &basicPublish{})
		srv.recv(1, &basicPublish{})
		srv.recv(1, &basicPublish{})
		srv.recv(1, &basicPublish{})

		// And some more, but in reverse order, multiple then one
		// 5, 6, 7, 8
		srv.send(1, &basicAck{DeliveryTag: 6, Multiple: true})
		srv.send(1, &basicAck{DeliveryTag: 8})
		srv.send(1, &basicAck{DeliveryTag: 7})
	}()

	c, err := Open(rwc, defaultConfig())
	if err != nil {
		t.Fatalf("could not create connection: %v (%s)", c, err)
	}

	ch, err := c.Channel()
	if err != nil {
		t.Fatalf("could not open channel: %v (%s)", ch, err)
	}

	confirm := ch.NotifyPublish(make(chan Confirmation))

	ch.Confirm(false)

	go func() {
		ch.Publish("", "q", false, false, Publishing{Body: []byte("pub 1")})
		ch.Publish("", "q", false, false, Publishing{Body: []byte("pub 2")})
		ch.Publish("", "q", false, false, Publishing{Body: []byte("pub 3")})
		ch.Publish("", "q", false, false, Publishing{Body: []byte("pub 4")})
	}()

	// received out of order, consumed in order
	for i, tag := range []uint64{1, 2, 3, 4} {
		if ack := <-confirm; tag != ack.DeliveryTag {
			t.Fatalf("failed ack, expected ack#%d to be %d, got %d", i, tag, ack.DeliveryTag)
		}
	}

	go func() {
		ch.Publish("", "q", false, false, Publishing{Body: []byte("pub 5")})
		ch.Publish("", "q", false, false, Publishing{Body: []byte("pub 6")})
		ch.Publish("", "q", false, false, Publishing{Body: []byte("pub 7")})
		ch.Publish("", "q", false, false, Publishing{Body: []byte("pub 8")})
	}()

	for i, tag := range []uint64{5, 6, 7, 8} {
		if ack := <-confirm; tag != ack.DeliveryTag {
			t.Fatalf("failed ack, expected ack#%d to be %d, got %d", i, tag, ack.DeliveryTag)
		}
	}

}

func TestNotifyClosesReusedPublisherConfirmChan(t *testing.T) {
	rwc, srv := newSession(t)

	go func() {
		srv.connectionOpen()
		srv.channelOpen(1)

		srv.recv(1, &confirmSelect{})
		srv.send(1, &confirmSelectOk{})

		srv.recv(0, &connectionClose{})
		srv.send(0, &connectionCloseOk{})
	}()

	c, err := Open(rwc, defaultConfig())
	if err != nil {
		t.Fatalf("could not create connection: %v (%s)", c, err)
	}

	ch, err := c.Channel()
	if err != nil {
		t.Fatalf("could not open channel: %v (%s)", ch, err)
	}

	ackAndNack := make(chan uint64)
	ch.NotifyConfirm(ackAndNack, ackAndNack)

	if err := ch.Confirm(false); err != nil {
		t.Fatalf("expected to enter confirm mode: %v", err)
	}

	if err := c.Close(); err != nil {
		t.Fatalf("could not close connection: %v (%s)", c, err)
	}
}

func TestNotifyClosesAllChansAfterConnectionClose(t *testing.T) {
	rwc, srv := newSession(t)

	go func() {
		srv.connectionOpen()
		srv.channelOpen(1)

		srv.recv(0, &connectionClose{})
		srv.send(0, &connectionCloseOk{})
	}()

	c, err := Open(rwc, defaultConfig())
	if err != nil {
		t.Fatalf("could not create connection: %v (%s)", c, err)
	}

	ch, err := c.Channel()
	if err != nil {
		t.Fatalf("could not open channel: %v (%s)", ch, err)
	}

	if err := c.Close(); err != nil {
		t.Fatalf("could not close connection: %v (%s)", c, err)
	}

	select {
	case <-c.NotifyClose(make(chan *Error)):
	case <-time.After(time.Millisecond):
		t.Errorf("expected to close NotifyClose chan after Connection.Close")
	}

	select {
	case <-ch.NotifyClose(make(chan *Error)):
	case <-time.After(time.Millisecond):
		t.Errorf("expected to close Connection.NotifyClose chan after Connection.Close")
	}

	select {
	case <-ch.NotifyFlow(make(chan bool)):
	case <-time.After(time.Millisecond):
		t.Errorf("expected to close Channel.NotifyFlow chan after Connection.Close")
	}

	select {
	case <-ch.NotifyCancel(make(chan string)):
	case <-time.After(time.Millisecond):
		t.Errorf("expected to close Channel.NofityCancel chan after Connection.Close")
	}

	select {
	case <-ch.NotifyReturn(make(chan Return)):
	case <-time.After(time.Millisecond):
		t.Errorf("expected to close Channel.NotifyReturn chan after Connection.Close")
	}

	confirms := ch.NotifyPublish(make(chan Confirmation))

	select {
	case <-confirms:
	case <-time.After(time.Millisecond):
		t.Errorf("expected to close confirms on Channel.NotifyPublish chan after Connection.Close")
	}
}

// Should not panic when sending bodies split at different boundaries
func TestPublishBodySliceIssue74(t *testing.T) {
	rwc, srv := newSession(t)
	defer rwc.Close()

	const frameSize = 100
	const publishings = frameSize * 3

	done := make(chan bool)
	base := make([]byte, publishings)

	go func() {
		srv.connectionOpen()
		srv.channelOpen(1)

		for i := 0; i < publishings; i++ {
			srv.recv(1, &basicPublish{})
		}

		done <- true
	}()

	cfg := defaultConfig()
	cfg.FrameSize = frameSize

	c, err := Open(rwc, cfg)
	if err != nil {
		t.Fatalf("could not create connection: %v (%s)", c, err)
	}

	ch, err := c.Channel()
	if err != nil {
		t.Fatalf("could not open channel: %v (%s)", ch, err)
	}

	for i := 0; i < publishings; i++ {
		go ch.Publish("", "q", false, false, Publishing{Body: base[0:i]})
	}

	<-done
}

// Should not panic when server and client have frame_size of 0
func TestPublishZeroFrameSizeIssue161(t *testing.T) {
	rwc, srv := newSession(t)
	defer rwc.Close()

	const frameSize = 0
	const publishings = 1
	done := make(chan bool)

	go func() {
		srv.connectionOpen()
		srv.channelOpen(1)

		for i := 0; i < publishings; i++ {
			srv.recv(1, &basicPublish{})
		}

		done <- true
	}()

	cfg := defaultConfig()
	cfg.FrameSize = frameSize

	c, err := Open(rwc, cfg)

	// override the tuned framesize with a hard 0, as would happen when rabbit is configured with 0
	c.Config.FrameSize = frameSize

	if err != nil {
		t.Fatalf("could not create connection: %v (%s)", c, err)
	}

	ch, err := c.Channel()
	if err != nil {
		t.Fatalf("could not open channel: %v (%s)", ch, err)
	}

	for i := 0; i < publishings; i++ {
		go ch.Publish("", "q", false, false, Publishing{Body: []byte("anything")})
	}

	<-done
}

func TestPublishAndShutdownDeadlockIssue84(t *testing.T) {
	rwc, srv := newSession(t)
	defer rwc.Close()

	go func() {
		srv.connectionOpen()
		srv.channelOpen(1)
		srv.recv(1, &basicPublish{})
		// Mimic a broken io pipe so that Publish catches the error and goes into shutdown
		srv.S.Close()
	}()

	c, err := Open(rwc, defaultConfig())
	if err != nil {
		t.Fatalf("couldn't create connection: %v (%s)", c, err)
	}

	ch, err := c.Channel()
	if err != nil {
		t.Fatalf("couldn't open channel: %v (%s)", ch, err)
	}

	defer time.AfterFunc(500*time.Millisecond, func() { panic("Publish deadlock") }).Stop()
	for {
		if err := ch.Publish("exchange", "q", false, false, Publishing{Body: []byte("test")}); err != nil {
			t.Log("successfully caught disconnect error", err)
			return
		}
	}
}

// TestChannelReturnsCloseRace ensures that receiving a basicReturn frame and
// sending the notification to the bound channel does not race with
// channel.shutdown() which closes all registered notification channels - checks
// for a "send on closed channel" panic
func TestChannelReturnsCloseRace(t *testing.T) {
	defer time.AfterFunc(5*time.Second, func() { panic("Shutdown deadlock") }).Stop()
	ch := newChannel(&Connection{}, 1)

	// Register a channel to close in channel.shutdown()
	notify := make(chan Return, 1)
	ch.NotifyReturn(notify)

	go func() {
		for range notify {
			// Drain notifications
		}
	}()

	// Simulate receiving a load of returns (triggering a write to the above
	// channel) while we call shutdown concurrently
	go func() {
		for i := 0; i < 100; i++ {
			ch.dispatch(&basicReturn{})
		}
	}()

	ch.shutdown(nil)
}

// TestLeakClosedConsumersIssue264 ensures that closing a consumer with
// prefetched messages does not leak the buffering goroutine.
func TestLeakClosedConsumersIssue264(t *testing.T) {
	const tag = "consumer-tag"

	rwc, srv := newSession(t)
	defer rwc.Close()

	go func() {
		srv.connectionOpen()
		srv.channelOpen(1)

		srv.recv(1, &basicQos{})
		srv.send(1, &basicQosOk{})

		srv.recv(1, &basicConsume{})
		srv.send(1, &basicConsumeOk{ConsumerTag: tag})

		// This delivery is intended to be consumed
		srv.send(1, &basicDeliver{ConsumerTag: tag, DeliveryTag: 1})

		// This delivery is intended to be dropped
		srv.send(1, &basicDeliver{ConsumerTag: tag, DeliveryTag: 2})

		srv.recv(0, &connectionClose{})
		srv.send(0, &connectionCloseOk{})
		srv.C.Close()
	}()

	c, err := Open(rwc, defaultConfig())
	if err != nil {
		t.Fatalf("could not create connection: %v (%s)", c, err)
	}

	ch, err := c.Channel()
	if err != nil {
		t.Fatalf("could not open channel: %v (%s)", ch, err)
	}
	ch.Qos(2, 0, false)

	consumer, err := ch.Consume("queue", tag, false, false, false, false, nil)
	if err != nil {
		t.Fatalf("unexpected error during consumer: %v", err)
	}

	first := <-consumer
	if want, got := uint64(1), first.DeliveryTag; want != got {
		t.Fatalf("unexpected delivery tag: want: %d, got: %d", want, got)
	}

	if err := c.Close(); err != nil {
		t.Fatalf("unexpected error during connection close: %v", err)
	}

	if _, open := <-consumer; open {
		t.Fatalf("expected deliveries channel to be closed immediately when the connection is closed so not to leak the bufferDeliveries goroutine")
	}
}
