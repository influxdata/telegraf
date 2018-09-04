package mssql

import (
	"bytes"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"golang.org/x/net/context"
)

type MockTransport struct {
	bytes.Buffer
}

func (t *MockTransport) Close() error {
	return nil
}

func TestSendLogin(t *testing.T) {
	memBuf := new(MockTransport)
	buf := newTdsBuffer(1024, memBuf)
	login := login{
		TDSVersion:     verTDS73,
		PacketSize:     0x1000,
		ClientProgVer:  0x01060100,
		ClientPID:      100,
		ClientTimeZone: -4 * 60,
		ClientID:       [6]byte{0x12, 0x34, 0x56, 0x78, 0x90, 0xab},
		OptionFlags1:   0xe0,
		OptionFlags3:   8,
		HostName:       "subdev1",
		UserName:       "test",
		Password:       "testpwd",
		AppName:        "appname",
		ServerName:     "servername",
		CtlIntName:     "library",
		Language:       "en",
		Database:       "database",
		ClientLCID:     0x204,
		AtchDBFile:     "filepath",
	}
	err := sendLogin(buf, login)
	if err != nil {
		t.Error("sendLogin should succeed")
	}
	ref := []byte{
		16, 1, 0, 222, 0, 0, 1, 0, 198 + 16, 0, 0, 0, 3, 0, 10, 115, 0, 16, 0, 0, 0, 1,
		6, 1, 100, 0, 0, 0, 0, 0, 0, 0, 224, 0, 0, 8, 16, 255, 255, 255, 4, 2, 0,
		0, 94, 0, 7, 0, 108, 0, 4, 0, 116, 0, 7, 0, 130, 0, 7, 0, 144, 0, 10, 0, 0,
		0, 0, 0, 164, 0, 7, 0, 178, 0, 2, 0, 182, 0, 8, 0, 18, 52, 86, 120, 144, 171,
		198, 0, 0, 0, 198, 0, 8, 0, 214, 0, 0, 0, 0, 0, 0, 0, 115, 0, 117, 0, 98,
		0, 100, 0, 101, 0, 118, 0, 49, 0, 116, 0, 101, 0, 115, 0, 116, 0, 226, 165,
		243, 165, 146, 165, 226, 165, 162, 165, 210, 165, 227, 165, 97, 0, 112,
		0, 112, 0, 110, 0, 97, 0, 109, 0, 101, 0, 115, 0, 101, 0, 114, 0, 118, 0,
		101, 0, 114, 0, 110, 0, 97, 0, 109, 0, 101, 0, 108, 0, 105, 0, 98, 0, 114,
		0, 97, 0, 114, 0, 121, 0, 101, 0, 110, 0, 100, 0, 97, 0, 116, 0, 97, 0, 98,
		0, 97, 0, 115, 0, 101, 0, 102, 0, 105, 0, 108, 0, 101, 0, 112, 0, 97, 0,
		116, 0, 104, 0}
	out := memBuf.Bytes()
	if !bytes.Equal(ref, out) {
		fmt.Println("Expected:")
		fmt.Print(hex.Dump(ref))
		fmt.Println("Returned:")
		fmt.Print(hex.Dump(out))
		t.Error("input output don't match")
	}
}

func TestSendSqlBatch(t *testing.T) {
	checkConnStr(t)
	p, err := parseConnectParams(makeConnStr(t).String())
	if err != nil {
		t.Error("parseConnectParams failed:", err.Error())
		return
	}

	conn, err := connect(optionalLogger{testLogger{t}}, p)
	if err != nil {
		t.Error("Open connection failed:", err.Error())
		return
	}
	defer conn.buf.transport.Close()

	headers := []headerStruct{
		{hdrtype: dataStmHdrTransDescr,
			data: transDescrHdr{0, 1}.pack()},
	}
	err = sendSqlBatch72(conn.buf, "select 1", headers)
	if err != nil {
		t.Error("Sending sql batch failed", err.Error())
		return
	}

	ch := make(chan tokenStruct, 5)
	go processResponse(context.Background(), conn, ch)

	var lastRow []interface{}
loop:
	for tok := range ch {
		switch token := tok.(type) {
		case doneStruct:
			break loop
		case []columnStruct:
			conn.columns = token
		case []interface{}:
			lastRow = token
		default:
			fmt.Println("unknown token", tok)
		}
	}

	if len(lastRow) == 0 {
		t.Fatal("expected row but no row set")
	}

	switch value := lastRow[0].(type) {
	case int32:
		if value != 1 {
			t.Error("Invalid value returned, should be 1", value)
			return
		}
	}
}

