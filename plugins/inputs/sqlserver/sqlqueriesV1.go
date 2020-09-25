package sqlserver

import (
	_ "github.com/denisenkom/go-mssqldb" // go-mssqldb initialization
)

// Queries V1
const sqlPerformanceMetrics string = `SET DEADLOCK_PRIORITY -10;
SET NOCOUNT ON;
SET ARITHABORT ON;
SET QUOTED_IDENTIFIER ON;
SET TRANSACTION ISOLATION LEVEL READ UNCOMMITTED

DECLARE @PCounters TABLE
(
	counter_name nvarchar(64),
	cntr_value bigint,
	Primary Key(counter_name)
);

INSERT @PCounters (counter_name, cntr_value)
SELECT 'Point In Time Recovery', Value = CASE
	WHEN  1 > 1.0 * COUNT(*)  / NULLIF((SELECT COUNT(*) FROM sys.databases d WHERE database_id > 4), 0)
	THEN 0 ELSE 1 END
FROM sys.databases d
WHERE database_id > 4
	AND recovery_model IN (1)
UNION ALL
SELECT 'Page File Usage (%)', CAST(100 * (1 - available_page_file_kb * 1. / total_page_file_kb) as decimal(9,2)) as [PageFileUsage (%)]
FROM sys.dm_os_sys_memory
UNION ALL
SELECT 'Connection memory per connection (bytes)',  Ratio = CAST((cntr_value / (SELECT 1.0 * cntr_value FROM sys.dm_os_performance_counters WHERE counter_name = 'User Connections')) * 1024 as int)
FROM sys.dm_os_performance_counters
WHERE counter_name = 'Connection Memory (KB)'
UNION ALL
SELECT 'Available physical memory (bytes)', available_physical_memory_kb * 1024
FROM sys.dm_os_sys_memory
UNION ALL
SELECT 'Signal wait (%)', SignalWaitPercent = CAST(100.0 * SUM(signal_wait_time_ms) / SUM (wait_time_ms) AS NUMERIC(20,2))
FROM sys.dm_os_wait_stats
UNION ALL
SELECT 'Sql compilation per batch request',  SqlCompilationPercent = 100.0 * cntr_value / (SELECT 1.0*cntr_value FROM sys.dm_os_performance_counters WHERE counter_name = 'Batch Requests/sec')
FROM sys.dm_os_performance_counters
WHERE counter_name = 'SQL Compilations/sec'
UNION ALL
SELECT 'Sql recompilation per batch request', SqlReCompilationPercent = 100.0 *cntr_value / (SELECT 1.0*cntr_value FROM sys.dm_os_performance_counters WHERE counter_name = 'Batch Requests/sec')
FROM sys.dm_os_performance_counters
WHERE counter_name = 'SQL Re-Compilations/sec'
UNION ALL
SELECT 'Page lookup per batch request',PageLookupPercent = 100.0 * cntr_value / (SELECT 1.0*cntr_value FROM sys.dm_os_performance_counters WHERE counter_name = 'Batch Requests/sec')
FROM sys.dm_os_performance_counters
WHERE counter_name = 'Page lookups/sec'
UNION ALL
SELECT 'Page split per batch request',PageSplitPercent = 100.0 * cntr_value / (SELECT 1.0*cntr_value FROM sys.dm_os_performance_counters WHERE counter_name = 'Batch Requests/sec')
FROM sys.dm_os_performance_counters
WHERE counter_name = 'Page splits/sec'
UNION ALL
SELECT 'Average tasks', AverageTaskCount = (SELECT AVG(current_tasks_count) FROM sys.dm_os_schedulers WITH (NOLOCK) WHERE scheduler_id < 255 )
UNION ALL
SELECT 'Average runnable tasks', AverageRunnableTaskCount = (SELECT AVG(runnable_tasks_count) FROM sys.dm_os_schedulers WITH (NOLOCK) WHERE scheduler_id < 255 )
UNION ALL
SELECT 'Average pending disk IO', AveragePendingDiskIOCount = (SELECT AVG(pending_disk_io_count) FROM sys.dm_os_schedulers WITH (NOLOCK) WHERE scheduler_id < 255 )
UNION ALL
SELECT 'Buffer pool rate (bytes/sec)', BufferPoolRate = (1.0*cntr_value * 8 * 1024) /
	(SELECT 1.0*cntr_value FROM sys.dm_os_performance_counters  WHERE object_name like '%Buffer Manager%' AND counter_name = 'Page life expectancy')
FROM sys.dm_os_performance_counters
WHERE object_name like '%Buffer Manager%'
AND counter_name = 'Database pages'
UNION ALL
SELECT 'Memory grant pending', MemoryGrantPending = cntr_value
FROM sys.dm_os_performance_counters
WHERE counter_name = 'Memory Grants Pending'
UNION ALL
SELECT 'Readahead per page read', Readahead = 100.0 *cntr_value / (SELECT 1.0*cntr_value FROM sys.dm_os_performance_counters WHERE counter_name = 'Page Reads/sec')
FROM sys.dm_os_performance_counters
WHERE counter_name = 'Readahead pages/sec'
UNION ALL
SELECT 'Total target memory ratio', TotalTargetMemoryRatio = 100.0 * cntr_value / (SELECT 1.0*cntr_value FROM sys.dm_os_performance_counters WHERE counter_name = 'Target Server Memory (KB)')
FROM sys.dm_os_performance_counters
WHERE counter_name = 'Total Server Memory (KB)'

IF OBJECT_ID('tempdb..#PCounters') IS NOT NULL DROP TABLE #PCounters;
SELECT * INTO #PCounters FROM @PCounters

DECLARE @DynamicPivotQuery AS NVARCHAR(MAX)
DECLARE @ColumnName AS NVARCHAR(MAX)
SELECT @ColumnName= ISNULL(@ColumnName + ',','') + QUOTENAME(counter_name)
FROM (SELECT DISTINCT counter_name FROM @PCounters) AS bl

SET @DynamicPivotQuery = N'
SELECT measurement = ''Performance metrics'', servername = REPLACE(@@SERVERNAME, ''\'', '':''), type = ''Performance metrics''
, ' + @ColumnName + '  FROM
(
SELECT counter_name, cntr_value
FROM #PCounters
) as V
PIVOT(SUM(cntr_value) FOR counter_name IN (' + @ColumnName + ')) AS PVTTable
'
EXEC sp_executesql @DynamicPivotQuery;
`

const sqlMemoryClerk string = `SET DEADLOCK_PRIORITY -10;
SET NOCOUNT ON;
SET TRANSACTION ISOLATION LEVEL READ UNCOMMITTED;

DECLARE @sqlVers numeric(4,2)
SELECT @sqlVers = LEFT(CAST(SERVERPROPERTY('productversion') as varchar), 4)

IF OBJECT_ID('tempdb..#clerk') IS NOT NULL
	DROP TABLE #clerk;

CREATE TABLE #clerk (
    ClerkCategory nvarchar(64) NOT NULL,
    UsedPercent decimal(9,2),
    UsedBytes bigint
);

DECLARE @DynamicClerkQuery AS NVARCHAR(MAX)

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
measurement
-- tags
, servername= REPLACE(@@SERVERNAME, '\', ':')
, type = 'Memory clerk'
-- value
, [Buffer pool]
, [Cache (objects)]
, [Cache (sql plans)]
, [Other]
FROM
(
SELECT measurement = 'Memory breakdown (%)'
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

SELECT measurement = 'Memory breakdown (bytes)'
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
`

const sqlDatabaseSize string = `SET DEADLOCK_PRIORITY -10;
SET NOCOUNT ON;
SET TRANSACTION ISOLATION LEVEL READ UNCOMMITTED

IF OBJECT_ID('tempdb..#baseline') IS NOT NULL
	DROP TABLE #baseline;
SELECT
    DB_NAME(mf.database_id) AS database_name ,
    CAST(mf.size AS BIGINT) as database_size_8k_pages,
    CAST(mf.max_size AS BIGINT) as database_max_size_8k_pages,
    size_on_disk_bytes ,
	type_desc as datafile_type,
    GETDATE() AS baselineDate
INTO #baseline
FROM sys.dm_io_virtual_file_stats(NULL, NULL) AS divfs
INNER JOIN sys.master_files AS mf ON mf.database_id = divfs.database_id
	AND mf.file_id = divfs.file_id

DECLARE @DynamicPivotQuery AS NVARCHAR(MAX)
DECLARE @ColumnName AS NVARCHAR(MAX), @ColumnName2 AS NVARCHAR(MAX)

SELECT @ColumnName= ISNULL(@ColumnName + ',','') + QUOTENAME(database_name)
FROM (SELECT DISTINCT database_name FROM #baseline) AS bl

--Prepare the PIVOT query using the dynamic
SET @DynamicPivotQuery = N'
SELECT measurement = ''Log size (bytes)'', servername = REPLACE(@@SERVERNAME, ''\'', '':''), type = ''Database size''
, ' + @ColumnName + '  FROM
(
SELECT database_name, size_on_disk_bytes
FROM #baseline
WHERE datafile_type = ''LOG''
) as V
PIVOT(SUM(size_on_disk_bytes) FOR database_name IN (' + @ColumnName + ')) AS PVTTable

UNION ALL

SELECT measurement = ''Rows size (bytes)'', servername = REPLACE(@@SERVERNAME, ''\'', '':''), type = ''Database size''
, ' + @ColumnName + '  FROM
(
SELECT database_name, size_on_disk_bytes
FROM #baseline
WHERE datafile_type = ''ROWS''
) as V
PIVOT(SUM(size_on_disk_bytes) FOR database_name IN (' + @ColumnName + ')) AS PVTTable

UNION ALL

SELECT measurement = ''Rows size (8KB pages)'', servername = REPLACE(@@SERVERNAME, ''\'', '':''), type = ''Database size''
, ' + @ColumnName + '  FROM
(
SELECT database_name, database_size_8k_pages
FROM #baseline
WHERE datafile_type = ''ROWS''
) as V
PIVOT(SUM(database_size_8k_pages) FOR database_name IN (' + @ColumnName + ')) AS PVTTable

UNION ALL

SELECT measurement = ''Log size (8KB pages)'', servername = REPLACE(@@SERVERNAME, ''\'', '':''), type = ''Database size''
, ' + @ColumnName + '  FROM
(
SELECT database_name, database_size_8k_pages
FROM #baseline
WHERE datafile_type = ''LOG''
) as V
PIVOT(SUM(database_size_8k_pages) FOR database_name IN (' + @ColumnName + ')) AS PVTTable

UNION ALL

SELECT measurement = ''Rows max size (8KB pages)'', servername = REPLACE(@@SERVERNAME, ''\'', '':''), type = ''Database size''
, ' + @ColumnName + '  FROM
(
SELECT database_name, database_max_size_8k_pages
FROM #baseline
WHERE datafile_type = ''ROWS''
) as V
PIVOT(SUM(database_max_size_8k_pages) FOR database_name IN (' + @ColumnName + ')) AS PVTTable

UNION ALL

SELECT measurement = ''Logs max size (8KB pages)'', servername = REPLACE(@@SERVERNAME, ''\'', '':''), type = ''Database size''
, ' + @ColumnName + '  FROM
(
SELECT database_name, database_max_size_8k_pages
FROM #baseline
WHERE datafile_type = ''LOG''
) as V
PIVOT(SUM(database_max_size_8k_pages) FOR database_name IN (' + @ColumnName + ')) AS PVTTable
'
--PRINT @DynamicPivotQuery
EXEC sp_executesql @DynamicPivotQuery;
`

