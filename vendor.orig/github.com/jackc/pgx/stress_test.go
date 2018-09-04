package pgx_test

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/jackc/fake"
	"github.com/jackc/pgx"
)

type execer interface {
	Exec(sql string, arguments ...interface{}) (commandTag pgx.CommandTag, err error)
}
type queryer interface {
	Query(sql string, args ...interface{}) (*pgx.Rows, error)
}
type queryRower interface {
	QueryRow(sql string, args ...interface{}) *pgx.Row
}

func TestStressConnPool(t *testing.T) {
	t.Parallel()

	maxConnections := 8
	pool := createConnPool(t, maxConnections)
	defer pool.Close()

	setupStressDB(t, pool)

	actions := []struct {
		name string
		fn   func(*pgx.ConnPool, int) error
	}{
		{"insertUnprepared", func(p *pgx.ConnPool, n int) error { return insertUnprepared(p, n) }},
		{"queryRowWithoutParams", func(p *pgx.ConnPool, n int) error { return queryRowWithoutParams(p, n) }},
		{"query", func(p *pgx.ConnPool, n int) error { return queryCloseEarly(p, n) }},
		{"queryCloseEarly", func(p *pgx.ConnPool, n int) error { return query(p, n) }},
		{"queryErrorWhileReturningRows", func(p *pgx.ConnPool, n int) error { return queryErrorWhileReturningRows(p, n) }},
		{"txInsertRollback", txInsertRollback},
		{"txInsertCommit", txInsertCommit},
		{"txMultipleQueries", txMultipleQueries},
		{"notify", notify},
		{"listenAndPoolUnlistens", listenAndPoolUnlistens},
		{"reset", func(p *pgx.ConnPool, n int) error { p.Reset(); return nil }},
		{"poolPrepareUseAndDeallocate", poolPrepareUseAndDeallocate},
		{"canceledQueryExContext", canceledQueryExContext},
		{"canceledExecExContext", canceledExecExContext},
	}

	actionCount := 1000
	if s := os.Getenv("STRESS_FACTOR"); s != "" {
		stressFactor, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			t.Fatalf("failed to parse STRESS_FACTOR: %v", s)
		}
		actionCount *= int(stressFactor)
	}

	workerCount := 16

	workChan := make(chan int)
	doneChan := make(chan struct{})
	errChan := make(chan error)

	work := func() {
		for n := range workChan {
			action := actions[rand.Intn(len(actions))]
			err := action.fn(pool, n)
			if err != nil {
				errChan <- errors.Errorf("%s: %v", action.name, err)
				break
			}
		}
		doneChan <- struct{}{}
	}

	for i := 0; i < workerCount; i++ {
		go work()
	}

	for i := 0; i < actionCount; i++ {
		select {
		case workChan <- i:
		case err := <-errChan:
			close(workChan)
			t.Fatal(err)
		}
	}
	close(workChan)

	for i := 0; i < workerCount; i++ {
		<-doneChan
	}
}

func setupStressDB(t *testing.T, pool *pgx.ConnPool) {
	_, err := pool.Exec(`
		drop table if exists widgets;
		create table widgets(
			id serial primary key,
			name varchar not null,
			description text,
			creation_time timestamptz
		);
`)
	if err != nil {
		t.Fatal(err)
	}
}

func insertUnprepared(e execer, actionNum int) error {
	sql := `
		insert into widgets(name, description, creation_time)
		values($1, $2, $3)`

	_, err := e.Exec(sql, fake.ProductName(), fake.Sentences(), time.Now())
	return err
}

func queryRowWithoutParams(qr queryRower, actionNum int) error {
	var id int32
	var name, description string
	var creationTime time.Time

	sql := `select * from widgets order by random() limit 1`

	err := qr.QueryRow(sql).Scan(&id, &name, &description, &creationTime)
	if err == pgx.ErrNoRows {
		return nil
	}
	return err
}

func query(q queryer, actionNum int) error {
	sql := `select * from widgets order by random() limit $1`

	rows, err := q.Query(sql, 10)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id int32
		var name, description string
		var creationTime time.Time
		rows.Scan(&id, &name, &description, &creationTime)
	}

	return rows.Err()
}

func queryCloseEarly(q queryer, actionNum int) error {
	sql := `select * from generate_series(1,$1)`

	rows, err := q.Query(sql, 100)
	if err != nil {
		return err
	}
	defer rows.Close()

	for i := 0; i < 10 && rows.Next(); i++ {
		var n int32
		rows.Scan(&n)
	}
	rows.Close()

	return rows.Err()
}

