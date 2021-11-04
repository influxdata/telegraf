package noise

import (
	"reflect"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

// Helper function which returns a map for metrics for easy lookup
func createTestMetrics() map[string]telegraf.Metric {
	m := make(map[string]telegraf.Metric)
	m["cpu"] = testutil.MustMetric(
		"cpu",
		map[string]string{},
		map[string]interface{}{
			"usage_guest":  2.5,
			"usage_system": 1.5,
			"usage_nice":   0.5,
			"usage_irq":    1.0,
		},
		time.Unix(0, 0),
	)
	m["disk"] = testutil.MustMetric(
		"disk",
		map[string]string{},
		map[string]interface{}{
			"free":        250,
			"inodes_free": 1500.0,
			"inodes_used": 1337,
		},
		time.Unix(0, 0),
	)
	return m
}

// Verifies that field values are modified by the Laplace noise
func TestAddNoiseToMetric(t *testing.T) {
	processor := Noise{
		Sensitivity: 1.0,
		Epsilon:     1.0,
		Log:         testutil.Logger{},
	}
	metrics := createTestMetrics()
	for _, metric := range metrics {
		after := processor.Apply(metric.Copy())[0]
		require.NotEqual(t, metric, after)
	}
}

// Tests that fields in the IgnoreField set are not affected by the Laplace
// noise
func TestLaplaceNoiseWithIgnoreField(t *testing.T) {
	processor := Noise{
		Sensitivity:  1.0,
		Epsilon:      1.0,
		IgnoreFields: []string{"usage_guest", "usage_system"},
		Log:          testutil.Logger{},
	}

	// call Init as we want to initialize the excludeLists
	_ = processor.Init()
	metric := createTestMetrics()
	after := processor.Apply(metric["cpu"].Copy())[0]

	// check that some values in the struct have been changed
	require.NotEqual(t, metric, after)

	// check that ignore values were not changed
	for _, ignore := range processor.IgnoreFields {
		have, _ := metric["cpu"].GetField(ignore)
		should, _ := after.GetField(ignore)
		require.Equal(t, have, should)
	}
}

// Tests that Measurements in the IgnoreMeasurement set are not affected by the
// Laplace noise
func TestLaplaceNoiseWithIgnoreMeasurement(t *testing.T) {
	processor := Noise{
		Sensitivity:        1.0,
		Epsilon:            1.0,
		IgnoreMeasurements: []string{"disk"},
		Log:                testutil.Logger{},
	}

	// call Init as we want to initialize the excludeLists
	_ = processor.Init()
	metrics := createTestMetrics()
	after := metrics

	// Run Apply for all metrics
	for _, metric := range metrics {
		after[metric.Name()] = processor.Apply(metric.Copy())[0]
	}

	// iterate metrics which should have been ignored if they still have the
	// value
	for _, ignoreMeasurement := range processor.IgnoreMeasurements {
		processor.Log.Infof("ignoreMeasurement", ignoreMeasurement)
		for key, value := range metrics[ignoreMeasurement].Fields() {
			require.Equal(t, value, after[ignoreMeasurement].Fields()[key])
		}
	}
}

func TestAddNoiseToValue(t *testing.T) {
	processor := Noise{
		Sensitivity: 1.0,
		Epsilon:     1.0,
		Log:         testutil.Logger{},
	}

	haveValues := []interface{}{
		int64(-5),
		uint64(4),
		float64(1.337),
	}

	haveValuesInvalid := []interface{}{
		string("helloworld"),
	}

	for _, value := range haveValues {
		after := processor.addNoiseToValue(value)
		// check value is not the same
		require.NotEqual(t, value, after)
		// check type is still the same
		require.Equal(t, reflect.TypeOf(value), reflect.TypeOf(after))
	}

	// check that nothing happens to non numerical types:
	for _, value := range haveValuesInvalid {
		after := processor.addNoiseToValue(value)
		require.Equal(t, value, after)
		require.Equal(t, reflect.TypeOf(value), reflect.TypeOf(after))
	}
}
