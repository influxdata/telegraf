package google_pubsub

import (
	"context"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	msgInflux = "cpu_load_short,host=server01 value=23422.0 1422568543702900257\n"
)

// Test ingesting InfluxDB-format PubSub message
func TestRunParse(t *testing.T) {
	subId := "sub-run-parse"

	acc := &testutil.Accumulator{}
	ctx, cancel := context.WithCancel(context.Background())
	ps, s := getTestPubsub(cancel, acc, subId)

	go ps.subReceive(ctx)
	go ps.receiveDelivered(ctx)

	testTracker := &testTracker{}
	msg := &testMsg{
		value:   msgInflux,
		tracker: testTracker,
	}
	s.messages <- msg

	acc.Wait(1)
	assert.Equal(t, acc.NFields(), 1)
	metric := acc.Metrics[0]
	validateTestInfluxMetric(t, metric)
}

// Test ingesting InfluxDB-format PubSub message with added subscription tag
func TestRunParseWithSubTag(t *testing.T) {
	subId := "sub-run-parse-subtagged"
	tag := "my-sub-tag"

	acc := &testutil.Accumulator{}
	ctx, cancel := context.WithCancel(context.Background())
	ps, s := getTestPubsub(cancel, acc, subId)

	// Add SubscriptionTag param
	ps.SubscriptionTag = tag

	go ps.subReceive(ctx)
	go ps.receiveDelivered(ctx)

	testTracker := &testTracker{}
	msg := &testMsg{
		value:   msgInflux,
		tracker: testTracker,
	}
	s.messages <- msg

	acc.Wait(1)

	assert.Equal(t, acc.NFields(), 1)
	metric := acc.Metrics[0]
	validateTestInfluxMetric(t, metric)

	assert.Contains(t, metric.Tags, tag)
	assert.Equal(t, metric.Tags[tag], subId)
}

func TestRunInvalidMessages(t *testing.T) {
	subId := "sub-invalid-messages"

	acc := &testutil.Accumulator{}
	ctx, cancel := context.WithCancel(context.Background())
	ps, s := getTestPubsub(cancel, acc, subId)

	go ps.subReceive(ctx)
	go ps.receiveDelivered(ctx)

	testTracker := &testTracker{}

	// Use invalid message
	msg := &testMsg{
		value:   "~invalidInfluxMsg~",
		tracker: testTracker,
	}
	s.messages <- msg

	acc.WaitError(1)

	// Make sure we acknowledged message so we don't receive it again.
	testTracker.WaitForAck(1)

	assert.Equal(t, acc.NFields(), 0)
}

func TestRunOverlongMessages(t *testing.T) {
	subId := "sub-message-too-long"

	acc := &testutil.Accumulator{}
	ctx, cancel := context.WithCancel(context.Background())
	ps, s := getTestPubsub(cancel, acc, subId)

	// Add MaxMessageLen param
	ps.MaxMessageLen = 1

	go ps.subReceive(ctx)
	go ps.receiveDelivered(ctx)

	testTracker := &testTracker{}
	msg := &testMsg{
		value:   msgInflux,
		tracker: testTracker,
	}
	s.messages <- msg

	acc.WaitError(1)

	// Make sure we acknowledged message so we don't receive it again.
	testTracker.WaitForAck(1)

	assert.Equal(t, acc.NFields(), 0)

}

// Test ingesting InfluxDB-format PubSub message
func getTestPubsub(cancel context.CancelFunc, acc *testutil.Accumulator, subId string) (*PubSub, *testSub) {
	testParser, _ := parsers.NewInfluxParser()

	s := &testSub{
		id:       subId,
		messages: make(chan *testMsg, 100),
	}
	ps := &PubSub{
		sub:    s,
		acc:    acc,
		sem:    make(semaphore, 100),
		parser: testParser,
		cancel: cancel,
	}

	return ps, s
}

func validateTestInfluxMetric(t *testing.T, m *testutil.Metric) {
	assert.Equal(t, "cpu_load_short", m.Measurement)
	assert.Equal(t, "server01", m.Tags["host"])
	assert.Equal(t, 23422.0, m.Fields["value"])
	assert.Equal(t, int64(1422568543702900257), m.Time.UnixNano())
}
