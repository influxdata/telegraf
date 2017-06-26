package logparser

import (
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"

	"github.com/influxdata/telegraf/plugins/inputs/logparser/grok"

	"github.com/stretchr/testify/assert"
)

func TestStartNoParsers(t *testing.T) {
	logparser := &LogParserPlugin{
		FromBeginning: true,
		Files:         []string{"grok/testdata/*.log"},
	}

	acc := testutil.Accumulator{}
	assert.Error(t, logparser.Start(&acc))
}

func TestGrokParseLogFilesNonExistPattern(t *testing.T) {
	thisdir := getCurrentDir()
	p := &grok.Parser{
		Patterns:           []string{"%{FOOBAR}"},
		CustomPatternFiles: []string{thisdir + "grok/testdata/test-patterns"},
	}

	logparser := &LogParserPlugin{
		FromBeginning: true,
		Files:         []string{thisdir + "grok/testdata/*.log"},
		GrokParser:    p,
	}

	acc := testutil.Accumulator{}
	err := logparser.Start(&acc)
	assert.Error(t, err)
}

func TestGrokParseLogFiles(t *testing.T) {
	thisdir := getCurrentDir()
	p := &grok.Parser{
		Patterns:           []string{"%{TEST_LOG_A}", "%{TEST_LOG_B}"},
		CustomPatternFiles: []string{thisdir + "grok/testdata/test-patterns"},
	}

	logparser := &LogParserPlugin{
		FromBeginning: true,
		Files:         []string{thisdir + "grok/testdata/*.log"},
		GrokParser:    p,
	}

	acc := testutil.Accumulator{}
	assert.NoError(t, logparser.Start(&acc))

	acc.Wait(2)

	logparser.Stop()

	acc.AssertContainsTaggedFields(t, "logparser_grok",
		map[string]interface{}{
			"clientip":      "192.168.1.1",
			"myfloat":       float64(1.25),
			"response_time": int64(5432),
			"myint":         int64(101),
		},
		map[string]string{"response_code": "200"})

	acc.AssertContainsTaggedFields(t, "logparser_grok",
		map[string]interface{}{
			"myfloat":    1.25,
			"mystring":   "mystring",
			"nomodifier": "nomodifier",
		},
		map[string]string{})
}

func TestGrokParseLogFilesAppearLater(t *testing.T) {
	emptydir, err := ioutil.TempDir("", "TestGrokParseLogFilesAppearLater")
	defer os.RemoveAll(emptydir)
	assert.NoError(t, err)

	thisdir := getCurrentDir()
	p := &grok.Parser{
		Patterns:           []string{"%{TEST_LOG_A}", "%{TEST_LOG_B}"},
		CustomPatternFiles: []string{thisdir + "grok/testdata/test-patterns"},
	}

	logparser := &LogParserPlugin{
		FromBeginning: true,
		Files:         []string{emptydir + "/*.log"},
		GrokParser:    p,
	}

	acc := testutil.Accumulator{}
	assert.NoError(t, logparser.Start(&acc))

	assert.Equal(t, acc.NFields(), 0)

	_ = os.Symlink(thisdir+"grok/testdata/test_a.log", emptydir+"/test_a.log")
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
		map[string]string{"response_code": "200"})
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
	thisdir := getCurrentDir()
	p := &grok.Parser{
		Patterns:           []string{"%{TEST_LOG_A}", "%{TEST_LOG_BAD}"},
		CustomPatternFiles: []string{thisdir + "grok/testdata/test-patterns"},
	}
	assert.NoError(t, p.Compile())

	logparser := &LogParserPlugin{
		FromBeginning: true,
		Files:         []string{thisdir + "grok/testdata/test_a.log"},
		GrokParser:    p,
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
		map[string]string{"response_code": "200"})
}

func getCurrentDir() string {
	_, filename, _, _ := runtime.Caller(1)
	return strings.Replace(filename, "logparser_test.go", "", 1)
}
