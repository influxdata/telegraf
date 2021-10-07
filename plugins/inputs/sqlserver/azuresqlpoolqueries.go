package sqlserver

import (
	_ "github.com/denisenkom/go-mssqldb" // go-mssqldb initialization
)

//------------------------------------------------------------------------------------------------
//------------------ Azure Sql Elastic Pool ------------------------------------------------------
//------------------------------------------------------------------------------------------------
const sqlAzurePoolResourceStats = `
IF SERVERPROPERTY('EngineEdition') <> 5 
   OR NOT EXISTS (SELECT 1 FROM sys.database_service_objectives WHERE database_id = DB_ID() AND elastic_pool_name IS NOT NULL) BEGIN
	DECLARE @ErrorMessage AS nvarchar(500) = 'Telegraf - Connection string Server:'+ @@SERVERNAME + ',Database:' + DB_NAME() +' is not an Azure SQL database in an elastic pool. Check the database_type parameter in the telegraf configuration.';
	RAISERROR (@ErrorMessage,11,1)
	RETURN
END

SELECT TOP(1)
   'sqlserver_pool_resource_stats' AS [measurement]
  ,REPLACE(@@SERVERNAME,'\',':') AS [sql_instance]
  ,(SELECT [elastic_pool_name] FROM sys.database_service_objectives WHERE database_id = DB_ID()) AS [elastic_pool_name]
  ,[snapshot_time]
  ,cast([cap_vcores_used_percent] as float) AS [avg_cpu_percent]
  ,cast([avg_data_io_percent] as float) AS [avg_data_io_percent]
  ,cast([avg_log_write_percent] as float) AS [avg_log_write_percent]
  ,cast([avg_storage_percent] as float) AS [avg_storage_percent]
  ,cast([max_worker_percent] as float) AS [max_worker_percent]
  ,cast([max_session_percent] as float) AS [max_session_percent]
  ,cast([max_data_space_kb]/1024. as int) AS [storage_limit_mb]
  ,cast([avg_instance_cpu_percent] as float) AS [avg_instance_cpu_percent]
  ,cast([avg_allocated_storage_percent] as float) AS [avg_allocated_storage_percent]
FROM 
  sys.dm_resource_governor_resource_pools_history_ex
WHERE 
  [name] = 'SloSharedPool1'
ORDER BY
  [snapshot_time] DESC;
`

const sqlAzurePoolResourceGovernance = `
IF SERVERPROPERTY('EngineEdition') <> 5 
   OR NOT EXISTS (SELECT 1 FROM sys.database_service_objectives WHERE database_id = DB_ID() AND elastic_pool_name IS NOT NULL) BEGIN
	DECLARE @ErrorMessage AS nvarchar(500) = 'Telegraf - Connection string Server:'+ @@SERVERNAME + ',Database:' + DB_NAME() +' is not an Azure SQL database in an elastic pool. Check the database_type parameter in the telegraf configuration.';
	RAISERROR (@ErrorMessage,11,1)
	RETURN
END

SELECT
	 'sqlserver_pool_resource_governance' AS [measurement]
	,REPLACE(@@SERVERNAME,'\',':') AS [sql_instance]
	,(SELECT [elastic_pool_name] FROM sys.database_service_objectives WHERE database_id = DB_ID()) AS [elastic_pool_name]
	,[slo_name]
	,[dtu_limit]
	,[cpu_limit]
	,[max_cpu]
	,[cap_cpu]
	,[max_db_memory]
	,[max_db_max_size_in_mb]
	,[db_file_growth_in_mb]
	,[log_size_in_mb]
	,[instance_cap_cpu]
	,[instance_max_log_rate]
	,[instance_max_worker_threads]
	,[checkpoint_rate_mbps]
	,[checkpoint_rate_io]
	,[primary_group_max_workers]
	,[primary_min_log_rate]
	,[primary_max_log_rate]
	,[primary_group_min_io]
	,[primary_group_max_io]
	,[primary_group_min_cpu]
	,[primary_group_max_cpu]
	,[primary_pool_max_workers]
	,[pool_max_io]
	,[volume_local_iops]
	,[volume_managed_xstore_iops]
	,[volume_external_xstore_iops]
	,[volume_type_local_iops]
	,[volume_type_managed_xstore_iops]
	,[volume_type_external_xstore_iops]
	,[volume_pfs_iops]
	,[volume_type_pfs_iops]
FROM 
	sys.dm_user_db_resource_governance
WHERE database_id = DB_ID();
`

