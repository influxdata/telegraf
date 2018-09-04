package pgx_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
)

func TestConnBeginBatch(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	sql := `create temporary table ledger(
  id serial primary key,
  description varchar not null,
  amount int not null
);`
	mustExec(t, conn, sql)

	batch := conn.BeginBatch()
	batch.Queue("insert into ledger(description, amount) values($1, $2)",
		[]interface{}{"q1", 1},
		[]pgtype.OID{pgtype.VarcharOID, pgtype.Int4OID},
		nil,
	)
	batch.Queue("insert into ledger(description, amount) values($1, $2)",
		[]interface{}{"q2", 2},
		[]pgtype.OID{pgtype.VarcharOID, pgtype.Int4OID},
		nil,
	)
	batch.Queue("insert into ledger(description, amount) values($1, $2)",
		[]interface{}{"q3", 3},
		[]pgtype.OID{pgtype.VarcharOID, pgtype.Int4OID},
		nil,
	)
	batch.Queue("select id, description, amount from ledger order by id",
		nil,
		nil,
		[]int16{pgx.BinaryFormatCode, pgx.TextFormatCode, pgx.BinaryFormatCode},
	)
	batch.Queue("select sum(amount) from ledger",
		nil,
		nil,
		[]int16{pgx.BinaryFormatCode},
	)

	err := batch.Send(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}

	ct, err := batch.ExecResults()
	if err != nil {
		t.Error(err)
	}
	if ct.RowsAffected() != 1 {
		t.Errorf("ct.RowsAffected() => %v, want %v", ct.RowsAffected(), 1)
	}

	ct, err = batch.ExecResults()
	if err != nil {
		t.Error(err)
	}
	if ct.RowsAffected() != 1 {
		t.Errorf("ct.RowsAffected() => %v, want %v", ct.RowsAffected(), 1)
	}

	rows, err := batch.QueryResults()
	if err != nil {
		t.Error(err)
	}

	var id int32
	var description string
	var amount int32
	if !rows.Next() {
		t.Fatal("expected a row to be available")
	}
	if err := rows.Scan(&id, &description, &amount); err != nil {
		t.Fatal(err)
	}
	if id != 1 {
		t.Errorf("id => %v, want %v", id, 1)
	}
	if description != "q1" {
		t.Errorf("description => %v, want %v", description, "q1")
	}
	if amount != 1 {
		t.Errorf("amount => %v, want %v", amount, 1)
	}

	if !rows.Next() {
		t.Fatal("expected a row to be available")
	}
	if err := rows.Scan(&id, &description, &amount); err != nil {
		t.Fatal(err)
	}
	if id != 2 {
		t.Errorf("id => %v, want %v", id, 2)
	}
	if description != "q2" {
		t.Errorf("description => %v, want %v", description, "q2")
	}
	if amount != 2 {
		t.Errorf("amount => %v, want %v", amount, 2)
	}

	if !rows.Next() {
		t.Fatal("expected a row to be available")
	}
	if err := rows.Scan(&id, &description, &amount); err != nil {
		t.Fatal(err)
	}
	if id != 3 {
		t.Errorf("id => %v, want %v", id, 3)
	}
	if description != "q3" {
		t.Errorf("description => %v, want %v", description, "q3")
	}
	if amount != 3 {
		t.Errorf("amount => %v, want %v", amount, 3)
	}

	if rows.Next() {
		t.Fatal("did not expect a row to be available")
	}

	if rows.Err() != nil {
		t.Fatal(rows.Err())
	}

	err = batch.QueryRowResults().Scan(&amount)
	if err != nil {
		t.Error(err)
	}
	if amount != 6 {
		t.Errorf("amount => %v, want %v", amount, 6)
	}

	err = batch.Close()
	if err != nil {
		t.Fatal(err)
	}

	ensureConnValid(t, conn)
}

func TestConnBeginBatchWithPreparedStatement(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	_, err := conn.Prepare("ps1", "select n from generate_series(0,$1::int) n")
	if err != nil {
		t.Fatal(err)
	}

	batch := conn.BeginBatch()

	queryCount := 3
	for i := 0; i < queryCount; i++ {
		batch.Queue("ps1",
			[]interface{}{5},
			nil,
			[]int16{pgx.BinaryFormatCode},
		)
	}

	err = batch.Send(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < queryCount; i++ {
		rows, err := batch.QueryResults()
		if err != nil {
			t.Fatal(err)
		}

		for k := 0; rows.Next(); k++ {
			var n int
			if err := rows.Scan(&n); err != nil {
				t.Fatal(err)
			}
			if n != k {
				t.Fatalf("n => %v, want %v", n, k)
			}
		}

		if rows.Err() != nil {
			t.Fatal(rows.Err())
		}
	}

	err = batch.Close()
	if err != nil {
		t.Fatal(err)
	}

	ensureConnValid(t, conn)
}

func TestConnBeginBatchContextCancelBeforeExecResults(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)

	sql := `create temporary table ledger(
  id serial primary key,
  description varchar not null,
  amount int not null
);`
	mustExec(t, conn, sql)

	batch := conn.BeginBatch()
	batch.Queue("insert into ledger(description, amount) values($1, $2)",
		[]interface{}{"q1", 1},
		[]pgtype.OID{pgtype.VarcharOID, pgtype.Int4OID},
		nil,
	)
	batch.Queue("select pg_sleep(2)",
		nil,
		nil,
		nil,
	)

	ctx, cancelFn := context.WithCancel(context.Background())

	err := batch.Send(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}

	cancelFn()

	_, err = batch.ExecResults()
	if err != context.Canceled {
		t.Errorf("err => %v, want %v", err, context.Canceled)
	}

	if conn.IsAlive() {
		t.Error("conn should be dead, but was alive")
	}
}

