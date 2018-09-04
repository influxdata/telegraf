package stdlib_test

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgmock"
	"github.com/jackc/pgx/pgproto3"
	"github.com/jackc/pgx/stdlib"
)

func openDB(t *testing.T) *sql.DB {
	db, err := sql.Open("pgx", "postgres://pgx_md5:secret@127.0.0.1:5432/pgx_test")
	if err != nil {
		t.Fatalf("sql.Open failed: %v", err)
	}

	return db
}

func closeDB(t *testing.T, db *sql.DB) {
	err := db.Close()
	if err != nil {
		t.Fatalf("db.Close unexpectedly failed: %v", err)
	}
}

// Do a simple query to ensure the connection is still usable
func ensureConnValid(t *testing.T, db *sql.DB) {
	var sum, rowCount int32

	rows, err := db.Query("select generate_series(1,$1)", 10)
	if err != nil {
		t.Fatalf("db.Query failed: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var n int32
		rows.Scan(&n)
		sum += n
		rowCount++
	}

	if rows.Err() != nil {
		t.Fatalf("db.Query failed: %v", err)
	}

	if rowCount != 10 {
		t.Error("Select called onDataRow wrong number of times")
	}
	if sum != 55 {
		t.Error("Wrong values returned")
	}
}

type preparer interface {
	Prepare(query string) (*sql.Stmt, error)
}

func prepareStmt(t *testing.T, p preparer, sql string) *sql.Stmt {
	stmt, err := p.Prepare(sql)
	if err != nil {
		t.Fatalf("%v Prepare unexpectedly failed: %v", p, err)
	}

	return stmt
}

func closeStmt(t *testing.T, stmt *sql.Stmt) {
	err := stmt.Close()
	if err != nil {
		t.Fatalf("stmt.Close unexpectedly failed: %v", err)
	}
}

func TestNormalLifeCycle(t *testing.T) {
	db := openDB(t)
	defer closeDB(t, db)

	stmt := prepareStmt(t, db, "select 'foo', n from generate_series($1::int, $2::int) n")
	defer closeStmt(t, stmt)

	rows, err := stmt.Query(int32(1), int32(10))
	if err != nil {
		t.Fatalf("stmt.Query unexpectedly failed: %v", err)
	}

	rowCount := int64(0)

	for rows.Next() {
		rowCount++

		var s string
		var n int64
		if err := rows.Scan(&s, &n); err != nil {
			t.Fatalf("rows.Scan unexpectedly failed: %v", err)
		}
		if s != "foo" {
			t.Errorf(`Expected "foo", received "%v"`, s)
		}
		if n != rowCount {
			t.Errorf("Expected %d, received %d", rowCount, n)
		}
	}
	err = rows.Err()
	if err != nil {
		t.Fatalf("rows.Err unexpectedly is: %v", err)
	}
	if rowCount != 10 {
		t.Fatalf("Expected to receive 10 rows, instead received %d", rowCount)
	}

	err = rows.Close()
	if err != nil {
		t.Fatalf("rows.Close unexpectedly failed: %v", err)
	}

	ensureConnValid(t, db)
}

func TestOpenWithDriverConfigAfterConnect(t *testing.T) {
	driverConfig := stdlib.DriverConfig{
		AfterConnect: func(c *pgx.Conn) error {
			_, err := c.Exec("create temporary sequence pgx")
			return err
		},
	}

	stdlib.RegisterDriverConfig(&driverConfig)
	defer stdlib.UnregisterDriverConfig(&driverConfig)

	db, err := sql.Open("pgx", driverConfig.ConnectionString("postgres://pgx_md5:secret@127.0.0.1:5432/pgx_test"))
	if err != nil {
		t.Fatalf("sql.Open failed: %v", err)
	}
	defer closeDB(t, db)

	var n int64
	err = db.QueryRow("select nextval('pgx')").Scan(&n)
	if err != nil {
		t.Fatalf("db.QueryRow unexpectedly failed: %v", err)
	}
	if n != 1 {
		t.Fatalf("n => %d, want %d", n, 1)
	}
}

func TestStmtExec(t *testing.T) {
	db := openDB(t)
	defer closeDB(t, db)

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("db.Begin unexpectedly failed: %v", err)
	}

	createStmt := prepareStmt(t, tx, "create temporary table t(a varchar not null)")
	_, err = createStmt.Exec()
	if err != nil {
		t.Fatalf("stmt.Exec unexpectedly failed: %v", err)
	}
	closeStmt(t, createStmt)

	insertStmt := prepareStmt(t, tx, "insert into t values($1::text)")
	result, err := insertStmt.Exec("foo")
	if err != nil {
		t.Fatalf("stmt.Exec unexpectedly failed: %v", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		t.Fatalf("result.RowsAffected unexpectedly failed: %v", err)
	}
	if n != 1 {
		t.Fatalf("Expected 1, received %d", n)
	}
	closeStmt(t, insertStmt)

	if err != nil {
		t.Fatalf("tx.Commit unexpectedly failed: %v", err)
	}

	ensureConnValid(t, db)
}

func TestQueryCloseRowsEarly(t *testing.T) {
	db := openDB(t)
	defer closeDB(t, db)

	stmt := prepareStmt(t, db, "select 'foo', n from generate_series($1::int, $2::int) n")
	defer closeStmt(t, stmt)

	rows, err := stmt.Query(int32(1), int32(10))
	if err != nil {
		t.Fatalf("stmt.Query unexpectedly failed: %v", err)
	}

	// Close rows immediately without having read them
	err = rows.Close()
	if err != nil {
		t.Fatalf("rows.Close unexpectedly failed: %v", err)
	}

	// Run the query again to ensure the connection and statement are still ok
	rows, err = stmt.Query(int32(1), int32(10))
	if err != nil {
		t.Fatalf("stmt.Query unexpectedly failed: %v", err)
	}

	rowCount := int64(0)

	for rows.Next() {
		rowCount++

		var s string
		var n int64
		if err := rows.Scan(&s, &n); err != nil {
			t.Fatalf("rows.Scan unexpectedly failed: %v", err)
		}
		if s != "foo" {
			t.Errorf(`Expected "foo", received "%v"`, s)
		}
		if n != rowCount {
			t.Errorf("Expected %d, received %d", rowCount, n)
		}
	}
	err = rows.Err()
	if err != nil {
		t.Fatalf("rows.Err unexpectedly is: %v", err)
	}
	if rowCount != 10 {
		t.Fatalf("Expected to receive 10 rows, instead received %d", rowCount)
	}

	err = rows.Close()
	if err != nil {
		t.Fatalf("rows.Close unexpectedly failed: %v", err)
	}

	ensureConnValid(t, db)
}

func TestConnExec(t *testing.T) {
	db := openDB(t)
	defer closeDB(t, db)

	_, err := db.Exec("create temporary table t(a varchar not null)")
	if err != nil {
		t.Fatalf("db.Exec unexpectedly failed: %v", err)
	}

	result, err := db.Exec("insert into t values('hey')")
	if err != nil {
		t.Fatalf("db.Exec unexpectedly failed: %v", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		t.Fatalf("result.RowsAffected unexpectedly failed: %v", err)
	}
	if n != 1 {
		t.Fatalf("Expected 1, received %d", n)
	}

	ensureConnValid(t, db)
}

func TestConnQuery(t *testing.T) {
	db := openDB(t)
	defer closeDB(t, db)

	rows, err := db.Query("select 'foo', n from generate_series($1::int, $2::int) n", int32(1), int32(10))
	if err != nil {
		t.Fatalf("db.Query unexpectedly failed: %v", err)
	}

	rowCount := int64(0)

	for rows.Next() {
		rowCount++

		var s string
		var n int64
		if err := rows.Scan(&s, &n); err != nil {
			t.Fatalf("rows.Scan unexpectedly failed: %v", err)
		}
		if s != "foo" {
			t.Errorf(`Expected "foo", received "%v"`, s)
		}
		if n != rowCount {
			t.Errorf("Expected %d, received %d", rowCount, n)
		}
	}
	err = rows.Err()
	if err != nil {
		t.Fatalf("rows.Err unexpectedly is: %v", err)
	}
	if rowCount != 10 {
		t.Fatalf("Expected to receive 10 rows, instead received %d", rowCount)
	}

	err = rows.Close()
	if err != nil {
		t.Fatalf("rows.Close unexpectedly failed: %v", err)
	}

	ensureConnValid(t, db)
}

type testLog struct {
	lvl  pgx.LogLevel
	msg  string
	data map[string]interface{}
}

type testLogger struct {
	logs []testLog
}

func (l *testLogger) Log(lvl pgx.LogLevel, msg string, data map[string]interface{}) {
	l.logs = append(l.logs, testLog{lvl: lvl, msg: msg, data: data})
}

func TestConnQueryLog(t *testing.T) {
	logger := &testLogger{}

	driverConfig := stdlib.DriverConfig{
		ConnConfig: pgx.ConnConfig{
			Host:     "127.0.0.1",
			User:     "pgx_md5",
			Password: "secret",
			Database: "pgx_test",
			Logger:   logger,
		},
	}

	stdlib.RegisterDriverConfig(&driverConfig)
	defer stdlib.UnregisterDriverConfig(&driverConfig)

	db, err := sql.Open("pgx", driverConfig.ConnectionString(""))
	if err != nil {
		t.Fatalf("sql.Open failed: %v", err)
	}
	defer closeDB(t, db)

	var n int64
	err = db.QueryRow("select 1").Scan(&n)
	if err != nil {
		t.Fatalf("db.QueryRow unexpectedly failed: %v", err)
	}

	l := logger.logs[len(logger.logs)-1]
	if l.msg != "Query" {
		t.Errorf("Expected to log Query, but got %v", l)
	}

	if l.data["sql"] != "select 1" {
		t.Errorf("Expected to log Query with sql 'select 1', but got %v", l)
	}
}

func TestConnQueryNull(t *testing.T) {
	db := openDB(t)
	defer closeDB(t, db)

	rows, err := db.Query("select $1::int", nil)
	if err != nil {
		t.Fatalf("db.Query unexpectedly failed: %v", err)
	}

	rowCount := int64(0)

	for rows.Next() {
		rowCount++

		var n sql.NullInt64
		if err := rows.Scan(&n); err != nil {
			t.Fatalf("rows.Scan unexpectedly failed: %v", err)
		}
		if n.Valid != false {
			t.Errorf("Expected n to be null, but it was %v", n)
		}
	}
	err = rows.Err()
	if err != nil {
		t.Fatalf("rows.Err unexpectedly is: %v", err)
	}
	if rowCount != 1 {
		t.Fatalf("Expected to receive 11 rows, instead received %d", rowCount)
	}

	err = rows.Close()
	if err != nil {
		t.Fatalf("rows.Close unexpectedly failed: %v", err)
	}

	ensureConnValid(t, db)
}

func TestConnQueryRowByteSlice(t *testing.T) {
	db := openDB(t)
	defer closeDB(t, db)

	expected := []byte{222, 173, 190, 239}
	var actual []byte

	err := db.QueryRow(`select E'\\xdeadbeef'::bytea`).Scan(&actual)
	if err != nil {
		t.Fatalf("db.QueryRow unexpectedly failed: %v", err)
	}

	if bytes.Compare(actual, expected) != 0 {
		t.Fatalf("Expected %v, but got %v", expected, actual)
	}

	ensureConnValid(t, db)
}

func TestConnQueryFailure(t *testing.T) {
	db := openDB(t)
	defer closeDB(t, db)

	_, err := db.Query("select 'foo")
	if _, ok := err.(pgx.PgError); !ok {
		t.Fatalf("Expected db.Query to return pgx.PgError, but instead received: %v", err)
	}

	ensureConnValid(t, db)
}

// Test type that pgx would handle natively in binary, but since it is not a
// database/sql native type should be passed through as a string
func TestConnQueryRowPgxBinary(t *testing.T) {
	db := openDB(t)
	defer closeDB(t, db)

	sql := "select $1::int4[]"
	expected := "{1,2,3}"
	var actual string

	err := db.QueryRow(sql, expected).Scan(&actual)
	if err != nil {
		t.Errorf("Unexpected failure: %v (sql -> %v)", err, sql)
	}

	if actual != expected {
		t.Errorf(`Expected "%v", got "%v" (sql -> %v)`, expected, actual, sql)
	}

	ensureConnValid(t, db)
}

func TestConnQueryRowUnknownType(t *testing.T) {
	db := openDB(t)
	defer closeDB(t, db)

	sql := "select $1::point"
	expected := "(1,2)"
	var actual string

	err := db.QueryRow(sql, expected).Scan(&actual)
	if err != nil {
		t.Errorf("Unexpected failure: %v (sql -> %v)", err, sql)
	}

	if actual != expected {
		t.Errorf(`Expected "%v", got "%v" (sql -> %v)`, expected, actual, sql)
	}

	ensureConnValid(t, db)
}

func TestConnQueryJSONIntoByteSlice(t *testing.T) {
	db := openDB(t)
	defer closeDB(t, db)

	_, err := db.Exec(`
		create temporary table docs(
			body json not null
		);

		insert into docs(body) values('{"foo":"bar"}');
`)
	if err != nil {
		t.Fatalf("db.Exec unexpectedly failed: %v", err)
	}

	sql := `select * from docs`
	expected := []byte(`{"foo":"bar"}`)
	var actual []byte

	err = db.QueryRow(sql).Scan(&actual)
	if err != nil {
		t.Errorf("Unexpected failure: %v (sql -> %v)", err, sql)
	}

	if bytes.Compare(actual, expected) != 0 {
		t.Errorf(`Expected "%v", got "%v" (sql -> %v)`, string(expected), string(actual), sql)
	}

	_, err = db.Exec(`drop table docs`)
	if err != nil {
		t.Fatalf("db.Exec unexpectedly failed: %v", err)
	}

	ensureConnValid(t, db)
}

func TestConnExecInsertByteSliceIntoJSON(t *testing.T) {
	db := openDB(t)
	defer closeDB(t, db)

	_, err := db.Exec(`
		create temporary table docs(
			body json not null
		);
`)
	if err != nil {
		t.Fatalf("db.Exec unexpectedly failed: %v", err)
	}

	expected := []byte(`{"foo":"bar"}`)

	_, err = db.Exec(`insert into docs(body) values($1)`, expected)
	if err != nil {
		t.Fatalf("db.Exec unexpectedly failed: %v", err)
	}

	var actual []byte
	err = db.QueryRow(`select body from docs`).Scan(&actual)
	if err != nil {
		t.Fatalf("db.QueryRow unexpectedly failed: %v", err)
	}

	if bytes.Compare(actual, expected) != 0 {
		t.Errorf(`Expected "%v", got "%v"`, string(expected), string(actual))
	}

	_, err = db.Exec(`drop table docs`)
	if err != nil {
		t.Fatalf("db.Exec unexpectedly failed: %v", err)
	}

	ensureConnValid(t, db)
}

func TestTransactionLifeCycle(t *testing.T) {
	db := openDB(t)
	defer closeDB(t, db)

	_, err := db.Exec("create temporary table t(a varchar not null)")
	if err != nil {
		t.Fatalf("db.Exec unexpectedly failed: %v", err)
	}

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("db.Begin unexpectedly failed: %v", err)
	}

	_, err = tx.Exec("insert into t values('hi')")
	if err != nil {
		t.Fatalf("tx.Exec unexpectedly failed: %v", err)
	}

	err = tx.Rollback()
	if err != nil {
		t.Fatalf("tx.Rollback unexpectedly failed: %v", err)
	}

	var n int64
	err = db.QueryRow("select count(*) from t").Scan(&n)
	if err != nil {
		t.Fatalf("db.QueryRow.Scan unexpectedly failed: %v", err)
	}
	if n != 0 {
		t.Fatalf("Expected 0 rows due to rollback, instead found %d", n)
	}

	tx, err = db.Begin()
	if err != nil {
		t.Fatalf("db.Begin unexpectedly failed: %v", err)
	}

	_, err = tx.Exec("insert into t values('hi')")
	if err != nil {
		t.Fatalf("tx.Exec unexpectedly failed: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		t.Fatalf("tx.Commit unexpectedly failed: %v", err)
	}

	err = db.QueryRow("select count(*) from t").Scan(&n)
	if err != nil {
		t.Fatalf("db.QueryRow.Scan unexpectedly failed: %v", err)
	}
	if n != 1 {
		t.Fatalf("Expected 1 rows due to rollback, instead found %d", n)
	}

	ensureConnValid(t, db)
}

func TestConnBeginTxIsolation(t *testing.T) {
	db := openDB(t)
	defer closeDB(t, db)

	var defaultIsoLevel string
	err := db.QueryRow("show transaction_isolation").Scan(&defaultIsoLevel)
	if err != nil {
		t.Fatalf("QueryRow failed: %v", err)
	}

	supportedTests := []struct {
		sqlIso sql.IsolationLevel
		pgIso  string
	}{
		{sqlIso: sql.LevelDefault, pgIso: defaultIsoLevel},
		{sqlIso: sql.LevelReadUncommitted, pgIso: "read uncommitted"},
		{sqlIso: sql.LevelReadCommitted, pgIso: "read committed"},
		{sqlIso: sql.LevelSnapshot, pgIso: "repeatable read"},
		{sqlIso: sql.LevelSerializable, pgIso: "serializable"},
	}
	for i, tt := range supportedTests {
		func() {
			tx, err := db.BeginTx(context.Background(), &sql.TxOptions{Isolation: tt.sqlIso})
			if err != nil {
				t.Errorf("%d. BeginTx failed: %v", i, err)
				return
			}
			defer tx.Rollback()

			var pgIso string
			err = tx.QueryRow("show transaction_isolation").Scan(&pgIso)
			if err != nil {
				t.Errorf("%d. QueryRow failed: %v", i, err)
			}

			if pgIso != tt.pgIso {
				t.Errorf("%d. pgIso => %s, want %s", i, pgIso, tt.pgIso)
			}
		}()
	}

	unsupportedTests := []struct {
		sqlIso sql.IsolationLevel
	}{
		{sqlIso: sql.LevelWriteCommitted},
		{sqlIso: sql.LevelLinearizable},
	}
	for i, tt := range unsupportedTests {
		tx, err := db.BeginTx(context.Background(), &sql.TxOptions{Isolation: tt.sqlIso})
		if err == nil {
			t.Errorf("%d. BeginTx should have failed", i)
			tx.Rollback()
		}
	}

	ensureConnValid(t, db)
}

func TestConnBeginTxReadOnly(t *testing.T) {
	db := openDB(t)
	defer closeDB(t, db)

	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{ReadOnly: true})
	if err != nil {
		t.Fatalf("BeginTx failed: %v", err)
	}
	defer tx.Rollback()

	var pgReadOnly string
	err = tx.QueryRow("show transaction_read_only").Scan(&pgReadOnly)
	if err != nil {
		t.Errorf("QueryRow failed: %v", err)
	}

	if pgReadOnly != "on" {
		t.Errorf("pgReadOnly => %s, want %s", pgReadOnly, "on")
	}

	ensureConnValid(t, db)
}

func TestBeginTxContextCancel(t *testing.T) {
	db := openDB(t)
	defer closeDB(t, db)

	_, err := db.Exec("drop table if exists t")
	if err != nil {
		t.Fatalf("db.Exec failed: %v", err)
	}

	ctx, cancelFn := context.WithCancel(context.Background())

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx failed: %v", err)
	}

	_, err = tx.Exec("create table t(id serial)")
	if err != nil {
		t.Fatalf("tx.Exec failed: %v", err)
	}

	cancelFn()

	err = tx.Commit()
	if err != context.Canceled && err != sql.ErrTxDone {
		t.Fatalf("err => %v, want %v or %v", err, context.Canceled, sql.ErrTxDone)
	}

	var n int
	err = db.QueryRow("select count(*) from t").Scan(&n)
	if pgErr, ok := err.(pgx.PgError); !ok || pgErr.Code != "42P01" {
		t.Fatalf(`err => %v, want PgError{Code: "42P01"}`, err)
	}

	ensureConnValid(t, db)
}

