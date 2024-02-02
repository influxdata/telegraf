package procstat

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/shirou/gopsutil/v3/process"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
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
func TestMockExecCommand(_ *testing.T) {
	var cmd []string //nolint:prealloc // Pre-allocated this slice would break the algorithm
	for _, arg := range os.Args {
		if arg == "--" {
			cmd = []string{}
			continue
		}
		if cmd == nil {
			continue
		}
		cmd = append(cmd, arg)
	}
	if cmd == nil {
		return
	}
	cmdline := strings.Join(cmd, " ")

	if cmdline == "systemctl show TestGather_systemdUnitPIDs" {
		fmt.Printf(`PIDFile=
GuessMainPID=yes
MainPID=11408
ControlPID=0
ExecMainPID=11408
`)
		//nolint:revive // error code is important for this "test"
		os.Exit(0)
	}

	if cmdline == "supervisorctl status TestGather_supervisorUnitPIDs" {
		fmt.Printf(`TestGather_supervisorUnitPIDs                             RUNNING   pid 7311, uptime 0:00:19
`)
		//nolint:revive // error code is important for this "test"
		os.Exit(0)
	}

	if cmdline == "supervisorctl status TestGather_STARTINGsupervisorUnitPIDs TestGather_FATALsupervisorUnitPIDs" {
		fmt.Printf(`TestGather_FATALsupervisorUnitPIDs                       FATAL     Exited too quickly (process log may have details)
TestGather_STARTINGsupervisorUnitPIDs                          STARTING`)
		//nolint:revive // error code is important for this "test"
		os.Exit(0)
	}

	fmt.Printf("command not found\n")
	//nolint:revive // error code is important for this "test"
	os.Exit(1)
}

type testPgrep struct {
	pids []PID
	err  error
}

func newTestFinder(pids []PID) PIDFinder {
	return &testPgrep{
		pids: pids,
		err:  nil,
	}
}

func (pg *testPgrep) PidFile(_ string) ([]PID, error) {
	return pg.pids, pg.err
}

func (p *testProc) Cmdline() (string, error) {
	return "test_proc", nil
}

func (pg *testPgrep) Pattern(_ string) ([]PID, error) {
	return pg.pids, pg.err
}

func (pg *testPgrep) UID(_ string) ([]PID, error) {
	return pg.pids, pg.err
}

func (pg *testPgrep) FullPattern(_ string) ([]PID, error) {
	return pg.pids, pg.err
}

func (pg *testPgrep) Children(_ PID) ([]PID, error) {
	pids := []PID{7311, 8111, 8112}
	return pids, pg.err
}

type testProc struct {
	pid  PID
	tags map[string]string
}

func newTestProc(pid PID) (Process, error) {
	proc := &testProc{
		pid:  pid,
		tags: make(map[string]string),
	}
	return proc, nil
}

func (p *testProc) PID() PID {
	return p.pid
}

func (p *testProc) Name() (string, error) {
	return "test_proc", nil
}

func (p *testProc) SetTag(k, v string) {
	p.tags[k] = v
}

func (p *testProc) MemoryMaps(bool) (*[]process.MemoryMapsStat, error) {
	return &[]process.MemoryMapsStat{}, nil
}

func (p *testProc) Metric(prefix string, tagging map[string]bool, _ bool) telegraf.Metric {
	if prefix != "" {
		prefix += "_"
	}

	fields := map[string]interface{}{
		prefix + "num_fds":                      int32(0),
		prefix + "num_threads":                  int32(0),
		prefix + "voluntary_context_switches":   int64(0),
		prefix + "involuntary_context_switches": int64(0),
		prefix + "minor_faults":                 uint64(0),
		prefix + "major_faults":                 uint64(0),
		prefix + "child_major_faults":           uint64(0),
		prefix + "child_minor_faults":           uint64(0),
		prefix + "read_bytes":                   uint64(0),
		prefix + "read_count":                   uint64(0),
		prefix + "write_bytes":                  uint64(0),
		prefix + "write_count":                  uint64(0),
		prefix + "created_at":                   int64(0),
		prefix + "cpu_time_user":                float64(0),
		prefix + "cpu_time_system":              float64(0),
		prefix + "cpu_time_iowait":              float64(0),
		prefix + "cpu_usage":                    float64(0),
		prefix + "memory_rss":                   uint64(0),
		prefix + "memory_vms":                   uint64(0),
		prefix + "memory_usage":                 float32(0),
	}

	tags := map[string]string{
		"process_name": "test_proc",
	}
	for k, v := range p.tags {
		tags[k] = v
	}

	// Add the tags as requested by the user
	if tagging["cmdline"] {
		tags["cmdline"] = "test_proc"
	} else {
		fields[prefix+"cmdline"] = "test_proc"
	}

	if tagging["pid"] {
		tags["pid"] = strconv.Itoa(int(p.pid))
	} else {
		fields["pid"] = int32(p.pid)
	}

	if tagging["ppid"] {
		tags["ppid"] = "0"
	} else {
		fields[prefix+"ppid"] = int32(0)
	}

	if tagging["status"] {
		tags["status"] = "running"
	} else {
		fields[prefix+"status"] = "running"
	}

	if tagging["user"] {
		tags["user"] = "testuser"
	} else {
		fields[prefix+"user"] = "testuser"
	}

	return metric.New("procstat", tags, fields, time.Time{})
}

