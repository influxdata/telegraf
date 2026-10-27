//go:build linux

package powercap

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

type zoneFiles struct {
	zone        string
	name        string
	energy      string
	energyRange string
}

func TestGatherReportsEnergyOfEveryZone(t *testing.T) {
	sysPath := t.TempDir()
	t.Setenv("HOST_SYS", sysPath)
	writeControlType(t, sysPath, "intel-rapl")
	writeZones(t, sysPath,
		zoneFiles{zone: "intel-rapl:0", name: "package-0", energy: "172380552916", energyRange: "262143328850"},
		zoneFiles{zone: "intel-rapl:0:0", name: "dram", energy: "21843017463", energyRange: "65712999613"},
		zoneFiles{zone: "intel-rapl:1", name: "package-1", energy: "160121449030", energyRange: "262143328850"},
		zoneFiles{zone: "intel-rapl-mmio:0", name: "package-0", energy: "3184012553", energyRange: "262143328850"},
	)

	plugin := &Powercap{Log: testutil.Logger{}}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	expected := []telegraf.Metric{
		energyMetric("intel-rapl:0", "package-0", 172380552916),
		energyMetric("intel-rapl:0:0", "dram", 21843017463),
		energyMetric("intel-rapl:1", "package-1", 160121449030),
		energyMetric("intel-rapl-mmio:0", "package-0", 3184012553),
	}
	require.Empty(t, acc.Errors)
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), testutil.SortMetrics())
}

func TestGatherSkipsZonesWithoutEnergyCounter(t *testing.T) {
	sysPath := t.TempDir()
	t.Setenv("HOST_SYS", sysPath)
	writeControlType(t, sysPath, "dtpm")
	writeFile(t, filepath.Join(sysPath, "class", "powercap", "dtpm:0", "name"), "soc")

	plugin := &Powercap{Log: testutil.Logger{}}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	require.Empty(t, acc.Errors)
	require.Empty(t, acc.GetTelegrafMetrics())
}

func TestGatherReportsPowerFromTheSecondReading(t *testing.T) {
	sysPath := t.TempDir()
	t.Setenv("HOST_SYS", sysPath)
	package0 := zoneFiles{zone: "intel-rapl:0", name: "package-0", energy: "1000000", energyRange: "65532610987"}
	writeZones(t, sysPath, package0)

	plugin := &Powercap{Log: testutil.Logger{}}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))
	require.False(t, acc.HasField("powercap", "power_watts"))

	package0.energy = "2000000"
	writeZones(t, sysPath, package0)
	acc.ClearMetrics()
	require.NoError(t, plugin.Gather(&acc))

	require.Empty(t, acc.Errors)
	watts, found := acc.FloatField("powercap", "power_watts")
	require.True(t, found)
	require.Positive(t, watts)
}

func TestGatherReportsUnreadableZoneAndContinues(t *testing.T) {
	sysPath := t.TempDir()
	t.Setenv("HOST_SYS", sysPath)
	writeZones(t, sysPath,
		zoneFiles{zone: "intel-rapl:0", name: "package-0", energy: "not a number", energyRange: "65532610987"},
		zoneFiles{zone: "intel-rapl:0:0", name: "core", energy: "13382443108", energyRange: "65532610987"},
	)

	plugin := &Powercap{Log: testutil.Logger{}}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	require.Len(t, acc.Errors, 1)
	require.ErrorContains(t, acc.Errors[0], `reading energy of zone "intel-rapl:0" failed`)
	expected := []telegraf.Metric{energyMetric("intel-rapl:0:0", "core", 13382443108)}
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
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

func energyMetric(zone, name string, microjoules uint64) telegraf.Metric {
	return metric.New(
		"powercap",
		map[string]string{"zone": zone, "name": name},
		map[string]any{"energy_uj": microjoules},
		time.Unix(0, 0),
	)
}

func writeControlType(t *testing.T, sysPath, controlType string) {
	writeFile(t, filepath.Join(sysPath, "class", "powercap", controlType, "enabled"), "1")
}

func writeZones(t *testing.T, sysPath string, zones ...zoneFiles) {
	for _, zone := range zones {
		zonePath := filepath.Join(sysPath, "class", "powercap", zone.zone)
		writeFile(t, filepath.Join(zonePath, "name"), zone.name)
		writeFile(t, filepath.Join(zonePath, "energy_uj"), zone.energy)
		writeFile(t, filepath.Join(zonePath, "max_energy_range_uj"), zone.energyRange)
	}
}

func writeFile(t *testing.T, path, content string) {
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o750))
	require.NoError(t, os.WriteFile(path, []byte(content+"\n"), 0o600))
}
