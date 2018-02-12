package postgresql_extensible

import (
	"fmt"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func queryRunner(t *testing.T, q query) (*Postgresql, *testutil.Accumulator) {
	p := &Postgresql{
		Address: fmt.Sprintf("host=%s user=postgres sslmode=disable",
			testutil.GetLocalHost()),
		Databases: []string{"postgres"},
		Query:     q,
	}
	var acc testutil.Accumulator

	require.NoError(t, acc.GatherError(p.Gather))
	return p, &acc
}

func TestPostgresqlGeneratesMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p, acc := queryRunner(t, query{{
		Sqlquery:   "select * from pg_stat_database",
		Version:    901,
		Withdbname: false,
		Tagvalue:   "",
	}})

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
		"datid",
	}

	int32Metrics := []string{}

	floatMetrics := []string{
		"blk_read_time",
		"blk_write_time",
	}

	stringMetrics := []string{
		"datname",
	}

	metricsCounted := 0

	for _, metric := range intMetrics {
		_, ok := availableColumns[metric]
		if ok {
			assert.True(t, acc.HasInt64Field("postgresql", metric))
			metricsCounted++
		}
	}

	for _, metric := range int32Metrics {
		_, ok := availableColumns[metric]
		if ok {
			assert.True(t, acc.HasInt32Field("postgresql", metric))
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

	for _, metric := range stringMetrics {
		_, ok := availableColumns[metric]
		if ok {
			assert.True(t, acc.HasStringField("postgresql", metric))
			metricsCounted++
		}
	}

	assert.True(t, metricsCounted > 0)
	assert.Equal(t, len(availableColumns)-len(p.IgnoredColumns()), metricsCounted)
}

func TestPostgresqlQueryOutputTests(t *testing.T) {
	const measurement = "postgresql"

	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	examples := map[string]func(*testutil.Accumulator){
		"SELECT 10.0::float AS myvalue": func(acc *testutil.Accumulator) {
			v, found := acc.FloatField(measurement, "myvalue")
			assert.True(t, found)
			assert.Equal(t, 10.0, v)
		},
		"SELECT 10.0 AS myvalue": func(acc *testutil.Accumulator) {
			v, found := acc.StringField(measurement, "myvalue")
			assert.True(t, found)
			assert.Equal(t, "10.0", v)
		},
		"SELECT 'hello world' AS myvalue": func(acc *testutil.Accumulator) {
			v, found := acc.StringField(measurement, "myvalue")
			assert.True(t, found)
			assert.Equal(t, "hello world", v)
		},
		"SELECT true AS myvalue": func(acc *testutil.Accumulator) {
			v, found := acc.BoolField(measurement, "myvalue")
			assert.True(t, found)
			assert.Equal(t, true, v)
		},
	}

	for q, assertions := range examples {
		_, acc := queryRunner(t, query{{
			Sqlquery:   q,
			Version:    901,
			Withdbname: false,
			Tagvalue:   "",
		}})
		assertions(acc)
	}
}

func TestPostgresqlFieldOutput(t *testing.T) {
	const measurement = "postgresql"
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	_, acc := queryRunner(t, query{{
		Sqlquery:   "select * from pg_stat_database",
		Version:    901,
		Withdbname: false,
		Tagvalue:   "",
	}})

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
		"datid",
	}

	int32Metrics := []string{}

	floatMetrics := []string{
		"blk_read_time",
		"blk_write_time",
	}

	stringMetrics := []string{
		"datname",
	}

	for _, field := range intMetrics {
		_, found := acc.Int64Field(measurement, field)
		assert.True(t, found, fmt.Sprintf("expected %s to be an integer", field))
	}

	for _, field := range int32Metrics {
		_, found := acc.Int32Field(measurement, field)
		assert.True(t, found, fmt.Sprintf("expected %s to be an int32", field))
	}

	for _, field := range floatMetrics {
		_, found := acc.FloatField(measurement, field)
		assert.True(t, found, fmt.Sprintf("expected %s to be a float64", field))
	}

	for _, field := range stringMetrics {
		_, found := acc.StringField(measurement, field)
		assert.True(t, found, fmt.Sprintf("expected %s to be a str", field))
	}
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
	require.NoError(t, acc.GatherError(p.Gather))

	assert.NotEmpty(t, p.IgnoredColumns())
	for col := range p.IgnoredColumns() {
		assert.False(t, acc.HasMeasurement(col))
	}
}
