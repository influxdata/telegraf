package postgresql

import (
	"fmt"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

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
	defer func() {
		require.NoError(t, container.Terminate(), "terminating container failed")
	}()

	p := &Postgresql{
		Service: Service{
			Address: fmt.Sprintf(
				"host=%s port=%s user=postgres sslmode=disable",
				container.Address,
				container.Ports[servicePort],
			),
			IsPgBouncer: false,
		},
		Databases: []string{"postgres"},
	}

	var acc testutil.Accumulator
	require.NoError(t, p.Start(&acc))
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
		"buffers_backend",
		"buffers_backend_fsync",
		"buffers_checkpoint",
		"buffers_clean",
		"checkpoints_req",
		"checkpoints_timed",
		"maxwritten_clean",
		"datid",
		"numbackends",
	}

	int32Metrics := []string{}

	floatMetrics := []string{
		"blk_read_time",
		"blk_write_time",
		"checkpoint_write_time",
		"checkpoint_sync_time",
	}

	stringMetrics := []string{
		"datname",
	}

	metricsCounted := 0

	for _, metric := range intMetrics {
		require.True(t, acc.HasInt64Field("postgresql", metric))
		metricsCounted++
	}

	for _, metric := range int32Metrics {
		require.True(t, acc.HasInt32Field("postgresql", metric))
		metricsCounted++
	}

	for _, metric := range floatMetrics {
		require.True(t, acc.HasFloatField("postgresql", metric))
		metricsCounted++
	}

	for _, metric := range stringMetrics {
		require.True(t, acc.HasStringField("postgresql", metric))
		metricsCounted++
	}

	require.True(t, metricsCounted > 0)
	require.Equal(t, len(floatMetrics)+len(intMetrics)+len(int32Metrics)+len(stringMetrics), metricsCounted)
}

func TestPostgresqlTagsMetricsWithDatabaseNameIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := launchTestContainer(t)
	defer func() {
		require.NoError(t, container.Terminate(), "terminating container failed")
	}()

	p := &Postgresql{
		Service: Service{
			Address: fmt.Sprintf(
				"host=%s port=%s user=postgres sslmode=disable",
				container.Address,
				container.Ports[servicePort],
			),
		},
		Databases: []string{"postgres"},
	}

	var acc testutil.Accumulator

	require.NoError(t, p.Start(&acc))
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
	defer func() {
		require.NoError(t, container.Terminate(), "terminating container failed")
	}()

	p := &Postgresql{
		Service: Service{
			Address: fmt.Sprintf(
				"host=%s port=%s user=postgres sslmode=disable",
				container.Address,
				container.Ports[servicePort],
			),
		},
	}

	var acc testutil.Accumulator

	require.NoError(t, p.Start(&acc))
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
	defer func() {
		require.NoError(t, container.Terminate(), "terminating container failed")
	}()

	p := &Postgresql{
		Service: Service{
			Address: fmt.Sprintf(
				"host=%s port=%s user=postgres sslmode=disable",
				container.Address,
				container.Ports[servicePort],
			),
		},
	}

	var acc testutil.Accumulator
	require.NoError(t, p.Start(&acc))
	require.NoError(t, p.Gather(&acc))

	for col := range p.IgnoredColumns() {
		require.False(t, acc.HasMeasurement(col))
	}
}

func TestPostgresqlDatabaseWhitelistTestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := launchTestContainer(t)
	defer func() {
		require.NoError(t, container.Terminate(), "terminating container failed")
	}()

	p := &Postgresql{
		Service: Service{
			Address: fmt.Sprintf(
				"host=%s port=%s user=postgres sslmode=disable",
				container.Address,
				container.Ports[servicePort],
			),
		},
		Databases: []string{"template0"},
	}

	var acc testutil.Accumulator

	require.NoError(t, p.Start(&acc))
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
	defer func() {
		require.NoError(t, container.Terminate(), "terminating container failed")
	}()

	p := &Postgresql{
		Service: Service{
			Address: fmt.Sprintf(
				"host=%s port=%s user=postgres sslmode=disable",
				container.Address,
				container.Ports[servicePort],
			),
		},
		IgnoredDatabases: []string{"template0"},
	}

	var acc testutil.Accumulator
	require.NoError(t, p.Start(&acc))
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