func acceptStandardPgxConn(backend *pgproto3.Backend) error {
	script := pgmock.Script{
		Steps: pgmock.AcceptUnauthenticatedConnRequestSteps(),
	}

	err := script.Run(backend)
	if err != nil {
		return err
	}

	typeScript := pgmock.Script{
		Steps: pgmock.PgxInitSteps(),
	}

	return typeScript.Run(backend)
}

func TestBeginTxContextCancelWithDeadConn(t *testing.T) {
	script := &pgmock.Script{
		Steps: pgmock.AcceptUnauthenticatedConnRequestSteps(),
	}
	script.Steps = append(script.Steps, pgmock.PgxInitSteps()...)
	script.Steps = append(script.Steps,
		pgmock.ExpectMessage(&pgproto3.Query{String: "begin"}),
		pgmock.SendMessage(&pgproto3.CommandComplete{CommandTag: "BEGIN"}),
		pgmock.SendMessage(&pgproto3.ReadyForQuery{TxStatus: 'T'}),
	)

	server, err := pgmock.NewServer(script)
	if err != nil {
		t.Fatal(err)
	}

	errChan := make(chan error)
	go func() {
		errChan <- server.ServeOne()
	}()

	db, err := sql.Open("pgx", fmt.Sprintf("postgres://pgx_md5:secret@%s/pgx_test?sslmode=disable", server.Addr()))
	if err != nil {
		t.Fatalf("sql.Open failed: %v", err)
	}
	defer closeDB(t, db)

	ctx, cancelFn := context.WithCancel(context.Background())

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx failed: %v", err)
	}

	cancelFn()

	err = tx.Commit()
	if err != context.Canceled && err != sql.ErrTxDone {
		t.Fatalf("err => %v, want %v or %v", err, context.Canceled, sql.ErrTxDone)
	}

	if err := <-errChan; err != nil {
		t.Fatalf("mock server err: %v", err)
	}
}

