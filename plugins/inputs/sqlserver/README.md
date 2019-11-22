# SQL Server Input Plugin

The `sqlserver` plugin provides metrics for your SQL Server instance. It
currently works with SQL Server 2008 SP3 and newer. Recorded metrics are
lightweight and use Dynamic Management Views supplied by SQL Server.

### Additional Setup:

You have to create a login on every instance you want to monitor, with following script:

```sql
USE master;
GO
CREATE LOGIN [telegraf] WITH PASSWORD = N'mystrongpassword';
GO
GRANT VIEW SERVER STATE TO [telegraf];
GO
GRANT VIEW ANY DEFINITION TO [telegraf];
GO
```

For Azure SQL Database, you require the View Database State permission and can create a user with a password directly in the database.

```sql
CREATE USER [telegraf] WITH PASSWORD = N'mystrongpassword';
GO
GRANT VIEW DATABASE STATE TO [telegraf];
GO
```

### Configuration:

```toml
[agent]
  ## Default data collection interval for all inputs, can be changed as per collection interval needs
  interval = "10s"

# Read metrics from Microsoft SQL Server
[[inputs.sqlserver]]
  ## Specify instances to monitor with a list of connection strings.
  ## All connection parameters are optional.
  ## By default, the host is localhost, listening on default port, TCP 1433.
  ##   for Windows, the user is the currently running AD user (SSO).
  ##   See https://github.com/denisenkom/go-mssqldb for detailed connection
  ##   parameters, in particular, tls connections can be created like so:
  ##   "encrypt=true;certificate=<cert>;hostNameInCertificate=<SqlServer host fqdn>"
  # servers = [
  #  "Server=192.168.1.10;Port=1433;User Id=<user>;Password=<pw>;app name=telegraf;log=1;",
  # ]

  ## Optional parameter, setting this to 2 will use a new version
  ## of the collection queries that break compatibility with the original
  ## dashboards. All new functionality is under V2
  query_version = 2

  ## If you are using AzureDB, setting this to true will gather resource utilization metrics
  # azuredb = true

  ## If you would like to exclude some of the metrics queries, list them here
  ## Possible choices:
  ## - PerformanceCounters
  ## - WaitStatsCategorized
  ## - DatabaseIO
  ## - DatabaseProperties
  ## - CPUHistory
  ## - DatabaseSize
  ## - DatabaseStats
  ## - MemoryClerk
  ## - VolumeSpace
  ## - Schedulers
  ## - AzureDBResourceStats
  ## - AzureDBResourceGovernance
  ## - SqlRequests
  ## - ServerProperties
  exclude_query = [ 'Schedulers' , 'SqlRequests']

  ## Sets a timeout for the query in milliseconds.  A value of 0 or less will be ignored.
  ## Define time in the format 5s = 5 seconds, 10 ms = 10 microseconds, 1 us = 1 microsecond, etc
  # query_timeout = 0
```

### Metrics:

To provide backwards compatibility, this plugin support two versions of metrics queries.

**Note**: Version 2 queries are not backwards compatible with the old queries. Any dashboards or queries based on the old query format will not work with the new format. The version 2 queries are written in such a way as to only gather SQL specific metrics (no disk space or overall CPU related metrics) and they only report raw metrics, no math has been done to calculate deltas. To graph this data you must calculate deltas in your dashboarding software.

#### Version 1 (deprecated in 1.6):

The original metrics queries provide:

- _Performance counters_: 1000+ metrics from `sys.dm_os_performance_counters`
- _Performance metrics_: special performance and ratio metrics
- _Wait stats_: wait tasks categorized from `sys.dm_os_wait_stats`
- _Memory clerk_: memory breakdown from `sys.dm_os_memory_clerks`
- _Database size_: databases size trend from `sys.dm_io_virtual_file_stats`
- _Database IO_: databases I/O from `sys.dm_io_virtual_file_stats`
- _Database latency_: databases latency from `sys.dm_io_virtual_file_stats`
- _Database properties_: databases properties, state and recovery model, from `sys.databases`
- _OS Volume_: available, used and total space from `sys.dm_os_volume_stats`
- _CPU_: cpu usage from `sys.dm_os_ring_buffers`

If you are using the original queries all stats have the following tags:

- `servername`: hostname:instance
- `type`: type of stats to easily filter measurements

#### Version 2:

The new (version 2) metrics provide:

