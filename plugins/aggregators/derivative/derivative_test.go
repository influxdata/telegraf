package derivative

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

var start, _ = metric.New("Test Metric",
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

var finish, _ = metric.New("Test Metric",
	map[string]string{"state": "full"},
	map[string]interface{}{
		"increasing": int64(1000),
		"decreasing": int64(0),
		"unchanged":  int64(42),
		"ignored":    "strings are not supported",
		"parameter":  float64(10.0),
	},
	time.Now(),
)

func TestTwoFullEventsWithParameter(t *testing.T) {
	acc := testutil.Accumulator{}
	derivative := &Derivative{
		Variable: "parameter",
		Infix:    "_by_",
		cache:    make(map[uint64]aggregate),
	}
	derivative.Init()

	emitMetrics(&acc, derivative, start, finish)

	expectedFields := map[string]interface{}{
		"increasing_by_parameter": 100.0,
		"decreasing_by_parameter": -10.0,
		"unchanged_by_parameter":  0.0,
	}
	expectedTags := map[string]string{
		"state": "full",
	}
	acc.AssertContainsTaggedFields(t, "Test Metric", expectedFields, expectedTags)
}

func emitMetrics(acc *testutil.Accumulator, aggregator telegraf.Aggregator, metrics ...telegraf.Metric) {
	for _, metric := range metrics {
		aggregator.Add(metric)
	}
	aggregator.Push(acc)
	aggregator.Reset()
}

func TestTwoFullEventsWithParameterReverseSequence(t *testing.T) {
	acc := testutil.Accumulator{}
	derivative := &Derivative{
		Variable: "parameter",
		Infix:    "_by_",
		cache:    make(map[uint64]aggregate),
	}
	derivative.Init()

	emitMetrics(&acc, derivative, finish, start)

	expectedFields := map[string]interface{}{
		"increasing_by_parameter": 100.0,
		"decreasing_by_parameter": -10.0,
		"unchanged_by_parameter":  0.0,
	}
	expectedTags := map[string]string{
		"state": "full",
	}
	acc.AssertContainsTaggedFields(t, "Test Metric", expectedFields, expectedTags)
}

func TestTwoFullEventsWithoutParameter(t *testing.T) {
	acc := testutil.Accumulator{}
	derivative := NewDerivative()
	startTime := time.Now()
	duration, _ := time.ParseDuration("2s")
	endTime := startTime.Add(duration)

	first, _ := metric.New("One Field",
		map[string]string{},
		map[string]interface{}{
			"value": int64(10),
		},
		startTime,
	)
	last, _ := metric.New("One Field",
		map[string]string{},
		map[string]interface{}{
			"value": int64(20),
		},
		endTime,
	)

	emitMetrics(&acc, derivative, first, last)

	acc.AssertContainsFields(t,
		"One Field",
		map[string]interface{}{
			"value_by_seconds": float64(5),
		},
	)

}

func TestTwoFullEventsInSeperatePushes(t *testing.T) {
	acc := testutil.Accumulator{}
	derivative := &Derivative{
		Variable:    " parameter",
		Infix:       "_wrt_",
		MaxRollOver: 10,
		cache:       make(map[uint64]aggregate),
	}
	derivative.Init()
	
	emitMetrics(&acc, derivative, start)
	acc.AssertDoesNotContainMeasurement(t, "Test Metric")

	acc.ClearMetrics()
	emitMetrics(&acc, derivative, finish)

	expectedFields := map[string]interface{}{
		"increasing_wrt_parameter": 100.0,
		"decreasing_wrt_parameter": -10.0,
		"unchanged_wrt_parameter":  0.0,
	}
	expectedTags := map[string]string{
		"state": "full",
	}
	acc.AssertContainsTaggedFields(t, "Test Metric", expectedFields, expectedTags)
}

func TestTwoFullEventsInSeperatePushesWithSeveralRollOvers(t *testing.T) {
	acc := testutil.Accumulator{}
	derivative := &Derivative{
		Variable:    "parameter",
		Infix:       "_wrt_",
		MaxRollOver: 10,
		cache:       make(map[uint64]aggregate),
	}
	derivative.Init()

	emitMetrics(&acc, derivative, start)
	acc.AssertDoesNotContainMeasurement(t, "Test Metric")

	emitMetrics(&acc, derivative)
	emitMetrics(&acc, derivative)
	emitMetrics(&acc, derivative)
	emitMetrics(&acc, derivative, finish)

	expectedFields := map[string]interface{}{
		"increasing_wrt_parameter": 100.0,
		"decreasing_wrt_parameter": -10.0,
		"unchanged_wrt_parameter":  0.0,
	}
	acc.AssertContainsFields(t, "Test Metric", expectedFields)
}

func TestTwoFullEventsInSeperatePushesWithOutRollOver(t *testing.T) {
	acc := testutil.Accumulator{}
	derivative := &Derivative{
		Variable:    "parameter",
		Infix:       "_by_",
		MaxRollOver: 0,
		cache:       make(map[uint64]aggregate),
	}
	derivative.Init()

	emitMetrics(&acc, derivative, start)
	acc.AssertDoesNotContainMeasurement(t, "Test Metric")

	acc.ClearMetrics()
	emitMetrics(&acc, derivative, finish)
	acc.AssertDoesNotContainMeasurement(t, "Test Metric")
}

func TestIgnoresMissingVariable(t *testing.T) {
	acc := testutil.Accumulator{}
	derivative := &Derivative{
		Variable: "parameter",
		Infix:    "_by_",
		cache:    make(map[uint64]aggregate),
	}
	derivative.Init()

	noParameter, _ := metric.New("Test Metric",
		map[string]string{"state": "no_parameter"},
		map[string]interface{}{
			"increasing": int64(100),
			"decreasing": int64(0),
			"unchanged":  int64(42),
		},
		time.Now(),
	)

	emitMetrics(&acc, derivative, noParameter)
	acc.AssertDoesNotContainMeasurement(t, "Test Metric")

	acc.ClearMetrics()
	emitMetrics(&acc, derivative, noParameter, start, noParameter, finish, noParameter)
	expectedFields := map[string]interface{}{
		"increasing_by_parameter": 100.0,
		"decreasing_by_parameter": -10.0,
		"unchanged_by_parameter":  0.0,
	}
	expectedTags := map[string]string{
		"state": "full",
	}
	acc.AssertContainsTaggedFields(t, "Test Metric", expectedFields, expectedTags)
}

func TestMergesDifferenMetricsWithSameHash(t *testing.T) {
	acc := testutil.Accumulator{}
	derivative := NewDerivative()

	startTime := time.Now()
	duration, _ := time.ParseDuration("2s")
	endTime := startTime.Add(duration)
	part1, _ := metric.New("Test Metric",
		map[string]string{"state": "full"},
		map[string]interface{}{"field1": int64(10)},
		startTime,
	)
	part2, _ := metric.New("Test Metric",
		map[string]string{"state": "full"},
		map[string]interface{}{"field2": int64(20)},
		startTime,
	)
	final, _ := metric.New("Test Metric",
		map[string]string{"state": "full"},
		map[string]interface{}{
			"field1": int64(30),
			"field2": int64(30),
		},
		endTime,
	)

	emitMetrics(&acc, derivative, part1)
	emitMetrics(&acc, derivative, part2)
	emitMetrics(&acc, derivative, final)

	expectedFields := map[string]interface{}{
		"field1_by_seconds": 10.0,
		"field2_by_seconds": 5.0,
	}
	expectedTags := map[string]string{
		"state": "full",
	}
	acc.AssertContainsTaggedFields(t, "Test Metric", expectedFields, expectedTags)
}

func TestDropsAggregatesOnMaxRollOver(t *testing.T) {
	acc := testutil.Accumulator{}
	derivative := &Derivative{
		Infix:       "_by_",
		MaxRollOver: 1,
		cache:       make(map[uint64]aggregate),
	}
	derivative.Init()

	emitMetrics(&acc, derivative, start)
	emitMetrics(&acc, derivative)
	emitMetrics(&acc, derivative, finish)

	acc.AssertDoesNotContainMeasurement(t, "Test Metric")
}

func TestAddMetricsResetsRollOver(t *testing.T) {
	acc := testutil.Accumulator{}
	derivative := &Derivative{
		Variable:    "parameter",
		Infix:       "_by_",
		MaxRollOver: 1,
		cache:       make(map[uint64]aggregate),
	}
	derivative.Init()

	emitMetrics(&acc, derivative, start)
	emitMetrics(&acc, derivative, start)
	emitMetrics(&acc, derivative, finish)

	expectedFields := map[string]interface{}{
		"increasing_by_parameter": 100.0,
		"decreasing_by_parameter": -10.0,
		"unchanged_by_parameter":  0.0,
	}
	acc.AssertContainsFields(t, "Test Metric", expectedFields)
}
