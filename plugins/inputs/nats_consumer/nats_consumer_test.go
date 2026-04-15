package nats_consumer

import (
	"fmt"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
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

func TestInitValidation(t *testing.T) {
	tests := []struct {
		name    string
		plugin  *NatsConsumer
		wantErr string
	}{
		{
			name: "invalid deliver policy",
			plugin: &NatsConsumer{
				JsDeliverPolicy: "invalid",
			},
			wantErr: `invalid jetstream_deliver_policy "invalid"`,
		},
		{
			name: "by_start_sequence without sequence",
			plugin: &NatsConsumer{
				JsDeliverPolicy: "by_start_sequence",
				JsStartSequence: 0,
			},
			wantErr: "jetstream_start_sequence must be set",
		},
		{
			name: "by_start_time without time",
			plugin: &NatsConsumer{
				JsDeliverPolicy: "by_start_time",
			},
			wantErr: "jetstream_start_time must be set",
		},
		{
			name: "by_start_time with invalid format",
			plugin: &NatsConsumer{
				JsDeliverPolicy: "by_start_time",
				JsStartTime:     "not-a-time",
			},
			wantErr: "must be RFC3339 format",
		},
		{
			name: "valid deliver policy all",
			plugin: &NatsConsumer{
				JsDeliverPolicy: "all",
			},
		},
		{
			name: "valid deliver policy new",
			plugin: &NatsConsumer{
				JsDeliverPolicy: "new",
			},
		},
		{
			name: "valid deliver policy last",
			plugin: &NatsConsumer{
				JsDeliverPolicy: "last",
			},
		},
		{
			name: "valid by_start_sequence",
			plugin: &NatsConsumer{
				JsDeliverPolicy: "by_start_sequence",
				JsStartSequence: 5,
			},
		},
		{
			name: "valid by_start_time",
			plugin: &NatsConsumer{
				JsDeliverPolicy: "by_start_time",
				JsStartTime:     "2024-01-01T00:00:00Z",
			},
		},
		{
			name:   "empty policy is valid",
			plugin: &NatsConsumer{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.plugin.Init()
			if tt.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestJetStreamDurableConsumerRestart(t *testing.T) {
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

	nc, err := nats.Connect(addr)
	require.NoError(t, err)
	defer nc.Close()
	js, err := nc.JetStream()
	require.NoError(t, err)

	streamName := "DURABLE_TEST"
	subject := "durable.test"
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     streamName,
		Subjects: []string{subject},
	})
	require.NoError(t, err)

	newPlugin := func() *NatsConsumer {
		return &NatsConsumer{
			Servers:                []string{addr},
			JsSubjects:             []string{subject},
			JsStream:               streamName,
			JsDurableName:          "test-durable",
			JsDeliverPolicy:        "all",
			JsAckWait:              config.Duration(30 * time.Second),
			JsMaxDeliver:           -1,
			PendingBytesLimit:      nats.DefaultSubPendingBytesLimit,
			PendingMessageLimit:    nats.DefaultSubPendingMsgsLimit,
			MaxUndeliveredMessages: defaultMaxUndeliveredMessages,
			Log:                    testutil.Logger{},
		}
	}

	plugin1 := newPlugin()
	require.NoError(t, plugin1.Init())
	parser := &influx.Parser{}
	require.NoError(t, parser.Init())
	plugin1.SetParser(parser)

	var acc1 testutil.Accumulator
	require.NoError(t, plugin1.Start(&acc1))

	for i := 0; i < 3; i++ {
		_, err = js.Publish(subject, []byte(fmt.Sprintf("test,batch=first value=%di", i)))
		require.NoError(t, err)
	}

	require.Eventually(t, func() bool {
		acc1.Lock()
		defer acc1.Unlock()
		return acc1.NMetrics() >= 3
	}, 5*time.Second, 100*time.Millisecond)

	for _, m := range acc1.GetTelegrafMetrics() {
		m.Accept()
	}

	require.Eventually(t, func() bool {
		plugin1.Lock()
		defer plugin1.Unlock()
		return len(plugin1.undelivered) == 0
	}, 5*time.Second, 100*time.Millisecond)

	plugin1.Stop()

	for i := 3; i < 6; i++ {
		_, err = js.Publish(subject, []byte(fmt.Sprintf("test,batch=second value=%di", i)))
		require.NoError(t, err)
	}

	plugin2 := newPlugin()
	require.NoError(t, plugin2.Init())
	parser2 := &influx.Parser{}
	require.NoError(t, parser2.Init())
	plugin2.SetParser(parser2)

	var acc2 testutil.Accumulator
	require.NoError(t, plugin2.Start(&acc2))
	defer plugin2.Stop()

	require.Eventually(t, func() bool {
		acc2.Lock()
		defer acc2.Unlock()
		return acc2.NMetrics() >= 3
	}, 5*time.Second, 100*time.Millisecond)

	actual := acc2.GetTelegrafMetrics()
	require.Len(t, actual, 3, "should only receive messages published after the first consumer stopped")

	for _, m := range actual {
		v, ok := m.GetField("value")
		require.True(t, ok)
		val := v.(int64)
		require.True(t, val >= 3 && val <= 5, "expected values 3-5 from second batch, got %d", val)
	}
}

func TestJetStreamDeliverPolicyNew(t *testing.T) {
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

	nc, err := nats.Connect(addr)
	require.NoError(t, err)
	defer nc.Close()
	js, err := nc.JetStream()
	require.NoError(t, err)

	streamName := "POLICY_NEW"
	subject := "policy.new"
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     streamName,
		Subjects: []string{subject},
	})
	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		_, err = js.Publish(subject, []byte(fmt.Sprintf("test,batch=old value=%di", i)))
		require.NoError(t, err)
	}

	plugin := &NatsConsumer{
		Servers:                []string{addr},
		JsSubjects:             []string{subject},
		JsStream:               streamName,
		JsDurableName:          "new-policy-durable",
		JsDeliverPolicy:        "new",
		JsAckWait:              config.Duration(30 * time.Second),
		JsMaxDeliver:           -1,
		PendingBytesLimit:      nats.DefaultSubPendingBytesLimit,
		PendingMessageLimit:    nats.DefaultSubPendingMsgsLimit,
		MaxUndeliveredMessages: defaultMaxUndeliveredMessages,
		Log:                    testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	parser := &influx.Parser{}
	require.NoError(t, parser.Init())
	plugin.SetParser(parser)

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	time.Sleep(500 * time.Millisecond)

	for i := 10; i < 13; i++ {
		_, err = js.Publish(subject, []byte(fmt.Sprintf("test,batch=new value=%di", i)))
		require.NoError(t, err)
	}

	require.Eventually(t, func() bool {
		acc.Lock()
		defer acc.Unlock()
		return acc.NMetrics() >= 3
	}, 5*time.Second, 100*time.Millisecond)

	actual := acc.GetTelegrafMetrics()
	require.Len(t, actual, 3, "should only receive new messages, not historical ones")

	for _, m := range actual {
		v, ok := m.GetField("value")
		require.True(t, ok)
		val := v.(int64)
		require.True(t, val >= 10 && val <= 12, "expected values 10-12, got %d", val)
	}
}

func TestJetStreamDeliverPolicyByStartSequence(t *testing.T) {
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

	nc, err := nats.Connect(addr)
	require.NoError(t, err)
	defer nc.Close()
	js, err := nc.JetStream()
	require.NoError(t, err)

	streamName := "POLICY_SEQ"
	subject := "policy.seq"
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     streamName,
		Subjects: []string{subject},
	})
	require.NoError(t, err)

	for i := 1; i <= 5; i++ {
		_, err = js.Publish(subject, []byte(fmt.Sprintf("test,idx=%d value=%di", i, i)))
		require.NoError(t, err)
	}

	plugin := &NatsConsumer{
		Servers:                []string{addr},
		JsSubjects:             []string{subject},
		JsStream:               streamName,
		JsDurableName:          "seq-policy-durable",
		JsDeliverPolicy:        "by_start_sequence",
		JsStartSequence:        3,
		JsAckWait:              config.Duration(30 * time.Second),
		JsMaxDeliver:           -1,
		PendingBytesLimit:      nats.DefaultSubPendingBytesLimit,
		PendingMessageLimit:    nats.DefaultSubPendingMsgsLimit,
		MaxUndeliveredMessages: defaultMaxUndeliveredMessages,
		Log:                    testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	parser := &influx.Parser{}
	require.NoError(t, parser.Init())
	plugin.SetParser(parser)

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	require.Eventually(t, func() bool {
		acc.Lock()
		defer acc.Unlock()
		return acc.NMetrics() >= 3
	}, 5*time.Second, 100*time.Millisecond)

	actual := acc.GetTelegrafMetrics()
	require.Len(t, actual, 3, "should receive messages starting from sequence 3")

	for _, m := range actual {
		v, ok := m.GetField("value")
		require.True(t, ok)
		val := v.(int64)
		require.True(t, val >= 3 && val <= 5, "expected values 3-5, got %d", val)
	}
}

