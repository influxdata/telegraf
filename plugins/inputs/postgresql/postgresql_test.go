package postgresql

import (
	"fmt"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresqlGeneratesMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p := &Postgresql{
		Service: Service{
			Address: fmt.Sprintf(
				"host=%s user=postgres sslmode=disable",
				testutil.GetLocalHost(),
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
		assert.True(t, acc.HasInt64Field("postgresql", metric))
		metricsCounted++
	}

	for _, metric := range int32Metrics {
		assert.True(t, acc.HasInt32Field("postgresql", metric))
		metricsCounted++
	}

	for _, metric := range floatMetrics {
		assert.True(t, acc.HasFloatField("postgresql", metric))
		metricsCounted++
	}

	for _, metric := range stringMetrics {
		assert.True(t, acc.HasStringField("postgresql", metric))
		metricsCounted++
	}

	assert.True(t, metricsCounted > 0)
	assert.Equal(t, len(floatMetrics)+len(intMetrics)+len(int32Metrics)+len(stringMetrics), metricsCounted)
}

func TestPostgresqlTagsMetricsWithDatabaseName(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p := &Postgresql{
		Service: Service{
			Address: fmt.Sprintf(
				"host=%s user=postgres sslmode=disable",
				testutil.GetLocalHost(),
			),
		},
		Databases: []string{"postgres"},
	}

	var acc testutil.Accumulator

	require.NoError(t, p.Start(&acc))
	require.NoError(t, p.Gather(&acc))

	point, ok := acc.Get("postgresql")
	require.True(t, ok)

	assert.Equal(t, "postgres", point.Tags["db"])
}

func TestPostgresqlDefaultsToAllDatabases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p := &Postgresql{
		Service: Service{
			Address: fmt.Sprintf(
				"host=%s user=postgres sslmode=disable",
				testutil.GetLocalHost(),
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

	assert.True(t, found)
}

func TestPostgresqlIgnoresUnwantedColumns(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p := &Postgresql{
		Service: Service{
			Address: fmt.Sprintf(
				"host=%s user=postgres sslmode=disable",
				testutil.GetLocalHost(),
			),
		},
	}

	var acc testutil.Accumulator
	require.NoError(t, p.Start(&acc))
	require.NoError(t, p.Gather(&acc))

	for col := range p.IgnoredColumns() {
		assert.False(t, acc.HasMeasurement(col))
	}
}

func TestPostgresqlDatabaseWhitelistTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p := &Postgresql{
		Service: Service{
			Address: fmt.Sprintf(
				"host=%s user=postgres sslmode=disable",
				testutil.GetLocalHost(),
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

	assert.True(t, foundTemplate0)
	assert.False(t, foundTemplate1)
}

func TestPostgresqlDatabaseBlacklistTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p := &Postgresql{
		Service: Service{
			Address: fmt.Sprintf(
				"host=%s user=postgres sslmode=disable",
				testutil.GetLocalHost(),
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

	assert.False(t, foundTemplate0)
	assert.True(t, foundTemplate1)
}
