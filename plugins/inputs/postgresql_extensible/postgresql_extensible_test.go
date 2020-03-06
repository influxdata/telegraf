package postgresql_extensible

import (
	"errors"
	"fmt"
	"testing"

	"github.com/influxdata/telegraf/plugins/inputs/postgresql"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	p.Start(&acc)
	p.Init()
	require.NoError(t, acc.GatherError(p.Gather))
	return &acc
}

func TestPostgresqlGeneratesMetrics(t *testing.T) {
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
		acc := queryRunner(t, query{{
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
	p.Start(&acc)
	p.Init()

	require.NoError(t, acc.GatherError(p.Gather))
}

func TestPostgresqlIgnoresUnwantedColumns(t *testing.T) {
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
	assert.NotEmpty(t, p.IgnoredColumns())
	for col := range p.IgnoredColumns() {
		assert.False(t, acc.HasMeasurement(col))
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
		return errors.New("Nada matchy buddy")
	}

	for i, d := range dest {
		switch d.(type) {
		case (*interface{}):
			*d.(*interface{}) = f.fields[i]
		default:
			return fmt.Errorf("Bad type %T", d)
		}
	}
	return nil
}
