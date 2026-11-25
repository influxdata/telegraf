package postgresql_extensible

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/testutil"
)

func newPlugin() *Postgresql {
	return &Postgresql{
		Log:                testutil.Logger{},
		Address:            config.NewSecret(nil),
		OutputAddress:      "server",
		MaxIdle:            1,
		MaxOpen:            1,
		PreparedStatements: true,
	}
}

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
// options, and for a query inheriting from the plugin or overriding it
func TestInitDefaults(t *testing.T) {
	creator, found := inputs.Inputs["postgresql_extensible"]
	require.True(t, found)
	p, ok := creator().(*Postgresql)
	require.True(t, ok)
	p.Log = testutil.Logger{}
	p.Address = config.NewSecret(nil)
	p.Query = []query{{Sqlquery: "SELECT 1"}}
	require.NoError(t, p.Init())

	require.Equal(t, "any", p.Role)
	require.Equal(t, "any", p.Query[0].Role)

	// A query takes the settings of the plugin unless it sets its own
	p = newPlugin()
	p.Role = "primary"
	p.Query = []query{
		{Sqlquery: "SELECT 1"},
		{Sqlquery: "SELECT 2", Role: "replica"},
	}
	require.NoError(t, p.Init())

	require.Equal(t, "primary", p.Query[0].Role)
	require.Equal(t, "replica", p.Query[1].Role)
}

func TestInitErrors(t *testing.T) {
	tests := []struct {
		name     string
		plugin   *Postgresql
		expected string
	}{
		{
			name: "invalid plugin role",
			plugin: func() *Postgresql {
				p := newPlugin()
				p.Role = "master"
				return p
			}(),
			expected: "config option 'role': unknown choice master",
		},
		{
			name: "invalid query role",
			plugin: func() *Postgresql {
				p := newPlugin()
				p.Query = []query{{Sqlquery: "SELECT 1", Role: "standby"}}
				return p
			}(),
			expected: "query 0: config option 'role': unknown choice standby",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.plugin.Init()
			require.ErrorContains(t, err, tt.expected)
		})
	}
}

func TestMatchesRole(t *testing.T) {
	require.True(t, matchesRole("any", false))
	require.True(t, matchesRole("any", true))
	require.True(t, matchesRole("primary", false))
	require.False(t, matchesRole("primary", true))
	require.False(t, matchesRole("replica", false))
	require.True(t, matchesRole("replica", true))
}

func TestRoleFilterIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// The container is always a primary
	acc := queryRunner(t, []query{
		{Measurement: "primary_only", Sqlquery: "SELECT 1 AS value", Role: "primary"},
		{Measurement: "replica_only", Sqlquery: "SELECT 1 AS value", Role: "replica"},
		{Measurement: "everywhere", Sqlquery: "SELECT 1 AS value"},
	})

	require.True(t, acc.HasMeasurement("primary_only"))
	require.False(t, acc.HasMeasurement("replica_only"))
	require.True(t, acc.HasMeasurement("everywhere"))
}

func TestMetadataFailure(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		expected []string
	}{
		// Every query is still tried and reports its own error
		{name: "without role", role: "", expected: []string{"connect", "connect"}},
		// The queries are not run without knowing the server role
		{name: "with role", role: "primary", expected: []string{"querying server role failed"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Nothing listens on this port, so reading the server information fails
			p := newPlugin()
			p.Address = config.NewSecret([]byte("host=127.0.0.1 port=1 user=postgres sslmode=disable connect_timeout=5"))
			p.Query = []query{
				{Measurement: "role", Sqlquery: "SELECT 1 AS value", Role: tt.role},
				{Measurement: "everywhere", Sqlquery: "SELECT 1 AS value"},
			}
			require.NoError(t, p.Init())

			var acc testutil.Accumulator
			require.NoError(t, p.Start(&acc))
			defer p.Stop()
			require.NoError(t, p.Gather(&acc))

			require.Len(t, acc.Errors, len(tt.expected))
			for i, expected := range tt.expected {
				require.ErrorContains(t, acc.Errors[i], expected)
			}
			require.Empty(t, acc.GetTelegrafMetrics())
		})
	}
}
