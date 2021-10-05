package sqlserver

import (
	"os"
	"testing"
	"github.com/stretchr/testify/require"
	"github.com/influxdata/telegraf/testutil"
)

func TestAzureSQL_Database_ResourceStats_Query(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("AZURESQL_DB_CONNECTION_STRING") == "" {
		t.Skip("Missing environment variable AZURESQL_DB_CONNECTION_STRING")
	}

	connectionString := os.Getenv("AZURESQL_DB_CONNECTION_STRING")
	
	server := &SQLServer {
		Servers:      []string{connectionString},
		IncludeQuery: []string{"AzureSQLDBResourceStats"},
		AuthMethod:   "connection_string",
		DatabaseType: "AzureSQLDB",
	}

	var acc testutil.Accumulator

	require.NoError(t, server.Start(&acc))
	require.NoError(t, server.Gather(&acc))

	require.True(t, acc.HasMeasurement("sqlserver_azure_db_resource_stats"))
	require.True(t, acc.HasTag("sqlserver_azure_db_resource_stats", "sql_instance"))
	require.True(t, acc.HasTag("sqlserver_azure_db_resource_stats", "database_name"))
	require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_cpu_percent"))
	require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_data_io_percent"))
	require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_log_write_percent"))
	require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_memory_usage_percent"))
	require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "xtp_storage_percent"))
	require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "max_worker_percent"))
	require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "max_session_percent"))
	require.True(t, acc.HasField("sqlserver_azure_db_resource_stats", "dtu_limit"))	// Can be null
	require.True(t, acc.HasField("sqlserver_azure_db_resource_stats", "avg_login_rate_percent")) // Can be null. Identified for informational purposes only. Not supported. Future compatibility is not guaranteed.
	require.True(t, acc.HasField("sqlserver_azure_db_resource_stats", "end_time")) // Time
	require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_instance_memory_percent"))
	require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_instance_cpu_percent"))
	require.True(t, acc.HasTag("sqlserver_azure_db_resource_stats", "replica_updateability"))

	// This query should only return one row
	require.Equal(t, 1, len(acc.Metrics))
	server.Stop()
}