package zfs

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/influxdb/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const arcstatsContents = `5 1 0x01 86 4128 23617128247 12081618582809582
name                            type data
hits                            4    5968846374
misses                          4    1659178751
demand_data_hits                4    4860247322
demand_data_misses              4    501499535
demand_metadata_hits            4    708608325
demand_metadata_misses          4    156591375
prefetch_data_hits              4    367047144
prefetch_data_misses            4    974529898
prefetch_metadata_hits          4    32943583
prefetch_metadata_misses        4    26557943
mru_hits                        4    301176811
mru_ghost_hits                  4    47066067
mfu_hits                        4    5520612438
mfu_ghost_hits                  4    45784009
deleted                         4    1718937704
recycle_miss                    4    481222994
mutex_miss                      4    20575623
evict_skip                      4    14655903906543
evict_l2_cached                 4    145310202998272
evict_l2_eligible               4    16345402777088
evict_l2_ineligible             4    7437226893312
hash_elements                   4    36617980
hash_elements_max               4    36618318
hash_collisions                 4    554145157
hash_chains                     4    4187651
hash_chain_max                  4    26
p                               4    13963222064
c                               4    16381258376
c_min                           4    4194304
c_max                           4    16884125696
size                            4    16319887096
hdr_size                        4    42567864
data_size                       4    60066304
meta_size                       4    1701534208
other_size                      4    1661543168
anon_size                       4    94720
anon_evict_data                 4    0
anon_evict_metadata             4    0
mru_size                        4    973099008
mru_evict_data                  4    9175040
mru_evict_metadata              4    32768
mru_ghost_size                  4    32768
mru_ghost_evict_data            4    0
mru_ghost_evict_metadata        4    32768
mfu_size                        4    788406784
mfu_evict_data                  4    50881024
mfu_evict_metadata              4    81920
mfu_ghost_size                  4    0
mfu_ghost_evict_data            4    0
mfu_ghost_evict_metadata        4    0
l2_hits                         4    573868618
l2_misses                       4    1085309718
l2_feeds                        4    12182087
l2_rw_clash                     4    9610
l2_read_bytes                   4    32695938336768
l2_write_bytes                  4    2826774778880
l2_writes_sent                  4    4267687
l2_writes_done                  4    4267687
l2_writes_error                 4    0
l2_writes_hdr_miss              4    164
l2_evict_lock_retry             4    5
l2_evict_reading                4    0
l2_free_on_write                4    1606914
l2_cdata_free_on_write          4    1775
l2_abort_lowmem                 4    83462
l2_cksum_bad                    4    393860640
l2_io_error                     4    53881460
l2_size                         4    2471466648576
l2_asize                        4    2461690072064
l2_hdr_size                     4    12854175552
l2_compress_successes           4    12184849
l2_compress_zeros               4    0
l2_compress_failures            4    0
memory_throttle_count           4    0
duplicate_buffers               4    0
duplicate_buffers_size          4    0
duplicate_reads                 4    0
memory_direct_count             4    5159942
memory_indirect_count           4    3034640
arc_no_grow                     4    0
arc_tempreserve                 4    0
arc_loaned_bytes                4    0
arc_prune                       4    114554259559
arc_meta_used                   4    16259820792
arc_meta_limit                  4    12663094272
arc_meta_max                    4    18327165696
`

const zfetchstatsContents = `3 1 0x01 11 528 23607270446 12081656848148208
name                            type data
hits                            4    7812959060
misses                          4    4154484207
colinear_hits                   4    1366368
colinear_misses                 4    4153117839
stride_hits                     4    7309776732
stride_misses                   4    222766182
reclaim_successes               4    107788388
reclaim_failures                4    4045329451
streams_resets                  4    20989756
streams_noresets                4    503182328
bogus_streams                   4    0
`
const vdev_cache_statsContents = `7 1 0x01 3 144 23617323692 12081684236238879
name                            type data
delegations                     4    0
hits                            4    0
misses                          4    0
`
const pool_ioContents = `11 3 0x00 1 80 2225326830828 32953476980628
nread    nwritten reads    writes   wtime    wlentime wupdate  rtime    rlentime rupdate  wcnt     rcnt    
1884160  6450688  22       978      272187126 2850519036 2263669418655 424226814 2850519036 2263669871823 0        0       
`

var testKstatPath = os.TempDir() + "/telegraf/proc/spl/kstat/zfs"

type metrics struct {
	name  string
	value int64
}

