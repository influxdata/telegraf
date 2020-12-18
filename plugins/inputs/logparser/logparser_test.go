package logparser

import (
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartNoParsers(t *testing.T) {
	logparser := &LogParserPlugin{
		Log:           testutil.Logger{},
		FromBeginning: true,
		Files:         []string{"testdata/*.log"},
	}

	acc := testutil.Accumulator{}
	assert.Error(t, logparser.Start(&acc))
}

func TestGrokParseLogFilesNonExistPattern(t *testing.T) {
	thisdir := getCurrentDir()

	logparser := &LogParserPlugin{
		Log:           testutil.Logger{},
		FromBeginning: true,
		Files:         []string{thisdir + "testdata/*.log"},
		GrokConfig: GrokConfig{
			Patterns:           []string{"%{FOOBAR}"},
			CustomPatternFiles: []string{thisdir + "testdata/test-patterns"},
		},
	}

	acc := testutil.Accumulator{}
	err := logparser.Start(&acc)
	assert.Error(t, err)
}

func TestGrokParseLogFiles(t *testing.T) {
	thisdir := getCurrentDir()

	logparser := &LogParserPlugin{
		Log: testutil.Logger{},
		GrokConfig: GrokConfig{
			MeasurementName:    "logparser_grok",
			Patterns:           []string{"%{TEST_LOG_A}", "%{TEST_LOG_B}", "%{TEST_LOG_C}"},
			CustomPatternFiles: []string{thisdir + "testdata/test-patterns"},
		},
		FromBeginning: true,
		Files:         []string{thisdir + "testdata/*.log"},
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
				"path":          thisdir + "testdata/test_a.log",
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
				"path": thisdir + "testdata/test_b.log",
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
				"path":          thisdir + "testdata/test_c.log",
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

	thisdir := getCurrentDir()

	logparser := &LogParserPlugin{
		Log:           testutil.Logger{},
		FromBeginning: true,
		Files:         []string{emptydir + "/*.log"},
		GrokConfig: GrokConfig{
			MeasurementName:    "logparser_grok",
			Patterns:           []string{"%{TEST_LOG_A}", "%{TEST_LOG_B}"},
			CustomPatternFiles: []string{thisdir + "testdata/test-patterns"},
		},
	}

	acc := testutil.Accumulator{}
	assert.NoError(t, logparser.Start(&acc))

	assert.Equal(t, acc.NFields(), 0)

	_ = os.Symlink(thisdir+"testdata/test_a.log", emptydir+"/test_a.log")
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
			"path":          emptydir + "/test_a.log",
		})
}

// Test that test_a.log line gets parsed even though we don't have the correct
// pattern available for test_b.log
func TestGrokParseLogFilesOneBad(t *testing.T) {
	thisdir := getCurrentDir()

	logparser := &LogParserPlugin{
		Log:           testutil.Logger{},
		FromBeginning: true,
		Files:         []string{thisdir + "testdata/test_a.log"},
		GrokConfig: GrokConfig{
			MeasurementName:    "logparser_grok",
			Patterns:           []string{"%{TEST_LOG_A}", "%{TEST_LOG_BAD}"},
			CustomPatternFiles: []string{thisdir + "testdata/test-patterns"},
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
			"path":          thisdir + "testdata/test_a.log",
		})
}

func TestGrokParseLogFiles_TimestampInEpochMilli(t *testing.T) {
	thisdir := getCurrentDir()

	logparser := &LogParserPlugin{
		Log: testutil.Logger{},
		GrokConfig: GrokConfig{
			MeasurementName:    "logparser_grok",
			Patterns:           []string{"%{TEST_LOG_C}"},
			CustomPatternFiles: []string{thisdir + "testdata/test-patterns"},
		},
		FromBeginning: true,
		Files:         []string{thisdir + "testdata/test_c.log"},
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
			"path":          thisdir + "testdata/test_c.log",
		})
}

func getCurrentDir() string {
	_, filename, _, _ := runtime.Caller(1)
	return strings.Replace(filename, "logparser_test.go", "", 1)
}
