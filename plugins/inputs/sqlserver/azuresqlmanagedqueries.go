package sqlserver

import (
	_ "github.com/denisenkom/go-mssqldb" // go-mssqldb initialization
)

//------------------------------------------------------------------------------------------------
//------------------ Azure Managed Instance ------------------------------------------------------
//------------------------------------------------------------------------------------------------
const sqlAzureMIProperties = `
IF SERVERPROPERTY('EngineEdition') <> 8 BEGIN /*not Azure Managed Instance*/
	DECLARE @ErrorMessage AS nvarchar(500) = 'Telegraf - Connection string Server:'+ @@SERVERNAME + ',Database:' + DB_NAME() +' is not an Azure Managed Instance. Check the database_type parameter in the telegraf configuration.';
	RAISERROR (@ErrorMessage,11,1)
	RETURN
END

SELECT TOP 1 
	 'sqlserver_server_properties' AS [measurement]
	,REPLACE(@@SERVERNAME,'\',':') AS [sql_instance]
	,[virtual_core_count] AS [cpu_count]
	,(SELECT [process_memory_limit_mb] FROM sys.dm_os_job_object) AS [server_memory]
	,[sku]
	,SERVERPROPERTY('EngineEdition') AS [engine_edition]
	,[hardware_generation] AS [hardware_type]
	,cast([reserved_storage_mb] as bigint) AS [total_storage_mb]
	,cast(([reserved_storage_mb] - [storage_space_used_mb]) as bigint) AS [available_storage_mb]
	,(SELECT DATEDIFF(MINUTE,[sqlserver_start_time],GETDATE()) from sys.dm_os_sys_info) as [uptime]
	,SERVERPROPERTY('ProductVersion') AS [sql_version]
	,LEFT(@@VERSION,CHARINDEX(' - ',@@VERSION)) AS [sql_version_desc]
	,[db_online]
	,[db_restoring]
	,[db_recovering]
	,[db_recoveryPending]
	,[db_suspect]
	,DATABASEPROPERTYEX(DB_NAME(), 'Updateability') as replica_updateability
FROM sys.server_resource_stats
CROSS APPLY	(
	SELECT  
		 SUM( CASE WHEN [state] = 0 THEN 1 ELSE 0 END ) AS [db_online]
		,SUM( CASE WHEN [state] = 1 THEN 1 ELSE 0 END ) AS [db_restoring]
		,SUM( CASE WHEN [state] = 2 THEN 1 ELSE 0 END ) AS [db_recovering]
		,SUM( CASE WHEN [state] = 3 THEN 1 ELSE 0 END ) AS [db_recoveryPending]
		,SUM( CASE WHEN [state] = 4 THEN 1 ELSE 0 END ) AS [db_suspect]
		,SUM( CASE WHEN [state] IN (6,10) THEN 1 ELSE 0 END ) AS [db_offline]
	FROM sys.databases
) AS dbs	
ORDER BY 
	[start_time] DESC;
`

const sqlAzureMIResourceStats = `
IF SERVERPROPERTY('EngineEdition') <> 8 BEGIN /*not Azure Managed Instance*/
	DECLARE @ErrorMessage AS nvarchar(500) = 'Telegraf - Connection string Server:'+ @@SERVERNAME + ',Database:' + DB_NAME() +' is not an Azure Managed Instance. Check the database_type parameter in the telegraf configuration.';
	RAISERROR (@ErrorMessage,11,1)
	RETURN
END

SELECT TOP(1)
	 'sqlserver_azure_db_resource_stats' AS [measurement]
	,REPLACE(@@SERVERNAME,'\',':') AS [sql_instance]
	,cast([avg_cpu_percent] as float) as [avg_cpu_percent]
	,DATABASEPROPERTYEX(DB_NAME(), 'Updateability') as replica_updateability
FROM
    sys.server_resource_stats
ORDER BY
    [end_time] DESC;
`

