package defaulter

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
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
			name: "Test single default set",
			defaulter: &Defaulter{
				DefaultFieldsSets: []DefaultFieldsSet{
					{
						Fields: []string{"usage", "temperature", "is_dead"},
						Value:  "Foobar",
						Metric: "CPU metrics",
					},
				},
			},
			input: testutil.MustMetric(
				"CPU metrics",
				map[string]string{},
				map[string]interface{}{
					"usage":       "30%",
					"temperature": "70F",
					"is_dead":     "nopes",
				},
				time.Unix(0, 0),
			),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"CPU metrics",
					map[string]string{},
					map[string]interface{}{
						"usage":       "30%",
						"temperature": "70F",
						"is_dead":     "nopes",
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "Test single default set",
			defaulter: &Defaulter{
				DefaultFieldsSets: []DefaultFieldsSet{
					{
						Fields: []string{"usage", "temperature", "is_dead"},
						Value:  "Foobar",
						Metric: "CPU metrics",
					},
				},
			},
			input: testutil.MustMetric(
				"CPU metrics",
				map[string]string{},
				map[string]interface{}{
					"usage":       "",
					"temperature": "0",
				},
				time.Unix(0, 0),
			),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"CPU metrics",
					map[string]string{},
					map[string]interface{}{
						"usage":       "Foobar",
						"temperature": "Foobar",
						"is_dead":     "Foobar",
					},
					time.Unix(0, 0),
				),
			},
		},
	}

	for _, scenario := range scenarios {
		defaulter := scenario.defaulter
		err := defaulter.Init()
		assert.NoError(t, err, "There was an error initializing the Defaulter")

		resultMetrics := defaulter.Apply(scenario.input)
		assert.Len(t, resultMetrics, 1)

		testutil.RequireMetricsEqual(t, scenario.expected, resultMetrics)
	}
}
