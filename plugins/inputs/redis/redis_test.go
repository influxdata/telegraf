package redis

import (
	"bufio"
	"fmt"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestRedisConnect(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	addr := fmt.Sprintf(testutil.GetLocalHost() + ":6379")

	r := &Redis{
		Servers: []string{addr},
	}

	var acc testutil.Accumulator

	err := r.Gather(&acc)
	require.NoError(t, err)
}

func TestRedis_ParseMetrics(t *testing.T) {
	var acc testutil.Accumulator
	tags := map[string]string{"host": "redis.net"}
	rdr := bufio.NewReader(strings.NewReader(testOutput))

	err := gatherInfoOutput(rdr, &acc, tags)
	require.NoError(t, err)

	tags = map[string]string{"host": "redis.net", "role": "master"}
	fields := map[string]interface{}{
		"uptime":                      uint64(238),
		"clients":                     uint64(1),
		"used_memory":                 uint64(1003936),
		"used_memory_rss":             uint64(811008),
		"used_memory_peak":            uint64(1003936),
		"used_memory_lua":             uint64(33792),
		"rdb_changes_since_last_save": uint64(0),
		"total_connections_received":  uint64(2),
		"total_commands_processed":    uint64(1),
		"instantaneous_ops_per_sec":   uint64(0),
		"sync_full":                   uint64(0),
		"sync_partial_ok":             uint64(0),
		"sync_partial_err":            uint64(0),
		"expired_keys":                uint64(0),
		"evicted_keys":                uint64(0),
		"keyspace_hits":               uint64(1),
		"keyspace_misses":             uint64(1),
		"pubsub_channels":             uint64(0),
		"pubsub_patterns":             uint64(0),
		"latest_fork_usec":            uint64(0),
		"connected_slaves":            uint64(0),
		"master_repl_offset":          uint64(0),
		"repl_backlog_active":         uint64(0),
		"repl_backlog_size":           uint64(1048576),
		"repl_backlog_histlen":        uint64(0),
		"mem_fragmentation_ratio":     float64(0.81),
		"instantaneous_input_kbps":    float64(876.16),
		"instantaneous_output_kbps":   float64(3010.23),
		"used_cpu_sys":                float64(0.14),
		"used_cpu_user":               float64(0.05),
		"used_cpu_sys_children":       float64(0.00),
		"used_cpu_user_children":      float64(0.00),
		"keyspace_hitrate":            float64(0.50),
	}
	keyspaceTags := map[string]string{"host": "redis.net", "role": "master", "database": "db0"}
	keyspaceFields := map[string]interface{}{
		"avg_ttl": uint64(0),
		"expires": uint64(0),
		"keys":    uint64(2),
	}
	acc.AssertContainsTaggedFields(t, "redis", fields, tags)
	acc.AssertContainsTaggedFields(t, "redis_keyspace", keyspaceFields, keyspaceTags)
}

const testOutput = `# Server
redis_version:2.8.9
redis_git_sha1:00000000
redis_git_dirty:0
redis_build_id:9ccc8119ea98f6e1
redis_mode:standalone
os:Darwin 14.1.0 x86_64
arch_bits:64
multiplexing_api:kqueue
gcc_version:4.2.1
process_id:40235
run_id:37d020620aadf0627282c0f3401405d774a82664
tcp_port:6379
uptime_in_seconds:238
uptime_in_days:0
hz:10
lru_clock:2364819
config_file:/usr/local/etc/redis.conf

# Clients
connected_clients:1
client_longest_output_list:0
client_biggest_input_buf:0
blocked_clients:0

# Memory
used_memory:1003936
used_memory_human:980.41K
used_memory_rss:811008
used_memory_peak:1003936
used_memory_peak_human:980.41K
used_memory_lua:33792
mem_fragmentation_ratio:0.81
mem_allocator:libc

# Persistence
loading:0
rdb_changes_since_last_save:0
rdb_bgsave_in_progress:0
rdb_last_save_time:1428427941
rdb_last_bgsave_status:ok
rdb_last_bgsave_time_sec:-1
rdb_current_bgsave_time_sec:-1
aof_enabled:0
aof_rewrite_in_progress:0
aof_rewrite_scheduled:0
aof_last_rewrite_time_sec:-1
aof_current_rewrite_time_sec:-1
aof_last_bgrewrite_status:ok
aof_last_write_status:ok

# Stats
total_connections_received:2
total_commands_processed:1
instantaneous_ops_per_sec:0
instantaneous_input_kbps:876.16
instantaneous_output_kbps:3010.23
rejected_connections:0
sync_full:0
sync_partial_ok:0
sync_partial_err:0
expired_keys:0
evicted_keys:0
keyspace_hits:1
keyspace_misses:1
pubsub_channels:0
pubsub_patterns:0
latest_fork_usec:0

# Replication
role:master
connected_slaves:0
master_repl_offset:0
repl_backlog_active:0
repl_backlog_size:1048576
repl_backlog_first_byte_offset:0
repl_backlog_histlen:0

# CPU
used_cpu_sys:0.14
used_cpu_user:0.05
used_cpu_sys_children:0.00
used_cpu_user_children:0.00

# Keyspace
db0:keys=2,expires=0,avg_ttl=0

(error) ERR unknown command 'eof'
`
