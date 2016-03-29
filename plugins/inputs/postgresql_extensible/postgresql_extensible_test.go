package postgresql_extensible

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
		Address: fmt.Sprintf("host=%s user=postgres sslmode=disable",
			testutil.GetLocalHost()),
		Databases: []string{"postgres"},
		Query: query{
			{Sqlquery: "select * from pg_stat_database",
				Version:    901,
				Withdbname: false,
				Tagvalue:   ""},
		},
	}
	var acc testutil.Accumulator
	err := p.Gather(&acc)
	require.NoError(t, err)

	availableColumns := make(map[string]bool)
	for _, col := range p.AllColumns {
		availableColumns[col] = true
	}
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
		"numbackends",
	}

	floatMetrics := []string{
		"blk_read_time",
		"blk_write_time",
	}

	metricsCounted := 0

	for _, metric := range intMetrics {
		_, ok := availableColumns[metric]
		if ok {
			assert.True(t, acc.HasIntField("postgresql", metric))
			metricsCounted++
		}
	}

	for _, metric := range floatMetrics {
		_, ok := availableColumns[metric]
		if ok {
			assert.True(t, acc.HasFloatField("postgresql", metric))
			metricsCounted++
		}
	}

	assert.True(t, metricsCounted > 0)
	assert.Equal(t, len(availableColumns)-len(p.IgnoredColumns()), metricsCounted)
}

func TestPostgresqlIgnoresUnwantedColumns(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p := &Postgresql{
		Address: fmt.Sprintf("host=%s user=postgres sslmode=disable",
			testutil.GetLocalHost()),
	}

	var acc testutil.Accumulator

	err := p.Gather(&acc)
	require.NoError(t, err)

	for col := range p.IgnoredColumns() {
		assert.False(t, acc.HasMeasurement(col))
	}
}
