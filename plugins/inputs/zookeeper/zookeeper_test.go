package zookeeper

import (
	"bufio"
	"fmt"
	"net"
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

func TestGatherIncompleteResponse(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	// Answer the request with the first lines of a response and stall
	// afterwards to make the plugin run into its timeout
	done := make(chan struct{})
	defer close(done)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		if _, err := bufio.NewReader(conn).ReadString('\n'); err != nil {
			return
		}
		if _, err := conn.Write([]byte("zk_version\t3.9.3\nzk_avg_latency\t0\n")); err != nil {
			return
		}
		<-done
	}()

	plugin := &Zookeeper{
		Servers: []string{listener.Addr().String()},
		Timeout: config.Duration(time.Second),
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.ErrorContains(t, acc.GatherError(plugin.Gather), "reading mntr response failed")
	require.Empty(t, acc.GetTelegrafMetrics())
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
