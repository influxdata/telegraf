package zerobus

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func TestConnectAndWriteIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	requiredEnvironment := []string{
		"ZEROBUS_SERVER_ENDPOINT",
		"DATABRICKS_WORKSPACE_URL",
		"ZEROBUS_TABLE_NAME",
		"DATABRICKS_CLIENT_ID",
		"DATABRICKS_CLIENT_SECRET",
	}
	values := make(map[string]string, len(requiredEnvironment))
	for _, name := range requiredEnvironment {
		value := os.Getenv(name)
		if value == "" {
			t.Skipf("Skipping integration test because %s is not set", name)
		}
		values[name] = value
	}
	schemaMode := os.Getenv("ZEROBUS_SCHEMA_MODE")
	if schemaMode == "" {
		schemaMode = schemaModeStatic
	}
	timestampColumn := os.Getenv("ZEROBUS_TIMESTAMP_COLUMN")
	if timestampColumn == "" {
		timestampColumn = "timestamp"
	}

	plugin := &Zerobus{
		ServerEndpoint:    values["ZEROBUS_SERVER_ENDPOINT"],
		WorkspaceURL:      values["DATABRICKS_WORKSPACE_URL"],
		TableName:         values["ZEROBUS_TABLE_NAME"],
		ClientID:          values["DATABRICKS_CLIENT_ID"],
		ClientSecret:      config.NewSecret([]byte(values["DATABRICKS_CLIENT_SECRET"])),
		ApplicationName:   "telegraf-integration-test",
		SchemaMode:        schemaMode,
		TimestampColumn:   timestampColumn,
		MeasurementColumn: os.Getenv("ZEROBUS_MEASUREMENT_COLUMN"),
		Log:               testutil.Logger{},
	}
	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Connect())
	t.Cleanup(func() { require.NoError(t, plugin.Close()) })

	input := metric.New(
		"zerobus_integration",
		map[string]string{"source": "telegraf-test"},
		map[string]interface{}{
			"integer":  int64(-42),
			"unsigned": uint64(42),
			"float":    1.25,
			"boolean":  true,
			"string":   "ready",
		},
		time.Now(),
		telegraf.Gauge,
	)
	require.NoError(t, plugin.Write([]telegraf.Metric{input}))
}
