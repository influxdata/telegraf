// +build go1.8

package mssql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"net"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestNextResultSet(t *testing.T) {
	conn := open(t)
	defer conn.Close()
	rows, err := conn.Query("select 1; select 2")
	if err != nil {
		t.Fatal("Query failed", err.Error())
	}
	defer func() {
		err := rows.Err()
		if err != nil {
			t.Error("unexpected error:", err)
		}
	}()

	defer rows.Close()

	if !rows.Next() {
		t.Fatal("Query didn't return row")
	}
	var fld1, fld2 int32
	err = rows.Scan(&fld1)
	if err != nil {
		t.Fatal("Scan failed", err)
	}
	if fld1 != 1 {
		t.Fatal("Returned value doesn't match")
	}
	if rows.Next() {
		t.Fatal("Query returned unexpected second row.")
	}
	// calling next again should still return false
	if rows.Next() {
		t.Fatal("Query returned unexpected second row.")
	}
	if !rows.NextResultSet() {
		t.Fatal("NextResultSet should return true but returned false")
	}
	if !rows.Next() {
		t.Fatal("Query didn't return row")
	}
	err = rows.Scan(&fld2)
	if err != nil {
		t.Fatal("Scan failed", err)
	}
	if fld2 != 2 {
		t.Fatal("Returned value doesn't match")
	}
	if rows.NextResultSet() {
		t.Fatal("NextResultSet should return false but returned true")
	}
}

func TestColumnTypeIntrospection(t *testing.T) {
	type tst struct {
		expr         string
		typeName     string
		reflType     reflect.Type
		hasSize      bool
		size         int64
		hasPrecScale bool
		precision    int64
		scale        int64
	}
	tests := []tst{
		{"cast(1 as bit)", "BIT", reflect.TypeOf(true), false, 0, false, 0, 0},
		{"cast(1 as tinyint)", "TINYINT", reflect.TypeOf(int64(0)), false, 0, false, 0, 0},
		{"cast(1 as smallint)", "SMALLINT", reflect.TypeOf(int64(0)), false, 0, false, 0, 0},
		{"1", "INT", reflect.TypeOf(int64(0)), false, 0, false, 0, 0},
		{"cast(1 as bigint)", "BIGINT", reflect.TypeOf(int64(0)), false, 0, false, 0, 0},
		{"cast(1 as real)", "REAL", reflect.TypeOf(0.0), false, 0, false, 0, 0},
		{"cast(1 as float)", "FLOAT", reflect.TypeOf(0.0), false, 0, false, 0, 0},
		{"cast('abc' as varbinary(3))", "VARBINARY", reflect.TypeOf([]byte{}), true, 3, false, 0, 0},
		{"cast('abc' as varbinary(max))", "VARBINARY", reflect.TypeOf([]byte{}), true, 2147483645, false, 0, 0},
		{"cast(1 as datetime)", "DATETIME", reflect.TypeOf(time.Time{}), false, 0, false, 0, 0},
		{"cast(1 as smalldatetime)", "SMALLDATETIME", reflect.TypeOf(time.Time{}), false, 0, false, 0, 0},
		{"cast(getdate() as datetime2(7))", "DATETIME2", reflect.TypeOf(time.Time{}), false, 0, false, 0, 0},
		{"cast(getdate() as datetimeoffset(7))", "DATETIMEOFFSET", reflect.TypeOf(time.Time{}), false, 0, false, 0, 0},
		{"cast(getdate() as date)", "DATE", reflect.TypeOf(time.Time{}), false, 0, false, 0, 0},
		{"cast(getdate() as time)", "TIME", reflect.TypeOf(time.Time{}), false, 0, false, 0, 0},
		{"'abc'", "VARCHAR", reflect.TypeOf(""), true, 3, false, 0, 0},
		{"cast('abc' as varchar(max))", "VARCHAR", reflect.TypeOf(""), true, 2147483645, false, 0, 0},
		{"N'abc'", "NVARCHAR", reflect.TypeOf(""), true, 3, false, 0, 0},
		{"cast(N'abc' as NVARCHAR(MAX))", "NVARCHAR", reflect.TypeOf(""), true, 1073741822, false, 0, 0},
		{"cast(1 as decimal)", "DECIMAL", reflect.TypeOf([]byte{}), false, 0, true, 18, 0},
		{"cast(1 as decimal(5, 2))", "DECIMAL", reflect.TypeOf([]byte{}), false, 0, true, 5, 2},
		{"cast(1 as numeric(10, 4))", "DECIMAL", reflect.TypeOf([]byte{}), false, 0, true, 10, 4},
		{"cast(1 as money)", "MONEY", reflect.TypeOf([]byte{}), false, 0, false, 0, 0},
		{"cast(1 as smallmoney)", "SMALLMONEY", reflect.TypeOf([]byte{}), false, 0, false, 0, 0},
		{"cast(0x6F9619FF8B86D011B42D00C04FC964FF as uniqueidentifier)", "UNIQUEIDENTIFIER", reflect.TypeOf([]byte{}), false, 0, false, 0, 0},
		{"cast('<root/>' as xml)", "XML", reflect.TypeOf(""), true, 1073741822, false, 0, 0},
		{"cast('abc' as text)", "TEXT", reflect.TypeOf(""), true, 2147483647, false, 0, 0},
		{"cast(N'abc' as ntext)", "NTEXT", reflect.TypeOf(""), true, 1073741823, false, 0, 0},
		{"cast('abc' as image)", "IMAGE", reflect.TypeOf([]byte{}), true, 2147483647, false, 0, 0},
		{"cast('abc' as char(3))", "CHAR", reflect.TypeOf(""), true, 3, false, 0, 0},
		{"cast(N'abc' as nchar(3))", "NCHAR", reflect.TypeOf(""), true, 3, false, 0, 0},
		{"cast(1 as sql_variant)", "SQL_VARIANT", reflect.TypeOf(nil), false, 0, false, 0, 0},
	}
	conn := open(t)
	defer conn.Close()
	for _, tt := range tests {
		rows, err := conn.Query("select " + tt.expr)
		if err != nil {
			t.Errorf("Query failed with unexpected error %s", err)
		}
		ct, err := rows.ColumnTypes()
		if err != nil {
			t.Errorf("Query failed with unexpected error %s", err)
		}
		if ct[0].DatabaseTypeName() != tt.typeName {
			t.Errorf("Expected type %s but returned %s", tt.typeName, ct[0].DatabaseTypeName())
		}
		size, ok := ct[0].Length()
		if ok != tt.hasSize {
			t.Errorf("Expected has size %v but returned %v for %s", tt.hasSize, ok, tt.expr)
		} else {
			if ok && size != tt.size {
				t.Errorf("Expected size %d but returned %d for %s", tt.size, size, tt.expr)
			}
		}

		prec, scale, ok := ct[0].DecimalSize()
		if ok != tt.hasPrecScale {
			t.Errorf("Expected has prec/scale %v but returned %v for %s", tt.hasPrecScale, ok, tt.expr)
		} else {
			if ok && prec != tt.precision {
				t.Errorf("Expected precision %d but returned %d for %s", tt.precision, prec, tt.expr)
			}
			if ok && scale != tt.scale {
				t.Errorf("Expected scale %d but returned %d for %s", tt.scale, scale, tt.expr)
			}
		}

		if ct[0].ScanType() != tt.reflType {
			t.Errorf("Expected ScanType %v but got %v for %s", tt.reflType, ct[0].ScanType(), tt.expr)
		}
	}
}

