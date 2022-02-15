package memcached

import (
	"bufio"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func TestMemcachedGeneratesMetricsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	m := &Memcached{
		Servers: []string{testutil.GetLocalHost()},
	}

	var acc testutil.Accumulator

	err := acc.GatherError(m.Gather)
	require.NoError(t, err)

	intMetrics := []string{"get_hits", "get_misses", "evictions",
		"limit_maxbytes", "bytes", "uptime", "curr_items", "total_items",
		"curr_connections", "total_connections", "connection_structures", "cmd_get",
		"cmd_set", "delete_hits", "delete_misses", "incr_hits", "incr_misses",
		"decr_hits", "decr_misses", "cas_hits", "cas_misses",
		"bytes_read", "bytes_written", "threads", "conn_yields"}

	for _, metric := range intMetrics {
		require.True(t, acc.HasInt64Field("memcached", metric), metric)
	}
}

func TestMemcachedParseMetrics(t *testing.T) {
	r := bufio.NewReader(strings.NewReader(memcachedStats))
	values, err := parseResponse(r)
	require.NoError(t, err, "Error parsing memcached response")

	tests := []struct {
		key   string
		value string
	}{
		{"pid", "5619"},
		{"uptime", "11"},
		{"time", "1644765868"},
		{"version", "1.6.14_5_ge03751b"},
		{"libevent", "2.1.11-stable"},
		{"pointer_size", "64"},
		{"rusage_user", "0.080905"},
		{"rusage_system", "0.059330"},
		{"max_connections", "1024"},
		{"curr_connections", "2"},
		{"total_connections", "3"},
		{"rejected_connections", "0"},
		{"connection_structures", "3"},
		{"response_obj_oom", "0"},
		{"response_obj_count", "1"},
		{"response_obj_bytes", "16384"},
		{"read_buf_count", "2"},
		{"read_buf_bytes", "32768"},
		{"read_buf_bytes_free", "0"},
		{"read_buf_oom", "0"},
		{"reserved_fds", "20"},
		{"cmd_get", "0"},
		{"cmd_set", "0"},
		{"cmd_flush", "0"},
		{"cmd_touch", "0"},
		{"cmd_meta", "0"},
		{"get_hits", "0"},
		{"get_misses", "0"},
		{"get_expired", "0"},
		{"get_flushed", "0"},
		{"delete_misses", "0"},
		{"delete_hits", "0"},
		{"incr_misses", "0"},
		{"incr_hits", "0"},
		{"decr_misses", "0"},
		{"decr_hits", "0"},
		{"cas_misses", "0"},
		{"cas_hits", "0"},
		{"cas_badval", "0"},
		{"touch_hits", "0"},
		{"touch_misses", "0"},
		{"store_too_large", "0"},
		{"store_no_memory", "0"},
		{"auth_cmds", "0"},
		{"auth_errors", "0"},
		{"bytes_read", "6"},
		{"bytes_written", "0"},
		{"limit_maxbytes", "67108864"},
		{"accepting_conns", "1"},
		{"listen_disabled_num", "0"},
		{"time_in_listen_disabled_us", "0"},
		{"threads", "4"},
		{"conn_yields", "0"},
		{"hash_power_level", "16"},
		{"hash_bytes", "524288"},
		{"hash_is_expanding", "0"},
		{"slab_reassign_rescues", "0"},
		{"slab_reassign_chunk_rescues", "0"},
		{"slab_reassign_evictions_nomem", "0"},
		{"slab_reassign_inline_reclaim", "0"},
		{"slab_reassign_busy_items", "0"},
		{"slab_reassign_busy_deletes", "0"},
		{"slab_reassign_running", "0"},
		{"slabs_moved", "0"},
		{"lru_crawler_running", "0"},
		{"lru_crawler_starts", "1"},
		{"lru_maintainer_juggles", "60"},
		{"malloc_fails", "0"},
		{"log_worker_dropped", "0"},
		{"log_worker_written", "0"},
		{"log_watcher_skipped", "0"},
		{"log_watcher_sent", "0"},
		{"log_watchers", "0"},
		{"unexpected_napi_ids", "0"},
		{"round_robin_fallback", "0"},
		{"bytes", "0"},
		{"curr_items", "0"},
		{"total_items", "0"},
		{"slab_global_page_pool", "0"},
		{"expired_unfetched", "0"},
		{"evicted_unfetched", "0"},
		{"evicted_active", "0"},
		{"evictions", "0"},
		{"reclaimed", "0"},
		{"crawler_reclaimed", "0"},
		{"crawler_items_checked", "0"},
		{"lrutail_reflocked", "0"},
		{"moves_to_cold", "0"},
		{"moves_to_warm", "0"},
		{"moves_within_lru", "0"},
		{"direct_reclaims", "0"},
		{"lru_bumps_dropped", "0"},
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

var memcachedStats = `STAT pid 5619
STAT uptime 11
STAT time 1644765868
STAT version 1.6.14_5_ge03751b
STAT libevent 2.1.11-stable
STAT pointer_size 64
STAT rusage_user 0.080905
STAT rusage_system 0.059330
STAT max_connections 1024
STAT curr_connections 2
STAT total_connections 3
STAT rejected_connections 0
STAT connection_structures 3
STAT response_obj_oom 0
STAT response_obj_count 1
STAT response_obj_bytes 16384
STAT read_buf_count 2
STAT read_buf_bytes 32768
STAT read_buf_bytes_free 0
STAT read_buf_oom 0
STAT reserved_fds 20
STAT cmd_get 0
STAT cmd_set 0
STAT cmd_flush 0
STAT cmd_touch 0
STAT cmd_meta 0
STAT get_hits 0
STAT get_misses 0
STAT get_expired 0
STAT get_flushed 0
STAT delete_misses 0
STAT delete_hits 0
STAT incr_misses 0
STAT incr_hits 0
STAT decr_misses 0
STAT decr_hits 0
STAT cas_misses 0
STAT cas_hits 0
STAT cas_badval 0
STAT touch_hits 0
STAT touch_misses 0
STAT store_too_large 0
STAT store_no_memory 0
STAT auth_cmds 0
STAT auth_errors 0
STAT bytes_read 6
STAT bytes_written 0
STAT limit_maxbytes 67108864
STAT accepting_conns 1
STAT listen_disabled_num 0
STAT time_in_listen_disabled_us 0
STAT threads 4
STAT conn_yields 0
STAT hash_power_level 16
STAT hash_bytes 524288
STAT hash_is_expanding 0
STAT slab_reassign_rescues 0
STAT slab_reassign_chunk_rescues 0
STAT slab_reassign_evictions_nomem 0
STAT slab_reassign_inline_reclaim 0
STAT slab_reassign_busy_items 0
STAT slab_reassign_busy_deletes 0
STAT slab_reassign_running 0
STAT slabs_moved 0
STAT lru_crawler_running 0
STAT lru_crawler_starts 1
STAT lru_maintainer_juggles 60
STAT malloc_fails 0
STAT log_worker_dropped 0
STAT log_worker_written 0
STAT log_watcher_skipped 0
STAT log_watcher_sent 0
STAT log_watchers 0
STAT unexpected_napi_ids 0
STAT round_robin_fallback 0
STAT bytes 0
STAT curr_items 0
STAT total_items 0
STAT slab_global_page_pool 0
STAT expired_unfetched 0
STAT evicted_unfetched 0
STAT evicted_active 0
STAT evictions 0
STAT reclaimed 0
STAT crawler_reclaimed 0
STAT crawler_items_checked 0
STAT lrutail_reflocked 0
STAT moves_to_cold 0
STAT moves_to_warm 0
STAT moves_within_lru 0
STAT direct_reclaims 0
STAT lru_bumps_dropped 0
END
`
