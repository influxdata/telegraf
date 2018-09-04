package mssql

import (
	"bytes"
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"math"
	"net"
	"strings"
	"testing"
	"time"
)

func driverWithProcess(t *testing.T) *MssqlDriver {
	return &MssqlDriver{
		log:              optionalLogger{testLogger{t}},
		processQueryText: true,
	}
}
func driverNoProcess(t *testing.T) *MssqlDriver {
	return &MssqlDriver{
		log:              optionalLogger{testLogger{t}},
		processQueryText: false,
	}
}

func TestSelect(t *testing.T) {
	conn := open(t)
	defer conn.Close()

	type testStruct struct {
		sql string
		val interface{}
	}

	longstr := strings.Repeat("x", 10000)

	values := []testStruct{
		{"1", int64(1)},
		{"-1", int64(-1)},
		{"cast(1 as int)", int64(1)},
		{"cast(-1 as int)", int64(-1)},
		{"cast(1 as tinyint)", int64(1)},
		{"cast(255 as tinyint)", int64(255)},
		{"cast(1 as smallint)", int64(1)},
		{"cast(-1 as smallint)", int64(-1)},
		{"cast(1 as bigint)", int64(1)},
		{"cast(-1 as bigint)", int64(-1)},
		{"cast(1 as bit)", true},
		{"cast(0 as bit)", false},
		{"'abc'", string("abc")},
		{"cast(0.5 as float)", float64(0.5)},
		{"cast(0.5 as real)", float64(0.5)},
		{"cast(1 as decimal)", []byte("1")},
		{"cast(1.2345 as money)", []byte("1.2345")},
		{"cast(-1.2345 as money)", []byte("-1.2345")},
		{"cast(1.2345 as smallmoney)", []byte("1.2345")},
		{"cast(-1.2345 as smallmoney)", []byte("-1.2345")},
		{"cast(0.5 as decimal(18,1))", []byte("0.5")},
		{"cast(-0.5 as decimal(18,1))", []byte("-0.5")},
		{"cast(-0.5 as numeric(18,1))", []byte("-0.5")},
		{"cast(4294967296 as numeric(20,0))", []byte("4294967296")},
		{"cast(-0.5 as numeric(18,2))", []byte("-0.50")},
		{"N'abc'", string("abc")},
		{"cast(null as nvarchar(3))", nil},
		{"NULL", nil},
		{"cast('2000-01-01' as datetime)", time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"cast('2000-01-01T12:13:14.12' as datetime)",
			time.Date(2000, 1, 1, 12, 13, 14, 120000000, time.UTC)},
		{"cast('2014-06-26 11:08:09.673' as datetime)", time.Date(2014, 06, 26, 11, 8, 9, 673000000, time.UTC)},
		{"cast(NULL as datetime)", nil},
		{"cast('2000-01-01T12:13:00' as smalldatetime)",
			time.Date(2000, 1, 1, 12, 13, 0, 0, time.UTC)},
		{"cast(0x6F9619FF8B86D011B42D00C04FC964FF as uniqueidentifier)",
			[]byte{0x6F, 0x96, 0x19, 0xFF, 0x8B, 0x86, 0xD0, 0x11, 0xB4, 0x2D, 0x00, 0xC0, 0x4F, 0xC9, 0x64, 0xFF}},
		{"cast(NULL as uniqueidentifier)", nil},
		{"cast(0x1234 as varbinary(2))", []byte{0x12, 0x34}},
		{"cast(N'abc' as nvarchar(max))", "abc"},
		{"cast(null as nvarchar(max))", nil},
		{"cast('<root/>' as xml)", "<root/>"},
		{"cast('abc' as text)", "abc"},
		{"cast(null as text)", nil},
		{"cast(N'abc' as ntext)", "abc"},
		{"cast(0x1234 as image)", []byte{0x12, 0x34}},
		{"cast('abc' as char(3))", "abc"},
		{"cast('abc' as varchar(3))", "abc"},
		{"cast(N'проверка' as nvarchar(max))", "проверка"},
		{"cast(N'Δοκιμή' as nvarchar(max))", "Δοκιμή"},
		{"cast(cast(N'สวัสดี' as nvarchar(max)) collate Thai_CI_AI as varchar(max))", "สวัสดี"},                // cp874
		{"cast(cast(N'你好' as nvarchar(max)) collate Chinese_PRC_CI_AI as varchar(max))", "你好"},                 // cp936
		{"cast(cast(N'こんにちは' as nvarchar(max)) collate Japanese_CI_AI as varchar(max))", "こんにちは"},              // cp939
		{"cast(cast(N'안녕하세요.' as nvarchar(max)) collate Korean_90_CI_AI as varchar(max))", "안녕하세요."},           // cp949
		{"cast(cast(N'你好' as nvarchar(max)) collate Chinese_Hong_Kong_Stroke_90_CI_AI as varchar(max))", "你好"}, // cp950
		{"cast(cast(N'cześć' as nvarchar(max)) collate Polish_CI_AI as varchar(max))", "cześć"},                // cp1250
		{"cast(cast(N'Алло' as nvarchar(max)) collate Cyrillic_General_CI_AI as varchar(max))", "Алло"},        // cp1251
		{"cast(cast(N'Bonjour' as nvarchar(max)) collate French_CI_AI as varchar(max))", "Bonjour"},            // cp1252
		{"cast(cast(N'Γεια σας' as nvarchar(max)) collate Greek_CI_AI as varchar(max))", "Γεια σας"},           // cp1253
		{"cast(cast(N'Merhaba' as nvarchar(max)) collate Turkish_CI_AI as varchar(max))", "Merhaba"},           // cp1254
		{"cast(cast(N'שלום' as nvarchar(max)) collate Hebrew_CI_AI as varchar(max))", "שלום"},                  // cp1255
		{"cast(cast(N'مرحبا' as nvarchar(max)) collate Arabic_CI_AI as varchar(max))", "مرحبا"},                // cp1256
		{"cast(cast(N'Sveiki' as nvarchar(max)) collate Lithuanian_CI_AI as varchar(max))", "Sveiki"},          // cp1257
		{"cast(cast(N'chào' as nvarchar(max)) collate Vietnamese_CI_AI as varchar(max))", "chào"},              // cp1258
		{fmt.Sprintf("cast(N'%s' as nvarchar(max))", longstr), longstr},
		{"cast(NULL as sql_variant)", nil},
		{"cast(cast(0x6F9619FF8B86D011B42D00C04FC964FF as uniqueidentifier) as sql_variant)",
			[]byte{0x6F, 0x96, 0x19, 0xFF, 0x8B, 0x86, 0xD0, 0x11, 0xB4, 0x2D, 0x00, 0xC0, 0x4F, 0xC9, 0x64, 0xFF}},
		{"cast(cast(1 as bit) as sql_variant)", true},
		{"cast(cast(10 as tinyint) as sql_variant)", int64(10)},
		{"cast(cast(-10 as smallint) as sql_variant)", int64(-10)},
		{"cast(cast(-20 as int) as sql_variant)", int64(-20)},
		{"cast(cast(-20 as bigint) as sql_variant)", int64(-20)},
		{"cast(cast('2000-01-01' as datetime) as sql_variant)", time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"cast(cast('2000-01-01T12:13:00' as smalldatetime) as sql_variant)",
			time.Date(2000, 1, 1, 12, 13, 0, 0, time.UTC)},
		{"cast(cast(0.125 as real) as sql_variant)", float64(0.125)},
		{"cast(cast(0.125 as float) as sql_variant)", float64(0.125)},
		{"cast(cast(1.2345 as smallmoney) as sql_variant)", []byte("1.2345")},
		{"cast(cast(1.2345 as money) as sql_variant)", []byte("1.2345")},
		{"cast(cast(0x1234 as varbinary(2)) as sql_variant)", []byte{0x12, 0x34}},
		{"cast(cast(0x1234 as binary(2)) as sql_variant)", []byte{0x12, 0x34}},
		{"cast(cast(-0.5 as decimal(18,1)) as sql_variant)", []byte("-0.5")},
		{"cast(cast(-0.5 as numeric(18,1)) as sql_variant)", []byte("-0.5")},
		{"cast(cast('abc' as varchar(3)) as sql_variant)", "abc"},
		{"cast(cast('abc' as char(3)) as sql_variant)", "abc"},
		{"cast(N'abc' as sql_variant)", "abc"},
	}

	for _, test := range values {
		stmt, err := conn.Prepare("select " + test.sql)
		if err != nil {
			t.Error("Prepare failed:", test.sql, err.Error())
			return
		}
		defer stmt.Close()

		row := stmt.QueryRow()
		var retval interface{}
		err = row.Scan(&retval)
		if err != nil {
			t.Error("Scan failed:", test.sql, err.Error())
			continue
		}
		var same bool
		switch decodedval := retval.(type) {
		case []byte:
			switch decodedvaltest := test.val.(type) {
			case []byte:
				same = bytes.Equal(decodedval, decodedvaltest)
			default:
				same = false
			}
		default:
			same = retval == test.val
		}
		if !same {
			t.Errorf("Values don't match '%s' '%s' for test: %s", retval, test.val, test.sql)
			continue
		}
	}
}

