package procstat

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/process"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testPgrep struct {
	pids []PID
	err  error
}

func pidFinder(pids []PID, err error) func() (PIDFinder, error) {
	return func() (PIDFinder, error) {
		return &testPgrep{
			pids: pids,
			err:  err,
		}, nil
	}
}

func (pg *testPgrep) PidFile(path string) ([]PID, error) {
	return pg.pids, pg.err
}

func (pg *testPgrep) Pattern(pattern string) ([]PID, error) {
	return pg.pids, pg.err
}

func (pg *testPgrep) Uid(user string) ([]PID, error) {
	return pg.pids, pg.err
}

func (pg *testPgrep) FullPattern(pattern string) ([]PID, error) {
	return pg.pids, pg.err
}

type testProc struct {
	pid  PID
	tags map[string]string
}

func newTestProc(pid PID) (Process, error) {
	proc := &testProc{
		tags: make(map[string]string),
	}
	return proc, nil
}

func (p *testProc) PID() PID {
	return p.pid
}

func (p *testProc) Tags() map[string]string {
	return p.tags
}

func (p *testProc) IOCounters() (*process.IOCountersStat, error) {
	return &process.IOCountersStat{}, nil
}

func (p *testProc) MemoryInfo() (*process.MemoryInfoStat, error) {
	return &process.MemoryInfoStat{}, nil
}

func (p *testProc) Name() (string, error) {
	return "test_proc", nil
}

func (p *testProc) NumCtxSwitches() (*process.NumCtxSwitchesStat, error) {
	return &process.NumCtxSwitchesStat{}, nil
}

func (p *testProc) NumFDs() (int32, error) {
	return 0, nil
}

func (p *testProc) NumThreads() (int32, error) {
	return 0, nil
}

func (p *testProc) Percent(interval time.Duration) (float64, error) {
	return 0, nil
}

func (p *testProc) Times() (*cpu.TimesStat, error) {
	return &cpu.TimesStat{}, nil
}

var pid PID = PID(42)
var exe string = "foo"

func TestGather_CreateProcessErrorOk(t *testing.T) {
	var acc testutil.Accumulator

	p := Procstat{
		Exe:             exe,
		createPIDFinder: pidFinder([]PID{pid}, nil),
		createProcess: func(PID) (Process, error) {
			return nil, fmt.Errorf("createProcess error")
		},
	}
	require.NoError(t, acc.GatherError(p.Gather))
}

func TestGather_CreatePIDFinderError(t *testing.T) {
	var acc testutil.Accumulator

	p := Procstat{
		createPIDFinder: func() (PIDFinder, error) {
			return nil, fmt.Errorf("createPIDFinder error")
		},
		createProcess: newTestProc,
	}
	require.Error(t, acc.GatherError(p.Gather))
}

func TestGather_ProcessName(t *testing.T) {
	var acc testutil.Accumulator

	p := Procstat{
		Exe:             exe,
		ProcessName:     "custom_name",
		createPIDFinder: pidFinder([]PID{pid}, nil),
		createProcess:   newTestProc,
	}
	require.NoError(t, acc.GatherError(p.Gather))

	assert.Equal(t, "custom_name", acc.TagValue("procstat", "process_name"))
}

func TestGather_NoProcessNameUsesReal(t *testing.T) {
	var acc testutil.Accumulator
	pid := PID(os.Getpid())

	p := Procstat{
		Exe:             exe,
		createPIDFinder: pidFinder([]PID{pid}, nil),
		createProcess:   newTestProc,
	}
	require.NoError(t, acc.GatherError(p.Gather))

	assert.True(t, acc.HasTag("procstat", "process_name"))
}

func TestGather_NoPidTag(t *testing.T) {
	var acc testutil.Accumulator

	p := Procstat{
		Exe:             exe,
		createPIDFinder: pidFinder([]PID{pid}, nil),
		createProcess:   newTestProc,
	}
	require.NoError(t, acc.GatherError(p.Gather))
	assert.True(t, acc.HasInt32Field("procstat", "pid"))
	assert.False(t, acc.HasTag("procstat", "pid"))
}

