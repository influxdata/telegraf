package logparser

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

var (
	testdataDir = getTestdataDir()
)

func TestStartNoParsers(t *testing.T) {
	logparser := &LogParserPlugin{
		Log:           testutil.Logger{},
		FromBeginning: true,
		Files:         []string{filepath.Join(testdataDir, "*.log")},
	}

	acc := testutil.Accumulator{}
	assert.Error(t, logparser.Start(&acc))
}

func TestGrokParseLogFilesNonExistPattern(t *testing.T) {
	logparser := &LogParserPlugin{
		Log:           testutil.Logger{},
		FromBeginning: true,
		Files:         []string{filepath.Join(testdataDir, "*.log")},
		GrokConfig: GrokConfig{
			Patterns:           []string{"%{FOOBAR}"},
			CustomPatternFiles: []string{filepath.Join(testdataDir, "test-patterns")},
		},
	}

	acc := testutil.Accumulator{}
	err := logparser.Start(&acc)
	assert.Error(t, err)
}

func TestGrokParseLogFiles(t *testing.T) {
	logparser := &LogParserPlugin{
		Log: testutil.Logger{},
		GrokConfig: GrokConfig{
			MeasurementName:    "logparser_grok",
			Patterns:           []string{"%{TEST_LOG_A}", "%{TEST_LOG_B}", "%{TEST_LOG_C}"},
			CustomPatternFiles: []string{filepath.Join(testdataDir, "test-patterns")},
		},
		FromBeginning: true,
		Files:         []string{filepath.Join(testdataDir, "*.log")},
	}

	acc := testutil.Accumulator{}
	require.NoError(t, logparser.Start(&acc))
	acc.Wait(3)

	logparser.Stop()

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"logparser_grok",
			map[string]string{
				"response_code": "200",
				"path":          filepath.Join(testdataDir, "test_a.log"),
			},
			map[string]interface{}{
				"clientip":      "192.168.1.1",
				"myfloat":       float64(1.25),
				"response_time": int64(5432),
				"myint":         int64(101),
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"logparser_grok",
			map[string]string{
				"path": filepath.Join(testdataDir, "test_b.log"),
			},
			map[string]interface{}{
				"myfloat":    1.25,
				"mystring":   "mystring",
				"nomodifier": "nomodifier",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"logparser_grok",
			map[string]string{
				"path":          filepath.Join(testdataDir, "test_c.log"),
				"response_code": "200",
			},
			map[string]interface{}{
				"clientip":      "192.168.1.1",
				"myfloat":       1.25,
				"myint":         101,
				"response_time": 5432,
			},
			time.Unix(0, 0),
		),
	}

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(),
		testutil.IgnoreTime(), testutil.SortMetrics())
}

func TestGrokParseLogFilesAppearLater(t *testing.T) {
	emptydir, err := ioutil.TempDir("", "TestGrokParseLogFilesAppearLater")
	defer os.RemoveAll(emptydir)
	assert.NoError(t, err)

	logparser := &LogParserPlugin{
		Log:           testutil.Logger{},
		FromBeginning: true,
		Files:         []string{filepath.Join(emptydir, "*.log")},
		GrokConfig: GrokConfig{
			MeasurementName:    "logparser_grok",
			Patterns:           []string{"%{TEST_LOG_A}", "%{TEST_LOG_B}"},
			CustomPatternFiles: []string{filepath.Join(testdataDir, "test-patterns")},
		},
	}

	acc := testutil.Accumulator{}
	assert.NoError(t, logparser.Start(&acc))

	assert.Equal(t, acc.NFields(), 0)

	input, err := ioutil.ReadFile(filepath.Join(testdataDir, "test_a.log"))
	assert.NoError(t, err)

	err = ioutil.WriteFile(filepath.Join(emptydir, "test_a.log"), input, 0644)
	assert.NoError(t, err)

	assert.NoError(t, acc.GatherError(logparser.Gather))
	acc.Wait(1)

	logparser.Stop()

	acc.AssertContainsTaggedFields(t, "logparser_grok",
		map[string]interface{}{
			"clientip":      "192.168.1.1",
			"myfloat":       float64(1.25),
			"response_time": int64(5432),
			"myint":         int64(101),
		},
		map[string]string{
			"response_code": "200",
			"path":          filepath.Join(emptydir, "test_a.log"),
		})
}

// Test that test_a.log line gets parsed even though we don't have the correct
// pattern available for test_b.log
func TestGrokParseLogFilesOneBad(t *testing.T) {
	logparser := &LogParserPlugin{
		Log:           testutil.Logger{},
		FromBeginning: true,
		Files:         []string{filepath.Join(testdataDir, "test_a.log")},
		GrokConfig: GrokConfig{
			MeasurementName:    "logparser_grok",
			Patterns:           []string{"%{TEST_LOG_A}", "%{TEST_LOG_BAD}"},
			CustomPatternFiles: []string{filepath.Join(testdataDir, "test-patterns")},
		},
	}

	acc := testutil.Accumulator{}
	acc.SetDebug(true)
	assert.NoError(t, logparser.Start(&acc))

	acc.Wait(1)
	logparser.Stop()

	acc.AssertContainsTaggedFields(t, "logparser_grok",
		map[string]interface{}{
			"clientip":      "192.168.1.1",
			"myfloat":       float64(1.25),
			"response_time": int64(5432),
			"myint":         int64(101),
		},
		map[string]string{
			"response_code": "200",
			"path":          filepath.Join(testdataDir, "test_a.log"),
		})
}

func TestGrokParseLogFiles_TimestampInEpochMilli(t *testing.T) {
	logparser := &LogParserPlugin{
		Log: testutil.Logger{},
		GrokConfig: GrokConfig{
			MeasurementName:    "logparser_grok",
			Patterns:           []string{"%{TEST_LOG_C}"},
			CustomPatternFiles: []string{filepath.Join(testdataDir, "test-patterns")},
		},
		FromBeginning: true,
		Files:         []string{filepath.Join(testdataDir, "test_c.log")},
	}

	acc := testutil.Accumulator{}
	acc.SetDebug(true)
	assert.NoError(t, logparser.Start(&acc))
	acc.Wait(1)

	logparser.Stop()

	acc.AssertContainsTaggedFields(t, "logparser_grok",
		map[string]interface{}{
			"clientip":      "192.168.1.1",
			"myfloat":       float64(1.25),
			"response_time": int64(5432),
			"myint":         int64(101),
		},
		map[string]string{
			"response_code": "200",
			"path":          filepath.Join(testdataDir, "test_c.log"),
		})
}

func getTestdataDir() string {
	dir, err := os.Getwd()
	if err != nil {
		// if we cannot even establish the test directory, further progress is meaningless
		panic(err)
	}

	return filepath.Join(dir, "testdata")
}
