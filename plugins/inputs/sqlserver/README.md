# SQL Server plugin

This sqlserver plugin provides metrics for your SQL Server instance. 
It currently works with SQL Server versions 2008+. 
Recorded metrics are lightweight and use Dynamic Management Views supplied by SQL Server:
```
Performance counters  : 1000+ metrics from sys.dm_os_performance_counters
Performance metrics   : special performance and ratio metrics
Wait stats 			  : wait tasks categorized from sys.dm_os_wait_stats
Memory clerk		  : memory breakdown from sys.dm_os_memory_clerks
Database size         : databases size trend from sys.dm_io_virtual_file_stats
Database IO			  : databases I/O from sys.dm_io_virtual_file_stats
Database latency	  : databases latency from sys.dm_io_virtual_file_stats
Database properties   : databases properties, state and recovery model, from sys.databases
OS Volume             : available, used and total space from sys.dm_os_volume_stats
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

Overview
![telegraf-sqlserver-0](https://cloud.githubusercontent.com/assets/16494280/12538189/ec1b70aa-c2d3-11e5-97ec-1a4f575e8a07.png)

General Activity
![telegraf-sqlserver-1](https://cloud.githubusercontent.com/assets/16494280/12591410/f098b602-c467-11e5-9acf-2edea077ed7e.png)

Memory
![telegraf-sqlserver-2](https://cloud.githubusercontent.com/assets/16494280/12591412/f2075688-c467-11e5-9d0f-d256e032cd0e.png)

I/O
![telegraf-sqlserver-3](https://cloud.githubusercontent.com/assets/16494280/12591417/f40ccb84-c467-11e5-89ff-498fb1bc3110.png)

Disks
![telegraf-sqlserver-4](https://cloud.githubusercontent.com/assets/16494280/12591420/f5de5f68-c467-11e5-90c8-9185444ac490.png)

CPU
![telegraf-sqlserver-5](https://cloud.githubusercontent.com/assets/16494280/12591446/11dfe7b8-c468-11e5-9681-6e33296e70e8.png)

Full view
![telegraf-sqlserver-full](https://cloud.githubusercontent.com/assets/16494280/12591426/fa2b17b4-c467-11e5-9c00-929f4c4aea57.png)
