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

func TestMysql(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := testutil.Logger{}

	addr := "127.0.0.1"
	port := "3306"
	passwd := ""
	database := "nation"

	if *spinup {
		logger.Infof("Spinning up container...")

		// Generate a random password
		passwd = pwgen(32)

		// Determine the test-data mountpoint
		testdata, err := filepath.Abs("testdata")
		require.NoError(t, err, "determining absolute path of test-data failed")

		// Spin-up the container
		ctx := context.Background()
		req := testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image: "mariadb",
				Env: map[string]string{
					"MYSQL_ROOT_PASSWORD": passwd,
				},
				BindMounts: map[string]string{
					testdata: "/docker-entrypoint-initdb.d",
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
		addr, err = mariadbContainer.Host(ctx)
		require.NoError(t, err, "getting container host address failed")
		p, err := mariadbContainer.MappedPort(ctx, "3306/tcp")
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
			name: "guests",
			queries: []Query{
				{
					Query:               "SELECT * FROM guests",
					TagColumnsInclude:   []string{"name"},
					FieldColumnsExclude: []string{"name"},
				},
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"sql",
					map[string]string{"name": "John"},
					map[string]interface{}{"guest_id": int64(1)},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"sql",
					map[string]string{"name": "Jane"},
					map[string]interface{}{"guest_id": int64(2)},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"sql",
					map[string]string{"name": "Jean"},
					map[string]interface{}{"guest_id": int64(3)},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"sql",
					map[string]string{"name": "Storm"},
					map[string]interface{}{"guest_id": int64(4)},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"sql",
					map[string]string{"name": "Beast"},
					map[string]interface{}{"guest_id": int64(5)},
					time.Unix(0, 0),
				),
			},
		},
	}

	for _, tt := range testset {
		t.Run(tt.name, func(t *testing.T) {
			// Setup the plugin-under-test
			plugin := &SQL{
				Driver:  "mysql",
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

			// Stopping the plugin
			plugin.Stop()

			// Do the comparison
			testutil.RequireMetricsEqual(t, tt.expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
		})
	}
}
