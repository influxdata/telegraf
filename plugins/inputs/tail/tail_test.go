package tail

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/parsers/csv"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/plugins/parsers/json"
	"github.com/influxdata/telegraf/testutil"
)

var (
	testdataDir = getTestdataDir()
)

func NewTestTail() *Tail {
	offsetsMutex.Lock()
	offsetsCopy := make(map[string]int64, len(offsets))
	for k, v := range offsets {
		offsetsCopy[k] = v
	}
	offsetsMutex.Unlock()
	watchMethod := defaultWatchMethod

	if runtime.GOOS == "windows" {
		watchMethod = "poll"
	}

	return &Tail{
		FromBeginning:       false,
		MaxUndeliveredLines: 1000,
		offsets:             offsetsCopy,
		WatchMethod:         watchMethod,
		PathTag:             "path",
	}
}

func TestTailBadLine(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.WriteString("cpu mytag= foo usage_idle= 100\n")
	require.NoError(t, err)

	// Write good metric so we can detect when processing is complete
	_, err = tmpfile.WriteString("cpu usage_idle=100\n")
	require.NoError(t, err)

	require.NoError(t, tmpfile.Close())

	buf := &bytes.Buffer{}
	log.SetOutput(buf)

	tt := NewTestTail()
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
	require.Contains(t, buf.String(), "Malformed log line")
}

func TestColoredLine(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())
	_, err = tmpfile.WriteString("cpu usage_idle=\033[4A\033[4A100\ncpu2 usage_idle=200\n")
	require.NoError(t, err)
	require.NoError(t, tmpfile.Close())

	tt := NewTestTail()
	tt.Log = testutil.Logger{}
	tt.FromBeginning = true
	tt.Filters = []string{"ansi_color"}
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

func TestTailDosLineEndings(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())
	_, err = tmpfile.WriteString("cpu usage_idle=100\r\ncpu2 usage_idle=200\r\n")
	require.NoError(t, err)
	require.NoError(t, tmpfile.Close())

	tt := NewTestTail()
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
	//we make sure the timeout won't kick in
	d, _ := time.ParseDuration("100s")
	duration := config.Duration(d)
	tt := NewTail()
	tt.Log = testutil.Logger{}
	tt.FromBeginning = true
	tt.Files = []string{filepath.Join(testdataDir, "test_multiline.log")}
	tt.MultilineConfig = MultilineConfig{
		Pattern:        `^[^\[]`,
		MatchWhichLine: Previous,
		InvertMatch:    false,
		Timeout:        &duration,
	}
	tt.SetParserFunc(createGrokParser)

	err := tt.Init()
	require.NoError(t, err)

	acc := testutil.Accumulator{}
	require.NoError(t, tt.Start(&acc))
	defer tt.Stop()

	acc.Wait(3)

	expectedPath := filepath.Join(testdataDir, "test_multiline.log")
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

	require.Equal(t, uint64(3), acc.NMetrics())
}

func TestGrokParseLogFilesWithMultilineTimeout(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	// This seems necessary in order to get the test to read the following lines.
	_, err = tmpfile.WriteString("[04/Jun/2016:12:41:48 +0100] INFO HelloExample: This is fluff\r\n")
	require.NoError(t, err)
	require.NoError(t, tmpfile.Sync())

	// set tight timeout for tests
	d := 10 * time.Millisecond
	duration := config.Duration(d)
	tt := NewTail()

	tt.Log = testutil.Logger{}
	tt.FromBeginning = true
	tt.Files = []string{tmpfile.Name()}
	tt.MultilineConfig = MultilineConfig{
		Pattern:        `^[^\[]`,
		MatchWhichLine: Previous,
		InvertMatch:    false,
		Timeout:        &duration,
	}
	tt.SetParserFunc(createGrokParser)

	err = tt.Init()
	require.NoError(t, err)

	acc := testutil.Accumulator{}
	require.NoError(t, tt.Start(&acc))
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
	require.Equal(t, uint64(3), acc.NMetrics())
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
	//we make sure the timeout won't kick in
	duration := config.Duration(100 * time.Second)

	tt := NewTestTail()
	tt.Log = testutil.Logger{}
	tt.FromBeginning = true
	tt.Files = []string{filepath.Join(testdataDir, "test_multiline.log")}
	tt.MultilineConfig = MultilineConfig{
		Pattern:        `^[^\[]`,
		MatchWhichLine: Previous,
		InvertMatch:    false,
		Timeout:        &duration,
	}
	tt.SetParserFunc(createGrokParser)

	err := tt.Init()
	require.NoError(t, err)

	acc := testutil.Accumulator{}
	require.NoError(t, tt.Start(&acc))
	acc.Wait(3)
	require.Equal(t, uint64(3), acc.NMetrics())
	// Close tailer, so multiline buffer is flushed
	tt.Stop()
	acc.Wait(4)

	expectedPath := filepath.Join(testdataDir, "test_multiline.log")
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
		GrokCustomPatternFiles: []string{filepath.Join(testdataDir, "test-patterns")},
		DataFormat:             "grok",
	}
	parser, err := parsers.NewParser(grokConfig)
	return parser, err
}

