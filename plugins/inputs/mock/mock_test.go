package mock

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGather(t *testing.T) {
	testRandomFloat := &RandomFloat{
		Name: "random",
		Min:  1.0,
		Max:  6.0,
	}
	testSineWave := &SineWave{
		Name:      "sine",
		Amplitude: 10,
	}
	testStep := &Step{
		Name:  "step",
		Start: 0.0,
		Step:  1.0,
	}
	testStock := &Stock{
		Name:       "abc",
		Price:      50.00,
		Volatility: 0.2,
	}

	tags := map[string]string{
		"buildling": "tbd",
		"site":      "nowhere",
	}

	m := &Mock{
		counter:    0.0,
		MetricName: "test",
		Tags:       tags,

		RandomFloat: []*RandomFloat{testRandomFloat},
		SineWave:    []*SineWave{testSineWave},
		Step:        []*Step{testStep},
		Stock:       []*Stock{testStock},
	}

	var acc testutil.Accumulator
	require.NoError(t, m.Gather(&acc))

	assert.Len(t, acc.Metrics, 1)

	metric := acc.Metrics[0]
	assert.Equal(t, "test", metric.Measurement)
	assert.Equal(t, tags, metric.Tags)
	for k, v := range metric.Fields {
		if k == "abc" {
			assert.Equal(t, 50.0, v)
		} else if k == "random" {
			assert.GreaterOrEqual(t, 6.0, v)
			assert.LessOrEqual(t, 1.0, v)
		} else if k == "sine" {
			assert.Equal(t, 0.0, v)
		} else if k == "step" {
			assert.Equal(t, 0.0, v)
		}
	}
}

func TestGatherEmpty(t *testing.T) {
	m := &Mock{
		counter:    0.0,
		MetricName: "test_empty",
	}

	var acc testutil.Accumulator
	require.NoError(t, m.Gather(&acc))

	acc.AssertDoesNotContainMeasurement(t, "test_empty")
}
