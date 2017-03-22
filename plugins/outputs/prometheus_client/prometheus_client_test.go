package prometheus_client

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/inputs/prometheus"
	"github.com/influxdata/telegraf/testutil"
)

var pTesting *PrometheusClient

func TestPrometheusWritePointEmptyTag(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pClient, p, err := setupPrometheus()
	require.NoError(t, err)
	defer pClient.Stop()

	now := time.Now()
	tags := make(map[string]string)
	pt1, _ := metric.New(
		"test_point_1",
		tags,
		map[string]interface{}{"value": 0.0},
		now)
	pt2, _ := metric.New(
		"test_point_2",
		tags,
		map[string]interface{}{"value": 1.0},
		now)
	var metrics = []telegraf.Metric{
		pt1,
		pt2,
	}
	require.NoError(t, pClient.Write(metrics))

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
		acc.AssertContainsFields(t, e.name,
			map[string]interface{}{"value": e.value})
	}

	tags = make(map[string]string)
	tags["testtag"] = "testvalue"
	pt3, _ := metric.New(
		"test_point_3",
		tags,
		map[string]interface{}{"value": 0.0},
		now)
	pt4, _ := metric.New(
		"test_point_4",
		tags,
		map[string]interface{}{"value": 1.0},
		now)
	metrics = []telegraf.Metric{
		pt3,
		pt4,
	}
	require.NoError(t, pClient.Write(metrics))

	expected2 := []struct {
		name  string
		value float64
	}{
		{"test_point_3", 0.0},
		{"test_point_4", 1.0},
	}

	require.NoError(t, p.Gather(&acc))
	for _, e := range expected2 {
		acc.AssertContainsFields(t, e.name,
			map[string]interface{}{"value": e.value})
	}
}

func TestPrometheusExpireOldMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pClient, p, err := setupPrometheus()
	pClient.ExpirationInterval = internal.Duration{Duration: time.Second * 10}
	require.NoError(t, err)
	defer pClient.Stop()

	now := time.Now()
	tags := make(map[string]string)
	pt1, _ := metric.New(
		"test_point_1",
		tags,
		map[string]interface{}{"value": 0.0},
		now)
	var metrics = []telegraf.Metric{pt1}
	require.NoError(t, pClient.Write(metrics))

	for _, m := range pClient.metrics {
		m.Expiration = now.Add(time.Duration(-15) * time.Second)
	}

	pt2, _ := metric.New(
		"test_point_2",
		tags,
		map[string]interface{}{"value": 1.0},
		now)
	var metrics2 = []telegraf.Metric{pt2}
	require.NoError(t, pClient.Write(metrics2))

	expected := []struct {
		name  string
		value float64
		tags  map[string]string
	}{
		{"test_point_2", 1.0, tags},
	}

	var acc testutil.Accumulator

	require.NoError(t, p.Gather(&acc))
	for _, e := range expected {
		acc.AssertContainsFields(t, e.name,
			map[string]interface{}{"value": e.value})
	}

	acc.AssertDoesNotContainMeasurement(t, "test_point_1")

	// Confirm that it's not in the PrometheusClient map anymore
	assert.Equal(t, 1, len(pClient.metrics))
}

func setupPrometheus() (*PrometheusClient, *prometheus.Prometheus, error) {
	if pTesting == nil {
		pTesting = &PrometheusClient{Listen: "localhost:9127"}
		err := pTesting.Start()
		if err != nil {
			return nil, nil, err
		}
	} else {
		pTesting.metrics = make(map[string]*MetricWithExpiration)
	}

	time.Sleep(time.Millisecond * 200)

	p := &prometheus.Prometheus{
		Urls: []string{"http://localhost:9127/metrics"},
	}

	return pTesting, p, nil
}
