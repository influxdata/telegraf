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

## Getting started :

You have to create a login on every instance you want to monitor, with following script:
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


## Configuration:

``` 
# Read metrics from Microsoft SQL Server
[[inputs.sqlserver]]
  # Specify instances to monitor with a list of connection strings.
  # All connection parameters are optional. 
  # By default, the host is localhost, listening on default port (TCP/1433) 
  #    for Windows, the user is the currently running AD user (SSO).
  #    See https://github.com/denisenkom/go-mssqldb for detailed connection parameters.

  servers = [
	"Server=192.168.1.30;Port=1433;User Id=telegraf;Password=T$l$gr@f69*;app name=telegraf;log=1;",
    "Server=192.168.1.30;Port=2222;User Id=telegraf;Password=T$l$gr@f69*;app name=telegraf;log=1;"
	]
```


## Measurement | Fields:

- Wait stats 
	- Wait time (ms) | I/O, Latch, Lock, Network, Service broker, Memory, Buffer, CLR, XEvent, Other, Total
	- Wait tasks | I/O, Latch, Lock, Network, Service broker, Memory, Buffer, CLR, XEvent, Other, Total
- Memory clerk	
	- Memory breakdown (%) | Buffer pool, Cache (objects), Cache (sql plans), Other
	- Memory breakdown (bytes) | Buffer pool, Cache (objects), Cache (sql plans), Other
- Database size 
	- Log size (bytes) | databases (included sysdb)
	- Rows size (bytes) | databases (included sysdb)
- Database IO	
	- Log writes (bytes/sec) | databases (included sysdb)
	- Rows writes (bytes/sec) | databases (included sysdb)
	- Log reads (bytes/sec) | databases (included sysdb)
	- Rows reads (bytes/sec) | databases (included sysdb)
	- Log (writes/sec) | databases (included sysdb)
	- Rows (writes/sec) | databases (included sysdb)
	- Log (reads/sec) | databases (included sysdb)
	- Rows (reads/sec) | databases (included sysdb)
- Database latency	 
 	- Log read latency (ms) | databases (included sysdb)
 	- Log write latency (ms) | databases (included sysdb)
 	- Rows read latency (ms) | databases (included sysdb)
 	- Rows write latency (ms) | databases (included sysdb)
	- Log (average bytes/read) | databases (included sysdb)
	- Log (average bytes/write) | databases (included sysdb)
	- Rows (average bytes/read) | databases (included sysdb)
	- Rows (average bytes/write) | databases (included sysdb)
- Database properties  
	- Recovery Model FULL | databases (included sysdb)
	- Recovery Model BULK_LOGGED | databases (included sysdb)
	- Recovery Model SIMPLE | databases (included sysdb)
	- State ONLINE | databases (included sysdb)
	- State RESTORING | databases (included sysdb)
	- State RECOVERING | databases (included sysdb)
	- State RECOVERY_PENDING | databases (included sysdb)
	- State SUSPECT | databases (included sysdb)
	- State EMERGENCY | databases (included sysdb)
	- State OFFLINE | databases (included sysdb)
- OS Volume    
	- Volume total space (bytes) | logical volumes 
	- Volume available space (bytes) | logical volumes 
	- Volume used space (bytes) | logical volumes
	- Volume used space (%) | logical volumes
- CPU	
	- CPU (%) | SQL process, External process, SystemIdle
- Performance metrics
	- Performance metrics | Point In Time Recovery, Available physical memory (bytes), Average pending disk IO, Average runnable tasks, Average tasks, Buffer pool rate (bytes/sec), Connection memory per connection (bytes), Memory grant pending, Page File Usage (%), Page lookup per batch request, Page split per batch request, Readahead per page read, Signal wait (%), Sql compilation per batch request, Sql recompilation per batch request, Total target memory ratio
- Performance counters
	- AU cleanup batches/sec | Value
	- ... 1000+ metrics
	  See https://msdn.microsoft.com/fr-fr/library/ms190382(v=sql.120).aspx

	  
## Tags:
- All stats have the following tags:
	- servername (server name:instance ID)
	- type (type of stats to easily filter measurements)

	
## Overview in Grafana:

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


## Example Output:

