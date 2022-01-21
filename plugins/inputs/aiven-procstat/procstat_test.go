package aiven_procstat

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	execCommand = mockExecCommand
}
func mockExecCommand(arg0 string, args ...string) *exec.Cmd {
	args = append([]string{"-test.run=TestMockExecCommand", "--", arg0}, args...)
	cmd := exec.Command(os.Args[0], args...)
	cmd.Stderr = os.Stderr
	return cmd
}
func TestMockExecCommand(t *testing.T) {
	var cmd []string
	for _, arg := range os.Args {
		if string(arg) == "--" {
			cmd = []string{}
			continue
		}
		if cmd == nil {
			continue
		}
		cmd = append(cmd, string(arg))
	}
	if cmd == nil {
		return
	}
	cmdline := strings.Join(cmd, " ")

	if cmdline == "systemctl status" {
		fmt.Printf(`● hostname
    State: degraded
     Jobs: 0 queued
   Failed: 1 units
    Since: Fri 2019-09-13 06:08:15 UTC; 50min ago
   CGroup: /
           ├─user.slice
           │ └─user-1002.slice
           │   ├─user@1002.service
           │   │ └─init.scope
           │   │   ├─3790 /usr/lib/systemd/systemd --user
           │   │   └─3792 (sd-pam)
           │   └─session-5.scope
           │     ├─3819 sshd: root [priv]
           │     ├─3822 sshd: root@pts/1
           │     ├─3823 -bash
           │     ├─3858 sudo -i
           │     ├─3860 -bash
           │     ├─7445 systemctl status
           │     └─7446 less
           ├─init.scope
           │ └─1 /usr/lib/systemd/systemd --switched-root --system --deserialize 31
           ├─system.slice
           │ ├─systemd-udevd.service
           │ │ └─455 /usr/lib/systemd/systemd-udevd
           │ ├─dbus-broker.service
           │ │ ├─538 /usr/bin/dbus-broker-launch --scope system --audit
           │ │ └─539 dbus-broker --log 4 --controller 9 --machine-id 5305bf75e3144e22a1c4d297f26ec42d --max-bytes 536870912 --max-fds 4096 --max-matches 131072 --audit
           │ ├─system-serial\x2dgetty.slice
           │ │ └─serial-getty@ttyS0.service
           │ │   └─739 /sbin/agetty -o -p -- \u --keep-baud 115200,38400,9600 ttyS0 vt220
           │ ├─TestGather_systemdUnitPIDs.service
           │ │ └─11408 /usr/bin/foo
           │ │ └─11420 /usr/bin/bar
           │ ├─TestTrailingSpaces_systemdUnitPIDs.service 	
           │ │ └─11428 /usr/bin/foo
           │ ├─chronyd.service
           │ │ └─1931 /usr/sbin/chronyd
           └─machine.slice
             └─machine-foo\x2dmgmt\x2d1.scope
               └─payload
                 ├─init.scope
                 │ └─2334 /usr/lib/systemd/systemd
                 └─system.slice
                   └─foo.service
					 └─2371 /bin/python3 -m aiven.almond.almond
`)
		os.Exit(0)
	}

	fmt.Printf("command not found\n")
	os.Exit(1)
}

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

func (p *testProc) Cmdline() (string, error) {
	return "test_proc", nil
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

func (p *testProc) Username() (string, error) {
	return "testuser", nil
}

func (p *testProc) Tags() map[string]string {
	return p.tags
}

func (p *testProc) PageFaults() (*process.PageFaultsStat, error) {
	return &process.PageFaultsStat{}, nil
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

func (p *testProc) MemoryPercent() (float32, error) {
	return 0, nil
}

func (p *testProc) CreateTime() (int64, error) {
	return 0, nil
}

func (p *testProc) Times() (*cpu.TimesStat, error) {
	return &cpu.TimesStat{}, nil
}

func (p *testProc) RlimitUsage(gatherUsage bool) ([]process.RlimitStat, error) {
	return []process.RlimitStat{}, nil
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

func TestGather_systemdUnitPIDs(t *testing.T) {
	p := Procstat{
		createPIDFinder: pidFinder([]PID{}, nil),
		SystemdUnit:     "TestGather_systemdUnitPIDs",
	}
	pidsArray, tagsArray, err := p.findPids()
	require.NoError(t, err)
	assert.Equal(t, []PID{11408, 11420}, pidsArray[0])
	assert.Equal(t, "TestGather_systemdUnitPIDs", tagsArray[0]["systemd_unit"])

	p = Procstat{
		createPIDFinder: pidFinder([]PID{}, nil),
		SystemdUnits:    []string{"TestGather_systemdUnitPIDs", "foo.service"},
	}
	pidsArray, tagsArray, err = p.findPids()
	require.NoError(t, err)
	assert.Equal(t, []PID{11408, 11420}, pidsArray[0])
	assert.Equal(t, "TestGather_systemdUnitPIDs", tagsArray[0]["systemd_unit"])
	assert.Equal(t, []PID{2371}, pidsArray[1])
	assert.Equal(t, "foo.service", tagsArray[1]["systemd_unit"])
}

func TestTrailingSpaces_systemdUnitPIDs(t *testing.T) {
	p := Procstat{
		createPIDFinder: pidFinder([]PID{}, nil),
		SystemdUnit:     "TestTrailingSpaces_systemdUnitPIDs",
	}
	pidsArray, tagsArray, err := p.findPids()
	require.NoError(t, err)
	assert.Equal(t, []PID{11428}, pidsArray[0])
	assert.Equal(t, "TestTrailingSpaces_systemdUnitPIDs", tagsArray[0]["systemd_unit"])
}

func TestGather_cgroupPIDs(t *testing.T) {
	//no cgroups in windows
	if runtime.GOOS == "windows" {
		t.Skip("no cgroups in windows")
	}
	td, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(td)
	err = ioutil.WriteFile(filepath.Join(td, "cgroup.procs"), []byte("1234\n5678\n"), 0644)
	require.NoError(t, err)

	p := Procstat{
		createPIDFinder: pidFinder([]PID{}, nil),
		CGroup:          td,
	}
	pidsArray, tagsArray, err := p.findPids()
	require.NoError(t, err)
	assert.Equal(t, []PID{1234, 5678}, pidsArray[0])
	assert.Equal(t, td, tagsArray[0]["cgroup"])
}

func TestProcstatLookupMetric(t *testing.T) {
	p := Procstat{
		createPIDFinder: pidFinder([]PID{543}, nil),
		Exe:             "-Gsys",
	}
	var acc testutil.Accumulator
	err := acc.GatherError(p.Gather)
	require.NoError(t, err)
	require.Equal(t, len(p.procs)+1, len(acc.Metrics))
}
