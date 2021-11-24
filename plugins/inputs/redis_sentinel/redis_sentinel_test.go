package redis_sentinel

import (
	"bufio"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/require"
)

const masterName = "mymaster"

func TestRedisSentinelConnect(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	addr := fmt.Sprintf("tcp://" + testutil.GetLocalHost() + ":26379")

	r := &RedisSentinel{
		Servers: []string{addr},
	}

	var acc testutil.Accumulator

	err := acc.GatherError(r.Gather)
	require.NoError(t, err)
}

func TestRedisSentinelMasters(t *testing.T) {
	t.Logf("Redis Sentinel: 'sentinel masters <name>'")

	now := time.Now()

	globalTags := map[string]string{
		"port":   "6379",
		"source": "redis.io",
	}

	expectedTags := map[string]string{
		"port":        "6379",
		"source":      "redis.io",
		"master_name": masterName,
	}

	// has_quorum is a custom field
	expectedFields := map[string]interface{}{
		"config_epoch":            0,
		"down_after_milliseconds": 30000,
		"failover_timeout":        180000,
		"flags":                   "master",
		"info_refresh":            8819,
		"ip":                      "127.0.0.1",
		"last_ok_ping_reply":      174,
		"last_ping_reply":         174,
		"last_ping_sent":          0,
		"link_pending_commands":   0,
		"link_refcount":           1,
		"name":                    "mymaster",
		"num_other_sentinels":     1,
		"num_slaves":              0,
		"parallel_syncs":          1,
		"port":                    6379,
		"quorum":                  2,
		"role_reported":           "master",
		"role_reported_time":      83138826,
		"runid":                   "ff3dadd1cfea3043de4d25711d93f01a564562f7",
		"has_quorum":              1,
	}

	expectedMetrics := []telegraf.Metric{
		testutil.MustMetric(measurementMasters, expectedTags, expectedFields, now),
	}

	sentinelMastersOutput := map[string]string{
		"config_epoch":            "0",
		"down_after_milliseconds": "30000",
		"failover_timeout":        "180000",
		"flags":                   "master",
		"info_refresh":            "8819",
		"ip":                      "127.0.0.1",
		"last_ok_ping_reply":      "174",
		"last_ping_reply":         "174",
		"last_ping_sent":          "0",
		"link_pending_commands":   "0",
		"link_refcount":           "1",
		"name":                    "mymaster",
		"num_other_sentinels":     "1",
		"num_slaves":              "0",
		"parallel_syncs":          "1",
		"port":                    "6379",
		"quorum":                  "2",
		"role_reported":           "master",
		"role_reported_time":      "83138826",
		"runid":                   "ff3dadd1cfea3043de4d25711d93f01a564562f7",
	}

	smTags, smFields := convertSentinelMastersOutput(globalTags, sentinelMastersOutput, nil)
	actualMetrics := []telegraf.Metric{
		testutil.MustMetric(measurementMasters, smTags, smFields, now),
	}

	testutil.RequireMetricsEqual(t, expectedMetrics, actualMetrics, testutil.IgnoreTime())
}

func TestRedisSentinels(t *testing.T) {
	t.Logf("Redis Sentinel: 'sentinel sentinels <name>'")

	now := time.Now()

	globalTags := make(map[string]string)

	expectedTags := map[string]string{
		"sentinel_ip":   "127.0.0.1",
		"sentinel_port": "26380",
		"master_name":   masterName,
	}
	expectedFields := map[string]interface{}{
		"name":                    "adfd343f6b6ecc77e2b9636de6d9f28d4b827521",
		"ip":                      "127.0.0.1",
		"port":                    26380,
		"runid":                   "adfd343f6b6ecc77e2b9636de6d9f28d4b827521",
		"flags":                   "sentinel",
		"link_pending_commands":   0,
		"link_refcount":           1,
		"last_ping_sent":          0,
		"last_ok_ping_reply":      516,
		"last_ping_reply":         516,
		"down_after_milliseconds": 30000,
		"last_hello_message":      1905,
		"voted_leader":            "?",
		"voted_leader_epoch":      0,
	}

	expectedMetrics := []telegraf.Metric{
		testutil.MustMetric(measurementSentinels, expectedTags, expectedFields, now),
	}

	sentinelsOutput := map[string]string{
		"name":                    "adfd343f6b6ecc77e2b9636de6d9f28d4b827521",
		"ip":                      "127.0.0.1",
		"port":                    "26380",
		"runid":                   "adfd343f6b6ecc77e2b9636de6d9f28d4b827521",
		"flags":                   "sentinel",
		"link_pending_commands":   "0",
		"link_refcount":           "1",
		"last_ping_sent":          "0",
		"last_ok_ping_reply":      "516",
		"last_ping_reply":         "516",
		"down_after_milliseconds": "30000",
		"last_hello_message":      "1905",
		"voted_leader":            "?",
		"voted_leader_epoch":      "0",
	}

	sentinelTags, sentinelFields := convertSentinelSentinelsOutput(globalTags, masterName, sentinelsOutput)
	actualMetrics := []telegraf.Metric{
		testutil.MustMetric(measurementSentinels, sentinelTags, sentinelFields, now),
	}

	testutil.RequireMetricsEqual(t, expectedMetrics, actualMetrics)
}

