package ceph

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

const (
	epsilon = float64(0.00000001)
)

type expectedResult struct {
	metric string
	fields map[string]interface{}
	tags   map[string]string
}

func TestParseSockId(t *testing.T) {
	s := parseSockID(sockFile(osdPrefix, 1), osdPrefix, sockSuffix)
	require.Equal(t, s, "1")
}

func TestParseMonDump(t *testing.T) {
	c := &Ceph{Log: testutil.Logger{}}
	dump, err := c.parseDump(monPerfDump)
	require.NoError(t, err)
	require.InEpsilon(t, int64(5678670180), dump["cluster"]["osd_kb_used"], epsilon)
	require.InEpsilon(t, 6866.540527000, dump["paxos"]["store_state_latency.sum"], epsilon)
}

func TestParseOsdDump(t *testing.T) {
	c := &Ceph{Log: testutil.Logger{}}
	dump, err := c.parseDump(osdPerfDump)
	require.NoError(t, err)
	require.InEpsilon(t, 552132.109360000, dump["filestore"]["commitcycle_interval.sum"], epsilon)
	require.Equal(t, float64(0), dump["mutex-FileJournal::finisher_lock"]["wait.avgcount"])
}

func TestParseMdsDump(t *testing.T) {
	c := &Ceph{Log: testutil.Logger{}}
	dump, err := c.parseDump(mdsPerfDump)
	require.NoError(t, err)
	require.InEpsilon(t, 2408386.600934982, dump["mds"]["reply_latency.sum"], epsilon)
	require.Equal(t, float64(0), dump["throttle-write_buf_throttle"]["wait.avgcount"])
}

func TestParseRgwDump(t *testing.T) {
	c := &Ceph{Log: testutil.Logger{}}
	dump, err := c.parseDump(rgwPerfDump)
	require.NoError(t, err)
	require.InEpsilon(t, 0.002219876, dump["rgw"]["get_initial_lat.sum"], epsilon)
	require.Equal(t, float64(0), dump["rgw"]["put_initial_lat.avgcount"])
}

func TestDecodeStatus(t *testing.T) {
	acc := &testutil.Accumulator{}
	err := decodeStatus(acc, clusterStatusDump)
	require.NoError(t, err)

	for _, r := range cephStatusResults {
		acc.AssertContainsTaggedFields(t, r.metric, r.fields, r.tags)
	}
}

func TestDecodeDf(t *testing.T) {
	acc := &testutil.Accumulator{}
	err := decodeDf(acc, cephDFDump)
	require.NoError(t, err)

	for _, r := range cephDfResults {
		acc.AssertContainsTaggedFields(t, r.metric, r.fields, r.tags)
	}
}

func TestDecodeOSDPoolStats(t *testing.T) {
	acc := &testutil.Accumulator{}
	err := decodeOsdPoolStats(acc, cephODSPoolStatsDump)
	require.NoError(t, err)

	for _, r := range cephOSDPoolStatsResults {
		acc.AssertContainsTaggedFields(t, r.metric, r.fields, r.tags)
	}
}

func TestGather(t *testing.T) {
	saveFind := findSockets
	saveDump := perfDump
	defer func() {
		findSockets = saveFind
		perfDump = saveDump
	}()

	findSockets = func(c *Ceph) ([]*socket, error) {
		return []*socket{{"osd.1", typeOsd, ""}}, nil
	}

	perfDump = func(binary string, s *socket) (string, error) {
		return osdPerfDump, nil
	}

	acc := &testutil.Accumulator{}
	c := &Ceph{}
	require.NoError(t, c.Gather(acc))
}

func TestFindSockets(t *testing.T) {
	tmpdir := t.TempDir()
	c := &Ceph{
		CephBinary:             "foo",
		OsdPrefix:              "ceph-osd",
		MonPrefix:              "ceph-mon",
		MdsPrefix:              "ceph-mds",
		RgwPrefix:              "ceph-client",
		SocketDir:              tmpdir,
		SocketSuffix:           "asok",
		CephUser:               "client.admin",
		CephConfig:             "/etc/ceph/ceph.conf",
		GatherAdminSocketStats: true,
		GatherClusterStats:     false,
	}

	for _, st := range sockTestParams {
		require.NoError(t, createTestFiles(tmpdir, st))

		sockets, err := findSockets(c)
		require.NoError(t, err)

		for i := 1; i <= st.osds; i++ {
			assertFoundSocket(t, tmpdir, typeOsd, i, sockets)
		}

		for i := 1; i <= st.mons; i++ {
			assertFoundSocket(t, tmpdir, typeMon, i, sockets)
		}
		for i := 1; i <= st.mdss; i++ {
			assertFoundSocket(t, tmpdir, typeMds, i, sockets)
		}
		for i := 1; i <= st.rgws; i++ {
			assertFoundSocket(t, tmpdir, typeRgw, i, sockets)
		}
		require.NoError(t, cleanupTestFiles(tmpdir, st))
	}
}

func assertFoundSocket(t *testing.T, dir, sockType string, i int, sockets []*socket) {
	var prefix string
	if sockType == typeOsd {
		prefix = osdPrefix
	} else if sockType == typeMds {
		prefix = mdsPrefix
	} else if sockType == typeRgw {
		prefix = rgwPrefix
	} else {
		prefix = monPrefix
	}
	expected := filepath.Join(dir, sockFile(prefix, i))
	found := false
	for _, s := range sockets {
		_, err := fmt.Printf("Checking %s\n", s.socket)
		require.NoError(t, err)
		if s.socket == expected {
			found = true
			require.Equal(t, s.sockType, sockType, "Unexpected socket type for '%s'", s)
			require.Equal(t, s.sockID, strconv.Itoa(i))
		}
	}
	require.True(t, found, "Did not find socket: %s", expected)
}

func sockFile(prefix string, i int) string {
	return strings.Join([]string{prefix, strconv.Itoa(i), sockSuffix}, ".")
}

func createTestFiles(dir string, st *SockTest) error {
	writeFile := func(prefix string, i int) error {
		f := sockFile(prefix, i)
		fpath := filepath.Join(dir, f)
		return os.WriteFile(fpath, []byte(""), 0777)
	}
	return tstFileApply(st, writeFile)
}

func cleanupTestFiles(dir string, st *SockTest) error {
	rmFile := func(prefix string, i int) error {
		f := sockFile(prefix, i)
		fpath := filepath.Join(dir, f)
		return os.Remove(fpath)
	}
	return tstFileApply(st, rmFile)
}

func tstFileApply(st *SockTest, fn func(string, int) error) error {
	for i := 1; i <= st.osds; i++ {
		if err := fn(osdPrefix, i); err != nil {
			return err
		}
	}
	for i := 1; i <= st.mons; i++ {
		if err := fn(monPrefix, i); err != nil {
			return err
		}
	}
	for i := 1; i <= st.mdss; i++ {
		if err := fn(mdsPrefix, i); err != nil {
			return err
		}
	}
	for i := 1; i <= st.rgws; i++ {
		if err := fn(rgwPrefix, i); err != nil {
			return err
		}
	}
	return nil
}

type SockTest struct {
	osds int
	mons int
	mdss int
	rgws int
}

var sockTestParams = []*SockTest{
	{
		osds: 2,
		mons: 2,
		mdss: 2,
		rgws: 2,
	},
	{
		mons: 1,
	},
	{
		osds: 1,
	},
	{
		mdss: 1,
	},
	{
		rgws: 1,
	},
	{},
}

