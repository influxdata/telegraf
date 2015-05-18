package postgresql

import (
	"testing"

	"github.com/influxdb/tivan/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresqlGeneratesMetrics(t *testing.T) {
	p := &Postgresql{
		Servers: []*Server{
			{
				Address:   "sslmode=disable",
				Databases: []string{"postgres"},
			},
		},
	}

	var acc testutil.Accumulator

	err := p.Gather(&acc)
	require.NoError(t, err)

	intMetrics := []string{
		"postgresql_xact_commit",
		"postgresql_xact_rollback",
		"postgresql_blks_read",
		"postgresql_blks_hit",
		"postgresql_tup_returned",
		"postgresql_tup_fetched",
		"postgresql_tup_inserted",
		"postgresql_tup_updated",
		"postgresql_tup_deleted",
		"postgresql_conflicts",
		"postgresql_temp_files",
		"postgresql_temp_bytes",
		"postgresql_deadlocks",
	}

	floatMetrics := []string{
		"postgresql_blk_read_time",
		"postgresql_blk_write_time",
	}

	for _, metric := range intMetrics {
		assert.True(t, acc.HasIntValue(metric))
	}

	for _, metric := range floatMetrics {
		assert.True(t, acc.HasFloatValue(metric))
	}
}

func TestPostgresqlTagsMetricsWithDatabaseName(t *testing.T) {
	p := &Postgresql{
		Servers: []*Server{
			{
				Address:   "sslmode=disable",
				Databases: []string{"postgres"},
			},
		},
	}

	var acc testutil.Accumulator

	err := p.Gather(&acc)
	require.NoError(t, err)

	point, ok := acc.Get("postgresql_xact_commit")
	require.True(t, ok)

	assert.Equal(t, "postgres", point.Tags["db"])
}

func TestPostgresqlDefaultsToAllDatabases(t *testing.T) {
	p := &Postgresql{
		Servers: []*Server{
			{
				Address: "sslmode=disable",
			},
		},
	}

	var acc testutil.Accumulator

	err := p.Gather(&acc)
	require.NoError(t, err)

	var found bool

	for _, pnt := range acc.Points {
		if pnt.Name == "postgresql_xact_commit" {
			if pnt.Tags["db"] == "postgres" {
				found = true
				break
			}
		}
	}

	assert.True(t, found)
}
