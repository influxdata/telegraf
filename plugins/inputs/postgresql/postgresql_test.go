package postgresql

import (
	"fmt"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/postgresql"
	"github.com/influxdata/telegraf/testutil"
)

const servicePort = "5432"

func launchTestContainer(t *testing.T) *testutil.Container {
	container := testutil.Container{
		Image:        "postgres:alpine",
		ExposedPorts: []string{servicePort},
		Env: map[string]string{
			"POSTGRES_HOST_AUTH_METHOD": "trust",
		},
		WaitingFor: wait.ForAll(
			// the database comes up twice, once right away, then again a second
			// time after the docker entrypoint starts configuration
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
			wait.ForListeningPort(nat.Port(servicePort)),
		),
	}

	err := container.Start()
	require.NoError(t, err, "failed to start container")

	return &container
}

func TestPostgresqlGeneratesMetricsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := launchTestContainer(t)
	defer container.Terminate()

	addr := fmt.Sprintf(
		"host=%s port=%s user=postgres sslmode=disable",
		container.Address,
		container.Ports[servicePort],
	)

	p := &Postgresql{
		Config: postgresql.Config{
			Address:     config.NewSecret([]byte(addr)),
			IsPgBouncer: false,
		},
		Databases: []string{"postgres"},
	}
	require.NoError(t, p.Init())

	var acc testutil.Accumulator
	require.NoError(t, p.Start(&acc))
	defer p.Stop()
	require.NoError(t, p.Gather(&acc))

	intMetrics := []string{
		"xact_commit",
		"xact_rollback",
		"blks_read",
		"blks_hit",
		"tup_returned",
		"tup_fetched",
		"tup_inserted",
		"tup_updated",
		"tup_deleted",
		"conflicts",
		"temp_files",
		"temp_bytes",
		"deadlocks",
		"buffers_alloc",
		"buffers_clean",
		"maxwritten_clean",
		"datid",
		"numbackends",
		"sessions",
		"sessions_killed",
		"sessions_fatal",
		"sessions_abandoned",
	}

	var int32Metrics []string

	floatMetrics := []string{
		"blk_read_time",
		"blk_write_time",
		"active_time",
		"idle_in_transaction_time",
		"session_time",
	}

	stringMetrics := []string{
		"datname",
	}

	metricsCounted := 0

	for _, metric := range intMetrics {
		require.True(t, acc.HasInt64Field("postgresql", metric), "%q not found in int metrics", metric)
		metricsCounted++
	}

	//nolint:gosec // G602: False positive — this is a safe range iteration over a slice (it may be empty)
	for _, metric := range int32Metrics {
		require.True(t, acc.HasInt32Field("postgresql", metric), "%q not found in int32 metrics", metric)
		metricsCounted++
	}

	for _, metric := range floatMetrics {
		require.True(t, acc.HasFloatField("postgresql", metric), "%q not found in float metrics", metric)
		metricsCounted++
	}

	for _, metric := range stringMetrics {
		require.True(t, acc.HasStringField("postgresql", metric), "%q not found in string metrics", metric)
		metricsCounted++
	}

	require.Positive(t, metricsCounted)
	require.Equal(t, len(floatMetrics)+len(intMetrics)+len(int32Metrics)+len(stringMetrics), metricsCounted)
}

func TestPostgresqlTagsMetricsWithDatabaseNameIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := launchTestContainer(t)
	defer container.Terminate()

	addr := fmt.Sprintf(
		"host=%s port=%s user=postgres sslmode=disable",
		container.Address,
		container.Ports[servicePort],
	)

	p := &Postgresql{
		Config: postgresql.Config{
			Address: config.NewSecret([]byte(addr)),
		},
		Databases: []string{"postgres"},
	}
	require.NoError(t, p.Init())

	var acc testutil.Accumulator
	require.NoError(t, p.Start(&acc))
	defer p.Stop()
	require.NoError(t, p.Gather(&acc))

	point, ok := acc.Get("postgresql")
	require.True(t, ok)

	require.Equal(t, "postgres", point.Tags["db"])
}

func TestPostgresqlDefaultsToAllDatabasesIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := launchTestContainer(t)
	defer container.Terminate()

	addr := fmt.Sprintf(
		"host=%s port=%s user=postgres sslmode=disable",
		container.Address,
		container.Ports[servicePort],
	)

	p := &Postgresql{
		Config: postgresql.Config{
			Address: config.NewSecret([]byte(addr)),
		},
	}
	require.NoError(t, p.Init())

	var acc testutil.Accumulator
	require.NoError(t, p.Start(&acc))
	defer p.Stop()
	require.NoError(t, p.Gather(&acc))

	var found bool

	for _, pnt := range acc.Metrics {
		if pnt.Measurement == "postgresql" {
			if pnt.Tags["db"] == "postgres" {
				found = true
				break
			}
		}
	}

	require.True(t, found)
}

