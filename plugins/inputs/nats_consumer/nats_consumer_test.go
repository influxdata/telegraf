package nats_consumer

import (
	"fmt"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/parsing"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"
)

type FakeClient struct {
	url             string
	closed          bool
	queueSubscribeF func(sub, group string, callback nats.MsgHandler) (*nats.Subscription, error)

	closeCalled          int
	queueSubscribeCalled int
	jetStreamCalled      int
}

func (f *FakeClient) QueueSubscribe(subj string, group string, callback nats.MsgHandler) (*nats.Subscription, error) {
	f.queueSubscribeCalled++
	return f.queueSubscribeF(subj, group, callback)
}

func (f *FakeClient) JetStream(...nats.JSOpt) (nats.JetStreamContext, error) {
	f.jetStreamCalled++
	return &FakeJetStreamContext{}, nil
}

//revive:disable:var-naming
func (f *FakeClient) ConnectedUrl() string {
	return f.url
}

//revive:enable:var-naming

func (f *FakeClient) IsClosed() bool {
	return f.closed
}

func (f *FakeClient) Close() {
	f.closeCalled++
}

func TestQueueSubscription(t *testing.T) {
	fc := &FakeClient{
		url:    "tcp://127.0.0.1:8444",
		closed: false,
		queueSubscribeF: func(sub, group string, callback nats.MsgHandler) (*nats.Subscription, error) {
			return &nats.Subscription{
				Subject: sub,
				Queue:   group,
			}, nil
		},
	}

	plugin := New(func(url string, opts ...nats.Option) (client, error) {
		return fc, nil
	}, func(subscription *nats.Subscription, msgLimit, bytesLimit int) error {
		return nil
	})

	plugin.Log = testutil.Logger{}
	plugin.Subjects = []string{"one", "two"}

	err := plugin.Init()
	require.NoError(t, err)

	var acc testutil.Accumulator
	err = plugin.Start(&acc)
	require.NoError(t, err)

	plugin.Stop()
	require.Equal(t, fc.queueSubscribeCalled, 2)
}

