package prometheusremotewrite

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/prometheus/prompb"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

const testCasesDir = "testcases"

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
		inputFilename := filepath.Join(testdataPath, "input.json")
		input, err := loadInput(inputFilename)
		require.NoError(t, err)
		inputBytes, err := input.Marshal()
		require.NoError(t, err)

		// Run tests for both metric versions
		runTestCaseForVersion(t, fname, testdataPath, inputBytes, 1)
		runTestCaseForVersion(t, fname, testdataPath, inputBytes, 2)
	}
}

func runTestCaseForVersion(t *testing.T, caseName, testdataPath string, inputBytes []byte, version int) {
	t.Run(caseName+"_v"+strconv.Itoa(version), func(t *testing.T) {
		// Load parser
		var parser *Parser
		configFilename := filepath.Join(testdataPath, "config.json")
		if _, err := os.Stat(configFilename); os.IsNotExist(err) {
			parser = newTestParser(map[string]string{}, version)
		} else {
			parser, err = loadConfig(configFilename)
			require.NoError(t, err)
			parser.MetricVersion = version
		}

		// Load expected output
		outputFilename := filepath.Join(testdataPath, "expected_v"+strconv.Itoa(version)+".json")
		expected, err := loadExpected(outputFilename)
		require.NoError(t, err)

		// Test
		parsed, err := parser.Parse(inputBytes)
		require.NoError(t, err)
		require.Len(t, parsed, len(expected))
		testutil.RequireMetricsEqual(t, expected, parsed, testutil.SortMetrics())
	})
}

func loadConfig(path string) (*Parser, error) {
	var config Parser
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func loadInput(path string) (prompb.WriteRequest, error) {
	var input prompb.WriteRequest
	data, err := os.ReadFile(path)
	if err != nil {
		return input, err
	}
	err = json.Unmarshal(data, &input)
	if err != nil {
		return input, err
	}
	for i, ts := range input.Timeseries {
		for j, h := range ts.Histograms {
			// We need to assign the count field manually as it cannot be assigned from json
			// This calculation assumes that zero_count is 0
			if h.PositiveDeltas != nil || h.NegativeDeltas != nil {
				// Int histogram always uses PositiveDeltas and NegativeDeltas
				// We calculate count from the deltas and sum them up
				count := int64(0)
				cumulative := int64(0)
				for _, pd := range h.PositiveDeltas {
					cumulative += pd
					count += cumulative
				}
				cumulative = 0
				for _, nd := range h.NegativeDeltas {
					cumulative += nd
					count += cumulative
				}
				input.Timeseries[i].Histograms[j].Count = &prompb.Histogram_CountInt{
					CountInt: uint64(count),
				}
			} else {
				// Float histogram always uses PositiveCounts and NegativeCounts
				// We just need to sum them up and get the count
				count := float64(0)
				for _, pd := range h.PositiveCounts {
					count += pd
				}
				for _, nd := range h.NegativeCounts {
					count += nd
				}
				input.Timeseries[i].Histograms[j].Count = &prompb.Histogram_CountFloat{
					CountFloat: count,
				}
			}
		}
	}
	return input, err
}

func loadExpected(path string) ([]telegraf.Metric, error) {
	var expected []struct {
		Name      string                 `json:"name"`
		Tags      map[string]string      `json:"tags"`
		Fields    map[string]interface{} `json:"fields"`
		Timestamp int64                  `json:"timestamp"`
		ValueType telegraf.ValueType     `json:"value_type"`
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &expected)
	if err != nil {
		return nil, err
	}

	var metrics []telegraf.Metric
	for _, e := range expected {
		// Convert fields to specific types
		fields := make(map[string]interface{})
		for k, v := range e.Fields {
			switch {
			case k == "schema" ||
				(strings.HasPrefix(k, "positive_span_") && strings.HasSuffix(k, "_offset")) ||
				(strings.HasPrefix(k, "negative_span_") && strings.HasSuffix(k, "_offset")):
				fields[k] = int64(v.(float64))
			case k == "counter_reset_hint" ||
				(strings.HasPrefix(k, "positive_span_") && strings.HasSuffix(k, "_length")) ||
				(strings.HasPrefix(k, "negative_span_") && strings.HasSuffix(k, "_length")):
				fields[k] = uint64(v.(float64))
			default:
				fields[k] = v
			}
		}
		m := metric.New(e.Name, e.Tags, fields, time.Unix(0, e.Timestamp), e.ValueType)
		metrics = append(metrics, m)
	}
	return metrics, nil
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

	plugin := newTestParser(map[string]string{}, 2)

	actual, err := plugin.Parse(benchmarkData)
	require.NoError(t, err)
	testutil.RequireMetricsEqual(t, expected, actual, testutil.SortMetrics())
}

func BenchmarkParsing(b *testing.B) {
	benchmarkData, err := benchmarkData.Marshal()
	require.NoError(b, err)

	plugin := newTestParser(map[string]string{}, 2)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		//nolint:errcheck // Benchmarking so skip the error check to avoid the unnecessary operations
		plugin.Parse(benchmarkData)
	}
}

func newTestParser(defaultTags map[string]string, metricVersion int) *Parser {
	parser := &Parser{
		DefaultTags:   defaultTags,
		MetricVersion: metricVersion,
	}
	return parser
}