const sqlDatabaseStats string = `SET DEADLOCK_PRIORITY -10;
SET NOCOUNT ON;
SET TRANSACTION ISOLATION LEVEL READ UNCOMMITTED;

IF OBJECT_ID('tempdb..#baseline') IS NOT NULL
	DROP TABLE #baseline;

SELECT
[ReadLatency] =
    CASE WHEN [num_of_reads] = 0
        THEN 0 ELSE ([io_stall_read_ms] / [num_of_reads]) END,
[WriteLatency] =
    CASE WHEN [num_of_writes] = 0
        THEN 0 ELSE ([io_stall_write_ms] / [num_of_writes]) END,
[Latency] =
    CASE WHEN ([num_of_reads] = 0 AND [num_of_writes] = 0)
        THEN 0 ELSE ([io_stall] / ([num_of_reads] + [num_of_writes])) END,
[AvgBytesPerRead] =
    CASE WHEN [num_of_reads] = 0
        THEN 0 ELSE ([num_of_bytes_read] / [num_of_reads]) END,
[AvgBytesPerWrite] =
    CASE WHEN [num_of_writes] = 0
        THEN 0 ELSE ([num_of_bytes_written] / [num_of_writes]) END,
[AvgBytesPerTransfer] =
    CASE WHEN ([num_of_reads] = 0 AND [num_of_writes] = 0)
        THEN 0 ELSE
            (([num_of_bytes_read] + [num_of_bytes_written]) /
            ([num_of_reads] + [num_of_writes])) END,
DB_NAME ([vfs].[database_id]) AS DatabaseName,
[mf].type_desc  as datafile_type
INTO #baseline
FROM sys.dm_io_virtual_file_stats (NULL,NULL) AS [vfs]
JOIN sys.master_files AS [mf] ON [vfs].[database_id] = [mf].[database_id]
    AND [vfs].[file_id] = [mf].[file_id]



DECLARE @DynamicPivotQuery AS NVARCHAR(MAX)
DECLARE @ColumnName AS NVARCHAR(MAX), @ColumnName2 AS NVARCHAR(MAX)

SELECT @ColumnName= ISNULL(@ColumnName + ',','') + QUOTENAME(DatabaseName)
FROM (SELECT DISTINCT DatabaseName FROM #baseline) AS bl

--Prepare the PIVOT query using the dynamic
SET @DynamicPivotQuery = N'
SELECT measurement = ''Log read latency (ms)'', servername = REPLACE(@@SERVERNAME, ''\'', '':''), type = ''Database stats''
, ' + @ColumnName + '  FROM
(
SELECT DatabaseName,  ReadLatency
FROM #baseline
WHERE datafile_type = ''LOG''
) as V
PIVOT(MAX(ReadLatency) FOR DatabaseName IN (' + @ColumnName + ')) AS PVTTable

UNION ALL

SELECT measurement = ''Log write latency (ms)'', servername = REPLACE(@@SERVERNAME, ''\'', '':''), type = ''Database stats''
, ' + @ColumnName + '  FROM
(
SELECT DatabaseName,  WriteLatency
FROM #baseline
WHERE datafile_type = ''LOG''
) as V
PIVOT(MAX(WriteLatency) FOR DatabaseName IN (' + @ColumnName + ')) AS PVTTable

UNION ALL

SELECT measurement = ''Rows read latency (ms)'', servername = REPLACE(@@SERVERNAME, ''\'', '':''), type = ''Database stats''
, ' + @ColumnName + '  FROM
(
SELECT DatabaseName,  ReadLatency
FROM #baseline
WHERE datafile_type = ''ROWS''
) as V
PIVOT(MAX(ReadLatency) FOR DatabaseName IN (' + @ColumnName + ')) AS PVTTable

UNION ALL

SELECT measurement = ''Rows write latency (ms)'', servername = REPLACE(@@SERVERNAME, ''\'', '':''), type = ''Database stats''
, ' + @ColumnName + '  FROM
(
SELECT DatabaseName,  WriteLatency
FROM #baseline
WHERE datafile_type = ''ROWS''
) as V
PIVOT(MAX(WriteLatency) FOR DatabaseName IN (' + @ColumnName + ')) AS PVTTable

UNION ALL

SELECT measurement = ''Rows (average bytes/read)'', servername = REPLACE(@@SERVERNAME, ''\'', '':''), type = ''Database stats''
, ' + @ColumnName + '  FROM
(
SELECT DatabaseName,  AvgBytesPerRead
FROM #baseline
WHERE datafile_type = ''ROWS''
) as V
PIVOT(SUM(AvgBytesPerRead) FOR DatabaseName IN (' + @ColumnName + ')) AS PVTTable

UNION ALL

SELECT measurement = ''Rows (average bytes/write)'', servername = REPLACE(@@SERVERNAME, ''\'', '':''), type = ''Database stats''
, ' + @ColumnName + '  FROM
(
SELECT DatabaseName,  AvgBytesPerWrite
FROM #baseline
WHERE datafile_type = ''ROWS''
) as V
PIVOT(SUM(AvgBytesPerWrite) FOR DatabaseName IN (' + @ColumnName + ')) AS PVTTable

UNION ALL

SELECT measurement = ''Log (average bytes/read)'', servername = REPLACE(@@SERVERNAME, ''\'', '':''), type = ''Database stats''
, ' + @ColumnName + '  FROM
(
SELECT DatabaseName,  AvgBytesPerRead
FROM #baseline
WHERE datafile_type = ''LOG''
) as V
PIVOT(SUM(AvgBytesPerRead) FOR DatabaseName IN (' + @ColumnName + ')) AS PVTTable

UNION ALL

SELECT measurement = ''Log (average bytes/write)'', servername = REPLACE(@@SERVERNAME, ''\'', '':''), type = ''Database stats''
, ' + @ColumnName + '  FROM
(
SELECT DatabaseName,  AvgBytesPerWrite
FROM #baseline
WHERE datafile_type = ''LOG''
) as V
PIVOT(SUM(AvgBytesPerWrite) FOR DatabaseName IN (' + @ColumnName + ')) AS PVTTable
'
--PRINT @DynamicPivotQuery
EXEC sp_executesql @DynamicPivotQuery;
`

const sqlDatabaseIO string = `SET DEADLOCK_PRIORITY -10;
SET NOCOUNT ON;
SET TRANSACTION ISOLATION LEVEL READ UNCOMMITTED;
DECLARE @secondsBetween tinyint = 5;
DECLARE @delayInterval char(8) = CONVERT(Char(8), DATEADD(SECOND, @secondsBetween, '00:00:00'), 108);
IF OBJECT_ID('tempdb..#baseline') IS NOT NULL
	DROP TABLE #baseline;
IF OBJECT_ID('tempdb..#baselinewritten') IS NOT NULL
	DROP TABLE #baselinewritten;
SELECT DB_NAME(mf.database_id) AS databaseName ,
    mf.physical_name,
    divfs.num_of_bytes_read,
    divfs.num_of_bytes_written,
	divfs.num_of_reads,
	divfs.num_of_writes,
    GETDATE() AS baselinedate
INTO #baseline
FROM sys.dm_io_virtual_file_stats(NULL, NULL) AS divfs
INNER JOIN sys.master_files AS mf ON mf.database_id = divfs.database_id
	AND mf.file_id = divfs.file_id
WAITFOR DELAY @delayInterval;
;WITH currentLine AS
(
	SELECT DB_NAME(mf.database_id) AS databaseName ,
		type_desc,
		mf.physical_name,
		divfs.num_of_bytes_read,
		divfs.num_of_bytes_written,
		divfs.num_of_reads,
	    divfs.num_of_writes,
		GETDATE() AS currentlinedate
	FROM sys.dm_io_virtual_file_stats(NULL, NULL) AS divfs
	INNER JOIN sys.master_files AS mf ON mf.database_id = divfs.database_id
			AND mf.file_id = divfs.file_id
)
SELECT database_name
, datafile_type
, num_of_bytes_read_persec = SUM(num_of_bytes_read_persec)
, num_of_bytes_written_persec = SUM(num_of_bytes_written_persec)
, num_of_reads_persec = SUM(num_of_reads_persec)
, num_of_writes_persec = SUM(num_of_writes_persec)
INTO #baselinewritten
FROM
(
SELECT
  database_name = currentLine.databaseName
, datafile_type = type_desc
, num_of_bytes_read_persec = (currentLine.num_of_bytes_read - T1.num_of_bytes_read) / (DATEDIFF(SECOND,baselinedate,currentlinedate))
, num_of_bytes_written_persec = (currentLine.num_of_bytes_written - T1.num_of_bytes_written) / (DATEDIFF(SECOND,baselinedate,currentlinedate))
, num_of_reads_persec =  (currentLine.num_of_reads - T1.num_of_reads) / (DATEDIFF(SECOND,baselinedate,currentlinedate))
, num_of_writes_persec =  (currentLine.num_of_writes - T1.num_of_writes) / (DATEDIFF(SECOND,baselinedate,currentlinedate))
FROM currentLine
INNER JOIN #baseline T1 ON T1.databaseName = currentLine.databaseName
	AND T1.physical_name = currentLine.physical_name
) as T
GROUP BY database_name, datafile_type
DECLARE @DynamicPivotQuery AS NVARCHAR(MAX)
DECLARE @ColumnName AS NVARCHAR(MAX), @ColumnName2 AS NVARCHAR(MAX)
SELECT @ColumnName = ISNULL(@ColumnName + ',','') + QUOTENAME(database_name)
	FROM (SELECT DISTINCT database_name FROM #baselinewritten) AS bl
SELECT @ColumnName2 = ISNULL(@ColumnName2 + '+','') + QUOTENAME(database_name)
	FROM (SELECT DISTINCT database_name FROM #baselinewritten) AS bl
SET @DynamicPivotQuery = N'
SELECT measurement = ''Log writes (bytes/sec)'', servername = REPLACE(@@SERVERNAME, ''\'', '':''), type = ''Database IO''
, ' + @ColumnName + ', Total = ' + @ColumnName2 + ' FROM
(
SELECT database_name, num_of_bytes_written_persec
FROM #baselinewritten
WHERE datafile_type = ''LOG''
) as V
PIVOT(SUM(num_of_bytes_written_persec) FOR database_name IN (' + @ColumnName + ')) AS PVTTable
UNION ALL
SELECT measurement = ''Rows writes (bytes/sec)'', servername = REPLACE(@@SERVERNAME, ''\'', '':''), type = ''Database IO''
, ' + @ColumnName + ', Total = ' + @ColumnName2 + ' FROM
(
SELECT database_name, num_of_bytes_written_persec
FROM #baselinewritten
WHERE datafile_type = ''ROWS''
) as V
PIVOT(SUM(num_of_bytes_written_persec) FOR database_name IN (' + @ColumnName + ')) AS PVTTable
UNION ALL
SELECT measurement = ''Log reads (bytes/sec)'', servername = REPLACE(@@SERVERNAME, ''\'', '':''), type = ''Database IO''
, ' + @ColumnName + ', Total = ' + @ColumnName2 + ' FROM
(
SELECT database_name, num_of_bytes_read_persec
FROM #baselinewritten
WHERE datafile_type = ''LOG''
) as V
PIVOT(SUM(num_of_bytes_read_persec) FOR database_name IN (' + @ColumnName + ')) AS PVTTable
UNION ALL
SELECT measurement = ''Rows reads (bytes/sec)'', servername = REPLACE(@@SERVERNAME, ''\'', '':''), type = ''Database IO''
, ' + @ColumnName + ', Total = ' + @ColumnName2 + ' FROM
(
SELECT database_name, num_of_bytes_read_persec
FROM #baselinewritten
WHERE datafile_type = ''ROWS''
) as V
PIVOT(SUM(num_of_bytes_read_persec) FOR database_name IN (' + @ColumnName + ')) AS PVTTable
UNION ALL
SELECT measurement = ''Log (writes/sec)'', servername = REPLACE(@@SERVERNAME, ''\'', '':''), type = ''Database IO''
, ' + @ColumnName + ', Total = ' + @ColumnName2 + ' FROM
(
SELECT database_name, num_of_writes_persec
FROM #baselinewritten
WHERE datafile_type = ''LOG''
) as V
PIVOT(SUM(num_of_writes_persec) FOR database_name IN (' + @ColumnName + ')) AS PVTTable
UNION ALL
SELECT measurement = ''Rows (writes/sec)'', servername = REPLACE(@@SERVERNAME, ''\'', '':''), type = ''Database IO''
, ' + @ColumnName + ', Total = ' + @ColumnName2 + ' FROM
(
SELECT database_name, num_of_writes_persec
FROM #baselinewritten
WHERE datafile_type = ''ROWS''
) as V
PIVOT(SUM(num_of_writes_persec) FOR database_name IN (' + @ColumnName + ')) AS PVTTabl
UNION ALL
SELECT measurement = ''Log (reads/sec)'', servername = REPLACE(@@SERVERNAME, ''\'', '':''), type = ''Database IO''
, ' + @ColumnName + ', Total = ' + @ColumnName2 + ' FROM
(
SELECT database_name, num_of_reads_persec
FROM #baselinewritten
WHERE datafile_type = ''LOG''
) as V
PIVOT(SUM(num_of_reads_persec) FOR database_name IN (' + @ColumnName + ')) AS PVTTable
UNION ALL
SELECT measurement = ''Rows (reads/sec)'', servername = REPLACE(@@SERVERNAME, ''\'', '':''), type = ''Database IO''
, ' + @ColumnName + ', Total = ' + @ColumnName2 + ' FROM
(
SELECT database_name, num_of_reads_persec
FROM #baselinewritten
WHERE datafile_type = ''ROWS''
) as V
PIVOT(SUM(num_of_reads_persec) FOR database_name IN (' + @ColumnName + ')) AS PVTTable
'
EXEC sp_executesql @DynamicPivotQuery;
`