func TestColumnIntrospection(t *testing.T) {
	type tst struct {
		expr         string
		fieldName    string
		typeName     string
		nullable     bool
		hasSize      bool
		size         int64
		hasPrecScale bool
		precision    int64
		scale        int64
	}
	tests := []tst{
		{"f1 int null", "f1", "INT", true, false, 0, false, 0, 0},
		{"f2 varchar(15) not null", "f2", "VARCHAR", false, true, 15, false, 0, 0},
		{"f3 decimal(5, 2) null", "f3", "DECIMAL", true, false, 0, true, 5, 2},
	}
	conn := open(t)
	defer conn.Close()

	// making table variable with specified fields and making a select from it
	exprs := make([]string, len(tests))
	for i, test := range tests {
		exprs[i] = test.expr
	}
	exprJoined := strings.Join(exprs, ",")
	rows, err := conn.Query(fmt.Sprintf("declare @tbl table(%s); select * from @tbl", exprJoined))
	if err != nil {
		t.Errorf("Query failed with unexpected error %s", err)
	}

	ct, err := rows.ColumnTypes()
	if err != nil {
		t.Errorf("ColumnTypes failed with unexpected error %s", err)
	}
	for i, test := range tests {
		if ct[i].Name() != test.fieldName {
			t.Errorf("Field expected have name %s but it has name %s", test.fieldName, ct[i].Name())
		}

		if ct[i].DatabaseTypeName() != test.typeName {
			t.Errorf("Invalid type name returned %s expected %s", ct[i].DatabaseTypeName(), test.typeName)
		}

		nullable, ok := ct[i].Nullable()
		if ok {
			if nullable != test.nullable {
				t.Errorf("Invalid nullable value returned %v", nullable)
			}
		} else {
			t.Error("Nullable was expected to support Nullable but it didn't")
		}

		size, ok := ct[i].Length()
		if ok != test.hasSize {
			t.Errorf("Expected has size %v but returned %v for %s", test.hasSize, ok, test.expr)
		} else {
			if ok && size != test.size {
				t.Errorf("Expected size %d but returned %d for %s", test.size, size, test.expr)
			}
		}

		prec, scale, ok := ct[i].DecimalSize()
		if ok != test.hasPrecScale {
			t.Errorf("Expected has prec/scale %v but returned %v for %s", test.hasPrecScale, ok, test.expr)
		} else {
			if ok && prec != test.precision {
				t.Errorf("Expected precision %d but returned %d for %s", test.precision, prec, test.expr)
			}
			if ok && scale != test.scale {
				t.Errorf("Expected scale %d but returned %d for %s", test.scale, scale, test.expr)
			}
		}
	}
}

