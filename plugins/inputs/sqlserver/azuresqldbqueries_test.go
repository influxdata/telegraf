package sqlserver

import (
	"os"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestAzureSQL_Database_ResourceStats_Query(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("AZURESQL_DB_CONNECTION_STRING") == "" {
		t.Skip("Missing environment variable AZURESQL_DB_CONNECTION_STRING")
	}

	connectionString := os.Getenv("AZURESQL_DB_CONNECTION_STRING")

	server := &SQLServer{
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
	require.True(t, acc.HasField("sqlserver_azure_db_resource_stats", "dtu_limit"))              // Can be null.
	require.True(t, acc.HasField("sqlserver_azure_db_resource_stats", "avg_login_rate_percent")) // Can be null.
	require.True(t, acc.HasField("sqlserver_azure_db_resource_stats", "end_time"))               // Time field.
	require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_instance_memory_percent"))
	require.True(t, acc.HasFloatField("sqlserver_azure_db_resource_stats", "avg_instance_cpu_percent"))
	require.True(t, acc.HasTag("sqlserver_azure_db_resource_stats", "replica_updateability"))

	// This query should only return one row
	require.Equal(t, 1, len(acc.Metrics))
	server.Stop()
}

func TestAzureSQL_Database_ResourceGovernance_Query(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("AZURESQL_DB_CONNECTION_STRING") == "" {
		t.Skip("Missing environment variable AZURESQL_DB_CONNECTION_STRING")
	}

	connectionString := os.Getenv("AZURESQL_DB_CONNECTION_STRING")

	server := &SQLServer{
		Servers:      []string{connectionString},
		IncludeQuery: []string{"AzureSQLDBResourceGovernance"},
		AuthMethod:   "connection_string",
		DatabaseType: "AzureSQLDB",
	}

	var acc testutil.Accumulator

	require.NoError(t, server.Start(&acc))
	require.NoError(t, server.Gather(&acc))

	require.True(t, acc.HasMeasurement("sqlserver_db_resource_governance"))
	require.True(t, acc.HasTag("sqlserver_db_resource_governance", "sql_instance"))
	require.True(t, acc.HasTag("sqlserver_db_resource_governance", "database_name"))
	require.True(t, acc.HasTag("sqlserver_db_resource_governance", "slo_name"))
	require.True(t, acc.HasInt64Field("sqlserver_db_resource_governance", "dtu_limit"))
	require.True(t, acc.HasInt64Field("sqlserver_db_resource_governance", "max_cpu"))
	require.True(t, acc.HasInt64Field("sqlserver_db_resource_governance", "cap_cpu"))
	require.True(t, acc.HasInt64Field("sqlserver_db_resource_governance", "instance_cap_cpu"))
	require.True(t, acc.HasInt64Field("sqlserver_db_resource_governance", "max_db_memory"))
	require.True(t, acc.HasInt64Field("sqlserver_db_resource_governance", "max_db_max_size_in_mb"))
	require.True(t, acc.HasInt64Field("sqlserver_db_resource_governance", "db_file_growth_in_mb"))
	require.True(t, acc.HasInt64Field("sqlserver_db_resource_governance", "log_size_in_mb"))
	require.True(t, acc.HasInt64Field("sqlserver_db_resource_governance", "instance_max_worker_threads"))
	require.True(t, acc.HasInt64Field("sqlserver_db_resource_governance", "primary_group_max_workers"))
	require.True(t, acc.HasInt64Field("sqlserver_db_resource_governance", "instance_max_log_rate"))
	require.True(t, acc.HasInt64Field("sqlserver_db_resource_governance", "primary_min_log_rate"))
	require.True(t, acc.HasInt64Field("sqlserver_db_resource_governance", "primary_max_log_rate"))
	require.True(t, acc.HasInt64Field("sqlserver_db_resource_governance", "primary_group_min_io"))
	require.True(t, acc.HasInt64Field("sqlserver_db_resource_governance", "primary_group_max_io"))
	require.True(t, acc.HasFloatField("sqlserver_db_resource_governance", "primary_group_min_cpu"))
	require.True(t, acc.HasFloatField("sqlserver_db_resource_governance", "primary_group_max_cpu"))
	require.True(t, acc.HasInt64Field("sqlserver_db_resource_governance", "primary_pool_max_workers"))
	require.True(t, acc.HasInt64Field("sqlserver_db_resource_governance", "pool_max_io"))
	require.True(t, acc.HasInt64Field("sqlserver_db_resource_governance", "checkpoint_rate_mbps"))
	require.True(t, acc.HasInt64Field("sqlserver_db_resource_governance", "checkpoint_rate_io"))
	require.True(t, acc.HasInt64Field("sqlserver_db_resource_governance", "volume_local_iops"))
	require.True(t, acc.HasInt64Field("sqlserver_db_resource_governance", "volume_managed_xstore_iops"))
	require.True(t, acc.HasInt64Field("sqlserver_db_resource_governance", "volume_external_xstore_iops"))
	require.True(t, acc.HasInt64Field("sqlserver_db_resource_governance", "volume_type_local_iops"))
	require.True(t, acc.HasInt64Field("sqlserver_db_resource_governance", "volume_type_managed_xstore_iops"))
	require.True(t, acc.HasInt64Field("sqlserver_db_resource_governance", "volume_type_external_xstore_iops"))
	require.True(t, acc.HasInt64Field("sqlserver_db_resource_governance", "volume_pfs_iops"))
	require.True(t, acc.HasInt64Field("sqlserver_db_resource_governance", "volume_type_pfs_iops"))
	require.True(t, acc.HasTag("sqlserver_db_resource_governance", "replica_updateability"))

	server.Stop()
}

func TestAzureSQL_Database_WaitStats_Query(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("AZURESQL_DB_CONNECTION_STRING") == "" {
		t.Skip("Missing environment variable AZURESQL_DB_CONNECTION_STRING")
	}

	connectionString := os.Getenv("AZURESQL_DB_CONNECTION_STRING")

	server := &SQLServer{
		Servers:      []string{connectionString},
		IncludeQuery: []string{"AzureSQLDBWaitStats"},
		AuthMethod:   "connection_string",
		DatabaseType: "AzureSQLDB",
	}

	var acc testutil.Accumulator

	require.NoError(t, server.Start(&acc))
	require.NoError(t, server.Gather(&acc))

	require.True(t, acc.HasMeasurement("sqlserver_azuredb_waitstats"))
	require.True(t, acc.HasTag("sqlserver_azuredb_waitstats", "sql_instance"))
	require.True(t, acc.HasTag("sqlserver_azuredb_waitstats", "database_name"))
	require.True(t, acc.HasTag("sqlserver_azuredb_waitstats", "wait_type"))
	require.True(t, acc.HasInt64Field("sqlserver_azuredb_waitstats", "wait_time_ms"))
	require.True(t, acc.HasInt64Field("sqlserver_azuredb_waitstats", "resource_wait_ms"))
	require.True(t, acc.HasInt64Field("sqlserver_azuredb_waitstats", "signal_wait_time_ms"))
	require.True(t, acc.HasInt64Field("sqlserver_azuredb_waitstats", "max_wait_time_ms"))
	require.True(t, acc.HasInt64Field("sqlserver_azuredb_waitstats", "waiting_tasks_count"))
	require.True(t, acc.HasTag("sqlserver_azuredb_waitstats", "replica_updateability"))

	server.Stop()
}

func TestAzureSQL_Database_DatabaseIO_Query(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("AZURESQL_DB_CONNECTION_STRING") == "" {
		t.Skip("Missing environment variable AZURESQL_DB_CONNECTION_STRING")
	}

	connectionString := os.Getenv("AZURESQL_DB_CONNECTION_STRING")

	server := &SQLServer{
		Servers:      []string{connectionString},
		IncludeQuery: []string{"AzureSQLDBDatabaseIO"},
		AuthMethod:   "connection_string",
		DatabaseType: "AzureSQLDB",
	}

	var acc testutil.Accumulator

	require.NoError(t, server.Start(&acc))
	require.NoError(t, server.Gather(&acc))

	require.True(t, acc.HasMeasurement("sqlserver_database_io"))
	require.True(t, acc.HasTag("sqlserver_database_io", "sql_instance"))
	require.True(t, acc.HasTag("sqlserver_database_io", "database_name"))
	require.True(t, acc.HasInt64Field("sqlserver_database_io", "database_id"))
	require.True(t, acc.HasInt64Field("sqlserver_database_io", "file_id"))
	require.True(t, acc.HasInt64Field("sqlserver_database_io", "read_latency_ms"))
	require.True(t, acc.HasInt64Field("sqlserver_database_io", "reads"))
	require.True(t, acc.HasInt64Field("sqlserver_database_io", "read_bytes"))
	require.True(t, acc.HasInt64Field("sqlserver_database_io", "write_latency_ms"))
	require.True(t, acc.HasInt64Field("sqlserver_database_io", "writes"))
	require.True(t, acc.HasInt64Field("sqlserver_database_io", "write_bytes"))
	require.True(t, acc.HasInt64Field("sqlserver_database_io", "rg_read_stall_ms"))
	require.True(t, acc.HasInt64Field("sqlserver_database_io", "rg_write_stall_ms"))
	require.True(t, acc.HasTag("sqlserver_database_io", "logical_filename"))
	require.True(t, acc.HasTag("sqlserver_database_io", "physical_filename"))
	require.True(t, acc.HasTag("sqlserver_database_io", "file_type"))
	require.True(t, acc.HasInt64Field("sqlserver_database_io", "current_size_mb"))
	require.True(t, acc.HasInt64Field("sqlserver_database_io", "space_used_mb"))
	require.True(t, acc.HasTag("sqlserver_database_io", "replica_updateability"))

	server.Stop()
}

func TestAzureSQL_Database_ServerProperties_Query(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("AZURESQL_DB_CONNECTION_STRING") == "" {
		t.Skip("Missing environment variable AZURESQL_DB_CONNECTION_STRING")
	}

	connectionString := os.Getenv("AZURESQL_DB_CONNECTION_STRING")

	server := &SQLServer{
		Servers:      []string{connectionString},
		IncludeQuery: []string{"AzureSQLDBServerProperties"},
		AuthMethod:   "connection_string",
		DatabaseType: "AzureSQLDB",
	}

	var acc testutil.Accumulator

	require.NoError(t, server.Start(&acc))
	require.NoError(t, server.Gather(&acc))

	require.True(t, acc.HasMeasurement("sqlserver_server_properties"))
	require.True(t, acc.HasTag("sqlserver_server_properties", "sql_instance"))
	require.True(t, acc.HasTag("sqlserver_server_properties", "database_name"))
	require.True(t, acc.HasInt64Field("sqlserver_server_properties", "cpu_count"))
	require.True(t, acc.HasInt64Field("sqlserver_server_properties", "server_memory"))
	require.True(t, acc.HasTag("sqlserver_server_properties", "sku"))
	require.True(t, acc.HasInt64Field("sqlserver_server_properties", "engine_edition"))
	require.True(t, acc.HasTag("sqlserver_server_properties", "hardware_type"))
	require.True(t, acc.HasInt64Field("sqlserver_server_properties", "total_storage_mb"))
	require.True(t, acc.HasInt64Field("sqlserver_server_properties", "available_storage_mb"))
	require.True(t, acc.HasField("sqlserver_server_properties", "uptime")) // Time field.
	require.True(t, acc.HasTag("sqlserver_server_properties", "replica_updateability"))

	// This query should only return one row
	require.Equal(t, 1, len(acc.Metrics))
	server.Stop()
}

func TestAzureSQL_Database_OsWaitstats_Query(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("AZURESQL_DB_CONNECTION_STRING") == "" {
		t.Skip("Missing environment variable AZURESQL_DB_CONNECTION_STRING")
	}

	connectionString := os.Getenv("AZURESQL_DB_CONNECTION_STRING")

	server := &SQLServer{
		Servers:      []string{connectionString},
		IncludeQuery: []string{"AzureSQLDBOsWaitstats"},
		AuthMethod:   "connection_string",
		DatabaseType: "AzureSQLDB",
	}

	var acc testutil.Accumulator

	require.NoError(t, server.Start(&acc))
	require.NoError(t, server.Gather(&acc))

	require.True(t, acc.HasMeasurement("sqlserver_waitstats"))
	require.True(t, acc.HasTag("sqlserver_waitstats", "sql_instance"))
	require.True(t, acc.HasTag("sqlserver_waitstats", "database_name"))
	require.True(t, acc.HasTag("sqlserver_waitstats", "wait_type"))
	require.True(t, acc.HasInt64Field("sqlserver_waitstats", "wait_time_ms"))
	require.True(t, acc.HasInt64Field("sqlserver_waitstats", "resource_wait_ms"))
	require.True(t, acc.HasInt64Field("sqlserver_waitstats", "signal_wait_time_ms"))
	require.True(t, acc.HasInt64Field("sqlserver_waitstats", "max_wait_time_ms"))
	require.True(t, acc.HasInt64Field("sqlserver_waitstats", "waiting_tasks_count"))
	require.True(t, acc.HasTag("sqlserver_waitstats", "wait_category"))
	require.True(t, acc.HasTag("sqlserver_waitstats", "replica_updateability"))

	server.Stop()
}

func TestAzureSQL_Database_MemoryClerks_Query(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("AZURESQL_DB_CONNECTION_STRING") == "" {
		t.Skip("Missing environment variable AZURESQL_DB_CONNECTION_STRING")
	}

	connectionString := os.Getenv("AZURESQL_DB_CONNECTION_STRING")

	server := &SQLServer{
		Servers:      []string{connectionString},
		IncludeQuery: []string{"AzureSQLDBMemoryClerks"},
		AuthMethod:   "connection_string",
		DatabaseType: "AzureSQLDB",
	}

	var acc testutil.Accumulator

	require.NoError(t, server.Start(&acc))
	require.NoError(t, server.Gather(&acc))

	require.True(t, acc.HasMeasurement("sqlserver_memory_clerks"))
	require.True(t, acc.HasTag("sqlserver_memory_clerks", "sql_instance"))
	require.True(t, acc.HasTag("sqlserver_memory_clerks", "database_name"))
	require.True(t, acc.HasTag("sqlserver_memory_clerks", "clerk_type"))
	require.True(t, acc.HasInt64Field("sqlserver_memory_clerks", "size_kb"))
	require.True(t, acc.HasTag("sqlserver_memory_clerks", "replica_updateability"))

	server.Stop()
}

func TestAzureSQL_Database_PerformanceCounters_Query(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("AZURESQL_DB_CONNECTION_STRING") == "" {
		t.Skip("Missing environment variable AZURESQL_DB_CONNECTION_STRING")
	}

	connectionString := os.Getenv("AZURESQL_DB_CONNECTION_STRING")

	server := &SQLServer{
		Servers:      []string{connectionString},
		IncludeQuery: []string{"AzureSQLDBPerformanceCounters"},
		AuthMethod:   "connection_string",
		DatabaseType: "AzureSQLDB",
	}

	var acc testutil.Accumulator

	require.NoError(t, server.Start(&acc))
	require.NoError(t, server.Gather(&acc))

	require.True(t, acc.HasMeasurement("sqlserver_performance"))
	require.True(t, acc.HasTag("sqlserver_performance", "sql_instance"))
	require.True(t, acc.HasTag("sqlserver_performance", "database_name"))
	require.True(t, acc.HasTag("sqlserver_performance", "object"))
	require.True(t, acc.HasTag("sqlserver_performance", "counter"))
	require.True(t, acc.HasTag("sqlserver_performance", "instance"))
	require.True(t, acc.HasFloatField("sqlserver_performance", "value"))
	require.True(t, acc.HasTag("sqlserver_performance", "counter_type"))
	require.True(t, acc.HasTag("sqlserver_performance", "replica_updateability"))

	server.Stop()
}

func TestAzureSQL_Database_Requests_Query(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("AZURESQL_DB_CONNECTION_STRING") == "" {
		t.Skip("Missing environment variable AZURESQL_DB_CONNECTION_STRING")
	}

	connectionString := os.Getenv("AZURESQL_DB_CONNECTION_STRING")

	server := &SQLServer{
		Servers:      []string{connectionString},
		IncludeQuery: []string{"AzureSQLDBRequests"},
		AuthMethod:   "connection_string",
		DatabaseType: "AzureSQLDB",
	}

	var acc testutil.Accumulator

	require.NoError(t, server.Start(&acc))
	require.NoError(t, server.Gather(&acc))

	require.True(t, acc.HasMeasurement("sqlserver_requests"))
	require.True(t, acc.HasTag("sqlserver_requests", "sql_instance"))
	require.True(t, acc.HasTag("sqlserver_requests", "database_name"))
	require.True(t, acc.HasInt64Field("sqlserver_requests", "session_id"))
	require.True(t, acc.HasInt64Field("sqlserver_requests", "request_id"))
	require.True(t, acc.HasTag("sqlserver_requests", "session_db_name"))
	require.True(t, acc.HasTag("sqlserver_requests", "status"))
	require.True(t, acc.HasInt64Field("sqlserver_requests", "cpu_time_ms"))
	require.True(t, acc.HasInt64Field("sqlserver_requests", "total_elapsed_time_ms"))
	require.True(t, acc.HasInt64Field("sqlserver_requests", "logical_reads"))
	require.True(t, acc.HasInt64Field("sqlserver_requests", "writes"))
	require.True(t, acc.HasTag("sqlserver_requests", "command"))
	require.True(t, acc.HasInt64Field("sqlserver_requests", "wait_time_ms"))
	require.True(t, acc.HasField("sqlserver_requests", "wait_type")) // Can be null.
	require.True(t, acc.HasTag("sqlserver_requests", "wait_resource"))
	require.True(t, acc.HasInt64Field("sqlserver_requests", "blocking_session_id"))
	require.True(t, acc.HasTag("sqlserver_requests", "program_name"))
	require.True(t, acc.HasTag("sqlserver_requests", "host_name"))
	require.True(t, acc.HasField("sqlserver_requests", "nt_user_name")) // Can be null.
	require.True(t, acc.HasTag("sqlserver_requests", "login_name"))
	require.True(t, acc.HasInt64Field("sqlserver_requests", "open_transaction"))
	require.True(t, acc.HasTag("sqlserver_requests", "transaction_isolation_level"))
	require.True(t, acc.HasInt64Field("sqlserver_requests", "granted_query_memory_pages"))
	require.True(t, acc.HasFloatField("sqlserver_requests", "percent_complete"))
	require.True(t, acc.HasTag("sqlserver_requests", "statement_text"))
	require.True(t, acc.HasField("sqlserver_requests", "objectid"))         // Can be null.
	require.True(t, acc.HasField("sqlserver_requests", "stmt_object_name")) // Can be null.
	require.True(t, acc.HasField("sqlserver_requests", "stmt_db_name"))     // Can be null.
	require.True(t, acc.HasTag("sqlserver_requests", "query_hash"))
	require.True(t, acc.HasTag("sqlserver_requests", "query_plan_hash"))
	require.True(t, acc.HasTag("sqlserver_requests", "replica_updateability"))

	server.Stop()
}

func TestAzureSQL_Database_Schedulers_Query(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("AZURESQL_DB_CONNECTION_STRING") == "" {
		t.Skip("Missing environment variable AZURESQL_DB_CONNECTION_STRING")
	}

	connectionString := os.Getenv("AZURESQL_DB_CONNECTION_STRING")

	server := &SQLServer{
		Servers:      []string{connectionString},
		IncludeQuery: []string{"AzureSQLDBSchedulers"},
		AuthMethod:   "connection_string",
		DatabaseType: "AzureSQLDB",
	}

	var acc testutil.Accumulator

	require.NoError(t, server.Start(&acc))
	require.NoError(t, server.Gather(&acc))

	require.True(t, acc.HasMeasurement("sqlserver_schedulers"))
	require.True(t, acc.HasTag("sqlserver_schedulers", "sql_instance"))
	require.True(t, acc.HasTag("sqlserver_schedulers", "scheduler_id"))
	require.True(t, acc.HasTag("sqlserver_schedulers", "cpu_id"))
	require.True(t, acc.HasField("sqlserver_schedulers", "is_online")) // Bool field.
	require.True(t, acc.HasField("sqlserver_schedulers", "is_idle"))   // Bool field.
	require.True(t, acc.HasInt64Field("sqlserver_schedulers", "preemptive_switches_count"))
	require.True(t, acc.HasInt64Field("sqlserver_schedulers", "context_switches_count"))
	require.True(t, acc.HasInt64Field("sqlserver_schedulers", "current_tasks_count"))
	require.True(t, acc.HasInt64Field("sqlserver_schedulers", "runnable_tasks_count"))
	require.True(t, acc.HasInt64Field("sqlserver_schedulers", "current_workers_count"))
	require.True(t, acc.HasInt64Field("sqlserver_schedulers", "active_workers_count"))
	require.True(t, acc.HasInt64Field("sqlserver_schedulers", "work_queue_count"))
	require.True(t, acc.HasInt64Field("sqlserver_schedulers", "pending_disk_io_count"))
	require.True(t, acc.HasInt64Field("sqlserver_schedulers", "load_factor"))
	require.True(t, acc.HasInt64Field("sqlserver_schedulers", "yield_count"))
	require.True(t, acc.HasInt64Field("sqlserver_schedulers", "total_cpu_usage_ms"))
	require.True(t, acc.HasInt64Field("sqlserver_schedulers", "total_scheduler_delay_ms"))
	require.True(t, acc.HasTag("sqlserver_schedulers", "replica_updateability"))

	server.Stop()
}