const sqlAzureMIResourceGovernance string = `
IF SERVERPROPERTY('EngineEdition') <> 8 BEGIN /*not Azure Managed Instance*/
	DECLARE @ErrorMessage AS nvarchar(500) = 'Telegraf - Connection string Server:'+ @@SERVERNAME + ',Database:' + DB_NAME() +' is not an Azure Managed Instance. Check the database_type parameter in the telegraf configuration.';
	RAISERROR (@ErrorMessage,11,1)
	RETURN
END

SELECT
	 'sqlserver_instance_resource_governance' AS [measurement]
	,REPLACE(@@SERVERNAME,'\',':') AS [sql_instance]
	,[instance_cap_cpu]
	,[instance_max_log_rate]
	,[instance_max_worker_threads]
	,[tempdb_log_file_number]
	,[volume_local_iops]
	,[volume_external_xstore_iops]
	,[volume_managed_xstore_iops]
	,[volume_type_local_iops] as [voltype_local_iops]
	,[volume_type_managed_xstore_iops] as [voltype_man_xtore_iops]
	,[volume_type_external_xstore_iops] as [voltype_ext_xtore_iops]
	,[volume_external_xstore_iops] as [vol_ext_xtore_iops]
	,DATABASEPROPERTYEX(DB_NAME(), 'Updateability') as replica_updateability
FROM sys.dm_instance_resource_governance;
`

const sqlAzureMIDatabaseIO = `
SET DEADLOCK_PRIORITY -10;
IF SERVERPROPERTY('EngineEdition') <> 8 BEGIN /*not Azure Managed Instance*/
	DECLARE @ErrorMessage AS nvarchar(500) = 'Telegraf - Connection string Server:'+ @@SERVERNAME + ',Database:' + DB_NAME() +' is not an Azure Managed Instance. Check the database_type parameter in the telegraf configuration.';
	RAISERROR (@ErrorMessage,11,1)
	RETURN
END

SELECT
	'sqlserver_database_io' AS [measurement]
	,REPLACE(@@SERVERNAME,'\',':') AS [sql_instance]
	,COALESCE(mf.[physical_name],'RBPEX') AS [physical_filename]	--RPBEX = Resilient Buffer Pool Extension
	,COALESCE(mf.[name],'RBPEX') AS [logical_filename]	--RPBEX = Resilient Buffer Pool Extension	
	,mf.[type_desc] AS [file_type]
	,vfs.[io_stall_read_ms] AS [read_latency_ms]
	,vfs.[num_of_reads] AS [reads]
	,vfs.[num_of_bytes_read] AS [read_bytes]
	,vfs.[io_stall_write_ms] AS [write_latency_ms]
	,vfs.[num_of_writes] AS [writes]
	,vfs.[num_of_bytes_written] AS [write_bytes]
	,vfs.io_stall_queued_read_ms AS [rg_read_stall_ms] 
	,vfs.io_stall_queued_write_ms AS [rg_write_stall_ms]
	,DATABASEPROPERTYEX(DB_NAME(), 'Updateability') as replica_updateability
FROM sys.dm_io_virtual_file_stats(NULL, NULL) AS vfs
LEFT OUTER JOIN sys.master_files AS mf WITH (NOLOCK)
	ON vfs.[database_id] = mf.[database_id] 
	AND vfs.[file_id] = mf.[file_id]
WHERE
	vfs.[database_id] < 32760
`

const sqlAzureMIMemoryClerks = `
IF SERVERPROPERTY('EngineEdition') <> 8 BEGIN /*not Azure Managed Instance*/
	DECLARE @ErrorMessage AS nvarchar(500) = 'Telegraf - Connection string Server:'+ @@SERVERNAME + ',Database:' + DB_NAME() +' is not an Azure Managed Instance. Check the database_type parameter in the telegraf configuration.';
	RAISERROR (@ErrorMessage,11,1)
	RETURN
END

SELECT
	 'sqlserver_memory_clerks' AS [measurement]
	,REPLACE(@@SERVERNAME, '\', ':') AS [sql_instance]
	,mc.[type] AS [clerk_type]
	,SUM(mc.[pages_kb]) AS [size_kb]
	,DATABASEPROPERTYEX(DB_NAME(), 'Updateability') as replica_updateability
FROM sys.[dm_os_memory_clerks] AS mc WITH (NOLOCK)
GROUP BY
	 mc.[type]
HAVING
	SUM(mc.[pages_kb]) >= 1024
OPTION(RECOMPILE);
`

