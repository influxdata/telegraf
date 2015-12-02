package memcached

import (
	"testing"

	"github.com/influxdb/telegraf/testutil"
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

	err := m.Gather(&acc)
	require.NoError(t, err)

	intMetrics := []string{"get_hits", "get_misses", "evictions", "limit_maxbytes", "bytes", "uptime", "curr_items", "total_items", "curr_connections", "total_connections", "connection_structures", "cmd_get", "cmd_set", "delete_hits", "delete_misses", "incr_hits", "incr_misses", "decr_hits", "decr_misses", "cas_hits", "cas_misses", "cas_badvalue", "evictions", "bytes_read", "bytes_written", "threads", "conn_yields"}

	for _, metric := range intMetrics {
		assert.True(t, acc.HasIntValue(metric), metric)
	}
}
