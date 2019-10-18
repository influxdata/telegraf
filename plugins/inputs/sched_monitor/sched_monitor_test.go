package sched_monitor

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

const sampleSchedFile = `migration (10, #threads: 1)
-------------------------------------------------------------------
se.exec_start                                :      43042098.108342
se.vruntime                                  :             0.000000
se.sum_exec_runtime                          :            87.654321
se.nr_migrations                             :                    0
nr_switches                                  :                   42
nr_voluntary_switches                        :                   32
nr_involuntary_switches                      :                   10
se.load.weight                               :              1048576
se.runnable_weight                           :              1048576
`

func TestParseSchedStats(t *testing.T) {
	f := bytes.NewBufferString(sampleSchedFile)
	cmd, cpuTime, ctxSwtch, invCtxSwtch := parseSchedStats(f)

	require.Equal(t, "migration", cmd)
	require.Equal(t, int64(87654321), cpuTime)
	require.Equal(t, int64(32), ctxSwtch)
	require.Equal(t, int64(10), invCtxSwtch)
}

func TestDecodeTaskPath(t *testing.T) {
	pid, tid := decodeTaskPath("/proc/123/task/456")

	require.Equal(t, 123, pid)
	require.Equal(t, 456, tid)
}

func TestParseCpuList(t *testing.T) {
	parseCPUList("1,2,5-8,10,12-15")

	expected := map[int]bool{1: true, 2: true, 5: true, 6: true, 7: true, 8: true, 10: true, 12: true, 13: true, 14: true, 15: true}
	require.Equal(t, expected, monitoredCPUS)
}

func TestIsKernelTask(t *testing.T) {
	require.True(t, isKernelTask("[migration]"))
}
