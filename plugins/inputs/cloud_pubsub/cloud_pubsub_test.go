package cloud_pubsub

import (
	"encoding/base64"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/testutil"
)

const (
	msgInflux = "cpu_load_short,host=server01 value=23422.0 1422568543702900257\n"
)

// Test ingesting InfluxDB-format PubSub message
func TestRunParse(t *testing.T) {
	subID := "sub-run-parse"

	testParser, _ := parsers.NewInfluxParser()

	sub := &stubSub{
		id:       subID,
		messages: make(chan *testMsg, 100),
	}
	sub.receiver = testMessagesReceive(sub)

	ps := &PubSub{
		Log:                    testutil.Logger{},
		parser:                 testParser,
		stubSub:                func() subscription { return sub },
		Project:                "projectIDontMatterForTests",
		Subscription:           subID,
		MaxUndeliveredMessages: defaultMaxUndeliveredMessages,
	}

	acc := &testutil.Accumulator{}
	if err := ps.Start(acc); err != nil {
		t.Fatalf("test PubSub failed to start: %s", err)
	}
	defer ps.Stop()

	if ps.sub == nil {
		t.Fatal("expected plugin subscription to be non-nil")
	}

	testTracker := &testTracker{}
	msg := &testMsg{
		value:   msgInflux,
		tracker: testTracker,
	}
	sub.messages <- msg

	acc.Wait(1)
	require.Equal(t, acc.NFields(), 1)
	metric := acc.Metrics[0]
	validateTestInfluxMetric(t, metric)
}

// Test ingesting InfluxDB-format PubSub message
func TestRunBase64(t *testing.T) {
	subID := "sub-run-base64"

	testParser, _ := parsers.NewInfluxParser()

	sub := &stubSub{
		id:       subID,
		messages: make(chan *testMsg, 100),
	}
	sub.receiver = testMessagesReceive(sub)

	ps := &PubSub{
		Log:                    testutil.Logger{},
		parser:                 testParser,
		stubSub:                func() subscription { return sub },
		Project:                "projectIDontMatterForTests",
		Subscription:           subID,
		MaxUndeliveredMessages: defaultMaxUndeliveredMessages,
		Base64Data:             true,
	}

	acc := &testutil.Accumulator{}
	if err := ps.Start(acc); err != nil {
		t.Fatalf("test PubSub failed to start: %s", err)
	}
	defer ps.Stop()

	if ps.sub == nil {
		t.Fatal("expected plugin subscription to be non-nil")
	}

	testTracker := &testTracker{}
	msg := &testMsg{
		value:   base64.StdEncoding.EncodeToString([]byte(msgInflux)),
		tracker: testTracker,
	}
	sub.messages <- msg

	acc.Wait(1)
	require.Equal(t, acc.NFields(), 1)
	metric := acc.Metrics[0]
	validateTestInfluxMetric(t, metric)
}

func TestRunInvalidMessages(t *testing.T) {
	subID := "sub-invalid-messages"

	testParser, _ := parsers.NewInfluxParser()

	sub := &stubSub{
		id:       subID,
		messages: make(chan *testMsg, 100),
	}
	sub.receiver = testMessagesReceive(sub)

	ps := &PubSub{
		Log:                    testutil.Logger{},
		parser:                 testParser,
		stubSub:                func() subscription { return sub },
		Project:                "projectIDontMatterForTests",
		Subscription:           subID,
		MaxUndeliveredMessages: defaultMaxUndeliveredMessages,
	}

	acc := &testutil.Accumulator{}

	if err := ps.Start(acc); err != nil {
		t.Fatalf("test PubSub failed to start: %s", err)
	}
	defer ps.Stop()
	if ps.sub == nil {
		t.Fatal("expected plugin subscription to be non-nil")
	}

	testTracker := &testTracker{}
	msg := &testMsg{
		value:   "~invalidInfluxMsg~",
		tracker: testTracker,
	}
	sub.messages <- msg

	acc.WaitError(1)

	// Make sure we acknowledged message so we don't receive it again.
	testTracker.WaitForAck(1)

	require.Equal(t, acc.NFields(), 0)
}

func TestRunOverlongMessages(t *testing.T) {
	subID := "sub-message-too-long"

	acc := &testutil.Accumulator{}

	testParser, _ := parsers.NewInfluxParser()

	sub := &stubSub{
		id:       subID,
		messages: make(chan *testMsg, 100),
	}
	sub.receiver = testMessagesReceive(sub)

	ps := &PubSub{
		Log:                    testutil.Logger{},
		parser:                 testParser,
		stubSub:                func() subscription { return sub },
		Project:                "projectIDontMatterForTests",
		Subscription:           subID,
		MaxUndeliveredMessages: defaultMaxUndeliveredMessages,
		// Add MaxMessageLen Param
		MaxMessageLen: 1,
	}

	if err := ps.Start(acc); err != nil {
		t.Fatalf("test PubSub failed to start: %s", err)
	}
	defer ps.Stop()
	if ps.sub == nil {
		t.Fatal("expected plugin subscription to be non-nil")
	}

	testTracker := &testTracker{}
	msg := &testMsg{
		value:   msgInflux,
		tracker: testTracker,
	}
	sub.messages <- msg

	acc.WaitError(1)

	// Make sure we acknowledged message so we don't receive it again.
	testTracker.WaitForAck(1)

	require.Equal(t, acc.NFields(), 0)
}

func TestRunErrorInSubscriber(t *testing.T) {
	// hard-coded retry in seconds
	if testing.Short() {
		t.Skip("Skipping long test in short mode")
	}

	subID := "sub-unexpected-error"

	acc := &testutil.Accumulator{}

	testParser, _ := parsers.NewInfluxParser()

	sub := &stubSub{
		id:       subID,
		messages: make(chan *testMsg, 100),
	}
	fakeErrStr := "a fake error"
	sub.receiver = testMessagesError(errors.New("a fake error"))

	ps := &PubSub{
		Log:                      testutil.Logger{},
		parser:                   testParser,
		stubSub:                  func() subscription { return sub },
		Project:                  "projectIDontMatterForTests",
		Subscription:             subID,
		MaxUndeliveredMessages:   defaultMaxUndeliveredMessages,
		RetryReceiveDelaySeconds: 1,
	}

	if err := ps.Start(acc); err != nil {
		t.Fatalf("test PubSub failed to start: %s", err)
	}
	defer ps.Stop()

	if ps.sub == nil {
		t.Fatal("expected plugin subscription to be non-nil")
	}
	acc.WaitError(1)
	require.Regexp(t, fakeErrStr, acc.Errors[0])
}

func validateTestInfluxMetric(t *testing.T, m *testutil.Metric) {
	require.Equal(t, "cpu_load_short", m.Measurement)
	require.Equal(t, "server01", m.Tags["host"])
	require.Equal(t, 23422.0, m.Fields["value"])
	require.Equal(t, int64(1422568543702900257), m.Time.UnixNano())
}