func TestAcquireConn(t *testing.T) {
	db := openDB(t)
	defer closeDB(t, db)

	var conns []*pgx.Conn

	for i := 1; i < 6; i++ {
		conn, err := stdlib.AcquireConn(db)
		if err != nil {
			t.Errorf("%d. AcquireConn failed: %v", i, err)
			continue
		}

		var n int32
		err = conn.QueryRow("select 1").Scan(&n)
		if err != nil {
			t.Errorf("%d. QueryRow failed: %v", i, err)
		}
		if n != 1 {
			t.Errorf("%d. n => %d, want %d", i, n, 1)
		}

		stats := db.Stats()
		if stats.OpenConnections != i {
			t.Errorf("%d. stats.OpenConnections => %d, want %d", i, stats.OpenConnections, i)
		}

		conns = append(conns, conn)
	}

	for i, conn := range conns {
		if err := stdlib.ReleaseConn(db, conn); err != nil {
			t.Errorf("%d. stdlib.ReleaseConn failed: %v", i, err)
		}
	}

	ensureConnValid(t, db)
}

func TestConnPingContextSuccess(t *testing.T) {
	db := openDB(t)
	defer closeDB(t, db)

	if err := db.PingContext(context.Background()); err != nil {
		t.Fatalf("db.PingContext failed: %v", err)
	}

	ensureConnValid(t, db)
}

