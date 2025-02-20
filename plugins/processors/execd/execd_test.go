package execd

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	_ "github.com/influxdata/telegraf/plugins/parsers/all"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/plugins/processors"
	_ "github.com/influxdata/telegraf/plugins/serializers/all"
	serializers_influx "github.com/influxdata/telegraf/plugins/serializers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestExternalProcessorWorks(t *testing.T) {
	// Determine name of the test executable for mocking an external program
	exe, err := os.Executable()
	require.NoError(t, err)

	// Setup the plugin
	plugin := &Execd{
		Command: []string{
			exe,
			"-case", "multiply",
			"-field", "count",
		},
		Environment:  []string{"PLUGINS_PROCESSORS_EXECD_MODE=application"},
		RestartDelay: config.Duration(5 * time.Second),
		Log:          testutil.Logger{},
	}

	// Setup the parser and serializer in the processor
	parser := &influx.Parser{}
	require.NoError(t, parser.Init())
	plugin.SetParser(parser)

	serializer := &serializers_influx.Serializer{}
	require.NoError(t, serializer.Init())
	plugin.SetSerializer(serializer)

	// Setup the input and expected output metrucs
	now := time.Now()
	var input []telegraf.Metric
	var expected []telegraf.Metric
	for i := 0; i < 10; i++ {
		m := metric.New(
			"test",
			map[string]string{"city": "Toronto"},
			map[string]interface{}{"population": 6000000, "count": 1},
			now.Add(time.Duration(i)),
		)
		input = append(input, m)

		e := m.Copy()
		e.AddField("count", 2)
		expected = append(expected, e)
	}

	// Perform the test and check the result
	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()
	for _, m := range input {
		require.NoError(t, plugin.Add(m, &acc))
	}

	require.Eventually(t, func() bool {
		return acc.NMetrics() >= uint64(len(expected))
	}, 3*time.Second, 100*time.Millisecond)

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics())
}

func TestParseLinesWithNewLines(t *testing.T) {
	// Determine name of the test executable for mocking an external program
	exe, err := os.Executable()
	require.NoError(t, err)

	// Setup the plugin
	plugin := &Execd{
		Command: []string{
			exe,
			"-case", "multiply",
			"-field", "count",
		},
		Environment:  []string{"PLUGINS_PROCESSORS_EXECD_MODE=application"},
		RestartDelay: config.Duration(5 * time.Second),
		Log:          testutil.Logger{},
	}

	// Setup the parser and serializer in the processor
	parser := &influx.Parser{}
	require.NoError(t, parser.Init())
	plugin.SetParser(parser)

	serializer := &serializers_influx.Serializer{}
	require.NoError(t, serializer.Init())
	plugin.SetSerializer(serializer)

	// Setup the input and expected output metrucs
	now := time.Now()
	input := metric.New(
		"test",
		map[string]string{
			"author": "Mr. Gopher",
		},
		map[string]interface{}{
			"phrase": "Gophers are amazing creatures.\nAbsolutely amazing.",
			"count":  3,
		},
		now,
	)
	expected := []telegraf.Metric{
		metric.New(
			"test",
			map[string]string{"author": "Mr. Gopher"},
			map[string]interface{}{
				"phrase": "Gophers are amazing creatures.\nAbsolutely amazing.",
				"count":  6,
			},
			now,
		),
	}

	// Perform the test and check the result
	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()
	require.NoError(t, plugin.Add(input, &acc))

	require.Eventually(t, func() bool {
		return acc.NMetrics() >= uint64(len(expected))
	}, 3*time.Second, 100*time.Millisecond)

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics())
}

