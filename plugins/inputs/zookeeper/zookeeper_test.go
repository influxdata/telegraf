package zookeeper

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
)

func TestInit(t *testing.T) {
	plugin := &Zookeeper{}

	require.NoError(t, plugin.Init())

	require.Len(t, plugin.Servers, 1)
	require.Equal(t, ":2181", plugin.Servers[0])

	require.Equal(t, config.Duration(5*time.Second), plugin.Timeout)

	require.Nil(t, plugin.tlsConfig)
}

func TestParseMntrSkipsQuantileLines(t *testing.T) {
	mntr := "zk_version	3.6.3\n" +
		"zk_avg_latency	0\n" +
		"zk_om_commit_process_time_ms{quantile=\"0.5\"}	NaN\n" +
		"zk_om_commit_process_time_ms{quantile=\"0.9\"}	NaN\n" +
		"zk_proposal_latency{quantile=\"0.99\"}	66.0\n" +
		"zk_packets_received	42\n" +
		"zk_server_state	standalone\n"

	plugin := &Zookeeper{Log: testutil.Logger{}}

	fields, state := plugin.parseMntr(strings.NewReader(mntr))

	require.Equal(t, "standalone", state)
	require.Equal(t, int64(42), fields["packets_received"])
	require.Equal(t, int64(0), fields["avg_latency"])
	require.Equal(t, "3.6.3", fields["version"])
	require.NotContains(t, fields, "om_commit_process_time_ms")
	require.NotContains(t, fields, "proposal_latency")
}

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
			wait.ForListeningPort(servicePort),
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
			require.NoError(t, tt.zookeeper.Init())

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
