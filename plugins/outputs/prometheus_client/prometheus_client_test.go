package prometheus_client

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
	prometheus_input "github.com/influxdata/telegraf/plugins/inputs/prometheus"
	"github.com/influxdata/telegraf/testutil"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func setUnixTime(client *PrometheusClient, sec int64) {
	client.now = func() time.Time {
		return time.Unix(sec, 0)
	}
}

// NewClient initializes a PrometheusClient.
func NewClient() *PrometheusClient {
	return &PrometheusClient{
		ExpirationInterval: internal.Duration{Duration: time.Second * 60},
		fam:                make(map[string]*MetricFamily),
		now:                time.Now,
	}
}

func TestWrite_Basic(t *testing.T) {
	now := time.Now()
	pt1, err := metric.New(
		"foo",
		make(map[string]string),
		map[string]interface{}{"value": 0.0},
		now)
	var metrics = []telegraf.Metric{
		pt1,
	}

	client := NewClient()
	err = client.Write(metrics)
	require.NoError(t, err)

	fam, ok := client.fam["foo"]
	require.True(t, ok)
	require.Equal(t, prometheus.UntypedValue, fam.ValueType)
	require.Equal(t, map[string]int{}, fam.LabelSet)

	sample, ok := fam.Samples[CreateSampleID(pt1.Tags())]
	require.True(t, ok)

	require.Equal(t, 0.0, sample.Value)
	require.True(t, now.Before(sample.Expiration))
}

func TestWrite_IntField(t *testing.T) {
	client := NewClient()

	p1, err := metric.New(
		"foo",
		make(map[string]string),
		map[string]interface{}{"value": 42},
		time.Now())
	err = client.Write([]telegraf.Metric{p1})
	require.NoError(t, err)

	fam, ok := client.fam["foo"]
	require.True(t, ok)
	for _, v := range fam.Samples {
		require.Equal(t, 42.0, v.Value)
	}

}

func TestWrite_FieldNotValue(t *testing.T) {
	client := NewClient()

	p1, err := metric.New(
		"foo",
		make(map[string]string),
		map[string]interface{}{"howdy": 0.0},
		time.Now())
	err = client.Write([]telegraf.Metric{p1})
	require.NoError(t, err)

	fam, ok := client.fam["foo_howdy"]
	require.True(t, ok)
	for _, v := range fam.Samples {
		require.Equal(t, 0.0, v.Value)
	}
}

func TestWrite_SkipNonNumberField(t *testing.T) {
	client := NewClient()

	p1, err := metric.New(
		"foo",
		make(map[string]string),
		map[string]interface{}{"value": "howdy"},
		time.Now())
	err = client.Write([]telegraf.Metric{p1})
	require.NoError(t, err)

	_, ok := client.fam["foo"]
	require.False(t, ok)
}

func TestWrite_Counter(t *testing.T) {
	client := NewClient()

	p1, err := metric.New(
		"foo",
		make(map[string]string),
		map[string]interface{}{"value": 42},
		time.Now(),
		telegraf.Counter)
	err = client.Write([]telegraf.Metric{p1})
	require.NoError(t, err)

	fam, ok := client.fam["foo"]
	require.True(t, ok)
	require.Equal(t, prometheus.CounterValue, fam.ValueType)
}

func TestWrite_Sanitize(t *testing.T) {
	client := NewClient()

	p1, err := metric.New(
		"foo.bar",
		map[string]string{"tag-with-dash": "localhost.local"},
		map[string]interface{}{"field-with-dash": 42},
		time.Now(),
		telegraf.Counter)
	err = client.Write([]telegraf.Metric{p1})
	require.NoError(t, err)

	fam, ok := client.fam["foo_bar_field_with_dash"]
	require.True(t, ok)
	require.Equal(t, map[string]int{"tag_with_dash": 1}, fam.LabelSet)

	sample1, ok := fam.Samples[CreateSampleID(p1.Tags())]
	require.True(t, ok)

	require.Equal(t, map[string]string{
		"tag_with_dash": "localhost.local"}, sample1.Labels)
}

func TestWrite_Gauge(t *testing.T) {
	client := NewClient()

	p1, err := metric.New(
		"foo",
		make(map[string]string),
		map[string]interface{}{"value": 42},
		time.Now(),
		telegraf.Gauge)
	err = client.Write([]telegraf.Metric{p1})
	require.NoError(t, err)

	fam, ok := client.fam["foo"]
	require.True(t, ok)
	require.Equal(t, prometheus.GaugeValue, fam.ValueType)
}

func TestWrite_MixedValueType(t *testing.T) {
	now := time.Now()
	p1, err := metric.New(
		"foo",
		make(map[string]string),
		map[string]interface{}{"value": 1.0},
		now,
		telegraf.Counter)
	p2, err := metric.New(
		"foo",
		make(map[string]string),
		map[string]interface{}{"value": 2.0},
		now,
		telegraf.Gauge)
	var metrics = []telegraf.Metric{p1, p2}

	client := NewClient()
	err = client.Write(metrics)
	require.NoError(t, err)

	fam, ok := client.fam["foo"]
	require.True(t, ok)
	require.Equal(t, 1, len(fam.Samples))
}