// The csv parser should only parse the header line once per file.
func TestCSVHeadersParsedOnce(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.WriteString(`
measurement,time_idle
cpu,42
cpu,42
`)
	require.NoError(t, err)
	require.NoError(t, tmpfile.Close())

	plugin := NewTestTail()
	plugin.Log = testutil.Logger{}
	plugin.FromBeginning = true
	plugin.Files = []string{tmpfile.Name()}
	plugin.SetParserFunc(func() (parsers.Parser, error) {
		parser := csv.Parser{
			MeasurementColumn: "measurement",
			HeaderRowCount:    1,
			TimeFunc:          func() time.Time { return time.Unix(0, 0) },
		}
		err := parser.Init()
		return &parser, err
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

func TestCSVMultiHeaderWithSkipRowANDColumn(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.WriteString(`garbage nonsense
skip,measurement,value
row,1,2
skip1,cpu,42
skip2,mem,100
`)
	require.NoError(t, err)
	require.NoError(t, tmpfile.Close())

	plugin := NewTestTail()
	plugin.Log = testutil.Logger{}
	plugin.FromBeginning = true
	plugin.Files = []string{tmpfile.Name()}
	plugin.SetParserFunc(func() (parsers.Parser, error) {
		parser := csv.Parser{
			MeasurementColumn: "measurement1",
			HeaderRowCount:    2,
			SkipRows:          1,
			SkipColumns:       1,
			TimeFunc:          func() time.Time { return time.Unix(0, 0) },
		}
		err := parser.Init()
		return &parser, err
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
				"value2": 42,
			},
			time.Unix(0, 0)),
		testutil.MustMetric("mem",
			map[string]string{
				"path": tmpfile.Name(),
			},
			map[string]interface{}{
				"value2": 100,
			},
			time.Unix(0, 0)),
	}
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics())
}

// Ensure that the first line can produce multiple metrics (#6138)
func TestMultipleMetricsOnFirstLine(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.WriteString(`
[{"time_idle": 42}, {"time_idle": 42}]
`)
	require.NoError(t, err)
	require.NoError(t, tmpfile.Close())

	plugin := NewTestTail()
	plugin.Log = testutil.Logger{}
	plugin.FromBeginning = true
	plugin.Files = []string{tmpfile.Name()}
	plugin.PathTag = "customPathTagMyFile"
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
				"customPathTagMyFile": tmpfile.Name(),
			},
			map[string]interface{}{
				"time_idle": 42.0,
			},
			time.Unix(0, 0)),
		testutil.MustMetric("cpu",
			map[string]string{
				"customPathTagMyFile": tmpfile.Name(),
			},
			map[string]interface{}{
				"time_idle": 42.0,
			},
			time.Unix(0, 0)),
	}
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(),
		testutil.IgnoreTime())
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

	watchMethod := defaultWatchMethod
	if runtime.GOOS == "windows" {
		watchMethod = "poll"
	}

	tests := []struct {
		name              string
		testfiles         string
		fromBeginning     bool
		characterEncoding string
		offset            int64
		expected          []telegraf.Metric
	}{
		{
			name:              "utf-8",
			testfiles:         "cpu-utf-8.influx",
			fromBeginning:     true,
			characterEncoding: "utf-8",
			expected:          full,
		},
		{
			name:              "utf-8 seek",
			testfiles:         "cpu-utf-8.influx",
			characterEncoding: "utf-8",
			offset:            0x33,
			expected:          full[1:],
		},
		{
			name:              "utf-16le",
			testfiles:         "cpu-utf-16le.influx",
			fromBeginning:     true,
			characterEncoding: "utf-16le",
			expected:          full,
		},
		{
			name:              "utf-16le seek",
			testfiles:         "cpu-utf-16le.influx",
			characterEncoding: "utf-16le",
			offset:            0x68,
			expected:          full[1:],
		},
		{
			name:              "utf-16be",
			testfiles:         "cpu-utf-16be.influx",
			fromBeginning:     true,
			characterEncoding: "utf-16be",
			expected:          full,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &Tail{
				Files:               []string{filepath.Join(testdataDir, tt.testfiles)},
				FromBeginning:       tt.fromBeginning,
				MaxUndeliveredLines: 1000,
				Log:                 testutil.Logger{},
				CharacterEncoding:   tt.characterEncoding,
				WatchMethod:         watchMethod,
			}

			plugin.SetParserFunc(func() (parsers.Parser, error) {
				handler := influx.NewMetricHandler()
				return influx.NewParser(handler), nil
			})

			if tt.offset != 0 {
				plugin.offsets = map[string]int64{
					plugin.Files[0]: tt.offset,
				}
			}

			err := plugin.Init()
			require.NoError(t, err)

			var acc testutil.Accumulator
			err = plugin.Start(&acc)
			require.NoError(t, err)
			acc.Wait(len(tt.expected))
			plugin.Stop()

			actual := acc.GetTelegrafMetrics()
			for _, m := range actual {
				m.RemoveTag("path")
			}

			testutil.RequireMetricsEqual(t, tt.expected, actual, testutil.IgnoreTime())
		})
	}
}

func TestTailEOF(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())
	_, err = tmpfile.WriteString("cpu usage_idle=100\r\n")
	require.NoError(t, err)
	err = tmpfile.Sync()
	require.NoError(t, err)

	tt := NewTestTail()
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

func getTestdataDir() string {
	dir, err := os.Getwd()
	if err != nil {
		// if we cannot even establish the test directory, further progress is meaningless
		panic(err)
	}

	return filepath.Join(dir, "testdata")
}
