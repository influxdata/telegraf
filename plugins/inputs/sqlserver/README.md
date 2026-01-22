# Microsoft SQL Server Input Plugin

This plugin provides metrics for your [SQL Server][sqlserver] instance. Recorded
metrics are lightweight and use Dynamic Management Views supplied by SQL Server.

> [!NOTE]
> This plugin supports SQL server versions supported by Microsoft (see
> [lifecycle dates][lifecycle]), Azure SQL Databases (Single), Azure SQL Managed
> Instances, Azure SQL Elastic Pools and Azure Arc-enabled SQL Managed
> Instances.

⭐ Telegraf v0.10.1
🏷️ datastore
💻 all

[sqlserver]: https://docs.microsoft.com/en-us/sql/sql-server
[lifecycle]: https://docs.microsoft.com/en-us/sql/sql-server/end-of-support/sql-server-end-of-life-overview?view=sql-server-ver15#lifecycle-dates

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

Plugins support additional global and plugin configuration settings for tasks
such as modifying metrics, tags, and fields, creating aliases, and configuring
plugin ordering. See [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Secret-store support

This plugin supports secrets from secret-stores for the `servers` option.
See the [secret-store documentation][SECRETSTORE] for more details on how
to use them.

[SECRETSTORE]: ../../../docs/CONFIGURATION.md#secret-store-secrets

## Configuration

```toml @sample.conf
# Read metrics from Microsoft SQL Server
[[inputs.sqlserver]]
  ## Specify instances to monitor with a list of connection strings.
  ## All connection parameters are optional.
  ## By default, the host is localhost, listening on default port, TCP 1433.
  ##   for Windows, the user is the currently running AD user (SSO).
  ##   See https://github.com/microsoft/go-mssqldb for detailed connection
  ##   parameters, in particular, tls connections can be created like so:
  ##   "encrypt=true;certificate=<cert>;hostNameInCertificate=<SqlServer host fqdn>"
  servers = [
    "Server=192.168.1.10;Port=1433;User Id=<user>;Password=<pw>;app name=telegraf;log=1;",
  ]

  ## Timeout for query execution operation
  ## Note that the timeout for queries is per query not per gather.
  ## 0 value means no timeout
  # query_timeout = "0s"

  ## Authentication method
  ## valid methods: "connection_string", "AAD"
  # auth_method = "connection_string"

  ## ClientID is the is the client ID of the user assigned identity of the VM
  ## that should be used to authenticate to the Azure SQL server.
  # client_id = ""

  ## "database_type" enables a specific set of queries depending on the database type. If specified, it replaces azuredb = true/false and query_version = 2
  ## In the config file, the sql server plugin section should be repeated each with a set of servers for a specific database_type.
  ## Possible values for database_type are - "SQLServer" or "AzureSQLDB" or "AzureSQLManagedInstance" or "AzureSQLPool"
  database_type = "SQLServer"

  ## A list of queries to include. If not specified, all the below listed queries are used.
  include_query = []

  ## A list of queries to explicitly ignore.
  exclude_query = ["SQLServerAvailabilityReplicaStates", "SQLServerDatabaseReplicaStates"]

  ## Force using the deprecated ADAL authentication method instead of the recommended
  ## MSAL method. Setting this option is not recommended and only exists for backward
  ## compatibility.
  # use_deprecated_adal_authentication = false

  ## Queries enabled by default for database_type = "SQLServer" are -
  ## SQLServerPerformanceCounters, SQLServerWaitStatsCategorized, SQLServerDatabaseIO, SQLServerProperties, SQLServerMemoryClerks,
  ## SQLServerSchedulers, SQLServerRequests, SQLServerVolumeSpace, SQLServerCpu, SQLServerAvailabilityReplicaStates, SQLServerDatabaseReplicaStates,
  ## SQLServerRecentBackups

  ## Queries enabled by default for database_type = "AzureSQLDB" are -
  ## AzureSQLDBResourceStats, AzureSQLDBResourceGovernance, AzureSQLDBWaitStats, AzureSQLDBDatabaseIO, AzureSQLDBServerProperties,
  ## AzureSQLDBOsWaitstats, AzureSQLDBMemoryClerks, AzureSQLDBPerformanceCounters, AzureSQLDBRequests, AzureSQLDBSchedulers

  ## Queries enabled by default for database_type = "AzureSQLManagedInstance" are -
  ## AzureSQLMIResourceStats, AzureSQLMIResourceGovernance, AzureSQLMIDatabaseIO, AzureSQLMIServerProperties, AzureSQLMIOsWaitstats,
  ## AzureSQLMIMemoryClerks, AzureSQLMIPerformanceCounters, AzureSQLMIRequests, AzureSQLMISchedulers

  ## Queries enabled by default for database_type = "AzureSQLPool" are -
  ## AzureSQLPoolResourceStats, AzureSQLPoolResourceGovernance, AzureSQLPoolDatabaseIO, AzureSQLPoolWaitStats,
  ## AzureSQLPoolMemoryClerks, AzureSQLPoolPerformanceCounters, AzureSQLPoolSchedulers

  ## Queries enabled by default for database_type = "AzureArcSQLManagedInstance" are -
  ## AzureSQLMIDatabaseIO, AzureSQLMIServerProperties, AzureSQLMIOsWaitstats,
  ## AzureSQLMIMemoryClerks, AzureSQLMIPerformanceCounters, AzureSQLMIRequests, AzureSQLMISchedulers

  ## Following are old config settings
  ## You may use them only if you are using the earlier flavor of queries, however it is recommended to use
  ## the new mechanism of identifying the database_type there by use it's corresponding queries

  ## Optional parameter, setting this to 2 will use a new version
  ## of the collection queries that break compatibility with the original
  ## dashboards.
  ## Version 2 - is compatible from SQL Server 2012 and later versions and also for SQL Azure DB
  # query_version = 2

  ## If you are using AzureDB, setting this to true will gather resource utilization metrics
  # azuredb = false

  ## Toggling this to true will emit an additional metric called "sqlserver_telegraf_health".
  ## This metric tracks the count of attempted queries and successful queries for each SQL instance specified in "servers".
  ## The purpose of this metric is to assist with identifying and diagnosing any connectivity or query issues.
  ## This setting/metric is optional and is disabled by default.
  # health_metric = false

  ## Possible queries across different versions of the collectors
  ## Queries enabled by default for specific Database Type

  ## database_type =  AzureSQLDB  by default collects the following queries
  ## - AzureSQLDBWaitStats
  ## - AzureSQLDBResourceStats
  ## - AzureSQLDBResourceGovernance
  ## - AzureSQLDBDatabaseIO
  ## - AzureSQLDBServerProperties
  ## - AzureSQLDBOsWaitstats
  ## - AzureSQLDBMemoryClerks
  ## - AzureSQLDBPerformanceCounters
  ## - AzureSQLDBRequests
  ## - AzureSQLDBSchedulers

  ## database_type =  AzureSQLManagedInstance by default collects the following queries
  ## - AzureSQLMIResourceStats
  ## - AzureSQLMIResourceGovernance
  ## - AzureSQLMIDatabaseIO
  ## - AzureSQLMIServerProperties
  ## - AzureSQLMIOsWaitstats
  ## - AzureSQLMIMemoryClerks
  ## - AzureSQLMIPerformanceCounters
  ## - AzureSQLMIRequests
  ## - AzureSQLMISchedulers

  ## database_type =  AzureSQLPool by default collects the following queries
  ## - AzureSQLPoolResourceStats
  ## - AzureSQLPoolResourceGovernance
  ## - AzureSQLPoolDatabaseIO
  ## - AzureSQLPoolOsWaitStats,
  ## - AzureSQLPoolMemoryClerks
  ## - AzureSQLPoolPerformanceCounters
  ## - AzureSQLPoolSchedulers

  ## database_type =  SQLServer by default collects the following queries
  ## - SQLServerPerformanceCounters
  ## - SQLServerWaitStatsCategorized
  ## - SQLServerDatabaseIO
  ## - SQLServerProperties
  ## - SQLServerMemoryClerks
  ## - SQLServerSchedulers
  ## - SQLServerRequests
  ## - SQLServerVolumeSpace
  ## - SQLServerCpu
  ## - SQLServerRecentBackups
  ## and following as optional (if mentioned in the include_query list)
  ## - SQLServerAvailabilityReplicaStates
  ## - SQLServerDatabaseReplicaStates

  ## Maximum number of open connections to the database, 0 allows the driver to decide.
  # max_open_connections = 0

  ## Maximum number of idle connections in the connection pool, 0 allows the driver to decide.
  # max_idle_connections = 0
```

For available options in the `servers` DSN check the [driver documentation][driver].

The plugin supports the named-pipe and LPC protocol **on Windows AMD64 and i386**
for connections. On other platforms those protocols are not available. See the
[protocol configuration section][driver_protos] of the driver documentation on
how to specify the protocols.

[driver]: https://github.com/microsoft/go-mssqldb
[driver_protos]: https://github.com/microsoft/go-mssqldb?tab=readme-ov-file#protocol-configuration

### Additional Setup

You have to create a login on every SQL Server instance or Azure SQL Managed
instance you want to monitor, with following script:

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

For Azure SQL Database, you require the View Database State permission
and can create a user with a password directly in the database.

```sql
CREATE USER [telegraf] WITH PASSWORD = N'mystrongpassword';
GO
GRANT VIEW DATABASE STATE TO [telegraf];
GO
```

For Azure SQL Elastic Pool, please follow the following instructions to collect
metrics. On master logical database, create an SQL login 'telegraf' and assign
it to the server-level role ##MS_ServerStateReader##.

```sql
CREATE LOGIN [telegraf] WITH PASSWORD = N'mystrongpassword';
GO
ALTER SERVER ROLE ##MS_ServerStateReader##
  ADD MEMBER [telegraf];
GO
```

Elastic pool metrics can be collected from any database in the pool if a user
for the `telegraf` login is created in that database. For collection to work,
this database must remain in the pool, and must not be renamed. If you plan
to add/remove databases from this pool, create a separate database for
monitoring purposes that will remain in the pool.

> [!NOTE]
> To avoid duplicate monitoring data, do not collect elastic pool metrics
> from more than one database in the same pool.

```sql
GO
CREATE USER [telegraf] FOR LOGIN telegraf;
```

For Service SID authentication to SQL Server (Windows service installations
only) check the [howto document][sid_howto]. In an administrative command prompt
configure the telegraf service for use with a service SID

```Batchfile
sc.exe sidtype "telegraf" unrestricted
```

To create the login for the telegraf service run the following script:

```sql
USE master;
GO
CREATE LOGIN [NT SERVICE\telegraf] FROM WINDOWS;
GO
GRANT VIEW SERVER STATE TO [NT SERVICE\telegraf];
GO
GRANT VIEW ANY DEFINITION TO [NT SERVICE\telegraf];
GO
```

Remove User Id and Password keywords from the connection string in your
config file to use windows authentication.

```toml
[[inputs.sqlserver]]
  servers = ["Server=192.168.1.10;Port=1433;app name=telegraf;log=1;",]
```

To set up a configurable timeout, add timeout to the connections string
in your config file.

```toml
servers = [
  "Server=192.168.1.10;Port=1433;User Id=<user>;Password=<pw>;app name=telegraf;log=1;dial timeout=30",
]
```

[sid_howto]: https://docs.microsoft.com/en-us/sql/relational-databases/security/using-service-sids-to-grant-permissions-to-services-in-sql-server

### Azure Active Directory (AAD) authentication using Managed Identity

Azure SQL Database instances support two main methods of
[authentication][auth_methods]: SQL authentication and AAD authentication. The
recommended practice is to use [AAD authentication][auth_aad] when possible as
it is a more modern authentication protocol, allows for easier credential and
role management and can eliminate the need to include passwords in connection
strings.

If more then one managed identity is assigned to the VM, you need specify the
`client_id` of the identity you wish to use to authenticate with the SQL Server.
Please check [SQL Server driver][driver_azure] documentation for available options.

[auth_methods]: https://docs.microsoft.com/en-us/azure/azure-sql/database/security-overview#authentication
[auth_aad]: https://docs.microsoft.com/en-us/azure/azure-sql/database/authentication-aad-overview
[driver_azure]: https://github.com/microsoft/go-mssqldb#azure-active-directory-authentication

### Azure Active Directory (AAD) authentication using MSI

AAD based auth is currently only supported for Azure SQL Database and Azure SQL
Managed Instance but not for SQL Server. To use MSI configure
"system-assigned managed identity" for Azure resources on the Monitoring VM
(the VM connecting to the SQL server/database) [using the Azure portal][portal].
Create a user with the name of the Monitoring VM as the principal on the
database being monitored using the below script. This might require
allow-listing the client machine's IP address (from where the below SQL script
is being run) on the SQL Server resource.

