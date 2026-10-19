package postgresql_extensible

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/postgresql"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/testutil"
)

func newPlugin() *Postgresql {
	return &Postgresql{
		Log: testutil.Logger{},
		Config: postgresql.Config{
			Address:       config.NewSecret(nil),
			OutputAddress: "server",
			MaxIdle:       1,
			MaxOpen:       1,
		},
		PreparedStatements: true,
		NumericAsFloat:     true,
	}
}

// TestInitDefaults checks the values Init fills in, both for the plugin and
// for a query inheriting from it or overriding it
func TestInitDefaults(t *testing.T) {
	p := newPlugin()
	p.Queries = []query{{Sqlquery: "SELECT 1"}}
	require.NoError(t, p.Init())

	require.Zero(t, p.Timeout)
	require.Zero(t, p.MetadataRefreshInterval)
	require.Equal(t, 1, p.MaxConnectionsPerDatabase)
	require.Equal(t, 1, p.MaxParallelDatabases)
	require.Equal(t, "any", p.Role)

	q := p.Queries[0]
	require.Equal(t, "postgresql", q.Measurement)
	require.Equal(t, "any", q.Role)
	require.False(t, q.stringTags)
	require.True(t, q.numericFloat)

	// A query takes the settings of the plugin unless it sets its own
	p = newPlugin()
	p.Role = "primary"
	p.StringColumnsAsTags = true
	p.NumericAsFloat = true
	p.Queries = []query{
		{Sqlquery: "SELECT 1"},
		{
			Sqlquery:            "SELECT 2",
			Role:                "replica",
			StringColumnsAsTags: boolPtr(false),
			NumericAsFloat:      boolPtr(false),
		},
	}
	require.NoError(t, p.Init())

	inherited := p.Queries[0]
	require.Equal(t, "primary", inherited.Role)
	require.True(t, inherited.stringTags)
	require.True(t, inherited.numericFloat)

	overridden := p.Queries[1]
	require.Equal(t, "replica", overridden.Role)
	require.False(t, overridden.stringTags)
	require.False(t, overridden.numericFloat)
}

// TestRegisteredDefaults pins the defaults of a registered plugin: no
// timeout, no caching, numeric columns as strings and the connection kept
// between collections.
func TestRegisteredDefaults(t *testing.T) {
	creator, found := inputs.Inputs["postgresql_extensible"]
	require.True(t, found)

	p, ok := creator().(*Postgresql)
	require.True(t, ok)

	require.Zero(t, p.Timeout)
	require.Zero(t, p.MetadataRefreshInterval)
	require.Zero(t, p.MaxConnectionsPerDatabase)
	require.Zero(t, p.MaxParallelDatabases)
	require.Zero(t, p.Role)
	require.False(t, p.NumericAsFloat)
	require.False(t, p.StringColumnsAsTags)
	require.True(t, p.KeepIdleConnections)
	require.True(t, p.PreparedStatements)
	require.Equal(t, 1, p.MaxIdle)
	require.Equal(t, 1, p.MaxOpen)
}

func TestInitDeprecatedOptions(t *testing.T) {
	tests := []struct {
		name      string
		databases []string
		query     query
		expected  string
	}{
		{
			name:     "version replaces min_version",
			query:    query{Sqlquery: "SELECT 1", Version: 901},
			expected: "SELECT 1",
		},
		{
			name:     "withdbname without databases",
			query:    query{Sqlquery: "SELECT * FROM pg_stat_database WHERE datname", Withdbname: true},
			expected: "SELECT * FROM pg_stat_database WHERE datname is not null",
		},
		{
			name:      "withdbname with databases",
			databases: []string{"app_1", "app_2"},
			query:     query{Sqlquery: "SELECT * FROM pg_stat_database WHERE datname", Withdbname: true},
			expected:  "SELECT * FROM pg_stat_database WHERE datname IN ('app_1','app_2')",
		},
		{
			name:      "databases without withdbname",
			databases: []string{"app_1"},
			query:     query{Sqlquery: "SELECT 1"},
			expected:  "SELECT 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newPlugin()
			p.Databases = tt.databases
			p.Queries = []query{tt.query}
			require.NoError(t, p.Init())

			require.Equal(t, tt.expected, p.Queries[0].Sqlquery)
			if tt.query.Version != 0 {
				require.Equal(t, tt.query.Version, p.Queries[0].MinVersion)
			}
		})
	}
}

