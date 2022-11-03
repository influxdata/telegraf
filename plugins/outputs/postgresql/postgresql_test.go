package postgresql

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs/postgresql/utils"
	"github.com/influxdata/telegraf/testutil"
)

type Log struct {
	level  pgx.LogLevel
	format string
	args   []interface{}
}

func (l Log) String() string {
	// We have to use Errorf() as Sprintf() doesn't allow usage of %w.
	return fmt.Errorf("%s: "+l.format, append([]interface{}{l.level}, l.args...)...).Error()
}

// LogAccumulator is a log collector that satisfies telegraf.Logger.
type LogAccumulator struct {
	logs      []Log
	cond      *sync.Cond
	tb        testing.TB
	emitLevel pgx.LogLevel
}

func NewLogAccumulator(tb testing.TB) *LogAccumulator {
	return &LogAccumulator{
		cond: sync.NewCond(&sync.Mutex{}),
		tb:   tb,
	}
}

func (la *LogAccumulator) append(level pgx.LogLevel, format string, args []interface{}) {
	la.tb.Helper()

	la.cond.L.Lock()
	log := Log{level, format, args}
	la.logs = append(la.logs, log)

	if la.emitLevel == 0 || log.level <= la.emitLevel {
		la.tb.Log(log.String())
	}

	la.cond.Broadcast()
	la.cond.L.Unlock()
}

func (la *LogAccumulator) HasLevel(level pgx.LogLevel) bool {
	la.cond.L.Lock()
	defer la.cond.L.Unlock()
	for _, log := range la.logs {
		if log.level > 0 && log.level <= level {
			return true
		}
	}
	return false
}

func (la *LogAccumulator) WaitLen(n int) []Log {
	la.cond.L.Lock()
	defer la.cond.L.Unlock()
	for len(la.logs) < n {
		la.cond.Wait()
	}
	return la.logs[:]
}

// Waits for a specific query log from pgx to show up.
func (la *LogAccumulator) WaitFor(f func(l Log) bool, waitCommit bool) {
	la.cond.L.Lock()
	defer la.cond.L.Unlock()
	i := 0
	var commitPid uint32
	for {
		for ; i < len(la.logs); i++ {
			log := la.logs[i]
			if commitPid == 0 {
				if f(log) {
					if !waitCommit {
						return
					}
					commitPid = log.args[1].(MSI)["pid"].(uint32)
				}
			} else {
				if len(log.args) < 2 {
					continue
				}
				data, ok := log.args[1].(MSI)
				if !ok || data["pid"] != commitPid {
					continue
				}
				if log.args[0] == "Exec" && data["sql"] == "commit" {
					return
				} else if log.args[0] == "Exec" && data["sql"] == "rollback" {
					// transaction aborted, start looking for another match
					commitPid = 0
				} else if log.level == pgx.LogLevelError {
					commitPid = 0
				}
			}
		}
		la.cond.Wait()
	}
}

func (la *LogAccumulator) WaitForQuery(str string, waitCommit bool) {
	la.WaitFor(func(log Log) bool {
		return log.format == "PG %s - %+v" &&
			(log.args[0].(string) == "Query" || log.args[0].(string) == "Exec") &&
			strings.Contains(log.args[1].(MSI)["sql"].(string), str)
	}, waitCommit)
}

func (la *LogAccumulator) WaitForCopy(tableName string, waitCommit bool) {
	la.WaitFor(func(log Log) bool {
		return log.format == "PG %s - %+v" &&
			log.args[0].(string) == "CopyFrom" &&
			log.args[1].(MSI)["tableName"].(pgx.Identifier)[1] == tableName
	}, waitCommit)
}

// Clear any stored logs.
// Do not run this while any WaitFor* operations are in progress.
func (la *LogAccumulator) Clear() {
	la.cond.L.Lock()
	if len(la.logs) > 0 {
		la.logs = nil
	}
	la.cond.L.Unlock()
}

func (la *LogAccumulator) Logs() []Log {
	la.cond.L.Lock()
	defer la.cond.L.Unlock()
	return la.logs[:]
}

func (la *LogAccumulator) Errorf(format string, args ...interface{}) {
	la.tb.Helper()
	la.append(pgx.LogLevelError, format, args)
}