func TestConnPingContextCancel(t *testing.T) {
	script := &pgmock.Script{
		Steps: pgmock.AcceptUnauthenticatedConnRequestSteps(),
	}
	script.Steps = append(script.Steps, pgmock.PgxInitSteps()...)
	script.Steps = append(script.Steps,
		pgmock.ExpectMessage(&pgproto3.Query{String: ";"}),
		pgmock.WaitForClose(),
	)

	server, err := pgmock.NewServer(script)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()

	errChan := make(chan error, 1)
	go func() {
		errChan <- server.ServeOne()
	}()

	db, err := sql.Open("pgx", fmt.Sprintf("postgres://pgx_md5:secret@%s/pgx_test?sslmode=disable", server.Addr()))
	if err != nil {
		t.Fatalf("sql.Open failed: %v", err)
	}
	defer closeDB(t, db)

	ctx, _ := context.WithTimeout(context.Background(), 100*time.Millisecond)

	err = db.PingContext(ctx)
	if err != context.DeadlineExceeded {
		t.Errorf("err => %v, want %v", err, context.DeadlineExceeded)
	}

	if err := <-errChan; err != nil {
		t.Errorf("mock server err: %v", err)
	}
}

func TestConnPrepareContextSuccess(t *testing.T) {
	db := openDB(t)
	defer closeDB(t, db)

	stmt, err := db.PrepareContext(context.Background(), "select now()")
	if err != nil {
		t.Fatalf("db.PrepareContext failed: %v", err)
	}
	stmt.Close()

	ensureConnValid(t, db)
}

