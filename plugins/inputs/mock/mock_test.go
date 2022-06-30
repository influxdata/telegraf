package mock

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestGather(t *testing.T) {
	testConstantString := &constant{
		Name:  "constant_string",
		Value: "a string",
	}
	testConstantFloat := &constant{
		Name:  "constant_float",
		Value: 3.1415,
	}
	testConstantInt := &constant{
		Name:  "constant_int",
		Value: 42,
	}
	testConstantBool := &constant{
		Name:  "constant_bool",
		Value: true,
	}
	testRandom := &random{
		Name: "random",
		Min:  1.0,
		Max:  6.0,
	}
	testSineWave := &sineWave{
		Name:      "sine",
		Amplitude: 1.0,
		Period:    0.5,
	}
	testStep := &step{
		Name:  "step",
		Start: 0.0,
		Step:  1.0,
	}
	testStock := &stock{
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

		Constant: []*constant{testConstantString, testConstantFloat, testConstantInt, testConstantBool},
		Random:   []*random{testRandom},
		SineWave: []*sineWave{testSineWave},
		Step:     []*step{testStep},
		Stock:    []*stock{testStock},
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
		case "constant_string":
			require.Equal(t, testConstantString.Value, v)
		case "constant_float":
			require.Equal(t, testConstantFloat.Value, v)
		case "constant_int":
			require.Equal(t, testConstantInt.Value, v)
		case "constant_bool":
			require.Equal(t, testConstantBool.Value, v)
		case "random":
			require.GreaterOrEqual(t, 6.0, v)
			require.LessOrEqual(t, 1.0, v)
		case "sine":
			require.Equal(t, 0.0, v)
		case "step":
			require.Equal(t, 0.0, v)
		default:
			require.Failf(t, "unexpected field %q", k)
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
