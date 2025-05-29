package noise

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
	"time"
)

func TestRoundSignificantFigures(t *testing.T) {
	tests := []struct {
		name     string
		sf       int
		input    float64
		expected float64
	}{
		{
			name:     "3sf with decimal part",
			sf:       3,
			input:    12.3456789,
			expected: 12.3,
		},
		{
			name:     "6sf with decimal part",
			sf:       6,
			input:    12.3456789,
			expected: 12.3457,
		},
		{
			name:     "3sf with decimal removes",
			sf:       3,
			input:    103.6,
			expected: 104,
		},
		{
			name:     "3sf without decimal",
			sf:       3,
			input:    1068,
			expected: 1070,
		},
		{
			name:     "3sf with leading zeros",
			sf:       3,
			input:    0.005,
			expected: 0.005,
		},
		{
			name:     "3sf with padding zeros",
			sf:       3,
			input:    1.006,
			expected: 1.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := roundToSignificantFigures(tt.input, tt.sf)
			if actual != tt.expected {
				t.Errorf("roundToSignificantFigures() actual = %v, expected = %v", actual, tt.expected)
			}
		})
	}
}

// Verifies that noise is correctly removed from values
func TestDenoise(t *testing.T) {
	tests := []struct {
		name     string
		input    []telegraf.Metric
		expected []telegraf.Metric
	}{
		{
			name: "int64",
			input: []telegraf.Metric{
				testutil.MustMetric("cpu",
					map[string]string{},
					map[string]interface{}{"value": int64(5567)},
					time.Unix(0, 0),
				),
				testutil.MustMetric("cpu",
					map[string]string{},
					map[string]interface{}{"value": float64(-1043)},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric("cpu",
					map[string]string{},
					map[string]interface{}{"value": float64(5570)},
					time.Unix(0, 0),
				),
				testutil.MustMetric("cpu",
					map[string]string{},
					map[string]interface{}{"value": float64(-1040)},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "uint64",
			input: []telegraf.Metric{
				testutil.MustMetric("cpu",
					map[string]string{},
					map[string]interface{}{"value": uint64(2505)},
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
					map[string]interface{}{"value": float64(2510)},
					time.Unix(0, 0),
				),
				testutil.MustMetric("cpu",
					map[string]string{},
					map[string]interface{}{"value": uint64(0)},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "float64",
			input: []telegraf.Metric{
				testutil.MustMetric("cpu",
					map[string]string{},
					map[string]interface{}{"value": float64(1.0798567)},
					time.Unix(0, 0),
				),
				testutil.MustMetric("cpu",
					map[string]string{},
					map[string]interface{}{"value": float64(10570.34507)},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric("cpu",
					map[string]string{},
					map[string]interface{}{"value": float64(1.08)},
					time.Unix(0, 0),
				),
				testutil.MustMetric("cpu",
					map[string]string{},
					map[string]interface{}{"value": float64(10600)},
					time.Unix(0, 0),
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := Denoise{
				SignificantFigures: 3,
				Log:                testutil.Logger{},
			}
			require.NoError(t, plugin.Init())

			actual := plugin.Apply(tt.input...)
			testutil.RequireMetricsEqual(t, tt.expected, actual)
		})
	}
}

// Verifies that denoise() returns zero values as 0
func TestDenoiseWithZeroValue(t *testing.T) {
	tests := []struct {
		name     string
		input    []telegraf.Metric
		expected []telegraf.Metric
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
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := Denoise{
				SignificantFigures: 3,
				Log:                testutil.Logger{},
			}
			require.NoError(t, plugin.Init())

			actual := plugin.Apply(tt.input...)
			testutil.RequireMetricsEqual(t, tt.expected, actual)
		})
	}
}

// Verifies that any invalid significant figures value raises an error
func TestInvalidSignificantFigures(t *testing.T) {
	p := Denoise{
		SignificantFigures: 0,
		Log:                testutil.Logger{},
	}
	err := p.Init()
	require.EqualError(t, err, "significant figures must be at least 1, got 0")
}

func TestTracking(t *testing.T) {
	// Setup raw input and expected output
	inputRaw := []telegraf.Metric{
		metric.New(
			"uint64",
			map[string]string{},
			map[string]interface{}{"value": uint64(1236)},
			time.Unix(0, 0),
		),
		metric.New(
			"int64",
			map[string]string{},
			map[string]interface{}{"value": int64(-234)},
			time.Unix(0, 0),
		),
		metric.New(
			"float",
			map[string]string{},
			map[string]interface{}{"value": float64(-45.7894)},
			time.Unix(0, 0),
		),
	}

	expected := []telegraf.Metric{
		metric.New(
			"uint64",
			map[string]string{},
			map[string]interface{}{"value": float64(1240)},
			time.Unix(0, 0),
		),
		metric.New(
			"int64",
			map[string]string{},
			map[string]interface{}{"value": float64(-234)},
			time.Unix(0, 0),
		),
		metric.New(
			"float",
			map[string]string{},
			map[string]interface{}{"value": float64(-45.8)},
			time.Unix(0, 0),
		),
	}

	// Create fake notification for testing
	var mu sync.Mutex
	delivered := make([]telegraf.DeliveryInfo, 0, len(inputRaw))
	notify := func(di telegraf.DeliveryInfo) {
		mu.Lock()
		defer mu.Unlock()
		delivered = append(delivered, di)
	}

	// Convert raw input to tracking metric
	input := make([]telegraf.Metric, 0, len(inputRaw))
	for _, m := range inputRaw {
		tm, _ := metric.WithTracking(m, notify)
		input = append(input, tm)
	}

	// Prepare and start the plugin
	plugin := &Denoise{
		SignificantFigures: 3,
		Log:                testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	// Process expected metrics and compare with resulting metrics
	actual := plugin.Apply(input...)
	testutil.RequireMetricsEqual(t, expected, actual)

	// Simulate output acknowledging delivery
	for _, m := range actual {
		m.Accept()
	}

	// Check delivery
	require.Eventuallyf(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(input) == len(delivered)
	}, time.Second, 100*time.Millisecond, "%d delivered but %d expected", len(delivered), len(expected))
}