func TestLifecycleSanity(t *testing.T) {
	var acc testutil.Accumulator

	fc := &FakeClient{
		url:    "tcp://127.0.0.1:8444",
		closed: false,
		queueSubscribeF: func(sub, group string, callback nats.MsgHandler) (*nats.Subscription, error) {
			return &nats.Subscription{
				Subject: sub,
				Queue:   group,
			}, nil
		},
	}

	plugin := New(func(url string, opts ...nats.Option) (client, error) {
		return fc, nil
	}, func(subscription *nats.Subscription, msgLimit, bytesLimit int) error {
		return nil
	})

	plugin.Log = testutil.Logger{}
	plugin.Servers = []string{"tcp://127.0.0.1:8444"}

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

func TestMessageParsing(t *testing.T) {
	tests := []struct {
		name           string
		subject        string
		subjectTag     func() *string
		expectedError  error
		subjectParsing []SubjectParsingConfig
		expected       []telegraf.Metric
	}{
		{
			name:    "no additional tag if subject tag is not set for backwards compatibility",
			subject: "telegraf",
			subjectTag: func() *string {
				return nil
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"time_idle": 42,
					},
					time.Unix(0, 0)),
			},
		},
		{
			name:    "subject tag is used when set",
			subject: "telegraf",
			subjectTag: func() *string {
				tag := "subject_tag"
				return &tag
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"subject_tag": "telegraf",
					},
					map[string]interface{}{
						"time_idle": 42,
					},
					time.Unix(0, 0)),
			},
		},
		{
			name:    "no subject tag is added when subject tag is set to empty string",
			subject: "telegraf",
			subjectTag: func() *string {
				tag := ""
				return &tag
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"time_idle": 42,
					},
					time.Unix(0, 0)),
			},
		},
		{
			name:    "subject parsing configured",
			subject: "telegraf.123.test",
			subjectTag: func() *string {
				tag := ""
				return &tag
			},
			subjectParsing: []SubjectParsingConfig{
				{
					ConfigEntry: parsing.ConfigEntry{
						Measurement: "_._.measurement",
						Tags:        "testTag._._",
						Fields:      "_.testNumber._",
						FieldTypes: map[string]string{
							"testNumber": "int",
						},
					},
					Base: "telegraf.123.test",
				},
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"test",
					map[string]string{
						"testTag": "telegraf",
					},
					map[string]interface{}{
						"testNumber": 123,
						"time_idle":  42,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name:    "subject parsing configured with the nats wildcard `*`",
			subject: "telegraf.123.test.hello",
			subjectTag: func() *string {
				tag := ""
				return &tag
			},
			subjectParsing: []SubjectParsingConfig{
				{
					ConfigEntry: parsing.ConfigEntry{
						Measurement: "_._.measurement._",
						Tags:        "testTag._._._",
						Fields:      "_.testNumber._.testString",
						FieldTypes: map[string]string{
							"testNumber": "int",
						},
					},
					Base: "telegraf.*.test.hello",
				},
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"test",
					map[string]string{
						"testTag": "telegraf",
					},
					map[string]interface{}{
						"testNumber": 123,
						"testString": "hello",
						"time_idle":  42,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name:    "subject parsing configured incorrectly with invalid fields length",
			subject: "telegraf.123.test.hello",
			subjectTag: func() *string {
				tag := ""
				return &tag
			},
			expectedError: fmt.Errorf("config error subject parsing: fields length does not equal subject length"),
			subjectParsing: []SubjectParsingConfig{
				{
					ConfigEntry: parsing.ConfigEntry{
						Measurement: "_._.measurement._",
						Tags:        "testTag._._._",
						Fields:      "_._.testNumber._.testString",
						FieldTypes: map[string]string{
							"testNumber": "int",
						},
					},
					Base: "telegraf.*.test.hello",
				},
			},
		},
		{
			name:    "subject parsing configured incorrectly with invalid tags length",
			subject: "telegraf.123.test.hello",
			subjectTag: func() *string {
				tag := ""
				return &tag
			},
			expectedError: fmt.Errorf("config error subject parsing: tags length does not equal subject length"),
			subjectParsing: []SubjectParsingConfig{
				{
					ConfigEntry: parsing.ConfigEntry{
						Measurement: "_._.measurement._",
						Tags:        "testTag._._._._",
						Fields:      "_._.testNumber.testString",
						FieldTypes: map[string]string{
							"testNumber": "int",
						},
					},
					Base: "telegraf.*.test.hello",
				},
			},
		},
		{
			name:    "subject parsing configured incorrectly with invalid measurement length",
			subject: "telegraf.123.test.hello",
			subjectTag: func() *string {
				tag := ""
				return &tag
			},
			expectedError: fmt.Errorf("config error subject parsing: measurement length 5 does not equal subject length 4"),
			subjectParsing: []SubjectParsingConfig{
				{
					ConfigEntry: parsing.ConfigEntry{
						Measurement: "_._.measurement._._",
						Tags:        "testTag._._._",
						Fields:      "_._.testNumber.testString",
						FieldTypes: map[string]string{
							"testNumber": "int",
						},
					},
					Base: "telegraf.*.test.hello",
				},
			},
		},
		{
			name:    "subject parsing configured without fields",
			subject: "telegraf.123.test.hello",
			subjectTag: func() *string {
				tag := ""
				return &tag
			},
			subjectParsing: []SubjectParsingConfig{
				{
					ConfigEntry: parsing.ConfigEntry{
						Measurement: "_._.measurement._",
						Tags:        "testTag._._._",
						FieldTypes: map[string]string{
							"testNumber": "int",
						},
					},
					Base: "telegraf.*.test.hello",
				},
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"test",
					map[string]string{
						"testTag": "telegraf",
					},
					map[string]interface{}{
						"time_idle": 42,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name:    "subject parsing configured without measurement",
			subject: "telegraf.123.test.hello",
			subjectTag: func() *string {
				tag := ""
				return &tag
			},
			subjectParsing: []SubjectParsingConfig{
				{
					ConfigEntry: parsing.ConfigEntry{
						Tags:   "testTag._._._",
						Fields: "_.testNumber._.testString",
						FieldTypes: map[string]string{
							"testNumber": "int",
						},
					},
					Base: "telegraf.*.test.hello",
				},
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"testTag": "telegraf",
					},
					map[string]interface{}{
						"testNumber": 123,
						"testString": "hello",
						"time_idle":  42,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name:    "subject parsing configured without tags",
			subject: "telegraf.123.test.hello",
			subjectTag: func() *string {
				tag := ""
				return &tag
			},
			subjectParsing: []SubjectParsingConfig{
				{
					ConfigEntry: parsing.ConfigEntry{
						Measurement: "_._.measurement._",
						Fields:      "_.testNumber._.testString",
						FieldTypes: map[string]string{
							"testNumber": "int",
						},
					},
					Base: "telegraf.*.test.hello",
				},
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"test",
					map[string]string{},
					map[string]interface{}{
						"testNumber": 123,
						"testString": "hello",
						"time_idle":  42,
					},
					time.Unix(0, 0),
				),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var handler nats.MsgHandler
			fc := &FakeClient{
				queueSubscribeF: func(sub, group string, callback nats.MsgHandler) (*nats.Subscription, error) {
					handler = callback
					return &nats.Subscription{
						Subject: sub,
						Queue:   group,
					}, nil
				},
			}

			plugin := New(func(url string, opts ...nats.Option) (client, error) {
				return fc, nil
			}, func(subscription *nats.Subscription, msgLimit, bytesLimit int) error {
				return nil
			})
			plugin.Log = testutil.Logger{}
			plugin.Subjects = []string{tt.subject}
			plugin.SubjectTag = tt.subjectTag()
			plugin.SubjectParsing = tt.subjectParsing

			parser := &influx.Parser{}
			require.NoError(t, parser.Init())
			plugin.SetParser(parser)

			err := plugin.Init()
			require.Equal(t, tt.expectedError, err)
			if tt.expectedError != nil {
				return
			}

			var acc testutil.Accumulator
			err = plugin.Start(&acc)
			require.NoError(t, err)

			var msg nats.Msg
			msg.Subject = tt.subject
			msg.Data = []byte("cpu time_idle=42i")
			handler(&msg)

			plugin.Stop()
			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, tt.expected, actual, testutil.IgnoreTime())
		})
	}
}

func TestConfig(t *testing.T) {
	c := config.NewConfig()
	require.NoError(t, c.LoadConfig("./testdata/config.toml"))

	plugin, ok := c.Inputs[0].Input.(*natsConsumer)
	require.True(t, ok)
	require.Len(t, plugin.SubjectParsing, 1)
	require.Equal(t, "telegraf.one.cpu.23", plugin.SubjectParsing[0].Base)
	require.Equal(t, "_._.measurement._", plugin.SubjectParsing[0].Measurement)
	require.Equal(t, "tag._._._", plugin.SubjectParsing[0].Tags)
	require.Equal(t, "_._._.test", plugin.SubjectParsing[0].Fields)
	require.Len(t, plugin.SubjectParsing[0].FieldTypes, 1)
	require.Equal(t, "int", plugin.SubjectParsing[0].FieldTypes["test"])
}

// FakeParser satisfies parsers.Parser
var _ parsers.Parser = &FakeParser{}

type FakeParser struct {
}

func (p *FakeParser) Parse([]byte) ([]telegraf.Metric, error) {
	panic("implement me")
}

func (p *FakeParser) ParseLine(string) (telegraf.Metric, error) {
	panic("implement me")
}

func (p *FakeParser) SetDefaultTags(map[string]string) {
	panic("implement me")
}

type FakeJetStreamContext struct {
}

func (f FakeJetStreamContext) Publish(string, []byte, ...nats.PubOpt) (*nats.PubAck, error) {
	panic("implement me")
}

func (f FakeJetStreamContext) PublishMsg(*nats.Msg, ...nats.PubOpt) (*nats.PubAck, error) {
	panic("implement me")
}

func (f FakeJetStreamContext) PublishAsync(string, []byte, ...nats.PubOpt) (nats.PubAckFuture, error) {
	panic("implement me")
}

func (f FakeJetStreamContext) PublishMsgAsync(*nats.Msg, ...nats.PubOpt) (nats.PubAckFuture, error) {
	panic("implement me")
}

func (f FakeJetStreamContext) PublishAsyncPending() int {
	panic("implement me")
}

func (f FakeJetStreamContext) PublishAsyncComplete() <-chan struct{} {
	panic("implement me")
}

func (f FakeJetStreamContext) Subscribe(string, nats.MsgHandler, ...nats.SubOpt) (*nats.Subscription, error) {
	panic("implement me")
}

func (f FakeJetStreamContext) SubscribeSync(string, ...nats.SubOpt) (*nats.Subscription, error) {
	panic("implement me")
}

func (f FakeJetStreamContext) ChanSubscribe(string, chan *nats.Msg, ...nats.SubOpt) (*nats.Subscription, error) {
	panic("implement me")
}

func (f FakeJetStreamContext) ChanQueueSubscribe(string, string, chan *nats.Msg, ...nats.SubOpt) (*nats.Subscription, error) {
	panic("implement me")
}

func (f FakeJetStreamContext) QueueSubscribe(string, string, nats.MsgHandler, ...nats.SubOpt) (*nats.Subscription, error) {
	panic("implement me")
}

func (f FakeJetStreamContext) QueueSubscribeSync(string, string, ...nats.SubOpt) (*nats.Subscription, error) {
	panic("implement me")
}

func (f FakeJetStreamContext) PullSubscribe(string, string, ...nats.SubOpt) (*nats.Subscription, error) {
	panic("implement me")
}

func (f FakeJetStreamContext) AddStream(*nats.StreamConfig, ...nats.JSOpt) (*nats.StreamInfo, error) {
	panic("implement me")
}

func (f FakeJetStreamContext) UpdateStream(*nats.StreamConfig, ...nats.JSOpt) (*nats.StreamInfo, error) {
	panic("implement me")
}

func (f FakeJetStreamContext) DeleteStream(string, ...nats.JSOpt) error {
	panic("implement me")
}

func (f FakeJetStreamContext) StreamInfo(string, ...nats.JSOpt) (*nats.StreamInfo, error) {
	panic("implement me")
}

func (f FakeJetStreamContext) PurgeStream(string, ...nats.JSOpt) error {
	panic("implement me")
}

func (f FakeJetStreamContext) StreamsInfo(...nats.JSOpt) <-chan *nats.StreamInfo {
	panic("implement me")
}

func (f FakeJetStreamContext) Streams(...nats.JSOpt) <-chan *nats.StreamInfo {
	panic("implement me")
}

func (f FakeJetStreamContext) StreamNames(...nats.JSOpt) <-chan string {
	panic("implement me")
}

func (f FakeJetStreamContext) GetMsg(string, uint64, ...nats.JSOpt) (*nats.RawStreamMsg, error) {
	panic("implement me")
}

func (f FakeJetStreamContext) GetLastMsg(string, string, ...nats.JSOpt) (*nats.RawStreamMsg, error) {
	panic("implement me")
}

func (f FakeJetStreamContext) DeleteMsg(string, uint64, ...nats.JSOpt) error {
	panic("implement me")
}

func (f FakeJetStreamContext) SecureDeleteMsg(string, uint64, ...nats.JSOpt) error {
	panic("implement me")
}

func (f FakeJetStreamContext) AddConsumer(string, *nats.ConsumerConfig, ...nats.JSOpt) (*nats.ConsumerInfo, error) {
	panic("implement me")
}

func (f FakeJetStreamContext) UpdateConsumer(string, *nats.ConsumerConfig, ...nats.JSOpt) (*nats.ConsumerInfo, error) {
	panic("implement me")
}

func (f FakeJetStreamContext) DeleteConsumer(string, string, ...nats.JSOpt) error {
	panic("implement me")
}

func (f FakeJetStreamContext) ConsumerInfo(string, string, ...nats.JSOpt) (*nats.ConsumerInfo, error) {
	panic("implement me")
}

func (f FakeJetStreamContext) ConsumersInfo(string, ...nats.JSOpt) <-chan *nats.ConsumerInfo {
	panic("implement me")
}

func (f FakeJetStreamContext) Consumers(string, ...nats.JSOpt) <-chan *nats.ConsumerInfo {
	panic("implement me")
}

func (f FakeJetStreamContext) ConsumerNames(string, ...nats.JSOpt) <-chan string {
	panic("implement me")
}

func (f FakeJetStreamContext) AccountInfo(...nats.JSOpt) (*nats.AccountInfo, error) {
	panic("implement me")
}

func (f FakeJetStreamContext) KeyValue(string) (nats.KeyValue, error) {
	panic("implement me")
}

func (f FakeJetStreamContext) CreateKeyValue(*nats.KeyValueConfig) (nats.KeyValue, error) {
	panic("implement me")
}

func (f FakeJetStreamContext) DeleteKeyValue(string) error {
	panic("implement me")
}

func (f FakeJetStreamContext) KeyValueStoreNames() <-chan string {
	panic("implement me")
}

func (f FakeJetStreamContext) KeyValueStores() <-chan nats.KeyValueStatus {
	panic("implement me")
}

func (f FakeJetStreamContext) ObjectStore(string) (nats.ObjectStore, error) {
	panic("implement me")
}

func (f FakeJetStreamContext) CreateObjectStore(*nats.ObjectStoreConfig) (nats.ObjectStore, error) {
	panic("implement me")
}

func (f FakeJetStreamContext) DeleteObjectStore(string) error {
	panic("implement me")
}

func (f FakeJetStreamContext) ObjectStoreNames(...nats.ObjectOpt) <-chan string {
	panic("implement me")
}

func (f FakeJetStreamContext) ObjectStores(...nats.ObjectOpt) <-chan nats.ObjectStoreStatus {
	panic("implement me")
}
