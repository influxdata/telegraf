package mock

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestGather(t *testing.T) {
	testRandom := &Random{
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
		MetricName: "test",
		Tags:       tags,

		Random:   []*Random{testRandom},
		SineWave: []*SineWave{testSineWave},
		Step:     []*Step{testStep},
		Stock:    []*Stock{testStock},
	}

	var acc testutil.Accumulator
	require.NoError(t, m.Gather(&acc))

	require.Len(t, acc.Metrics, 1)

	metric := acc.Metrics[0]
	require.Equal(t, "test", metric.Measurement)
	require.Equal(t, tags, metric.Tags)
	for k, v := range metric.Fields {
		switch k {
		case "abc":
			require.Equal(t, 50.0, v)
		case "random":
			require.GreaterOrEqual(t, 6.0, v)
			require.LessOrEqual(t, 1.0, v)
		case "sine":
			require.Equal(t, 0.0, v)
		case "step":
			require.Equal(t, 0.0, v)
		default:
			t.Errorf("unexpected field %q", k)
			t.Fail()
		}
	}
}

func TestGatherEmpty(t *testing.T) {
	m := &Mock{
		MetricName: "test_empty",
	}

	var acc testutil.Accumulator
	require.NoError(t, m.Gather(&acc))

	acc.AssertDoesNotContainMeasurement(t, "test_empty")
}