func TestSelectDateTimeOffset(t *testing.T) {
	type testStruct struct {
		sql string
		val time.Time
	}
	values := []testStruct{
		{"cast('2010-11-15T11:56:45.123+01:00' as datetimeoffset(3))",
			time.Date(2010, 11, 15, 11, 56, 45, 123000000, time.FixedZone("", 60*60))},
		{"cast(cast('2010-11-15T11:56:45.123+10:00' as datetimeoffset(3)) as sql_variant)",
			time.Date(2010, 11, 15, 11, 56, 45, 123000000, time.FixedZone("", 10*60*60))},
	}

	conn := open(t)
	defer conn.Close()
	for _, test := range values {
		row := conn.QueryRow("select " + test.sql)
		var retval interface{}
		err := row.Scan(&retval)
		if err != nil {
			t.Error("Scan failed:", test.sql, err.Error())
			continue
		}
		retvalDate := retval.(time.Time)
		if retvalDate.UTC() != test.val.UTC() {
			t.Errorf("UTC values don't match '%v' '%v' for test: %s", retvalDate, test.val, test.sql)
			continue
		}
		if retvalDate.String() != test.val.String() {
			t.Errorf("Locations don't match '%v' '%v' for test: %s", retvalDate.String(), test.val.String(), test.sql)
			continue
		}
	}
}

