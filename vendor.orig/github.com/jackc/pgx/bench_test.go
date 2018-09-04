package pgx_test

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
)

func BenchmarkConnPool(b *testing.B) {
	config := pgx.ConnPoolConfig{ConnConfig: *defaultConnConfig, MaxConnections: 5}
	pool, err := pgx.NewConnPool(config)
	if err != nil {
		b.Fatalf("Unable to create connection pool: %v", err)
	}
	defer pool.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var conn *pgx.Conn
		if conn, err = pool.Acquire(); err != nil {
			b.Fatalf("Unable to acquire connection: %v", err)
		}
		pool.Release(conn)
	}
}

func BenchmarkConnPoolQueryRow(b *testing.B) {
	config := pgx.ConnPoolConfig{ConnConfig: *defaultConnConfig, MaxConnections: 5}
	pool, err := pgx.NewConnPool(config)
	if err != nil {
		b.Fatalf("Unable to create connection pool: %v", err)
	}
	defer pool.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		num := float64(-1)
		if err := pool.QueryRow("select random()").Scan(&num); err != nil {
			b.Fatal(err)
		}

		if num < 0 {
			b.Fatalf("expected `select random()` to return between 0 and 1 but it was: %v", num)
		}
	}
}

func BenchmarkPointerPointerWithNullValues(b *testing.B) {
	conn := mustConnect(b, *defaultConnConfig)
	defer closeConn(b, conn)

	_, err := conn.Prepare("selectNulls", "select 1::int4, 'johnsmith', null::text, null::text, null::text, null::date, null::timestamptz")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var record struct {
			id            int32
			userName      string
			email         *string
			name          *string
			sex           *string
			birthDate     *time.Time
			lastLoginTime *time.Time
		}

		err = conn.QueryRow("selectNulls").Scan(
			&record.id,
			&record.userName,
			&record.email,
			&record.name,
			&record.sex,
			&record.birthDate,
			&record.lastLoginTime,
		)
		if err != nil {
			b.Fatal(err)
		}

		// These checks both ensure that the correct data was returned
		// and provide a benchmark of accessing the returned values.
		if record.id != 1 {
			b.Fatalf("bad value for id: %v", record.id)
		}
		if record.userName != "johnsmith" {
			b.Fatalf("bad value for userName: %v", record.userName)
		}
		if record.email != nil {
			b.Fatalf("bad value for email: %v", record.email)
		}
		if record.name != nil {
			b.Fatalf("bad value for name: %v", record.name)
		}
		if record.sex != nil {
			b.Fatalf("bad value for sex: %v", record.sex)
		}
		if record.birthDate != nil {
			b.Fatalf("bad value for birthDate: %v", record.birthDate)
		}
		if record.lastLoginTime != nil {
			b.Fatalf("bad value for lastLoginTime: %v", record.lastLoginTime)
		}
	}
}

func BenchmarkPointerPointerWithPresentValues(b *testing.B) {
	conn := mustConnect(b, *defaultConnConfig)
	defer closeConn(b, conn)

	_, err := conn.Prepare("selectNulls", "select 1::int4, 'johnsmith', 'johnsmith@example.com', 'John Smith', 'male', '1970-01-01'::date, '2015-01-01 00:00:00'::timestamptz")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var record struct {
			id            int32
			userName      string
			email         *string
			name          *string
			sex           *string
			birthDate     *time.Time
			lastLoginTime *time.Time
		}

		err = conn.QueryRow("selectNulls").Scan(
			&record.id,
			&record.userName,
			&record.email,
			&record.name,
			&record.sex,
			&record.birthDate,
			&record.lastLoginTime,
		)
		if err != nil {
			b.Fatal(err)
		}

		// These checks both ensure that the correct data was returned
		// and provide a benchmark of accessing the returned values.
		if record.id != 1 {
			b.Fatalf("bad value for id: %v", record.id)
		}
		if record.userName != "johnsmith" {
			b.Fatalf("bad value for userName: %v", record.userName)
		}
		if record.email == nil || *record.email != "johnsmith@example.com" {
			b.Fatalf("bad value for email: %v", record.email)
		}
		if record.name == nil || *record.name != "John Smith" {
			b.Fatalf("bad value for name: %v", record.name)
		}
		if record.sex == nil || *record.sex != "male" {
			b.Fatalf("bad value for sex: %v", record.sex)
		}
		if record.birthDate == nil || *record.birthDate != time.Date(1970, 1, 1, 0, 0, 0, 0, time.Local) {
			b.Fatalf("bad value for birthDate: %v", record.birthDate)
		}
		if record.lastLoginTime == nil || *record.lastLoginTime != time.Date(2015, 1, 1, 0, 0, 0, 0, time.Local) {
			b.Fatalf("bad value for lastLoginTime: %v", record.lastLoginTime)
		}
	}
}

