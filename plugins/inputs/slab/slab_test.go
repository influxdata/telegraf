//go:build linux
// +build linux

package slab

import (
	"os"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func makeFakeStatFile(content []byte) string {
	tmpfile, err := os.CreateTemp("", "slab_test")
	if err != nil {
		panic(err)
	}

	if _, err := tmpfile.Write(content); err != nil {
		panic(err)
	}
	if err := tmpfile.Close(); err != nil {
		panic(err)
	}

	return tmpfile.Name()
}

func TestSlab(t *testing.T) {
	slabStats := SlabStats{
		statFile: makeFakeStatFile([]byte(procSlabInfo)),
	}

	var acc testutil.Accumulator
	err := acc.GatherError(slabStats.Gather)
	require.NoError(t, err)

	fields := map[string]interface{}{
		"ext4_allocation_context_size_in_bytes": int(16384),
		"ext4_extent_status_size_in_bytes":      int(8160),
		"ext4_free_data_size_in_bytes":          int(0),
		"ext4_inode_cache_size_in_bytes":        int(491520),
		"ext4_io_end_size_in_bytes":             int(4032),
		"ext4_xattr_size_in_bytes":              int(0),
		"kmalloc-1024_size_in_bytes":            int(239927296),
		"kmalloc-128_size_in_bytes":             int(5586944),
		"kmalloc-16_size_in_bytes":              int(17002496),
		"kmalloc-192_size_in_bytes":             int(4015872),
		"kmalloc-2048_size_in_bytes":            int(3309568),
		"kmalloc-256_size_in_bytes":             int(5423104),
		"kmalloc-32_size_in_bytes":              int(3657728),
		"kmalloc-4096_size_in_bytes":            int(2359296),
		"kmalloc-512_size_in_bytes":             int(41435136),
		"kmalloc-64_size_in_bytes":              int(8536064),
		"kmalloc-8_size_in_bytes":               int(229376),
		"kmalloc-8192_size_in_bytes":            int(1048576),
		"kmalloc-96_size_in_bytes":              int(12378240),
		"kmem_cache_size_in_bytes":              int(81920),
		"kmem_cache_node_size_in_bytes":         int(36864),
	}

	acc.AssertContainsFields(t, "slab", fields)
}

var procSlabInfo = `slabinfo - version: 2.1
# name            <active_objs> <num_objs> <objsize> <objperslab> <pagesperslab> : tunables <limit> <batchcount> <sharedfactor> : slabdata <active_slabs> <num_slabs> <sharedavail>
ext4_inode_cache     480    480   1024   32    8 : tunables    0    0    0 : slabdata     15     15      0
ext4_xattr             0      0     88   46    1 : tunables    0    0    0 : slabdata      0      0      0
ext4_free_data         0      0     64   64    1 : tunables    0    0    0 : slabdata      0      0      0
ext4_allocation_context    128    128    128   32    1 : tunables    0    0    0 : slabdata      4      4      0
ext4_io_end           56     56     72   56    1 : tunables    0    0    0 : slabdata      1      1      0
ext4_extent_status    204    204     40  102    1 : tunables    0    0    0 : slabdata      2      2      0
kmalloc-8192         106    128   8192    4    8 : tunables    0    0    0 : slabdata     32     32      0
kmalloc-4096         486    576   4096    8    8 : tunables    0    0    0 : slabdata     72     72      0
kmalloc-2048        1338   1616   2048   16    8 : tunables    0    0    0 : slabdata    101    101      0
kmalloc-1024      155845 234304   1024   32    8 : tunables    0    0    0 : slabdata   7329   7329      0
kmalloc-512        18995  80928    512   32    4 : tunables    0    0    0 : slabdata   2529   2529      0
kmalloc-256        16366  21184    256   32    2 : tunables    0    0    0 : slabdata    662    662      0
kmalloc-192        18835  20916    192   21    1 : tunables    0    0    0 : slabdata    996    996      0
kmalloc-128        23600  43648    128   32    1 : tunables    0    0    0 : slabdata   1364   1364      0
kmalloc-96         95106 128940     96   42    1 : tunables    0    0    0 : slabdata   3070   3070      0
kmalloc-64         82432 133376     64   64    1 : tunables    0    0    0 : slabdata   2084   2084      0
kmalloc-32         78477 114304     32  128    1 : tunables    0    0    0 : slabdata    893    893      0
kmalloc-16        885605 1062656     16  256    1 : tunables    0    0    0 : slabdata   4151   4151      0
kmalloc-8          28672  28672      8  512    1 : tunables    0    0    0 : slabdata     56     56      0
kmem_cache_node      576    576     64   64    1 : tunables    0    0    0 : slabdata      9      9      0
kmem_cache           320    320    256   32    2 : tunables    0    0    0 : slabdata     10     10      0
`
