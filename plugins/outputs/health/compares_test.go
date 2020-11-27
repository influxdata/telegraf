package health_test

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs/health"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func addr(v float64) *float64 {
	return &v
}

func TestFieldNotFoundIsSuccess(t *testing.T) {
	metrics := []telegraf.Metric{
		testutil.MustMetric(
			"cpu",
			map[string]string{},
			map[string]interface{}{},
			time.Now()),
	}

	compares := &health.Compares{
		Field: "time_idle",
		GT:    addr(42.0),
	}
	result := compares.Check(metrics)
	require.True(t, result)
}

func TestStringFieldIsFailure(t *testing.T) {
	metrics := []telegraf.Metric{
		testutil.MustMetric(
			"cpu",
			map[string]string{},
			map[string]interface{}{
				"time_idle": "foo",
			},
			time.Now()),
	}

	compares := &health.Compares{
		Field: "time_idle",
		GT:    addr(42.0),
	}
	result := compares.Check(metrics)
	require.False(t, result)
}

func TestFloatConvert(t *testing.T) {
	tests := []struct {
		name     string
		metrics  []telegraf.Metric
		expected bool
	}{
		{
			name: "int64 field",
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"time_idle": int64(42.0),
					},
					time.Now()),
			},
			expected: true,
		},
		{
			name: "uint64 field",
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"time_idle": uint64(42.0),
					},
					time.Now()),
			},
			expected: true,
		},
		{
			name: "float64 field",
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"time_idle": float64(42.0),
					},
					time.Now()),
			},
			expected: true,
		},
		{
			name: "bool field true",
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"time_idle": true,
					},
					time.Now()),
			},
			expected: true,
		},
		{
			name: "bool field false",
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"time_idle": false,
					},
					time.Now()),
			},
			expected: false,
		},
		{
			name: "string field",
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"time_idle": "42.0",
					},
					time.Now()),
			},
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compares := &health.Compares{
				Field: "time_idle",
				GT:    addr(0.0),
			}
			actual := compares.Check(tt.metrics)
			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestOperators(t *testing.T) {
	tests := []struct {
		name     string
		compares *health.Compares
		expected bool
	}{
		{
			name: "gt",
			compares: &health.Compares{
				Field: "time_idle",
				GT:    addr(41.0),
			},
			expected: true,
		},
		{
			name: "not gt",
			compares: &health.Compares{
				Field: "time_idle",
				GT:    addr(42.0),
			},
			expected: false,
		},
		{
			name: "ge",
			compares: &health.Compares{
				Field: "time_idle",
				GE:    addr(42.0),
			},
			expected: true,
		},
		{
			name: "not ge",
			compares: &health.Compares{
				Field: "time_idle",
				GE:    addr(43.0),
			},
			expected: false,
		},
		{
			name: "lt",
			compares: &health.Compares{
				Field: "time_idle",
				LT:    addr(43.0),
			},
			expected: true,
		},
		{
			name: "not lt",
			compares: &health.Compares{
				Field: "time_idle",
				LT:    addr(42.0),
			},
			expected: false,
		},
		{
			name: "le",
			compares: &health.Compares{
				Field: "time_idle",
				LE:    addr(42.0),
			},
			expected: true,
		},
		{
			name: "not le",
			compares: &health.Compares{
				Field: "time_idle",
				LE:    addr(41.0),
			},
			expected: false,
		},
		{
			name: "eq",
			compares: &health.Compares{
				Field: "time_idle",
				EQ:    addr(42.0),
			},
			expected: true,
		},
		{
			name: "not eq",
			compares: &health.Compares{
				Field: "time_idle",
				EQ:    addr(41.0),
			},
			expected: false,
		},
		{
			name: "ne",
			compares: &health.Compares{
				Field: "time_idle",
				NE:    addr(41.0),
			},
			expected: true,
		},
		{
			name: "not ne",
			compares: &health.Compares{
				Field: "time_idle",
				NE:    addr(42.0),
			},
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"time_idle": 42.0,
					},
					time.Now()),
			}
			actual := tt.compares.Check(metrics)
			require.Equal(t, tt.expected, actual)
		})
	}
}