func TestContext(t *testing.T) {
	conn := open(t)
	defer conn.Close()

	opts := &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	}
	ctx := context.Background()
	tx, err := conn.BeginTx(ctx, opts)
	if err != nil {
		t.Errorf("BeginTx failed with unexpected error %s", err)
		return
	}
	rows, err := tx.QueryContext(ctx, "DBCC USEROPTIONS")
	properties := make(map[string]string)
	for rows.Next() {
		var name, value string
		if err = rows.Scan(&name, &value); err != nil {
			t.Errorf("Scan failed with unexpected error %s", err)
		}
		properties[name] = value
	}

	if properties["isolation level"] != "serializable" {
		t.Errorf("Expected isolation level to be serializable but it is %s", properties["isolation level"])
	}

	row := tx.QueryRowContext(ctx, "select 1")
	var val int64
	if err = row.Scan(&val); err != nil {
		t.Errorf("QueryRowContext failed with unexpected error %s", err)
	}
	if val != 1 {
		t.Error("Incorrect value returned from query")
	}

	_, err = tx.ExecContext(ctx, "select 1")
	if err != nil {
		t.Errorf("ExecContext failed with unexpected error %s", err)
		return
	}

	_, err = tx.PrepareContext(ctx, "select 1")
	if err != nil {
		t.Errorf("PrepareContext failed with unexpected error %s", err)
		return
	}
}

func TestBeginTxtReadOnlyNotSupported(t *testing.T) {
	conn := open(t)
	defer conn.Close()
	opts := &sql.TxOptions{ReadOnly: true}
	_, err := conn.BeginTx(context.Background(), opts)
	if err == nil {
		t.Error("BeginTx expected to fail for read only transaction because MSSQL doesn't support it, but it succeeded")
	}
}

func TestMssqlConn_BeginTx(t *testing.T) {
	conn := open(t)
	defer conn.Close()
	_, err := conn.Exec("create table test (f int)")
	defer conn.Exec("drop table test")
	if err != nil {
		t.Fatal("create table failed with error", err)
	}

	tx1, err := conn.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal("BeginTx failed with error", err)
	}
	tx2, err := conn.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal("BeginTx failed with error", err)
	}
	_, err = tx1.Exec("insert into test (f) values (1)")
	if err != nil {
		t.Fatal("insert failed with error", err)
	}
	_, err = tx2.Exec("insert into test (f) values (2)")
	if err != nil {
		t.Fatal("insert failed with error", err)
	}
	tx1.Rollback()
	tx2.Commit()

	rows, err := conn.Query("select f from test")
	if err != nil {
		t.Fatal("select failed with error", err)
	}
	values := []int64{}
	for rows.Next() {
		var val int64
		err = rows.Scan(&val)
		if err != nil {
			t.Fatal("scan failed with error", err)
		}
		values = append(values, val)
	}
	if !reflect.DeepEqual(values, []int64{2}) {
		t.Errorf("Values is expected to be [1] but it is %v", values)
	}
}

func TestNamedParameters(t *testing.T) {
	conn := open(t)
	defer conn.Close()
	row := conn.QueryRow(
		"select :param2, :param1, :param2",
		sql.Named("param1", 1),
		sql.Named("param2", 2))
	var col1, col2, col3 int64
	err := row.Scan(&col1, &col2, &col3)
	if err != nil {
		t.Errorf("Scan failed with unexpected error %s", err)
		return
	}
	if col1 != 2 || col2 != 1 || col3 != 2 {
		t.Errorf("Unexpected values returned col1=%d, col2=%d, col3=%d", col1, col2, col3)
	}
}

func TestBadNamedParameters(t *testing.T) {
	conn := open(t)
	defer conn.Close()
	row := conn.QueryRow(
		"select :param2, :param1, :param2",
		sql.Named("badparam1", 1),
		sql.Named("param2", 2))
	var col1, col2, col3 int64
	err := row.Scan(&col1, &col2, &col3)
	if err == nil {
		t.Error("Scan succeeded unexpectedly")
		return
	}
	t.Logf("Scan failed as expected with error %s", err)
}

func TestMixedParameters(t *testing.T) {
	conn := open(t)
	defer conn.Close()
	row := conn.QueryRow(
		"select :2, :param1, :param2",
		5, // this parameter will be unused
		6,
		sql.Named("param1", 1),
		sql.Named("param2", 2))
	var col1, col2, col3 int64
	err := row.Scan(&col1, &col2, &col3)
	if err != nil {
		t.Errorf("Scan failed with unexpected error %s", err)
		return
	}
	if col1 != 6 || col2 != 1 || col3 != 2 {
		t.Errorf("Unexpected values returned col1=%d, col2=%d, col3=%d", col1, col2, col3)
	}
}

/*
func TestMixedParametersExample(t *testing.T) {
	conn := open(t)
	defer conn.Close()
	row := conn.QueryRow(
		"select :id, ?",
		sql.Named("id", 1),
		2,
		)
	var col1, col2 int64
	err := row.Scan(&col1, &col2)
	if err != nil {
		t.Errorf("Scan failed with unexpected error %s", err)
		return
	}
	if col1 != 1 || col2 != 2 {
		t.Errorf("Unexpected values returned col1=%d, col2=%d", col1, col2)
	}
}
*/

func TestPinger(t *testing.T) {
	conn := open(t)
	defer conn.Close()
	err := conn.Ping()
	if err != nil {
		t.Errorf("Failed to hit database")
	}
}

