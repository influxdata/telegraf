package sql

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
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
	testMetrics = []telegraf.Metric{
		stableMetric(
			"metric_one",
			[]telegraf.Tag{
				telegraf.Tag{
					Key:   "tag_one",
					Value: "tag1",
				},
				telegraf.Tag{
					Key:   "tag_two",
					Value: "tag2",
				},
			},
			[]telegraf.Field{
				telegraf.Field{
					Key:   "int64_one",
					Value: int64(1234),
				},
				telegraf.Field{
					Key:   "int64_two",
					Value: int64(2345),
				},
			},
			time.Unix(1621289085, 0),
		),
		stableMetric(
			"metric_two",
			[]telegraf.Tag{
				telegraf.Tag{
					Key:   "tag_three",
					Value: "tag3",
				},
			},
			[]telegraf.Field{
				telegraf.Field{
					Key:   "string_one",
					Value: "string1",
				},
			},
			time.Unix(1621289085, 0),
		),
	}
)

func writeMysql(t *testing.T, host string, port string, username string, password string, database string) {
	address := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v",
		username, password, host, port, database,
	)
	p := newSql()
	p.Driver = "mysql"
	p.Address = address
	p.Convert.Timestamp = "TEXT" //disable mysql default current_timestamp()

	require.NoError(t, p.Connect())
	require.NoError(t, p.Write(
		testMetrics,
	))
}

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
	outDir, err := ioutil.TempDir("", "tg-mysql-*")
	require.NoError(t, err)
	defer os.RemoveAll(outDir)

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
	p, err := mariadbContainer.MappedPort(ctx, "3306/tcp")
	require.NoError(t, err, "getting container host port failed")
	port := p.Port()
	require.NotEmpty(t, port)

	//use the plugin to write to the database
	writeMysql(t, host, port, username, password, dbname)

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
	require.Equal(t, 0, rc)
	dumpfile := filepath.Join(outDir, "dump")
	require.FileExists(t, dumpfile)

	//compare the dump to what we expected
	expected, err := ioutil.ReadFile("testdata/mariadb/expected.sql")
	require.NoError(t, err)
	actual, err := ioutil.ReadFile(dumpfile)
	require.NoError(t, err)
	require.Equal(t, string(expected), string(actual))
}
