package quantile

import (
	"math/rand"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestConfigInvalidAlgorithm(t *testing.T) {
	q := Quantile{AlgorithmType: "a strange one"}
	err := q.Init()
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown algorithm type")
}

func TestConfigInvalidCompression(t *testing.T) {
	q := Quantile{Compression: 0, AlgorithmType: "t-digest"}
	err := q.Init()
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot create \"t-digest\" algorithm")
}

func TestConfigInvalidQuantiles(t *testing.T) {
	q := Quantile{Compression: 100, Quantiles: []float64{-0.5}}
	err := q.Init()
	require.Error(t, err)
	require.Contains(t, err.Error(), "quantile -0.5 out of range")

	q = Quantile{Compression: 100, Quantiles: []float64{1.5}}
	err = q.Init()
	require.Error(t, err)
	require.Contains(t, err.Error(), "quantile 1.5 out of range")

	q = Quantile{Compression: 100, Quantiles: []float64{0.1, 0.2, 0.3, 0.1}}
	err = q.Init()
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate quantile")
}

func TestSingleMetricTDigest(t *testing.T) {
	acc := testutil.Accumulator{}

	q := Quantile{Compression: 100}
	err := q.Init()
	require.NoError(t, err)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"test",
			map[string]string{"foo": "bar"},
			map[string]interface{}{
				"a_025": 24.75,
				"a_050": 49.50,
				"a_075": 74.25,
				"b_025": 24.75,
				"b_050": 49.50,
				"b_075": 74.25,
				"c_025": 24.75,
				"c_050": 49.50,
				"c_075": 74.25,
				"d_025": 24.75,
				"d_050": 49.50,
				"d_075": 74.25,
				"e_025": 24.75,
				"e_050": 49.50,
				"e_075": 74.25,
				"f_025": 24.75,
				"f_050": 49.50,
				"f_075": 74.25,
				"g_025": 0.2475,
				"g_050": 0.4950,
				"g_075": 0.7425,
			},
			time.Now(),
		),
	}

	metrics := make([]telegraf.Metric, 100)
	for i := range metrics {
		metrics[i] = testutil.MustMetric(
			"test",
			map[string]string{"foo": "bar"},
			map[string]interface{}{
				"a":  int32(i),
				"b":  int64(i),
				"c":  uint32(i),
				"d":  uint64(i),
				"e":  float32(i),
				"f":  float64(i),
				"g":  float64(i) / 100.0,
				"x1": "string",
				"x2": true,
			},
			time.Now(),
		)
	}

	for _, m := range metrics {
		q.Add(m)
	}
	q.Push(&acc)

	epsilon := cmpopts.EquateApprox(0, 1e-3)
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), epsilon)
}

func TestMultipleMetricsTDigest(t *testing.T) {
	acc := testutil.Accumulator{}

	q := Quantile{Compression: 100}
	err := q.Init()
	require.NoError(t, err)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"test",
			map[string]string{"series": "foo"},
			map[string]interface{}{
				"a_025": 24.75, "a_050": 49.50, "a_075": 74.25,
				"b_025": 24.75, "b_050": 49.50, "b_075": 74.25,
			},
			time.Now(),
		),
		testutil.MustMetric(
			"test",
			map[string]string{"series": "bar"},
			map[string]interface{}{
				"a_025": 49.50, "a_050": 99.00, "a_075": 148.50,
				"b_025": 49.50, "b_050": 99.00, "b_075": 148.50,
			},
			time.Now(),
		),
	}

	metricsA := make([]telegraf.Metric, 100)
	metricsB := make([]telegraf.Metric, 100)
	for i := range metricsA {
		metricsA[i] = testutil.MustMetric(
			"test",
			map[string]string{"series": "foo"},
			map[string]interface{}{"a": int64(i), "b": float64(i), "x1": "string", "x2": true},
			time.Now(),
		)
	}
	for i := range metricsB {
		metricsB[i] = testutil.MustMetric(
			"test",
			map[string]string{"series": "bar"},
			map[string]interface{}{"a": int64(2 * i), "b": float64(2 * i), "x1": "string", "x2": true},
			time.Now(),
		)
	}

	for _, m := range metricsA {
		q.Add(m)
	}
	for _, m := range metricsB {
		q.Add(m)
	}
	q.Push(&acc)

	epsilon := cmpopts.EquateApprox(0, 1e-3)
	sort := testutil.SortMetrics()
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), epsilon, sort)
}

