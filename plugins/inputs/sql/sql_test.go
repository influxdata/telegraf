package sql

import (
	"context"
	"flag"
	"fmt"
	"testing"
	"time"

	"math/rand"
	"path/filepath"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
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

var spinup = flag.Bool("spinup", false, "Spin-up the required test containers")

func TestMariaDB(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := testutil.Logger{}

	addr := "127.0.0.1"
	port := "3306"
	passwd := ""
	database := "foo"

	if *spinup {
		logger.Infof("Spinning up container...")

		// Generate a random password
		passwd = pwgen(32)

		// Determine the test-data mountpoint
		testdata, err := filepath.Abs("testdata/mariadb")
		require.NoError(t, err, "determining absolute path of test-data failed")

		// Spin-up the container
		ctx := context.Background()
		req := testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image: "mariadb",
				Env: map[string]string{
					"MYSQL_ROOT_PASSWORD": passwd,
					"MYSQL_DATABASE":      database,
				},
				BindMounts: map[string]string{
					testdata: "/docker-entrypoint-initdb.d",
				},
				ExposedPorts: []string{"3306/tcp"},
				WaitingFor:   wait.ForListeningPort("3306/tcp"),
			},
			Started: true,
		}
		container, err := testcontainers.GenericContainer(ctx, req)
		require.NoError(t, err, "starting container failed")
		defer func() {
			require.NoError(t, container.Terminate(ctx), "terminating container failed")
		}()

		// Get the connection details from the container
		addr, err = container.Host(ctx)
		require.NoError(t, err, "getting container host address failed")
		p, err := container.MappedPort(ctx, "3306/tcp")
		require.NoError(t, err, "getting container host port failed")
		port = p.Port()
	}

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
				Driver:  "maria",
				Dsn:     fmt.Sprintf("root:%s@tcp(%s:%s)/%s", passwd, addr, port, database),
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

func TestPostgreSQL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := testutil.Logger{}

	addr := "127.0.0.1"
	port := "5432"
	passwd := ""
	database := "foo"

	if *spinup {
		logger.Infof("Spinning up container...")

		// Generate a random password
		passwd = pwgen(32)

		// Determine the test-data mountpoint
		testdata, err := filepath.Abs("testdata/postgres")
		require.NoError(t, err, "determining absolute path of test-data failed")

		// Spin-up the container
		ctx := context.Background()
		req := testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image: "postgres",
				Env: map[string]string{
					"POSTGRES_PASSWORD": passwd,
					"POSTGRES_DB":       database,
				},
				BindMounts: map[string]string{
					testdata: "/docker-entrypoint-initdb.d",
				},
				ExposedPorts: []string{"5432/tcp"},
				WaitingFor:   wait.ForListeningPort("5432/tcp"),
			},
			Started: true,
		}
		container, err := testcontainers.GenericContainer(ctx, req)
		require.NoError(t, err, "starting container failed")
		defer func() {
			require.NoError(t, container.Terminate(ctx), "terminating container failed")
		}()

		// Get the connection details from the container
		addr, err = container.Host(ctx)
		require.NoError(t, err, "getting container host address failed")
		p, err := container.MappedPort(ctx, "5432/tcp")
		require.NoError(t, err, "getting container host port failed")
		port = p.Port()
	}

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
				Driver:  "pgx",
				Dsn:     fmt.Sprintf("postgres://postgres:%v@%v:%v/%v", passwd, addr, port, database),
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

func TestClickHouse(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := testutil.Logger{}

	addr := "127.0.0.1"
	port := "9000"
	user := "default"

	if *spinup {
		logger.Infof("Spinning up container...")

		// Determine the test-data mountpoint
		testdata, err := filepath.Abs("testdata/clickhouse")
		require.NoError(t, err, "determining absolute path of test-data failed")

		// Spin-up the container
		ctx := context.Background()
		req := testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image: "yandex/clickhouse-server",
				BindMounts: map[string]string{
					testdata: "/docker-entrypoint-initdb.d",
				},
				ExposedPorts: []string{"9000/tcp", "8123/tcp"},
				WaitingFor:   wait.NewHTTPStrategy("/").WithPort("8123/tcp"),
			},
			Started: true,
		}
		container, err := testcontainers.GenericContainer(ctx, req)
		require.NoError(t, err, "starting container failed")
		defer func() {
			require.NoError(t, container.Terminate(ctx), "terminating container failed")
		}()

		// Get the connection details from the container
		addr, err = container.Host(ctx)
		require.NoError(t, err, "getting container host address failed")
		p, err := container.MappedPort(ctx, "9000/tcp")
		require.NoError(t, err, "getting container host port failed")
		port = p.Port()
	}

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
				Driver:  "clickhouse",
				Dsn:     fmt.Sprintf("tcp://%v:%v?username=%v", addr, port, user),
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