In case of multiple assigned identities on one VM you can use the parameter
user_assigned_id to specify the `client_id`.

```sql
EXECUTE ('IF EXISTS(SELECT * FROM sys.database_principals WHERE name = ''<Monitoring_VM_Name>'')
    BEGIN
        DROP USER [<Monitoring_VM_Name>]
    END')
EXECUTE ('CREATE USER [<Monitoring_VM_Name>] FROM EXTERNAL PROVIDER')
EXECUTE ('GRANT VIEW DATABASE STATE TO [<Monitoring_VM_Name>]')
```

On the SQL Server resource of the database(s) being monitored, go to
"Firewalls and Virtual Networks" tab and allowlist the monitoring VM IP address.
On the Monitoring VM, update the telegraf config file with the database
connection string in the following format. The connection string only provides
the server and database name, but no password (since the VM's system-assigned
managed identity would be used for authentication). The auth method must be
set to "AAD"

```toml
  servers = [
    "Server=<Azure_SQL_Server_Name>.database.windows.net;Port=1433;Database=<Azure_SQL_Database_Name>;app name=telegraf;log=1;",
  ]
  auth_method = "AAD"
```

[portal]: https://docs.microsoft.com/en-us/azure/active-directory/managed-identities-azure-resources/qs-configure-portal-windows-vm

