// +build linux

package zfs

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/influxdata/telegraf/testutil"
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

	z := &Zfs{KstatPath: testKstatPath, KstatMetrics: []string{"arcstats"}}
	err = z.Gather(&acc)
	require.NoError(t, err)

	require.False(t, acc.HasMeasurement("zfs_pool"))
	acc.Metrics = nil

	z = &Zfs{KstatPath: testKstatPath, KstatMetrics: []string{"arcstats"}, PoolMetrics: true}
	err = z.Gather(&acc)
	require.NoError(t, err)

	//one pool, all metrics
	tags := map[string]string{
		"pool": "HOME",
	}

	acc.AssertContainsTaggedFields(t, "zfs_pool", poolMetrics, tags)

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

	acc.AssertContainsTaggedFields(t, "zfs", intMetrics, tags)
	acc.Metrics = nil

	//two pools, all metrics
	err = os.MkdirAll(testKstatPath+"/STORAGE", 0755)
	require.NoError(t, err)

	err = ioutil.WriteFile(testKstatPath+"/STORAGE/io", []byte(""), 0644)
	require.NoError(t, err)

	tags = map[string]string{
		"pools": "HOME::STORAGE",
	}

	z = &Zfs{KstatPath: testKstatPath}
	acc2 := testutil.Accumulator{}
	err = z.Gather(&acc2)
	require.NoError(t, err)

	acc2.AssertContainsTaggedFields(t, "zfs", intMetrics, tags)
	acc2.Metrics = nil

	intMetrics = getKstatMetricsArcOnly()

	//two pools, one metric
	z = &Zfs{KstatPath: testKstatPath, KstatMetrics: []string{"arcstats"}}
	acc3 := testutil.Accumulator{}
	err = z.Gather(&acc3)
	require.NoError(t, err)

	acc3.AssertContainsTaggedFields(t, "zfs", intMetrics, tags)

	err = os.RemoveAll(os.TempDir() + "/telegraf")
	require.NoError(t, err)
}