const sqlAzurePoolDatabaseIO = `
IF SERVERPROPERTY('EngineEdition') <> 5 
   OR NOT EXISTS (SELECT 1 FROM sys.database_service_objectives WHERE database_id = DB_ID() AND elastic_pool_name IS NOT NULL) BEGIN
	DECLARE @ErrorMessage AS nvarchar(500) = 'Telegraf - Connection string Server:'+ @@SERVERNAME + ',Database:' + DB_NAME() +' is not an Azure SQL database in an elastic pool. Check the database_type parameter in the telegraf configuration.';
	RAISERROR (@ErrorMessage,11,1)
	RETURN
END

SELECT
	 'sqlserver_database_io' AS [measurement]
	,REPLACE(@@SERVERNAME,'\',':') AS [sql_instance]
	,(SELECT [elastic_pool_name] FROM sys.database_service_objectives WHERE database_id = DB_ID()) AS [elastic_pool_name]
	,CASE
		WHEN vfs.[database_id] = 1 THEN 'master'
		WHEN vfs.[database_id] = 2 THEN 'tempdb'
		WHEN vfs.[database_id] = 3 THEN 'model'
		WHEN vfs.[database_id] = 4 THEN 'msdb'
		ELSE gov.[database_name]
	 END AS [database_name]
	,vfs.[database_id]
	,vfs.[file_id]
	,CASE 
		WHEN vfs.[file_id] = 2 THEN 'LOG'
		ELSE 'ROWS' 
	 END AS [file_type]
	,vfs.[num_of_reads] AS [reads]
	,vfs.[num_of_bytes_read] AS [read_bytes]
	,vfs.[io_stall_read_ms] AS [read_latency_ms]
	,vfs.[io_stall_write_ms] AS [write_latency_ms]
	,vfs.[num_of_writes] AS [writes]
	,vfs.[num_of_bytes_written] AS [write_bytes]
	,vfs.[io_stall_queued_read_ms] AS [rg_read_stall_ms]
	,vfs.[io_stall_queued_write_ms] AS [rg_write_stall_ms]
	,[size_on_disk_bytes]
	,ISNULL([size_on_disk_bytes],0)/(1024*1024) AS [size_on_disk_mb]
FROM 
	sys.dm_io_virtual_file_stats(NULL,NULL) AS vfs
LEFT OUTER JOIN 
	sys.dm_user_db_resource_governance AS gov
ON vfs.[database_id] = gov.[database_id];
`

