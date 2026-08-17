package cloud_pubsub

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"cloud.google.com/go/pubsub/v2"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/support/bundler"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	serializers_influx "github.com/influxdata/telegraf/plugins/serializers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestPubSub_WriteSingle(t *testing.T) {
	metrics := []telegraf.Metric{testutil.TestMetric("value_1", "test")}
	tpc := &stubTopic{
		T:               t,
		ReturnErr:       make(map[string]bool),
		published:       make(map[string]*pubsub.Message),
		ContentEncoding: "identity",
	}

	// Setup the plugin
	settings := pubsub.DefaultPublishSettings
	plugin := &PubSub{
		Project:               "test-project",
		Topic:                 "test-topic",
		PublishCountThreshold: 1,
		PublishByteThreshold:  settings.ByteThreshold,
		PublishNumGoroutines:  settings.NumGoroutines,
		PublishTimeout:        config.Duration(settings.Timeout),
		ContentEncoding:       "identity",
		stubTopic:             func(string) topic { return tpc },
	}
	require.NoError(t, plugin.Init())

	// Instantiate a Influx line-protocol serializer
	serializer := &serializers_influx.Serializer{}
	require.NoError(t, serializer.Init())
	plugin.SetSerializer(serializer)

	// Write the metric and check
	require.NoError(t, plugin.Write(metrics))

	expected := map[string]string{
		"value_1": "test,tag1=value1 value=\"value_1\" 1257894000000000000\n",
		"value_2": "test,tag1=value1 value=\"value_2\" 1257894000000000000\n",
	}
	for name, msg := range tpc.published {
		require.Equalf(t, expected[name], string(msg.Data), "mismatch for %q", name)
	}
}

func TestPubSub_WriteWithAttribute(t *testing.T) {
	metrics := []telegraf.Metric{testutil.TestMetric("value_1", "test")}
	tpc := &stubTopic{
		T:               t,
		ReturnErr:       make(map[string]bool),
		published:       make(map[string]*pubsub.Message),
		ContentEncoding: "identity",
	}

	// Setup the plugin
	settings := pubsub.DefaultPublishSettings
	plugin := &PubSub{
		Project:               "test-project",
		Topic:                 "test-topic",
		PublishCountThreshold: settings.CountThreshold,
		PublishByteThreshold:  settings.ByteThreshold,
		PublishNumGoroutines:  settings.NumGoroutines,
		PublishTimeout:        config.Duration(settings.Timeout),
		ContentEncoding:       "identity",
		Attributes: map[string]string{
			"foo1": "bar1",
			"foo2": "bar2",
		},
		stubTopic: func(string) topic { return tpc },
	}
	require.NoError(t, plugin.Init())

	// Instantiate a Influx line-protocol serializer
	serializer := &serializers_influx.Serializer{}
	require.NoError(t, serializer.Init())
	plugin.SetSerializer(serializer)

	// Write the metric and check
	require.NoError(t, plugin.Write(metrics))

	expected := map[string]string{
		"value_1": "test,tag1=value1 value=\"value_1\" 1257894000000000000\n",
		"value_2": "test,tag1=value1 value=\"value_2\" 1257894000000000000\n",
	}
	expectedAttr := map[string]string{
		"foo1": "bar1",
		"foo2": "bar2",
	}
	for name, msg := range tpc.published {
		require.Equalf(t, expected[name], string(msg.Data), "mismatch for %q", name)
		require.Equalf(t, expectedAttr, msg.Attributes, "attribute mismatch for %q", name)
	}
}

