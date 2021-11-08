package starlark

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	common "github.com/influxdata/telegraf/plugins/common/starlark"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

var m1 = metric.New("m1",
	map[string]string{"foo": "bar"},
	map[string]interface{}{
		"a": int64(1),
		"b": int64(1),
		"c": int64(1),
		"d": int64(1),
		"e": int64(1),
		"f": int64(2),
		"g": int64(2),
		"h": int64(2),
		"i": int64(2),
		"j": int64(3),
	},
	time.Now(),
)
var m2 = metric.New("m1",
	map[string]string{"foo": "bar"},
	map[string]interface{}{
		"a":        int64(1),
		"b":        int64(3),
		"c":        int64(3),
		"d":        int64(3),
		"e":        int64(3),
		"f":        int64(1),
		"g":        int64(1),
		"h":        int64(1),
		"i":        int64(1),
		"j":        int64(1),
		"k":        int64(200),
		"l":        int64(200),
		"ignoreme": "string",
		"andme":    true,
	},
	time.Now(),
)

func BenchmarkApply(b *testing.B) {
	minmax, _ := newMinMax()

	for n := 0; n < b.N; n++ {
		minmax.Add(m1)
		minmax.Add(m2)
	}
}

// Test two metrics getting added.
func TestMinMaxWithPeriod(t *testing.T) {
	acc := testutil.Accumulator{}
	minmax, err := newMinMax()
	require.NoError(t, err)

	minmax.Add(m1)
	minmax.Add(m2)
	minmax.Push(&acc)

	expectedFields := map[string]interface{}{
		"a_max": int64(1),
		"a_min": int64(1),
		"b_max": int64(3),
		"b_min": int64(1),
		"c_max": int64(3),
		"c_min": int64(1),
		"d_max": int64(3),
		"d_min": int64(1),
		"e_max": int64(3),
		"e_min": int64(1),
		"f_max": int64(2),
		"f_min": int64(1),
		"g_max": int64(2),
		"g_min": int64(1),
		"h_max": int64(2),
		"h_min": int64(1),
		"i_max": int64(2),
		"i_min": int64(1),
		"j_max": int64(3),
		"j_min": int64(1),
		"k_max": int64(200),
		"k_min": int64(200),
		"l_max": int64(200),
		"l_min": int64(200),
	}
	expectedTags := map[string]string{
		"foo": "bar",
	}
	acc.AssertContainsTaggedFields(t, "m1", expectedFields, expectedTags)
}

// Test two metrics getting added with a push/reset in between (simulates
// getting added in different periods.)
func TestMinMaxDifferentPeriods(t *testing.T) {
	acc := testutil.Accumulator{}
	minmax, err := newMinMax()
	require.NoError(t, err)
	minmax.Add(m1)
	minmax.Push(&acc)
	expectedFields := map[string]interface{}{
		"a_max": int64(1),
		"a_min": int64(1),
		"b_max": int64(1),
		"b_min": int64(1),
		"c_max": int64(1),
		"c_min": int64(1),
		"d_max": int64(1),
		"d_min": int64(1),
		"e_max": int64(1),
		"e_min": int64(1),
		"f_max": int64(2),
		"f_min": int64(2),
		"g_max": int64(2),
		"g_min": int64(2),
		"h_max": int64(2),
		"h_min": int64(2),
		"i_max": int64(2),
		"i_min": int64(2),
		"j_max": int64(3),
		"j_min": int64(3),
	}
	expectedTags := map[string]string{
		"foo": "bar",
	}
	acc.AssertContainsTaggedFields(t, "m1", expectedFields, expectedTags)

	acc.ClearMetrics()
	minmax.Reset()
	minmax.Add(m2)
	minmax.Push(&acc)
	expectedFields = map[string]interface{}{
		"a_max": int64(1),
		"a_min": int64(1),
		"b_max": int64(3),
		"b_min": int64(3),
		"c_max": int64(3),
		"c_min": int64(3),
		"d_max": int64(3),
		"d_min": int64(3),
		"e_max": int64(3),
		"e_min": int64(3),
		"f_max": int64(1),
		"f_min": int64(1),
		"g_max": int64(1),
		"g_min": int64(1),
		"h_max": int64(1),
		"h_min": int64(1),
		"i_max": int64(1),
		"i_min": int64(1),
		"j_max": int64(1),
		"j_min": int64(1),
		"k_max": int64(200),
		"k_min": int64(200),
		"l_max": int64(200),
		"l_min": int64(200),
	}
	expectedTags = map[string]string{
		"foo": "bar",
	}
	acc.AssertContainsTaggedFields(t, "m1", expectedFields, expectedTags)
}