func TestWrite_MixedValueTypeUpgrade(t *testing.T) {
	now := time.Now()
	p1, err := metric.New(
		"foo",
		map[string]string{"a": "x"},
		map[string]interface{}{"value": 1.0},
		now,
		telegraf.Untyped)
	p2, err := metric.New(
		"foo",
		map[string]string{"a": "y"},
		map[string]interface{}{"value": 2.0},
		now,
		telegraf.Gauge)
	var metrics = []telegraf.Metric{p1, p2}

	client := NewClient()
	err = client.Write(metrics)
	require.NoError(t, err)

	fam, ok := client.fam["foo"]
	require.True(t, ok)
	require.Equal(t, 2, len(fam.Samples))
}

func TestWrite_MixedValueTypeDowngrade(t *testing.T) {
	now := time.Now()
	p1, err := metric.New(
		"foo",
		map[string]string{"a": "x"},
		map[string]interface{}{"value": 1.0},
		now,
		telegraf.Gauge)
	p2, err := metric.New(
		"foo",
		map[string]string{"a": "y"},
		map[string]interface{}{"value": 2.0},
		now,
		telegraf.Untyped)
	var metrics = []telegraf.Metric{p1, p2}

	client := NewClient()
	err = client.Write(metrics)
	require.NoError(t, err)

	fam, ok := client.fam["foo"]
	require.True(t, ok)
	require.Equal(t, 2, len(fam.Samples))
}

func TestWrite_Tags(t *testing.T) {
	now := time.Now()
	p1, err := metric.New(
		"foo",
		make(map[string]string),
		map[string]interface{}{"value": 1.0},
		now)
	p2, err := metric.New(
		"foo",
		map[string]string{"host": "localhost"},
		map[string]interface{}{"value": 2.0},
		now)
	var metrics = []telegraf.Metric{p1, p2}

	client := NewClient()
	err = client.Write(metrics)
	require.NoError(t, err)

	fam, ok := client.fam["foo"]
	require.True(t, ok)
	require.Equal(t, prometheus.UntypedValue, fam.ValueType)

	require.Equal(t, map[string]int{"host": 1}, fam.LabelSet)

	sample1, ok := fam.Samples[CreateSampleID(p1.Tags())]
	require.True(t, ok)

	require.Equal(t, 1.0, sample1.Value)
	require.True(t, now.Before(sample1.Expiration))

	sample2, ok := fam.Samples[CreateSampleID(p2.Tags())]
	require.True(t, ok)

	require.Equal(t, 2.0, sample2.Value)
	require.True(t, now.Before(sample2.Expiration))
}

func TestExpire(t *testing.T) {
	client := NewClient()

	p1, err := metric.New(
		"foo",
		make(map[string]string),
		map[string]interface{}{"value": 1.0},
		time.Now())
	setUnixTime(client, 0)
	err = client.Write([]telegraf.Metric{p1})
	require.NoError(t, err)

	p2, err := metric.New(
		"bar",
		make(map[string]string),
		map[string]interface{}{"value": 2.0},
		time.Now())
	setUnixTime(client, 1)
	err = client.Write([]telegraf.Metric{p2})

	setUnixTime(client, 61)
	require.Equal(t, 2, len(client.fam))
	client.Expire()
	require.Equal(t, 1, len(client.fam))
}

func TestExpire_TagsNoDecrement(t *testing.T) {
	client := NewClient()

	p1, err := metric.New(
		"foo",
		make(map[string]string),
		map[string]interface{}{"value": 1.0},
		time.Now())
	setUnixTime(client, 0)
	err = client.Write([]telegraf.Metric{p1})
	require.NoError(t, err)

	p2, err := metric.New(
		"foo",
		map[string]string{"host": "localhost"},
		map[string]interface{}{"value": 2.0},
		time.Now())
	setUnixTime(client, 1)
	err = client.Write([]telegraf.Metric{p2})

	setUnixTime(client, 61)
	fam, ok := client.fam["foo"]
	require.True(t, ok)
	require.Equal(t, 2, len(fam.Samples))
	client.Expire()
	require.Equal(t, 1, len(fam.Samples))

	require.Equal(t, map[string]int{"host": 1}, fam.LabelSet)
}

func TestExpire_TagsWithDecrement(t *testing.T) {
	client := NewClient()

	p1, err := metric.New(
		"foo",
		map[string]string{"host": "localhost"},
		map[string]interface{}{"value": 1.0},
		time.Now())
	setUnixTime(client, 0)
	err = client.Write([]telegraf.Metric{p1})
	require.NoError(t, err)

	p2, err := metric.New(
		"foo",
		make(map[string]string),
		map[string]interface{}{"value": 2.0},
		time.Now())
	setUnixTime(client, 1)
	err = client.Write([]telegraf.Metric{p2})

	setUnixTime(client, 61)
	fam, ok := client.fam["foo"]
	require.True(t, ok)
	require.Equal(t, 2, len(fam.Samples))
	client.Expire()
	require.Equal(t, 1, len(fam.Samples))

	require.Equal(t, map[string]int{"host": 0}, fam.LabelSet)
}

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

func setupPrometheus() (*PrometheusClient, *prometheus_input.Prometheus, error) {
	if pTesting == nil {
		pTesting = NewClient()
		pTesting.Listen = "localhost:9127"
		pTesting.Path = "/metrics"
		err := pTesting.Start()
		if err != nil {
			return nil, nil, err
		}
	} else {
		pTesting.fam = make(map[string]*MetricFamily)
	}

	time.Sleep(time.Millisecond * 200)

	p := &prometheus_input.Prometheus{
		Urls: []string{"http://localhost:9127/metrics"},
	}

	return pTesting, p, nil
}
