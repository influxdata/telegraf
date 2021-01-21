package amqp

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf/internal"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/require"
)

type MockClient struct {
	PublishF func(key string, body []byte) error
	CloseF   func() error

	PublishCallCount int
	CloseCallCount   int

	t *testing.T
}

func (c *MockClient) Publish(key string, body []byte) error {
	c.PublishCallCount++
	return c.PublishF(key, body)
}

func (c *MockClient) Close() error {
	c.CloseCallCount++
	return c.CloseF()
}

func MockConnect(config *ClientConfig) (Client, error) {
	return &MockClient{}, nil
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
				Timeout:            internal.Duration{Duration: time.Second * 5},
				connect: func(config *ClientConfig) (Client, error) {
					return NewMockClient(), nil
				},
			},
			errFunc: func(t *testing.T, output *AMQP, err error) {
				config := output.config
				require.Equal(t, []string{DefaultURL}, config.brokers)
				require.Equal(t, "", config.exchange)
				require.Equal(t, "topic", config.exchangeType)
				require.Equal(t, false, config.exchangePassive)
				require.Equal(t, true, config.exchangeDurable)
				require.Equal(t, amqp.Table(nil), config.exchangeArguments)
				require.Equal(t, amqp.Table{
					"database":         DefaultDatabase,
					"retention_policy": DefaultRetentionPolicy,
				}, config.headers)
				require.Equal(t, amqp.Transient, config.deliveryMode)
				require.NoError(t, err)
			},
		},
		{
			name: "headers overrides deprecated dbrp",
			output: &AMQP{
				Headers: map[string]string{
					"foo": "bar",
				},
				connect: func(config *ClientConfig) (Client, error) {
					return NewMockClient(), nil
				},
			},
			errFunc: func(t *testing.T, output *AMQP, err error) {
				config := output.config
				require.Equal(t, amqp.Table{
					"foo": "bar",
				}, config.headers)
				require.NoError(t, err)
			},
		},
		{
			name: "exchange args",
			output: &AMQP{
				ExchangeArguments: map[string]string{
					"foo": "bar",
				},
				connect: func(config *ClientConfig) (Client, error) {
					return NewMockClient(), nil
				},
			},
			errFunc: func(t *testing.T, output *AMQP, err error) {
				config := output.config
				require.Equal(t, amqp.Table{
					"foo": "bar",
				}, config.exchangeArguments)
				require.NoError(t, err)
			},
		},
		{
			name: "username password",
			output: &AMQP{
				URL:      "amqp://foo:bar@localhost",
				Username: "telegraf",
				Password: "pa$$word",
				connect: func(config *ClientConfig) (Client, error) {
					return NewMockClient(), nil
				},
			},
			errFunc: func(t *testing.T, output *AMQP, err error) {
				config := output.config
				require.Equal(t, []amqp.Authentication{
					&amqp.PlainAuth{
						Username: "telegraf",
						Password: "pa$$word",
					},
				}, config.auth)

				require.NoError(t, err)
			},
		},
		{
			name: "url support",
			output: &AMQP{
				URL: DefaultURL,
				connect: func(config *ClientConfig) (Client, error) {
					return NewMockClient(), nil
				},
			},
			errFunc: func(t *testing.T, output *AMQP, err error) {
				config := output.config
				require.Equal(t, []string{DefaultURL}, config.brokers)
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
