//go:build linux && amd64

package turbostat

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestBuildCmd(t *testing.T) {
	tests := []struct {
		in  Turbostat
		out []string
	}{
		{
			in: Turbostat{
				UseSudo:  true,
				Path:     "/usr/bin/turbostat",
				Interval: config.Duration(10 * time.Second),
			},
			out: []string{"sudo", "/usr/bin/turbostat", "--quiet", "--interval", "10", "--show", "all"},
		},
		{
			in: Turbostat{
				UseSudo:  false,
				Path:     "turbostat",
				Interval: config.Duration(1 * time.Minute),
			},
			out: []string{"turbostat", "--quiet", "--interval", "60", "--show", "all"},
		},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			require.Equal(t, tt.out, tt.in.buildCmd())
		})
	}
}

func TestSplitTokens(t *testing.T) {
	tests := []struct {
		in  string
		out []string
	}{
		{in: "X2APIC", out: []string{"X2APIC"}},
		{in: "Bzy_MHz", out: []string{"Bzy", "MHz"}},
		{in: "Busy%", out: []string{"Busy", "%"}},
		{in: "CPU%c6", out: []string{"CPU", "%", "c6"}},
		{in: "PKG_%", out: []string{"PKG", "%"}},
		{in: "C3+", out: []string{"C3", "+"}},
		{in: "UMHz1.0", out: []string{"UMHz1", "0"}},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			require.Equal(t, tt.out, splitTokens(tt.in))
		})
	}
}

func TestSplitKnownTokens(t *testing.T) {
	tests := []struct {
		in  string
		out []string
	}{
		{in: "Core", out: []string{"Core"}},
		{in: "MHz", out: []string{"MHz"}},
		{in: "UncMHz", out: []string{"Unc", "MHz"}},
		{in: "UMHz1", out: []string{"U", "MHz", "1"}},
		{in: "PkgTmp", out: []string{"Pkg", "Tmp"}},
		{in: "CoreThr", out: []string{"Core", "Thr"}},
		{in: "CorWatt", out: []string{"Cor", "Watt"}},
		{in: "GFXMHz", out: []string{"GFX", "MHz"}},
		{in: "GFXAMHz", out: []string{"GFX", "A", "MHz"}},
		{in: "SAMMHz", out: []string{"SAM", "MHz"}},
		{in: "SAMAMHz", out: []string{"SAM", "A", "MHz"}},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			require.Equal(t, tt.out, splitKnownTokens(tt.in))
		})
	}
}

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

func TestIsValidTagValue(t *testing.T) {
	tests := []struct {
		in  string
		out bool
	}{
		{in: "0", out: true},
		{in: "123", out: true},
		{in: "-", out: true},
		{in: "abc", out: false},
		{in: "*", out: false},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			require.Equal(t, tt.out, isValidTagValue(tt.in))
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
		err    string
	}{
		{
			values: []string{"-", "-", "1.23", "4.56"},
			metric: metric.New(
				"turbostat",
				map[string]string{"cpu": "-", "core": "-"},
				map[string]interface{}{"busy_frequency_mhz": 1.23, "core_power_watt": 4.56},
				time.Time{},
			),
			err: "",
		},
		{
			values: []string{"0", "1", "1.23"},
			metric: metric.New(
				"turbostat",
				map[string]string{"cpu": "0", "core": "1"},
				map[string]interface{}{"busy_frequency_mhz": 1.23},
				time.Time{},
			),
			err: "",
		},
		{
			values: []string{"0", "1"},
			metric: nil,
			err:    "no value for any field",
		},
		{
			values: []string{"0", "1", "123", "456", "789"},
			metric: nil,
			err:    "too many values: 4 columns, 5 values",
		},
		{
			values: []string{"?", "1", "123"},
			metric: nil,
			err:    "invalid tag: ?",
		},
		{
			values: []string{"0", "1", "xyz"},
			metric: nil,
			err:    "unable to parse column \"busy_frequency_mhz\": strconv.ParseFloat: parsing \"xyz\": invalid syntax",
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s", tt.values), func(t *testing.T) {
			acc := &testutil.Accumulator{}
			err := processValues(columns, tt.values, acc)
			if tt.err == "" {
				require.NoError(t, err)
				actual := acc.GetTelegrafMetrics()
				testutil.RequireMetricsEqual(t, []telegraf.Metric{tt.metric}, actual, testutil.IgnoreTime())
			} else {
				require.Equal(t, tt.err, err.Error())
			}
		})
	}
}

func TestProcessStdout(t *testing.T) {
	tests := []struct {
		stream  string
		metrics []telegraf.Metric
	}{
		{
			stream: strings.Join([]string{
				"Time_Of_Day_Seconds     Core    Bzy_MHz",
				"1753026271.021766       -       3484",
				"1753026271.021581       0       3493",
			}, "\n"),
			metrics: []telegraf.Metric{
				metric.New(
					"turbostat",
					map[string]string{"core": "-"},
					map[string]interface{}{"busy_frequency_mhz": 3484.0},
					time.Time{},
				),
				metric.New(
					"turbostat",
					map[string]string{"core": "0"},
					map[string]interface{}{"busy_frequency_mhz": 3493.0},
					time.Time{},
				),
			},
		},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			s := strings.NewReader(tt.stream)
			acc := &testutil.Accumulator{}
			require.NoError(t, processStdout(s, acc))
			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, tt.metrics, actual, testutil.IgnoreTime())
		})
	}
}

func TestProcessStderr(t *testing.T) {
	tests := []struct {
		stream string
		errors []error
	}{
		{
			stream: strings.Join([]string{
				"here is an error",
				"and another",
			}, "\n"),
			errors: []error{
				errors.New("here is an error"),
				errors.New("and another"),
			},
		},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			s := strings.NewReader(tt.stream)
			acc := &testutil.Accumulator{}
			require.NoError(t, processStderr(s, acc))
			actual := acc.Errors
			require.Equal(t, tt.errors, actual)
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
		pipeFilename := filepath.Join(testcasePath, "pipe")
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

			// Replace Turbostat with a process which outputs the desired lines.
			// We use a named pipe that we don't write to to block like Turbostat
			// would, instead of exiting after outputing the lines.
			require.NoError(t, syscall.Mkfifo(pipeFilename, 0666))
			defer os.Remove(pipeFilename)
			plugin.command = []string{"cat", inputFilename, pipeFilename}

			// Start the plugin and wait for the expected metrics.
			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			acc.Wait(len(expected))
			require.Empty(t, acc.Errors)

			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime())
		})
	}
}