var monPerfDump = `
{ "cluster": { "num_mon": 2,
      "num_mon_quorum": 2,
      "num_osd": 26,
      "num_osd_up": 26,
      "num_osd_in": 26,
      "osd_epoch": 3306,
      "osd_kb": 11487846448,
      "osd_kb_used": 5678670180,
      "osd_kb_avail": 5809176268,
      "num_pool": 12,
      "num_pg": 768,
      "num_pg_active_clean": 768,
      "num_pg_active": 768,
      "num_pg_peering": 0,
      "num_object": 397616,
      "num_object_degraded": 0,
      "num_object_unfound": 0,
      "num_bytes": 2917848227467,
      "num_mds_up": 0,
      "num_mds_in": 0,
      "num_mds_failed": 0,
      "mds_epoch": 1},
  "leveldb": { "leveldb_get": 321950312,
      "leveldb_transaction": 18729922,
      "leveldb_compact": 0,
      "leveldb_compact_range": 74141,
      "leveldb_compact_queue_merge": 0,
      "leveldb_compact_queue_len": 0},
  "mon": {},
  "paxos": { "start_leader": 0,
      "start_peon": 1,
      "restart": 4,
      "refresh": 9363435,
      "refresh_latency": { "avgcount": 9363435,
          "sum": 5378.794002000},
      "begin": 9363435,
      "begin_keys": { "avgcount": 0,
          "sum": 0},
      "begin_bytes": { "avgcount": 9363435,
          "sum": 110468605489},
      "begin_latency": { "avgcount": 9363435,
          "sum": 5850.060682000},
      "commit": 9363435,
      "commit_keys": { "avgcount": 0,
          "sum": 0},
      "commit_bytes": { "avgcount": 0,
          "sum": 0},
      "commit_latency": { "avgcount": 0,
          "sum": 0.000000000},
      "collect": 1,
      "collect_keys": { "avgcount": 1,
          "sum": 1},
      "collect_bytes": { "avgcount": 1,
          "sum": 24},
      "collect_latency": { "avgcount": 1,
          "sum": 0.000280000},
      "collect_uncommitted": 0,
      "collect_timeout": 0,
      "accept_timeout": 0,
      "lease_ack_timeout": 0,
      "lease_timeout": 0,
      "store_state": 9363435,
      "store_state_keys": { "avgcount": 9363435,
          "sum": 176572789},
      "store_state_bytes": { "avgcount": 9363435,
          "sum": 216355887217},
      "store_state_latency": { "avgcount": 9363435,
          "sum": 6866.540527000},
      "share_state": 0,
      "share_state_keys": { "avgcount": 0,
          "sum": 0},
      "share_state_bytes": { "avgcount": 0,
          "sum": 0},
      "new_pn": 0,
      "new_pn_latency": { "avgcount": 0,
          "sum": 0.000000000}},
  "throttle-mon_client_bytes": { "val": 246,
      "max": 104857600,
      "get": 896030,
      "get_sum": 45854374,
      "get_or_fail_fail": 0,
      "get_or_fail_success": 0,
      "take": 0,
      "take_sum": 0,
      "put": 896026,
      "put_sum": 45854128,
      "wait": { "avgcount": 0,
          "sum": 0.000000000}},
  "throttle-mon_daemon_bytes": { "val": 0,
      "max": 419430400,
      "get": 2773768,
      "get_sum": 3627676976,
      "get_or_fail_fail": 0,
      "get_or_fail_success": 0,
      "take": 0,
      "take_sum": 0,
      "put": 2773768,
      "put_sum": 3627676976,
      "wait": { "avgcount": 0,
          "sum": 0.000000000}},
  "throttle-msgr_dispatch_throttler-mon": { "val": 0,
      "max": 104857600,
      "get": 34504949,
      "get_sum": 226860281124,
      "get_or_fail_fail": 0,
      "get_or_fail_success": 0,
      "take": 0,
      "take_sum": 0,
      "put": 34504949,
      "put_sum": 226860281124,
      "wait": { "avgcount": 0,
          "sum": 0.000000000}}}
`

