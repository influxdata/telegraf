//go:build linux

package powercap

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestCases(t *testing.T) {
	folders, err := os.ReadDir("testcases")
	require.NoError(t, err)

	parser := &influx.Parser{}
	require.NoError(t, parser.Init())

	for _, f := range folders {
		if !f.IsDir() {
			continue
		}
		testcasePath := filepath.Join("testcases", f.Name())
		configFilename := filepath.Join(testcasePath, "telegraf.conf")
		expectedFilename := filepath.Join(testcasePath, "expected.out")
		expectedErrorFilename := filepath.Join(testcasePath, "expected.err")

		t.Run(f.Name(), func(t *testing.T) {
			var expected []telegraf.Metric
			if _, err := os.Stat(expectedFilename); err == nil {
				expected, err = testutil.ParseMetricsFromFile(expectedFilename, parser)
				require.NoError(t, err)
			}

			var expectedErrors []string
			if _, err := os.Stat(expectedErrorFilename); err == nil {
				expectedErrors, err = testutil.ParseLinesFromFile(expectedErrorFilename)
				require.NoError(t, err)
				require.NotEmpty(t, expectedErrors)
			}

			t.Setenv("HOST_SYS", filepath.Join(testcasePath, "sys"))

			cfg := config.NewConfig()
			require.NoError(t, cfg.LoadConfig(configFilename))
			require.Len(t, cfg.Inputs, 1)
			plugin := cfg.Inputs[0].Input.(*Powercap)
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			require.NoError(t, plugin.Gather(&acc))

			actualErrors := make([]string, 0, len(acc.Errors))
			for _, err := range acc.Errors {
				actualErrors = append(actualErrors, err.Error())
			}
			require.ElementsMatch(t, expectedErrors, actualErrors)
			testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), testutil.SortMetrics())
		})
	}
}

func TestGatherReportsPowerFromTheSecondReading(t *testing.T) {
	sysPath := t.TempDir()
	t.Setenv("HOST_SYS", sysPath)
	zonePath := filepath.Join(sysPath, "class", "powercap", "intel-rapl:0")
	writeFile(t, filepath.Join(zonePath, "name"), "package-0")
	writeFile(t, filepath.Join(zonePath, "max_energy_range_uj"), "65532610987")
	writeFile(t, filepath.Join(zonePath, "energy_uj"), "1000000")

	plugin := &Powercap{Log: testutil.Logger{}}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))
	require.False(t, acc.HasField("powercap", "power_watts"))

	writeFile(t, filepath.Join(zonePath, "energy_uj"), "2000000")
	acc.ClearMetrics()
	require.NoError(t, plugin.Gather(&acc))

	require.Empty(t, acc.Errors)
	watts, found := acc.FloatField("powercap", "power_watts")
	require.True(t, found)
	require.Positive(t, watts)
}

func TestAverageWatts(t *testing.T) {
	start := time.Unix(0, 0)
	tests := []struct {
		name     string
		earlier  energyReading
		later    energyReading
		wrapsAt  uint64
		expected float64
	}{
		{
			name:     "counter increases",
			earlier:  energyReading{microjoules: 1_000_000, time: start},
			later:    energyReading{microjoules: 101_000_000, time: start.Add(10 * time.Second)},
			wrapsAt:  65_532_610_987,
			expected: 10,
		},
		{
			name:     "counter wraps around",
			earlier:  energyReading{microjoules: 65_482_610_987, time: start},
			later:    energyReading{microjoules: 50_000_000, time: start.Add(10 * time.Second)},
			wrapsAt:  65_532_610_987,
			expected: 10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.InDelta(t, tt.expected, averageWatts(tt.earlier, tt.later, tt.wrapsAt), 1e-9)
		})
	}
}

func writeFile(t *testing.T, path, content string) {
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o750))
	require.NoError(t, os.WriteFile(path, []byte(content+"\n"), 0o600))
}
