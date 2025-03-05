package prometheusremotewrite

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/prometheus/prometheus/model/histogram"
	"github.com/prometheus/prometheus/prompb"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/models"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
	test "github.com/influxdata/telegraf/testutil/plugin_input"
)

const (
	testCasesDir     = "testcases"
	benchmarkFolder  = "benchmark"
	inputFilename    = "input.json"
	expectedFilename = "expected_v%d.out"
	configFilename   = "telegraf.conf"
)

func TestCases(t *testing.T) {
	// Get all directories in testcases
	folders, err := os.ReadDir(testCasesDir)
	require.NoError(t, err)
	// Make sure testdata contains data
	require.NotEmpty(t, folders)

	for _, f := range folders {
		if !f.IsDir() {
			continue
		}

		fname := f.Name()
		testdataPath := filepath.Join(testCasesDir, fname)

		// Load input data
		inputFilePath := filepath.Join(testdataPath, inputFilename)
		buf, err := os.ReadFile(inputFilePath)
		require.NoError(t, err)
		var writeRequest prompb.WriteRequest
		err = jsonpb.Unmarshal(bytes.NewReader(buf), &writeRequest)
		require.NoError(t, err)
		inputBytes, err := writeRequest.Marshal()
		require.NoError(t, err)

		versions := []int{1, 2}
		for _, version := range versions {
			t.Run(fmt.Sprintf("%s_v%d", fname, version), func(t *testing.T) {
				// Load parser
				configFilePath := filepath.Join(testdataPath, configFilename)
				cfg := config.NewConfig()
				require.NoError(t, cfg.LoadConfig(configFilePath))
				require.Len(t, cfg.Inputs, 1)
				plugin := cfg.Inputs[0].Input.(*test.Plugin)
				parser := plugin.Parser.(*models.RunningParser).Parser.(*Parser)
				parser.MetricVersion = version

				// Load expected output
				expectedFilePath := filepath.Join(testdataPath, fmt.Sprintf(expectedFilename, version))
				var expected []telegraf.Metric
				influxParser := &influx.Parser{}
				require.NoError(t, influxParser.Init())
				expected, err := testutil.ParseMetricsFromFile(expectedFilePath, influxParser)
				require.NoError(t, err)

				// Act and assert
				parsed, err := parser.Parse(inputBytes)
				require.NoError(t, err)
				require.Len(t, parsed, len(expected))
				// Ignore type when comparing, because expected metrics are parsed from influx lines and thus always untyped
				testutil.RequireMetricsEqual(t, expected, parsed, testutil.SortMetrics(), testutil.IgnoreType())
			})
		}
	}
}

func BenchmarkParsingMetricVersion1(b *testing.B) {
	parser := &Parser{
		MetricVersion: 1,
	}

	benchmarkData, err := os.ReadFile(filepath.Join(testCasesDir, benchmarkFolder, inputFilename))
	require.NoError(b, err)
	require.NotEmpty(b, benchmarkData)

	for n := 0; n < b.N; n++ {
		//nolint:errcheck // Benchmarking so skip the error check to avoid the unnecessary operations
		parser.Parse(benchmarkData)
	}
}

func BenchmarkParsingMetricVersion2(b *testing.B) {
	parser := &Parser{
		MetricVersion: 2,
	}

	benchmarkData, err := os.ReadFile(filepath.Join(testCasesDir, benchmarkFolder, inputFilename))
	require.NoError(b, err)
	require.NotEmpty(b, benchmarkData)

	for n := 0; n < b.N; n++ {
		//nolint:errcheck // Benchmarking so skip the error check to avoid the unnecessary operations
		parser.Parse(benchmarkData)
	}
}