func checkConnStr(t *testing.T) {
	if len(os.Getenv("SQLSERVER_DSN")) > 0 {
		return
	}
	if len(os.Getenv("HOST")) > 0 && len(os.Getenv("DATABASE")) > 0 {
		return
	}
	t.Skip("no database connection string")
}

// makeConnStr returns a URL struct so it may be modified by various
// tests before used as a DSN.
func makeConnStr(t *testing.T) *url.URL {
	dsn := os.Getenv("SQLSERVER_DSN")
	if len(dsn) > 0 {
		parsed, err := url.Parse(dsn)
		if err != nil {
			t.Fatal("unable to parse SQLSERVER_DSN as URL", err)
		}
		values := parsed.Query()
		values.Set("log", "127")
		parsed.RawQuery = values.Encode()
		return parsed
	}
	values := url.Values{}
	values.Set("log", "127")
	values.Set("database", os.Getenv("DATABASE"))
	return &url.URL{
		Scheme:   "sqlserver",
		Host:     os.Getenv("HOST"),
		Path:     os.Getenv("INSTANCE"),
		User:     url.UserPassword(os.Getenv("SQLUSER"), os.Getenv("SQLPASSWORD")),
		RawQuery: values.Encode(),
	}
}

type testLogger struct {
	t *testing.T
}

func (l testLogger) Printf(format string, v ...interface{}) {
	l.t.Logf(format, v...)
}

func (l testLogger) Println(v ...interface{}) {
	l.t.Log(v...)
}

func open(t *testing.T) *sql.DB {
	checkConnStr(t)
	SetLogger(testLogger{t})
	conn, err := sql.Open("mssql", makeConnStr(t).String())
	if err != nil {
		t.Error("Open connection failed:", err.Error())
		return nil
	}

	return conn
}

func TestConnect(t *testing.T) {
	checkConnStr(t)
	SetLogger(testLogger{t})
	conn, err := sql.Open("mssql", makeConnStr(t).String())
	if err != nil {
		t.Error("Open connection failed:", err.Error())
		return
	}
	defer conn.Close()
}

func simpleQuery(conn *sql.DB, t *testing.T) (stmt *sql.Stmt) {
	stmt, err := conn.Prepare("select 1 as a")
	if err != nil {
		t.Error("Prepare failed:", err.Error())
		return nil
	}
	return stmt
}

func checkSimpleQuery(rows *sql.Rows, t *testing.T) {
	numrows := 0
	for rows.Next() {
		var val int
		err := rows.Scan(&val)
		if err != nil {
			t.Error("Scan failed:", err.Error())
		}
		if val != 1 {
			t.Error("query should return 1")
		}
		numrows++
	}
	if numrows != 1 {
		t.Error("query should return 1 row, returned", numrows)
	}
}

func TestQuery(t *testing.T) {
	conn := open(t)
	if conn == nil {
		return
	}
	defer conn.Close()

	stmt := simpleQuery(conn, t)
	if stmt == nil {
		return
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		t.Error("Query failed:", err.Error())
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		t.Error("getting columns failed", err.Error())
	}
	if len(columns) != 1 && columns[0] != "a" {
		t.Error("returned incorrect columns (expected ['a']):", columns)
	}

	checkSimpleQuery(rows, t)
}

func TestMultipleQueriesSequentialy(t *testing.T) {

	conn := open(t)
	defer conn.Close()

	stmt, err := conn.Prepare("select 1 as a")
	if err != nil {
		t.Error("Prepare failed:", err.Error())
		return
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		t.Error("Query failed:", err.Error())
		return
	}
	defer rows.Close()
	checkSimpleQuery(rows, t)

	rows, err = stmt.Query()
	if err != nil {
		t.Error("Query failed:", err.Error())
		return
	}
	defer rows.Close()
	checkSimpleQuery(rows, t)
}

func TestMultipleQueryClose(t *testing.T) {
	conn := open(t)
	defer conn.Close()

	stmt, err := conn.Prepare("select 1 as a")
	if err != nil {
		t.Error("Prepare failed:", err.Error())
		return
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		t.Error("Query failed:", err.Error())
		return
	}
	rows.Close()

	rows, err = stmt.Query()
	if err != nil {
		t.Error("Query failed:", err.Error())
		return
	}
	defer rows.Close()
	checkSimpleQuery(rows, t)
}

func TestPing(t *testing.T) {
	conn := open(t)
	defer conn.Close()
	conn.Ping()
}

