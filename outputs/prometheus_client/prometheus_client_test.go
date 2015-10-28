package prometheus_client

import (
	"github.com/influxdb/influxdb/client/v2"
	"github.com/influxdb/telegraf/plugins/prometheus"
	"github.com/influxdb/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

var pTesting *PrometheusClient = &PrometheusClient{}

func TestPrometheusStart(t *testing.T) {
	require.NoError(t, pTesting.Start())
}

func TestPrometheusWritePointEmptyTag(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	p := &prometheus.Prometheus{
		Urls: []string{"http://" + testutil.GetLocalHost() + ":9126/metrics"},
	}
	tags := make(map[string]string)
	var points = []*client.Point{
		client.NewPoint(
			"test_point_1",
			tags,
			map[string]interface{}{"value": 0.0}),
		client.NewPoint(
			"test_point_2",
			tags,
			map[string]interface{}{"value": 1.0}),
	}
	require.NoError(t, pTesting.Write(points))

	expected := []struct {
		name  string
		value float64
		tags  map[string]string
	}{
		{"test_point_1", 0.0, tags},
		{"test_point_2", 1.0, tags},
	}

	var acc testutil.Accumulator

	require.NoError(t, p.Gather(&acc))
	for _, e := range expected {
		assert.NoError(t, acc.ValidateValue(e.name, e.value))
	}
}

func TestPrometheusWritePointTag(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	p := &prometheus.Prometheus{
		Urls: []string{"http://" + testutil.GetLocalHost() + ":9126/metrics"},
	}
	tags := make(map[string]string)
	tags["testtag"] = "testvalue"
	var points = []*client.Point{
		client.NewPoint(
			"test_point_3",
			tags,
			map[string]interface{}{"value": 0.0}),
		client.NewPoint(
			"test_point_4",
			tags,
			map[string]interface{}{"value": 1.0}),
	}
	require.NoError(t, pTesting.Write(points))

	expected := []struct {
		name  string
		value float64
	}{
		{"test_point_3", 0.0},
		{"test_point_4", 1.0},
	}

	var acc testutil.Accumulator

	require.NoError(t, p.Gather(&acc))
	for _, e := range expected {
		assert.True(t, acc.CheckTaggedValue(e.name, e.value, tags))
	}
}