func TestInitScript(t *testing.T) {
	p := newPlugin()
	p.Queries = []query{{Script: "testdata/test.sql"}}
	require.NoError(t, p.Init())
	require.Equal(t, "select * from pg_stat_database", p.Queries[0].Sqlquery)
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
			expected: `invalid role "master"`,
		},
		{
			name: "invalid query role",
			plugin: func() *Postgresql {
				p := newPlugin()
				p.Queries = []query{{Sqlquery: "SELECT 1", Role: "standby"}}
				return p
			}(),
			expected: `query 0: invalid role "standby"`,
		},
		{
			name: "invalid include regex",
			plugin: func() *Postgresql {
				p := newPlugin()
				p.DatnameInclude = []string{"app_("}
				return p
			}(),
			expected: `compiling datname_include "app_(" failed`,
		},
		{
			name: "invalid exclude regex",
			plugin: func() *Postgresql {
				p := newPlugin()
				p.DatnameExclude = []string{"[a-"}
				return p
			}(),
			expected: `compiling datname_exclude "[a-" failed`,
		},
		{
			name: "missing script",
			plugin: func() *Postgresql {
				p := newPlugin()
				p.Queries = []query{{Script: "testdata/does_not_exist.sql"}}
				return p
			}(),
			expected: "no such file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.plugin.Init()
			require.ErrorContains(t, err, tt.expected)
		})
	}
}