const sqlDatabaseProperties string = `SET DEADLOCK_PRIORITY -10;
SET NOCOUNT ON;
SET ARITHABORT ON;
SET QUOTED_IDENTIFIER ON;
SET TRANSACTION ISOLATION LEVEL READ UNCOMMITTED

IF OBJECT_ID('tempdb..#Databases') IS NOT NULL
	DROP TABLE #Databases;
CREATE  TABLE #Databases
(
	Measurement nvarchar(64) NOT NULL,
	DatabaseName nvarchar(128) NOT NULL,
	Value tinyint NOT NULL
	Primary Key(DatabaseName, Measurement)
);

INSERT #Databases (	Measurement, DatabaseName, Value)
SELECT
  Measurement = 'Recovery Model FULL'
, DatabaseName = d.Name
, Value = CASE WHEN d.recovery_model = 1 THEN 1 ELSE 0 END
FROM sys.databases d
UNION ALL
SELECT
  Measurement = 'Recovery Model BULK_LOGGED'
, DatabaseName = d.Name
, Value = CASE WHEN d.recovery_model = 2 THEN 1 ELSE 0 END
FROM sys.databases d
UNION ALL
SELECT
  Measurement = 'Recovery Model SIMPLE'
, DatabaseName = d.Name
, Value = CASE WHEN d.recovery_model = 3 THEN 1 ELSE 0 END
FROM sys.databases d

UNION ALL
SELECT
  Measurement = 'State ONLINE'
, DatabaseName = d.Name
, Value = CASE WHEN d.state = 0 THEN 1 ELSE 0 END
FROM sys.databases d
UNION ALL
SELECT
  Measurement = 'State RESTORING'
, DatabaseName = d.Name
, Value = CASE WHEN d.state = 1 THEN 1 ELSE 0 END
FROM sys.databases d
UNION ALL
SELECT
  Measurement = 'State RECOVERING'
, DatabaseName = d.Name
, Value = CASE WHEN d.state = 2 THEN 1 ELSE 0 END
FROM sys.databases d
UNION ALL
SELECT
  Measurement = 'State RECOVERY_PENDING'
, DatabaseName = d.Name
, Value = CASE WHEN d.state = 3 THEN 1 ELSE 0 END
FROM sys.databases d
UNION ALL
SELECT
  Measurement = 'State SUSPECT'
, DatabaseName = d.Name
, Value = CASE WHEN d.state = 4 THEN 1 ELSE 0 END
FROM sys.databases d
UNION ALL
SELECT
  Measurement = 'State EMERGENCY'
, DatabaseName = d.Name
, Value = CASE WHEN d.state = 5 THEN 1 ELSE 0 END
FROM sys.databases d
UNION ALL
SELECT
  Measurement = 'State OFFLINE'
, DatabaseName = d.Name
, Value = CASE WHEN d.state = 6 THEN 1 ELSE 0 END
FROM sys.databases d

DECLARE @DynamicPivotQuery AS NVARCHAR(MAX)
DECLARE @ColumnName AS NVARCHAR(MAX)
SELECT @ColumnName= ISNULL(@ColumnName + ',','') + QUOTENAME(DatabaseName)
FROM (SELECT DISTINCT DatabaseName FROM #Databases) AS bl

SET @DynamicPivotQuery = N'
SELECT measurement = Measurement, servername = REPLACE(@@SERVERNAME, ''\'', '':'')
, type = ''Database properties''
, ' + @ColumnName + ', Total FROM
(
SELECT Measurement, DatabaseName, Value
, Total = (SELECT SUM(Value) FROM #Databases WHERE Measurement = d.Measurement)
FROM #Databases d
WHERE d.Measurement = ''Recovery Model FULL''
) as V
PIVOT(SUM(Value) FOR DatabaseName IN (' + @ColumnName + ')) AS PVTTable

UNION ALL

SELECT measurement = Measurement, servername = REPLACE(@@SERVERNAME, ''\'', '':'')
, type = ''Database properties''
, ' + @ColumnName + ', Total FROM
(
SELECT Measurement, DatabaseName, Value
, Total = (SELECT SUM(Value) FROM #Databases WHERE Measurement = d.Measurement)
FROM #Databases d
WHERE d.Measurement = ''Recovery Model BULK_LOGGED''
) as V
PIVOT(SUM(Value) FOR DatabaseName IN (' + @ColumnName + ')) AS PVTTable

UNION ALL

SELECT measurement = Measurement, servername = REPLACE(@@SERVERNAME, ''\'', '':'')
, type = ''Database properties''
, ' + @ColumnName + ', Total FROM
(
SELECT Measurement, DatabaseName, Value
, Total = (SELECT SUM(Value) FROM #Databases WHERE Measurement = d.Measurement)
FROM #Databases d
WHERE d.Measurement = ''Recovery Model SIMPLE''
) as V
PIVOT(SUM(Value) FOR DatabaseName IN (' + @ColumnName + ')) AS PVTTable


UNION ALL

SELECT measurement = Measurement, servername = REPLACE(@@SERVERNAME, ''\'', '':'')
, type = ''Database properties''
, ' + @ColumnName + ', Total FROM
(
SELECT Measurement, DatabaseName, Value
, Total = (SELECT SUM(Value) FROM #Databases WHERE Measurement = d.Measurement)
FROM #Databases d
WHERE d.Measurement = ''State ONLINE''
) as V
PIVOT(SUM(Value) FOR DatabaseName IN (' + @ColumnName + ')) AS PVTTable

UNION ALL

SELECT measurement = Measurement, servername = REPLACE(@@SERVERNAME, ''\'', '':'')
, type = ''Database properties''
, ' + @ColumnName + ', Total FROM
(
SELECT Measurement, DatabaseName, Value
, Total = (SELECT SUM(Value) FROM #Databases WHERE Measurement = d.Measurement)
FROM #Databases d
WHERE d.Measurement = ''State RESTORING''
) as V
PIVOT(SUM(Value) FOR DatabaseName IN (' + @ColumnName + ')) AS PVTTable

UNION ALL

SELECT measurement = Measurement, servername = REPLACE(@@SERVERNAME, ''\'', '':'')
, type = ''Database properties''
, ' + @ColumnName + ', Total FROM
(
SELECT Measurement, DatabaseName, Value
, Total = (SELECT SUM(Value) FROM #Databases WHERE Measurement = d.Measurement)
FROM #Databases d
WHERE d.Measurement = ''State RECOVERING''
) as V
PIVOT(SUM(Value) FOR DatabaseName IN (' + @ColumnName + ')) AS PVTTable

UNION ALL

SELECT measurement = Measurement, servername = REPLACE(@@SERVERNAME, ''\'', '':'')
, type = ''Database properties''
, ' + @ColumnName + ', Total FROM
(
SELECT Measurement, DatabaseName, Value
, Total = (SELECT SUM(Value) FROM #Databases WHERE Measurement = d.Measurement)
FROM #Databases d
WHERE d.Measurement = ''State RECOVERY_PENDING''
) as V
PIVOT(SUM(Value) FOR DatabaseName IN (' + @ColumnName + ')) AS PVTTable

UNION ALL

SELECT measurement = Measurement, servername = REPLACE(@@SERVERNAME, ''\'', '':'')
, type = ''Database properties''
, ' + @ColumnName + ', Total FROM
(
SELECT Measurement, DatabaseName, Value
, Total = (SELECT SUM(Value) FROM #Databases WHERE Measurement = d.Measurement)
FROM #Databases d
WHERE d.Measurement = ''State SUSPECT''
) as V
PIVOT(SUM(Value) FOR DatabaseName IN (' + @ColumnName + ')) AS PVTTable

UNION ALL

SELECT measurement = Measurement, servername = REPLACE(@@SERVERNAME, ''\'', '':'')
, type = ''Database properties''
, ' + @ColumnName + ', Total FROM
(
SELECT Measurement, DatabaseName, Value
, Total = (SELECT SUM(Value) FROM #Databases WHERE Measurement = d.Measurement)
FROM #Databases d
WHERE d.Measurement = ''State EMERGENCY''
) as V
PIVOT(SUM(Value) FOR DatabaseName IN (' + @ColumnName + ')) AS PVTTable

UNION ALL

SELECT measurement = Measurement, servername = REPLACE(@@SERVERNAME, ''\'', '':'')
, type = ''Database properties''
, ' + @ColumnName + ', Total FROM
(
SELECT Measurement, DatabaseName, Value
, Total = (SELECT SUM(Value) FROM #Databases WHERE Measurement = d.Measurement)
FROM #Databases d
WHERE d.Measurement = ''State OFFLINE''
) as V
PIVOT(SUM(Value) FOR DatabaseName IN (' + @ColumnName + ')) AS PVTTable
'
EXEC sp_executesql @DynamicPivotQuery;
`