## Metrics

To provide backwards compatibility, this plugin support two versions of
metrics queries.

> [!NOTE]
> Version 2 queries are not backwards compatible with the old queries. Any
> dashboards or queries based on the old query format will not work with the new
> format. The version 2 queries only report raw metrics, no math has been done
> to calculate deltas. To graph this data you must calculate deltas in your
> dashboarding software.

### Query Version 1

> [!CAUTION]
> The `query_version` option was **deprecated** in Telegraf v1.16. All future
> development will be under configuration option`database_type`.

The original metrics queries provide:

- *Performance counters*: 1000+ metrics from `sys.dm_os_performance_counters`
- *Performance metrics*: special performance and ratio metrics
- *Wait stats*: wait tasks categorized from `sys.dm_os_wait_stats`
- *Memory clerk*: memory breakdown from `sys.dm_os_memory_clerks`
- *Database size*: databases size trend from `sys.dm_io_virtual_file_stats`
- *Database IO*: databases I/O from `sys.dm_io_virtual_file_stats`
- *Database latency*: databases latency from `sys.dm_io_virtual_file_stats`
- *Database properties*: databases properties, state and recovery model, from `sys.databases`
- *OS Volume*: available, used and total space from `sys.dm_os_volume_stats`
- *CPU*: cpu usage from `sys.dm_os_ring_buffers`

