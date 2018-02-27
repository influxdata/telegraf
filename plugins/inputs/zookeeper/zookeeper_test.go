package zookeeper

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestZookeeperGeneratesMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	z := &Zookeeper{
		Servers: []string{testutil.GetLocalHost() + ":2181"},
	}

	var acc testutil.Accumulator

	require.NoError(t, acc.GatherError(z.Gather))

	intMetrics := []string{
		"avg_latency",
		"max_latency",
		"min_latency",
		"packets_received",
		"packets_sent",
		"outstanding_requests",
		"znode_count",
		"watch_count",
		"ephemerals_count",
		"approximate_data_size",
		"open_file_descriptor_count",
		"max_file_descriptor_count",
	}

	for _, metric := range intMetrics {
		assert.True(t, acc.HasInt64Field("zookeeper", metric), metric)
	}
}