func TestSelectNewTypes(t *testing.T) {
	conn := open(t)
	defer conn.Close()
	var ver string
	err := conn.QueryRow("select SERVERPROPERTY('productversion')").Scan(&ver)
	if err != nil {
		t.Fatalf("cannot select productversion: %s", err)
	}
	var n int
	_, err = fmt.Sscanf(ver, "%d", &n)
	if err != nil {
		t.Fatalf("cannot parse productversion: %s", err)
	}
	// 8 is SQL 2000, 9 is SQL 2005, 10 is SQL 2008, 11 is SQL 2012
	if n < 10 {
		return
	}
	// run tests for new data types available only in SQL Server 2008 and later
	type testStruct struct {
		sql string
		val interface{}
	}
	values := []testStruct{
		{"cast('2000-01-01' as date)",
			time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"cast(NULL as date)", nil},
		{"cast('00:00:45.123' as time(3))",
			time.Date(1, 1, 1, 00, 00, 45, 123000000, time.UTC)},
		{"cast('11:56:45.123' as time(3))",
			time.Date(1, 1, 1, 11, 56, 45, 123000000, time.UTC)},
		{"cast('11:56:45' as time(0))",
			time.Date(1, 1, 1, 11, 56, 45, 0, time.UTC)},
		{"cast('2010-11-15T11:56:45.123' as datetime2(3))",
			time.Date(2010, 11, 15, 11, 56, 45, 123000000, time.UTC)},
		{"cast('2010-11-15T11:56:45' as datetime2(0))",
			time.Date(2010, 11, 15, 11, 56, 45, 0, time.UTC)},
		{"cast(cast('2000-01-01' as date) as sql_variant)",
			time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"cast(cast('00:00:45.123' as time(3)) as sql_variant)",
			time.Date(1, 1, 1, 00, 00, 45, 123000000, time.UTC)},
		{"cast(cast('2010-11-15T11:56:45.123' as datetime2(3)) as sql_variant)",
			time.Date(2010, 11, 15, 11, 56, 45, 123000000, time.UTC)},
	}
	for _, test := range values {
		stmt, err := conn.Prepare("select " + test.sql)
		if err != nil {
			t.Error("Prepare failed:", test.sql, err.Error())
			return
		}
		defer stmt.Close()

		row := stmt.QueryRow()
		var retval interface{}
		err = row.Scan(&retval)
		if err != nil {
			t.Error("Scan failed:", test.sql, err.Error())
			continue
		}
		if retval != test.val {
			t.Errorf("Values don't match '%s' '%s' for test: %s", retval, test.val, test.sql)
			continue
		}
	}
}

func TestTrans(t *testing.T) {
	conn := open(t)
	defer conn.Close()

	var tx *sql.Tx
	var err error
	if tx, err = conn.Begin(); err != nil {
		t.Fatal("Begin failed", err.Error())
	}
	if err = tx.Commit(); err != nil {
		t.Fatal("Commit failed", err.Error())
	}

	if tx, err = conn.Begin(); err != nil {
		t.Fatal("Begin failed", err.Error())
	}
	if _, err = tx.Exec("create table #abc (fld int)"); err != nil {
		t.Fatal("Create table failed", err.Error())
	}
	if err = tx.Rollback(); err != nil {
		t.Fatal("Rollback failed", err.Error())
	}
}