func TestPostgresqlIgnoresUnwantedColumnsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := launchTestContainer(t)
	defer container.Terminate()

	addr := fmt.Sprintf(
		"host=%s port=%s user=postgres sslmode=disable",
		container.Address,
		container.Ports[servicePort],
	)

	p := &Postgresql{
		Config: postgresql.Config{
			Address: config.NewSecret([]byte(addr)),
		},
	}
	require.NoError(t, p.Init())

	var acc testutil.Accumulator
	require.NoError(t, p.Start(&acc))
	defer p.Stop()
	require.NoError(t, p.Gather(&acc))

	for col := range ignoredColumns {
		require.False(t, acc.HasMeasurement(col))
	}
}

func TestPostgresqlDatabaseWhitelistTestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := launchTestContainer(t)
	defer container.Terminate()

	addr := fmt.Sprintf(
		"host=%s port=%s user=postgres sslmode=disable",
		container.Address,
		container.Ports[servicePort],
	)

	p := &Postgresql{
		Config: postgresql.Config{
			Address: config.NewSecret([]byte(addr)),
		},
		Databases: []string{"template0"},
	}
	require.NoError(t, p.Init())

	var acc testutil.Accumulator
	require.NoError(t, p.Start(&acc))
	defer p.Stop()
	require.NoError(t, p.Gather(&acc))

	var foundTemplate0 = false
	var foundTemplate1 = false

	for _, pnt := range acc.Metrics {
		if pnt.Measurement == "postgresql" {
			if pnt.Tags["db"] == "template0" {
				foundTemplate0 = true
			}
		}
		if pnt.Measurement == "postgresql" {
			if pnt.Tags["db"] == "template1" {
				foundTemplate1 = true
			}
		}
	}

	require.True(t, foundTemplate0)
	require.False(t, foundTemplate1)
}

func TestPostgresqlDatabaseBlacklistTestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := launchTestContainer(t)
	defer container.Terminate()

	addr := fmt.Sprintf(
		"host=%s port=%s user=postgres sslmode=disable",
		container.Address,
		container.Ports[servicePort],
	)

	p := &Postgresql{
		Config: postgresql.Config{
			Address: config.NewSecret([]byte(addr)),
		},
		IgnoredDatabases: []string{"template0"},
	}
	require.NoError(t, p.Init())

	var acc testutil.Accumulator
	require.NoError(t, p.Start(&acc))
	defer p.Stop()
	require.NoError(t, p.Gather(&acc))

	var foundTemplate0 = false
	var foundTemplate1 = false

	for _, pnt := range acc.Metrics {
		if pnt.Measurement == "postgresql" {
			if pnt.Tags["db"] == "template0" {
				foundTemplate0 = true
			}
		}
		if pnt.Measurement == "postgresql" {
			if pnt.Tags["db"] == "template1" {
				foundTemplate1 = true
			}
		}
	}

	require.False(t, foundTemplate0)
	require.True(t, foundTemplate1)
}

func TestInitialConnectivityIssueIntegration(t *testing.T) {
	// Test case for https://github.com/influxdata/telegraf/issues/8586
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Startup the container
	container := testutil.Container{
		Image:        "postgres:alpine",
		ExposedPorts: []string{servicePort},
		Env: map[string]string{
			"POSTGRES_HOST_AUTH_METHOD": "trust",
		},
		WaitingFor: wait.ForAll(
			// the database comes up twice, once right away, then again a second
			// time after the docker entrypoint starts configuration
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
			wait.ForListeningPort(nat.Port(servicePort)),
		),
	}
	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()

	// Pause the container to simulate connectivity issues
	require.NoError(t, container.Pause())

	// Setup and start the plugin. This should work as the SQL framework will
	// not connect immediately but on the first query/access to the server
	addr := fmt.Sprintf("host=%s port=%s user=postgres sslmode=disable connect_timeout=1", container.Address, container.Ports[servicePort])
	plugin := &Postgresql{
		Config: postgresql.Config{
			Address: config.NewSecret([]byte(addr)),
		},
		IgnoredDatabases: []string{"template0"},
	}
	require.NoError(t, plugin.Init())

	// Startup the plugin
	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	// This should fail because we cannot connect
	require.ErrorContains(t, acc.GatherError(plugin.Gather), "failed to connect")

	// Unpause the container, now gather should succeed
	require.NoError(t, container.Resume())
	require.NoError(t, acc.GatherError(plugin.Gather))
	require.NotEmpty(t, acc.GetTelegrafMetrics())
}