func TestRedisSentinelReplicas(t *testing.T) {
	t.Logf("Redis Sentinel: 'sentinel replicas <name>'")

	now := time.Now()

	globalTags := make(map[string]string)

	expectedTags := map[string]string{
		"replica_ip":   "127.0.0.1",
		"replica_port": "6380",
		"master_name":  masterName,
	}
	expectedFields := map[string]interface{}{
		"down_after_milliseconds": 30000,
		"flags":                   "slave",
		"info_refresh":            8476,
		"ip":                      "127.0.0.1",
		"last_ok_ping_reply":      987,
		"last_ping_reply":         987,
		"last_ping_sent":          0,
		"link_pending_commands":   0,
		"link_refcount":           1,
		"master_host":             "127.0.0.1",
		"master_link_down_time":   0,
		"master_link_status":      "ok",
		"master_port":             6379,
		"name":                    "127.0.0.1:6380",
		"port":                    6380,
		"role_reported":           "slave",
		"role_reported_time":      10267432,
		"runid":                   "70e07dad9e450e2d35f1b75338e0a5341b59d710",
		"slave_priority":          100,
		"slave_repl_offset":       1392400,
	}

	expectedMetrics := []telegraf.Metric{
		testutil.MustMetric(measurementReplicas, expectedTags, expectedFields, now),
	}

	replicasOutput := map[string]string{
		"down_after_milliseconds": "30000",
		"flags":                   "slave",
		"info_refresh":            "8476",
		"ip":                      "127.0.0.1",
		"last_ok_ping_reply":      "987",
		"last_ping_reply":         "987",
		"last_ping_sent":          "0",
		"link_pending_commands":   "0",
		"link_refcount":           "1",
		"master_host":             "127.0.0.1",
		"master_link_down_time":   "0",
		"master_link_status":      "ok",
		"master_port":             "6379",
		"name":                    "127.0.0.1:6380",
		"port":                    "6380",
		"role_reported":           "slave",
		"role_reported_time":      "10267432",
		"runid":                   "70e07dad9e450e2d35f1b75338e0a5341b59d710",
		"slave_priority":          "100",
		"slave_repl_offset":       "1392400",
	}

	sentinelTags, sentinelFields := convertSentinelReplicaOutput(globalTags, masterName, replicasOutput)
	actualMetrics := []telegraf.Metric{
		testutil.MustMetric(measurementReplicas, sentinelTags, sentinelFields, now),
	}

	testutil.RequireMetricsEqual(t, expectedMetrics, actualMetrics)
}