const sqlCPUHistory string = `SET DEADLOCK_PRIORITY -10;
SET NOCOUNT ON;
SET ARITHABORT ON;
SET QUOTED_IDENTIFIER ON;
SET TRANSACTION ISOLATION LEVEL READ UNCOMMITTED;

DECLARE @ms_ticks bigint;
SET @ms_ticks = (Select ms_ticks From sys.dm_os_sys_info);
DECLARE @maxEvents int = 1

SELECT
---- measurement
  measurement = 'CPU (%)'
---- tags
, servername= REPLACE(@@SERVERNAME, '\', ':')
, type = 'CPU usage'
-- value
, [SQL process] = ProcessUtilization
, [External process]= 100 - SystemIdle - ProcessUtilization
, [SystemIdle]
FROM
(
SELECT TOP (@maxEvents)
  EventTime = CAST(DateAdd(ms, -1 * (@ms_ticks - timestamp_ms), GetUTCDate()) as datetime)
, ProcessUtilization = CAST(ProcessUtilization as int)
, SystemIdle = CAST(SystemIdle as int)
FROM (SELECT Record.value('(./Record/SchedulerMonitorEvent/SystemHealth/SystemIdle)[1]', 'int') as SystemIdle,
		     Record.value('(./Record/SchedulerMonitorEvent/SystemHealth/ProcessUtilization)[1]', 'int') as ProcessUtilization,
		     timestamp as timestamp_ms
FROM (SELECT timestamp, convert(xml, record) As Record
		FROM sys.dm_os_ring_buffers
		WHERE ring_buffer_type = N'RING_BUFFER_SCHEDULER_MONITOR'
		    And record Like '%<SystemHealth>%') x) y
ORDER BY timestamp_ms Desc
) as T;
`

const sqlPerformanceCounters string = `SET DEADLOCK_PRIORITY -10;
SET NOCOUNT ON;
SET TRANSACTION ISOLATION LEVEL READ UNCOMMITTED;
IF OBJECT_ID('tempdb..#PCounters') IS NOT NULL DROP TABLE #PCounters
CREATE TABLE #PCounters
(
	object_name nvarchar(128),
	counter_name nvarchar(128),
	instance_name nvarchar(128),
	cntr_value bigint,
	cntr_type INT,
	Primary Key(object_name, counter_name, instance_name)
);
INSERT #PCounters
SELECT DISTINCT RTrim(spi.object_name) object_name
, RTrim(spi.counter_name) counter_name
, RTrim(spi.instance_name) instance_name
, spi.cntr_value
, spi.cntr_type
FROM sys.dm_os_performance_counters spi
WHERE spi.object_name NOT LIKE '%Backup Device%'
	AND NOT EXISTS (SELECT 1 FROM sys.databases WHERE Name = spi.instance_name);

WAITFOR DELAY '00:00:01';

IF OBJECT_ID('tempdb..#CCounters') IS NOT NULL DROP TABLE #CCounters
CREATE TABLE #CCounters
(
	object_name nvarchar(128),
	counter_name nvarchar(128),
	instance_name nvarchar(128),
	cntr_value bigint,
	cntr_type INT,
	Primary Key(object_name, counter_name, instance_name)
);
INSERT #CCounters
SELECT DISTINCT RTrim(spi.object_name) object_name
, RTrim(spi.counter_name) counter_name
, RTrim(spi.instance_name) instance_name
, spi.cntr_value
, spi.cntr_type
FROM sys.dm_os_performance_counters spi
WHERE spi.object_name NOT LIKE '%Backup Device%'
	AND NOT EXISTS (SELECT 1 FROM sys.databases WHERE Name = spi.instance_name);

SELECT
 measurement = cc.counter_name
	+ CASE WHEN LEN(cc.instance_name) > 0 THEN ' | ' + cc.instance_name ELSE '' END
	+ ' | '
	+ SUBSTRING( cc.object_name, CHARINDEX(':',  cc.object_name) + 1, LEN( cc.object_name) - CHARINDEX(':',  cc.object_name))
-- tags
, servername = REPLACE(@@SERVERNAME, '\', ':')
, type = 'Performance counters'
--, countertype = CASE cc.cntr_type
--    When 65792 Then 'Count'
--    When 537003264 Then 'Ratio'
--    When 272696576 Then 'Per second'
--    When 1073874176 Then 'Average'
--    When 272696320 Then 'Average Per Second'
--    When 1073939712 Then 'Base'
--    END
-- value
, value = CAST(CASE cc.cntr_type
    When 65792 Then cc.cntr_value -- Count
    When 537003264 Then IsNull(Cast(cc.cntr_value as decimal(19,4)) / NullIf(cbc.cntr_value, 0), 0) -- Ratio
    When 272696576 Then cc.cntr_value - pc.cntr_value -- Per Second
    When 1073874176 Then IsNull(Cast(cc.cntr_value - pc.cntr_value as decimal(19,4)) / NullIf(cbc.cntr_value - pbc.cntr_value, 0), 0) -- Avg
    When 272696320 Then IsNull(Cast(cc.cntr_value - pc.cntr_value as decimal(19,4)) / NullIf(cbc.cntr_value - pbc.cntr_value, 0), 0) -- Avg/sec
    When 1073939712 Then cc.cntr_value - pc.cntr_value -- Base
    Else cc.cntr_value End as bigint)
--, currentvalue= CAST(cc.cntr_value as bigint)
FROM #CCounters cc
INNER JOIN #PCounters pc On cc.object_name = pc.object_name
        And cc.counter_name = pc.counter_name
        And cc.instance_name = pc.instance_name
        And cc.cntr_type = pc.cntr_type
LEFT JOIN #CCounters cbc On cc.object_name = cbc.object_name
        And (Case When cc.counter_name Like '%(ms)' Then Replace(cc.counter_name, ' (ms)',' Base')
                  When cc.object_name like '%FileTable' Then Replace(cc.counter_name, 'Avg ','') + ' base'
                  When cc.counter_name = 'Worktables From Cache Ratio' Then 'Worktables From Cache Base'
                  When cc.counter_name = 'Avg. Length of Batched Writes' Then 'Avg. Length of Batched Writes BS'
                  Else cc.counter_name + ' base'
             End) = cbc.counter_name
        And cc.instance_name = cbc.instance_name
        And cc.cntr_type In (537003264, 1073874176)
        And cbc.cntr_type = 1073939712
LEFT JOIN #PCounters pbc On pc.object_name = pbc.object_name
        And pc.instance_name = pbc.instance_name
        And (Case When pc.counter_name Like '%(ms)' Then Replace(pc.counter_name, ' (ms)',' Base')
                  When pc.object_name like '%FileTable' Then Replace(pc.counter_name, 'Avg ','') + ' base'
                  When pc.counter_name = 'Worktables From Cache Ratio' Then 'Worktables From Cache Base'
                  When pc.counter_name = 'Avg. Length of Batched Writes' Then 'Avg. Length of Batched Writes BS'
                  Else pc.counter_name + ' base'
             End) = pbc.counter_name
        And pc.cntr_type In (537003264, 1073874176)

IF OBJECT_ID('tempdb..#CCounters') IS NOT NULL DROP TABLE #CCounters;
IF OBJECT_ID('tempdb..#PCounters') IS NOT NULL DROP TABLE #PCounters;
`

