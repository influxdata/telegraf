package sqlserver

import (
	"os"
	"testing"
	"github.com/stretchr/testify/require"
	"github.com/influxdata/telegraf/testutil"
)

func TestAzureSQL_Managed_ResourceStats_Query(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("AZURESQL_MI_CONNECTION_STRING") == "" {
		t.Skip("Missing environment variable AZURESQL_MI_CONNECTION_STRING")
	}

	connectionString := os.Getenv("AZURESQL_MI_CONNECTION_STRING")
	
	server := &SQLServer {
		Servers:      []string{connectionString},
		IncludeQuery: []string{"AzureSQLMIResourceStats"},
		AuthMethod:   "connection_string",
		DatabaseType: "AzureSQLManagedInstance",
	}

	var acc testutil.Accumulator

	require.NoError(t, server.Start(&acc))
	require.NoError(t, server.Gather(&acc))

	// require.True(t, acc.HasMeasurement("sqlserver_azure_db_resource_stats"))
	// require.True(t, acc.HasTag("sqlserver_azure_db_resource_stats", "sql_instance"))
	// require.True(t, acc.HasTag("sqlserver_azure_db_resource_stats", "database_name"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_cpu_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_data_io_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_log_write_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_memory_usage_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "xtp_storage_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "max_worker_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "max_session_percent"))
	// require.True(t, acc.HasField("sqlserver_azure_db_resource_stats", "dtu_limit"))	// Can be null
	// require.True(t, acc.HasField("sqlserver_azure_db_resource_stats", "avg_login_rate_percent")) // Can be null. Identified for informational purposes only. Not supported. Future compatibility is not guaranteed.
	// require.True(t, acc.HasField("sqlserver_azure_db_resource_stats", "end_time")) // Time
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_instance_memory_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_instance_cpu_percent"))
	// require.True(t, acc.HasTag("sqlserver_azure_db_resource_stats", "replica_updateability"))

	// This query should only return one row
	require.Equal(t, 1, len(acc.Metrics))
	server.Stop()
}

func TestAzureSQL_Managed_ResourceGovernance_Query(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("AZURESQL_MI_CONNECTION_STRING") == "" {
		t.Skip("Missing environment variable AZURESQL_MI_CONNECTION_STRING")
	}

	connectionString := os.Getenv("AZURESQL_MI_CONNECTION_STRING")
	
	server := &SQLServer {
		Servers:      []string{connectionString},
		IncludeQuery: []string{"AzureSQLMIResourceGovernance"},
		AuthMethod:   "connection_string",
		DatabaseType: "AzureSQLManagedInstance",
	}

	var acc testutil.Accumulator

	require.NoError(t, server.Start(&acc))
	require.NoError(t, server.Gather(&acc))

	// require.True(t, acc.HasMeasurement("sqlserver_azure_db_resource_stats"))
	// require.True(t, acc.HasTag("sqlserver_azure_db_resource_stats", "sql_instance"))
	// require.True(t, acc.HasTag("sqlserver_azure_db_resource_stats", "database_name"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_cpu_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_data_io_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_log_write_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_memory_usage_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "xtp_storage_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "max_worker_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "max_session_percent"))
	// require.True(t, acc.HasField("sqlserver_azure_db_resource_stats", "dtu_limit"))	// Can be null
	// require.True(t, acc.HasField("sqlserver_azure_db_resource_stats", "avg_login_rate_percent")) // Can be null. Identified for informational purposes only. Not supported. Future compatibility is not guaranteed.
	// require.True(t, acc.HasField("sqlserver_azure_db_resource_stats", "end_time")) // Time
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_instance_memory_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_instance_cpu_percent"))
	// require.True(t, acc.HasTag("sqlserver_azure_db_resource_stats", "replica_updateability"))

	// This query should only return one row
	require.Equal(t, 1, len(acc.Metrics))
	server.Stop()
}

