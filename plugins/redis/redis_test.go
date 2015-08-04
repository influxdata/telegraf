package redis

import (
	"bufio"
	"fmt"
	"net"
	"testing"

	"github.com/influxdb/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisGeneratesMetrics(t *testing.T) {
	l, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	defer l.Close()

	go func() {
		c, err := l.Accept()
		if err != nil {
			return
		}

		buf := bufio.NewReader(c)

		for {
			line, err := buf.ReadString('\n')
			if err != nil {
				return
			}

			if line != "info\r\n" {
				return
			}

			fmt.Fprintf(c, "$%d\n", len(testOutput))
			c.Write([]byte(testOutput))
		}
	}()

	addr := fmt.Sprintf("redis://%s", l.Addr().String())

	r := &Redis{
		Servers: []string{addr},
	}

	var acc testutil.Accumulator

	err = r.Gather(&acc)
	require.NoError(t, err)

	checkInt := []struct {
		name  string
		value uint64
	}{
		{"uptime", 238},
		{"clients", 1},
		{"used_memory", 1003936},
		{"used_memory_rss", 811008},
		{"used_memory_peak", 1003936},
		{"used_memory_lua", 33792},
		{"rdb_changes_since_last_save", 0},
		{"total_connections_received", 2},
		{"total_commands_processed", 1},
		{"instantaneous_ops_per_sec", 0},
		{"sync_full", 0},
		{"sync_partial_ok", 0},
		{"sync_partial_err", 0},
		{"expired_keys", 0},
		{"evicted_keys", 0},
		{"keyspace_hits", 0},
		{"keyspace_misses", 0},
		{"pubsub_channels", 0},
		{"pubsub_patterns", 0},
		{"latest_fork_usec", 0},
		{"connected_slaves", 0},
		{"master_repl_offset", 0},
		{"repl_backlog_active", 0},
		{"repl_backlog_size", 1048576},
		{"repl_backlog_histlen", 0},
	}

	for _, c := range checkInt {
		assert.True(t, acc.CheckValue(c.name, c.value))
	}

	checkFloat := []struct {
		name  string
		value float64
	}{
		{"mem_fragmentation_ratio", 0.81},
		{"used_cpu_sys", 0.14},
		{"used_cpu_user", 0.05},
		{"used_cpu_sys_children", 0.00},
		{"used_cpu_user_children", 0.00},
	}

	for _, c := range checkFloat {
		assert.True(t, acc.CheckValue(c.name, c.value))
	}
}

func TestRedisCanPullStatsFromMultipleServers(t *testing.T) {
	l, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	defer l.Close()

	go func() {
		c, err := l.Accept()
		if err != nil {
			return
		}

		buf := bufio.NewReader(c)

		for {
			line, err := buf.ReadString('\n')
			if err != nil {
				return
			}

			if line != "info\r\n" {
				return
			}

			fmt.Fprintf(c, "$%d\n", len(testOutput))
			c.Write([]byte(testOutput))
		}
	}()

	addr := fmt.Sprintf("redis://%s", l.Addr().String())

	r := &Redis{
		Servers: []string{addr},
	}

	var acc testutil.Accumulator

	err = r.Gather(&acc)
	require.NoError(t, err)

	checkInt := []struct {
		name  string
		value uint64
	}{
		{"uptime", 238},
		{"clients", 1},
		{"used_memory", 1003936},
		{"used_memory_rss", 811008},
		{"used_memory_peak", 1003936},
		{"used_memory_lua", 33792},
		{"rdb_changes_since_last_save", 0},
		{"total_connections_received", 2},
		{"total_commands_processed", 1},
		{"instantaneous_ops_per_sec", 0},
		{"sync_full", 0},
		{"sync_partial_ok", 0},
		{"sync_partial_err", 0},
		{"expired_keys", 0},
		{"evicted_keys", 0},
		{"keyspace_hits", 0},
		{"keyspace_misses", 0},
		{"pubsub_channels", 0},
		{"pubsub_patterns", 0},
		{"latest_fork_usec", 0},
		{"connected_slaves", 0},
		{"master_repl_offset", 0},
		{"repl_backlog_active", 0},
		{"repl_backlog_size", 1048576},
		{"repl_backlog_histlen", 0},
	}

	for _, c := range checkInt {
		assert.True(t, acc.CheckValue(c.name, c.value))
	}

	checkFloat := []struct {
		name  string
		value float64
	}{
		{"mem_fragmentation_ratio", 0.81},
		{"used_cpu_sys", 0.14},
		{"used_cpu_user", 0.05},
		{"used_cpu_sys_children", 0.00},
		{"used_cpu_user_children", 0.00},
	}

	for _, c := range checkFloat {
		assert.True(t, acc.CheckValue(c.name, c.value))
	}
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
rejected_connections:0
sync_full:0
sync_partial_ok:0
sync_partial_err:0
expired_keys:0
evicted_keys:0
keyspace_hits:0
keyspace_misses:0
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

`
