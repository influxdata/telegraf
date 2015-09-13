package postgresql

import (
	"fmt"
	"testing"

	"github.com/influxdb/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresqlGeneratesMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p := &Postgresql{
		Servers: []*Server{
			{
				Address: fmt.Sprintf("host=%s user=postgres sslmode=disable",
					testutil.GetLocalHost()),
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
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p := &Postgresql{
		Servers: []*Server{
			{
				Address: fmt.Sprintf("host=%s user=postgres sslmode=disable",
					testutil.GetLocalHost()),
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
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p := &Postgresql{
		Servers: []*Server{
			{
				Address: fmt.Sprintf("host=%s user=postgres sslmode=disable",
					testutil.GetLocalHost()),
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

func TestPostgresqlIgnoresUnwantedColumns(t *testing.T) {
	// if testing.Short() {
	// 	t.Skip("Skipping integration test in short mode")
	// }

	p := &Postgresql{
		Servers: []*Server{
			{
				Address: fmt.Sprintf("host=%s user=postgres sslmode=disable",
					testutil.GetLocalHost()),
			},
		},
	}

	var acc testutil.Accumulator

	err := p.Gather(&acc)
	require.NoError(t, err)

	var found bool

	for _, pnt := range acc.Points {
		if pnt.Measurement == "datname" || pnt.Measurement == "datid" || pnt.Measurement == "stats_reset" {
			found = true
			break
		}
	}

	assert.False(t, found)
}