func TestZfsPoolMetrics(t *testing.T) {
	err := os.MkdirAll(testKstatPath, 0755)
	require.NoError(t, err)

	err = os.MkdirAll(testKstatPath+"/HOME", 0755)
	require.NoError(t, err)

	err = ioutil.WriteFile(testKstatPath+"/HOME/io", []byte(pool_ioContents), 0644)
	require.NoError(t, err)

	err = ioutil.WriteFile(testKstatPath+"/arcstats", []byte(arcstatsContents), 0644)
	require.NoError(t, err)

	poolMetrics := getPoolMetrics()

	var acc testutil.Accumulator

	//one pool, all metrics
	tags := map[string]string{
		"pool": "HOME",
	}

	z := &Zfs{KstatPath: testKstatPath, KstatMetrics: []string{"arcstats"}}
	err = z.Gather(&acc)
	require.NoError(t, err)

	for _, metric := range poolMetrics {
		assert.True(t, !acc.HasIntValue(metric.name), metric.name)
		assert.True(t, !acc.CheckTaggedValue(metric.name, metric.value, tags))
	}

	z = &Zfs{KstatPath: testKstatPath, KstatMetrics: []string{"arcstats"}, PoolMetrics: true}
	err = z.Gather(&acc)
	require.NoError(t, err)

	for _, metric := range poolMetrics {
		assert.True(t, acc.HasIntValue(metric.name), metric.name)
		assert.True(t, acc.CheckTaggedValue(metric.name, metric.value, tags))
	}

	err = os.RemoveAll(os.TempDir() + "/telegraf")
	require.NoError(t, err)
}

func TestZfsGeneratesMetrics(t *testing.T) {
	err := os.MkdirAll(testKstatPath, 0755)
	require.NoError(t, err)

	err = os.MkdirAll(testKstatPath+"/HOME", 0755)
	require.NoError(t, err)

	err = ioutil.WriteFile(testKstatPath+"/HOME/io", []byte(""), 0644)
	require.NoError(t, err)

	err = ioutil.WriteFile(testKstatPath+"/arcstats", []byte(arcstatsContents), 0644)
	require.NoError(t, err)

	err = ioutil.WriteFile(testKstatPath+"/zfetchstats", []byte(zfetchstatsContents), 0644)
	require.NoError(t, err)

	err = ioutil.WriteFile(testKstatPath+"/vdev_cache_stats", []byte(vdev_cache_statsContents), 0644)
	require.NoError(t, err)

	intMetrics := getKstatMetricsAll()

	var acc testutil.Accumulator

	//one pool, all metrics
	tags := map[string]string{
		"pools": "HOME",
	}

	z := &Zfs{KstatPath: testKstatPath}
	err = z.Gather(&acc)
	require.NoError(t, err)

	for _, metric := range intMetrics {
		assert.True(t, acc.HasIntValue(metric.name), metric.name)
		assert.True(t, acc.CheckTaggedValue(metric.name, metric.value, tags))
	}

	//two pools, all metrics
	err = os.MkdirAll(testKstatPath+"/STORAGE", 0755)
	require.NoError(t, err)

	err = ioutil.WriteFile(testKstatPath+"/STORAGE/io", []byte(""), 0644)
	require.NoError(t, err)

	tags = map[string]string{
		"pools": "HOME::STORAGE",
	}

	z = &Zfs{KstatPath: testKstatPath}
	acc = testutil.Accumulator{}
	err = z.Gather(&acc)
	require.NoError(t, err)

	for _, metric := range intMetrics {
		assert.True(t, acc.HasIntValue(metric.name), metric.name)
		assert.True(t, acc.CheckTaggedValue(metric.name, metric.value, tags))
	}

	intMetrics = getKstatMetricsArcOnly()

	//two pools, one metric
	z = &Zfs{KstatPath: testKstatPath, KstatMetrics: []string{"arcstats"}}
	acc = testutil.Accumulator{}
	err = z.Gather(&acc)
	require.NoError(t, err)

	for _, metric := range intMetrics {
		assert.True(t, acc.HasIntValue(metric.name), metric.name)
		assert.True(t, acc.CheckTaggedValue(metric.name, metric.value, tags))
	}

	err = os.RemoveAll(os.TempDir() + "/telegraf")
	require.NoError(t, err)
}

