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

	/*
		tests := []struct {
				name      string
				converter *Converter
				input     telegraf.Metric
				expected  []telegraf.Metric
			}
	*/

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
					},
				},
			},
			input: testutil.MustMetric(
				"CPU metrics",
				map[string]string{},
				map[string]interface{}{
					"usage" : "",
				},
				time.Unix(0, 0),
			),
		},
	}

	for _, scenario := range scenarios {
		/*
		tt.converter.Log = testutil.Logger{}

				err := tt.converter.Init()
				require.NoError(t, err)
				actual := tt.converter.Apply(tt.input)

				testutil.RequireMetricsEqual(t, tt.expected, actual)*/
		err := scenario.defaulter.Init()
		assert.NoError(t, err, "There was an error initializing the Defaulter")
	}
}
