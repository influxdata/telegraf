package postgresql

import (
	"fmt"
	"strings"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
)

const servicePort = "5432"

func launchTestContainer(t *testing.T) *testutil.Container {
	container := testutil.Container{
		Image:        "postgres:alpine",
		ExposedPorts: []string{servicePort},
		Env: map[string]string{
			"POSTGRES_HOST_AUTH_METHOD": "trust",
		},
		WaitingFor: wait.ForAll(
			// the database comes up twice, once right away, then again a second
			// time after the docker entrypoint starts configuration
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
			wait.ForListeningPort(nat.Port(servicePort)),
		),
	}

	err := container.Start()
	require.NoError(t, err, "failed to start container")

	return &container
}

func TestPostgresqlGeneratesMetricsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := launchTestContainer(t)
	defer container.Terminate()

	addr := fmt.Sprintf(
		"host=%s port=%s user=postgres sslmode=disable",
		container.Address,
		container.Ports[servicePort],
	)

	p := &Postgresql{
		Service: Service{
			Address:     config.NewSecret([]byte(addr)),
			IsPgBouncer: false,
		},
		Databases: []string{"postgres"},
	}

	var acc testutil.Accumulator
	require.NoError(t, p.Start(&acc))
	require.NoError(t, p.Gather(&acc))

	intMetrics := []string{
		"xact_commit",
		"xact_rollback",
		"blks_read",
		"blks_hit",
		"tup_returned",
		"tup_fetched",
		"tup_inserted",
		"tup_updated",
		"tup_deleted",
		"conflicts",
		"temp_files",
		"temp_bytes",
		"deadlocks",
		"buffers_alloc",
		"buffers_backend",
		"buffers_backend_fsync",
		"buffers_checkpoint",
		"buffers_clean",
		"checkpoints_req",
		"checkpoints_timed",
		"maxwritten_clean",
		"datid",
		"numbackends",
	}

	int32Metrics := []string{}

	floatMetrics := []string{
		"blk_read_time",
		"blk_write_time",
		"checkpoint_write_time",
		"checkpoint_sync_time",
	}

	stringMetrics := []string{
		"datname",
	}

	metricsCounted := 0

	for _, metric := range intMetrics {
		require.True(t, acc.HasInt64Field("postgresql", metric))
		metricsCounted++
	}

	for _, metric := range int32Metrics {
		require.True(t, acc.HasInt32Field("postgresql", metric))
		metricsCounted++
	}

	for _, metric := range floatMetrics {
		require.True(t, acc.HasFloatField("postgresql", metric))
		metricsCounted++
	}

	for _, metric := range stringMetrics {
		require.True(t, acc.HasStringField("postgresql", metric))
		metricsCounted++
	}

	require.Greater(t, metricsCounted, 0)
	require.Equal(t, len(floatMetrics)+len(intMetrics)+len(int32Metrics)+len(stringMetrics), metricsCounted)
}

func TestPostgresqlTagsMetricsWithDatabaseNameIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := launchTestContainer(t)
	defer container.Terminate()

	addr := fmt.Sprintf(
		"host=%s port=%s user=postgres sslmode=disable",
		container.Address,
		container.Ports[servicePort],
	)

	p := &Postgresql{
		Service: Service{
			Address: config.NewSecret([]byte(addr)),
		},
		Databases: []string{"postgres"},
	}

	var acc testutil.Accumulator

	require.NoError(t, p.Start(&acc))
	require.NoError(t, p.Gather(&acc))

	point, ok := acc.Get("postgresql")
	require.True(t, ok)

	require.Equal(t, "postgres", point.Tags["db"])
}

func TestPostgresqlDefaultsToAllDatabasesIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := launchTestContainer(t)
	defer container.Terminate()

	addr := fmt.Sprintf(
		"host=%s port=%s user=postgres sslmode=disable",
		container.Address,
		container.Ports[servicePort],
	)

	p := &Postgresql{
		Service: Service{
			Address: config.NewSecret([]byte(addr)),
		},
	}

	var acc testutil.Accumulator

	require.NoError(t, p.Start(&acc))
	require.NoError(t, p.Gather(&acc))

	var found bool

	for _, pnt := range acc.Metrics {
		if pnt.Measurement == "postgresql" {
			if pnt.Tags["db"] == "postgres" {
				found = true
				break
			}
		}
	}

	require.True(t, found)
}

func TestPostgresqlIgnoresUnwantedColumnsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := launchTestContainer(t)
	defer container.Terminate()

	addr := fmt.Sprintf(
		"host=%s port=%s user=postgres sslmode=disable",
		container.Address,
		container.Ports[servicePort],
	)

	p := &Postgresql{
		Service: Service{
			Address: config.NewSecret([]byte(addr)),
		},
	}

	var acc testutil.Accumulator
	require.NoError(t, p.Start(&acc))
	require.NoError(t, p.Gather(&acc))

	for col := range p.IgnoredColumns() {
		require.False(t, acc.HasMeasurement(col))
	}
}

func TestPostgresqlDatabaseWhitelistTestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := launchTestContainer(t)
	defer container.Terminate()

	addr := fmt.Sprintf(
		"host=%s port=%s user=postgres sslmode=disable",
		container.Address,
		container.Ports[servicePort],
	)

	p := &Postgresql{
		Service: Service{
			Address: config.NewSecret([]byte(addr)),
		},
		Databases: []string{"template0"},
	}

	var acc testutil.Accumulator

	require.NoError(t, p.Start(&acc))
	require.NoError(t, p.Gather(&acc))

	var foundTemplate0 = false
	var foundTemplate1 = false

	for _, pnt := range acc.Metrics {
		if pnt.Measurement == "postgresql" {
			if pnt.Tags["db"] == "template0" {
				foundTemplate0 = true
			}
		}
		if pnt.Measurement == "postgresql" {
			if pnt.Tags["db"] == "template1" {
				foundTemplate1 = true
			}
		}
	}

	require.True(t, foundTemplate0)
	require.False(t, foundTemplate1)
}

func TestPostgresqlDatabaseBlacklistTestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := launchTestContainer(t)
	defer container.Terminate()

	addr := fmt.Sprintf(
		"host=%s port=%s user=postgres sslmode=disable",
		container.Address,
		container.Ports[servicePort],
	)

	p := &Postgresql{
		Service: Service{
			Address: config.NewSecret([]byte(addr)),
		},
		IgnoredDatabases: []string{"template0"},
	}

	var acc testutil.Accumulator
	require.NoError(t, p.Start(&acc))
	require.NoError(t, p.Gather(&acc))

	var foundTemplate0 = false
	var foundTemplate1 = false

	for _, pnt := range acc.Metrics {
		if pnt.Measurement == "postgresql" {
			if pnt.Tags["db"] == "template0" {
				foundTemplate0 = true
			}
		}
		if pnt.Measurement == "postgresql" {
			if pnt.Tags["db"] == "template1" {
				foundTemplate1 = true
			}
		}
	}

	require.False(t, foundTemplate0)
	require.True(t, foundTemplate1)
}

func TestSanitizeAddressKeyValue(t *testing.T) {
	keys := []string{"password", "sslcert", "sslkey", "sslmode", "sslrootcert"}
	tests := []struct {
		name  string
		value string
	}{
		{
			name:  "simple text",
			value: `foo`,
		},
		{
			name:  "empty values",
			value: `''`,
		},
		{
			name:  "space in value",
			value: `'foo bar'`,
		},
		{
			name:  "escaped quote",
			value: `'foo\'s bar'`,
		},
		{
			name:  "escaped quote no space",
			value: `\'foobar\'s\'`,
		},
		{
			name:  "escaped backslash",
			value: `'foo bar\\'`,
		},
		{
			name:  "escaped quote and backslash",
			value: `'foo\\\'s bar'`,
		},
		{
			name:  "two escaped backslashes",
			value: `'foo bar\\\\'`,
		},
	}

	for _, tt := range tests {
		// Key value without spaces around equal sign
		t.Run(tt.name, func(t *testing.T) {
			// Generate the DSN from the given keys and value
			parts := make([]string, 0, len(keys))
			for _, k := range keys {
				parts = append(parts, k+"="+tt.value)
			}
			dsn := strings.Join(parts, " canary=ok ")

			plugin := &Postgresql{
				Service: Service{
					Address: config.NewSecret([]byte(dsn)),
				},
			}

			expected := strings.Join(make([]string, len(keys)), "canary=ok ")
			expected = strings.TrimSpace(expected)
			actual, err := plugin.SanitizedAddress()
			require.NoError(t, err)
			require.Equalf(t, expected, actual, "initial: %s", dsn)
		})

		// Key value with spaces around equal sign
		t.Run("spaced "+tt.name, func(t *testing.T) {
			// Generate the DSN from the given keys and value
			parts := make([]string, 0, len(keys))
			for _, k := range keys {
				parts = append(parts, k+" = "+tt.value)
			}
			dsn := strings.Join(parts, " canary=ok ")

			plugin := &Postgresql{
				Service: Service{
					Address: config.NewSecret([]byte(dsn)),
				},
			}

			expected := strings.Join(make([]string, len(keys)), "canary=ok ")
			expected = strings.TrimSpace(expected)
			actual, err := plugin.SanitizedAddress()
			require.NoError(t, err)
			require.Equalf(t, expected, actual, "initial: %s", dsn)
		})
	}
}

func TestSanitizeAddressURI(t *testing.T) {
	keys := []string{"password", "sslcert", "sslkey", "sslmode", "sslrootcert"}
	tests := []struct {
		name  string
		value string
	}{
		{
			name:  "simple text",
			value: `foo`,
		},
		{
			name:  "empty values",
			value: ``,
		},
		{
			name:  "space in value",
			value: `foo bar`,
		},
		{
			name:  "equal sign in value",
			value: `foo=bar`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate the DSN from the given keys and value
			value := strings.ReplaceAll(tt.value, "=", "%3D")
			value = strings.ReplaceAll(value, " ", "%20")
			parts := make([]string, 0, len(keys))
			for _, k := range keys {
				parts = append(parts, k+"="+value)
			}
			dsn := "postgresql://user:passwd@localhost:5432/db?" + strings.Join(parts, "&")

			plugin := &Postgresql{
				Service: Service{
					Address: config.NewSecret([]byte(dsn)),
				},
			}

			expected := "dbname=db host=localhost port=5432 user=user"
			actual, err := plugin.SanitizedAddress()
			require.NoError(t, err)
			require.Equalf(t, expected, actual, "initial: %s", dsn)
		})
	}
}