func TestAzureSQL_Managed_DatabaseIO_Query(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("AZURESQL_MI_CONNECTION_STRING") == "" {
		t.Skip("Missing environment variable AZURESQL_MI_CONNECTION_STRING")
	}

	connectionString := os.Getenv("AZURESQL_MI_CONNECTION_STRING")
	
	server := &SQLServer {
		Servers:      []string{connectionString},
		IncludeQuery: []string{"AzureSQLMIDatabaseIO"},
		AuthMethod:   "connection_string",
		DatabaseType: "AzureSQLManagedInstance",
	}

	var acc testutil.Accumulator

	require.NoError(t, server.Start(&acc))
	require.NoError(t, server.Gather(&acc))

	// require.True(t, acc.HasMeasurement("sqlserver_azure_db_resource_stats"))
	// require.True(t, acc.HasTag("sqlserver_azure_db_resource_stats", "sql_instance"))
	// require.True(t, acc.HasTag("sqlserver_azure_db_resource_stats", "database_name"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_cpu_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_data_io_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_log_write_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_memory_usage_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "xtp_storage_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "max_worker_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "max_session_percent"))
	// require.True(t, acc.HasField("sqlserver_azure_db_resource_stats", "dtu_limit"))	// Can be null
	// require.True(t, acc.HasField("sqlserver_azure_db_resource_stats", "avg_login_rate_percent")) // Can be null. Identified for informational purposes only. Not supported. Future compatibility is not guaranteed.
	// require.True(t, acc.HasField("sqlserver_azure_db_resource_stats", "end_time")) // Time
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_instance_memory_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_instance_cpu_percent"))
	// require.True(t, acc.HasTag("sqlserver_azure_db_resource_stats", "replica_updateability"))

	// This query should only return one row
	require.Equal(t, 1, len(acc.Metrics))
	server.Stop()
}

func TestAzureSQL_Managed_ServerProperties_Query(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("AZURESQL_MI_CONNECTION_STRING") == "" {
		t.Skip("Missing environment variable AZURESQL_MI_CONNECTION_STRING")
	}

	connectionString := os.Getenv("AZURESQL_MI_CONNECTION_STRING")
	
	server := &SQLServer {
		Servers:      []string{connectionString},
		IncludeQuery: []string{"AzureSQLMIServerProperties"},
		AuthMethod:   "connection_string",
		DatabaseType: "AzureSQLManagedInstance",
	}

	var acc testutil.Accumulator

	require.NoError(t, server.Start(&acc))
	require.NoError(t, server.Gather(&acc))

	// require.True(t, acc.HasMeasurement("sqlserver_azure_db_resource_stats"))
	// require.True(t, acc.HasTag("sqlserver_azure_db_resource_stats", "sql_instance"))
	// require.True(t, acc.HasTag("sqlserver_azure_db_resource_stats", "database_name"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_cpu_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_data_io_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_log_write_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_memory_usage_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "xtp_storage_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "max_worker_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "max_session_percent"))
	// require.True(t, acc.HasField("sqlserver_azure_db_resource_stats", "dtu_limit"))	// Can be null
	// require.True(t, acc.HasField("sqlserver_azure_db_resource_stats", "avg_login_rate_percent")) // Can be null. Identified for informational purposes only. Not supported. Future compatibility is not guaranteed.
	// require.True(t, acc.HasField("sqlserver_azure_db_resource_stats", "end_time")) // Time
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_instance_memory_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_instance_cpu_percent"))
	// require.True(t, acc.HasTag("sqlserver_azure_db_resource_stats", "replica_updateability"))

	// This query should only return one row
	require.Equal(t, 1, len(acc.Metrics))
	server.Stop()
}

func TestAzureSQL_Managed_OsWaitStats_Query(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("AZURESQL_MI_CONNECTION_STRING") == "" {
		t.Skip("Missing environment variable AZURESQL_MI_CONNECTION_STRING")
	}

	connectionString := os.Getenv("AZURESQL_MI_CONNECTION_STRING")
	
	server := &SQLServer {
		Servers:      []string{connectionString},
		IncludeQuery: []string{"AzureSQLMIOsWaitstats"},
		AuthMethod:   "connection_string",
		DatabaseType: "AzureSQLManagedInstance",
	}

	var acc testutil.Accumulator

	require.NoError(t, server.Start(&acc))
	require.NoError(t, server.Gather(&acc))

	// require.True(t, acc.HasMeasurement("sqlserver_azure_db_resource_stats"))
	// require.True(t, acc.HasTag("sqlserver_azure_db_resource_stats", "sql_instance"))
	// require.True(t, acc.HasTag("sqlserver_azure_db_resource_stats", "database_name"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_cpu_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_data_io_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_log_write_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_memory_usage_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "xtp_storage_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "max_worker_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "max_session_percent"))
	// require.True(t, acc.HasField("sqlserver_azure_db_resource_stats", "dtu_limit"))	// Can be null
	// require.True(t, acc.HasField("sqlserver_azure_db_resource_stats", "avg_login_rate_percent")) // Can be null. Identified for informational purposes only. Not supported. Future compatibility is not guaranteed.
	// require.True(t, acc.HasField("sqlserver_azure_db_resource_stats", "end_time")) // Time
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_instance_memory_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_instance_cpu_percent"))
	// require.True(t, acc.HasTag("sqlserver_azure_db_resource_stats", "replica_updateability"))

	// This query should only return one row
	require.Equal(t, 1, len(acc.Metrics))
	server.Stop()
}