const sqlWaitStatsCategorized string = `SET DEADLOCK_PRIORITY -10;
SET NOCOUNT ON;
SET TRANSACTION ISOLATION LEVEL READ UNCOMMITTED
DECLARE @secondsBetween tinyint = 5
DECLARE @delayInterval char(8) = CONVERT(Char(8), DATEADD(SECOND, @secondsBetween, '00:00:00'), 108);

DECLARE @w1 TABLE
(
	WaitType nvarchar(64) collate SQL_Latin1_General_CP1_CI_AS NOT NULL,
	WaitTimeInMs bigint NOT NULL,
	WaitTaskCount bigint NOT NULL,
	CollectionDate datetime NOT NULL
)
DECLARE @w2 TABLE
(
	WaitType nvarchar(64) collate SQL_Latin1_General_CP1_CI_AS NOT NULL,
	WaitTimeInMs bigint NOT NULL,
	WaitTaskCount bigint NOT NULL,
	CollectionDate datetime NOT NULL
)
DECLARE @w3 TABLE
(
	WaitType nvarchar(64) collate SQL_Latin1_General_CP1_CI_AS NOT NULL
)
DECLARE @w4 TABLE
(
	WaitType nvarchar(64) collate SQL_Latin1_General_CP1_CI_AS NOT NULL,
	WaitCategory nvarchar(64) collate SQL_Latin1_General_CP1_CI_AS NOT NULL
)
DECLARE @w5 TABLE
(
	WaitCategory nvarchar(64) collate SQL_Latin1_General_CP1_CI_AS NOT NULL,
	WaitTimeInMs bigint NOT NULL,
	WaitTaskCount bigint NOT NULL
)

INSERT @w3 (WaitType)
VALUES (N'QDS_SHUTDOWN_QUEUE'), (N'HADR_FILESTREAM_IOMGR_IOCOMPLETION'),
	(N'BROKER_EVENTHANDLER'),            (N'BROKER_RECEIVE_WAITFOR'),
	(N'BROKER_TASK_STOP'),               (N'BROKER_TO_FLUSH'),
	(N'BROKER_TRANSMITTER'),             (N'CHECKPOINT_QUEUE'),
	(N'CHKPT'),                          (N'CLR_AUTO_EVENT'),
	(N'CLR_MANUAL_EVENT'),               (N'CLR_SEMAPHORE'),
	(N'DBMIRROR_DBM_EVENT'),             (N'DBMIRROR_EVENTS_QUEUE'),
	(N'DBMIRROR_WORKER_QUEUE'),          (N'DBMIRRORING_CMD'),
	(N'DIRTY_PAGE_POLL'),                (N'DISPATCHER_QUEUE_SEMAPHORE'),
	(N'EXECSYNC'),                       (N'FSAGENT'),
	(N'FT_IFTS_SCHEDULER_IDLE_WAIT'),    (N'FT_IFTSHC_MUTEX'),
	(N'HADR_CLUSAPI_CALL'),              (N'HADR_FILESTREAM_IOMGR_IOCOMPLETIO(N'),
	(N'HADR_LOGCAPTURE_WAIT'),           (N'HADR_NOTIFICATION_DEQUEUE'),
	(N'HADR_TIMER_TASK'),                (N'HADR_WORK_QUEUE'),
	(N'KSOURCE_WAKEUP'),                 (N'LAZYWRITER_SLEEP'),
	(N'LOGMGR_QUEUE'),                   (N'ONDEMAND_TASK_QUEUE'),
	(N'PWAIT_ALL_COMPONENTS_INITIALIZED'),
	(N'QDS_PERSIST_TASK_MAIN_LOOP_SLEEP'),
	(N'QDS_CLEANUP_STALE_QUERIES_TASK_MAIN_LOOP_SLEEP'),
	(N'REQUEST_FOR_DEADLOCK_SEARCH'),    (N'RESOURCE_QUEUE'),
	(N'SERVER_IDLE_CHECK'),              (N'SLEEP_BPOOL_FLUSH'),
	(N'SLEEP_DBSTARTUP'),                (N'SLEEP_DCOMSTARTUP'),
	(N'SLEEP_MASTERDBREADY'),            (N'SLEEP_MASTERMDREADY'),
	(N'SLEEP_MASTERUPGRADED'),           (N'SLEEP_MSDBSTARTUP'),
	(N'SLEEP_SYSTEMTASK'),               (N'SLEEP_TASK'),
	(N'SLEEP_TEMPDBSTARTUP'),            (N'SNI_HTTP_ACCEPT'),
	(N'SP_SERVER_DIAGNOSTICS_SLEEP'),    (N'SQLTRACE_BUFFER_FLUSH'),
	(N'SQLTRACE_INCREMENTAL_FLUSH_SLEEP'),
	(N'SQLTRACE_WAIT_ENTRIES'),          (N'WAIT_FOR_RESULTS'),
	(N'WAITFOR'),                        (N'WAITFOR_TASKSHUTDOW(N'),
	(N'WAIT_XTP_HOST_WAIT'),             (N'WAIT_XTP_OFFLINE_CKPT_NEW_LOG'),
	(N'WAIT_XTP_CKPT_CLOSE'),            (N'XE_DISPATCHER_JOI(N'),
	(N'XE_DISPATCHER_WAIT'),             (N'XE_TIMER_EVENT');

INSERT @w4 (WaitType, WaitCategory) VALUES ('ABR', 'OTHER') ,
('ASSEMBLY_LOAD' , 'OTHER') , ('ASYNC_DISKPOOL_LOCK' , 'I/O') , ('ASYNC_IO_COMPLETION' , 'I/O') ,
('ASYNC_NETWORK_IO' , 'NETWORK') , ('AUDIT_GROUPCACHE_LOCK' , 'OTHER') , ('AUDIT_LOGINCACHE_LOCK' ,
'OTHER') , ('AUDIT_ON_DEMAND_TARGET_LOCK' , 'OTHER') , ('AUDIT_XE_SESSION_MGR' , 'OTHER') , ('BACKUP' ,
'BACKUP') , ('BACKUP_CLIENTLOCK ' , 'BACKUP') , ('BACKUP_OPERATOR' , 'BACKUP') , ('BACKUPBUFFER' ,
'BACKUP') , ('BACKUPIO' , 'BACKUP') , ('BACKUPTHREAD' , 'BACKUP') , ('BAD_PAGE_PROCESS' , 'MEMORY') ,
('BROKER_CONNECTION_RECEIVE_TASK' , 'SERVICE BROKER') , ('BROKER_ENDPOINT_STATE_MUTEX' , 'SERVICE BROKER')
, ('BROKER_EVENTHANDLER' , 'SERVICE BROKER') , ('BROKER_INIT' , 'SERVICE BROKER') , ('BROKER_MASTERSTART'
, 'SERVICE BROKER') , ('BROKER_RECEIVE_WAITFOR' , 'SERVICE BROKER') , ('BROKER_REGISTERALLENDPOINTS' ,
'SERVICE BROKER') , ('BROKER_SERVICE' , 'SERVICE BROKER') , ('BROKER_SHUTDOWN' , 'SERVICE BROKER') ,
('BROKER_TASK_STOP' , 'SERVICE BROKER') , ('BROKER_TO_FLUSH' , 'SERVICE BROKER') , ('BROKER_TRANSMITTER' ,
'SERVICE BROKER') , ('BUILTIN_HASHKEY_MUTEX' , 'OTHER') , ('CHECK_PRINT_RECORD' , 'OTHER') ,
('CHECKPOINT_QUEUE' , 'BUFFER') , ('CHKPT' , 'BUFFER') , ('CLEAR_DB' , 'OTHER') , ('CLR_AUTO_EVENT' ,
'CLR') , ('CLR_CRST' , 'CLR') , ('CLR_JOIN' , 'CLR') , ('CLR_MANUAL_EVENT' , 'CLR') , ('CLR_MEMORY_SPY' ,
'CLR') , ('CLR_MONITOR' , 'CLR') , ('CLR_RWLOCK_READER' , 'CLR') , ('CLR_RWLOCK_WRITER' , 'CLR') ,
('CLR_SEMAPHORE' , 'CLR') , ('CLR_TASK_START' , 'CLR') , ('CLRHOST_STATE_ACCESS' , 'CLR') , ('CMEMTHREAD'
, 'MEMORY') , ('COMMIT_TABLE' , 'OTHER') , ('CURSOR' , 'OTHER') , ('CURSOR_ASYNC' , 'OTHER') , ('CXPACKET'
, 'OTHER') , ('CXROWSET_SYNC' , 'OTHER') , ('DAC_INIT' , 'OTHER') , ('DBMIRROR_DBM_EVENT ' , 'OTHER') ,
('DBMIRROR_DBM_MUTEX ' , 'OTHER') , ('DBMIRROR_EVENTS_QUEUE' , 'OTHER') , ('DBMIRROR_SEND' , 'OTHER') ,
('DBMIRROR_WORKER_QUEUE' , 'OTHER') , ('DBMIRRORING_CMD' , 'OTHER') , ('DBTABLE' , 'OTHER') ,
('DEADLOCK_ENUM_MUTEX' , 'LOCK') , ('DEADLOCK_TASK_SEARCH' , 'LOCK') , ('DEBUG' , 'OTHER') ,
('DISABLE_VERSIONING' , 'OTHER') , ('DISKIO_SUSPEND' , 'BACKUP') , ('DISPATCHER_QUEUE_SEMAPHORE' ,
'OTHER') , ('DLL_LOADING_MUTEX' , 'XML') , ('DROPTEMP' , 'TEMPORARY OBJECTS') , ('DTC' , 'OTHER') ,
('DTC_ABORT_REQUEST' , 'OTHER') , ('DTC_RESOLVE' , 'OTHER') , ('DTC_STATE' , 'DOTHERTC') ,
('DTC_TMDOWN_REQUEST' , 'OTHER') , ('DTC_WAITFOR_OUTCOME' , 'OTHER') , ('DUMP_LOG_COORDINATOR' , 'OTHER')
, ('DUMP_LOG_COORDINATOR_QUEUE' , 'OTHER') , ('DUMPTRIGGER' , 'OTHER') , ('EC' , 'OTHER') , ('EE_PMOLOCK'
, 'MEMORY') , ('EE_SPECPROC_MAP_INIT' , 'OTHER') , ('ENABLE_VERSIONING' , 'OTHER') ,
('ERROR_REPORTING_MANAGER' , 'OTHER') , ('EXCHANGE' , 'OTHER') , ('EXECSYNC' , 'OTHER') ,
('EXECUTION_PIPE_EVENT_OTHER' , 'OTHER') , ('Failpoint' , 'OTHER') , ('FCB_REPLICA_READ' , 'OTHER') ,
('FCB_REPLICA_WRITE' , 'OTHER') , ('FS_FC_RWLOCK' , 'OTHER') , ('FS_GARBAGE_COLLECTOR_SHUTDOWN' , 'OTHER')
, ('FS_HEADER_RWLOCK' , 'OTHER') , ('FS_LOGTRUNC_RWLOCK' , 'OTHER') , ('FSA_FORCE_OWN_XACT' , 'OTHER') ,
('FSAGENT' , 'OTHER') , ('FSTR_CONFIG_MUTEX' , 'OTHER') , ('FSTR_CONFIG_RWLOCK' , 'OTHER') ,
('FT_COMPROWSET_RWLOCK' , 'OTHER') , ('FT_IFTS_RWLOCK' , 'OTHER') , ('FT_IFTS_SCHEDULER_IDLE_WAIT' ,
'OTHER') , ('FT_IFTSHC_MUTEX' , 'OTHER') , ('FT_IFTSISM_MUTEX' , 'OTHER') , ('FT_MASTER_MERGE' , 'OTHER')
, ('FT_METADATA_MUTEX' , 'OTHER') , ('FT_RESTART_CRAWL' , 'OTHER') , ('FT_RESUME_CRAWL' , 'OTHER') ,
('FULLTEXT GATHERER' , 'OTHER') , ('GUARDIAN' , 'OTHER') , ('HTTP_ENDPOINT_COLLCREATE' , 'SERVICE BROKER')
, ('HTTP_ENUMERATION' , 'SERVICE BROKER') , ('HTTP_START' , 'SERVICE BROKER') , ('IMP_IMPORT_MUTEX' ,
'OTHER') , ('IMPPROV_IOWAIT' , 'I/O') , ('INDEX_USAGE_STATS_MUTEX' , 'OTHER') , ('OTHER_TESTING' ,
'OTHER') , ('IO_AUDIT_MUTEX' , 'OTHER') , ('IO_COMPLETION' , 'I/O') , ('IO_RETRY' , 'I/O') ,
('IOAFF_RANGE_QUEUE' , 'OTHER') , ('KSOURCE_WAKEUP' , 'SHUTDOWN') , ('KTM_ENLISTMENT' , 'OTHER') ,
('KTM_RECOVERY_MANAGER' , 'OTHER') , ('KTM_RECOVERY_RESOLUTION' , 'OTHER') , ('LATCH_DT' , 'LATCH') ,
('LATCH_EX' , 'LATCH') , ('LATCH_KP' , 'LATCH') , ('LATCH_NL' , 'LATCH') , ('LATCH_SH' , 'LATCH') ,
('LATCH_UP' , 'LATCH') , ('LAZYWRITER_SLEEP' , 'BUFFER') , ('LCK_M_BU' , 'LOCK') , ('LCK_M_IS' , 'LOCK') ,
('LCK_M_IU' , 'LOCK') , ('LCK_M_IX' , 'LOCK') , ('LCK_M_RIn_NL' , 'LOCK') , ('LCK_M_RIn_S' , 'LOCK') ,
('LCK_M_RIn_U' , 'LOCK') , ('LCK_M_RIn_X' , 'LOCK') , ('LCK_M_RS_S' , 'LOCK') , ('LCK_M_RS_U' , 'LOCK') ,
('LCK_M_RX_S' , 'LOCK') , ('LCK_M_RX_U' , 'LOCK') , ('LCK_M_RX_X' , 'LOCK') , ('LCK_M_S' , 'LOCK') ,
('LCK_M_SCH_M' , 'LOCK') , ('LCK_M_SCH_S' , 'LOCK') , ('LCK_M_SIU' , 'LOCK') , ('LCK_M_SIX' , 'LOCK') ,
('LCK_M_U' , 'LOCK') , ('LCK_M_UIX' , 'LOCK') , ('LCK_M_X' , 'LOCK') , ('LOGBUFFER' , 'OTHER') ,
('LOGGENERATION' , 'OTHER') , ('LOGMGR' , 'OTHER') , ('LOGMGR_FLUSH' , 'OTHER') , ('LOGMGR_QUEUE' ,
'OTHER') , ('LOGMGR_RESERVE_APPEND' , 'OTHER') , ('LOWFAIL_MEMMGR_QUEUE' , 'MEMORY') ,
('METADATA_LAZYCACHE_RWLOCK' , 'OTHER') , ('MIRROR_SEND_MESSAGE' , 'OTHER') , ('MISCELLANEOUS' , 'IGNORE')
, ('MSQL_DQ' , 'DISTRIBUTED QUERY') , ('MSQL_SYNC_PIPE' , 'OTHER') , ('MSQL_XACT_MGR_MUTEX' , 'OTHER') ,
('MSQL_XACT_MUTEX' , 'OTHER') , ('MSQL_XP' , 'OTHER') , ('MSSEARCH' , 'OTHER') , ('NET_WAITFOR_PACKET' ,
'NETWORK') , ('NODE_CACHE_MUTEX' , 'OTHER') , ('OTHER' , 'OTHER') , ('ONDEMAND_TASK_QUEUE' , 'OTHER') ,
('PAGEIOLATCH_DT' , 'LATCH') , ('PAGEIOLATCH_EX' , 'LATCH') , ('PAGEIOLATCH_KP' , 'LATCH') ,
('PAGEIOLATCH_NL' , 'LATCH') , ('PAGEIOLATCH_SH' , 'LATCH') , ('PAGEIOLATCH_UP' , 'LATCH') ,
('PAGELATCH_DT' , 'LATCH') , ('PAGELATCH_EX' , 'LATCH') , ('PAGELATCH_KP' , 'LATCH') , ('PAGELATCH_NL' ,
'LATCH') , ('PAGELATCH_SH' , 'LATCH') , ('PAGELATCH_UP' , 'LATCH') , ('PARALLEL_BACKUP_QUEUE' , 'BACKUP')
, ('PERFORMANCE_COUNTERS_RWLOCK' , 'OTHER') , ('PREEMPTIVE_ABR' , 'OTHER') ,
('PREEMPTIVE_AUDIT_ACCESS_EVENTLOG' , 'OTHER') , ('PREEMPTIVE_AUDIT_ACCESS_SECLOG' , 'OTHER') ,
('PREEMPTIVE_CLOSEBACKUPMEDIA' , 'OTHER') , ('PREEMPTIVE_CLOSEBACKUPTAPE' , 'OTHER') ,
('PREEMPTIVE_CLOSEBACKUPVDIDEVICE' , 'OTHER') , ('PREEMPTIVE_CLUSAPI_CLUSTERRESOURCECONTROL' , 'OTHER') ,
('PREEMPTIVE_COM_COCREATEINSTANCE' , 'OTHER') , ('PREEMPTIVE_COM_COGETCLASSOBJECT' , 'OTHER') ,
('PREEMPTIVE_COM_CREATEACCESSOR' , 'OTHER') , ('PREEMPTIVE_COM_DELETEROWS' , 'OTHER') ,
('PREEMPTIVE_COM_GETCOMMANDTEXT' , 'OTHER') , ('PREEMPTIVE_COM_GETDATA' , 'OTHER') ,
('PREEMPTIVE_COM_GETNEXTROWS' , 'OTHER') , ('PREEMPTIVE_COM_GETRESULT' , 'OTHER') ,
('PREEMPTIVE_COM_GETROWSBYBOOKMARK' , 'OTHER') , ('PREEMPTIVE_COM_LBFLUSH' , 'OTHER') ,
('PREEMPTIVE_COM_LBLOCKREGION' , 'OTHER') , ('PREEMPTIVE_COM_LBREADAT' , 'OTHER') ,
('PREEMPTIVE_COM_LBSETSIZE' , 'OTHER') , ('PREEMPTIVE_COM_LBSTAT' , 'OTHER') ,
('PREEMPTIVE_COM_LBUNLOCKREGION' , 'OTHER') , ('PREEMPTIVE_COM_LBWRITEAT' , 'OTHER') ,
('PREEMPTIVE_COM_QUERYINTERFACE' , 'OTHER') , ('PREEMPTIVE_COM_RELEASE' , 'OTHER') ,
('PREEMPTIVE_COM_RELEASEACCESSOR' , 'OTHER') , ('PREEMPTIVE_COM_RELEASEROWS' , 'OTHER') ,
('PREEMPTIVE_COM_RELEASESESSION' , 'OTHER') , ('PREEMPTIVE_COM_RESTARTPOSITION' , 'OTHER') ,
('PREEMPTIVE_COM_SEQSTRMREAD' , 'OTHER') , ('PREEMPTIVE_COM_SEQSTRMREADANDWRITE' , 'OTHER') ,
('PREEMPTIVE_COM_SETDATAFAILURE' , 'OTHER') , ('PREEMPTIVE_COM_SETPARAMETERINFO' , 'OTHER') ,
('PREEMPTIVE_COM_SETPARAMETERPROPERTIES' , 'OTHER') , ('PREEMPTIVE_COM_STRMLOCKREGION' , 'OTHER') ,
('PREEMPTIVE_COM_STRMSEEKANDREAD' , 'OTHER') , ('PREEMPTIVE_COM_STRMSEEKANDWRITE' , 'OTHER') ,
('PREEMPTIVE_COM_STRMSETSIZE' , 'OTHER') , ('PREEMPTIVE_COM_STRMSTAT' , 'OTHER') ,
('PREEMPTIVE_COM_STRMUNLOCKREGION' , 'OTHER') , ('PREEMPTIVE_CONSOLEWRITE' , 'OTHER') ,
('PREEMPTIVE_CREATEPARAM' , 'OTHER') , ('PREEMPTIVE_DEBUG' , 'OTHER') , ('PREEMPTIVE_DFSADDLINK' ,
'OTHER') , ('PREEMPTIVE_DFSLINKEXISTCHECK' , 'OTHER') , ('PREEMPTIVE_DFSLINKHEALTHCHECK' , 'OTHER') ,
('PREEMPTIVE_DFSREMOVELINK' , 'OTHER') , ('PREEMPTIVE_DFSREMOVEROOT' , 'OTHER') ,
('PREEMPTIVE_DFSROOTFOLDERCHECK' , 'OTHER') , ('PREEMPTIVE_DFSROOTINIT' , 'OTHER') ,
('PREEMPTIVE_DFSROOTSHARECHECK' , 'OTHER') , ('PREEMPTIVE_DTC_ABORT' , 'OTHER') ,
('PREEMPTIVE_DTC_ABORTREQUESTDONE' , 'OTHER') , ('PREEMPTIVE_DTC_BEGINOTHER' , 'OTHER') ,
('PREEMPTIVE_DTC_COMMITREQUESTDONE' , 'OTHER') , ('PREEMPTIVE_DTC_ENLIST' , 'OTHER') ,
('PREEMPTIVE_DTC_PREPAREREQUESTDONE' , 'OTHER') , ('PREEMPTIVE_FILESIZEGET' , 'OTHER') ,
('PREEMPTIVE_FSAOTHER_ABORTOTHER' , 'OTHER') , ('PREEMPTIVE_FSAOTHER_COMMITOTHER' , 'OTHER') ,
('PREEMPTIVE_FSAOTHER_STARTOTHER' , 'OTHER') , ('PREEMPTIVE_FSRECOVER_UNCONDITIONALUNDO' , 'OTHER') ,
('PREEMPTIVE_GETRMINFO' , 'OTHER') , ('PREEMPTIVE_LOCKMONITOR' , 'OTHER') , ('PREEMPTIVE_MSS_RELEASE' ,
'OTHER') , ('PREEMPTIVE_ODBCOPS' , 'OTHER') , ('PREEMPTIVE_OLE_UNINIT' , 'OTHER') ,
('PREEMPTIVE_OTHER_ABORTORCOMMITTRAN' , 'OTHER') , ('PREEMPTIVE_OTHER_ABORTTRAN' , 'OTHER') ,
('PREEMPTIVE_OTHER_GETDATASOURCE' , 'OTHER') , ('PREEMPTIVE_OTHER_GETLITERALINFO' , 'OTHER') ,
('PREEMPTIVE_OTHER_GETPROPERTIES' , 'OTHER') , ('PREEMPTIVE_OTHER_GETPROPERTYINFO' , 'OTHER') ,
('PREEMPTIVE_OTHER_GETSCHEMALOCK' , 'OTHER') , ('PREEMPTIVE_OTHER_JOINOTHER' , 'OTHER') ,
('PREEMPTIVE_OTHER_RELEASE' , 'OTHER') , ('PREEMPTIVE_OTHER_SETPROPERTIES' , 'OTHER') ,
('PREEMPTIVE_OTHEROPS' , 'OTHER') , ('PREEMPTIVE_OS_ACCEPTSECURITYCONTEXT' , 'OTHER') ,
('PREEMPTIVE_OS_ACQUIRECREDENTIALSHANDLE' , 'OTHER') , ('PREEMPTIVE_OS_AU,TICATIONOPS' , 'OTHER') ,
('PREEMPTIVE_OS_AUTHORIZATIONOPS' , 'OTHER') , ('PREEMPTIVE_OS_AUTHZGETINFORMATIONFROMCONTEXT' , 'OTHER')
, ('PREEMPTIVE_OS_AUTHZINITIALIZECONTEXTFROMSID' , 'OTHER') ,
('PREEMPTIVE_OS_AUTHZINITIALIZERESOURCEMANAGER' , 'OTHER') , ('PREEMPTIVE_OS_BACKUPREAD' , 'OTHER') ,
('PREEMPTIVE_OS_CLOSEHANDLE' , 'OTHER') , ('PREEMPTIVE_OS_CLUSTEROPS' , 'OTHER') , ('PREEMPTIVE_OS_COMOPS'
, 'OTHER') , ('PREEMPTIVE_OS_COMPLETEAUTHTOKEN' , 'OTHER') , ('PREEMPTIVE_OS_COPYFILE' , 'OTHER') ,
('PREEMPTIVE_OS_CREATEDIRECTORY' , 'OTHER') , ('PREEMPTIVE_OS_CREATEFILE' , 'OTHER') ,
('PREEMPTIVE_OS_CRYPTACQUIRECONTEXT' , 'OTHER') , ('PREEMPTIVE_OS_CRYPTIMPORTKEY' , 'OTHER') ,
('PREEMPTIVE_OS_CRYPTOPS' , 'OTHER') , ('PREEMPTIVE_OS_DECRYPTMESSAGE' , 'OTHER') ,
('PREEMPTIVE_OS_DELETEFILE' , 'OTHER') , ('PREEMPTIVE_OS_DELETESECURITYCONTEXT' , 'OTHER') ,
('PREEMPTIVE_OS_DEVICEIOCONTROL' , 'OTHER') , ('PREEMPTIVE_OS_DEVICEOPS' , 'OTHER') ,
('PREEMPTIVE_OS_DIRSVC_NETWORKOPS' , 'OTHER') , ('PREEMPTIVE_OS_DISCONNECTNAMEDPIPE' , 'OTHER') ,
('PREEMPTIVE_OS_DOMAINSERVICESOPS' , 'OTHER') , ('PREEMPTIVE_OS_DSGETDCNAME' , 'OTHER') ,
('PREEMPTIVE_OS_DTCOPS' , 'OTHER') , ('PREEMPTIVE_OS_ENCRYPTMESSAGE' , 'OTHER') , ('PREEMPTIVE_OS_FILEOPS'
, 'OTHER') , ('PREEMPTIVE_OS_FINDFILE' , 'OTHER') , ('PREEMPTIVE_OS_FLUSHFILEBUFFERS' , 'OTHER') ,
('PREEMPTIVE_OS_FORMATMESSAGE' , 'OTHER') , ('PREEMPTIVE_OS_FREECREDENTIALSHANDLE' , 'OTHER') ,
('PREEMPTIVE_OS_FREELIBRARY' , 'OTHER') , ('PREEMPTIVE_OS_GENERICOPS' , 'OTHER') ,
('PREEMPTIVE_OS_GETADDRINFO' , 'OTHER') , ('PREEMPTIVE_OS_GETCOMPRESSEDFILESIZE' , 'OTHER') ,
('PREEMPTIVE_OS_GETDISKFREESPACE' , 'OTHER') , ('PREEMPTIVE_OS_GETFILEATTRIBUTES' , 'OTHER') ,
('PREEMPTIVE_OS_GETFILESIZE' , 'OTHER') , ('PREEMPTIVE_OS_GETLONGPATHNAME' , 'OTHER') ,
('PREEMPTIVE_OS_GETPROCADDRESS' , 'OTHER') , ('PREEMPTIVE_OS_GETVOLUMENAMEFORVOLUMEMOUNTPOINT' , 'OTHER')
, ('PREEMPTIVE_OS_GETVOLUMEPATHNAME' , 'OTHER') , ('PREEMPTIVE_OS_INITIALIZESECURITYCONTEXT' , 'OTHER') ,
('PREEMPTIVE_OS_LIBRARYOPS' , 'OTHER') , ('PREEMPTIVE_OS_LOADLIBRARY' , 'OTHER') ,
('PREEMPTIVE_OS_LOGONUSER' , 'OTHER') , ('PREEMPTIVE_OS_LOOKUPACCOUNTSID' , 'OTHER') ,
('PREEMPTIVE_OS_MESSAGEQUEUEOPS' , 'OTHER') , ('PREEMPTIVE_OS_MOVEFILE' , 'OTHER') ,
('PREEMPTIVE_OS_NETGROUPGETUSERS' , 'OTHER') , ('PREEMPTIVE_OS_NETLOCALGROUPGETMEMBERS' , 'OTHER') ,
('PREEMPTIVE_OS_NETUSERGETGROUPS' , 'OTHER') , ('PREEMPTIVE_OS_NETUSERGETLOCALGROUPS' , 'OTHER') ,
('PREEMPTIVE_OS_NETUSERMODALSGET' , 'OTHER') , ('PREEMPTIVE_OS_NETVALIDATEPASSWORDPOLICY' , 'OTHER') ,
('PREEMPTIVE_OS_NETVALIDATEPASSWORDPOLICYFREE' , 'OTHER') , ('PREEMPTIVE_OS_OPENDIRECTORY' , 'OTHER') ,
('PREEMPTIVE_OS_PIPEOPS' , 'OTHER') , ('PREEMPTIVE_OS_PROCESSOPS' , 'OTHER') ,
('PREEMPTIVE_OS_QUERYREGISTRY' , 'OTHER') , ('PREEMPTIVE_OS_QUERYSECURITYCONTEXTTOKEN' , 'OTHER') ,
('PREEMPTIVE_OS_REMOVEDIRECTORY' , 'OTHER') , ('PREEMPTIVE_OS_REPORTEVENT' , 'OTHER') ,
('PREEMPTIVE_OS_REVERTTOSELF' , 'OTHER') , ('PREEMPTIVE_OS_RSFXDEVICEOPS' , 'OTHER') ,
('PREEMPTIVE_OS_SECURITYOPS' , 'OTHER') , ('PREEMPTIVE_OS_SERVICEOPS' , 'OTHER') ,
('PREEMPTIVE_OS_SETENDOFFILE' , 'OTHER') , ('PREEMPTIVE_OS_SETFILEPOINTER' , 'OTHER') ,
('PREEMPTIVE_OS_SETFILEVALIDDATA' , 'OTHER') , ('PREEMPTIVE_OS_SETNAMEDSECURITYINFO' , 'OTHER') ,
('PREEMPTIVE_OS_SQLCLROPS' , 'OTHER') , ('PREEMPTIVE_OS_SQMLAUNCH' , 'OTHER') ,
('PREEMPTIVE_OS_VERIFYSIGNATURE' , 'OTHER') , ('PREEMPTIVE_OS_VSSOPS' , 'OTHER') ,
('PREEMPTIVE_OS_WAITFORSINGLEOBJECT' , 'OTHER') , ('PREEMPTIVE_OS_WINSOCKOPS' , 'OTHER') ,
('PREEMPTIVE_OS_WRITEFILE' , 'OTHER') , ('PREEMPTIVE_OS_WRITEFILEGATHER' , 'OTHER') ,
('PREEMPTIVE_OS_WSASETLASTERROR' , 'OTHER') , ('PREEMPTIVE_REENLIST' , 'OTHER') , ('PREEMPTIVE_RESIZELOG'
, 'OTHER') , ('PREEMPTIVE_ROLLFORWARDREDO' , 'OTHER') , ('PREEMPTIVE_ROLLFORWARDUNDO' , 'OTHER') ,
('PREEMPTIVE_SB_STOPENDPOINT' , 'OTHER') , ('PREEMPTIVE_SERVER_STARTUP' , 'OTHER') ,
('PREEMPTIVE_SETRMINFO' , 'OTHER') , ('PREEMPTIVE_SHAREDMEM_GETDATA' , 'OTHER') , ('PREEMPTIVE_SNIOPEN' ,
'OTHER') , ('PREEMPTIVE_SOSHOST' , 'OTHER') , ('PREEMPTIVE_SOSTESTING' , 'OTHER') , ('PREEMPTIVE_STARTRM'
, 'OTHER') , ('PREEMPTIVE_STREAMFCB_CHECKPOINT' , 'OTHER') , ('PREEMPTIVE_STREAMFCB_RECOVER' , 'OTHER') ,
('PREEMPTIVE_STRESSDRIVER' , 'OTHER') , ('PREEMPTIVE_TESTING' , 'OTHER') , ('PREEMPTIVE_TRANSIMPORT' ,
'OTHER') , ('PREEMPTIVE_UNMARSHALPROPAGATIONTOKEN' , 'OTHER') , ('PREEMPTIVE_VSS_CREATESNAPSHOT' ,
'OTHER') , ('PREEMPTIVE_VSS_CREATEVOLUMESNAPSHOT' , 'OTHER') , ('PREEMPTIVE_XE_CALLBACKEXECUTE' , 'OTHER')
, ('PREEMPTIVE_XE_DISPATCHER' , 'OTHER') , ('PREEMPTIVE_XE_ENGINEINIT' , 'OTHER') ,
('PREEMPTIVE_XE_GETTARGETSTATE' , 'OTHER') , ('PREEMPTIVE_XE_SESSIONCOMMIT' , 'OTHER') ,
('PREEMPTIVE_XE_TARGETFINALIZE' , 'OTHER') , ('PREEMPTIVE_XE_TARGETINIT' , 'OTHER') ,
('PREEMPTIVE_XE_TIMERRUN' , 'OTHER') , ('PREEMPTIVE_XETESTING' , 'OTHER') , ('PREEMPTIVE_XXX' , 'OTHER') ,
('PRINT_ROLLBACK_PROGRESS' , 'OTHER') , ('QNMANAGER_ACQUIRE' , 'OTHER') , ('QPJOB_KILL' , 'OTHER') ,
('QPJOB_WAITFOR_ABORT' , 'OTHER') , ('QRY_MEM_GRANT_INFO_MUTEX' , 'OTHER') , ('QUERY_ERRHDL_SERVICE_DONE'
, 'OTHER') , ('QUERY_EXECUTION_INDEX_SORT_EVENT_OPEN' , 'OTHER') , ('QUERY_NOTIFICATION_MGR_MUTEX' ,
'OTHER') , ('QUERY_NOTIFICATION_SUBSCRIPTION_MUTEX' , 'OTHER') , ('QUERY_NOTIFICATION_TABLE_MGR_MUTEX' ,
'OTHER') , ('QUERY_NOTIFICATION_UNITTEST_MUTEX' , 'OTHER') , ('QUERY_OPTIMIZER_PRINT_MUTEX' , 'OTHER') ,
('QUERY_TRACEOUT' , 'OTHER') , ('QUERY_WAIT_ERRHDL_SERVICE' , 'OTHER') , ('RECOVER_CHANGEDB' , 'OTHER') ,
('REPL_CACHE_ACCESS' , 'REPLICATION') , ('REPL_HISTORYCACHE_ACCESS' , 'OTHER') , ('REPL_SCHEMA_ACCESS' ,
'OTHER') , ('REPL_TRANHASHTABLE_ACCESS' , 'OTHER') , ('REPLICA_WRITES' , 'OTHER') ,
('REQUEST_DISPENSER_PAUSE' , 'BACKUP') , ('REQUEST_FOR_DEADLOCK_SEARCH' , 'LOCK') , ('RESMGR_THROTTLED' ,
'OTHER') , ('RESOURCE_QUERY_SEMAPHORE_COMPILE' , 'QUERY') , ('RESOURCE_QUEUE' , 'OTHER') ,
('RESOURCE_SEMAPHORE' , 'OTHER') , ('RESOURCE_SEMAPHORE_MUTEX' , 'MEMORY') ,
('RESOURCE_SEMAPHORE_QUERY_COMPILE' , 'MEMORY') , ('RESOURCE_SEMAPHORE_SMALL_QUERY' , 'MEMORY') ,
('RG_RECONFIG' , 'OTHER') , ('SEC_DROP_TEMP_KEY' , 'SECURITY') , ('SECURITY_MUTEX' , 'OTHER') ,
('SEQUENTIAL_GUID' , 'OTHER') , ('SERVER_IDLE_CHECK' , 'OTHER') , ('SHUTDOWN' , 'OTHER') ,
('SLEEP_BPOOL_FLUSH' , 'OTHER') , ('SLEEP_DBSTARTUP' , 'OTHER') , ('SLEEP_DCOMSTARTUP' , 'OTHER') ,
('SLEEP_MSDBSTARTUP' , 'OTHER') , ('SLEEP_SYSTEMTASK' , 'OTHER') , ('SLEEP_TASK' , 'OTHER') ,
('SLEEP_TEMPDBSTARTUP' , 'OTHER') , ('SNI_CRITICAL_SECTION' , 'OTHER') , ('SNI_HTTP_ACCEPT' , 'OTHER') ,
('SNI_HTTP_WAITFOR_0_DISCON' , 'OTHER') , ('SNI_LISTENER_ACCESS' , 'OTHER') , ('SNI_TASK_COMPLETION' ,
'OTHER') , ('SOAP_READ' , 'OTHER') , ('SOAP_WRITE' , 'OTHER') , ('SOS_CALLBACK_REMOVAL' , 'OTHER') ,
('SOS_DISPATCHER_MUTEX' , 'OTHER') , ('SOS_LOCALALLOCATORLIST' , 'OTHER') , ('SOS_MEMORY_USAGE_ADJUSTMENT'
, 'OTHER') , ('SOS_OBJECT_STORE_DESTROY_MUTEX' , 'OTHER') , ('SOS_PROCESS_AFFINITY_MUTEX' , 'OTHER') ,
('SOS_RESERVEDMEMBLOCKLIST' , 'OTHER') , ('SOS_SCHEDULER_YIELD' , 'SQLOS') , ('SOS_SMALL_PAGE_ALLOC' ,
'OTHER') , ('SOS_STACKSTORE_INIT_MUTEX' , 'OTHER') , ('SOS_SYNC_TASK_ENQUEUE_EVENT' , 'OTHER') ,
('SOS_VIRTUALMEMORY_LOW' , 'OTHER') , ('SOSHOST_EVENT' , 'CLR') , ('SOSHOST_OTHER' , 'CLR') ,
('SOSHOST_MUTEX' , 'CLR') , ('SOSHOST_ROWLOCK' , 'CLR') , ('SOSHOST_RWLOCK' , 'CLR') ,
('SOSHOST_SEMAPHORE' , 'CLR') , ('SOSHOST_SLEEP' , 'CLR') , ('SOSHOST_TRACELOCK' , 'CLR') ,
('SOSHOST_WAITFORDONE' , 'CLR') , ('SQLCLR_APPDOMAIN' , 'CLR') , ('SQLCLR_ASSEMBLY' , 'CLR') ,
('SQLCLR_DEADLOCK_DETECTION' , 'CLR') , ('SQLCLR_QUANTUM_PUNISHMENT' , 'CLR') , ('SQLSORT_NORMMUTEX' ,
'OTHER') , ('SQLSORT_SORTMUTEX' , 'OTHER') , ('SQLTRACE_BUFFER_FLUSH ' , 'TRACE') , ('SQLTRACE_LOCK' ,
'OTHER') , ('SQLTRACE_SHUTDOWN' , 'OTHER') , ('SQLTRACE_WAIT_ENTRIES' , 'OTHER') , ('SRVPROC_SHUTDOWN' ,
'OTHER') , ('TEMPOBJ' , 'OTHER') , ('THREADPOOL' , 'SQLOS') , ('TIMEPRIV_TIMEPERIOD' , 'OTHER') ,
('TRACE_EVTNOTIF' , 'OTHER') , ('TRACEWRITE' , 'OTHER') , ('TRAN_MARKLATCH_DT' , 'TRAN_MARKLATCH') ,
('TRAN_MARKLATCH_EX' , 'TRAN_MARKLATCH') , ('TRAN_MARKLATCH_KP' , 'TRAN_MARKLATCH') , ('TRAN_MARKLATCH_NL'
, 'TRAN_MARKLATCH') , ('TRAN_MARKLATCH_SH' , 'TRAN_MARKLATCH') , ('TRAN_MARKLATCH_UP' , 'TRAN_MARKLATCH')
, ('OTHER_MUTEX' , 'OTHER') , ('UTIL_PAGE_ALLOC' , 'OTHER') , ('VIA_ACCEPT' , 'OTHER') ,
('VIEW_DEFINITION_MUTEX' , 'OTHER') , ('WAIT_FOR_RESULTS' , 'OTHER') , ('WAITFOR' , 'WAITFOR') ,
('WAITFOR_TASKSHUTDOWN' , 'OTHER') , ('WAITSTAT_MUTEX' , 'OTHER') , ('WCC' , 'OTHER') , ('WORKTBL_DROP' ,
'OTHER') , ('WRITE_COMPLETION' , 'OTHER') , ('WRITELOG' , 'I/O') , ('XACT_OWN_OTHER' , 'OTHER') ,
('XACT_RECLAIM_SESSION' , 'OTHER') , ('XACTLOCKINFO' , 'OTHER') , ('XACTWORKSPACE_MUTEX' , 'OTHER') ,
('XE_BUFFERMGR_ALLPROCESSED_EVENT' , 'XEVENT') , ('XE_BUFFERMGR_FREEBUF_EVENT' , 'XEVENT') ,
('XE_DISPATCHER_CONFIG_SESSION_LIST' , 'XEVENT') , ('XE_DISPATCHER_JOIN' , 'XEVENT') ,
('XE_DISPATCHER_WAIT' , 'XEVENT') , ('XE_MODULEMGR_SYNC' , 'XEVENT') , ('XE_OLS_LOCK' , 'XEVENT') ,
('XE_PACKAGE_LOCK_BACKOFF' , 'XEVENT') , ('XE_SERVICES_EVENTMANUAL' , 'XEVENT') , ('XE_SERVICES_MUTEX' ,
'XEVENT') , ('XE_SERVICES_RWLOCK' , 'XEVENT') , ('XE_SESSION_CREATE_SYNC' , 'XEVENT') ,
('XE_SESSION_FLUSH' , 'XEVENT') , ('XE_SESSION_SYNC' , 'XEVENT') , ('XE_STM_CREATE' , 'XEVENT') ,
('XE_TIMER_EVENT' , 'XEVENT') , ('XE_TIMER_MUTEX' , 'XEVENT')
, ('XE_TIMER_TASK_DONE' , 'XEVENT');


INSERT @w1 (WaitType, WaitTimeInMs, WaitTaskCount, CollectionDate)
SELECT
  WaitType = wait_type  collate SQL_Latin1_General_CP1_CI_AS
, WaitTimeInMs = SUM(wait_time_ms)
, WaitTaskCount = SUM(waiting_tasks_count)
, CollectionDate = GETDATE()
FROM sys.dm_os_wait_stats
WHERE [wait_type]  collate SQL_Latin1_General_CP1_CI_AS NOT IN
(
	SELECT WaitType FROM  @w3
)
AND [waiting_tasks_count] > 0
GROUP BY wait_type

WAITFOR DELAY @delayInterval;

INSERT @w2 (WaitType, WaitTimeInMs, WaitTaskCount, CollectionDate)
SELECT
  WaitType = wait_type  collate SQL_Latin1_General_CP1_CI_AS
, WaitTimeInMs = SUM(wait_time_ms)
, WaitTaskCount = SUM(waiting_tasks_count)
, CollectionDate = GETDATE()
FROM sys.dm_os_wait_stats
WHERE [wait_type]  collate SQL_Latin1_General_CP1_CI_AS NOT IN
(
	SELECT WaitType FROM  @w3
)
AND [waiting_tasks_count] > 0
GROUP BY wait_type;


INSERT @w5 (WaitCategory, WaitTimeInMs, WaitTaskCount)
SELECT WaitCategory
, WaitTimeInMs = SUM(WaitTimeInMs)
, WaitTaskCount = SUM(WaitTaskCount)
FROM
(
SELECT
  WaitCategory = ISNULL(T4.WaitCategory, 'OTHER')
, WaitTimeInMs = (T2.WaitTimeInMs - T1.WaitTimeInMs)
, WaitTaskCount = (T2.WaitTaskCount - T1.WaitTaskCount)
--, WaitTimeInMsPerSec = ((T2.WaitTimeInMs - T1.WaitTimeInMs) / CAST(DATEDIFF(SECOND, T1.CollectionDate, T2.CollectionDate) as float))
FROM @w1 T1
INNER JOIN @w2 T2 ON T2.WaitType = T1.WaitType
LEFT JOIN @w4 T4 ON T4.WaitType = T1.WaitType
WHERE T2.WaitTaskCount - T1.WaitTaskCount > 0
) as G
GROUP BY G.WaitCategory;



SELECT
---- measurement
  measurement = 'Wait time (ms)'
---- tags
, servername= REPLACE(@@SERVERNAME, '\', ':')
, type = 'Wait stats'
---- values
, [I/O] = SUM([I/O])
, [Latch] = SUM([LATCH])
, [Lock] = SUM([LOCK])
, [Network] = SUM([NETWORK])
, [Service broker] = SUM([SERVICE BROKER])
, [Memory] = SUM([MEMORY])
, [Buffer] = SUM([BUFFER])
, [CLR] = SUM([CLR])
, [SQLOS] = SUM([SQLOS])
, [XEvent] = SUM([XEVENT])
, [Other] = SUM([OTHER])
, [Total] = SUM([I/O]+[LATCH]+[LOCK]+[NETWORK]+[SERVICE BROKER]+[MEMORY]+[BUFFER]+[CLR]+[XEVENT]+[SQLOS]+[OTHER])
FROM
(
SELECT
  [I/O] = ISNULL([I/O] , 0)
, [MEMORY] = ISNULL([MEMORY] , 0)
, [BUFFER] = ISNULL([BUFFER] , 0)
, [LATCH] = ISNULL([LATCH] , 0)
, [LOCK] = ISNULL([LOCK] , 0)
, [NETWORK] = ISNULL([NETWORK] , 0)
, [SERVICE BROKER] = ISNULL([SERVICE BROKER] , 0)
, [CLR] = ISNULL([CLR] , 0)
, [XEVENT] = ISNULL([XEVENT] , 0)
, [SQLOS] = ISNULL([SQLOS] , 0)
, [OTHER] = ISNULL([OTHER] , 0)
FROM @w5 as P
PIVOT
(
	SUM(WaitTimeInMs)
	FOR WaitCategory IN ([I/O], [LATCH], [LOCK], [NETWORK], [SERVICE BROKER], [MEMORY], [BUFFER], [CLR], [XEVENT], [SQLOS], [OTHER])
) AS PivotTable
) as T

UNION ALL

SELECT
---- measurement
  measurement = 'Wait tasks'
---- tags
, server_name= REPLACE(@@SERVERNAME, '\', ':')
, type = 'Wait stats'
---- values
, [I/O] = SUM([I/O])
, [Latch] = SUM([LATCH])
, [Lock] = SUM([LOCK])
, [Network] = SUM([NETWORK])
, [Service broker] = SUM([SERVICE BROKER])
, [Memory] = SUM([MEMORY])
, [Buffer] = SUM([BUFFER])
, [CLR] = SUM([CLR])
, [SQLOS] = SUM([SQLOS])
, [XEvent] = SUM([XEVENT])
, [Other] = SUM([OTHER])
, [Total] = SUM([I/O]+[LATCH]+[LOCK]+[NETWORK]+[SERVICE BROKER]+[MEMORY]+[BUFFER]+[CLR]+[XEVENT]+[SQLOS]+[OTHER])
FROM
(
SELECT
  [I/O] = ISNULL([I/O] , 0)
, [MEMORY] = ISNULL([MEMORY] , 0)
, [BUFFER] = ISNULL([BUFFER] , 0)
, [LATCH] = ISNULL([LATCH] , 0)
, [LOCK] = ISNULL([LOCK] , 0)
, [NETWORK] = ISNULL([NETWORK] , 0)
, [SERVICE BROKER] = ISNULL([SERVICE BROKER] , 0)
, [CLR] = ISNULL([CLR] , 0)
, [XEVENT] = ISNULL([XEVENT] , 0)
, [SQLOS] = ISNULL([SQLOS] , 0)
, [OTHER] = ISNULL([OTHER] , 0)
FROM @w5 as P
PIVOT
(
	SUM(WaitTaskCount)
	FOR WaitCategory IN ([I/O], [LATCH], [LOCK], [NETWORK], [SERVICE BROKER], [MEMORY], [BUFFER], [CLR], [XEVENT], [SQLOS], [OTHER])
) AS PivotTable
) as T;
`