func TestPubSub_WriteMultiple(t *testing.T) {
	metrics := []telegraf.Metric{
		testutil.TestMetric("value_1", "test"),
		testutil.TestMetric("value_2", "test"),
	}

	tpc := &stubTopic{
		T:               t,
		ReturnErr:       make(map[string]bool),
		published:       make(map[string]*pubsub.Message),
		ContentEncoding: "identity",
	}

	// Setup the plugin
	settings := pubsub.DefaultPublishSettings
	plugin := &PubSub{
		Project:               "test-project",
		Topic:                 "test-topic",
		PublishCountThreshold: settings.CountThreshold,
		PublishByteThreshold:  settings.ByteThreshold,
		PublishNumGoroutines:  settings.NumGoroutines,
		PublishTimeout:        config.Duration(settings.Timeout),
		ContentEncoding:       "identity",
		stubTopic:             func(string) topic { return tpc },
	}
	require.NoError(t, plugin.Init())

	// Instantiate a Influx line-protocol serializer
	serializer := &serializers_influx.Serializer{}
	require.NoError(t, serializer.Init())
	plugin.SetSerializer(serializer)

	// Write the metric and check
	require.NoError(t, plugin.Write(metrics))

	expected := map[string]string{
		"value_1": "test,tag1=value1 value=\"value_1\" 1257894000000000000\n",
		"value_2": "test,tag1=value1 value=\"value_2\" 1257894000000000000\n",
	}
	for name, msg := range tpc.published {
		require.Equalf(t, expected[name], string(msg.Data), "mismatch for %q", name)
	}
	require.EqualValues(t, 1, tpc.bundleCount.Load(), "unexpected bundle count")
}

func TestPubSub_WriteOverCountThreshold(t *testing.T) {
	metrics := []telegraf.Metric{
		testutil.TestMetric("value_1", "test"),
		testutil.TestMetric("value_2", "test"),
		testutil.TestMetric("value_3", "test"),
		testutil.TestMetric("value_4", "test"),
	}

	tpc := &stubTopic{
		T:               t,
		ReturnErr:       make(map[string]bool),
		published:       make(map[string]*pubsub.Message),
		ContentEncoding: "identity",
	}

	// Setup the plugin
	settings := pubsub.DefaultPublishSettings
	plugin := &PubSub{
		Project:               "test-project",
		Topic:                 "test-topic",
		PublishCountThreshold: 2,
		PublishByteThreshold:  settings.ByteThreshold,
		PublishNumGoroutines:  settings.NumGoroutines,
		PublishTimeout:        config.Duration(settings.Timeout),
		ContentEncoding:       "identity",
		stubTopic:             func(string) topic { return tpc },
	}
	require.NoError(t, plugin.Init())

	// Instantiate a Influx line-protocol serializer
	serializer := &serializers_influx.Serializer{}
	require.NoError(t, serializer.Init())
	plugin.SetSerializer(serializer)

	// Write the metric and check
	require.NoError(t, plugin.Write(metrics))

	expected := map[string]string{
		"value_1": "test,tag1=value1 value=\"value_1\" 1257894000000000000\n",
		"value_2": "test,tag1=value1 value=\"value_2\" 1257894000000000000\n",
		"value_3": "test,tag1=value1 value=\"value_3\" 1257894000000000000\n",
		"value_4": "test,tag1=value1 value=\"value_4\" 1257894000000000000\n",
	}
	for name, msg := range tpc.published {
		require.Equalf(t, expected[name], string(msg.Data), "mismatch for %q", name)
	}
	require.EqualValues(t, 2, tpc.bundleCount.Load(), "unexpected bundle count")
}