const sqlAzureMIOsWaitStats = `
IF SERVERPROPERTY('EngineEdition') <> 8 BEGIN /*not Azure Managed Instance*/
	DECLARE @ErrorMessage AS nvarchar(500) = 'Telegraf - Connection string Server:'+ @@SERVERNAME + ',Database:' + DB_NAME() +' is not an Azure Managed Instance. Check the database_type parameter in the telegraf configuration.';
	RAISERROR (@ErrorMessage,11,1)
	RETURN
END

SELECT
	 'sqlserver_waitstats' AS [measurement]
	,REPLACE(@@SERVERNAME,'\',':') AS [sql_instance]
	,ws.[wait_type]
	,[wait_time_ms]
	,[wait_time_ms] - [signal_wait_time_ms] AS [resource_wait_ms]
	,[signal_wait_time_ms]
	,[max_wait_time_ms]
	,[waiting_tasks_count]
	,CASE 
		WHEN ws.[wait_type] LIKE 'SOS_SCHEDULER_YIELD' then 'CPU'
		WHEN ws.[wait_type] = 'THREADPOOL' THEN 'Worker Thread'
		WHEN ws.[wait_type] LIKE 'LCK[_]%' THEN 'Lock'
		WHEN ws.[wait_type] LIKE 'LATCH[_]%' THEN 'Latch'
		WHEN ws.[wait_type] LIKE 'PAGELATCH[_]%' THEN 'Buffer Latch'
		WHEN ws.[wait_type] LIKE 'PAGEIOLATCH[_]%' THEN 'Buffer IO'
		WHEN ws.[wait_type] LIKE 'RESOURCE_SEMAPHORE_QUERY_COMPILE%' THEN 'Compilation'
		WHEN ws.[wait_type] LIKE 'CLR[_]%' or ws.[wait_type] like 'SQLCLR%' THEN 'SQL CLR'
		WHEN ws.[wait_type] LIKE 'DBMIRROR_%' THEN 'Mirroring'
		WHEN ws.[wait_type] LIKE 'DTC[_]%' or ws.[wait_type] LIKE 'DTCNEW%' or ws.[wait_type] LIKE 'TRAN_%' 
     		or ws.[wait_type] LIKE 'XACT%' or ws.[wait_type] like 'MSQL_XACT%' THEN 'Transaction'
		WHEN ws.[wait_type] LIKE 'SLEEP[_]%'
			or ws.[wait_type] IN (
				'LAZYWRITER_SLEEP', 'SQLTRACE_BUFFER_FLUSH', 'SQLTRACE_INCREMENTAL_FLUSH_SLEEP',
				'SQLTRACE_WAIT_ENTRIES', 'FT_IFTS_SCHEDULER_IDLE_WAIT', 'XE_DISPATCHER_WAIT',
				'REQUEST_FOR_DEADLOCK_SEARCH', 'LOGMGR_QUEUE', 'ONDEMAND_TASK_QUEUE',
				'CHECKPOINT_QUEUE', 'XE_TIMER_EVENT') THEN 'Idle'
		WHEN ws.[wait_type] IN(
			'ASYNC_IO_COMPLETION','BACKUPIO','CHKPT','WRITE_COMPLETION',
			'IO_QUEUE_LIMIT', 'IO_RETRY') THEN 'Other Disk IO'
		WHEN ws.[wait_type] LIKE 'PREEMPTIVE_%' THEN 'Preemptive'
		WHEN ws.[wait_type] LIKE 'BROKER[_]%' THEN 'Service Broker'
		WHEN ws.[wait_type] IN (
			'WRITELOG','LOGBUFFER','LOGMGR_RESERVE_APPEND',
			'LOGMGR_FLUSH', 'LOGMGR_PMM_LOG')  THEN 'Tran Log IO'
		WHEN ws.[wait_type] LIKE 'LOG_RATE%' then 'Log Rate Governor'
		WHEN ws.[wait_type] LIKE 'HADR_THROTTLE[_]%' 
			or ws.[wait_type] = 'THROTTLE_LOG_RATE_LOG_STORAGE' THEN 'HADR Log Rate Governor'
		WHEN ws.[wait_type] LIKE 'RBIO_RG%' or ws.[wait_type] like 'WAIT_RBIO_RG%' then 'VLDB Log Rate Governor'
		WHEN ws.[wait_type] LIKE 'RBIO[_]%' or ws.[wait_type] like 'WAIT_RBIO[_]%' then 'VLDB RBIO'
		WHEN ws.[wait_type] IN(
			'ASYNC_NETWORK_IO','EXTERNAL_SCRIPT_NETWORK_IOF',
			'NET_WAITFOR_PACKET','PROXY_NETWORK_IO') THEN 'Network IO'
		WHEN ws.[wait_type] IN ( 'CXPACKET', 'CXCONSUMER')
			or ws.[wait_type] like 'HT%' or ws.[wait_type] like 'BMP%'
			or ws.[wait_type] like 'BP%' THEN 'Parallelism'
		WHEN ws.[wait_type] IN(
			'CMEMTHREAD','CMEMPARTITIONED','EE_PMOLOCK','EXCHANGE',
			'RESOURCE_SEMAPHORE','MEMORY_ALLOCATION_EXT',
			'RESERVED_MEMORY_ALLOCATION_EXT', 'MEMORY_GRANT_UPDATE')  THEN 'Memory'
		WHEN ws.[wait_type] IN ('WAITFOR','WAIT_FOR_RESULTS')  THEN 'User Wait'
		WHEN ws.[wait_type] LIKE 'HADR[_]%' or ws.[wait_type] LIKE 'PWAIT_HADR%'
			or ws.[wait_type] LIKE 'REPLICA[_]%' or ws.[wait_type] LIKE 'REPL_%' 
			or ws.[wait_type] LIKE 'SE_REPL[_]%'
			or ws.[wait_type] LIKE 'FCB_REPLICA%' THEN 'Replication' 
		WHEN ws.[wait_type] LIKE 'SQLTRACE[_]%' 
			or ws.[wait_type] IN (
				'TRACEWRITE', 'SQLTRACE_LOCK', 'SQLTRACE_FILE_BUFFER', 'SQLTRACE_FILE_WRITE_IO_COMPLETION',
				'SQLTRACE_FILE_READ_IO_COMPLETION', 'SQLTRACE_PENDING_BUFFER_WRITERS', 'SQLTRACE_SHUTDOWN',
				'QUERY_TRACEOUT', 'TRACE_EVTNOTIF') THEN 'Tracing'
		WHEN ws.[wait_type] IN (
			'FT_RESTART_CRAWL', 'FULLTEXT GATHERER', 'MSSEARCH', 'FT_METADATA_MUTEX', 
  			'FT_IFTSHC_MUTEX', 'FT_IFTSISM_MUTEX', 'FT_IFTS_RWLOCK', 'FT_COMPROWSET_RWLOCK',
  			'FT_MASTER_MERGE', 'FT_PROPERTYLIST_CACHE', 'FT_MASTER_MERGE_COORDINATOR',
  			'PWAIT_RESOURCE_SEMAPHORE_FT_PARALLEL_QUERY_SYNC') THEN 'Full Text Search'
 		ELSE 'Other'
	END as [wait_category]
	,DATABASEPROPERTYEX(DB_NAME(), 'Updateability') as replica_updateability
FROM sys.dm_os_wait_stats AS ws WITH (NOLOCK)
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
        N'QDS_PERSIST_TASK_MAIN_LOOP_SLEEP',
        N'QDS_ASYNC_QUEUE',
        N'QDS_CLEANUP_STALE_QUERIES_TASK_MAIN_LOOP_SLEEP', N'REQUEST_FOR_DEADLOCK_SEARCH',
        N'RESOURCE_QUEUE', N'SERVER_IDLE_CHECK', N'SLEEP_BPOOL_FLUSH', N'SLEEP_DBSTARTUP',
        N'SLEEP_DCOMSTARTUP', N'SLEEP_MASTERDBREADY', N'SLEEP_MASTERMDREADY',
        N'SLEEP_MASTERUPGRADED', N'SLEEP_MSDBSTARTUP', N'SLEEP_SYSTEMTASK', N'SLEEP_TASK',
        N'SLEEP_TEMPDBSTARTUP', N'SNI_HTTP_ACCEPT', N'SP_SERVER_DIAGNOSTICS_SLEEP',
		N'SQLTRACE_BUFFER_FLUSH', N'SQLTRACE_INCREMENTAL_FLUSH_SLEEP',
        N'SQLTRACE_WAIT_ENTRIES',
        N'WAIT_FOR_RESULTS', N'WAITFOR', N'WAITFOR_TASKSHUTDOWN', N'WAIT_XTP_HOST_WAIT',
        N'WAIT_XTP_OFFLINE_CKPT_NEW_LOG', N'WAIT_XTP_CKPT_CLOSE',
        N'XE_BUFFERMGR_ALLPROCESSED_EVENT', N'XE_DISPATCHER_JOIN',
        N'XE_DISPATCHER_WAIT', N'XE_LIVE_TARGET_TVF', N'XE_TIMER_EVENT',
        N'SOS_WORK_DISPATCHER','RESERVED_MEMORY_ALLOCATION_EXT','SQLTRACE_WAIT_ENTRIES',
		N'RBIO_COMM_RETRY')
AND [waiting_tasks_count] > 10
AND [wait_time_ms] > 100;
`

