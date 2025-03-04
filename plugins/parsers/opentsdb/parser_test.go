package opentsdb

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func TestParseLine(t *testing.T) {
	testTime := time.Now()
	testTimeSec := testTime.Round(time.Second)
	testTimeMilli := testTime.Round(time.Millisecond)
	strTimeSec := strconv.FormatInt(testTimeSec.Unix(), 10)
	strTimeMilli := strconv.FormatInt(testTimeMilli.UnixNano()/1000000, 10)

	var tests = []struct {
		name     string
		input    string
		expected telegraf.Metric
	}{
		{
			name:  "minimal case",
			input: "put sys.cpu.user " + strTimeSec + " 50",
			expected: testutil.MustMetric(
				"sys.cpu.user",
				map[string]string{},
				map[string]interface{}{
					"value": float64(50),
				},
				testTimeSec,
			),
		},
		{
			name:  "millisecond timestamp",
			input: "put sys.cpu.user " + strTimeMilli + " 50",
			expected: testutil.MustMetric(
				"sys.cpu.user",
				map[string]string{},
				map[string]interface{}{
					"value": float64(50),
				},
				testTimeMilli,
			),
		},
		{
			name:  "floating point value",
			input: "put sys.cpu.user " + strTimeSec + " 42.5",
			expected: testutil.MustMetric(
				"sys.cpu.user",
				map[string]string{},
				map[string]interface{}{
					"value": float64(42.5),
				},
				testTimeSec,
			),
		},
		{
			name:  "single tag",
			input: "put sys.cpu.user " + strTimeSec + " 42.5 host=webserver01",
			expected: testutil.MustMetric(
				"sys.cpu.user",
				map[string]string{
					"host": "webserver01",
				},
				map[string]interface{}{
					"value": float64(42.5),
				},
				testTimeSec,
			),
		},
		{
			name:  "double tags",
			input: "put sys.cpu.user " + strTimeSec + " 42.5 host=webserver01 cpu=7",
			expected: testutil.MustMetric(
				"sys.cpu.user",
				map[string]string{
					"host": "webserver01",
					"cpu":  "7",
				},
				map[string]interface{}{
					"value": float64(42.5),
				},
				testTimeSec,
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{Log: testutil.Logger{}}

			actual, err := p.ParseLine(tt.input)
			require.NoError(t, err)

			testutil.RequireMetricEqual(t, tt.expected, actual)
		})
	}
}

func TestParse(t *testing.T) {
	testTime := time.Now()
	testTimeSec := testTime.Round(time.Second)
	strTimeSec := strconv.FormatInt(testTimeSec.Unix(), 10)

	var tests = []struct {
		name     string
		input    []byte
		expected []telegraf.Metric
	}{
		{
			name:  "single line with no newline",
			input: []byte("put sys.cpu.user " + strTimeSec + " 42.5 host=webserver01 cpu=7"),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"sys.cpu.user",
					map[string]string{
						"host": "webserver01",
						"cpu":  "7",
					},
					map[string]interface{}{
						"value": float64(42.5),
					},
					testTimeSec,
				),
			},
		},
		{
			name:  "single line with LF",
			input: []byte("put sys.cpu.user " + strTimeSec + " 42.5 host=webserver01 cpu=7\n"),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"sys.cpu.user",
					map[string]string{
						"host": "webserver01",
						"cpu":  "7",
					},
					map[string]interface{}{
						"value": float64(42.5),
					},
					testTimeSec,
				),
			},
		},
		{
			name:  "single line with CR+LF",
			input: []byte("put sys.cpu.user " + strTimeSec + " 42.5 host=webserver01 cpu=7\r\n"),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"sys.cpu.user",
					map[string]string{
						"host": "webserver01",
						"cpu":  "7",
					},
					map[string]interface{}{
						"value": float64(42.5),
					},
					testTimeSec,
				),
			},
		},
		{
			name: "double lines",
			input: []byte("put sys.cpu.user " + strTimeSec + " 42.5 host=webserver01 cpu=7\r\n" +
				"put sys.cpu.user " + strTimeSec + " 53.5 host=webserver02 cpu=3\r\n"),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"sys.cpu.user",
					map[string]string{
						"host": "webserver01",
						"cpu":  "7",
					},
					map[string]interface{}{
						"value": float64(42.5),
					},
					testTimeSec,
				),
				testutil.MustMetric(
					"sys.cpu.user",
					map[string]string{
						"host": "webserver02",
						"cpu":  "3",
					},
					map[string]interface{}{
						"value": float64(53.5),
					},
					testTimeSec,
				),
			},
		},
		{
			name: "mixed valid/invalid input",
			input: []byte(
				"version\r\n" +
					"put sys.cpu.user " + strTimeSec + " 42.5 host=webserver01 cpu=7\r\n" +
					"put sys.cpu.user " + strTimeSec + " 53.5 host=webserver02 cpu=3\r\n",
			),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"sys.cpu.user",
					map[string]string{
						"host": "webserver01",
						"cpu":  "7",
					},
					map[string]interface{}{
						"value": float64(42.5),
					},
					testTimeSec,
				),
				testutil.MustMetric(
					"sys.cpu.user",
					map[string]string{
						"host": "webserver02",
						"cpu":  "3",
					},
					map[string]interface{}{
						"value": float64(53.5),
					},
					testTimeSec,
				),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{Log: testutil.Logger{}}

			actual, err := p.Parse(tt.input)
			require.NoError(t, err)

			testutil.RequireMetricsEqual(t, tt.expected, actual)
		})
	}
}

