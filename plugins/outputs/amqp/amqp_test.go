package amqp

import (
	"errors"
	"fmt"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
	"github.com/influxdata/telegraf/testutil"
)

type MockClient struct {
	PublishF func() error
	CloseF   func() error

	PublishCallCount int
	CloseCallCount   int
}

func (c *MockClient) Publish(string, []byte) error {
	c.PublishCallCount++
	return c.PublishF()
}

func (c *MockClient) Close() error {
	c.CloseCallCount++
	return c.CloseF()
}

func NewMockClient() Client {
	return &MockClient{
		PublishF: func() error {
			return nil
		},
		CloseF: func() error {
			return nil
		},
	}
}

func TestConnect(t *testing.T) {
	tests := []struct {
		name    string
		output  *AMQP
		errFunc func(t *testing.T, cfg *ClientConfig, err error)
	}{
		{
			name: "defaults",
			output: &AMQP{
				Brokers:            []string{DefaultURL},
				ExchangeType:       DefaultExchangeType,
				ExchangeDurability: "durable",
				AuthMethod:         DefaultAuthMethod,
				Headers: map[string]string{
					"database":         DefaultDatabase,
					"retention_policy": DefaultRetentionPolicy,
				},
				Timeout: config.Duration(time.Second * 5),
			},
			errFunc: func(t *testing.T, cfg *ClientConfig, err error) {
				require.Equal(t, []string{DefaultURL}, cfg.brokers)
				require.Empty(t, cfg.exchange)
				require.Equal(t, "topic", cfg.exchangeType)
				require.False(t, cfg.exchangePassive)
				require.True(t, cfg.exchangeDurable)
				require.Equal(t, amqp.Table(nil), cfg.exchangeArguments)
				require.Equal(t, amqp.Table{
					"database":         DefaultDatabase,
					"retention_policy": DefaultRetentionPolicy,
				}, cfg.headers)
				require.Equal(t, amqp.Transient, cfg.deliveryMode)
				require.NoError(t, err)
			},
		},
		{
			name: "headers overrides deprecated dbrp",
			output: &AMQP{
				Headers: map[string]string{
					"foo": "bar",
				},
			},
			errFunc: func(t *testing.T, cfg *ClientConfig, err error) {
				require.Equal(t, amqp.Table{
					"foo": "bar",
				}, cfg.headers)
				require.NoError(t, err)
			},
		},
		{
			name: "exchange args",
			output: &AMQP{
				ExchangeArguments: map[string]string{
					"foo": "bar",
				},
			},
			errFunc: func(t *testing.T, cfg *ClientConfig, err error) {
				require.Equal(t, amqp.Table{
					"foo": "bar",
				}, cfg.exchangeArguments)
				require.NoError(t, err)
			},
		},
		{
			name: "username password",
			output: &AMQP{
				Brokers:  []string{"amqp://foo:bar@localhost"},
				Username: config.NewSecret([]byte("telegraf")),
				Password: config.NewSecret([]byte("pa$$word")),
			},
			errFunc: func(t *testing.T, cfg *ClientConfig, err error) {
				require.Equal(t, []amqp.Authentication{
					&amqp.PlainAuth{
						Username: "telegraf",
						Password: "pa$$word",
					},
				}, cfg.auth)

				require.NoError(t, err)
			},
		},
		{
			name: "url support",
			output: &AMQP{
				Brokers: []string{DefaultURL},
			},
			errFunc: func(t *testing.T, cfg *ClientConfig, err error) {
				require.Equal(t, []string{DefaultURL}, cfg.brokers)
				require.NoError(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg *ClientConfig
			tt.output.connect = func(c *ClientConfig) (Client, error) {
				cfg = c
				return NewMockClient(), nil
			}
			require.NoError(t, tt.output.Init())
			err := tt.output.Connect()
			tt.errFunc(t, cfg, err)
		})
	}
}

func TestWriteReturnsErrorWhenBrokerUnavailable(t *testing.T) {
	serializer := &influx.Serializer{}
	require.NoError(t, serializer.Init())

	q := &AMQP{
		Brokers:      []string{DefaultURL},
		ExchangeType: DefaultExchangeType,
		AuthMethod:   DefaultAuthMethod,
		Timeout:      config.Duration(5 * time.Second),
		Log:          testutil.Logger{},
		connect: func(*ClientConfig) (Client, error) {
			return nil, errors.New("could not connect to any broker")
		},
	}
	require.NoError(t, q.Init())
	q.SetSerializer(serializer)

	// The broker is unreachable, so publish() calls connect() (q.client is
	// nil) and gets back a plain error that is not amqp.ErrClosed. Write must
	// surface that error so the framework keeps the metrics buffered for
	// retry instead of silently dropping them.
	require.ErrorContains(t, q.Write(testutil.MockMetrics()), "could not connect to any broker")
}

func TestReconnectIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	servicePort := "5672"
	container := &testutil.Container{
		Image:        "rabbitmq",
		ExposedPorts: []string{servicePort},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort(servicePort),
			wait.ForLog("Server startup complete"),
		),
	}
	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()

	serializer := &influx.Serializer{}
	require.NoError(t, serializer.Init())

	// Close the connection after every message to force a reconnect with the
	// same plugin configuration on each write
	plugin := &AMQP{
		Brokers:      []string{fmt.Sprintf("amqp://%s:%s/", container.Address, container.Ports[servicePort])},
		Exchange:     "telegraf",
		ExchangeType: DefaultExchangeType,
		AuthMethod:   DefaultAuthMethod,
		Username:     config.NewSecret([]byte("guest")),
		Password:     config.NewSecret([]byte("guest")),
		MaxMessages:  1,
		Timeout:      config.Duration(5 * time.Second),
		Log:          testutil.Logger{},
		connect:      connect,
	}
	plugin.SetSerializer(serializer)
	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Connect())
	defer plugin.Close()

	for range 3 {
		require.NoError(t, plugin.Write(testutil.MockMetrics()))
	}
}
