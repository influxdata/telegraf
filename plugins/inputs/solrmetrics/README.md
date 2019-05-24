# Solr input plugin

The [solr](http://lucene.apache.org/solr/) plugin collects stats via the
[Metrics API](https://lucene.apache.org/solr/guide/metrics-reporting.html#metrics-api)

Solr Metrics API was introduced with Solr of version 6.4. The plugin supports ONLY usage of `key` and `prefix` parameters for `/solr/admin/metrics` endpoint.
Since, if any parameter isn't set for `/solr/admin/metrics` the endpoint returns an enourmous amount of metrics, the plugin requires explicit setting metrics that shall be returned, by using `keys` and `prefixes` plugin parameters.

## Configuration

```toml
[[inputs.solrmetrics]]
  ## specify a list of one or more Solr servers
  servers = ["http://localhost:8983"]
  
  ## Optional HTTP Basic Auth Credentials
  # username = "username"
  # password = "pa$$word"
  #
  ## Optional HTTP timeout, sec
  # httptimeout = 5
  #
  # https://lucene.apache.org/solr/guide/metrics-reporting.html#metrics-api
  ## Prefixes 
  prefixes = [
	"REPLICATION./replication.replicationEnabled",
	"REPLICATION./replication.isSlave",
	"REPLICATION./replication.isMaster",
	"CACHE.searcher.queryResultCache",
	"INDEX.sizeInBytes",
	"SEARCHER.searcher.numDocs",
	"SEARCHER.searcher.deletedDocs",
	"SEARCHER.searcher.maxDoc"
	]
  #
  ## Keys
  keys = [
	"solr.jetty:org.eclipse.jetty.server.handler.DefaultHandler.1xx-responses:count",
	"solr.jetty:org.eclipse.jetty.server.handler.DefaultHandler.2xx-responses:count",
	"solr.jetty:org.eclipse.jetty.server.handler.DefaultHandler.3xx-responses:count",
	"solr.jetty:org.eclipse.jetty.server.handler.DefaultHandler.4xx-responses:count",
	"solr.jetty:org.eclipse.jetty.server.handler.DefaultHandler.5xx-responses:count",
	"solr.jetty:org.eclipse.jetty.server.handler.DefaultHandler.connect-requests:count",
	"solr.jetty:org.eclipse.jetty.server.handler.DefaultHandler.options-requests:count",
	"solr.jetty:org.eclipse.jetty.server.handler.DefaultHandler.head-requests:count",
	"solr.jetty:org.eclipse.jetty.server.handler.DefaultHandler.move-requests:count",
	"solr.jetty:org.eclipse.jetty.server.handler.DefaultHandler.delete-requests:count",
	"solr.jetty:org.eclipse.jetty.server.handler.DefaultHandler.get-requests:count",
	"solr.jetty:org.eclipse.jetty.server.handler.DefaultHandler.post-requests:count",
	"solr.jetty:org.eclipse.jetty.server.handler.DefaultHandler.put-requests:count",
	"solr.jetty:org.eclipse.jetty.server.handler.DefaultHandler.other-requests:count",
	"solr.jvm:buffers.direct.Count",
	"solr.jvm:buffers.direct.MemoryUsed",
	"solr.jvm:buffers.direct.TotalCapacity",
	"solr.jvm:buffers.mapped.Count",
	"solr.jvm:buffers.mapped.MemoryUsed",
	"solr.jvm:buffers.mapped.TotalCapacity",
	"solr.jvm:threads.blocked.count",
	"solr.jvm:threads.count",
	"solr.jvm:threads.daemon.count",
	"solr.jvm:threads.deadlock.count",
	"solr.jvm:threads.new.count",
	"solr.jvm:threads.runnable.count",
	"solr.jvm:threads.terminated.count",
	"solr.jvm:threads.timed_waiting.count",
	"solr.jvm:threads.waiting.count",
	"solr.jvm:os.maxFileDescriptorCount",
	"solr.jvm:os.openFileDescriptorCount",
	"solr.jvm:memory.total.init",
	"solr.jvm:memory.total.max",
	"solr.jvm:memory.total.used",
	"solr.jvm:memory.heap.init",
	"solr.jvm:memory.heap.max",
	"solr.jvm:memory.heap.used",
	"solr.jvm:gc.ConcurrentMarkSweep.count",
	"solr.jvm:gc.ConcurrentMarkSweep.time",
	"solr.jvm:gc.ParNew.count",
	"solr.jvm:gc.ParNew.time",
	"solr.node:CONTAINER.fs.totalSpace",
	"solr.node:CONTAINER.fs.usableSpace",
	"solr.node:ADMIN./admin/zookeeper.errors:count",
	"solr.node:ADMIN./admin/zookeeper.timeouts:count",
	"solr.node:CONTAINER.cores.lazy",
	"solr.node:CONTAINER.cores.loaded",
	"solr.node:CONTAINER.cores.unloaded"
	]
```

## Example output of gathered metrics

```toml
2019-05-24T14:14:27Z I! Starting Telegraf 
> jvm,host=solrhost,port=9999 memory.heap.used=826104064,memory.total.used=947384648,os.maxFileDescriptorCount=8192,os.openFileDescriptorCount=291 1558707268000000000
> jetty,host=solrhost,port=9999 org.eclipse.jetty.server.handler.DefaultHandler.2xx-responses:count=1294225 1558707268000000000
> jvm,host=solrhost,port=7777 memory.heap.used=516535240,memory.total.used=623985304,os.maxFileDescriptorCount=8192,os.openFileDescriptorCount=219 1558707268000000000
> node,host=solrhost,port=9999 ADMIN./admin/zookeeper.errors:count=0,ADMIN./admin/zookeeper.timeouts:count=0,CONTAINER.cores.lazy=0,CONTAINER.cores.loaded=12,CONTAINER.cores.unloaded=0,CONTAINER.fs.usableSpace=3097767170048 1558707268000000000
> jetty,host=solrhost,port=7777 org.eclipse.jetty.server.handler.DefaultHandler.2xx-responses:count=1053396 1558707268000000000
> node,host=solrhost,port=7777 ADMIN./admin/zookeeper.errors:count=0,ADMIN./admin/zookeeper.timeouts:count=0,CONTAINER.cores.lazy=0,CONTAINER.cores.loaded=8,CONTAINER.cores.unloaded=0,CONTAINER.fs.usableSpace=3097767170048 1558707268000000000
> core,collection=collection01,host=solrhost,port=7777,replica=replica_n1,shard=shard1 CACHE.searcher.queryResultCache.cumulative_evictions=0,CACHE.searcher.queryResultCache.cumulative_hitratio=0.8,CACHE.searcher.queryResultCache.cumulative_hits=4,CACHE.searcher.queryResultCache.cumulative_inserts=1,CACHE.searcher.queryResultCache.cumulative_lookups=5,CACHE.searcher.queryResultCache.evictions=0,CACHE.searcher.queryResultCache.hitratio=0.8,CACHE.searcher.queryResultCache.hits=4,CACHE.searcher.queryResultCache.inserts=1,CACHE.searcher.queryResultCache.lookups=5,CACHE.searcher.queryResultCache.size=1,CACHE.searcher.queryResultCache.warmupTime=0,INDEX.sizeInBytes=151848557,REPLICATION./replication.isMaster=true,REPLICATION./replication.isSlave=false,REPLICATION./replication.replicationEnabled=true,SEARCHER.searcher.deletedDocs=68673,SEARCHER.searcher.maxDoc=894329,SEARCHER.searcher.numDocs=825656 1558707268000000000
> core,collection=collection02,host=solrhost,port=7777,replica=replica_n1,shard=shard1 CACHE.searcher.queryResultCache.cumulative_evictions=0,CACHE.searcher.queryResultCache.cumulative_hitratio=0.66,CACHE.searcher.queryResultCache.cumulative_hits=29,CACHE.searcher.queryResultCache.cumulative_inserts=15,CACHE.searcher.queryResultCache.cumulative_lookups=44,CACHE.searcher.queryResultCache.evictions=0,CACHE.searcher.queryResultCache.hitratio=0,CACHE.searcher.queryResultCache.hits=0,CACHE.searcher.queryResultCache.inserts=1,CACHE.searcher.queryResultCache.lookups=1,CACHE.searcher.queryResultCache.size=1,CACHE.searcher.queryResultCache.warmupTime=0,INDEX.sizeInBytes=228090478,REPLICATION./replication.isMaster=true,REPLICATION./replication.isSlave=false,REPLICATION./replication.replicationEnabled=true,SEARCHER.searcher.deletedDocs=217445,SEARCHER.searcher.maxDoc=1093614,SEARCHER.searcher.numDocs=876169 1558707268000000000
> core,collection=collection03,host=solrhost,port=9999,replica=replica_n1,shard=shard1 CACHE.searcher.queryResultCache.cumulative_evictions=0,CACHE.searcher.queryResultCache.cumulative_hitratio=0.67,CACHE.searcher.queryResultCache.cumulative_hits=4,CACHE.searcher.queryResultCache.cumulative_inserts=2,CACHE.searcher.queryResultCache.cumulative_lookups=6,CACHE.searcher.queryResultCache.evictions=0,CACHE.searcher.queryResultCache.hitratio=0.8,CACHE.searcher.queryResultCache.hits=4,CACHE.searcher.queryResultCache.inserts=1,CACHE.searcher.queryResultCache.lookups=5,CACHE.searcher.queryResultCache.size=1,CACHE.searcher.queryResultCache.warmupTime=0,INDEX.sizeInBytes=1917643,REPLICATION./replication.isMaster=true,REPLICATION./replication.isSlave=false,REPLICATION./replication.replicationEnabled=true,SEARCHER.searcher.deletedDocs=0,SEARCHER.searcher.maxDoc=5285,SEARCHER.searcher.numDocs=5285 1558707268000000000
> core,collection=collection04,host=solrhost,port=9999,replica=replica1,shard=shard1 CACHE.searcher.queryResultCache.cumulative_evictions=0,CACHE.searcher.queryResultCache.cumulative_hitratio=0.95,CACHE.searcher.queryResultCache.cumulative_hits=1392,CACHE.searcher.queryResultCache.cumulative_inserts=125,CACHE.searcher.queryResultCache.cumulative_lookups=1464,CACHE.searcher.queryResultCache.evictions=0,CACHE.searcher.queryResultCache.hitratio=0.96,CACHE.searcher.queryResultCache.hits=1030,CACHE.searcher.queryResultCache.inserts=80,CACHE.searcher.queryResultCache.lookups=1069,CACHE.searcher.queryResultCache.size=38,CACHE.searcher.queryResultCache.warmupTime=0,INDEX.sizeInBytes=231513419,REPLICATION./replication.isMaster=true,REPLICATION./replication.isSlave=false,REPLICATION./replication.replicationEnabled=true,SEARCHER.searcher.deletedDocs=0,SEARCHER.searcher.maxDoc=658660,SEARCHER.searcher.numDocs=658660 1558707268000000000
```
