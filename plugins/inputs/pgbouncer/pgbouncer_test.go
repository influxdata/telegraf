package pgbouncer

import (
	"fmt"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPgbouncerGeneratesMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p := &Pgbouncer{
		Address: fmt.Sprintf("host=%s port=6432 user=postgres dbname=pgbouncer sslmode=disable",
			testutil.GetLocalHost()),
		Databases: []string{"pgbouncer"},
	}

	var acc testutil.Accumulator
	err := p.Gather(&acc)
	require.NoError(t, err)

	availableColumns := make(map[string]bool)
	for _, col := range p.AllColumns {
		availableColumns[col] = true
	}
	poolMetrics := []string{
		"cl_active",
		"cl_waiting",
		"maxwait",
		"pool_mode",
		"sv_active",
		"sv_idle",
		"sv_login",
		"sv_tested",
		"sv_used",
	}

	statMetrics := []string{
		"avg_query",
		"avg_recv",
		"avg_req",
		"avg_sent",
		"total_query_time",
		"total_received",
		"total_requests",
		"total_sent",
	}

	metricsCounted := 0

	for _, metric := range poolMetrics {
		_, ok := availableColumns[metric]
		if ok {
			assert.True(t, acc.HasIntField("pgbouncer_pools", metric))
			metricsCounted++
		}
	}

	for _, metric := range statMetrics {
		_, ok := availableColumns[metric]
		if ok {
			assert.True(t, acc.HasIntField("pgbouncer_stats", metric))
			metricsCounted++
		}
	}

	assert.True(t, metricsCounted > 0)
	// assert.Equal(t, len(availableColumns)-len(p.IgnoredColumns()), metricsCounted)
}

func TestPgbouncerTagsMetricsWithDatabaseName(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p := &Pgbouncer{
		Address: fmt.Sprintf("host=%s port=6432 user=postgres dbname=pgbouncer sslmode=disable",
			testutil.GetLocalHost()),
		Databases: []string{"pgbouncer"},
	}

	var acc testutil.Accumulator

	err := p.Gather(&acc)
	require.NoError(t, err)

	point, ok := acc.Get("pgbouncer_pools")
	require.True(t, ok)

	assert.Equal(t, "pgbouncer", point.Tags["db"])

	point, ok = acc.Get("pgbouncer_stats")
	require.True(t, ok)

	assert.Equal(t, "pgbouncer", point.Tags["db"])
}

func TestPgbouncerTagsMetricsWithSpecifiedDatabaseName(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p := &Pgbouncer{
		Address: fmt.Sprintf("host=%s port=6432 user=postgres dbname=pgbouncer sslmode=disable",
			testutil.GetLocalHost()),
		Databases: []string{"foo"},
	}

	var acc testutil.Accumulator

	err := p.Gather(&acc)
	require.NoError(t, err)

	_, ok := acc.Get("pgbouncer_pools")
	require.False(t, ok)

	_, ok = acc.Get("pgbouncer_stats")
	require.False(t, ok)
}

func TestPgbouncerDefaultsToAllDatabases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p := &Pgbouncer{
		Address: fmt.Sprintf("host=%s port=6432 user=postgres dbname=pgbouncer sslmode=disable",
			testutil.GetLocalHost()),
	}

	var acc testutil.Accumulator

	err := p.Gather(&acc)
	require.NoError(t, err)

	var found bool

	for _, pnt := range acc.Metrics {
		if pnt.Measurement == "pgbouncer_pools" {
			if pnt.Tags["db"] == "pgbouncer" {
				found = true
				break
			}
		}

		if pnt.Measurement == "pgbouncer_stats" {
			if pnt.Tags["db"] == "pgbouncer" {
				found = true
				break
			}
		}
	}

	assert.True(t, found)
}

func TestPgbouncerIgnoresUnwantedColumns(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p := &Pgbouncer{
		Address: fmt.Sprintf("host=%s port=6432 user=postgres dbname=pgbouncer sslmode=disable",
			testutil.GetLocalHost()),
	}

	var acc testutil.Accumulator

	err := p.Gather(&acc)
	require.NoError(t, err)

	for col := range p.IgnoredColumns() {
		assert.False(t, acc.HasMeasurement(col))
	}
}