func (la *LogAccumulator) Error(args ...interface{}) {
	la.tb.Helper()
	la.append(pgx.LogLevelError, "%v", args)
}

func (la *LogAccumulator) Debugf(format string, args ...interface{}) {
	la.tb.Helper()
	la.append(pgx.LogLevelDebug, format, args)
}

func (la *LogAccumulator) Debug(args ...interface{}) {
	la.tb.Helper()
	la.append(pgx.LogLevelDebug, "%v", args)
}

func (la *LogAccumulator) Warnf(format string, args ...interface{}) {
	la.tb.Helper()
	la.append(pgx.LogLevelWarn, format, args)
}

func (la *LogAccumulator) Warn(args ...interface{}) {
	la.tb.Helper()
	la.append(pgx.LogLevelWarn, "%v", args)
}

func (la *LogAccumulator) Infof(format string, args ...interface{}) {
	la.tb.Helper()
	la.append(pgx.LogLevelInfo, format, args)
}

func (la *LogAccumulator) Info(args ...interface{}) {
	la.tb.Helper()
	la.append(pgx.LogLevelInfo, "%v", args)
}

var ctx = context.Background()

type PostgresqlTest struct {
	*Postgresql
	Logger *LogAccumulator
}

func newPostgresqlTest(tb testing.TB) *PostgresqlTest {
	if testing.Short() {
		tb.Skip("Skipping integration test in short mode")
	}

	servicePort := "5432"
	username := "postgres"
	password := "postgres"
	testDatabaseName := "telegraf_test"

	container := testutil.Container{
		Image:        "postgres:alpine",
		ExposedPorts: []string{servicePort},
		Env: map[string]string{
			"POSTGRES_USER":     username,
			"POSTGRES_PASSWORD": password,
			"POSTGRES_DB":       "telegraf_test",
		},
		WaitingFor: wait.ForAll(
			// the database comes up twice, once right away, then again a second
			// time after the docker entrypoint starts configuraiton
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
			wait.ForListeningPort(nat.Port(servicePort)),
		),
	}
	tb.Cleanup(container.Terminate)

	err := container.Start()
	require.NoError(tb, err, "failed to start container")

	p := newPostgresql()
	p.Connection = fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s",
		container.Address,
		container.Ports[servicePort],
		username,
		password,
		testDatabaseName,
	)
	logger := NewLogAccumulator(tb)
	p.Logger = logger
	p.LogLevel = "debug"
	require.NoError(tb, p.Init())

	pt := &PostgresqlTest{Postgresql: p}
	pt.Logger = logger

	return pt
}

func TestPostgresqlConnectIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p := newPostgresqlTest(t)
	require.NoError(t, p.Connect())
	require.EqualValues(t, 1, p.db.Stat().MaxConns())

	p = newPostgresqlTest(t)
	p.Connection += " pool_max_conns=2"
	_ = p.Init()
	require.NoError(t, p.Connect())
	require.EqualValues(t, 2, p.db.Stat().MaxConns())
}

func newMetric(
	t *testing.T,
	suffix string,
	tags map[string]string,
	fields map[string]interface{},
) telegraf.Metric {
	return testutil.MustMetric(t.Name()+suffix, tags, fields, time.Now())
}

type MSS = map[string]string
type MSI = map[string]interface{}

func dbTableDump(t *testing.T, db *pgxpool.Pool, suffix string) []MSI {
	rows, err := db.Query(ctx, "SELECT * FROM "+pgx.Identifier{t.Name() + suffix}.Sanitize())
	require.NoError(t, err)
	defer rows.Close()

	var dump []MSI
	for rows.Next() {
		msi := MSI{}
		vals, err := rows.Values()
		require.NoError(t, err)
		for i, fd := range rows.FieldDescriptions() {
			msi[string(fd.Name)] = vals[i]
		}
		dump = append(dump, msi)
	}
	require.NoError(t, rows.Err())
	return dump
}