func TestSingleMetricExactR7(t *testing.T) {
	acc := testutil.Accumulator{}

	q := Quantile{AlgorithmType: "exact R7"}
	err := q.Init()
	require.NoError(t, err)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"test",
			map[string]string{"foo": "bar"},
			map[string]interface{}{
				"a_025": 24.75,
				"a_050": 49.50,
				"a_075": 74.25,
				"b_025": 24.75,
				"b_050": 49.50,
				"b_075": 74.25,
				"c_025": 24.75,
				"c_050": 49.50,
				"c_075": 74.25,
				"d_025": 24.75,
				"d_050": 49.50,
				"d_075": 74.25,
				"e_025": 24.75,
				"e_050": 49.50,
				"e_075": 74.25,
				"f_025": 24.75,
				"f_050": 49.50,
				"f_075": 74.25,
				"g_025": 0.2475,
				"g_050": 0.4950,
				"g_075": 0.7425,
			},
			time.Now(),
		),
	}

	metrics := make([]telegraf.Metric, 100)
	for i := range metrics {
		metrics[i] = testutil.MustMetric(
			"test",
			map[string]string{"foo": "bar"},
			map[string]interface{}{
				"a":  int32(i),
				"b":  int64(i),
				"c":  uint32(i),
				"d":  uint64(i),
				"e":  float32(i),
				"f":  float64(i),
				"g":  float64(i) / 100.0,
				"x1": "string",
				"x2": true,
			},
			time.Now(),
		)
	}

	for _, m := range metrics {
		q.Add(m)
	}
	q.Push(&acc)

	epsilon := cmpopts.EquateApprox(0, 1e-3)
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), epsilon)
}

func TestMultipleMetricsExactR7(t *testing.T) {
	acc := testutil.Accumulator{}

	q := Quantile{AlgorithmType: "exact R7"}
	err := q.Init()
	require.NoError(t, err)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"test",
			map[string]string{"series": "foo"},
			map[string]interface{}{
				"a_025": 24.75, "a_050": 49.50, "a_075": 74.25,
				"b_025": 24.75, "b_050": 49.50, "b_075": 74.25,
			},
			time.Now(),
		),
		testutil.MustMetric(
			"test",
			map[string]string{"series": "bar"},
			map[string]interface{}{
				"a_025": 49.50, "a_050": 99.00, "a_075": 148.50,
				"b_025": 49.50, "b_050": 99.00, "b_075": 148.50,
			},
			time.Now(),
		),
	}

	metricsA := make([]telegraf.Metric, 100)
	metricsB := make([]telegraf.Metric, 100)
	for i := range metricsA {
		metricsA[i] = testutil.MustMetric(
			"test",
			map[string]string{"series": "foo"},
			map[string]interface{}{"a": int64(i), "b": float64(i), "x1": "string", "x2": true},
			time.Now(),
		)
	}
	for i := range metricsB {
		metricsB[i] = testutil.MustMetric(
			"test",
			map[string]string{"series": "bar"},
			map[string]interface{}{"a": int64(2 * i), "b": float64(2 * i), "x1": "string", "x2": true},
			time.Now(),
		)
	}

	for _, m := range metricsA {
		q.Add(m)
	}
	for _, m := range metricsB {
		q.Add(m)
	}
	q.Push(&acc)

	epsilon := cmpopts.EquateApprox(0, 1e-3)
	sort := testutil.SortMetrics()
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), epsilon, sort)
}