func TestConnPrepareContextCancel(t *testing.T) {
	script := &pgmock.Script{
		Steps: pgmock.AcceptUnauthenticatedConnRequestSteps(),
	}
	script.Steps = append(script.Steps, pgmock.PgxInitSteps()...)
	script.Steps = append(script.Steps,
		pgmock.ExpectMessage(&pgproto3.Parse{Name: "pgx_0", Query: "select now()"}),
		pgmock.ExpectMessage(&pgproto3.Describe{ObjectType: 'S', Name: "pgx_0"}),
		pgmock.ExpectMessage(&pgproto3.Sync{}),
		pgmock.WaitForClose(),
	)

	server, err := pgmock.NewServer(script)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()

	errChan := make(chan error)
	go func() {
		errChan <- server.ServeOne()
	}()

	db, err := sql.Open("pgx", fmt.Sprintf("postgres://pgx_md5:secret@%s/pgx_test?sslmode=disable", server.Addr()))
	if err != nil {
		t.Fatalf("sql.Open failed: %v", err)
	}
	defer closeDB(t, db)

	ctx, _ := context.WithTimeout(context.Background(), 100*time.Millisecond)

	_, err = db.PrepareContext(ctx, "select now()")
	if err != context.DeadlineExceeded {
		t.Errorf("err => %v, want %v", err, context.DeadlineExceeded)
	}

	if err := <-errChan; err != nil {
		t.Errorf("mock server err: %v", err)
	}
}

