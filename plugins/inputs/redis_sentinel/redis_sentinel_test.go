package redis_sentinel

import (
	"bufio"
	"fmt"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestRedisSentinelConnect(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	addr := fmt.Sprintf(testutil.GetLocalHost() + ":26379")

	r := &RedisSentinel{
		Servers: []string{addr},
	}

	var acc testutil.Accumulator

	err := acc.GatherError(r.Gather)
	require.NoError(t, err)
}

func TestRedisSentinelParseInfo(t *testing.T) {
	var acc testutil.Accumulator

	tags := map[string]string{"host": "redis.net"}
	rdr := bufio.NewReader(strings.NewReader(testINFOOutput))

	gatherSentinelInfoOutput(rdr, &acc, tags)

	fields := map[string]interface{}{
		"lru_clock":     int64(15585808),
		"uptime":        int64(901),
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

	acc.AssertContainsTaggedFields(t, "redis_sentinel", fields, tags)
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
