package nats_consumer

import (
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestIntegrationStartStop(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := testutil.Container{
		Image:        "nats",
		ExposedPorts: []string{"4222"},
		WaitingFor:   wait.ForLog("Server is ready"),
	}
	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()

	plugin := &NatsConsumer{
		Servers:                []string{"nats://" + container.Address + ":" + container.Ports["4222"]},
		Subjects:               []string{"telegraf"},
		QueueGroup:             "telegraf_consumers",
		PendingBytesLimit:      nats.DefaultSubPendingBytesLimit,
		PendingMessageLimit:    nats.DefaultSubPendingMsgsLimit,
		MaxUndeliveredMessages: defaultMaxUndeliveredMessages,
		Log:                    testutil.Logger{},
	}

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	plugin.Stop()
}

func TestIntegrationSendReceive(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := testutil.Container{
		Image:        "nats",
		ExposedPorts: []string{"4222"},
		WaitingFor:   wait.ForLog("Server is ready"),
	}
	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()
	addr := "nats://" + container.Address + ":" + container.Ports["4222"]

	tests := []struct {
		name     string
		msgs     map[string][]string
		expected []telegraf.Metric
	}{
		{
			name: "single message",
			msgs: map[string][]string{
				"telegraf": {"test,source=foo value=42i"},
			},
			expected: []telegraf.Metric{
				metric.New(
					"test",
					map[string]string{
						"source":  "foo",
						"subject": "telegraf",
					},
					map[string]interface{}{"value": int64(42)},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "multiple message",
			msgs: map[string][]string{
				"telegraf": {
					"test,source=foo value=42i",
					"test,source=bar value=23i",
				},
				"hitchhiker": {
					"wale,part=front named=true",
					"wale,part=back named=false",
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"test",
					map[string]string{
						"source":  "foo",
						"subject": "telegraf",
					},
					map[string]interface{}{"value": int64(42)},
					time.Unix(0, 0),
				),
				metric.New(
					"test",
					map[string]string{
						"source":  "bar",
						"subject": "telegraf",
					},
					map[string]interface{}{"value": int64(23)},
					time.Unix(0, 0),
				),
				metric.New(
					"wale",
					map[string]string{
						"part":    "front",
						"subject": "hitchhiker",
					},
					map[string]interface{}{"named": true},
					time.Unix(0, 0),
				),
				metric.New(
					"wale",
					map[string]string{
						"part":    "back",
						"subject": "hitchhiker",
					},
					map[string]interface{}{"named": false},
					time.Unix(0, 0),
				),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subjects := make([]string, 0, len(tt.msgs))
			for k := range tt.msgs {
				subjects = append(subjects, k)
			}

			// Setup the plugin
			plugin := &NatsConsumer{
				Servers:                []string{addr},
				Subjects:               subjects,
				QueueGroup:             "telegraf_consumers",
				PendingBytesLimit:      nats.DefaultSubPendingBytesLimit,
				PendingMessageLimit:    nats.DefaultSubPendingMsgsLimit,
				MaxUndeliveredMessages: defaultMaxUndeliveredMessages,
				Log:                    testutil.Logger{},
			}

			// Add a line-protocol parser
			parser := &influx.Parser{}
			require.NoError(t, parser.Init())
			plugin.SetParser(parser)

			// Startup the plugin
			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			defer plugin.Stop()

			// Send all messages to the topics (random order due to Golang map)
			publisher := &sender{addr: addr}
			require.NoError(t, publisher.connect())
			defer publisher.disconnect()
			for topic, msgs := range tt.msgs {
				for _, msg := range msgs {
					require.NoError(t, publisher.send(topic, msg))
				}
			}
			publisher.disconnect()

			// Wait for the metrics to be collected
			require.Eventually(t, func() bool {
				acc.Lock()
				defer acc.Unlock()
				return acc.NMetrics() >= uint64(len(tt.expected))
			}, time.Second, 100*time.Millisecond)

			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, tt.expected, actual, testutil.IgnoreTime(), testutil.SortMetrics())

			plugin.Lock()
			defer plugin.Unlock()
			require.Empty(t, plugin.undelivered)
		})
	}
}

func TestJetStreamIntegrationSendReceive(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := testutil.Container{
		Image:        "nats",
		ExposedPorts: []string{"4222"},
		Cmd:          []string{"-js"},
		WaitingFor:   wait.ForLog("Server is ready"),
	}
	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()
	addr := "nats://" + container.Address + ":" + container.Ports["4222"]

	// Add a JetStream stream for testing
	nc, err := nats.Connect(addr)
	require.NoError(t, err)
	defer nc.Close()
	js, err := nc.JetStream()
	require.NoError(t, err)

	streamName := "TESTSTREAM"
	subject := "telegraf"
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     streamName,
		Subjects: []string{subject},
	})
	require.NoError(t, err)

	// Setup the plugin for JetStream
	log := testutil.CaptureLogger{}
	plugin := &NatsConsumer{
		Servers:                []string{addr},
		JsSubjects:             []string{subject},
		JsStream:               streamName,
		QueueGroup:             "telegraf_consumers",
		PendingBytesLimit:      nats.DefaultSubPendingBytesLimit,
		PendingMessageLimit:    nats.DefaultSubPendingMsgsLimit,
		MaxUndeliveredMessages: defaultMaxUndeliveredMessages,
		Log:                    &log,
	}

	parser := &influx.Parser{}
	require.NoError(t, parser.Init())
	plugin.SetParser(parser)

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	// Publish a message to JetStream
	msg := "test,source=js value=99i"
	_, err = js.Publish(subject, []byte(msg))
	require.NoError(t, err)

	// Wait for the metric to be collected
	require.Eventually(t, func() bool {
		acc.Lock()
		defer acc.Unlock()
		return acc.NMetrics() >= 1
	}, time.Second, 100*time.Millisecond)

	actual := acc.GetTelegrafMetrics()
	expected := []telegraf.Metric{
		metric.New(
			"test",
			map[string]string{
				"source":  "js",
				"subject": subject,
			},
			map[string]interface{}{"value": int64(99)},
			time.Unix(0, 0),
		),
	}
	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime(), testutil.SortMetrics())

	// Acknowledge the message and check undelivered tracking
	log.Clear()
	plugin.Lock()
	require.Len(t, plugin.undelivered, 1)
	plugin.Unlock()
	for _, m := range actual {
		m.Accept()
	}

	require.Eventually(t, func() bool {
		plugin.Lock()
		defer plugin.Unlock()
		return len(plugin.undelivered) == 0
	}, time.Second, 100*time.Millisecond, "undelivered messages not cleared")

	require.Empty(t, log.Messages(), "no warnings or errors should be logged")
}

func TestJetStreamIntegrationSourcedStreamNotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := testutil.Container{
		Image:        "nats",
		ExposedPorts: []string{"4222"},
		Cmd:          []string{"-js"},
		WaitingFor:   wait.ForLog("Server is ready"),
	}
	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()
	addr := "nats://" + container.Address + ":" + container.Ports["4222"]

	// Add a JetStream stream for testing
	nc, err := nats.Connect(addr)
	require.NoError(t, err)
	defer nc.Close()
	js, err := nc.JetStream()
	require.NoError(t, err)

	// Create a stream with no subject
	streamName := "TESTSTREAM"
	_, err = js.AddStream(&nats.StreamConfig{
		Name: streamName,
		Sources: []*nats.StreamSource{
			{Name: "NONEXISTENT"},
		},
	})
	require.NoError(t, err)

	// Setup the plugin for JetStream
	plugin := &NatsConsumer{
		Servers:                []string{addr},
		JsSubjects:             []string{"TESTSTREAM"},
		QueueGroup:             "telegraf_consumers",
		PendingBytesLimit:      nats.DefaultSubPendingBytesLimit,
		PendingMessageLimit:    nats.DefaultSubPendingMsgsLimit,
		MaxUndeliveredMessages: defaultMaxUndeliveredMessages,
		Log:                    testutil.Logger{},
	}

	// Add a line-protocol parser
	parser := &influx.Parser{}
	require.NoError(t, parser.Init())
	plugin.SetParser(parser)

	// Startup the plugin
	var acc testutil.Accumulator
	err = plugin.Start(&acc)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no stream matches subject")
}

func TestJetStreamIntegrationSourcedStreamFound(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := testutil.Container{
		Image:        "nats",
		ExposedPorts: []string{"4222"},
		Cmd:          []string{"-js"},
		WaitingFor:   wait.ForLog("Server is ready"),
	}
	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()
	addr := "nats://" + container.Address + ":" + container.Ports["4222"]

	// Add a JetStream stream for testing
	nc, err := nats.Connect(addr)
	require.NoError(t, err)
	defer nc.Close()
	js, err := nc.JetStream()
	require.NoError(t, err)

	// Create a stream with no subject
	streamName := "TESTSTREAM"
	_, err = js.AddStream(&nats.StreamConfig{
		Name: streamName,
		Sources: []*nats.StreamSource{
			{Name: "NONEXISTENT"},
		},
	})
	require.NoError(t, err)

	// Setup the plugin for JetStream
	plugin := &NatsConsumer{
		Servers:                []string{addr},
		JsSubjects:             []string{"TESTSTREAM"},
		JsStream:               streamName,
		QueueGroup:             "telegraf_consumers",
		PendingBytesLimit:      nats.DefaultSubPendingBytesLimit,
		PendingMessageLimit:    nats.DefaultSubPendingMsgsLimit,
		MaxUndeliveredMessages: defaultMaxUndeliveredMessages,
		Log:                    testutil.Logger{},
	}

	// Add a line-protocol parser
	parser := &influx.Parser{}
	require.NoError(t, parser.Init())
	plugin.SetParser(parser)

	// Startup the plugin
	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	plugin.Stop()
}

type sender struct {
	addr string
	conn *nats.Conn
}

func (s *sender) connect() error {
	conn, err := nats.Connect(s.addr)
	if err != nil {
		return err
	}
	s.conn = conn

	return nil
}

func (s *sender) disconnect() {
	if s.conn != nil && !s.conn.IsClosed() {
		_ = s.conn.Flush()
		s.conn.Close()
	}
	s.conn = nil
}

func (s *sender) send(topic, msg string) error {
	return s.conn.Publish(topic, []byte(msg))
}