func TestQueryCancelLowLevel(t *testing.T) {
	checkConnStr(t)
	drv := driverWithProcess(t)
	conn, err := drv.open(makeConnStr(t).String())
	if err != nil {
		t.Fatalf("Open failed with error %v", err)
	}

	defer conn.Close()
	ctx, cancel := context.WithCancel(context.Background())
	stmt, err := conn.prepareContext(ctx, "waitfor delay '00:00:03'")
	if err != nil {
		t.Fatalf("Prepare failed with error %v", err)
	}
	err = stmt.sendQuery([]namedValue{})
	if err != nil {
		t.Fatalf("sendQuery failed with error %v", err)
	}

	cancel()

	_, err = stmt.processExec(ctx)
	if err != context.Canceled {
		t.Errorf("Expected error to be Cancelled but got %v", err)
	}

	// same connection should be usable again after it was cancelled
	stmt, err = conn.prepareContext(context.Background(), "select 1")
	if err != nil {
		t.Fatalf("Prepare failed with error %v", err)
	}
	rows, err := stmt.Query([]driver.Value{})
	if err != nil {
		t.Fatalf("Query failed with error %v", err)
	}

	values := []driver.Value{nil}
	err = rows.Next(values)
	if err != nil {
		t.Fatalf("Next failed with error %v", err)
	}
}

func TestQueryCancelHighLevel(t *testing.T) {
	conn := open(t)
	defer conn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(200 * time.Millisecond)
		cancel()
	}()
	_, err := conn.ExecContext(ctx, "waitfor delay '00:00:03'")
	if err != context.Canceled {
		t.Errorf("ExecContext expected to fail with Cancelled but it returned %v", err)
	}

	// connection should be usable after timeout
	row := conn.QueryRow("select 1")
	var val int64
	err = row.Scan(&val)
	if err != nil {
		t.Fatal("Scan failed with", err)
	}
}

func TestQueryTimeout(t *testing.T) {
	conn := open(t)
	defer conn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	_, err := conn.ExecContext(ctx, "waitfor delay '00:00:03'")
	if err != context.DeadlineExceeded {
		t.Errorf("ExecContext expected to fail with DeadlineExceeded but it returned %v", err)
	}

	// connection should be usable after timeout
	row := conn.QueryRow("select 1")
	var val int64
	err = row.Scan(&val)
	if err != nil {
		t.Fatal("Scan failed with", err)
	}
}

func TestDriverParams(t *testing.T) {
	checkConnStr(t)
	SetLogger(testLogger{t})
	type sqlCmd struct {
		Name   string
		Driver string
		Query  string
		Param  []interface{}
		Expect []interface{}
	}

	list := []sqlCmd{
		{
			Name:   "preprocess-ordinal",
			Driver: "mssql",
			Query:  `select V1=:1`,
			Param:  []interface{}{"abc"},
			Expect: []interface{}{"abc"},
		},
		{
			Name:   "preprocess-name",
			Driver: "mssql",
			Query:  `select V1=:First`,
			Param:  []interface{}{sql.Named("First", "abc")},
			Expect: []interface{}{"abc"},
		},
		{
			Name:   "raw-ordinal",
			Driver: "sqlserver",
			Query:  `select V1=@p1`,
			Param:  []interface{}{"abc"},
			Expect: []interface{}{"abc"},
		},
		{
			Name:   "raw-name",
			Driver: "sqlserver",
			Query:  `select V1=@First`,
			Param:  []interface{}{sql.Named("First", "abc")},
			Expect: []interface{}{"abc"},
		},
	}

	for cmdIndex, cmd := range list {
		t.Run(cmd.Name, func(t *testing.T) {
			db, err := sql.Open(cmd.Driver, makeConnStr(t).String())
			if err != nil {
				t.Fatalf("failed to open driver %q", cmd.Driver)
			}
			defer db.Close()

			rows, err := db.Query(cmd.Query, cmd.Param...)
			if err != nil {
				t.Fatalf("failed to run query %q %v", cmd.Query, err)
			}
			defer rows.Close()

			columns, err := rows.Columns()
			if err != nil {
				t.Fatal("failed to get column schema %v", err)
			}
			clen := len(columns)

			if clen != len(cmd.Expect) {
				t.Fatal("query column has %d, expect %d columns", clen, len(cmd.Expect))
			}

			values := make([]interface{}, clen)
			into := make([]interface{}, clen)
			for i := 0; i < clen; i++ {
				into[i] = &values[i]
			}
			for rows.Next() {
				err = rows.Scan(into...)
				if err != nil {
					t.Fatalf("failed to scan into row for %d %q", cmdIndex, cmd.Driver)
				}
				for i := range cmd.Expect {
					if values[i] != cmd.Expect[i] {
						t.Fatal("expected value in index %d %v != actual value %v", i, cmd.Expect[i], values[i])
					}
				}
			}
		})
	}
}

type connInterrupt struct {
	net.Conn

	mu           sync.Mutex
	disruptRead  bool
	disruptWrite bool
}

func (c *connInterrupt) Interrupt(write bool) {
	c.mu.Lock()
	if write {
		c.disruptWrite = true
	} else {
		c.disruptRead = true
	}
	c.mu.Unlock()
}

