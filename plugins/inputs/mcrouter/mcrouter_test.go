package mcrouter

import (
	"bufio"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func TestAddressParsing(t *testing.T) {
	m := &Mcrouter{
		Servers: []string{"tcp://" + testutil.GetLocalHost()},
	}

	var acceptTests = [][3]string{
		{"tcp://localhost:8086", "localhost:8086", "tcp"},
		{"tcp://localhost", "localhost:" + defaultServerURL.Port(), "tcp"},
		{"tcp://localhost:", "localhost:" + defaultServerURL.Port(), "tcp"},
		{"tcp://:8086", defaultServerURL.Hostname() + ":8086", "tcp"},
		{"tcp://:", defaultServerURL.Host, "tcp"},
	}

	var rejectTests = []string{
		"tcp://",
	}

	for _, args := range acceptTests {
		address, protocol, err := m.ParseAddress(args[0])

		require.Nil(t, err, args[0])
		require.Equal(t, args[1], address, args[0])
		require.Equal(t, args[2], protocol, args[0])
	}

	for _, addr := range rejectTests {
		address, protocol, err := m.ParseAddress(addr)

		require.NotNil(t, err, addr)
		require.Empty(t, address, addr)
		require.Empty(t, protocol, addr)
	}
}

func TestMcrouterGeneratesMetricsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	m := &Mcrouter{
		Servers: []string{"tcp://" + testutil.GetLocalHost()},
	}

	var acc testutil.Accumulator

	err := acc.GatherError(m.Gather)
	require.NoError(t, err)

	intMetrics := []string{
		"uptime",
		// "num_servers",
		// "num_servers_new",
		// "num_servers_up",
		// "num_servers_down",
		// "num_servers_closed",
		// "num_clients",
		// "num_suspect_servers",
		// "destination_batches_sum",
		// "destination_requests_sum",
		// "outstanding_route_get_reqs_queued",
		// "outstanding_route_update_reqs_queued",
		// "outstanding_route_get_avg_queue_size",
		// "outstanding_route_update_avg_queue_size",
		// "outstanding_route_get_avg_wait_time_sec",
		// "outstanding_route_update_avg_wait_time_sec",
		// "retrans_closed_connections",
		// "destination_pending_reqs",
		// "destination_inflight_reqs",
		// "destination_batch_size",
		// "asynclog_requests",
		// "proxy_reqs_processing",
		// "proxy_reqs_waiting",
		// "client_queue_notify_period",
		// "ps_num_minor_faults",
		// "ps_num_major_faults",
		// "ps_vsize",
		// "ps_rss",
		// "fibers_allocated",
		// "fibers_pool_size",
		// "fibers_stack_high_watermark",
		// "successful_client_connections",
		// "duration_us",
		// "destination_max_pending_reqs",
		// "destination_max_inflight_reqs",
		// "retrans_per_kbyte_max",
		// "cmd_get_count",
		// "cmd_delete_out",
		// "cmd_lease_get",
		"cmd_set",
		// "cmd_get_out_all",
		// "cmd_get_out",
		// "cmd_lease_set_count",
		// "cmd_other_out_all",
		// "cmd_lease_get_out",
		// "cmd_set_count",
		// "cmd_lease_set_out",
		// "cmd_delete_count",
		// "cmd_other",
		// "cmd_delete",
		"cmd_get",
		// "cmd_lease_set",
		// "cmd_set_out",
		// "cmd_lease_get_count",
		// "cmd_other_out",
		// "cmd_lease_get_out_all",
		// "cmd_set_out_all",
		// "cmd_other_count",
		// "cmd_delete_out_all",
		// "cmd_lease_set_out_all"
	}

	floatMetrics := []string{
		"rusage_system",
		"rusage_user",
		// "ps_user_time_sec",
		// "ps_system_time_sec",
	}

	for _, metric := range intMetrics {
		require.True(t, acc.HasInt64Field("mcrouter", metric), metric)
	}

	for _, metric := range floatMetrics {
		require.True(t, acc.HasFloatField("mcrouter", metric), metric)
	}
}

