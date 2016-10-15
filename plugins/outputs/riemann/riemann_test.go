package riemann

import (
	"testing"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestAttributes(t *testing.T) {
	tags := map[string]string{"tag1": "value1", "tag2": "value2"}

	r := &Riemann{}
	require.Equal(t,
		map[string]string{"tag1": "value1", "tag2": "value2"},
		r.attributes("test", tags))

	r.MeasurementAsAttribute = true
	require.Equal(t,
		map[string]string{"tag1": "value1", "tag2": "value2", "measurement": "test"},
		r.attributes("test", tags))
}

func TestService(t *testing.T) {
	r := &Riemann{
		Separator: "/",
	}
	require.Equal(t, "test/value", r.service("test", "value"))

	r.MeasurementAsAttribute = true
	require.Equal(t, "value", r.service("test", "value"))
}

func TestTags(t *testing.T) {
	tags := map[string]string{"tag1": "value1", "tag2": "value2"}

	r := &Riemann{
		Tags: []string{"test"},
	}
	require.Equal(t,
		[]string{"test", "value1", "value2"},
		r.tags(tags))

	r.TagKeys = []string{"tag2"}
	require.Equal(t,
		[]string{"test", "value2"},
		r.tags(tags))

	r.Tags = nil
	r.TagKeys = []string{"tag1"}
	require.Equal(t,
		[]string{"value1"},
		r.tags(tags))
}

func TestConnectAndWrite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	addr := testutil.GetLocalHost() + ":5555"

	r := &Riemann{
		Address:                addr,
		Transport:              "tcp",
		Separator:              "/",
		MeasurementAsAttribute: false,
		DescriptionText:        "metrics from telegraf",
		Tags:                   []string{"telegraf"},
	}

	err := r.Connect()
	require.NoError(t, err)

	err = r.Write(testutil.MockMetrics())
	require.NoError(t, err)

	metrics := make([]telegraf.Metric, 0)
	metrics = append(metrics, testutil.TestMetric(2))
	metrics = append(metrics, testutil.TestMetric(3.456789))
	metrics = append(metrics, testutil.TestMetric(uint(0)))
	metrics = append(metrics, testutil.TestMetric("ok"))
	metrics = append(metrics, testutil.TestMetric("running"))
	err = r.Write(metrics)
	require.NoError(t, err)
}