func TestGather_PidTag(t *testing.T) {
	var acc testutil.Accumulator

	p := Procstat{
		Exe:             exe,
		PidTag:          true,
		createPIDFinder: pidFinder([]PID{pid}, nil),
		createProcess:   newTestProc,
	}
	require.NoError(t, acc.GatherError(p.Gather))
	assert.Equal(t, "42", acc.TagValue("procstat", "pid"))
	assert.False(t, acc.HasInt32Field("procstat", "pid"))
}

func TestGather_Prefix(t *testing.T) {
	var acc testutil.Accumulator

	p := Procstat{
		Exe:             exe,
		Prefix:          "custom_prefix",
		createPIDFinder: pidFinder([]PID{pid}, nil),
		createProcess:   newTestProc,
	}
	require.NoError(t, acc.GatherError(p.Gather))
	assert.True(t, acc.HasInt32Field("procstat", "custom_prefix_num_fds"))
}

func TestGather_Exe(t *testing.T) {
	var acc testutil.Accumulator

	p := Procstat{
		Exe:             exe,
		createPIDFinder: pidFinder([]PID{pid}, nil),
		createProcess:   newTestProc,
	}
	require.NoError(t, acc.GatherError(p.Gather))

	assert.Equal(t, exe, acc.TagValue("procstat", "exe"))
}

func TestGather_User(t *testing.T) {
	var acc testutil.Accumulator
	user := "ada"

	p := Procstat{
		User:            user,
		createPIDFinder: pidFinder([]PID{pid}, nil),
		createProcess:   newTestProc,
	}
	require.NoError(t, acc.GatherError(p.Gather))

	assert.Equal(t, user, acc.TagValue("procstat", "user"))
}

func TestGather_Pattern(t *testing.T) {
	var acc testutil.Accumulator
	pattern := "foo"

	p := Procstat{
		Pattern:         pattern,
		createPIDFinder: pidFinder([]PID{pid}, nil),
		createProcess:   newTestProc,
	}
	require.NoError(t, acc.GatherError(p.Gather))

	assert.Equal(t, pattern, acc.TagValue("procstat", "pattern"))
}

func TestGather_MissingPidMethod(t *testing.T) {
	var acc testutil.Accumulator

	p := Procstat{
		createPIDFinder: pidFinder([]PID{pid}, nil),
		createProcess:   newTestProc,
	}
	require.Error(t, acc.GatherError(p.Gather))
}

func TestGather_PidFile(t *testing.T) {
	var acc testutil.Accumulator
	pidfile := "/path/to/pidfile"

	p := Procstat{
		PidFile:         pidfile,
		createPIDFinder: pidFinder([]PID{pid}, nil),
		createProcess:   newTestProc,
	}
	require.NoError(t, acc.GatherError(p.Gather))

	assert.Equal(t, pidfile, acc.TagValue("procstat", "pidfile"))
}

func TestGather_PercentFirstPass(t *testing.T) {
	var acc testutil.Accumulator
	pid := PID(os.Getpid())

	p := Procstat{
		Pattern:         "foo",
		PidTag:          true,
		createPIDFinder: pidFinder([]PID{pid}, nil),
		createProcess:   NewProc,
	}
	require.NoError(t, acc.GatherError(p.Gather))

	assert.True(t, acc.HasFloatField("procstat", "cpu_time_user"))
	assert.False(t, acc.HasFloatField("procstat", "cpu_usage"))
}

func TestGather_PercentSecondPass(t *testing.T) {
	var acc testutil.Accumulator
	pid := PID(os.Getpid())

	p := Procstat{
		Pattern:         "foo",
		PidTag:          true,
		createPIDFinder: pidFinder([]PID{pid}, nil),
		createProcess:   NewProc,
	}
	require.NoError(t, acc.GatherError(p.Gather))
	require.NoError(t, acc.GatherError(p.Gather))

	assert.True(t, acc.HasFloatField("procstat", "cpu_time_user"))
	assert.True(t, acc.HasFloatField("procstat", "cpu_usage"))
}
