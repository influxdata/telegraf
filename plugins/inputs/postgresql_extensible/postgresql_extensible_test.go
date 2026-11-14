package postgresql_extensible

import (
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
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

func TestPostgresqlSqlScript(t *testing.T) {
	q := []query{{
		Script:     "testdata/test.sql",
		MinVersion: 901,
		Withdbname: false,
		Tagvalue:   "",
	}}

	addr := fmt.Sprintf(
		"host=%s user=postgres sslmode=disable",
		testutil.GetLocalHost(),
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
}

func TestPostgresqlIgnoresUnwantedColumnsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	addr := fmt.Sprintf(
		"host=%s user=postgres sslmode=disable",
		testutil.GetLocalHost(),
	)

	p := &Postgresql{
		Log:     testutil.Logger{},
		Address: config.NewSecret([]byte(addr)),
	}
	require.NoError(t, p.Init())

	var acc testutil.Accumulator
	require.NoError(t, p.Start(&acc))
	defer p.Stop()
	require.NoError(t, acc.GatherError(p.Gather))

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

// TestInitDefaults checks the values Init fills in for a plugin created by its
// registered factory, as Telegraf does for a configuration without the new
// options
func TestInitDefaults(t *testing.T) {
	creator, found := inputs.Inputs["postgresql_extensible"]
	require.True(t, found)
	p, ok := creator().(*Postgresql)
	require.True(t, ok)
	p.Log = testutil.Logger{}
	p.Address = config.NewSecret(nil)
	p.Query = []query{{Sqlquery: "SELECT 1"}}
	require.NoError(t, p.Init())

	require.Zero(t, p.Timeout)
}

type testServer struct {
	container *testutil.Container
	address   string
}

func startServer(t *testing.T) *testServer {
	servicePort := "5432"
	container := &testutil.Container{
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

	address := fmt.Sprintf(
		"host=%s port=%s user=postgres dbname=postgres sslmode=disable",
		container.Address,
		container.Ports[servicePort],
	)

	return &testServer{
		container: container,
		address:   address,
	}
}

func (s *testServer) stop() {
	s.container.Terminate()
}

func (s *testServer) run(t *testing.T, p *Postgresql) *testutil.Accumulator {
	p.Log = testutil.Logger{}
	p.Address = config.NewSecret([]byte(s.address))
	if p.MaxOpen == 0 {
		p.MaxIdle = 1
		p.MaxOpen = 1
	}
	require.NoError(t, p.Init())

	var acc testutil.Accumulator
	require.NoError(t, p.Start(&acc))
	defer p.Stop()
	require.NoError(t, p.Gather(&acc))

	return &acc
}

func TestTimeoutUnresponsiveServer(t *testing.T) {
	// The connections to this listener are never answered
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	host, port, err := net.SplitHostPort(listener.Addr().String())
	require.NoError(t, err)
	server := &testServer{
		address: fmt.Sprintf("host=%s port=%s user=postgres dbname=postgres sslmode=disable", host, port),
	}

	acc := server.run(t, &Postgresql{
		Timeout: config.Duration(500 * time.Millisecond),
		Query:   []query{{Measurement: "not_started", Sqlquery: "SELECT 1 AS value"}},
	})

	require.Len(t, acc.Errors, 1)
	require.ErrorContains(t, acc.Errors[0], "querying metadata failed")
	require.Empty(t, acc.GetTelegrafMetrics())
}

func TestTimeoutsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	server := startServer(t)
	defer server.stop()

	t.Run("query", func(t *testing.T) {
		acc := server.run(t, &Postgresql{
			Query: []query{
				{Measurement: "slow", Sqlquery: "SELECT pg_sleep(5), 1 AS value", Timeout: config.Duration(500 * time.Millisecond)},
				{Measurement: "fast", Sqlquery: "SELECT 1 AS value"},
			},
		})

		require.Len(t, acc.Errors, 1)
		require.ErrorContains(t, acc.Errors[0], `query "slow": query timeout 500ms exceeded`)
		require.False(t, acc.HasMeasurement("slow"))
		require.True(t, acc.HasInt64Field("fast", "value"))
	})

	t.Run("collection", func(t *testing.T) {
		acc := server.run(t, &Postgresql{
			Timeout: config.Duration(500 * time.Millisecond),
			Query: []query{
				{Measurement: "slow", Sqlquery: "SELECT pg_sleep(5), 1 AS value"},
				{Measurement: "not_started", Sqlquery: "SELECT 1 AS value"},
			},
		})

		require.Len(t, acc.Errors, 1)
		require.ErrorContains(t, acc.Errors[0], `collection timeout 500ms exceeded in database "postgres": query "slow"`)
		require.False(t, acc.HasMeasurement("not_started"))
	})

	t.Run("query while reading rows", func(t *testing.T) {
		acc := server.run(t, &Postgresql{
			Query: []query{
				{
					Measurement: "slow",
					// Every row is larger than the send buffer of the server, so
					// the first rows arrive long before the query is done
					Sqlquery: "SELECT repeat('x', 10000) AS pad, pg_sleep(0.1)::text AS s FROM generate_series(1, 50)",
					Timeout:  config.Duration(time.Second),
				},
				{Measurement: "fast", Sqlquery: "SELECT 1 AS value"},
			},
		})

		require.Len(t, acc.Errors, 1)
		require.ErrorContains(t, acc.Errors[0], `query "slow": query timeout 1s exceeded`)
		require.True(t, acc.HasMeasurement("slow"))
		require.True(t, acc.HasInt64Field("fast", "value"))
	})
}
