package postgresql_extensible

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/plugins/inputs/postgresql"
	"github.com/influxdata/telegraf/testutil"
)

func queryRunner(t *testing.T, q query) *testutil.Accumulator {
	p := &Postgresql{
		Log: testutil.Logger{},
		Service: postgresql.Service{
			Address: fmt.Sprintf(
				"host=%s user=postgres sslmode=disable",
				testutil.GetLocalHost(),
			),
			IsPgBouncer: false,
		},
		Databases: []string{"postgres"},
		Query:     q,
	}
	var acc testutil.Accumulator
	require.NoError(t, p.Init())
	require.NoError(t, p.Start(&acc))
	require.NoError(t, acc.GatherError(p.Gather))
	return &acc
}

func TestPostgresqlGeneratesMetricsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	acc := queryRunner(t, query{{
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

	metricsCounted := 0

	for _, metric := range intMetrics {
		require.True(t, acc.HasInt64Field("postgresql", metric))
		metricsCounted++
	}

	for _, metric := range int32Metrics {
		require.True(t, acc.HasInt32Field("postgresql", metric))
		metricsCounted++
	}

	for _, metric := range floatMetrics {
		require.True(t, acc.HasFloatField("postgresql", metric))
		metricsCounted++
	}

	for _, metric := range stringMetrics {
		require.True(t, acc.HasStringField("postgresql", metric))
		metricsCounted++
	}

	require.True(t, metricsCounted > 0)
	require.Equal(t, len(floatMetrics)+len(intMetrics)+len(int32Metrics)+len(stringMetrics), metricsCounted)
}

func TestPostgresqlQueryOutputTestsIntegration(t *testing.T) {
	const measurement = "postgresql"

	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	examples := map[string]func(*testutil.Accumulator){
		"SELECT 10.0::float AS myvalue": func(acc *testutil.Accumulator) {
			v, found := acc.FloatField(measurement, "myvalue")
			require.True(t, found)
			require.Equal(t, 10.0, v)
		},
		"SELECT 10.0 AS myvalue": func(acc *testutil.Accumulator) {
			v, found := acc.StringField(measurement, "myvalue")
			require.True(t, found)
			require.Equal(t, "10.0", v)
		},
		"SELECT 'hello world' AS myvalue": func(acc *testutil.Accumulator) {
			v, found := acc.StringField(measurement, "myvalue")
			require.True(t, found)
			require.Equal(t, "hello world", v)
		},
		"SELECT true AS myvalue": func(acc *testutil.Accumulator) {
			v, found := acc.BoolField(measurement, "myvalue")
			require.True(t, found)
			require.Equal(t, true, v)
		},
		"SELECT timestamp'1980-07-23' as ts, true AS myvalue": func(acc *testutil.Accumulator) {
			expectedTime := time.Date(1980, 7, 23, 0, 0, 0, 0, time.UTC)
			v, found := acc.BoolField(measurement, "myvalue")
			require.True(t, found)
			require.Equal(t, true, v)
			require.True(t, acc.HasTimestamp(measurement, expectedTime))
		},
	}

	for q, assertions := range examples {
		acc := queryRunner(t, query{{
			Sqlquery:   q,
			Version:    901,
			Withdbname: false,
			Tagvalue:   "",
			Timestamp:  "ts",
		}})
		assertions(acc)
	}
}

func TestPostgresqlFieldOutputIntegration(t *testing.T) {
	const measurement = "postgresql"
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	acc := queryRunner(t, query{{
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
		require.True(t, found, fmt.Sprintf("expected %s to be an integer", field))
	}

	for _, field := range int32Metrics {
		_, found := acc.Int32Field(measurement, field)
		require.True(t, found, fmt.Sprintf("expected %s to be an int32", field))
	}

	for _, field := range floatMetrics {
		_, found := acc.FloatField(measurement, field)
		require.True(t, found, fmt.Sprintf("expected %s to be a float64", field))
	}

	for _, field := range stringMetrics {
		_, found := acc.StringField(measurement, field)
		require.True(t, found, fmt.Sprintf("expected %s to be a str", field))
	}
}

func TestPostgresqlSqlScript(t *testing.T) {
	q := query{{
		Script:     "testdata/test.sql",
		Version:    901,
		Withdbname: false,
		Tagvalue:   "",
	}}
	p := &Postgresql{
		Log: testutil.Logger{},
		Service: postgresql.Service{
			Address: fmt.Sprintf(
				"host=%s user=postgres sslmode=disable",
				testutil.GetLocalHost(),
			),
			IsPgBouncer: false,
		},
		Databases: []string{"postgres"},
		Query:     q,
	}
	var acc testutil.Accumulator
	require.NoError(t, p.Init())
	require.NoError(t, p.Start(&acc))

	require.NoError(t, acc.GatherError(p.Gather))
}

func TestPostgresqlIgnoresUnwantedColumnsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p := &Postgresql{
		Log: testutil.Logger{},
		Service: postgresql.Service{
			Address: fmt.Sprintf(
				"host=%s user=postgres sslmode=disable",
				testutil.GetLocalHost(),
			),
		},
	}

	var acc testutil.Accumulator

	require.NoError(t, p.Start(&acc))
	require.NoError(t, acc.GatherError(p.Gather))
	require.NotEmpty(t, p.IgnoredColumns())
	for col := range p.IgnoredColumns() {
		require.False(t, acc.HasMeasurement(col))
	}
}

func TestAccRow(t *testing.T) {
	p := Postgresql{
		Log: testutil.Logger{},
	}

	var acc testutil.Accumulator
	columns := []string{"datname", "cat"}

	testRows := []fakeRow{
		{fields: []interface{}{1, "gato"}},
		{fields: []interface{}{nil, "gato"}},
		{fields: []interface{}{"name", "gato"}},
	}
	for i := range testRows {
		err := p.accRow("pgTEST", testRows[i], &acc, columns)
		if err != nil {
			t.Fatalf("Scan failed: %s", err)
		}
	}
}

type fakeRow struct {
	fields []interface{}
}

func (f fakeRow) Scan(dest ...interface{}) error {
	if len(f.fields) != len(dest) {
		return errors.New("nada matchy buddy")
	}

	for i, d := range dest {
		switch d := d.(type) {
		case *interface{}:
			*d = f.fields[i]
		default:
			return fmt.Errorf("bad type %T", d)
		}
	}
	return nil
}