func TestAzureSQL_Managed_MemoryClerks_Query(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("AZURESQL_MI_CONNECTION_STRING") == "" {
		t.Skip("Missing environment variable AZURESQL_MI_CONNECTION_STRING")
	}

	connectionString := os.Getenv("AZURESQL_MI_CONNECTION_STRING")
	
	server := &SQLServer {
		Servers:      []string{connectionString},
		IncludeQuery: []string{"AzureSQLMIMemoryClerks"},
		AuthMethod:   "connection_string",
		DatabaseType: "AzureSQLManagedInstance",
	}

	var acc testutil.Accumulator

	require.NoError(t, server.Start(&acc))
	require.NoError(t, server.Gather(&acc))

	// require.True(t, acc.HasMeasurement("sqlserver_azure_db_resource_stats"))
	// require.True(t, acc.HasTag("sqlserver_azure_db_resource_stats", "sql_instance"))
	// require.True(t, acc.HasTag("sqlserver_azure_db_resource_stats", "database_name"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_cpu_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_data_io_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_log_write_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_memory_usage_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "xtp_storage_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "max_worker_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "max_session_percent"))
	// require.True(t, acc.HasField("sqlserver_azure_db_resource_stats", "dtu_limit"))	// Can be null
	// require.True(t, acc.HasField("sqlserver_azure_db_resource_stats", "avg_login_rate_percent")) // Can be null. Identified for informational purposes only. Not supported. Future compatibility is not guaranteed.
	// require.True(t, acc.HasField("sqlserver_azure_db_resource_stats", "end_time")) // Time
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_instance_memory_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_instance_cpu_percent"))
	// require.True(t, acc.HasTag("sqlserver_azure_db_resource_stats", "replica_updateability"))

	// This query should only return one row
	require.Equal(t, 1, len(acc.Metrics))
	server.Stop()
}

func TestAzureSQL_Managed_PerformanceCounters_Query(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("AZURESQL_MI_CONNECTION_STRING") == "" {
		t.Skip("Missing environment variable AZURESQL_MI_CONNECTION_STRING")
	}

	connectionString := os.Getenv("AZURESQL_MI_CONNECTION_STRING")
	
	server := &SQLServer {
		Servers:      []string{connectionString},
		IncludeQuery: []string{"AzureSQLMIPerformanceCounters"},
		AuthMethod:   "connection_string",
		DatabaseType: "AzureSQLManagedInstance",
	}

	var acc testutil.Accumulator

	require.NoError(t, server.Start(&acc))
	require.NoError(t, server.Gather(&acc))

	// require.True(t, acc.HasMeasurement("sqlserver_azure_db_resource_stats"))
	// require.True(t, acc.HasTag("sqlserver_azure_db_resource_stats", "sql_instance"))
	// require.True(t, acc.HasTag("sqlserver_azure_db_resource_stats", "database_name"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_cpu_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_data_io_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_log_write_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_memory_usage_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "xtp_storage_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "max_worker_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "max_session_percent"))
	// require.True(t, acc.HasField("sqlserver_azure_db_resource_stats", "dtu_limit"))	// Can be null
	// require.True(t, acc.HasField("sqlserver_azure_db_resource_stats", "avg_login_rate_percent")) // Can be null. Identified for informational purposes only. Not supported. Future compatibility is not guaranteed.
	// require.True(t, acc.HasField("sqlserver_azure_db_resource_stats", "end_time")) // Time
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_instance_memory_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_instance_cpu_percent"))
	// require.True(t, acc.HasTag("sqlserver_azure_db_resource_stats", "replica_updateability"))

	// This query should only return one row
	require.Equal(t, 1, len(acc.Metrics))
	server.Stop()
}