var osdPerfDump = `
{ "WBThrottle": { "bytes_dirtied": 28405539,
      "bytes_wb": 0,
      "ios_dirtied": 93,
      "ios_wb": 0,
      "inodes_dirtied": 86,
      "inodes_wb": 0},
  "filestore": { "journal_queue_max_ops": 0,
      "journal_queue_ops": 0,
      "journal_ops": 1108008,
      "journal_queue_max_bytes": 0,
      "journal_queue_bytes": 0,
      "journal_bytes": 73233416196,
      "journal_latency": { "avgcount": 1108008,
          "sum": 290.981036000},
      "journal_wr": 1091866,
      "journal_wr_bytes": { "avgcount": 1091866,
          "sum": 74925682688},
      "journal_full": 0,
      "committing": 0,
      "commitcycle": 110389,
      "commitcycle_interval": { "avgcount": 110389,
          "sum": 552132.109360000},
      "commitcycle_latency": { "avgcount": 110389,
          "sum": 178.657804000},
      "op_queue_max_ops": 50,
      "op_queue_ops": 0,
      "ops": 1108008,
      "op_queue_max_bytes": 104857600,
      "op_queue_bytes": 0,
      "bytes": 73226768148,
      "apply_latency": { "avgcount": 1108008,
          "sum": 947.742722000},
      "queue_transaction_latency_avg": { "avgcount": 1108008,
          "sum": 0.511327000}},
  "leveldb": { "leveldb_get": 4361221,
      "leveldb_transaction": 4351276,
      "leveldb_compact": 0,
      "leveldb_compact_range": 0,
      "leveldb_compact_queue_merge": 0,
      "leveldb_compact_queue_len": 0},
  "mutex-FileJournal::completions_lock": { "wait": { "avgcount": 0,
          "sum": 0.000000000}},
  "mutex-FileJournal::finisher_lock": { "wait": { "avgcount": 0,
          "sum": 0.000000000}},
  "mutex-FileJournal::write_lock": { "wait": { "avgcount": 0,
          "sum": 0.000000000}},
  "mutex-FileJournal::writeq_lock": { "wait": { "avgcount": 0,
          "sum": 0.000000000}},
  "mutex-JOS::ApplyManager::apply_lock": { "wait": { "avgcount": 0,
          "sum": 0.000000000}},
  "mutex-JOS::ApplyManager::com_lock": { "wait": { "avgcount": 0,
          "sum": 0.000000000}},
  "mutex-JOS::SubmitManager::lock": { "wait": { "avgcount": 0,
          "sum": 0.000000000}},
  "mutex-WBThrottle::lock": { "wait": { "avgcount": 0,
          "sum": 0.000000000}},
  "objecter": { "op_active": 0,
      "op_laggy": 0,
      "op_send": 0,
      "op_send_bytes": 0,
      "op_resend": 0,
      "op_ack": 0,
      "op_commit": 0,
      "op": 0,
      "op_r": 0,
      "op_w": 0,
      "op_rmw": 0,
      "op_pg": 0,
      "osdop_stat": 0,
      "osdop_create": 0,
      "osdop_read": 0,
      "osdop_write": 0,
      "osdop_writefull": 0,
      "osdop_append": 0,
      "osdop_zero": 0,
      "osdop_truncate": 0,
      "osdop_delete": 0,
      "osdop_mapext": 0,
      "osdop_sparse_read": 0,
      "osdop_clonerange": 0,
      "osdop_getxattr": 0,
      "osdop_setxattr": 0,
      "osdop_cmpxattr": 0,
      "osdop_rmxattr": 0,
      "osdop_resetxattrs": 0,
      "osdop_tmap_up": 0,
      "osdop_tmap_put": 0,
      "osdop_tmap_get": 0,
      "osdop_call": 0,
      "osdop_watch": 0,
      "osdop_notify": 0,
      "osdop_src_cmpxattr": 0,
      "osdop_pgls": 0,
      "osdop_pgls_filter": 0,
      "osdop_other": 0,
      "linger_active": 0,
      "linger_send": 0,
      "linger_resend": 0,
      "poolop_active": 0,
      "poolop_send": 0,
      "poolop_resend": 0,
      "poolstat_active": 0,
      "poolstat_send": 0,
      "poolstat_resend": 0,
      "statfs_active": 0,
      "statfs_send": 0,
      "statfs_resend": 0,
      "command_active": 0,
      "command_send": 0,
      "command_resend": 0,
      "map_epoch": 3300,
      "map_full": 0,
      "map_inc": 3293,
      "osd_sessions": 0,
      "osd_session_open": 0,
      "osd_session_close": 0,
      "osd_laggy": 0},
  "osd": { "opq": 0,
      "op_wip": 0,
      "op": 23939,
      "op_in_bytes": 1245903961,
      "op_out_bytes": 29103083856,
      "op_latency": { "avgcount": 23939,
          "sum": 440.192015000},
      "op_process_latency": { "avgcount": 23939,
          "sum": 30.170685000},
      "op_r": 23112,
      "op_r_out_bytes": 29103056146,
      "op_r_latency": { "avgcount": 23112,
          "sum": 19.373526000},
      "op_r_process_latency": { "avgcount": 23112,
          "sum": 14.625928000},
      "op_w": 549,
      "op_w_in_bytes": 1245804358,
      "op_w_rlat": { "avgcount": 549,
          "sum": 17.022299000},
      "op_w_latency": { "avgcount": 549,
          "sum": 418.494610000},
      "op_w_process_latency": { "avgcount": 549,
          "sum": 13.316555000},
      "op_rw": 278,
      "op_rw_in_bytes": 99603,
      "op_rw_out_bytes": 27710,
      "op_rw_rlat": { "avgcount": 278,
          "sum": 2.213785000},
      "op_rw_latency": { "avgcount": 278,
          "sum": 2.323879000},
      "op_rw_process_latency": { "avgcount": 278,
          "sum": 2.228202000},
      "subop": 1074774,
      "subop_in_bytes": 26841811636,
      "subop_latency": { "avgcount": 1074774,
          "sum": 745.509160000},
      "subop_w": 0,
      "subop_w_in_bytes": 26841811636,
      "subop_w_latency": { "avgcount": 1074774,
          "sum": 745.509160000},
      "subop_pull": 0,
      "subop_pull_latency": { "avgcount": 0,
          "sum": 0.000000000},
      "subop_push": 0,
      "subop_push_in_bytes": 0,
      "subop_push_latency": { "avgcount": 0,
          "sum": 0.000000000},
      "pull": 0,
      "push": 28,
      "push_out_bytes": 103483392,
      "push_in": 0,
      "push_in_bytes": 0,
      "recovery_ops": 15,
      "loadavg": 202,
      "buffer_bytes": 0,
      "numpg": 18,
      "numpg_primary": 8,
      "numpg_replica": 10,
      "numpg_stray": 0,
      "heartbeat_to_peers": 10,
      "heartbeat_from_peers": 0,
      "map_messages": 7413,
      "map_message_epochs": 9792,
      "map_message_epoch_dups": 10105,
      "messages_delayed_for_map": 83,
      "stat_bytes": 102123175936,
      "stat_bytes_used": 49961820160,
      "stat_bytes_avail": 52161355776,
      "copyfrom": 0,
      "tier_promote": 0,
      "tier_flush": 0,
      "tier_flush_fail": 0,
      "tier_try_flush": 0,
      "tier_try_flush_fail": 0,
      "tier_evict": 0,
      "tier_whiteout": 0,
      "tier_dirty": 230,
      "tier_clean": 0,
      "tier_delay": 0,
      "agent_wake": 0,
      "agent_skip": 0,
      "agent_flush": 0,
      "agent_evict": 0},
  "recoverystate_perf": { "initial_latency": { "avgcount": 473,
          "sum": 0.027207000},
      "started_latency": { "avgcount": 1480,
          "sum": 9854902.397648000},
      "reset_latency": { "avgcount": 1953,
          "sum": 0.096206000},
      "start_latency": { "avgcount": 1953,
          "sum": 0.059947000},
      "primary_latency": { "avgcount": 765,
          "sum": 4688922.186935000},
      "peering_latency": { "avgcount": 704,
          "sum": 1668.652135000},
      "backfilling_latency": { "avgcount": 0,
          "sum": 0.000000000},
      "waitremotebackfillreserved_latency": { "avgcount": 0,
          "sum": 0.000000000},
      "waitlocalbackfillreserved_latency": { "avgcount": 0,
          "sum": 0.000000000},
      "notbackfilling_latency": { "avgcount": 0,
          "sum": 0.000000000},
      "repnotrecovering_latency": { "avgcount": 462,
          "sum": 5158922.114600000},
      "repwaitrecoveryreserved_latency": { "avgcount": 15,
          "sum": 0.008275000},
      "repwaitbackfillreserved_latency": { "avgcount": 1,
          "sum": 0.000095000},
      "RepRecovering_latency": { "avgcount": 16,
          "sum": 2274.944727000},
      "activating_latency": { "avgcount": 514,
          "sum": 261.008520000},
      "waitlocalrecoveryreserved_latency": { "avgcount": 20,
          "sum": 0.175422000},
      "waitremoterecoveryreserved_latency": { "avgcount": 20,
          "sum": 0.682778000},
      "recovering_latency": { "avgcount": 20,
          "sum": 0.697551000},
      "recovered_latency": { "avgcount": 511,
          "sum": 0.011038000},
      "clean_latency": { "avgcount": 503,
          "sum": 4686961.154278000},
      "active_latency": { "avgcount": 506,
          "sum": 4687223.640464000},
      "replicaactive_latency": { "avgcount": 446,
          "sum": 5161197.078966000},
      "stray_latency": { "avgcount": 794,
          "sum": 4805.105128000},
      "getinfo_latency": { "avgcount": 704,
          "sum": 1138.477937000},
      "getlog_latency": { "avgcount": 678,
          "sum": 0.036393000},
      "waitactingchange_latency": { "avgcount": 69,
          "sum": 59.172893000},
      "incomplete_latency": { "avgcount": 0,
          "sum": 0.000000000},
      "getmissing_latency": { "avgcount": 609,
          "sum": 0.012288000},
      "waitupthru_latency": { "avgcount": 576,
          "sum": 530.106999000}},
  "throttle-filestore_bytes": { "val": 0,
      "max": 0,
      "get": 0,
      "get_sum": 0,
      "get_or_fail_fail": 0,
      "get_or_fail_success": 0,
      "take": 0,
      "take_sum": 0,
      "put": 0,
      "put_sum": 0,
      "wait": { "avgcount": 0,
          "sum": 0.000000000}},
  "throttle-filestore_ops": { "val": 0,
      "max": 0,
      "get": 0,
      "get_sum": 0,
      "get_or_fail_fail": 0,
      "get_or_fail_success": 0,
      "take": 0,
      "take_sum": 0,
      "put": 0,
      "put_sum": 0,
      "wait": { "avgcount": 0,
          "sum": 0.000000000}},
  "throttle-msgr_dispatch_throttler-client": { "val": 0,
      "max": 104857600,
      "get": 130730,
      "get_sum": 1246039872,
      "get_or_fail_fail": 0,
      "get_or_fail_success": 0,
      "take": 0,
      "take_sum": 0,
      "put": 130730,
      "put_sum": 1246039872,
      "wait": { "avgcount": 0,
          "sum": 0.000000000}},
  "throttle-msgr_dispatch_throttler-cluster": { "val": 0,
      "max": 104857600,
      "get": 1108033,
      "get_sum": 71277949992,
      "get_or_fail_fail": 0,
      "get_or_fail_success": 0,
      "take": 0,
      "take_sum": 0,
      "put": 1108033,
      "put_sum": 71277949992,
      "wait": { "avgcount": 0,
          "sum": 0.000000000}},
  "throttle-msgr_dispatch_throttler-hb_back_server": { "val": 0,
      "max": 104857600,
      "get": 18320575,
      "get_sum": 861067025,
      "get_or_fail_fail": 0,
      "get_or_fail_success": 0,
      "take": 0,
      "take_sum": 0,
      "put": 18320575,
      "put_sum": 861067025,
      "wait": { "avgcount": 0,
          "sum": 0.000000000}},
  "throttle-msgr_dispatch_throttler-hb_front_server": { "val": 0,
      "max": 104857600,
      "get": 18320575,
      "get_sum": 861067025,
      "get_or_fail_fail": 0,
      "get_or_fail_success": 0,
      "take": 0,
      "take_sum": 0,
      "put": 18320575,
      "put_sum": 861067025,
      "wait": { "avgcount": 0,
          "sum": 0.000000000}},
  "throttle-msgr_dispatch_throttler-hbclient": { "val": 0,
      "max": 104857600,
      "get": 40479394,
      "get_sum": 1902531518,
      "get_or_fail_fail": 0,
      "get_or_fail_success": 0,
      "take": 0,
      "take_sum": 0,
      "put": 40479394,
      "put_sum": 1902531518,
      "wait": { "avgcount": 0,
          "sum": 0.000000000}},
  "throttle-msgr_dispatch_throttler-ms_objecter": { "val": 0,
      "max": 104857600,
      "get": 0,
      "get_sum": 0,
      "get_or_fail_fail": 0,
      "get_or_fail_success": 0,
      "take": 0,
      "take_sum": 0,
      "put": 0,
      "put_sum": 0,
      "wait": { "avgcount": 0,
          "sum": 0.000000000}},
  "throttle-objecter_bytes": { "val": 0,
      "max": 104857600,
      "get": 0,
      "get_sum": 0,
      "get_or_fail_fail": 0,
      "get_or_fail_success": 0,
      "take": 0,
      "take_sum": 0,
      "put": 0,
      "put_sum": 0,
      "wait": { "avgcount": 0,
          "sum": 0.000000000}},
  "throttle-objecter_ops": { "val": 0,
      "max": 1024,
      "get": 0,
      "get_sum": 0,
      "get_or_fail_fail": 0,
      "get_or_fail_success": 0,
      "take": 0,
      "take_sum": 0,
      "put": 0,
      "put_sum": 0,
      "wait": { "avgcount": 0,
          "sum": 0.000000000}},
  "throttle-osd_client_bytes": { "val": 0,
      "max": 524288000,
      "get": 24241,
      "get_sum": 1241992581,
      "get_or_fail_fail": 0,
      "get_or_fail_success": 0,
      "take": 0,
      "take_sum": 0,
      "put": 25958,
      "put_sum": 1241992581,
      "wait": { "avgcount": 0,
          "sum": 0.000000000}},
  "throttle-osd_client_messages": { "val": 0,
      "max": 100,
      "get": 49214,
      "get_sum": 49214,
      "get_or_fail_fail": 0,
      "get_or_fail_success": 0,
      "take": 0,
      "take_sum": 0,
      "put": 49214,
      "put_sum": 49214,
      "wait": { "avgcount": 0,
          "sum": 0.000000000}}}
`
var mdsPerfDump = `
{
    "AsyncMessenger::Worker-0": {
        "msgr_recv_messages": 2723536628,
        "msgr_send_messages": 1160771414,
        "msgr_recv_bytes": 1112936719134,
        "msgr_send_bytes": 1368194904867,
        "msgr_created_connections": 18281,
        "msgr_active_connections": 83,
        "msgr_running_total_time": 109001.938705141,
        "msgr_running_send_time": 33686.215323581,
        "msgr_running_recv_time": 8374950.111041426,
        "msgr_running_fast_dispatch_time": 5828.083761243
    },
    "AsyncMessenger::Worker-1": {
        "msgr_recv_messages": 1426105165,
        "msgr_send_messages": 783174767,
        "msgr_recv_bytes": 800620150187,
        "msgr_send_bytes": 1394738277392,
        "msgr_created_connections": 17677,
        "msgr_active_connections": 100,
        "msgr_running_total_time": 70660.929329800,
        "msgr_running_send_time": 24190.940207198,
        "msgr_running_recv_time": 3920894.209204916,
        "msgr_running_fast_dispatch_time": 8206.816536602
    },
    "AsyncMessenger::Worker-2": {
        "msgr_recv_messages": 3471200310,
        "msgr_send_messages": 2757725529,
        "msgr_recv_bytes": 1331676471794,
        "msgr_send_bytes": 2593968875674,
        "msgr_created_connections": 16714,
        "msgr_active_connections": 73,
        "msgr_running_total_time": 167020.893916556,
        "msgr_running_send_time": 61197.682840176,
        "msgr_running_recv_time": 5816036.495319415,
        "msgr_running_fast_dispatch_time": 8581.768789481
    },
    "finisher-PurgeQueue": {
        "queue_len": 0,
        "complete_latency": {
            "avgcount": 20170260,
            "sum": 70213.859039869,
            "avgtime": 0.003481058
        }
    },
    "mds": {
        "request": 2167457412,
        "reply": 2167457403,
        "reply_latency": {
            "avgcount": 2167457403,
            "sum": 2408386.600934982,
            "avgtime": 0.001111157
        },
        "forward": 0,
        "dir_fetch": 585012985,
        "dir_commit": 58926158,
        "dir_split": 8,
        "dir_merge": 7,
        "inode_max": 2147483647,
        "inodes": 39604287,
        "inodes_top": 9743493,
        "inodes_bottom": 29063656,
        "inodes_pin_tail": 797138,
        "inodes_pinned": 25685011,
        "inodes_expired": 1302542128,
        "inodes_with_caps": 4517329,
        "caps": 6370838,
        "subtrees": 2,
        "traverse": 2426357623,
        "traverse_hit": 2202314009,
        "traverse_forward": 0,
        "traverse_discover": 0,
        "traverse_dir_fetch": 35332112,
        "traverse_remote_ino": 0,
        "traverse_lock": 4371557,
        "load_cent": 1966748,
        "q": 976,
        "exported": 0,
        "exported_inodes": 0,
        "imported": 0,
        "imported_inodes": 0,
        "openino_dir_fetch": 22725418,
        "openino_backtrace_fetch": 6,
        "openino_peer_discover": 0
    },
    "mds_cache": {
        "num_strays": 384,
        "num_strays_delayed": 0,
        "num_strays_enqueuing": 0,
        "strays_created": 29140050,
        "strays_enqueued": 29134399,
        "strays_reintegrated": 10171,
        "strays_migrated": 0,
        "num_recovering_processing": 0,
        "num_recovering_enqueued": 0,
        "num_recovering_prioritized": 0,
        "recovery_started": 229,
        "recovery_completed": 229,
        "ireq_enqueue_scrub": 0,
        "ireq_exportdir": 0,
        "ireq_flush": 0,
        "ireq_fragmentdir": 15,
        "ireq_fragstats": 0,
        "ireq_inodestats": 0
    },
    "mds_log": {
        "evadd": 1920368707,
        "evex": 1920372003,
        "evtrm": 1920372003,
        "ev": 106627,
        "evexg": 0,
        "evexd": 4369,
        "segadd": 2247990,
        "segex": 2247995,
        "segtrm": 2247995,
        "seg": 123,
        "segexg": 0,
        "segexd": 5,
        "expos": 24852063335817,
        "wrpos": 24852205446582,
        "rdpos": 22044255640175,
        "jlat": {
            "avgcount": 182241259,
            "sum": 1732094.198366820,
            "avgtime": 0.009504402
        },
        "replayed": 109923
    },
    "mds_mem": {
        "ino": 39604292,
        "ino+": 1307214891,
        "ino-": 1267610599,
        "dir": 22827008,
        "dir+": 591593031,
        "dir-": 568766023,
        "dn": 39604761,
        "dn+": 1376976677,
        "dn-": 1337371916,
        "cap": 6370838,
        "cap+": 1720930015,
        "cap-": 1714559177,
        "rss": 167723320,
        "heap": 322260,
        "buf": 0
    },
    "mds_server": {
        "dispatch_client_request": 2932764331,
        "dispatch_server_request": 0,
        "handle_client_request": 2167457412,
        "handle_client_session": 10929454,
        "handle_slave_request": 0,
        "req_create_latency": {
            "avgcount": 30590326,
            "sum": 23887.274170412,
            "avgtime": 0.000780876
        },
        "req_getattr_latency": {
            "avgcount": 124767480,
            "sum": 718160.497644305,
            "avgtime": 0.005755991
        },
        "req_getfilelock_latency": {
            "avgcount": 0,
            "sum": 0.000000000,
            "avgtime": 0.000000000
        },
        "req_link_latency": {
            "avgcount": 5636,
            "sum": 2.371499732,
            "avgtime": 0.000420777
        },
        "req_lookup_latency": {
            "avgcount": 474590034,
            "sum": 452548.849373476,
            "avgtime": 0.000953557
        },
        "req_lookuphash_latency": {
            "avgcount": 0,
            "sum": 0.000000000,
            "avgtime": 0.000000000
        },
        "req_lookupino_latency": {
            "avgcount": 0,
            "sum": 0.000000000,
            "avgtime": 0.000000000
        },
        "req_lookupname_latency": {
            "avgcount": 9794,
            "sum": 54.118496591,
            "avgtime": 0.005525678
        },
        "req_lookupparent_latency": {
            "avgcount": 0,
            "sum": 0.000000000,
            "avgtime": 0.000000000
        },
        "req_lookupsnap_latency": {
            "avgcount": 0,
            "sum": 0.000000000,
            "avgtime": 0.000000000
        },
        "req_lssnap_latency": {
            "avgcount": 0,
            "sum": 0.000000000,
            "avgtime": 0.000000000
        },
        "req_mkdir_latency": {
            "avgcount": 13394317,
            "sum": 13025.982105531,
            "avgtime": 0.000972500
        },
        "req_mknod_latency": {
            "avgcount": 0,
            "sum": 0.000000000,
            "avgtime": 0.000000000
        },
        "req_mksnap_latency": {
            "avgcount": 0,
            "sum": 0.000000000,
            "avgtime": 0.000000000
        },
        "req_open_latency": {
            "avgcount": 32849768,
            "sum": 12862.382994977,
            "avgtime": 0.000391551
        },
        "req_readdir_latency": {
            "avgcount": 654394394,
            "sum": 715669.609601541,
            "avgtime": 0.001093636
        },
        "req_rename_latency": {
            "avgcount": 6058807,
            "sum": 2126.232719555,
            "avgtime": 0.000350932
        },
        "req_renamesnap_latency": {
            "avgcount": 0,
            "sum": 0.000000000,
            "avgtime": 0.000000000
        },
        "req_rmdir_latency": {
            "avgcount": 1901530,
            "sum": 4064.121157858,
            "avgtime": 0.002137290
        },
        "req_rmsnap_latency": {
            "avgcount": 0,
            "sum": 0.000000000,
            "avgtime": 0.000000000
        },
        "req_rmxattr_latency": {
            "avgcount": 0,
            "sum": 0.000000000,
            "avgtime": 0.000000000
        },
        "req_setattr_latency": {
            "avgcount": 37051209,
            "sum": 171198.037329531,
            "avgtime": 0.004620578
        },
        "req_setdirlayout_latency": {
            "avgcount": 0,
            "sum": 0.000000000,
            "avgtime": 0.000000000
        },
        "req_setfilelock_latency": {
            "avgcount": 765439143,
            "sum": 262660.582883819,
            "avgtime": 0.000343150
        },
        "req_setlayout_latency": {
            "avgcount": 0,
            "sum": 0.000000000,
            "avgtime": 0.000000000
        },
        "req_setxattr_latency": {
            "avgcount": 41572,
            "sum": 7.273371375,
            "avgtime": 0.000174958
        },
        "req_symlink_latency": {
            "avgcount": 329,
            "sum": 0.117859965,
            "avgtime": 0.000358236
        },
        "req_unlink_latency": {
            "avgcount": 26363064,
            "sum": 32119.149726314,
            "avgtime": 0.001218339
        },
        "cap_revoke_eviction": 0
    },
    "mds_sessions": {
        "session_count": 80,
        "session_add": 90,
        "session_remove": 10,
        "sessions_open": 80,
        "sessions_stale": 0,
        "total_load": 112490,
        "average_load": 1406,
        "avg_session_uptime": 2221807
    },
    "objecter": {
        "op_active": 0,
        "op_laggy": 0,
        "op_send": 955060080,
        "op_send_bytes": 3178832110019,
        "op_resend": 67,
        "op_reply": 955060013,
        "op": 955060013,
        "op_r": 585982837,
        "op_w": 369077176,
        "op_rmw": 0,
        "op_pg": 0,
        "osdop_stat": 45924375,
        "osdop_create": 31162274,
        "osdop_read": 969513,
        "osdop_write": 183211164,
        "osdop_writefull": 1063233,
        "osdop_writesame": 0,
        "osdop_append": 0,
        "osdop_zero": 2,
        "osdop_truncate": 8,
        "osdop_delete": 60594735,
        "osdop_mapext": 0,
        "osdop_sparse_read": 0,
        "osdop_clonerange": 0,
        "osdop_getxattr": 584941886,
        "osdop_setxattr": 62324548,
        "osdop_cmpxattr": 0,
        "osdop_rmxattr": 0,
        "osdop_resetxattrs": 0,
        "osdop_tmap_up": 0,
        "osdop_tmap_put": 0,
        "osdop_tmap_get": 0,
        "osdop_call": 0,
        "osdop_watch": 0,
        "osdop_notify": 0,
        "osdop_src_cmpxattr": 0,
        "osdop_pgls": 0,
        "osdop_pgls_filter": 0,
        "osdop_other": 32053182,
        "linger_active": 0,
        "linger_send": 0,
        "linger_resend": 0,
        "linger_ping": 0,
        "poolop_active": 0,
        "poolop_send": 0,
        "poolop_resend": 0,
        "poolstat_active": 0,
        "poolstat_send": 0,
        "poolstat_resend": 0,
        "statfs_active": 0,
        "statfs_send": 0,
        "statfs_resend": 0,
        "command_active": 0,
        "command_send": 0,
        "command_resend": 0,
        "map_epoch": 66793,
        "map_full": 0,
        "map_inc": 1762,
        "osd_sessions": 120,
        "osd_session_open": 52554,
        "osd_session_close": 52434,
        "osd_laggy": 0,
        "omap_wr": 106692727,
        "omap_rd": 1170026044,
        "omap_del": 5674762
    },
    "purge_queue": {
        "pq_executing_ops": 0,
        "pq_executing": 0,
        "pq_executed": 29134399
    },
    "throttle-msgr_dispatch_throttler-mds": {
        "val": 0,
        "max": 104857600,
        "get_started": 0,
        "get": 7620842095,
        "get_sum": 2681291022887,
        "get_or_fail_fail": 53,
        "get_or_fail_success": 7620842095,
        "take": 0,
        "take_sum": 0,
        "put": 7620842095,
        "put_sum": 2681291022887,
        "wait": {
            "avgcount": 0,
            "sum": 0.000000000,
            "avgtime": 0.000000000
        }
    },
    "throttle-objecter_bytes": {
        "val": 0,
        "max": 104857600,
        "get_started": 0,
        "get": 0,
        "get_sum": 0,
        "get_or_fail_fail": 0,
        "get_or_fail_success": 0,
        "take": 955060013,
        "take_sum": 3172776432475,
        "put": 862340641,
        "put_sum": 3172776432475,
        "wait": {
            "avgcount": 0,
            "sum": 0.000000000,
            "avgtime": 0.000000000
        }
    },
    "throttle-objecter_ops": {
        "val": 0,
        "max": 1024,
        "get_started": 0,
        "get": 0,
        "get_sum": 0,
        "get_or_fail_fail": 0,
        "get_or_fail_success": 0,
        "take": 955060013,
        "take_sum": 955060013,
        "put": 955060013,
        "put_sum": 955060013,
        "wait": {
            "avgcount": 0,
            "sum": 0.000000000,
            "avgtime": 0.000000000
        }
    },
    "throttle-write_buf_throttle": {
        "val": 0,
        "max": 3758096384,
        "get_started": 0,
        "get": 29134399,
        "get_sum": 3160498139,
        "get_or_fail_fail": 0,
        "get_or_fail_success": 29134399,
        "take": 0,
        "take_sum": 0,
        "put": 969905,
        "put_sum": 3160498139,
        "wait": {
            "avgcount": 0,
            "sum": 0.000000000,
            "avgtime": 0.000000000
        }
    },
    "throttle-write_buf_throttle-0x561894f0b8e0": {
        "val": 286270,
        "max": 3758096384,
        "get_started": 0,
        "get": 1920368707,
        "get_sum": 2807949805409,
        "get_or_fail_fail": 0,
        "get_or_fail_success": 1920368707,
        "take": 0,
        "take_sum": 0,
        "put": 182241259,
        "put_sum": 2807949519139,
        "wait": {
            "avgcount": 0,
            "sum": 0.000000000,
            "avgtime": 0.000000000
        }
    }
}
`

