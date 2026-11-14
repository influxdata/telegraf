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
		require.NoError(t, p.accRow(&acc, tt.fields, &resultSet{columns: columns}, q, time.Now()))
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

	q := p.Query[0]
	require.False(t, q.stringTags)
	require.False(t, q.numericFloat)

	// A query takes the settings of the plugin unless it sets its own
	p = newPlugin()
	p.StringColumnsAsTags = true
	p.NumericType = "float"
	p.Query = []query{
		{Sqlquery: "SELECT 1"},
		{
			Sqlquery:            "SELECT 2",
			StringColumnsAsTags: new(false),
			NumericType:         "string",
		},
	}
	require.NoError(t, p.Init())

	inherited := p.Query[0]
	require.True(t, inherited.stringTags)
	require.True(t, inherited.numericFloat)

	overridden := p.Query[1]
	require.False(t, overridden.stringTags)
	require.False(t, overridden.numericFloat)
}

func TestInitErrors(t *testing.T) {
	tests := []struct {
		name     string
		plugin   *Postgresql
		expected string
	}{
		{
			name: "invalid numeric type",
			plugin: func() *Postgresql {
				p := newPlugin()
				p.NumericType = "decimal"
				return p
			}(),
			expected: "config option 'numeric_type'",
		},
		{
			name: "invalid query numeric type",
			plugin: func() *Postgresql {
				p := newPlugin()
				p.Query = []query{{Sqlquery: "SELECT 1", NumericType: "decimal"}}
				return p
			}(),
			expected: "query 0: config option 'numeric_type'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.plugin.Init()
			require.ErrorContains(t, err, tt.expected)
		})
	}
}

func TestAccRowColumnTypes(t *testing.T) {
	tests := []struct {
		name           string
		query          query
		columns        []string
		numeric        map[string]bool
		text           map[string]bool
		values         []any
		expectedTags   map[string]string
		expectedFields map[string]any
	}{
		{
			name:    "string columns as tags",
			query:   query{StringColumnsAsTags: new(true)},
			columns: []string{"relname", "raw", "count"},
			text:    map[string]bool{"relname": true, "raw": true},
			values:  []any{"users", []byte("bytes"), int64(3)},
			expectedTags: map[string]string{
				"server":  "server",
				"db":      "postgres",
				"relname": "users",
				"raw":     "bytes",
			},
			expectedFields: map[string]any{"count": int64(3)},
		},
		{
			name:    "string columns as tags skip numeric columns",
			query:   query{StringColumnsAsTags: new(true), NumericType: "string"},
			columns: []string{"relname", "ratio", "uuid"},
			numeric: map[string]bool{"ratio": true},
			text:    map[string]bool{"relname": true},
			values:  []any{"users", "2.5", "a0eebc99"},
			expectedTags: map[string]string{
				"server":  "server",
				"db":      "postgres",
				"relname": "users",
			},
			expectedFields: map[string]any{"ratio": "2.5", "uuid": "a0eebc99"},
		},
		{
			name:    "string columns as fields",
			columns: []string{"relname", "raw"},
			values:  []any{"users", []byte("bytes")},
			expectedTags: map[string]string{
				"server": "server",
				"db":     "postgres",
			},
			expectedFields: map[string]any{"relname": "users", "raw": "bytes"},
		},
		{
			name:    "numeric as float",
			query:   query{NumericType: "float"},
			columns: []string{"ratio", "text", "big", "nan", "inf", "bad"},
			numeric: map[string]bool{"ratio": true, "big": true, "nan": true, "inf": true, "bad": true},
			values:  []any{"10.5", "10.5", []byte("1e3"), "NaN", "Infinity", "abc"},
			expectedTags: map[string]string{
				"server": "server",
				"db":     "postgres",
			},
			expectedFields: map[string]any{"ratio": 10.5, "text": "10.5", "big": 1000.0},
		},
		{
			name:    "numeric as string",
			query:   query{NumericType: "string"},
			columns: []string{"ratio"},
			numeric: map[string]bool{"ratio": true},
			values:  []any{"10.5"},
			expectedTags: map[string]string{
				"server": "server",
				"db":     "postgres",
			},
			expectedFields: map[string]any{"ratio": "10.5"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newPlugin()
			tt.query.Sqlquery = "SELECT 1"
			tt.query.Measurement = "pgTEST"
			p.Query = []query{tt.query}
			require.NoError(t, p.Init())

			var acc testutil.Accumulator
			rs := &resultSet{columns: tt.columns, numeric: tt.numeric, text: tt.text}
			require.NoError(t, p.accRow(&acc, fakeRow{fields: tt.values}, rs, p.Query[0], time.Now()))
			require.Len(t, acc.Metrics, 1)

			metric := acc.Metrics[0]
			require.Equal(t, "pgTEST", metric.Measurement)
			require.Equal(t, tt.expectedTags, metric.Tags)
			require.Equal(t, tt.expectedFields, metric.Fields)
		})
	}
}

func TestColumnTypesIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tests := []struct {
		name           string
		query          query
		expectedTags   map[string]string
		expectedFields map[string]any
	}{
		{
			name:           "numeric as float",
			query:          query{Sqlquery: "SELECT 10.5::numeric AS myvalue", NumericType: "float"},
			expectedFields: map[string]any{"myvalue": 10.5},
		},
		{
			name: "string columns as tags",
			query: query{
				Sqlquery: "SELECT 'a'::text AS t, 'b'::varchar AS v, 'c'::char(1) AS c, 'd'::name AS n, " +
					"'e'::\"char\" AS k, 2 / 0.8 AS ratio, 3 AS count",
				StringColumnsAsTags: new(true),
			},
			expectedTags:   map[string]string{"t": "a", "v": "b", "c": "c", "n": "d", "k": "e"},
			expectedFields: map[string]any{"ratio": "2.5000000000000000", "count": int64(3)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := queryRunner(t, []query{tt.query})
			require.Empty(t, acc.Errors)
			require.Len(t, acc.Metrics, 1)

			m := acc.Metrics[0]
			require.Equal(t, tt.expectedFields, m.Fields)
			// The db and server tags come on top of the expected ones
			require.Len(t, m.Tags, len(tt.expectedTags)+2)
			for k, v := range tt.expectedTags {
				require.Equal(t, v, m.Tags[k], k)
			}
		})
	}
}