const sqlAzureMIPerformanceCounters = `
SET DEADLOCK_PRIORITY -10;
IF SERVERPROPERTY('EngineEdition') <> 8 BEGIN /*not Azure Managed Instance*/
	DECLARE @ErrorMessage AS nvarchar(500) = 'Telegraf - Connection string Server:'+ @@SERVERNAME + ',Database:' + DB_NAME() +' is not an Azure Managed Instance. Check the database_type parameter in the telegraf configuration.';
	RAISERROR (@ErrorMessage,11,1)
	RETURN
END

DECLARE @PCounters TABLE
(
	[object_name] nvarchar(128),
	[counter_name] nvarchar(128),
	[instance_name] nvarchar(128),
	[cntr_value] bigint,
	[cntr_type] INT ,
	Primary Key([object_name],[counter_name],[instance_name])
);

WITH PerfCounters AS (
	SELECT DISTINCT
	 RTrim(spi.[object_name]) [object_name]
	,RTrim(spi.[counter_name]) [counter_name]
	,CASE WHEN (
		   RTRIM(spi.[object_name]) LIKE '%:Databases'
		OR RTRIM(spi.[object_name]) LIKE '%:Database Replica'
		OR RTRIM(spi.[object_name]) LIKE '%:Catalog Metadata'
		OR RTRIM(spi.[object_name]) LIKE '%:Query Store'
		OR RTRIM(spi.[object_name]) LIKE '%:Columnstore'
		OR RTRIM(spi.[object_name]) LIKE '%:Advanced Analytics')
		AND TRY_CONVERT([uniqueidentifier], spi.[instance_name]) IS NOT NULL -- for cloud only
			THEN ISNULL(d.[name],RTRIM(spi.instance_name)) -- Elastic Pools counters exist for all databases but sys.databases only has current DB value
		WHEN 
			RTRIM([object_name]) LIKE '%:Availability Replica'
			AND TRY_CONVERT([uniqueidentifier], spi.[instance_name]) IS NOT NULL -- for cloud only
				THEN ISNULL(d.[name],RTRIM(spi.[instance_name])) + RTRIM(SUBSTRING(spi.[instance_name], 37, LEN(spi.[instance_name])))
		ELSE RTRIM(spi.instance_name)
	END AS [instance_name]
	,CAST(spi.[cntr_value] AS BIGINT) AS [cntr_value]
	,spi.[cntr_type]
	FROM sys.dm_os_performance_counters AS spi 
	LEFT JOIN sys.databases AS d
		ON LEFT(spi.[instance_name], 36) -- some instance_name values have an additional identifier appended after the GUID
		= CASE
			/*in SQL DB standalone, physical_database_name for master is the GUID of the user database*/
			WHEN d.[name] = 'master' AND TRY_CONVERT([uniqueidentifier], d.[physical_database_name]) IS NOT NULL
				THEN d.[name]
			ELSE d.[physical_database_name]
		END
	WHERE
		counter_name IN (
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
			,'Distributed Query'
			,'DTC calls'
			,'Query Store CPU usage'
		) OR (
			spi.[object_name] LIKE '%User Settable%'
			OR spi.[object_name] LIKE '%SQL Errors%'
			OR spi.[object_name] LIKE '%Batch Resp Statistics%'
		) OR (
			spi.[instance_name] IN ('_Total')
			AND spi.[counter_name] IN (
				 'Lock Timeouts/sec'
				,'Lock Timeouts (timeout > 0)/sec'
				,'Number of Deadlocks/sec'
				,'Lock Waits/sec'
				,'Latch Waits/sec'
			)
		)
)

INSERT INTO @PCounters select * from PerfCounters

SELECT 
	'sqlserver_performance' AS [measurement]
	,REPLACE(@@SERVERNAME,'\',':') AS [sql_instance]
	,pc.[object_name] AS [object]
	,pc.[counter_name] AS [counter]
	,CASE pc.[instance_name] 
		WHEN '_Total' THEN 'Total' 
		ELSE ISNULL(pc.[instance_name],'') 
	END AS [instance]
	,CAST(CASE WHEN pc.[cntr_type] = 537003264 AND pc1.[cntr_value] > 0 THEN (pc.[cntr_value] * 1.0) / (pc1.[cntr_value] * 1.0) * 100 ELSE pc.[cntr_value] END AS float(10)) AS [value]
	,cast(pc.[cntr_type] as varchar(25)) as [counter_type]
	,DATABASEPROPERTYEX(DB_NAME(), 'Updateability') as replica_updateability
from @PCounters pc
LEFT OUTER JOIN @PCounters AS pc1
	ON (
		pc.[counter_name] = REPLACE(pc1.[counter_name],' base','')
		OR pc.[counter_name] = REPLACE(pc1.[counter_name],' base',' (ms)')
	)
	AND pc.[object_name] = pc1.[object_name]
	AND pc.[instance_name] = pc1.[instance_name]
	AND pc1.[counter_name] LIKE '%base'
WHERE
	pc.[counter_name] NOT LIKE '% base'
OPTION (RECOMPILE);
`