var rgwPerfDump = `
{
    "AsyncMessenger::Worker-0": {
        "msgr_recv_messages": 10684185,
        "msgr_send_messages": 13448962,
        "msgr_recv_bytes": 2622531258,
        "msgr_send_bytes": 4195038384,
        "msgr_created_connections": 8029,
        "msgr_active_connections": 3,
        "msgr_running_total_time": 3249.441108544,
        "msgr_running_send_time": 739.821446096,
        "msgr_running_recv_time": 310.354319110,
        "msgr_running_fast_dispatch_time": 1915.410317430
    },
    "AsyncMessenger::Worker-1": {
        "msgr_recv_messages": 2137773,
        "msgr_send_messages": 3850070,
        "msgr_recv_bytes": 503824366,
        "msgr_send_bytes": 1130107261,
        "msgr_created_connections": 11030,
        "msgr_active_connections": 1,
        "msgr_running_total_time": 445.055291782,
        "msgr_running_send_time": 227.817750758,
        "msgr_running_recv_time": 78.974093226,
        "msgr_running_fast_dispatch_time": 47.587740615
    },
    "AsyncMessenger::Worker-2": {
        "msgr_recv_messages": 2809014,
        "msgr_send_messages": 4126613,
        "msgr_recv_bytes": 653093470,
        "msgr_send_bytes": 1022041970,
        "msgr_created_connections": 14810,
        "msgr_active_connections": 5,
        "msgr_running_total_time": 453.384703728,
        "msgr_running_send_time": 208.580910390,
        "msgr_running_recv_time": 80.075306670,
        "msgr_running_fast_dispatch_time": 46.854112208
    },
    "cct": {
        "total_workers": 0,
        "unhealthy_workers": 0
    },
    "finisher-radosclient": {
        "queue_len": 0,
        "complete_latency": {
            "avgcount": 0,
            "sum": 0.000000000,
            "avgtime": 0.000000000
        }
    },
    "finisher-radosclient-0x55994098e460": {
        "queue_len": 0,
        "complete_latency": {
            "avgcount": 0,
            "sum": 0.000000000,
            "avgtime": 0.000000000
        }
    },
    "finisher-radosclient-0x5599409901c0": {
        "queue_len": 0,
        "complete_latency": {
            "avgcount": 0,
            "sum": 0.000000000,
            "avgtime": 0.000000000
        }
    },
    "mempool": {
        "bloom_filter_bytes": 0,
        "bloom_filter_items": 0,
        "bluestore_alloc_bytes": 0,
        "bluestore_alloc_items": 0,
        "bluestore_cache_data_bytes": 0,
        "bluestore_cache_data_items": 0,
        "bluestore_cache_onode_bytes": 0,
        "bluestore_cache_onode_items": 0,
        "bluestore_cache_other_bytes": 0,
        "bluestore_cache_other_items": 0,
        "bluestore_fsck_bytes": 0,
        "bluestore_fsck_items": 0,
        "bluestore_txc_bytes": 0,
        "bluestore_txc_items": 0,
        "bluestore_writing_deferred_bytes": 0,
        "bluestore_writing_deferred_items": 0,
        "bluestore_writing_bytes": 0,
        "bluestore_writing_items": 0,
        "bluefs_bytes": 0,
        "bluefs_items": 0,
        "buffer_anon_bytes": 258469,
        "buffer_anon_items": 201,
        "buffer_meta_bytes": 0,
        "buffer_meta_items": 0,
        "osd_bytes": 0,
        "osd_items": 0,
        "osd_mapbl_bytes": 0,
        "osd_mapbl_items": 0,
        "osd_pglog_bytes": 0,
        "osd_pglog_items": 0,
        "osdmap_bytes": 74448,
        "osdmap_items": 732,
        "osdmap_mapping_bytes": 0,
        "osdmap_mapping_items": 0,
        "pgmap_bytes": 0,
        "pgmap_items": 0,
        "mds_co_bytes": 0,
        "mds_co_items": 0,
        "unittest_1_bytes": 0,
        "unittest_1_items": 0,
        "unittest_2_bytes": 0,
        "unittest_2_items": 0
    },
    "objecter": {
        "op_active": 0,
        "op_laggy": 0,
        "op_send": 9377910,
        "op_send_bytes": 312,
        "op_resend": 0,
        "op_reply": 9377904,
        "op": 9377910,
        "op_r": 2755291,
        "op_w": 6622619,
        "op_rmw": 0,
        "op_pg": 0,
        "osdop_stat": 2755258,
        "osdop_create": 8,
        "osdop_read": 25,
        "osdop_write": 0,
        "osdop_writefull": 0,
        "osdop_writesame": 0,
        "osdop_append": 0,
        "osdop_zero": 0,
        "osdop_truncate": 0,
        "osdop_delete": 0,
        "osdop_mapext": 0,
        "osdop_sparse_read": 0,
        "osdop_clonerange": 0,
        "osdop_getxattr": 0,
        "osdop_setxattr": 0,
        "osdop_cmpxattr": 0,
        "osdop_rmxattr": 0,
        "osdop_resetxattrs": 0,
        "osdop_call": 0,
        "osdop_watch": 6622611,
        "osdop_notify": 0,
        "osdop_src_cmpxattr": 0,
        "osdop_pgls": 0,
        "osdop_pgls_filter": 0,
        "osdop_other": 2755266,
        "linger_active": 8,
        "linger_send": 35,
        "linger_resend": 27,
        "linger_ping": 6622576,
        "poolop_active": 0,
        "poolop_send": 0,
        "poolop_resend": 0,
        "poolstat_active": 0,
        "poolstat_send": 0,
        "poolstat_resend": 0,
        "statfs_active": 0,
        "statfs_send": 0,
        "statfs_resend": 0,
        "command_active": 0,
        "command_send": 0,
        "command_resend": 0,
        "map_epoch": 1064,
        "map_full": 0,
        "map_inc": 106,
        "osd_sessions": 8,
        "osd_session_open": 11928,
        "osd_session_close": 11920,
        "osd_laggy": 5,
        "omap_wr": 0,
        "omap_rd": 0,
        "omap_del": 0
    },
    "objecter-0x55994098e500": {
        "op_active": 0,
        "op_laggy": 0,
        "op_send": 827839,
        "op_send_bytes": 0,
        "op_resend": 0,
        "op_reply": 827839,
        "op": 827839,
        "op_r": 0,
        "op_w": 827839,
        "op_rmw": 0,
        "op_pg": 0,
        "osdop_stat": 0,
        "osdop_create": 0,
        "osdop_read": 0,
        "osdop_write": 0,
        "osdop_writefull": 0,
        "osdop_writesame": 0,
        "osdop_append": 0,
        "osdop_zero": 0,
        "osdop_truncate": 0,
        "osdop_delete": 0,
        "osdop_mapext": 0,
        "osdop_sparse_read": 0,
        "osdop_clonerange": 0,
        "osdop_getxattr": 0,
        "osdop_setxattr": 0,
        "osdop_cmpxattr": 0,
        "osdop_rmxattr": 0,
        "osdop_resetxattrs": 0,
        "osdop_call": 0,
        "osdop_watch": 827839,
        "osdop_notify": 0,
        "osdop_src_cmpxattr": 0,
        "osdop_pgls": 0,
        "osdop_pgls_filter": 0,
        "osdop_other": 0,
        "linger_active": 1,
        "linger_send": 3,
        "linger_resend": 2,
        "linger_ping": 827836,
        "poolop_active": 0,
        "poolop_send": 0,
        "poolop_resend": 0,
        "poolstat_active": 0,
        "poolstat_send": 0,
        "poolstat_resend": 0,
        "statfs_active": 0,
        "statfs_send": 0,
        "statfs_resend": 0,
        "command_active": 0,
        "command_send": 0,
        "command_resend": 0,
        "map_epoch": 1064,
        "map_full": 0,
        "map_inc": 106,
        "osd_sessions": 1,
        "osd_session_open": 1,
        "osd_session_close": 0,
        "osd_laggy": 1,
        "omap_wr": 0,
        "omap_rd": 0,
        "omap_del": 0
    },
    "objecter-0x55994098f720": {
        "op_active": 0,
        "op_laggy": 0,
        "op_send": 5415951,
        "op_send_bytes": 205291238,
        "op_resend": 8,
        "op_reply": 5415943,
        "op": 5415943,
        "op_r": 3612105,
        "op_w": 1803838,
        "op_rmw": 0,
        "op_pg": 0,
        "osdop_stat": 0,
        "osdop_create": 0,
        "osdop_read": 0,
        "osdop_write": 0,
        "osdop_writefull": 0,
        "osdop_writesame": 0,
        "osdop_append": 0,
        "osdop_zero": 0,
        "osdop_truncate": 0,
        "osdop_delete": 0,
        "osdop_mapext": 0,
        "osdop_sparse_read": 0,
        "osdop_clonerange": 0,
        "osdop_getxattr": 0,
        "osdop_setxattr": 0,
        "osdop_cmpxattr": 0,
        "osdop_rmxattr": 0,
        "osdop_resetxattrs": 0,
        "osdop_call": 5415567,
        "osdop_watch": 0,
        "osdop_notify": 0,
        "osdop_src_cmpxattr": 0,
        "osdop_pgls": 0,
        "osdop_pgls_filter": 0,
        "osdop_other": 376,
        "linger_active": 0,
        "linger_send": 0,
        "linger_resend": 0,
        "linger_ping": 0,
        "poolop_active": 0,
        "poolop_send": 0,
        "poolop_resend": 0,
        "poolstat_active": 0,
        "poolstat_send": 0,
        "poolstat_resend": 0,
        "statfs_active": 0,
        "statfs_send": 0,
        "statfs_resend": 0,
        "command_active": 0,
        "command_send": 0,
        "command_resend": 0,
        "map_epoch": 1064,
        "map_full": 0,
        "map_inc": 106,
        "osd_sessions": 8,
        "osd_session_open": 8834,
        "osd_session_close": 8826,
        "osd_laggy": 0,
        "omap_wr": 0,
        "omap_rd": 0,
        "omap_del": 0
    },
    "rgw": {
        "req": 2755258,
        "failed_req": 0,
        "get": 0,
        "get_b": 0,
        "get_initial_lat": {
            "avgcount": 0,
            "sum": 0.002219876,
            "avgtime": 0.000000000
        },
        "put": 0,
        "put_b": 0,
        "put_initial_lat": {
            "avgcount": 0,
            "sum": 0.000000000,
            "avgtime": 0.000000000
        },
        "qlen": 0,
        "qactive": 0,
        "cache_hit": 0,
        "cache_miss": 2755261,
        "keystone_token_cache_hit": 0,
        "keystone_token_cache_miss": 0,
        "gc_retire_object": 0,
        "pubsub_event_triggered": 0,
        "pubsub_event_lost": 0,
        "pubsub_store_ok": 0,
        "pubsub_store_fail": 0,
        "pubsub_events": 0,
        "pubsub_push_ok": 0,
        "pubsub_push_failed": 0,
        "pubsub_push_pending": 0
    },
    "simple-throttler": {
        "throttle": 0
    },
    "throttle-msgr_dispatch_throttler-radosclient": {
        "val": 0,
        "max": 104857600,
        "get_started": 0,
        "get": 9379775,
        "get_sum": 1545393284,
        "get_or_fail_fail": 0,
        "get_or_fail_success": 9379775,
        "take": 0,
        "take_sum": 0,
        "put": 9379775,
        "put_sum": 1545393284,
        "wait": {
            "avgcount": 0,
            "sum": 0.000000000,
            "avgtime": 0.000000000
        }
    },
    "throttle-msgr_dispatch_throttler-radosclient-0x55994098e320": {
        "val": 0,
        "max": 104857600,
        "get_started": 0,
        "get": 829631,
        "get_sum": 162850310,
        "get_or_fail_fail": 0,
        "get_or_fail_success": 829631,
        "take": 0,
        "take_sum": 0,
        "put": 829631,
        "put_sum": 162850310,
        "wait": {
            "avgcount": 0,
            "sum": 0.000000000,
            "avgtime": 0.000000000
        }
    },
    "throttle-msgr_dispatch_throttler-radosclient-0x55994098fa40": {
        "val": 0,
        "max": 104857600,
        "get_started": 0,
        "get": 5421553,
        "get_sum": 914508527,
        "get_or_fail_fail": 0,
        "get_or_fail_success": 5421553,
        "take": 0,
        "take_sum": 0,
        "put": 5421553,
        "put_sum": 914508527,
        "wait": {
            "avgcount": 0,
            "sum": 0.000000000,
            "avgtime": 0.000000000
        }
    },
    "throttle-objecter_bytes": {
        "val": 0,
        "max": 104857600,
        "get_started": 0,
        "get": 2755292,
        "get_sum": 0,
        "get_or_fail_fail": 0,
        "get_or_fail_success": 2755292,
        "take": 0,
        "take_sum": 0,
        "put": 0,
        "put_sum": 0,
        "wait": {
            "avgcount": 0,
            "sum": 0.000000000,
            "avgtime": 0.000000000
        }
    },
    "throttle-objecter_bytes-0x55994098e780": {
        "val": 0,
        "max": 104857600,
        "get_started": 0,
        "get": 0,
        "get_sum": 0,
        "get_or_fail_fail": 0,
        "get_or_fail_success": 0,
        "take": 0,
        "take_sum": 0,
        "put": 0,
        "put_sum": 0,
        "wait": {
            "avgcount": 0,
            "sum": 0.000000000,
            "avgtime": 0.000000000
        }
    },
    "throttle-objecter_bytes-0x55994098f7c0": {
        "val": 0,
        "max": 104857600,
        "get_started": 0,
        "get": 5415614,
        "get_sum": 0,
        "get_or_fail_fail": 0,
        "get_or_fail_success": 5415614,
        "take": 0,
        "take_sum": 0,
        "put": 0,
        "put_sum": 0,
        "wait": {
            "avgcount": 0,
            "sum": 0.000000000,
            "avgtime": 0.000000000
        }
    },
    "throttle-objecter_ops": {
        "val": 0,
        "max": 24576,
        "get_started": 0,
        "get": 2755292,
        "get_sum": 2755292,
        "get_or_fail_fail": 0,
        "get_or_fail_success": 2755292,
        "take": 0,
        "take_sum": 0,
        "put": 2755292,
        "put_sum": 2755292,
        "wait": {
            "avgcount": 0,
            "sum": 0.000000000,
            "avgtime": 0.000000000
        }
    },
    "throttle-objecter_ops-0x55994098e640": {
        "val": 0,
        "max": 24576,
        "get_started": 0,
        "get": 0,
        "get_sum": 0,
        "get_or_fail_fail": 0,
        "get_or_fail_success": 0,
        "take": 0,
        "take_sum": 0,
        "put": 0,
        "put_sum": 0,
        "wait": {
            "avgcount": 0,
            "sum": 0.000000000,
            "avgtime": 0.000000000
        }
    },
    "throttle-objecter_ops-0x55994098f0e0": {
        "val": 0,
        "max": 24576,
        "get_started": 0,
        "get": 5415614,
        "get_sum": 5415614,
        "get_or_fail_fail": 0,
        "get_or_fail_success": 5415614,
        "take": 0,
        "take_sum": 0,
        "put": 5415614,
        "put_sum": 5415614,
        "wait": {
            "avgcount": 0,
            "sum": 0.000000000,
            "avgtime": 0.000000000
        }
    },
    "throttle-rgw_async_rados_ops": {
        "val": 0,
        "max": 64,
        "get_started": 0,
        "get": 0,
        "get_sum": 0,
        "get_or_fail_fail": 0,
        "get_or_fail_success": 0,
        "take": 0,
        "take_sum": 0,
        "put": 0,
        "put_sum": 0,
        "wait": {
            "avgcount": 0,
            "sum": 0.000000000,
            "avgtime": 0.000000000
        }
    }
}
`