If you are using the original queries all stats have the following tags:

- `servername`:  hostname:instance
- `type`: type of stats to easily filter measurements

### Query Version 2

> [!CAUTION]
> The `query_version` option was **deprecated** in Telegraf v1.16. All future
> development will be under configuration option`database_type`.

The new (version 2) metrics provide:

- *Database IO*: IO stats from `sys.dm_io_virtual_file_stats`.
- *Memory Clerk*: Memory clerk breakdown from `sys.dm_os_memory_clerks`, most
                  clerks have been given a friendly name.
- *Performance Counters*:  A select list of performance counters from
                           `sys.dm_os_performance_counters`. Some of the
                           important metrics included:
  - *Activity*:            Transactions/sec/database, Batch requests/sec,
                           blocked processes and more
  - *Availability Groups*: Bytes sent to replica, Bytes received from replica,
                           Log bytes received, Log send queue, transaction delay
                           and more
  - *Log activity*:        Log bytes flushed/sec, Log flushes/sec, Log Flush
                           Wait Time
  - *Memory*:              PLE, Page reads/sec, Page writes/sec and more
  - *TempDB*:              Free space, Version store usage, Active temp tables,
                           temp table creation rate and more
  - *Resource Governor*:   CPU Usage, Requests/sec, Queued Requests, and Blocked
                           tasks per workload group and more
