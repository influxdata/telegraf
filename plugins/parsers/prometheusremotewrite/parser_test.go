package prometheusremotewrite

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/prometheus/prompb"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

const (
	testCasesDir     = "testcases"
	benchmarkFolder  = "benchmark"
	inputFilename    = "input.json"
	expectedFilename = "expected_v%d.json"
	configFilename   = "config.json"
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
		inputFilename := filepath.Join(testdataPath, inputFilename)
		input, err := loadInput(inputFilename)
		require.NoError(t, err)
		inputBytes, err := input.Marshal()
		require.NoError(t, err)

		// Run tests for both metric versions
		runTestCaseForVersion(t, fname, testdataPath, inputBytes, 1)
		runTestCaseForVersion(t, fname, testdataPath, inputBytes, 2)
	}
}

func BenchmarkParsingMetricVersion1(b *testing.B) {
	parser := newTestParser(map[string]string{}, 1)

	benchmarkData, err := os.ReadFile(filepath.Join(testCasesDir, benchmarkFolder, inputFilename))
	require.NoError(b, err)
	require.NotEmpty(b, benchmarkData)

	for n := 0; n < b.N; n++ {
		//nolint:errcheck // Benchmarking so skip the error check to avoid the unnecessary operations
		parser.Parse(benchmarkData)
	}
}

func BenchmarkParsingMetricVersion2(b *testing.B) {
	parser := newTestParser(map[string]string{}, 2)

	benchmarkData, err := os.ReadFile(filepath.Join(testCasesDir, benchmarkFolder, inputFilename))
	require.NoError(b, err)
	require.NotEmpty(b, benchmarkData)

	for n := 0; n < b.N; n++ {
		//nolint:errcheck // Benchmarking so skip the error check to avoid the unnecessary operations
		parser.Parse(benchmarkData)
	}
}

func runTestCaseForVersion(t *testing.T, caseName, testdataPath string, inputBytes []byte, version int) {
	t.Run(fmt.Sprintf("%s_v%d", caseName, version), func(t *testing.T) {
		// Load parser
		var parser *Parser
		configFilename := filepath.Join(testdataPath, configFilename)
		if _, err := os.Stat(configFilename); os.IsNotExist(err) {
			parser = newTestParser(map[string]string{}, version)
		} else {
			parser, err = loadConfig(configFilename)
			require.NoError(t, err)
			parser.MetricVersion = version
		}

		// Load expected output
		outputFilename := filepath.Join(testdataPath, fmt.Sprintf(expectedFilename, version))
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
			// We need to assign the count field manually
			// as count cannot be assigned from json and can be inferred from other fields anyway
			// This calculation assumes that zero_count is 0
			if h.PositiveDeltas != nil || h.NegativeDeltas != nil {
				// Int histogram always uses PositiveDeltas and NegativeDeltas
				// We calculate count from the deltas and sum them up
				count := int64(0)
				if h.PositiveDeltas != nil {
					cumulative := int64(0)
					for _, pd := range h.PositiveDeltas {
						cumulative += pd
						count += cumulative
					}
				}
				if h.NegativeDeltas != nil {
					cumulative := int64(0)
					for _, nd := range h.NegativeDeltas {
						cumulative += nd
						count += cumulative
					}
				}
				input.Timeseries[i].Histograms[j].Count = &prompb.Histogram_CountInt{
					CountInt: uint64(count),
				}
			} else {
				// Float histogram always uses PositiveCounts and NegativeCounts
				// They are absolute counts, so we just need to sum them up and get the count
				count := float64(0)
				if h.PositiveCounts != nil {
					for _, pd := range h.PositiveCounts {
						count += pd
					}
				}
				if h.NegativeCounts != nil {
					for _, nd := range h.NegativeCounts {
						count += nd
					}
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
		// Convert fields to specific types, as json unmarshal converts all fields to float64
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

func newTestParser(defaultTags map[string]string, metricVersion int) *Parser {
	parser := &Parser{
		DefaultTags:   defaultTags,
		MetricVersion: metricVersion,
	}
	return parser
}