var pid = PID(42)
var exe = "foo"

func TestInitInvalidFinder(t *testing.T) {
	plugin := Procstat{
		PidFinder:     "foo",
		Log:           testutil.Logger{},
		createProcess: newTestProc,
	}
	require.Error(t, plugin.Init())
}

func TestInitRequiresChildDarwin(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping test on non-darwin platform")
	}

	p := Procstat{
		Pattern:         "somepattern",
		SupervisorUnits: []string{"a_unit"},
		PidFinder:       "native",
		Log:             testutil.Logger{},
	}
	require.ErrorContains(t, p.Init(), "requires 'pgrep' finder")
}

func TestInitMissingPidMethod(t *testing.T) {
	p := Procstat{
		Log:           testutil.Logger{},
		createProcess: newTestProc,
	}
	require.ErrorContains(t, p.Init(), "require filter option but none set")
}

func TestGather_CreateProcessErrorOk(t *testing.T) {
	expected := []telegraf.Metric{
		testutil.MustMetric(
			"procstat_lookup",
			map[string]string{
				"exe":        "foo",
				"pid_finder": "test",
				"result":     "success",
			},
			map[string]interface{}{
				"pid_count":   int64(1),
				"result_code": int64(0),
				"running":     int64(0),
			},
			time.Unix(0, 0),
			telegraf.Untyped,
		),
	}

	p := Procstat{
		Exe:       exe,
		PidFinder: "test",
		Log:       testutil.Logger{},
		finder:    newTestFinder([]PID{pid}),
		createProcess: func(PID) (Process, error) {
			return nil, fmt.Errorf("createProcess error")
		},
	}
	require.NoError(t, p.Init())

	var acc testutil.Accumulator
	require.NoError(t, p.Gather(&acc))
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestGather_ProcessName(t *testing.T) {
	expected := []telegraf.Metric{
		testutil.MustMetric(
			"procstat",
			map[string]string{
				"exe":          "foo",
				"process_name": "custom_name",
			},
			map[string]interface{}{
				"child_major_faults":           uint64(0),
				"child_minor_faults":           uint64(0),
				"cmdline":                      "test_proc",
				"cpu_time_iowait":              float64(0),
				"cpu_time_system":              float64(0),
				"cpu_time_user":                float64(0),
				"cpu_usage":                    float64(0),
				"created_at":                   int64(0),
				"involuntary_context_switches": int64(0),
				"major_faults":                 uint64(0),
				"memory_rss":                   uint64(0),
				"memory_usage":                 float32(0),
				"memory_vms":                   uint64(0),
				"minor_faults":                 uint64(0),
				"num_fds":                      int32(0),
				"num_threads":                  int32(0),
				"pid":                          int32(42),
				"ppid":                         int32(0),
				"read_bytes":                   uint64(0),
				"read_count":                   uint64(0),
				"status":                       "running",
				"user":                         "testuser",
				"voluntary_context_switches":   int64(0),
				"write_bytes":                  uint64(0),
				"write_count":                  uint64(0),
			},
			time.Unix(0, 0),
			telegraf.Untyped,
		),
		testutil.MustMetric(
			"procstat_lookup",
			map[string]string{
				"exe":        "foo",
				"pid_finder": "test",
				"result":     "success",
			},
			map[string]interface{}{
				"pid_count":   int64(1),
				"result_code": int64(0),
				"running":     int64(1),
			},
			time.Unix(0, 0),
			telegraf.Untyped,
		),
	}

	p := Procstat{
		Exe:           exe,
		ProcessName:   "custom_name",
		PidFinder:     "test",
		Log:           testutil.Logger{},
		finder:        newTestFinder([]PID{pid}),
		createProcess: newTestProc,
	}
	require.NoError(t, p.Init())

	var acc testutil.Accumulator
	require.NoError(t, p.Gather(&acc))

	testutil.PrintMetrics(acc.GetTelegrafMetrics())
	require.Equal(t, "custom_name", acc.TagValue("procstat", "process_name"))
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestGather_NoProcessNameUsesReal(t *testing.T) {
	pid := PID(os.Getpid())

	p := Procstat{
		Exe:           exe,
		PidFinder:     "test",
		Log:           testutil.Logger{},
		finder:        newTestFinder([]PID{pid}),
		createProcess: newTestProc,
	}
	require.NoError(t, p.Init())

	var acc testutil.Accumulator
	require.NoError(t, p.Gather(&acc))

	require.True(t, acc.HasTag("procstat", "process_name"))
}

func TestGather_NoPidTag(t *testing.T) {
	p := Procstat{
		Exe:           exe,
		PidFinder:     "test",
		Log:           testutil.Logger{},
		finder:        newTestFinder([]PID{pid}),
		createProcess: newTestProc,
	}
	require.NoError(t, p.Init())

	var acc testutil.Accumulator
	require.NoError(t, p.Gather(&acc))

	require.True(t, acc.HasInt64Field("procstat", "pid"))
	require.False(t, acc.HasTag("procstat", "pid"))
}

func TestGather_PidTag(t *testing.T) {
	p := Procstat{
		Exe:           exe,
		PidTag:        true,
		PidFinder:     "test",
		Log:           testutil.Logger{},
		finder:        newTestFinder([]PID{pid}),
		createProcess: newTestProc,
	}
	require.NoError(t, p.Init())

	var acc testutil.Accumulator
	require.NoError(t, p.Gather(&acc))

	require.Equal(t, "42", acc.TagValue("procstat", "pid"))
	require.False(t, acc.HasInt32Field("procstat", "pid"))
}

func TestGather_Prefix(t *testing.T) {
	p := Procstat{
		Exe:           exe,
		Prefix:        "custom_prefix",
		PidFinder:     "test",
		Log:           testutil.Logger{},
		finder:        newTestFinder([]PID{pid}),
		createProcess: newTestProc,
	}
	require.NoError(t, p.Init())

	var acc testutil.Accumulator
	require.NoError(t, p.Gather(&acc))

	require.True(t, acc.HasInt64Field("procstat", "custom_prefix_num_fds"))
}

func TestGather_Exe(t *testing.T) {
	p := Procstat{
		Exe:           exe,
		PidFinder:     "test",
		Log:           testutil.Logger{},
		finder:        newTestFinder([]PID{pid}),
		createProcess: newTestProc,
	}
	require.NoError(t, p.Init())

	var acc testutil.Accumulator
	require.NoError(t, p.Gather(&acc))

	require.Equal(t, exe, acc.TagValue("procstat", "exe"))
}

func TestGather_User(t *testing.T) {
	user := "ada"

	p := Procstat{
		User:          user,
		PidFinder:     "test",
		Log:           testutil.Logger{},
		finder:        newTestFinder([]PID{pid}),
		createProcess: newTestProc,
	}
	require.NoError(t, p.Init())

	var acc testutil.Accumulator
	require.NoError(t, p.Gather(&acc))

	require.Equal(t, user, acc.TagValue("procstat", "user"))
}

func TestGather_Pattern(t *testing.T) {
	pattern := "foo"

	p := Procstat{
		Pattern:       pattern,
		PidFinder:     "test",
		Log:           testutil.Logger{},
		finder:        newTestFinder([]PID{pid}),
		createProcess: newTestProc,
	}
	require.NoError(t, p.Init())

	var acc testutil.Accumulator
	require.NoError(t, p.Gather(&acc))

	require.Equal(t, pattern, acc.TagValue("procstat", "pattern"))
}

func TestGather_PidFile(t *testing.T) {
	pidfile := "/path/to/pidfile"

	p := Procstat{
		PidFile:       pidfile,
		PidFinder:     "test",
		Log:           testutil.Logger{},
		finder:        newTestFinder([]PID{pid}),
		createProcess: newTestProc,
	}
	require.NoError(t, p.Init())

	var acc testutil.Accumulator
	require.NoError(t, p.Gather(&acc))

	require.Equal(t, pidfile, acc.TagValue("procstat", "pidfile"))
}

func TestGather_PercentFirstPass(t *testing.T) {
	pid := PID(os.Getpid())

	p := Procstat{
		Pattern:       "foo",
		PidTag:        true,
		PidFinder:     "test",
		Log:           testutil.Logger{},
		finder:        newTestFinder([]PID{pid}),
		createProcess: newProc,
	}
	require.NoError(t, p.Init())

	var acc testutil.Accumulator
	require.NoError(t, p.Gather(&acc))

	require.True(t, acc.HasFloatField("procstat", "cpu_time_user"))
	require.False(t, acc.HasFloatField("procstat", "cpu_usage"))
}

func TestGather_PercentSecondPass(t *testing.T) {
	pid := PID(os.Getpid())

	p := Procstat{
		Pattern:       "foo",
		PidTag:        true,
		PidFinder:     "test",
		Log:           testutil.Logger{},
		finder:        newTestFinder([]PID{pid}),
		createProcess: newProc,
	}
	require.NoError(t, p.Init())

	var acc testutil.Accumulator
	require.NoError(t, p.Gather(&acc))
	require.NoError(t, p.Gather(&acc))

	require.True(t, acc.HasFloatField("procstat", "cpu_time_user"))
	require.True(t, acc.HasFloatField("procstat", "cpu_usage"))
}

func TestGather_systemdUnitPIDs(t *testing.T) {
	p := Procstat{
		SystemdUnit: "TestGather_systemdUnitPIDs",
		PidFinder:   "test",
		Log:         testutil.Logger{},
		finder:      newTestFinder([]PID{pid}),
	}
	require.NoError(t, p.Init())

	pidsTags, err := p.findPids()
	require.NoError(t, err)

	for _, pidsTag := range pidsTags {
		require.Equal(t, []PID{11408}, pidsTag.PIDs)
		require.Equal(t, "TestGather_systemdUnitPIDs", pidsTag.Tags["systemd_unit"])
	}
}

func TestGather_cgroupPIDs(t *testing.T) {
	//no cgroups in windows
	if runtime.GOOS == "windows" {
		t.Skip("no cgroups in windows")
	}
	td := t.TempDir()
	err := os.WriteFile(filepath.Join(td, "cgroup.procs"), []byte("1234\n5678\n"), 0640)
	require.NoError(t, err)

	p := Procstat{
		CGroup:    td,
		PidFinder: "test",
		Log:       testutil.Logger{},
		finder:    newTestFinder([]PID{pid}),
	}
	require.NoError(t, p.Init())

	pidsTags, err := p.findPids()
	require.NoError(t, err)
	for _, pidsTag := range pidsTags {
		require.Equal(t, []PID{1234, 5678}, pidsTag.PIDs)
		require.Equal(t, td, pidsTag.Tags["cgroup"])
	}
}

func TestProcstatLookupMetric(t *testing.T) {
	p := Procstat{
		Exe:           "-Gsys",
		PidFinder:     "test",
		Log:           testutil.Logger{},
		finder:        newTestFinder([]PID{543}),
		createProcess: newProc,
	}
	require.NoError(t, p.Init())

	var acc testutil.Accumulator
	require.NoError(t, p.Gather(&acc))
	require.Len(t, acc.GetTelegrafMetrics(), 1)
}

func TestGather_SameTimestamps(t *testing.T) {
	pidfile := "/path/to/pidfile"

	p := Procstat{
		PidFile:       pidfile,
		PidFinder:     "test",
		Log:           testutil.Logger{},
		finder:        newTestFinder([]PID{pid}),
		createProcess: newTestProc,
	}
	require.NoError(t, p.Init())

	var acc testutil.Accumulator
	require.NoError(t, p.Gather(&acc))

	procstat, _ := acc.Get("procstat")
	procstatLookup, _ := acc.Get("procstat_lookup")

	require.Equal(t, procstat.Time, procstatLookup.Time)
}

func TestGather_supervisorUnitPIDs(t *testing.T) {
	p := Procstat{
		SupervisorUnits: []string{"TestGather_supervisorUnitPIDs"},
		PidFinder:       "test",
		Log:             testutil.Logger{},
		finder:          newTestFinder([]PID{pid}),
	}
	require.NoError(t, p.Init())

	pidsTags, err := p.findPids()
	require.NoError(t, err)
	for _, pidsTag := range pidsTags {
		require.Equal(t, []PID{7311, 8111, 8112}, pidsTag.PIDs)
		require.Equal(t, "TestGather_supervisorUnitPIDs", pidsTag.Tags["supervisor_unit"])
	}
}

func TestGather_MoresupervisorUnitPIDs(t *testing.T) {
	p := Procstat{
		SupervisorUnits: []string{"TestGather_STARTINGsupervisorUnitPIDs", "TestGather_FATALsupervisorUnitPIDs"},
		PidFinder:       "test",
		Log:             testutil.Logger{},
		finder:          newTestFinder([]PID{pid}),
	}
	require.NoError(t, p.Init())

	pidsTags, err := p.findPids()
	require.NoError(t, err)
	for _, pidsTag := range pidsTags {
		require.Empty(t, pidsTag.PIDs)
		switch pidsTag.Tags["supervisor_unit"] {
		case "TestGather_STARTINGsupervisorUnitPIDs":
			require.Equal(t, "STARTING", pidsTag.Tags["status"])
		case "TestGather_FATALsupervisorUnitPIDs":
			require.Equal(t, "FATAL", pidsTag.Tags["status"])
			require.Equal(t, "Exited too quickly (process log may have details)", pidsTag.Tags["error"])
		default:
			t.Fatalf("unexpected value for tag 'supervisor_unit': %q", pidsTag.Tags["supervisor_unit"])
		}
	}
}
