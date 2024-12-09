package zookeeper

import (
	"fmt"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf/testutil"
)

func TestZookeeperGeneratesMetricsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	servicePort := "2181"
	container := testutil.Container{
		Image:        "zookeeper",
		ExposedPorts: []string{servicePort},
		Env: map[string]string{
			"ZOO_4LW_COMMANDS_WHITELIST": "mntr",
		},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort(nat.Port(servicePort)),
			wait.ForLog("ZooKeeper audit is disabled."),
		),
	}
	err := container.Start()
	require.NoError(t, err, "failed to start container")
	defer container.Terminate()

	var testset = []struct {
		name      string
		zookeeper Zookeeper
	}{
		{
			name: "floats as strings",
			zookeeper: Zookeeper{
				Servers: []string{
					fmt.Sprintf("%s:%s", container.Address, container.Ports[servicePort]),
				},
			},
		},
		{
			name: "floats as floats",
			zookeeper: Zookeeper{
				Servers: []string{
					fmt.Sprintf("%s:%s", container.Address, container.Ports[servicePort]),
				},
				ParseFloats: "float",
			},
		},
	}
	for _, tt := range testset {
		t.Run(tt.name, func(t *testing.T) {
			var acc testutil.Accumulator
			require.NoError(t, acc.GatherError(tt.zookeeper.Gather))

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

			if tt.zookeeper.ParseFloats == "float" {
				require.True(t, acc.HasFloatField("zookeeper", "avg_latency"), "avg_latency not a float")
			} else {
				require.True(t, acc.HasStringField("zookeeper", "avg_latency"), "avg_latency not a string")
			}
		})
	}
}
