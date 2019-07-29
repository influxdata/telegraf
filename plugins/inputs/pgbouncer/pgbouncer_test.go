package pgbouncer

import (
	"fmt"
	"github.com/influxdata/telegraf/plugins/inputs/postgresql"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPgBouncerGeneratesMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

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

	intMetrics := []string{
		"total_requests",
		"total_received",
		"total_sent",
		"total_query_time",
		"avg_req",
		"avg_recv",
		"avg_sent",
		"avg_query",
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

	for _, metric := range intMetrics {
		assert.True(t, acc.HasInt64Field("pgbouncer", metric))
		metricsCounted++
	}

	for _, metric := range int32Metrics {
		assert.True(t, acc.HasInt32Field("pgbouncer", metric))
		metricsCounted++
	}

	assert.True(t, metricsCounted > 0)
	assert.Equal(t, len(intMetrics)+len(int32Metrics), metricsCounted)
}
