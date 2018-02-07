package redis

import (
	"bufio"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
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

	err := acc.GatherError(r.Gather)
	require.NoError(t, err)
}

func TestRedis_ParseMetrics(t *testing.T) {
	var acc testutil.Accumulator
	tags := map[string]string{"host": "redis.net"}
	rdr := bufio.NewReader(strings.NewReader(testOutput))

	err := gatherInfoOutput(rdr, &acc, tags)
	require.NoError(t, err)

	tags = map[string]string{"host": "redis.net", "replication_role": "master"}
	fields := map[string]interface{}{
		"uptime":                         int64(238),
		"lru_clock":                      int64(2364819),
		"clients":                        int64(1),
		"client_longest_output_list":     int64(0),
		"client_biggest_input_buf":       int64(0),
		"blocked_clients":                int64(0),
		"used_memory":                    int64(1003936),
		"used_memory_rss":                int64(811008),
		"used_memory_peak":               int64(1003936),
		"used_memory_lua":                int64(33792),
		"mem_fragmentation_ratio":        float64(0.81),
		"loading":                        int64(0),
		"rdb_changes_since_last_save":    int64(0),
		"rdb_bgsave_in_progress":         int64(0),
		"rdb_last_save_time":             int64(1428427941),
		"rdb_last_bgsave_status":         "ok",
		"rdb_last_bgsave_time_sec":       int64(-1),
		"rdb_current_bgsave_time_sec":    int64(-1),
		"aof_enabled":                    int64(0),
		"aof_rewrite_in_progress":        int64(0),
		"aof_rewrite_scheduled":          int64(0),
		"aof_last_rewrite_time_sec":      int64(-1),
		"aof_current_rewrite_time_sec":   int64(-1),
		"aof_last_bgrewrite_status":      "ok",
		"aof_last_write_status":          "ok",
		"total_connections_received":     int64(2),
		"total_commands_processed":       int64(1),
		"instantaneous_ops_per_sec":      int64(0),
		"instantaneous_input_kbps":       float64(876.16),
		"instantaneous_output_kbps":      float64(3010.23),
		"rejected_connections":           int64(0),
		"sync_full":                      int64(0),
		"sync_partial_ok":                int64(0),
		"sync_partial_err":               int64(0),
		"expired_keys":                   int64(0),
		"evicted_keys":                   int64(0),
		"keyspace_hits":                  int64(1),
		"keyspace_misses":                int64(1),
		"pubsub_channels":                int64(0),
		"pubsub_patterns":                int64(0),
		"latest_fork_usec":               int64(0),
		"connected_slaves":               int64(0),
		"master_repl_offset":             int64(0),
		"repl_backlog_active":            int64(0),
		"repl_backlog_size":              int64(1048576),
		"repl_backlog_first_byte_offset": int64(0),
		"repl_backlog_histlen":           int64(0),
		"second_repl_offset":             int64(-1),
		"used_cpu_sys":                   float64(0.14),
		"used_cpu_user":                  float64(0.05),
		"used_cpu_sys_children":          float64(0.00),
		"used_cpu_user_children":         float64(0.00),
		"keyspace_hitrate":               float64(0.50),
		"redis_version":                  "2.8.9",
	}

	// We have to test rdb_last_save_time_offset manually because the value is based on the time when gathered
	for _, m := range acc.Metrics {
		for k, v := range m.Fields {
			if k == "rdb_last_save_time_elapsed" {
				fields[k] = v
			}
		}
	}
	assert.InDelta(t,
		time.Now().Unix()-fields["rdb_last_save_time"].(int64),
		fields["rdb_last_save_time_elapsed"].(int64),
		2) // allow for 2 seconds worth of offset

	keyspaceTags := map[string]string{"host": "redis.net", "replication_role": "master", "database": "db0"}
	keyspaceFields := map[string]interface{}{
		"avg_ttl": int64(0),
		"expires": int64(0),
		"keys":    int64(2),
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
master_replid:8c4d7b768b26826825ceb20ff4a2c7c54616350b
master_replid2:0000000000000000000000000000000000000000
master_repl_offset:0
second_repl_offset:-1
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
