package pgbouncer

import (
	"fmt"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs/postgresql"
	"github.com/influxdata/telegraf/testutil"
)

func TestPgBouncerGeneratesMetricsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	postgresServicePort := "5432"
	pgBouncerServicePort := "6432"

	backend := testutil.Container{
		Image:        "postgres:alpine",
		ExposedPorts: []string{postgresServicePort},
		Env: map[string]string{
			"POSTGRES_HOST_AUTH_METHOD": "trust",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
	}
	err := backend.Start()
	require.NoError(t, err, "failed to start container")
	defer backend.Terminate()

	container := testutil.Container{
		Image:        "z9pascal/pgbouncer-container:1.18.0-latest",
		ExposedPorts: []string{pgBouncerServicePort},
		Env: map[string]string{
			"PG_ENV_POSTGRESQL_USER": "pgbouncer",
			"PG_ENV_POSTGRESQL_PASS": "pgbouncer",
		},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort(nat.Port(pgBouncerServicePort)),
			wait.ForLog("LOG process up"),
		),
	}
	err = container.Start()
	require.NoError(t, err, "failed to start container")
	defer container.Terminate()

	addr := fmt.Sprintf(
		"host=%s user=pgbouncer password=pgbouncer dbname=pgbouncer port=%s sslmode=disable",
		container.Address,
		container.Ports[pgBouncerServicePort],
	)

	p := &PgBouncer{
		Service: postgresql.Service{
			Address:     config.NewSecret([]byte(addr)),
			IsPgBouncer: true,
		},
	}

	var acc testutil.Accumulator
	require.NoError(t, p.Start(&acc))
	require.NoError(t, p.Gather(&acc))

	// Return value of pgBouncer
	// [pgbouncer map[db:pgbouncer server:host=localhost user=pgbouncer dbname=pgbouncer port=6432 ]
	// map[avg_query_count:0 avg_query_time:0 avg_wait_time:0 avg_xact_count:0 avg_xact_time:0 total_query_count:3 total_query_time:0 total_received:0
	// total_sent:0 total_wait_time:0 total_xact_count:3 total_xact_time:0] 1620163750039747891 pgbouncer_pools map[db:pgbouncer pool_mode:statement
	// server:host=localhost user=pgbouncer dbname=pgbouncer port=6432  user:pgbouncer] map[cl_active:1 cl_waiting:0 maxwait:0 maxwait_us:0
	// sv_active:0 sv_idle:0 sv_login:0 sv_tested:0 sv_used:0] 1620163750041444466]

	intMetricsPgBouncer := []string{
		"total_received",
		"total_sent",
		"total_query_time",
		"avg_query_count",
		"avg_query_time",
		"avg_wait_time",
	}

	intMetricsPgBouncerPools := []string{
		"cl_active",
		"cl_waiting",
		"sv_active",
		"sv_idle",
		"sv_used",
		"sv_tested",
		"sv_login",
		"maxwait",
	}

	metricsCounted := 0

	for _, metric := range intMetricsPgBouncer {
		require.True(t, acc.HasInt64Field("pgbouncer", metric))
		metricsCounted++
	}

	for _, metric := range intMetricsPgBouncerPools {
		require.True(t, acc.HasInt64Field("pgbouncer_pools", metric))
		metricsCounted++
	}

	require.True(t, metricsCounted > 0)
	require.Equal(t, len(intMetricsPgBouncer)+len(intMetricsPgBouncerPools), metricsCounted)
}

func TestPgBouncerGeneratesMetricsIntegrationShowCommands(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	postgresServicePort := "5432"
	pgBouncerServicePort := "6432"

	backend := testutil.Container{
		Image:        "postgres:alpine",
		ExposedPorts: []string{postgresServicePort},
		Env: map[string]string{
			"POSTGRES_HOST_AUTH_METHOD": "trust",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
	}
	err := backend.Start()
	require.NoError(t, err, "failed to start container")
	defer backend.Terminate()

	container := testutil.Container{
		Image:        "z9pascal/pgbouncer-container:1.18.0-latest",
		ExposedPorts: []string{pgBouncerServicePort},
		Env: map[string]string{
			"PG_ENV_POSTGRESQL_USER": "pgbouncer",
			"PG_ENV_POSTGRESQL_PASS": "pgbouncer",
		},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort(nat.Port(pgBouncerServicePort)),
			wait.ForLog("LOG process up"),
		),
	}
	err = container.Start()
	require.NoError(t, err, "failed to start container")
	defer container.Terminate()

	addr := fmt.Sprintf(
		"host=%s user=pgbouncer password=pgbouncer dbname=pgbouncer port=%s sslmode=disable",
		container.Address,
		container.Ports[pgBouncerServicePort],
	)

	p := &PgBouncer{
		Service: postgresql.Service{
			Address:     config.NewSecret([]byte(addr)),
			IsPgBouncer: true,
		},
		ShowCommands: []string{"pools", "lists", "databases"},
	}

	var acc testutil.Accumulator
	require.NoError(t, p.Start(&acc))
	require.NoError(t, p.Gather(&acc))

	// Return value of pgBouncer
	// [pgbouncer_pools map[db:pgbouncer pool_mode:statement server:host=localhost user=pgbouncer dbname=pgbouncer port=6432  user:pgbouncer]
	// map[cl_active:1 cl_waiting:0 maxwait:0 maxwait_us:0 sv_active:0 sv_idle:0 sv_login:0 sv_tested:0 sv_used:0] 1620163750041444466
	// [pgbouncer_lists map[server:host=localhost]
	// map[databases:1 dns_names:0 dns_queries:0 dns_zones:0 free_clients:49 free_servers:0 login_clients:0 pools:1 used_clients:1 used_servers:0 users:2]
	// [pgbouncer_databases map[db:pgbouncer pg_dbname:pgbouncer server:host=localhost]
	// map[current_connections:0 disabled:0 max_connections:0 min_pool_size:0 paused:0 pool_size:2 reserve_pool:0]

	intMetricsPgBouncerPools := []string{
		"cl_active",
		"cl_waiting",
		"sv_active",
		"sv_idle",
		"sv_used",
		"sv_tested",
		"sv_login",
		"maxwait",
	}

	intMetricsPgBouncerLists := []string{
		"databases",
		"users",
		"pools",
		"free_clients",
		"used_clients",
		"login_clients",
		"free_servers",
		"used_servers",
		"dns_names",
		"dns_zones",
		"dns_queries",
	}

	intMetricsPgBouncerDatabases := []string{
		"pool_size",
		"min_pool_size",
		"reserve_pool",
		"max_connections",
		"current_connections",
		"paused",
		"disabled",
	}

	metricsCounted := 0

	for _, metric := range intMetricsPgBouncerPools {
		require.True(t, acc.HasInt64Field("pgbouncer_pools", metric))
		metricsCounted++
	}

	for _, metric := range intMetricsPgBouncerLists {
		require.True(t, acc.HasInt64Field("pgbouncer_lists", metric))
		metricsCounted++
	}

	for _, metric := range intMetricsPgBouncerDatabases {
		fmt.Println(acc)
		require.True(t, acc.HasInt64Field("pgbouncer_databases", metric))
		metricsCounted++
	}

	require.True(t, metricsCounted > 0)
	require.Equal(t, len(intMetricsPgBouncerPools)+len(intMetricsPgBouncerLists)+len(intMetricsPgBouncerDatabases), metricsCounted)
}