func (c *connInterrupt) Read(b []byte) (n int, err error) {
	c.mu.Lock()
	dis := c.disruptRead
	c.mu.Unlock()
	if dis {
		return 0, disconnectError{}
	}
	return c.Conn.Read(b)
}

func (c *connInterrupt) Write(b []byte) (n int, err error) {
	c.mu.Lock()
	dis := c.disruptWrite
	c.mu.Unlock()
	if dis {
		return 0, disconnectError{}
	}
	return c.Conn.Write(b)
}

type dialerInterrupt struct {
	nd tcpDialer

	mu   sync.Mutex
	list []*connInterrupt
}

func (d *dialerInterrupt) Dial(addr string) (net.Conn, error) {
	conn, err := d.nd.Dial(addr)
	if err != nil {
		return nil, err
	}
	ci := &connInterrupt{Conn: conn}
	d.mu.Lock()
	d.list = append(d.list, ci)
	d.mu.Unlock()
	return ci, err
}

func (d *dialerInterrupt) Interrupt(write bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	for _, ci := range d.list {
		ci.Interrupt(write)
	}
}

var _ net.Error = disconnectError{}

type disconnectError struct{}

func (disconnectError) Error() string {
	return "disconnect"
}

func (disconnectError) Timeout() bool {
	return true
}

func (disconnectError) Temporary() bool {
	return true
}

// TestDisconnect1 ensures errors and states are handled correctly if
// the server is disconnected mid-query.
func TestDisconnect1(t *testing.T) {
	if testing.Short() {
		t.Skip("short")
	}
	checkConnStr(t)
	SetLogger(testLogger{t})

	// Revert to the normal dialer after the test is done.
	normalCreateDialer := createDialer
	defer func() {
		createDialer = normalCreateDialer
	}()

	waitDisrupt := make(chan struct{})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	createDialer = func(p *connectParams) dialer {
		nd := tcpDialer{&net.Dialer{Timeout: p.dial_timeout, KeepAlive: p.keepAlive}}
		di := &dialerInterrupt{nd: nd}
		go func() {
			<-waitDisrupt
			di.Interrupt(true)
			di.Interrupt(false)
		}()
		return di
	}
	db, err := sql.Open("sqlserver", makeConnStr(t).String())
	if err != nil {
		t.Fatal(err)
	}

	if err := db.PingContext(ctx); err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.ExecContext(ctx, `SET LOCK_TIMEOUT 1800;`)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		time.Sleep(time.Second * 1)
		close(waitDisrupt)
	}()
	t.Log("prepare for query")
	_, err = db.ExecContext(ctx, `waitfor delay '00:00:3';`)
	if err != nil {
		t.Log("expected error after disconnect", err)
		return
	}
	t.Fatal("wanted error after Exec")
}

// TestDisconnect2 tests a read error so the query is started
// but results cannot be read.
func TestDisconnect2(t *testing.T) {
	if testing.Short() {
		t.Skip("short")
	}
	checkConnStr(t)
	SetLogger(testLogger{t})

	// Revert to the normal dialer after the test is done.
	normalCreateDialer := createDialer
	defer func() {
		createDialer = normalCreateDialer
	}()

	end := make(chan error)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		waitDisrupt := make(chan struct{})
		ctx, cancel = context.WithTimeout(ctx, time.Second*2)
		defer cancel()

		createDialer = func(p *connectParams) dialer {
			nd := tcpDialer{&net.Dialer{Timeout: p.dial_timeout, KeepAlive: p.keepAlive}}
			di := &dialerInterrupt{nd: nd}
			go func() {
				<-waitDisrupt
				di.Interrupt(false)
			}()
			return di
		}
		db, err := sql.Open("sqlserver", makeConnStr(t).String())
		if err != nil {
			t.Fatal(err)
		}

		if err := db.PingContext(ctx); err != nil {
			t.Fatal(err)
		}
		defer db.Close()

		_, err = db.ExecContext(ctx, `SET LOCK_TIMEOUT 1800;`)
		if err != nil {
			t.Fatal(err)
		}
		close(waitDisrupt)

		_, err = db.ExecContext(ctx, `waitfor delay '00:00:3';`)
		end <- err
	}()

	timeout := time.After(10 * time.Second)
	select {
	case err := <-end:
		if err == nil {
			t.Fatal("test err")
		}
	case <-timeout:
		t.Fatal("timeout")
	}
}

