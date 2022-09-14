//go:build linux

package linux_cpu

import (
	"github.com/influxdata/telegraf/testutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNoMetrics(t *testing.T) {
	plugin := &LinuxCPU{}
	require.Error(t, plugin.Init())
}

func TestNoCPUs(t *testing.T) {
	td := t.TempDir()

	plugin := &LinuxCPU{
		Log:       testutil.Logger{Name: "LinuxCPUPluginTest"},
		Metrics:   []string{"cpufreq"},
		PathSysfs: td,
	}
	require.Error(t, plugin.Init())
}

func TestNoCPUMetrics(t *testing.T) {
	td := t.TempDir()

	require.NoError(t, os.MkdirAll(td+"/devices/system/cpu/cpu0/cpufreq", os.ModePerm))

	plugin := &LinuxCPU{
		Log:       testutil.Logger{Name: "LinuxCPUPluginTest"},
		Metrics:   []string{"cpufreq"},
		PathSysfs: td,
	}
	require.Error(t, plugin.Init())
}

func TestGatherCPUFreq(t *testing.T) {
	td := t.TempDir()

	require.NoError(t, os.MkdirAll(td+"/devices/system/cpu/cpu0/cpufreq", os.ModePerm))
	require.NoError(t, os.WriteFile(td+"/devices/system/cpu/cpu0/cpufreq/scaling_cur_freq", []byte("250\n"), 0644))
	require.NoError(t, os.WriteFile(td+"/devices/system/cpu/cpu0/cpufreq/scaling_min_freq", []byte("100\n"), 0644))
	require.NoError(t, os.WriteFile(td+"/devices/system/cpu/cpu0/cpufreq/scaling_max_freq", []byte("255\n"), 0644))

	require.NoError(t, os.MkdirAll(td+"/devices/system/cpu/cpu1/cpufreq", os.ModePerm))
	require.NoError(t, os.WriteFile(td+"/devices/system/cpu/cpu1/cpufreq/scaling_cur_freq", []byte("123\n"), 0644))
	require.NoError(t, os.WriteFile(td+"/devices/system/cpu/cpu1/cpufreq/scaling_min_freq", []byte("80\n"), 0644))
	require.NoError(t, os.WriteFile(td+"/devices/system/cpu/cpu1/cpufreq/scaling_max_freq", []byte("230\n"), 0644))

	plugin := &LinuxCPU{
		Log:       testutil.Logger{Name: "LinuxCPUPluginTest"},
		Metrics:   []string{"cpufreq"},
		PathSysfs: td,
	}

	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	tags1 := map[string]string{
		"cpu": "0",
	}

	tags2 := map[string]string{
		"cpu": "1",
	}

	fields1 := map[string]interface{}{
		"scaling_cur_freq": uint64(250),
		"scaling_min_freq": uint64(100),
		"scaling_max_freq": uint64(255),
	}

	fields2 := map[string]interface{}{
		"scaling_cur_freq": uint64(123),
		"scaling_min_freq": uint64(80),
		"scaling_max_freq": uint64(230),
	}

	acc.AssertContainsTaggedFields(t, "linux_cpu", fields1, tags1)
	acc.AssertContainsTaggedFields(t, "linux_cpu", fields2, tags2)
}

func TestGatherThermal(t *testing.T) {
	td := t.TempDir()

	require.NoError(t, os.MkdirAll(td+"/devices/system/cpu/cpu0/thermal_throttle", os.ModePerm))
	require.NoError(t, os.WriteFile(td+"/devices/system/cpu/cpu0/thermal_throttle/core_throttle_count", []byte("250\n"), 0644))
	require.NoError(t, os.WriteFile(td+"/devices/system/cpu/cpu0/thermal_throttle/core_throttle_max_time_ms", []byte("100\n"), 0644))
	require.NoError(t, os.WriteFile(td+"/devices/system/cpu/cpu0/thermal_throttle/core_throttle_total_time_ms", []byte("255\n"), 0644))

	plugin := &LinuxCPU{
		Log:       testutil.Logger{Name: "LinuxCPUPluginTest"},
		Metrics:   []string{"thermal"},
		PathSysfs: td,
	}

	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	acc.AssertContainsFields(t, "linux_cpu", map[string]interface{}{
		"throttle_count":      uint64(250),
		"throttle_max_time":   uint64(100),
		"throttle_total_time": uint64(255),
	})
}

func TestGatherPropertyRemoved(t *testing.T) {
	td := t.TempDir()

	require.NoError(t, os.MkdirAll(td+"/devices/system/cpu/cpu0/cpufreq", os.ModePerm))
	require.NoError(t, os.WriteFile(td+"/devices/system/cpu/cpu0/cpufreq/scaling_cur_freq", []byte("250\n"), 0644))
	require.NoError(t, os.WriteFile(td+"/devices/system/cpu/cpu0/cpufreq/scaling_min_freq", []byte("100\n"), 0644))
	require.NoError(t, os.WriteFile(td+"/devices/system/cpu/cpu0/cpufreq/scaling_max_freq", []byte("255\n"), 0644))

	plugin := &LinuxCPU{
		Log:       testutil.Logger{Name: "LinuxCPUPluginTest"},
		Metrics:   []string{"cpufreq"},
		PathSysfs: td,
	}

	require.NoError(t, plugin.Init())

	// Remove one of the properties
	require.NoError(t, os.RemoveAll(td+"/devices/system/cpu/cpu0/cpufreq/scaling_max_freq"))

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	tags1 := map[string]string{
		"cpu": "0",
	}

	fields1 := map[string]interface{}{
		"scaling_cur_freq": uint64(250),
		"scaling_min_freq": uint64(100),
		"scaling_max_freq": uint64(255),
	}

	acc.AssertDoesNotContainsTaggedFields(t, "linux_cpu", fields1, tags1)
	require.NotEmpty(t, acc.Errors)
}

func TestGatherPropertyInvalid(t *testing.T) {
	td := t.TempDir()

	require.NoError(t, os.MkdirAll(td+"/devices/system/cpu/cpu0/cpufreq", os.ModePerm))
	require.NoError(t, os.WriteFile(td+"/devices/system/cpu/cpu0/cpufreq/scaling_cur_freq", []byte("ABC\n"), 0644))
	require.NoError(t, os.WriteFile(td+"/devices/system/cpu/cpu0/cpufreq/scaling_min_freq", []byte("100\n"), 0644))
	require.NoError(t, os.WriteFile(td+"/devices/system/cpu/cpu0/cpufreq/scaling_max_freq", []byte("255\n"), 0644))

	plugin := &LinuxCPU{
		Log:       testutil.Logger{Name: "LinuxCPUPluginTest"},
		Metrics:   []string{"cpufreq"},
		PathSysfs: td,
	}

	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	tags1 := map[string]string{
		"cpu": "0",
	}

	fields1 := map[string]interface{}{
		"scaling_cur_freq": uint64(250),
		"scaling_min_freq": uint64(100),
		"scaling_max_freq": uint64(255),
	}

	acc.AssertDoesNotContainsTaggedFields(t, "linux_cpu", fields1, tags1)
	require.NotEmpty(t, acc.Errors)
}
