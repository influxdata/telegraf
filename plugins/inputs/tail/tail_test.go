package tail

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/parsers/csv"
	"github.com/influxdata/telegraf/plugins/parsers/json"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTailFromBeginning(t *testing.T) {
	if os.Getenv("CIRCLE_PROJECT_REPONAME") != "" {
		t.Skip("Skipping CI testing due to race conditions")
	}

	tmpfile, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()
	_, err = tmpfile.WriteString("cpu,mytag=foo usage_idle=100\n")
	require.NoError(t, err)

	tt := NewTail()
	tt.Log = testutil.Logger{}
	tt.FromBeginning = true
	tt.Files = []string{tmpfile.Name()}
	tt.SetParserFunc(parsers.NewInfluxParser)

	err = tt.Init()
	require.NoError(t, err)

	acc := testutil.Accumulator{}
	require.NoError(t, tt.Start(&acc))
	defer tt.Stop()
	require.NoError(t, acc.GatherError(tt.Gather))

	acc.Wait(1)
	acc.AssertContainsTaggedFields(t, "cpu",
		map[string]interface{}{
			"usage_idle": float64(100),
		},
		map[string]string{
			"mytag": "foo",
			"path":  tmpfile.Name(),
		})
}

func TestTailFromEnd(t *testing.T) {
	if os.Getenv("CIRCLE_PROJECT_REPONAME") != "" {
		t.Skip("Skipping CI testing due to race conditions")
	}

	tmpfile, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()
	_, err = tmpfile.WriteString("cpu,mytag=foo usage_idle=100\n")
	require.NoError(t, err)

	tt := NewTail()
	tt.Log = testutil.Logger{}
	tt.Files = []string{tmpfile.Name()}
	tt.SetParserFunc(parsers.NewInfluxParser)

	err = tt.Init()
	require.NoError(t, err)

	acc := testutil.Accumulator{}
	require.NoError(t, tt.Start(&acc))
	defer tt.Stop()
	for _, tailer := range tt.tailers {
		for n, err := tailer.Tell(); err == nil && n == 0; n, err = tailer.Tell() {
			// wait for tailer to jump to end
			runtime.Gosched()
		}
	}

	_, err = tmpfile.WriteString("cpu,othertag=foo usage_idle=100\n")
	require.NoError(t, err)
	require.NoError(t, acc.GatherError(tt.Gather))

	acc.Wait(1)
	acc.AssertContainsTaggedFields(t, "cpu",
		map[string]interface{}{
			"usage_idle": float64(100),
		},
		map[string]string{
			"othertag": "foo",
			"path":     tmpfile.Name(),
		})
	assert.Len(t, acc.Metrics, 1)
}

func TestTailBadLine(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	tt := NewTail()
	tt.Log = testutil.Logger{}
	tt.FromBeginning = true
	tt.Files = []string{tmpfile.Name()}
	tt.SetParserFunc(parsers.NewInfluxParser)

	err = tt.Init()
	require.NoError(t, err)

	acc := testutil.Accumulator{}
	require.NoError(t, tt.Start(&acc))
	defer tt.Stop()

	buf := &bytes.Buffer{}
	log.SetOutput(buf)

	require.NoError(t, acc.GatherError(tt.Gather))

	_, err = tmpfile.WriteString("cpu mytag= foo usage_idle= 100\n")
	require.NoError(t, err)

	time.Sleep(500 * time.Millisecond)
	assert.Contains(t, buf.String(), "Malformed log line")
}

func TestTailDosLineendings(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()
	_, err = tmpfile.WriteString("cpu usage_idle=100\r\ncpu2 usage_idle=200\r\n")
	require.NoError(t, err)

	tt := NewTail()
	tt.Log = testutil.Logger{}
	tt.FromBeginning = true
	tt.Files = []string{tmpfile.Name()}
	tt.SetParserFunc(parsers.NewInfluxParser)

	err = tt.Init()
	require.NoError(t, err)

	acc := testutil.Accumulator{}
	require.NoError(t, tt.Start(&acc))
	defer tt.Stop()
	require.NoError(t, acc.GatherError(tt.Gather))

	acc.Wait(2)
	acc.AssertContainsFields(t, "cpu",
		map[string]interface{}{
			"usage_idle": float64(100),
		})
	acc.AssertContainsFields(t, "cpu2",
		map[string]interface{}{
			"usage_idle": float64(200),
		})
}

