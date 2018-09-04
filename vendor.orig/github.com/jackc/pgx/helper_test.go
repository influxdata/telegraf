package pgx_test

import (
	"testing"

	"github.com/jackc/pgx"
)

func mustConnect(t testing.TB, config pgx.ConnConfig) *pgx.Conn {
	conn, err := pgx.Connect(config)
	if err != nil {
		t.Fatalf("Unable to establish connection: %v", err)
	}
	return conn
}

func mustReplicationConnect(t testing.TB, config pgx.ConnConfig) *pgx.ReplicationConn {
	conn, err := pgx.ReplicationConnect(config)
	if err != nil {
		t.Fatalf("Unable to establish connection: %v", err)
	}
	return conn
}

func closeConn(t testing.TB, conn *pgx.Conn) {
	err := conn.Close()
	if err != nil {
		t.Fatalf("conn.Close unexpectedly failed: %v", err)
	}
}

func closeReplicationConn(t testing.TB, conn *pgx.ReplicationConn) {
	err := conn.Close()
	if err != nil {
		t.Fatalf("conn.Close unexpectedly failed: %v", err)
	}
}

func mustExec(t testing.TB, conn *pgx.Conn, sql string, arguments ...interface{}) (commandTag pgx.CommandTag) {
	var err error
	if commandTag, err = conn.Exec(sql, arguments...); err != nil {
		t.Fatalf("Exec unexpectedly failed with %v: %v", sql, err)
	}
	return
}

// Do a simple query to ensure the connection is still usable
func ensureConnValid(t *testing.T, conn *pgx.Conn) {
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