func getKstatMetricsArcOnly() []*metrics {
	return []*metrics{
		{
			name:  "arcstats_hits",
			value: 5968846374,
		},
		{
			name:  "arcstats_misses",
			value: 1659178751,
		},
		{
			name:  "arcstats_demand_data_hits",
			value: 4860247322,
		},
		{
			name:  "arcstats_demand_data_misses",
			value: 501499535,
		},
		{
			name:  "arcstats_demand_metadata_hits",
			value: 708608325,
		},
		{
			name:  "arcstats_demand_metadata_misses",
			value: 156591375,
		},
		{
			name:  "arcstats_prefetch_data_hits",
			value: 367047144,
		},
		{
			name:  "arcstats_prefetch_data_misses",
			value: 974529898,
		},
		{
			name:  "arcstats_prefetch_metadata_hits",
			value: 32943583,
		},
		{
			name:  "arcstats_prefetch_metadata_misses",
			value: 26557943,
		},
		{
			name:  "arcstats_mru_hits",
			value: 301176811,
		},
		{
			name:  "arcstats_mru_ghost_hits",
			value: 47066067,
		},
		{
			name:  "arcstats_mfu_hits",
			value: 5520612438,
		},
		{
			name:  "arcstats_mfu_ghost_hits",
			value: 45784009,
		},
		{
			name:  "arcstats_deleted",
			value: 1718937704,
		},
		{
			name:  "arcstats_recycle_miss",
			value: 481222994,
		},
		{
			name:  "arcstats_mutex_miss",
			value: 20575623,
		},
		{
			name:  "arcstats_evict_skip",
			value: 14655903906543,
		},
		{
			name:  "arcstats_evict_l2_cached",
			value: 145310202998272,
		},
		{
			name:  "arcstats_evict_l2_eligible",
			value: 16345402777088,
		},
		{
			name:  "arcstats_evict_l2_ineligible",
			value: 7437226893312,
		},
		{
			name:  "arcstats_hash_elements",
			value: 36617980,
		},
		{
			name:  "arcstats_hash_elements_max",
			value: 36618318,
		},
		{
			name:  "arcstats_hash_collisions",
			value: 554145157,
		},
		{
			name:  "arcstats_hash_chains",
			value: 4187651,
		},
		{
			name:  "arcstats_hash_chain_max",
			value: 26,
		},
		{
			name:  "arcstats_p",
			value: 13963222064,
		},
		{
			name:  "arcstats_c",
			value: 16381258376,
		},
		{
			name:  "arcstats_c_min",
			value: 4194304,
		},
		{
			name:  "arcstats_c_max",
			value: 16884125696,
		},
		{
			name:  "arcstats_size",
			value: 16319887096,
		},
		{
			name:  "arcstats_hdr_size",
			value: 42567864,
		},
		{
			name:  "arcstats_data_size",
			value: 60066304,
		},
		{
			name:  "arcstats_meta_size",
			value: 1701534208,
		},
		{
			name:  "arcstats_other_size",
			value: 1661543168,
		},
		{
			name:  "arcstats_anon_size",
			value: 94720,
		},
		{
			name:  "arcstats_anon_evict_data",
			value: 0,
		},
		{
			name:  "arcstats_anon_evict_metadata",
			value: 0,
		},
		{
			name:  "arcstats_mru_size",
			value: 973099008,
		},
		{
			name:  "arcstats_mru_evict_data",
			value: 9175040,
		},
		{
			name:  "arcstats_mru_evict_metadata",
			value: 32768,
		},
		{
			name:  "arcstats_mru_ghost_size",
			value: 32768,
		},
		{
			name:  "arcstats_mru_ghost_evict_data",
			value: 0,
		},
		{
			name:  "arcstats_mru_ghost_evict_metadata",
			value: 32768,
		},
		{
			name:  "arcstats_mfu_size",
			value: 788406784,
		},
		{
			name:  "arcstats_mfu_evict_data",
			value: 50881024,
		},
		{
			name:  "arcstats_mfu_evict_metadata",
			value: 81920,
		},
		{
			name:  "arcstats_mfu_ghost_size",
			value: 0,
		},
		{
			name:  "arcstats_mfu_ghost_evict_data",
			value: 0,
		},
		{
			name:  "arcstats_mfu_ghost_evict_metadata",
			value: 0,
		},
		{
			name:  "arcstats_l2_hits",
			value: 573868618,
		},
		{
			name:  "arcstats_l2_misses",
			value: 1085309718,
		},
		{
			name:  "arcstats_l2_feeds",
			value: 12182087,
		},
		{
			name:  "arcstats_l2_rw_clash",
			value: 9610,
		},
		{
			name:  "arcstats_l2_read_bytes",
			value: 32695938336768,
		},
		{
			name:  "arcstats_l2_write_bytes",
			value: 2826774778880,
		},
		{
			name:  "arcstats_l2_writes_sent",
			value: 4267687,
		},
		{
			name:  "arcstats_l2_writes_done",
			value: 4267687,
		},
		{
			name:  "arcstats_l2_writes_error",
			value: 0,
		},
		{
			name:  "arcstats_l2_writes_hdr_miss",
			value: 164,
		},
		{
			name:  "arcstats_l2_evict_lock_retry",
			value: 5,
		},
		{
			name:  "arcstats_l2_evict_reading",
			value: 0,
		},
		{
			name:  "arcstats_l2_free_on_write",
			value: 1606914,
		},
		{
			name:  "arcstats_l2_cdata_free_on_write",
			value: 1775,
		},
		{
			name:  "arcstats_l2_abort_lowmem",
			value: 83462,
		},
		{
			name:  "arcstats_l2_cksum_bad",
			value: 393860640,
		},
		{
			name:  "arcstats_l2_io_error",
			value: 53881460,
		},
		{
			name:  "arcstats_l2_size",
			value: 2471466648576,
		},
		{
			name:  "arcstats_l2_asize",
			value: 2461690072064,
		},
		{
			name:  "arcstats_l2_hdr_size",
			value: 12854175552,
		},
		{
			name:  "arcstats_l2_compress_successes",
			value: 12184849,
		},
		{
			name:  "arcstats_l2_compress_zeros",
			value: 0,
		},
		{
			name:  "arcstats_l2_compress_failures",
			value: 0,
		},
		{
			name:  "arcstats_memory_throttle_count",
			value: 0,
		},
		{
			name:  "arcstats_duplicate_buffers",
			value: 0,
		},
		{
			name:  "arcstats_duplicate_buffers_size",
			value: 0,
		},
		{
			name:  "arcstats_duplicate_reads",
			value: 0,
		},
		{
			name:  "arcstats_memory_direct_count",
			value: 5159942,
		},
		{
			name:  "arcstats_memory_indirect_count",
			value: 3034640,
		},
		{
			name:  "arcstats_arc_no_grow",
			value: 0,
		},
		{
			name:  "arcstats_arc_tempreserve",
			value: 0,
		},
		{
			name:  "arcstats_arc_loaned_bytes",
			value: 0,
		},
		{
			name:  "arcstats_arc_prune",
			value: 114554259559,
		},
		{
			name:  "arcstats_arc_meta_used",
			value: 16259820792,
		},
		{
			name:  "arcstats_arc_meta_limit",
			value: 12663094272,
		},
		{
			name:  "arcstats_arc_meta_max",
			value: 18327165696,
		},
	}
}