func TestConnBeginBatchContextCancelBeforeQueryResults(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)

	batch := conn.BeginBatch()
	batch.Queue("select pg_sleep(2)",
		nil,
		nil,
		nil,
	)
	batch.Queue("select pg_sleep(2)",
		nil,
		nil,
		nil,
	)

	ctx, cancelFn := context.WithCancel(context.Background())

	err := batch.Send(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}

	cancelFn()

	_, err = batch.QueryResults()
	if err != context.Canceled {
		t.Errorf("err => %v, want %v", err, context.Canceled)
	}

	if conn.IsAlive() {
		t.Error("conn should be dead, but was alive")
	}
}

func TestConnBeginBatchContextCancelBeforeFinish(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)

	batch := conn.BeginBatch()
	batch.Queue("select pg_sleep(2)",
		nil,
		nil,
		nil,
	)
	batch.Queue("select pg_sleep(2)",
		nil,
		nil,
		nil,
	)

	ctx, cancelFn := context.WithCancel(context.Background())

	err := batch.Send(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}

	cancelFn()

	err = batch.Close()
	if err != context.Canceled {
		t.Errorf("err => %v, want %v", err, context.Canceled)
	}

	if conn.IsAlive() {
		t.Error("conn should be dead, but was alive")
	}
}

func TestConnBeginBatchCloseRowsPartiallyRead(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	batch := conn.BeginBatch()
	batch.Queue("select n from generate_series(0,5) n",
		nil,
		nil,
		[]int16{pgx.BinaryFormatCode},
	)
	batch.Queue("select n from generate_series(0,5) n",
		nil,
		nil,
		[]int16{pgx.BinaryFormatCode},
	)

	err := batch.Send(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}

	rows, err := batch.QueryResults()
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < 3; i++ {
		if !rows.Next() {
			t.Error("expected a row to be available")
		}

		var n int
		if err := rows.Scan(&n); err != nil {
			t.Error(err)
		}
		if n != i {
			t.Errorf("n => %v, want %v", n, i)
		}
	}

	rows.Close()

	rows, err = batch.QueryResults()
	if err != nil {
		t.Error(err)
	}

	for i := 0; rows.Next(); i++ {
		var n int
		if err := rows.Scan(&n); err != nil {
			t.Error(err)
		}
		if n != i {
			t.Errorf("n => %v, want %v", n, i)
		}
	}

	if rows.Err() != nil {
		t.Error(rows.Err())
	}

	err = batch.Close()
	if err != nil {
		t.Fatal(err)
	}

	ensureConnValid(t, conn)
}

func TestConnBeginBatchQueryError(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	batch := conn.BeginBatch()
	batch.Queue("select n from generate_series(0,5) n where 100/(5-n) > 0",
		nil,
		nil,
		[]int16{pgx.BinaryFormatCode},
	)
	batch.Queue("select n from generate_series(0,5) n",
		nil,
		nil,
		[]int16{pgx.BinaryFormatCode},
	)

	err := batch.Send(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}

	rows, err := batch.QueryResults()
	if err != nil {
		t.Error(err)
	}

	for i := 0; rows.Next(); i++ {
		var n int
		if err := rows.Scan(&n); err != nil {
			t.Error(err)
		}
		if n != i {
			t.Errorf("n => %v, want %v", n, i)
		}
	}

	if pgErr, ok := rows.Err().(pgx.PgError); !(ok && pgErr.Code == "22012") {
		t.Errorf("rows.Err() => %v, want error code %v", rows.Err(), 22012)
	}

	err = batch.Close()
	if pgErr, ok := err.(pgx.PgError); !(ok && pgErr.Code == "22012") {
		t.Errorf("rows.Err() => %v, want error code %v", err, 22012)
	}

	if conn.IsAlive() {
		t.Error("conn should be dead, but was alive")
	}
}

func TestConnBeginBatchQuerySyntaxError(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	batch := conn.BeginBatch()
	batch.Queue("select 1 1",
		nil,
		nil,
		[]int16{pgx.BinaryFormatCode},
	)

	err := batch.Send(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}

	var n int32
	err = batch.QueryRowResults().Scan(&n)
	if pgErr, ok := err.(pgx.PgError); !(ok && pgErr.Code == "42601") {
		t.Errorf("rows.Err() => %v, want error code %v", err, 42601)
	}

	err = batch.Close()
	if err == nil {
		t.Error("Expected error")
	}

	if conn.IsAlive() {
		t.Error("conn should be dead, but was alive")
	}
}