func getKstatMetricsArcOnly() map[string]interface{} {
	return map[string]interface{}{
		"arcstats_hits":                     int64(5968846374),
		"arcstats_misses":                   int64(1659178751),
		"arcstats_demand_data_hits":         int64(4860247322),
		"arcstats_demand_data_misses":       int64(501499535),
		"arcstats_demand_metadata_hits":     int64(708608325),
		"arcstats_demand_metadata_misses":   int64(156591375),
		"arcstats_prefetch_data_hits":       int64(367047144),
		"arcstats_prefetch_data_misses":     int64(974529898),
		"arcstats_prefetch_metadata_hits":   int64(32943583),
		"arcstats_prefetch_metadata_misses": int64(26557943),
		"arcstats_mru_hits":                 int64(301176811),
		"arcstats_mru_ghost_hits":           int64(47066067),
		"arcstats_mfu_hits":                 int64(5520612438),
		"arcstats_mfu_ghost_hits":           int64(45784009),
		"arcstats_deleted":                  int64(1718937704),
		"arcstats_recycle_miss":             int64(481222994),
		"arcstats_mutex_miss":               int64(20575623),
		"arcstats_evict_skip":               int64(14655903906543),
		"arcstats_evict_l2_cached":          int64(145310202998272),
		"arcstats_evict_l2_eligible":        int64(16345402777088),
		"arcstats_evict_l2_ineligible":      int64(7437226893312),
		"arcstats_hash_elements":            int64(36617980),
		"arcstats_hash_elements_max":        int64(36618318),
		"arcstats_hash_collisions":          int64(554145157),
		"arcstats_hash_chains":              int64(4187651),
		"arcstats_hash_chain_max":           int64(26),
		"arcstats_p":                        int64(13963222064),
		"arcstats_c":                        int64(16381258376),
		"arcstats_c_min":                    int64(4194304),
		"arcstats_c_max":                    int64(16884125696),
		"arcstats_size":                     int64(16319887096),
		"arcstats_hdr_size":                 int64(42567864),
		"arcstats_data_size":                int64(60066304),
		"arcstats_meta_size":                int64(1701534208),
		"arcstats_other_size":               int64(1661543168),
		"arcstats_anon_size":                int64(94720),
		"arcstats_anon_evict_data":          int64(0),
		"arcstats_anon_evict_metadata":      int64(0),
		"arcstats_mru_size":                 int64(973099008),
		"arcstats_mru_evict_data":           int64(9175040),
		"arcstats_mru_evict_metadata":       int64(32768),
		"arcstats_mru_ghost_size":           int64(32768),
		"arcstats_mru_ghost_evict_data":     int64(0),
		"arcstats_mru_ghost_evict_metadata": int64(32768),
		"arcstats_mfu_size":                 int64(788406784),
		"arcstats_mfu_evict_data":           int64(50881024),
		"arcstats_mfu_evict_metadata":       int64(81920),
		"arcstats_mfu_ghost_size":           int64(0),
		"arcstats_mfu_ghost_evict_data":     int64(0),
		"arcstats_mfu_ghost_evict_metadata": int64(0),
		"arcstats_l2_hits":                  int64(573868618),
		"arcstats_l2_misses":                int64(1085309718),
		"arcstats_l2_feeds":                 int64(12182087),
		"arcstats_l2_rw_clash":              int64(9610),
		"arcstats_l2_read_bytes":            int64(32695938336768),
		"arcstats_l2_write_bytes":           int64(2826774778880),
		"arcstats_l2_writes_sent":           int64(4267687),
		"arcstats_l2_writes_done":           int64(4267687),
		"arcstats_l2_writes_error":          int64(0),
		"arcstats_l2_writes_hdr_miss":       int64(164),
		"arcstats_l2_evict_lock_retry":      int64(5),
		"arcstats_l2_evict_reading":         int64(0),
		"arcstats_l2_free_on_write":         int64(1606914),
		"arcstats_l2_cdata_free_on_write":   int64(1775),
		"arcstats_l2_abort_lowmem":          int64(83462),
		"arcstats_l2_cksum_bad":             int64(393860640),
		"arcstats_l2_io_error":              int64(53881460),
		"arcstats_l2_size":                  int64(2471466648576),
		"arcstats_l2_asize":                 int64(2461690072064),
		"arcstats_l2_hdr_size":              int64(12854175552),
		"arcstats_l2_compress_successes":    int64(12184849),
		"arcstats_l2_compress_zeros":        int64(0),
		"arcstats_l2_compress_failures":     int64(0),
		"arcstats_memory_throttle_count":    int64(0),
		"arcstats_duplicate_buffers":        int64(0),
		"arcstats_duplicate_buffers_size":   int64(0),
		"arcstats_duplicate_reads":          int64(0),
		"arcstats_memory_direct_count":      int64(5159942),
		"arcstats_memory_indirect_count":    int64(3034640),
		"arcstats_arc_no_grow":              int64(0),
		"arcstats_arc_tempreserve":          int64(0),
		"arcstats_arc_loaned_bytes":         int64(0),
		"arcstats_arc_prune":                int64(114554259559),
		"arcstats_arc_meta_used":            int64(16259820792),
		"arcstats_arc_meta_limit":           int64(12663094272),
		"arcstats_arc_meta_max":             int64(18327165696),
	}
}

func getKstatMetricsAll() map[string]interface{} {
	otherMetrics := map[string]interface{}{
		"zfetchstats_hits":              int64(7812959060),
		"zfetchstats_misses":            int64(4154484207),
		"zfetchstats_colinear_hits":     int64(1366368),
		"zfetchstats_colinear_misses":   int64(4153117839),
		"zfetchstats_stride_hits":       int64(7309776732),
		"zfetchstats_stride_misses":     int64(222766182),
		"zfetchstats_reclaim_successes": int64(107788388),
		"zfetchstats_reclaim_failures":  int64(4045329451),
		"zfetchstats_streams_resets":    int64(20989756),
		"zfetchstats_streams_noresets":  int64(503182328),
		"zfetchstats_bogus_streams":     int64(0),
		"vdev_cache_stats_delegations":  int64(0),
		"vdev_cache_stats_hits":         int64(0),
		"vdev_cache_stats_misses":       int64(0),
	}
	arcMetrics := getKstatMetricsArcOnly()
	for k, v := range otherMetrics {
		arcMetrics[k] = v
	}
	return arcMetrics
}

func getPoolMetrics() map[string]interface{} {
	return map[string]interface{}{
		"nread":    int64(1884160),
		"nwritten": int64(6450688),
		"reads":    int64(22),
		"writes":   int64(978),
		"wtime":    int64(272187126),
		"wlentime": int64(2850519036),
		"wupdate":  int64(2263669418655),
		"rtime":    int64(424226814),
		"rlentime": int64(2850519036),
		"rupdate":  int64(2263669871823),
		"wcnt":     int64(0),
		"rcnt":     int64(0),
	}
}
