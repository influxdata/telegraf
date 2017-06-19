package memcached

import (
	"bufio"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemcachedGeneratesMetrics(t *testing.T) {
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
		assert.True(t, acc.HasInt64Field("memcached", metric), metric)
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
		{"pid", "23235"},
		{"uptime", "194"},
		{"time", "1449174679"},
		{"version", "1.4.14 (Ubuntu)"},
		{"libevent", "2.0.21-stable"},
		{"pointer_size", "64"},
		{"rusage_user", "0.000000"},
		{"rusage_system", "0.007566"},
		{"curr_connections", "5"},
		{"total_connections", "6"},
		{"connection_structures", "6"},
		{"reserved_fds", "20"},
		{"cmd_get", "0"},
		{"cmd_set", "0"},
		{"cmd_flush", "0"},
		{"cmd_touch", "0"},
		{"get_hits", "0"},
		{"get_misses", "0"},
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
		{"auth_cmds", "0"},
		{"auth_errors", "0"},
		{"bytes_read", "7"},
		{"bytes_written", "0"},
		{"limit_maxbytes", "67108864"},
		{"accepting_conns", "1"},
		{"listen_disabled_num", "0"},
		{"threads", "4"},
		{"conn_yields", "0"},
		{"hash_power_level", "16"},
		{"hash_bytes", "524288"},
		{"hash_is_expanding", "0"},
		{"expired_unfetched", "0"},
		{"evicted_unfetched", "0"},
		{"bytes", "0"},
		{"curr_items", "0"},
		{"total_items", "0"},
		{"evictions", "0"},
		{"reclaimed", "0"},
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

var memcachedStats = `STAT pid 23235
STAT uptime 194
STAT time 1449174679
STAT version 1.4.14 (Ubuntu)
STAT libevent 2.0.21-stable
STAT pointer_size 64
STAT rusage_user 0.000000
STAT rusage_system 0.007566
STAT curr_connections 5
STAT total_connections 6
STAT connection_structures 6
STAT reserved_fds 20
STAT cmd_get 0
STAT cmd_set 0
STAT cmd_flush 0
STAT cmd_touch 0
STAT get_hits 0
STAT get_misses 0
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
STAT auth_cmds 0
STAT auth_errors 0
STAT bytes_read 7
STAT bytes_written 0
STAT limit_maxbytes 67108864
STAT accepting_conns 1
STAT listen_disabled_num 0
STAT threads 4
STAT conn_yields 0
STAT hash_power_level 16
STAT hash_bytes 524288
STAT hash_is_expanding 0
STAT expired_unfetched 0
STAT evicted_unfetched 0
STAT bytes 0
STAT curr_items 0
STAT total_items 0
STAT evictions 0
STAT reclaimed 0
END
`
