//go:build linux && amd64

package parser

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestSplitSymbols(t *testing.T) {
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
			assert.Equal(t, tt.out, splitSymbols(tt.in))
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
			assert.Equal(t, tt.out, splitKnownTokens(tt.in))
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
		{in: "Time_Of_Day_Seconds", out: column{name: "time_of_day_seconds", isIgnored: true}},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			assert.Equal(t, tt.out, createColumn(tt.in))
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
			assert.Equal(t, tt.out, isValidTagValue(tt.in))
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
		tags   tagMap
		fields fieldMap
		err    string
	}{
		{
			values: []string{"-", "-", "1.23", "4.56"},
			tags:   tagMap{"cpu": "-", "core": "-"},
			fields: fieldMap{"busy_frequency_mhz": 1.23, "core_power_watt": 4.56},
			err:    "",
		},
		{
			values: []string{"0", "1", "1.23"},
			tags:   tagMap{"cpu": "0", "core": "1"},
			fields: fieldMap{"busy_frequency_mhz": 1.23},
			err:    "",
		},
		{
			values: []string{"0", "1"},
			tags:   nil,
			fields: nil,
			err:    "no value for any field",
		},
		{
			values: []string{"0", "1", "123", "456", "789"},
			tags:   nil,
			fields: nil,
			err:    "too many values: 4 columns, 5 values",
		},
		{
			values: []string{"?", "1", "123"},
			tags:   nil,
			fields: nil,
			err:    "invalid tag: ?",
		},
		{
			values: []string{"0", "1", "xyz"},
			tags:   nil,
			fields: nil,
			err:    "strconv.ParseFloat: parsing \"xyz\": invalid syntax",
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s", tt.values), func(t *testing.T) {
			tags, fields, err := processValues(columns, tt.values)
			if tt.err == "" {
				require.NoError(t, err)
			} else {
				assert.Equal(t, tt.err, err.Error())
			}
			assert.Equal(t, tt.tags, tags)
			assert.Equal(t, tt.fields, fields)
		})
	}
}

func TestProcessStdout(t *testing.T) {
	tests := []struct {
		in  string
		out string
	}{
		{in: "testdata/in1", out: "testdata/out1"},
		{in: "testdata/in2", out: "testdata/out2"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			parser := &influx.Parser{}
			require.NoError(t, parser.Init())
			expected, err := testutil.ParseMetricsFromFile(tt.out, parser)
			require.NoError(t, err)
			f, err := os.Open(tt.in)
			require.NoError(t, err)
			defer f.Close()
			acc := &testutil.Accumulator{}
			require.NoError(t, ProcessStream(f, acc))
			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime())
		})
	}
}