- *Server properties*:     Number of databases in all possible states (online,
                           offline, suspect, etc.), cpu count, total physical
                           memory, available physical memory, SQL Server service
                           uptime, SQL Server SPID, and SQL Server version. In
                           the case of Azure SQL relevant properties such as
                           Tier, #Vcores, Memory etc.
- *Wait stats*:            Wait time in ms, number of waiting tasks, resource
                           wait time, signal wait time, max wait time in ms,
                           wait type, and wait category. The waits are
                           categorized using the same categories used in
                           Query Store.
- *Schedulers*:            This captures `sys.dm_os_schedulers`.
- *SqlRequests*:           This captures a snapshot of `sys.dm_exec_requests`
                           and `sys.dm_exec_sessions` that gives you running
                           requests as well as wait types and blocking sessions.
                           Telegraf's monitoring request is omitted unless it is
                           a heading blocker. Also includes sleeping sessions
                           with open transactions.
- *VolumeSpace*:           uses `sys.dm_os_volume_stats` to get total, used and
                           occupied space on every disk that contains a data or
                           log file. (Note that even if enabled it won't get any
                           data from Azure SQL Database or SQL Managed Instance).
                           It is pointless to run this with high frequency
                           (ie: every 10s), but it won't cause any problem.
- *Cpu*:                   uses the buffer ring (`sys.dm_os_ring_buffers`) to
                           get CPU data, the table is updated once per minute.
                           (Note that even if enabled it won't get any data from
                           Azure SQL Database or SQL Managed Instance).

  In order to allow tracking on a per statement basis this query produces a
  unique tag for each query.  Depending on the database workload, this may
  result in a high cardinality series.  Reference the FAQ for tips on
  [managing series cardinality][cardinality].

- *Azure Managed Instances*
  - Stats from `sys.server_resource_stats`
  - Resource governance stats from `sys.dm_instance_resource_governance`
- *Azure SQL Database* in addition to other stats
  - Stats from `sys.dm_db_wait_stats`
  - Resource governance stats from `sys.dm_user_db_resource_governance`
  - Stats from `sys.dm_db_resource_stats`

### Database Type "AzureSQLDB"

These are metrics for Azure SQL Database (single database) and are very
similar to version 2 but split out for maintenance reasons, better ability
to test,differences in DMVs:

- *AzureSQLDBDatabaseIO*:          IO stats from `sys.dm_io_virtual_file_stats`
                                   including resource governance time, RBPEX, IO
                                   for Hyperscale.
- *AzureSQLDBMemoryClerks*:        Memory clerk breakdown from `sys.dm_os_memory_clerks`.
- *AzureSQLDBResourceGovernance*:  Relevant properties indicatign resource
                                   limits from `sys.dm_user_db_resource_governance`