const sqlAzurePoolOsWaitStats = `
IF SERVERPROPERTY('EngineEdition') <> 5 
   OR NOT EXISTS (SELECT 1 FROM sys.database_service_objectives WHERE database_id = DB_ID() AND elastic_pool_name IS NOT NULL) BEGIN
	DECLARE @ErrorMessage AS nvarchar(500) = 'Telegraf - Connection string Server:'+ @@SERVERNAME + ',Database:' + DB_NAME() +' is not an Azure SQL database in an elastic pool. Check the database_type parameter in the telegraf configuration.';
	RAISERROR (@ErrorMessage,11,1)
	RETURN
END

SELECT
	 'sqlserver_waitstats' AS [measurement]
	,REPLACE(@@SERVERNAME,'\',':') AS [sql_instance]
	,(SELECT [elastic_pool_name] FROM sys.database_service_objectives WHERE database_id = DB_ID()) AS [elastic_pool_name]
	,[wait_type]
	,[waiting_tasks_count]
	,[wait_time_ms]
	,[max_wait_time_ms]
	,[signal_wait_time_ms]
	,[wait_time_ms]-[signal_wait_time_ms] AS [resource_wait_ms]
	,CASE
		WHEN ws.[wait_type] LIKE 'SOS_SCHEDULER_YIELD' THEN 'CPU'
		WHEN ws.[wait_type] = 'THREADPOOL' THEN 'Worker Thread'
		WHEN ws.[wait_type] LIKE 'LCK[_]%' THEN 'Lock'
		WHEN ws.[wait_type] LIKE 'LATCH[_]%' THEN 'Latch'
		WHEN ws.[wait_type] LIKE 'PAGELATCH[_]%' THEN 'Buffer Latch'
		WHEN ws.[wait_type] LIKE 'PAGEIOLATCH[_]%' THEN 'Buffer IO'
		WHEN ws.[wait_type] LIKE 'RESOURCE_SEMAPHORE_QUERY_COMPILE%' THEN 'Compilation'
		WHEN ws.[wait_type] LIKE 'CLR[_]%' OR ws.[wait_type] LIKE 'SQLCLR%' THEN 'SQL CLR'
		WHEN ws.[wait_type] LIKE 'DBMIRROR_%' THEN 'Mirroring'
		WHEN ws.[wait_type] LIKE 'DTC[_]%' OR ws.[wait_type] LIKE 'DTCNEW%' OR ws.[wait_type] LIKE 'TRAN_%' 
     		OR ws.[wait_type] LIKE 'XACT%' OR ws.[wait_type] LIKE 'MSQL_XACT%' THEN 'Transaction'
		WHEN ws.[wait_type] LIKE 'SLEEP[_]%' OR ws.[wait_type] IN (
			'LAZYWRITER_SLEEP', 'SQLTRACE_BUFFER_FLUSH', 'SQLTRACE_INCREMENTAL_FLUSH_SLEEP',
			'SQLTRACE_WAIT_ENTRIES', 'FT_IFTS_SCHEDULER_IDLE_WAIT', 'XE_DISPATCHER_WAIT',
			'REQUEST_FOR_DEADLOCK_SEARCH', 'LOGMGR_QUEUE', 'ONDEMAND_TASK_QUEUE',
			'CHECKPOINT_QUEUE', 'XE_TIMER_EVENT') THEN 'Idle'
		WHEN ws.[wait_type] IN (
			'ASYNC_IO_COMPLETION','BACKUPIO','CHKPT','WRITE_COMPLETION',
			'IO_QUEUE_LIMIT', 'IO_RETRY') THEN 'Other Disk IO'
		WHEN ws.[wait_type] LIKE 'PREEMPTIVE_%' THEN 'Preemptive'
		WHEN ws.[wait_type] LIKE 'BROKER[_]%' THEN 'Service Broker'
		WHEN ws.[wait_type] IN (
			'WRITELOG','LOGBUFFER','LOGMGR_RESERVE_APPEND',
			'LOGMGR_FLUSH', 'LOGMGR_PMM_LOG')  THEN 'Tran Log IO'
		WHEN ws.[wait_type] LIKE 'LOG_RATE%' then 'Log Rate Governor'
		WHEN ws.[wait_type] LIKE 'HADR_THROTTLE[_]%' 
			OR ws.[wait_type] = 'THROTTLE_LOG_RATE_LOG_STORAGE' THEN 'HADR Log Rate Governor'
		WHEN ws.[wait_type] LIKE 'RBIO_RG%' OR ws.[wait_type] LIKE 'WAIT_RBIO_RG%' THEN 'VLDB Log Rate Governor'
		WHEN ws.[wait_type] LIKE 'RBIO[_]%' OR ws.[wait_type] LIKE 'WAIT_RBIO[_]%' THEN 'VLDB RBIO'
		WHEN ws.[wait_type] IN(
			'ASYNC_NETWORK_IO','EXTERNAL_SCRIPT_NETWORK_IOF',
			'NET_WAITFOR_PACKET','PROXY_NETWORK_IO') THEN 'Network IO'
		WHEN ws.[wait_type] IN ( 'CXPACKET', 'CXCONSUMER')
			OR ws.[wait_type] LIKE 'HT%' or ws.[wait_type] LIKE 'BMP%'
			OR ws.[wait_type] LIKE 'BP%' THEN 'Parallelism'
		WHEN ws.[wait_type] IN(
			'CMEMTHREAD','CMEMPARTITIONED','EE_PMOLOCK','EXCHANGE',
			'RESOURCE_SEMAPHORE','MEMORY_ALLOCATION_EXT',
			'RESERVED_MEMORY_ALLOCATION_EXT', 'MEMORY_GRANT_UPDATE')  THEN 'Memory'
		WHEN ws.[wait_type] IN ('WAITFOR','WAIT_FOR_RESULTS')  THEN 'User Wait'
		WHEN ws.[wait_type] LIKE 'HADR[_]%' or ws.[wait_type] LIKE 'PWAIT_HADR%'
			OR ws.[wait_type] LIKE 'REPLICA[_]%' or ws.[wait_type] LIKE 'REPL_%' 
			OR ws.[wait_type] LIKE 'SE_REPL[_]%'
			OR ws.[wait_type] LIKE 'FCB_REPLICA%' THEN 'Replication' 
		WHEN ws.[wait_type] LIKE 'SQLTRACE[_]%' 
			OR ws.[wait_type] IN (
				'TRACEWRITE', 'SQLTRACE_LOCK', 'SQLTRACE_FILE_BUFFER', 'SQLTRACE_FILE_WRITE_IO_COMPLETION',
				'SQLTRACE_FILE_READ_IO_COMPLETION', 'SQLTRACE_PENDING_BUFFER_WRITERS', 'SQLTRACE_SHUTDOWN',
				'QUERY_TRACEOUT', 'TRACE_EVTNOTIF') THEN 'Tracing'
		WHEN ws.[wait_type] IN (
			'FT_RESTART_CRAWL', 'FULLTEXT GATHERER', 'MSSEARCH', 'FT_METADATA_MUTEX', 
  			'FT_IFTSHC_MUTEX', 'FT_IFTSISM_MUTEX', 'FT_IFTS_RWLOCK', 'FT_COMPROWSET_RWLOCK',
  			'FT_MASTER_MERGE', 'FT_PROPERTYLIST_CACHE', 'FT_MASTER_MERGE_COORDINATOR',
  			'PWAIT_RESOURCE_SEMAPHORE_FT_PARALLEL_QUERY_SYNC') THEN 'Full Text Search'
 		ELSE 'Other'
	 END AS [wait_category]
FROM sys.dm_os_wait_stats AS ws
WHERE
	ws.[wait_type] NOT IN (
        N'BROKER_EVENTHANDLER', N'BROKER_RECEIVE_WAITFOR', N'BROKER_TASK_STOP',
        N'BROKER_TO_FLUSH', N'BROKER_TRANSMITTER', N'CHECKPOINT_QUEUE',
        N'CHKPT', N'CLR_AUTO_EVENT', N'CLR_MANUAL_EVENT', N'CLR_SEMAPHORE',
        N'DBMIRROR_DBM_EVENT', N'DBMIRROR_EVENTS_QUEUE', N'DBMIRROR_QUEUE',
        N'DBMIRRORING_CMD', N'DIRTY_PAGE_POLL', N'DISPATCHER_QUEUE_SEMAPHORE',
        N'EXECSYNC', N'FSAGENT', N'FT_IFTS_SCHEDULER_IDLE_WAIT', N'FT_IFTSHC_MUTEX',
        N'HADR_CLUSAPI_CALL', N'HADR_FILESTREAM_IOMGR_IOCOMPLETION', N'HADR_LOGCAPTURE_WAIT',
        N'HADR_NOTIFICATION_DEQUEUE', N'HADR_TIMER_TASK', N'HADR_WORK_QUEUE',
        N'KSOURCE_WAKEUP', N'LAZYWRITER_SLEEP', N'LOGMGR_QUEUE',
        N'MEMORY_ALLOCATION_EXT', N'ONDEMAND_TASK_QUEUE',
        N'PARALLEL_REDO_WORKER_WAIT_WORK',
        N'PREEMPTIVE_HADR_LEASE_MECHANISM', N'PREEMPTIVE_SP_SERVER_DIAGNOSTICS',
        N'PREEMPTIVE_OS_LIBRARYOPS', N'PREEMPTIVE_OS_COMOPS', N'PREEMPTIVE_OS_CRYPTOPS',
        N'PREEMPTIVE_OS_PIPEOPS','PREEMPTIVE_OS_GENERICOPS', N'PREEMPTIVE_OS_VERIFYTRUST',
        N'PREEMPTIVE_OS_DEVICEOPS',
        N'PREEMPTIVE_XE_CALLBACKEXECUTE', N'PREEMPTIVE_XE_DISPATCHER',
        N'PREEMPTIVE_XE_GETTARGETSTATE', N'PREEMPTIVE_XE_SESSIONCOMMIT',
        N'PREEMPTIVE_XE_TARGETINIT', N'PREEMPTIVE_XE_TARGETFINALIZE',
        N'PWAIT_ALL_COMPONENTS_INITIALIZED', N'PWAIT_DIRECTLOGCONSUMER_GETNEXT',
        N'QDS_PERSIST_TASK_MAIN_LOOP_SLEEP', N'QDS_ASYNC_QUEUE',
        N'QDS_CLEANUP_STALE_QUERIES_TASK_MAIN_LOOP_SLEEP', N'REQUEST_FOR_DEADLOCK_SEARCH',
        N'RESOURCE_QUEUE', N'SERVER_IDLE_CHECK', N'SLEEP_BPOOL_FLUSH', N'SLEEP_DBSTARTUP',
        N'SLEEP_DCOMSTARTUP', N'SLEEP_MASTERDBREADY', N'SLEEP_MASTERMDREADY',
        N'SLEEP_MASTERUPGRADED', N'SLEEP_MSDBSTARTUP', N'SLEEP_SYSTEMTASK', N'SLEEP_TASK',
        N'SLEEP_TEMPDBSTARTUP', N'SNI_HTTP_ACCEPT', N'SP_SERVER_DIAGNOSTICS_SLEEP',
		N'SQLTRACE_BUFFER_FLUSH', N'SQLTRACE_INCREMENTAL_FLUSH_SLEEP', 
		N'SQLTRACE_WAIT_ENTRIES', N'WAIT_FOR_RESULTS', N'WAITFOR', N'WAITFOR_TASKSHUTDOWN',
		N'WAIT_XTP_HOST_WAIT', N'WAIT_XTP_OFFLINE_CKPT_NEW_LOG', N'WAIT_XTP_CKPT_CLOSE',
        N'XE_BUFFERMGR_ALLPROCESSED_EVENT', N'XE_DISPATCHER_JOIN',
        N'XE_DISPATCHER_WAIT', N'XE_LIVE_TARGET_TVF', N'XE_TIMER_EVENT',
        N'SOS_WORK_DISPATCHER','RESERVED_MEMORY_ALLOCATION_EXT','SQLTRACE_WAIT_ENTRIES',
		N'RBIO_COMM_RETRY')
AND [waiting_tasks_count] > 10
AND [wait_time_ms] > 100;
`

