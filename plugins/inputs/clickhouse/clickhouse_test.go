package clickhouse

import (
	"fmt"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClickHouseGeneratesMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	ch := ClickHouse{
		DSN:          fmt.Sprintf("native://%s:9000", testutil.GetLocalHost()),
		connect:      &connect{},
		clustersConn: make(map[string]*connect),
	}
	var acc testutil.Accumulator
	{
		require.NoError(t, ch.Start(&acc))
		require.NoError(t, ch.Gather(&acc))
	}

	for _, event := range []string{
		"FileOpen",
		"ReadBufferFromFileDescriptorRead",
		"IOBufferAllocs",
		"IOBufferAllocBytes",
	} {
		assert.True(t, acc.HasUIntField("clickhouse_events", event))
	}
	for _, metric := range []string{
		"Query",
		"Merge",
		"TCPConnection",
		"HTTPConnection",
	} {
		assert.True(t, acc.HasUIntField("clickhouse_metrics", metric))
	}
	for _, metric := range []string{
		"tcmalloc.pageheap_free_bytes",
		"tcmalloc.thread_cache_free_bytes",
		"MarkCacheFiles",
		"UncompressedCacheBytes",
	} {
		assert.True(t, acc.HasUIntField("clickhouse_asynchronous_metrics", metric))
	}
}
