//go:build linux

package kernel

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func TestFullProcFile(t *testing.T) {
	k := Kernel{
		statFile:        "testdata/stat_file_full",
		entropyStatFile: "testdata/entropy_stat_file_full",
	}

	acc := testutil.Accumulator{}
	require.NoError(t, k.Gather(&acc))

	fields := map[string]interface{}{
		"boot_time":        int64(1457505775),
		"context_switches": int64(2626618),
		"disk_pages_in":    int64(5741),
		"disk_pages_out":   int64(1808),
		"interrupts":       int64(1472736),
		"processes_forked": int64(10673),
		"entropy_avail":    int64(1024),
	}
	acc.AssertContainsFields(t, "kernel", fields)
}

func TestPartialProcFile(t *testing.T) {
	k := Kernel{
		statFile:        "testdata/stat_file_partial",
		entropyStatFile: "testdata/entropy_stat_file_partial",
	}

	acc := testutil.Accumulator{}
	require.NoError(t, k.Gather(&acc))

	fields := map[string]interface{}{
		"boot_time":        int64(1457505775),
		"context_switches": int64(2626618),
		"disk_pages_in":    int64(5741),
		"disk_pages_out":   int64(1808),
		"interrupts":       int64(1472736),
		"entropy_avail":    int64(1024),
	}
	acc.AssertContainsFields(t, "kernel", fields)
}

func TestInvalidProcFile1(t *testing.T) {
	// missing btime measurement
	k := Kernel{
		statFile:        "testdata/stat_file_invalid",
		entropyStatFile: "testdata/entropy_stat_file_invalid",
	}

	acc := testutil.Accumulator{}
	err := k.Gather(&acc)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid syntax")
}

func TestInvalidProcFile2(t *testing.T) {
	// missing second page measurement
	k := Kernel{
		statFile: "testdata/stat_file_invalid2",
	}

	acc := testutil.Accumulator{}
	err := k.Gather(&acc)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no such file")
}

func TestNoProcFile(t *testing.T) {
	k := Kernel{
		statFile: "testdata/this_file_does_not_exist",
	}

	acc := testutil.Accumulator{}
	err := k.Gather(&acc)
	require.Error(t, err)
	require.Contains(t, err.Error(), "does not exist")
}