func TestJetStreamMaxDeliver(t *testing.T) {
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

	nc, err := nats.Connect(addr)
	require.NoError(t, err)
	defer nc.Close()
	js, err := nc.JetStream()
	require.NoError(t, err)

	streamName := "MAX_DELIVER"
	subject := "maxdeliver.test"
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     streamName,
		Subjects: []string{subject},
	})
	require.NoError(t, err)

	plugin := &NatsConsumer{
		Servers:                []string{addr},
		JsSubjects:             []string{subject},
		JsStream:               streamName,
		JsDurableName:          "maxdeliver-durable",
		JsDeliverPolicy:        "all",
		JsAckWait:              config.Duration(1 * time.Second),
		JsMaxDeliver:           2,
		PendingBytesLimit:      nats.DefaultSubPendingBytesLimit,
		PendingMessageLimit:    nats.DefaultSubPendingMsgsLimit,
		MaxUndeliveredMessages: defaultMaxUndeliveredMessages,
		Log:                    testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	parser := &influx.Parser{}
	require.NoError(t, parser.Init())
	plugin.SetParser(parser)

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	_, err = js.Publish(subject, []byte("test,source=maxdeliver value=1i"))
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		acc.Lock()
		defer acc.Unlock()
		return acc.NMetrics() >= 1
	}, 5*time.Second, 100*time.Millisecond)

	consInfo, err := js.ConsumerInfo(streamName, "maxdeliver-durable")
	require.NoError(t, err)
	require.Equal(t, 2, consInfo.Config.MaxDeliver)
}

