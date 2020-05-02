package defaulter

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

func TestDefaulter(t *testing.T) {
	assert.Equal(t, 1, 1)

	scenarios := []struct {
		name      string
		defaulter *Defaulter
		input     telegraf.Metric
		expected  []telegraf.Metric
	}{
		{
			name: "Test that no values are changed since they are not nil or empty",
			defaulter: &Defaulter{
				DefaultFieldsSets: map[string]interface{}{
					"usage":     30,
					"wind_feel": "very chill",
					"is_dead":   true,
				},
			},
			input: testutil.MustMetric(
				"CPU metrics",
				map[string]string{},
				map[string]interface{}{
					"usage":     45,
					"wind_feel": "a dragon's breath",
					"is_dead":   false,
				},
				time.Unix(0, 0),
			),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"CPU metrics",
					map[string]string{},
					map[string]interface{}{
						"usage":     45,
						"wind_feel": "a dragon's breath",
						"is_dead":   false,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "Tests that the missing fields are set on the metric",
			defaulter: &Defaulter{
				DefaultFieldsSets: map[string]interface{}{
					"max_clock_gz":  6,
					"wind_feel":     "Unknown",
					"boost_enabled": false,
					"variance":      1.2,
				},
			},
			input: testutil.MustMetric(
				"CPU metrics",
				map[string]string{},
				map[string]interface{}{
					"usage":       45,
					"temperature": 64,
				},
				time.Unix(0, 0),
			),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"CPU metrics",
					map[string]string{},
					map[string]interface{}{
						"usage":         45,
						"temperature":   64,
						"max_clock_gz":  6,
						"wind_feel":     "Unknown",
						"boost_enabled": false,
						"variance":      1.2,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "Tests that empty fields are replaced by specified defaults",
			defaulter: &Defaulter{
				DefaultFieldsSets: map[string]interface{}{
					"max_clock_gz":  6,
					"wind_feel":     "Unknown",
					"boost_enabled": false,
				},
			},
			input: testutil.MustMetric(
				"CPU metrics",
				map[string]string{},
				map[string]interface{}{
					"max_clock_gz": "",
					"wind_feel":    " ",
				},
				time.Unix(0, 0),
			),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"CPU metrics",
					map[string]string{},
					map[string]interface{}{
						"max_clock_gz":  6,
						"wind_feel":     "Unknown",
						"boost_enabled": false,
					},
					time.Unix(0, 0),
				),
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			defaulter := scenario.defaulter
			defaulter.Log = testutil.Logger{}

			resultMetrics := defaulter.Apply(scenario.input)
			assert.Len(t, resultMetrics, 1)
			testutil.RequireMetricsEqual(t, scenario.expected, resultMetrics)
		})
	}
}