func TestWriteIntegration_sequential(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p := newPostgresqlTest(t)
	require.NoError(t, p.Connect())

	metrics := []telegraf.Metric{
		newMetric(t, "_a", MSS{}, MSI{"v": 1}),
		newMetric(t, "_b", MSS{}, MSI{"v": 2}),
		newMetric(t, "_a", MSS{}, MSI{"v": 3}),
	}
	require.NoError(t, p.Write(metrics))

	dumpA := dbTableDump(t, p.db, "_a")
	dumpB := dbTableDump(t, p.db, "_b")

	require.Len(t, dumpA, 2)
	require.EqualValues(t, 1, dumpA[0]["v"])
	require.EqualValues(t, 3, dumpA[1]["v"])

	require.Len(t, dumpB, 1)
	require.EqualValues(t, 2, dumpB[0]["v"])

	p.Logger.Clear()
	require.NoError(t, p.Write(metrics))

	stmtCount := 0
	for _, log := range p.Logger.Logs() {
		if strings.Contains(log.String(), "info: PG ") {
			stmtCount++
		}
	}
	require.Equal(t, 6, stmtCount) // BEGIN, SAVEPOINT, COPY table _a, SAVEPOINT, COPY table _b, COMMIT
}

func TestWriteIntegration_concurrent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p := newPostgresqlTest(t)
	p.dbConfig.MaxConns = 3
	require.NoError(t, p.Connect())

	// Write a metric so it creates a table we can lock.
	metrics := []telegraf.Metric{
		newMetric(t, "_a", MSS{}, MSI{"v": 1}),
	}
	require.NoError(t, p.Write(metrics))
	p.Logger.WaitForCopy(t.Name()+"_a", false)
	// clear so that the WaitForCopy calls below don't pick up this one
	p.Logger.Clear()

	// Lock the table so that we ensure the writes hangs and the plugin has to open another connection.
	tx, err := p.db.Begin(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx) //nolint:errcheck
	_, err = tx.Exec(ctx, "LOCK TABLE "+utils.QuoteIdentifier(t.Name()+"_a"))
	require.NoError(t, err)

	metrics = []telegraf.Metric{
		newMetric(t, "_a", MSS{}, MSI{"v": 2}),
	}
	require.NoError(t, p.Write(metrics))

	// Note, there is technically a possible race here, where it doesn't try to insert into _a until after _b. However
	// this should be practically impossible, and trying to engineer a solution to account for it would be even more
	// complex than we already are.

	metrics = []telegraf.Metric{
		newMetric(t, "_b", MSS{}, MSI{"v": 3}),
	}
	require.NoError(t, p.Write(metrics))

	p.Logger.WaitForCopy(t.Name()+"_b", false)
	// release the lock on table _a
	_ = tx.Rollback(ctx)
	p.Logger.WaitForCopy(t.Name()+"_a", false)

	dumpA := dbTableDump(t, p.db, "_a")
	dumpB := dbTableDump(t, p.db, "_b")

	require.Len(t, dumpA, 2)
	require.EqualValues(t, 1, dumpA[0]["v"])
	require.EqualValues(t, 2, dumpA[1]["v"])

	require.Len(t, dumpB, 1)
	require.EqualValues(t, 3, dumpB[0]["v"])

	// We should have had 3 connections. One for the lock, and one for each table.
	require.EqualValues(t, 3, p.db.Stat().TotalConns())
}

// Test that the bad metric is dropped, and the rest of the batch succeeds.
func TestWriteIntegration_sequentialPermError(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p := newPostgresqlTest(t)
	require.NoError(t, p.Connect())

	metrics := []telegraf.Metric{
		newMetric(t, "_a", MSS{}, MSI{"v": 1}),
		newMetric(t, "_b", MSS{}, MSI{"v": 2}),
	}
	require.NoError(t, p.Write(metrics))

	metrics = []telegraf.Metric{
		newMetric(t, "_a", MSS{}, MSI{"v": "a"}),
		newMetric(t, "_b", MSS{}, MSI{"v": 3}),
	}
	require.NoError(t, p.Write(metrics))

	dumpA := dbTableDump(t, p.db, "_a")
	dumpB := dbTableDump(t, p.db, "_b")
	require.Len(t, dumpA, 1)
	require.Len(t, dumpB, 2)

	haveError := false
	for _, l := range p.Logger.Logs() {
		if strings.Contains(l.String(), "write error") {
			haveError = true
			break
		}
	}
	require.True(t, haveError, "write error not found in log")
}

