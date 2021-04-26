package disque

import (
	"bufio"
	"fmt"
	"net"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestDisqueGeneratesMetricsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	l, err := net.Listen("tcp", "localhost:0")
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

			if _, err := fmt.Fprintf(c, "$%d\n", len(testOutput)); err != nil {
				return
			}
			if _, err := c.Write([]byte(testOutput)); err != nil {
				return
			}
		}
	}()

	addr := fmt.Sprintf("disque://%s", l.Addr().String())

	r := &Disque{
		Servers: []string{addr},
	}

	var acc testutil.Accumulator

	err = acc.GatherError(r.Gather)
	require.NoError(t, err)

	fields := map[string]interface{}{
		"uptime":                     uint64(1452705),
		"clients":                    uint64(31),
		"blocked_clients":            uint64(13),
		"used_memory":                uint64(1840104),
		"used_memory_rss":            uint64(3227648),
		"used_memory_peak":           uint64(89603656),
		"total_connections_received": uint64(5062777),
		"total_commands_processed":   uint64(12308396),
		"instantaneous_ops_per_sec":  uint64(18),
		"latest_fork_usec":           uint64(1644),
		"registered_jobs":            uint64(360),
		"registered_queues":          uint64(12),
		"mem_fragmentation_ratio":    float64(1.75),
		"used_cpu_sys":               float64(19585.73),
		"used_cpu_user":              float64(11255.96),
		"used_cpu_sys_children":      float64(1.75),
		"used_cpu_user_children":     float64(1.91),
	}
	acc.AssertContainsFields(t, "disque", fields)
}

func TestDisqueCanPullStatsFromMultipleServersIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	l, err := net.Listen("tcp", "localhost:0")
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

			if _, err := fmt.Fprintf(c, "$%d\n", len(testOutput)); err != nil {
				return
			}
			if _, err := c.Write([]byte(testOutput)); err != nil {
				return
			}
		}
	}()

	addr := fmt.Sprintf("disque://%s", l.Addr().String())

	r := &Disque{
		Servers: []string{addr},
	}

	var acc testutil.Accumulator

	err = acc.GatherError(r.Gather)
	require.NoError(t, err)

	fields := map[string]interface{}{
		"uptime":                     uint64(1452705),
		"clients":                    uint64(31),
		"blocked_clients":            uint64(13),
		"used_memory":                uint64(1840104),
		"used_memory_rss":            uint64(3227648),
		"used_memory_peak":           uint64(89603656),
		"total_connections_received": uint64(5062777),
		"total_commands_processed":   uint64(12308396),
		"instantaneous_ops_per_sec":  uint64(18),
		"latest_fork_usec":           uint64(1644),
		"registered_jobs":            uint64(360),
		"registered_queues":          uint64(12),
		"mem_fragmentation_ratio":    float64(1.75),
		"used_cpu_sys":               float64(19585.73),
		"used_cpu_user":              float64(11255.96),
		"used_cpu_sys_children":      float64(1.75),
		"used_cpu_user_children":     float64(1.91),
	}
	acc.AssertContainsFields(t, "disque", fields)
}

const testOutput = `# Server
disque_version:0.0.1
disque_git_sha1:b5247598
disque_git_dirty:0
disque_build_id:379fda78983a60c6
os:Linux 3.13.0-44-generic x86_64
arch_bits:64
multiplexing_api:epoll
gcc_version:4.8.2
process_id:32420
run_id:1cfdfa4c6bc3f285182db5427522a8a4c16e42e4
tcp_port:7711
uptime_in_seconds:1452705
uptime_in_days:16
hz:10
config_file:/usr/local/etc/disque/disque.conf

# Clients
connected_clients:31
client_longest_output_list:0
client_biggest_input_buf:0
blocked_clients:13

# Memory
used_memory:1840104
used_memory_human:1.75M
used_memory_rss:3227648
used_memory_peak:89603656
used_memory_peak_human:85.45M
mem_fragmentation_ratio:1.75
mem_allocator:jemalloc-3.6.0

# Jobs
registered_jobs:360

# Queues
registered_queues:12

# Persistence
loading:0
aof_enabled:1
aof_state:on
aof_rewrite_in_progress:0
aof_rewrite_scheduled:0
aof_last_rewrite_time_sec:0
aof_current_rewrite_time_sec:-1
aof_last_bgrewrite_status:ok
aof_last_write_status:ok
aof_current_size:41952430
aof_base_size:9808
aof_pending_rewrite:0
aof_buffer_length:0
aof_rewrite_buffer_length:0
aof_pending_bio_fsync:0
aof_delayed_fsync:1

# Stats
total_connections_received:5062777
total_commands_processed:12308396
instantaneous_ops_per_sec:18
total_net_input_bytes:1346996528
total_net_output_bytes:1967551763
instantaneous_input_kbps:1.38
instantaneous_output_kbps:1.78
rejected_connections:0
latest_fork_usec:1644

# CPU
used_cpu_sys:19585.73
used_cpu_user:11255.96
used_cpu_sys_children:1.75
used_cpu_user_children:1.91
`