func TestParse(t *testing.T) {
	prompbInput := prompb.WriteRequest{
		Timeseries: []prompb.TimeSeries{
			{
				Labels: []prompb.Label{
					{Name: "__name__", Value: "go_gc_duration_seconds"},
					{Name: "quantile", Value: "0.99"},
				},
				Samples: []prompb.Sample{
					{Value: 4.63, Timestamp: time.Date(2020, 4, 1, 0, 0, 0, 0, time.UTC).UnixNano()},
				},
			},
			{
				Labels: []prompb.Label{
					{Name: "__name__", Value: "prometheus_target_interval_length_seconds"},
					{Name: "job", Value: "prometheus"},
				},
				Samples: []prompb.Sample{
					{Value: 14.99, Timestamp: time.Date(2020, 4, 1, 0, 0, 0, 0, time.UTC).UnixNano()},
				},
			},
		},
	}

	inoutBytes, err := prompbInput.Marshal()
	require.NoError(t, err)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"prometheus_remote_write",
			map[string]string{
				"quantile": "0.99",
			},
			map[string]interface{}{
				"go_gc_duration_seconds": float64(4.63),
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"prometheus_remote_write",
			map[string]string{
				"job": "prometheus",
			},
			map[string]interface{}{
				"prometheus_target_interval_length_seconds": float64(14.99),
			},
			time.Unix(0, 0),
		),
	}

	parser := Parser{
		DefaultTags: map[string]string{},
	}

	metrics, err := parser.Parse(inoutBytes)
	require.NoError(t, err)
	require.Len(t, metrics, 2)
	testutil.RequireMetricsEqual(t, expected, metrics, testutil.IgnoreTime(), testutil.SortMetrics())
}

func generateTestHistogram(i int) *histogram.Histogram {
	return &histogram.Histogram{
		Count:         12 + uint64(i*9),
		ZeroCount:     2 + uint64(i),
		ZeroThreshold: 0.001,
		Sum:           18.4 * float64(i+1),
		Schema:        1,
		PositiveSpans: []histogram.Span{
			{Offset: 0, Length: 2},
			{Offset: 1, Length: 2},
		},
		PositiveBuckets: []int64{int64(i + 1), 1, -1, 0},
		NegativeSpans: []histogram.Span{
			{Offset: 0, Length: 2},
			{Offset: 1, Length: 2},
		},
		NegativeBuckets: []int64{int64(i + 1), 1, -1, 0},
	}
}

func generateTestFloatHistogram(i int) *histogram.FloatHistogram {
	return &histogram.FloatHistogram{
		Count:         12 + float64(i*9),
		ZeroCount:     2 + float64(i),
		ZeroThreshold: 0.001,
		Sum:           18.4 * float64(i+1),
		Schema:        1,
		PositiveSpans: []histogram.Span{
			{Offset: 0, Length: 2},
			{Offset: 1, Length: 2},
		},
		PositiveBuckets: []float64{float64(i + 1), float64(i + 2), float64(i + 1), float64(i + 1)},
		NegativeSpans: []histogram.Span{
			{Offset: 0, Length: 2},
			{Offset: 1, Length: 2},
		},
		NegativeBuckets: []float64{float64(i + 1), float64(i + 2), float64(i + 1), float64(i + 1)},
	}
}

func TestHistograms(t *testing.T) {
	prompbInput := prompb.WriteRequest{
		Timeseries: []prompb.TimeSeries{
			{
				Labels: []prompb.Label{
					{Name: "__name__", Value: "test_metric_seconds"},
				},
				Histograms: []prompb.Histogram{
					prompb.FromIntHistogram(0, generateTestHistogram(1)),
				},
			},
			{
				Labels: []prompb.Label{
					{Name: "__name__", Value: "test_float_metric_seconds"},
				},
				Histograms: []prompb.Histogram{
					prompb.FromFloatHistogram(0, generateTestFloatHistogram(2)),
				},
			},
		},
	}

	inoutBytes, err := prompbInput.Marshal()
	require.NoError(t, err)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"prometheus_remote_write",
			map[string]string{},
			map[string]interface{}{
				"test_metric_seconds_sum": float64(36.8),
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"prometheus_remote_write",
			map[string]string{},
			map[string]interface{}{
				"test_metric_seconds_count": float64(21),
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"prometheus_remote_write",
			map[string]string{},
			map[string]interface{}{
				"test_float_metric_seconds_sum": float64(55.199999999999996),
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"prometheus_remote_write",
			map[string]string{},
			map[string]interface{}{
				"test_float_metric_seconds_count": float64(30),
			},
			time.Unix(0, 0),
		),
	}

	parser := Parser{
		DefaultTags: map[string]string{},
	}
	metrics, err := parser.Parse(inoutBytes)
	require.NoError(t, err)
	require.Len(t, metrics, 22)
	testutil.RequireMetricsSubset(t, expected, metrics, testutil.IgnoreTime(), testutil.SortMetrics())
}

func TestDefaultTags(t *testing.T) {
	prompbInput := prompb.WriteRequest{
		Timeseries: []prompb.TimeSeries{
			{
				Labels: []prompb.Label{
					{Name: "__name__", Value: "foo"},
					{Name: "__eg__", Value: "bar"},
				},
				Samples: []prompb.Sample{
					{Value: 1, Timestamp: time.Date(2020, 4, 1, 0, 0, 0, 0, time.UTC).UnixNano()},
				},
			},
		},
	}

	inoutBytes, err := prompbInput.Marshal()
	require.NoError(t, err)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"prometheus_remote_write",
			map[string]string{
				"defaultTag": "defaultTagValue",
				"__eg__":     "bar",
			},
			map[string]interface{}{
				"foo": float64(1),
			},
			time.Unix(0, 0),
		),
	}

	parser := Parser{
		DefaultTags: map[string]string{
			"defaultTag": "defaultTagValue",
		},
	}

	metrics, err := parser.Parse(inoutBytes)
	require.NoError(t, err)
	require.Len(t, metrics, 1)
	testutil.RequireMetricsEqual(t, expected, metrics, testutil.IgnoreTime(), testutil.SortMetrics())
}