const sqlAzureMIRequests string = `
IF SERVERPROPERTY('EngineEdition') <> 8 BEGIN /*not Azure Managed Instance*/
	DECLARE @ErrorMessage AS nvarchar(500) = 'Telegraf - Connection string Server:'+ @@SERVERNAME + ',Database:' + DB_NAME() +' is not an Azure Managed Instance. Check the database_type parameter in the telegraf configuration.';
	RAISERROR (@ErrorMessage,11,1)
	RETURN
END



SELECT
	 [measurement],[sql_instance],[database_name],[session_id]
	,ISNULL([request_id],0) AS [request_id]
	,[blocking_session_id],[status],[cpu_time_ms]
	,[total_elapsed_time_ms],[logical_reads],[writes]
	,[command],[wait_time_ms],[wait_type]
	,[wait_resource],[program_name]
	,[host_name],[nt_user_name],[login_name]
	,[transaction_isolation_level],[granted_query_memory_pages],[percent_complete]
	,[statement_text],[objectid],[stmt_object_name]
	,[stmt_db_name],[query_hash],[query_plan_hash]
	,replica_updateability
	,[session_db_name],[open_transaction]
FROM (
	SELECT	
		'sqlserver_requests' AS [measurement]
		,REPLACE(@@SERVERNAME,'\',':') AS [sql_instance]
		,DB_NAME() as [database_name]
		,s.[session_id]
		,ISNULL(r.[request_id], 0) as [request_id]	
		,DB_NAME(COALESCE(r.[database_id], s.[database_id])) AS [session_db_name]
		,COALESCE(r.[status], s.[status]) AS [status]
		,COALESCE(r.[cpu_time], s.[cpu_time]) AS [cpu_time_ms]
		,COALESCE(r.[total_elapsed_time], s.[total_elapsed_time]) AS [total_elapsed_time_ms]
		,COALESCE(r.[logical_reads], s.[logical_reads]) AS [logical_reads]
		,COALESCE(r.[writes], s.[writes]) AS [writes]
		,r.[command]
		,r.[wait_time] AS [wait_time_ms]
		,r.[wait_type]
		,r.[wait_resource]
		,NULLIF(r.[blocking_session_id],0) AS [blocking_session_id]
		,s.[program_name]
		,s.[host_name]
		,s.[nt_user_name]
		,s.[login_name]
		,COALESCE(r.[open_transaction_count], s.[open_transaction_count]) AS [open_transaction]
		,LEFT (CASE COALESCE(r.[transaction_isolation_level], s.[transaction_isolation_level])
			WHEN 0 THEN '0-Read Committed' 
			WHEN 1 THEN '1-Read Uncommitted (NOLOCK)' 
			WHEN 2 THEN '2-Read Committed' 
			WHEN 3 THEN '3-Repeatable Read' 
			WHEN 4 THEN '4-Serializable' 
			WHEN 5 THEN '5-Snapshot' 
			ELSE CONVERT (varchar(30), r.[transaction_isolation_level]) + '-UNKNOWN' 
		END, 30) AS [transaction_isolation_level]
		,r.[granted_query_memory] AS [granted_query_memory_pages]
		,r.[percent_complete]
		,SUBSTRING(
			qt.[text], 
			r.[statement_start_offset] / 2 + 1,
			(CASE WHEN r.[statement_end_offset] = -1
				THEN DATALENGTH(qt.[text])
				ELSE r.[statement_end_offset]
			END - r.[statement_start_offset]) / 2 + 1
		) AS [statement_text]
		,qt.[objectid]
		,QUOTENAME(OBJECT_SCHEMA_NAME(qt.[objectid], qt.[dbid])) + '.' +  QUOTENAME(OBJECT_NAME(qt.[objectid], qt.[dbid])) as [stmt_object_name]
		,DB_NAME(qt.[dbid]) AS [stmt_db_name]
		,CONVERT(varchar(20),r.[query_hash],1) AS [query_hash]
		,CONVERT(varchar(20),r.[query_plan_hash],1) AS [query_plan_hash]
		,DATABASEPROPERTYEX(DB_NAME(), 'Updateability') as replica_updateability
		,s.[is_user_process]
		,[blocking_or_blocked] = COUNT(*) OVER(PARTITION BY ISNULL(NULLIF(r.[blocking_session_id], 0),s.[session_id]))
	FROM sys.dm_exec_sessions AS s
	LEFT OUTER JOIN sys.dm_exec_requests AS r 
		ON s.[session_id] = r.[session_id]
	OUTER APPLY sys.dm_exec_sql_text(r.[sql_handle]) AS qt
) AS data
WHERE
	[blocking_or_blocked] > 1	--Always include blocking or blocked sessions/requests
	OR (
		[request_id] IS NOT NULL	--A request must exists
		AND (	--Always fetch user process (in any state), fetch system process only if active
			[is_user_process] = 1
			OR [status] COLLATE Latin1_General_BIN NOT IN ('background', 'sleeping')
		)
	)  
OPTION(MAXDOP 1);
`

