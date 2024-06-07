package cloud_pubsub

import (
	"encoding/base64"
	"testing"

	"cloud.google.com/go/pubsub"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestPubSub_WriteSingle(t *testing.T) {
	testMetrics := []testMetric{
		{testutil.TestMetric("value_1", "test"), false /*return error */},
	}

	settings := pubsub.DefaultPublishSettings
	settings.CountThreshold = 1
	ps, topic, metrics := getTestResources(t, settings, testMetrics)

	require.NoError(t, ps.Write(metrics))

	for _, testM := range testMetrics {
		verifyRawMetricPublished(t, testM.m, topic.published)
	}
}

func TestPubSub_WriteWithAttribute(t *testing.T) {
	testMetrics := []testMetric{
		{testutil.TestMetric("value_1", "test"), false /*return error*/},
	}

	settings := pubsub.DefaultPublishSettings
	ps, topic, metrics := getTestResources(t, settings, testMetrics)
	ps.Attributes = map[string]string{
		"foo1": "bar1",
		"foo2": "bar2",
	}

	require.NoError(t, ps.Write(metrics))

	for _, testM := range testMetrics {
		msg := verifyRawMetricPublished(t, testM.m, topic.published)
		require.Equalf(t, "bar1", msg.Attributes["foo1"], "expected attribute foo1=bar1")
		require.Equalf(t, "bar2", msg.Attributes["foo2"], "expected attribute foo2=bar2")
	}
}

func TestPubSub_WriteMultiple(t *testing.T) {
	testMetrics := []testMetric{
		{testutil.TestMetric("value_1", "test"), false /*return error*/},
		{testutil.TestMetric("value_2", "test"), false},
	}

	settings := pubsub.DefaultPublishSettings

	ps, topic, metrics := getTestResources(t, settings, testMetrics)

	require.NoError(t, ps.Write(metrics))

	for _, testM := range testMetrics {
		verifyRawMetricPublished(t, testM.m, topic.published)
	}
	require.Equalf(t, 1, topic.getBundleCount(), "unexpected bundle count")
}

func TestPubSub_WriteOverCountThreshold(t *testing.T) {
	testMetrics := []testMetric{
		{testutil.TestMetric("value_1", "test"), false /*return error*/},
		{testutil.TestMetric("value_2", "test"), false},
		{testutil.TestMetric("value_3", "test"), false},
		{testutil.TestMetric("value_4", "test"), false},
	}

	settings := pubsub.DefaultPublishSettings
	settings.CountThreshold = 2

	ps, topic, metrics := getTestResources(t, settings, testMetrics)

	require.NoError(t, ps.Write(metrics))

	for _, testM := range testMetrics {
		verifyRawMetricPublished(t, testM.m, topic.published)
	}
	require.Equalf(t, 2, topic.getBundleCount(), "unexpected bundle count")
}

func TestPubSub_WriteOverByteThreshold(t *testing.T) {
	testMetrics := []testMetric{
		{testutil.TestMetric("value_1", "test"), false /*return error*/},
		{testutil.TestMetric("value_2", "test"), false},
	}

	settings := pubsub.DefaultPublishSettings
	settings.CountThreshold = 10
	settings.ByteThreshold = 1

	ps, topic, metrics := getTestResources(t, settings, testMetrics)

	require.NoError(t, ps.Write(metrics))

	for _, testM := range testMetrics {
		verifyRawMetricPublished(t, testM.m, topic.published)
	}
	require.Equalf(t, 2, topic.getBundleCount(), "unexpected bundle count")
}

func TestPubSub_WriteBase64Single(t *testing.T) {
	testMetrics := []testMetric{
		{testutil.TestMetric("value_1", "test"), false /*return error */},
		{testutil.TestMetric("value_2", "test"), false},
	}

	settings := pubsub.DefaultPublishSettings
	settings.CountThreshold = 1
	ps, topic, metrics := getTestResources(t, settings, testMetrics)
	ps.Base64Data = true
	topic.Base64Data = true

	require.NoError(t, ps.Write(metrics))

	for _, testM := range testMetrics {
		verifyMetricPublished(t, testM.m, topic.published, true /* base64encoded */, false /* gzipEncoded */)
	}
}

func TestPubSub_Error(t *testing.T) {
	testMetrics := []testMetric{
		// Force this batch to return error
		{testutil.TestMetric("value_1", "test"), true},
		{testutil.TestMetric("value_2", "test"), false},
	}

	settings := pubsub.DefaultPublishSettings
	ps, _, metrics := getTestResources(t, settings, testMetrics)

	err := ps.Write(metrics)
	require.Error(t, err)
	require.ErrorContains(t, err, errMockFail)
}

func TestPubSub_WriteGzipSingle(t *testing.T) {
	testMetrics := []testMetric{
		{testutil.TestMetric("value_1", "test"), false /*return error */},
		{testutil.TestMetric("value_2", "test"), false},
	}

	settings := pubsub.DefaultPublishSettings
	settings.CountThreshold = 1
	ps, topic, metrics := getTestResources(t, settings, testMetrics)
	topic.ContentEncoding = "gzip"
	ps.ContentEncoding = "gzip"
	var err error
	ps.encoder, err = internal.NewContentEncoder(ps.ContentEncoding)

	require.NoError(t, err)
	require.NoError(t, ps.Write(metrics))

	for _, testM := range testMetrics {
		verifyMetricPublished(t, testM.m, topic.published, false /* base64encoded */, true /* Gzipencoded */)
	}
}

func TestPubSub_WriteGzipAndBase64Single(t *testing.T) {
	testMetrics := []testMetric{
		{testutil.TestMetric("value_1", "test"), false /*return error */},
		{testutil.TestMetric("value_2", "test"), false},
	}

	settings := pubsub.DefaultPublishSettings
	settings.CountThreshold = 1
	ps, topic, metrics := getTestResources(t, settings, testMetrics)
	topic.ContentEncoding = "gzip"
	topic.Base64Data = true
	ps.ContentEncoding = "gzip"
	ps.Base64Data = true
	var err error
	ps.encoder, err = internal.NewContentEncoder(ps.ContentEncoding)

	require.NoError(t, err)
	require.NoError(t, ps.Write(metrics))

	for _, testM := range testMetrics {
		verifyMetricPublished(t, testM.m, topic.published, true /* base64encoded */, true /* Gzipencoded */)
	}
}

func verifyRawMetricPublished(t *testing.T, m telegraf.Metric, published map[string]*pubsub.Message) *pubsub.Message {
	return verifyMetricPublished(t, m, published, false, false)
}

func verifyMetricPublished(t *testing.T, m telegraf.Metric, published map[string]*pubsub.Message, base64Encoded bool, gzipEncoded bool) *pubsub.Message {
	p := influx.Parser{}
	require.NoError(t, p.Init())

	v, _ := m.GetField("value")
	psMsg, ok := published[v.(string)]
	if !ok {
		t.Fatalf("expected metric to get published (value: %s)", v.(string))
	}

	data := psMsg.Data

	if gzipEncoded {
		decoder, _ := internal.NewContentDecoder("gzip")
		var err error
		data, err = decoder.Decode(data)
		if err != nil {
			t.Fatalf("Unable to decode expected gzip encoded message: %s", err)
		}
	}

	if base64Encoded {
		v, err := base64.StdEncoding.DecodeString(string(data))
		if err != nil {
			t.Fatalf("Unable to decode expected base64-encoded message: %s", err)
		}
		data = v
	}

	parsed, err := p.Parse(data)
	if err != nil {
		t.Fatalf("could not parse influxdb metric from published message: %s", string(data))
	}
	if len(parsed) > 1 {
		t.Fatalf("expected only one influxdb metric per published message, got %d", len(published))
	}

	publishedV, ok := parsed[0].GetField("value")
	if !ok {
		t.Fatalf("expected published metric to have a value")
	}
	require.Equal(t, v, publishedV, "incorrect published value")

	return psMsg
}