// Test that in a bach with only 1 sub-batch, that we don't return an error.
func TestWriteIntegration_sequentialSinglePermError(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p := newPostgresqlTest(t)
	require.NoError(t, p.Connect())

	metrics := []telegraf.Metric{
		newMetric(t, "", MSS{}, MSI{"v": 1}),
	}
	require.NoError(t, p.Write(metrics))

	metrics = []telegraf.Metric{
		newMetric(t, "", MSS{}, MSI{"v": "a"}),
	}
	require.NoError(t, p.Write(metrics))
}

// Test that the bad metric is dropped, and the rest of the batch succeeds.
func TestWriteIntegration_concurrentPermError(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p := newPostgresqlTest(t)
	p.dbConfig.MaxConns = 2
	require.NoError(t, p.Connect())

	metrics := []telegraf.Metric{
		newMetric(t, "_a", MSS{}, MSI{"v": 1}),
	}
	require.NoError(t, p.Write(metrics))
	p.Logger.WaitForCopy(t.Name()+"_a", false)

	metrics = []telegraf.Metric{
		newMetric(t, "_a", MSS{}, MSI{"v": "a"}),
		newMetric(t, "_b", MSS{}, MSI{"v": 2}),
	}
	require.NoError(t, p.Write(metrics))
	p.Logger.WaitFor(func(l Log) bool {
		return strings.Contains(l.String(), "write error")
	}, false)
	p.Logger.WaitForCopy(t.Name()+"_b", false)

	dumpA := dbTableDump(t, p.db, "_a")
	dumpB := dbTableDump(t, p.db, "_b")
	require.Len(t, dumpA, 1)
	require.Len(t, dumpB, 1)
}

// Verify that in sequential mode, errors are returned allowing telegraf agent to handle & retry
func TestWriteIntegration_sequentialTempError(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p := newPostgresqlTest(t)
	require.NoError(t, p.Connect())

	// To avoid a race condition, we need to know when our goroutine has started listening to the log.
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		// Wait for the CREATE TABLE, and then kill the connection.
		// The WaitFor callback holds a lock on the log. Meaning it will block logging of the next action. So we trigger
		// on CREATE TABLE so that there's a few statements to go before the COMMIT.
		p.Logger.WaitFor(func(log Log) bool {
			if strings.Contains(log.String(), "release wg") {
				wg.Done()
			}

			if !strings.Contains(log.String(), "CREATE TABLE") {
				return false
			}
			pid := log.args[1].(MSI)["pid"].(uint32)

			conf := p.db.Config().ConnConfig
			conf.Logger = nil
			c, err := pgx.ConnectConfig(context.Background(), conf)
			if err != nil {
				return true
			}
			_, err = c.Exec(context.Background(), "SELECT pg_terminate_backend($1)", pid)
			require.NoError(t, err)
			return true
		}, false)
	}()

	p.Logger.Infof("release wg")
	wg.Wait()

	metrics := []telegraf.Metric{
		newMetric(t, "_a", MSS{}, MSI{"v": 1}),
	}
	require.Error(t, p.Write(metrics))
}

// Verify that when using concurrency, errors are not returned, but instead logged and automatically retried
func TestWriteIntegration_concurrentTempError(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p := newPostgresqlTest(t)
	p.dbConfig.MaxConns = 2
	require.NoError(t, p.Connect())

	// To avoid a race condition, we need to know when our goroutine has started listening to the log.
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		// Wait for the CREATE TABLE, and then kill the connection.
		// The WaitFor callback holds a lock on the log. Meaning it will block logging of the next action. So we trigger
		// on CREATE TABLE so that there's a few statements to go before the COMMIT.
		p.Logger.WaitFor(func(log Log) bool {
			if strings.Contains(log.String(), "release wg") {
				wg.Done()
			}

			if !strings.Contains(log.String(), "CREATE TABLE") {
				return false
			}
			pid := log.args[1].(MSI)["pid"].(uint32)

			conf := p.db.Config().ConnConfig
			conf.Logger = nil
			c, err := pgx.ConnectConfig(context.Background(), conf)
			if err != nil {
				return true
			}
			_, err = c.Exec(context.Background(), "SELECT pg_terminate_backend($1)", pid)
			require.NoError(t, err)
			return true
		}, false)
	}()
	p.Logger.Infof("release wg")
	wg.Wait()

	metrics := []telegraf.Metric{
		newMetric(t, "_a", MSS{}, MSI{"v": 1}),
	}
	require.NoError(t, p.Write(metrics))

	p.Logger.WaitForCopy(t.Name()+"_a", false)
	dumpA := dbTableDump(t, p.db, "_a")
	require.Len(t, dumpA, 1)

	haveError := false
	for _, l := range p.Logger.Logs() {
		if strings.Contains(l.String(), "write error") {
			haveError = true
			break
		}
	}
	require.True(t, haveError, "write error not found in log")
}

func TestWriteTagTableIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p := newPostgresqlTest(t)
	p.TagsAsForeignKeys = true
	require.NoError(t, p.Connect())

	metrics := []telegraf.Metric{
		newMetric(t, "", MSS{"tag": "foo"}, MSI{"v": 1}),
	}
	require.NoError(t, p.Write(metrics))

	dump := dbTableDump(t, p.db, "")
	require.Len(t, dump, 1)
	require.EqualValues(t, 1, dump[0]["v"])

	dumpTags := dbTableDump(t, p.db, p.TagTableSuffix)
	require.Len(t, dumpTags, 1)
	require.EqualValues(t, dump[0]["tag_id"], dumpTags[0]["tag_id"])
	require.EqualValues(t, "foo", dumpTags[0]["tag"])

	p.Logger.Clear()
	require.NoError(t, p.Write(metrics))

	stmtCount := 0
	for _, log := range p.Logger.Logs() {
		if strings.Contains(log.String(), "info: PG ") {
			stmtCount++
		}
	}
	require.Equal(t, 3, stmtCount) // BEGIN, COPY metrics table, COMMIT
}

// Verify that when using TagsAsForeignKeys and a tag can't be written, that we still add the metrics.
func TestWriteIntegration_tagError(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p := newPostgresqlTest(t)
	p.TagsAsForeignKeys = true
	require.NoError(t, p.Connect())

	metrics := []telegraf.Metric{
		newMetric(t, "", MSS{"tag": "foo"}, MSI{"v": 1}),
	}
	require.NoError(t, p.Write(metrics))

	// It'll have the table cached, so won't know we dropped it, will try insert, and get error.
	_, err := p.db.Exec(ctx, "DROP TABLE \""+t.Name()+"_tag\"")
	require.NoError(t, err)

	metrics = []telegraf.Metric{
		newMetric(t, "", MSS{"tag": "foo"}, MSI{"v": 2}),
	}
	require.NoError(t, p.Write(metrics))

	dump := dbTableDump(t, p.db, "")
	require.Len(t, dump, 2)
	require.EqualValues(t, 1, dump[0]["v"])
	require.EqualValues(t, 2, dump[1]["v"])
}

// Verify that when using TagsAsForeignKeys and ForeignTagConstraing and a tag can't be written, that we drop the metrics.
func TestWriteIntegration_tagError_foreignConstraint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p := newPostgresqlTest(t)
	p.TagsAsForeignKeys = true
	p.ForeignTagConstraint = true
	require.NoError(t, p.Connect())

	metrics := []telegraf.Metric{
		newMetric(t, "", MSS{"tag": "foo"}, MSI{"v": 1}),
	}
	require.NoError(t, p.Write(metrics))

	// It'll have the table cached, so won't know we dropped it, will try insert, and get error.
	_, err := p.db.Exec(ctx, "DROP TABLE \""+t.Name()+"_tag\"")
	require.NoError(t, err)

	metrics = []telegraf.Metric{
		newMetric(t, "", MSS{"tag": "bar"}, MSI{"v": 2}),
	}
	require.NoError(t, p.Write(metrics))
	haveError := false
	for _, l := range p.Logger.Logs() {
		if strings.Contains(l.String(), "write error") {
			haveError = true
			break
		}
	}
	require.True(t, haveError, "write error not found in log")

	dump := dbTableDump(t, p.db, "")
	require.Len(t, dump, 1)
	require.EqualValues(t, 1, dump[0]["v"])
}