``` 
./telegraf --config telegraf.conf --test 
* Plugin: sqlserver, Collection 1
> Memory\ breakdown\ (%),servername=WIN8-DEV,type=Memory\ clerk Buffer\ pool=27.20,Cache\ (objects)=6.50,Cache\ (sql\ plans)=31.50,Other=34.80 1453876411474582877
> Memory\ breakdown\ (bytes),servername=WIN8-DEV,type=Memory\ clerk Buffer\ pool=100016128.00,Cache\ (objects)=23904256.00,Cache\ (sql\ plans)=115621888.00,Other=127942656.00 1453876411474655779
> Log\ size\ (bytes),servername=WIN8-DEV,type=Database\ size AdventureWorks2014=538968064i,Australian=1048576i,DOC.Azure=786432i,ResumeCloud=786432i,master=2359296i,model=4325376i,msdb=30212096i,ngMon=1048576i,tempdb=4194304i 1453876411497506186
> Rows\ size\ (bytes),servername=WIN8-DEV,type=Database\ size AdventureWorks2014=2362703872i,Australian=3211264i,DOC.Azure=26083328i,ResumeCloud=10551296i,master=5111808i,model=3342336i,msdb=24051712i,ngMon=46137344i,tempdb=1073741824i 1453876411497557587
> CPU\ (%),servername=WIN8-DEV,type=CPU\ usage External\ process=1i,SQL\ process=0i,SystemIdle=99i 1453876411546452385
> Log\ read\ latency\ (ms),servername=WIN8-DEV,type=Database\ stats AdventureWorks2014=28i,Australian=7i,DOC.Azure=2i,ResumeCloud=3i,master=13i,model=35i,msdb=33i,ngMon=12i,tempdb=2i 1453876411559643935
> Log\ write\ latency\ (ms),servername=WIN8-DEV,type=Database\ stats AdventureWorks2014=0i,Australian=0i,DOC.Azure=0i,ResumeCloud=0i,master=1i,model=0i,msdb=0i,ngMon=0i,tempdb=15i 1453876411559700937
> Rows\ read\ latency\ (ms),servername=WIN8-DEV,type=Database\ stats AdventureWorks2014=39i,Australian=18i,DOC.Azure=39i,ResumeCloud=48i,master=26i,model=9i,msdb=29i,ngMon=42i,tempdb=63i 1453876411559729338
> Rows\ write\ latency\ (ms),servername=WIN8-DEV,type=Database\ stats AdventureWorks2014=0i,Australian=0i,DOC.Azure=0i,ResumeCloud=0i,master=7i,model=0i,msdb=5i,ngMon=0i,tempdb=3i 1453876411559767639
> Rows\ (average\ bytes/read),servername=WIN8-DEV,type=Database\ stats AdventureWorks2014=786310i,Australian=62968i,DOC.Azure=515034i,ResumeCloud=203113i,master=63126i,model=60503i,msdb=56555i,ngMon=156254i,tempdb=148091i 1453876411559790439
> Rows\ (average\ bytes/write),servername=WIN8-DEV,type=Database\ stats AdventureWorks2014=9011i,Australian=8192i,DOC.Azure=8192i,ResumeCloud=8192i,master=8192i,model=8192i,msdb=8960i,ngMon=8192i,tempdb=32768i 1453876411559815940
> Log\ (average\ bytes/read),servername=WIN8-DEV,type=Database\ stats AdventureWorks2014=66560i,Australian=57856i,DOC.Azure=13312i,ResumeCloud=28086i,master=39384i,model=13653i,msdb=29857i,ngMon=50688i,tempdb=143945i 1453876411559838741
> Log\ (average\ bytes/write),servername=WIN8-DEV,type=Database\ stats AdventureWorks2014=4949i,Australian=5213i,DOC.Azure=6264i,ResumeCloud=5213i,master=4256i,model=5120i,msdb=5412i,ngMon=5120i,tempdb=55033i 1453876411559861941
> Performance\ metrics,servername=WIN8-DEV,type=Performance\ metrics Available\ physical\ memory\ (bytes)=5708369920i,Average\ pending\ disk\ IO=0i,Average\ runnable\ tasks=0i,Average\ tasks=9i,Buffer\ pool\ rate\ (bytes/sec)=2194i,Connection\ memory\ per\ connection\ (bytes)=120832i,Memory\ grant\ pending=0i,Page\ File\ Usage\ (%)=32i,Page\ lookup\ per\ batch\ request=162800i,Page\ split\ per\ batch\ request=377i,Point\ In\ Time\ Recovery=0i,Readahead\ per\ page\ read=19i,Signal\ wait\ (%)=18i,Sql\ compilation\ per\ batch\ request=501i,Sql\ recompilation\ per\ batch\ request=230i,Total\ target\ memory\ ratio=27i 1453876411621085067
> Volume\ total\ space\ (bytes),servername=WIN8-DEV,type=OS\ Volume\ space C:=135338651648.00,D:\ (DATA)=32075874304.00,L:\ (LOG)=10701701120.00 1453876411645554817
> Volume\ available\ space\ (bytes),servername=WIN8-DEV,type=OS\ Volume\ space C:=59092901888.00,D:\ (DATA)=28439674880.00,L:\ (LOG)=10107617280.00 1453876411645599018
> Volume\ used\ space\ (bytes),servername=WIN8-DEV,type=OS\ Volume\ space C:=76245749760.00,D:\ (DATA)=3636199424.00,L:\ (LOG)=594083840.00 1453876411645619518
> Volume\ used\ space\ (%),servername=WIN8-DEV,type=OS\ Volume\ space C:=56.00,D:\ (DATA)=11.00,L:\ (LOG)=6.00 1453876411645639419
> Recovery\ Model\ FULL,servername=WIN8-DEV,type=Database\ properties AdventureWorks2014=1i,Australian=0i,DOC.Azure=0i,ResumeCloud=0i,master=0i,model=1i,msdb=0i,ngMon=0i,tempdb=0i,total=2i 1453876411649946833
> Recovery\ Model\ BULK_LOGGED,servername=WIN8-DEV,type=Database\ properties AdventureWorks2014=0i,Australian=0i,DOC.Azure=0i,ResumeCloud=0i,master=0i,model=0i,msdb=0i,ngMon=0i,tempdb=0i,total=0i 1453876411650019335
> Recovery\ Model\ SIMPLE,servername=WIN8-DEV,type=Database\ properties AdventureWorks2014=0i,Australian=1i,DOC.Azure=1i,ResumeCloud=1i,master=1i,model=0i,msdb=1i,ngMon=1i,tempdb=1i,total=7i 1453876411650045636
> State\ ONLINE,servername=WIN8-DEV,type=Database\ properties AdventureWorks2014=1i,Australian=1i,DOC.Azure=1i,ResumeCloud=1i,master=1i,model=1i,msdb=1i,ngMon=1i,tempdb=1i,total=9i 1453876411650071137
> State\ RESTORING,servername=WIN8-DEV,type=Database\ properties AdventureWorks2014=0i,Australian=0i,DOC.Azure=0i,ResumeCloud=0i,master=0i,model=0i,msdb=0i,ngMon=0i,tempdb=0i,total=0i 1453876411650095937
> State\ RECOVERING,servername=WIN8-DEV,type=Database\ properties AdventureWorks2014=0i,Australian=0i,DOC.Azure=0i,ResumeCloud=0i,master=0i,model=0i,msdb=0i,ngMon=0i,tempdb=0i,total=0i 1453876411650119138
> State\ RECOVERY_PENDING,servername=WIN8-DEV,type=Database\ properties AdventureWorks2014=0i,Australian=0i,DOC.Azure=0i,ResumeCloud=0i,master=0i,model=0i,msdb=0i,ngMon=0i,tempdb=0i,total=0i 1453876411650154939
> State\ SUSPECT,servername=WIN8-DEV,type=Database\ properties AdventureWorks2014=0i,Australian=0i,DOC.Azure=0i,ResumeCloud=0i,master=0i,model=0i,msdb=0i,ngMon=0i,tempdb=0i,total=0i 1453876411650177939
> State\ EMERGENCY,servername=WIN8-DEV,type=Database\ properties AdventureWorks2014=0i,Australian=0i,DOC.Azure=0i,ResumeCloud=0i,master=0i,model=0i,msdb=0i,ngMon=0i,tempdb=0i,total=0i 1453876411650201140
> State\ OFFLINE,servername=WIN8-DEV,type=Database\ properties AdventureWorks2014=0i,Australian=0i,DOC.Azure=0i,ResumeCloud=0i,master=0i,model=0i,msdb=0i,ngMon=0i,tempdb=0i,total=0i 1453876411650222841
> AU\ cleanup\ batches/sec\ |\ Access\ Methods,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412515666120
> AU\ cleanups/sec\ |\ Access\ Methods,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412515723222
> By-reference\ Lob\ Create\ Count\ |\ Access\ Methods,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412515741722
> By-reference\ Lob\ Use\ Count\ |\ Access\ Methods,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412515754423
> Count\ Lob\ Readahead\ |\ Access\ Methods,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412515780423
> Count\ Pull\ In\ Row\ |\ Access\ Methods,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412515806724
> Count\ Push\ Off\ Row\ |\ Access\ Methods,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412515818924
> Deferred\ dropped\ AUs\ |\ Access\ Methods,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412515832925
> Deferred\ Dropped\ rowsets\ |\ Access\ Methods,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412515844225
> Dropped\ rowset\ cleanups/sec\ |\ Access\ Methods,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412515856425
> Dropped\ rowsets\ skipped/sec\ |\ Access\ Methods,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412515882726
> Extent\ Deallocations/sec\ |\ Access\ Methods,servername=WIN8-DEV,type=Performance\ counters value=2i 1453876412515902227
> Extents\ Allocated/sec\ |\ Access\ Methods,servername=WIN8-DEV,type=Performance\ counters value=5i 1453876412515947828
> Failed\ AU\ cleanup\ batches/sec\ |\ Access\ Methods,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412515963628
> Failed\ leaf\ page\ cookie\ |\ Access\ Methods,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412515981729
> Failed\ tree\ page\ cookie\ |\ Access\ Methods,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412516030230
> Forwarded\ Records/sec\ |\ Access\ Methods,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412516043030
> FreeSpace\ Page\ Fetches/sec\ |\ Access\ Methods,servername=WIN8-DEV,type=Performance\ counters value=6i 1453876412516065131
> FreeSpace\ Scans/sec\ |\ Access\ Methods,servername=WIN8-DEV,type=Performance\ counters value=6i 1453876412516076631
> Full\ Scans/sec\ |\ Access\ Methods,servername=WIN8-DEV,type=Performance\ counters value=132i 1453876412516088332
> Index\ Searches/sec\ |\ Access\ Methods,servername=WIN8-DEV,type=Performance\ counters value=3387i 1453876412516122932
> InSysXact\ waits/sec\ |\ Access\ Methods,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412516138833
> LobHandle\ Create\ Count\ |\ Access\ Methods,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412516151733
> LobHandle\ Destroy\ Count\ |\ Access\ Methods,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412516163134
> LobSS\ Provider\ Create\ Count\ |\ Access\ Methods,servername=WIN8-DEV,type=Performance\ counters value=5i 1453876412516176834
> LobSS\ Provider\ Destroy\ Count\ |\ Access\ Methods,servername=WIN8-DEV,type=Performance\ counters value=5i 1453876412516237036
> LobSS\ Provider\ Truncation\ Count\ |\ Access\ Methods,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412516249536
> Mixed\ page\ allocations/sec\ |\ Access\ Methods,servername=WIN8-DEV,type=Performance\ counters value=36i 1453876412516261536
> Page\ compression\ attempts/sec\ |\ Access\ Methods,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412516272536
> Page\ Deallocations/sec\ |\ Access\ Methods,servername=WIN8-DEV,type=Performance\ counters value=14i 1453876412516284237
> Page\ Splits/sec\ |\ Access\ Methods,servername=WIN8-DEV,type=Performance\ counters value=24i 1453876412516310037
> Pages\ Allocated/sec\ |\ Access\ Methods,servername=WIN8-DEV,type=Performance\ counters value=50i 1453876412516321338
> Pages\ compressed/sec\ |\ Access\ Methods,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412516333038
> Probe\ Scans/sec\ |\ Access\ Methods,servername=WIN8-DEV,type=Performance\ counters value=724i 1453876412516343338
> Range\ Scans/sec\ |\ Access\ Methods,servername=WIN8-DEV,type=Performance\ counters value=610i 1453876412516356639
> Scan\ Point\ Revalidations/sec\ |\ Access\ Methods,servername=WIN8-DEV,type=Performance\ counters value=4i 1453876412516381539
> Skipped\ Ghosted\ Records/sec\ |\ Access\ Methods,servername=WIN8-DEV,type=Performance\ counters value=9i 1453876412516392940
> Table\ Lock\ Escalations/sec\ |\ Access\ Methods,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412516404740
> Used\ leaf\ page\ cookie\ |\ Access\ Methods,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412516415540
> Used\ tree\ page\ cookie\ |\ Access\ Methods,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412516427241
> Workfiles\ Created/sec\ |\ Access\ Methods,servername=WIN8-DEV,type=Performance\ counters value=144i 1453876412516451441
> Worktables\ Created/sec\ |\ Access\ Methods,servername=WIN8-DEV,type=Performance\ counters value=46i 1453876412516464242
> Worktables\ From\ Cache\ Base\ |\ Access\ Methods,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412516475942
> Worktables\ From\ Cache\ Ratio\ |\ Access\ Methods,servername=WIN8-DEV,type=Performance\ counters value=1i 1453876412516487742
> Bytes\ Received\ from\ Replica/sec\ |\ _Total\ |\ Availability\ Replica,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412516499642
> Bytes\ Sent\ to\ Replica/sec\ |\ _Total\ |\ Availability\ Replica,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412516524443
> Bytes\ Sent\ to\ Transport/sec\ |\ _Total\ |\ Availability\ Replica,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412516537143
> Flow\ Control\ Time\ (ms/sec)\ |\ _Total\ |\ Availability\ Replica,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412516549644
> Flow\ Control/sec\ |\ _Total\ |\ Availability\ Replica,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412516560444
> Receives\ from\ Replica/sec\ |\ _Total\ |\ Availability\ Replica,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412516576045
> Resent\ Messages/sec\ |\ _Total\ |\ Availability\ Replica,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412516624246
> Sends\ to\ Replica/sec\ |\ _Total\ |\ Availability\ Replica,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412516643746
> Sends\ to\ Transport/sec\ |\ _Total\ |\ Availability\ Replica,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412516656347
> Batches\ >=000000ms\ &\ <000001ms\ |\ CPU\ Time:Requests\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=265i 1453876412516666947
> Batches\ >=000000ms\ &\ <000001ms\ |\ CPU\ Time:Total(ms)\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412516678947
> Batches\ >=000000ms\ &\ <000001ms\ |\ Elapsed\ Time:Requests\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=261i 1453876412516719248
> Batches\ >=000000ms\ &\ <000001ms\ |\ Elapsed\ Time:Total(ms)\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412516731249
> Batches\ >=000001ms\ &\ <000002ms\ |\ CPU\ Time:Requests\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=86i 1453876412516745149
> Batches\ >=000001ms\ &\ <000002ms\ |\ CPU\ Time:Total(ms)\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=86i 1453876412516756149
> Batches\ >=000001ms\ &\ <000002ms\ |\ Elapsed\ Time:Requests\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=31i 1453876412516767750
> Batches\ >=000001ms\ &\ <000002ms\ |\ Elapsed\ Time:Total(ms)\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=31i 1453876412516794050
> Batches\ >=000002ms\ &\ <000005ms\ |\ CPU\ Time:Requests\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=116i 1453876412516805351
> Batches\ >=000002ms\ &\ <000005ms\ |\ CPU\ Time:Total(ms)\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=257i 1453876412516817051
> Batches\ >=000002ms\ &\ <000005ms\ |\ Elapsed\ Time:Requests\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=43i 1453876412516830151
> Batches\ >=000002ms\ &\ <000005ms\ |\ Elapsed\ Time:Total(ms)\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=108i 1453876412516842152
> Batches\ >=000005ms\ &\ <000010ms\ |\ CPU\ Time:Requests\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=199i 1453876412516869752
> Batches\ >=000005ms\ &\ <000010ms\ |\ CPU\ Time:Total(ms)\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=1676i 1453876412516881053
> Batches\ >=000005ms\ &\ <000010ms\ |\ Elapsed\ Time:Requests\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=60i 1453876412516918754
> Batches\ >=000005ms\ &\ <000010ms\ |\ Elapsed\ Time:Total(ms)\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=440i 1453876412516935054
> Batches\ >=000010ms\ &\ <000020ms\ |\ CPU\ Time:Requests\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=193i 1453876412516947954
> Batches\ >=000010ms\ &\ <000020ms\ |\ CPU\ Time:Total(ms)\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=2327i 1453876412516976455
> Batches\ >=000010ms\ &\ <000020ms\ |\ Elapsed\ Time:Requests\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=117i 1453876412516987855
> Batches\ >=000010ms\ &\ <000020ms\ |\ Elapsed\ Time:Total(ms)\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=1619i 1453876412516999656
> Batches\ >=000020ms\ &\ <000050ms\ |\ CPU\ Time:Requests\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=586i 1453876412517010356
> Batches\ >=000020ms\ &\ <000050ms\ |\ CPU\ Time:Total(ms)\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=18826i 1453876412517021956
> Batches\ >=000020ms\ &\ <000050ms\ |\ Elapsed\ Time:Requests\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=307i 1453876412517066858
> Batches\ >=000020ms\ &\ <000050ms\ |\ Elapsed\ Time:Total(ms)\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=10144i 1453876412517079258
> Batches\ >=000050ms\ &\ <000100ms\ |\ CPU\ Time:Requests\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=283i 1453876412517095158
> Batches\ >=000050ms\ &\ <000100ms\ |\ CPU\ Time:Total(ms)\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=22015i 1453876412517112259
> Batches\ >=000050ms\ &\ <000100ms\ |\ Elapsed\ Time:Requests\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=234i 1453876412517138159
> Batches\ >=000050ms\ &\ <000100ms\ |\ Elapsed\ Time:Total(ms)\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=16312i 1453876412517169260
> Batches\ >=000100ms\ &\ <000200ms\ |\ CPU\ Time:Requests\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=338i 1453876412517180961
> Batches\ >=000100ms\ &\ <000200ms\ |\ CPU\ Time:Total(ms)\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=54584i 1453876412517192961
> Batches\ >=000100ms\ &\ <000200ms\ |\ Elapsed\ Time:Requests\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=239i 1453876412517203761
> Batches\ >=000100ms\ &\ <000200ms\ |\ Elapsed\ Time:Total(ms)\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=37309i 1453876412517217362
> Batches\ >=000200ms\ &\ <000500ms\ |\ CPU\ Time:Requests\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=101i 1453876412517245062
> Batches\ >=000200ms\ &\ <000500ms\ |\ CPU\ Time:Total(ms)\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=25711i 1453876412517256463
> Batches\ >=000200ms\ &\ <000500ms\ |\ Elapsed\ Time:Requests\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=321i 1453876412517269763
> Batches\ >=000200ms\ &\ <000500ms\ |\ Elapsed\ Time:Total(ms)\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=85623i 1453876412517280463
> Batches\ >=000500ms\ &\ <001000ms\ |\ CPU\ Time:Requests\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=2i 1453876412517292064
> Batches\ >=000500ms\ &\ <001000ms\ |\ CPU\ Time:Total(ms)\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=1270i 1453876412517320064
> Batches\ >=000500ms\ &\ <001000ms\ |\ Elapsed\ Time:Requests\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=10i 1453876412517332465
> Batches\ >=000500ms\ &\ <001000ms\ |\ Elapsed\ Time:Total(ms)\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=6518i 1453876412517344465
> Batches\ >=001000ms\ &\ <002000ms\ |\ CPU\ Time:Requests\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412517354965
> Batches\ >=001000ms\ &\ <002000ms\ |\ CPU\ Time:Total(ms)\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412517366666
> Batches\ >=001000ms\ &\ <002000ms\ |\ Elapsed\ Time:Requests\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=184i 1453876412517406967
> Batches\ >=001000ms\ &\ <002000ms\ |\ Elapsed\ Time:Total(ms)\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=209039i 1453876412517420367
> Batches\ >=002000ms\ &\ <005000ms\ |\ CPU\ Time:Requests\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412517432467
> Batches\ >=002000ms\ &\ <005000ms\ |\ CPU\ Time:Total(ms)\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412517443168
> Batches\ >=002000ms\ &\ <005000ms\ |\ Elapsed\ Time:Requests\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412517458668
> Batches\ >=002000ms\ &\ <005000ms\ |\ Elapsed\ Time:Total(ms)\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412517486269
> Batches\ >=005000ms\ &\ <010000ms\ |\ CPU\ Time:Requests\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412517499269
> Batches\ >=005000ms\ &\ <010000ms\ |\ CPU\ Time:Total(ms)\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412517512669
> Batches\ >=005000ms\ &\ <010000ms\ |\ Elapsed\ Time:Requests\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=360i 1453876412517523870
> Batches\ >=005000ms\ &\ <010000ms\ |\ Elapsed\ Time:Total(ms)\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=1816241i 1453876412517535770
> Batches\ >=010000ms\ &\ <020000ms\ |\ CPU\ Time:Requests\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412517564171
> Batches\ >=010000ms\ &\ <020000ms\ |\ CPU\ Time:Total(ms)\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412517575571
> Batches\ >=010000ms\ &\ <020000ms\ |\ Elapsed\ Time:Requests\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412517587571
> Batches\ >=010000ms\ &\ <020000ms\ |\ Elapsed\ Time:Total(ms)\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412517597872
> Batches\ >=020000ms\ &\ <050000ms\ |\ CPU\ Time:Requests\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=1i 1453876412517609472
> Batches\ >=020000ms\ &\ <050000ms\ |\ CPU\ Time:Total(ms)\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=20954i 1453876412517634873
> Batches\ >=020000ms\ &\ <050000ms\ |\ Elapsed\ Time:Requests\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=1i 1453876412517645673
> Batches\ >=020000ms\ &\ <050000ms\ |\ Elapsed\ Time:Total(ms)\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=28069i 1453876412517657173
> Batches\ >=050000ms\ &\ <100000ms\ |\ CPU\ Time:Requests\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412517667574
> Batches\ >=050000ms\ &\ <100000ms\ |\ CPU\ Time:Total(ms)\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412517679174
> Batches\ >=050000ms\ &\ <100000ms\ |\ Elapsed\ Time:Requests\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412517734775
> Batches\ >=050000ms\ &\ <100000ms\ |\ Elapsed\ Time:Total(ms)\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412517747276
> Batches\ >=100000ms\ |\ CPU\ Time:Requests\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412517759476
> Batches\ >=100000ms\ |\ CPU\ Time:Total(ms)\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412517771476
> Batches\ >=100000ms\ |\ Elapsed\ Time:Requests\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412517783377
> Batches\ >=100000ms\ |\ Elapsed\ Time:Total(ms)\ |\ Batch\ Resp\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412517809077
> Stored\ Procedures\ Invoked/sec\ |\ _Total\ |\ Broker\ Activation,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412517820378
> Task\ Limit\ Reached\ |\ _Total\ |\ Broker\ Activation,servername=WIN8-DEV,type=Performance\ counters value=3i 1453876412517831978
> Task\ Limit\ Reached/sec\ |\ _Total\ |\ Broker\ Activation,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412517843478
> Tasks\ Aborted/sec\ |\ _Total\ |\ Broker\ Activation,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412517870579
> Tasks\ Running\ |\ _Total\ |\ Broker\ Activation,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412517915980
> Tasks\ Started/sec\ |\ _Total\ |\ Broker\ Activation,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412517930480
> Activation\ Errors\ Total\ |\ Broker\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412517942581
> Broker\ Transaction\ Rollbacks\ |\ Broker\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412517956181
> Corrupted\ Messages\ Total\ |\ Broker\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412517968982
> Dequeued\ TransmissionQ\ Msgs/sec\ |\ Broker\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412517992382
> Dialog\ Timer\ Event\ Count\ |\ Broker\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412518002982
> Dropped\ Messages\ Total\ |\ Broker\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412518016083
> Enqueued\ Local\ Messages\ Total\ |\ Broker\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412518026883
> Enqueued\ Local\ Messages/sec\ |\ Broker\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412518038183
> Enqueued\ Messages\ Total\ |\ Broker\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412518061484
> Enqueued\ Messages/sec\ |\ Broker\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412518072084
> Enqueued\ P1\ Messages/sec\ |\ Broker\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412518083385
> Enqueued\ P10\ Messages/sec\ |\ Broker\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412518093785
> Enqueued\ P2\ Messages/sec\ |\ Broker\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412518111585
> Enqueued\ P3\ Messages/sec\ |\ Broker\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412518145586
> Enqueued\ P4\ Messages/sec\ |\ Broker\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412518166287
> Enqueued\ P5\ Messages/sec\ |\ Broker\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412518187587
> Enqueued\ P6\ Messages/sec\ |\ Broker\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412518203488
> Enqueued\ P7\ Messages/sec\ |\ Broker\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412518223488
> Enqueued\ P8\ Messages/sec\ |\ Broker\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412518286490
> Enqueued\ P9\ Messages/sec\ |\ Broker\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412518310291
> Enqueued\ TransmissionQ\ Msgs/sec\ |\ Broker\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412518330291
> Enqueued\ Transport\ Msg\ Frag\ Tot\ |\ Broker\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412518346092
> Enqueued\ Transport\ Msg\ Frags/sec\ |\ Broker\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412518363392
> Enqueued\ Transport\ Msgs\ Total\ |\ Broker\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412518404593
> Enqueued\ Transport\ Msgs/sec\ |\ Broker\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412518424494
> Forwarded\ Messages\ Total\ |\ Broker\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412518446394
> Forwarded\ Messages/sec\ |\ Broker\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412518486895
> Forwarded\ Msg\ Byte\ Total\ |\ Broker\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412518508596
> Forwarded\ Msg\ Bytes/sec\ |\ Broker\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412518548897
> Forwarded\ Msg\ Discarded\ Total\ |\ Broker\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412518572198
> Forwarded\ Msgs\ Discarded/sec\ |\ Broker\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412518592098
> Forwarded\ Pending\ Msg\ Bytes\ |\ Broker\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412518608698
> Forwarded\ Pending\ Msg\ Count\ |\ Broker\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412518626999
> SQL\ RECEIVE\ Total\ |\ Broker\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412518671100
> SQL\ RECEIVEs/sec\ |\ Broker\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412518683400
> SQL\ SEND\ Total\ |\ Broker\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412518694801
> SQL\ SENDs/sec\ |\ Broker\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412518705301
> Avg.\ Length\ of\ Batched\ Writes\ |\ Broker\ TO\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412518716501
> Avg.\ Length\ of\ Batched\ Writes\ BS\ |\ Broker\ TO\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=1i 1453876412518741102
> Avg.\ Time\ Between\ Batches\ (ms)\ |\ Broker\ TO\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=2062i 1453876412518752002
> Avg.\ Time\ Between\ Batches\ Base\ |\ Broker\ TO\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=1i 1453876412518763803
> Avg.\ Time\ to\ Write\ Batch\ (ms)\ |\ Broker\ TO\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412518774503
> Avg.\ Time\ to\ Write\ Batch\ Base\ |\ Broker\ TO\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=1i 1453876412518787903
> Transmission\ Obj\ Gets/Sec\ |\ Broker\ TO\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412518812104
> Transmission\ Obj\ Set\ Dirty/Sec\ |\ Broker\ TO\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412518824504
> Transmission\ Obj\ Writes/Sec\ |\ Broker\ TO\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412518853005
> Current\ Bytes\ for\ Recv\ I/O\ |\ Broker/DBM\ Transport,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412518864005
> Current\ Bytes\ for\ Send\ I/O\ |\ Broker/DBM\ Transport,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412518876906
> Current\ Msg\ Frags\ for\ Send\ I/O\ |\ Broker/DBM\ Transport,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412518928607
> Message\ Fragment\ P1\ Sends/sec\ |\ Broker/DBM\ Transport,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412518944407
> Message\ Fragment\ P10\ Sends/sec\ |\ Broker/DBM\ Transport,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412518956508
> Message\ Fragment\ P2\ Sends/sec\ |\ Broker/DBM\ Transport,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412518966908
> Message\ Fragment\ P3\ Sends/sec\ |\ Broker/DBM\ Transport,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412518978108
> Message\ Fragment\ P4\ Sends/sec\ |\ Broker/DBM\ Transport,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412519004309
> Message\ Fragment\ P5\ Sends/sec\ |\ Broker/DBM\ Transport,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412519016509
> Message\ Fragment\ P6\ Sends/sec\ |\ Broker/DBM\ Transport,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412519027810
> Message\ Fragment\ P7\ Sends/sec\ |\ Broker/DBM\ Transport,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412519038010
> Message\ Fragment\ P8\ Sends/sec\ |\ Broker/DBM\ Transport,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412519049010
> Message\ Fragment\ P9\ Sends/sec\ |\ Broker/DBM\ Transport,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412519074911
> Message\ Fragment\ Receives/sec\ |\ Broker/DBM\ Transport,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412519087211
> Message\ Fragment\ Sends/sec\ |\ Broker/DBM\ Transport,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412519104212
> Msg\ Fragment\ Recv\ Size\ Avg\ |\ Broker/DBM\ Transport,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412519137113
> Msg\ Fragment\ Recv\ Size\ Avg\ Base\ |\ Broker/DBM\ Transport,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412519150413
> Msg\ Fragment\ Send\ Size\ Avg\ |\ Broker/DBM\ Transport,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412519175214
> Msg\ Fragment\ Send\ Size\ Avg\ Base\ |\ Broker/DBM\ Transport,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412519186014
> Open\ Connection\ Count\ |\ Broker/DBM\ Transport,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412519197514
> Pending\ Bytes\ for\ Recv\ I/O\ |\ Broker/DBM\ Transport,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412519207814
> Pending\ Bytes\ for\ Send\ I/O\ |\ Broker/DBM\ Transport,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412519221515
> Pending\ Msg\ Frags\ for\ Recv\ I/O\ |\ Broker/DBM\ Transport,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412519245015
> Pending\ Msg\ Frags\ for\ Send\ I/O\ |\ Broker/DBM\ Transport,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412519255516
> Receive\ I/O\ bytes/sec\ |\ Broker/DBM\ Transport,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412519267116
> Receive\ I/O\ Len\ Avg\ |\ Broker/DBM\ Transport,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412519277516
> Receive\ I/O\ Len\ Avg\ Base\ |\ Broker/DBM\ Transport,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412519289017
> Receive\ I/Os/sec\ |\ Broker/DBM\ Transport,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412519327718
> Recv\ I/O\ Buffer\ Copies\ bytes/sec\ |\ Broker/DBM\ Transport,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412519339718
> Recv\ I/O\ Buffer\ Copies\ Count\ |\ Broker/DBM\ Transport,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412519355918
> Send\ I/O\ bytes/sec\ |\ Broker/DBM\ Transport,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412519372119
> Send\ I/O\ Len\ Avg\ |\ Broker/DBM\ Transport,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412519383919
> Send\ I/O\ Len\ Avg\ Base\ |\ Broker/DBM\ Transport,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412519414420
> Send\ I/Os/sec\ |\ Broker/DBM\ Transport,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412519425520
> Background\ writer\ pages/sec\ |\ Buffer\ Manager,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412519436720
> Buffer\ cache\ hit\ ratio\ |\ Buffer\ Manager,servername=WIN8-DEV,type=Performance\ counters value=1i 1453876412519447021
> Buffer\ cache\ hit\ ratio\ base\ |\ Buffer\ Manager,servername=WIN8-DEV,type=Performance\ counters value=7341i 1453876412519458421
> Checkpoint\ pages/sec\ |\ Buffer\ Manager,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412519481922
> Database\ pages\ |\ Buffer\ Manager,servername=WIN8-DEV,type=Performance\ counters value=12217i 1453876412519493222
> Extension\ allocated\ pages\ |\ Buffer\ Manager,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412519518323
> Extension\ free\ pages\ |\ Buffer\ Manager,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412519529023
> Extension\ in\ use\ as\ percentage\ |\ Buffer\ Manager,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412519540323
> Extension\ outstanding\ IO\ counter\ |\ Buffer\ Manager,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412519567224
> Extension\ page\ evictions/sec\ |\ Buffer\ Manager,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412519582924
> Extension\ page\ reads/sec\ |\ Buffer\ Manager,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412519595025
> Extension\ page\ unreferenced\ time\ |\ Buffer\ Manager,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412519609425
> Extension\ page\ writes/sec\ |\ Buffer\ Manager,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412519621425
> Free\ list\ stalls/sec\ |\ Buffer\ Manager,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412519644326
> Integral\ Controller\ Slope\ |\ Buffer\ Manager,servername=WIN8-DEV,type=Performance\ counters value=10i 1453876412519655926
> Lazy\ writes/sec\ |\ Buffer\ Manager,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412519683827
> Page\ life\ expectancy\ |\ Buffer\ Manager,servername=WIN8-DEV,type=Performance\ counters value=45614i 1453876412519698327
> Page\ lookups/sec\ |\ Buffer\ Manager,servername=WIN8-DEV,type=Performance\ counters value=7427i 1453876412519709828
> Page\ reads/sec\ |\ Buffer\ Manager,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412519732628
> Page\ writes/sec\ |\ Buffer\ Manager,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412519743529
> Readahead\ pages/sec\ |\ Buffer\ Manager,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412519754729
> Readahead\ time/sec\ |\ Buffer\ Manager,servername=WIN8-DEV,type=Performance\ counters value=2i 1453876412519764929
> Target\ pages\ |\ Buffer\ Manager,servername=WIN8-DEV,type=Performance\ counters value=16367616i 1453876412519779730
> Database\ pages\ |\ 000\ |\ Buffer\ Node,servername=WIN8-DEV,type=Performance\ counters value=12217i 1453876412519818131
> Local\ node\ page\ lookups/sec\ |\ 000\ |\ Buffer\ Node,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412519830031
> Page\ life\ expectancy\ |\ 000\ |\ Buffer\ Node,servername=WIN8-DEV,type=Performance\ counters value=45614i 1453876412519841831
> Remote\ node\ page\ lookups/sec\ |\ 000\ |\ Buffer\ Node,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412519852232
> Cache\ Entries\ Count\ |\ _Total\ |\ Catalog\ Metadata,servername=WIN8-DEV,type=Performance\ counters value=2589i 1453876412519863832
> Cache\ Entries\ Count\ |\ mssqlsystemresource\ |\ Catalog\ Metadata,servername=WIN8-DEV,type=Performance\ counters value=2212i 1453876412519893033
> Cache\ Entries\ Pinned\ Count\ |\ _Total\ |\ Catalog\ Metadata,servername=WIN8-DEV,type=Performance\ counters value=5i 1453876412519916933
> Cache\ Entries\ Pinned\ Count\ |\ mssqlsystemresource\ |\ Catalog\ Metadata,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412519932234
> Cache\ Hit\ Ratio\ |\ _Total\ |\ Catalog\ Metadata,servername=WIN8-DEV,type=Performance\ counters value=1i 1453876412519943534
> Cache\ Hit\ Ratio\ |\ mssqlsystemresource\ |\ Catalog\ Metadata,servername=WIN8-DEV,type=Performance\ counters value=1i 1453876412519955034
> Cache\ Hit\ Ratio\ Base\ |\ _Total\ |\ Catalog\ Metadata,servername=WIN8-DEV,type=Performance\ counters value=1883i 1453876412519978935
> Cache\ Hit\ Ratio\ Base\ |\ mssqlsystemresource\ |\ Catalog\ Metadata,servername=WIN8-DEV,type=Performance\ counters value=444i 1453876412519989735
> CLR\ Execution\ |\ CLR,servername=WIN8-DEV,type=Performance\ counters value=339517i 1453876412520001435
> Active\ cursors\ |\ _Total\ |\ Cursor\ Manager\ by\ Type,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520011836
> Active\ cursors\ |\ API\ Cursor\ |\ Cursor\ Manager\ by\ Type,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520025336
> Active\ cursors\ |\ TSQL\ Global\ Cursor\ |\ Cursor\ Manager\ by\ Type,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520049937
> Active\ cursors\ |\ TSQL\ Local\ Cursor\ |\ Cursor\ Manager\ by\ Type,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520060837
> Cache\ Hit\ Ratio\ |\ _Total\ |\ Cursor\ Manager\ by\ Type,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520072437
> Cache\ Hit\ Ratio\ |\ API\ Cursor\ |\ Cursor\ Manager\ by\ Type,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520084438
> Cache\ Hit\ Ratio\ |\ TSQL\ Global\ Cursor\ |\ Cursor\ Manager\ by\ Type,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520101238
> Cache\ Hit\ Ratio\ |\ TSQL\ Local\ Cursor\ |\ Cursor\ Manager\ by\ Type,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520156440
> Cache\ Hit\ Ratio\ Base\ |\ _Total\ |\ Cursor\ Manager\ by\ Type,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520168740
> Cache\ Hit\ Ratio\ Base\ |\ API\ Cursor\ |\ Cursor\ Manager\ by\ Type,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520180740
> Cache\ Hit\ Ratio\ Base\ |\ TSQL\ Global\ Cursor\ |\ Cursor\ Manager\ by\ Type,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520193241
> Cache\ Hit\ Ratio\ Base\ |\ TSQL\ Local\ Cursor\ |\ Cursor\ Manager\ by\ Type,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520209441
> Cached\ Cursor\ Counts\ |\ _Total\ |\ Cursor\ Manager\ by\ Type,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520237442
> Cached\ Cursor\ Counts\ |\ API\ Cursor\ |\ Cursor\ Manager\ by\ Type,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520248742
> Cached\ Cursor\ Counts\ |\ TSQL\ Global\ Cursor\ |\ Cursor\ Manager\ by\ Type,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520260542
> Cached\ Cursor\ Counts\ |\ TSQL\ Local\ Cursor\ |\ Cursor\ Manager\ by\ Type,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520271243
> Cursor\ Cache\ Use\ Counts/sec\ |\ _Total\ |\ Cursor\ Manager\ by\ Type,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520285143
> Cursor\ Cache\ Use\ Counts/sec\ |\ API\ Cursor\ |\ Cursor\ Manager\ by\ Type,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520325944
> Cursor\ Cache\ Use\ Counts/sec\ |\ TSQL\ Global\ Cursor\ |\ Cursor\ Manager\ by\ Type,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520338144
> Cursor\ Cache\ Use\ Counts/sec\ |\ TSQL\ Local\ Cursor\ |\ Cursor\ Manager\ by\ Type,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520350145
> Cursor\ memory\ usage\ |\ _Total\ |\ Cursor\ Manager\ by\ Type,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520360945
> Cursor\ memory\ usage\ |\ API\ Cursor\ |\ Cursor\ Manager\ by\ Type,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520372545
> Cursor\ memory\ usage\ |\ TSQL\ Global\ Cursor\ |\ Cursor\ Manager\ by\ Type,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520397346
> Cursor\ memory\ usage\ |\ TSQL\ Local\ Cursor\ |\ Cursor\ Manager\ by\ Type,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520408246
> Cursor\ Requests/sec\ |\ _Total\ |\ Cursor\ Manager\ by\ Type,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520420147
> Cursor\ Requests/sec\ |\ API\ Cursor\ |\ Cursor\ Manager\ by\ Type,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520444947
> Cursor\ Requests/sec\ |\ TSQL\ Global\ Cursor\ |\ Cursor\ Manager\ by\ Type,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520458648
> Cursor\ Requests/sec\ |\ TSQL\ Local\ Cursor\ |\ Cursor\ Manager\ by\ Type,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520483948
> Cursor\ worktable\ usage\ |\ _Total\ |\ Cursor\ Manager\ by\ Type,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520494849
> Cursor\ worktable\ usage\ |\ API\ Cursor\ |\ Cursor\ Manager\ by\ Type,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520506249
> Cursor\ worktable\ usage\ |\ TSQL\ Global\ Cursor\ |\ Cursor\ Manager\ by\ Type,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520516749
> Cursor\ worktable\ usage\ |\ TSQL\ Local\ Cursor\ |\ Cursor\ Manager\ by\ Type,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520528349
> Number\ of\ active\ cursor\ plans\ |\ _Total\ |\ Cursor\ Manager\ by\ Type,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520552950
> Number\ of\ active\ cursor\ plans\ |\ API\ Cursor\ |\ Cursor\ Manager\ by\ Type,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520565650
> Number\ of\ active\ cursor\ plans\ |\ TSQL\ Global\ Cursor\ |\ Cursor\ Manager\ by\ Type,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520582651
> Number\ of\ active\ cursor\ plans\ |\ TSQL\ Local\ Cursor\ |\ Cursor\ Manager\ by\ Type,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520595851
> Async\ population\ count\ |\ Cursor\ Manager\ Total,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520607652
> Cursor\ conversion\ rate\ |\ Cursor\ Manager\ Total,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520653453
> Cursor\ flushes\ |\ Cursor\ Manager\ Total,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520676953
> File\ Bytes\ Received/sec\ |\ _Total\ |\ Database\ Replica,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520691854
> Log\ Bytes\ Received/sec\ |\ _Total\ |\ Database\ Replica,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520702454
> Log\ remaining\ for\ undo\ |\ _Total\ |\ Database\ Replica,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520713954
> Log\ Send\ Queue\ |\ _Total\ |\ Database\ Replica,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520737655
> Mirrored\ Write\ Transactions/sec\ |\ _Total\ |\ Database\ Replica,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520750455
> Recovery\ Queue\ |\ _Total\ |\ Database\ Replica,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520763456
> Redo\ blocked/sec\ |\ _Total\ |\ Database\ Replica,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520773956
> Redo\ Bytes\ Remaining\ |\ _Total\ |\ Database\ Replica,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520785256
> Redone\ Bytes/sec\ |\ _Total\ |\ Database\ Replica,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520808457
> Total\ Log\ requiring\ undo\ |\ _Total\ |\ Database\ Replica,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520818857
> Transaction\ Delay\ |\ _Total\ |\ Database\ Replica,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520830357
> Active\ Transactions\ |\ _Total\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520840458
> Active\ Transactions\ |\ mssqlsystemresource\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520852958
> Backup/Restore\ Throughput/sec\ |\ _Total\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520877359
> Backup/Restore\ Throughput/sec\ |\ mssqlsystemresource\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520887959
> Bulk\ Copy\ Rows/sec\ |\ _Total\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520899059
> Bulk\ Copy\ Rows/sec\ |\ mssqlsystemresource\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520925360
> Bulk\ Copy\ Throughput/sec\ |\ _Total\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520939560
> Bulk\ Copy\ Throughput/sec\ |\ mssqlsystemresource\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520965161
> Commit\ table\ entries\ |\ _Total\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520976061
> Commit\ table\ entries\ |\ mssqlsystemresource\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412520999262
> Data\ File(s)\ Size\ (KB)\ |\ _Total\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=3512576i 1453876412521010862
> Data\ File(s)\ Size\ (KB)\ |\ mssqlsystemresource\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=40960i 1453876412521022563
> DBCC\ Logical\ Scan\ Bytes/sec\ |\ _Total\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521048563
> DBCC\ Logical\ Scan\ Bytes/sec\ |\ mssqlsystemresource\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521062964
> Group\ Commit\ Time/sec\ |\ _Total\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521075064
> Group\ Commit\ Time/sec\ |\ mssqlsystemresource\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521086864
> Log\ Bytes\ Flushed/sec\ |\ _Total\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=483328i 1453876412521100065
> Log\ Bytes\ Flushed/sec\ |\ mssqlsystemresource\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521160066
> Log\ Cache\ Hit\ Ratio\ |\ _Total\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521177567
> Log\ Cache\ Hit\ Ratio\ |\ mssqlsystemresource\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521190167
> Log\ Cache\ Hit\ Ratio\ Base\ |\ _Total\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521200567
> Log\ Cache\ Hit\ Ratio\ Base\ |\ mssqlsystemresource\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521212368
> Log\ Cache\ Reads/sec\ |\ _Total\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521236568
> Log\ Cache\ Reads/sec\ |\ mssqlsystemresource\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521247369
> Log\ File(s)\ Size\ (KB)\ |\ _Total\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=570992i 1453876412521258769
> Log\ File(s)\ Size\ (KB)\ |\ mssqlsystemresource\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=1016i 1453876412521269969
> Log\ File(s)\ Used\ Size\ (KB)\ |\ _Total\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=317272i 1453876412521284270
> Log\ File(s)\ Used\ Size\ (KB)\ |\ mssqlsystemresource\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=634i 1453876412521310470
> Log\ Flush\ Wait\ Time\ |\ _Total\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521321671
> Log\ Flush\ Wait\ Time\ |\ mssqlsystemresource\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521336771
> Log\ Flush\ Waits/sec\ |\ _Total\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521347671
> Log\ Flush\ Waits/sec\ |\ mssqlsystemresource\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521361972
> Log\ Flush\ Write\ Time\ (ms)\ |\ _Total\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=13i 1453876412521387272
> Log\ Flush\ Write\ Time\ (ms)\ |\ mssqlsystemresource\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521398473
> Log\ Flushes/sec\ |\ _Total\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=8i 1453876412521410273
> Log\ Flushes/sec\ |\ mssqlsystemresource\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521420973
> Log\ Growths\ |\ _Total\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521432473
> Log\ Growths\ |\ mssqlsystemresource\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521457774
> Log\ Pool\ Cache\ Misses/sec\ |\ _Total\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521468574
> Log\ Pool\ Cache\ Misses/sec\ |\ mssqlsystemresource\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521479975
> Log\ Pool\ Disk\ Reads/sec\ |\ _Total\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521490375
> Log\ Pool\ Disk\ Reads/sec\ |\ mssqlsystemresource\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521501775
> Log\ Pool\ Requests/sec\ |\ _Total\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521527776
> Log\ Pool\ Requests/sec\ |\ mssqlsystemresource\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521540876
> Log\ Shrinks\ |\ _Total\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521556077
> Log\ Shrinks\ |\ mssqlsystemresource\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521577877
> Log\ Truncations\ |\ _Total\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=81i 1453876412521591878
> Log\ Truncations\ |\ mssqlsystemresource\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521628379
> Percent\ Log\ Used\ |\ _Total\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=55i 1453876412521639779
> Percent\ Log\ Used\ |\ mssqlsystemresource\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=62i 1453876412521653179
> Repl.\ Pending\ Xacts\ |\ _Total\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521663780
> Repl.\ Pending\ Xacts\ |\ mssqlsystemresource\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521675280
> Repl.\ Trans.\ Rate\ |\ _Total\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521701881
> Repl.\ Trans.\ Rate\ |\ mssqlsystemresource\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521712981
> Shrink\ Data\ Movement\ Bytes/sec\ |\ _Total\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521724581
> Shrink\ Data\ Movement\ Bytes/sec\ |\ mssqlsystemresource\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521734982
> Tracked\ transactions/sec\ |\ _Total\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521746582
> Tracked\ transactions/sec\ |\ mssqlsystemresource\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521784183
> Transactions/sec\ |\ _Total\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=111i 1453876412521798783
> Transactions/sec\ |\ mssqlsystemresource\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=9i 1453876412521810484
> Write\ Transactions/sec\ |\ _Total\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=54i 1453876412521820884
> Write\ Transactions/sec\ |\ mssqlsystemresource\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521832084
> XTP\ Memory\ Used\ (KB)\ |\ _Total\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521858885
> XTP\ Memory\ Used\ (KB)\ |\ mssqlsystemresource\ |\ Databases,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521869885
> Usage\ |\ '#'\ and\ '##'\ as\ the\ name\ of\ temporary\ tables\ and\ stored\ procedures\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521881185
> Usage\ |\ '::'\ function\ calling\ syntax\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521892486
> Usage\ |\ '@'\ and\ names\ that\ start\ with\ '@@'\ as\ Transact-SQL\ identifiers\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521914186
> Usage\ |\ ADDING\ TAPE\ DEVICE\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521944987
> Usage\ |\ ALL\ Permission\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521956487
> Usage\ |\ ALTER\ DATABASE\ WITH\ TORN_PAGE_DETECTION\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521971288
> Usage\ |\ ALTER\ LOGIN\ WITH\ SET\ CREDENTIAL\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412521985288
> Usage\ |\ asymmetric_keys\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522000189
> Usage\ |\ asymmetric_keys.attested_by\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522049290
> Usage\ |\ Azeri_Cyrillic_90\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522066290
> Usage\ |\ Azeri_Latin_90\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522082791
> Usage\ |\ BACKUP\ DATABASE\ or\ LOG\ TO\ TAPE\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522102491
> Usage\ |\ certificates\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522124292
> Usage\ |\ certificates.attested_by\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522150593
> Usage\ |\ Create/alter\ SOAP\ endpoint\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522161493
> Usage\ |\ CREATE_DROP_DEFAULT\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522175393
> Usage\ |\ CREATE_DROP_RULE\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522186093
> Usage\ |\ Data\ types:\ text\ ntext\ or\ image\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522197594
> Usage\ |\ Database\ compatibility\ level\ 100\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522223894
> Usage\ |\ Database\ compatibility\ level\ 110\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=4i 1453876412522234995
> Usage\ |\ Database\ compatibility\ level\ 90\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522249695
> Usage\ |\ Database\ Mirroring\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522260995
> Usage\ |\ DATABASEPROPERTY\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522272596
> Usage\ |\ DATABASEPROPERTYEX('IsFullTextEnabled')\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522297496
> Usage\ |\ DBCC\ [UN]PINTABLE\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522308597
> Usage\ |\ DBCC\ DBREINDEX\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522320097
> Usage\ |\ DBCC\ INDEXDEFRAG\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522330797
> Usage\ |\ DBCC\ SHOWCONTIG\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522342298
> Usage\ |\ DBCC_EXTENTINFO\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522367798
> Usage\ |\ DBCC_IND\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522378699
> Usage\ |\ DEFAULT\ keyword\ as\ a\ default\ value\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522390399
> Usage\ |\ Deprecated\ Attested\ Option\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522400899
> Usage\ |\ Deprecated\ encryption\ algorithm\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522425100
> Usage\ |\ DESX\ algorithm\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522467701
> Usage\ |\ dm_fts_active_catalogs\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522483201
> Usage\ |\ dm_fts_active_catalogs.is_paused\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522494802
> Usage\ |\ dm_fts_active_catalogs.previous_status\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522505102
> Usage\ |\ dm_fts_active_catalogs.previous_status_description\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522518302
> Usage\ |\ dm_fts_active_catalogs.row_count_in_thousands\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522543103
> Usage\ |\ dm_fts_active_catalogs.status\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522559203
> Usage\ |\ dm_fts_active_catalogs.status_description\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522572204
> Usage\ |\ dm_fts_active_catalogs.worker_count\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522595904
> Usage\ |\ dm_fts_memory_buffers\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522611105
> Usage\ |\ dm_fts_memory_buffers.row_count\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522635505
> Usage\ |\ DROP\ INDEX\ with\ two-part\ name\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522646406
> Usage\ |\ endpoint_webmethods\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522658106
> Usage\ |\ EXTPROP_LEVEL0TYPE\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522668306
> Usage\ |\ EXTPROP_LEVEL0USER\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522681607
> Usage\ |\ FILE_ID\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522705307
> Usage\ |\ fn_get_sql\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522717808
> Usage\ |\ fn_servershareddrives\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522731008
> Usage\ |\ fn_trace_geteventinfo\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522741408
> Usage\ |\ fn_trace_getfilterinfo\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522752709
> Usage\ |\ fn_trace_getinfo\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522778009
> Usage\ |\ fn_trace_gettable\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522788809
> Usage\ |\ fn_virtualservernodes\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522800010
> Usage\ |\ fulltext_catalogs\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522810210
> Usage\ |\ fulltext_catalogs.data_space_id\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522823010
> Usage\ |\ fulltext_catalogs.file_id\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522864011
> Usage\ |\ fulltext_catalogs.path\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522877112
> Usage\ |\ FULLTEXTCATALOGPROPERTY('LogSize')\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522889212
> Usage\ |\ FULLTEXTCATALOGPROPERTY('PopulateStatus')\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522899712
> Usage\ |\ FULLTEXTSERVICEPROPERTY('ConnectTimeout')\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522939013
> Usage\ |\ FULLTEXTSERVICEPROPERTY('DataTimeout')\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522982115
> Usage\ |\ FULLTEXTSERVICEPROPERTY('ResourceUsage')\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412522994915
> Usage\ |\ GROUP\ BY\ ALL\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523006415
> Usage\ |\ Hindi\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523019416
> Usage\ |\ IDENTITYCOL\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523030816
> Usage\ |\ IN\ PATH\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523054817
> Usage\ |\ Index\ view\ select\ list\ without\ COUNT_BIG(*)\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523070517
> Usage\ |\ INDEX_OPTION\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523082917
> Usage\ |\ INDEXKEY_PROPERTY\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523093118
> Usage\ |\ Indirect\ TVF\ hints\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523109018
> Usage\ |\ INSERT\ NULL\ into\ TIMESTAMP\ columns\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523136519
> Usage\ |\ INSERT_HINTS\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523149419
> Usage\ |\ Korean_Wansung_Unicode\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523164019
> Usage\ |\ Lithuanian_Classic\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523174620
> Usage\ |\ Macedonian\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523186520
> Usage\ |\ MODIFY\ FILEGROUP\ READONLY\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523214321
> Usage\ |\ MODIFY\ FILEGROUP\ READWRITE\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523225521
> Usage\ |\ More\ than\ two-part\ column\ name\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523237121
> Usage\ |\ Multiple\ table\ hints\ without\ comma\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523247822
> Usage\ |\ NOLOCK\ or\ READUNCOMMITTED\ in\ UPDATE\ or\ DELETE\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523259522
> Usage\ |\ Numbered\ stored\ procedures\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523285723
> Usage\ |\ numbered_procedure_parameters\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523296823
> Usage\ |\ numbered_procedures\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523309923
> Usage\ |\ objidupdate\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523320524
> Usage\ |\ Old\ NEAR\ Syntax\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523335124
> Usage\ |\ OLEDB\ for\ ad\ hoc\ connections\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523374225
> Usage\ |\ PERMISSIONS\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523400326
> Usage\ |\ READTEXT\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523412626
> Usage\ |\ REMSERVER\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523425326
> Usage\ |\ RESTORE\ DATABASE\ or\ LOG\ WITH\ MEDIAPASSWORD\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523436927
> Usage\ |\ RESTORE\ DATABASE\ or\ LOG\ WITH\ PASSWORD\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523461727
> Usage\ |\ Returning\ results\ from\ trigger\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523472828
> Usage\ |\ ROWGUIDCOL\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523486128
> Usage\ |\ SET\ ANSI_NULLS\ OFF\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523496828
> Usage\ |\ SET\ ANSI_PADDING\ OFF\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523510229
> Usage\ |\ SET\ CONCAT_NULL_YIELDS_NULL\ OFF\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523539229
> Usage\ |\ SET\ ERRLVL\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523551030
> Usage\ |\ SET\ FMTONLY\ ON\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523564530
> Usage\ |\ SET\ OFFSETS\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523576930
> Usage\ |\ SET\ REMOTE_PROC_TRANSACTIONS\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523591931
> Usage\ |\ SET\ ROWCOUNT\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523625132
> Usage\ |\ SETUSER\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523636932
> Usage\ |\ soap_endpoints\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523648532
> Usage\ |\ sp_addapprole\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523659133
> Usage\ |\ sp_addextendedproc\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523670433
> Usage\ |\ sp_addlogin\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523695334
> Usage\ |\ sp_addremotelogin\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523706634
> Usage\ |\ sp_addrole\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523718134
> Usage\ |\ sp_addrolemember\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523728534
> Usage\ |\ sp_addserver\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523740035
> Usage\ |\ sp_addsrvrolemember\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523764935
> Usage\ |\ sp_addtype\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523777236
> Usage\ |\ sp_adduser\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523788836
> Usage\ |\ sp_approlepassword\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523798936
> Usage\ |\ sp_attach_db\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523811837
> Usage\ |\ sp_attach_single_file_db\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523856038
> Usage\ |\ sp_bindefault\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523868238
> Usage\ |\ sp_bindrule\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523879738
> Usage\ |\ sp_bindsession\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523889939
> Usage\ |\ sp_certify_removable\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523901139
> Usage\ |\ sp_change_users_login\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523943340
> Usage\ |\ sp_changedbowner\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523954940
> Usage\ |\ sp_changeobjectowner\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523966741
> Usage\ |\ sp_configure\ 'affinity\ mask'\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523978741
> Usage\ |\ sp_configure\ 'affinity64\ mask'\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412523990641
> Usage\ |\ sp_configure\ 'allow\ updates'\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524016742
> Usage\ |\ sp_configure\ 'c2\ audit\ mode'\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524031542
> Usage\ |\ sp_configure\ 'default\ trace\ enabled'\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524046343
> Usage\ |\ sp_configure\ 'disallow\ results\ from\ triggers'\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524057343
> Usage\ |\ sp_configure\ 'ft\ crawl\ bandwidth\ (max)'\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524069043
> Usage\ |\ sp_configure\ 'ft\ crawl\ bandwidth\ (min)'\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524093444
> Usage\ |\ sp_configure\ 'ft\ notify\ bandwidth\ (max)'\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524111545
> Usage\ |\ sp_configure\ 'ft\ notify\ bandwidth\ (min)'\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524144745
> Usage\ |\ sp_configure\ 'locks'\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524155646
> Usage\ |\ sp_configure\ 'open\ objects'\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524167246
> Usage\ |\ sp_configure\ 'priority\ boost'\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524192247
> Usage\ |\ sp_configure\ 'remote\ proc\ trans'\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524206447
> Usage\ |\ sp_configure\ 'set\ working\ set\ size'\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524218347
> Usage\ |\ sp_control_dbmasterkey_password\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524230848
> Usage\ |\ sp_create_removable\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524242348
> Usage\ |\ sp_db_increased_partitions\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524281149
> Usage\ |\ sp_db_selective_xml_index\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524294249
> Usage\ |\ sp_db_vardecimal_storage_format\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524305750
> Usage\ |\ sp_dbcmptlevel\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524316050
> Usage\ |\ sp_dbfixedrolepermission\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524329050
> Usage\ |\ sp_dbremove\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524352651
> Usage\ |\ sp_defaultdb\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524363451
> Usage\ |\ sp_defaultlanguage\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524374852
> Usage\ |\ sp_denylogin\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524384952
> Usage\ |\ sp_depends\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524399052
> Usage\ |\ sp_detach_db\ @keepfulltextindexfile\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524423953
> Usage\ |\ sp_dropapprole\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524434753
> Usage\ |\ sp_dropextendedproc\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524446053
> Usage\ |\ sp_droplogin\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524456454
> Usage\ |\ sp_dropremotelogin\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524467554
> Usage\ |\ sp_droprole\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524492655
> Usage\ |\ sp_droprolemember\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524506655
> Usage\ |\ sp_dropsrvrolemember\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524518355
> Usage\ |\ sp_droptype\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524528756
> Usage\ |\ sp_dropuser\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524539756
> Usage\ |\ sp_estimated_rowsize_reduction_for_vardecimal\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524565457
> Usage\ |\ sp_fulltext_catalog\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524576357
> Usage\ |\ sp_fulltext_column\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524587457
> Usage\ |\ sp_fulltext_database\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524605858
> Usage\ |\ sp_fulltext_service\ @action=clean_up\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524617758
> Usage\ |\ sp_fulltext_service\ @action=connect_timeout\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524659159
> Usage\ |\ sp_fulltext_service\ @action=data_timeout\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524674860
> Usage\ |\ sp_fulltext_service\ @action=resource_usage\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524686560
> Usage\ |\ sp_fulltext_table\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524697160
> Usage\ |\ sp_getbindtoken\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524708460
> Usage\ |\ sp_grantdbaccess\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524736461
> Usage\ |\ sp_grantlogin\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524747662
> Usage\ |\ sp_help_fulltext_catalog_components\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524759062
> Usage\ |\ sp_help_fulltext_catalogs\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524782162
> Usage\ |\ sp_help_fulltext_catalogs_cursor\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524806463
> Usage\ |\ sp_help_fulltext_columns\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524838864
> Usage\ |\ sp_help_fulltext_columns_cursor\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524850564
> Usage\ |\ sp_help_fulltext_tables\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524862165
> Usage\ |\ sp_help_fulltext_tables_cursor\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524872665
> Usage\ |\ sp_helpdevice\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524884065
> Usage\ |\ sp_helpextendedproc\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524922766
> Usage\ |\ sp_helpremotelogin\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524936967
> Usage\ |\ sp_indexoption\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524948967
> Usage\ |\ sp_lock\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524959267
> Usage\ |\ sp_password\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524970667
> Usage\ |\ sp_remoteoption\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412524993668
> Usage\ |\ sp_renamedb\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525007168
> Usage\ |\ sp_resetstatus\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525018969
> Usage\ |\ sp_revokedbaccess\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525032869
> Usage\ |\ sp_revokelogin\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525059370
> Usage\ |\ sp_srvrolepermission\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525086671
> Usage\ |\ sp_trace_create\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525102771
> Usage\ |\ sp_trace_getdata\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525118471
> Usage\ |\ sp_trace_setevent\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525128972
> Usage\ |\ sp_trace_setfilter\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525140272
> Usage\ |\ sp_trace_setstatus\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525184773
> Usage\ |\ sp_unbindefault\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525197073
> Usage\ |\ sp_unbindrule\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525208574
> Usage\ |\ SQL_AltDiction_CP1253_CS_AS\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525218674
> Usage\ |\ sql_dependencies\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525230174
> Usage\ |\ String\ literals\ as\ column\ aliases\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=4i 1453876412525257075
> Usage\ |\ sysaltfiles\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525268475
> Usage\ |\ syscacheobjects\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525279876
> Usage\ |\ syscolumns\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525299176
> Usage\ |\ syscomments\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525313677
> Usage\ |\ sysconfigures\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525340677
> Usage\ |\ sysconstraints\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525351478
> Usage\ |\ syscurconfigs\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525362778
> Usage\ |\ sysdatabases\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=2i 1453876412525372978
> Usage\ |\ sysdepends\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525384178
> Usage\ |\ sysdevices\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525407279
> Usage\ |\ sysfilegroups\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525421379
> Usage\ |\ sysfiles\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525432780
> Usage\ |\ sysforeignkeys\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525442980
> Usage\ |\ sysfulltextcatalogs\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525454180
> Usage\ |\ sysindexes\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525476481
> Usage\ |\ sysindexkeys\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525486781
> Usage\ |\ syslockinfo\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525503582
> Usage\ |\ syslogins\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=8i 1453876412525514682
> Usage\ |\ sysmembers\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525525982
> Usage\ |\ sysmessages\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525553883
> Usage\ |\ sysobjects\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525564483
> Usage\ |\ sysoledbusers\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525575683
> Usage\ |\ sysopentapes\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525585884
> Usage\ |\ sysperfinfo\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525596984
> Usage\ |\ syspermissions\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525639885
> Usage\ |\ sysprocesses\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525651786
> Usage\ |\ sysprotects\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525663186
> Usage\ |\ sysreferences\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525673286
> Usage\ |\ sysremotelogins\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525687986
> Usage\ |\ sysservers\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525720987
> Usage\ |\ systypes\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525740588
> Usage\ |\ sysusers\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525752688
> Usage\ |\ Table\ hint\ without\ WITH\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525762788
> Usage\ |\ Text\ in\ row\ table\ option\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525774389
> Usage\ |\ TEXTPTR\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525798089
> Usage\ |\ TEXTVALID\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525808390
> Usage\ |\ TIMESTAMP\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525819790
> Usage\ |\ UPDATETEXT\ or\ WRITETEXT\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525830090
> Usage\ |\ USER_ID\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=1i 1453876412525841291
> Usage\ |\ Using\ OLEDB\ for\ linked\ servers\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525877392
> Usage\ |\ Vardecimal\ storage\ format\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525889492
> Usage\ |\ XMLDATA\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525901192
> Usage\ |\ XP_API\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=7i 1453876412525952493
> Usage\ |\ xp_grantlogin\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412525964994
> Usage\ |\ xp_loginconfig\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526001095
> Usage\ |\ xp_revokelogin\ |\ Deprecated\ Features,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526015695
> Distributed\ Query\ |\ Average\ execution\ time\ (ms)\ |\ Exec\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526035096
> Distributed\ Query\ |\ Cumulative\ execution\ time\ (ms)\ per\ second\ |\ Exec\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526046596
> Distributed\ Query\ |\ Execs\ in\ progress\ |\ Exec\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526061696
> Distributed\ Query\ |\ Execs\ started\ per\ second\ |\ Exec\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526116398
> DTC\ calls\ |\ Average\ execution\ time\ (ms)\ |\ Exec\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526137598
> DTC\ calls\ |\ Cumulative\ execution\ time\ (ms)\ per\ second\ |\ Exec\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526153299
> DTC\ calls\ |\ Execs\ in\ progress\ |\ Exec\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526164599
> DTC\ calls\ |\ Execs\ started\ per\ second\ |\ Exec\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526176299
> Extended\ Procedures\ |\ Average\ execution\ time\ (ms)\ |\ Exec\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526201600
> Extended\ Procedures\ |\ Cumulative\ execution\ time\ (ms)\ per\ second\ |\ Exec\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526212600
> Extended\ Procedures\ |\ Execs\ in\ progress\ |\ Exec\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526230901
> Extended\ Procedures\ |\ Execs\ started\ per\ second\ |\ Exec\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526244401
> OLEDB\ calls\ |\ Average\ execution\ time\ (ms)\ |\ Exec\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526256002
> OLEDB\ calls\ |\ Cumulative\ execution\ time\ (ms)\ per\ second\ |\ Exec\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526282802
> OLEDB\ calls\ |\ Execs\ in\ progress\ |\ Exec\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526294103
> OLEDB\ calls\ |\ Execs\ started\ per\ second\ |\ Exec\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526305703
> Avg\ time\ delete\ FileTable\ item\ |\ FileTable,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526316203
> Avg\ time\ FileTable\ enumeration\ |\ FileTable,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526327703
> Avg\ time\ FileTable\ handle\ kill\ |\ FileTable,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526350504
> Avg\ time\ move\ FileTable\ item\ |\ FileTable,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526361104
> Avg\ time\ per\ file\ I/O\ request\ |\ FileTable,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526372405
> Avg\ time\ per\ file\ I/O\ response\ |\ FileTable,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526386705
> Avg\ time\ rename\ FileTable\ item\ |\ FileTable,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526398605
> Avg\ time\ to\ get\ FileTable\ item\ |\ FileTable,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526421406
> Avg\ time\ update\ FileTable\ item\ |\ FileTable,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526432106
> FileTable\ db\ operations/sec\ |\ FileTable,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526457907
> FileTable\ enumeration\ reqs/sec\ |\ FileTable,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526468807
> FileTable\ file\ I/O\ requests/sec\ |\ FileTable,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526480108
> FileTable\ file\ I/O\ response/sec\ |\ FileTable,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526507008
> FileTable\ item\ delete\ reqs/sec\ |\ FileTable,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526525809
> FileTable\ item\ get\ requests/sec\ |\ FileTable,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526540209
> FileTable\ item\ move\ reqs/sec\ |\ FileTable,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526550709
> FileTable\ item\ rename\ reqs/sec\ |\ FileTable,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526565010
> FileTable\ item\ update\ reqs/sec\ |\ FileTable,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526609511
> FileTable\ kill\ handle\ ops/sec\ |\ FileTable,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526621511
> FileTable\ table\ operations/sec\ |\ FileTable,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526633012
> Time\ delete\ FileTable\ item\ BASE\ |\ FileTable,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526643112
> Time\ FileTable\ enumeration\ BASE\ |\ FileTable,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526654312
> Time\ FileTable\ handle\ kill\ BASE\ |\ FileTable,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526677313
> Time\ move\ FileTable\ item\ BASE\ |\ FileTable,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526692713
> Time\ per\ file\ I/O\ request\ BASE\ |\ FileTable,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526704613
> Time\ per\ file\ I/O\ response\ BASE\ |\ FileTable,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526715014
> Time\ rename\ FileTable\ item\ BASE\ |\ FileTable,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526726314
> Time\ to\ get\ FileTable\ item\ BASE\ |\ FileTable,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526753015
> Time\ update\ FileTable\ item\ BASE\ |\ FileTable,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526763915
> Active\ Temp\ Tables\ |\ General\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=26i 1453876412526775915
> Connection\ Reset/sec\ |\ General\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526801716
> Event\ Notifications\ Delayed\ Drop\ |\ General\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526816616
> HTTP\ Authenticated\ Requests\ |\ General\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526839717
> Logical\ Connections\ |\ General\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=9i 1453876412526850217
> Logins/sec\ |\ General\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=3i 1453876412526861518
> Logouts/sec\ |\ General\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=7i 1453876412526871718
> Mars\ Deadlocks\ |\ General\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526882918
> Non-atomic\ yield\ rate\ |\ General\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526920619
> Processes\ blocked\ |\ General\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526938020
> SOAP\ Empty\ Requests\ |\ General\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526950020
> SOAP\ Method\ Invocations\ |\ General\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526960220
> SOAP\ Session\ Initiate\ Requests\ |\ General\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412526971521
> SOAP\ Session\ Terminate\ Requests\ |\ General\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527004821
> SOAP\ SQL\ Requests\ |\ General\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527015822
> SOAP\ WSDL\ Requests\ |\ General\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527027222
> SQL\ Trace\ IO\ Provider\ Lock\ Waits\ |\ General\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527040822
> Temp\ Tables\ Creation\ Rate\ |\ General\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=11i 1453876412527052523
> Temp\ Tables\ For\ Destruction\ |\ General\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527095024
> Tempdb\ recovery\ unit\ id\ |\ General\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527114824
> Tempdb\ rowset\ id\ |\ General\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527136125
> Trace\ Event\ Notification\ Queue\ |\ General\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527147925
> Transactions\ |\ General\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=1i 1453876412527168326
> User\ Connections\ |\ General\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=9i 1453876412527195927
> Avg.\ Bytes/Read\ |\ _Total\ |\ HTTP\ Storage,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527206527
> Avg.\ Bytes/Read\ BASE\ |\ _Total\ |\ HTTP\ Storage,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527217927
> Avg.\ Bytes/Transfer\ |\ _Total\ |\ HTTP\ Storage,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527228427
> Avg.\ Bytes/Transfer\ BASE\ |\ _Total\ |\ HTTP\ Storage,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527239828
> Avg.\ Bytes/Write\ |\ _Total\ |\ HTTP\ Storage,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527266328
> Avg.\ Bytes/Write\ BASE\ |\ _Total\ |\ HTTP\ Storage,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527277429
> Avg.\ microsec/Read\ |\ _Total\ |\ HTTP\ Storage,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527289029
> Avg.\ microsec/Read\ BASE\ |\ _Total\ |\ HTTP\ Storage,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527299229
> Avg.\ microsec/Transfer\ |\ _Total\ |\ HTTP\ Storage,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527310630
> Avg.\ microsec/Transfer\ BASE\ |\ _Total\ |\ HTTP\ Storage,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527334230
> Avg.\ microsec/Write\ |\ _Total\ |\ HTTP\ Storage,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527344830
> Avg.\ microsec/Write\ BASE\ |\ _Total\ |\ HTTP\ Storage,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527360331
> HTTP\ Storage\ IO\ retry/sec\ |\ _Total\ |\ HTTP\ Storage,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527374031
> Outstanding\ HTTP\ Storage\ IO\ |\ _Total\ |\ HTTP\ Storage,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527385632
> Read\ Bytes/Sec\ |\ _Total\ |\ HTTP\ Storage,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527415432
> Reads/Sec\ |\ _Total\ |\ HTTP\ Storage,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527427033
> Total\ Bytes/Sec\ |\ _Total\ |\ HTTP\ Storage,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527441733
> Transfers/Sec\ |\ _Total\ |\ HTTP\ Storage,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527452233
> Write\ Bytes/Sec\ |\ _Total\ |\ HTTP\ Storage,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527466634
> Writes/Sec\ |\ _Total\ |\ HTTP\ Storage,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527496134
> Average\ Latch\ Wait\ Time\ (ms)\ |\ Latches,servername=WIN8-DEV,type=Performance\ counters value=1i 1453876412527507035
> Average\ Latch\ Wait\ Time\ Base\ |\ Latches,servername=WIN8-DEV,type=Performance\ counters value=33i 1453876412527518435
> Latch\ Waits/sec\ |\ Latches,servername=WIN8-DEV,type=Performance\ counters value=39i 1453876412527528935
> Number\ of\ SuperLatches\ |\ Latches,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527540136
> SuperLatch\ Demotions/sec\ |\ Latches,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527562936
> SuperLatch\ Promotions/sec\ |\ Latches,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527573137
> Total\ Latch\ Wait\ Time\ (ms)\ |\ Latches,servername=WIN8-DEV,type=Performance\ counters value=33i 1453876412527587337
> Average\ Wait\ Time\ (ms)\ |\ _Total\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527606237
> Average\ Wait\ Time\ (ms)\ |\ AllocUnit\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527622138
> Average\ Wait\ Time\ (ms)\ |\ Application\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527662339
> Average\ Wait\ Time\ (ms)\ |\ Database\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527674539
> Average\ Wait\ Time\ (ms)\ |\ Extent\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527686040
> Average\ Wait\ Time\ (ms)\ |\ File\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527696540
> Average\ Wait\ Time\ (ms)\ |\ HoBT\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527707640
> Average\ Wait\ Time\ (ms)\ |\ Key\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527730341
> Average\ Wait\ Time\ (ms)\ |\ Metadata\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527754641
> Average\ Wait\ Time\ (ms)\ |\ Object\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527767442
> Average\ Wait\ Time\ (ms)\ |\ OIB\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527784142
> Average\ Wait\ Time\ (ms)\ |\ Page\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527795842
> Average\ Wait\ Time\ (ms)\ |\ RID\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527826443
> Average\ Wait\ Time\ (ms)\ |\ RowGroup\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527844244
> Average\ Wait\ Time\ Base\ |\ _Total\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527865044
> Average\ Wait\ Time\ Base\ |\ AllocUnit\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527875645
> Average\ Wait\ Time\ Base\ |\ Application\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527887045
> Average\ Wait\ Time\ Base\ |\ Database\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527924846
> Average\ Wait\ Time\ Base\ |\ Extent\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527939346
> Average\ Wait\ Time\ Base\ |\ File\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527951347
> Average\ Wait\ Time\ Base\ |\ HoBT\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527961847
> Average\ Wait\ Time\ Base\ |\ Key\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527973347
> Average\ Wait\ Time\ Base\ |\ Metadata\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412527999748
> Average\ Wait\ Time\ Base\ |\ Object\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412528010848
> Average\ Wait\ Time\ Base\ |\ OIB\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412528022348
> Average\ Wait\ Time\ Base\ |\ Page\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412528032549
> Average\ Wait\ Time\ Base\ |\ RID\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412528043949
> Average\ Wait\ Time\ Base\ |\ RowGroup\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412528075750
> Lock\ Requests/sec\ |\ _Total\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=8075i 1453876412528086950
> Lock\ Requests/sec\ |\ AllocUnit\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=2i 1453876412528098250
> Lock\ Requests/sec\ |\ Application\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412528127151
> Lock\ Requests/sec\ |\ Database\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=229i 1453876412528146652
> Lock\ Requests/sec\ |\ Extent\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=70i 1453876412528197753
> Lock\ Requests/sec\ |\ File\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412528213754
> Lock\ Requests/sec\ |\ HoBT\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=49i 1453876412528232254
> Lock\ Requests/sec\ |\ Key\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=2282i 1453876412528242954
> Lock\ Requests/sec\ |\ Metadata\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=3501i 1453876412528254355
> Lock\ Requests/sec\ |\ Object\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=1799i 1453876412528277055
> Lock\ Requests/sec\ |\ OIB\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412528287355
> Lock\ Requests/sec\ |\ Page\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=103i 1453876412528298756
> Lock\ Requests/sec\ |\ RID\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=40i 1453876412528312756
> Lock\ Requests/sec\ |\ RowGroup\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412528327357
> Lock\ Timeouts\ (timeout\ >\ 0)/sec\ |\ _Total\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412528363258
> Lock\ Timeouts\ (timeout\ >\ 0)/sec\ |\ AllocUnit\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412528374958
> Lock\ Timeouts\ (timeout\ >\ 0)/sec\ |\ Application\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412528386858
> Lock\ Timeouts\ (timeout\ >\ 0)/sec\ |\ Database\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412528397558
> Lock\ Timeouts\ (timeout\ >\ 0)/sec\ |\ Extent\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412528408859
> Lock\ Timeouts\ (timeout\ >\ 0)/sec\ |\ File\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412528432359
> Lock\ Timeouts\ (timeout\ >\ 0)/sec\ |\ HoBT\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412528442860
> Lock\ Timeouts\ (timeout\ >\ 0)/sec\ |\ Key\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412528454260
> Lock\ Timeouts\ (timeout\ >\ 0)/sec\ |\ Metadata\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412529628991
> Lock\ Timeouts\ (timeout\ >\ 0)/sec\ |\ Object\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412529651592
> Lock\ Timeouts\ (timeout\ >\ 0)/sec\ |\ OIB\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412529663992
> Lock\ Timeouts\ (timeout\ >\ 0)/sec\ |\ Page\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412529676692
> Lock\ Timeouts\ (timeout\ >\ 0)/sec\ |\ RID\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412529688093
> Lock\ Timeouts\ (timeout\ >\ 0)/sec\ |\ RowGroup\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412529700393
> Lock\ Timeouts/sec\ |\ _Total\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=7i 1453876412529725794
> Lock\ Timeouts/sec\ |\ AllocUnit\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412529737294
> Lock\ Timeouts/sec\ |\ Application\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412529749294
> Lock\ Timeouts/sec\ |\ Database\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412529759995
> Lock\ Timeouts/sec\ |\ Extent\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412529771695
> Lock\ Timeouts/sec\ |\ File\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412529794996
> Lock\ Timeouts/sec\ |\ HoBT\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412529805696
> Lock\ Timeouts/sec\ |\ Key\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=7i 1453876412529818596
> Lock\ Timeouts/sec\ |\ Metadata\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412529829396
> Lock\ Timeouts/sec\ |\ Object\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412529840797
> Lock\ Timeouts/sec\ |\ OIB\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412529902898
> Lock\ Timeouts/sec\ |\ Page\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412529931399
> Lock\ Timeouts/sec\ |\ RID\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412529944099
> Lock\ Timeouts/sec\ |\ RowGroup\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412529955000
> Lock\ Wait\ Time\ (ms)\ |\ _Total\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412529966800
> Lock\ Wait\ Time\ (ms)\ |\ AllocUnit\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530001301
> Lock\ Wait\ Time\ (ms)\ |\ Application\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530012901
> Lock\ Wait\ Time\ (ms)\ |\ Database\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530025002
> Lock\ Wait\ Time\ (ms)\ |\ Extent\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530036502
> Lock\ Wait\ Time\ (ms)\ |\ File\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530052302
> Lock\ Wait\ Time\ (ms)\ |\ HoBT\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530075903
> Lock\ Wait\ Time\ (ms)\ |\ Key\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530088903
> Lock\ Wait\ Time\ (ms)\ |\ Metadata\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530101204
> Lock\ Wait\ Time\ (ms)\ |\ Object\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530120204
> Lock\ Wait\ Time\ (ms)\ |\ OIB\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530136505
> Lock\ Wait\ Time\ (ms)\ |\ Page\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530159905
> Lock\ Wait\ Time\ (ms)\ |\ RID\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530171306
> Lock\ Wait\ Time\ (ms)\ |\ RowGroup\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530183206
> Lock\ Waits/sec\ |\ _Total\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530193706
> Lock\ Waits/sec\ |\ AllocUnit\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530205606
> Lock\ Waits/sec\ |\ Application\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530229807
> Lock\ Waits/sec\ |\ Database\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530240707
> Lock\ Waits/sec\ |\ Extent\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530255108
> Lock\ Waits/sec\ |\ File\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530265708
> Lock\ Waits/sec\ |\ HoBT\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530277308
> Lock\ Waits/sec\ |\ Key\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530309209
> Lock\ Waits/sec\ |\ Metadata\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530320709
> Lock\ Waits/sec\ |\ Object\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530332210
> Lock\ Waits/sec\ |\ OIB\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530342910
> Lock\ Waits/sec\ |\ Page\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530364311
> Lock\ Waits/sec\ |\ RID\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530389911
> Lock\ Waits/sec\ |\ RowGroup\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530401012
> Number\ of\ Deadlocks/sec\ |\ _Total\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530412312
> Number\ of\ Deadlocks/sec\ |\ AllocUnit\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530423812
> Number\ of\ Deadlocks/sec\ |\ Application\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530435413
> Number\ of\ Deadlocks/sec\ |\ Database\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530476714
> Number\ of\ Deadlocks/sec\ |\ Extent\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530488814
> Number\ of\ Deadlocks/sec\ |\ File\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530500114
> Number\ of\ Deadlocks/sec\ |\ HoBT\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530510415
> Number\ of\ Deadlocks/sec\ |\ Key\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530524215
> Number\ of\ Deadlocks/sec\ |\ Metadata\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530547416
> Number\ of\ Deadlocks/sec\ |\ Object\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530558916
> Number\ of\ Deadlocks/sec\ |\ OIB\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530570416
> Number\ of\ Deadlocks/sec\ |\ Page\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530580716
> Number\ of\ Deadlocks/sec\ |\ RID\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530592117
> Number\ of\ Deadlocks/sec\ |\ RowGroup\ |\ Locks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530623618
> Internal\ benefit\ |\ Buffer\ Pool\ |\ Memory\ Broker\ Clerks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530634618
> Internal\ benefit\ |\ Column\ store\ object\ pool\ |\ Memory\ Broker\ Clerks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530646418
> Memory\ broker\ clerk\ size\ |\ Buffer\ Pool\ |\ Memory\ Broker\ Clerks,servername=WIN8-DEV,type=Performance\ counters value=12217i 1453876412530662619
> Memory\ broker\ clerk\ size\ |\ Column\ store\ object\ pool\ |\ Memory\ Broker\ Clerks,servername=WIN8-DEV,type=Performance\ counters value=4i 1453876412530674519
> Periodic\ evictions\ (pages)\ |\ Buffer\ Pool\ |\ Memory\ Broker\ Clerks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530700020
> Periodic\ evictions\ (pages)\ |\ Column\ store\ object\ pool\ |\ Memory\ Broker\ Clerks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530711120
> Pressure\ evictions\ (pages/sec)\ |\ Buffer\ Pool\ |\ Memory\ Broker\ Clerks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530722720
> Pressure\ evictions\ (pages/sec)\ |\ Column\ store\ object\ pool\ |\ Memory\ Broker\ Clerks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530735621
> Simulation\ benefit\ |\ Buffer\ Pool\ |\ Memory\ Broker\ Clerks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530747421
> Simulation\ benefit\ |\ Column\ store\ object\ pool\ |\ Memory\ Broker\ Clerks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530773322
> Simulation\ size\ |\ Buffer\ Pool\ |\ Memory\ Broker\ Clerks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530784522
> Simulation\ size\ |\ Column\ store\ object\ pool\ |\ Memory\ Broker\ Clerks,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530796322
> Connection\ Memory\ (KB)\ |\ Memory\ Manager,servername=WIN8-DEV,type=Performance\ counters value=1416i 1453876412530807022
> Database\ Cache\ Memory\ (KB)\ |\ Memory\ Manager,servername=WIN8-DEV,type=Performance\ counters value=97736i 1453876412530818623
> External\ benefit\ of\ memory\ |\ Memory\ Manager,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530855524
> Free\ Memory\ (KB)\ |\ Memory\ Manager,servername=WIN8-DEV,type=Performance\ counters value=14080i 1453876412530867224
> Granted\ Workspace\ Memory\ (KB)\ |\ Memory\ Manager,servername=WIN8-DEV,type=Performance\ counters value=1024i 1453876412530881024
> Lock\ Blocks\ |\ Memory\ Manager,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530892625
> Lock\ Blocks\ Allocated\ |\ Memory\ Manager,servername=WIN8-DEV,type=Performance\ counters value=3050i 1453876412530915725
> Lock\ Memory\ (KB)\ |\ Memory\ Manager,servername=WIN8-DEV,type=Performance\ counters value=768i 1453876412530967127
> Lock\ Owner\ Blocks\ |\ Memory\ Manager,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412530983227
> Lock\ Owner\ Blocks\ Allocated\ |\ Memory\ Manager,servername=WIN8-DEV,type=Performance\ counters value=5550i 1453876412530997627
> Log\ Pool\ Memory\ (KB)\ |\ Memory\ Manager,servername=WIN8-DEV,type=Performance\ counters value=1552i 1453876412531013928
> Maximum\ Workspace\ Memory\ (KB)\ |\ Memory\ Manager,servername=WIN8-DEV,type=Performance\ counters value=1041480i 1453876412531026628
> Memory\ Grants\ Outstanding\ |\ Memory\ Manager,servername=WIN8-DEV,type=Performance\ counters value=1i 1453876412531059429
> Memory\ Grants\ Pending\ |\ Memory\ Manager,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412531070729
> Optimizer\ Memory\ (KB)\ |\ Memory\ Manager,servername=WIN8-DEV,type=Performance\ counters value=21264i 1453876412531082430
> Reserved\ Server\ Memory\ (KB)\ |\ Memory\ Manager,servername=WIN8-DEV,type=Performance\ counters value=1016i 1453876412531092830
> SQL\ Cache\ Memory\ (KB)\ |\ Memory\ Manager,servername=WIN8-DEV,type=Performance\ counters value=2072i 1453876412531128431
> Stolen\ Server\ Memory\ (KB)\ |\ Memory\ Manager,servername=WIN8-DEV,type=Performance\ counters value=314424i 1453876412531155632
> Target\ Server\ Memory\ (KB)\ |\ Memory\ Manager,servername=WIN8-DEV,type=Performance\ counters value=1536000i 1453876412531166632
> Total\ Server\ Memory\ (KB)\ |\ Memory\ Manager,servername=WIN8-DEV,type=Performance\ counters value=426240i 1453876412531178132
> Database\ Node\ Memory\ (KB)\ |\ 000\ |\ Memory\ Node,servername=WIN8-DEV,type=Performance\ counters value=97736i 1453876412531190433
> Foreign\ Node\ Memory\ (KB)\ |\ 000\ |\ Memory\ Node,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412531202033
> Free\ Node\ Memory\ (KB)\ |\ 000\ |\ Memory\ Node,servername=WIN8-DEV,type=Performance\ counters value=14072i 1453876412531225434
> Stolen\ Node\ Memory\ (KB)\ |\ 000\ |\ Memory\ Node,servername=WIN8-DEV,type=Performance\ counters value=314408i 1453876412531237434
> Target\ Node\ Memory\ (KB)\ |\ 000\ |\ Memory\ Node,servername=WIN8-DEV,type=Performance\ counters value=1535976i 1453876412531249334
> Total\ Node\ Memory\ (KB)\ |\ 000\ |\ Memory\ Node,servername=WIN8-DEV,type=Performance\ counters value=426216i 1453876412531262634
> Cache\ Hit\ Ratio\ |\ _Total\ |\ Plan\ Cache,servername=WIN8-DEV,type=Performance\ counters value=1i 1453876412531274235
> Cache\ Hit\ Ratio\ |\ Bound\ Trees\ |\ Plan\ Cache,servername=WIN8-DEV,type=Performance\ counters value=1i 1453876412531302236
> Cache\ Hit\ Ratio\ |\ Extended\ Stored\ Procedures\ |\ Plan\ Cache,servername=WIN8-DEV,type=Performance\ counters value=1i 1453876412531313536
> Cache\ Hit\ Ratio\ |\ Object\ Plans\ |\ Plan\ Cache,servername=WIN8-DEV,type=Performance\ counters value=1i 1453876412531326736
> Cache\ Hit\ Ratio\ |\ SQL\ Plans\ |\ Plan\ Cache,servername=WIN8-DEV,type=Performance\ counters value=1i 1453876412531337436
> Cache\ Hit\ Ratio\ |\ Temporary\ Tables\ &\ Table\ Variables\ |\ Plan\ Cache,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412531349137
> Cache\ Hit\ Ratio\ Base\ |\ _Total\ |\ Plan\ Cache,servername=WIN8-DEV,type=Performance\ counters value=-436i 1453876412531393738
> Cache\ Hit\ Ratio\ Base\ |\ Bound\ Trees\ |\ Plan\ Cache,servername=WIN8-DEV,type=Performance\ counters value=-449i 1453876412531406038
> Cache\ Hit\ Ratio\ Base\ |\ Extended\ Stored\ Procedures\ |\ Plan\ Cache,servername=WIN8-DEV,type=Performance\ counters value=5i 1453876412531418039
> Cache\ Hit\ Ratio\ Base\ |\ Object\ Plans\ |\ Plan\ Cache,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412531428839
> Cache\ Hit\ Ratio\ Base\ |\ SQL\ Plans\ |\ Plan\ Cache,servername=WIN8-DEV,type=Performance\ counters value=8i 1453876412531443539
> Cache\ Hit\ Ratio\ Base\ |\ Temporary\ Tables\ &\ Table\ Variables\ |\ Plan\ Cache,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412531488941
> Cache\ Object\ Counts\ |\ _Total\ |\ Plan\ Cache,servername=WIN8-DEV,type=Performance\ counters value=369i 1453876412531508441
> Cache\ Object\ Counts\ |\ Bound\ Trees\ |\ Plan\ Cache,servername=WIN8-DEV,type=Performance\ counters value=98i 1453876412531521041
> Cache\ Object\ Counts\ |\ Extended\ Stored\ Procedures\ |\ Plan\ Cache,servername=WIN8-DEV,type=Performance\ counters value=9i 1453876412531531942
> Cache\ Object\ Counts\ |\ Object\ Plans\ |\ Plan\ Cache,servername=WIN8-DEV,type=Performance\ counters value=38i 1453876412531543742
> Cache\ Object\ Counts\ |\ SQL\ Plans\ |\ Plan\ Cache,servername=WIN8-DEV,type=Performance\ counters value=207i 1453876412531572243
> Cache\ Object\ Counts\ |\ Temporary\ Tables\ &\ Table\ Variables\ |\ Plan\ Cache,servername=WIN8-DEV,type=Performance\ counters value=17i 1453876412531584843
> Cache\ Objects\ in\ use\ |\ _Total\ |\ Plan\ Cache,servername=WIN8-DEV,type=Performance\ counters value=3i 1453876412531597943
> Cache\ Objects\ in\ use\ |\ Bound\ Trees\ |\ Plan\ Cache,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412531608644
> Cache\ Objects\ in\ use\ |\ Extended\ Stored\ Procedures\ |\ Plan\ Cache,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412531620444
> Cache\ Objects\ in\ use\ |\ Object\ Plans\ |\ Plan\ Cache,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412531644845
> Cache\ Objects\ in\ use\ |\ SQL\ Plans\ |\ Plan\ Cache,servername=WIN8-DEV,type=Performance\ counters value=3i 1453876412531658345
> Cache\ Objects\ in\ use\ |\ Temporary\ Tables\ &\ Table\ Variables\ |\ Plan\ Cache,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412531673945
> Cache\ Pages\ |\ _Total\ |\ Plan\ Cache,servername=WIN8-DEV,type=Performance\ counters value=18687i 1453876412531685046
> Cache\ Pages\ |\ Bound\ Trees\ |\ Plan\ Cache,servername=WIN8-DEV,type=Performance\ counters value=1104i 1453876412531700446
> Cache\ Pages\ |\ Extended\ Stored\ Procedures\ |\ Plan\ Cache,servername=WIN8-DEV,type=Performance\ counters value=11i 1453876412531724847
> Cache\ Pages\ |\ Object\ Plans\ |\ Plan\ Cache,servername=WIN8-DEV,type=Performance\ counters value=2918i 1453876412531735747
> Cache\ Pages\ |\ SQL\ Plans\ |\ Plan\ Cache,servername=WIN8-DEV,type=Performance\ counters value=14650i 1453876412531747447
> Cache\ Pages\ |\ Temporary\ Tables\ &\ Table\ Variables\ |\ Plan\ Cache,servername=WIN8-DEV,type=Performance\ counters value=4i 1453876412531757948
> Active\ memory\ grant\ amount\ (KB)\ |\ default\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412531769848
> Active\ memory\ grant\ amount\ (KB)\ |\ internal\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412531798249
> Active\ memory\ grants\ count\ |\ default\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412531809749
> Active\ memory\ grants\ count\ |\ internal\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412531821449
> Avg\ Disk\ Read\ IO\ (ms)\ |\ default\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412531832350
> Avg\ Disk\ Read\ IO\ (ms)\ |\ internal\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412531846150
> Avg\ Disk\ Read\ IO\ (ms)\ Base\ |\ default\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412531889551
> Avg\ Disk\ Read\ IO\ (ms)\ Base\ |\ internal\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412531902551
> Avg\ Disk\ Write\ IO\ (ms)\ |\ default\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412531931052
> Avg\ Disk\ Write\ IO\ (ms)\ |\ internal\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412531942553
> Avg\ Disk\ Write\ IO\ (ms)\ Base\ |\ default\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412531954453
> Avg\ Disk\ Write\ IO\ (ms)\ Base\ |\ internal\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412531982654
> Cache\ memory\ target\ (KB)\ |\ default\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=1231200i 1453876412531994254
> Cache\ memory\ target\ (KB)\ |\ internal\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=1122336i 1453876412532007754
> Compile\ memory\ target\ (KB)\ |\ default\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=1171672i 1453876412532018455
> Compile\ memory\ target\ (KB)\ |\ internal\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=1173240i 1453876412532030155
> CPU\ control\ effect\ %\ |\ default\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=31i 1453876412532054056
> CPU\ control\ effect\ %\ |\ internal\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412532065056
> CPU\ usage\ %\ |\ default\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412532076556
> CPU\ usage\ %\ |\ internal\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412532091256
> CPU\ usage\ %\ base\ |\ default\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412532122657
> CPU\ usage\ %\ base\ |\ internal\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412532153158
> CPU\ usage\ target\ %\ |\ default\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=56i 1453876412532164558
> CPU\ usage\ target\ %\ |\ internal\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412532176059
> Disk\ Read\ Bytes/sec\ |\ default\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412532188659
> Disk\ Read\ Bytes/sec\ |\ internal\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412532200459
> Disk\ Read\ IO\ Throttled/sec\ |\ default\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412532224560
> Disk\ Read\ IO\ Throttled/sec\ |\ internal\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412532235560
> Disk\ Read\ IO/sec\ |\ default\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412532252661
> Disk\ Read\ IO/sec\ |\ internal\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412532263461
> Disk\ Write\ Bytes/sec\ |\ default\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412532275061
> Disk\ Write\ Bytes/sec\ |\ internal\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412532312762
> Disk\ Write\ IO\ Throttled/sec\ |\ default\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412532324463
> Disk\ Write\ IO\ Throttled/sec\ |\ internal\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412532336363
> Disk\ Write\ IO/sec\ |\ default\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412532348863
> Disk\ Write\ IO/sec\ |\ internal\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412532361464
> Max\ memory\ (KB)\ |\ default\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=1459200i 1453876412532387564
> Max\ memory\ (KB)\ |\ internal\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=1459200i 1453876412532398765
> Memory\ grant\ timeouts/sec\ |\ default\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412532410465
> Memory\ grant\ timeouts/sec\ |\ internal\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412532421265
> Memory\ grants/sec\ |\ default\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=16i 1453876412532436666
> Memory\ grants/sec\ |\ internal\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412532461866
> Pending\ memory\ grants\ count\ |\ default\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412532472567
> Pending\ memory\ grants\ count\ |\ internal\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412532484667
> Query\ exec\ memory\ target\ (KB)\ |\ default\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=1041480i 1453876412532498167
> Query\ exec\ memory\ target\ (KB)\ |\ internal\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=1039560i 1453876412532510168
> Target\ memory\ (KB)\ |\ default\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=1459200i 1453876412532534968
> Target\ memory\ (KB)\ |\ internal\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=1459200i 1453876412532546269
> Used\ memory\ (KB)\ |\ default\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=189552i 1453876412532557769
> Used\ memory\ (KB)\ |\ internal\ |\ Resource\ Pool\ Stats,servername=WIN8-DEV,type=Performance\ counters value=125856i 1453876412532569269
> Errors/sec\ |\ _Total\ |\ SQL\ Errors,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412532580769
> Errors/sec\ |\ DB\ Offline\ Errors\ |\ SQL\ Errors,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412532610970
> Errors/sec\ |\ Info\ Errors\ |\ SQL\ Errors,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412532622271
> Errors/sec\ |\ Kill\ Connection\ Errors\ |\ SQL\ Errors,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412532633671
> Errors/sec\ |\ User\ Errors\ |\ SQL\ Errors,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412532644371
> Auto-Param\ Attempts/sec\ |\ SQL\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412532655671
> Batch\ Requests/sec\ |\ SQL\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=3i 1453876412532684772
> Failed\ Auto-Params/sec\ |\ SQL\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412532696073
> Forced\ Parameterizations/sec\ |\ SQL\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412532707573
> Guided\ plan\ executions/sec\ |\ SQL\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412532717873
> Misguided\ plan\ executions/sec\ |\ SQL\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412532729073
> Safe\ Auto-Params/sec\ |\ SQL\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412532764674
> SQL\ Attention\ rate\ |\ SQL\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412532776175
> SQL\ Compilations/sec\ |\ SQL\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=32i 1453876412532788575
> SQL\ Re-Compilations/sec\ |\ SQL\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=12i 1453876412532798875
> Unsafe\ Auto-Params/sec\ |\ SQL\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412532809976
> Free\ Space\ in\ tempdb\ (KB)\ |\ Transactions,servername=WIN8-DEV,type=Performance\ counters value=1044992i 1453876412532834776
> Longest\ Transaction\ Running\ Time\ |\ Transactions,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412532846077
> NonSnapshot\ Version\ Transactions\ |\ Transactions,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412532857377
> Snapshot\ Transactions\ |\ Transactions,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412532868977
> Transactions\ |\ Transactions,servername=WIN8-DEV,type=Performance\ counters value=14i 1453876412532883178
> Update\ conflict\ ratio\ |\ Transactions,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412532925579
> Update\ conflict\ ratio\ base\ |\ Transactions,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412532939679
> Update\ Snapshot\ Transactions\ |\ Transactions,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412532951879
> Version\ Cleanup\ rate\ (KB/s)\ |\ Transactions,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412532962180
> Version\ Generation\ rate\ (KB/s)\ |\ Transactions,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412532976080
> Version\ Store\ Size\ (KB)\ |\ Transactions,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412532999681
> Version\ Store\ unit\ count\ |\ Transactions,servername=WIN8-DEV,type=Performance\ counters value=2i 1453876412533011581
> Version\ Store\ unit\ creation\ |\ Transactions,servername=WIN8-DEV,type=Performance\ counters value=12i 1453876412533023081
> Version\ Store\ unit\ truncation\ |\ Transactions,servername=WIN8-DEV,type=Performance\ counters value=10i 1453876412533033482
> Query\ |\ User\ counter\ 1\ |\ User\ Settable,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533044782
> Query\ |\ User\ counter\ 10\ |\ User\ Settable,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533068482
> Query\ |\ User\ counter\ 2\ |\ User\ Settable,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533079683
> Query\ |\ User\ counter\ 3\ |\ User\ Settable,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533090983
> Query\ |\ User\ counter\ 4\ |\ User\ Settable,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533104283
> Query\ |\ User\ counter\ 5\ |\ User\ Settable,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533119384
> Query\ |\ User\ counter\ 6\ |\ User\ Settable,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533143584
> Query\ |\ User\ counter\ 7\ |\ User\ Settable,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533154485
> Query\ |\ User\ counter\ 8\ |\ User\ Settable,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533165885
> Query\ |\ User\ counter\ 9\ |\ User\ Settable,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533176285
> Lock\ waits\ |\ Average\ wait\ time\ (ms)\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533188986
> Lock\ waits\ |\ Cumulative\ wait\ time\ (ms)\ per\ second\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533231087
> Lock\ waits\ |\ Waits\ in\ progress\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533243487
> Lock\ waits\ |\ Waits\ started\ per\ second\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533260588
> Log\ buffer\ waits\ |\ Average\ wait\ time\ (ms)\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533273688
> Log\ buffer\ waits\ |\ Cumulative\ wait\ time\ (ms)\ per\ second\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533288388
> Log\ buffer\ waits\ |\ Waits\ in\ progress\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533316189
> Log\ buffer\ waits\ |\ Waits\ started\ per\ second\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533327689
> Log\ write\ waits\ |\ Average\ wait\ time\ (ms)\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533339290
> Log\ write\ waits\ |\ Cumulative\ wait\ time\ (ms)\ per\ second\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533349990
> Log\ write\ waits\ |\ Waits\ in\ progress\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533361690
> Log\ write\ waits\ |\ Waits\ started\ per\ second\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533388291
> Memory\ grant\ queue\ waits\ |\ Average\ wait\ time\ (ms)\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533402291
> Memory\ grant\ queue\ waits\ |\ Cumulative\ wait\ time\ (ms)\ per\ second\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533415792
> Memory\ grant\ queue\ waits\ |\ Waits\ in\ progress\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533426892
> Memory\ grant\ queue\ waits\ |\ Waits\ started\ per\ second\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533438392
> Network\ IO\ waits\ |\ Average\ wait\ time\ (ms)\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533463393
> Network\ IO\ waits\ |\ Cumulative\ wait\ time\ (ms)\ per\ second\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533474493
> Network\ IO\ waits\ |\ Waits\ in\ progress\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533486194
> Network\ IO\ waits\ |\ Waits\ started\ per\ second\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533496594
> Non-Page\ latch\ waits\ |\ Average\ wait\ time\ (ms)\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533511394
> Non-Page\ latch\ waits\ |\ Cumulative\ wait\ time\ (ms)\ per\ second\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533536695
> Non-Page\ latch\ waits\ |\ Waits\ in\ progress\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533547895
> Non-Page\ latch\ waits\ |\ Waits\ started\ per\ second\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533561496
> Page\ IO\ latch\ waits\ |\ Average\ wait\ time\ (ms)\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533572096
> Page\ IO\ latch\ waits\ |\ Cumulative\ wait\ time\ (ms)\ per\ second\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533583896
> Page\ IO\ latch\ waits\ |\ Waits\ in\ progress\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533629397
> Page\ IO\ latch\ waits\ |\ Waits\ started\ per\ second\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533641398
> Page\ latch\ waits\ |\ Average\ wait\ time\ (ms)\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533653298
> Page\ latch\ waits\ |\ Cumulative\ wait\ time\ (ms)\ per\ second\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533663798
> Page\ latch\ waits\ |\ Waits\ in\ progress\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533675899
> Page\ latch\ waits\ |\ Waits\ started\ per\ second\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533705599
> Thread-safe\ memory\ objects\ waits\ |\ Average\ wait\ time\ (ms)\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533716800
> Thread-safe\ memory\ objects\ waits\ |\ Cumulative\ wait\ time\ (ms)\ per\ second\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533728700
> Thread-safe\ memory\ objects\ waits\ |\ Waits\ in\ progress\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533739700
> Thread-safe\ memory\ objects\ waits\ |\ Waits\ started\ per\ second\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533751501
> Transaction\ ownership\ waits\ |\ Average\ wait\ time\ (ms)\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533778501
> Transaction\ ownership\ waits\ |\ Cumulative\ wait\ time\ (ms)\ per\ second\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533801602
> Transaction\ ownership\ waits\ |\ Waits\ in\ progress\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533818402
> Transaction\ ownership\ waits\ |\ Waits\ started\ per\ second\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533829403
> Wait\ for\ the\ worker\ |\ Average\ wait\ time\ (ms)\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533841103
> Wait\ for\ the\ worker\ |\ Cumulative\ wait\ time\ (ms)\ per\ second\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533868204
> Wait\ for\ the\ worker\ |\ Waits\ in\ progress\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533879704
> Wait\ for\ the\ worker\ |\ Waits\ started\ per\ second\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533891604
> Workspace\ synchronization\ waits\ |\ Average\ wait\ time\ (ms)\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533902305
> Workspace\ synchronization\ waits\ |\ Cumulative\ wait\ time\ (ms)\ per\ second\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533929505
> Workspace\ synchronization\ waits\ |\ Waits\ in\ progress\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533976207
> Workspace\ synchronization\ waits\ |\ Waits\ started\ per\ second\ |\ Wait\ Statistics,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412533988307
> Active\ parallel\ threads\ |\ default\ |\ Workload\ Group\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534000107
> Active\ parallel\ threads\ |\ internal\ |\ Workload\ Group\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534013008
> Active\ requests\ |\ default\ |\ Workload\ Group\ Stats,servername=WIN8-DEV,type=Performance\ counters value=3i 1453876412534024608
> Active\ requests\ |\ internal\ |\ Workload\ Group\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534049608
> Blocked\ tasks\ |\ default\ |\ Workload\ Group\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534061609
> Blocked\ tasks\ |\ internal\ |\ Workload\ Group\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534073209
> CPU\ usage\ %\ |\ default\ |\ Workload\ Group\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534083709
> CPU\ usage\ %\ |\ internal\ |\ Workload\ Group\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534095410
> CPU\ usage\ %\ base\ |\ default\ |\ Workload\ Group\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534134711
> CPU\ usage\ %\ base\ |\ internal\ |\ Workload\ Group\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534148011
> Max\ request\ cpu\ time\ (ms)\ |\ default\ |\ Workload\ Group\ Stats,servername=WIN8-DEV,type=Performance\ counters value=731i 1453876412534160011
> Max\ request\ cpu\ time\ (ms)\ |\ internal\ |\ Workload\ Group\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534170612
> Max\ request\ memory\ grant\ (KB)\ |\ default\ |\ Workload\ Group\ Stats,servername=WIN8-DEV,type=Performance\ counters value=11056i 1453876412534183912
> Max\ request\ memory\ grant\ (KB)\ |\ internal\ |\ Workload\ Group\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534208113
> Query\ optimizations/sec\ |\ default\ |\ Workload\ Group\ Stats,servername=WIN8-DEV,type=Performance\ counters value=30i 1453876412534219313
> Query\ optimizations/sec\ |\ internal\ |\ Workload\ Group\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534232113
> Queued\ requests\ |\ default\ |\ Workload\ Group\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534242914
> Queued\ requests\ |\ internal\ |\ Workload\ Group\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534254514
> Reduced\ memory\ grants/sec\ |\ default\ |\ Workload\ Group\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534284615
> Reduced\ memory\ grants/sec\ |\ internal\ |\ Workload\ Group\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534296215
> Requests\ completed/sec\ |\ default\ |\ Workload\ Group\ Stats,servername=WIN8-DEV,type=Performance\ counters value=7i 1453876412534307715
> Requests\ completed/sec\ |\ internal\ |\ Workload\ Group\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534323716
> Suboptimal\ plans/sec\ |\ default\ |\ Workload\ Group\ Stats,servername=WIN8-DEV,type=Performance\ counters value=4i 1453876412534335316
> Suboptimal\ plans/sec\ |\ internal\ |\ Workload\ Group\ Stats,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534374117
> Cursor\ deletes/sec\ |\ MSSQLSERVER\ |\ XTP\ Cursors,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534385817
> Cursor\ inserts/sec\ |\ MSSQLSERVER\ |\ XTP\ Cursors,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534397818
> Cursor\ scans\ started/sec\ |\ MSSQLSERVER\ |\ XTP\ Cursors,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534408518
> Cursor\ unique\ violations/sec\ |\ MSSQLSERVER\ |\ XTP\ Cursors,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534420118
> Cursor\ updates/sec\ |\ MSSQLSERVER\ |\ XTP\ Cursors,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534445819
> Cursor\ write\ conflicts/sec\ |\ MSSQLSERVER\ |\ XTP\ Cursors,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534457119
> Dusty\ corner\ scan\ retries/sec\ (user-issued)\ |\ MSSQLSERVER\ |\ XTP\ Cursors,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534468620
> Expired\ rows\ removed/sec\ |\ MSSQLSERVER\ |\ XTP\ Cursors,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534482420
> Expired\ rows\ touched/sec\ |\ MSSQLSERVER\ |\ XTP\ Cursors,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534494220
> Rows\ returned/sec\ |\ MSSQLSERVER\ |\ XTP\ Cursors,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534521721
> Rows\ touched/sec\ |\ MSSQLSERVER\ |\ XTP\ Cursors,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534532821
> Tentatively-deleted\ rows\ touched/sec\ |\ MSSQLSERVER\ |\ XTP\ Cursors,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534544222
> Dusty\ corner\ scan\ retries/sec\ (GC-issued)\ |\ MSSQLSERVER\ |\ XTP\ Garbage\ Collection,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534556122
> Main\ GC\ work\ items/sec\ |\ MSSQLSERVER\ |\ XTP\ Garbage\ Collection,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534569422
> Parallel\ GC\ work\ item/sec\ |\ MSSQLSERVER\ |\ XTP\ Garbage\ Collection,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534597823
> Rows\ processed/sec\ |\ MSSQLSERVER\ |\ XTP\ Garbage\ Collection,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534608923
> Rows\ processed/sec\ (first\ in\ bucket\ and\ removed)\ |\ MSSQLSERVER\ |\ XTP\ Garbage\ Collection,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534620624
> Rows\ processed/sec\ (first\ in\ bucket)\ |\ MSSQLSERVER\ |\ XTP\ Garbage\ Collection,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534631824
> Rows\ processed/sec\ (marked\ for\ unlink)\ |\ MSSQLSERVER\ |\ XTP\ Garbage\ Collection,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534643624
> Rows\ processed/sec\ (no\ sweep\ needed)\ |\ MSSQLSERVER\ |\ XTP\ Garbage\ Collection,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534673825
> Sweep\ expired\ rows\ removed/sec\ |\ MSSQLSERVER\ |\ XTP\ Garbage\ Collection,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534687525
> Sweep\ expired\ rows\ touched/sec\ |\ MSSQLSERVER\ |\ XTP\ Garbage\ Collection,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534699526
> Sweep\ expiring\ rows\ touched/sec\ |\ MSSQLSERVER\ |\ XTP\ Garbage\ Collection,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534714626
> Sweep\ rows\ touched/sec\ |\ MSSQLSERVER\ |\ XTP\ Garbage\ Collection,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534727026
> Sweep\ scans\ started/sec\ |\ MSSQLSERVER\ |\ XTP\ Garbage\ Collection,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534765628
> Dusty\ corner\ scan\ retries/sec\ (Phantom-issued)\ |\ MSSQLSERVER\ |\ XTP\ Phantom\ Processor,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534777328
> Phantom\ expired\ rows\ removed/sec\ |\ MSSQLSERVER\ |\ XTP\ Phantom\ Processor,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534789428
> Phantom\ expired\ rows\ touched/sec\ |\ MSSQLSERVER\ |\ XTP\ Phantom\ Processor,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534800128
> Phantom\ expiring\ rows\ touched/sec\ |\ MSSQLSERVER\ |\ XTP\ Phantom\ Processor,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534811529
> Phantom\ rows\ touched/sec\ |\ MSSQLSERVER\ |\ XTP\ Phantom\ Processor,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534835629
> Phantom\ scans\ started/sec\ |\ MSSQLSERVER\ |\ XTP\ Phantom\ Processor,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534846630
> Checkpoints\ Closed\ |\ MSSQLSERVER\ |\ XTP\ Storage,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534858230
> Checkpoints\ Completed\ |\ MSSQLSERVER\ |\ XTP\ Storage,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534869630
> Core\ Merges\ Completed\ |\ MSSQLSERVER\ |\ XTP\ Storage,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534881131
> Merge\ Policy\ Evaluations\ |\ MSSQLSERVER\ |\ XTP\ Storage,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534919832
> Merge\ Requests\ Outstanding\ |\ MSSQLSERVER\ |\ XTP\ Storage,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534934132
> Merges\ Abandoned\ |\ MSSQLSERVER\ |\ XTP\ Storage,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534946332
> Merges\ Installed\ |\ MSSQLSERVER\ |\ XTP\ Storage,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534959133
> Total\ Files\ Merged\ |\ MSSQLSERVER\ |\ XTP\ Storage,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412534970533
> Log\ bytes\ written/sec\ |\ MSSQLSERVER\ |\ XTP\ Transaction\ Log,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412535001434
> Log\ records\ written/sec\ |\ MSSQLSERVER\ |\ XTP\ Transaction\ Log,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412535014034
> Cascading\ aborts/sec\ |\ MSSQLSERVER\ |\ XTP\ Transactions,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412535025934
> Commit\ dependencies\ taken/sec\ |\ MSSQLSERVER\ |\ XTP\ Transactions,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412535036635
> Read-only\ transactions\ prepared/sec\ |\ MSSQLSERVER\ |\ XTP\ Transactions,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412535051635
> Save\ point\ refreshes/sec\ |\ MSSQLSERVER\ |\ XTP\ Transactions,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412535076736
> Save\ point\ rollbacks/sec\ |\ MSSQLSERVER\ |\ XTP\ Transactions,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412535088536
> Save\ points\ created/sec\ |\ MSSQLSERVER\ |\ XTP\ Transactions,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412535102536
> Transaction\ validation\ failures/sec\ |\ MSSQLSERVER\ |\ XTP\ Transactions,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412535116437
> Transactions\ aborted\ by\ user/sec\ |\ MSSQLSERVER\ |\ XTP\ Transactions,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412535128337
> Transactions\ aborted/sec\ |\ MSSQLSERVER\ |\ XTP\ Transactions,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412535146038
> Transactions\ created/sec\ |\ MSSQLSERVER\ |\ XTP\ Transactions,servername=WIN8-DEV,type=Performance\ counters value=0i 1453876412535159838
> Wait\ time\ (ms),servername=WIN8-DEV,type=Wait\ stats Buffer=0i,CLR=0i,I/O=0i,Latch=13i,Lock=0i,Memory=0i,Network=0i,Other=98i,SQLOS=115i,Service\ broker=0i,Total=226i,XEvent=0i 1453876416521666190
> Wait\ tasks,servername=WIN8-DEV,type=Wait\ stats Buffer=0i,CLR=0i,I/O=0i,Latch=22i,Lock=0i,Memory=0i,Network=0i,Other=30i,SQLOS=80i,Service\ broker=0i,Total=132i,XEvent=0i 1453876416521726692
> Log\ writes\ (bytes/sec),servername=WIN8-DEV,type=Database\ IO AdventureWorks2014=0i,Australian=0i,DOC.Azure=0i,ResumeCloud=0i,Total=86835i,master=0i,model=0i,msdb=0i,ngMon=0i,tempdb=86835i 1453876416602799144
> Rows\ writes\ (bytes/sec),servername=WIN8-DEV,type=Database\ IO AdventureWorks2014=0i,Australian=0i,DOC.Azure=0i,ResumeCloud=0i,Total=0i,master=0i,model=0i,msdb=0i,ngMon=0i,tempdb=0i 1453876416602853046
> Log\ reads\ (bytes/sec),servername=WIN8-DEV,type=Database\ IO AdventureWorks2014=0i,Australian=0i,DOC.Azure=0i,ResumeCloud=0i,Total=0i,master=0i,model=0i,msdb=0i,ngMon=0i,tempdb=0i 1453876416602897047
> Rows\ reads\ (bytes/sec),servername=WIN8-DEV,type=Database\ IO AdventureWorks2014=0i,Australian=0i,DOC.Azure=0i,ResumeCloud=0i,Total=6553i,master=0i,model=0i,msdb=0i,ngMon=0i,tempdb=6553i 1453876416602943248
> Log\ (writes/sec),servername=WIN8-DEV,type=Database\ IO AdventureWorks2014=0i,Australian=0i,DOC.Azure=0i,ResumeCloud=0i,Total=1i,master=0i,model=0i,msdb=0i,ngMon=0i,tempdb=1i 1453876416602969849
> Rows\ (writes/sec),servername=WIN8-DEV,type=Database\ IO AdventureWorks2014=0i,Australian=0i,DOC.Azure=0i,ResumeCloud=0i,Total=0i,master=0i,model=0i,msdb=0i,ngMon=0i,tempdb=0i 1453876416602993950
> Log\ (reads/sec),servername=WIN8-DEV,type=Database\ IO AdventureWorks2014=0i,Australian=0i,DOC.Azure=0i,ResumeCloud=0i,Total=0i,master=0i,model=0i,msdb=0i,ngMon=0i,tempdb=0i 1453876416603018350
> Rows\ (reads/sec),servername=WIN8-DEV,type=Database\ IO AdventureWorks2014=0i,Australian=0i,DOC.Azure=0i,ResumeCloud=0i,Total=0i,master=0i,model=0i,msdb=0i,ngMon=0i,tempdb=0i 1453876416603047751
``` 