func TestMatchesDatname(t *testing.T) {
	tests := []struct {
		name     string
		include  []string
		exclude  []string
		datname  string
		expected bool
	}{
		{name: "no filters", datname: "app", expected: true},
		{name: "include exact", include: []string{"app"}, datname: "app", expected: true},
		{name: "include anchored", include: []string{"app"}, datname: "myapp", expected: false},
		{name: "include anchored suffix", include: []string{"app"}, datname: "app2", expected: false},
		{name: "include regex", include: []string{"app_.*"}, datname: "app_prod", expected: true},
		{name: "include regex no match", include: []string{"app_.*"}, datname: "billing", expected: false},
		{name: "include multiple", include: []string{"app_.*", "billing"}, datname: "billing", expected: true},
		{name: "exclude only", exclude: []string{".*_test"}, datname: "app_test", expected: false},
		{name: "exclude only no match", exclude: []string{".*_test"}, datname: "app_prod", expected: true},
		{name: "include and exclude", include: []string{"app_.*"}, exclude: []string{".*_test"}, datname: "app_test", expected: false},
		{name: "include and exclude pass", include: []string{"app_.*"}, exclude: []string{".*_test"}, datname: "app_prod", expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newPlugin()
			p.DatnameInclude = tt.include
			p.DatnameExclude = tt.exclude
			require.NoError(t, p.Init())
			require.Equal(t, tt.expected, p.matchesDatname(tt.datname))
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

func TestMatchesVersion(t *testing.T) {
	tests := []struct {
		name     string
		query    query
		version  int
		expected bool
	}{
		{name: "no limits", query: query{}, version: 1500, expected: true},
		{name: "min version met", query: query{MinVersion: 901}, version: 1500, expected: true},
		{name: "min version not met", query: query{MinVersion: 1600}, version: 1500, expected: false},
		{name: "max version exclusive", query: query{MaxVersion: 1500}, version: 1500, expected: false},
		{name: "max version below", query: query{MaxVersion: 1600}, version: 1500, expected: true},
		{name: "unknown version", query: query{MinVersion: 901}, version: 0, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, matchesVersion(&tt.query, tt.version))
		})
	}
}

func TestAccRow(t *testing.T) {
	tests := []struct {
		name           string
		query          query
		columns        []string
		numeric        map[string]bool
		values         []interface{}
		expectedTags   map[string]string
		expectedFields map[string]interface{}
	}{
		{
			name:    "db tag from job",
			columns: []string{"cat"},
			values:  []interface{}{"gato"},
			expectedTags: map[string]string{
				"server": "server",
				"db":     "mydb",
			},
			expectedFields: map[string]interface{}{"cat": "gato"},
		},
		{
			name:    "db tag from datname column",
			columns: []string{"datname", "cat"},
			values:  []interface{}{"name", "gato"},
			expectedTags: map[string]string{
				"server": "server",
				"db":     "name",
			},
			expectedFields: map[string]interface{}{"datname": "name", "cat": "gato"},
		},
		{
			name:    "non-string datname column is ignored",
			columns: []string{"datname", "cat"},
			values:  []interface{}{1, "gato"},
			expectedTags: map[string]string{
				"server": "server",
				"db":     "mydb",
			},
			expectedFields: map[string]interface{}{"datname": 1, "cat": "gato"},
		},
		{
			name:    "null and stats_reset columns are skipped",
			columns: []string{"stats_reset", "nothing", "value"},
			values:  []interface{}{time.Now(), nil, int64(42)},
			expectedTags: map[string]string{
				"server": "server",
				"db":     "mydb",
			},
			expectedFields: map[string]interface{}{"value": int64(42)},
		},
		{
			name:    "tagvalue columns become tags",
			query:   query{Tagvalue: "state,pid"},
			columns: []string{"state", "pid", "count"},
			values:  []interface{}{"idle", int64(7), int64(3)},
			expectedTags: map[string]string{
				"server": "server",
				"db":     "mydb",
				"state":  "idle",
				"pid":    "7",
			},
			expectedFields: map[string]interface{}{"count": int64(3)},
		},
		{
			name:    "string columns as tags",
			query:   query{StringColumnsAsTags: boolPtr(true)},
			columns: []string{"relname", "raw", "count"},
			values:  []interface{}{"users", []byte("bytes"), int64(3)},
			expectedTags: map[string]string{
				"server":  "server",
				"db":      "mydb",
				"relname": "users",
				"raw":     "bytes",
			},
			expectedFields: map[string]interface{}{"count": int64(3)},
		},
		{
			name:    "string columns as fields",
			columns: []string{"relname", "raw"},
			values:  []interface{}{"users", []byte("bytes")},
			expectedTags: map[string]string{
				"server": "server",
				"db":     "mydb",
			},
			expectedFields: map[string]interface{}{"relname": "users", "raw": "bytes"},
		},
		{
			name:    "numeric as float",
			columns: []string{"ratio", "text", "big", "nan", "bad"},
			numeric: map[string]bool{"ratio": true, "big": true, "nan": true, "bad": true},
			values:  []interface{}{"10.5", "10.5", []byte("1e3"), "NaN", "abc"},
			expectedTags: map[string]string{
				"server": "server",
				"db":     "mydb",
			},
			expectedFields: map[string]interface{}{"ratio": 10.5, "text": "10.5", "big": 1000.0},
		},
		{
			name:    "numeric as string",
			query:   query{NumericAsFloat: boolPtr(false)},
			columns: []string{"ratio"},
			numeric: map[string]bool{"ratio": true},
			values:  []interface{}{"10.5"},
			expectedTags: map[string]string{
				"server": "server",
				"db":     "mydb",
			},
			expectedFields: map[string]interface{}{"ratio": "10.5"},
		},
		{
			name:    "timestamp column",
			query:   query{Timestamp: "ts"},
			columns: []string{"ts", "value"},
			values:  []interface{}{time.Date(1980, 7, 23, 0, 0, 0, 0, time.UTC), int64(1)},
			expectedTags: map[string]string{
				"server": "server",
				"db":     "mydb",
			},
			expectedFields: map[string]interface{}{"value": int64(1)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newPlugin()
			tt.query.Sqlquery = "SELECT 1"
			tt.query.Measurement = "pgTEST"
			p.Queries = []query{tt.query}
			require.NoError(t, p.Init())

			var acc testutil.Accumulator
			j := job{datname: "mydb", query: &p.Queries[0]}
			now := time.Now()
			p.accRow(&acc, tt.columns, tt.values, tt.numeric, j, now, false)
			require.Len(t, acc.Metrics, 1)

			metric := acc.Metrics[0]
			require.Equal(t, "pgTEST", metric.Measurement)
			require.Equal(t, tt.expectedTags, metric.Tags)
			require.Equal(t, tt.expectedFields, metric.Fields)
			if tt.query.Timestamp != "" {
				require.Equal(t, time.Date(1980, 7, 23, 0, 0, 0, 0, time.UTC), metric.Time)
			} else {
				require.Equal(t, now, metric.Time)
			}
		})
	}
}

func TestMetadataCached(t *testing.T) {
	p := newPlugin()
	p.MetadataRefreshInterval = config.Duration(time.Hour)
	require.NoError(t, p.Init())

	// The pool is not started, so any query would panic on a nil database
	p.dbVersion = 1700
	p.inRecovery = true
	p.datnames = []string{"app_1"}
	p.metadataUpdated = time.Now()

	require.NoError(t, p.refreshMetadata(context.Background()))
	require.Equal(t, 1700, p.dbVersion)
	require.True(t, p.inRecovery)
	require.Equal(t, []string{"app_1"}, p.datnames)
}

func TestRefreshDue(t *testing.T) {
	const interval = 2 * time.Minute
	updated := time.Date(2026, 9, 20, 17, 46, 0, 50*int(time.Millisecond), time.UTC)

	tests := []struct {
		name     string
		updated  time.Time
		now      time.Time
		interval time.Duration
		expected bool
	}{
		{name: "never updated", now: updated, interval: interval, expected: true},
		{name: "zero interval reads every time", updated: updated, now: updated, interval: 0, expected: true},
		{name: "same tick", updated: updated, now: updated, interval: interval, expected: false},
		{name: "within the interval", updated: updated, now: updated.Add(time.Minute), interval: interval, expected: false},
		{
			name:     "boundary reached slightly earlier than the update time",
			updated:  updated,
			now:      time.Date(2026, 9, 20, 17, 48, 0, 20*int(time.Millisecond), time.UTC),
			interval: interval,
			expected: true,
		},
		{
			name:     "boundary reached exactly",
			updated:  updated,
			now:      time.Date(2026, 9, 20, 17, 48, 0, 0, time.UTC),
			interval: interval,
			expected: true,
		},
		{
			name:     "just before the boundary",
			updated:  updated,
			now:      time.Date(2026, 9, 20, 17, 47, 59, 0, time.UTC),
			interval: interval,
			expected: false,
		},
		{name: "long overdue", updated: updated, now: updated.Add(time.Hour), interval: interval, expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, refreshDue(tt.updated, tt.now, tt.interval))
		})
	}
}