func TestConnBeginBatchQueryRowInsert(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	sql := `create temporary table ledger(
  id serial primary key,
  description varchar not null,
  amount int not null
);`
	mustExec(t, conn, sql)

	batch := conn.BeginBatch()
	batch.Queue("select 1",
		nil,
		nil,
		[]int16{pgx.BinaryFormatCode},
	)
	batch.Queue("insert into ledger(description, amount) values($1, $2),($1, $2)",
		[]interface{}{"q1", 1},
		[]pgtype.OID{pgtype.VarcharOID, pgtype.Int4OID},
		nil,
	)

	err := batch.Send(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}

	var value int
	err = batch.QueryRowResults().Scan(&value)
	if err != nil {
		t.Error(err)
	}

	ct, err := batch.ExecResults()
	if err != nil {
		t.Error(err)
	}
	if ct.RowsAffected() != 2 {
		t.Errorf("ct.RowsAffected() => %v, want %v", ct.RowsAffected(), 2)
	}

	batch.Close()

	ensureConnValid(t, conn)
}

func TestConnBeginBatchQueryPartialReadInsert(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	sql := `create temporary table ledger(
  id serial primary key,
  description varchar not null,
  amount int not null
);`
	mustExec(t, conn, sql)

	batch := conn.BeginBatch()
	batch.Queue("select 1 union all select 2 union all select 3",
		nil,
		nil,
		[]int16{pgx.BinaryFormatCode},
	)
	batch.Queue("insert into ledger(description, amount) values($1, $2),($1, $2)",
		[]interface{}{"q1", 1},
		[]pgtype.OID{pgtype.VarcharOID, pgtype.Int4OID},
		nil,
	)

	err := batch.Send(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}

	rows, err := batch.QueryResults()
	if err != nil {
		t.Error(err)
	}
	rows.Close()

	ct, err := batch.ExecResults()
	if err != nil {
		t.Error(err)
	}
	if ct.RowsAffected() != 2 {
		t.Errorf("ct.RowsAffected() => %v, want %v", ct.RowsAffected(), 2)
	}

	batch.Close()

	ensureConnValid(t, conn)
}

func TestTxBeginBatch(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	sql := `create temporary table ledger1(
  id serial primary key,
  description varchar not null
);`
	mustExec(t, conn, sql)

	sql = `create temporary table ledger2(
  id int primary key,
  amount int not null
);`
	mustExec(t, conn, sql)

	tx, _ := conn.Begin()
	batch := tx.BeginBatch()
	batch.Queue("insert into ledger1(description) values($1) returning id",
		[]interface{}{"q1"},
		[]pgtype.OID{pgtype.VarcharOID},
		[]int16{pgx.BinaryFormatCode},
	)

	err := batch.Send(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	var id int
	err = batch.QueryRowResults().Scan(&id)
	if err != nil {
		t.Error(err)
	}
	batch.Close()

	batch = tx.BeginBatch()
	batch.Queue("insert into ledger2(id,amount) values($1, $2)",
		[]interface{}{id, 2},
		[]pgtype.OID{pgtype.Int4OID, pgtype.Int4OID},
		nil,
	)

	batch.Queue("select amount from ledger2 where id = $1",
		[]interface{}{id},
		[]pgtype.OID{pgtype.Int4OID},
		nil,
	)

	err = batch.Send(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	ct, err := batch.ExecResults()
	if err != nil {
		t.Error(err)
	}
	if ct.RowsAffected() != 1 {
		t.Errorf("ct.RowsAffected() => %v, want %v", ct.RowsAffected(), 1)
	}

	var amout int
	err = batch.QueryRowResults().Scan(&amout)
	if err != nil {
		t.Error(err)
	}

	batch.Close()
	tx.Commit()

	var count int
	conn.QueryRow("select count(1) from ledger1 where id = $1", id).Scan(&count)
	if count != 1 {
		t.Errorf("count => %v, want %v", count, 1)
	}

	err = batch.Close()
	if err != nil {
		t.Fatal(err)
	}

	ensureConnValid(t, conn)
}

func TestTxBeginBatchRollback(t *testing.T) {
	t.Parallel()

	conn := mustConnect(t, *defaultConnConfig)
	defer closeConn(t, conn)

	sql := `create temporary table ledger1(
  id serial primary key,
  description varchar not null
);`
	mustExec(t, conn, sql)

	tx, _ := conn.Begin()
	batch := tx.BeginBatch()
	batch.Queue("insert into ledger1(description) values($1) returning id",
		[]interface{}{"q1"},
		[]pgtype.OID{pgtype.VarcharOID},
		[]int16{pgx.BinaryFormatCode},
	)

	err := batch.Send(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	var id int
	err = batch.QueryRowResults().Scan(&id)
	if err != nil {
		t.Error(err)
	}
	batch.Close()
	tx.Rollback()

	row := conn.QueryRow("select count(1) from ledger1 where id = $1", id)
	var count int
	row.Scan(&count)
	if count != 0 {
		t.Errorf("count => %v, want %v", count, 0)
	}

	ensureConnValid(t, conn)
}