func TestPubSub_WriteOverByteThreshold(t *testing.T) {
	metrics := []telegraf.Metric{
		testutil.TestMetric("value_1", "test"),
		testutil.TestMetric("value_2", "test"),
	}

	tpc := &stubTopic{
		T:               t,
		ReturnErr:       make(map[string]bool),
		published:       make(map[string]*pubsub.Message),
		ContentEncoding: "identity",
	}

	// Setup the plugin
	settings := pubsub.DefaultPublishSettings
	plugin := &PubSub{
		Project:               "test-project",
		Topic:                 "test-topic",
		PublishCountThreshold: 10,
		PublishByteThreshold:  1,
		PublishNumGoroutines:  settings.NumGoroutines,
		PublishTimeout:        config.Duration(settings.Timeout),
		ContentEncoding:       "identity",
		stubTopic:             func(string) topic { return tpc },
	}
	require.NoError(t, plugin.Init())

	// Instantiate a Influx line-protocol serializer
	serializer := &serializers_influx.Serializer{}
	require.NoError(t, serializer.Init())
	plugin.SetSerializer(serializer)

	// Write the metric and check
	require.NoError(t, plugin.Write(metrics))

	expected := map[string]string{
		"value_1": "test,tag1=value1 value=\"value_1\" 1257894000000000000\n",
		"value_2": "test,tag1=value1 value=\"value_2\" 1257894000000000000\n",
	}
	for name, msg := range tpc.published {
		require.Equalf(t, expected[name], string(msg.Data), "mismatch for %q", name)
	}
	require.EqualValues(t, 2, tpc.bundleCount.Load(), "unexpected bundle count")
}

func TestPubSub_WriteBase64Single(t *testing.T) {
	metrics := []telegraf.Metric{
		testutil.TestMetric("value_1", "test"),
		testutil.TestMetric("value_2", "test"),
	}

	tpc := &stubTopic{
		T:               t,
		ReturnErr:       make(map[string]bool),
		published:       make(map[string]*pubsub.Message),
		ContentEncoding: "identity",
		Base64Data:      true,
	}

	// Setup the plugin
	settings := pubsub.DefaultPublishSettings
	plugin := &PubSub{
		Project:               "test-project",
		Topic:                 "test-topic",
		PublishCountThreshold: 1,
		PublishByteThreshold:  settings.ByteThreshold,
		PublishNumGoroutines:  settings.NumGoroutines,
		PublishTimeout:        config.Duration(settings.Timeout),
		ContentEncoding:       "identity",
		Base64Data:            true,
		stubTopic:             func(string) topic { return tpc },
	}
	require.NoError(t, plugin.Init())

	// Instantiate a Influx line-protocol serializer
	serializer := &serializers_influx.Serializer{}
	require.NoError(t, serializer.Init())
	plugin.SetSerializer(serializer)

	// Write the metric and check
	require.NoError(t, plugin.Write(metrics))

	expected := map[string]string{
		"value_1": "dGVzdCx0YWcxPXZhbHVlMSB2YWx1ZT0idmFsdWVfMSIgMTI1Nzg5NDAwMDAwMDAwMDAwMAo=",
		"value_2": "dGVzdCx0YWcxPXZhbHVlMSB2YWx1ZT0idmFsdWVfMiIgMTI1Nzg5NDAwMDAwMDAwMDAwMAo=",
	}
	for name, msg := range tpc.published {
		require.Equalf(t, expected[name], string(msg.Data), "mismatch for %q", name)
	}
}

func TestPubSub_Error(t *testing.T) {
	metrics := []telegraf.Metric{
		testutil.TestMetric("value_1", "test"),
		testutil.TestMetric("value_2", "test"),
	}

	tpc := &stubTopic{
		T:               t,
		ReturnErr:       map[string]bool{"value_1": true},
		published:       make(map[string]*pubsub.Message),
		ContentEncoding: "identity",
	}

	// Setup the plugin
	settings := pubsub.DefaultPublishSettings
	plugin := &PubSub{
		Project:               "test-project",
		Topic:                 "test-topic",
		PublishCountThreshold: settings.CountThreshold,
		PublishByteThreshold:  settings.ByteThreshold,
		PublishNumGoroutines:  settings.NumGoroutines,
		PublishTimeout:        config.Duration(settings.Timeout),
		ContentEncoding:       "identity",
		stubTopic:             func(string) topic { return tpc },
	}
	require.NoError(t, plugin.Init())

	// Instantiate a Influx line-protocol serializer
	serializer := &serializers_influx.Serializer{}
	require.NoError(t, serializer.Init())
	plugin.SetSerializer(serializer)

	require.ErrorContains(t, plugin.Write(metrics), "this is an error")
}

