package zookeeper

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

	z := &Zookeeper{
		Servers: []string{testutil.GetLocalHost()},
	}

	var acc testutil.Accumulator

	err := z.Gather(&acc)
	require.NoError(t, err)

	intMetrics := []string{"zookeeper_avg_latency", "zookeeper_packets_sent", "zookeeper_znode_count", "zookeeper_open_file_descriptor_count", "zookeeper_max_file_descriptor_count"}

	for _, metric := range intMetrics {
		assert.True(t, acc.HasIntValue(metric), metric)
	}
}
