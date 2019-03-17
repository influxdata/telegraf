package logparser

import (
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/assert"
)

func TestStartNoParsers(t *testing.T) {
	logparser := &LogParserPlugin{
		FromBeginning: true,
		Files:         []string{"testdata/*.log"},
	}

	acc := testutil.Accumulator{}
	assert.Error(t, logparser.Start(&acc))
}

func TestGrokParseLogFilesNonExistPattern(t *testing.T) {
	thisdir := getCurrentDir()

	logparser := &LogParserPlugin{
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
		GrokConfig: GrokConfig{
			MeasurementName:    "logparser_grok",
			Patterns:           []string{"%{TEST_LOG_A}", "%{TEST_LOG_B}"},
			CustomPatternFiles: []string{thisdir + "testdata/test-patterns"},
		},
		FromBeginning: true,
		Files:         []string{thisdir + "testdata/*.log"},
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
		map[string]string{
			"response_code": "200",
			"path":          strings.Replace(thisdir+"testdata/test_a.log", "/", string(os.PathSeparator), -1),
		})

	acc.AssertContainsTaggedFields(t, "logparser_grok",
		map[string]interface{}{
			"myfloat":    1.25,
			"mystring":   "mystring",
			"nomodifier": "nomodifier",
		},
		map[string]string{
			"path": strings.Replace(thisdir+"testdata/test_b.log", "/", string(os.PathSeparator), -1),
		})
}

func TestGrokParseLogFilesAppearLater(t *testing.T) {
	emptydir, err := ioutil.TempDir("", "TestGrokParseLogFilesAppearLater")
	defer os.RemoveAll(emptydir)
	assert.NoError(t, err)

	thisdir := getCurrentDir()

	logparser := &LogParserPlugin{
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

func TestGrokParseLogFilesWithMultiline(t *testing.T) {
	thisdir := getCurrentDir()
	//we make sure the timeout won't kick in
	duration, _ := time.ParseDuration("100s")

	logparser := &LogParserPlugin{
		GrokConfig: GrokConfig{
			MeasurementName:    "logparser_grok",
			Patterns:           []string{"%{TEST_LOG_MULTILINE}"},
			CustomPatternFiles: []string{thisdir + "testdata/test-patterns"},
		},
		MultilineConfig: MultilineConfig{
			Pattern: `^[^\[]`,
			What:    Previous,
			Negate:  false,
			Timeout: &internal.Duration{Duration: duration},
		},
		FromBeginning: true,
		Files:         []string{thisdir + "testdata/multiline/test_multiline.log"},
	}

	acc := testutil.Accumulator{}
	assert.NoError(t, logparser.Start(&acc))
	acc.Wait(3)

	logparser.Stop()

	assert.Equal(t, uint64(3), acc.NMetrics())
	expectedPath := thisdir + "testdata/multiline/test_multiline.log"
	acc.AssertContainsTaggedFields(t, "logparser_grok",
		map[string]interface{}{
			"message": "HelloExample: This is debug",
		},
		map[string]string{
			"path":     expectedPath,
			"loglevel": "DEBUG",
		})
	acc.AssertContainsTaggedFields(t, "logparser_grok",
		map[string]interface{}{
			"message": "HelloExample: This is info",
		},
		map[string]string{
			"path":     expectedPath,
			"loglevel": "INFO",
		})
	acc.AssertContainsTaggedFields(t, "logparser_grok",
		map[string]interface{}{
			"message": "HelloExample: Sorry, something wrong! java.lang.ArithmeticException: / by zero\tat com.foo.HelloExample2.divide(HelloExample2.java:24)\tat com.foo.HelloExample2.main(HelloExample2.java:14)",
		},
		map[string]string{
			"path":     expectedPath,
			"loglevel": "ERROR",
		})
}

func TestGrokParseLogFilesWithMultilineWithTimeout(t *testing.T) {
	thisdir := getCurrentDir()
	// set tight timeout for tests
	duration, _ := time.ParseDuration("10ms")

	logparser := &LogParserPlugin{
		GrokConfig: GrokConfig{
			MeasurementName:    "logparser_grok",
			Patterns:           []string{"%{TEST_LOG_MULTILINE}"},
			CustomPatternFiles: []string{thisdir + "testdata/test-patterns"},
		},
		MultilineConfig: MultilineConfig{
			Pattern: `^[^\[]`,
			What:    Previous,
			Negate:  false,
			Timeout: &internal.Duration{Duration: duration},
		},
		FromBeginning: true,
		Files:         []string{thisdir + "testdata/multiline/test_multiline.log"},
	}

	acc := testutil.Accumulator{}
	assert.NoError(t, logparser.Start(&acc))
	acc.Wait(4)

	logparser.Stop()

	assert.Equal(t, uint64(4), acc.NMetrics())
	expectedPath := thisdir + "testdata/multiline/test_multiline.log"
	acc.AssertContainsTaggedFields(t, "logparser_grok",
		map[string]interface{}{
			"message": "HelloExample: This is warn",
		},
		map[string]string{
			"path":     expectedPath,
			"loglevel": "WARN",
		})
}

func getCurrentDir() string {
	_, filename, _, _ := runtime.Caller(1)
	return strings.Replace(filename, "logparser_test.go", "", 1)
}
