# SQL Server Extensible Plugin

This sqlserver plugin provides metrics for your SQL Server instance. 
It has been designed to parse SQL statements defined in the plugin section of your `telegraf.conf` or imported from a file.

The example below has two statements, one defined in the configuration file and one stored in an external script file, with the following parameters:

* version: the minimum SQL Server Product Version able to run defined statements:
	2008   = 10
	2008R2 = 10.50
	2012   = 11
	2014   = 12
	2016   = 13
	2017   = 14
  If the sqlserver instance version is lower, the statement will not be executed
* statement:  T-SQL statement 
* scriptfile: path to the script file (if statement is empty or not defined)
* measurement: name of the measurement
* tags: list of the columns to be defined as tags
* fields: list of the columns to be defined as fields
* fieldname: list of the columns to replace the field name, works only if one field defined (see sql_server_perf_counters example)

```
  [[inputs.sqlserver_extensible.query]]
   version="11"
   statement="select replace(rtrim(counter_name),' ','_') as counter_name, replace(rtrim(instance_name),' ','_') as instance_name, cntr_value from sys.dm_os_performance_counters where (counter_name in ('SQL Compilations/sec','SQL Re-Compilations/sec','User Connections','Batch Requests/sec','Logouts/sec','Logins/sec','Processes blocked','Latch Waits/sec','Full Scans/sec','Index Searches/sec','Page Splits/sec','Page Lookups/sec','Page Reads/sec','Page Writes/sec','Readahead Pages/sec','Lazy Writes/sec','Checkpoint Pages/sec','Database Cache Memory (KB)','Log Pool Memory (KB)','Optimizer Memory (KB)','SQL Cache Memory (KB)','Connection Memory (KB)','Lock Memory (KB)', 'Memory broker clerk size','Page life expectancy')) or (instance_name in ('_Total','Column store object pool') and counter_name in ('Transactions/sec','Write Transactions/sec','Log Flushes/sec','Log Flush Wait Time','Lock Timeouts/sec','Number of Deadlocks/sec','Lock Waits/sec','Latch Waits/sec','Memory broker clerk size','Log Bytes Flushed/sec','Bytes Sent to Replica/sec','Log Send Queue','Bytes Sent to Transport/sec','Sends to Replica/sec','Bytes Sent to Transport/sec','Sends to Transport/sec','Bytes Received from Replica/sec','Receives from Replica/sec','Flow Control Time (ms/sec)','Flow Control/sec','Resent Messages/sec','Redone Bytes/sec') or (object_name = 'SQLServer:Database Replica' and counter_name in ('Log Bytes Received/sec','Log Apply Pending Queue','Redone Bytes/sec','Recovery Queue','Log Apply Ready Queue') and instance_name = '_Total')) or (object_name = 'SQLServer:Database Replica' and counter_name in ('Transaction Delay'));"
   measurement="sql_server_perf_counters"
   tags=["server_name"]
   fields=["cntr_value"]
   fieldname=["counter_name", "instance_name"] # replace the field cntr_value

  [[inputs.sqlserver_extensible.query]]
  version="11"
  scriptfile = "/path/to/memory_clerk.sql" 
  tags=["counter_name", "server_name"]
  fields=["Buffer pool", "Cache (objects)", "Cache (sql plans)", "Other"]
```

## Example output:
```
sql_server_memory_clerk,server_name=WIN8-DEV,host=CentosInfluxDB,counter_name=Memory\ breakdown\ (%) Cache\ (sql\ plans)="2.50",Other="72.90",Buffer\ pool="24.60",Cache\ (objects)="0.00" 1501270780000000000
sql_server_memory_clerk,server_name=WIN8-DEV,host=CentosInfluxDB,counter_name=Memory\ breakdown\ (bytes) Buffer\ pool="26394624.00",Cache\ (objects)="8192.00",Cache\ (sql\ plans)="2695168.00",Other="78364672.00" 1501270780000000000

# with the counter_name field replacement
sql_server_perf_counters,server_name=WIN8-DEV,host=CentosInfluxDB Page_lookups/sec=568869i 1501270780000000000
sql_server_perf_counters,server_name=WIN8-DEV,host=CentosInfluxDB Lazy_writes/sec=0i 1501270780000000000
sql_server_perf_counters,server_name=WIN8-DEV,host=CentosInfluxDB Readahead_pages/sec=292i 1501270780000000000
sql_server_perf_counters,server_name=WIN8-DEV,host=CentosInfluxDB Page_reads/sec=3067i 1501270780000000000
sql_server_perf_counters,server_name=WIN8-DEV,host=CentosInfluxDB Page_writes/sec=17i 1501270780000000000
sql_server_perf_counters,server_name=WIN8-DEV,host=CentosInfluxDB Checkpoint_pages/sec=5i 1501270780000000000
sql_server_perf_counters,server_name=WIN8-DEV,host=CentosInfluxDB Page_life_expectancy=30356i 1501270780000000000

# with the counter_name as tags
sql_server_perf_counters,server_name=WIN8-DEV,counter_name=Page_lookups/sec,host=CentosInfluxDB cntr_value=577383i 1501271194000000000
sql_server_perf_counters,server_name=WIN8-DEV,host=CentosInfluxDB,counter_name=Lazy_writes/sec cntr_value=0i 1501271194000000000
sql_server_perf_counters,server_name=WIN8-DEV,host=CentosInfluxDB,counter_name=Readahead_pages/sec cntr_value=292i 1501271194000000000
sql_server_perf_counters,server_name=WIN8-DEV,host=CentosInfluxDB,counter_name=Page_reads/sec cntr_value=3227i 1501271194000000000
sql_server_perf_counters,server_name=WIN8-DEV,host=CentosInfluxDB,counter_name=Page_writes/sec cntr_value=17i 1501271194000000000
sql_server_perf_counters,server_name=WIN8-DEV,host=CentosInfluxDB,counter_name=Checkpoint_pages/sec cntr_value=5i 1501271194000000000
```