func TestPubSub_WriteGzipSingle(t *testing.T) {
	metrics := []telegraf.Metric{
		testutil.TestMetric("value_1", "test"),
		testutil.TestMetric("value_2", "test"),
	}

	tpc := &stubTopic{
		T:               t,
		ReturnErr:       make(map[string]bool),
		published:       make(map[string]*pubsub.Message),
		ContentEncoding: "gzip",
	}

	// Setup the plugin
	settings := pubsub.DefaultPublishSettings
	plugin := &PubSub{
		Project:               "test-project",
		Topic:                 "test-topic",
		PublishCountThreshold: 1,
		PublishByteThreshold:  settings.ByteThreshold,
		PublishNumGoroutines:  settings.NumGoroutines,
		PublishTimeout:        config.Duration(settings.Timeout),
		ContentEncoding:       "gzip",
		stubTopic:             func(string) topic { return tpc },
	}
	require.NoError(t, plugin.Init())

	// Instantiate a Influx line-protocol serializer
	serializer := &serializers_influx.Serializer{}
	require.NoError(t, serializer.Init())
	plugin.SetSerializer(serializer)

	// Write the metric and check
	require.NoError(t, plugin.Write(metrics))

	//nolint:lll // Long data line would not benefit from wrapping
	expected := map[string]string{
		"value_1": "1f8b080000096e8800ff003500caff746573742c746167313d76616c7565312076616c75653d2276616c75655f312220313235373839343030303030303030303030300a03001da5187935000000",
		"value_2": "1f8b080000096e8800ff003500caff746573742c746167313d76616c7565312076616c75653d2276616c75655f322220313235373839343030303030303030303030300a0300209cfd0f35000000",
	}
	for name, msg := range tpc.published {
		require.Equalf(t, expected[name], hex.EncodeToString(msg.Data), "mismatch for %q", name)
	}
}

func TestPubSub_WriteGzipAndBase64Single(t *testing.T) {
	metrics := []telegraf.Metric{
		testutil.TestMetric("value_1", "test"),
		testutil.TestMetric("value_2", "test"),
	}

	tpc := &stubTopic{
		T:               t,
		ReturnErr:       make(map[string]bool),
		published:       make(map[string]*pubsub.Message),
		ContentEncoding: "gzip",
		Base64Data:      true,
	}

	// Setup the plugin
	settings := pubsub.DefaultPublishSettings
	plugin := &PubSub{
		Project:               "test-project",
		Topic:                 "test-topic",
		PublishCountThreshold: 1,
		PublishByteThreshold:  settings.ByteThreshold,
		PublishNumGoroutines:  settings.NumGoroutines,
		PublishTimeout:        config.Duration(settings.Timeout),
		ContentEncoding:       "gzip",
		Base64Data:            true,
		stubTopic:             func(string) topic { return tpc },
	}
	require.NoError(t, plugin.Init())

	// Instantiate a Influx line-protocol serializer
	serializer := &serializers_influx.Serializer{}
	require.NoError(t, serializer.Init())
	plugin.SetSerializer(serializer)

	// Write the metric and check
	require.NoError(t, plugin.Write(metrics))

	//nolint:lll // Long data line would not benefit from wrapping
	expected := map[string]string{
		"value_1": "1f8b080000096e8800ff004800b7ff6447567a644378305957637850585a686248566c4d534232595778315a543069646d4673645756664d5349674d5449314e7a67354e4441774d4441774d4441774d4441774d416f3d03005ec5b25d48000000",
		"value_2": "1f8b080000096e8800ff004800b7ff6447567a644378305957637850585a686248566c4d534232595778315a543069646d4673645756664d6949674d5449314e7a67354e4441774d4441774d4441774d4441774d416f3d030005c3b0de48000000",
	}
	for name, msg := range tpc.published {
		require.Equalf(t, expected[name], hex.EncodeToString(msg.Data), "mismatch for %q", name)
	}
}

type bundledMsg struct {
	*pubsub.Message
	*stubResult
}

