package pgbouncer

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/plugins/inputs/postgresql"
	"github.com/influxdata/telegraf/testutil"
)

func TestPgBouncerGeneratesMetricsIntegration(t *testing.T) {
	t.Skip("Skipping test, connection refused")

	p := &PgBouncer{
		Service: postgresql.Service{
			Address: fmt.Sprintf(
				"host=%s user=pgbouncer password=pgbouncer dbname=pgbouncer port=6432 sslmode=disable",
				testutil.GetLocalHost(),
			),
			IsPgBouncer: true,
		},
	}

	var acc testutil.Accumulator
	require.NoError(t, p.Start(&acc))
	require.NoError(t, p.Gather(&acc))

	// Return value of pgBouncer
	// [pgbouncer map[db:pgbouncer server:host=localhost user=pgbouncer dbname=pgbouncer port=6432 ] map[avg_query_count:0 avg_query_time:0 avg_wait_time:0 avg_xact_count:0 avg_xact_time:0 total_query_count:3 total_query_time:0 total_received:0 total_sent:0 total_wait_time:0 total_xact_count:3 total_xact_time:0] 1620163750039747891 pgbouncer_pools map[db:pgbouncer pool_mode:statement server:host=localhost user=pgbouncer dbname=pgbouncer port=6432  user:pgbouncer] map[cl_active:1 cl_waiting:0 maxwait:0 maxwait_us:0 sv_active:0 sv_idle:0 sv_login:0 sv_tested:0 sv_used:0] 1620163750041444466]

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

	int32Metrics := []string{}

	metricsCounted := 0

	for _, metric := range intMetricsPgBouncer {
		require.True(t, acc.HasInt64Field("pgbouncer", metric))
		metricsCounted++
	}

	for _, metric := range intMetricsPgBouncerPools {
		require.True(t, acc.HasInt64Field("pgbouncer_pools", metric))
		metricsCounted++
	}

	for _, metric := range int32Metrics {
		require.True(t, acc.HasInt32Field("pgbouncer", metric))
		metricsCounted++
	}

	require.True(t, metricsCounted > 0)
	require.Equal(t, len(intMetricsPgBouncer)+len(intMetricsPgBouncerPools)+len(int32Metrics), metricsCounted)
}