func queryErrorWhileReturningRows(q queryer, actionNum int) error {
	// This query should divide by 0 within the first number of rows
	sql := `select 42 / (random() * 20)::integer from generate_series(1,100000)`

	rows, err := q.Query(sql)
	if err != nil {
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		var n int32
		rows.Scan(&n)
	}

	if _, ok := rows.Err().(pgx.PgError); ok {
		return nil
	}
	return rows.Err()
}

func notify(pool *pgx.ConnPool, actionNum int) error {
	_, err := pool.Exec("notify stress")
	return err
}

func listenAndPoolUnlistens(pool *pgx.ConnPool, actionNum int) error {
	conn, err := pool.Acquire()
	if err != nil {
		return err
	}
	defer pool.Release(conn)

	err = conn.Listen("stress")
	if err != nil {
		return err
	}

	ctx, _ := context.WithTimeout(context.Background(), 100*time.Millisecond)
	_, err = conn.WaitForNotification(ctx)
	if err == context.DeadlineExceeded {
		return nil
	}
	return err
}

func poolPrepareUseAndDeallocate(pool *pgx.ConnPool, actionNum int) error {
	psName := fmt.Sprintf("poolPreparedStatement%d", actionNum)

	_, err := pool.Prepare(psName, "select $1::text")
	if err != nil {
		return err
	}

	var s string
	err = pool.QueryRow(psName, "hello").Scan(&s)
	if err != nil {
		return err
	}

	if s != "hello" {
		return errors.Errorf("Prepared statement did not return expected value: %v", s)
	}

	return pool.Deallocate(psName)
}

func txInsertRollback(pool *pgx.ConnPool, actionNum int) error {
	tx, err := pool.Begin()
	if err != nil {
		return err
	}

	sql := `
		insert into widgets(name, description, creation_time)
		values($1, $2, $3)`

	_, err = tx.Exec(sql, fake.ProductName(), fake.Sentences(), time.Now())
	if err != nil {
		return err
	}

	return tx.Rollback()
}

func txInsertCommit(pool *pgx.ConnPool, actionNum int) error {
	tx, err := pool.Begin()
	if err != nil {
		return err
	}

	sql := `
		insert into widgets(name, description, creation_time)
		values($1, $2, $3)`

	_, err = tx.Exec(sql, fake.ProductName(), fake.Sentences(), time.Now())
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func txMultipleQueries(pool *pgx.ConnPool, actionNum int) error {
	tx, err := pool.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	errExpectedTxDeath := errors.New("Expected tx death")

	actions := []struct {
		name string
		fn   func() error
	}{
		{"insertUnprepared", func() error { return insertUnprepared(tx, actionNum) }},
		{"queryRowWithoutParams", func() error { return queryRowWithoutParams(tx, actionNum) }},
		{"query", func() error { return query(tx, actionNum) }},
		{"queryCloseEarly", func() error { return queryCloseEarly(tx, actionNum) }},
		{"queryErrorWhileReturningRows", func() error {
			err := queryErrorWhileReturningRows(tx, actionNum)
			if err != nil {
				return err
			}
			return errExpectedTxDeath
		}},
	}

	for i := 0; i < 20; i++ {
		action := actions[rand.Intn(len(actions))]
		err := action.fn()
		if err == errExpectedTxDeath {
			return nil
		} else if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func canceledQueryExContext(pool *pgx.ConnPool, actionNum int) error {
	ctx, cancelFunc := context.WithCancel(context.Background())
	go func() {
		time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
		cancelFunc()
	}()

	rows, err := pool.QueryEx(ctx, "select pg_sleep(2)", nil)
	if err == context.Canceled {
		return nil
	} else if err != nil {
		return errors.Errorf("Only allowed error is context.Canceled, got %v", err)
	}

	for rows.Next() {
		return errors.New("should never receive row")
	}

	if rows.Err() != context.Canceled {
		return errors.Errorf("Expected context.Canceled error, got %v", rows.Err())
	}

	return nil
}

func canceledExecExContext(pool *pgx.ConnPool, actionNum int) error {
	ctx, cancelFunc := context.WithCancel(context.Background())
	go func() {
		time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
		cancelFunc()
	}()

	_, err := pool.ExecEx(ctx, "select pg_sleep(2)", nil)
	if err != context.Canceled {
		return errors.Errorf("Expected context.Canceled error, got %v", err)
	}

	return nil
}
