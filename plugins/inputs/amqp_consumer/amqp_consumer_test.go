package amqp_consumer

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/models"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestAutoEncoding(t *testing.T) {
	// Setup a gzipped payload
	enc, err := internal.NewGzipEncoder()
	require.NoError(t, err)
	payloadGZip, err := enc.Encode([]byte(`measurementName fieldKey="gzip" 1556813561098000000`))
	require.NoError(t, err)

	// Setup the plugin including the message parser
	decoder, err := internal.NewContentDecoder("auto")
	require.NoError(t, err)
	plugin := &AMQPConsumer{
		deliveries: make(map[telegraf.TrackingID]amqp091.Delivery),
		decoder:    decoder,
	}

	parser := &influx.Parser{}
	require.NoError(t, parser.Init())
	plugin.SetParser(parser)

	// Setup the message creator
	msg := amqp091.Delivery{
		ContentEncoding: "gzip",
		Body:            payloadGZip,
	}

	// Simulate a message receive event
	var acc testutil.Accumulator
	require.NoError(t, plugin.onMessage(&acc, msg))
	acc.AssertContainsFields(t, "measurementName", map[string]interface{}{"fieldKey": "gzip"})

	// Check the decoding
	encIdentity, err := internal.NewIdentityEncoder()
	require.NoError(t, err)
	payload, err := encIdentity.Encode([]byte(`measurementName2 fieldKey="identity" 1556813561098000000`))
	require.NoError(t, err)

	// Setup a non-encoded payload
	msg = amqp091.Delivery{
		ContentEncoding: "not_gzip",
		Body:            payload,
	}

	// Simulate a message receive event
	require.NoError(t, plugin.onMessage(&acc, msg))
	require.NoError(t, err)
	acc.AssertContainsFields(t, "measurementName2", map[string]interface{}{"fieldKey": "identity"})
}

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Define common properties
	servicePort := "5672"

	// Setup the container
	container := testutil.Container{
		Image:        "rabbitmq",
		ExposedPorts: []string{servicePort},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort(servicePort),
			wait.ForLog("Server startup complete"),
		),
	}
	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()
	url := fmt.Sprintf("amqp://%s:%s/", container.Address, container.Ports[servicePort])

	// Setup a AMQP producer to send messages
	client, err := newProducer(url)
	require.NoError(t, err)
	defer client.close()

	// Setup the plugin with an Influx line-protocol parser
	plugin := &AMQPConsumer{
		Brokers:      []string{url},
		Username:     config.NewSecret([]byte("guest")),
		Password:     config.NewSecret([]byte("guest")),
		Timeout:      config.Duration(3 * time.Second),
		Exchange:     "telegraf",
		ExchangeType: "direct",
		Queue:        "test",
		BindingKey:   "test",
		Log:          testutil.Logger{},
	}

	parser := &influx.Parser{}
	require.NoError(t, parser.Init())
	plugin.SetParser(parser)
	require.NoError(t, plugin.Init())

	// Setup the metrics
	metrics := []string{
		"test,source=A value=0i 1712780301000000000",
		"test,source=B value=1i 1712780301000000100",
		"test,source=C value=2i 1712780301000000200",
	}
	expected := make([]telegraf.Metric, 0, len(metrics))
	for _, x := range metrics {
		m, err := parser.Parse([]byte(x))
		require.NoError(t, err)
		expected = append(expected, m...)
	}

	// Start the plugin
	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	// Write metrics
	for _, x := range metrics {
		require.NoError(t, client.write(t.Context(), []byte(x)))
	}

	// Verify that the metrics were actually written
	require.Eventually(t, func() bool {
		return acc.NMetrics() >= uint64(len(expected))
	}, 3*time.Second, 100*time.Millisecond)

	client.close()
	plugin.Stop()
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics())
}

func TestReconnectIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Define common properties
	servicePort := "5672"

	// Setup the container
	container := testutil.Container{
		Image:        "rabbitmq",
		ExposedPorts: []string{servicePort},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort(servicePort),
			wait.ForLog("Server startup complete"),
		),
	}
	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()
	url := fmt.Sprintf("amqp://%s:%s/", container.Address, container.Ports[servicePort])

	// Setup a AMQP producer to send messages
	client, err := newProducer(url)
	require.NoError(t, err)
	defer client.close()

	// Setup the plugin with an Influx line-protocol parser
	logger := &testutil.CaptureLogger{}
	plugin := &AMQPConsumer{
		Brokers:      []string{url},
		Username:     config.NewSecret([]byte("guest")),
		Password:     config.NewSecret([]byte("guest")),
		Timeout:      config.Duration(500 * time.Millisecond),
		Exchange:     "telegraf",
		ExchangeType: "direct",
		Queue:        "test",
		BindingKey:   "test",
		Log:          logger,
	}

	parser := &influx.Parser{}
	require.NoError(t, parser.Init())
	plugin.SetParser(parser)
	require.NoError(t, plugin.Init())

	// Setup the metrics
	metrics := []string{
		"test,source=A value=0i 1712780301000000000",
		"test,source=B value=1i 1712780301000000100",
		"test,source=C value=2i 1712780301000000200",
	}
	expected := make([]telegraf.Metric, 0, len(metrics))
	for _, x := range metrics {
		m, err := parser.Parse([]byte(x))
		require.NoError(t, err)
		expected = append(expected, m...)
	}

	// Start the plugin
	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	// Write metrics
	for _, x := range metrics {
		require.NoError(t, client.write(t.Context(), []byte(x)))
	}

	// Verify that the metrics were actually written
	require.Eventually(t, func() bool {
		return acc.NMetrics() >= uint64(len(expected))
	}, 3*time.Second, 100*time.Millisecond)
	client.close()
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics())

	// Remove the metrics to make sure we can receive after reconnecting
	acc.ClearMetrics()
	require.Empty(t, acc.GetTelegrafMetrics())

	// Stop the container to force the plugin to disconnect and wait for the
	// corresponding message
	require.NoError(t, container.Stop())
	require.Eventually(t, func() bool {
		plugin.Lock()
		defer plugin.Unlock()
		return plugin.conn.IsClosed()
	}, 30*time.Second, 100*time.Millisecond, "plugin never lost connection")
	require.Eventually(t, func() bool {
		var disconnected, firstAttempt bool
		for _, msg := range logger.Messages() {
			// Check for the disconnect message
			if msg.Level == testutil.LevelInfo &&
				strings.Contains(msg.Text, "Connection closed") &&
				strings.Contains(msg.Text, "trying to reconnect") {
				disconnected = true
				continue
			}

			// Check for the first attempt to reconnect
			if msg.Level == testutil.LevelError && strings.Contains(msg.Text, "AMQP connection failed") {
				firstAttempt = true
			}
		}
		return disconnected && firstAttempt
	}, 3*time.Second, 100*time.Millisecond)

	// Restart the container and wait for the plugin to reconnect with a long
	// timeout. This is necessary because the plugin has a fixed reconnect
	// interval of 10 seconds so give it some tries.
	require.NoError(t, container.Start())
	url = fmt.Sprintf("amqp://%s:%s/", container.Address, container.Ports[servicePort])
	plugin.Brokers = []string{url}

	require.Eventually(t, func() bool {
		plugin.Lock()
		defer plugin.Unlock()
		return !plugin.conn.IsClosed()
	}, 15*time.Second, 100*time.Millisecond, "plugin never reconnected")

	// Setup another AMQP producer to send messages
	client2, err := newProducer(url)
	require.NoError(t, err)
	defer client2.close()

	// Write metrics
	for _, x := range metrics {
		require.NoError(t, client2.write(t.Context(), []byte(x)))
	}

	// Verify that the metrics were actually written
	require.Eventually(t, func() bool {
		return acc.NMetrics() >= uint64(len(expected))
	}, 3*time.Second, 100*time.Millisecond)
	client2.close()
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics())
}

func TestStartupErrorBehaviorErrorIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Define common properties
	servicePort := "5672"

	// Setup the container
	container := testutil.Container{
		Image:        "rabbitmq",
		ExposedPorts: []string{servicePort},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort(servicePort),
			wait.ForLog("Server startup complete"),
		),
	}
	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()
	url := fmt.Sprintf("amqp://%s:%s/", container.Address, container.Ports[servicePort])

	// Pause the container for simulating connectivity issues
	require.NoError(t, container.Pause())
	defer container.Resume() //nolint:errcheck // Ignore the returned error as we cannot do anything about it anyway

	// Setup the plugin with an Influx line-protocol parser
	plugin := &AMQPConsumer{
		Brokers:      []string{url},
		Username:     config.NewSecret([]byte("guest")),
		Password:     config.NewSecret([]byte("guest")),
		Timeout:      config.Duration(1 * time.Second),
		Exchange:     "telegraf",
		ExchangeType: "direct",
		Queue:        "test",
		BindingKey:   "test",
		Log:          testutil.Logger{},
	}

	parser := &influx.Parser{}
	require.NoError(t, parser.Init())
	plugin.SetParser(parser)

	// Create a model to be able to use the startup retry strategy
	model := models.NewRunningInput(
		plugin,
		&models.InputConfig{
			Name: "amqp",
		},
	)
	require.NoError(t, model.Init())

	// Starting the plugin will fail with an error because the container
	// is paused.
	var acc testutil.Accumulator
	require.ErrorContains(t, model.Start(&acc), "could not connect to any broker")
}

func TestStartupErrorBehaviorIgnoreIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Define common properties
	servicePort := "5672"

	// Setup the container
	container := testutil.Container{
		Image:        "rabbitmq",
		ExposedPorts: []string{servicePort},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort(servicePort),
			wait.ForLog("Server startup complete"),
		),
	}
	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()
	url := fmt.Sprintf("amqp://%s:%s/", container.Address, container.Ports[servicePort])

	// Pause the container for simulating connectivity issues
	require.NoError(t, container.Pause())
	defer container.Resume() //nolint:errcheck // Ignore the returned error as we cannot do anything about it anyway

	// Setup the plugin with an Influx line-protocol parser
	plugin := &AMQPConsumer{
		Brokers:      []string{url},
		Username:     config.NewSecret([]byte("guest")),
		Password:     config.NewSecret([]byte("guest")),
		Timeout:      config.Duration(1 * time.Second),
		Exchange:     "telegraf",
		ExchangeType: "direct",
		Queue:        "test",
		BindingKey:   "test",
		Log:          testutil.Logger{},
	}

	parser := &influx.Parser{}
	require.NoError(t, parser.Init())
	plugin.SetParser(parser)

	// Create a model to be able to use the startup retry strategy
	model := models.NewRunningInput(
		plugin,
		&models.InputConfig{
			Name:                 "amqp",
			StartupErrorBehavior: "ignore",
		},
	)
	require.NoError(t, model.Init())

	// Starting the plugin will fail because the container is paused.
	// The model code should convert it to a fatal error for the agent to remove
	// the plugin.
	var acc testutil.Accumulator
	err := model.Start(&acc)
	require.ErrorContains(t, err, "could not connect to any broker")
	var fatalErr *internal.FatalError
	require.ErrorAs(t, err, &fatalErr)
}

func TestStartupErrorBehaviorRetryIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Define common properties
	servicePort := "5672"

	// Setup the container
	container := testutil.Container{
		Image:        "rabbitmq",
		ExposedPorts: []string{servicePort},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort(servicePort),
			wait.ForLog("Server startup complete"),
		),
	}
	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()
	url := fmt.Sprintf("amqp://%s:%s/", container.Address, container.Ports[servicePort])

	// Pause the container for simulating connectivity issues
	require.NoError(t, container.Pause())
	defer container.Resume() //nolint:errcheck // Ignore the returned error as we cannot do anything about it anyway

	// Setup the plugin with an Influx line-protocol parser
	plugin := &AMQPConsumer{
		Brokers:      []string{url},
		Username:     config.NewSecret([]byte("guest")),
		Password:     config.NewSecret([]byte("guest")),
		Timeout:      config.Duration(1 * time.Second),
		Exchange:     "telegraf",
		ExchangeType: "direct",
		Queue:        "test",
		BindingKey:   "test",
		Log:          testutil.Logger{},
	}

	parser := &influx.Parser{}
	require.NoError(t, parser.Init())
	plugin.SetParser(parser)

	// Create a model to be able to use the startup retry strategy
	model := models.NewRunningInput(
		plugin,
		&models.InputConfig{
			Name:                 "amqp",
			StartupErrorBehavior: "retry",
		},
	)
	require.NoError(t, model.Init())
	model.StartupErrors.Set(0)

	// Setup the metrics
	metrics := []string{
		"test,source=A value=0i 1712780301000000000",
		"test,source=B value=1i 1712780301000000100",
		"test,source=C value=2i 1712780301000000200",
	}
	expected := make([]telegraf.Metric, 0, len(metrics))
	for _, x := range metrics {
		m, err := parser.Parse([]byte(x))
		require.NoError(t, err)
		expected = append(expected, m...)
	}

	// Starting the plugin should succeed as we will retry to startup later
	var acc testutil.Accumulator
	require.NoError(t, model.Start(&acc))
	defer model.Stop()

	// There should be no metrics as the plugin is not fully started up yet
	require.Empty(t, acc.GetTelegrafMetrics())
	require.ErrorIs(t, model.Gather(&acc), internal.ErrNotConnected)
	require.Equal(t, int64(2), model.StartupErrors.Get())

	// Unpause the container, now writes should succeed
	require.NoError(t, container.Resume())
	require.NoError(t, model.Gather(&acc))

	// Setup a AMQP producer and send messages
	client, err := newProducer(url)
	require.NoError(t, err)
	defer client.close()

	// Write metrics
	for _, x := range metrics {
		require.NoError(t, client.write(t.Context(), []byte(x)))
	}

	// Verify that the metrics were actually collected
	require.Eventually(t, func() bool {
		return acc.NMetrics() >= uint64(len(expected))
	}, 3*time.Second, 100*time.Millisecond)

	client.close()
	plugin.Stop()
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics())
}

type producer struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
}

func newProducer(url string) (*producer, error) {
	cfg := amqp091.Config{Properties: amqp091.NewConnectionProperties()}
	cfg.Properties.SetClientConnectionName("test-producer")
	conn, err := amqp091.DialConfig(url, cfg)
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	if err := channel.ExchangeDeclare("telegraf", "direct", true, false, false, false, nil); err != nil {
		return nil, err
	}

	queue, err := channel.QueueDeclare("test", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	if err := channel.QueueBind(queue.Name, "test", "telegraf", false, nil); err != nil {
		return nil, err
	}

	return &producer{
		conn:    conn,
		channel: channel,
	}, nil
}

func (p *producer) close() {
	p.channel.Close()
	p.conn.Close()
}

func (p *producer) write(testContext context.Context, payload []byte) error {
	msg := amqp091.Publishing{
		DeliveryMode: amqp091.Persistent,
		Timestamp:    time.Now(),
		ContentType:  "text/plain",
		Body:         payload,
	}

	ctx, cancel := context.WithTimeout(testContext, 3*time.Second)
	defer cancel()

	return p.channel.PublishWithContext(ctx, "telegraf", "test", true, false, msg)
}