- *AzureSQLDBPerformanceCounters*: A select list of performance counters from
                                   `sys.dm_os_performance_counters` including
                                   cloud specific counters for SQL Hyperscale.
- *AzureSQLDBServerProperties*:    Relevant Azure SQL relevant properties from
                                   such as Tier, #Vcores, Memory etc, storage, etc.
- *AzureSQLDBWaitstats*:           Wait time in ms from `sys.dm_db_wait_stats`,
                                   number of waiting tasks, resource wait time,
                                   signal wait time, max wait time in ms, wait
                                   type, and wait category. The waits are
                                   categorized using the same categories used in
                                   Query Store. These waits are collected only
                                   as of the end of the a statement. and for a
                                   specific database only.
- *AzureSQLOsWaitstats*:           Wait time in ms from `sys.dm_os_wait_stats`,
                                   number of waiting tasks, resource wait time,
                                   signal wait time, max wait time in ms, wait
                                   type, and wait category. The waits are
                                   categorized using the same categories used in
                                   Query Store. These waits are collected as
                                   they occur and instance wide
- *AzureSQLDBRequests*:            Requests which are blocked or have a wait
                                   type from `sys.dm_exec_sessions` and
                                   `sys.dm_exec_requests`. Telegraf's monitoring
                                   request is omitted unless it is a heading blocker
- *AzureSQLDBSchedulers*:          This captures `sys.dm_os_schedulers` snapshots.

### Database Type "AzureSQLManagedInstance"

These are metrics for Azure SQL Managed instance, are very similar to version
2 but split out for maintenance reasons, better ability to test, differences
in DMVs:

- *AzureSQLMIDatabaseIO*:          IO stats from `sys.dm_io_virtual_file_stats`
                                   including resource governance time, RBPEX, IO
                                   for Hyperscale.
- *AzureSQLMIMemoryClerks*:        Memory clerk breakdown from `sys.dm_os_memory_clerks`.
- *AzureSQLMIResourceGovernance*:  Relevant properties indicatign resource limits
                                   from `sys.dm_instance_resource_governance`
- *AzureSQLMIPerformanceCounters*: A select list of performance counters from
                                  `sys.dm_os_performance_counters` including
                                   cloud specific counters for SQL Hyperscale.
- *AzureSQLMIServerProperties*:    Relevant Azure SQL relevant properties such
                                   as Tier, #Vcores, Memory etc, storage, etc.
- *AzureSQLMIOsWaitstats*:         Wait time in ms from `sys.dm_os_wait_stats`,
                                   number of waiting tasks, resource wait time,
                                   signal wait time, max wait time in ms, wait
                                   type, and wait category. The waits are
                                   categorized using the same categories used in
                                   Query Store. These waits are collected as
                                   they occur and instance wide
- *AzureSQLMIRequests*:            Requests which are blocked or have a wait
                                   type from `sys.dm_exec_sessions` and
                                   `sys.dm_exec_requests`. Telegraf's monitoring
                                   request is omitted unless it is a heading blocker
- *AzureSQLMISchedulers*: This captures `sys.dm_os_schedulers` snapshots.

### Database Type "AzureSQLPool"

These are metrics for Azure SQL to monitor resources usage at Elastic Pool
level. These metrics require additional permissions to be collected, please
ensure to check additional setup section in this documentation.

- *AzureSQLPoolResourceStats*:       Returns resource usage statistics for the
                                     current elastic pool in a SQL Database
                                     server. Queried from `sys.dm_resource_governor_resource_pools_history_ex`.
- *AzureSQLPoolResourceGovernance*:  Returns actual configuration and capacity
                                     settings used by resource governance
                                     mechanisms in the current elastic pool.
                                     Queried from `sys.dm_user_db_resource_governance`.
- *AzureSQLPoolDatabaseIO*:          Returns I/O statistics for data and log
                                     files for each database in the pool.
                                     Queried from `sys.dm_io_virtual_file_stats`.
