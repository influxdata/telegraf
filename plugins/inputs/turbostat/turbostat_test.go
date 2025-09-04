//go:build linux && amd64

package turbostat

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestCreateColumn(t *testing.T) {
	tests := []struct {
		in  string
		out column
	}{
		{in: "Core", out: column{name: "core", isTag: true}},
		{in: "Busy%", out: column{name: "busy_percent"}},
		{in: "Bzy_MHz", out: column{name: "busy_frequency_mhz"}},
		{in: "C3-", out: column{name: "c3_minus"}},
		{in: "CoreThr", out: column{name: "core_throttle"}},
		{in: "Cor_J", out: column{name: "core_energy_joule"}},
		{in: "CorWatt", out: column{name: "core_power_watt"}},
		{in: "UMHz1.0", out: column{name: "uncore_frequency_mhz_1_0"}},
		{in: "Time_Of_Day_Seconds", out: column{name: "time_of_day_seconds", isTime: true}},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			require.Equal(t, tt.out, createColumn(tt.in))
		})
	}
}

func TestProcessValues(t *testing.T) {
	columns := []column{
		{name: "cpu", isTag: true},
		{name: "core", isTag: true},
		{name: "busy_frequency_mhz", isTag: false},
		{name: "core_power_watt", isTag: false},
	}
	tests := []struct {
		values []string
		metric telegraf.Metric
	}{
		{
			values: []string{"-", "-", "1.23", "4.56"},
			metric: metric.New(
				"turbostat",
				map[string]string{"cpu": "-", "core": "-"},
				map[string]interface{}{"busy_frequency_mhz": 1.23, "core_power_watt": 4.56},
				time.Time{},
			),
		},
		{
			values: []string{"0", "1", "1.23"},
			metric: metric.New(
				"turbostat",
				map[string]string{"cpu": "0", "core": "1"},
				map[string]interface{}{"busy_frequency_mhz": 1.23},
				time.Time{},
			),
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s", tt.values), func(t *testing.T) {
			acc := &testutil.Accumulator{}
			err := processValues(acc, columns, tt.values)
			require.NoError(t, err)
			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, []telegraf.Metric{tt.metric}, actual, testutil.IgnoreTime())
		})
	}
}

func TestProcessValuesErrors(t *testing.T) {
	columns := []column{
		{name: "cpu", isTag: true},
		{name: "core", isTag: true},
		{name: "busy_frequency_mhz", isTag: false},
		{name: "core_power_watt", isTag: false},
	}
	tests := []struct {
		values []string
		err    string
	}{
		{
			values: []string{"0", "1"},
			err:    "no value for any field",
		},
		{
			values: []string{"0", "1", "123", "456", "789"},
			err:    "too many values: 4 columns, 5 values",
		},
		{
			values: []string{"0", "1", "xyz"},
			err:    "unable to parse column \"busy_frequency_mhz\": strconv.ParseFloat: parsing \"xyz\": invalid syntax",
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s", tt.values), func(t *testing.T) {
			var acc testutil.Accumulator
			err := processValues(&acc, columns, tt.values)
			require.ErrorContains(t, err, tt.err)
		})
	}
}

func TestCases(t *testing.T) {
	dirs, err := os.ReadDir("testcases")
	require.NoError(t, err)
	for _, f := range dirs {
		// Only handle directories.
		if !f.IsDir() {
			continue
		}

		testcasePath := filepath.Join("testcases", f.Name())
		configFilename := filepath.Join(testcasePath, "telegraf.conf")
		inputFilename := filepath.Join(testcasePath, "input.txt")
		expectedFilename := filepath.Join(testcasePath, "expected.out")

		t.Run(f.Name(), func(t *testing.T) {
			// Load the expected output.
			parser := &influx.Parser{}
			require.NoError(t, parser.Init())
			expected, err := testutil.ParseMetricsFromFile(expectedFilename, parser)
			require.NoError(t, err)

			// Configure the plugin.
			cfg := config.NewConfig()
			require.NoError(t, cfg.LoadConfig(configFilename))
			require.Len(t, cfg.Inputs, 1)
			plugin := cfg.Inputs[0].Input.(*Turbostat)
			require.NoError(t, plugin.Init())

			// Replace Turbostat with a mock process which outputs the desired lines.
			exe, err := os.Executable()
			require.NoError(t, err)
			plugin.command = []string{exe, "-mockInput=" + inputFilename}

			// Start the plugin and wait for the expected metrics.
			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))

			require.Eventually(t, func() bool {
				return acc.NMetrics() == uint64(len(expected))
			}, 3*time.Second, 100*time.Millisecond)

			require.Empty(t, acc.Errors)

			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime())
		})
	}
}

// This function lets the test executable to behave as a mock Turbostat process.
func TestMain(m *testing.M) {
	var mockInput string
	flag.StringVar(&mockInput, "mockInput", "", "mock turbostat output")
	flag.Parse()

	if mockInput == "" {
		os.Exit(m.Run())
	}

	err := func() error {
		file, err := os.Open(mockInput)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(os.Stdout, file)
		if err != nil {
			return err
		}
		return nil
	}()
	if err != nil {
		os.Exit(5) // EIO
	}

	// Sleep. Terminating immediately would look like Turbostat crashed.
	time.Sleep(3 * time.Second)
}