func BenchmarkSelectWithoutLogging(b *testing.B) {
	conn := mustConnect(b, *defaultConnConfig)
	defer closeConn(b, conn)

	benchmarkSelectWithLog(b, conn)
}

type discardLogger struct{}

func (dl discardLogger) Log(level pgx.LogLevel, msg string, data map[string]interface{}) {}

func BenchmarkSelectWithLoggingTraceDiscard(b *testing.B) {
	conn := mustConnect(b, *defaultConnConfig)
	defer closeConn(b, conn)

	var logger discardLogger
	conn.SetLogger(logger)
	conn.SetLogLevel(pgx.LogLevelTrace)

	benchmarkSelectWithLog(b, conn)
}

func BenchmarkSelectWithLoggingDebugWithDiscard(b *testing.B) {
	conn := mustConnect(b, *defaultConnConfig)
	defer closeConn(b, conn)

	var logger discardLogger
	conn.SetLogger(logger)
	conn.SetLogLevel(pgx.LogLevelDebug)

	benchmarkSelectWithLog(b, conn)
}

func BenchmarkSelectWithLoggingInfoWithDiscard(b *testing.B) {
	conn := mustConnect(b, *defaultConnConfig)
	defer closeConn(b, conn)

	var logger discardLogger
	conn.SetLogger(logger)
	conn.SetLogLevel(pgx.LogLevelInfo)

	benchmarkSelectWithLog(b, conn)
}

func BenchmarkSelectWithLoggingErrorWithDiscard(b *testing.B) {
	conn := mustConnect(b, *defaultConnConfig)
	defer closeConn(b, conn)

	var logger discardLogger
	conn.SetLogger(logger)
	conn.SetLogLevel(pgx.LogLevelError)

	benchmarkSelectWithLog(b, conn)
}

func benchmarkSelectWithLog(b *testing.B, conn *pgx.Conn) {
	_, err := conn.Prepare("test", "select 1::int4, 'johnsmith', 'johnsmith@example.com', 'John Smith', 'male', '1970-01-01'::date, '2015-01-01 00:00:00'::timestamptz")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var record struct {
			id            int32
			userName      string
			email         string
			name          string
			sex           string
			birthDate     time.Time
			lastLoginTime time.Time
		}

		err = conn.QueryRow("test").Scan(
			&record.id,
			&record.userName,
			&record.email,
			&record.name,
			&record.sex,
			&record.birthDate,
			&record.lastLoginTime,
		)
		if err != nil {
			b.Fatal(err)
		}

		// These checks both ensure that the correct data was returned
		// and provide a benchmark of accessing the returned values.
		if record.id != 1 {
			b.Fatalf("bad value for id: %v", record.id)
		}
		if record.userName != "johnsmith" {
			b.Fatalf("bad value for userName: %v", record.userName)
		}
		if record.email != "johnsmith@example.com" {
			b.Fatalf("bad value for email: %v", record.email)
		}
		if record.name != "John Smith" {
			b.Fatalf("bad value for name: %v", record.name)
		}
		if record.sex != "male" {
			b.Fatalf("bad value for sex: %v", record.sex)
		}
		if record.birthDate != time.Date(1970, 1, 1, 0, 0, 0, 0, time.Local) {
			b.Fatalf("bad value for birthDate: %v", record.birthDate)
		}
		if record.lastLoginTime != time.Date(2015, 1, 1, 0, 0, 0, 0, time.Local) {
			b.Fatalf("bad value for lastLoginTime: %v", record.lastLoginTime)
		}
	}
}

const benchmarkWriteTableCreateSQL = `drop table if exists t;

create table t(
	varchar_1 varchar not null,
	varchar_2 varchar not null,
	varchar_null_1 varchar,
	date_1 date not null,
	date_null_1 date,
	int4_1 int4 not null,
	int4_2 int4 not null,
	int4_null_1 int4,
	tstz_1 timestamptz not null,
	tstz_2 timestamptz,
	bool_1 bool not null,
	bool_2 bool not null,
	bool_3 bool not null
);
`

const benchmarkWriteTableInsertSQL = `insert into t(
	varchar_1,
	varchar_2,
	varchar_null_1,
	date_1,
	date_null_1,
	int4_1,
	int4_2,
	int4_null_1,
	tstz_1,
	tstz_2,
	bool_1,
	bool_2,
	bool_3
) values (
	$1::varchar,
	$2::varchar,
	$3::varchar,
	$4::date,
	$5::date,
	$6::int4,
	$7::int4,
	$8::int4,
	$9::timestamptz,
	$10::timestamptz,
	$11::bool,
	$12::bool,
	$13::bool
)`

