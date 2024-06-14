package derivative

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

var start = metric.New("TestMetric",
	map[string]string{"state": "full"},
	map[string]interface{}{
		"increasing": int64(0),
		"decreasing": int64(100),
		"unchanged":  int64(42),
		"ignored":    "strings are not supported",
		"parameter":  float64(0.0),
	},
	time.Now(),
)

var finish = metric.New("TestMetric",
	map[string]string{"state": "full"},
	map[string]interface{}{
		"increasing": int64(1000),
		"decreasing": int64(0),
		"unchanged":  int64(42),
		"ignored":    "strings are not supported",
		"parameter":  float64(10.0),
	},
	time.Now().Add(time.Second),
)

func TestTwoFullEventsWithParameter(t *testing.T) {
	acc := testutil.Accumulator{}
	derivative := &Derivative{
		Variable: "parameter",
		Suffix:   "_by_parameter",
		cache:    make(map[uint64]*aggregate),
	}
	derivative.Log = testutil.Logger{}
	err := derivative.Init()
	require.NoError(t, err)

	derivative.Add(start)
	derivative.Add(finish)
	derivative.Push(&acc)

	expectedFields := map[string]interface{}{
		"increasing_by_parameter": 100.0,
		"decreasing_by_parameter": -10.0,
		"unchanged_by_parameter":  0.0,
	}
	expectedTags := map[string]string{
		"state": "full",
	}

	acc.AssertContainsTaggedFields(t, "TestMetric", expectedFields, expectedTags)
}

func TestTwoFullEventsWithParameterReverseSequence(t *testing.T) {
	acc := testutil.Accumulator{}
	derivative := &Derivative{
		Variable: "parameter",
		Suffix:   "_by_parameter",
		cache:    make(map[uint64]*aggregate),
	}
	derivative.Log = testutil.Logger{}
	err := derivative.Init()
	require.NoError(t, err)

	derivative.Add(finish)
	derivative.Add(start)
	derivative.Push(&acc)

	expectedFields := map[string]interface{}{
		"increasing_by_parameter": 100.0,
		"decreasing_by_parameter": -10.0,
		"unchanged_by_parameter":  0.0,
	}
	expectedTags := map[string]string{
		"state": "full",
	}

	acc.AssertContainsTaggedFields(t, "TestMetric", expectedFields, expectedTags)
}

func TestTwoFullEventsWithoutParameter(t *testing.T) {
	acc := testutil.Accumulator{}
	derivative := NewDerivative()
	derivative.Log = testutil.Logger{}
	err := derivative.Init()
	require.NoError(t, err)

	startTime := time.Now()
	duration, err := time.ParseDuration("2s")
	require.NoError(t, err)
	endTime := startTime.Add(duration)

	first := metric.New("One Field",
		map[string]string{},
		map[string]interface{}{
			"value": int64(10),
		},
		startTime,
	)
	last := metric.New("One Field",
		map[string]string{},
		map[string]interface{}{
			"value": int64(20),
		},
		endTime,
	)

	derivative.Add(first)
	derivative.Add(last)
	derivative.Push(&acc)

	acc.AssertContainsFields(t,
		"One Field",
		map[string]interface{}{
			"value_rate": float64(5),
		},
	)
}

func TestTwoFullEventsInSeparatePushes(t *testing.T) {
	acc := testutil.Accumulator{}
	derivative := &Derivative{
		Variable:    " parameter",
		Suffix:      "_wrt_parameter",
		MaxRollOver: 10,
		cache:       make(map[uint64]*aggregate),
	}
	derivative.Log = testutil.Logger{}
	err := derivative.Init()
	require.NoError(t, err)

	derivative.Add(start)
	derivative.Push(&acc)

	acc.AssertDoesNotContainMeasurement(t, "TestMetric")

	acc.ClearMetrics()

	derivative.Add(finish)
	derivative.Push(&acc)

	expectedFields := map[string]interface{}{
		"increasing_wrt_parameter": 100.0,
		"decreasing_wrt_parameter": -10.0,
		"unchanged_wrt_parameter":  0.0,
	}
	expectedTags := map[string]string{
		"state": "full",
	}

	acc.AssertContainsTaggedFields(t, "TestMetric", expectedFields, expectedTags)
}