type stubResult struct {
	metricIDs []string

	sendError bool
	err       chan error
	done      chan struct{}
}

type stubTopic struct {
	Settings  pubsub.PublishSettings
	ReturnErr map[string]bool
	telegraf.Parser
	*testing.T
	Base64Data      bool
	ContentEncoding string

	stopped bool
	pLock   sync.Mutex

	published map[string]*pubsub.Message

	bundler     *bundler.Bundler
	bLock       sync.Mutex
	bundleCount atomic.Uint32
}

func (*stubTopic) ID() string {
	return "test-topic"
}

func (t *stubTopic) Stop() {
	t.pLock.Lock()
	defer t.pLock.Unlock()

	t.stopped = true
	t.bundler.Flush()
}

func (t *stubTopic) Publish(ctx context.Context, msg *pubsub.Message) publishResult {
	t.pLock.Lock()
	defer t.pLock.Unlock()

	if t.stopped || ctx.Err() != nil {
		t.Fatalf("publish called after stop")
	}

	ids := t.parseIDs(msg)
	r := &stubResult{
		metricIDs: ids,
		err:       make(chan error, 1),
		done:      make(chan struct{}, 1),
	}

	for _, id := range ids {
		_, ok := t.ReturnErr[id]
		r.sendError = r.sendError || ok
	}

	bundled := &bundledMsg{msg, r}
	if err := t.bundler.Add(bundled, len(msg.Data)); err != nil {
		t.Fatalf("unexpected error while adding to bundle: %v", err)
	}
	return r
}

func (t *stubTopic) PublishSettings() pubsub.PublishSettings {
	return t.Settings
}

func (t *stubTopic) SetPublishSettings(settings pubsub.PublishSettings) {
	t.Settings = settings
	t.bundler = bundler.NewBundler(&bundledMsg{}, t.sendBundle())
	t.bundler.DelayThreshold = 10 * time.Second
	t.bundler.BundleCountThreshold = t.Settings.CountThreshold
	if t.bundler.BundleCountThreshold > pubsub.MaxPublishRequestCount {
		t.bundler.BundleCountThreshold = pubsub.MaxPublishRequestCount
	}
	t.bundler.BundleByteThreshold = t.Settings.ByteThreshold
	t.bundler.BundleByteLimit = pubsub.MaxPublishRequestBytes
	t.bundler.HandlerLimit = 25 * runtime.GOMAXPROCS(0)
}

func (r *stubResult) Get(ctx context.Context) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case err := <-r.err:
		return "", err
	case <-r.done:
		return "id-" + r.metricIDs[0], nil
	}
}

func (t *stubTopic) sendBundle() func(items interface{}) {
	return func(items interface{}) {
		t.bLock.Lock()
		defer t.bLock.Unlock()

		bundled := items.([]*bundledMsg)

		for _, msg := range bundled {
			r := msg.stubResult
			for _, id := range r.metricIDs {
				t.published[id] = msg.Message
			}

			if r.sendError {
				r.err <- errors.New("this is an error")
			} else {
				r.done <- struct{}{}
			}
		}

		t.bundleCount.Add(1)
	}
}

func (t *stubTopic) parseIDs(msg *pubsub.Message) []string {
	p := influx.Parser{}
	err := p.Init()
	require.NoError(t, err)

	decoder, err := internal.NewContentDecoder(t.ContentEncoding)
	require.NoError(t, err)
	d, err := decoder.Decode(msg.Data)
	if err != nil {
		t.Errorf("unable to decode message: %v", err)
	}
	if t.Base64Data {
		strData, err := base64.StdEncoding.DecodeString(string(d))
		if err != nil {
			t.Errorf("unable to base64 decode message: %v", err)
		}
		d = strData
	}
	metrics, err := p.Parse(d)
	if err != nil {
		t.Fatalf("unexpected parsing error: %v", err)
	}

	ids := make([]string, 0, len(metrics))
	for _, met := range metrics {
		id, _ := met.GetField("value")
		ids = append(ids, id.(string))
	}
	return ids
}