- _Database IO_: IO stats from `sys.dm_io_virtual_file_stats`
- _Memory Clerk_: Memory clerk breakdown from `sys.dm_os_memory_clerks`, most clerks have been given a friendly name.
- _Performance Counters_: A select list of performance counters from `sys.dm_os_performance_counters`. Some of the important metrics included:
  - _Activity_: Transactions/sec/database, Batch requests/sec, blocked processes, + more
  - _Availability Groups_: Bytes sent to replica, Bytes received from replica, Log bytes received, Log send queue, transaction delay, + more
  - _Log activity_: Log bytes flushed/sec, Log flushes/sec, Log Flush Wait Time
  - _Memory_: PLE, Page reads/sec, Page writes/sec, + more
  - _TempDB_: Free space, Version store usage, Active temp tables, temp table creation rate, + more
  - _Resource Governor_: CPU Usage, Requests/sec, Queued Requests, and Blocked tasks per workload group + more
- _Server properties_: Number of databases in all possible states (online, offline, suspect, etc.), cpu count, physical memory, SQL Server service uptime, and SQL Server version. In the case of Azure SQL relevent properties such as Tier, #Vcores, Memory etc.
- _Wait stats_: Wait time in ms, number of waiting tasks, resource wait time, signal wait time, max wait time in ms, wait type, and wait category. The waits are categorized using the same categories used in Query Store.
- _Schedulers_ - This captures sys.dm_os_schedulers.
- _SqlRequests_ - This captures a snapshot of dm_exec_requests and dm_exec_sessions that gives you running requests as well as wait types and blocking sessions
- _Azure Managed Instances_
  - Stats from `sys.server_resource_stats`:
    - cpu_count
    - server_memory
    - sku
    - engine_edition
    - hardware_type
    - total_storage_mb
    - available_storage_mb
    - uptime
  - Resource governance stats from sys.dm_instance_resource_governance
- _Azure SQL Database_
  - Stats from sys.dm_db_wait_stats
  - Resource governance stats from sys.dm_user_db_resource_governance
  - Stats from sys.dm_db_resource_stats

The following metrics can be used directly, with no delta calculations:

- SQLServer:Buffer Manager\Buffer cache hit ratio
- SQLServer:Buffer Manager\Page life expectancy
- SQLServer:Buffer Node\Page life expectancy
- SQLServer:Database Replica\Log Apply Pending Queue
- SQLServer:Database Replica\Log Apply Ready Queue
- SQLServer:Database Replica\Log Send Queue
- SQLServer:Database Replica\Recovery Queue
- SQLServer:Databases\Data File(s) Size (KB)
- SQLServer:Databases\Log File(s) Size (KB)
- SQLServer:Databases\Log File(s) Used Size (KB)
- SQLServer:Databases\XTP Memory Used (KB)
- SQLServer:General Statistics\Active Temp Tables
- SQLServer:General Statistics\Processes blocked
- SQLServer:General Statistics\Temp Tables For Destruction
- SQLServer:General Statistics\User Connections
- SQLServer:Memory Broker Clerks\Memory broker clerk size
- SQLServer:Memory Manager\Memory Grants Pending
- SQLServer:Memory Manager\Target Server Memory (KB)
- SQLServer:Memory Manager\Total Server Memory (KB)
- SQLServer:Resource Pool Stats\Active memory grant amount (KB)
- SQLServer:Resource Pool Stats\Disk Read Bytes/sec
- SQLServer:Resource Pool Stats\Disk Read IO Throttled/sec
- SQLServer:Resource Pool Stats\Disk Read IO/sec
- SQLServer:Resource Pool Stats\Disk Write Bytes/sec
- SQLServer:Resource Pool Stats\Disk Write IO Throttled/sec
- SQLServer:Resource Pool Stats\Disk Write IO/sec
- SQLServer:Resource Pool Stats\Used memory (KB)
- SQLServer:Transactions\Free Space in tempdb (KB)
- SQLServer:Transactions\Version Store Size (KB)
- SQLServer:User Settable\Query
- SQLServer:Workload Group Stats\Blocked tasks
- SQLServer:Workload Group Stats\CPU usage %
- SQLServer:Workload Group Stats\Queued requests
- SQLServer:Workload Group Stats\Requests completed/sec

Version 2 queries have the following tags:

- `sql_instance`: Physical host and instance name (hostname:instance)
- database_name: For Azure SQLDB, database_name denotes the name of the Azure SQL Database as server name is a logical construct.
