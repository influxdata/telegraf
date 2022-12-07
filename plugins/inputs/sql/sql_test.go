package sql

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

func pwgen(n int) string {
	charset := []byte("abcdedfghijklmnopqrstABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	nchars := len(charset)
	buffer := make([]byte, n)

	for i := range buffer {
		buffer[i] = charset[rand.Intn(nchars)]
	}

	return string(buffer)
}

func TestMariaDBIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := testutil.Logger{}

	port := "3306"
	passwd := pwgen(32)
	database := "foo"

	// Determine the test-data mountpoint
	testdata, err := filepath.Abs("testdata/mariadb")
	require.NoError(t, err, "determining absolute path of test-data failed")

	container := testutil.Container{
		Image:        "mariadb",
		ExposedPorts: []string{port},
		Env: map[string]string{
			"MYSQL_ROOT_PASSWORD": passwd,
			"MYSQL_DATABASE":      database,
		},
		BindMounts: map[string]string{
			"/docker-entrypoint-initdb.d": testdata,
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("mariadbd: ready for connections.").WithOccurrence(2),
			wait.ForListeningPort(nat.Port(port)),
		),
	}
	err = container.Start()
	require.NoError(t, err, "failed to start container")
	defer container.Terminate()

	// Define the testset
	var testset = []struct {
		name     string
		queries  []Query
		expected []telegraf.Metric
	}{
		{
			name: "metric_one",
			queries: []Query{
				{
					Query:               "SELECT * FROM metric_one",
					TagColumnsInclude:   []string{"tag_*"},
					FieldColumnsExclude: []string{"tag_*", "timestamp"},
					TimeColumn:          "timestamp",
					TimeFormat:          "2006-01-02 15:04:05",
				},
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"sql",
					map[string]string{
						"tag_one": "tag1",
						"tag_two": "tag2",
					},
					map[string]interface{}{
						"int64_one": int64(1234),
						"int64_two": int64(2345),
					},
					time.Date(2021, 5, 17, 22, 4, 45, 0, time.UTC),
				),
			},
		},
	}

	for _, tt := range testset {
		t.Run(tt.name, func(t *testing.T) {
			// Setup the plugin-under-test
			plugin := &SQL{
				Driver: "maria",
				Dsn: fmt.Sprintf("root:%s@tcp(%s:%s)/%s",
					passwd,
					container.Address,
					container.Ports[port],
					database,
				),
				Queries: tt.queries,
				Log:     logger,
			}

			var acc testutil.Accumulator

			// Startup the plugin
			err := plugin.Init()
			require.NoError(t, err)
			err = plugin.Start(&acc)
			require.NoError(t, err)

			// Gather
			err = plugin.Gather(&acc)
			require.NoError(t, err)
			require.Len(t, acc.Errors, 0)

			// Stopping the plugin
			plugin.Stop()

			// Do the comparison
			testutil.RequireMetricsEqual(t, tt.expected, acc.GetTelegrafMetrics())
		})
	}
}

func TestPostgreSQLIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := testutil.Logger{}

	port := "5432"
	passwd := pwgen(32)
	database := "foo"

	// Determine the test-data mountpoint
	testdata, err := filepath.Abs("testdata/postgres")
	require.NoError(t, err, "determining absolute path of test-data failed")

	container := testutil.Container{
		Image:        "postgres",
		ExposedPorts: []string{port},
		Env: map[string]string{
			"POSTGRES_PASSWORD": passwd,
			"POSTGRES_DB":       database,
		},
		BindMounts: map[string]string{
			"/docker-entrypoint-initdb.d": testdata,
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
			wait.ForListeningPort(nat.Port(port)),
		),
	}
	err = container.Start()
	require.NoError(t, err, "failed to start container")
	defer container.Terminate()

	// Define the testset
	var testset = []struct {
		name     string
		queries  []Query
		expected []telegraf.Metric
	}{
		{
			name: "metric_one",
			queries: []Query{
				{
					Query:               "SELECT * FROM metric_one",
					TagColumnsInclude:   []string{"tag_*"},
					FieldColumnsExclude: []string{"tag_*", "timestamp"},
					TimeColumn:          "timestamp",
					TimeFormat:          "2006-01-02 15:04:05",
				},
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"sql",
					map[string]string{
						"tag_one": "tag1",
						"tag_two": "tag2",
					},
					map[string]interface{}{
						"int64_one": int64(1234),
						"int64_two": int64(2345),
					},
					time.Date(2021, 5, 17, 22, 4, 45, 0, time.UTC),
				),
			},
		},
	}

	for _, tt := range testset {
		t.Run(tt.name, func(t *testing.T) {
			// Setup the plugin-under-test
			plugin := &SQL{
				Driver: "pgx",
				Dsn: fmt.Sprintf("postgres://postgres:%v@%v:%v/%v",
					passwd,
					container.Address,
					container.Ports[port],
					database,
				),
				Queries: tt.queries,
				Log:     logger,
			}

			var acc testutil.Accumulator

			// Startup the plugin
			err := plugin.Init()
			require.NoError(t, err)
			err = plugin.Start(&acc)
			require.NoError(t, err)

			// Gather
			err = plugin.Gather(&acc)
			require.NoError(t, err)
			require.Len(t, acc.Errors, 0)

			// Stopping the plugin
			plugin.Stop()

			// Do the comparison
			testutil.RequireMetricsEqual(t, tt.expected, acc.GetTelegrafMetrics())
		})
	}
}

func TestClickHouseIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := testutil.Logger{}

	port := "9000"
	user := "default"

	// Determine the test-data mountpoint
	testdata, err := filepath.Abs("testdata/clickhouse")
	require.NoError(t, err, "determining absolute path of test-data failed")

	container := testutil.Container{
		Image:        "yandex/clickhouse-server",
		ExposedPorts: []string{port, "8123"},
		BindMounts: map[string]string{
			"/docker-entrypoint-initdb.d": testdata,
		},
		WaitingFor: wait.ForAll(
			wait.NewHTTPStrategy("/").WithPort(nat.Port("8123")),
			wait.ForListeningPort(nat.Port(port)),
			wait.ForLog("Saved preprocessed configuration to '/var/lib/clickhouse/preprocessed_configs/users.xml'.").WithOccurrence(2),
		),
	}
	err = container.Start()
	require.NoError(t, err, "failed to start container")
	defer container.Terminate()

	// Define the testset
	var testset = []struct {
		name     string
		queries  []Query
		expected []telegraf.Metric
	}{
		{
			name: "metric_one",
			queries: []Query{
				{
					Query:               "SELECT * FROM default.metric_one",
					TagColumnsInclude:   []string{"tag_*"},
					FieldColumnsExclude: []string{"tag_*", "timestamp"},
					TimeColumn:          "timestamp",
					TimeFormat:          "unix",
				},
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"sql",
					map[string]string{
						"tag_one": "tag1",
						"tag_two": "tag2",
					},
					map[string]interface{}{
						"int64_one": int64(1234),
						"int64_two": int64(2345),
					},
					time.Unix(1621289085, 0),
				),
			},
		},
	}

	for _, tt := range testset {
		t.Run(tt.name, func(t *testing.T) {
			// Setup the plugin-under-test
			plugin := &SQL{
				Driver: "clickhouse",
				Dsn: fmt.Sprintf("tcp://%v:%v?username=%v",
					container.Address,
					container.Ports[port],
					user,
				),
				Queries: tt.queries,
				Log:     logger,
			}

			var acc testutil.Accumulator

			// Startup the plugin
			err := plugin.Init()
			require.NoError(t, err)
			err = plugin.Start(&acc)
			require.NoError(t, err)

			// Gather
			err = plugin.Gather(&acc)
			require.NoError(t, err)
			require.Len(t, acc.Errors, 0)

			// Stopping the plugin
			plugin.Stop()

			// Do the comparison
			testutil.RequireMetricsEqual(t, tt.expected, acc.GetTelegrafMetrics())
		})
	}
}