func boolPtr(v bool) *bool {
	return &v
}

// Integration tests

type testServer struct {
	container *testutil.Container
	address   string
	host      string
	port      string
}

// dsn builds a connection string for the given database, optionally tagging
// the sessions with an application_name so a test can recognize them in
// pg_stat_activity
func (s *testServer) dsn(datname, appName string) string {
	dsn := fmt.Sprintf("host=%s port=%s user=postgres dbname=%s sslmode=disable", s.host, s.port, datname)
	if appName != "" {
		dsn += " application_name=" + appName
	}
	return dsn
}

// seedSchema creates a few tables with indexes in the given database so the
// table and index statistics views return rows
func (s *testServer) seedSchema(t *testing.T, datname string) {
	db, err := sql.Open("pgx", s.dsn(datname, ""))
	require.NoError(t, err)
	defer db.Close()

	for _, table := range []string{"orders", "customers", "invoices"} {
		_, err := db.Exec("CREATE TABLE " + table + " (id integer PRIMARY KEY, label text, amount numeric)")
		require.NoErrorf(t, err, "creating table %q in %q failed", table, datname)
		_, err = db.Exec("CREATE INDEX ON " + table + " (label)")
		require.NoErrorf(t, err, "creating index on %q in %q failed", table, datname)
		_, err = db.Exec("INSERT INTO " + table + " SELECT g, 'row-' || g, g * 1.5 FROM generate_series(1, 50) AS g")
		require.NoErrorf(t, err, "filling table %q in %q failed", table, datname)
	}
	_, err = db.Exec("ANALYZE")
	require.NoError(t, err)
}

func startServer(t *testing.T, databases ...string) *testServer {
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

	if len(databases) > 0 {
		db, err := sql.Open("pgx", address)
		require.NoError(t, err)
		defer db.Close()

		for _, name := range databases {
			_, err := db.Exec("CREATE DATABASE " + name)
			require.NoErrorf(t, err, "creating database %q failed", name)
		}
	}

	return &testServer{
		container: container,
		address:   address,
		host:      container.Address,
		port:      container.Ports[servicePort],
	}
}

func (s *testServer) stop() {
	s.container.Terminate()
}

func (s *testServer) run(t *testing.T, p *Postgresql) *testutil.Accumulator {
	p.Log = testutil.Logger{}
	p.Config.Address = config.NewSecret([]byte(s.address))
	if p.Config.MaxOpen == 0 {
		p.Config.MaxIdle = 1
		p.Config.MaxOpen = 1
	}
	require.NoError(t, p.Init())

	var acc testutil.Accumulator
	require.NoError(t, p.Start(&acc))
	defer p.Stop()
	require.NoError(t, p.Gather(&acc))

	return &acc
}

func dbTags(metrics []telegraf.Metric, measurement string) map[string]int {
	result := make(map[string]int)
	for _, m := range metrics {
		if m.Name() != measurement {
			continue
		}
		db, _ := m.GetTag("db")
		result[db]++
	}
	return result
}

func TestGeneratesMetricsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	server := startServer(t)
	defer server.stop()

	acc := server.run(t, &Postgresql{
		NumericAsFloat: true,
		Queries: []query{{
			Sqlquery:   "SELECT * FROM pg_stat_database WHERE datname = 'postgres'",
			MinVersion: 901,
		}},
	})
	require.Empty(t, acc.Errors)

	// A few representative columns of each type. Naming them all would make
	// the test fail on a PostgreSQL release that changes the view.
	for _, metric := range []string{"numbackends", "xact_commit", "blks_read"} {
		require.Truef(t, acc.HasInt64Field("postgresql", metric), "expected %s to be an integer", metric)
	}
	require.True(t, acc.HasFloatField("postgresql", "blk_read_time"))
	require.True(t, acc.HasStringField("postgresql", "datname"))
	require.False(t, acc.HasField("postgresql", "stats_reset"))
	require.Equal(t, "postgres", acc.TagValue("postgresql", "db"))
}

func TestMultipleDatabasesIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	server := startServer(t, "app_1", "app_2", "skip_me")
	defer server.stop()

	acc := server.run(t, &Postgresql{
		NumericAsFloat: true,
		DatnameInclude: []string{"app_.*", "postgres"},
		DatnameExclude: []string{"app_2"},
		Queries: []query{{
			Measurement: "current",
			Sqlquery:    "SELECT current_database() AS name, 1 AS value",
		}},
	})
	require.Empty(t, acc.Errors)

	expected := map[string]int{"app_1": 1, "postgres": 1}
	require.Equal(t, expected, dbTags(acc.GetTelegrafMetrics(), "current"))

	// The db tag must reflect the database the query was run in
	for _, m := range acc.GetTelegrafMetrics() {
		db, _ := m.GetTag("db")
		name, _ := m.GetField("name")
		require.Equal(t, db, name)
	}
}