- *AzureSQLPoolOsWaitStats*:         Returns information about all the waits
                                     encountered by threads that executed.
                                     Queried from `sys.dm_os_wait_stats`.
- *AzureSQLPoolMemoryClerks*:        Memory clerk breakdown from `sys.dm_os_memory_clerks`.
- *AzureSQLPoolPerformanceCounters*: A selected list of performance counters
                                     from `sys.dm_os_performance_counters`.
                                     Note: Performance counters where the
                                     cntr_type column value is 537003264 are
                                     already returned with a percentage format
                                     between 0 and 100. For other counters,
                                     please check [sys.dm_os_performance_counters][os_perf_counters]
                                     documentation.
- *AzureSQLPoolSchedulers*: This captures `sys.dm_os_schedulers` snapshots.

[os_perf_counters]: https://docs.microsoft.com/en-us/sql/relational-databases/system-dynamic-management-views/sys-dm-os-performance-counters-transact-sql?view=azuresqldb-current

### Database Type "SQLServer"

- *SQLServerDatabaseIO*:           IO stats from `sys.dm_io_virtual_file_stats`
- *SQLServerMemoryClerks*:         Memory clerk breakdown from `sys.dm_os_memory_clerks`,
                                   most clerks have been given a friendly name.
- *SQLServerPerformanceCounters*:  A select list of performance counters from
                                   `sys.dm_os_performance_counters`. Some of the
                                   important metrics included:
  - *Activity*:                    Transactions/sec/database, Batch requests/sec,
                                   blocked processes and more
  - *Availability Groups*:         Bytes sent to replica, Bytes received from
                                   replica, Log bytes received, Log send queue,
                                   transaction delay and more
  - *Log activity*:                Log bytes flushed/sec, Log flushes/sec,
                                   Log Flush Wait Time
  - *Memory*:                      PLE, Page reads/sec, Page writes/sec and more
  - *TempDB*:                      Free space, Version store usage, Active temp
                                   tables, temp table creation rate and more
  - *Resource Governor*:           CPU Usage, Requests/sec, Queued Requests, and
                                   Blocked tasks per workload group and more
- *SQLServerProperties*:           Number of databases in all possible states
                                   (online, offline, suspect, etc.), cpu count,
                                   total physical memory, available physical memory,
                                   SQL Server service uptime, SQL Server SPID and
                                   SQL Server version. In the case of Azure SQL
                                   relevant properties such as Tier, #Vcores,
                                   Memory etc.
- *SQLServerWaitStatsCategorized*: Wait time in ms, number of waiting tasks,
                                   resource wait time, signal wait time, max
                                   wait time in ms, wait type, and wait category.
                                   The waits are categorized using the same
                                   categories used in Query Store.
- *SQLServerSchedulers*:           This captures `sys.dm_os_schedulers`.
- *SQLServerRequests*:             This captures a snapshot of `sys.dm_exec_requests`
                                   and `sys.dm_exec_sessions` that gives you
                                   running requests as well as wait types and
                                   blocking sessions.
- *SQLServerVolumeSpace*:          Uses `sys.dm_os_volume_stats` to get total,
                                   used and occupied space on every disk that
                                   contains a data or log file. (Note that even
                                   if enabled it won't get any data from Azure
                                   SQL Database or SQL Managed Instance). It is
                                   pointless to run this with high frequency
                                   (ie: every 10s), but it won't cause any problem.
- SQLServerCpu:                    Uses the buffer ring (`sys.dm_os_ring_buffers`)
                                   to get CPU data, the table is updated once
                                   per minute. (Note that even if enabled it
                                   won't get any data from Azure SQL Database or
                                   SQL Managed Instance).
- SQLServerAvailabilityReplicaStates: Collects availability replica state
                                      information from `sys.dm_hadr_availability_replica_states`
                                      for a High Availability / Disaster
                                      Recovery (HADR) setup