func TestParams(t *testing.T) {
	longstr := strings.Repeat("x", 10000)
	longbytes := make([]byte, 10000)
	values := []interface{}{
		int64(5),
		"hello",
		"",
		[]byte{1, 2, 3},
		[]byte{},
		float64(1.12313554),
		true,
		false,
		nil,
		longstr,
		longbytes,
	}

	conn := open(t)
	defer conn.Close()

	for _, val := range values {
		row := conn.QueryRow("select ?", val)
		var retval interface{}
		err := row.Scan(&retval)
		if err != nil {
			t.Error("Scan failed", err.Error())
			return
		}
		var same bool
		switch decodedval := retval.(type) {
		case []byte:
			switch decodedvaltest := val.(type) {
			case []byte:
				same = bytes.Equal(decodedval, decodedvaltest)
			default:
				same = false
			}
		default:
			same = retval == val
		}
		if !same {
			t.Error("Value don't match", retval, val)
			return
		}
	}
}

func TestExec(t *testing.T) {
	conn := open(t)
	defer conn.Close()

	res, err := conn.Exec("create table #abc (fld int)")
	if err != nil {
		t.Fatal("Exec failed", err.Error())
	}
	_ = res
}

func TestShortTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("short")
	}
	checkConnStr(t)
	SetLogger(testLogger{t})
	dsn := makeConnStr(t)
	dsnParams := dsn.Query()
	dsnParams.Set("Connection Timeout", "2")
	dsn.RawQuery = dsnParams.Encode()
	conn, err := sql.Open("mssql", dsn.String())
	if err != nil {
		t.Fatal("Open connection failed:", err.Error())
	}
	defer conn.Close()

	_, err = conn.Exec("waitfor delay '00:00:15'")
	if err == nil {
		t.Fatal("Exec should fail with timeout, but no failure occurred")
	}
	if neterr, ok := err.(net.Error); !ok || !neterr.Timeout() {
		t.Fatal("failure not a timeout, failed with", err)
	}

	// connection should be usable after timeout
	row := conn.QueryRow("select 1")
	var val int64
	err = row.Scan(&val)
	if err != nil {
		t.Fatal("Scan failed with", err)
	}
}

func TestTwoQueries(t *testing.T) {
	conn := open(t)
	defer conn.Close()

	rows, err := conn.Query("select 1")
	if err != nil {
		t.Fatal("First exec failed", err)
	}
	if !rows.Next() {
		t.Fatal("First query didn't return row")
	}
	var i int
	if err = rows.Scan(&i); err != nil {
		t.Fatal("Scan failed", err)
	}
	if i != 1 {
		t.Fatalf("Wrong value returned %d, should be 1", i)
	}

	if rows, err = conn.Query("select 2"); err != nil {
		t.Fatal("Second query failed", err)
	}
	if !rows.Next() {
		t.Fatal("Second query didn't return row")
	}
	if err = rows.Scan(&i); err != nil {
		t.Fatal("Scan failed", err)
	}
	if i != 2 {
		t.Fatalf("Wrong value returned %d, should be 2", i)
	}
}

func TestError(t *testing.T) {
	conn := open(t)
	defer conn.Close()

	_, err := conn.Query("exec bad")
	if err == nil {
		t.Fatal("Query should fail")
	}

	if sqlerr, ok := err.(Error); !ok {
		t.Fatalf("Should be sql error, actually %T, %v", err, err)
	} else {
		if sqlerr.Number != 2812 { // Could not find stored procedure 'bad'
			t.Fatalf("Should be specific error code 2812, actually %d %s", sqlerr.Number, sqlerr)
		}
	}
}

func TestQueryNoRows(t *testing.T) {
	conn := open(t)
	defer conn.Close()

	var rows *sql.Rows
	var err error
	if rows, err = conn.Query("create table #abc (fld int)"); err != nil {
		t.Fatal("Query failed", err)
	}
	if rows.Next() {
		t.Fatal("Query shoulnd't return any rows")
	}
}

func TestQueryManyNullsRow(t *testing.T) {
	conn := open(t)
	defer conn.Close()

	var row *sql.Row
	var err error
	if row = conn.QueryRow("select null, null, null, null, null, null, null, null"); err != nil {
		t.Fatal("Query failed", err)
	}
	var v [8]sql.NullInt64
	if err = row.Scan(&v[0], &v[1], &v[2], &v[3], &v[4], &v[5], &v[6], &v[7]); err != nil {
		t.Fatal("Scan failed", err)
	}
}