// TestKeepConnectionsIntegration counts the sessions still open after a
// collection, which is what keep_idle_connections decides
func TestKeepConnectionsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	server := startServer(t, "app_1", "app_2")
	defer server.stop()

	const appName = "pgext_keep_connections"

	// Observe the server from an independent connection, as the sessions of
	// the plugin can only be counted once its collection is done. The
	// observer carries an application_name of its own, so it never counts
	// itself.
	observer, err := sql.Open("pgx", server.dsn("postgres", "keep_observer"))
	require.NoError(t, err)
	defer observer.Close()
	observer.SetMaxOpenConns(1)

	sessions := func() int {
		var count int
		query := `SELECT count(*) FROM pg_stat_activity WHERE application_name = $1`
		require.NoError(t, observer.QueryRow(query, appName).Scan(&count))
		return count
	}

	tests := []struct {
		name     string
		keep     bool
		include  []string
		expected int
	}{
		{name: "single database closed", keep: false, expected: 0},
		{name: "single database kept", keep: true, expected: 3},
		{name: "multiple databases closed", keep: false, include: []string{"app_.*"}, expected: 0},
		{name: "multiple databases kept", keep: true, include: []string{"app_.*"}, expected: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Postgresql{
				Log:                       testutil.Logger{},
				Config:                    postgresql.Config{Address: config.NewSecret([]byte(server.dsn("postgres", appName))), MaxIdle: 1, MaxOpen: 1},
				NumericAsFloat:            true,
				MaxConnectionsPerDatabase: 3,
				KeepIdleConnections:       tt.keep,
				DatnameInclude:            tt.include,
				Queries: []query{
					{Measurement: "q1", Sqlquery: "SELECT pg_sleep(0.1), 1 AS v"},
					{Measurement: "q2", Sqlquery: "SELECT pg_sleep(0.1), 2 AS v"},
					{Measurement: "q3", Sqlquery: "SELECT pg_sleep(0.1), 3 AS v"},
				},
			}
			require.NoError(t, p.Init())

			var acc testutil.Accumulator
			require.NoError(t, p.Start(&acc))
			defer p.Stop()
			require.NoError(t, p.Gather(&acc))
			require.Empty(t, acc.Errors)

			// Give the server a moment to register closed sessions
			require.Eventually(t, func() bool {
				return sessions() == tt.expected
			}, 5*time.Second, 100*time.Millisecond)
		})
	}
}

func TestQueryTimeoutIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	server := startServer(t)
	defer server.stop()

	acc := server.run(t, &Postgresql{
		NumericAsFloat: true,
		Queries: []query{
			{Measurement: "slow", Sqlquery: "SELECT pg_sleep(5), 1 AS value", Timeout: config.Duration(500 * time.Millisecond)},
			{Measurement: "fast", Sqlquery: "SELECT 1 AS value"},
		},
	})

	require.Len(t, acc.Errors, 1)
	require.ErrorContains(t, acc.Errors[0], `query "slow"`)
	require.False(t, acc.HasMeasurement("slow"))
	require.True(t, acc.HasInt64Field("fast", "value"))
}

func TestCachedRoleIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	server := startServer(t)
	defer server.stop()

	p := &Postgresql{
		Log:                     testutil.Logger{},
		Config:                  postgresql.Config{Address: config.NewSecret([]byte(server.address)), MaxIdle: 1, MaxOpen: 1},
		NumericAsFloat:          true,
		MetadataRefreshInterval: config.Duration(time.Hour),
		Queries:                 []query{{Measurement: "primary_only", Sqlquery: "SELECT 1 AS value", Role: "primary"}},
	}
	require.NoError(t, p.Init())

	var acc testutil.Accumulator
	require.NoError(t, p.Start(&acc))
	defer p.Stop()

	// The container is a primary, so the query runs
	require.NoError(t, p.Gather(&acc))
	require.Empty(t, acc.Errors)
	require.True(t, acc.HasMeasurement("primary_only"))

	// Pretend the server became a replica without the cache knowing about
	// it: the next collection must use the cached role instead of asking
	// the server again
	acc.ClearMetrics()
	p.inRecovery = true
	require.NoError(t, p.Gather(&acc))
	require.Empty(t, acc.Errors)
	require.False(t, acc.HasMeasurement("primary_only"))
}

func TestRoleFilterIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	server := startServer(t)
	defer server.stop()

	// The container is always a primary
	acc := server.run(t, &Postgresql{
		NumericAsFloat: true,
		Queries: []query{
			{Measurement: "primary_only", Sqlquery: "SELECT 1 AS value", Role: "primary"},
			{Measurement: "replica_only", Sqlquery: "SELECT 1 AS value", Role: "replica"},
			{Measurement: "everywhere", Sqlquery: "SELECT 1 AS value"},
		},
	})
	require.Empty(t, acc.Errors)

	require.True(t, acc.HasMeasurement("primary_only"))
	require.False(t, acc.HasMeasurement("replica_only"))
	require.True(t, acc.HasMeasurement("everywhere"))
}

