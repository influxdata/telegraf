package procstat

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func init() {
	t128execCommand = mockExecCommand
}

func TestT128Gather_CreateProcessErrorOk(t *testing.T) {
	var acc testutil.Accumulator
	p := T128Procstat{
		Exe:             exe,
		createPIDFinder: pidFinder([]PID{pid}),
		createProcess: func(PID) (Process, error) {
			return nil, fmt.Errorf("createProcess error")
		},
	}
	require.NoError(t, acc.GatherError(p.Gather))
}

func TestT128Gather_CreatePIDFinderError(t *testing.T) {
	var acc testutil.Accumulator

	p := T128Procstat{
		createPIDFinder: func() (PIDFinder, error) {
			return nil, fmt.Errorf("createPIDFinder error")
		},
		createProcess: newTestProc,
	}
	require.Error(t, acc.GatherError(p.Gather))
}

func TestT128Gather_ProcessName(t *testing.T) {
	var acc testutil.Accumulator

	p := T128Procstat{
		Exe:             exe,
		ProcessName:     "custom_name",
		createPIDFinder: pidFinder([]PID{pid}),
		createProcess:   newTestProc,
	}
	require.NoError(t, acc.GatherError(p.Gather))

	require.Equal(t, "custom_name", acc.TagValue("procstat", "process_name"))
}

func TestT128Gather_NoProcessNameUsesReal(t *testing.T) {
	var acc testutil.Accumulator
	pid := PID(os.Getpid())

	p := T128Procstat{
		Exe:             exe,
		createPIDFinder: pidFinder([]PID{pid}),
		createProcess:   newTestProc,
	}
	require.NoError(t, acc.GatherError(p.Gather))

	require.True(t, acc.HasTag("procstat", "process_name"))
}

func TestT128Gather_PercentFirstPass(t *testing.T) {
	var acc testutil.Accumulator
	pid := PID(os.Getpid())

	p := T128Procstat{
		Pattern:         "foo",
		PidTag:          true,
		createPIDFinder: pidFinder([]PID{pid}),
		createProcess:   NewProc,
	}
	require.NoError(t, acc.GatherError(p.Gather))

	require.True(t, acc.HasFloatField("procstat", "cpu_time_user"))
	require.False(t, acc.HasFloatField("procstat", "cpu_usage"))
}

func TestT128Gather_PercentSecondPass(t *testing.T) {
	var acc testutil.Accumulator
	pid := PID(os.Getpid())

	p := T128Procstat{
		Pattern:         "foo",
		PidTag:          true,
		createPIDFinder: pidFinder([]PID{pid}),
		createProcess:   NewProc,
	}
	require.NoError(t, acc.GatherError(p.Gather))
	require.NoError(t, acc.GatherError(p.Gather))

	require.True(t, acc.HasFloatField("procstat", "cpu_time_user"))
	require.True(t, acc.HasFloatField("procstat", "cpu_usage"))
}

func TestT128Gather_systemdUnitPIDs(t *testing.T) {
	p := T128Procstat{
		createPIDFinder: pidFinder([]PID{}),
		SystemdUnit:     "TestGather_systemdUnitPIDs",
	}
	pidsTags := p.findPids()
	for _, pidsTag := range pidsTags {
		pids := pidsTag.PIDS
		tags := pidsTag.Tags
		err := pidsTag.Err
		require.NoError(t, err)
		require.Equal(t, []PID{11408}, pids)
		require.Equal(t, "TestGather_systemdUnitPIDs", tags["systemd_unit"])
	}
}

func TestT128Gather_cgroupPIDs(t *testing.T) {
	//no cgroups in windows
	if runtime.GOOS == "windows" {
		t.Skip("no cgroups in windows")
	}
	td, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	defer os.RemoveAll(td)
	err = os.WriteFile(filepath.Join(td, "cgroup.procs"), []byte("1234\n5678\n"), 0644)
	require.NoError(t, err)

	p := T128Procstat{
		createPIDFinder: pidFinder([]PID{}),
		CGroup:          td,
	}
	pidsTags := p.findPids()
	for _, pidsTag := range pidsTags {
		pids := pidsTag.PIDS
		tags := pidsTag.Tags
		err := pidsTag.Err
		require.NoError(t, err)
		require.Equal(t, []PID{1234, 5678}, pids)
		require.Equal(t, td, tags["cgroup"])
	}
}

func TestT128ProcstatLookupMetric(t *testing.T) {
	p := Procstat{
		createPIDFinder: pidFinder([]PID{543}),
		Exe:             "-Gsys",
	}
	var acc testutil.Accumulator
	err := acc.GatherError(p.Gather)
	require.NoError(t, err)
	require.Equal(t, len(p.procs)+1, len(acc.Metrics))
}

func TestT128Gather_SameTimestamps(t *testing.T) {
	var acc testutil.Accumulator
	pidfile := "/path/to/pidfile"

	p := T128Procstat{
		PidFile:         pidfile,
		createPIDFinder: pidFinder([]PID{pid}),
		createProcess:   newTestProc,
	}
	require.NoError(t, acc.GatherError(p.Gather))

	procstat, _ := acc.Get("procstat")
	procstatLookup, _ := acc.Get("procstat_lookup")

	require.Equal(t, procstat.Time, procstatLookup.Time)
}