const sqlAzurePoolMemoryClerks = `
IF SERVERPROPERTY('EngineEdition') <> 5 
   OR NOT EXISTS (SELECT 1 FROM sys.database_service_objectives WHERE database_id = DB_ID() AND elastic_pool_name IS NOT NULL) BEGIN
	DECLARE @ErrorMessage AS nvarchar(500) = 'Telegraf - Connection string Server:'+ @@SERVERNAME + ',Database:' + DB_NAME() +' is not an Azure SQL database in an elastic pool. Check the database_type parameter in the telegraf configuration.';
	RAISERROR (@ErrorMessage,11,1)
	RETURN
END

SELECT
	 'sqlserver_memory_clerks' AS [measurement]
	,REPLACE(@@SERVERNAME, '\', ':') AS [sql_instance]
	,(SELECT [elastic_pool_name] FROM sys.database_service_objectives WHERE database_id = DB_ID()) AS [elastic_pool_name]
	,mc.[type] AS [clerk_type]
	,SUM(mc.[pages_kb]) AS [size_kb]
FROM 
	sys.dm_os_memory_clerks AS mc
GROUP BY
	mc.[type]
HAVING 
	SUM(mc.[pages_kb]) >= 1024
OPTION(RECOMPILE);
`

// Specific case on this query when cntr_type = 537003264 to return a percentage value between 0 and 100
// cf. https://docs.microsoft.com/en-us/sql/relational-databases/system-dynamic-management-views/sys-dm-os-performance-counters-transact-sql?view=azuresqldb-current
// Performance counters where the cntr_type column value is 537003264 display the ratio of a subset to its set as a percentage.
// For example, the Buffer Manager:Buffer cache hit ratio counter compares the total number of cache hits and the total number of cache lookups.
// As such, to get a snapshot-like reading of the last second only, you must compare the delta between the current value and the base value (denominator)
// between two collection points that are one second apart.
// The corresponding base value is the performance counter Buffer Manager:Buffer cache hit ratio base where the cntr_type column value is 1073939712.
const sqlAzurePoolPerformanceCounters = `
SET DEADLOCK_PRIORITY -10;
IF SERVERPROPERTY('EngineEdition') <> 5 
   OR NOT EXISTS (SELECT 1 FROM sys.database_service_objectives WHERE database_id = DB_ID() AND elastic_pool_name IS NOT NULL) BEGIN
	DECLARE @ErrorMessage AS nvarchar(500) = 'Telegraf - Connection string Server:'+ @@SERVERNAME + ',Database:' + DB_NAME() +' is not an Azure SQL database in an elastic pool. Check the database_type parameter in the telegraf configuration.';
	RAISERROR (@ErrorMessage,11,1)
	RETURN
END

DECLARE @PCounters TABLE
(
	[object_name] nvarchar(128),
	[counter_name] nvarchar(128),
	[instance_name] nvarchar(128),
	[cntr_value] bigint,
	[cntr_type] int
	Primary Key([object_name],[counter_name],[instance_name])
);

WITH PerfCounters AS (
	SELECT DISTINCT 
		 RTRIM(pc.[object_name]) AS [object_name]
		,RTRIM(pc.[counter_name]) AS [counter_name]
		,ISNULL(gov.[database_name], RTRIM(pc.instance_name)) AS [instance_name]
		,pc.[cntr_value] AS [cntr_value]
		,pc.[cntr_type] AS [cntr_type]
	FROM sys.dm_os_performance_counters AS pc
	LEFT JOIN sys.dm_user_db_resource_governance AS gov
	ON 
		TRY_CONVERT([uniqueidentifier], pc.[instance_name]) = gov.[physical_database_guid]
	WHERE
		/*filter out unnecessary SQL DB system database counters, other than master and tempdb*/
		NOT (pc.[object_name] LIKE 'MSSQL%:Databases%' AND pc.[instance_name] IN ('model','model_masterdb','model_userdb','msdb','mssqlsystemresource'))
		AND
		(
			pc.[counter_name] IN (
				 'SQL Compilations/sec'
				,'SQL Re-Compilations/sec'
				,'User Connections'
				,'Batch Requests/sec'
				,'Logouts/sec'
				,'Logins/sec'
				,'Processes blocked'
				,'Latch Waits/sec'
				,'Full Scans/sec'
				,'Index Searches/sec'
				,'Page Splits/sec'
				,'Page lookups/sec'
				,'Page reads/sec'
				,'Page writes/sec'
				,'Readahead pages/sec'
				,'Lazy writes/sec'
				,'Checkpoint pages/sec'
				,'Table Lock Escalations/sec'
				,'Page life expectancy'
				,'Log File(s) Size (KB)'
				,'Log File(s) Used Size (KB)'
				,'Data File(s) Size (KB)'
				,'Transactions/sec'
				,'Write Transactions/sec'
				,'Active Transactions'
				,'Log Growths'
				,'Active Temp Tables'
				,'Logical Connections'
				,'Temp Tables Creation Rate'
				,'Temp Tables For Destruction'
				,'Free Space in tempdb (KB)'
				,'Version Store Size (KB)'
				,'Memory Grants Pending'
				,'Memory Grants Outstanding'
				,'Free list stalls/sec'
				,'Buffer cache hit ratio'
				,'Buffer cache hit ratio base'
				,'Backup/Restore Throughput/sec'
				,'Total Server Memory (KB)'
				,'Target Server Memory (KB)'
				,'Log Flushes/sec'
				,'Log Flush Wait Time'
				,'Memory broker clerk size'
				,'Log Bytes Flushed/sec'
				,'Bytes Sent to Replica/sec'
				,'Log Send Queue'
				,'Bytes Sent to Transport/sec'
				,'Sends to Replica/sec'
				,'Bytes Sent to Transport/sec'
				,'Sends to Transport/sec'
				,'Bytes Received from Replica/sec'
				,'Receives from Replica/sec'
				,'Flow Control Time (ms/sec)'
				,'Flow Control/sec'
				,'Resent Messages/sec'
				,'Redone Bytes/sec'
				,'XTP Memory Used (KB)'
				,'Transaction Delay'
				,'Log Bytes Received/sec'
				,'Log Apply Pending Queue'
				,'Redone Bytes/sec'
				,'Recovery Queue'
				,'Log Apply Ready Queue'
				,'CPU usage %'
				,'CPU usage % base'
				,'Queued requests'
				,'Requests completed/sec'
				,'Blocked tasks'
				,'Active memory grant amount (KB)'
				,'Disk Read Bytes/sec'
				,'Disk Read IO Throttled/sec'
				,'Disk Read IO/sec'
				,'Disk Write Bytes/sec'
				,'Disk Write IO Throttled/sec'
				,'Disk Write IO/sec'
				,'Used memory (KB)'
				,'Forwarded Records/sec'
				,'Background Writer pages/sec'
				,'Percent Log Used'
				,'Log Send Queue KB'
				,'Redo Queue KB'
				,'Mirrored Write Transactions/sec'
				,'Group Commit Time'
				,'Group Commits/Sec'
				,'Workfiles Created/sec'
				,'Worktables Created/sec'
				,'Query Store CPU usage'
			) OR (
				   pc.[object_name] LIKE '%User Settable%'
				OR pc.[object_name] LIKE '%SQL Errors%'
				OR pc.[object_name] LIKE '%Batch Resp Statistics%'
			) OR (
				    pc.[instance_name] IN ('_Total')
				AND pc.[counter_name] IN (
					 'Lock Timeouts/sec'
					,'Lock Timeouts (timeout > 0)/sec'
					,'Number of Deadlocks/sec'
					,'Lock Waits/sec'
					,'Latch Waits/sec'
				)
			)
		)
)

INSERT INTO @PCounters select * from PerfCounters

SELECT
	 'sqlserver_performance' AS [measurement]
	,REPLACE(@@SERVERNAME,'\',':') AS [sql_instance]
	,pc.[object_name] AS [object]
	,pc.[counter_name] AS [counter]
	,CASE pc.[instance_name] WHEN '_Total' THEN 'Total' ELSE ISNULL(pc.[instance_name],'') END AS [instance]
	,CAST(
		 CASE WHEN pc.[cntr_type] = 537003264 AND base.[cntr_value] > 0 
			THEN (pc.[cntr_value] * 1.0) / (base.[cntr_value] * 1.0) * 100 
			ELSE pc.[cntr_value] 
		 END 
	 AS float) AS [value]
	,CAST(pc.[cntr_type] AS varchar(25)) AS [counter_type]
FROM @PCounters AS pc
LEFT OUTER JOIN @PCounters AS base
ON 
	pc.[counter_name] = REPLACE(base.[counter_name],' base','')
	AND pc.[object_name] = base.[object_name]
	AND pc.[instance_name] = base.[instance_name]
	AND base.[cntr_type] = 1073939712
WHERE
	pc.[cntr_type] <> 1073939712
OPTION(RECOMPILE)
`

