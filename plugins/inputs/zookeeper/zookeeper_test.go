package zookeeper

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf/testutil"
)

func TestZookeeperGeneratesMetricsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := testutil.Container{
		Image:        "zookeeper",
		ExposedPorts: []string{"2181"},
		Env: map[string]string{
			"ZOO_4LW_COMMANDS_WHITELIST": "mntr",
		},
		WaitingFor: wait.ForListeningPort("2181"),
	}
	err := container.Start()
	require.NoError(t, err, "failed to start container")
	defer func() {
		require.NoError(t, container.Terminate(), "terminating container failed")
	}()

	z := &Zookeeper{
		Servers: []string{
			fmt.Sprintf("%s:%s", container.Address, container.Port),
		},
	}

	var acc testutil.Accumulator

	require.NoError(t, acc.GatherError(z.Gather))

	intMetrics := []string{
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
		require.True(t, acc.HasInt64Field("zookeeper", metric), metric)
	}

	// Currently we output floats as strings (see #8863), but the desired behavior is to have floats
	require.True(t, acc.HasStringField("zookeeper", "avg_latency"), "avg_latency")
	// require.True(t, acc.HasFloat64Field("zookeeper", "avg_latency"), "avg_latency")
}
