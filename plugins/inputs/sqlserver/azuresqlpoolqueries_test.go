package sqlserver

import (
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestAzureSQL_ElasticPool_ResourceStats_Query(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("AZURESQL_POOL_CONNECTION_STRING") == "" {
		t.Skip("Missing environment variable AZURESQL_POOL_CONNECTION_STRING")
	}

	connectionString := os.Getenv("AZURESQL_POOL_CONNECTION_STRING")

	server := &SQLServer{
		Servers:      []string{connectionString},
		IncludeQuery: []string{"AzureSQLPoolResourceStats"},
		AuthMethod:   "connection_string",
		DatabaseType: "AzureSQLPool",
	}

	var acc testutil.Accumulator

	require.NoError(t, server.Start(&acc))
	require.NoError(t, server.Gather(&acc))

	require.True(t, acc.HasMeasurement("sqlserver_pool_resource_stats"))
	require.True(t, acc.HasTag("sqlserver_pool_resource_stats", "sql_instance"))
	require.True(t, acc.HasTag("sqlserver_pool_resource_stats", "elastic_pool_name"))
	require.True(t, acc.HasField("sqlserver_pool_resource_stats", "snapshot_time"))
	require.True(t, acc.HasFloatField("sqlserver_pool_resource_stats", "avg_cpu_percent"))
	require.True(t, acc.HasFloatField("sqlserver_pool_resource_stats", "avg_data_io_percent"))
	require.True(t, acc.HasFloatField("sqlserver_pool_resource_stats", "avg_log_write_percent"))
	require.True(t, acc.HasFloatField("sqlserver_pool_resource_stats", "avg_storage_percent"))
	require.True(t, acc.HasFloatField("sqlserver_pool_resource_stats", "max_worker_percent"))
	require.True(t, acc.HasFloatField("sqlserver_pool_resource_stats", "max_session_percent"))
	require.True(t, acc.HasInt64Field("sqlserver_pool_resource_stats", "storage_limit_mb"))
	require.True(t, acc.HasFloatField("sqlserver_pool_resource_stats", "avg_instance_cpu_percent"))
	require.True(t, acc.HasFloatField("sqlserver_pool_resource_stats", "avg_allocated_storage_percent"))

	// This query should only return one row
	require.Equal(t, 1, len(acc.Metrics))
	server.Stop()
}

func TestAzureSQL_ElasticPool_ResourceGovernance_Query(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("AZURESQL_POOL_CONNECTION_STRING") == "" {
		t.Skip("Missing environment variable AZURESQL_POOL_CONNECTION_STRING")
	}

	connectionString := os.Getenv("AZURESQL_POOL_CONNECTION_STRING")

	server := &SQLServer{
		Servers:      []string{connectionString},
		IncludeQuery: []string{"AzureSQLPoolResourceGovernance"},
		AuthMethod:   "connection_string",
		DatabaseType: "AzureSQLPool",
	}

	var acc testutil.Accumulator

	require.NoError(t, server.Start(&acc))
	require.NoError(t, server.Gather(&acc))

	require.True(t, acc.HasMeasurement("sqlserver_pool_resource_governance"))
	require.True(t, acc.HasTag("sqlserver_pool_resource_governance", "sql_instance"))
	require.True(t, acc.HasTag("sqlserver_pool_resource_governance", "elastic_pool_name"))
	require.True(t, acc.HasTag("sqlserver_pool_resource_governance", "slo_name"))
	require.True(t, acc.HasInt64Field("sqlserver_pool_resource_governance", "dtu_limit"))
	require.True(t, acc.HasInt64Field("sqlserver_pool_resource_governance", "cpu_limit"))
	require.True(t, acc.HasInt64Field("sqlserver_pool_resource_governance", "max_cpu"))
	require.True(t, acc.HasInt64Field("sqlserver_pool_resource_governance", "cap_cpu"))
	require.True(t, acc.HasInt64Field("sqlserver_pool_resource_governance", "max_db_memory"))
	require.True(t, acc.HasInt64Field("sqlserver_pool_resource_governance", "max_db_max_size_in_mb"))
	require.True(t, acc.HasInt64Field("sqlserver_pool_resource_governance", "db_file_growth_in_mb"))
	require.True(t, acc.HasInt64Field("sqlserver_pool_resource_governance", "log_size_in_mb"))
	require.True(t, acc.HasInt64Field("sqlserver_pool_resource_governance", "instance_cap_cpu"))
	require.True(t, acc.HasInt64Field("sqlserver_pool_resource_governance", "instance_max_log_rate"))
	require.True(t, acc.HasInt64Field("sqlserver_pool_resource_governance", "instance_max_worker_threads"))
	require.True(t, acc.HasInt64Field("sqlserver_pool_resource_governance", "checkpoint_rate_mbps"))
	require.True(t, acc.HasInt64Field("sqlserver_pool_resource_governance", "checkpoint_rate_io"))
	require.True(t, acc.HasInt64Field("sqlserver_pool_resource_governance", "primary_group_max_workers"))
	require.True(t, acc.HasInt64Field("sqlserver_pool_resource_governance", "primary_min_log_rate"))
	require.True(t, acc.HasInt64Field("sqlserver_pool_resource_governance", "primary_max_log_rate"))
	require.True(t, acc.HasInt64Field("sqlserver_pool_resource_governance", "primary_group_min_io"))
	require.True(t, acc.HasInt64Field("sqlserver_pool_resource_governance", "primary_group_max_io"))
	require.True(t, acc.HasFloatField("sqlserver_pool_resource_governance", "primary_group_min_cpu"))
	require.True(t, acc.HasFloatField("sqlserver_pool_resource_governance", "primary_group_max_cpu"))
	require.True(t, acc.HasInt64Field("sqlserver_pool_resource_governance", "primary_pool_max_workers"))
	require.True(t, acc.HasInt64Field("sqlserver_pool_resource_governance", "pool_max_io"))
	require.True(t, acc.HasInt64Field("sqlserver_pool_resource_governance", "volume_local_iops"))
	require.True(t, acc.HasInt64Field("sqlserver_pool_resource_governance", "volume_managed_xstore_iops"))
	require.True(t, acc.HasInt64Field("sqlserver_pool_resource_governance", "volume_external_xstore_iops"))
	require.True(t, acc.HasInt64Field("sqlserver_pool_resource_governance", "volume_type_local_iops"))
	require.True(t, acc.HasInt64Field("sqlserver_pool_resource_governance", "volume_type_managed_xstore_iops"))
	require.True(t, acc.HasInt64Field("sqlserver_pool_resource_governance", "volume_type_external_xstore_iops"))
	require.True(t, acc.HasInt64Field("sqlserver_pool_resource_governance", "volume_pfs_iops"))
	require.True(t, acc.HasInt64Field("sqlserver_pool_resource_governance", "volume_type_pfs_iops"))

	// This query should only return one row
	require.Equal(t, 1, len(acc.Metrics))
	server.Stop()
}

func TestAzureSQL_ElasticPool_DatabaseIO_Query(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("AZURESQL_POOL_CONNECTION_STRING") == "" {
		t.Skip("Missing environment variable AZURESQL_POOL_CONNECTION_STRING")
	}

	connectionString := os.Getenv("AZURESQL_POOL_CONNECTION_STRING")

	server := &SQLServer{
		Servers:      []string{connectionString},
		IncludeQuery: []string{"AzureSQLPoolDatabaseIO"},
		AuthMethod:   "connection_string",
		DatabaseType: "AzureSQLPool",
	}

	var acc testutil.Accumulator

	require.NoError(t, server.Start(&acc))
	require.NoError(t, server.Gather(&acc))

	require.True(t, acc.HasMeasurement("sqlserver_database_io"))
	require.True(t, acc.HasTag("sqlserver_database_io", "sql_instance"))
	require.True(t, acc.HasTag("sqlserver_database_io", "elastic_pool_name"))
	require.True(t, acc.HasTag("sqlserver_database_io", "database_name"))
	require.True(t, acc.HasInt64Field("sqlserver_database_io", "database_id"))
	require.True(t, acc.HasInt64Field("sqlserver_database_io", "file_id"))
	require.True(t, acc.HasTag("sqlserver_database_io", "file_type"))
	require.True(t, acc.HasInt64Field("sqlserver_database_io", "reads"))
	require.True(t, acc.HasInt64Field("sqlserver_database_io", "read_bytes"))
	require.True(t, acc.HasInt64Field("sqlserver_database_io", "read_latency_ms"))
	require.True(t, acc.HasInt64Field("sqlserver_database_io", "write_latency_ms"))
	require.True(t, acc.HasInt64Field("sqlserver_database_io", "writes"))
	require.True(t, acc.HasInt64Field("sqlserver_database_io", "write_bytes"))
	require.True(t, acc.HasInt64Field("sqlserver_database_io", "rg_read_stall_ms"))
	require.True(t, acc.HasInt64Field("sqlserver_database_io", "rg_write_stall_ms"))
	require.True(t, acc.HasInt64Field("sqlserver_database_io", "size_on_disk_bytes"))
	require.True(t, acc.HasInt64Field("sqlserver_database_io", "size_on_disk_mb"))

	server.Stop()
}

func TestAzureSQL_ElasticPool_OsWaitStats_Query(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("AZURESQL_POOL_CONNECTION_STRING") == "" {
		t.Skip("Missing environment variable AZURESQL_POOL_CONNECTION_STRING")
	}

	connectionString := os.Getenv("AZURESQL_POOL_CONNECTION_STRING")

	server := &SQLServer{
		Servers:      []string{connectionString},
		IncludeQuery: []string{"AzureSQLPoolOsWaitStats"},
		AuthMethod:   "connection_string",
		DatabaseType: "AzureSQLPool",
	}

	var acc testutil.Accumulator

	require.NoError(t, server.Start(&acc))
	require.NoError(t, server.Gather(&acc))

	require.True(t, acc.HasMeasurement("sqlserver_waitstats"))
	require.True(t, acc.HasTag("sqlserver_waitstats", "sql_instance"))
	require.True(t, acc.HasTag("sqlserver_waitstats", "elastic_pool_name"))
	require.True(t, acc.HasTag("sqlserver_waitstats", "wait_type"))
	require.True(t, acc.HasInt64Field("sqlserver_waitstats", "waiting_tasks_count"))
	require.True(t, acc.HasInt64Field("sqlserver_waitstats", "wait_time_ms"))
	require.True(t, acc.HasInt64Field("sqlserver_waitstats", "max_wait_time_ms"))
	require.True(t, acc.HasInt64Field("sqlserver_waitstats", "signal_wait_time_ms"))
	require.True(t, acc.HasInt64Field("sqlserver_waitstats", "resource_wait_ms"))
	require.True(t, acc.HasTag("sqlserver_waitstats", "wait_category"))

	server.Stop()
}

func TestAzureSQL_ElasticPool_MemoryClerks_Query(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("AZURESQL_POOL_CONNECTION_STRING") == "" {
		t.Skip("Missing environment variable AZURESQL_POOL_CONNECTION_STRING")
	}

	connectionString := os.Getenv("AZURESQL_POOL_CONNECTION_STRING")

	server := &SQLServer{
		Servers:      []string{connectionString},
		IncludeQuery: []string{"AzureSQLPoolMemoryClerks"},
		AuthMethod:   "connection_string",
		DatabaseType: "AzureSQLPool",
	}

	var acc testutil.Accumulator

	require.NoError(t, server.Start(&acc))
	require.NoError(t, server.Gather(&acc))

	require.True(t, acc.HasMeasurement("sqlserver_memory_clerks"))
	require.True(t, acc.HasTag("sqlserver_memory_clerks", "sql_instance"))
	require.True(t, acc.HasTag("sqlserver_memory_clerks", "elastic_pool_name"))
	require.True(t, acc.HasTag("sqlserver_memory_clerks", "clerk_type"))
	require.True(t, acc.HasInt64Field("sqlserver_memory_clerks", "size_kb"))

	server.Stop()
}

func TestAzureSQL_ElasticPool_PerformanceCounters_Query(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("AZURESQL_POOL_CONNECTION_STRING") == "" {
		t.Skip("Missing environment variable AZURESQL_POOL_CONNECTION_STRING")
	}

	connectionString := os.Getenv("AZURESQL_POOL_CONNECTION_STRING")

	server := &SQLServer{
		Servers:      []string{connectionString},
		IncludeQuery: []string{"AzureSQLPoolPerformanceCounters"},
		AuthMethod:   "connection_string",
		DatabaseType: "AzureSQLPool",
	}

	var acc testutil.Accumulator

	require.NoError(t, server.Start(&acc))
	require.NoError(t, server.Gather(&acc))

	require.True(t, acc.HasMeasurement("sqlserver_performance"))
	require.True(t, acc.HasTag("sqlserver_performance", "sql_instance"))
	require.True(t, acc.HasTag("sqlserver_performance", "object"))
	require.True(t, acc.HasTag("sqlserver_performance", "counter"))
	require.True(t, acc.HasTag("sqlserver_performance", "instance"))
	require.True(t, acc.HasFloatField("sqlserver_performance", "value"))
	require.True(t, acc.HasTag("sqlserver_performance", "counter_type"))

	server.Stop()
}

func TestAzureSQL_ElasticPool_Schedulers_Query(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("AZURESQL_POOL_CONNECTION_STRING") == "" {
		t.Skip("Missing environment variable AZURESQL_POOL_CONNECTION_STRING")
	}

	connectionString := os.Getenv("AZURESQL_POOL_CONNECTION_STRING")

	server := &SQLServer{
		Servers:      []string{connectionString},
		IncludeQuery: []string{"AzureSQLPoolSchedulers"},
		AuthMethod:   "connection_string",
		DatabaseType: "AzureSQLPool",
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

	server.Stop()
}
