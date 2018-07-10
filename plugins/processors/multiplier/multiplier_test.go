package multiplier

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/require"
)

func Metric(v telegraf.Metric, err error) telegraf.Metric {
	if err != nil {
		panic(err)
	}
	return v
}

func TestMultiplier(t *testing.T) {
	tests := []struct {
		name       string
		multiplier *Multiplier
		input      telegraf.Metric
		expected   telegraf.Metric
	}{
		{
			name:       "empty",
			multiplier: &Multiplier{},
			input: Metric(
				metric.New(
					"system",
					nil,
					map[string]interface{}{
						"value": 42,
					},
					time.Unix(0, 0),
				),
			),
			expected: Metric(
				metric.New(
					"system",
					nil,
					map[string]interface{}{
						"value": 42,
					},
					time.Unix(0, 0),
				),
			),
		},
		{
			name: "Integer multiplication",
			multiplier: &Multiplier{
				Config: []string{"system value=10"},
			},
			input: Metric(
				metric.New(
					"system",
					nil,
					map[string]interface{}{
						"value": 42,
					},
					time.Unix(0, 0),
				),
			),
			expected: Metric(
				metric.New(
					"system",
					nil,
					map[string]interface{}{
						"value": 420,
					},
					time.Unix(0, 0),
				),
			),
		},
		{
			name: "Float multiplication",
			multiplier: &Multiplier{
				Config: []string{"system value=10"},
			},
			input: Metric(
				metric.New(
					"system",
					nil,
					map[string]interface{}{
						"value": 42.0,
					},
					time.Unix(0, 0),
				),
			),
			expected: Metric(
				metric.New(
					"system",
					nil,
					map[string]interface{}{
						"value": 420.0,
					},
					time.Unix(0, 0),
				),
			),
		},
		{
			name: "Integer devision",
			multiplier: &Multiplier{
				Config: []string{"system value=0.5"},
			},
			input: Metric(
				metric.New(
					"system",
					nil,
					map[string]interface{}{
						"value": 42,
					},
					time.Unix(0, 0),
				),
			),
			expected: Metric(
				metric.New(
					"system",
					nil,
					map[string]interface{}{
						"value": 21,
					},
					time.Unix(0, 0),
				),
			),
		},
		{
			name: "Float devision",
			multiplier: &Multiplier{
				Config: []string{"system value=0.5"},
			},
			input: Metric(
				metric.New(
					"system",
					nil,
					map[string]interface{}{
						"value": 42.0,
					},
					time.Unix(0, 0),
				),
			),
			expected: Metric(
				metric.New(
					"system",
					nil,
					map[string]interface{}{
						"value": 21.0,
					},
					time.Unix(0, 0),
				),
			),
		},
		{
			name: "Integer precision",
			multiplier: &Multiplier{
				Config: []string{"system value=0.001"},
			},
			input: Metric(
				metric.New(
					"system",
					nil,
					map[string]interface{}{
						"value": 999,
					},
					time.Unix(0, 0),
				),
			),
			expected: Metric(
				metric.New(
					"system",
					nil,
					map[string]interface{}{
						"value": 0,
					},
					time.Unix(0, 0),
				),
			),
		},
		{
			name: "Float precision",
			multiplier: &Multiplier{
				Config: []string{"system value=0.001"},
			},
			input: Metric(
				metric.New(
					"system",
					nil,
					map[string]interface{}{
						"value": 999.0,
					},
					time.Unix(0, 0),
				),
			),
			expected: Metric(
				metric.New(
					"system",
					nil,
					map[string]interface{}{
						"value": 0.999,
					},
					time.Unix(0, 0),
				),
			),
		},
		{
			name: "String unchangeability",
			multiplier: &Multiplier{
				Config: []string{"system value=10"},
			},
			input: Metric(
				metric.New(
					"system",
					nil,
					map[string]interface{}{
						"value": "unchanged",
					},
					time.Unix(0, 0),
				),
			),
			expected: Metric(
				metric.New(
					"system",
					nil,
					map[string]interface{}{
						"value": "unchanged",
					},
					time.Unix(0, 0),
				),
			),
		},
		{
			name: "Wrong configuration",
			multiplier: &Multiplier{
				Config: []string{"system value=\"10\""},
			},
			input: Metric(
				metric.New(
					"system",
					nil,
					map[string]interface{}{
						"value": 42,
					},
					time.Unix(0, 0),
				),
			),
			expected: Metric(
				metric.New(
					"system",
					nil,
					map[string]interface{}{
						"value": 0,
					},
					time.Unix(0, 0),
				),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := tt.multiplier.Apply(tt.input)

			require.Equal(t, 1, len(metrics))
			require.Equal(t, tt.expected.Name(), metrics[0].Name())
			require.Equal(t, tt.expected.Tags(), metrics[0].Tags())
			require.Equal(t, tt.expected.Fields(), metrics[0].Fields())
			require.Equal(t, tt.expected.Time(), metrics[0].Time())
		})
	}
}
