package amqp_consumer

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/rabbitmq/amqp091-go"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
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
	vhost := "/"
	exchange := "telegraf"
	exchangeType := "direct"
	queueName := "test"
	bindingKey := "test"

	// Setup the container
	container := testutil.Container{
		Image:        "rabbitmq",
		ExposedPorts: []string{servicePort},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort(nat.Port(servicePort)),
			wait.ForLog("Server startup complete"),
		),
	}
	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()
	url := fmt.Sprintf("amqp://%s:%s%s", container.Address, container.Ports[servicePort], vhost)

	// Setup a AMQP producer to send messages
	client, err := newProducer(url, vhost, exchange, exchangeType, queueName, bindingKey)
	require.NoError(t, err)
	defer client.close()

	// Setup the plugin with an Influx line-protocol parser
	plugin := &AMQPConsumer{
		Brokers:      []string{url},
		Username:     config.NewSecret([]byte("guest")),
		Password:     config.NewSecret([]byte("guest")),
		Timeout:      config.Duration(3 * time.Second),
		Exchange:     exchange,
		ExchangeType: exchangeType,
		Queue:        queueName,
		BindingKey:   bindingKey,
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
	expexted := make([]telegraf.Metric, 0, len(metrics))
	for _, x := range metrics {
		m, err := parser.Parse([]byte(x))
		require.NoError(t, err)
		expexted = append(expexted, m...)
	}

	// Start the plugin
	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	// Write metrics
	for _, x := range metrics {
		require.NoError(t, client.write(exchange, queueName, []byte(x)))
	}

	// Verify that the metrics were actually written
	require.Eventually(t, func() bool {
		return acc.NMetrics() >= uint64(len(expexted))
	}, 3*time.Second, 100*time.Millisecond)

	client.close()
	plugin.Stop()
	testutil.RequireMetricsEqual(t, expexted, acc.GetTelegrafMetrics())
}

type producer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   amqp.Queue
}

func newProducer(url, vhost, exchange, exchangeType, queueName, key string) (*producer, error) {
	cfg := amqp.Config{
		Vhost:      vhost,
		Properties: amqp.NewConnectionProperties(),
	}
	cfg.Properties.SetClientConnectionName("test-producer")
	conn, err := amqp.DialConfig(url, cfg)
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	if err := channel.ExchangeDeclare(exchange, exchangeType, true, false, false, false, nil); err != nil {
		return nil, err
	}

	queue, err := channel.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	if err := channel.QueueBind(queue.Name, key, exchange, false, nil); err != nil {
		return nil, err
	}

	return &producer{
		conn:    conn,
		channel: channel,
		queue:   queue,
	}, nil
}

func (p *producer) close() {
	p.channel.Close()
	p.conn.Close()
}

func (p *producer) write(exchange, key string, payload []byte) error {
	msg := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		ContentType:  "text/plain",
		Body:         payload,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return p.channel.PublishWithContext(ctx, exchange, key, true, false, msg)
}