func TestOrderBy(t *testing.T) {
	conn := open(t)
	defer conn.Close()

	tx, err := conn.Begin()
	if err != nil {
		t.Fatal("Begin tran failed", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec("if (exists(select * from INFORMATION_SCHEMA.TABLES where TABLE_NAME='tbl')) drop table tbl")
	if err != nil {
		t.Fatal("Drop table failed", err)
	}

	_, err = tx.Exec("create table tbl (fld1 int primary key, fld2 int)")
	if err != nil {
		t.Fatal("Create table failed", err)
	}
	_, err = tx.Exec("insert into tbl (fld1, fld2) values (1, 2)")
	if err != nil {
		t.Fatal("Insert failed", err)
	}
	_, err = tx.Exec("insert into tbl (fld1, fld2) values (2, 1)")
	if err != nil {
		t.Fatal("Insert failed", err)
	}

	rows, err := tx.Query("select * from tbl order by fld1")
	if err != nil {
		t.Fatal("Query failed", err)
	}

	for rows.Next() {
		var fld1 int32
		var fld2 int32
		err = rows.Scan(&fld1, &fld2)
		if err != nil {
			t.Fatal("Scan failed", err)
		}
	}

	err = rows.Err()
	if err != nil {
		t.Fatal("Rows have errors", err)
	}
}

func TestScanDecimal(t *testing.T) {
	conn := open(t)
	defer conn.Close()

	var f float64
	err := conn.QueryRow("select cast(0.5 as numeric(25,1))").Scan(&f)
	if err != nil {
		t.Error("query row / scan failed:", err.Error())
		return
	}
	if math.Abs(f-0.5) > 0.000001 {
		t.Error("Value is not 0.5:", f)
		return
	}

	var s string
	err = conn.QueryRow("select cast(-0.05 as numeric(25,2))").Scan(&s)
	if err != nil {
		t.Error("query row / scan failed:", err.Error())
		return
	}
	if s != "-0.05" {
		t.Error("Value is not -0.05:", s)
		return
	}
}

func TestAffectedRows(t *testing.T) {
	conn := open(t)
	defer conn.Close()

	tx, err := conn.Begin()
	if err != nil {
		t.Fatal("Begin tran failed", err)
	}
	defer tx.Rollback()

	res, err := tx.Exec("create table #foo (bar int)")
	if err != nil {
		t.Fatal("create table failed")
	}
	n, err := res.RowsAffected()
	if err != nil {
		t.Fatal("rows affected failed")
	}
	if n != 0 {
		t.Error("Expected 0 rows affected, got ", n)
	}

	res, err = tx.Exec("insert into #foo (bar) values (1)")
	if err != nil {
		t.Fatal("insert failed")
	}
	n, err = res.RowsAffected()
	if err != nil {
		t.Fatal("rows affected failed")
	}
	if n != 1 {
		t.Error("Expected 1 row affected, got ", n)
	}

	res, err = tx.Exec("insert into #foo (bar) values (?)", 2)
	if err != nil {
		t.Fatal("insert failed")
	}
	n, err = res.RowsAffected()
	if err != nil {
		t.Fatal("rows affected failed")
	}
	if n != 1 {
		t.Error("Expected 1 row affected, got ", n)
	}
}

func TestIdentity(t *testing.T) {
	conn := open(t)
	defer conn.Close()

	tx, err := conn.Begin()
	if err != nil {
		t.Fatal("Begin tran failed", err)
	}
	defer tx.Rollback()

	res, err := tx.Exec("create table #foo (bar int identity, baz int unique)")
	if err != nil {
		t.Fatal("create table failed")
	}

	res, err = tx.Exec("insert into #foo (baz) values (1)")
	if err != nil {
		t.Fatal("insert failed")
	}
	n, err := res.LastInsertId()
	if err != nil {
		t.Fatal("last insert id failed")
	}
	if n != 1 {
		t.Error("Expected 1 for identity, got ", n)
	}

	res, err = tx.Exec("insert into #foo (baz) values (20)")
	if err != nil {
		t.Fatal("insert failed")
	}
	n, err = res.LastInsertId()
	if err != nil {
		t.Fatal("last insert id failed")
	}
	if n != 2 {
		t.Error("Expected 2 for identity, got ", n)
	}

	res, err = tx.Exec("insert into #foo (baz) values (1)")
	if err == nil {
		t.Fatal("insert should fail")
	}

	res, err = tx.Exec("insert into #foo (baz) values (?)", 1)
	if err == nil {
		t.Fatal("insert should fail")
	}
}

func TestDateTimeParam(t *testing.T) {
	conn := open(t)
	defer conn.Close()
	type testStruct struct {
		t time.Time
	}
	values := []testStruct{
		{time.Date(2004, 6, 3, 12, 13, 14, 150000000, time.UTC)},
		{time.Date(4, 6, 3, 12, 13, 14, 150000000, time.UTC)},
	}
	for _, test := range values {
		var t2 time.Time
		err := conn.QueryRow("select ?", test.t).Scan(&t2)
		if err != nil {
			t.Error("select / scan failed", err.Error())
			continue
		}
		if test.t.Sub(t2) != 0 {
			t.Errorf("datetime does not match: '%s' '%s' delta: %d", test.t, t2, test.t.Sub(t2))
		}
	}

}

func TestUniqueIdentifierParam(t *testing.T) {
	conn := open(t)
	defer conn.Close()
	type testStruct struct {
		name string
		uuid interface{}
	}

	expected := UniqueIdentifier{0x01, 0x23, 0x45, 0x67,
		0x89, 0xAB,
		0xCD, 0xEF,
		0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF,
	}

	values := []testStruct{
		{
			"[]byte",
			[]byte{0x67, 0x45, 0x23, 0x01,
				0xAB, 0x89,
				0xEF, 0xCD,
				0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF}},
		{
			"string",
			"01234567-89ab-cdef-0123-456789abcdef"},
	}

	for _, test := range values {
		t.Run(test.name, func(t *testing.T) {
			var uuid2 UniqueIdentifier
			err := conn.QueryRow("select ?", test.uuid).Scan(&uuid2)
			if err != nil {
				t.Fatal("select / scan failed", err.Error())
			}

			if expected != uuid2 {
				t.Errorf("uniqueidentifier does not match: '%s' '%s'", expected, uuid2)
			}
		})
	}
}

func TestBigQuery(t *testing.T) {
	conn := open(t)
	defer conn.Close()
	rows, err := conn.Query(`WITH n(n) AS
		(
		    SELECT 1
		    UNION ALL
		    SELECT n+1 FROM n WHERE n < 10000
		)
		SELECT n, @@version FROM n ORDER BY n
		OPTION (MAXRECURSION 10000);`)
	if err != nil {
		t.Fatal("cannot exec query", err)
	}
	rows.Next()
	rows.Close()
	var res int
	err = conn.QueryRow("select 0").Scan(&res)
	if err != nil {
		t.Fatal("cannot scan value", err)
	}
	if res != 0 {
		t.Fatal("expected 0, got ", res)
	}
}

func TestBug32(t *testing.T) {
	conn := open(t)
	defer conn.Close()

	tx, err := conn.Begin()
	if err != nil {
		t.Fatal("Begin tran failed", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec("if (exists(select * from INFORMATION_SCHEMA.TABLES where TABLE_NAME='tbl')) drop table tbl")
	if err != nil {
		t.Fatal("Drop table failed", err)
	}

	_, err = tx.Exec("create table tbl(a int primary key,fld bit null)")
	if err != nil {
		t.Fatal("Create table failed", err)
	}

	_, err = tx.Exec("insert into tbl (a,fld) values (1,nullif(?, ''))", "")
	if err != nil {
		t.Fatal("Insert failed", err)
	}
}

func TestIgnoreEmptyResults(t *testing.T) {
	conn := open(t)
	defer conn.Close()
	rows, err := conn.Query("set nocount on; select 2")
	if err != nil {
		t.Fatal("Query failed", err.Error())
	}
	if !rows.Next() {
		t.Fatal("Query didn't return row")
	}
	var fld1 int32
	err = rows.Scan(&fld1)
	if err != nil {
		t.Fatal("Scan failed", err)
	}
	if fld1 != 2 {
		t.Fatal("Returned value doesn't match")
	}
}

func TestMssqlStmt_SetQueryNotification(t *testing.T) {
	checkConnStr(t)
	mssqldriver := driverWithProcess(t)
	cn, err := mssqldriver.Open(makeConnStr(t).String())
	stmt, err := cn.Prepare("SELECT 1")
	if err != nil {
		t.Error("Connection failed", err)
	}

	sqlstmt := stmt.(*MssqlStmt)
	sqlstmt.SetQueryNotification("ABC", "service=WebCacheNotifications", time.Hour)

	rows, err := sqlstmt.Query(nil)
	if err == nil {
		rows.Close()
	}
	// notifications are sent to Service Broker
	// see for more info: https://github.com/denisenkom/go-mssqldb/pull/90
}

func TestErrorInfo(t *testing.T) {
	conn := open(t)
	defer conn.Close()

	_, err := conn.Exec("select bad")
	if sqlError, ok := err.(Error); ok {
		if sqlError.SQLErrorNumber() != 207 /*invalid column name*/ {
			t.Errorf("Query failed with unexpected error number %d %s", sqlError.SQLErrorNumber(), sqlError.SQLErrorMessage())
		}
	} else {
		t.Error("Failed to convert error to SQLErorr", err)
	}
}

func TestSetLanguage(t *testing.T) {
	conn := open(t)
	defer conn.Close()

	_, err := conn.Exec("set language russian")
	if err != nil {
		t.Errorf("Query failed with unexpected error %s", err)
	}

	row := conn.QueryRow("select cast(getdate() as varchar(50))")
	var val interface{}
	err = row.Scan(&val)
	if err != nil {
		t.Errorf("Query failed with unexpected error %s", err)
	}
	t.Log("Returned value", val)
}

func TestConnectionClosing(t *testing.T) {
	conn := open(t)
	defer conn.Close()
	for i := 1; i <= 100; i++ {
		if conn.Stats().OpenConnections > 1 {
			t.Errorf("Open connections is expected to stay <= 1, but it is %d", conn.Stats().OpenConnections)
			return
		}

		stmt, err := conn.Query("select 1")
		if err != nil {
			t.Errorf("Query failed with unexpected error %s", err)
		}
		for stmt.Next() {
			var val interface{}
			err := stmt.Scan(&val)
			if err != nil {
				t.Errorf("Query failed with unexpected error %s", err)
			}
		}
	}
}

func TestBeginTranError(t *testing.T) {
	checkConnStr(t)
	drv := driverWithProcess(t)
	conn, err := drv.open(makeConnStr(t).String())
	if err != nil {
		t.Fatalf("Open failed with error %v", err)
	}

	defer conn.Close()
	// close actual connection to make begin transaction to fail during sending of a packet
	conn.sess.buf.transport.Close()

	ctx := context.Background()
	_, err = conn.begin(ctx, isolationSnapshot)
	if err != driver.ErrBadConn {
		t.Errorf("begin should fail with ErrBadConn but it returned %v", err)
	}

	// reopen connection
	conn, err = drv.open(makeConnStr(t).String())
	if err != nil {
		t.Fatalf("Open failed with error %v", err)
	}
	err = conn.sendBeginRequest(ctx, isolationSerializable)
	if err != nil {
		t.Fatalf("sendBeginRequest failed with error %v", err)
	}

	// close connection to cause processBeginResponse to fail
	conn.sess.buf.transport.Close()
	_, err = conn.processBeginResponse(ctx)
	switch err {
	case nil:
		t.Error("processBeginResponse should fail but it succeeded")
	case driver.ErrBadConn:
		t.Error("processBeginResponse should fail with error different from ErrBadConn but it did")
	}

	if conn.connectionGood {
		t.Fatal("Connection should be in a bad state")
	}
}

func TestCommitTranError(t *testing.T) {
	checkConnStr(t)
	drv := driverWithProcess(t)
	conn, err := drv.open(makeConnStr(t).String())
	if err != nil {
		t.Fatalf("Open failed with error %v", err)
	}

	defer conn.Close()
	// close actual connection to make commit transaction to fail during sending of a packet
	conn.sess.buf.transport.Close()

	ctx := context.Background()
	err = conn.Commit()
	if err != driver.ErrBadConn {
		t.Errorf("begin should fail with ErrBadConn but it returned %v", err)
	}

	// reopen connection
	conn, err = drv.open(makeConnStr(t).String())
	if err != nil {
		t.Fatalf("Open failed with error %v", err)
	}
	err = conn.sendCommitRequest()
	if err != nil {
		t.Fatalf("sendCommitRequest failed with error %v", err)
	}

	// close connection to cause processBeginResponse to fail
	conn.sess.buf.transport.Close()
	err = conn.simpleProcessResp(ctx)
	switch err {
	case nil:
		t.Error("simpleProcessResp should fail but it succeeded")
	case driver.ErrBadConn:
		t.Error("simpleProcessResp should fail with error different from ErrBadConn but it did")
	}

	if conn.connectionGood {
		t.Fatal("Connection should be in a bad state")
	}

	// reopen connection
	conn, err = drv.open(makeConnStr(t).String())
	defer conn.Close()
	if err != nil {
		t.Fatalf("Open failed with error %v", err)
	}
	// should fail because there is no transaction
	err = conn.Commit()
	switch err {
	case nil:
		t.Error("Commit should fail but it succeeded")
	case driver.ErrBadConn:
		t.Error("Commit should fail with error different from ErrBadConn but it did")
	}
}

func TestRollbackTranError(t *testing.T) {
	checkConnStr(t)
	drv := driverWithProcess(t)
	conn, err := drv.open(makeConnStr(t).String())
	if err != nil {
		t.Fatalf("Open failed with error %v", err)
	}

	defer conn.Close()
	// close actual connection to make commit transaction to fail during sending of a packet
	conn.sess.buf.transport.Close()

	ctx := context.Background()
	err = conn.Rollback()
	if err != driver.ErrBadConn {
		t.Errorf("Rollback should fail with ErrBadConn but it returned %v", err)
	}

	// reopen connection
	conn, err = drv.open(makeConnStr(t).String())
	if err != nil {
		t.Fatalf("Open failed with error %v", err)
	}
	err = conn.sendRollbackRequest()
	if err != nil {
		t.Fatalf("sendCommitRequest failed with error %v", err)
	}

	// close connection to cause processBeginResponse to fail
	conn.sess.buf.transport.Close()
	err = conn.simpleProcessResp(ctx)
	switch err {
	case nil:
		t.Error("simpleProcessResp should fail but it succeeded")
	case driver.ErrBadConn:
		t.Error("simpleProcessResp should fail with error different from ErrBadConn but it did")
	}

	if conn.connectionGood {
		t.Fatal("Connection should be in a bad state")
	}

	// reopen connection
	conn, err = drv.open(makeConnStr(t).String())
	defer conn.Close()
	if err != nil {
		t.Fatalf("Open failed with error %v", err)
	}
	// should fail because there is no transaction
	err = conn.Rollback()
	switch err {
	case nil:
		t.Error("Commit should fail but it succeeded")
	case driver.ErrBadConn:
		t.Error("Commit should fail with error different from ErrBadConn but it did")
	}
}

func TestSendQueryErrors(t *testing.T) {
	checkConnStr(t)
	drv := driverWithProcess(t)
	conn, err := drv.open(makeConnStr(t).String())
	if err != nil {
		t.FailNow()
	}

	defer conn.Close()
	stmt, err := conn.prepareContext(context.Background(), "select 1")
	if err != nil {
		t.FailNow()
	}

	// should fail because parameter is invalid
	_, err = stmt.Query([]driver.Value{conn})
	if err == nil {
		t.Fail()
	}

	// close actual connection to make commit transaction to fail during sending of a packet
	conn.sess.buf.transport.Close()

	// should fail because connection is closed
	_, err = stmt.Query([]driver.Value{})
	if err != driver.ErrBadConn {
		t.Fail()
	}

	stmt, err = conn.prepareContext(context.Background(), "select ?")
	if err != nil {
		t.FailNow()
	}
	// should fail because connection is closed
	_, err = stmt.Query([]driver.Value{int64(1)})
	if err != driver.ErrBadConn {
		t.Fail()
	}
}

func TestProcessQueryErrors(t *testing.T) {
	checkConnStr(t)
	drv := driverWithProcess(t)
	conn, err := drv.open(makeConnStr(t).String())
	if err != nil {
		t.Fatal("open expected to succeed, but it failed with", err)
	}
	stmt, err := conn.prepareContext(context.Background(), "select 1")
	if err != nil {
		t.Fatal("prepareContext expected to succeed, but it failed with", err)
	}
	err = stmt.sendQuery([]namedValue{})
	if err != nil {
		t.Fatal("sendQuery expected to succeed, but it failed with", err)
	}
	// close actual connection to make reading response to fail
	conn.sess.buf.transport.Close()
	_, err = stmt.processQueryResponse(context.Background())
	if err == nil {
		t.Error("processQueryResponse expected to fail but it succeeded")
	}
	// should not fail with ErrBadConn because query was successfully sent to server
	if err == driver.ErrBadConn {
		t.Error("processQueryResponse expected to fail with error other than ErrBadConn but it failed with it")
	}

	if conn.connectionGood {
		t.Fatal("Connection should be in a bad state")
	}
}

func TestSendExecErrors(t *testing.T) {
	checkConnStr(t)
	drv := driverWithProcess(t)
	conn, err := drv.open(makeConnStr(t).String())
	if err != nil {
		t.FailNow()
	}

	defer conn.Close()
	stmt, err := conn.prepareContext(context.Background(), "select 1")
	if err != nil {
		t.FailNow()
	}

	// should fail because parameter is invalid
	_, err = stmt.Exec([]driver.Value{conn})
	if err == nil {
		t.Fail()
	}

	// close actual connection to make commit transaction to fail during sending of a packet
	conn.sess.buf.transport.Close()

	// should fail because connection is closed
	_, err = stmt.Exec([]driver.Value{})
	if err != driver.ErrBadConn {
		t.Fail()
	}

	stmt, err = conn.prepareContext(context.Background(), "select ?")
	if err != nil {
		t.FailNow()
	}
	// should fail because connection is closed
	_, err = stmt.Exec([]driver.Value{int64(1)})
	if err != driver.ErrBadConn {
		t.Fail()
	}
}
