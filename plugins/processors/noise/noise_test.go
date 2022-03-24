package noise

import (
	"math"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"gonum.org/v1/gonum/stat/distuv"
)

type testDistribution struct {
	value float64
}

func (t *testDistribution) Rand() float64 {
	return t.value
}

// Verifies that field values are modified by the Laplace noise
func TestAddNoiseToMetric(t *testing.T) {
	generators := []string{"laplacian", "gaussian", "uniform"}
	for _, generator := range generators {
		p := Noise{
			NoiseType: generator,
			Scale:     1.0,
			Mu:        0.0,
			Min:       -1,
			Max:       1,
			Log:       testutil.Logger{},
		}
		require.NoError(t, p.Init())
		for _, m := range testutil.MockMetrics() {
			after := p.Apply(m.Copy())
			require.Len(t, after, 1)
			require.NotEqual(t, m, after[0])
		}
	}
}

// Verifies that a given noise is added correctly to values
func TestAddNoise(t *testing.T) {
	tests := []struct {
		name         string
		input        []telegraf.Metric
		expected     []telegraf.Metric
		distribution distuv.Rander
	}{
		{
			name: "int64",
			input: []telegraf.Metric{
				testutil.MustMetric("cpu",
					map[string]string{},
					map[string]interface{}{"value": int64(5)},
					time.Unix(0, 0),
				),
				testutil.MustMetric("cpu",
					map[string]string{},
					map[string]interface{}{"value": int64(-10)},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric("cpu",
					map[string]string{},
					map[string]interface{}{"value": int64(4)},
					time.Unix(0, 0),
				),
				testutil.MustMetric("cpu",
					map[string]string{},
					map[string]interface{}{"value": int64(-11)},
					time.Unix(0, 0),
				),
			},
			distribution: &testDistribution{value: -1.5},
		},
		{
			name: "uint64",
			input: []telegraf.Metric{
				testutil.MustMetric("cpu",
					map[string]string{},
					map[string]interface{}{"value": uint64(25)},
					time.Unix(0, 0),
				),
				testutil.MustMetric("cpu",
					map[string]string{},
					map[string]interface{}{"value": uint64(0)},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric("cpu",
					map[string]string{},
					map[string]interface{}{"value": uint64(26)},
					time.Unix(0, 0),
				),
				testutil.MustMetric("cpu",
					map[string]string{},
					map[string]interface{}{"value": uint64(1)},
					time.Unix(0, 0),
				),
			},
			distribution: &testDistribution{value: 1.5},
		},
		{
			name: "float64",
			input: []telegraf.Metric{
				testutil.MustMetric("cpu",
					map[string]string{},
					map[string]interface{}{"value": float64(0.0005)},
					time.Unix(0, 0),
				),
				testutil.MustMetric("cpu",
					map[string]string{},
					map[string]interface{}{"value": float64(1000.5)},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric("cpu",
					map[string]string{},
					map[string]interface{}{"value": float64(5.0005)},
					time.Unix(0, 0),
				),
				testutil.MustMetric("cpu",
					map[string]string{},
					map[string]interface{}{"value": float64(1005.5)},
					time.Unix(0, 0),
				),
			},
			distribution: &testDistribution{value: 5.0},
		},
		{
			name: "float64",
			input: []telegraf.Metric{
				testutil.MustMetric("cpu",
					map[string]string{},
					map[string]interface{}{"value": float64(0.0005)},
					time.Unix(0, 0),
				),
				testutil.MustMetric("cpu",
					map[string]string{},
					map[string]interface{}{"value": float64(1000.5)},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric("cpu",
					map[string]string{},
					map[string]interface{}{"value": float64(-0.4995)},
					time.Unix(0, 0),
				),
				testutil.MustMetric("cpu",
					map[string]string{},
					map[string]interface{}{"value": float64(1000)},
					time.Unix(0, 0),
				),
			},
			distribution: &testDistribution{value: -0.5},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := Noise{
				NoiseType: "laplacian",
				Scale:     1.0,
				Log:       testutil.Logger{},
			}
			require.NoError(t, plugin.Init())
			plugin.generator = tt.distribution

			actual := plugin.Apply(tt.input...)
			testutil.RequireMetricsEqual(t, tt.expected, actual)
		})
	}
}

// Tests that int64 & uint64 overflow errors are catched
func TestAddNoiseOverflowCheck(t *testing.T) {
	tests := []struct {
		name         string
		input        []telegraf.Metric
		expected     []telegraf.Metric
		distribution distuv.Rander
	}{
		{
			name: "underflow",
			input: []telegraf.Metric{
				testutil.MustMetric("underflow_int64",
					map[string]string{},
					map[string]interface{}{"value": int64(math.MinInt64)},
					time.Unix(0, 0),
				),
				testutil.MustMetric("underflow_uint64_1",
					map[string]string{},
					map[string]interface{}{"value": uint64(5)},
					time.Unix(0, 0),
				),
				testutil.MustMetric("underflow_uint64_2",
					map[string]string{},
					map[string]interface{}{"value": uint64(0)},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric("underflow_int64",
					map[string]string{},
					map[string]interface{}{"value": int64(math.MinInt64)},
					time.Unix(0, 0),
				),
				testutil.MustMetric("underflow_uint64_1",
					map[string]string{},
					map[string]interface{}{"value": uint64(4)},
					time.Unix(0, 0),
				),
				testutil.MustMetric("underflow_uint64_2",
					map[string]string{},
					map[string]interface{}{"value": uint64(0)},
					time.Unix(0, 0),
				),
			},
			distribution: &testDistribution{value: -1.0},
		},
		{
			name: "overflow",
			input: []telegraf.Metric{
				testutil.MustMetric("overflow_int64",
					map[string]string{},
					map[string]interface{}{"value": int64(math.MaxInt64)},
					time.Unix(0, 0),
				),
				testutil.MustMetric("overflow_uint",
					map[string]string{},
					map[string]interface{}{"value": uint64(math.MaxUint)},
					time.Unix(0, 0),
				),
				testutil.MustMetric("overflow_uint64",
					map[string]string{},
					map[string]interface{}{"value": uint64(math.MaxUint64)},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric("overflow_int64",
					map[string]string{},
					map[string]interface{}{"value": int64(math.MaxInt64)},
					time.Unix(0, 0),
				),
				testutil.MustMetric("overflow_uint",
					map[string]string{},
					map[string]interface{}{"value": uint64(math.MaxUint)},
					time.Unix(0, 0),
				),
				testutil.MustMetric("overflow_uint64",
					map[string]string{},
					map[string]interface{}{"value": uint64(math.MaxUint64)},
					time.Unix(0, 0),
				),
			},
			distribution: &testDistribution{value: 0.0},
		},
		{
			name: "non-numeric fields",
			input: []telegraf.Metric{
				testutil.MustMetric("cpu",
					map[string]string{},
					map[string]interface{}{
						"a": "test",
						"b": true,
					},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric("cpu",
					map[string]string{},
					map[string]interface{}{
						"a": "test",
						"b": true,
					},
					time.Unix(0, 0),
				),
			},
			distribution: &testDistribution{value: 1.0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := Noise{
				NoiseType: "laplacian",
				Scale:     1.0,
				Log:       testutil.Logger{},
			}
			require.NoError(t, plugin.Init())
			plugin.generator = tt.distribution

			actual := plugin.Apply(tt.input...)
			testutil.RequireMetricsEqual(t, tt.expected, actual)
		})
	}
}

// Verifies that even addNoise() modifies 0 values as well
func TestAddNoiseWithZeroValue(t *testing.T) {
	tests := []struct {
		name         string
		input        []telegraf.Metric
		expected     []telegraf.Metric
		distribution distuv.Rander
	}{
		{
			name: "zeros",
			input: []telegraf.Metric{
				testutil.MustMetric("zero_uint64",
					map[string]string{},
					map[string]interface{}{"value": uint64(0)},
					time.Unix(0, 0),
				),
				testutil.MustMetric("zero_int64",
					map[string]string{},
					map[string]interface{}{"value": int64(0)},
					time.Unix(0, 0),
				),
				testutil.MustMetric("zero_float",
					map[string]string{},
					map[string]interface{}{"value": float64(0.0)},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric("zero_uint64",
					map[string]string{},
					map[string]interface{}{"value": uint64(13)},
					time.Unix(0, 0),
				),
				testutil.MustMetric("zero_int64",
					map[string]string{},
					map[string]interface{}{"value": int64(13)},
					time.Unix(0, 0),
				),
				testutil.MustMetric("zero_float",
					map[string]string{},
					map[string]interface{}{"value": float64(13.37)},
					time.Unix(0, 0),
				),
			},
			distribution: &testDistribution{value: 13.37},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := Noise{
				NoiseType: "laplacian",
				Scale:     1.0,
				Log:       testutil.Logger{},
			}
			require.NoError(t, plugin.Init())
			plugin.generator = tt.distribution

			actual := plugin.Apply(tt.input...)
			testutil.RequireMetricsEqual(t, tt.expected, actual)
		})
	}
}

// Verifies that any invalid generator setting (not "laplacian", "gaussian" or
// "uniform") raises an error
func TestInvalidDistributionFunction(t *testing.T) {
	p := Noise{
		NoiseType: "invalid",
		Log:       testutil.Logger{},
	}
	err := p.Init()
	require.EqualError(t, err, "unknown distribution type \"invalid\"")
}
