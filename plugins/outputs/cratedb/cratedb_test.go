package cratedb

import (
	"database/sql"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestConnectAndWrite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("CIRCLE_PROJECT_REPONAME") != "" {
		t.Skip("Skipping test on CircleCI due to docker failures")
	}

	url := testURL()
	table := "test"

	// dropSQL drops our table before each test. This simplifies changing the
	// schema during development :).
	dropSQL := "DROP TABLE IF EXISTS " + escapeString(table, `"`)
	db, err := sql.Open("pgx", url)
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
		hashIDVal, err := escapeValue(hashID(m))
		require.NoError(t, err)
		timestamp, err := escapeValue(m.Time())
		require.NoError(t, err)

		var id int64
		row := db.QueryRow(
			"SELECT hash_id FROM " + escapeString(table, `"`) + " " +
				"WHERE hash_id = " + hashIDVal + " " +
				"AND timestamp = " + timestamp,
		)
		require.NoError(t, row.Scan(&id))
		// We could check the whole row, but this is meant to be more of a smoke
		// test, so just checking the HashID seems fine.
		require.Equal(t, id, hashID(m))
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
(-4023501406646044814, '2009-11-10T23:00:00+0000', 'test1', {"tag1" = 'value1'}, {"value" = 1});
`),
		},
	}

	for _, test := range tests {
		if got, err := insertSQL("my_table", test.Metrics); err != nil {
			t.Error(err)
		} else if got != test.Want {
			t.Errorf("got:\n%s\n\nwant:\n%s", got, test.Want)
		}
	}
}

func Test_escapeValue(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("CIRCLE_PROJECT_REPONAME") != "" {
		t.Skip("Skipping test on CircleCI due to docker failures")
	}

	tests := []struct {
		Val  interface{}
		Want string
	}{
		// string
		{`foo`, `'foo'`},
		{`foo'bar 'yeah`, `'foo''bar ''yeah'`},
		// int types
		{int64(123), `123`},
		{uint64(123), `123`},
		{uint64(MaxInt64) + 1, `9223372036854775807`},
		{true, `true`},
		{false, `false`},
		// float types
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
		{map[string]interface{}{`fo"o`: `b'ar`, `ab'c`: `xy"z`, `on"""e`: `mo'''re`}, `{"ab'c" = 'xy"z', "fo""o" = 'b''ar', "on""""""e" = 'mo''''''re'}`},
	}

	url := testURL()
	db, err := sql.Open("pgx", url)
	require.NoError(t, err)
	defer db.Close()

	for _, test := range tests {
		got, err := escapeValue(test.Val)
		if err != nil {
			t.Errorf("val: %#v: %s", test.Val, err)
		} else if got != test.Want {
			t.Errorf("got:\n%s\n\nwant:\n%s", got, test.Want)
		}

		// This is a smoke test that will blow up if our escaping causing a SQL
		// syntax error, which may allow for an attack.
		var reply interface{}
		row := db.QueryRow("SELECT " + got)
		require.NoError(t, row.Scan(&reply))
	}
}

func Test_hashID(t *testing.T) {
	tests := []struct {
		Name   string
		Tags   map[string]string
		Fields map[string]interface{}
		Want   int64
	}{
		{
			Name:   "metric1",
			Tags:   map[string]string{"tag1": "val1", "tag2": "val2"},
			Fields: map[string]interface{}{"field1": "val1", "field2": "val2"},
			Want:   8973971082006474188,
		},

		// This metric has a different tag order (in a perhaps non-ideal attempt to
		// trigger different pseudo-random map iteration)) and fields (none)
		// compared to the previous metric, but should still get the same hash.
		{
			Name:   "metric1",
			Tags:   map[string]string{"tag2": "val2", "tag1": "val1"},
			Fields: map[string]interface{}{"field3": "val3"},
			Want:   8973971082006474188,
		},

		// Different metric name -> different hash
		{
			Name:   "metric2",
			Tags:   map[string]string{"tag1": "val1", "tag2": "val2"},
			Fields: map[string]interface{}{"field1": "val1", "field2": "val2"},
			Want:   306487682448261783,
		},

		// Different tag val -> different hash
		{
			Name:   "metric1",
			Tags:   map[string]string{"tag1": "new-val", "tag2": "val2"},
			Fields: map[string]interface{}{"field1": "val1", "field2": "val2"},
			Want:   1938713695181062970,
		},

		// Different tag key -> different hash
		{
			Name:   "metric1",
			Tags:   map[string]string{"new-key": "val1", "tag2": "val2"},
			Fields: map[string]interface{}{"field1": "val1", "field2": "val2"},
			Want:   7678889081527706328,
		},
	}

	for i, test := range tests {
		m, err := metric.New(
			test.Name,
			test.Tags,
			test.Fields,
			time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
		)
		require.NoError(t, err)
		if got := hashID(m); got != test.Want {
			t.Errorf("test #%d: got=%d want=%d", i, got, test.Want)
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
