package prometheusremotewrite

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/models"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/prometheus/prometheus/prompb"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
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
		jsonpb.Unmarshal(bytes.NewReader(buf), &writeRequest)
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
				outputFilename := filepath.Join(testdataPath, fmt.Sprintf(expectedFilename, version))
				var expected []telegraf.Metric
				influxParser := &influx.Parser{}
				require.NoError(t, influxParser.Init())
				expected, err := testutil.ParseMetricsFromFile(outputFilename, influxParser)
				require.NoError(t, err)

				// Test
				parsed, err := parser.Parse(inputBytes)
				require.NoError(t, err)
				require.Len(t, parsed, len(expected))
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
