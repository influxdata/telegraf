package postgresql

import (
	"testing"

	"github.com/influxdb/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresqlGeneratesMetrics(t *testing.T) {
	p := &Postgresql{
		Servers: []*Server{
			{
				Address:   "host=localhost user=postgres sslmode=disable",
				Databases: []string{"postgres"},
			},
		},
	}

	var acc testutil.Accumulator

	err := p.Gather(&acc)
	require.NoError(t, err)

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
	}

	floatMetrics := []string{
		"blk_read_time",
		"blk_write_time",
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
				Address:   "host=localhost user=postgres sslmode=disable",
				Databases: []string{"postgres"},
			},
		},
	}

	var acc testutil.Accumulator

	err := p.Gather(&acc)
	require.NoError(t, err)

	point, ok := acc.Get("xact_commit")
	require.True(t, ok)

	assert.Equal(t, "postgres", point.Tags["db"])
}

func TestPostgresqlDefaultsToAllDatabases(t *testing.T) {
	p := &Postgresql{
		Servers: []*Server{
			{
				Address: "host=localhost user=postgres sslmode=disable",
			},
		},
	}

	var acc testutil.Accumulator

	err := p.Gather(&acc)
	require.NoError(t, err)

	var found bool

	for _, pnt := range acc.Points {
		if pnt.Measurement == "xact_commit" {
			if pnt.Tags["db"] == "postgres" {
				found = true
				break
			}
		}
	}

	assert.True(t, found)
}