func TestWriteIntegration_utf8(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p := newPostgresqlTest(t)
	p.TagsAsForeignKeys = true
	require.NoError(t, p.Connect())

	metrics := []telegraf.Metric{
		newMetric(t, "Ѧ𝙱Ƈᗞ",
			MSS{"ăѣ𝔠ծ": "𝘈Ḇ𝖢𝕯٤ḞԍНǏ𝙅ƘԸⲘ𝙉০Ρ𝗤Ɍ𝓢ȚЦ𝒱Ѡ𝓧ƳȤ"},
			MSI{"АḂⲤ𝗗": "𝘢ƀ𝖼ḋếᵮℊ𝙝Ꭵ𝕛кιṃդⱺ𝓅𝘲𝕣𝖘ŧ𝑢ṽẉ𝘅ყž𝜡"},
		),
	}
	require.NoError(t, p.Write(metrics))

	dump := dbTableDump(t, p.db, "Ѧ𝙱Ƈᗞ")
	require.Len(t, dump, 1)
	require.EqualValues(t, "𝘢ƀ𝖼ḋếᵮℊ𝙝Ꭵ𝕛кιṃդⱺ𝓅𝘲𝕣𝖘ŧ𝑢ṽẉ𝘅ყž𝜡", dump[0]["АḂⲤ𝗗"])

	dumpTags := dbTableDump(t, p.db, "Ѧ𝙱Ƈᗞ"+p.TagTableSuffix)
	require.Len(t, dumpTags, 1)
	require.EqualValues(t, dump[0]["tag_id"], dumpTags[0]["tag_id"])
	require.EqualValues(t, "𝘈Ḇ𝖢𝕯٤ḞԍНǏ𝙅ƘԸⲘ𝙉০Ρ𝗤Ɍ𝓢ȚЦ𝒱Ѡ𝓧ƳȤ", dumpTags[0]["ăѣ𝔠ծ"])
}

func TestWriteIntegration_UnsignedIntegers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	p := newPostgresqlTest(t)
	p.Uint64Type = PgUint8
	_ = p.Init()
	if err := p.Connect(); err != nil {
		if strings.Contains(err.Error(), "retreiving OID for uint8 data type") {
			t.Skipf("pguint extension is not installed")
			t.SkipNow()
		}
		require.NoError(t, err)
	}

	metrics := []telegraf.Metric{
		newMetric(t, "", MSS{}, MSI{"v": uint64(math.MaxUint64)}),
	}
	require.NoError(t, p.Write(metrics))

	dump := dbTableDump(t, p.db, "")

	require.Len(t, dump, 1)
	require.EqualValues(t, uint64(math.MaxUint64), dump[0]["v"])

}

// Last ditch effort to find any concurrency issues.
func TestStressConcurrencyIntegration(t *testing.T) {
	t.Skip("Skipping very long test - run locally with no timeout")

	metrics := []telegraf.Metric{
		newMetric(t, "", MSS{"foo": "bar"}, MSI{"a": 1}),
		newMetric(t, "", MSS{"pop": "tart"}, MSI{"b": 1}),
		newMetric(t, "", MSS{"foo": "bar", "pop": "tart"}, MSI{"a": 2, "b": 2}),
		newMetric(t, "_b", MSS{"foo": "bar"}, MSI{"a": 1}),
	}

	concurrency := 4
	loops := 100

	pctl := newPostgresqlTest(t)
	pctl.Logger.emitLevel = pgx.LogLevelWarn
	require.NoError(t, pctl.Connect())

	for i := 0; i < loops; i++ {
		var wgStart, wgDone sync.WaitGroup
		wgStart.Add(concurrency)
		wgDone.Add(concurrency)
		for j := 0; j < concurrency; j++ {
			go func() {
				mShuf := make([]telegraf.Metric, len(metrics))
				copy(mShuf, metrics)
				rand.Shuffle(len(mShuf), func(a, b int) { mShuf[a], mShuf[b] = mShuf[b], mShuf[a] })

				p := newPostgresqlTest(t)
				p.TagsAsForeignKeys = true
				p.Logger.emitLevel = pgx.LogLevelWarn
				p.dbConfig.MaxConns = int32(rand.Intn(3) + 1)
				require.NoError(t, p.Connect())
				wgStart.Done()
				wgStart.Wait()

				err := p.Write(mShuf)
				require.NoError(t, err)
				require.NoError(t, p.Close())
				require.False(t, p.Logger.HasLevel(pgx.LogLevelWarn))
				wgDone.Done()
			}()
		}
		wgDone.Wait()

		if t.Failed() {
			break
		}
	}
}
