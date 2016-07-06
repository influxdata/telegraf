# influxdb plugin

The InfluxDB plugin will collect metrics on the given InfluxDB servers.

This plugin can also gather metrics from endpoints that expose
InfluxDB-formatted endpoints. See below for more information.

### Configuration:

```toml
# Read InfluxDB-formatted JSON metrics from one or more HTTP endpoints
[[inputs.influxdb]]
  ## Works with InfluxDB debug endpoints out of the box,
  ## but other services can use this format too.
  ## See the influxdb plugin's README for more details.

  ## Multiple URLs from which to read InfluxDB-formatted JSON
  urls = [
    "http://localhost:8086/debug/vars"
  ]
```

### Measurements & Fields

- influxdb
  - n_shards
- influxdb_database
- influxdb_httpd
- influxdb_measurement
- influxdb_memstats
  - heap_inuse
  - heap_released
  - mspan_inuse
  - total_alloc
  - sys
  - mallocs
  - frees
  - heap_idle
  - pause_total_ns
  - lookups
  - heap_sys
  - mcache_sys
  - next_gc
  - gcc_pu_fraction
  - other_sys
  - alloc
  - stack_inuse
  - stack_sys
  - buck_hash_sys
  - gc_sys
  - num_gc
  - heap_alloc
  - heap_objects
  - mspan_sys
  - mcache_inuse
  - last_gc
- influxdb_shard
- influxdb_subscriber
- influxdb_tsm1_cache
- influxdb_tsm1_wal
- influxdb_write

### Example Output:

```
telegraf -config ~/ws/telegraf.conf -input-filter influxdb -test
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
> influxdb_memstats,host=tyrion,url=http://localhost:8086/debug/vars alloc=7642384i,buck_hash_sys=1463471i,frees=1169558i,gc_sys=653312i,gcc_pu_fraction=0.00003825652361068311,heap_alloc=7642384i,heap_idle=9912320i,heap_inuse=9125888i,heap_objects=48276i,heap_released=0i,heap_sys=19038208i,last_gc=1463590480877651621i,lookups=90i,mallocs=1217834i,mcache_inuse=4800i,mcache_sys=16384i,mspan_inuse=70920i,mspan_sys=81920i,next_gc=11679787i,num_gc=141i,other_sys=1244233i,pause_total_ns=24034027i,stack_inuse=884736i,stack_sys=884736i,sys=23382264i,total_alloc=679012200i 1463590500277918755
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

### InfluxDB-formatted endpoints

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