- SQLServerDatabaseReplicaStates:  Collects database replica state information
                                   from `sys.dm_hadr_database_replica_states`
                                   for a High Availability / Disaster Recovery
                                   (HADR) setup
- SQLServerRecentBackups:          Collects latest full, differential and
                                   transaction log backup date and size from `msdb.dbo.backupset`
- SQLServerPersistentVersionStore: Collects persistent version store information
                                   from `sys.dm_tran_persistent_version_store_stats`
                                   for databases with Accelerated Database
                                   Recovery enabled

### Output Measures

The guiding principal is that all data collected from the same primary DMV ends
up in the same measure irrespective of database_type.

- `sqlserver_database_io`       - Used by AzureSQLDBDatabaseIO,
                                  AzureSQLMIDatabaseIO, SQLServerDatabaseIO,
                                  DatabaseIO given the data is from `sys.dm_io_virtual_file_stats`
- `sqlserver_waitstats`         - Used by WaitStatsCategorized,
                                  AzureSQLDBOsWaitstats, AzureSQLMIOsWaitstats
- `sqlserver_server_properties` - Used by SQLServerProperties,
                                  AzureSQLDBServerProperties,
                                  AzureSQLMIServerProperties, ServerProperties
- `sqlserver_memory_clerks`     - Used by SQLServerMemoryClerks,
                                  AzureSQLDBMemoryClerks, AzureSQLMIMemoryClerks,
                                  MemoryClerk
- `sqlserver_performance`       - Used by SQLServerPerformanceCounters,
                                  AzureSQLDBPerformanceCounters,
                                  AzureSQLMIPerformanceCounters, PerformanceCounters
- `sys.dm_os_schedulers`        - Used by SQLServerSchedulers,
                                  AzureSQLDBServerSchedulers, AzureSQLMIServerSchedulers

The following Performance counter metrics can be used directly, with no delta
calculations:

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

- `sql_instance`:  Physical host and instance name (hostname:instance)
- `database_name`: For Azure SQLDB, database_name denotes the name of the
                   Azure SQL Database as server name is a logical construct.

### Health Metric

All collection versions (version 1, version 2, and database_type) support an
optional plugin health metric called `sqlserver_telegraf_health`. This metric
tracks if connections to SQL Server are succeeding or failing. Users can
leverage this metric to detect if their SQL Server monitoring is not working
as intended.

In the configuration file, toggling `health_metric` to `true` will enable
collection of this metric. By default, this value is set to `false` and
the metric is not collected. The health metric emits one record for each
connection specified by `servers` in the configuration file.

The health metric emits the following tags:

- `sql_instance` -   Name of the server specified in the connection string. This
                     value is emitted as-is in the connection string. If the
                     server could not be parsed from the connection string, a
                     constant placeholder value is emitted
- `database_name` -  Name of the database or (initial catalog) specified in the
                     connection string. This value is emitted as-is in the
                     connection string. If the database could not be parsed from
                     the connection string, a constant placeholder value is
                     emitted

The health metric emits the following fields:

- `attempted_queries`  - Number of queries that were attempted for this connection
- `successful_queries` - Number of queries that completed successfully for this connection
- `database_type`      - Type of database as specified by `database_type`.
                         If `database_type` is empty, the `QueryVersion` and
                         `AzureDB` fields are concatenated instead

If `attempted_queries` and `successful_queries` are not equal for
a given connection, some metrics were not successfully gathered for
that connection. If `successful_queries` is 0, no metrics were successfully
gathered.

[cardinality]: /docs/FAQ.md#user-content-q-how-can-i-manage-series-cardinality

## Example Output

```text
sqlserver_cpu_other_process_cpu{host="servername",measurement_db_type="SQLServer",sql_instance="SERVERNAME:INST"} 9
sqlserver_performance{counter="Log File(s) Size (KB)",counter_type="65792",host="servername",instance="instance_name",measurement_db_type="SQLServer",object="MSSQL$INSTANCE_NAME:Databases",sql_instance="SERVERNAME:INSTANCE_NAME"} 1.048568e+06
```