var clusterStatusDump = `
{
  "health": {
    "health": {
      "health_services": [
        {
          "mons": [
            {
              "name": "a",
              "kb_total": 114289256,
              "kb_used": 26995516,
              "kb_avail": 81465132,
              "avail_percent": 71,
              "last_updated": "2017-01-03 17:20:57.595004",
              "store_stats": {
                "bytes_total": 942117141,
                "bytes_sst": 0,
                "bytes_log": 4345406,
                "bytes_misc": 937771735,
                "last_updated": "0.000000"
              },
              "health": "HEALTH_OK"
            },
            {
              "name": "b",
              "kb_total": 114289256,
              "kb_used": 27871624,
              "kb_avail": 80589024,
              "avail_percent": 70,
              "last_updated": "2017-01-03 17:20:47.784331",
              "store_stats": {
                "bytes_total": 454853104,
                "bytes_sst": 0,
                "bytes_log": 5788320,
                "bytes_misc": 449064784,
                "last_updated": "0.000000"
              },
              "health": "HEALTH_OK"
            },
            {
              "name": "c",
              "kb_total": 130258508,
              "kb_used": 38076996,
              "kb_avail": 85541692,
              "avail_percent": 65,
              "last_updated": "2017-01-03 17:21:03.311123",
              "store_stats": {
                "bytes_total": 455555199,
                "bytes_sst": 0,
                "bytes_log": 6950876,
                "bytes_misc": 448604323,
                "last_updated": "0.000000"
              },
              "health": "HEALTH_OK"
            }
          ]
        }
      ]
    },
    "timechecks": {
      "epoch": 504,
      "round": 34642,
      "round_status": "finished",
      "mons": [
        { "name": "a", "skew": 0, "latency": 0, "health": "HEALTH_OK" },
        { "name": "b", "skew": -0, "latency": 0.000951, "health": "HEALTH_OK" },
        { "name": "c", "skew": -0, "latency": 0.000946, "health": "HEALTH_OK" }
      ]
    },
    "summary": [],
    "overall_status": "HEALTH_OK",
    "detail": []
  },
  "fsid": "01234567-abcd-9876-0123-ffeeddccbbaa",
  "election_epoch": 504,
  "quorum": [ 0, 1, 2 ],
  "quorum_names": [ "a", "b", "c" ],
  "monmap": {
    "epoch": 17,
    "fsid": "01234567-abcd-9876-0123-ffeeddccbbaa",
    "modified": "2016-04-11 14:01:52.600198",
    "created": "0.000000",
    "mons": [
      { "rank": 0, "name": "a", "addr": "192.168.0.1:6789/0" },
      { "rank": 1, "name": "b", "addr": "192.168.0.2:6789/0" },
      { "rank": 2, "name": "c", "addr": "192.168.0.3:6789/0" }
    ]
  },
  "osdmap": {
    "osdmap": {
      "epoch": 21734,
      "num_osds": 24,
      "num_up_osds": 24,
      "num_in_osds": 24,
      "full": false,
      "nearfull": false,
      "num_remapped_pgs": 0
    }
  },
  "pgmap": {
    "pgs_by_state": [
      { "state_name": "active+clean", "count": 2560 },
      { "state_name": "active+scrubbing", "count": 10 },
      { "state_name": "active+backfilling", "count": 5 }
    ],
    "version": 52314277,
    "num_pgs": 2560,
    "data_bytes": 2700031960713,
    "bytes_used": 7478347665408,
    "bytes_avail": 9857462382592,
    "bytes_total": 17335810048000,
    "read_bytes_sec": 0,
    "write_bytes_sec": 367217,
    "op_per_sec": 98,
    "read_op_per_sec": 322,
    "write_op_per_sec": 1022
  },
  "mdsmap": {
    "epoch": 1,
    "up": 0,
    "in": 0,
    "max": 0,
    "by_rank": []
  }
}
`

