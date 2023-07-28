//go:build linux

package kernel_ksm

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func TestFullDataSysDir(t *testing.T) {
	k := KernelKsm{
		sysPath: "testdata/valid",
	}

	acc := testutil.Accumulator{}
	require.NoError(t, k.Gather(&acc))

	fields := map[string]interface{}{

		"full_scans":                         int64(123),
		"max_page_sharing":                   int64(10000),
		"merge_across_nodes":                 int64(1),
		"pages_shared":                       int64(12922),
		"pages_sharing":                      int64(28384),
		"pages_to_scan":                      int64(12928),
		"pages_unshared":                     int64(92847),
		"pages_volatile":                     int64(2824171),
		"run":                                int64(1),
		"sleep_millisecs":                    int64(1000),
		"stable_node_chains":                 int64(0),
		"stable_node_chains_prune_millisecs": int64(0),
		"stable_node_dups":                   int64(0),
		"use_zero_pages":                     int64(1),
	}
	acc.AssertContainsFields(t, "kernel_ksm", fields)
}

func TestInvalidDataSysDir(t *testing.T) {
	k := KernelKsm{
		sysPath: "testdata/invalid",
	}

	acc := testutil.Accumulator{}
	err := k.Gather(&acc)

	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid syntax")
}

func TestMissingFile(t *testing.T) {
	k := KernelKsm{
		sysPath: "testdata/missing",
	}

	acc := testutil.Accumulator{}
	err := k.Gather(&acc)

	require.Error(t, err)
	require.Contains(t, err.Error(), "does not exist")
}

func TestNoSysDir(t *testing.T) {
	k := KernelKsm{
		sysPath: "testdata/this_directory_does_not_exist",
	}

	acc := testutil.Accumulator{}
	err := k.Gather(&acc)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Is KSM included in the kernel?")
}