// TestAzureDatabase tests reading from an encrypted azure database connection
// from a SELECT read.
func TestAzureDatabase(t *testing.T) {
	t.Skip("currently failing")

	query := `
;with
    config_cte (config) as (
            select *
                    from ( values
                    ('_partition:{\"Fill\":{\"PatternType\":\"solid\",\"FgColor\":\"99ff99\"}}')
                    , ('_separation:{\"Fill\":{\"PatternType\":\"solid\",\"FgColor\":\"99ffff\"}}')
                    , ('Monthly Earnings:\$#,##0.00 ;(\$#,##0.00)')
                    , ('Weekly Earnings:\$#,##0.00 ;(\$#,##0.00)')
                    , ('Total Earnings:\$#,##0.00 ;(\$#,##0.00)')
                    , ('Average Earnings:\$#,##0.00 ;(\$#,##0.00)')
                    , ('Last Month Earning:#,##0.00 ;(#,##0.00)')
                    , ('Award:\$#,##0.00 ;(\$#,##0.00)')
                    , ('Amount:\$#,##0.00 ;(\$#,##0.00)')
                    , ('Grand Total:\$#,##0.00 ;(\$#,##0.00)')
                    , ('Total:\$#,##0.00 ;(\$#,##0.00)')
                    , ('Price Each:\$#,##0.00 ;(\$#,##0.00)')
                    , ('Hyperwallet:\$#,##0.00 ;(\$#,##0.00)')
                    , ('Credit/Debit:\$#,##0.00 ;(\$#,##0.00)')
                    , ('Earning:#,##0.00 ;(#,##0.00)')
                    , ('Change Earning:#,##0.00 ;(#,##0.00)')
                    , ('CheckAmount:#,##0.00 ;(#,##0.00)')
                    , ('Residual:#,##0.00 ;(#,##0.00)')
                    , ('Prev Residual:#,##0.00 ;(#,##0.00)')
                    , ('Team Bonuses:#,##0.00 ;(#,##0.00)')
                    , ('Change:#,##0.00 ;(#,##0.00)')
                    , ('Shipping Total:#,##0.00 ;(#,##0.00)')
                    , ('SubTotal:\$#,##0.00 ;(\$#,##0.00)')
                    , ('Total Diff:#,##0.00 ;(#,##0.00)')
                    , ('SubTotal Diff:#,##0.00 ;(#,##0.00)')
                    , ('Return Total:#,##0.00 ;(#,##0.00)')
                    , ('Return SubTotal:#,##0.00 ;(#,##0.00)')
                    , ('Return Total Diff:#,##0.00 ;(#,##0.00)')
                    , ('Return SubTotal Diff:#,##0.00 ;(#,##0.00)')
                    , ('Cancel Total:#,##0.00 ;(#,##0.00)')
                    , ('Cancel SubTotal:#,##0.00 ;(#,##0.00)')
                    , ('Cancel Total Diff:#,##0.00 ;(#,##0.00)')
                    , ('Cancel SubTotal Diff:#,##0.00 ;(#,##0.00)')
                    , ('Replacement Total:#,##0.00 ;(#,##0.00)')
                    , ('Replacement SubTotal:#,##0.00 ;(#,##0.00)')
                    , ('Replacement Total Diff:#,##0.00 ;(#,##0.00)')
                    , ('Replacement SubTotal Diff:#,##0.00 ;(#,##0.00)')
                    , ('Jan Residual:#,##0.00 ;(#,##0.00)')
                    , ('Jan Bonus:#,##0.00 ;(#,##0.00)')
                    , ('Jan Total:#,##0.00 ;(#,##0.00)')
                    , ('January Residual:#,##0.00 ;(#,##0.00)')
                    , ('Feb Residual:#,##0.00 ;(#,##0.00)')
                    , ('Feb Bonus:#,##0.00 ;(#,##0.00)')
                    , ('Feb Total:#,##0.00 ;(#,##0.00)')
                    , ('February Residual:#,##0.00 ;(#,##0.00)')
                    , ('Mar Residual:#,##0.00 ;(#,##0.00)')
                    , ('Mar Bonus:#,##0.00 ;(#,##0.00)')
                    , ('Mar Total:#,##0.00 ;(#,##0.00)')
                    , ('March Residual:#,##0.00 ;(#,##0.00)')
                    , ('Apr Residual:#,##0.00 ;(#,##0.00)')
                    , ('Apr Bonus:#,##0.00 ;(#,##0.00)')
                    , ('Apr Total:#,##0.00 ;(#,##0.00)')
                    , ('April Residual:#,##0.00 ;(#,##0.00)')
                    , ('May Residual:#,##0.00 ;(#,##0.00)')
                    , ('May Bonus:#,##0.00 ;(#,##0.00)')
                    , ('May Total:#,##0.00 ;(#,##0.00)')
                    , ('Jun Residual:#,##0.00 ;(#,##0.00)')
                    , ('Jun Bonus:#,##0.00 ;(#,##0.00)')
                    , ('Jun Total:#,##0.00 ;(#,##0.00)')
                    , ('June Residual:#,##0.00 ;(#,##0.00)')
                    , ('Jul Residual:#,##0.00 ;(#,##0.00)')
                    , ('Jul Bonus:#,##0.00 ;(#,##0.00)')
                    , ('Jul Total:#,##0.00 ;(#,##0.00)')
                    , ('July Residual:#,##0.00 ;(#,##0.00)')
                    , ('Aug Residual:#,##0.00 ;(#,##0.00)')
                    , ('Aug Bonus:#,##0.00 ;(#,##0.00)')
                    , ('Aug Total:#,##0.00 ;(#,##0.00)')
                    , ('August Residual:#,##0.00 ;(#,##0.00)')
                    , ('Sep Residual:#,##0.00 ;(#,##0.00)')
                    , ('Sep Bonus:#,##0.00 ;(#,##0.00)')
                    , ('Sep Total:#,##0.00 ;(#,##0.00)')
                    , ('September Residual:#,##0.00 ;(#,##0.00)')
                    , ('Oct Residual:#,##0.00 ;(#,##0.00)')
                    , ('Oct Bonus:#,##0.00 ;(#,##0.00)')
                    , ('Oct Total:#,##0.00 ;(#,##0.00)')
                    , ('October Residual:#,##0.00 ;(#,##0.00)')
                    , ('Nov Residual:#,##0.00 ;(#,##0.00)')
                    , ('Nov Bonus:#,##0.00 ;(#,##0.00)')
                    , ('Nov Total:#,##0.00 ;(#,##0.00)')
                    , ('November Residual:#,##0.00 ;(#,##0.00)')
                    , ('Dec Residual:#,##0.00 ;(#,##0.00)')
                    , ('Dec Bonus:#,##0.00 ;(#,##0.00)')
                    , ('Dec Total:#,##0.00 ;(#,##0.00)')
                    , ('December Residual:#,##0.00 ;(#,##0.00)')
                    , ('January Bonus:#,##0.00 ;(#,##0.00)')
                    , ('February Bonus:#,##0.00 ;(#,##0.00)')
                    , ('March Bonus:#,##0.00 ;(#,##0.00)')
                    , ('April Bonus:#,##0.00 ;(#,##0.00)')
                    , ('May Bonus:#,##0.00 ;(#,##0.00)')
                    , ('June Bonus:#,##0.00 ;(#,##0.00)')
                    , ('July Bonus:#,##0.00 ;(#,##0.00)')
                    , ('August Bonus:#,##0.00 ;(#,##0.00)')
                    , ('September Bonus:#,##0.00 ;(#,##0.00)')
                    , ('October Bonus:#,##0.00 ;(#,##0.00)')
                    , ('November Bonus:#,##0.00 ;(#,##0.00)')
                    , ('December Bonus:#,##0.00 ;(#,##0.00)')
                    , ('January Adj:#,##0.00 ;(#,##0.00)')
                    , ('February Adj:#,##0.00 ;(#,##0.00)')
                    , ('March Adj:#,##0.00 ;(#,##0.00)')
                    , ('April Adj:#,##0.00 ;(#,##0.00)')
                    , ('May Adj:#,##0.00 ;(#,##0.00)')
                    , ('June Adj:#,##0.00 ;(#,##0.00)')
                    , ('July Adj:#,##0.00 ;(#,##0.00)')
                    , ('August Adj:#,##0.00 ;(#,##0.00)')
                    , ('September Adj:#,##0.00 ;(#,##0.00)')
                    , ('October Adj:#,##0.00 ;(#,##0.00)')
                    , ('November Adj:#,##0.00 ;(#,##0.00)')
                    , ('December Adj:#,##0.00 ;(#,##0.00)')
                    , ('2016- 2015 YTD Dif:#,##0.00 ;(#,##0.00)')
                    , ('2017- 2016 YTD Dif:#,##0.00 ;(#,##0.00)')
                    , ('2018- 2017 YTD Dif:#,##0.00 ;(#,##0.00)')
                    , ('Dec to Jan Dif Residual:#,##0.00 ;(#,##0.00)')
                    , ('Jan to Feb Dif Residual:#,##0.00 ;(#,##0.00)')
                    , ('Feb to Mar Dif Residual:#,##0.00 ;(#,##0.00)')
                    , ('Mar to Apr Dif Residual:#,##0.00 ;(#,##0.00)')
                    , ('Apr to May Dif Residual:#,##0.00 ;(#,##0.00)')
                    , ('May to Jun Dif Residual:#,##0.00 ;(#,##0.00)')
                    , ('Jun to Jul Dif Residual:#,##0.00 ;(#,##0.00)')
                    , ('Jul to Aug Dif Residual:#,##0.00 ;(#,##0.00)')
                    , ('Aug to Sep Dif Residual:#,##0.00 ;(#,##0.00)')
                    , ('Sep to Oct Dif Residual:#,##0.00 ;(#,##0.00)')
                    , ('Oct to Nov Dif Residual:#,##0.00 ;(#,##0.00)')
                    , ('Nov to Dec Dif Residual:#,##0.00 ;(#,##0.00)')
                    , ('Dec to Jan Dif Bonus:#,##0.00 ;(#,##0.00)')
                    , ('Jan to Feb Dif Bonus:#,##0.00 ;(#,##0.00)')
                    , ('Feb to Mar Dif Bonus:#,##0.00 ;(#,##0.00)')
                    , ('Mar to Apr Dif Bonus:#,##0.00 ;(#,##0.00)')
                    , ('Apr to May Dif Bonus:#,##0.00 ;(#,##0.00)')
                    , ('May to Jun Dif Bonus:#,##0.00 ;(#,##0.00)')
                    , ('Jun to Jul Dif Bonus:#,##0.00 ;(#,##0.00)')
                    , ('Jul to Aug Dif Bonus:#,##0.00 ;(#,##0.00)')
                    , ('Aug to Sep Dif Bonus:#,##0.00 ;(#,##0.00)')
                    , ('Sep to Oct Dif Bonus:#,##0.00 ;(#,##0.00)')
                    , ('Oct to Nov Dif Bonus:#,##0.00 ;(#,##0.00)')
                    , ('Nov to Dec Dif Bonus:#,##0.00 ;(#,##0.00)')
                    , ('Dec to Jan Dif Total:#,##0.00 ;(#,##0.00)')
                    , ('Jan to Feb Dif Total:#,##0.00 ;(#,##0.00)')
                    , ('Feb to Mar Dif Total:#,##0.00 ;(#,##0.00)')
                    , ('Mar to Apr Dif Total:#,##0.00 ;(#,##0.00)')
                    , ('Apr to May Dif Total:#,##0.00 ;(#,##0.00)')
                    , ('May to Jun Dif Total:#,##0.00 ;(#,##0.00)')
                    , ('Jun to Jul Dif Total:#,##0.00 ;(#,##0.00)')
                    , ('Jul to Aug Dif Total:#,##0.00 ;(#,##0.00)')
                    , ('Aug to Sep Dif Total:#,##0.00 ;(#,##0.00)')
                    , ('Sep to Oct Dif Total:#,##0.00 ;(#,##0.00)')
                    , ('Oct to Nov Dif Total:#,##0.00 ;(#,##0.00)')
                    , ('Nov to Dec Dif Total:#,##0.00 ;(#,##0.00)')
                    , ('Jan Refund Cnt:#,##0 ;(#,##0)')
                    , ('Feb Refund Cnt:#,##0 ;(#,##0)')
                    , ('Mar Refund Cnt:#,##0 ;(#,##0)')
                    , ('Apr Refund Cnt:#,##0 ;(#,##0)')
                    , ('May Refund Cnt:#,##0 ;(#,##0)')
                    , ('Jun Refund Cnt:#,##0 ;(#,##0)')
                    , ('Jul Refund Cnt:#,##0 ;(#,##0)')
                    , ('Aug Refund Cnt:#,##0 ;(#,##0)')
                    , ('Sep Refund Cnt:#,##0 ;(#,##0)')
                    , ('Oct Refund Cnt:#,##0 ;(#,##0)')
                    , ('Nov Refund Cnt:#,##0 ;(#,##0)')
                    , ('Dec Refund Cnt:#,##0 ;(#,##0)')
                    , ('Jan Purchase Cnt:#,##0 ;(#,##0)')
                    , ('Feb Purchase Cnt:#,##0 ;(#,##0)')
                    , ('Mar Purchase Cnt:#,##0 ;(#,##0)')
                    , ('Apr Purchase Cnt:#,##0 ;(#,##0)')
                    , ('May Purchase Cnt:#,##0 ;(#,##0)')
                    , ('Jun Purchase Cnt:#,##0 ;(#,##0)')
                    , ('Jul Purchase Cnt:#,##0 ;(#,##0)')
                    , ('Aug Purchase Cnt:#,##0 ;(#,##0)')
                    , ('Sep Purchase Cnt:#,##0 ;(#,##0)')
                    , ('Oct Purchase Cnt:#,##0 ;(#,##0)')
                    , ('Nov Purchase Cnt:#,##0 ;(#,##0)')
                    , ('Dec Purchase Cnt:#,##0 ;(#,##0)')
                    , ('Jan Refund Amt:#,##0.00 ;(#,##0.00)')
                    , ('Feb Refund Amt:#,##0.00 ;(#,##0.00)')
                    , ('Mar Refund Amt:#,##0.00 ;(#,##0.00)')
                    , ('Apr Refund Amt:#,##0.00 ;(#,##0.00)')
                    , ('May Refund Amt:#,##0.00 ;(#,##0.00)')
                    , ('Jun Refund Amt:#,##0.00 ;(#,##0.00)')
                    , ('Jul Refund Amt:#,##0.00 ;(#,##0.00)')
                    , ('Aug Refund Amt:#,##0.00 ;(#,##0.00)')
                    , ('Sep Refund Amt:#,##0.00 ;(#,##0.00)')
                    , ('Oct Refund Amt:#,##0.00 ;(#,##0.00)')
                    , ('Nov Refund Amt:#,##0.00 ;(#,##0.00)')
                    , ('Dec Refund Amt:#,##0.00 ;(#,##0.00)')
                    , ('Jan Purchase Amt:#,##0.00 ;(#,##0.00)')
                    , ('Feb Purchase Amt:#,##0.00 ;(#,##0.00)')
                    , ('Mar Purchase Amt:#,##0.00 ;(#,##0.00)')
                    , ('Apr Purchase Amt:#,##0.00 ;(#,##0.00)')
                    , ('May Purchase Amt:#,##0.00 ;(#,##0.00)')
                    , ('Jun Purchase Amt:#,##0.00 ;(#,##0.00)')
                    , ('Jul Purchase Amt:#,##0.00 ;(#,##0.00)')
                    , ('Aug Purchase Amt:#,##0.00 ;(#,##0.00)')
                    , ('Sep Purchase Amt:#,##0.00 ;(#,##0.00)')
                    , ('Oct Purchase Amt:#,##0.00 ;(#,##0.00)')
                    , ('Nov Purchase Amt:#,##0.00 ;(#,##0.00)')
                    , ('Dec Purchase Amt:#,##0.00 ;(#,##0.00)')
                    ) X(a))
    select * from config_cte
	`

	db := open(t)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var s string
	for rows.Next() {
		err = rows.Scan(&s)
		if err != nil {
			t.Fatal(err)
		}
	}
}