// TestColumnTypesIntegration checks the mapping of the column types to the
// field types
func TestColumnTypesIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	server := startServer(t)
	defer server.stop()

	tests := []struct {
		name   string
		sql    string
		plugin func(q *query, p *Postgresql)
		check  func(t *testing.T, acc *testutil.Accumulator)
	}{
		{
			name: "float",
			sql:  "SELECT 10.0::float AS myvalue",
			check: func(t *testing.T, acc *testutil.Accumulator) {
				v, found := acc.FloatField("postgresql", "myvalue")
				require.True(t, found)
				require.InDelta(t, 10.0, v, testutil.DefaultDelta)
			},
		},
		{
			name: "numeric as string",
			sql:  "SELECT 10.0 AS myvalue",
			check: func(t *testing.T, acc *testutil.Accumulator) {
				v, found := acc.StringField("postgresql", "myvalue")
				require.True(t, found)
				require.Equal(t, "10.0", v)
			},
		},
		{
			name:   "numeric as float",
			sql:    "SELECT 10.5::numeric AS myvalue",
			plugin: func(_ *query, p *Postgresql) { p.NumericAsFloat = true },
			check: func(t *testing.T, acc *testutil.Accumulator) {
				v, found := acc.FloatField("postgresql", "myvalue")
				require.True(t, found)
				require.InDelta(t, 10.5, v, testutil.DefaultDelta)
			},
		},
		{
			name: "query overrides numeric as float",
			sql:  "SELECT 10.5::numeric AS myvalue",
			plugin: func(q *query, p *Postgresql) {
				p.NumericAsFloat = true
				q.NumericAsFloat = boolPtr(false)
			},
			check: func(t *testing.T, acc *testutil.Accumulator) {
				v, found := acc.StringField("postgresql", "myvalue")
				require.True(t, found)
				require.Equal(t, "10.5", v)
			},
		},
		{
			name: "string",
			sql:  "SELECT 'hello world' AS myvalue",
			check: func(t *testing.T, acc *testutil.Accumulator) {
				v, found := acc.StringField("postgresql", "myvalue")
				require.True(t, found)
				require.Equal(t, "hello world", v)
			},
		},
		{
			name: "boolean",
			sql:  "SELECT true AS myvalue",
			check: func(t *testing.T, acc *testutil.Accumulator) {
				v, found := acc.BoolField("postgresql", "myvalue")
				require.True(t, found)
				require.True(t, v)
			},
		},
		{
			name: "timestamp column",
			sql:  "SELECT timestamp'1980-07-23' AS ts, true AS myvalue",
			check: func(t *testing.T, acc *testutil.Accumulator) {
				v, found := acc.BoolField("postgresql", "myvalue")
				require.True(t, found)
				require.True(t, v)
				require.True(t, acc.HasTimestamp("postgresql", time.Date(1980, 7, 23, 0, 0, 0, 0, time.UTC)))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := query{Sqlquery: tt.sql, MinVersion: 901, Timestamp: "ts"}
			p := &Postgresql{}
			if tt.plugin != nil {
				tt.plugin(&q, p)
			}
			p.Queries = []query{q}

			acc := server.run(t, p)
			require.Empty(t, acc.Errors)
			tt.check(t, acc)
		})
	}
}

// connSampler watches pg_stat_activity from an independent connection and
// records how many sessions the plugin had open at the same time
type connSampler struct {
	db      *sql.DB
	appName string
	done    chan struct{}
	stopped chan struct{}

	mu           sync.Mutex
	maxPerDB     map[string]int
	maxDBsAtOnce int
	maxTotal     int
	samples      int
}

func newConnSampler(t *testing.T, server *testServer, appName string) *connSampler {
	db, err := sql.Open("pgx", server.dsn("postgres", "conn_sampler"))
	require.NoError(t, err)

	s := &connSampler{
		db:       db,
		appName:  appName,
		done:     make(chan struct{}),
		stopped:  make(chan struct{}),
		maxPerDB: make(map[string]int),
	}

	go func() {
		defer close(s.stopped)
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-s.done:
				s.sample()
				return
			case <-ticker.C:
				s.sample()
			}
		}
	}()

	return s
}

func (s *connSampler) sample() {
	rows, err := s.db.Query(
		`SELECT datname, count(*) FROM pg_stat_activity WHERE application_name = $1 GROUP BY datname`,
		s.appName)
	if err != nil {
		return
	}
	defer rows.Close()

	current := make(map[string]int)
	total := 0
	for rows.Next() {
		var datname string
		var n int
		if err := rows.Scan(&datname, &n); err != nil {
			return
		}
		current[datname] = n
		total += n
	}
	if rows.Err() != nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.samples++
	for datname, n := range current {
		if n > s.maxPerDB[datname] {
			s.maxPerDB[datname] = n
		}
	}
	if len(current) > s.maxDBsAtOnce {
		s.maxDBsAtOnce = len(current)
	}
	if total > s.maxTotal {
		s.maxTotal = total
	}
}