func TestSingleMetricExactR8(t *testing.T) {
	acc := testutil.Accumulator{}

	q := Quantile{AlgorithmType: "exact R8"}
	err := q.Init()
	require.NoError(t, err)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"test",
			map[string]string{"foo": "bar"},
			map[string]interface{}{
				"a_025": 24.417,
				"a_050": 49.500,
				"a_075": 74.583,
				"b_025": 24.417,
				"b_050": 49.500,
				"b_075": 74.583,
				"c_025": 24.417,
				"c_050": 49.500,
				"c_075": 74.583,
				"d_025": 24.417,
				"d_050": 49.500,
				"d_075": 74.583,
				"e_025": 24.417,
				"e_050": 49.500,
				"e_075": 74.583,
				"f_025": 24.417,
				"f_050": 49.500,
				"f_075": 74.583,
				"g_025": 0.24417,
				"g_050": 0.49500,
				"g_075": 0.74583,
			},
			time.Now(),
		),
	}

	metrics := make([]telegraf.Metric, 100)
	for i := range metrics {
		metrics[i] = testutil.MustMetric(
			"test",
			map[string]string{"foo": "bar"},
			map[string]interface{}{
				"a":  int32(i),
				"b":  int64(i),
				"c":  uint32(i),
				"d":  uint64(i),
				"e":  float32(i),
				"f":  float64(i),
				"g":  float64(i) / 100.0,
				"x1": "string",
				"x2": true,
			},
			time.Now(),
		)
	}

	for _, m := range metrics {
		q.Add(m)
	}
	q.Push(&acc)

	epsilon := cmpopts.EquateApprox(0, 1e-3)
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), epsilon)
}

func TestMultipleMetricsExactR8(t *testing.T) {
	acc := testutil.Accumulator{}

	q := Quantile{AlgorithmType: "exact R8"}
	err := q.Init()
	require.NoError(t, err)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"test",
			map[string]string{"series": "foo"},
			map[string]interface{}{
				"a_025": 24.417, "a_050": 49.500, "a_075": 74.583,
				"b_025": 24.417, "b_050": 49.500, "b_075": 74.583,
			},
			time.Now(),
		),
		testutil.MustMetric(
			"test",
			map[string]string{"series": "bar"},
			map[string]interface{}{
				"a_025": 48.833, "a_050": 99.000, "a_075": 149.167,
				"b_025": 48.833, "b_050": 99.000, "b_075": 149.167,
			},
			time.Now(),
		),
	}

	metricsA := make([]telegraf.Metric, 100)
	metricsB := make([]telegraf.Metric, 100)
	for i := range metricsA {
		metricsA[i] = testutil.MustMetric(
			"test",
			map[string]string{"series": "foo"},
			map[string]interface{}{"a": int64(i), "b": float64(i), "x1": "string", "x2": true},
			time.Now(),
		)
	}
	for i := range metricsB {
		metricsB[i] = testutil.MustMetric(
			"test",
			map[string]string{"series": "bar"},
			map[string]interface{}{"a": int64(2 * i), "b": float64(2 * i), "x1": "string", "x2": true},
			time.Now(),
		)
	}

	for _, m := range metricsA {
		q.Add(m)
	}
	for _, m := range metricsB {
		q.Add(m)
	}
	q.Push(&acc)

	epsilon := cmpopts.EquateApprox(0, 1e-3)
	sort := testutil.SortMetrics()
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), epsilon, sort)
}

func BenchmarkDefaultTDigest(b *testing.B) {
	metrics := make([]telegraf.Metric, 100)
	for i := range metrics {
		metrics[i] = testutil.MustMetric(
			"test",
			map[string]string{"foo": "bar"},
			map[string]interface{}{
				"a":  rand.Int31(),
				"b":  rand.Int63(),
				"c":  rand.Uint32(),
				"d":  rand.Uint64(),
				"e":  rand.Float32(),
				"f":  rand.Float64(),
				"x1": "string",
				"x2": true,
			},
			time.Now(),
		)
	}

	q := Quantile{Compression: 100}
	err := q.Init()
	require.NoError(b, err)

	acc := testutil.Accumulator{}
	for n := 0; n < b.N; n++ {
		for _, m := range metrics {
			q.Add(m)
		}
		q.Push(&acc)
	}
}

func BenchmarkDefaultTDigest100Q(b *testing.B) {
	metrics := make([]telegraf.Metric, 100)
	for i := range metrics {
		metrics[i] = testutil.MustMetric(
			"test",
			map[string]string{"foo": "bar"},
			map[string]interface{}{
				"a":  rand.Int31(),
				"b":  rand.Int63(),
				"c":  rand.Uint32(),
				"d":  rand.Uint64(),
				"e":  rand.Float32(),
				"f":  rand.Float64(),
				"x1": "string",
				"x2": true,
			},
			time.Now(),
		)
	}
	quantiles := make([]float64, 100)
	for i := range quantiles {
		quantiles[i] = 0.01 * float64(i)
	}

	q := Quantile{Compression: 100, Quantiles: quantiles}
	err := q.Init()
	require.NoError(b, err)

	acc := testutil.Accumulator{}
	for n := 0; n < b.N; n++ {
		for _, m := range metrics {
			q.Add(m)
		}
		q.Push(&acc)
	}
}

