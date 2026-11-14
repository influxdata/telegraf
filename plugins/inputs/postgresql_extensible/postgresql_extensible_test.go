package postgresql_extensible

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/postgresql"
	"github.com/influxdata/telegraf/testutil"
)

func queryRunner(t *testing.T, q []query) *testutil.Accumulator {
	servicePort := "5432"
	container := testutil.Container{
		Image:        "postgres:alpine",
		ExposedPorts: []string{servicePort},
		Env: map[string]string{
			"POSTGRES_HOST_AUTH_METHOD": "trust",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
			wait.ForListeningPort(servicePort),
		),
	}

	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()

	addr := fmt.Sprintf(
		"host=%s port=%s user=postgres sslmode=disable",
		container.Address,
		container.Ports[servicePort],
	)

	p := &Postgresql{
		Log:         testutil.Logger{},
		Address:     config.NewSecret([]byte(addr)),
		IsPgBouncer: false,
		Databases:   []string{"postgres"},
		Query:       q,
	}
	require.NoError(t, p.Init())

	var acc testutil.Accumulator
	require.NoError(t, p.Start(&acc))
	defer p.Stop()
	require.NoError(t, acc.GatherError(p.Gather))

	return &acc
}

func TestPostgresqlGeneratesMetricsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	acc := queryRunner(t, []query{{
		Sqlquery:   "select * from pg_stat_database",
		MinVersion: 901,
		Withdbname: false,
		Tagvalue:   "",
	}})
	testutil.PrintMetrics(acc.GetTelegrafMetrics())

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

	for _, metric := range floatMetrics {
		require.True(t, acc.HasFloatField("postgresql", metric))
		metricsCounted++
	}

	for _, metric := range stringMetrics {
		require.True(t, acc.HasStringField("postgresql", metric))
		metricsCounted++
	}

	require.Positive(t, metricsCounted)
	require.Equal(t, len(floatMetrics)+len(intMetrics)+len(stringMetrics), metricsCounted)
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
			require.InDelta(t, 10.0, v, testutil.DefaultDelta)
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
			require.True(t, v)
		},
		"SELECT timestamp'1980-07-23' as ts, true AS myvalue": func(acc *testutil.Accumulator) {
			expectedTime := time.Date(1980, 7, 23, 0, 0, 0, 0, time.UTC)
			v, found := acc.BoolField(measurement, "myvalue")
			require.True(t, found)
			require.True(t, v)
			require.True(t, acc.HasTimestamp(measurement, expectedTime))
		},
	}

	for q, assertions := range examples {
		acc := queryRunner(t, []query{{
			Sqlquery:   q,
			MinVersion: 901,
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

	acc := queryRunner(t, []query{{
		Sqlquery:   "select * from pg_stat_database",
		MinVersion: 901,
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

	floatMetrics := []string{
		"blk_read_time",
		"blk_write_time",
	}

	stringMetrics := []string{
		"datname",
	}

	for _, field := range intMetrics {
		_, found := acc.Int64Field(measurement, field)
		require.Truef(t, found, "expected %s to be an integer", field)
	}

	for _, field := range floatMetrics {
		_, found := acc.FloatField(measurement, field)
		require.Truef(t, found, "expected %s to be a float64", field)
	}

	for _, field := range stringMetrics {
		_, found := acc.StringField(measurement, field)
		require.Truef(t, found, "expected %s to be a str", field)
	}
}

func TestPostgresqlSqlScriptIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	q := []query{{
		Script:     "testdata/test.sql",
		MinVersion: 901,
		Withdbname: false,
		Tagvalue:   "",
	}}
	queryRunner(t, q)
}

func TestPostgresqlIgnoresUnwantedColumnsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	acc := queryRunner(t, nil)

	require.NotEmpty(t, ignoredColumns)
	for col := range ignoredColumns {
		require.False(t, acc.HasMeasurement(col))
	}
}

func TestAccRow(t *testing.T) {
	p := Postgresql{
		Log:           testutil.Logger{},
		Address:       config.NewSecret(nil),
		OutputAddress: "server",
	}
	require.NoError(t, p.Init())

	var acc testutil.Accumulator
	columns := []string{"datname", "cat"}

	tests := []struct {
		fields fakeRow
		dbName string
		server string
	}{
		{
			fields: fakeRow{
				fields: []any{1, "gato"},
			},
			dbName: "postgres",
			server: "server",
		},
		{
			fields: fakeRow{
				fields: []any{nil, "gato"},
			},
			dbName: "postgres",
			server: "server",
		},
		{
			fields: fakeRow{
				fields: []any{"name", "gato"},
			},
			dbName: "name",
			server: "server",
		},
	}
	for _, tt := range tests {
		q := query{Measurement: "pgTEST", additionalTags: make(map[string]bool)}
		require.NoError(t, p.accRow(&acc, tt.fields, columns, q, time.Now()))
		require.Len(t, acc.Metrics, 1)
		metric := acc.Metrics[0]
		require.Equal(t, tt.dbName, metric.Tags["db"])
		require.Equal(t, tt.server, metric.Tags["server"])
		acc.ClearMetrics()
	}
}

func TestGatherRetriesVersionQuery(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// The first attempt hits a connection closed by the server
	mock.ExpectQuery("server_version_num").WillReturnError(errors.New("unexpected EOF"))
	mock.ExpectQuery("server_version_num").WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow(1700))
	mock.ExpectQuery("SELECT 'new'").WillReturnRows(sqlmock.NewRows([]string{"which"}).AddRow("new"))

	p := Postgresql{
		Log:     testutil.Logger{},
		service: &postgresql.Service{DB: db},
		Query: []query{
			{Sqlquery: "SELECT 'old' AS which", MaxVersion: 1400, Measurement: "test"},
			{Sqlquery: "SELECT 'new' AS which", MinVersion: 1400, Measurement: "test"},
		},
	}

	var acc testutil.Accumulator
	require.NoError(t, p.Gather(&acc))
	require.Empty(t, acc.Errors)
	require.NoError(t, mock.ExpectationsWereMet())
	require.Len(t, acc.Metrics, 1)
	require.Equal(t, "new", acc.Metrics[0].Fields["which"])
}

func TestGatherVersionQueryFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("server_version_num").WillReturnError(errors.New("unexpected EOF"))
	mock.ExpectQuery("server_version_num").WillReturnError(errors.New("connection refused"))

	p := Postgresql{
		Log:     testutil.Logger{},
		service: &postgresql.Service{DB: db},
		Query:   []query{{Sqlquery: "SELECT 'old' AS which", MaxVersion: 1400, Measurement: "test"}},
	}

	var acc testutil.Accumulator
	require.ErrorContains(t, p.Gather(&acc), "connection refused")
	require.NoError(t, mock.ExpectationsWereMet())
	require.Empty(t, acc.Metrics)
}

type fakeRow struct {
	fields []any
}

func (f fakeRow) Scan(dest ...any) error {
	if len(f.fields) != len(dest) {
		return errors.New("nada matchy buddy")
	}

	for i, d := range dest {
		switch d := d.(type) {
		case *any:
			*d = f.fields[i]
		default:
			return fmt.Errorf("bad type %T", d)
		}
	}
	return nil
}