const sqlAzurePoolSchedulers = `
IF SERVERPROPERTY('EngineEdition') <> 5 
   OR NOT EXISTS (SELECT 1 FROM sys.database_service_objectives WHERE database_id = DB_ID() AND elastic_pool_name IS NOT NULL) BEGIN
	DECLARE @ErrorMessage AS nvarchar(500) = 'Telegraf - Connection string Server:'+ @@SERVERNAME + ',Database:' + DB_NAME() +' is not an Azure SQL database in an elastic pool. Check the database_type parameter in the telegraf configuration.';
	RAISERROR (@ErrorMessage,11,1)
	RETURN
END

SELECT
	 'sqlserver_schedulers' AS [measurement]
	,REPLACE(@@SERVERNAME, '\', ':') AS [sql_instance]
	,(SELECT [elastic_pool_name] FROM sys.database_service_objectives WHERE database_id = DB_ID()) AS [elastic_pool_name]
	,[scheduler_id]
	,[cpu_id]
	,[status]
	,[is_online]
	,[is_idle]
	,[preemptive_switches_count]
	,[context_switches_count]
	,[idle_switches_count]
	,[current_tasks_count]
	,[runnable_tasks_count]
	,[current_workers_count]
	,[active_workers_count]
	,[work_queue_count]
	,[pending_disk_io_count]
	,[load_factor]
	,[failed_to_create_worker]
	,[quantum_length_us]
	,[yield_count]
	,[total_cpu_usage_ms]
	,[total_cpu_idle_capped_ms]
	,[total_scheduler_delay_ms]
	,[ideal_workers_limit]
FROM 
	sys.dm_os_schedulers;
`