func BenchmarkDefaultExactR7(b *testing.B) {
	metrics := make([]telegraf.Metric, 100)
	for i := range metrics {
		metrics[i] = testutil.MustMetric(
			"test",
			map[string]string{"foo": "bar"},
			map[string]interface{}{
				"a":  rand.Int31(),
				"b":  rand.Int63(),
				"c":  rand.Uint32(),
				"d":  rand.Uint64(),
				"e":  rand.Float32(),
				"f":  rand.Float64(),
				"x1": "string",
				"x2": true,
			},
			time.Now(),
		)
	}

	q := Quantile{AlgorithmType: "exact R7"}
	err := q.Init()
	require.NoError(b, err)

	acc := testutil.Accumulator{}
	for n := 0; n < b.N; n++ {
		for _, m := range metrics {
			q.Add(m)
		}
		q.Push(&acc)
	}
}

func BenchmarkDefaultExactR7100Q(b *testing.B) {
	metrics := make([]telegraf.Metric, 100)
	for i := range metrics {
		metrics[i] = testutil.MustMetric(
			"test",
			map[string]string{"foo": "bar"},
			map[string]interface{}{
				"a":  rand.Int31(),
				"b":  rand.Int63(),
				"c":  rand.Uint32(),
				"d":  rand.Uint64(),
				"e":  rand.Float32(),
				"f":  rand.Float64(),
				"x1": "string",
				"x2": true,
			},
			time.Now(),
		)
	}
	quantiles := make([]float64, 100)
	for i := range quantiles {
		quantiles[i] = 0.01 * float64(i)
	}

	q := Quantile{AlgorithmType: "exact R7", Quantiles: quantiles}
	err := q.Init()
	require.NoError(b, err)

	acc := testutil.Accumulator{}
	for n := 0; n < b.N; n++ {
		for _, m := range metrics {
			q.Add(m)
		}
		q.Push(&acc)
	}
}

func BenchmarkDefaultExactR8(b *testing.B) {
	metrics := make([]telegraf.Metric, 100)
	for i := range metrics {
		metrics[i] = testutil.MustMetric(
			"test",
			map[string]string{"foo": "bar"},
			map[string]interface{}{
				"a":  rand.Int31(),
				"b":  rand.Int63(),
				"c":  rand.Uint32(),
				"d":  rand.Uint64(),
				"e":  rand.Float32(),
				"f":  rand.Float64(),
				"x1": "string",
				"x2": true,
			},
			time.Now(),
		)
	}

	q := Quantile{AlgorithmType: "exact R8"}
	err := q.Init()
	require.NoError(b, err)

	acc := testutil.Accumulator{}
	for n := 0; n < b.N; n++ {
		for _, m := range metrics {
			q.Add(m)
		}
		q.Push(&acc)
	}
}

func BenchmarkDefaultExactR8100Q(b *testing.B) {
	metrics := make([]telegraf.Metric, 100)
	for i := range metrics {
		metrics[i] = testutil.MustMetric(
			"test",
			map[string]string{"foo": "bar"},
			map[string]interface{}{
				"a":  rand.Int31(),
				"b":  rand.Int63(),
				"c":  rand.Uint32(),
				"d":  rand.Uint64(),
				"e":  rand.Float32(),
				"f":  rand.Float64(),
				"x1": "string",
				"x2": true,
			},
			time.Now(),
		)
	}
	quantiles := make([]float64, 100)
	for i := range quantiles {
		quantiles[i] = 0.01 * float64(i)
	}

	q := Quantile{AlgorithmType: "exact R8", Quantiles: quantiles}
	err := q.Init()
	require.NoError(b, err)

	acc := testutil.Accumulator{}
	for n := 0; n < b.N; n++ {
		for _, m := range metrics {
			q.Add(m)
		}
		q.Push(&acc)
	}
}