func TestJetStreamFilterSubjects(t *testing.T) {
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

	nc, err := nats.Connect(addr)
	require.NoError(t, err)
	defer nc.Close()
	js, err := nc.JetStream()
	require.NoError(t, err)

	streamName := "FILTER_SUBJ"
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     streamName,
		Subjects: []string{"filter.*"},
	})
	require.NoError(t, err)

	plugin := &NatsConsumer{
		Servers:                []string{addr},
		JsSubjects:             []string{"filter.a"},
		JsStream:               streamName,
		JsDurableName:          "filter-durable",
		JsDeliverPolicy:        "all",
		JsAckWait:              config.Duration(30 * time.Second),
		JsMaxDeliver:           -1,
		PendingBytesLimit:      nats.DefaultSubPendingBytesLimit,
		PendingMessageLimit:    nats.DefaultSubPendingMsgsLimit,
		MaxUndeliveredMessages: defaultMaxUndeliveredMessages,
		Log:                    testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	parser := &influx.Parser{}
	require.NoError(t, parser.Init())
	plugin.SetParser(parser)

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	_, err = js.Publish("filter.a", []byte("test,source=a value=1i"))
	require.NoError(t, err)
	_, err = js.Publish("filter.b", []byte("test,source=b value=2i"))
	require.NoError(t, err)
	_, err = js.Publish("filter.a", []byte("test,source=a value=3i"))
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		acc.Lock()
		defer acc.Unlock()
		return acc.NMetrics() >= 2
	}, 5*time.Second, 100*time.Millisecond)

	time.Sleep(500 * time.Millisecond)

	actual := acc.GetTelegrafMetrics()
	require.Len(t, actual, 2, "should only receive messages on filter.a subject")

	for _, m := range actual {
		tag, ok := m.GetTag("subject")
		require.True(t, ok)
		require.Equal(t, "filter.a", tag)
	}
}

func TestJetStreamDurableWithQueueGroup(t *testing.T) {
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

	nc, err := nats.Connect(addr)
	require.NoError(t, err)
	defer nc.Close()
	js, err := nc.JetStream()
	require.NoError(t, err)

	streamName := "QUEUE_GROUP"
	subject := "queue.test"
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     streamName,
		Subjects: []string{subject},
	})
	require.NoError(t, err)

	newPlugin := func(queueGroup string) *NatsConsumer {
		return &NatsConsumer{
			Servers:                []string{addr},
			JsSubjects:             []string{subject},
			JsStream:               streamName,
			JsDurableName:          "queue-durable",
			JsDeliverPolicy:        "all",
			JsAckWait:              config.Duration(30 * time.Second),
			JsMaxDeliver:           -1,
			QueueGroup:             queueGroup,
			PendingBytesLimit:      nats.DefaultSubPendingBytesLimit,
			PendingMessageLimit:    nats.DefaultSubPendingMsgsLimit,
			MaxUndeliveredMessages: defaultMaxUndeliveredMessages,
			Log:                    testutil.Logger{},
		}
	}

	plugin1 := newPlugin("test_queue")
	require.NoError(t, plugin1.Init())
	parser1 := &influx.Parser{}
	require.NoError(t, parser1.Init())
	plugin1.SetParser(parser1)

	plugin2 := newPlugin("test_queue")
	require.NoError(t, plugin2.Init())
	parser2 := &influx.Parser{}
	require.NoError(t, parser2.Init())
	plugin2.SetParser(parser2)

	var acc1, acc2 testutil.Accumulator
	require.NoError(t, plugin1.Start(&acc1))
	defer plugin1.Stop()
	require.NoError(t, plugin2.Start(&acc2))
	defer plugin2.Stop()

	msgCount := 20
	for i := 0; i < msgCount; i++ {
		_, err = js.Publish(subject, []byte(fmt.Sprintf("test,idx=%d value=%di", i, i)))
		require.NoError(t, err)
	}

	require.Eventually(t, func() bool {
		acc1.Lock()
		acc2.Lock()
		defer acc1.Unlock()
		defer acc2.Unlock()
		return acc1.NMetrics()+acc2.NMetrics() >= uint64(msgCount)
	}, 10*time.Second, 100*time.Millisecond)

	total := acc1.NMetrics() + acc2.NMetrics()
	require.Equal(t, uint64(msgCount), total, "all messages should be consumed across both consumers")
	require.Positive(t, acc1.NMetrics(), "consumer 1 should receive some messages")
	require.Positive(t, acc2.NMetrics(), "consumer 2 should receive some messages")
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
