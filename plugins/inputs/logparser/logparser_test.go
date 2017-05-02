package logparser

import (
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"testing"

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

	os.Symlink(
		thisdir+"grok/testdata/test_a.log",
		emptydir+"/test_a.log")
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
