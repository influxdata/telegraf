# SQL Server plugin

This sqlserver plugin provides metrics for your SQL Server instance. 
It currently works with SQL Server versions 2008+. 
Recorded metrics are lightweight and use Dynamic Management Views supplied by SQL Server:
```
PerformanceCounters  : 1000+ metrics from sys.dm_os_performance_counters
PerformanceMetrics   : some special performance metrics
WaitStatsCategorized : list of wait tasks categorized from sys.dm_os_wait_stats
MemoryClerk			 : memory breakdown from sys.dm_os_memory_clerks
DatabaseSize         : database size trend, datafile and logfile from sys.dm_io_virtual_file_stats
DatabaseIO			 : database I/O from sys.dm_io_virtual_file_stats
CPUHistory			 : cpu usage from sys.dm_os_ring_buffers
```

You must create a login on every instance you want to monitor, with following script:
	```SQL 
	USE master; 
	GO
	CREATE LOGIN [linuxuser] WITH PASSWORD = N'mystrongpassword';
	GO
	GRANT VIEW SERVER STATE TO [linuxuser]; 
	GO
	GRANT VIEW ANY DEFINITION TO [linuxuser]; 
	GO