func TestLongLinesForLineProtocol(t *testing.T) {
	// Determine name of the test executable for mocking an external program
	exe, err := os.Executable()
	require.NoError(t, err)

	// Setup the plugin
	plugin := &Execd{
		Command: []string{
			exe,
			"-case", "long",
			"-field", "long",
		},
		Environment:  []string{"PLUGINS_PROCESSORS_EXECD_MODE=application"},
		RestartDelay: config.Duration(5 * time.Second),
		Log:          testutil.Logger{},
	}

	// Setup the parser and serializer in the processor
	parser := &influx.Parser{}
	require.NoError(t, parser.Init())
	plugin.SetParser(parser)

	serializer := &serializers_influx.Serializer{}
	require.NoError(t, serializer.Init())
	plugin.SetSerializer(serializer)

	// Setup the input and expected output metrucs
	now := time.Now()
	input := metric.New(
		"test",
		map[string]string{"author": "Mr. Gopher"},
		map[string]interface{}{"count": 3},
		now,
	)
	expected := []telegraf.Metric{
		metric.New(
			"test",
			map[string]string{"author": "Mr. Gopher"},
			map[string]interface{}{
				"long":  strings.Repeat("foobar", 280_000/6),
				"count": 3,
			},
			now,
		),
	}

	// Perform the test and check the result
	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()
	require.NoError(t, plugin.Add(input, &acc))

	require.Eventually(t, func() bool {
		return acc.NMetrics() >= uint64(len(expected))
	}, 3*time.Second, 100*time.Millisecond)

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics())
}

func TestCases(t *testing.T) {
	// Get all directories in testcases
	folders, err := os.ReadDir("testcases")
	require.NoError(t, err)

	// Make sure tests contains data
	require.NotEmpty(t, folders)

	// Set up for file inputs
	processors.AddStreaming("execd", func() telegraf.StreamingProcessor {
		return &Execd{RestartDelay: config.Duration(10 * time.Second)}
	})

	for _, f := range folders {
		// Only handle folders
		if !f.IsDir() {
			continue
		}

		fname := f.Name()
		t.Run(fname, func(t *testing.T) {
			testdataPath := filepath.Join("testcases", fname)
			configFilename := filepath.Join(testdataPath, "telegraf.conf")
			inputFilename := filepath.Join(testdataPath, "input.influx")
			expectedFilename := filepath.Join(testdataPath, "expected.out")

			// Get parser to parse input and expected output
			parser := &influx.Parser{}
			require.NoError(t, parser.Init())

			input, err := testutil.ParseMetricsFromFile(inputFilename, parser)
			require.NoError(t, err)

			expected, err := testutil.ParseMetricsFromFile(expectedFilename, parser)
			require.NoError(t, err)

			// Configure the plugin
			cfg := config.NewConfig()
			require.NoError(t, cfg.LoadConfig(configFilename))
			require.Len(t, cfg.Processors, 1, "wrong number of outputs")
			plugin := cfg.Processors[0].Processor

			// Process the metrics
			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			for _, m := range input {
				require.NoError(t, plugin.Add(m, &acc))
			}
			plugin.Stop()

			require.Eventually(t, func() bool {
				return acc.NMetrics() >= uint64(len(expected))
			}, time.Second, 100*time.Millisecond)

			// Check the expectations
			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, expected, actual)
		})
	}
}