func TestTwoFullEventsInSeparatePushesWithSeveralRollOvers(t *testing.T) {
	acc := testutil.Accumulator{}
	derivative := &Derivative{
		Variable:    "parameter",
		Suffix:      "_wrt_parameter",
		MaxRollOver: 10,
		cache:       make(map[uint64]*aggregate),
	}
	derivative.Log = testutil.Logger{}
	err := derivative.Init()
	require.NoError(t, err)

	derivative.Add(start)
	derivative.Push(&acc)

	acc.AssertDoesNotContainMeasurement(t, "TestMetric")

	derivative.Push(&acc)
	derivative.Push(&acc)
	derivative.Push(&acc)

	derivative.Add(finish)
	derivative.Push(&acc)

	expectedFields := map[string]interface{}{
		"increasing_wrt_parameter": 100.0,
		"decreasing_wrt_parameter": -10.0,
		"unchanged_wrt_parameter":  0.0,
	}

	acc.AssertContainsFields(t, "TestMetric", expectedFields)
}

func TestTwoFullEventsInSeparatePushesWithOutRollOver(t *testing.T) {
	acc := testutil.Accumulator{}
	derivative := &Derivative{
		Variable:    "parameter",
		Suffix:      "_by_parameter",
		MaxRollOver: 0,
		cache:       make(map[uint64]*aggregate),
	}
	derivative.Log = testutil.Logger{}
	err := derivative.Init()
	require.NoError(t, err)

	derivative.Add(start)
	// This test relies on RunningAggregator always callining Reset after Push
	// to remove the first metric after max-rollover of 0 has been reached.
	derivative.Push(&acc)
	derivative.Reset()

	acc.AssertDoesNotContainMeasurement(t, "TestMetric")

	acc.ClearMetrics()
	derivative.Add(finish)
	derivative.Push(&acc)

	acc.AssertDoesNotContainMeasurement(t, "TestMetric")
}

func TestIgnoresMissingVariable(t *testing.T) {
	acc := testutil.Accumulator{}
	derivative := &Derivative{
		Variable: "parameter",
		Suffix:   "_by_parameter",
		cache:    make(map[uint64]*aggregate),
	}
	derivative.Log = testutil.Logger{}
	err := derivative.Init()
	require.NoError(t, err)

	noParameter := metric.New("TestMetric",
		map[string]string{"state": "no_parameter"},
		map[string]interface{}{
			"increasing": int64(100),
			"decreasing": int64(0),
			"unchanged":  int64(42),
		},
		time.Now(),
	)

	derivative.Add(noParameter)
	derivative.Push(&acc)

	acc.AssertDoesNotContainMeasurement(t, "TestMetric")

	acc.ClearMetrics()
	derivative.Add(noParameter)
	derivative.Add(start)
	derivative.Add(noParameter)
	derivative.Add(finish)
	derivative.Add(noParameter)
	derivative.Push(&acc)
	expectedFields := map[string]interface{}{
		"increasing_by_parameter": 100.0,
		"decreasing_by_parameter": -10.0,
		"unchanged_by_parameter":  0.0,
	}
	expectedTags := map[string]string{
		"state": "full",
	}

	acc.AssertContainsTaggedFields(t, "TestMetric", expectedFields, expectedTags)
}