const sqlAzureMISchedulers string = `
IF SERVERPROPERTY('EngineEdition') <> 8 BEGIN /*not Azure Managed Instance*/
	DECLARE @ErrorMessage AS nvarchar(500) = 'Telegraf - Connection string Server:'+ @@SERVERNAME + ',Database:' + DB_NAME() +' is not an Azure Managed Instance. Check the database_type parameter in the telegraf configuration.';
	RAISERROR (@ErrorMessage,11,1)
	RETURN
END

SELECT
	 'sqlserver_schedulers' AS [measurement]
	,REPLACE(@@SERVERNAME, '\', ':') AS [sql_instance]
	,CAST(s.[scheduler_id] AS VARCHAR(4)) AS [scheduler_id]
	,CAST(s.[cpu_id] AS VARCHAR(4)) AS [cpu_id]
	,s.[is_online]
	,s.[is_idle]
	,s.[preemptive_switches_count]
	,s.[context_switches_count]
	,s.[current_tasks_count]
	,s.[runnable_tasks_count]
	,s.[current_workers_count]
	,s.[active_workers_count]
	,s.[work_queue_count]
	,s.[pending_disk_io_count]
	,s.[load_factor]
	,s.[yield_count]
	,s.[total_cpu_usage_ms]
	,s.[total_scheduler_delay_ms]
	,DATABASEPROPERTYEX(DB_NAME(), 'Updateability') as replica_updateability
FROM sys.dm_os_schedulers AS s
`
