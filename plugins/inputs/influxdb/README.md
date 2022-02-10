# InfluxDB Input Plugin

The InfluxDB plugin will collect metrics on the given InfluxDB servers. Read our
[documentation](https://docs.influxdata.com/platform/monitoring/influxdata-platform/tools/measurements-internal/)
for detailed information about `influxdb` metrics.

This plugin can also gather metrics from endpoints that expose
InfluxDB-formatted endpoints. See below for more information.

## Configuration

```toml
# Read InfluxDB-formatted JSON metrics from one or more HTTP endpoints
[[inputs.influxdb]]
  ## Works with InfluxDB debug endpoints out of the box,
  ## but other services can use this format too.
  ## See the influxdb plugin's README for more details.

  ## Multiple URLs from which to read InfluxDB-formatted JSON
  ## Default is "http://localhost:8086/debug/vars".
  urls = [
    "http://localhost:8086/debug/vars"
  ]

  ## Username and password to send using HTTP Basic Authentication.
  # username = ""
  # password = ""

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## http request & header timeout
  timeout = "5s"
```

## Measurements & Fields

**Note:** The measurements and fields included in this plugin are dynamically built from the InfluxDB source, and may vary between versions:

- **influxdb_ae** _(Enterprise Only)_ : Statistics related to the Anti-Entropy (AE) engine in InfluxDB Enterprise clusters.
  - **bytesRx**: Number of bytes received by the data node.
  - **errors**: Total number of anti-entropy jobs that have resulted in errors.
  - **jobs**: Total number of jobs executed by the data node.
  - **jobsActive**: Number of active (currently executing) jobs.
- **influxdb_cluster** _(Enterprise Only)_ : Statistics related to the clustering features of the data nodes in InfluxDB Enterprise clusters.
  - **copyShardReq**: Number of internal requests made to copy a shard from one data node to another.
  - **createIteratorReq**: Number of read requests from other data nodes in the cluster.
  - **expandSourcesReq**: Number of remote node requests made to find measurements on this node that match a particular regular expression.
  - **fieldDimensionsReq**: Number of remote node requests for information about the fields and associated types, and tag keys of measurements on this data node.
  - **iteratorCostReq**: Number of internal requests for iterator cost.
  - **openConnections**: Tracks the number of open connections being handled by the data node
    (including logical connections multiplexed onto a single yamux connection).
  - **removeShardReq**: Number of internal requests to delete a shard from this data node. Exclusively incremented by use of the influxd-ctl remove shard command.
  - **writeShardFail**: Total number of internal write requests from a remote node that failed.
  - **writeShardPointsReq**: Number of points in every internal write request from any remote node, regardless of success.
  - **writeShardReq**: Number of internal write requests from a remote data node, regardless of success.
- **influxdb_cq**: Metrics related to continuous queries (CQs).
  - **queryFail**: Total number of continuous queries that executed but failed.
  - **queryOk**: Total number of continuous queries that executed successfully.
- **influxdb_database**: Database metrics are collected from.
  - **numMeasurements**: Current number of measurements in the specified database.
  - **numSeries**: Current series cardinality of the specified database.
- **influxdb_hh** _(Enterprise Only)_ : Events resulting in new hinted handoff (HH) processors in InfluxDB Enterprise clusters.
  - **writeShardReq**: Number of initial write requests handled by the hinted handoff engine for a remote node.
  - **writeShardReqPoints**: Number of write requests for each point in the initial request to the hinted handoff engine for a remote node.
- **influxdb_hh_database** _(Enterprise Only)_ : Aggregates all hinted handoff queues for a single database and node.
  - **bytesRead**: Size, in bytes, of points read from the hinted handoff queue and sent to its destination data node.
  - **bytesWritten**: Total number of bytes written to the hinted handoff queue.
  - **queueBytes**: Total number of bytes remaining in the hinted handoff queue.
  - **queueDepth**: Total number of segments in the hinted handoff queue. The HH queue is a sequence of 10MB “segment” files.
  - **writeBlocked**: Number of writes blocked because the number of concurrent HH requests exceeds the limit.
  - **writeDropped**: Number of writes dropped from the HH queue because the write appeared to be corrupted.
  - **writeNodeReq**: Total number of write requests that succeeded in writing a batch to the destination node.
  - **writeNodeReqFail**: Total number of write requests that failed in writing a batch of data from the hinted handoff queue to the destination node.
  - **writeNodeReqPoints**: Total number of points successfully written from the HH queue to the destination node fr
  - **writeShardReq**: Total number of every write batch request enqueued into the hinted handoff queue.
  - **writeShardReqPoints**: Total number of points enqueued into the hinted handoff queue.  
- **influxdb_hh_processor** _(Enterprise Only)_: Statistics stored for a single queue (shard).
  - **bytesRead**: Size, in bytes, of points read from the hinted handoff queue and sent to its destination data node.
  - **bytesWritten**: Total number of bytes written to the hinted handoff queue.
  - **queueBytes**: Total number of bytes remaining in the hinted handoff queue.
  - **queueDepth**: Total number of segments in the hinted handoff queue. The HH queue is a sequence of 10MB “segment” files.
  - **writeBlocked**: Number of writes blocked because the number of concurrent HH requests exceeds the limit.
  - **writeDropped**: Number of writes dropped from the HH queue because the write appeared to be corrupted.
  - **writeNodeReq**: Total number of write requests that succeeded in writing a batch to the destination node.
  - **writeNodeReqFail**: Total number of write requests that failed in writing a batch of data from the hinted handoff queue to the destination node.
  - **writeNodeReqPoints**: Total number of points successfully written from the HH queue to the destination node fr
  - **writeShardReq**: Total number of every write batch request enqueued into the hinted handoff queue.
  - **writeShardReqPoints**: Total number of points enqueued into the hinted handoff queue.
- **influxdb_httpd**: Metrics related to the InfluxDB HTTP server.
  - **authFail**: Number of HTTP requests that were aborted due to authentication being required, but not supplied or incorrect.
  - **clientError**: Number of HTTP responses due to client errors, with a 4XX HTTP status code.
  - **fluxQueryReq**: Number of Flux query requests served.
  - **fluxQueryReqDurationNs**: Duration (wall-time), in nanoseconds, spent executing Flux query requests.
  - **pingReq**: Number of times InfluxDB HTTP server served the /ping HTTP endpoint.
  - **pointsWrittenDropped**: Number of points dropped by the storage engine.
  - **pointsWrittenFail**: Number of points accepted by the HTTP /write endpoint, but unable to be persisted.
  - **pointsWrittenOK**: Number of points successfully accepted and persisted by the HTTP /write endpoint.
  - **promReadReq**: Number of read requests to the Prometheus /read endpoint.
  - **promWriteReq**: Number of write requests to the Prometheus /write endpoint.
  - **queryReq**: Number of query requests.
  - **queryReqDurationNs**: Total query request duration, in nanosecond (ns).
  - **queryRespBytes**: Total number of bytes returned in query responses.
  - **recoveredPanics**: Total number of panics recovered by the HTTP handler.
  - **req**: Total number of HTTP requests served.
  - **reqActive**: Number of currently active requests.
  - **reqDurationNs**: Duration (wall time), in nanoseconds, spent inside HTTP requests.
  - **serverError**: Number of HTTP responses due to server errors.
  - **statusReq**: Number of status requests served using the HTTP /status endpoint.
  - **valuesWrittenOK**: Number of values (fields) successfully accepted and persisted by the HTTP /write endpoint.
  - **writeReq**: Number of write requests served using the HTTP /write endpoint.
  - **writeReqActive**: Number of currently active write requests.
  - **writeReqBytes**: Total number of bytes of line protocol data received by write requests, using the HTTP /write endpoint.
  - **writeReqDurationNs**: Duration, in nanoseconds, of write requests served using the /write HTTP endpoint.
- **influxdb_memstats**: Statistics about the memory allocator in the specified database.
  - **Alloc**: Number of bytes allocated to heap objects.
  - **BuckHashSys**: Number of bytes of memory in profiling bucket hash tables.
  - **Frees**: Cumulative count of heap objects freed.
  - **GCCPUFraction**: fraction of InfluxDB's available CPU time used by the garbage collector (GC) since InfluxDB started.
  - **GCSys**: Number of bytes of memory in garbage collection metadata.
  - **HeapAlloc**: Number of bytes of allocated heap objects.
  - **HeapIdle**: Number of bytes in idle (unused) spans.
  - **HeapInuse**: Number of bytes in in-use spans.
  - **HeapObjects**: Number of allocated heap objects.
  - **HeapReleased**: Number of bytes of physical memory returned to the OS.
  - **HeapSys**: Number of bytes of heap memory obtained from the OS.
  - **LastGC**: Time the last garbage collection finished.
  - **Lookups**: Number of pointer lookups performed by the runtime.
  - **MCacheInuse**: Number of bytes of allocated mcache structures.
  - **MCacheSys**: Number of bytes of memory obtained from the OS for mcache structures.
  - **MSpanInuse**: Number of bytes of allocated mspan structures.
  - **MSpanSys**: Number of bytes of memory obtained from the OS for mspan structures.
  - **Mallocs**: Cumulative count of heap objects allocated.
  - **NextGC**: Target heap size of the next GC cycle.
  - **NumForcedGC**: Number of GC cycles that were forced by the application calling the GC function.
  - **NumGC**: Number of completed GC cycles.
  - **OtherSys**: Number of bytes of memory in miscellaneous off-heap runtime allocations.
  - **PauseTotalNs**: Cumulative nanoseconds in GC stop-the-world pauses since the program started.
  - **StackInuse**: Number of bytes in stack spans.
  - **StackSys**: Number of bytes of stack memory obtained from the OS.
  - **Sys**: Total bytes of memory obtained from the OS.
  - **TotalAlloc**: Cumulative bytes allocated for heap objects.
- **influxdb_queryExecutor**: Metrics related to usage of the Query Executor of the InfluxDB engine.
  - **queriesActive**: Number of active queries currently being handled.
  - **queriesExecuted**: Number of queries executed (started).
  - **queriesFinished**: Number of queries that have finished executing.
  - **queryDurationNs**: Total duration, in nanoseconds, of executed queries.
  - **recoveredPanics**: Number of panics recovered by the Query Executor.
- **influxdb_rpc** _(Enterprise Only)_ : Statistics related to the use of RPC calls within InfluxDB Enterprise clusters.
  - **idleStreams**: Number of idle multiplexed streams across all live TCP connections.
  - **liveConnections**: Current number of live TCP connections to other nodes.
  - **liveStreams**: Current number of live multiplexed streams across all live TCP connections.
  - **rpcCalls**: Total number of RPC calls made to remote nodes.
  - **rpcFailures**: Total number of RPC failures, which are RPCs that did not recover.
  - **rpcReadBytes**: Total number of RPC bytes read.
  - **rpcRetries**: Total number of RPC calls that retried at least once.
  - **rpcWriteBytes**: Total number of RPC bytes written.
  - **singleUse**: Total number of single-use connections opened using Dial.
  - **singleUseOpen**: Number of single-use connections currently open.
  - **totalConnections**: Total number of TCP connections that have been established.
  - **totalStreams**: Total number of streams established.
- **influxdb_runtime**: Subset of memstat record statistics for the Go memory allocator.
  - **Alloc**: Currently allocated number of bytes of heap objects.
  - **Frees**: Cumulative number of freed (live) heap objects.
  - **HeapAlloc**: Size, in bytes, of all heap objects.
  - **HeapIdle**: Number of bytes of idle heap objects.
  - **HeapInUse**: Number of bytes in in-use spans.
  - **HeapObjects**: Number of allocated heap objects.
  - **HeapReleased**: Number of bytes of physical memory returned to the OS.
  - **HeapSys**: Number of bytes of heap memory obtained from the OS. Measures the amount of virtual address space reserved for the heap.
  - **Lookups**: Number of pointer lookups performed by the runtime. Primarily useful for debugging runtime internals.
  - **Mallocs**: Total number of heap objects allocated. The total number of live objects is Frees.
  - **NumGC**: Number of completed GC (garbage collection) cycles.
  - **NumGoroutine**: Total number of Go routines.
  - **PauseTotalNs**: Total duration, in nanoseconds, of total GC (garbage collection) pauses.
  - **Sys**: Total number of bytes of memory obtained from the OS. Measures the virtual address space reserved by the Go runtime for the heap, stacks, and other internal data structures.
  - **TotalAlloc**: Total number of bytes allocated for heap objects. This statistic does not decrease when objects are freed.
- **influxdb_shard**: Metrics related to InfluxDB shards.
  - **diskBytes**: Size, in bytes, of the shard, including the size of the data directory and the WAL directory.
  - **fieldsCreate**: Number of fields created.
  - **indexType**: Type of index inmem or tsi1.
  - **n_shards**: Total number of shards in the specified database.
  - **seriesCreate**: Number of series created.
  - **writeBytes**: Number of bytes written to the shard.
  - **writePointsDropped**: Number of requests to write points t dropped from a write.
  - **writePointsErr**: Number of requests to write points that failed to be written due to errors.
  - **writePointsOk**: Number of points written successfully.
  - **writeReq**: Total number of write requests.
  - **writeReqErr**: Total number of write requests that failed due to errors.
  - **writeReqOk**: Total number of successful write requests.
- **influxdb_subscriber**: InfluxDB subscription metrics.
  - **createFailures**: Number of subscriptions that failed to be created.
  - **pointsWritten**: Total number of points that were successfully written to subscribers.
  - **writeFailures**: Total number of batches that failed to be written to subscribers.
- **influxdb_tsm1_cache**: TSM cache metrics.
  - **cacheAgeMs**: Duration, in milliseconds, since the cache was last snapshotted at sample time.
  - **cachedBytes**: Total number of bytes that have been written into snapshots.
  - **diskBytes**: Size, in bytes, of on-disk snapshots.
  - **memBytes**: Size, in bytes, of in-memory cache.
  - **snapshotCount**: Current level (number) of active snapshots.
  - **WALCompactionTimeMs**: Duration, in milliseconds, that the commit lock is held while compacting snapshots.
  - **writeDropped**: Total number of writes dropped due to timeouts.
  - **writeErr**: Total number of writes that failed.
  - **writeOk**: Total number of successful writes.
- **influxdb_tsm1_engine**: TSM storage engine metrics.
  - **cacheCompactionDuration** Duration (wall time), in nanoseconds, spent in cache compactions.
  - **cacheCompactionErr** Number of cache compactions that have failed due to errors.
  - **cacheCompactions** Total number of cache compactions that have ever run.
  - **cacheCompactionsActive** Number of cache compactions that are currently running.
  - **tsmFullCompactionDuration** Duration (wall time), in nanoseconds, spent in full compactions.
  - **tsmFullCompactionErr** Total number of TSM full compactions that have failed due to errors.
  - **tsmFullCompactionQueue** Current number of pending TMS Full compactions.
  - **tsmFullCompactions** Total number of TSM full compactions that have ever run.
  - **tsmFullCompactionsActive** Number of TSM full compactions currently running.
  - **tsmLevel1CompactionDuration** Duration (wall time), in nanoseconds, spent in TSM level 1 compactions.
  - **tsmLevel1CompactionErr** Total number of TSM level 1 compactions that have failed due to errors.
  - **tsmLevel1CompactionQueue** Current number of pending TSM level 1 compactions.
  - **tsmLevel1Compactions** Total number of TSM level 1 compactions that have ever run.
  - **tsmLevel1CompactionsActive** Number of TSM level 1 compactions that are currently running.
  - **tsmLevel2CompactionDuration** Duration (wall time), in nanoseconds, spent in TSM level 2 compactions.
  - **tsmLevel2CompactionErr** Number of TSM level 2 compactions that have failed due to errors.
  - **tsmLevel2CompactionQueue** Current number of pending TSM level 2 compactions.
  - **tsmLevel2Compactions** Total number of TSM level 2 compactions that have ever run.
  - **tsmLevel2CompactionsActive** Number of TSM level 2 compactions that are currently running.
  - **tsmLevel3CompactionDuration** Duration (wall time), in nanoseconds, spent in TSM level 3 compactions.
  - **tsmLevel3CompactionErr** Number of TSM level 3 compactions that have failed due to errors.
  - **tsmLevel3CompactionQueue** Current number of pending TSM level 3 compactions.
  - **tsmLevel3Compactions** Total number of TSM level 3 compactions that have ever run.
  - **tsmLevel3CompactionsActive** Number of TSM level 3 compactions that are currently running.
  - **tsmOptimizeCompactionDuration** Duration (wall time), in nanoseconds, spent during TSM optimize compactions.
  - **tsmOptimizeCompactionErr** Total number of TSM optimize compactions that have failed due to errors.
  - **tsmOptimizeCompactionQueue** Current number of pending TSM optimize compactions.
  - **tsmOptimizeCompactions** Total number of TSM optimize compactions that have ever run.
  - **tsmOptimizeCompactionsActive** Number of TSM optimize compactions that are currently running.
- **influxdb_tsm1_filestore**: The TSM file store metrics.
  - **diskBytes**: Size, in bytes, of disk usage by the TSM file store.
  - **numFiles**: Total number of files in the TSM file store.
- **influxdb_tsm1_wal**: The TSM Write Ahead Log (WAL) metrics.
  - **currentSegmentDiskBytes**: Current size, in bytes, of the segment disk.
  - **oldSegmentDiskBytes**: Size, in bytes, of the segment disk.
  - **writeErr**: Number of writes that failed due to errors.
  - **writeOK**: Number of writes that succeeded.
- **influxdb_write**: Metrics related to InfluxDB writes.
  - **pointReq**: Total number of points requested to be written.
  - **pointReqHH** _(Enterprise only)_: Total number of points received for write by this node and then enqueued into hinted handoff for the destination node.
  - **pointReqLocal** _(Enterprise only)_: Total number of point requests that have been attempted to be written into a shard on the same (local) node.
  - **pointReqRemote** _(Enterprise only)_: Total number of points received for write by this node but needed to be forwarded into a shard on a remote node.
  - **pointsWrittenOK**: Number of points written to the HTTP /write endpoint and persisted successfully.
  - **req**: Total number of batches requested to be written.
  - **subWriteDrop**: Total number of batches that failed to be sent to the subscription dispatcher.
  - **subWriteOk**: Total number of batches successfully sent to the subscription dispatcher.
  - **valuesWrittenOK**: Number of values (fields) written to the HTTP /write endpoint and persisted successfully.
  - **writeDrop**: Total number of write requests for points that have been dropped due to timestamps not matching any existing retention policies.
  - **writeError**: Total number of batches of points that were not successfully written, due to a failure to write to a local or remote shard.
  - **writeOk**: Total number of batches of points written at the requested consistency level.
  - **writePartial** _(Enterprise only)_: Total number of batches written to at least one node, but did not meet the requested consistency level.
  - **writeTimeout**: Total number of write requests that failed to complete within the default write timeout duration.

## Example Output

```sh
telegraf --config ~/ws/telegraf.conf --input-filter influxdb --test
* Plugin: influxdb, Collection 1
> influxdb_database,database=_internal,host=tyrion,url=http://localhost:8086/debug/vars numMeasurements=10,numSeries=29 1463590500247354636
> influxdb_httpd,bind=:8086,host=tyrion,url=http://localhost:8086/debug/vars req=7,reqActive=1,reqDurationNs=14227734 1463590500247354636
> influxdb_measurement,database=_internal,host=tyrion,measurement=database,url=http://localhost:8086/debug/vars numSeries=1 1463590500247354636
> influxdb_measurement,database=_internal,host=tyrion,measurement=httpd,url=http://localhost:8086/debug/vars numSeries=1 1463590500247354636
> influxdb_measurement,database=_internal,host=tyrion,measurement=measurement,url=http://localhost:8086/debug/vars numSeries=10 1463590500247354636
> influxdb_measurement,database=_internal,host=tyrion,measurement=runtime,url=http://localhost:8086/debug/vars numSeries=1 1463590500247354636
> influxdb_measurement,database=_internal,host=tyrion,measurement=shard,url=http://localhost:8086/debug/vars numSeries=4 1463590500247354636
> influxdb_measurement,database=_internal,host=tyrion,measurement=subscriber,url=http://localhost:8086/debug/vars numSeries=1 1463590500247354636
> influxdb_measurement,database=_internal,host=tyrion,measurement=tsm1_cache,url=http://localhost:8086/debug/vars numSeries=4 1463590500247354636
> influxdb_measurement,database=_internal,host=tyrion,measurement=tsm1_filestore,url=http://localhost:8086/debug/vars numSeries=2 1463590500247354636
> influxdb_measurement,database=_internal,host=tyrion,measurement=tsm1_wal,url=http://localhost:8086/debug/vars numSeries=4 1463590500247354636
> influxdb_measurement,database=_internal,host=tyrion,measurement=write,url=http://localhost:8086/debug/vars numSeries=1 1463590500247354636
> influxdb_memstats,host=tyrion,url=http://localhost:8086/debug/vars alloc=7642384i,buck_hash_sys=1463471i,frees=1169558i,gc_sys=653312i,gc_cpu_fraction=0.00003825652361068311,heap_alloc=7642384i,heap_idle=9912320i,heap_inuse=9125888i,heap_objects=48276i,heap_released=0i,heap_sys=19038208i,last_gc=1463590480877651621i,lookups=90i,mallocs=1217834i,mcache_inuse=4800i,mcache_sys=16384i,mspan_inuse=70920i,mspan_sys=81920i,next_gc=11679787i,num_gc=141i,other_sys=1244233i,pause_total_ns=24034027i,stack_inuse=884736i,stack_sys=884736i,sys=23382264i,total_alloc=679012200i 1463590500277918755
> influxdb_shard,database=_internal,engine=tsm1,host=tyrion,id=4,path=/Users/sparrc/.influxdb/data/_internal/monitor/4,retentionPolicy=monitor,url=http://localhost:8086/debug/vars fieldsCreate=65,seriesCreate=26,writePointsOk=7274,writeReq=280 1463590500247354636
> influxdb_subscriber,host=tyrion,url=http://localhost:8086/debug/vars pointsWritten=7274 1463590500247354636
> influxdb_tsm1_cache,database=_internal,host=tyrion,path=/Users/sparrc/.influxdb/data/_internal/monitor/1,retentionPolicy=monitor,url=http://localhost:8086/debug/vars WALCompactionTimeMs=0,cacheAgeMs=2809192,cachedBytes=0,diskBytes=0,memBytes=0,snapshotCount=0 1463590500247354636
> influxdb_tsm1_cache,database=_internal,host=tyrion,path=/Users/sparrc/.influxdb/data/_internal/monitor/2,retentionPolicy=monitor,url=http://localhost:8086/debug/vars WALCompactionTimeMs=0,cacheAgeMs=2809184,cachedBytes=0,diskBytes=0,memBytes=0,snapshotCount=0 1463590500247354636
> influxdb_tsm1_cache,database=_internal,host=tyrion,path=/Users/sparrc/.influxdb/data/_internal/monitor/3,retentionPolicy=monitor,url=http://localhost:8086/debug/vars WALCompactionTimeMs=0,cacheAgeMs=2809180,cachedBytes=0,diskBytes=0,memBytes=42368,snapshotCount=0 1463590500247354636
> influxdb_tsm1_cache,database=_internal,host=tyrion,path=/Users/sparrc/.influxdb/data/_internal/monitor/4,retentionPolicy=monitor,url=http://localhost:8086/debug/vars WALCompactionTimeMs=0,cacheAgeMs=2799155,cachedBytes=0,diskBytes=0,memBytes=331216,snapshotCount=0 1463590500247354636
> influxdb_tsm1_filestore,database=_internal,host=tyrion,path=/Users/sparrc/.influxdb/data/_internal/monitor/1,retentionPolicy=monitor,url=http://localhost:8086/debug/vars diskBytes=37892 1463590500247354636
> influxdb_tsm1_filestore,database=_internal,host=tyrion,path=/Users/sparrc/.influxdb/data/_internal/monitor/2,retentionPolicy=monitor,url=http://localhost:8086/debug/vars diskBytes=52907 1463590500247354636
> influxdb_tsm1_wal,database=_internal,host=tyrion,path=/Users/sparrc/.influxdb/wal/_internal/monitor/1,retentionPolicy=monitor,url=http://localhost:8086/debug/vars currentSegmentDiskBytes=0,oldSegmentsDiskBytes=0 1463590500247354636
> influxdb_tsm1_wal,database=_internal,host=tyrion,path=/Users/sparrc/.influxdb/wal/_internal/monitor/2,retentionPolicy=monitor,url=http://localhost:8086/debug/vars currentSegmentDiskBytes=0,oldSegmentsDiskBytes=0 1463590500247354636
> influxdb_tsm1_wal,database=_internal,host=tyrion,path=/Users/sparrc/.influxdb/wal/_internal/monitor/3,retentionPolicy=monitor,url=http://localhost:8086/debug/vars currentSegmentDiskBytes=0,oldSegmentsDiskBytes=65651 1463590500247354636
> influxdb_tsm1_wal,database=_internal,host=tyrion,path=/Users/sparrc/.influxdb/wal/_internal/monitor/4,retentionPolicy=monitor,url=http://localhost:8086/debug/vars currentSegmentDiskBytes=495687,oldSegmentsDiskBytes=0 1463590500247354636
> influxdb_write,host=tyrion,url=http://localhost:8086/debug/vars pointReq=7274,pointReqLocal=7274,req=280,subWriteOk=280,writeOk=280 1463590500247354636
> influxdb_shard,host=tyrion n_shards=4i 1463590500247354636
```

## InfluxDB-formatted endpoints

The influxdb plugin can collect InfluxDB-formatted data from JSON endpoints.
Whether associated with an Influx database or not.

With a configuration of:

```toml
[[inputs.influxdb]]
  urls = [
    "http://127.0.0.1:8086/debug/vars",
    "http://192.168.2.1:8086/debug/vars"
  ]
```