func TestGrokParseLogFilesWithMultiline(t *testing.T) {
	thisdir := getCurrentDir()
	//we make sure the timeout won't kick in
	duration, _ := time.ParseDuration("100s")

	tt := NewTail()
	tt.Log = testutil.Logger{}
	tt.FromBeginning = true
	tt.Files = []string{thisdir + "testdata/test_multiline.log"}
	tt.MultilineConfig = MultilineConfig{
		Pattern:        `^[^\[]`,
		MatchWhichLine: Previous,
		Negate:         false,
		Timeout:        &internal.Duration{Duration: duration},
	}
	tt.SetParserFunc(createGrokParser)

	err := tt.Init()
	require.NoError(t, err)

	acc := testutil.Accumulator{}
	assert.NoError(t, tt.Start(&acc))
	defer tt.Stop()

	acc.Wait(3)

	expectedPath := thisdir + "testdata/test_multiline.log"
	acc.AssertContainsTaggedFields(t, "tail_grok",
		map[string]interface{}{
			"message": "HelloExample: This is debug",
		},
		map[string]string{
			"path":     expectedPath,
			"loglevel": "DEBUG",
		})
	acc.AssertContainsTaggedFields(t, "tail_grok",
		map[string]interface{}{
			"message": "HelloExample: This is info",
		},
		map[string]string{
			"path":     expectedPath,
			"loglevel": "INFO",
		})
	acc.AssertContainsTaggedFields(t, "tail_grok",
		map[string]interface{}{
			"message": "HelloExample: Sorry, something wrong! java.lang.ArithmeticException: / by zero\tat com.foo.HelloExample2.divide(HelloExample2.java:24)\tat com.foo.HelloExample2.main(HelloExample2.java:14)",
		},
		map[string]string{
			"path":     expectedPath,
			"loglevel": "ERROR",
		})

	assert.Equal(t, uint64(3), acc.NMetrics())
}

func TestGrokParseLogFilesWithMultilineTimeout(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	// set tight timeout for tests
	duration := 1 * time.Second

	tt := NewTail()
	tt.Log = testutil.Logger{}
	tt.FromBeginning = true
	tt.Files = []string{tmpfile.Name()}
	tt.MultilineConfig = MultilineConfig{
		Pattern:        `^[^\[]`,
		MatchWhichLine: Previous,
		Negate:         false,
		Timeout:        &internal.Duration{Duration: duration},
	}
	tt.SetParserFunc(createGrokParser)

	err = tt.Init()
	require.NoError(t, err)

	acc := testutil.Accumulator{}
	assert.NoError(t, tt.Start(&acc))
	time.Sleep(5 * time.Second) // will force timeout
	_, err = tmpfile.WriteString("[04/Jun/2016:12:41:48 +0100] INFO HelloExample: This is info\r\n")
	require.NoError(t, err)
	require.NoError(t, tmpfile.Sync())
	acc.Wait(1)
	time.Sleep(5 * time.Second) // will force timeout
	_, err = tmpfile.WriteString("[04/Jun/2016:12:41:48 +0100] WARN HelloExample: This is warn\r\n")
	require.NoError(t, err)
	require.NoError(t, tmpfile.Sync())
	acc.Wait(2)
	tt.Stop()
	assert.Equal(t, uint64(2), acc.NMetrics())
	expectedPath := tmpfile.Name()
	acc.AssertContainsTaggedFields(t, "tail_grok",
		map[string]interface{}{
			"message": "HelloExample: This is info",
		},
		map[string]string{
			"path":     expectedPath,
			"loglevel": "INFO",
		})
	acc.AssertContainsTaggedFields(t, "tail_grok",
		map[string]interface{}{
			"message": "HelloExample: This is warn",
		},
		map[string]string{
			"path":     expectedPath,
			"loglevel": "WARN",
		})
}

