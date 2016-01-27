package prometheus_client

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs/prometheus"
	"github.com/influxdata/telegraf/testutil"
)

var pTesting *PrometheusClient

func TestPrometheusWritePointEmptyTag(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p := &prometheus.Prometheus{
		Urls: []string{"http://localhost:9126/metrics"},
	}
	tags := make(map[string]string)
	pt1, _ := telegraf.NewMetric(
		"test_point_1",
		tags,
		map[string]interface{}{"value": 0.0})
	pt2, _ := telegraf.NewMetric(
		"test_point_2",
		tags,
		map[string]interface{}{"value": 1.0})
	var metrics = []telegraf.Metric{
		pt1,
		pt2,
	}
	require.NoError(t, pTesting.Write(metrics))

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
		acc.AssertContainsFields(t, "prometheus_"+e.name,
			map[string]interface{}{"value": e.value})
	}
}

func TestPrometheusWritePointTag(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p := &prometheus.Prometheus{
		Urls: []string{"http://localhost:9126/metrics"},
	}
	tags := make(map[string]string)
	tags["testtag"] = "testvalue"
	pt1, _ := telegraf.NewMetric(
		"test_point_3",
		tags,
		map[string]interface{}{"value": 0.0})
	pt2, _ := telegraf.NewMetric(
		"test_point_4",
		tags,
		map[string]interface{}{"value": 1.0})
	var metrics = []telegraf.Metric{
		pt1,
		pt2,
	}
	require.NoError(t, pTesting.Write(metrics))

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
		acc.AssertContainsFields(t, "prometheus_"+e.name,
			map[string]interface{}{"value": e.value})
	}
}

func init() {
	pTesting = &PrometheusClient{Listen: "localhost:9126"}
	pTesting.Start()
}