func TestSecureWithInvalidHostName(t *testing.T) {
	checkConnStr(t)
	SetLogger(testLogger{t})

	dsn := makeConnStr(t)
	dsnParams := dsn.Query()
	dsnParams.Set("encrypt", "true")
	dsnParams.Set("TrustServerCertificate", "false")
	dsnParams.Set("hostNameInCertificate", "foo.bar")
	dsn.RawQuery = dsnParams.Encode()

	conn, err := sql.Open("mssql", dsn.String())
	if err != nil {
		t.Fatal("Open connection failed:", err.Error())
	}
	defer conn.Close()
	err = conn.Ping()
	if err == nil {
		t.Fatal("Connected to fake foo.bar server")
	}
}

func TestSecureConnection(t *testing.T) {
	checkConnStr(t)
	SetLogger(testLogger{t})

	dsn := makeConnStr(t)
	dsnParams := dsn.Query()
	dsnParams.Set("encrypt", "true")
	dsnParams.Set("TrustServerCertificate", "true")
	dsn.RawQuery = dsnParams.Encode()

	conn, err := sql.Open("mssql", dsn.String())
	if err != nil {
		t.Fatal("Open connection failed:", err.Error())
	}
	defer conn.Close()
	var msg string
	err = conn.QueryRow("select 'secret'").Scan(&msg)
	if err != nil {
		t.Fatal("cannot scan value", err)
	}
	if msg != "secret" {
		t.Fatal("expected secret, got: ", msg)
	}
	var secure bool
	err = conn.QueryRow("select encrypt_option from sys.dm_exec_connections where session_id=@@SPID").Scan(&secure)
	if err != nil {
		t.Fatal("cannot scan value", err)
	}
	if !secure {
		t.Fatal("connection is not encrypted")
	}
}

func TestInvalidConnectionString(t *testing.T) {
	connStrings := []string{
		"log=invalid",
		"port=invalid",
		"packet size=invalid",
		"connection timeout=invalid",
		"dial timeout=invalid",
		"keepalive=invalid",
		"encrypt=invalid",
		"trustservercertificate=invalid",
		"failoverport=invalid",

		// ODBC mode
		"odbc:password={",
		"odbc:password={somepass",
		"odbc:password={somepass}}",
		"odbc:password={some}pass",
	}
	for _, connStr := range connStrings {
		_, err := parseConnectParams(connStr)
		if err == nil {
			t.Errorf("Connection expected to fail for connection string %s but it didn't", connStr)
			continue
		} else {
			t.Logf("Connection failed for %s as expected with error %v", connStr, err)
		}
	}
}