func TestConnExecContextSuccess(t *testing.T) {
	db := openDB(t)
	defer closeDB(t, db)

	_, err := db.ExecContext(context.Background(), "create temporary table exec_context_test(id serial primary key)")
	if err != nil {
		t.Fatalf("db.ExecContext failed: %v", err)
	}

	ensureConnValid(t, db)
}

func TestConnExecContextCancel(t *testing.T) {
	script := &pgmock.Script{
		Steps: pgmock.AcceptUnauthenticatedConnRequestSteps(),
	}
	script.Steps = append(script.Steps, pgmock.PgxInitSteps()...)
	script.Steps = append(script.Steps,
		pgmock.ExpectMessage(&pgproto3.Query{String: "create temporary table exec_context_test(id serial primary key)"}),
		pgmock.WaitForClose(),
	)

	server, err := pgmock.NewServer(script)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()

	errChan := make(chan error)
	go func() {
		errChan <- server.ServeOne()
	}()

	db, err := sql.Open("pgx", fmt.Sprintf("postgres://pgx_md5:secret@%s/pgx_test?sslmode=disable", server.Addr()))
	if err != nil {
		t.Fatalf("sql.Open failed: %v", err)
	}
	defer closeDB(t, db)

	ctx, _ := context.WithTimeout(context.Background(), 100*time.Millisecond)

	_, err = db.ExecContext(ctx, "create temporary table exec_context_test(id serial primary key)")
	if err != context.DeadlineExceeded {
		t.Errorf("err => %v, want %v", err, context.DeadlineExceeded)
	}

	if err := <-errChan; err != nil {
		t.Errorf("mock server err: %v", err)
	}
}

func TestConnQueryContextSuccess(t *testing.T) {
	db := openDB(t)
	defer closeDB(t, db)

	rows, err := db.QueryContext(context.Background(), "select * from generate_series(1,10) n")
	if err != nil {
		t.Fatalf("db.QueryContext failed: %v", err)
	}

	for rows.Next() {
		var n int64
		if err := rows.Scan(&n); err != nil {
			t.Error(err)
		}
	}

	if rows.Err() != nil {
		t.Error(rows.Err())
	}

	ensureConnValid(t, db)
}

func TestConnQueryContextCancel(t *testing.T) {
	script := &pgmock.Script{
		Steps: pgmock.AcceptUnauthenticatedConnRequestSteps(),
	}
	script.Steps = append(script.Steps, pgmock.PgxInitSteps()...)
	script.Steps = append(script.Steps,
		pgmock.ExpectMessage(&pgproto3.Parse{Query: "select * from generate_series(1,10) n"}),
		pgmock.ExpectMessage(&pgproto3.Describe{ObjectType: 'S'}),
		pgmock.ExpectMessage(&pgproto3.Sync{}),

		pgmock.SendMessage(&pgproto3.ParseComplete{}),
		pgmock.SendMessage(&pgproto3.ParameterDescription{}),
		pgmock.SendMessage(&pgproto3.RowDescription{
			Fields: []pgproto3.FieldDescription{
				{
					Name:         "n",
					DataTypeOID:  23,
					DataTypeSize: 4,
					TypeModifier: 4294967295,
				},
			},
		}),
		pgmock.SendMessage(&pgproto3.ReadyForQuery{TxStatus: 'I'}),

		pgmock.ExpectMessage(&pgproto3.Bind{ResultFormatCodes: []int16{1}}),
		pgmock.ExpectMessage(&pgproto3.Execute{}),
		pgmock.ExpectMessage(&pgproto3.Sync{}),

		pgmock.SendMessage(&pgproto3.BindComplete{}),
		pgmock.WaitForClose(),
	)

	server, err := pgmock.NewServer(script)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()

	errChan := make(chan error)
	go func() {
		errChan <- server.ServeOne()
	}()

	db, err := sql.Open("pgx", fmt.Sprintf("postgres://pgx_md5:secret@%s/pgx_test?sslmode=disable", server.Addr()))
	if err != nil {
		t.Fatalf("sql.Open failed: %v", err)
	}
	defer db.Close()

	ctx, cancelFn := context.WithCancel(context.Background())

	rows, err := db.QueryContext(ctx, "select * from generate_series(1,10) n")
	if err != nil {
		t.Fatalf("db.QueryContext failed: %v", err)
	}

	cancelFn()

	for rows.Next() {
		t.Fatalf("no rows should ever be received")
	}

	if rows.Err() != context.Canceled {
		t.Errorf("rows.Err() => %v, want %v", rows.Err(), context.Canceled)
	}

	if err := <-errChan; err != nil {
		t.Errorf("mock server err: %v", err)
	}
}

func TestRowsColumnTypeDatabaseTypeName(t *testing.T) {
	db := openDB(t)
	defer closeDB(t, db)

	rows, err := db.Query("select * from generate_series(1,10) n")
	if err != nil {
		t.Fatalf("db.Query failed: %v", err)
	}

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		t.Fatalf("rows.ColumnTypes failed: %v", err)
	}

	if len(columnTypes) != 1 {
		t.Fatalf("len(columnTypes) => %v, want %v", len(columnTypes), 1)
	}

	if columnTypes[0].DatabaseTypeName() != "INT4" {
		t.Errorf("columnTypes[0].DatabaseTypeName() => %v, want %v", columnTypes[0].DatabaseTypeName(), "INT4")
	}

	rows.Close()

	ensureConnValid(t, db)
}

