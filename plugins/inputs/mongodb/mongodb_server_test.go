package mongodb

import (
	"fmt"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf/testutil"
)

var ServicePort = "27017"

func createTestServer(t *testing.T) *testutil.Container {
	container := testutil.Container{
		Image:        "mongo",
		ExposedPorts: []string{ServicePort},
		WaitingFor: wait.ForAll(
			wait.NewHTTPStrategy("/").WithPort(nat.Port(ServicePort)),
			wait.ForLog("Waiting for connections"),
		),
	}
	err := container.Start()
	require.NoError(t, err, "failed to start container")

	return &container
}

func TestGetDefaultTagsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := createTestServer(t)
	defer func() {
		require.NoError(t, container.Terminate(), "terminating container failed")
	}()

	m := &MongoDB{
		Log: testutil.Logger{},
		Servers: []string{
			fmt.Sprintf("mongodb://%s:%s", container.Address, container.Ports[ServicePort]),
		},
	}
	err := m.Init()
	require.NoError(t, err)

	server := m.clients[0]

	var tagTests = []struct {
		in  string
		out string
	}{
		{"hostname", server.hostname},
	}
	defaultTags := server.getDefaultTags()
	for _, tt := range tagTests {
		if defaultTags[tt.in] != tt.out {
			t.Errorf("expected %q, got %q", tt.out, defaultTags[tt.in])
		}
	}
}

func TestAddDefaultStatsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := createTestServer(t)
	defer func() {
		require.NoError(t, container.Terminate(), "terminating container failed")
	}()

	m := &MongoDB{
		Log: testutil.Logger{},
		Servers: []string{
			fmt.Sprintf("mongodb://%s:%s", container.Address, container.Ports[ServicePort]),
		},
	}
	err := m.Init()
	require.NoError(t, err)

	server := m.clients[0]

	var acc testutil.Accumulator
	err = server.gatherData(&acc, false, true, true, true, []string{"local"})
	require.NoError(t, err)

	// need to call this twice so it can perform the diff
	err = server.gatherData(&acc, false, true, true, true, []string{"local"})
	require.NoError(t, err)

	for key := range defaultStats {
		require.True(t, acc.HasInt64Field("mongodb", key))
	}
}

func TestPoolStatsVersionCompatibility(t *testing.T) {
	tests := []struct {
		name            string
		version         string
		expectedCommand string
		err             bool
	}{
		{
			name:            "mongodb v3",
			version:         "3.0.0",
			expectedCommand: "shardConnPoolStats",
		},
		{
			name:            "mongodb v4",
			version:         "4.0.0",
			expectedCommand: "shardConnPoolStats",
		},
		{
			name:            "mongodb v5",
			version:         "5.0.0",
			expectedCommand: "connPoolStats",
		},
		{
			name:            "mongodb v6",
			version:         "6.0.0",
			expectedCommand: "connPoolStats",
		},
		{
			name:    "invalid version",
			version: "v4",
			err:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			command, err := poolStatsCommand(test.version)
			require.Equal(t, test.expectedCommand, command)
			if test.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