func TestMetricsWithTimestamp(t *testing.T) {
	testTime := time.Date(2020, time.October, 4, 17, 0, 0, 0, time.UTC)
	testTimeUnix := testTime.UnixNano() / int64(time.Millisecond)
	prompbInput := prompb.WriteRequest{
		Timeseries: []prompb.TimeSeries{
			{
				Labels: []prompb.Label{
					{Name: "__name__", Value: "foo"},
					{Name: "__eg__", Value: "bar"},
				},
				Samples: []prompb.Sample{
					{Value: 1, Timestamp: testTimeUnix},
				},
			},
		},
	}

	inoutBytes, err := prompbInput.Marshal()
	require.NoError(t, err)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"prometheus_remote_write",
			map[string]string{
				"__eg__": "bar",
			},
			map[string]interface{}{
				"foo": float64(1),
			},
			testTime,
		),
	}
	parser := Parser{
		DefaultTags: map[string]string{},
	}

	metrics, err := parser.Parse(inoutBytes)
	require.NoError(t, err)
	require.Len(t, metrics, 1)
	testutil.RequireMetricsEqual(t, expected, metrics, testutil.SortMetrics())
}

var benchmarkData = prompb.WriteRequest{
	Timeseries: []prompb.TimeSeries{
		{
			Labels: []prompb.Label{
				{Name: "__name__", Value: "benchmark_a"},
				{Name: "source", Value: "myhost"},
				{Name: "tags_platform", Value: "python"},
				{Name: "tags_sdkver", Value: "3.11.5"},
			},
			Samples: []prompb.Sample{
				{Value: 5.0, Timestamp: time.Date(2020, 4, 1, 0, 0, 0, 0, time.UTC).UnixMilli()},
			},
		},
		{
			Labels: []prompb.Label{
				{Name: "__name__", Value: "benchmark_b"},
				{Name: "source", Value: "myhost"},
				{Name: "tags_platform", Value: "python"},
				{Name: "tags_sdkver", Value: "3.11.4"},
			},
			Samples: []prompb.Sample{
				{Value: 4.0, Timestamp: time.Date(2020, 4, 1, 0, 0, 0, 0, time.UTC).UnixMilli()},
			},
		},
	},
}

func TestBenchmarkData(t *testing.T) {
	expected := []telegraf.Metric{
		metric.New(
			"prometheus_remote_write",
			map[string]string{
				"source":        "myhost",
				"tags_platform": "python",
				"tags_sdkver":   "3.11.5",
			},
			map[string]interface{}{
				"benchmark_a": 5.0,
			},
			time.Unix(1585699200, 0),
		),
		metric.New(
			"prometheus_remote_write",
			map[string]string{
				"source":        "myhost",
				"tags_platform": "python",
				"tags_sdkver":   "3.11.4",
			},
			map[string]interface{}{
				"benchmark_b": 4.0,
			},
			time.Unix(1585699200, 0),
		),
	}

	benchmarkData, err := benchmarkData.Marshal()
	require.NoError(t, err)

	plugin := &Parser{}
	actual, err := plugin.Parse(benchmarkData)
	require.NoError(t, err)
	testutil.RequireMetricsEqual(t, expected, actual, testutil.SortMetrics())
}

func BenchmarkParsing(b *testing.B) {
	benchmarkData, err := benchmarkData.Marshal()
	require.NoError(b, err)

	plugin := &Parser{}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		//nolint:errcheck // Benchmarking so skip the error check to avoid the unnecessary operations
		plugin.Parse(benchmarkData)
	}
}