func (s *connSampler) stop() {
	close(s.done)
	<-s.stopped
	s.db.Close()
}

// TestConnectionLimitsIntegration proves that the plugin never opens more
// connections than max_parallel_databases * max_connections_per_database by
// counting the real sessions in pg_stat_activity while a collection runs
func TestConnectionLimitsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	server := startServer(t, "app_1", "app_2", "app_3", "app_4", "app_5")
	defer server.stop()
	for _, datname := range []string{"postgres", "app_1", "app_2", "app_3", "app_4", "app_5"} {
		server.seedSchema(t, datname)
	}

	// Queries slow enough for the sampler to observe them. pg_sleep is put in
	// a CTE so it runs once per query instead of once per row.
	const sleep = "0.4"
	trivial := func(n int) string {
		return "WITH s AS (SELECT pg_sleep(" + sleep + ")) SELECT " + fmt.Sprint(n) + " AS value FROM s"
	}
	tableStats := "WITH s AS (SELECT pg_sleep(" + sleep + ")) " +
		"SELECT relname, n_live_tup, n_dead_tup FROM pg_stat_user_tables, s"
	indexStats := "WITH s AS (SELECT pg_sleep(" + sleep + ")) " +
		"SELECT indexrelname, idx_scan, idx_tup_read FROM pg_stat_user_indexes, s"

	tests := []struct {
		name string

		datnameInclude []string
		datnameExclude []string
		parallelDBs    int
		connsPerDB     int
		queries        []query
		// coldMetadata measures the first collection, which still has to read
		// the server version, the role and the database list
		coldMetadata bool

		wantDatnames    []string
		wantMaxPerDB    int
		wantMaxDBs      int
		wantMaxTotalCap int
	}{
		{
			name:        "single database, three queries",
			parallelDBs: 1,
			connsPerDB:  3,
			queries: []query{
				{Measurement: "q1", Sqlquery: trivial(1)},
				{Measurement: "q2", Sqlquery: trivial(2)},
				{Measurement: "q3", Sqlquery: trivial(3)},
			},
			wantDatnames:    []string{"postgres"},
			wantMaxPerDB:    3,
			wantMaxDBs:      1,
			wantMaxTotalCap: 3,
		},
		{
			name:           "five databases without postgres",
			datnameExclude: []string{"postgres"},
			parallelDBs:    2,
			connsPerDB:     2,
			queries: []query{
				{Measurement: "pg_tables", Sqlquery: tableStats, Tagvalue: "relname"},
				{Measurement: "pg_indexes", Sqlquery: indexStats, Tagvalue: "indexrelname"},
			},
			wantDatnames:    []string{"app_1", "app_2", "app_3", "app_4", "app_5"},
			wantMaxPerDB:    2,
			wantMaxDBs:      2,
			wantMaxTotalCap: 4,
		},
		{
			name:        "all databases including postgres",
			parallelDBs: 2,
			connsPerDB:  2,
			// Match every database so the connection database is processed
			// like any other one
			datnameInclude: []string{".*"},
			queries: []query{
				{Measurement: "pg_tables", Sqlquery: tableStats, Tagvalue: "relname"},
				{Measurement: "pg_indexes", Sqlquery: indexStats, Tagvalue: "indexrelname"},
			},
			wantDatnames:    []string{"app_1", "app_2", "app_3", "app_4", "app_5", "postgres"},
			wantMaxPerDB:    2,
			wantMaxDBs:      2,
			wantMaxTotalCap: 4,
		},
		{
			name:           "five databases with a cold metadata cache",
			datnameExclude: []string{"postgres"},
			parallelDBs:    2,
			connsPerDB:     2,
			coldMetadata:   true,
			queries: []query{
				{Measurement: "pg_tables", Sqlquery: tableStats, Tagvalue: "relname"},
				{Measurement: "pg_indexes", Sqlquery: indexStats, Tagvalue: "indexrelname"},
			},
			wantDatnames: []string{"app_1", "app_2", "app_3", "app_4", "app_5"},
			wantMaxPerDB: 2,
			// The connection reading the server version, the role and the
			// database list belongs to the connection database, which is not
			// one of the databases being queried here
			wantMaxDBs:      3,
			wantMaxTotalCap: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appName := "pgext_" + strings.ReplaceAll(tt.name, " ", "_")

			metadataInterval := config.Duration(time.Hour)
			if tt.coldMetadata {
				metadataInterval = 0
			}

			p := &Postgresql{
				Log: testutil.Logger{},
				Config: postgresql.Config{
					Address: config.NewSecret([]byte(server.dsn("postgres", appName))),
					MaxIdle: 1,
					MaxOpen: 1,
				},
				PreparedStatements: true,
				// Read the server version, the role and the database list on
				// the first collection only, and keep no connection open
				// afterwards, so the measured collection shows the query
				// connections alone
				MetadataRefreshInterval:   metadataInterval,
				KeepIdleConnections:       false,
				NumericAsFloat:            true,
				DatnameInclude:            tt.datnameInclude,
				DatnameExclude:            tt.datnameExclude,
				MaxParallelDatabases:      tt.parallelDBs,
				MaxConnectionsPerDatabase: tt.connsPerDB,
				Queries:                   tt.queries,
			}
			require.NoError(t, p.Init())

			var acc testutil.Accumulator
			require.NoError(t, p.Start(&acc))
			defer p.Stop()

			if !tt.coldMetadata {
				// First collection fills the metadata cache and closes its
				// connections, the second one is the one being measured
				require.NoError(t, p.Gather(&acc))
				require.Empty(t, acc.Errors)
				acc.ClearMetrics()
			}

			sampler := newConnSampler(t, server, appName)
			require.NoError(t, p.Gather(&acc))
			sampler.stop()
			require.Empty(t, acc.Errors)

			// Every expected database produced metrics for every query
			for _, q := range tt.queries {
				got := make([]string, 0, len(tt.wantDatnames))
				for datname := range dbTags(acc.GetTelegrafMetrics(), q.Measurement) {
					got = append(got, datname)
				}
				sort.Strings(got)
				require.Equalf(t, tt.wantDatnames, got, "databases for measurement %q", q.Measurement)
			}

			sampler.mu.Lock()
			defer sampler.mu.Unlock()

			datnames := make([]string, 0, len(sampler.maxPerDB))
			for datname := range sampler.maxPerDB {
				datnames = append(datnames, datname)
			}
			sort.Strings(datnames)

			t.Logf("max_parallel_databases=%d max_connections_per_database=%d queries=%d",
				tt.parallelDBs, tt.connsPerDB, len(tt.queries))
			t.Logf("%-12s %s", "database", "peak connections")
			for _, datname := range datnames {
				t.Logf("%-12s %d", datname, sampler.maxPerDB[datname])
			}
			t.Logf("%-12s %d", "TOTAL", sampler.maxTotal)
			t.Logf("databases seen at the same time: %d (samples taken: %d)", sampler.maxDBsAtOnce, sampler.samples)

			require.Positive(t, sampler.samples, "the sampler never ran")
			require.NotEmpty(t, sampler.maxPerDB, "no sessions of the plugin were observed")

			for _, datname := range datnames {
				require.LessOrEqualf(t, sampler.maxPerDB[datname], tt.connsPerDB,
					"database %q exceeded max_connections_per_database", datname)
			}
			require.Equal(t, tt.wantMaxDBs, sampler.maxDBsAtOnce,
				"number of databases processed at the same time")
			require.LessOrEqual(t, sampler.maxTotal, tt.wantMaxTotalCap,
				"total number of connections")

			// At least one database must have reached the per-database limit,
			// otherwise the test would pass without any concurrency at all
			var reached bool
			for _, datname := range datnames {
				if sampler.maxPerDB[datname] == tt.wantMaxPerDB {
					reached = true
					break
				}
			}
			require.Truef(t, reached, "no database reached %d connections", tt.wantMaxPerDB)
		})
	}
}