## External script example:

memory_clerk.sql

```SQL
SET NOCOUNT ON;
DECLARE @sqlVers numeric(4,2)
SELECT @sqlVers = LEFT(CAST(SERVERPROPERTY('productversion') as varchar), 4)

IF OBJECT_ID('tempdb..#clerk') IS NOT NULL
	DROP TABLE #clerk;
CREATE TABLE #clerk (
    ClerkCategory nvarchar(64) NOT NULL, 
    UsedPercent decimal(9,2), 
    UsedBytes bigint
);
DECLARE @DynamicClerkQuery AS nvarchar(max)
IF @sqlVers < 11
BEGIN
    SET @DynamicClerkQuery = N'
    INSERT #clerk (ClerkCategory, UsedPercent, UsedBytes)
    SELECT ClerkCategory
    , UsedPercent = SUM(UsedPercent)
    , UsedBytes = SUM(UsedBytes)
    FROM
    (
    SELECT ClerkCategory = CASE MC.[type]
        WHEN ''MEMORYCLERK_SQLBUFFERPOOL'' THEN ''Buffer pool''
        WHEN ''CACHESTORE_SQLCP'' THEN ''Cache (sql plans)''
        WHEN ''CACHESTORE_OBJCP'' THEN ''Cache (objects)''
        ELSE ''Other'' END
    , SUM((single_pages_kb + multi_pages_kb) * 1024) AS UsedBytes
    , Cast(100 * Sum((single_pages_kb + multi_pages_kb))*1.0/(Select Sum((single_pages_kb + multi_pages_kb)) From sys.dm_os_memory_clerks) as Decimal(7, 4)) UsedPercent
    FROM sys.dm_os_memory_clerks MC
    WHERE (single_pages_kb + multi_pages_kb) > 0
    GROUP BY CASE MC.[type]
        WHEN ''MEMORYCLERK_SQLBUFFERPOOL'' THEN ''Buffer pool''
        WHEN ''CACHESTORE_SQLCP'' THEN ''Cache (sql plans)''
        WHEN ''CACHESTORE_OBJCP'' THEN ''Cache (objects)''
        ELSE ''Other'' END
    ) as T
    GROUP BY ClerkCategory;
    '
END
ELSE
BEGIN
    SET @DynamicClerkQuery = N'
    INSERT #clerk (ClerkCategory, UsedPercent, UsedBytes)
    SELECT ClerkCategory
    , UsedPercent = SUM(UsedPercent)
    , UsedBytes = SUM(UsedBytes)
    FROM
    (
    SELECT ClerkCategory = CASE MC.[type]
        WHEN ''MEMORYCLERK_SQLBUFFERPOOL'' THEN ''Buffer pool''
        WHEN ''CACHESTORE_SQLCP'' THEN ''Cache (sql plans)''
        WHEN ''CACHESTORE_OBJCP'' THEN ''Cache (objects)''
        ELSE ''Other'' END
    , SUM(pages_kb * 1024) AS UsedBytes
    , Cast(100 * Sum(pages_kb)*1.0/(Select Sum(pages_kb) From sys.dm_os_memory_clerks) as Decimal(7, 4)) UsedPercent
    FROM sys.dm_os_memory_clerks MC
    WHERE pages_kb > 0
    GROUP BY CASE MC.[type]
        WHEN ''MEMORYCLERK_SQLBUFFERPOOL'' THEN ''Buffer pool''
        WHEN ''CACHESTORE_SQLCP'' THEN ''Cache (sql plans)''
        WHEN ''CACHESTORE_OBJCP'' THEN ''Cache (objects)''
        ELSE ''Other'' END
    ) as T
    GROUP BY ClerkCategory;
    '
END

EXEC sp_executesql @DynamicClerkQuery;
SELECT
-- measurement
  measurement = 'sql_server_memory_clerk'
-- tags
, server_name= REPLACE(@@SERVERNAME, '\', ':')
, counter_name
-- value
, [Buffer pool]
, [Cache (objects)]
, [Cache (sql plans)]
, [Other]
FROM
(
SELECT counter_name = 'Memory breakdown (%)'
, [Buffer pool] = ISNULL(ROUND([Buffer Pool], 1), 0)
, [Cache (objects)] = ISNULL(ROUND([Cache (objects)], 1), 0)
, [Cache (sql plans)] = ISNULL(ROUND([Cache (sql plans)], 1), 0)
, [Other] = ISNULL(ROUND([Other], 1), 0)
FROM (SELECT ClerkCategory, UsedPercent FROM #clerk) as G1
PIVOT
(
	SUM(UsedPercent)
	FOR ClerkCategory IN ([Buffer Pool], [Cache (objects)], [Cache (sql plans)], [Other])
) AS PivotTable
UNION ALL
SELECT counter_name = 'Memory breakdown (bytes)'
, [Buffer pool] = ISNULL(ROUND([Buffer Pool], 1), 0)
, [Cache (objects)] = ISNULL(ROUND([Cache (objects)], 1), 0)
, [Cache (sql plans)] = ISNULL(ROUND([Cache (sql plans)], 1), 0)
, [Other] = ISNULL(ROUND([Other], 1), 0)
FROM (SELECT ClerkCategory, UsedBytes FROM #clerk) as G2
PIVOT
(
	SUM(UsedBytes)
	FOR ClerkCategory IN ([Buffer Pool], [Cache (objects)], [Cache (sql plans)], [Other])
) AS PivotTable
) as T;
```