func TestMcrouterParseMetrics(t *testing.T) {
	r := bufio.NewReader(strings.NewReader(mcrouterStats))
	scanner := bufio.NewScanner(r)
	values, err := parseResponse(scanner)
	require.NoError(t, err, "Error parsing mcrouter response")

	tests := []struct {
		key   string
		value string
	}{
		{"uptime", "166"},
		{"num_servers", "1"},
		{"num_servers_new", "1"},
		{"num_servers_up", "0"},
		{"num_servers_down", "0"},
		{"num_servers_closed", "0"},
		{"num_clients", "1"},
		{"num_suspect_servers", "0"},
		{"destination_batches_sum", "0"},
		{"destination_requests_sum", "0"},
		{"outstanding_route_get_reqs_queued", "0"},
		{"outstanding_route_update_reqs_queued", "0"},
		{"outstanding_route_get_avg_queue_size", "0"},
		{"outstanding_route_update_avg_queue_size", "0"},
		{"outstanding_route_get_avg_wait_time_sec", "0"},
		{"outstanding_route_update_avg_wait_time_sec", "0"},
		{"retrans_closed_connections", "0"},
		{"destination_pending_reqs", "0"},
		{"destination_inflight_reqs", "0"},
		{"destination_batch_size", "0"},
		{"asynclog_requests", "0"},
		{"proxy_reqs_processing", "1"},
		{"proxy_reqs_waiting", "0"},
		{"client_queue_notify_period", "0"},
		{"rusage_system", "0.040966"},
		{"rusage_user", "0.020483"},
		{"ps_num_minor_faults", "2490"},
		{"ps_num_major_faults", "11"},
		{"ps_user_time_sec", "0.02"},
		{"ps_system_time_sec", "0.04"},
		{"ps_vsize", "697741312"},
		{"ps_rss", "10563584"},
		{"fibers_allocated", "0"},
		{"fibers_pool_size", "0"},
		{"fibers_stack_high_watermark", "0"},
		{"successful_client_connections", "18"},
		{"duration_us", "0"},
		{"destination_max_pending_reqs", "0"},
		{"destination_max_inflight_reqs", "0"},
		{"retrans_per_kbyte_max", "0"},
		{"cmd_get_count", "0"},
		{"cmd_delete_out", "0"},
		{"cmd_lease_get", "0"},
		{"cmd_set", "0"},
		{"cmd_get_out_all", "0"},
		{"cmd_get_out", "0"},
		{"cmd_lease_set_count", "0"},
		{"cmd_other_out_all", "0"},
		{"cmd_lease_get_out", "0"},
		{"cmd_set_count", "0"},
		{"cmd_lease_set_out", "0"},
		{"cmd_delete_count", "0"},
		{"cmd_other", "0"},
		{"cmd_delete", "0"},
		{"cmd_get", "0"},
		{"cmd_lease_set", "0"},
		{"cmd_set_out", "0"},
		{"cmd_lease_get_count", "0"},
		{"cmd_other_out", "0"},
		{"cmd_lease_get_out_all", "0"},
		{"cmd_set_out_all", "0"},
		{"cmd_other_count", "0"},
		{"cmd_delete_out_all", "0"},
		{"cmd_lease_set_out_all", "0"},
	}

	for _, test := range tests {
		value, ok := values[test.key]
		if !ok {
			t.Errorf("Did not find key for metric %s in values", test.key)
			continue
		}
		if value != test.value {
			t.Errorf("Metric: %s, Expected: %s, actual: %s",
				test.key, test.value, value)
		}
	}
}

var mcrouterStats = `STAT version 36.0.0 mcrouter
STAT commandargs --port 11211 --config-file /etc/mcrouter/mcrouter.json --async-dir /var/spool/mcrouter --log-path /var/log/mcrouter/mcrouter.log --stats-root /var/mcrouter/stats --server-timeout 100 --reset-inactive-connection-interval 10000 --proxy-threads auto
STAT pid 21357
STAT parent_pid 1
STAT time 1524673265
STAT uptime 166
STAT num_servers 1
STAT num_servers_new 1
STAT num_servers_up 0
STAT num_servers_down 0
STAT num_servers_closed 0
STAT num_clients 1
STAT num_suspect_servers 0
STAT destination_batches_sum 0
STAT destination_requests_sum 0
STAT outstanding_route_get_reqs_queued 0
STAT outstanding_route_update_reqs_queued 0
STAT outstanding_route_get_avg_queue_size 0
STAT outstanding_route_update_avg_queue_size 0
STAT outstanding_route_get_avg_wait_time_sec 0
STAT outstanding_route_update_avg_wait_time_sec 0
STAT retrans_closed_connections 0
STAT destination_pending_reqs 0
STAT destination_inflight_reqs 0
STAT destination_batch_size 0
STAT asynclog_requests 0
STAT proxy_reqs_processing 1
STAT proxy_reqs_waiting 0
STAT client_queue_notify_period 0
STAT rusage_system 0.040966
STAT rusage_user 0.020483
STAT ps_num_minor_faults 2490
STAT ps_num_major_faults 11
STAT ps_user_time_sec 0.02
STAT ps_system_time_sec 0.04
STAT ps_vsize 697741312
STAT ps_rss 10563584
STAT fibers_allocated 0
STAT fibers_pool_size 0
STAT fibers_stack_high_watermark 0
STAT successful_client_connections 18
STAT duration_us 0
STAT destination_max_pending_reqs 0
STAT destination_max_inflight_reqs 0
STAT retrans_per_kbyte_max 0
STAT cmd_get_count 0
STAT cmd_delete_out 0
STAT cmd_lease_get 0
STAT cmd_set 0
STAT cmd_get_out_all 0
STAT cmd_get_out 0
STAT cmd_lease_set_count 0
STAT cmd_other_out_all 0
STAT cmd_lease_get_out 0
STAT cmd_set_count 0
STAT cmd_lease_set_out 0
STAT cmd_delete_count 0
STAT cmd_other 0
STAT cmd_delete 0
STAT cmd_get 0
STAT cmd_lease_set 0
STAT cmd_set_out 0
STAT cmd_lease_get_count 0
STAT cmd_other_out 0
STAT cmd_lease_get_out_all 0
STAT cmd_set_out_all 0
STAT cmd_other_count 0
STAT cmd_delete_out_all 0
STAT cmd_lease_set_out_all 0
END
`