var cephStatusResults = []expectedResult{
	{
		metric: "ceph_osdmap",
		fields: map[string]interface{}{
			"epoch":            float64(21734),
			"num_osds":         float64(24),
			"num_up_osds":      float64(24),
			"num_in_osds":      float64(24),
			"full":             false,
			"nearfull":         false,
			"num_remapped_pgs": float64(0),
		},
		tags: map[string]string{},
	},
	{
		metric: "ceph_pgmap",
		fields: map[string]interface{}{
			"version":          float64(52314277),
			"num_pgs":          float64(2560),
			"data_bytes":       float64(2700031960713),
			"bytes_used":       float64(7478347665408),
			"bytes_avail":      float64(9857462382592),
			"bytes_total":      float64(17335810048000),
			"read_bytes_sec":   float64(0),
			"write_bytes_sec":  float64(367217),
			"op_per_sec":       pf(98),
			"read_op_per_sec":  float64(322),
			"write_op_per_sec": float64(1022),
		},
		tags: map[string]string{},
	},
	{
		metric: "ceph_pgmap_state",
		fields: map[string]interface{}{
			"count": float64(2560),
		},
		tags: map[string]string{
			"state": "active+clean",
		},
	},
	{
		metric: "ceph_pgmap_state",
		fields: map[string]interface{}{
			"count": float64(10),
		},
		tags: map[string]string{
			"state": "active+scrubbing",
		},
	},
	{
		metric: "ceph_pgmap_state",
		fields: map[string]interface{}{
			"count": float64(5),
		},
		tags: map[string]string{
			"state": "active+backfilling",
		},
	},
}

