package cratedb

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/models"
	"github.com/influxdata/telegraf/testutil"
)

const servicePort = "5432"

func createTestContainer(t *testing.T) *testutil.Container {
	container := testutil.Container{
		Image:        "crate",
		ExposedPorts: []string{servicePort},
		Entrypoint: []string{
			"/docker-entrypoint.sh",
			"-Cdiscovery.type=single-node",
		},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort(servicePort),
			wait.ForLog("recovered [0] indices into cluster_state"),
		),
	}
	err := container.Start()
	require.NoError(t, err, "failed to start container")

	return &container
}

func TestConnectAndWriteIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := createTestContainer(t)
	defer container.Terminate()
	url := fmt.Sprintf("postgres://crate@%s:%s/test", container.Address, container.Ports[servicePort])

	db, err := sql.Open("pgx", url)
	require.NoError(t, err)
	defer db.Close()

	c := &CrateDB{
		URL:         url,
		Table:       "testing",
		Timeout:     config.Duration(time.Second * 5),
		TableCreate: true,
	}

	metrics := testutil.MockMetrics()
	require.NoError(t, c.Connect())
	require.NoError(t, c.Write(metrics))

	// The code below verifies that the metrics were written. We have to select
	// the rows using their primary keys in order to take advantage of
	// read-after-write consistency in CrateDB.
	for _, m := range metrics {
		var id int64
		row := db.QueryRow("SELECT hash_id FROM testing WHERE hash_id = ? AND timestamp = ?", hashID(m), m.Time())
		require.NoError(t, row.Scan(&id))
		// We could check the whole row, but this is meant to be more of a smoke
		// test, so just checking the HashID seems fine.
		require.Equal(t, id, hashID(m))
	}

	require.NoError(t, c.Close())
}

func TestConnectionIssueAtStartup(t *testing.T) {
	// Test case for https://github.com/influxdata/telegraf/issues/13278
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := testutil.Container{
		Image:        "crate",
		ExposedPorts: []string{servicePort},
		Entrypoint: []string{
			"/docker-entrypoint.sh",
			"-Cdiscovery.type=single-node",
		},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort(servicePort),
			wait.ForLog("recovered [0] indices into cluster_state"),
		),
	}
	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()
	url := fmt.Sprintf("postgres://crate@%s:%s/test", container.Address, container.Ports[servicePort])

	// Pause the container for connectivity issues
	require.NoError(t, container.Pause())

	// Create a model to be able to use the startup retry strategy
	plugin := &CrateDB{
		URL:         url,
		Table:       "testing",
		Timeout:     config.Duration(time.Second * 5),
		TableCreate: true,
	}
	model := models.NewRunningOutput(
		plugin,
		&models.OutputConfig{
			Name:                 "cratedb",
			StartupErrorBehavior: "retry",
		},
		1000, 1000,
	)
	require.NoError(t, model.Init())

	// The connect call should succeed even though the table creation was not
	// successful due to the "retry" strategy
	require.NoError(t, model.Connect())

	// Writing the metrics in this state should fail because we are not fully
	// started up
	metrics := testutil.MockMetrics()
	for _, m := range metrics {
		model.AddMetric(m)
	}
	require.ErrorIs(t, model.WriteBatch(), internal.ErrNotConnected)

	// Unpause the container, now writes should succeed
	require.NoError(t, container.Resume())
	require.NoError(t, model.WriteBatch())
	defer model.Close()

	// Verify that the metrics were actually written
	for _, m := range metrics {
		mid := hashID(m)
		row := plugin.db.QueryRow("SELECT hash_id FROM testing WHERE hash_id = ? AND timestamp = ?", mid, m.Time())

		var id int64
		require.NoError(t, row.Scan(&id))
		require.Equal(t, id, mid)
	}
}

func TestInsertSQL(t *testing.T) {
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
		if got, err := insertSQL("my_table", "_", test.Metrics); err != nil {
			t.Error(err)
		} else if got != test.Want {
			t.Errorf("got:\n%s\n\nwant:\n%s", got, test.Want)
		}
	}
}

type escapeValueTest struct {
	Value interface{}
	Want  string
}

func escapeValueTests() []escapeValueTest {
	return []escapeValueTest{
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
		{map[string]string{"f.oo": "bar", "o.n.e": "more"}, `{"f_oo" = 'bar', "o_n_e" = 'more'}`},
		// map[string]interface{}
		{map[string]interface{}{}, `{}`},
		{map[string]interface{}(nil), `{}`},
		{map[string]interface{}{"foo": "bar"}, `{"foo" = 'bar'}`},
		{map[string]interface{}{"foo": "bar", "one": "more"}, `{"foo" = 'bar', "one" = 'more'}`},
		{map[string]interface{}{"foo": map[string]interface{}{"one": "more"}}, `{"foo" = {"one" = 'more'}}`},
		{map[string]interface{}{`fo"o`: `b'ar`, `ab'c`: `xy"z`, `on"""e`: `mo'''re`}, `{"ab'c" = 'xy"z', "fo""o" = 'b''ar', "on""""""e" = 'mo''''''re'}`},
	}
}

func TestEscapeValueIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := createTestContainer(t)
	defer container.Terminate()
	url := fmt.Sprintf("postgres://crate@%s:%s/test", container.Address, container.Ports[servicePort])

	db, err := sql.Open("pgx", url)
	require.NoError(t, err)
	defer db.Close()

	tests := escapeValueTests()
	for _, test := range tests {
		got, err := escapeValue(test.Value, "_")
		require.NoError(t, err, "value: %#v", test.Value)

		// This is a smoke test that will blow up if our escaping causing a SQL
		// syntax error, which may allow for an attack.=
		var reply interface{}
		row := db.QueryRow("SELECT ?", got)
		require.NoError(t, row.Scan(&reply))
	}
}

func TestEscapeValue(t *testing.T) {
	tests := escapeValueTests()
	for _, test := range tests {
		got, err := escapeValue(test.Value, "_")
		require.NoError(t, err, "value: %#v", test.Value)
		require.Equal(t, test.Want, got)
	}
}

func TestCircumventingStringEscape(t *testing.T) {
	value, err := escapeObject(map[string]interface{}{"a.b": "c"}, `_"`)
	require.NoError(t, err)
	require.Equal(t, `{"a_""b" = 'c'}`, value)
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
		m := metric.New(
			test.Name,
			test.Tags,
			test.Fields,
			time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
		)
		if got := hashID(m); got != test.Want {
			t.Errorf("test #%d: got=%d want=%d", i, got, test.Want)
		}
	}
}