func TestAzureSQL_Managed_Requests_Query(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("AZURESQL_MI_CONNECTION_STRING") == "" {
		t.Skip("Missing environment variable AZURESQL_MI_CONNECTION_STRING")
	}

	connectionString := os.Getenv("AZURESQL_MI_CONNECTION_STRING")
	
	server := &SQLServer {
		Servers:      []string{connectionString},
		IncludeQuery: []string{"AzureSQLMIRequests"},
		AuthMethod:   "connection_string",
		DatabaseType: "AzureSQLManagedInstance",
	}

	var acc testutil.Accumulator

	require.NoError(t, server.Start(&acc))
	require.NoError(t, server.Gather(&acc))

	// require.True(t, acc.HasMeasurement("sqlserver_azure_db_resource_stats"))
	// require.True(t, acc.HasTag("sqlserver_azure_db_resource_stats", "sql_instance"))
	// require.True(t, acc.HasTag("sqlserver_azure_db_resource_stats", "database_name"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_cpu_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_data_io_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_log_write_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_memory_usage_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "xtp_storage_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "max_worker_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "max_session_percent"))
	// require.True(t, acc.HasField("sqlserver_azure_db_resource_stats", "dtu_limit"))	// Can be null
	// require.True(t, acc.HasField("sqlserver_azure_db_resource_stats", "avg_login_rate_percent")) // Can be null. Identified for informational purposes only. Not supported. Future compatibility is not guaranteed.
	// require.True(t, acc.HasField("sqlserver_azure_db_resource_stats", "end_time")) // Time
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_instance_memory_percent"))
	// require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_instance_cpu_percent"))
	// require.True(t, acc.HasTag("sqlserver_azure_db_resource_stats", "replica_updateability"))

	// This query should only return one row
	require.Equal(t, 1, len(acc.Metrics))
	server.Stop()
}

func TestAzureSQL_Managed_Schedulers_Query(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("AZURESQL_MI_CONNECTION_STRING") == "" {
		t.Skip("Missing environment variable AZURESQL_MI_CONNECTION_STRING")
	}

	connectionString := os.Getenv("AZURESQL_MI_CONNECTION_STRING")
	
	server := &SQLServer {
		Servers:      []string{connectionString},
		IncludeQuery: []string{"AzureSQLMISchedulers"},
		AuthMethod:   "connection_string",
		DatabaseType: "AzureSQLManagedInstance",
	}

	var acc testutil.Accumulator

	require.NoError(t, server.Start(&acc))
	require.NoError(t, server.Gather(&acc))

	require.True(t, acc.HasMeasurement("sqlserver_schedulers"))
	require.True(t, acc.HasTag("sqlserver_schedulers", "sql_instance"))
	require.True(t, acc.HasTag("sqlserver_schedulers", "elastic_pool_name"))
	require.True(t, acc.HasInt64Field("sqlserver_schedulers", "scheduler_id"))
	require.True(t, acc.HasInt64Field("sqlserver_schedulers", "cpu_id"))
	require.True(t, acc.HasTag("sqlserver_schedulers", "status"))
	require.True(t, acc.HasField("sqlserver_schedulers", "is_online"))	
	require.True(t, acc.HasField("sqlserver_schedulers", "is_idle"))	
	require.True(t, acc.HasInt64Field("sqlserver_schedulers", "preemptive_switches_count"))
	require.True(t, acc.HasInt64Field("sqlserver_schedulers", "context_switches_count"))
	require.True(t, acc.HasInt64Field("sqlserver_schedulers", "idle_switches_count"))
	require.True(t, acc.HasInt64Field("sqlserver_schedulers", "current_tasks_count"))
	require.True(t, acc.HasInt64Field("sqlserver_schedulers", "runnable_tasks_count"))
	require.True(t, acc.HasInt64Field("sqlserver_schedulers", "current_workers_count"))
	require.True(t, acc.HasInt64Field("sqlserver_schedulers", "active_workers_count"))
	require.True(t, acc.HasInt64Field("sqlserver_schedulers", "work_queue_count"))
	require.True(t, acc.HasInt64Field("sqlserver_schedulers", "pending_disk_io_count"))
	require.True(t, acc.HasInt64Field("sqlserver_schedulers", "load_factor"))
	require.True(t, acc.HasField("sqlserver_schedulers", "failed_to_create_worker"))
	require.True(t, acc.HasInt64Field("sqlserver_schedulers", "quantum_length_us"))
	require.True(t, acc.HasInt64Field("sqlserver_schedulers", "yield_count"))
	require.True(t, acc.HasInt64Field("sqlserver_schedulers", "total_cpu_usage_ms"))
	require.True(t, acc.HasInt64Field("sqlserver_schedulers", "total_cpu_idle_capped_ms"))
	require.True(t, acc.HasInt64Field("sqlserver_schedulers", "total_scheduler_delay_ms"))
	require.True(t, acc.HasInt64Field("sqlserver_schedulers", "ideal_workers_limit"))

	// This query should only return one row
	require.Equal(t, 1, len(acc.Metrics))
	server.Stop()
}