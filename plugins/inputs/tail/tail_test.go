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
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/plugins/parsers/json"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTailBadLine(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.WriteString("cpu mytag= foo usage_idle= 100\n")
	require.NoError(t, err)

	// Write good metric so we can detect when processing is complete
	_, err = tmpfile.WriteString("cpu usage_idle=100\n")
	require.NoError(t, err)

	tmpfile.Close()

	buf := &bytes.Buffer{}
	log.SetOutput(buf)

	tt := NewTail()
	tt.Log = testutil.Logger{}
	tt.FromBeginning = true
	tt.Files = []string{tmpfile.Name()}
	tt.SetParserFunc(parsers.NewInfluxParser)

	err = tt.Init()
	require.NoError(t, err)

	acc := testutil.Accumulator{}
	require.NoError(t, tt.Start(&acc))

	require.NoError(t, acc.GatherError(tt.Gather))

	acc.Wait(1)

	tt.Stop()
	assert.Contains(t, buf.String(), "Malformed log line")
}

func TestTailDosLineendings(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())
	_, err = tmpfile.WriteString("cpu usage_idle=100\r\ncpu2 usage_idle=200\r\n")
	require.NoError(t, err)
	tmpfile.Close()

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
		InvertMatch:    false,
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

	// This seems neccessary in order to get the test to read the following lines.
	_, err = tmpfile.WriteString("[04/Jun/2016:12:41:48 +0100] INFO HelloExample: This is fluff\r\n")
	require.NoError(t, err)
	require.NoError(t, tmpfile.Sync())

	// set tight timeout for tests
	duration := 10 * time.Millisecond

	tt := NewTail()
	tt.Log = testutil.Logger{}
	tt.FromBeginning = true
	tt.Files = []string{tmpfile.Name()}
	tt.MultilineConfig = MultilineConfig{
		Pattern:        `^[^\[]`,
		MatchWhichLine: Previous,
		InvertMatch:    false,
		Timeout:        &internal.Duration{Duration: duration},
	}
	tt.SetParserFunc(createGrokParser)

	err = tt.Init()
	require.NoError(t, err)

	acc := testutil.Accumulator{}
	assert.NoError(t, tt.Start(&acc))
	time.Sleep(11 * time.Millisecond) // will force timeout
	_, err = tmpfile.WriteString("[04/Jun/2016:12:41:48 +0100] INFO HelloExample: This is info\r\n")
	require.NoError(t, err)
	require.NoError(t, tmpfile.Sync())
	acc.Wait(2)
	time.Sleep(11 * time.Millisecond) // will force timeout
	_, err = tmpfile.WriteString("[04/Jun/2016:12:41:48 +0100] WARN HelloExample: This is warn\r\n")
	require.NoError(t, err)
	require.NoError(t, tmpfile.Sync())
	acc.Wait(3)
	tt.Stop()
	assert.Equal(t, uint64(3), acc.NMetrics())
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
	duration := 100 * time.Second

	tt := NewTail()
	tt.Log = testutil.Logger{}
	tt.FromBeginning = true
	tt.Files = []string{thisdir + "testdata/test_multiline.log"}
	tt.MultilineConfig = MultilineConfig{
		Pattern:        `^[^\[]`,
		MatchWhichLine: Previous,
		InvertMatch:    false,
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
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.WriteString(`
measurement,time_idle
cpu,42
cpu,42
`)
	require.NoError(t, err)
	tmpfile.Close()

	plugin := NewTail()
	plugin.Log = testutil.Logger{}
	plugin.FromBeginning = true
	plugin.Files = []string{tmpfile.Name()}
	plugin.SetParserFunc(func() (parsers.Parser, error) {
		return csv.NewParser(&csv.Config{
			MeasurementColumn: "measurement",
			HeaderRowCount:    1,
			TimeFunc:          func() time.Time { return time.Unix(0, 0) },
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
				"time_idle": 42,
			},
			time.Unix(0, 0)),
		testutil.MustMetric("cpu",
			map[string]string{
				"path": tmpfile.Name(),
			},
			map[string]interface{}{
				"time_idle": 42,
			},
			time.Unix(0, 0)),
	}
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics())
}