func newMinMax() (*Starlark, error) {
	return newStarlarkFromScript("testdata/min_max.star")
}

func TestSimple(t *testing.T) {
	plugin, err := newMerge()
	require.NoError(t, err)

	plugin.Add(
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"cpu": "cpu0",
			},
			map[string]interface{}{
				"time_idle": 42,
			},
			time.Unix(0, 0),
		),
	)
	require.NoError(t, err)

	plugin.Add(
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"cpu": "cpu0",
			},
			map[string]interface{}{
				"time_guest": 42,
			},
			time.Unix(0, 0),
		),
	)
	require.NoError(t, err)

	var acc testutil.Accumulator
	plugin.Push(&acc)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"cpu": "cpu0",
			},
			map[string]interface{}{
				"time_idle":  42,
				"time_guest": 42,
			},
			time.Unix(0, 0),
		),
	}

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics())
}

func TestNanosecondPrecision(t *testing.T) {
	plugin, err := newMerge()

	require.NoError(t, err)

	plugin.Add(
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"cpu": "cpu0",
			},
			map[string]interface{}{
				"time_idle": 42,
			},
			time.Unix(0, 1),
		),
	)
	require.NoError(t, err)

	plugin.Add(
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"cpu": "cpu0",
			},
			map[string]interface{}{
				"time_guest": 42,
			},
			time.Unix(0, 1),
		),
	)
	require.NoError(t, err)

	var acc testutil.Accumulator
	acc.SetPrecision(time.Second)
	plugin.Push(&acc)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"cpu": "cpu0",
			},
			map[string]interface{}{
				"time_idle":  42,
				"time_guest": 42,
			},
			time.Unix(0, 1),
		),
	}

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics())
}

func TestReset(t *testing.T) {
	plugin, err := newMerge()

	require.NoError(t, err)

	plugin.Add(
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"cpu": "cpu0",
			},
			map[string]interface{}{
				"time_idle": 42,
			},
			time.Unix(0, 0),
		),
	)
	require.NoError(t, err)

	var acc testutil.Accumulator
	plugin.Push(&acc)

	plugin.Reset()

	plugin.Add(
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"cpu": "cpu0",
			},
			map[string]interface{}{
				"time_guest": 42,
			},
			time.Unix(0, 0),
		),
	)
	require.NoError(t, err)

	plugin.Push(&acc)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"cpu": "cpu0",
			},
			map[string]interface{}{
				"time_idle": 42,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"cpu": "cpu0",
			},
			map[string]interface{}{
				"time_guest": 42,
			},
			time.Unix(0, 0),
		),
	}

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics())
}

func newMerge() (*Starlark, error) {
	return newStarlarkFromScript("testdata/merge.star")
}

func TestLastFromSource(t *testing.T) {
	acc := testutil.Accumulator{}
	plugin, err := newStarlarkFromSource(`
state = {}
def add(metric):
  state["last"] = metric

def push():
  return state.get("last")

def reset():
  state.clear()
`)
	require.NoError(t, err)
	plugin.Add(
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"cpu": "cpu0",
			},
			map[string]interface{}{
				"time_idle": 42,
			},
			time.Unix(0, 0),
		),
	)
	require.NoError(t, err)
	plugin.Add(
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"cpu": "cpu2",
			},
			map[string]interface{}{
				"time_idle": 31,
			},
			time.Unix(0, 0),
		),
	)
	require.NoError(t, err)
	plugin.Push(&acc)
	expectedFields := map[string]interface{}{
		"time_idle": int64(31),
	}
	expectedTags := map[string]string{
		"cpu": "cpu2",
	}
	acc.AssertContainsTaggedFields(t, "cpu", expectedFields, expectedTags)
	plugin.Reset()
}

func newStarlarkFromSource(source string) (*Starlark, error) {
	plugin := &Starlark{
		StarlarkCommon: common.StarlarkCommon{
			StarlarkLoadFunc: common.LoadFunc,
			Log:              testutil.Logger{},
			Source:           source,
		},
	}
	err := plugin.Init()
	if err != nil {
		return nil, err
	}
	return plugin, nil
}

func newStarlarkFromScript(script string) (*Starlark, error) {
	plugin := &Starlark{
		StarlarkCommon: common.StarlarkCommon{
			StarlarkLoadFunc: common.LoadFunc,
			Log:              testutil.Logger{},
			Script:           script,
		},
	}
	err := plugin.Init()
	if err != nil {
		return nil, err
	}
	return plugin, nil
}
