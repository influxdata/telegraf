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

// TestGrokParseLogFilesOneFileRemovedandRecreated tests that after removing and re-creating an observed log file new
// log lines from that file are still monitored
func TestGrokParseLogFilesOneFileRemovedandRecreated(t *testing.T) {
	logdir, err := ioutil.TempDir("", "TestGrokParseLogFilesRemoveOneFile")
	defer os.RemoveAll(logdir)
	assert.NoError(t, err)

	file1, fileErr1 := os.Create(logdir + "/test_1.log")
	assert.NoError(t, fileErr1)

	file2, fileErr2 := os.Create(logdir + "/test_2.log")
	assert.NoError(t, fileErr2)

	thisdir := getCurrentDir()
	p := &grok.Parser{
		Patterns:           []string{"%{TEST_LOG_A}", "%{TEST_LOG_B}"},
		CustomPatternFiles: []string{thisdir + "grok/testdata/test-patterns"},
	}

	logparser := &LogParserPlugin{
		FromBeginning: true,
		Files:         []string{logdir + "/*.log"},
		GrokParser:    p,
	}

	acc := testutil.Accumulator{}
	assert.NoError(t, logparser.Start(&acc))

	acc.SetDebug(true)

	assert.Equal(t, acc.NFields(), 0)

	// write log into file1 and check accumulator
	_, fwErr1 := file1.WriteString("[04/Jun/2016:12:41:45 +0100] 1.25 200 192.168.1.1 5.432µs 101\n")
	assert.NoError(t, fwErr1)

	file1.Sync()

	assert.NoError(t, acc.GatherError(logparser.Gather))
	time.Sleep(100 * time.Millisecond)

	acc.AssertContainsTaggedFields(t, "logparser_grok",
		map[string]interface{}{
			"clientip":      "192.168.1.1",
			"myfloat":       float64(1.25),
			"response_time": int64(5432),
			"myint":         int64(101),
		},
		map[string]string{"response_code": "200"})

	acc.ClearMetrics()

	// write log into file2 and check accumulator
	_, fwErr3 := file2.WriteString("[04/Jun/2016:12:41:45 +0100] 1.25 200 192.168.1.3 5.432µs 101\n")
	assert.NoError(t, fwErr3)

	file2.Sync()

	assert.NoError(t, acc.GatherError(logparser.Gather))
	time.Sleep(100 * time.Millisecond)

	acc.AssertContainsTaggedFields(t, "logparser_grok",
		map[string]interface{}{
			"clientip":      "192.168.1.3",
			"myfloat":       float64(1.25),
			"response_time": int64(5432),
			"myint":         int64(101),
		},
		map[string]string{"response_code": "200"})

	acc.ClearMetrics()

	// close file1 and re-create it at the same place
	closeErr := file1.Close()
	assert.NoError(t, closeErr)

	var removeErr = os.Remove(logdir + "/test_1.log")
	assert.NoError(t, removeErr)

	fileNew1, fileNewErr1 := os.Create(logdir + "/test_1.log")
	assert.NoError(t, fileNewErr1)

	// write log into file2 again and check accumulator
	_, fwErr4 := file2.WriteString("[04/Jun/2016:12:41:45 +0100] 1.25 200 192.168.1.4 5.432µs 101\n")
	assert.NoError(t, fwErr4)

	file2.Sync()

	assert.NoError(t, acc.GatherError(logparser.Gather))
	time.Sleep(100 * time.Millisecond)

	acc.AssertContainsTaggedFields(t, "logparser_grok",
		map[string]interface{}{
			"clientip":      "192.168.1.4",
			"myfloat":       float64(1.25),
			"response_time": int64(5432),
			"myint":         int64(101),
		},
		map[string]string{"response_code": "200"})

	acc.ClearMetrics()

	// write log into re-created file1 again check accumulator
	_, fwErr5 := fileNew1.WriteString("[04/Jun/2016:12:41:45 +0100] 1.25 200 192.168.1.5 5.432µs 101\n")
	assert.NoError(t, fwErr5)

	fileNew1.Sync()

	assert.NoError(t, acc.GatherError(logparser.Gather))
	time.Sleep(100 * time.Millisecond)

	acc.AssertContainsTaggedFields(t, "logparser_grok",
		map[string]interface{}{
			"clientip":      "192.168.1.5",
			"myfloat":       float64(1.25),
			"response_time": int64(5432),
			"myint":         int64(101),
		},
		map[string]string{"response_code": "200"})

	acc.ClearMetrics()

	logparser.Stop()

	file2.Close()
	fileNew1.Close()
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