const sqlVolumeSpace string = `SET DEADLOCK_PRIORITY -10;
SET NOCOUNT ON;
SET TRANSACTION ISOLATION LEVEL READ UNCOMMITTED;

IF OBJECT_ID('tempdb..#volumestats') IS NOT NULL
	DROP TABLE #volumestats;
SELECT DISTINCT
  volume =  REPLACE(vs.volume_mount_point, '\', '')
	+ CASE WHEN LEN(vs.logical_volume_name) > 0
		THEN ' (' + vs.logical_volume_name + ')'
		ELSE '' END
, total_bytes = vs.total_bytes
, available_bytes = vs.available_bytes
, used_bytes = vs.total_bytes - vs.available_bytes
, used_percent = 100 * CAST(ROUND((vs.total_bytes - vs.available_bytes) * 1. / vs.total_bytes, 2) as decimal(5,2))
INTO #volumestats
FROM sys.master_files AS f
CROSS APPLY sys.dm_os_volume_stats(f.database_id, f.file_id) vs

DECLARE @DynamicPivotQuery AS NVARCHAR(MAX)
DECLARE @ColumnName AS NVARCHAR(MAX), @ColumnName2 AS NVARCHAR(MAX)

SELECT @ColumnName= ISNULL(@ColumnName + ',','') + QUOTENAME(volume)
FROM (SELECT DISTINCT volume FROM #volumestats) AS bl

--Prepare the PIVOT query using the dynamic
SET @DynamicPivotQuery = N'
SELECT measurement = ''Volume total space (bytes)'', servername = REPLACE(@@SERVERNAME, ''\'', '':''), type = ''OS Volume space''
, ' + @ColumnName + '  FROM
(
SELECT volume,  total_bytes
FROM #volumestats
) as V
PIVOT(SUM(total_bytes) FOR volume IN (' + @ColumnName + ')) AS PVTTable

UNION ALL

SELECT measurement = ''Volume available space (bytes)'', servername = REPLACE(@@SERVERNAME, ''\'', '':''), type = ''OS Volume space''
, ' + @ColumnName + '  FROM
(
SELECT volume,  available_bytes
FROM #volumestats
) as V
PIVOT(SUM(available_bytes) FOR volume IN (' + @ColumnName + ')) AS PVTTable

UNION ALL

SELECT measurement = ''Volume used space (bytes)'', servername = REPLACE(@@SERVERNAME, ''\'', '':''), type = ''OS Volume space''
, ' + @ColumnName + '  FROM
(
SELECT volume,  used_bytes
FROM #volumestats
) as V
PIVOT(SUM(used_bytes) FOR volume IN (' + @ColumnName + ')) AS PVTTable

UNION ALL

SELECT measurement = ''Volume used space (%)'', servername = REPLACE(@@SERVERNAME, ''\'', '':''), type = ''OS Volume space''
, ' + @ColumnName + '  FROM
(
SELECT volume,  used_percent
FROM #volumestats
) as V
PIVOT(SUM(used_percent) FOR volume IN (' + @ColumnName + ')) AS PVTTable'

EXEC sp_executesql @DynamicPivotQuery;
`