func TestValidConnectionString(t *testing.T) {
	type testStruct struct {
		connStr string
		check   func(connectParams) bool
	}
	connStrings := []testStruct{
		{"server=server\\instance;database=testdb;user id=tester;password=pwd", func(p connectParams) bool {
			return p.host == "server" && p.instance == "instance" && p.user == "tester" && p.password == "pwd"
		}},
		{"server=.", func(p connectParams) bool { return p.host == "localhost" }},
		{"server=(local)", func(p connectParams) bool { return p.host == "localhost" }},
		{"ServerSPN=serverspn;Workstation ID=workstid", func(p connectParams) bool { return p.serverSPN == "serverspn" && p.workstation == "workstid" }},
		{"failoverpartner=fopartner;failoverport=2000", func(p connectParams) bool { return p.failOverPartner == "fopartner" && p.failOverPort == 2000 }},
		{"app name=appname;applicationintent=ReadOnly", func(p connectParams) bool { return p.appname == "appname" && (p.typeFlags&fReadOnlyIntent != 0) }},
		{"encrypt=disable", func(p connectParams) bool { return p.disableEncryption }},
		{"encrypt=true", func(p connectParams) bool { return p.encrypt && !p.disableEncryption }},
		{"encrypt=false", func(p connectParams) bool { return !p.encrypt && !p.disableEncryption }},
		{"trustservercertificate=true", func(p connectParams) bool { return p.trustServerCertificate }},
		{"trustservercertificate=false", func(p connectParams) bool { return !p.trustServerCertificate }},
		{"certificate=abc", func(p connectParams) bool { return p.certificate == "abc" }},
		{"hostnameincertificate=abc", func(p connectParams) bool { return p.hostInCertificate == "abc" }},
		{"connection timeout=3;dial timeout=4;keepalive=5", func(p connectParams) bool {
			return p.conn_timeout == 3*time.Second && p.dial_timeout == 4*time.Second && p.keepAlive == 5*time.Second
		}},
		{"log=63", func(p connectParams) bool { return p.logFlags == 63 && p.port == 1433 }},
		{"log=63;port=1000", func(p connectParams) bool { return p.logFlags == 63 && p.port == 1000 }},
		{"log=64", func(p connectParams) bool { return p.logFlags == 64 && p.packetSize == 4096 }},
		{"log=64;packet size=0", func(p connectParams) bool { return p.logFlags == 64 && p.packetSize == 512 }},
		{"log=64;packet size=300", func(p connectParams) bool { return p.logFlags == 64 && p.packetSize == 512 }},
		{"log=64;packet size=8192", func(p connectParams) bool { return p.logFlags == 64 && p.packetSize == 8192 }},
		{"log=64;packet size=48000", func(p connectParams) bool { return p.logFlags == 64 && p.packetSize == 32767 }},

		// those are supported currently, but maybe should not be
		{"someparam", func(p connectParams) bool { return true }},
		{";;=;", func(p connectParams) bool { return true }},

		// ODBC mode
		{"odbc:server=somehost;user id=someuser;password=somepass", func(p connectParams) bool {
			return p.host == "somehost" && p.user == "someuser" && p.password == "somepass"
		}},
		{"odbc:server=somehost;user id=someuser;password=some{pass", func(p connectParams) bool {
			return p.host == "somehost" && p.user == "someuser" && p.password == "some{pass"
		}},
		{"odbc:server={somehost};user id={someuser};password={somepass}", func(p connectParams) bool {
			return p.host == "somehost" && p.user == "someuser" && p.password == "somepass"
		}},
		{"odbc:server={somehost};user id={someuser};password={some=pass}", func(p connectParams) bool {
			return p.host == "somehost" && p.user == "someuser" && p.password == "some=pass"
		}},
		{"odbc:server={somehost};user id={someuser};password={some;pass}", func(p connectParams) bool {
			return p.host == "somehost" && p.user == "someuser" && p.password == "some;pass"
		}},
		{"odbc:server={somehost};user id={someuser};password={some{pass}", func(p connectParams) bool {
			return p.host == "somehost" && p.user == "someuser" && p.password == "some{pass"
		}},
		{"odbc:server={somehost};user id={someuser};password={some}}pass}", func(p connectParams) bool {
			return p.host == "somehost" && p.user == "someuser" && p.password == "some}pass"
		}},
		{"odbc:server={somehost};user id={someuser};password={some{}}p=a;ss}", func(p connectParams) bool {
			return p.host == "somehost" && p.user == "someuser" && p.password == "some{}p=a;ss"
		}},
		{"odbc: server = somehost; user id =  someuser ; password = {some pass } ", func(p connectParams) bool {
			return p.host == "somehost" && p.user == "someuser" && p.password == "some pass "
		}},

		// URL mode
		{"sqlserver://somehost?connection+timeout=30", func(p connectParams) bool {
			return p.host == "somehost" && p.port == 1433 && p.instance == "" && p.conn_timeout == 30*time.Second
		}},
		{"sqlserver://someuser@somehost?connection+timeout=30", func(p connectParams) bool {
			return p.host == "somehost" && p.port == 1433 && p.instance == "" && p.user == "someuser" && p.password == "" && p.conn_timeout == 30*time.Second
		}},
		{"sqlserver://someuser:@somehost?connection+timeout=30", func(p connectParams) bool {
			return p.host == "somehost" && p.port == 1433 && p.instance == "" && p.user == "someuser" && p.password == "" && p.conn_timeout == 30*time.Second
		}},
		{"sqlserver://someuser:foo%3A%2F%5C%21~%40;bar@somehost?connection+timeout=30", func(p connectParams) bool {
			return p.host == "somehost" && p.port == 1433 && p.instance == "" && p.user == "someuser" && p.password == "foo:/\\!~@;bar" && p.conn_timeout == 30*time.Second
		}},
		{"sqlserver://someuser:foo%3A%2F%5C%21~%40;bar@somehost:1434?connection+timeout=30", func(p connectParams) bool {
			return p.host == "somehost" && p.port == 1434 && p.instance == "" && p.user == "someuser" && p.password == "foo:/\\!~@;bar" && p.conn_timeout == 30*time.Second
		}},
		{"sqlserver://someuser:foo%3A%2F%5C%21~%40;bar@somehost:1434/someinstance?connection+timeout=30", func(p connectParams) bool {
			return p.host == "somehost" && p.port == 1434 && p.instance == "someinstance" && p.user == "someuser" && p.password == "foo:/\\!~@;bar" && p.conn_timeout == 30*time.Second
		}},
	}
	for _, ts := range connStrings {
		p, err := parseConnectParams(ts.connStr)
		if err == nil {
			t.Logf("Connection string was parsed successfully %s", ts.connStr)
		} else {
			t.Errorf("Connection string %s failed to parse with error %s", ts.connStr, err)
			continue
		}

		if !ts.check(p) {
			t.Errorf("Check failed on conn str %s", ts.connStr)
		}
	}
}
