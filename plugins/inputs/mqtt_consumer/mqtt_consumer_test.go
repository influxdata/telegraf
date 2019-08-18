package mqtt_consumer

import (
	"testing"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

type FakeClient struct {
	ConnectF           func() mqtt.Token
	SubscribeMultipleF func(filters map[string]byte, callback mqtt.MessageHandler) mqtt.Token
	AddRouteF          func(topic string, callback mqtt.MessageHandler)
	DisconnectF        func(quiesce uint)

	connectCallCount    int
	subscribeCallCount  int
	addRouteCallCount   int
	disconnectCallCount int
}

func (c *FakeClient) Connect() mqtt.Token {
	c.connectCallCount++
	return c.ConnectF()
}

func (c *FakeClient) SubscribeMultiple(filters map[string]byte, callback mqtt.MessageHandler) mqtt.Token {
	c.subscribeCallCount++
	return c.SubscribeMultipleF(filters, callback)
}

func (c *FakeClient) AddRoute(topic string, callback mqtt.MessageHandler) {
	c.addRouteCallCount++
	c.AddRouteF(topic, callback)
}

func (c *FakeClient) Disconnect(quiesce uint) {
	c.disconnectCallCount++
	c.DisconnectF(quiesce)
}

type FakeParser struct {
}

// FakeParser satisfies parsers.Parser
var _ parsers.Parser = &FakeParser{}

func (p *FakeParser) Parse(buf []byte) ([]telegraf.Metric, error) {
	panic("not implemented")
}

func (p *FakeParser) ParseLine(line string) (telegraf.Metric, error) {
	panic("not implemented")
}

func (p *FakeParser) SetDefaultTags(tags map[string]string) {
	panic("not implemented")
}

type FakeToken struct {
	sessionPresent bool
}

// FakeToken satisfies mqtt.Token
var _ mqtt.Token = &FakeToken{}

func (t *FakeToken) Wait() bool {
	return true
}

func (t *FakeToken) WaitTimeout(time.Duration) bool {
	return true
}

func (t *FakeToken) Error() error {
	return nil
}

func (t *FakeToken) SessionPresent() bool {
	return t.sessionPresent
}

// Test the basic lifecycle transitions of the plugin.
func TestLifecycleSanity(t *testing.T) {
	var acc testutil.Accumulator

	plugin := New(func(o *mqtt.ClientOptions) Client {
		return &FakeClient{
			ConnectF: func() mqtt.Token {
				return &FakeToken{}
			},
			AddRouteF: func(topic string, callback mqtt.MessageHandler) {
			},
			SubscribeMultipleF: func(filters map[string]byte, callback mqtt.MessageHandler) mqtt.Token {
				return &FakeToken{}
			},
			DisconnectF: func(quiesce uint) {
			},
		}
	})
	plugin.Servers = []string{"tcp://127.0.0.1"}

	parser := &FakeParser{}
	plugin.SetParser(parser)

	err := plugin.Init()
	require.NoError(t, err)

	err = plugin.Start(&acc)
	require.NoError(t, err)

	err = plugin.Gather(&acc)
	require.NoError(t, err)

	plugin.Stop()
}

// Test that default client has random ID
func TestRandomClientID(t *testing.T) {
	var err error

	m1 := New(nil)
	err = m1.Init()
	require.NoError(t, err)

	m2 := New(nil)
	err = m2.Init()
	require.NoError(t, err)

	require.NotEqual(t, m1.opts.ClientID, m2.opts.ClientID)
}

// PersistentSession requires ClientID
func TestPersistentClientIDFail(t *testing.T) {
	plugin := New(nil)
	plugin.PersistentSession = true

	err := plugin.Init()
	require.Error(t, err)
}

func TestAddRouteCalledForEachTopic(t *testing.T) {
	client := &FakeClient{
		ConnectF: func() mqtt.Token {
			return &FakeToken{}
		},
		AddRouteF: func(topic string, callback mqtt.MessageHandler) {
		},
		SubscribeMultipleF: func(filters map[string]byte, callback mqtt.MessageHandler) mqtt.Token {
			return &FakeToken{}
		},
		DisconnectF: func(quiesce uint) {
		},
	}
	plugin := New(func(o *mqtt.ClientOptions) Client {
		return client
	})
	plugin.Topics = []string{"a", "b"}

	err := plugin.Init()
	require.NoError(t, err)

	var acc testutil.Accumulator
	err = plugin.Start(&acc)
	require.NoError(t, err)

	plugin.Stop()

	require.Equal(t, client.addRouteCallCount, 2)
}

func TestSubscribeCalledIfNoSession(t *testing.T) {
	client := &FakeClient{
		ConnectF: func() mqtt.Token {
			return &FakeToken{}
		},
		AddRouteF: func(topic string, callback mqtt.MessageHandler) {
		},
		SubscribeMultipleF: func(filters map[string]byte, callback mqtt.MessageHandler) mqtt.Token {
			return &FakeToken{}
		},
		DisconnectF: func(quiesce uint) {
		},
	}
	plugin := New(func(o *mqtt.ClientOptions) Client {
		return client
	})
	plugin.Topics = []string{"b"}

	err := plugin.Init()
	require.NoError(t, err)

	var acc testutil.Accumulator
	err = plugin.Start(&acc)
	require.NoError(t, err)

	plugin.Stop()

	require.Equal(t, client.subscribeCallCount, 1)
}

func TestSubscribeNotCalledIfSession(t *testing.T) {
	client := &FakeClient{
		ConnectF: func() mqtt.Token {
			return &FakeToken{sessionPresent: true}
		},
		AddRouteF: func(topic string, callback mqtt.MessageHandler) {
		},
		SubscribeMultipleF: func(filters map[string]byte, callback mqtt.MessageHandler) mqtt.Token {
			return &FakeToken{}
		},
		DisconnectF: func(quiesce uint) {
		},
	}
	plugin := New(func(o *mqtt.ClientOptions) Client {
		return client
	})
	plugin.Topics = []string{"b"}

	err := plugin.Init()
	require.NoError(t, err)

	var acc testutil.Accumulator
	err = plugin.Start(&acc)
	require.NoError(t, err)

	plugin.Stop()

	require.Equal(t, client.subscribeCallCount, 0)
}