// Ensure that the first line can produce multiple metrics (#6138)
func TestMultipleMetricsOnFirstLine(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.WriteString(`
[{"time_idle": 42}, {"time_idle": 42}]
`)
	require.NoError(t, err)
	tmpfile.Close()

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

func TestCharacterEncoding(t *testing.T) {
	full := []telegraf.Metric{
		testutil.MustMetric("cpu",
			map[string]string{
				"cpu": "cpu0",
			},
			map[string]interface{}{
				"usage_active": 11.9,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric("cpu",
			map[string]string{
				"cpu": "cpu1",
			},
			map[string]interface{}{
				"usage_active": 26.0,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric("cpu",
			map[string]string{
				"cpu": "cpu2",
			},
			map[string]interface{}{
				"usage_active": 14.0,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric("cpu",
			map[string]string{
				"cpu": "cpu3",
			},
			map[string]interface{}{
				"usage_active": 20.4,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric("cpu",
			map[string]string{
				"cpu": "cpu-total",
			},
			map[string]interface{}{
				"usage_active": 18.4,
			},
			time.Unix(0, 0),
		),
	}

	tests := []struct {
		name     string
		plugin   *Tail
		offset   int64
		expected []telegraf.Metric
	}{
		{
			name: "utf-8",
			plugin: &Tail{
				Files:               []string{"testdata/cpu-utf-8.influx"},
				FromBeginning:       true,
				MaxUndeliveredLines: 1000,
				Log:                 testutil.Logger{},
				CharacterEncoding:   "utf-8",
			},
			expected: full,
		},
		{
			name: "utf-8 seek",
			plugin: &Tail{
				Files:               []string{"testdata/cpu-utf-8.influx"},
				MaxUndeliveredLines: 1000,
				Log:                 testutil.Logger{},
				CharacterEncoding:   "utf-8",
			},
			offset:   0x33,
			expected: full[1:],
		},
		{
			name: "utf-16le",
			plugin: &Tail{
				Files:               []string{"testdata/cpu-utf-16le.influx"},
				FromBeginning:       true,
				MaxUndeliveredLines: 1000,
				Log:                 testutil.Logger{},
				CharacterEncoding:   "utf-16le",
			},
			expected: full,
		},
		{
			name: "utf-16le seek",
			plugin: &Tail{
				Files:               []string{"testdata/cpu-utf-16le.influx"},
				MaxUndeliveredLines: 1000,
				Log:                 testutil.Logger{},
				CharacterEncoding:   "utf-16le",
			},
			offset:   0x68,
			expected: full[1:],
		},
		{
			name: "utf-16be",
			plugin: &Tail{
				Files:               []string{"testdata/cpu-utf-16be.influx"},
				FromBeginning:       true,
				MaxUndeliveredLines: 1000,
				Log:                 testutil.Logger{},
				CharacterEncoding:   "utf-16be",
			},
			expected: full,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.plugin.SetParserFunc(func() (parsers.Parser, error) {
				handler := influx.NewMetricHandler()
				return influx.NewParser(handler), nil
			})

			if tt.offset != 0 {
				tt.plugin.offsets = map[string]int64{
					tt.plugin.Files[0]: tt.offset,
				}
			}

			err := tt.plugin.Init()
			require.NoError(t, err)

			var acc testutil.Accumulator
			err = tt.plugin.Start(&acc)
			require.NoError(t, err)
			acc.Wait(len(tt.expected))
			tt.plugin.Stop()

			actual := acc.GetTelegrafMetrics()
			for _, m := range actual {
				m.RemoveTag("path")
			}

			testutil.RequireMetricsEqual(t, tt.expected, actual, testutil.IgnoreTime())
		})
	}
}

func TestTailEOF(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())
	_, err = tmpfile.WriteString("cpu usage_idle=100\r\n")
	require.NoError(t, err)
	err = tmpfile.Sync()
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
	acc.Wait(1) // input hits eof

	_, err = tmpfile.WriteString("cpu2 usage_idle=200\r\n")
	require.NoError(t, err)
	err = tmpfile.Sync()
	require.NoError(t, err)

	acc.Wait(2)
	require.NoError(t, acc.GatherError(tt.Gather))
	acc.AssertContainsFields(t, "cpu",
		map[string]interface{}{
			"usage_idle": float64(100),
		})
	acc.AssertContainsFields(t, "cpu2",
		map[string]interface{}{
			"usage_idle": float64(200),
		})

	err = tmpfile.Close()
	require.NoError(t, err)
}