func getKstatMetricsAll() []*metrics {
	otherMetrics := []*metrics{
		{
			name:  "zfetchstats_hits",
			value: 7812959060,
		},
		{
			name:  "zfetchstats_misses",
			value: 4154484207,
		},
		{
			name:  "zfetchstats_colinear_hits",
			value: 1366368,
		},
		{
			name:  "zfetchstats_colinear_misses",
			value: 4153117839,
		},
		{
			name:  "zfetchstats_stride_hits",
			value: 7309776732,
		},
		{
			name:  "zfetchstats_stride_misses",
			value: 222766182,
		},
		{
			name:  "zfetchstats_reclaim_successes",
			value: 107788388,
		},
		{
			name:  "zfetchstats_reclaim_failures",
			value: 4045329451,
		},
		{
			name:  "zfetchstats_streams_resets",
			value: 20989756,
		},
		{
			name:  "zfetchstats_streams_noresets",
			value: 503182328,
		},
		{
			name:  "zfetchstats_bogus_streams",
			value: 0,
		},
		{
			name:  "vdev_cache_stats_delegations",
			value: 0,
		},
		{
			name:  "vdev_cache_stats_hits",
			value: 0,
		},
		{
			name:  "vdev_cache_stats_misses",
			value: 0,
		},
	}

	return append(getKstatMetricsArcOnly(), otherMetrics...)
}

func getPoolMetrics() []*metrics {
	return []*metrics{
		{
			name:  "nread",
			value: 1884160,
		},
		{
			name:  "nwritten",
			value: 6450688,
		},
		{
			name:  "reads",
			value: 22,
		},
		{
			name:  "writes",
			value: 978,
		},
		{
			name:  "wtime",
			value: 272187126,
		},
		{
			name:  "wlentime",
			value: 2850519036,
		},
		{
			name:  "wupdate",
			value: 2263669418655,
		},
		{
			name:  "rtime",
			value: 424226814,
		},
		{
			name:  "rlentime",
			value: 2850519036,
		},
		{
			name:  "rupdate",
			value: 2263669871823,
		},
		{
			name:  "wcnt",
			value: 0,
		},
		{
			name:  "rcnt",
			value: 0,
		},
	}
}