func TestParse_DefaultTags(t *testing.T) {
	testTime := time.Now()
	testTimeSec := testTime.Round(time.Second)
	strTimeSec := strconv.FormatInt(testTimeSec.Unix(), 10)

	var tests = []struct {
		name        string
		input       []byte
		defaultTags map[string]string
		expected    []telegraf.Metric
	}{
		{
			name:  "single default tag",
			input: []byte("put sys.cpu.user " + strTimeSec + " 42.5 host=webserver01 cpu=7"),
			defaultTags: map[string]string{
				"foo": "bar",
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"sys.cpu.user",
					map[string]string{
						"foo":  "bar",
						"host": "webserver01",
						"cpu":  "7",
					},
					map[string]interface{}{
						"value": float64(42.5),
					},
					testTimeSec,
				),
			},
		},
		{
			name:  "double default tags",
			input: []byte("put sys.cpu.user " + strTimeSec + " 42.5 host=webserver01 cpu=7"),
			defaultTags: map[string]string{
				"foo1": "bar1",
				"foo2": "bar2",
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"sys.cpu.user",
					map[string]string{
						"foo1": "bar1",
						"foo2": "bar2",
						"host": "webserver01",
						"cpu":  "7",
					},
					map[string]interface{}{
						"value": float64(42.5),
					},
					testTimeSec,
				),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{Log: testutil.Logger{}}
			p.SetDefaultTags(tt.defaultTags)

			actual, err := p.Parse(tt.input)
			require.NoError(t, err)

			testutil.RequireMetricsEqual(t, tt.expected, actual)
		})
	}
}

const benchmarkData = `put benchmark_a 1653643420 4 tags_host=myhost tags_platform=python tags_sdkver=3.11.4
put benchmark_b 1653643420 5 tags_host=myhost tags_platform=python tags_sdkver=3.11.5
`

func TestBenchmarkData(t *testing.T) {
	plugin := &Parser{}

	expected := []telegraf.Metric{
		metric.New(
			"benchmark_a",
			map[string]string{
				"tags_host":     "myhost",
				"tags_platform": "python",
				"tags_sdkver":   "3.11.4",
			},
			map[string]interface{}{
				"value": 4.0,
			},
			time.Unix(1653643420, 0),
		),
		metric.New(
			"benchmark_b",
			map[string]string{
				"tags_host":     "myhost",
				"tags_platform": "python",
				"tags_sdkver":   "3.11.5",
			},
			map[string]interface{}{
				"value": 5.0,
			},
			time.Unix(1653643420, 0),
		),
	}

	actual, err := plugin.Parse([]byte(benchmarkData))
	require.NoError(t, err)
	testutil.RequireMetricsEqual(t, expected, actual, testutil.SortMetrics())
}

func BenchmarkParsing(b *testing.B) {
	plugin := &Parser{}

	for n := 0; n < b.N; n++ {
		//nolint:errcheck // Benchmarking so skip the error check to avoid the unnecessary operations
		plugin.Parse([]byte(benchmarkData))
	}
}