var cephDFDump = `
{ "stats": { "total_space": 472345880,
      "total_used": 71058504,
      "total_avail": 377286864,
      "total_bytes": 472345880,
      "total_used_bytes": 71058504,
      "total_avail_bytes": 377286864},
  "pools": [
        { "name": "data",
          "id": 0,
          "stats": { "kb_used": 0,
              "bytes_used": 0,
              "objects": 0}},
        { "name": "metadata",
          "id": 1,
          "stats": { "kb_used": 25,
              "bytes_used": 25052,
              "objects": 53}},
        { "name": "rbd",
          "id": 2,
          "stats": { "kb_used": 0,
              "bytes_used": 0,
              "objects": 0}},
        { "name": "test",
          "id": 3,
          "stats": { "kb_used": 55476,
              "bytes_used": 56806602,
              "objects": 1}}]}`

var cephDfResults = []expectedResult{
	{
		metric: "ceph_usage",
		fields: map[string]interface{}{
			"total_space":       pf(472345880),
			"total_used":        pf(71058504),
			"total_avail":       pf(377286864),
			"total_bytes":       pf(472345880),
			"total_used_bytes":  pf(71058504),
			"total_avail_bytes": pf(377286864),
		},
		tags: map[string]string{},
	},
	{
		metric: "ceph_pool_usage",
		fields: map[string]interface{}{
			"kb_used":      float64(0),
			"bytes_used":   float64(0),
			"objects":      float64(0),
			"percent_used": (*float64)(nil),
			"max_avail":    (*float64)(nil),
		},
		tags: map[string]string{
			"name": "data",
		},
	},
	{
		metric: "ceph_pool_usage",
		fields: map[string]interface{}{
			"kb_used":      float64(25),
			"bytes_used":   float64(25052),
			"objects":      float64(53),
			"percent_used": (*float64)(nil),
			"max_avail":    (*float64)(nil),
		},
		tags: map[string]string{
			"name": "metadata",
		},
	},
	{
		metric: "ceph_pool_usage",
		fields: map[string]interface{}{
			"kb_used":      float64(0),
			"bytes_used":   float64(0),
			"objects":      float64(0),
			"percent_used": (*float64)(nil),
			"max_avail":    (*float64)(nil),
		},
		tags: map[string]string{
			"name": "rbd",
		},
	},
	{
		metric: "ceph_pool_usage",
		fields: map[string]interface{}{
			"kb_used":      float64(55476),
			"bytes_used":   float64(56806602),
			"objects":      float64(1),
			"percent_used": (*float64)(nil),
			"max_avail":    (*float64)(nil),
		},
		tags: map[string]string{
			"name": "test",
		},
	},
}

