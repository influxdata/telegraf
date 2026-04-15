package snowpipe_streaming

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

// ---------------------------------------------------------------------------
// Mock SQL driver
// ---------------------------------------------------------------------------

// mockDriver records every query executed against it.
type mockDriver struct{}

type mockConn struct {
	mu      sync.Mutex
	queries []executedQuery
	closed  bool

	// If set, Exec returns this error for the first N calls
	execErr      error
	execErrCount int32 // atomic: how many execs should fail
}

type executedQuery struct {
	query string
	args  []driver.Value
}

type mockStmt struct {
	conn  *mockConn
	query string
}

type mockTx struct {
	conn *mockConn
}

type mockRows struct {
	columns []string
	data    [][]driver.Value
	pos     int
}

var (
	globalMockConn *mockConn
	globalMockMu   sync.Mutex
)

func resetGlobalMock() *mockConn {
	globalMockMu.Lock()
	defer globalMockMu.Unlock()
	globalMockConn = &mockConn{}
	return globalMockConn
}

func getGlobalMock() *mockConn {
	globalMockMu.Lock()
	defer globalMockMu.Unlock()
	return globalMockConn
}

func init() {
	sql.Register("snowflake_mock", &mockDriver{})
}

func (*mockDriver) Open(_ string) (driver.Conn, error) {
	c := getGlobalMock()
	if c == nil {
		return nil, errors.New("no mock conn configured")
	}
	return c, nil
}

func (c *mockConn) Prepare(query string) (driver.Stmt, error) {
	return &mockStmt{conn: c, query: query}, nil
}

func (c *mockConn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
	return nil
}

func (c *mockConn) Begin() (driver.Tx, error) {
	return &mockTx{conn: c}, nil
}

func (c *mockConn) Exec(query string, args []driver.Value) (driver.Result, error) {
	c.mu.Lock()
	c.queries = append(c.queries, executedQuery{query: query, args: args})
	c.mu.Unlock()

	if c.execErr != nil {
		remaining := atomic.AddInt32(&c.execErrCount, -1)
		if remaining >= 0 {
			return nil, c.execErr
		}
	}

	return mockResult{}, nil
}

func (c *mockConn) Query(query string, args []driver.Value) (driver.Rows, error) {
	c.mu.Lock()
	c.queries = append(c.queries, executedQuery{query: query, args: args})
	c.mu.Unlock()

	if strings.Contains(strings.ToUpper(query), "INFORMATION_SCHEMA.COLUMNS") {
		return &mockRows{
			columns: []string{"COLUMN_NAME"},
			data:    make([][]driver.Value, 0),
		}, nil
	}

	return &mockRows{columns: make([]string, 0), data: make([][]driver.Value, 0)}, nil
}

func (c *mockConn) getQueries() []executedQuery {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]executedQuery, len(c.queries))
	copy(out, c.queries)
	return out
}

func (*mockStmt) Close() error  { return nil }
func (*mockStmt) NumInput() int { return -1 }

func (s *mockStmt) Exec(args []driver.Value) (driver.Result, error) {
	return s.conn.Exec(s.query, args)
}

func (s *mockStmt) Query(args []driver.Value) (driver.Rows, error) {
	return s.conn.Query(s.query, args)
}

func (*mockTx) Commit() error   { return nil }
func (*mockTx) Rollback() error { return nil }

type mockResult struct{}

func (mockResult) LastInsertId() (int64, error) { return 0, nil }
func (mockResult) RowsAffected() (int64, error) { return 1, nil }

func (r *mockRows) Columns() []string { return r.columns }
func (*mockRows) Close() error        { return nil }
func (r *mockRows) Next(dest []driver.Value) error {
	if r.pos >= len(r.data) {
		return io.EOF
	}
	copy(dest, r.data[r.pos])
	r.pos++
	return nil
}

// ---------------------------------------------------------------------------
// Helper: create a plugin wired to the mock driver
// ---------------------------------------------------------------------------