func TestStmtExecContextSuccess(t *testing.T) {
	db := openDB(t)
	defer closeDB(t, db)

	_, err := db.Exec("create temporary table t(id int primary key)")
	if err != nil {
		t.Fatalf("db.Exec failed: %v", err)
	}

	stmt, err := db.Prepare("insert into t(id) values ($1::int4)")
	if err != nil {
		t.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(context.Background(), 42)
	if err != nil {
		t.Fatal(err)
	}

	ensureConnValid(t, db)
}

func TestStmtExecContextCancel(t *testing.T) {
	db := openDB(t)
	defer closeDB(t, db)

	_, err := db.Exec("create temporary table t(id int primary key)")
	if err != nil {
		t.Fatalf("db.Exec failed: %v", err)
	}

	stmt, err := db.Prepare("insert into t(id) select $1::int4 from pg_sleep(5)")
	if err != nil {
		t.Fatal(err)
	}
	defer stmt.Close()

	ctx, _ := context.WithTimeout(context.Background(), 100*time.Millisecond)

	_, err = stmt.ExecContext(ctx, 42)
	if err != context.DeadlineExceeded {
		t.Errorf("err => %v, want %v", err, context.DeadlineExceeded)
	}

	ensureConnValid(t, db)
}

func TestStmtQueryContextSuccess(t *testing.T) {
	db := openDB(t)
	defer closeDB(t, db)

	stmt, err := db.Prepare("select * from generate_series(1,$1::int4) n")
	if err != nil {
		t.Fatal(err)
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(context.Background(), 5)
	if err != nil {
		t.Fatalf("stmt.QueryContext failed: %v", err)
	}

	for rows.Next() {
		var n int64
		if err := rows.Scan(&n); err != nil {
			t.Error(err)
		}
	}

	if rows.Err() != nil {
		t.Error(rows.Err())
	}

	ensureConnValid(t, db)
}

func TestStmtQueryContextCancel(t *testing.T) {
	script := &pgmock.Script{
		Steps: pgmock.AcceptUnauthenticatedConnRequestSteps(),
	}
	script.Steps = append(script.Steps, pgmock.PgxInitSteps()...)
	script.Steps = append(script.Steps,
		pgmock.ExpectMessage(&pgproto3.Parse{Name: "pgx_0", Query: "select * from generate_series(1, $1::int4) n"}),
		pgmock.ExpectMessage(&pgproto3.Describe{ObjectType: 'S', Name: "pgx_0"}),
		pgmock.ExpectMessage(&pgproto3.Sync{}),

		pgmock.SendMessage(&pgproto3.ParseComplete{}),
		pgmock.SendMessage(&pgproto3.ParameterDescription{ParameterOIDs: []uint32{23}}),
		pgmock.SendMessage(&pgproto3.RowDescription{
			Fields: []pgproto3.FieldDescription{
				{
					Name:         "n",
					DataTypeOID:  23,
					DataTypeSize: 4,
					TypeModifier: 4294967295,
				},
			},
		}),
		pgmock.SendMessage(&pgproto3.ReadyForQuery{TxStatus: 'I'}),

		pgmock.ExpectMessage(&pgproto3.Bind{PreparedStatement: "pgx_0", ParameterFormatCodes: []int16{1}, Parameters: [][]uint8{{0x0, 0x0, 0x0, 0x2a}}, ResultFormatCodes: []int16{1}}),
		pgmock.ExpectMessage(&pgproto3.Execute{}),
		pgmock.ExpectMessage(&pgproto3.Sync{}),

		pgmock.SendMessage(&pgproto3.BindComplete{}),
		pgmock.WaitForClose(),
	)

	server, err := pgmock.NewServer(script)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()

	errChan := make(chan error)
	go func() {
		errChan <- server.ServeOne()
	}()

	db, err := sql.Open("pgx", fmt.Sprintf("postgres://pgx_md5:secret@%s/pgx_test?sslmode=disable", server.Addr()))
	if err != nil {
		t.Fatalf("sql.Open failed: %v", err)
	}
	// defer closeDB(t, db) // mock DB doesn't close correctly yet

	stmt, err := db.Prepare("select * from generate_series(1, $1::int4) n")
	if err != nil {
		t.Fatal(err)
	}
	// defer stmt.Close()

	ctx, cancelFn := context.WithCancel(context.Background())

	rows, err := stmt.QueryContext(ctx, 42)
	if err != nil {
		t.Fatalf("stmt.QueryContext failed: %v", err)
	}

	cancelFn()

	for rows.Next() {
		t.Fatalf("no rows should ever be received")
	}

	if rows.Err() != context.Canceled {
		t.Errorf("rows.Err() => %v, want %v", rows.Err(), context.Canceled)
	}

	if err := <-errChan; err != nil {
		t.Errorf("mock server err: %v", err)
	}
}

func TestRowsColumnTypes(t *testing.T) {
	columnTypesTests := []struct {
		Name     string
		TypeName string
		Length   struct {
			Len int64
			OK  bool
		}
		DecimalSize struct {
			Precision int64
			Scale     int64
			OK        bool
		}
		ScanType reflect.Type
	}{
		{
			Name:     "a",
			TypeName: "INT4",
			Length: struct {
				Len int64
				OK  bool
			}{
				Len: 0,
				OK:  false,
			},
			DecimalSize: struct {
				Precision int64
				Scale     int64
				OK        bool
			}{
				Precision: 0,
				Scale:     0,
				OK:        false,
			},
			ScanType: reflect.TypeOf(int32(0)),
		}, {
			Name:     "bar",
			TypeName: "TEXT",
			Length: struct {
				Len int64
				OK  bool
			}{
				Len: math.MaxInt64,
				OK:  true,
			},
			DecimalSize: struct {
				Precision int64
				Scale     int64
				OK        bool
			}{
				Precision: 0,
				Scale:     0,
				OK:        false,
			},
			ScanType: reflect.TypeOf(""),
		}, {
			Name:     "dec",
			TypeName: "NUMERIC",
			Length: struct {
				Len int64
				OK  bool
			}{
				Len: 0,
				OK:  false,
			},
			DecimalSize: struct {
				Precision int64
				Scale     int64
				OK        bool
			}{
				Precision: 9,
				Scale:     2,
				OK:        true,
			},
			ScanType: reflect.TypeOf(float64(0)),
		},
	}

	db := openDB(t)
	defer closeDB(t, db)

	rows, err := db.Query("SELECT 1 AS a, text 'bar' AS bar, 1.28::numeric(9, 2) AS dec")
	if err != nil {
		t.Fatal(err)
	}

	columns, err := rows.ColumnTypes()
	if err != nil {
		t.Fatal(err)
	}
	if len(columns) != 3 {
		t.Errorf("expected 3 columns found %d", len(columns))
	}

	for i, tt := range columnTypesTests {
		c := columns[i]
		if c.Name() != tt.Name {
			t.Errorf("(%d) got: %s, want: %s", i, c.Name(), tt.Name)
		}
		if c.DatabaseTypeName() != tt.TypeName {
			t.Errorf("(%d) got: %s, want: %s", i, c.DatabaseTypeName(), tt.TypeName)
		}
		l, ok := c.Length()
		if l != tt.Length.Len {
			t.Errorf("(%d) got: %d, want: %d", i, l, tt.Length.Len)
		}
		if ok != tt.Length.OK {
			t.Errorf("(%d) got: %t, want: %t", i, ok, tt.Length.OK)
		}
		p, s, ok := c.DecimalSize()
		if p != tt.DecimalSize.Precision {
			t.Errorf("(%d) got: %d, want: %d", i, p, tt.DecimalSize.Precision)
		}
		if s != tt.DecimalSize.Scale {
			t.Errorf("(%d) got: %d, want: %d", i, s, tt.DecimalSize.Scale)
		}
		if ok != tt.DecimalSize.OK {
			t.Errorf("(%d) got: %t, want: %t", i, ok, tt.DecimalSize.OK)
		}
		if c.ScanType() != tt.ScanType {
			t.Errorf("(%d) got: %v, want: %v", i, c.ScanType(), tt.ScanType)
		}
	}
}

func TestSimpleQueryLifeCycle(t *testing.T) {
	driverConfig := stdlib.DriverConfig{
		ConnConfig: pgx.ConnConfig{PreferSimpleProtocol: true},
	}

	stdlib.RegisterDriverConfig(&driverConfig)
	defer stdlib.UnregisterDriverConfig(&driverConfig)

	db, err := sql.Open("pgx", driverConfig.ConnectionString("postgres://pgx_md5:secret@127.0.0.1:5432/pgx_test"))
	if err != nil {
		t.Fatalf("sql.Open failed: %v", err)
	}
	defer closeDB(t, db)

	rows, err := db.Query("SELECT 'foo', n FROM generate_series($1::int, $2::int) n WHERE 3 = $3", 1, 10, 3)
	if err != nil {
		t.Fatalf("stmt.Query unexpectedly failed: %v", err)
	}

	rowCount := int64(0)

	for rows.Next() {
		rowCount++
		var (
			s string
			n int64
		)

		if err := rows.Scan(&s, &n); err != nil {
			t.Fatalf("rows.Scan unexpectedly failed: %v", err)
		}

		if s != "foo" {
			t.Errorf(`Expected "foo", received "%v"`, s)
		}

		if n != rowCount {
			t.Errorf("Expected %d, received %d", rowCount, n)
		}
	}

	if err = rows.Err(); err != nil {
		t.Fatalf("rows.Err unexpectedly is: %v", err)
	}

	if rowCount != 10 {
		t.Fatalf("Expected to receive 10 rows, instead received %d", rowCount)
	}

	err = rows.Close()
	if err != nil {
		t.Fatalf("rows.Close unexpectedly failed: %v", err)
	}

	rows, err = db.Query("select 1 where false")
	if err != nil {
		t.Fatalf("stmt.Query unexpectedly failed: %v", err)
	}

	rowCount = int64(0)

	for rows.Next() {
		rowCount++
	}

	if err = rows.Err(); err != nil {
		t.Fatalf("rows.Err unexpectedly is: %v", err)
	}

	if rowCount != 0 {
		t.Fatalf("Expected to receive 10 rows, instead received %d", rowCount)
	}

	err = rows.Close()
	if err != nil {
		t.Fatalf("rows.Close unexpectedly failed: %v", err)
	}

	ensureConnValid(t, db)
}