func TestRedisSentinelInfoAll(t *testing.T) {
	t.Logf("Redis Sentinel: 'info all'")

	now := time.Now()

	globalTags := map[string]string{
		"port":   "6379",
		"source": "redis.io",
	}

	expectedTags := map[string]string{
		"port":   "6379",
		"source": "redis.io",
	}

	expectedFields := map[string]interface{}{
		"lru_clock":     int64(15585808),
		"uptime_ns":     int64(901000000000),
		"redis_version": "5.0.5",

		"clients":                         int64(2),
		"client_recent_max_input_buffer":  int64(2),
		"client_recent_max_output_buffer": int64(0),
		"blocked_clients":                 int64(0),

		"used_cpu_sys":           float64(0.786872),
		"used_cpu_user":          float64(0.939455),
		"used_cpu_sys_children":  float64(0.000000),
		"used_cpu_user_children": float64(0.000000),

		"total_connections_received":     int64(2),
		"total_commands_processed":       int64(6),
		"instantaneous_ops_per_sec":      int64(0),
		"total_net_input_bytes":          int64(124),
		"total_net_output_bytes":         int64(10148),
		"instantaneous_input_kbps":       float64(0.00),
		"instantaneous_output_kbps":      float64(0.00),
		"rejected_connections":           int64(0),
		"sync_full":                      int64(0),
		"sync_partial_ok":                int64(0),
		"sync_partial_err":               int64(0),
		"expired_keys":                   int64(0),
		"expired_stale_perc":             float64(0.00),
		"expired_time_cap_reached_count": int64(0),
		"evicted_keys":                   int64(0),
		"keyspace_hits":                  int64(0),
		"keyspace_misses":                int64(0),
		"pubsub_channels":                int64(0),
		"pubsub_patterns":                int64(0),
		"latest_fork_usec":               int64(0),
		"migrate_cached_sockets":         int64(0),
		"slave_expires_tracked_keys":     int64(0),
		"active_defrag_hits":             int64(0),
		"active_defrag_misses":           int64(0),
		"active_defrag_key_hits":         int64(0),
		"active_defrag_key_misses":       int64(0),

		"sentinel_masters":                int64(2),
		"sentinel_running_scripts":        int64(0),
		"sentinel_scripts_queue_length":   int64(0),
		"sentinel_simulate_failure_flags": int64(0),
		"sentinel_tilt":                   int64(0),
	}

	expectedMetrics := []telegraf.Metric{
		testutil.MustMetric(measurementSentinel, expectedTags, expectedFields, now),
	}

	const testINFOOutput = `
# Server
redis_version:5.0.5
redis_git_sha1:00000000
redis_git_dirty:0
redis_build_id:78473e0efb96880a
redis_mode:sentinel
os:Linux 5.1.3-arch1-1-ARCH x86_64
arch_bits:64
multiplexing_api:epoll
atomicvar_api:atomic-builtin
gcc_version:8.3.0
process_id:2837
run_id:ecbbb2ca0035a532b03748fbec9f3f8ca1967536
tcp_port:26379
uptime_in_seconds:901
uptime_in_days:0
hz:10
configured_hz:10
lru_clock:15585808
executable:/home/adam/redis-sentinel
config_file:/home/adam/rs1.conf

# Clients
connected_clients:2
client_recent_max_input_buffer:2
client_recent_max_output_buffer:0
blocked_clients:0

# CPU
used_cpu_sys:0.786872
used_cpu_user:0.939455
used_cpu_sys_children:0.000000
used_cpu_user_children:0.000000

# Stats
total_connections_received:2
total_commands_processed:6
instantaneous_ops_per_sec:0
total_net_input_bytes:124
total_net_output_bytes:10148
instantaneous_input_kbps:0.00
instantaneous_output_kbps:0.00
rejected_connections:0
sync_full:0
sync_partial_ok:0
sync_partial_err:0
expired_keys:0
expired_stale_perc:0.00
expired_time_cap_reached_count:0
evicted_keys:0
keyspace_hits:0
keyspace_misses:0
pubsub_channels:0
pubsub_patterns:0
latest_fork_usec:0
migrate_cached_sockets:0
slave_expires_tracked_keys:0
active_defrag_hits:0
active_defrag_misses:0
active_defrag_key_hits:0
active_defrag_key_misses:0

# Sentinel
sentinel_masters:2
sentinel_tilt:0
sentinel_running_scripts:0
sentinel_scripts_queue_length:0
sentinel_simulate_failure_flags:0
master0:name=myothermaster,status=ok,address=127.0.0.1:6380,slaves=1,sentinels=2
master0:name=myothermaster,status=ok,address=127.0.0.1:6381,slaves=1,sentinels=2
master1:name=mymaster,status=ok,address=127.0.0.1:6379,slaves=1,sentinels=1
`

	var acc testutil.Accumulator
	rdr := bufio.NewReader(strings.NewReader(testINFOOutput))

	infoTags, infoFields := convertSentinelInfoOutput(&acc, globalTags, rdr)

	actualMetrics := []telegraf.Metric{
		testutil.MustMetric(measurementSentinel, infoTags, infoFields, now),
	}

	testutil.RequireMetricsEqual(t, expectedMetrics, actualMetrics)
}
