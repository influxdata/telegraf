package redis_sentinel

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
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
	now := time.Now()

	globalTags := map[string]string{
		"port":   "6379",
		"source": "redis.io",
	}

	expectedTags := map[string]string{
		"port":   "6379",
		"source": "redis.io",
		"master": masterName,
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
		"num_other_sentinels":     1,
		"num_slaves":              0,
		"parallel_syncs":          1,
		"port":                    6379,
		"quorum":                  2,
		"role_reported":           "master",
		"role_reported_time":      83138826,
		"has_quorum":              true,
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

	sentinelTags, sentinelFields, sentinalErr := convertSentinelMastersOutput(globalTags, sentinelMastersOutput, nil)
	require.NoErrorf(t, sentinalErr, "failed converting output: %v", sentinalErr)

	actualMetrics := []telegraf.Metric{
		testutil.MustMetric(measurementMasters, sentinelTags, sentinelFields, now),
	}

	testutil.RequireMetricsEqual(t, expectedMetrics, actualMetrics, testutil.IgnoreTime())
}

func TestRedisSentinels(t *testing.T) {
	now := time.Now()

	globalTags := make(map[string]string)

	expectedTags := map[string]string{
		"sentinel_ip":   "127.0.0.1",
		"sentinel_port": "26380",
		"master":        masterName,
	}
	expectedFields := map[string]interface{}{
		"name":                    "adfd343f6b6ecc77e2b9636de6d9f28d4b827521",
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

	sentinelTags, sentinelFields, sentinelErr := convertSentinelSentinelsOutput(globalTags, masterName, sentinelsOutput)
	require.NoErrorf(t, sentinelErr, "failed converting output: %v", sentinelErr)

	actualMetrics := []telegraf.Metric{
		testutil.MustMetric(measurementSentinels, sentinelTags, sentinelFields, now),
	}

	testutil.RequireMetricsEqual(t, expectedMetrics, actualMetrics)
}

func TestRedisSentinelReplicas(t *testing.T) {
	now := time.Now()

	globalTags := make(map[string]string)

	expectedTags := map[string]string{
		"replica_ip":   "127.0.0.1",
		"replica_port": "6380",
		"master":       masterName,
	}
	expectedFields := map[string]interface{}{
		"down_after_milliseconds": 30000,
		"flags":                   "slave",
		"info_refresh":            8476,
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
		"role_reported":           "slave",
		"role_reported_time":      10267432,
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

	sentinelTags, sentinelFields, sentinalErr := convertSentinelReplicaOutput(globalTags, masterName, replicasOutput)
	require.NoErrorf(t, sentinalErr, "failed converting output: %v", sentinalErr)

	actualMetrics := []telegraf.Metric{
		testutil.MustMetric(measurementReplicas, sentinelTags, sentinelFields, now),
	}

	testutil.RequireMetricsEqual(t, expectedMetrics, actualMetrics)
}

func TestRedisSentinelInfoAll(t *testing.T) {
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

	sentinelInfoResponse, err := os.ReadFile("testdata/sentinel.info.response")
	require.NoErrorf(t, err, "could not init fixture: %v", err)

	rdr := bufio.NewReader(bytes.NewReader(sentinelInfoResponse))

	sentinelTags, sentinelFields, sentinalErr := convertSentinelInfoOutput(globalTags, rdr)
	require.NoErrorf(t, sentinalErr, "failed converting output: %v", sentinalErr)

	actualMetrics := []telegraf.Metric{
		testutil.MustMetric(measurementSentinel, sentinelTags, sentinelFields, now),
	}

	testutil.RequireMetricsEqual(t, expectedMetrics, actualMetrics)
}
