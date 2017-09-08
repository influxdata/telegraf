package cratedb

import (
	"database/sql"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestConnectAndWrite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	url := testURL()
	table := "test"

	// dropSQL drops our table before each test. This simplifies changing the
	// schema during development :).
	dropSQL := "DROP TABLE IF EXISTS " + escapeString(table, `"`)
	db, err := sql.Open("postgres", url)
	require.NoError(t, err)
	_, err = db.Exec(dropSQL)
	require.NoError(t, err)
	defer db.Close()

	c := &CrateDB{
		URL:         url,
		Table:       table,
		Timeout:     internal.Duration{Duration: time.Second * 5},
		TableCreate: true,
	}

	metrics := testutil.MockMetrics()
	require.NoError(t, c.Connect())
	require.NoError(t, c.Write(metrics))

	// The code below verifies that the metrics were written. We have to select
	// the rows using their primary keys in order to take advantage of
	// read-after-write consistency in CrateDB.
	for _, m := range metrics {
		hashID, err := escapeValue(int64(m.HashID()))
		require.NoError(t, err)
		timestamp, err := escapeValue(m.Time())
		require.NoError(t, err)

		var id int64
		row := db.QueryRow(
			"SELECT hash_id FROM " + escapeString(table, `"`) + " " +
				"WHERE hash_id = " + hashID + " " +
				"AND timestamp = " + timestamp,
		)
		require.NoError(t, row.Scan(&id))
		// We could check the whole row, but this is meant to be more of a smoke
		// test, so just checking the HashID seems fine.
		require.Equal(t, id, int64(m.HashID()))
	}

	require.NoError(t, c.Close())
}

func Test_insertSQL(t *testing.T) {
	tests := []struct {
		Metrics []telegraf.Metric
		Want    string
	}{
		{
			Metrics: testutil.MockMetrics(),
			Want: strings.TrimSpace(`
INSERT INTO my_table ("hash_id", "timestamp", "name", "tags", "fields")
VALUES
(1845393540509842047, '2009-11-10T23:00:00+0000', 'test1', {"tag1" = 'value1'}, {"value" = 1});
`),
		},
	}

	for _, test := range tests {
		if got, err := insertSQL("my_table", test.Metrics, time.UTC); err != nil {
			t.Error(err)
		} else if got != test.Want {
			t.Errorf("got:\n%s\n\nwant:\n%s", got, test.Want)
		}
	}
}

func Test_escapeValue(t *testing.T) {
	tests := []struct {
		Val  interface{}
		Want string
	}{
		// string
		{`foo`, `'foo'`},
		{`foo'bar 'yeah`, `'foo''bar ''yeah'`},
		// int types
		{123, `123`}, // int
		{int64(123), `123`},
		{int32(123), `123`},
		// float types
		{123.456, `123.456`},
		{float32(123.456), `123.456`}, // floating point SNAFU
		{float64(123.456), `123.456`},
		// time.Time
		{time.Date(2017, 8, 7, 16, 44, 52, 123*1000*1000, time.FixedZone("Dreamland", 5400)), `'2017-08-07T16:44:52.123+0130'`},
		// map[string]string
		{map[string]string{}, `{}`},
		{map[string]string(nil), `{}`},
		{map[string]string{"foo": "bar"}, `{"foo" = 'bar'}`},
		{map[string]string{"foo": "bar", "one": "more"}, `{"foo" = 'bar', "one" = 'more'}`},
		// map[string]interface{}
		{map[string]interface{}{}, `{}`},
		{map[string]interface{}(nil), `{}`},
		{map[string]interface{}{"foo": "bar"}, `{"foo" = 'bar'}`},
		{map[string]interface{}{"foo": "bar", "one": "more"}, `{"foo" = 'bar', "one" = 'more'}`},
		{map[string]interface{}{"foo": map[string]interface{}{"one": "more"}}, `{"foo" = {"one" = 'more'}}`},
	}

	for _, test := range tests {
		if got, err := escapeValue(test.Val); err != nil {
			t.Errorf("val: %#v: %s", test.Val, err)
		} else if got != test.Want {
			t.Errorf("got:\n%s\n\nwant:\n%s", got, test.Want)
		}
	}
}

func testURL() string {
	url := os.Getenv("CRATE_URL")
	if url == "" {
		return "postgres://" + testutil.GetLocalHost() + ":6543/test?sslmode=disable"
	}
	return url
}
