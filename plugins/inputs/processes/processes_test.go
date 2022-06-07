//go:build !windows
// +build !windows

package processes

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

func TestProcesses(t *testing.T) {
	tester := tester{}
	processes := &Processes{
		Log: testutil.Logger{},
		execPS: testExecPS("STAT\n		Ss  \n		S   \n		Z   \n		R   \n		S<  \n		SNs \n		Ss+ \n		\n		\n"),
		readProcFile: tester.testProcFile,
	}
	var acc testutil.Accumulator

	err := processes.Gather(&acc)
	require.NoError(t, err)

	require.True(t, acc.HasInt64Field("processes", "running"))
	require.True(t, acc.HasInt64Field("processes", "sleeping"))
	require.True(t, acc.HasInt64Field("processes", "stopped"))
	require.True(t, acc.HasInt64Field("processes", "total"))
	total, ok := acc.Get("processes")
	require.True(t, ok)
	require.True(t, total.Fields["total"].(int64) > 0)
}

func TestFromPS(t *testing.T) {
	processes := &Processes{
		Log:     testutil.Logger{},
		execPS:  testExecPS("\nSTAT\nD\nI\nL\nR\nR+\nS\nS+\nSNs\nSs\nU\nZ\n"),
		forcePS: true,
	}

	var acc testutil.Accumulator
	err := processes.Gather(&acc)
	require.NoError(t, err)

	fields := getEmptyFields()
	fields["blocked"] = int64(3)
	fields["zombies"] = int64(1)
	fields["running"] = int64(2)
	fields["sleeping"] = int64(4)
	fields["idle"] = int64(1)
	fields["total"] = int64(11)

	acc.AssertContainsTaggedFields(t, "processes", fields, map[string]string{})
}

func TestFromPSError(t *testing.T) {
	processes := &Processes{
		Log:     testutil.Logger{},
		execPS:  testExecPSError,
		forcePS: true,
	}

	var acc testutil.Accumulator
	err := processes.Gather(&acc)
	require.Error(t, err)
}

func TestFromProcFiles(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("This test only runs on linux")
	}
	tester := tester{}
	processes := &Processes{
		Log:          testutil.Logger{},
		readProcFile: tester.testProcFile,
		forceProc:    true,
	}

	var acc testutil.Accumulator
	err := processes.Gather(&acc)
	require.NoError(t, err)

	fields := getEmptyFields()
	fields["sleeping"] = tester.calls
	fields["total_threads"] = tester.calls * 2
	fields["total"] = tester.calls

	acc.AssertContainsTaggedFields(t, "processes", fields, map[string]string{})
}

func TestFromProcFilesWithSpaceInCmd(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("This test only runs on linux")
	}
	tester := tester{}
	processes := &Processes{
		Log:          testutil.Logger{},
		readProcFile: tester.testProcFile2,
		forceProc:    true,
	}

	var acc testutil.Accumulator
	err := processes.Gather(&acc)
	require.NoError(t, err)

	fields := getEmptyFields()
	fields["sleeping"] = tester.calls
	fields["total_threads"] = tester.calls * 2
	fields["total"] = tester.calls

	acc.AssertContainsTaggedFields(t, "processes", fields, map[string]string{})
}

// Based on `man 5 proc`, parked processes an be found in a
// limited range of Linux versions:
//
// >    P  Parked (Linux 3.9 to 3.13 only)
//
// However, we have had reports of this process state on Ubuntu
// Bionic w/ Linux 4.15 (#6270)
func TestParkedProcess(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Parked process test only relevant on linux")
	}
	procstat := `88 (watchdog/13) P 2 0 0 0 -1 69238848 0 0 0 0 0 0 0 0 20 0 1 0 20 0 0 18446744073709551615 0 0 0 0 0 0 0 2147483647 0 1 0 0 17 0 0 0 0 0 0 0 0 0 0 0 0 0 0
`
	plugin := &Processes{
		Log: testutil.Logger{},
		readProcFile: func(string) ([]byte, error) {
			return []byte(procstat), nil
		},
		forceProc: true,
	}

	var acc testutil.Accumulator
	err := plugin.Gather(&acc)
	require.NoError(t, err)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"processes",
			map[string]string{},
			map[string]interface{}{
				"blocked":  0,
				"dead":     0,
				"idle":     0,
				"paging":   0,
				"parked":   1,
				"running":  0,
				"sleeping": 0,
				"stopped":  0,
				"unknown":  0,
				"zombies":  0,
			},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
	}
	actual := acc.GetTelegrafMetrics()
	for _, a := range actual {
		a.RemoveField("total")
		a.RemoveField("total_threads")
	}
	testutil.RequireMetricsEqual(t, expected, actual,
		testutil.IgnoreTime())
}

func testExecPS(out string) func() ([]byte, error) {
	return func() ([]byte, error) { return []byte(out), nil }
}

// struct for counting calls to testProcFile
type tester struct {
	calls int64
}

func (t *tester) testProcFile(_ string) ([]byte, error) {
	t.calls++
	return []byte(fmt.Sprintf(testProcStat, "S", "2")), nil
}

func (t *tester) testProcFile2(_ string) ([]byte, error) {
	t.calls++
	return []byte(fmt.Sprintf(testProcStat2, "S", "2")), nil
}

func testExecPSError() ([]byte, error) {
	return []byte("\nSTAT\nD\nI\nL\nR\nR+\nS\nS+\nSNs\nSs\nU\nZ\n"), fmt.Errorf("error")
}

const testProcStat = `10 (rcuob/0) %s 2 0 0 0 -1 2129984 0 0 0 0 0 0 0 0 20 0 %s 0 11 0 0 18446744073709551615 0 0 0 0 0 0 0 2147483647 0 18446744073709551615 0 0 17 0 0 0 0 0 0 0 0 0 0 0 0 0 0
`

const testProcStat2 = `10 (rcuob 0) %s 2 0 0 0 -1 2129984 0 0 0 0 0 0 0 0 20 0 %s 0 11 0 0 18446744073709551615 0 0 0 0 0 0 0 2147483647 0 18446744073709551615 0 0 17 0 0 0 0 0 0 0 0 0 0 0 0 0 0
`