var cephODSPoolStatsDump = `
[
    { "pool_name": "data",
      "pool_id": 0,
      "recovery": {},
      "recovery_rate": {},
      "client_io_rate": {}},
    { "pool_name": "metadata",
      "pool_id": 1,
      "recovery": {},
      "recovery_rate": {},
      "client_io_rate": {}},
    { "pool_name": "rbd",
      "pool_id": 2,
      "recovery": {},
      "recovery_rate": {},
      "client_io_rate": {}},
    { "pool_name": "pbench",
      "pool_id": 3,
      "recovery": { "degraded_objects": 18446744073709551562,
          "degraded_total": 412,
          "degrated_ratio": "-13.107"},
      "recovery_rate": { "recovering_objects_per_sec": 279,
          "recovering_bytes_per_sec": 176401059,
          "recovering_keys_per_sec": 0},
      "client_io_rate": { "read_bytes_sec": 10566067,
          "write_bytes_sec": 15165220376,
          "op_per_sec": 9828,
          "read_op_per_sec": 182,
          "write_op_per_sec": 473}}]`

var cephOSDPoolStatsResults = []expectedResult{
	{
		metric: "ceph_pool_stats",
		fields: map[string]interface{}{
			"read_bytes_sec":             float64(0),
			"write_bytes_sec":            float64(0),
			"op_per_sec":                 (*float64)(nil),
			"read_op_per_sec":            float64(0),
			"write_op_per_sec":           float64(0),
			"recovering_objects_per_sec": float64(0),
			"recovering_bytes_per_sec":   float64(0),
			"recovering_keys_per_sec":    float64(0),
		},
		tags: map[string]string{
			"name": "data",
		},
	},
	{
		metric: "ceph_pool_stats",
		fields: map[string]interface{}{
			"read_bytes_sec":             float64(10566067),
			"write_bytes_sec":            float64(15165220376),
			"op_per_sec":                 pf(9828),
			"read_op_per_sec":            float64(182),
			"write_op_per_sec":           float64(473),
			"recovering_objects_per_sec": float64(279),
			"recovering_bytes_per_sec":   float64(176401059),
			"recovering_keys_per_sec":    float64(0),
		},
		tags: map[string]string{
			"name": "pbench",
		},
	},
}

func pf(i float64) *float64 {
	return &i
}
