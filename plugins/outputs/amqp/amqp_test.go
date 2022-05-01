package amqp

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf/config"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/require"
)

type MockClient struct {
	PublishF func(key string, body []byte) error
	CloseF   func() error

	PublishCallCount int
	CloseCallCount   int
}

func (c *MockClient) Publish(key string, body []byte) error {
	c.PublishCallCount++
	return c.PublishF(key, body)
}

func (c *MockClient) Close() error {
	c.CloseCallCount++
	return c.CloseF()
}

func NewMockClient() Client {
	return &MockClient{
		PublishF: func(key string, body []byte) error {
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
		errFunc func(t *testing.T, output *AMQP, err error)
	}{
		{
			name: "defaults",
			output: &AMQP{
				Brokers:            []string{DefaultURL},
				ExchangeType:       DefaultExchangeType,
				ExchangeDurability: "durable",
				AuthMethod:         DefaultAuthMethod,
				Database:           DefaultDatabase,
				RetentionPolicy:    DefaultRetentionPolicy,
				Timeout:            config.Duration(time.Second * 5),
				connect: func(_ *ClientConfig) (Client, error) {
					return NewMockClient(), nil
				},
			},
			errFunc: func(t *testing.T, output *AMQP, err error) {
				cfg := output.config
				require.Equal(t, []string{DefaultURL}, cfg.brokers)
				require.Equal(t, "", cfg.exchange)
				require.Equal(t, "topic", cfg.exchangeType)
				require.Equal(t, false, cfg.exchangePassive)
				require.Equal(t, true, cfg.exchangeDurable)
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
				connect: func(_ *ClientConfig) (Client, error) {
					return NewMockClient(), nil
				},
			},
			errFunc: func(t *testing.T, output *AMQP, err error) {
				cfg := output.config
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
				connect: func(_ *ClientConfig) (Client, error) {
					return NewMockClient(), nil
				},
			},
			errFunc: func(t *testing.T, output *AMQP, err error) {
				cfg := output.config
				require.Equal(t, amqp.Table{
					"foo": "bar",
				}, cfg.exchangeArguments)
				require.NoError(t, err)
			},
		},
		{
			name: "username password",
			output: &AMQP{
				URL:      "amqp://foo:bar@localhost",
				Username: "telegraf",
				Password: "pa$$word",
				connect: func(_ *ClientConfig) (Client, error) {
					return NewMockClient(), nil
				},
			},
			errFunc: func(t *testing.T, output *AMQP, err error) {
				cfg := output.config
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
				URL: DefaultURL,
				connect: func(_ *ClientConfig) (Client, error) {
					return NewMockClient(), nil
				},
			},
			errFunc: func(t *testing.T, output *AMQP, err error) {
				cfg := output.config
				require.Equal(t, []string{DefaultURL}, cfg.brokers)
				require.NoError(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.output.Connect()
			tt.errFunc(t, tt.output, err)
		})
	}
}