func TestGrokParseLogFilesWithMultilineTailerCloseFlushesMultilineBuffer(t *testing.T) {
	thisdir := getCurrentDir()
	//we make sure the timeout won't kick in
	duration, _ := time.ParseDuration("100s")

	tt := NewTail()
	tt.Log = testutil.Logger{}
	tt.FromBeginning = true
	tt.Files = []string{thisdir + "testdata/test_multiline.log"}
	tt.MultilineConfig = MultilineConfig{
		Pattern:        `^[^\[]`,
		MatchWhichLine: Previous,
		Negate:         false,
		Timeout:        &internal.Duration{Duration: duration},
	}
	tt.SetParserFunc(createGrokParser)

	err := tt.Init()
	require.NoError(t, err)

	acc := testutil.Accumulator{}
	assert.NoError(t, tt.Start(&acc))
	acc.Wait(3)
	assert.Equal(t, uint64(3), acc.NMetrics())
	// Close tailer, so multiline buffer is flushed
	tt.Stop()
	acc.Wait(4)

	expectedPath := thisdir + "testdata/test_multiline.log"
	acc.AssertContainsTaggedFields(t, "tail_grok",
		map[string]interface{}{
			"message": "HelloExample: This is warn",
		},
		map[string]string{
			"path":     expectedPath,
			"loglevel": "WARN",
		})
}

func createGrokParser() (parsers.Parser, error) {
	grokConfig := &parsers.Config{
		MetricName:             "tail_grok",
		GrokPatterns:           []string{"%{TEST_LOG_MULTILINE}"},
		GrokCustomPatternFiles: []string{getCurrentDir() + "testdata/test-patterns"},
		DataFormat:             "grok",
	}
	parser, err := parsers.NewParser(grokConfig)
	return parser, err
}

// The csv parser should only parse the header line once per file.
func TestCSVHeadersParsedOnce(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	defer func() {
		tmpfile.Close()
		os.Remove(tmpfile.Name())
	}()

	_, err = tmpfile.WriteString(`
measurement,time_idle
cpu,42
cpu,42
`)
	require.NoError(t, err)

	plugin := NewTail()
	plugin.Log = testutil.Logger{}
	plugin.FromBeginning = true
	plugin.Files = []string{tmpfile.Name()}
	plugin.SetParserFunc(func() (parsers.Parser, error) {
		return &csv.Parser{
			MeasurementColumn: "measurement",
			HeaderRowCount:    1,
			TimeFunc:          func() time.Time { return time.Unix(0, 0) },
		}, nil
	})

	err = plugin.Init()
	require.NoError(t, err)

	acc := testutil.Accumulator{}
	err = plugin.Start(&acc)
	require.NoError(t, err)
	defer plugin.Stop()
	err = plugin.Gather(&acc)
	require.NoError(t, err)
	acc.Wait(2)
	plugin.Stop()

	expected := []telegraf.Metric{
		testutil.MustMetric("cpu",
			map[string]string{
				"path": tmpfile.Name(),
			},
			map[string]interface{}{
				"time_idle":   42,
				"measurement": "cpu",
			},
			time.Unix(0, 0)),
		testutil.MustMetric("cpu",
			map[string]string{
				"path": tmpfile.Name(),
			},
			map[string]interface{}{
				"time_idle":   42,
				"measurement": "cpu",
			},
			time.Unix(0, 0)),
	}
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics())
}

// Ensure that the first line can produce multiple metrics (#6138)
func TestMultipleMetricsOnFirstLine(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	defer func() {
		tmpfile.Close()
		os.Remove(tmpfile.Name())
	}()

	_, err = tmpfile.WriteString(`
[{"time_idle": 42}, {"time_idle": 42}]
`)
	require.NoError(t, err)

	plugin := NewTail()
	plugin.Log = testutil.Logger{}
	plugin.FromBeginning = true
	plugin.Files = []string{tmpfile.Name()}
	plugin.SetParserFunc(func() (parsers.Parser, error) {
		return json.New(
			&json.Config{
				MetricName: "cpu",
			})
	})

	err = plugin.Init()
	require.NoError(t, err)

	acc := testutil.Accumulator{}
	err = plugin.Start(&acc)
	require.NoError(t, err)
	defer plugin.Stop()
	err = plugin.Gather(&acc)
	require.NoError(t, err)
	acc.Wait(2)
	plugin.Stop()

	expected := []telegraf.Metric{
		testutil.MustMetric("cpu",
			map[string]string{
				"path": tmpfile.Name(),
			},
			map[string]interface{}{
				"time_idle": 42.0,
			},
			time.Unix(0, 0)),
		testutil.MustMetric("cpu",
			map[string]string{
				"path": tmpfile.Name(),
			},
			map[string]interface{}{
				"time_idle": 42.0,
			},
			time.Unix(0, 0)),
	}
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(),
		testutil.IgnoreTime())
}

func getCurrentDir() string {
	_, filename, _, _ := runtime.Caller(1)
	return strings.Replace(filename, "tail_test.go", "", 1)
}