type benchmarkWriteTableCopyFromSrc struct {
	count int
	idx   int
	row   []interface{}
}

func (s *benchmarkWriteTableCopyFromSrc) Next() bool {
	s.idx++
	return s.idx < s.count
}

func (s *benchmarkWriteTableCopyFromSrc) Values() ([]interface{}, error) {
	return s.row, nil
}

func (s *benchmarkWriteTableCopyFromSrc) Err() error {
	return nil
}

func newBenchmarkWriteTableCopyFromSrc(count int) pgx.CopyFromSource {
	return &benchmarkWriteTableCopyFromSrc{
		count: count,
		row: []interface{}{
			"varchar_1",
			"varchar_2",
			pgtype.Text{},
			time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local),
			pgtype.Date{},
			1,
			2,
			pgtype.Int4{},
			time.Date(2001, 1, 1, 0, 0, 0, 0, time.Local),
			time.Date(2002, 1, 1, 0, 0, 0, 0, time.Local),
			true,
			false,
			true,
		},
	}
}

func benchmarkWriteNRowsViaInsert(b *testing.B, n int) {
	conn := mustConnect(b, *defaultConnConfig)
	defer closeConn(b, conn)

	mustExec(b, conn, benchmarkWriteTableCreateSQL)
	_, err := conn.Prepare("insert_t", benchmarkWriteTableInsertSQL)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		src := newBenchmarkWriteTableCopyFromSrc(n)

		tx, err := conn.Begin()
		if err != nil {
			b.Fatal(err)
		}

		for src.Next() {
			values, _ := src.Values()
			if _, err = tx.Exec("insert_t", values...); err != nil {
				b.Fatalf("Exec unexpectedly failed with: %v", err)
			}
		}

		err = tx.Commit()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// note this function is only used for benchmarks -- it doesn't escape tableName
// or columnNames
func multiInsert(conn *pgx.Conn, tableName string, columnNames []string, rowSrc pgx.CopyFromSource) (int, error) {
	maxRowsPerInsert := 65535 / len(columnNames)
	rowsThisInsert := 0
	rowCount := 0

	sqlBuf := &bytes.Buffer{}
	args := make(pgx.QueryArgs, 0)

	resetQuery := func() {
		sqlBuf.Reset()
		fmt.Fprintf(sqlBuf, "insert into %s(%s) values", tableName, strings.Join(columnNames, ", "))

		args = args[0:0]

		rowsThisInsert = 0
	}
	resetQuery()

	tx, err := conn.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	for rowSrc.Next() {
		if rowsThisInsert > 0 {
			sqlBuf.WriteByte(',')
		}

		sqlBuf.WriteByte('(')

		values, err := rowSrc.Values()
		if err != nil {
			return 0, err
		}

		for i, val := range values {
			if i > 0 {
				sqlBuf.WriteByte(',')
			}
			sqlBuf.WriteString(args.Append(val))
		}

		sqlBuf.WriteByte(')')

		rowsThisInsert++

		if rowsThisInsert == maxRowsPerInsert {
			_, err := tx.Exec(sqlBuf.String(), args...)
			if err != nil {
				return 0, err
			}

			rowCount += rowsThisInsert
			resetQuery()
		}
	}

	if rowsThisInsert > 0 {
		_, err := tx.Exec(sqlBuf.String(), args...)
		if err != nil {
			return 0, err
		}

		rowCount += rowsThisInsert
	}

	if err := tx.Commit(); err != nil {
		return 0, nil
	}

	return rowCount, nil

}

func benchmarkWriteNRowsViaMultiInsert(b *testing.B, n int) {
	conn := mustConnect(b, *defaultConnConfig)
	defer closeConn(b, conn)

	mustExec(b, conn, benchmarkWriteTableCreateSQL)
	_, err := conn.Prepare("insert_t", benchmarkWriteTableInsertSQL)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		src := newBenchmarkWriteTableCopyFromSrc(n)

		_, err := multiInsert(conn, "t",
			[]string{"varchar_1",
				"varchar_2",
				"varchar_null_1",
				"date_1",
				"date_null_1",
				"int4_1",
				"int4_2",
				"int4_null_1",
				"tstz_1",
				"tstz_2",
				"bool_1",
				"bool_2",
				"bool_3"},
			src)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func benchmarkWriteNRowsViaCopy(b *testing.B, n int) {
	conn := mustConnect(b, *defaultConnConfig)
	defer closeConn(b, conn)

	mustExec(b, conn, benchmarkWriteTableCreateSQL)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		src := newBenchmarkWriteTableCopyFromSrc(n)

		_, err := conn.CopyFrom(pgx.Identifier{"t"},
			[]string{"varchar_1",
				"varchar_2",
				"varchar_null_1",
				"date_1",
				"date_null_1",
				"int4_1",
				"int4_2",
				"int4_null_1",
				"tstz_1",
				"tstz_2",
				"bool_1",
				"bool_2",
				"bool_3"},
			src)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWrite5RowsViaInsert(b *testing.B) {
	benchmarkWriteNRowsViaInsert(b, 5)
}

func BenchmarkWrite5RowsViaMultiInsert(b *testing.B) {
	benchmarkWriteNRowsViaMultiInsert(b, 5)
}

func BenchmarkWrite5RowsViaCopy(b *testing.B) {
	benchmarkWriteNRowsViaCopy(b, 5)
}

func BenchmarkWrite10RowsViaInsert(b *testing.B) {
	benchmarkWriteNRowsViaInsert(b, 10)
}

func BenchmarkWrite10RowsViaMultiInsert(b *testing.B) {
	benchmarkWriteNRowsViaMultiInsert(b, 10)
}

func BenchmarkWrite10RowsViaCopy(b *testing.B) {
	benchmarkWriteNRowsViaCopy(b, 10)
}

func BenchmarkWrite100RowsViaInsert(b *testing.B) {
	benchmarkWriteNRowsViaInsert(b, 100)
}

func BenchmarkWrite100RowsViaMultiInsert(b *testing.B) {
	benchmarkWriteNRowsViaMultiInsert(b, 100)
}

func BenchmarkWrite100RowsViaCopy(b *testing.B) {
	benchmarkWriteNRowsViaCopy(b, 100)
}

func BenchmarkWrite1000RowsViaInsert(b *testing.B) {
	benchmarkWriteNRowsViaInsert(b, 1000)
}

func BenchmarkWrite1000RowsViaMultiInsert(b *testing.B) {
	benchmarkWriteNRowsViaMultiInsert(b, 1000)
}

func BenchmarkWrite1000RowsViaCopy(b *testing.B) {
	benchmarkWriteNRowsViaCopy(b, 1000)
}

func BenchmarkWrite10000RowsViaInsert(b *testing.B) {
	benchmarkWriteNRowsViaInsert(b, 10000)
}

func BenchmarkWrite10000RowsViaMultiInsert(b *testing.B) {
	benchmarkWriteNRowsViaMultiInsert(b, 10000)
}

func BenchmarkWrite10000RowsViaCopy(b *testing.B) {
	benchmarkWriteNRowsViaCopy(b, 10000)
}

func BenchmarkMultipleQueriesNonBatch(b *testing.B) {
	config := pgx.ConnPoolConfig{ConnConfig: *defaultConnConfig, MaxConnections: 5}
	pool, err := pgx.NewConnPool(config)
	if err != nil {
		b.Fatalf("Unable to create connection pool: %v", err)
	}
	defer pool.Close()

	queryCount := 3

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < queryCount; j++ {
			rows, err := pool.Query("select n from generate_series(0, 5) n")
			if err != nil {
				b.Fatal(err)
			}

			for k := 0; rows.Next(); k++ {
				var n int
				if err := rows.Scan(&n); err != nil {
					b.Fatal(err)
				}
				if n != k {
					b.Fatalf("n => %v, want %v", n, k)
				}
			}

			if rows.Err() != nil {
				b.Fatal(rows.Err())
			}
		}
	}
}

func BenchmarkMultipleQueriesBatch(b *testing.B) {
	config := pgx.ConnPoolConfig{ConnConfig: *defaultConnConfig, MaxConnections: 5}
	pool, err := pgx.NewConnPool(config)
	if err != nil {
		b.Fatalf("Unable to create connection pool: %v", err)
	}
	defer pool.Close()

	queryCount := 3

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		batch := pool.BeginBatch()
		for j := 0; j < queryCount; j++ {
			batch.Queue("select n from generate_series(0,5) n",
				nil,
				nil,
				[]int16{pgx.BinaryFormatCode},
			)
		}

		err := batch.Send(context.Background(), nil)
		if err != nil {
			b.Fatal(err)
		}

		for j := 0; j < queryCount; j++ {
			rows, err := batch.QueryResults()
			if err != nil {
				b.Fatal(err)
			}

			for k := 0; rows.Next(); k++ {
				var n int
				if err := rows.Scan(&n); err != nil {
					b.Fatal(err)
				}
				if n != k {
					b.Fatalf("n => %v, want %v", n, k)
				}
			}

			if rows.Err() != nil {
				b.Fatal(rows.Err())
			}
		}

		err = batch.Close()
		if err != nil {
			b.Fatal(err)
		}
	}
}
