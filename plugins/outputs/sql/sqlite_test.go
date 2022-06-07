//go:build linux && freebsd && (!mips || !mips64)
// +build linux
// +build freebsd
// +build !mips !mips64

package sql

import (
	gosql "database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestSqlite(t *testing.T) {
	outDir := t.TempDir()

	dbfile := filepath.Join(outDir, "db")

	// Use the plugin to write to the database address :=
	// fmt.Sprintf("file:%v", dbfile)
	address := dbfile // accepts a path or a file: URI
	p := newSQL()
	p.Log = testutil.Logger{}
	p.Driver = "sqlite"
	p.DataSourceName = address

	require.NoError(t, p.Connect())
	require.NoError(t, p.Write(
		testMetrics,
	))

	//read directly from the database
	db, err := gosql.Open("sqlite", address)
	require.NoError(t, err)
	defer db.Close()

	var countMetricOne int
	require.NoError(t, db.QueryRow("select count(*) from metric_one").Scan(&countMetricOne))
	require.Equal(t, 1, countMetricOne)

	var countMetricTwo int
	require.NoError(t, db.QueryRow("select count(*) from metric_one").Scan(&countMetricTwo))
	require.Equal(t, 1, countMetricTwo)

	var rows *gosql.Rows

	// Check that tables were created as expected
	rows, err = db.Query("select sql from sqlite_master")
	require.NoError(t, err)
	var sql string
	require.True(t, rows.Next())
	require.NoError(t, rows.Scan(&sql))
	require.Equal(t,
		`CREATE TABLE "metric_one"("timestamp" TIMESTAMP,"tag_one" TEXT,"tag_two" TEXT,"int64_one" INT,"int64_two" INT)`,
		sql,
	)
	require.True(t, rows.Next())
	require.NoError(t, rows.Scan(&sql))
	require.Equal(t,
		`CREATE TABLE "metric_two"("timestamp" TIMESTAMP,"tag_three" TEXT,"string_one" TEXT)`,
		sql,
	)
	require.True(t, rows.Next())
	require.NoError(t, rows.Scan(&sql))
	require.Equal(t,
		`CREATE TABLE "metric three"("timestamp" TIMESTAMP,"tag four" TEXT,"string two" TEXT)`,
		sql,
	)
	require.False(t, rows.Next())
	require.NoError(t, rows.Close()) //nolint:sqlclosecheck

	// sqlite stores dates as strings. They may be in the local
	// timezone. The test needs to parse them back into a time.Time to
	// check them.
	//timeLayout := "2006-01-02 15:04:05 -0700 MST"
	timeLayout := "2006-01-02T15:04:05Z"
	var actualTime time.Time

	// Check contents of tables
	rows, err = db.Query("select timestamp, tag_one, tag_two, int64_one, int64_two from metric_one")
	require.NoError(t, err)
	require.True(t, rows.Next())
	var (
		a    string
		b, c string
		d, e int64
	)
	require.NoError(t, rows.Scan(&a, &b, &c, &d, &e))
	actualTime, err = time.Parse(timeLayout, a)
	require.NoError(t, err)
	require.Equal(t, ts, actualTime.UTC())
	require.Equal(t, "tag1", b)
	require.Equal(t, "tag2", c)
	require.Equal(t, int64(1234), d)
	require.Equal(t, int64(2345), e)
	require.False(t, rows.Next())
	require.NoError(t, rows.Close()) //nolint:sqlclosecheck

	rows, err = db.Query("select timestamp, tag_three, string_one from metric_two")
	require.NoError(t, err)
	require.True(t, rows.Next())
	var (
		f, g, h string
	)
	require.NoError(t, rows.Scan(&f, &g, &h))
	actualTime, err = time.Parse(timeLayout, f)
	require.NoError(t, err)
	require.Equal(t, ts, actualTime.UTC())
	require.Equal(t, "tag3", g)
	require.Equal(t, "string1", h)
	require.False(t, rows.Next())
	require.NoError(t, rows.Close()) //nolint:sqlclosecheck

	rows, err = db.Query(`select timestamp, "tag four", "string two" from "metric three"`)
	require.NoError(t, err)
	require.True(t, rows.Next())
	var (
		i, j, k string
	)
	require.NoError(t, rows.Scan(&i, &j, &k))
	actualTime, err = time.Parse(timeLayout, i)
	require.NoError(t, err)
	require.Equal(t, ts, actualTime.UTC())
	require.Equal(t, "tag4", j)
	require.Equal(t, "string2", k)
	require.False(t, rows.Next())
	require.NoError(t, rows.Close()) //nolint:sqlclosecheck
}
