package cloud_pubsub

import (
	"cloud.google.com/go/pubsub"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPubSub_WriteSingle(t *testing.T) {

	testMetrics := []testMetric{
		{testutil.TestMetric("value_1", "test"), false /*return error */},
	}

	settings := pubsub.DefaultPublishSettings
	settings.CountThreshold = 1
	ps, topic, metrics := getTestResources(t, settings, testMetrics)

	err := ps.Write(metrics)
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	for _, testM := range testMetrics {
		verifyMetricPublished(t, testM.m, topic.published)
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

	err := ps.Write(metrics)
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	for _, testM := range testMetrics {
		msg := verifyMetricPublished(t, testM.m, topic.published)
		assert.Equalf(t, "bar1", msg.Attributes["foo1"], "expected attribute foo1=bar1")
		assert.Equalf(t, "bar2", msg.Attributes["foo2"], "expected attribute foo2=bar2")
	}
}

func TestPubSub_WriteMultiple(t *testing.T) {
	testMetrics := []testMetric{
		{testutil.TestMetric("value_1", "test"), false /*return error*/},
		{testutil.TestMetric("value_2", "test"), false},
	}

	settings := pubsub.DefaultPublishSettings

	ps, topic, metrics := getTestResources(t, settings, testMetrics)

	err := ps.Write(metrics)
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	for _, testM := range testMetrics {
		verifyMetricPublished(t, testM.m, topic.published)
	}
	assert.Equalf(t, 1, topic.bundleCount, "unexpected bundle count")
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

	err := ps.Write(metrics)
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	for _, testM := range testMetrics {
		verifyMetricPublished(t, testM.m, topic.published)
	}
	assert.Equalf(t, 2, topic.bundleCount, "unexpected bundle count")
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

	err := ps.Write(metrics)
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}

	for _, testM := range testMetrics {
		verifyMetricPublished(t, testM.m, topic.published)
	}
	assert.Equalf(t, 2, topic.bundleCount, "unexpected bundle count")
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
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Error() != errMockFail {
		t.Fatalf("expected fake error, got %v", err)
	}
}

func verifyMetricPublished(t *testing.T, m telegraf.Metric, published map[string]*pubsub.Message) *pubsub.Message {
	p, _ := parsers.NewInfluxParser()

	v, _ := m.GetField("value")
	psMsg, ok := published[v.(string)]
	if !ok {
		t.Fatalf("expected metric to get published (value: %s)", v.(string))
	}

	parsed, err := p.Parse(psMsg.Data)
	if err != nil {
		t.Fatalf("could not parse influxdb metric from published message: %s", string(psMsg.Data))
	}
	if len(parsed) > 1 {
		t.Fatalf("expected only one influxdb metric per published message, got %d", len(published))
	}

	publishedV, ok := parsed[0].GetField("value")
	if !ok {
		t.Fatalf("expected published metric to have a value")
	}
	assert.Equal(t, v, publishedV, "incorrect published value")

	return psMsg
}
