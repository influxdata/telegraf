# SQL Server plugin

This sqlserver plugin provides metrics for your SQL Server instance. 
It currently works with SQL Server versions 2008+. 
Recorded metrics are lightweight and use Dynamic Management Views supplied by SQL Server:
```
Performance counters  : 1000+ metrics from sys.dm_os_performance_counters
Performance metrics   : some special performance metrics
Wait stats 			  : list of wait tasks categorized from sys.dm_os_wait_stats
Memory clerk		  : memory breakdown from sys.dm_os_memory_clerks
Database size         : database size trend, data and log file from sys.dm_io_virtual_file_stats
Database IO			  : database I/O from sys.dm_io_virtual_file_stats
Database latency	  : database reads and writes latency from sys.dm_io_virtual_file_stats
CPU				      : cpu usage from sys.dm_os_ring_buffers
```

You must create a login on every instance you want to monitor, with following script:
```SQL 
USE master; 
GO
CREATE LOGIN [telegraf] WITH PASSWORD = N'mystrongpassword';
GO
GRANT VIEW SERVER STATE TO [telegraf]; 
GO
GRANT VIEW ANY DEFINITION TO [telegraf]; 
GO
```
