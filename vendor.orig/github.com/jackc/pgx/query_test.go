package pgx_test

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/cockroachdb/apd"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	satori "github.com/jackc/pgx/pgtype/ext/satori-uuid"
	"github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
)

func TestConnQueryScan(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	var sum, rowCount int32

	rows, err := conn.Query("select generate_series(1,$1)", 10)
	if err != nil {
		t.Fatalf("conn.Query failed: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var n int32
		rows.Scan(&n)
		sum += n
		rowCount++
	}

	if rows.Err() != nil {
		t.Fatalf("conn.Query failed: %v", err)
	}

	if rowCount != 10 {
		t.Error("Select called onDataRow wrong number of times")
	}
	if sum != 55 {
		t.Error("Wrong values returned")
	}
}

func TestConnQueryScanWithManyColumns(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	columnCount := 1000
	sql := "select "
	for i := 0; i < columnCount; i++ {
		if i > 0 {
			sql += ","
		}
		sql += fmt.Sprintf(" %d", i)
	}
	sql += " from generate_series(1,5)"

	dest := make([]int, columnCount)

	var rowCount int

	rows, err := conn.Query(sql)
	if err != nil {
		t.Fatalf("conn.Query failed: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		destPtrs := make([]interface{}, columnCount)
		for i := range destPtrs {
			destPtrs[i] = &dest[i]
		}
		if err := rows.Scan(destPtrs...); err != nil {
			t.Fatalf("rows.Scan failed: %v", err)
		}
		rowCount++

		for i := range dest {
			if dest[i] != i {
				t.Errorf("dest[%d] => %d, want %d", i, dest[i], i)
			}
		}
	}

	if rows.Err() != nil {
		t.Fatalf("conn.Query failed: %v", err)
	}

	if rowCount != 5 {
		t.Errorf("rowCount => %d, want %d", rowCount, 5)
	}
}

func TestConnQueryValues(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	var rowCount int32

	rows, err := conn.Query("select 'foo'::text, 'bar'::varchar, n, null, n from generate_series(1,$1) n", 10)
	if err != nil {
		t.Fatalf("conn.Query failed: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		rowCount++

		values, err := rows.Values()
		if err != nil {
			t.Fatalf("rows.Values failed: %v", err)
		}
		if len(values) != 5 {
			t.Errorf("Expected rows.Values to return 5 values, but it returned %d", len(values))
		}
		if values[0] != "foo" {
			t.Errorf(`Expected values[0] to be "foo", but it was %v`, values[0])
		}
		if values[1] != "bar" {
			t.Errorf(`Expected values[1] to be "bar", but it was %v`, values[1])
		}

		if values[2] != rowCount {
			t.Errorf(`Expected values[2] to be %d, but it was %d`, rowCount, values[2])
		}

		if values[3] != nil {
			t.Errorf(`Expected values[3] to be %v, but it was %d`, nil, values[3])
		}

		if values[4] != rowCount {
			t.Errorf(`Expected values[4] to be %d, but it was %d`, rowCount, values[4])
		}
	}

	if rows.Err() != nil {
		t.Fatalf("conn.Query failed: %v", err)
	}

	if rowCount != 10 {
		t.Error("Select called onDataRow wrong number of times")
	}
}

// https://github.com/jackc/pgx/issues/228
func TestRowsScanDoesNotAllowScanningBinaryFormatValuesIntoString(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	var s string

	err := conn.QueryRow("select 1").Scan(&s)
	if err == nil || !(strings.Contains(err.Error(), "cannot decode binary value into string") || strings.Contains(err.Error(), "cannot assign")) {
		t.Fatalf("Expected Scan to fail to encode binary value into string but: %v", err)
	}

	ensureConnValid(t, conn)
}

// Test that a connection stays valid when query results are closed early
func TestConnQueryCloseEarly(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	// Immediately close query without reading any rows
	rows, err := conn.Query("select generate_series(1,$1)", 10)
	if err != nil {
		t.Fatalf("conn.Query failed: %v", err)
	}
	rows.Close()

	ensureConnValid(t, conn)

	// Read partial response then close
	rows, err = conn.Query("select generate_series(1,$1)", 10)
	if err != nil {
		t.Fatalf("conn.Query failed: %v", err)
	}

	ok := rows.Next()
	if !ok {
		t.Fatal("rows.Next terminated early")
	}

	var n int32
	rows.Scan(&n)
	if n != 1 {
		t.Fatalf("Expected 1 from first row, but got %v", n)
	}

	rows.Close()

	ensureConnValid(t, conn)
}

func TestConnQueryCloseEarlyWithErrorOnWire(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	rows, err := conn.Query("select 1/(10-n) from generate_series(1,10) n")
	if err != nil {
		t.Fatalf("conn.Query failed: %v", err)
	}
	rows.Close()

	ensureConnValid(t, conn)
}

// Test that a connection stays valid when query results read incorrectly
func TestConnQueryReadWrongTypeError(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	// Read a single value incorrectly
	rows, err := conn.Query("select generate_series(1,$1)", 10)
	if err != nil {
		t.Fatalf("conn.Query failed: %v", err)
	}

	rowsRead := 0

	for rows.Next() {
		var t time.Time
		rows.Scan(&t)
		rowsRead++
	}

	if rowsRead != 1 {
		t.Fatalf("Expected error to cause only 1 row to be read, but %d were read", rowsRead)
	}

	if rows.Err() == nil {
		t.Fatal("Expected Rows to have an error after an improper read but it didn't")
	}

	if rows.Err().Error() != "can't scan into dest[0]: Can't convert OID 23 to time.Time" && !strings.Contains(rows.Err().Error(), "cannot assign") {
		t.Fatalf("Expected different Rows.Err(): %v", rows.Err())
	}

	ensureConnValid(t, conn)
}

// Test that a connection stays valid when query results read incorrectly
func TestConnQueryReadTooManyValues(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	// Read too many values
	rows, err := conn.Query("select generate_series(1,$1)", 10)
	if err != nil {
		t.Fatalf("conn.Query failed: %v", err)
	}

	rowsRead := 0

	for rows.Next() {
		var n, m int32
		rows.Scan(&n, &m)
		rowsRead++
	}

	if rowsRead != 1 {
		t.Fatalf("Expected error to cause only 1 row to be read, but %d were read", rowsRead)
	}

	if rows.Err() == nil {
		t.Fatal("Expected Rows to have an error after an improper read but it didn't")
	}

	ensureConnValid(t, conn)
}

func TestConnQueryScanIgnoreColumn(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	rows, err := conn.Query("select 1::int8, 2::int8, 3::int8")
	if err != nil {
		t.Fatalf("conn.Query failed: %v", err)
	}

	ok := rows.Next()
	if !ok {
		t.Fatal("rows.Next terminated early")
	}

	var n, m int64
	err = rows.Scan(&n, nil, &m)
	if err != nil {
		t.Fatalf("rows.Scan failed: %v", err)
	}
	rows.Close()

	if n != 1 {
		t.Errorf("Expected n to equal 1, but it was %d", n)
	}

	if m != 3 {
		t.Errorf("Expected n to equal 3, but it was %d", m)
	}

	ensureConnValid(t, conn)
}

func TestConnQueryErrorWhileReturningRows(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	for i := 0; i < 100; i++ {
		func() {
			sql := `select 42 / (random() * 20)::integer from generate_series(1,100000)`

			rows, err := conn.Query(sql)
			if err != nil {
				t.Fatal(err)
			}
			defer rows.Close()

			for rows.Next() {
				var n int32
				rows.Scan(&n)
			}

			if err, ok := rows.Err().(pgx.PgError); !ok {
				t.Fatalf("Expected pgx.PgError, got %v", err)
			}

			ensureConnValid(t, conn)
		}()
	}

}

func TestQueryEncodeError(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	rows, err := conn.Query("select $1::integer", "wrong")
	if err != nil {
		t.Errorf("conn.Query failure: %v", err)
	}
	defer rows.Close()

	rows.Next()

	if rows.Err() == nil {
		t.Error("Expected rows.Err() to return error, but it didn't")
	}
	if rows.Err().Error() != `ERROR: invalid input syntax for integer: "wrong" (SQLSTATE 22P02)` {
		t.Error("Expected rows.Err() to return different error:", rows.Err())
	}
}

func TestQueryRowCoreTypes(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	type allTypes struct {
		s   string
		f32 float32
		f64 float64
		b   bool
		t   time.Time
		oid pgtype.OID
	}

	var actual, zero allTypes

	tests := []struct {
		sql       string
		queryArgs []interface{}
		scanArgs  []interface{}
		expected  allTypes
	}{
		{"select $1::text", []interface{}{"Jack"}, []interface{}{&actual.s}, allTypes{s: "Jack"}},
		{"select $1::float4", []interface{}{float32(1.23)}, []interface{}{&actual.f32}, allTypes{f32: 1.23}},
		{"select $1::float8", []interface{}{float64(1.23)}, []interface{}{&actual.f64}, allTypes{f64: 1.23}},
		{"select $1::bool", []interface{}{true}, []interface{}{&actual.b}, allTypes{b: true}},
		{"select $1::timestamptz", []interface{}{time.Unix(123, 5000)}, []interface{}{&actual.t}, allTypes{t: time.Unix(123, 5000)}},
		{"select $1::timestamp", []interface{}{time.Date(2010, 1, 2, 3, 4, 5, 0, time.UTC)}, []interface{}{&actual.t}, allTypes{t: time.Date(2010, 1, 2, 3, 4, 5, 0, time.UTC)}},
		{"select $1::date", []interface{}{time.Date(1987, 1, 2, 0, 0, 0, 0, time.UTC)}, []interface{}{&actual.t}, allTypes{t: time.Date(1987, 1, 2, 0, 0, 0, 0, time.UTC)}},
		{"select $1::oid", []interface{}{pgtype.OID(42)}, []interface{}{&actual.oid}, allTypes{oid: 42}},
	}

	for i, tt := range tests {
		actual = zero

		err := conn.QueryRow(tt.sql, tt.queryArgs...).Scan(tt.scanArgs...)
		if err != nil {
			t.Errorf("%d. Unexpected failure: %v (sql -> %v, queryArgs -> %v)", i, err, tt.sql, tt.queryArgs)
		}

		if actual != tt.expected {
			t.Errorf("%d. Expected %v, got %v (sql -> %v, queryArgs -> %v)", i, tt.expected, actual, tt.sql, tt.queryArgs)
		}

		ensureConnValid(t, conn)

		// Check that Scan errors when a core type is null
		err = conn.QueryRow(tt.sql, nil).Scan(tt.scanArgs...)
		if err == nil {
			t.Errorf("%d. Expected null to cause error, but it didn't (sql -> %v)", i, tt.sql)
		}

		ensureConnValid(t, conn)
	}
}

func TestQueryRowCoreIntegerEncoding(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	type allTypes struct {
		ui   uint
		ui8  uint8
		ui16 uint16
		ui32 uint32
		ui64 uint64
		i    int
		i8   int8
		i16  int16
		i32  int32
		i64  int64
	}

	var actual, zero allTypes

	successfulEncodeTests := []struct {
		sql      string
		queryArg interface{}
		scanArg  interface{}
		expected allTypes
	}{
		// Check any integer type where value is within int2 range can be encoded
		{"select $1::int2", int(42), &actual.i16, allTypes{i16: 42}},
		{"select $1::int2", int8(42), &actual.i16, allTypes{i16: 42}},
		{"select $1::int2", int16(42), &actual.i16, allTypes{i16: 42}},
		{"select $1::int2", int32(42), &actual.i16, allTypes{i16: 42}},
		{"select $1::int2", int64(42), &actual.i16, allTypes{i16: 42}},
		{"select $1::int2", uint(42), &actual.i16, allTypes{i16: 42}},
		{"select $1::int2", uint8(42), &actual.i16, allTypes{i16: 42}},
		{"select $1::int2", uint16(42), &actual.i16, allTypes{i16: 42}},
		{"select $1::int2", uint32(42), &actual.i16, allTypes{i16: 42}},
		{"select $1::int2", uint64(42), &actual.i16, allTypes{i16: 42}},

		// Check any integer type where value is within int4 range can be encoded
		{"select $1::int4", int(42), &actual.i32, allTypes{i32: 42}},
		{"select $1::int4", int8(42), &actual.i32, allTypes{i32: 42}},
		{"select $1::int4", int16(42), &actual.i32, allTypes{i32: 42}},
		{"select $1::int4", int32(42), &actual.i32, allTypes{i32: 42}},
		{"select $1::int4", int64(42), &actual.i32, allTypes{i32: 42}},
		{"select $1::int4", uint(42), &actual.i32, allTypes{i32: 42}},
		{"select $1::int4", uint8(42), &actual.i32, allTypes{i32: 42}},
		{"select $1::int4", uint16(42), &actual.i32, allTypes{i32: 42}},
		{"select $1::int4", uint32(42), &actual.i32, allTypes{i32: 42}},
		{"select $1::int4", uint64(42), &actual.i32, allTypes{i32: 42}},

		// Check any integer type where value is within int8 range can be encoded
		{"select $1::int8", int(42), &actual.i64, allTypes{i64: 42}},
		{"select $1::int8", int8(42), &actual.i64, allTypes{i64: 42}},
		{"select $1::int8", int16(42), &actual.i64, allTypes{i64: 42}},
		{"select $1::int8", int32(42), &actual.i64, allTypes{i64: 42}},
		{"select $1::int8", int64(42), &actual.i64, allTypes{i64: 42}},
		{"select $1::int8", uint(42), &actual.i64, allTypes{i64: 42}},
		{"select $1::int8", uint8(42), &actual.i64, allTypes{i64: 42}},
		{"select $1::int8", uint16(42), &actual.i64, allTypes{i64: 42}},
		{"select $1::int8", uint32(42), &actual.i64, allTypes{i64: 42}},
		{"select $1::int8", uint64(42), &actual.i64, allTypes{i64: 42}},
	}

	for i, tt := range successfulEncodeTests {
		actual = zero

		err := conn.QueryRow(tt.sql, tt.queryArg).Scan(tt.scanArg)
		if err != nil {
			t.Errorf("%d. Unexpected failure: %v (sql -> %v, queryArg -> %v)", i, err, tt.sql, tt.queryArg)
			continue
		}

		if actual != tt.expected {
			t.Errorf("%d. Expected %v, got %v (sql -> %v, queryArg -> %v)", i, tt.expected, actual, tt.sql, tt.queryArg)
		}

		ensureConnValid(t, conn)
	}

	failedEncodeTests := []struct {
		sql      string
		queryArg interface{}
	}{
		// Check any integer type where value is outside pg:int2 range cannot be encoded
		{"select $1::int2", int(32769)},
		{"select $1::int2", int32(32769)},
		{"select $1::int2", int32(32769)},
		{"select $1::int2", int64(32769)},
		{"select $1::int2", uint(32769)},
		{"select $1::int2", uint16(32769)},
		{"select $1::int2", uint32(32769)},
		{"select $1::int2", uint64(32769)},

		// Check any integer type where value is outside pg:int4 range cannot be encoded
		{"select $1::int4", int64(2147483649)},
		{"select $1::int4", uint32(2147483649)},
		{"select $1::int4", uint64(2147483649)},

		// Check any integer type where value is outside pg:int8 range cannot be encoded
		{"select $1::int8", uint64(9223372036854775809)},
	}

	for i, tt := range failedEncodeTests {
		err := conn.QueryRow(tt.sql, tt.queryArg).Scan(nil)
		if err == nil {
			t.Errorf("%d. Expected failure to encode, but unexpectedly succeeded: %v (sql -> %v, queryArg -> %v)", i, err, tt.sql, tt.queryArg)
		} else if !strings.Contains(err.Error(), "is greater than") {
			t.Errorf("%d. Expected failure to encode, but got: %v (sql -> %v, queryArg -> %v)", i, err, tt.sql, tt.queryArg)
		}

		ensureConnValid(t, conn)
	}
}

func TestQueryRowCoreIntegerDecoding(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	type allTypes struct {
		ui   uint
		ui8  uint8
		ui16 uint16
		ui32 uint32
		ui64 uint64
		i    int
		i8   int8
		i16  int16
		i32  int32
		i64  int64
	}

	var actual, zero allTypes

	successfulDecodeTests := []struct {
		sql      string
		scanArg  interface{}
		expected allTypes
	}{
		// Check any integer type where value is within Go:int range can be decoded
		{"select 42::int2", &actual.i, allTypes{i: 42}},
		{"select 42::int4", &actual.i, allTypes{i: 42}},
		{"select 42::int8", &actual.i, allTypes{i: 42}},
		{"select -42::int2", &actual.i, allTypes{i: -42}},
		{"select -42::int4", &actual.i, allTypes{i: -42}},
		{"select -42::int8", &actual.i, allTypes{i: -42}},

		// Check any integer type where value is within Go:int8 range can be decoded
		{"select 42::int2", &actual.i8, allTypes{i8: 42}},
		{"select 42::int4", &actual.i8, allTypes{i8: 42}},
		{"select 42::int8", &actual.i8, allTypes{i8: 42}},
		{"select -42::int2", &actual.i8, allTypes{i8: -42}},
		{"select -42::int4", &actual.i8, allTypes{i8: -42}},
		{"select -42::int8", &actual.i8, allTypes{i8: -42}},

		// Check any integer type where value is within Go:int16 range can be decoded
		{"select 42::int2", &actual.i16, allTypes{i16: 42}},
		{"select 42::int4", &actual.i16, allTypes{i16: 42}},
		{"select 42::int8", &actual.i16, allTypes{i16: 42}},
		{"select -42::int2", &actual.i16, allTypes{i16: -42}},
		{"select -42::int4", &actual.i16, allTypes{i16: -42}},
		{"select -42::int8", &actual.i16, allTypes{i16: -42}},

		// Check any integer type where value is within Go:int32 range can be decoded
		{"select 42::int2", &actual.i32, allTypes{i32: 42}},
		{"select 42::int4", &actual.i32, allTypes{i32: 42}},
		{"select 42::int8", &actual.i32, allTypes{i32: 42}},
		{"select -42::int2", &actual.i32, allTypes{i32: -42}},
		{"select -42::int4", &actual.i32, allTypes{i32: -42}},
		{"select -42::int8", &actual.i32, allTypes{i32: -42}},

		// Check any integer type where value is within Go:int64 range can be decoded
		{"select 42::int2", &actual.i64, allTypes{i64: 42}},
		{"select 42::int4", &actual.i64, allTypes{i64: 42}},
		{"select 42::int8", &actual.i64, allTypes{i64: 42}},
		{"select -42::int2", &actual.i64, allTypes{i64: -42}},
		{"select -42::int4", &actual.i64, allTypes{i64: -42}},
		{"select -42::int8", &actual.i64, allTypes{i64: -42}},

		// Check any integer type where value is within Go:uint range can be decoded
		{"select 128::int2", &actual.ui, allTypes{ui: 128}},
		{"select 128::int4", &actual.ui, allTypes{ui: 128}},
		{"select 128::int8", &actual.ui, allTypes{ui: 128}},

		// Check any integer type where value is within Go:uint8 range can be decoded
		{"select 128::int2", &actual.ui8, allTypes{ui8: 128}},
		{"select 128::int4", &actual.ui8, allTypes{ui8: 128}},
		{"select 128::int8", &actual.ui8, allTypes{ui8: 128}},

		// Check any integer type where value is within Go:uint16 range can be decoded
		{"select 42::int2", &actual.ui16, allTypes{ui16: 42}},
		{"select 32768::int4", &actual.ui16, allTypes{ui16: 32768}},
		{"select 32768::int8", &actual.ui16, allTypes{ui16: 32768}},

		// Check any integer type where value is within Go:uint32 range can be decoded
		{"select 42::int2", &actual.ui32, allTypes{ui32: 42}},
		{"select 42::int4", &actual.ui32, allTypes{ui32: 42}},
		{"select 2147483648::int8", &actual.ui32, allTypes{ui32: 2147483648}},

		// Check any integer type where value is within Go:uint64 range can be decoded
		{"select 42::int2", &actual.ui64, allTypes{ui64: 42}},
		{"select 42::int4", &actual.ui64, allTypes{ui64: 42}},
		{"select 42::int8", &actual.ui64, allTypes{ui64: 42}},
	}

	for i, tt := range successfulDecodeTests {
		actual = zero

		err := conn.QueryRow(tt.sql).Scan(tt.scanArg)
		if err != nil {
			t.Errorf("%d. Unexpected failure: %v (sql -> %v)", i, err, tt.sql)
			continue
		}

		if actual != tt.expected {
			t.Errorf("%d. Expected %v, got %v (sql -> %v)", i, tt.expected, actual, tt.sql)
		}

		ensureConnValid(t, conn)
	}

	failedDecodeTests := []struct {
		sql         string
		scanArg     interface{}
		expectedErr string
	}{
		// Check any integer type where value is outside Go:int8 range cannot be decoded
		{"select 128::int2", &actual.i8, "is greater than"},
		{"select 128::int4", &actual.i8, "is greater than"},
		{"select 128::int8", &actual.i8, "is greater than"},
		{"select -129::int2", &actual.i8, "is less than"},
		{"select -129::int4", &actual.i8, "is less than"},
		{"select -129::int8", &actual.i8, "is less than"},

		// Check any integer type where value is outside Go:int16 range cannot be decoded
		{"select 32768::int4", &actual.i16, "is greater than"},
		{"select 32768::int8", &actual.i16, "is greater than"},
		{"select -32769::int4", &actual.i16, "is less than"},
		{"select -32769::int8", &actual.i16, "is less than"},

		// Check any integer type where value is outside Go:int32 range cannot be decoded
		{"select 2147483648::int8", &actual.i32, "is greater than"},
		{"select -2147483649::int8", &actual.i32, "is less than"},

		// Check any integer type where value is outside Go:uint range cannot be decoded
		{"select -1::int2", &actual.ui, "is less than"},
		{"select -1::int4", &actual.ui, "is less than"},
		{"select -1::int8", &actual.ui, "is less than"},

		// Check any integer type where value is outside Go:uint8 range cannot be decoded
		{"select 256::int2", &actual.ui8, "is greater than"},
		{"select 256::int4", &actual.ui8, "is greater than"},
		{"select 256::int8", &actual.ui8, "is greater than"},
		{"select -1::int2", &actual.ui8, "is less than"},
		{"select -1::int4", &actual.ui8, "is less than"},
		{"select -1::int8", &actual.ui8, "is less than"},

		// Check any integer type where value is outside Go:uint16 cannot be decoded
		{"select 65536::int4", &actual.ui16, "is greater than"},
		{"select 65536::int8", &actual.ui16, "is greater than"},
		{"select -1::int2", &actual.ui16, "is less than"},
		{"select -1::int4", &actual.ui16, "is less than"},
		{"select -1::int8", &actual.ui16, "is less than"},

		// Check any integer type where value is outside Go:uint32 range cannot be decoded
		{"select 4294967296::int8", &actual.ui32, "is greater than"},
		{"select -1::int2", &actual.ui32, "is less than"},
		{"select -1::int4", &actual.ui32, "is less than"},
		{"select -1::int8", &actual.ui32, "is less than"},

		// Check any integer type where value is outside Go:uint64 range cannot be decoded
		{"select -1::int2", &actual.ui64, "is less than"},
		{"select -1::int4", &actual.ui64, "is less than"},
		{"select -1::int8", &actual.ui64, "is less than"},
	}

	for i, tt := range failedDecodeTests {
		err := conn.QueryRow(tt.sql).Scan(tt.scanArg)
		if err == nil {
			t.Errorf("%d. Expected failure to decode, but unexpectedly succeeded: %v (sql -> %v)", i, err, tt.sql)
		} else if !strings.Contains(err.Error(), tt.expectedErr) {
			t.Errorf("%d. Expected failure to decode, but got: %v (sql -> %v)", i, err, tt.sql)
		}

		ensureConnValid(t, conn)
	}
}

func TestQueryRowCoreByteSlice(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	tests := []struct {
		sql      string
		queryArg interface{}
		expected []byte
	}{
		{"select $1::text", "Jack", []byte("Jack")},
		{"select $1::text", []byte("Jack"), []byte("Jack")},
		{"select $1::varchar", []byte("Jack"), []byte("Jack")},
		{"select $1::bytea", []byte{0, 15, 255, 17}, []byte{0, 15, 255, 17}},
	}

	for i, tt := range tests {
		var actual []byte

		err := conn.QueryRow(tt.sql, tt.queryArg).Scan(&actual)
		if err != nil {
			t.Errorf("%d. Unexpected failure: %v (sql -> %v)", i, err, tt.sql)
		}

		if !bytes.Equal(actual, tt.expected) {
			t.Errorf("%d. Expected %v, got %v (sql -> %v)", i, tt.expected, actual, tt.sql)
		}

		ensureConnValid(t, conn)
	}
}

func TestQueryRowUnknownType(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	// Clear existing type mappings
	conn.ConnInfo = pgtype.NewConnInfo()
	conn.ConnInfo.RegisterDataType(pgtype.DataType{
		Value: &pgtype.GenericText{},
		Name:  "point",
		OID:   600,
	})
	conn.ConnInfo.RegisterDataType(pgtype.DataType{
		Value: &pgtype.Int4{},
		Name:  "int4",
		OID:   pgtype.Int4OID,
	})

	sql := "select $1::point"
	expected := "(1,0)"
	var actual string

	err := conn.QueryRow(sql, expected).Scan(&actual)
	if err != nil {
		t.Errorf("Unexpected failure: %v (sql -> %v)", err, sql)
	}

	if actual != expected {
		t.Errorf(`Expected "%v", got "%v" (sql -> %v)`, expected, actual, sql)

	}

	ensureConnValid(t, conn)
}

func TestQueryRowErrors(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	type allTypes struct {
		i16 int16
		i   int
		s   string
	}

	var actual, zero allTypes

	tests := []struct {
		sql       string
		queryArgs []interface{}
		scanArgs  []interface{}
		err       string
	}{
		{"select $1::badtype", []interface{}{"Jack"}, []interface{}{&actual.i16}, `type "badtype" does not exist`},
		{"SYNTAX ERROR", []interface{}{}, []interface{}{&actual.i16}, "SQLSTATE 42601"},
		{"select $1::text", []interface{}{"Jack"}, []interface{}{&actual.i16}, "cannot decode"},
		{"select $1::point", []interface{}{int(705)}, []interface{}{&actual.s}, "cannot convert 705 to Point"},
	}

	for i, tt := range tests {
		actual = zero

		err := conn.QueryRow(tt.sql, tt.queryArgs...).Scan(tt.scanArgs...)
		if err == nil {
			t.Errorf("%d. Unexpected success (sql -> %v, queryArgs -> %v)", i, tt.sql, tt.queryArgs)
		}
		if err != nil && !strings.Contains(err.Error(), tt.err) {
			t.Errorf("%d. Expected error to contain %s, but got %v (sql -> %v, queryArgs -> %v)", i, tt.err, err, tt.sql, tt.queryArgs)
		}

		ensureConnValid(t, conn)
	}
}

func TestQueryRowExErrorsWrongParameterOIDs(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	sql := `
		with t as (
			select 1::int8 as some_int, 'foo'::text as some_text
		)
		select some_int from t where some_text = $1`
	paramOIDs := []pgtype.OID{pgtype.TextArrayOID}
	queryArgs := []interface{}{"bar"}
	expectedErr := "operator does not exist: text = text[] (SQLSTATE 42883)"
	var result int64

	err := conn.QueryRowEx(
		context.Background(),
		sql,
		&pgx.QueryExOptions{
			ParameterOIDs:     paramOIDs,
			ResultFormatCodes: []int16{pgx.BinaryFormatCode},
		},
		queryArgs...,
	).Scan(&result)

	if err == nil {
		t.Errorf("Unexpected success (sql -> %v, paramOIDs -> %v, queryArgs -> %v)", sql, paramOIDs, queryArgs)
	}
	if err != nil && !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("Expected error to contain %s, but got %v (sql -> %v, paramOIDs -> %v, queryArgs -> %v)",
			expectedErr, err, sql, paramOIDs, queryArgs)
	}

	ensureConnValid(t, conn)
}

func TestQueryRowNoResults(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	var n int32
	err := conn.QueryRow("select 1 where 1=0").Scan(&n)
	if err != pgx.ErrNoRows {
		t.Errorf("Expected pgx.ErrNoRows, got %v", err)
	}

	ensureConnValid(t, conn)
}

func TestReadingValueAfterEmptyArray(t *testing.T) {
	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	var a []string
	var b int32
	err := conn.QueryRow("select '{}'::text[], 42::integer").Scan(&a, &b)
	if err != nil {
		t.Fatalf("conn.QueryRow failed: %v", err)
	}

	if len(a) != 0 {
		t.Errorf("Expected 'a' to have length 0, but it was: %d", len(a))
	}

	if b != 42 {
		t.Errorf("Expected 'b' to 42, but it was: %d", b)
	}
}

func TestReadingNullByteArray(t *testing.T) {
	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	var a []byte
	err := conn.QueryRow("select null::text").Scan(&a)
	if err != nil {
		t.Fatalf("conn.QueryRow failed: %v", err)
	}

	if a != nil {
		t.Errorf("Expected 'a' to be nil, but it was: %v", a)
	}
}

func TestReadingNullByteArrays(t *testing.T) {
	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	rows, err := conn.Query("select null::text union all select null::text")
	if err != nil {
		t.Fatalf("conn.Query failed: %v", err)
	}

	count := 0
	for rows.Next() {
		count++
		var a []byte
		if err := rows.Scan(&a); err != nil {
			t.Fatalf("failed to scan row: %v", err)
		}
		if a != nil {
			t.Errorf("Expected 'a' to be nil, but it was: %v", a)
		}
	}
	if count != 2 {
		t.Errorf("Expected to read 2 rows, read: %d", count)
	}
}

// Use github.com/shopspring/decimal as real-world database/sql custom type
// to test against.
func TestConnQueryDatabaseSQLScanner(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	var num decimal.Decimal

	err := conn.QueryRow("select '1234.567'::decimal").Scan(&num)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	expected, err := decimal.NewFromString("1234.567")
	if err != nil {
		t.Fatal(err)
	}

	if !num.Equals(expected) {
		t.Errorf("Expected num to be %v, but it was %v", expected, num)
	}

	ensureConnValid(t, conn)
}

// Use github.com/shopspring/decimal as real-world database/sql custom type
// to test against.
func TestConnQueryDatabaseSQLDriverValuer(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	expected, err := decimal.NewFromString("1234.567")
	if err != nil {
		t.Fatal(err)
	}
	var num decimal.Decimal

	err = conn.QueryRow("select $1::decimal", &expected).Scan(&num)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if !num.Equals(expected) {
		t.Errorf("Expected num to be %v, but it was %v", expected, num)
	}

	ensureConnValid(t, conn)
}

// https://github.com/jackc/pgx/issues/339
func TestConnQueryDatabaseSQLDriverValuerWithAutoGeneratedPointerReceiver(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	mustExec(t, conn, "create temporary table t(n numeric)")

	var d *apd.Decimal
	commandTag, err := conn.Exec(`insert into t(n) values($1)`, d)
	if err != nil {
		t.Fatal(err)
	}
	if commandTag != "INSERT 0 1" {
		t.Fatalf("want %s, got %s", "INSERT 0 1", commandTag)
	}

	ensureConnValid(t, conn)
}

func TestConnQueryDatabaseSQLDriverValuerWithBinaryPgTypeThatAcceptsSameType(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	conn.ConnInfo.RegisterDataType(pgtype.DataType{
		Value: &satori.UUID{},
		Name:  "uuid",
		OID:   2950,
	})

	expected, err := uuid.FromString("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	if err != nil {
		t.Fatal(err)
	}

	var u2 uuid.UUID
	err = conn.QueryRow("select $1::uuid", expected).Scan(&u2)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if expected != u2 {
		t.Errorf("Expected u2 to be %v, but it was %v", expected, u2)
	}

	ensureConnValid(t, conn)
}

func TestConnQueryDatabaseSQLNullX(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	type row struct {
		boolValid    sql.NullBool
		boolNull     sql.NullBool
		int64Valid   sql.NullInt64
		int64Null    sql.NullInt64
		float64Valid sql.NullFloat64
		float64Null  sql.NullFloat64
		stringValid  sql.NullString
		stringNull   sql.NullString
	}

	expected := row{
		boolValid:    sql.NullBool{Bool: true, Valid: true},
		int64Valid:   sql.NullInt64{Int64: 123, Valid: true},
		float64Valid: sql.NullFloat64{Float64: 3.14, Valid: true},
		stringValid:  sql.NullString{String: "pgx", Valid: true},
	}

	var actual row

	err := conn.QueryRow(
		"select $1::bool, $2::bool, $3::int8, $4::int8, $5::float8, $6::float8, $7::text, $8::text",
		expected.boolValid,
		expected.boolNull,
		expected.int64Valid,
		expected.int64Null,
		expected.float64Valid,
		expected.float64Null,
		expected.stringValid,
		expected.stringNull,
	).Scan(
		&actual.boolValid,
		&actual.boolNull,
		&actual.int64Valid,
		&actual.int64Null,
		&actual.float64Valid,
		&actual.float64Null,
		&actual.stringValid,
		&actual.stringNull,
	)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if expected != actual {
		t.Errorf("Expected %v, but got %v", expected, actual)
	}

	ensureConnValid(t, conn)
}

func TestQueryExContextSuccess(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	rows, err := conn.QueryEx(ctx, "select 42::integer", nil)
	if err != nil {
		t.Fatal(err)
	}

	var result, rowCount int
	for rows.Next() {
		err = rows.Scan(&result)
		if err != nil {
			t.Fatal(err)
		}
		rowCount++
	}

	if rows.Err() != nil {
		t.Fatal(rows.Err())
	}

	if rowCount != 1 {
		t.Fatalf("Expected 1 row, got %d", rowCount)
	}
	if result != 42 {
		t.Fatalf("Expected result 42, got %d", result)
	}

	ensureConnValid(t, conn)
}

func TestQueryExContextErrorWhileReceivingRows(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	rows, err := conn.QueryEx(ctx, "select 10/(10-n) from generate_series(1, 100) n", nil)
	if err != nil {
		t.Fatal(err)
	}

	var result, rowCount int
	for rows.Next() {
		err = rows.Scan(&result)
		if err != nil {
			t.Fatal(err)
		}
		rowCount++
	}

	if rows.Err() == nil || rows.Err().Error() != "ERROR: division by zero (SQLSTATE 22012)" {
		t.Fatalf("Expected division by zero error, but got %v", rows.Err())
	}

	if rowCount != 9 {
		t.Fatalf("Expected 9 rows, got %d", rowCount)
	}
	if result != 10 {
		t.Fatalf("Expected result 10, got %d", result)
	}

	ensureConnValid(t, conn)
}

func TestQueryExContextCancelationCancelsQuery(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	ctx, cancelFunc := context.WithCancel(context.Background())
	go func() {
		time.Sleep(500 * time.Millisecond)
		cancelFunc()
	}()

	rows, err := conn.QueryEx(ctx, "select pg_sleep(5)", nil)
	if err != nil {
		t.Fatal(err)
	}

	for rows.Next() {
		t.Fatal("No rows should ever be ready -- context cancel apparently did not happen")
	}

	if rows.Err() != context.Canceled {
		t.Fatalf("Expected context.Canceled error, got %v", rows.Err())
	}

	ensureConnValid(t, conn)
}

func TestQueryRowExContextSuccess(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	var result int
	err := conn.QueryRowEx(ctx, "select 42::integer", nil).Scan(&result)
	if err != nil {
		t.Fatal(err)
	}
	if result != 42 {
		t.Fatalf("Expected result 42, got %d", result)
	}

	ensureConnValid(t, conn)
}

func TestQueryRowExContextErrorWhileReceivingRow(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	var result int
	err := conn.QueryRowEx(ctx, "select 10/0", nil).Scan(&result)
	if err == nil || err.Error() != "ERROR: division by zero (SQLSTATE 22012)" {
		t.Fatalf("Expected division by zero error, but got %v", err)
	}

	ensureConnValid(t, conn)
}

func TestQueryRowExContextCancelationCancelsQuery(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	ctx, cancelFunc := context.WithCancel(context.Background())
	go func() {
		time.Sleep(500 * time.Millisecond)
		cancelFunc()
	}()

	var result []byte
	err := conn.QueryRowEx(ctx, "select pg_sleep(5)", nil).Scan(&result)
	if err != context.Canceled {
		t.Fatalf("Expected context.Canceled error, got %v", err)
	}

	ensureConnValid(t, conn)
}

func TestConnQueryRowExSingleRoundTrip(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	var result int32
	err := conn.QueryRowEx(
		context.Background(),
		"select $1 + $2",
		&pgx.QueryExOptions{
			ParameterOIDs:     []pgtype.OID{pgtype.Int4OID, pgtype.Int4OID},
			ResultFormatCodes: []int16{pgx.BinaryFormatCode},
		},
		1, 2,
	).Scan(&result)
	if err != nil {
		t.Fatal(err)
	}
	if result != 3 {
		t.Fatalf("result => %d, want %d", result, 3)
	}

	ensureConnValid(t, conn)
}

func TestConnSimpleProtocol(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	// Test all supported low-level types

	{
		expected := int64(42)
		var actual int64
		err := conn.QueryRowEx(
			context.Background(),
			"select $1::int8",
			&pgx.QueryExOptions{SimpleProtocol: true},
			expected,
		).Scan(&actual)
		if err != nil {
			t.Error(err)
		}
		if expected != actual {
			t.Errorf("expected %v got %v", expected, actual)
		}
	}

	{
		expected := float64(1.23)
		var actual float64
		err := conn.QueryRowEx(
			context.Background(),
			"select $1::float8",
			&pgx.QueryExOptions{SimpleProtocol: true},
			expected,
		).Scan(&actual)
		if err != nil {
			t.Error(err)
		}
		if expected != actual {
			t.Errorf("expected %v got %v", expected, actual)
		}
	}

	{
		expected := true
		var actual bool
		err := conn.QueryRowEx(
			context.Background(),
			"select $1",
			&pgx.QueryExOptions{SimpleProtocol: true},
			expected,
		).Scan(&actual)
		if err != nil {
			t.Error(err)
		}
		if expected != actual {
			t.Errorf("expected %v got %v", expected, actual)
		}
	}

	{
		expected := []byte{0, 1, 20, 35, 64, 80, 120, 3, 255, 240, 128, 95}
		var actual []byte
		err := conn.QueryRowEx(
			context.Background(),
			"select $1::bytea",
			&pgx.QueryExOptions{SimpleProtocol: true},
			expected,
		).Scan(&actual)
		if err != nil {
			t.Error(err)
		}
		if bytes.Compare(actual, expected) != 0 {
			t.Errorf("expected %v got %v", expected, actual)
		}
	}

	{
		expected := "test"
		var actual string
		err := conn.QueryRowEx(
			context.Background(),
			"select $1::text",
			&pgx.QueryExOptions{SimpleProtocol: true},
			expected,
		).Scan(&actual)
		if err != nil {
			t.Error(err)
		}
		if expected != actual {
			t.Errorf("expected %v got %v", expected, actual)
		}
	}

	// Test high-level type

	{
		expected := pgtype.Circle{P: pgtype.Vec2{1, 2}, R: 1.5, Status: pgtype.Present}
		actual := expected
		err := conn.QueryRowEx(
			context.Background(),
			"select $1::circle",
			&pgx.QueryExOptions{SimpleProtocol: true},
			&expected,
		).Scan(&actual)
		if err != nil {
			t.Error(err)
		}
		if expected != actual {
			t.Errorf("expected %v got %v", expected, actual)
		}
	}

	// Test multiple args in single query

	{
		expectedInt64 := int64(234423)
		expectedFloat64 := float64(-0.2312)
		expectedBool := true
		expectedBytes := []byte{255, 0, 23, 16, 87, 45, 9, 23, 45, 223}
		expectedString := "test"
		var actualInt64 int64
		var actualFloat64 float64
		var actualBool bool
		var actualBytes []byte
		var actualString string
		err := conn.QueryRowEx(
			context.Background(),
			"select $1::int8, $2::float8, $3, $4::bytea, $5::text",
			&pgx.QueryExOptions{SimpleProtocol: true},
			expectedInt64, expectedFloat64, expectedBool, expectedBytes, expectedString,
		).Scan(&actualInt64, &actualFloat64, &actualBool, &actualBytes, &actualString)
		if err != nil {
			t.Error(err)
		}
		if expectedInt64 != actualInt64 {
			t.Errorf("expected %v got %v", expectedInt64, actualInt64)
		}
		if expectedFloat64 != actualFloat64 {
			t.Errorf("expected %v got %v", expectedFloat64, actualFloat64)
		}
		if expectedBool != actualBool {
			t.Errorf("expected %v got %v", expectedBool, actualBool)
		}
		if bytes.Compare(expectedBytes, actualBytes) != 0 {
			t.Errorf("expected %v got %v", expectedBytes, actualBytes)
		}
		if expectedString != actualString {
			t.Errorf("expected %v got %v", expectedString, actualString)
		}
	}

	// Test dangerous cases

	{
		expected := "foo';drop table users;"
		var actual string
		err := conn.QueryRowEx(
			context.Background(),
			"select $1",
			&pgx.QueryExOptions{SimpleProtocol: true},
			expected,
		).Scan(&actual)
		if err != nil {
			t.Error(err)
		}
		if expected != actual {
			t.Errorf("expected %v got %v", expected, actual)
		}
	}

	ensureConnValid(t, conn)
}

func TestConnSimpleProtocolRefusesNonUTF8ClientEncoding(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	mustExec(t, conn, "set client_encoding to 'SQL_ASCII'")

	var expected string
	err := conn.QueryRowEx(
		context.Background(),
		"select $1",
		&pgx.QueryExOptions{SimpleProtocol: true},
		"test",
	).Scan(&expected)
	if err == nil {
		t.Error("expected error when client_encoding not UTF8, but no error occurred")
	}

	ensureConnValid(t, conn)
}

func TestConnSimpleProtocolRefusesNonStandardConformingStrings(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	mustExec(t, conn, "set standard_conforming_strings to off")

	var expected string
	err := conn.QueryRowEx(
		context.Background(),
		"select $1",
		&pgx.QueryExOptions{SimpleProtocol: true},
		`\'; drop table users; --`,
	).Scan(&expected)
	if err == nil {
		t.Error("expected error when standard_conforming_strings is off, but no error occurred")
	}

	ensureConnValid(t, conn)
}