func newTestPlugin(t *testing.T) *SnowpipeStreaming {
	t.Helper()
	s := &SnowpipeStreaming{
		Account:             "test_account",
		User:                "test_user",
		Database:            "TEST_DB",
		Schema:              "PUBLIC",
		Table:               "METRICS",
		BatchSize:           1000,
		RetryMax:            3,
		RetryDelay:          config.Duration(10 * time.Millisecond),
		TimestampColumn:     "timestamp",
		TableSchemaCacheTTL: config.Duration(5 * time.Minute),
		Log:                 testutil.Logger{},
	}
	require.NoError(t, s.Init())
	return s
}

func connectTestPlugin(t *testing.T, s *SnowpipeStreaming) {
	t.Helper()
	mc := resetGlobalMock()
	_ = mc

	s.openDB = func() (*sql.DB, error) {
		return sql.Open("snowflake_mock", "mock")
	}
	require.NoError(t, s.Connect())
	t.Cleanup(func() { s.Close() })
}

func testMetric(name string, tags map[string]string, fields map[string]interface{}, ts time.Time) telegraf.Metric {
	m := metric.New(name, tags, fields, ts)
	return m
}

// ---------------------------------------------------------------------------
// Unit Tests
// ---------------------------------------------------------------------------

func TestInit(t *testing.T) {
	tests := []struct {
		name    string
		plugin  *SnowpipeStreaming
		wantErr string
	}{
		{
			name:    "missing account",
			plugin:  &SnowpipeStreaming{User: "u", Database: "d", Schema: "s", Table: "t"},
			wantErr: `"account" is required`,
		},
		{
			name:    "missing user",
			plugin:  &SnowpipeStreaming{Account: "a", Database: "d", Schema: "s", Table: "t"},
			wantErr: `"user" is required`,
		},
		{
			name:    "missing database",
			plugin:  &SnowpipeStreaming{Account: "a", User: "u", Schema: "s", Table: "t"},
			wantErr: `"database" is required`,
		},
		{
			name:    "missing schema",
			plugin:  &SnowpipeStreaming{Account: "a", User: "u", Database: "d", Table: "t"},
			wantErr: `"schema" is required`,
		},
		{
			name:    "missing table",
			plugin:  &SnowpipeStreaming{Account: "a", User: "u", Database: "d", Schema: "s"},
			wantErr: `"table" is required`,
		},
		{
			name:   "valid minimal config",
			plugin: &SnowpipeStreaming{Account: "a", User: "u", Database: "d", Schema: "s", Table: "t"},
		},
		{
			name:    "invalid table template",
			plugin:  &SnowpipeStreaming{Account: "a", User: "u", Database: "d", Schema: "s", Table: "{{.Invalid"},
			wantErr: "parsing table template",
		},
		{
			name:   "valid template table",
			plugin: &SnowpipeStreaming{Account: "a", User: "u", Database: "d", Schema: "s", Table: "metrics_{{.Name}}"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.plugin.Init()
			if tc.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMetricToRow(t *testing.T) {
	s := newTestPlugin(t)

	ts := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	m := testMetric("cpu", map[string]string{"host": "server1"}, map[string]interface{}{
		"usage_idle": float64(95.5),
		"count":      int64(42),
	}, ts)

	columns := s.buildColumnOrder(m)
	columnSet := make(map[string]bool, len(columns))
	for _, c := range columns {
		columnSet[c] = true
	}
	row := s.metricToRow(m, columns, columnSet)

	require.Contains(t, columns, "timestamp")
	require.Contains(t, columns, "name")
	require.Contains(t, columns, "host")
	require.Contains(t, columns, "usage_idle")
	require.Contains(t, columns, "count")

	// Check that row values match
	valMap := make(map[string]interface{}, len(columns))
	for i, col := range columns {
		valMap[col] = row[i]
	}

	require.Equal(t, ts, valMap["timestamp"])
	require.Equal(t, "cpu", valMap["name"])
	require.Equal(t, "server1", valMap["host"])
	require.InDelta(t, float64(95.5), valMap["usage_idle"], 1e-9)
	require.Equal(t, int64(42), valMap["count"])
}

func TestMetricToRowNaNInf(t *testing.T) {
	s := newTestPlugin(t)

	ts := time.Now()
	m := testMetric("test", nil, map[string]interface{}{
		"nan_val": math.NaN(),
		"ok_val":  float64(1.0),
	}, ts)

	columns := s.buildColumnOrder(m)
	columnSet := make(map[string]bool, len(columns))
	for _, c := range columns {
		columnSet[c] = true
	}
	row := s.metricToRow(m, columns, columnSet)

	valMap := make(map[string]interface{}, len(columns))
	for i, col := range columns {
		valMap[col] = row[i]
	}

	require.Nil(t, valMap["nan_val"])
	require.InDelta(t, float64(1.0), valMap["ok_val"], 1e-9)
}

func TestTableNameTemplate(t *testing.T) {
	s := &SnowpipeStreaming{
		Account:  "a",
		User:     "u",
		Database: "d",
		Schema:   "s",
		Table:    "metrics_{{.Name}}",
	}
	require.NoError(t, s.Init())

	m := testMetric("cpu", map[string]string{"host": "h1"}, map[string]interface{}{"val": 1.0}, time.Now())
	name := s.resolveTableName(m)
	require.Equal(t, "metrics_cpu", name)

	m2 := testMetric("mem", nil, map[string]interface{}{"val": 1.0}, time.Now())
	name2 := s.resolveTableName(m2)
	require.Equal(t, "metrics_mem", name2)
}

func TestTableNameNoTemplate(t *testing.T) {
	s := &SnowpipeStreaming{
		Account:  "a",
		User:     "u",
		Database: "d",
		Schema:   "s",
		Table:    "fixed_table",
	}
	require.NoError(t, s.Init())

	m := testMetric("cpu", nil, map[string]interface{}{"val": 1.0}, time.Now())
	require.Equal(t, "fixed_table", s.resolveTableName(m))
}

func TestBatching(t *testing.T) {
	s := newTestPlugin(t)
	s.BatchSize = 3
	connectTestPlugin(t, s)

	ts := time.Now()
	metrics := make([]telegraf.Metric, 7)
	for i := range 7 {
		metrics[i] = testMetric("cpu", nil, map[string]interface{}{
			"val": float64(i),
		}, ts)
	}

	require.NoError(t, s.Write(metrics))

	mc := getGlobalMock()
	queries := mc.getQueries()

	// With batch_size=3 and 7 metrics, expect 3 INSERT queries (3+3+1)
	insertCount := 0
	for _, q := range queries {
		if strings.HasPrefix(q.query, "INSERT INTO") {
			insertCount++
		}
	}
	require.Equal(t, 3, insertCount, "expected 3 batch inserts for 7 rows with batch_size=3")

	// Verify the last batch has only 1 row's worth of placeholders
	lastInsert := queries[len(queries)-1]
	require.Contains(t, lastInsert.query, "VALUES (?")
	// Count the number of VALUES groups
	valuesCount := strings.Count(lastInsert.query, "(?,")
	// Last batch should be 1 row only
	require.Equal(t, 1, valuesCount, "last batch should contain 1 row")
}

func TestRetryLogic(t *testing.T) {
	s := newTestPlugin(t)
	s.RetryMax = 2
	s.RetryDelay = config.Duration(1 * time.Millisecond)

	mc := resetGlobalMock()
	mc.execErr = errors.New("connection refused")
	atomic.StoreInt32(&mc.execErrCount, 2) // fail first 2, succeed on 3rd

	s.openDB = func() (*sql.DB, error) {
		return sql.Open("snowflake_mock", "mock")
	}
	require.NoError(t, s.Connect())
	t.Cleanup(func() { s.Close() })

	ts := time.Now()
	m := testMetric("test", nil, map[string]interface{}{"val": 1.0}, ts)
	require.NoError(t, s.Write([]telegraf.Metric{m}))

	queries := mc.getQueries()
	insertCount := 0
	for _, q := range queries {
		if strings.HasPrefix(q.query, "INSERT INTO") {
			insertCount++
		}
	}
	require.Equal(t, 3, insertCount, "expected 3 attempts (1 initial + 2 retries)")
}

func TestRetryExhausted(t *testing.T) {
	s := newTestPlugin(t)
	s.RetryMax = 1
	s.RetryDelay = config.Duration(1 * time.Millisecond)

	mc := resetGlobalMock()
	mc.execErr = errors.New("connection refused")
	atomic.StoreInt32(&mc.execErrCount, 100) // always fail

	s.openDB = func() (*sql.DB, error) {
		return sql.Open("snowflake_mock", "mock")
	}
	require.NoError(t, s.Connect())
	t.Cleanup(func() { s.Close() })

	ts := time.Now()
	m := testMetric("test", nil, map[string]interface{}{"val": 1.0}, ts)
	err := s.Write([]telegraf.Metric{m})
	require.Error(t, err)
	require.Contains(t, err.Error(), "after 1 retries")
}

func TestNonTransientErrorNoRetry(t *testing.T) {
	s := newTestPlugin(t)
	s.RetryMax = 3
	s.RetryDelay = config.Duration(1 * time.Millisecond)

	mc := resetGlobalMock()
	mc.execErr = errors.New("sql compilation error: invalid identifier")
	atomic.StoreInt32(&mc.execErrCount, 100)

	s.openDB = func() (*sql.DB, error) {
		return sql.Open("snowflake_mock", "mock")
	}
	require.NoError(t, s.Connect())
	t.Cleanup(func() { s.Close() })

	ts := time.Now()
	m := testMetric("test", nil, map[string]interface{}{"val": 1.0}, ts)
	err := s.Write([]telegraf.Metric{m})
	require.Error(t, err)
	require.Contains(t, err.Error(), "insert failed")

	queries := mc.getQueries()
	insertCount := 0
	for _, q := range queries {
		if strings.HasPrefix(q.query, "INSERT INTO") {
			insertCount++
		}
	}
	require.Equal(t, 1, insertCount, "non-transient error should not be retried")
}

func TestTagFieldFiltering(t *testing.T) {
	t.Run("filter tags", func(t *testing.T) {
		s := &SnowpipeStreaming{
			Account:         "a",
			User:            "u",
			Database:        "d",
			Schema:          "s",
			Table:           "t",
			BatchSize:       1000,
			TimestampColumn: "timestamp",
			TagColumns:      []string{"host"},
		}
		require.NoError(t, s.Init())

		m := testMetric("cpu", map[string]string{"host": "h1", "region": "us"}, map[string]interface{}{"val": 1.0}, time.Now())
		columns := s.buildColumnOrder(m)

		require.Contains(t, columns, "host")
		require.NotContains(t, columns, "region")
	})

	t.Run("filter fields", func(t *testing.T) {
		s := &SnowpipeStreaming{
			Account:         "a",
			User:            "u",
			Database:        "d",
			Schema:          "s",
			Table:           "t",
			BatchSize:       1000,
			TimestampColumn: "timestamp",
			FieldColumns:    []string{"usage_idle"},
		}
		require.NoError(t, s.Init())

		m := testMetric("cpu", nil, map[string]interface{}{
			"usage_idle":   95.5,
			"usage_system": 4.5,
		}, time.Now())
		columns := s.buildColumnOrder(m)

		require.Contains(t, columns, "usage_idle")
		require.NotContains(t, columns, "usage_system")
	})

	t.Run("no filter includes all", func(t *testing.T) {
		s := newTestPlugin(t)

		m := testMetric("cpu", map[string]string{"host": "h1", "region": "us"}, map[string]interface{}{
			"usage_idle":   95.5,
			"usage_system": 4.5,
		}, time.Now())
		columns := s.buildColumnOrder(m)

		require.Contains(t, columns, "host")
		require.Contains(t, columns, "region")
		require.Contains(t, columns, "usage_idle")
		require.Contains(t, columns, "usage_system")
	})
}

// ---------------------------------------------------------------------------
// Integration-style tests (with mocked Snowflake)
// ---------------------------------------------------------------------------

func TestConnectAndWriteIntegration(t *testing.T) {
	s := newTestPlugin(t)
	connectTestPlugin(t, s)

	ts := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	metrics := []telegraf.Metric{
		testMetric("cpu", map[string]string{"host": "srv1"}, map[string]interface{}{
			"usage_idle": float64(95.5),
		}, ts),
		testMetric("cpu", map[string]string{"host": "srv2"}, map[string]interface{}{
			"usage_idle": float64(80.0),
		}, ts),
	}

	require.NoError(t, s.Write(metrics))

	mc := getGlobalMock()
	queries := mc.getQueries()

	// Should have exactly one INSERT
	var insertQ executedQuery
	found := false
	for _, q := range queries {
		if strings.HasPrefix(q.query, "INSERT INTO") {
			insertQ = q
			found = true
			break
		}
	}
	require.True(t, found, "expected an INSERT query")

	// Verify fully qualified table
	require.Contains(t, insertQ.query, `"TEST_DB"."PUBLIC"."METRICS"`)

	// Verify columns
	require.Contains(t, insertQ.query, `"timestamp"`)
	require.Contains(t, insertQ.query, `"name"`)
	require.Contains(t, insertQ.query, `"host"`)
	require.Contains(t, insertQ.query, `"usage_idle"`)

	// 2 rows, each with 4 columns = 8 args
	require.Len(t, insertQ.args, 8)
}

func TestCreateTable(t *testing.T) {
	s := newTestPlugin(t)
	s.CreateTable = true
	connectTestPlugin(t, s)

	ts := time.Now()
	m := testMetric("new_metric", map[string]string{"host": "h1"}, map[string]interface{}{
		"val":    int64(42),
		"active": true,
	}, ts)

	require.NoError(t, s.Write([]telegraf.Metric{m}))

	mc := getGlobalMock()
	queries := mc.getQueries()

	// Find CREATE TABLE query
	var createQ string
	for _, q := range queries {
		if strings.Contains(q.query, "CREATE TABLE IF NOT EXISTS") {
			createQ = q.query
			break
		}
	}
	require.NotEmpty(t, createQ, "expected a CREATE TABLE query")
	require.Contains(t, createQ, `"TEST_DB"."PUBLIC"."METRICS"`)
	require.Contains(t, createQ, `"timestamp" TIMESTAMP_NTZ`)
	require.Contains(t, createQ, `"name" VARCHAR`)
	require.Contains(t, createQ, `"host" VARCHAR`)
	require.Contains(t, createQ, `"val" NUMBER`)
	require.Contains(t, createQ, `"active" BOOLEAN`)
}

func TestSchemaEvolution(t *testing.T) {
	s := newTestPlugin(t)
	s.CreateTable = true
	connectTestPlugin(t, s)

	mc := getGlobalMock()

	ts := time.Now()
	m1 := testMetric("evolve", nil, map[string]interface{}{"val": int64(1)}, ts)
	require.NoError(t, s.Write([]telegraf.Metric{m1}))

	// Now write a metric with an extra field — should trigger ALTER TABLE
	m2 := testMetric("evolve", nil, map[string]interface{}{
		"val":     int64(2),
		"new_col": "hello",
	}, ts)
	require.NoError(t, s.Write([]telegraf.Metric{m2}))

	queries := mc.getQueries()
	var alterFound bool
	for _, q := range queries {
		if strings.Contains(q.query, "ALTER TABLE") && strings.Contains(q.query, `"new_col"`) {
			alterFound = true
			break
		}
	}
	require.True(t, alterFound, "expected ALTER TABLE to add new_col")
}

func TestBatchRetry(t *testing.T) {
	s := newTestPlugin(t)
	s.BatchSize = 2
	s.RetryMax = 2
	s.RetryDelay = config.Duration(1 * time.Millisecond)

	mc := resetGlobalMock()
	mc.execErr = errors.New("i/o timeout")
	atomic.StoreInt32(&mc.execErrCount, 1) // fail first attempt, succeed second

	s.openDB = func() (*sql.DB, error) {
		return sql.Open("snowflake_mock", "mock")
	}
	require.NoError(t, s.Connect())
	t.Cleanup(func() { s.Close() })

	ts := time.Now()
	metrics := []telegraf.Metric{
		testMetric("cpu", nil, map[string]interface{}{"val": 1.0}, ts),
		testMetric("cpu", nil, map[string]interface{}{"val": 2.0}, ts),
	}

	require.NoError(t, s.Write(metrics))
}

func TestConcurrentWrites(t *testing.T) {
	s := newTestPlugin(t)
	s.Table = "metrics_{{.Name}}"
	require.NoError(t, s.Init())
	connectTestPlugin(t, s)

	ts := time.Now()
	var wg sync.WaitGroup

	for i := range 10 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			name := fmt.Sprintf("metric_%d", i)
			m := testMetric(name, nil, map[string]interface{}{"val": float64(i)}, ts)
			assert.NoError(t, s.Write([]telegraf.Metric{m}))
		}(i)
	}

	wg.Wait()

	mc := getGlobalMock()
	queries := mc.getQueries()
	insertCount := 0
	for _, q := range queries {
		if strings.HasPrefix(q.query, "INSERT INTO") {
			insertCount++
		}
	}
	require.Equal(t, 10, insertCount, "expected 10 inserts from 10 concurrent writers")
}

func TestBuildInsertQuery(t *testing.T) {
	s := newTestPlugin(t)

	query := s.buildInsertQuery("my_table", []string{"timestamp", "name", "host", "value"}, 2)
	require.Contains(t, query, `"TEST_DB"."PUBLIC"."my_table"`)
	require.Contains(t, query, `"timestamp", "name", "host", "value"`)
	require.Contains(t, query, "(?, ?, ?, ?), (?, ?, ?, ?)")
}

func TestQuoteIdent(t *testing.T) {
	require.Equal(t, `"simple"`, quoteIdent("simple"))
	require.Equal(t, `"has""quote"`, quoteIdent(`has"quote`))
	require.Equal(t, `"has space"`, quoteIdent("has space"))
}

func TestGroupByTable(t *testing.T) {
	s := &SnowpipeStreaming{
		Account:  "a",
		User:     "u",
		Database: "d",
		Schema:   "s",
		Table:    "metrics_{{.Name}}",
	}
	require.NoError(t, s.Init())

	ts := time.Now()
	metrics := []telegraf.Metric{
		testMetric("cpu", nil, map[string]interface{}{"val": 1.0}, ts),
		testMetric("mem", nil, map[string]interface{}{"val": 2.0}, ts),
		testMetric("cpu", nil, map[string]interface{}{"val": 3.0}, ts),
	}

	groups := s.groupByTable(metrics)
	require.Len(t, groups, 2)
	require.Len(t, groups["metrics_cpu"], 2)
	require.Len(t, groups["metrics_mem"], 1)
}

func TestGoTypeToSnowflake(t *testing.T) {
	require.Equal(t, "NUMBER", goTypeToSnowflake(int64(1)))
	require.Equal(t, "NUMBER", goTypeToSnowflake(uint64(1)))
	require.Equal(t, "DOUBLE", goTypeToSnowflake(float64(1.0)))
	require.Equal(t, "BOOLEAN", goTypeToSnowflake(true))
	require.Equal(t, "VARCHAR", goTypeToSnowflake("text"))
}

func TestIsTransientError(t *testing.T) {
	require.False(t, isTransientError(nil))
	require.True(t, isTransientError(errors.New("connection refused")))
	require.True(t, isTransientError(errors.New("i/o timeout")))
	require.True(t, isTransientError(errors.New("service unavailable")))
	require.True(t, isTransientError(errors.New("connection reset by peer")))
	require.False(t, isTransientError(errors.New("sql compilation error")))
}

func TestSanitizeFieldValue(t *testing.T) {
	require.InDelta(t, float64(1.0), sanitizeFieldValue(float64(1.0)), 1e-9)
	require.Equal(t, "hello", sanitizeFieldValue("hello"))
	require.Nil(t, sanitizeFieldValue(math.NaN()))
	require.Nil(t, sanitizeFieldValue(math.Inf(1)))
	require.Equal(t, int64(42), sanitizeFieldValue(int64(42)))
}

func TestSampleConfig(t *testing.T) {
	s := &SnowpipeStreaming{}
	conf := s.SampleConfig()
	require.NotEmpty(t, conf)
	require.Contains(t, conf, "snowpipe_streaming")
	require.Contains(t, conf, "account")
}
