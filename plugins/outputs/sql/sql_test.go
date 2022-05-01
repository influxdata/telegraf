package sql

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestSqlQuote(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
}

func TestSqlCreateStatement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
}

func TestSqlInsertStatement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
}

func pwgen(n int) string {
	charset := []byte("abcdedfghijklmnopqrstABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	nchars := len(charset)
	buffer := make([]byte, n)

	for i := range buffer {
		buffer[i] = charset[rand.Intn(nchars)]
	}

	return string(buffer)
}

func stableMetric(
	name string,
	tags []telegraf.Tag,
	fields []telegraf.Field,
	tm time.Time,
	tp ...telegraf.ValueType,
) telegraf.Metric {
	// We want to compare the output of this plugin with expected
	// output. Maps don't preserve order so comparison fails. There's
	// no metric constructor that takes a slice of tag and slice of
	// field, just the one that takes maps.
	//
	// To preserve order, construct the metric without tags and fields
	// and then add them using AddTag and AddField.  Those are stable.
	m := metric.New(name, map[string]string{}, map[string]interface{}{}, tm, tp...)
	for _, tag := range tags {
		m.AddTag(tag.Key, tag.Value)
	}
	for _, field := range fields {
		m.AddField(field.Key, field.Value)
	}
	return m
}

var (
	// 2021-05-17T22:04:45+00:00
	// or 2021-05-17T16:04:45-06:00
	ts = time.Unix(1621289085, 0).UTC()

	testMetrics = []telegraf.Metric{
		stableMetric(
			"metric_one",
			[]telegraf.Tag{
				{
					Key:   "tag_one",
					Value: "tag1",
				},
				{
					Key:   "tag_two",
					Value: "tag2",
				},
			},
			[]telegraf.Field{
				{
					Key:   "int64_one",
					Value: int64(1234),
				},
				{
					Key:   "int64_two",
					Value: int64(2345),
				},
				{
					Key:   "bool_one",
					Value: true,
				},
				{
					Key:   "bool_two",
					Value: false,
				},
				{
					Key:   "uint64_one",
					Value: uint64(1000000000),
				},
				{
					Key:   "float64_one",
					Value: float64(3.1415),
				},
			},
			ts,
		),
		stableMetric(
			"metric_two",
			[]telegraf.Tag{
				{
					Key:   "tag_three",
					Value: "tag3",
				},
			},
			[]telegraf.Field{
				{
					Key:   "string_one",
					Value: "string1",
				},
			},
			ts,
		),
		stableMetric( //test spaces in metric, tag, and field names
			"metric three",
			[]telegraf.Tag{
				{
					Key:   "tag four",
					Value: "tag4",
				},
			},
			[]telegraf.Field{
				{
					Key:   "string two",
					Value: "string2",
				},
			},
			ts,
		),
	}
)

func TestMysqlIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	initdb, err := filepath.Abs("testdata/mariadb/initdb")
	require.NoError(t, err)

	// initdb/script.sql creates this database
	const dbname = "foo"

	// The mariadb image lets you set the root password through an env
	// var. We'll use root to insert and query test data.
	const username = "root"

	password := pwgen(32)
	outDir := t.TempDir()

	ctx := context.Background()
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "mariadb",
			Env: map[string]string{
				"MARIADB_ROOT_PASSWORD": password,
			},
			BindMounts: map[string]string{
				initdb: "/docker-entrypoint-initdb.d",
				outDir: "/out",
			},
			ExposedPorts: []string{"3306/tcp"},
			WaitingFor:   wait.ForListeningPort("3306/tcp"),
		},
		Started: true,
	}
	mariadbContainer, err := testcontainers.GenericContainer(ctx, req)
	require.NoError(t, err, "starting container failed")
	defer func() {
		require.NoError(t, mariadbContainer.Terminate(ctx), "terminating container failed")
	}()

	// Get the connection details from the container
	host, err := mariadbContainer.Host(ctx)
	require.NoError(t, err, "getting container host address failed")
	require.NotEmpty(t, host)
	natPort, err := mariadbContainer.MappedPort(ctx, "3306/tcp")
	require.NoError(t, err, "getting container host port failed")
	port := natPort.Port()
	require.NotEmpty(t, port)

	//use the plugin to write to the database
	address := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v",
		username, password, host, port, dbname,
	)
	p := newSQL()
	p.Log = testutil.Logger{}
	p.Driver = "mysql"
	p.DataSourceName = address
	//p.Convert.Timestamp = "TEXT" //disable mysql default current_timestamp()
	p.InitSQL = "SET sql_mode='ANSI_QUOTES';"

	require.NoError(t, p.Connect())
	require.NoError(t, p.Write(
		testMetrics,
	))

	//dump the database
	var rc int
	rc, err = mariadbContainer.Exec(ctx, []string{
		"bash",
		"-c",
		"mariadb-dump --user=" + username +
			" --password=" + password +
			" --compact --skip-opt " +
			dbname +
			" > /out/dump",
	})
	require.NoError(t, err)
	require.Equal(t, 0, rc)
	dumpfile := filepath.Join(outDir, "dump")
	require.FileExists(t, dumpfile)

	//compare the dump to what we expected
	expected, err := os.ReadFile("testdata/mariadb/expected.sql")
	require.NoError(t, err)
	actual, err := os.ReadFile(dumpfile)
	require.NoError(t, err)
	require.Equal(t, string(expected), string(actual))
}

func TestPostgresIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	initdb, err := filepath.Abs("testdata/postgres/initdb")
	require.NoError(t, err)

	// initdb/init.sql creates this database
	const dbname = "foo"

	// default username for postgres is postgres
	const username = "postgres"

	password := pwgen(32)
	outDir := t.TempDir()

	ctx := context.Background()
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "postgres",
			Env: map[string]string{
				"POSTGRES_PASSWORD": password,
			},
			BindMounts: map[string]string{
				initdb: "/docker-entrypoint-initdb.d",
				outDir: "/out",
			},
			ExposedPorts: []string{"5432/tcp"},
			WaitingFor:   wait.ForListeningPort("5432/tcp"),
		},
		Started: true,
	}
	cont, err := testcontainers.GenericContainer(ctx, req)
	require.NoError(t, err, "starting container failed")
	defer func() {
		require.NoError(t, cont.Terminate(ctx), "terminating container failed")
	}()

	// Get the connection details from the container
	host, err := cont.Host(ctx)
	require.NoError(t, err, "getting container host address failed")
	require.NotEmpty(t, host)
	natPort, err := cont.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err, "getting container host port failed")
	port := natPort.Port()
	require.NotEmpty(t, port)

	//use the plugin to write to the database
	// host, port, username, password, dbname
	address := fmt.Sprintf("postgres://%v:%v@%v:%v/%v",
		username, password, host, port, dbname,
	)
	p := newSQL()
	p.Log = testutil.Logger{}
	p.Driver = "pgx"
	p.DataSourceName = address
	p.Convert.Real = "double precision"
	p.Convert.Unsigned = "bigint"
	p.Convert.ConversionStyle = "literal"

	require.NoError(t, p.Connect())
	require.NoError(t, p.Write(
		testMetrics,
	))

	//dump the database
	//psql -u postgres
	var rc int
	rc, err = cont.Exec(ctx, []string{
		"bash",
		"-c",
		"pg_dump" +
			" --username=" + username +
			//" --password=" + password +
			//			" --compact --skip-opt " +
			" --no-comments" +
			//" --data-only" +
			" " + dbname +
			// pg_dump's output has comments that include build info
			// of postgres and pg_dump. The build info changes with
			// each release. To prevent these changes from causing the
			// test to fail, we strip out comments. Also strip out
			// blank lines.
			"|grep -E -v '(^--|^$)'" +
			" > /out/dump 2>&1",
	})
	require.NoError(t, err)
	require.Equal(t, 0, rc)
	dumpfile := filepath.Join(outDir, "dump")
	require.FileExists(t, dumpfile)

	//compare the dump to what we expected
	expected, err := os.ReadFile("testdata/postgres/expected.sql")
	require.NoError(t, err)
	actual, err := os.ReadFile(dumpfile)
	require.NoError(t, err)
	require.Equal(t, string(expected), string(actual))
}

func TestClickHouseIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	initdb, err := filepath.Abs("testdata/clickhouse/initdb")
	// confd, err := filepath.Abs("testdata/clickhouse/config.d")
	require.NoError(t, err)

	// initdb/init.sql creates this database
	const dbname = "foo"

	// default username for clickhouse is default
	const username = "default"

	outDir := t.TempDir()

	ctx := context.Background()
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "yandex/clickhouse-server",
			BindMounts: map[string]string{
				initdb: "/docker-entrypoint-initdb.d",
				outDir: "/out",
			},
			ExposedPorts: []string{"9000/tcp", "8123/tcp"},
			WaitingFor:   wait.NewHTTPStrategy("/").WithPort("8123/tcp"),
		},
		Started: true,
	}
	cont, err := testcontainers.GenericContainer(ctx, req)
	require.NoError(t, err, "starting container failed")
	defer func() {
		require.NoError(t, cont.Terminate(ctx), "terminating container failed")
	}()

	// Get the connection details from the container
	host, err := cont.Host(ctx)
	require.NoError(t, err, "getting container host address failed")
	require.NotEmpty(t, host)
	natPort, err := cont.MappedPort(ctx, "9000/tcp")
	require.NoError(t, err, "getting container host port failed")
	port := natPort.Port()
	require.NotEmpty(t, port)

	//use the plugin to write to the database
	// host, port, username, password, dbname
	address := fmt.Sprintf("tcp://%v:%v?username=%v&database=%v", host, port, username, dbname)
	p := newSQL()
	p.Log = testutil.Logger{}
	p.Driver = "clickhouse"
	p.DataSourceName = address
	p.TableTemplate = "CREATE TABLE {TABLE}({COLUMNS}) ENGINE MergeTree() ORDER by timestamp"
	p.Convert.Integer = "Int64"
	p.Convert.Text = "String"
	p.Convert.Timestamp = "DateTime"
	p.Convert.Defaultvalue = "String"
	p.Convert.Unsigned = "UInt64"
	p.Convert.Bool = "UInt8"
	p.Convert.ConversionStyle = "literal"

	require.NoError(t, p.Connect())

	require.NoError(t, p.Write(testMetrics))

	// dump the database
	var rc int
	for _, testMetric := range testMetrics {
		rc, err = cont.Exec(ctx, []string{
			"bash",
			"-c",
			"clickhouse-client" +
				" --user=" + username +
				" --database=" + dbname +
				" --format=TabSeparatedRaw" +
				" --multiquery --query=" +
				"\"SELECT * FROM \\\"" + testMetric.Name() + "\\\";" +
				"SHOW CREATE TABLE \\\"" + testMetric.Name() + "\\\"\"" +
				" >> /out/dump 2>&1",
		})
		require.NoError(t, err)
		require.Equal(t, 0, rc)
	}

	dumpfile := filepath.Join(outDir, "dump")
	require.FileExists(t, dumpfile)

	//compare the dump to what we expected
	expected, err := os.ReadFile("testdata/clickhouse/expected.txt")
	require.NoError(t, err)
	actual, err := os.ReadFile(dumpfile)
	require.NoError(t, err)
	require.Equal(t, string(expected), string(actual))
}