func BenchmarkAccRow(b *testing.B) {
	columns := []string{
		"datname", "relname", "schemaname", "n_live_tup", "n_dead_tup",
		"seq_scan", "idx_scan", "table_size_b", "bloat_pct", "last_vacuum",
	}
	row := []interface{}{
		"app_1", "users", "public", int64(20000), int64(6667),
		int64(15), int64(3), int64(5652480), 34.6, "2026-09-22",
	}

	for _, tt := range []struct {
		name  string
		debug bool
	}{{name: "debug off", debug: false}, {name: "debug on", debug: true}} {
		b.Run(tt.name, func(b *testing.B) {
			p := newPlugin()
			p.Log = testutil.Logger{Quiet: true}
			p.Queries = []query{{Sqlquery: "SELECT 1", Measurement: "bench", Tagvalue: "relname,schemaname"}}
			require.NoError(b, p.Init())

			var acc testutil.Accumulator
			j := job{datname: "app_1", query: &p.Queries[0]}
			now := time.Now()

			// The row buffers are prepared once per result set, as in
			// gatherMetricsFromQuery
			values := make([]interface{}, len(columns))
			pointers := make([]interface{}, len(columns))
			for i := range values {
				pointers[i] = &values[i]
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// simulate rows.Scan writing into the reused buffer
				for k := range row {
					*(pointers[k].(*interface{})) = row[k]
				}
				p.accRow(&acc, columns, values, nil, j, now, tt.debug)
				acc.ClearMetrics()
			}
		})
	}
}