func TestTracking(t *testing.T) {
	now := time.Now()

	// Setup the raw  input and expected output data
	inputRaw := []telegraf.Metric{
		metric.New(
			"test",
			map[string]string{
				"city": "Toronto",
			},
			map[string]interface{}{
				"population": 6000000,
				"count":      1,
			},
			now,
		),
		metric.New(
			"test",
			map[string]string{
				"city": "Tokio",
			},
			map[string]interface{}{
				"population": 14000000,
				"count":      8,
			},
			now,
		),
	}

	expected := []telegraf.Metric{
		metric.New(
			"test",
			map[string]string{
				"city": "Toronto",
			},
			map[string]interface{}{
				"population": 6000000,
				"count":      2,
			},
			now,
		),
		metric.New(
			"test",
			map[string]string{
				"city": "Tokio",
			},
			map[string]interface{}{
				"population": 14000000,
				"count":      16,
			},
			now,
		),
	}

	// Create a testing notifier
	var mu sync.Mutex
	delivered := make([]telegraf.DeliveryInfo, 0, len(inputRaw))
	notify := func(di telegraf.DeliveryInfo) {
		mu.Lock()
		defer mu.Unlock()
		delivered = append(delivered, di)
	}

	// Convert raw input to tracking metrics
	input := make([]telegraf.Metric, 0, len(inputRaw))
	for _, m := range inputRaw {
		tm, _ := metric.WithTracking(m, notify)
		input = append(input, tm)
	}

	// Setup the plugin
	exe, err := os.Executable()
	require.NoError(t, err)

	plugin := &Execd{
		Command: []string{
			exe,
			"-case", "multiply",
			"-field", "count",
		},
		Environment:  []string{"PLUGINS_PROCESSORS_EXECD_MODE=application"},
		RestartDelay: config.Duration(5 * time.Second),
		Log:          testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	parser := &influx.Parser{}
	require.NoError(t, parser.Init())
	plugin.SetParser(parser)

	serializer := &serializers_influx.Serializer{}
	require.NoError(t, serializer.Init())
	plugin.SetSerializer(serializer)

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	// Process expected metrics and compare with resulting metrics
	for _, in := range input {
		require.NoError(t, plugin.Add(in, &acc))
	}
	require.Eventually(t, func() bool {
		return int(acc.NMetrics()) >= len(expected)
	}, 3*time.Second, 100*time.Millisecond)

	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual)

	// Simulate output acknowledging delivery
	for _, m := range actual {
		m.Accept()
	}

	// Check delivery
	require.Eventuallyf(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(input) == len(delivered)
	}, time.Second, 100*time.Millisecond, "%d delivered but %d expected", len(delivered), len(expected))
}

func TestMain(m *testing.M) {
	var testcase, field string
	flag.StringVar(&testcase, "case", "", "test-case to mock [multiply, long]")
	flag.StringVar(&field, "field", "count", "name of the field to multiply")
	flag.Parse()

	if os.Getenv("PLUGINS_PROCESSORS_EXECD_MODE") != "application" || testcase == "" {
		os.Exit(m.Run())
	}

	switch testcase {
	case "multiply":
		os.Exit(runTestCaseMultiply(field))
	case "long":
		os.Exit(runTestCaseLong(field))
	}
	os.Exit(5)
}

func runTestCaseMultiply(field string) int {
	parser := influx.NewStreamParser(os.Stdin)
	serializer := &serializers_influx.Serializer{}
	if err := serializer.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "initialization ERR %v\n", err)
		return 1
	}

	for {
		m, err := parser.Next()
		if err != nil {
			if errors.Is(err, influx.EOF) {
				return 0
			}
			var parseErr *influx.ParseError
			if errors.As(err, &parseErr) {
				fmt.Fprintf(os.Stderr, "parse ERR %v\n", parseErr)
				return 1
			}
			fmt.Fprintf(os.Stderr, "ERR %v\n", err)
			return 1
		}

		c, found := m.GetField(field)
		if !found {
			fmt.Fprintf(os.Stderr, "metric has no field %q\n", field)
			return 1
		}
		switch t := c.(type) {
		case float64:
			m.AddField(field, t*2)
		case int64:
			m.AddField(field, t*2)
		default:
			fmt.Fprintf(os.Stderr, "%s has an unknown type, it's a %T\n", field, c)
			return 1
		}
		b, err := serializer.Serialize(m)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERR %v\n", err)
			return 1
		}
		fmt.Fprint(os.Stdout, string(b))
	}
}

func runTestCaseLong(field string) int {
	parser := influx.NewStreamParser(os.Stdin)
	serializer := &serializers_influx.Serializer{}
	if err := serializer.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "initialization ERR %v\n", err)
		return 1
	}

	// Setup a field with a lot of characters to exceed the scanner limit
	long := strings.Repeat("foobar", 280_000/6)

	for {
		m, err := parser.Next()
		if err != nil {
			if errors.Is(err, influx.EOF) {
				return 0
			}
			var parseErr *influx.ParseError
			if errors.As(err, &parseErr) {
				fmt.Fprintf(os.Stderr, "parse ERR %v\n", parseErr)
				return 1
			}
			fmt.Fprintf(os.Stderr, "ERR %v\n", err)
			return 1
		}

		m.AddField(field, long)

		b, err := serializer.Serialize(m)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERR %v\n", err)
			return 1
		}
		fmt.Fprint(os.Stdout, string(b))
	}
}
