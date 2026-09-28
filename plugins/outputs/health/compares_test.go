package health_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/outputs/health"
)

func TestFieldNotFoundIsSuccess(t *testing.T) {
	metrics := []telegraf.Metric{
		metric.New(
			"cpu",
			map[string]string{},
			map[string]any{},
			time.Now()),
	}

	compares := &health.Compares{
		Field: "time_idle",
		GT:    new(42.0),
	}
	result := compares.Check(metrics)
	require.True(t, result)
}

func TestStringFieldIsFailure(t *testing.T) {
	metrics := []telegraf.Metric{
		metric.New(
			"cpu",
			map[string]string{},
			map[string]any{
				"time_idle": "foo",
			},
			time.Now()),
	}

	compares := &health.Compares{
		Field: "time_idle",
		GT:    new(42.0),
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
				metric.New(
					"cpu",
					map[string]string{},
					map[string]any{
						"time_idle": int64(42.0),
					},
					time.Now()),
			},
			expected: true,
		},
		{
			name: "uint64 field",
			metrics: []telegraf.Metric{
				metric.New(
					"cpu",
					map[string]string{},
					map[string]any{
						"time_idle": uint64(42.0),
					},
					time.Now()),
			},
			expected: true,
		},
		{
			name: "float64 field",
			metrics: []telegraf.Metric{
				metric.New(
					"cpu",
					map[string]string{},
					map[string]any{
						"time_idle": float64(42.0),
					},
					time.Now()),
			},
			expected: true,
		},
		{
			name: "bool field true",
			metrics: []telegraf.Metric{
				metric.New(
					"cpu",
					map[string]string{},
					map[string]any{
						"time_idle": true,
					},
					time.Now()),
			},
			expected: true,
		},
		{
			name: "bool field false",
			metrics: []telegraf.Metric{
				metric.New(
					"cpu",
					map[string]string{},
					map[string]any{
						"time_idle": false,
					},
					time.Now()),
			},
			expected: false,
		},
		{
			name: "string field",
			metrics: []telegraf.Metric{
				metric.New(
					"cpu",
					map[string]string{},
					map[string]any{
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
				GT:    new(0.0),
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
				GT:    new(41.0),
			},
			expected: true,
		},
		{
			name: "not gt",
			compares: &health.Compares{
				Field: "time_idle",
				GT:    new(42.0),
			},
			expected: false,
		},
		{
			name: "ge",
			compares: &health.Compares{
				Field: "time_idle",
				GE:    new(42.0),
			},
			expected: true,
		},
		{
			name: "not ge",
			compares: &health.Compares{
				Field: "time_idle",
				GE:    new(43.0),
			},
			expected: false,
		},
		{
			name: "lt",
			compares: &health.Compares{
				Field: "time_idle",
				LT:    new(43.0),
			},
			expected: true,
		},
		{
			name: "not lt",
			compares: &health.Compares{
				Field: "time_idle",
				LT:    new(42.0),
			},
			expected: false,
		},
		{
			name: "le",
			compares: &health.Compares{
				Field: "time_idle",
				LE:    new(42.0),
			},
			expected: true,
		},
		{
			name: "not le",
			compares: &health.Compares{
				Field: "time_idle",
				LE:    new(41.0),
			},
			expected: false,
		},
		{
			name: "eq",
			compares: &health.Compares{
				Field: "time_idle",
				EQ:    new(42.0),
			},
			expected: true,
		},
		{
			name: "not eq",
			compares: &health.Compares{
				Field: "time_idle",
				EQ:    new(41.0),
			},
			expected: false,
		},
		{
			name: "ne",
			compares: &health.Compares{
				Field: "time_idle",
				NE:    new(41.0),
			},
			expected: true,
		},
		{
			name: "not ne",
			compares: &health.Compares{
				Field: "time_idle",
				NE:    new(42.0),
			},
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := []telegraf.Metric{
				metric.New(
					"cpu",
					map[string]string{},
					map[string]any{
						"time_idle": 42.0,
					},
					time.Now()),
			}
			actual := tt.compares.Check(metrics)
			require.Equal(t, tt.expected, actual)
		})
	}
}