func TestMergesDifferentMetricsWithSameHash(t *testing.T) {
	acc := testutil.Accumulator{}
	derivative := NewDerivative()
	derivative.Log = testutil.Logger{}
	err := derivative.Init()
	require.NoError(t, err)

	startTime := time.Now()
	duration, err := time.ParseDuration("2s")
	require.NoError(t, err)
	endTime := startTime.Add(duration)
	part1 := metric.New("TestMetric",
		map[string]string{"state": "full"},
		map[string]interface{}{"field1": int64(10)},
		startTime,
	)
	part2 := metric.New("TestMetric",
		map[string]string{"state": "full"},
		map[string]interface{}{"field2": int64(20)},
		startTime,
	)
	final := metric.New("TestMetric",
		map[string]string{"state": "full"},
		map[string]interface{}{
			"field1": int64(30),
			"field2": int64(30),
		},
		endTime,
	)

	derivative.Add(part1)
	derivative.Push(&acc)
	derivative.Add(part2)
	derivative.Push(&acc)
	derivative.Add(final)
	derivative.Push(&acc)

	expectedFields := map[string]interface{}{
		"field1_rate": 10.0,
		"field2_rate": 5.0,
	}
	expectedTags := map[string]string{
		"state": "full",
	}

	acc.AssertContainsTaggedFields(t, "TestMetric", expectedFields, expectedTags)
}

func TestDropsAggregatesOnMaxRollOver(t *testing.T) {
	acc := testutil.Accumulator{}
	derivative := &Derivative{
		MaxRollOver: 1,
		cache:       make(map[uint64]*aggregate),
	}
	derivative.Log = testutil.Logger{}
	err := derivative.Init()
	require.NoError(t, err)

	derivative.Add(start)
	derivative.Push(&acc)
	derivative.Reset()
	derivative.Push(&acc)
	derivative.Reset()
	derivative.Add(finish)
	derivative.Push(&acc)
	derivative.Reset()

	acc.AssertDoesNotContainMeasurement(t, "TestMetric")
}

func TestAddMetricsResetsRollOver(t *testing.T) {
	acc := testutil.Accumulator{}
	derivative := &Derivative{
		Variable:    "parameter",
		Suffix:      "_by_parameter",
		MaxRollOver: 1,
		cache:       make(map[uint64]*aggregate),
		Log:         testutil.Logger{},
	}
	err := derivative.Init()
	require.NoError(t, err)

	derivative.Add(start)
	derivative.Push(&acc)
	derivative.Reset()
	derivative.Add(start)
	derivative.Reset()
	derivative.Add(finish)
	derivative.Push(&acc)

	expectedFields := map[string]interface{}{
		"increasing_by_parameter": 100.0,
		"decreasing_by_parameter": -10.0,
		"unchanged_by_parameter":  0.0,
	}

	acc.AssertContainsFields(t, "TestMetric", expectedFields)
}

func TestCalculatesCorrectDerivativeOnTwoConsecutivePeriods(t *testing.T) {
	acc := testutil.Accumulator{}
	period, err := time.ParseDuration("10s")
	require.NoError(t, err)
	derivative := NewDerivative()
	derivative.Log = testutil.Logger{}
	require.NoError(t, derivative.Init())

	startTime := time.Now()
	first := metric.New("One Field",
		map[string]string{},
		map[string]interface{}{
			"value": int64(10),
		},
		startTime,
	)
	derivative.Add(first)
	derivative.Push(&acc)
	derivative.Reset()

	second := metric.New("One Field",
		map[string]string{},
		map[string]interface{}{
			"value": int64(20),
		},
		startTime.Add(period),
	)
	derivative.Add(second)
	derivative.Push(&acc)
	derivative.Reset()

	acc.AssertContainsFields(t, "One Field", map[string]interface{}{
		"value_rate": 1.0,
	})

	acc.ClearMetrics()
	third := metric.New("One Field",
		map[string]string{},
		map[string]interface{}{
			"value": int64(40),
		},
		startTime.Add(period).Add(period),
	)
	derivative.Add(third)
	derivative.Push(&acc)
	derivative.Reset()

	acc.AssertContainsFields(t, "One Field", map[string]interface{}{
		"value_rate": 2.0,
	})
}
