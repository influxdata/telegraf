package split

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/plugins/processors"
	"github.com/influxdata/telegraf/testutil"
)

func TestCases(t *testing.T) {
	folders, err := os.ReadDir("testcases")
	require.NoError(t, err)
	require.NotEmpty(t, folders)

	processors.Add("split", func() telegraf.Processor {
		return &Split{}
	})

	for _, f := range folders {
		// Only handle folders
		if !f.IsDir() {
			continue
		}

		fname := f.Name()
		testdataPath := filepath.Join("testcases", fname)
		configFilename := filepath.Join(testdataPath, "config.toml")
		inputFilename := filepath.Join(testdataPath, "input.influx")
		expectedFilename := filepath.Join(testdataPath, "expected.out")

		t.Run(fname, func(t *testing.T) {
			// Get parser to parse input and expected output
			parser := &influx.Parser{}
			require.NoError(t, parser.Init())

			input, err := testutil.ParseMetricsFromFile(inputFilename, parser)
			require.NoError(t, err)

			var expected []telegraf.Metric
			if _, err := os.Stat(expectedFilename); err == nil {
				var err error
				expected, err = testutil.ParseMetricsFromFile(expectedFilename, parser)
				require.NoError(t, err)
			}

			// Configure the plugin
			cfg := config.NewConfig()
			require.NoError(t, cfg.LoadConfig(configFilename))
			require.Len(t, cfg.Processors, 1, "wrong number of processors")

			proc := cfg.Processors[0].Processor.(processors.HasUnwrap)
			plugin := proc.Unwrap().(*Split)
			require.NoError(t, plugin.Init())

			// Process expected metrics and compare with resulting metrics
			actual := plugin.Apply(input...)

			testutil.RequireMetricsEqual(t, expected, actual)
		})
	}
}

func TestTrackingMetrics(t *testing.T) {
	type testcase struct {
		name         string
		dropOriginal bool
		input        []telegraf.Metric
		expected     []telegraf.Metric
	}
	testcases := []testcase{
		{
			name:         "keep all",
			dropOriginal: false,
			input: []telegraf.Metric{
				metric.New("foo", map[string]string{}, map[string]interface{}{"value": 42}, time.Unix(0, 0)),
				metric.New("bar", map[string]string{}, map[string]interface{}{"value": 99}, time.Unix(0, 0)),
				metric.New("baz", map[string]string{}, map[string]interface{}{"value": 1}, time.Unix(0, 0)),
			},
			expected: []telegraf.Metric{
				metric.New("foo", map[string]string{}, map[string]interface{}{"value": 42}, time.Unix(0, 0)),
				metric.New("bar", map[string]string{}, map[string]interface{}{"value": 99}, time.Unix(0, 0)),
				metric.New("baz", map[string]string{}, map[string]interface{}{"value": 1}, time.Unix(0, 0)),
				metric.New("new", map[string]string{}, map[string]interface{}{"value": 42}, time.Unix(0, 0)),
				metric.New("new", map[string]string{}, map[string]interface{}{"value": 99}, time.Unix(0, 0)),
				metric.New("new", map[string]string{}, map[string]interface{}{"value": 1}, time.Unix(0, 0)),
			},
		},
		{
			name:         "drop original",
			dropOriginal: true,
			input: []telegraf.Metric{
				metric.New("foo", map[string]string{}, map[string]interface{}{"value": 42}, time.Unix(0, 0)),
				metric.New("bar", map[string]string{}, map[string]interface{}{"value": 99}, time.Unix(0, 0)),
				metric.New("baz", map[string]string{}, map[string]interface{}{"value": 1}, time.Unix(0, 0)),
			},
			expected: []telegraf.Metric{
				metric.New("new", map[string]string{}, map[string]interface{}{"value": 42}, time.Unix(0, 0)),
				metric.New("new", map[string]string{}, map[string]interface{}{"value": 99}, time.Unix(0, 0)),
				metric.New("new", map[string]string{}, map[string]interface{}{"value": 1}, time.Unix(0, 0)),
			},
		},
	}
	for _, tc := range testcases {
		var mu sync.Mutex
		delivered := make([]telegraf.DeliveryInfo, 0, len(tc.input))
		notify := func(di telegraf.DeliveryInfo) {
			mu.Lock()
			defer mu.Unlock()
			delivered = append(delivered, di)
		}

		input := make([]telegraf.Metric, 0, len(tc.input))
		for _, m := range tc.input {
			tm, _ := metric.WithTracking(m, notify)
			input = append(input, tm)
		}

		plugin := &Split{
			DropOriginal: tc.dropOriginal,
			Templates: []template{
				{
					Name:   "new",
					Fields: []string{"value"},
				},
			},
		}
		require.NoError(t, plugin.Init())

		// Process expected metrics and compare with resulting metrics
		actual := plugin.Apply(input...)
		testutil.RequireMetricsEqual(t, tc.expected, actual, testutil.SortMetrics())

		// Simulate output acknowledging delivery
		for _, m := range actual {
			m.Accept()
		}

		// Check delivery
		require.Eventuallyf(t, func() bool {
			mu.Lock()
			defer mu.Unlock()
			return len(tc.input) == len(delivered)
		}, time.Second, 100*time.Millisecond, "%d delivered but %d expected", len(delivered), len(tc.input))
	}
}
