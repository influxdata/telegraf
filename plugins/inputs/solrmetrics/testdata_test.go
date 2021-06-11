package solrmetrics

const (
	metricsPrefixesResponse = `
	{
		"responseHeader":{
		  "status":0,
		  "QTime":10},
		"metrics":{
		  "solr.core.collection01.shard1.replica1":{
			"CACHE.searcher.queryResultCache":{
			  "lookups":3,
			  "hits":2,
			  "cumulative_evictions":0,
			  "size":1,
			  "hitratio":0.67,
			  "evictions":0,
			  "cumulative_lookups":3,
			  "cumulative_hitratio":0.67,
			  "warmupTime":0,
			  "inserts":1,
			  "cumulative_inserts":1,
			  "cumulative_hits":2},
			"INDEX.sizeInBytes":18622286,
			"REPLICATION./replication.replicationEnabled":true},
		  "solr.core.collection02.shard1.replica1":{
			"CACHE.searcher.queryResultCache":{
			  "lookups":3,
			  "hits":2,
			  "cumulative_evictions":0,
			  "size":1,
			  "hitratio":0.67,
			  "evictions":0,
			  "cumulative_lookups":3,
			  "cumulative_hitratio":0.67,
			  "warmupTime":0,
			  "inserts":1,
			  "cumulative_inserts":1,
			  "cumulative_hits":2},
			"INDEX.sizeInBytes":169412206,
			"REPLICATION./replication.replicationEnabled":true},
		  "solr.core.collection03.shard1.replica1":{
			"CACHE.searcher.queryResultCache":{
			  "lookups":0,
			  "hits":0,
			  "cumulative_evictions":0,
			  "size":0,
			  "hitratio":0.0,
			  "evictions":0,
			  "cumulative_lookups":0,
			  "cumulative_hitratio":0.0,
			  "warmupTime":0,
			  "inserts":0,
			  "cumulative_inserts":0,
			  "cumulative_hits":0},
			"INDEX.sizeInBytes":259680378,
			"REPLICATION./replication.replicationEnabled":true},
		  "solr.core.collection04.shard1.replica1":{
			"CACHE.searcher.queryResultCache":{
			  "lookups":0,
			  "hits":0,
			  "cumulative_evictions":0,
			  "size":7,
			  "hitratio":0.0,
			  "evictions":0,
			  "cumulative_lookups":0,
			  "cumulative_hitratio":0.0,
			  "warmupTime":0,
			  "inserts":7,
			  "cumulative_inserts":0,
			  "cumulative_hits":0},
			"INDEX.sizeInBytes":8189000,
			"REPLICATION./replication.replicationEnabled":true}}}
	`
	metricsKeysResponse = `
	{
		"responseHeader":{
		  "status":0,
		  "QTime":3},
		"metrics":{
		  "solr.jetty:org.eclipse.jetty.server.handler.DefaultHandler.2xx-responses:count":393179,
		  "solr.jetty:org.eclipse.jetty.server.handler.DefaultHandler.get-requests:count":399787,
		  "solr.jvm:buffers.direct.MemoryUsed":1034909,
		  "solr.jvm:memory.heap.init":64424509440,
		  "solr.node:CONTAINER.fs.totalSpace":3348006576128,
		  "solr.node:CONTAINER.fs.usableSpace":3097767038976}}
	`
)

var (
	testPrefixes = []string{
		"REPLICATION./replication.replicationEnabled",
		"REPLICATION./replication.isSlave",
		"REPLICATION./replication.isMaster",
		"CACHE.searcher.queryResultCache",
		"INDEX.sizeInBytes",
		"SEARCHER.searcher.numDocs",
		"SEARCHER.searcher.deletedDocs",
		"SEARCHER.searcher.maxDoc",
	}

	testKeys = []string{
		"solr.jetty:org.eclipse.jetty.server.handler.DefaultHandler.2xx-responses:count",
		"solr.jetty:org.eclipse.jetty.server.handler.DefaultHandler.get-requests:count",
		"solr.jvm:buffers.direct.MemoryUsed",
		"solr.jvm:memory.heap.init",
		"solr.node:CONTAINER.fs.totalSpace",
		"solr.node:CONTAINER.fs.usableSpace",
	}
	solrCol01MetricsExpected = Metrics{
		"CACHE.searcher.queryResultCache.lookups":              float64(3),
		"CACHE.searcher.queryResultCache.hits":                 float64(2),
		"CACHE.searcher.queryResultCache.cumulative_evictions": float64(0),
		"CACHE.searcher.queryResultCache.size":                 float64(1),
		"CACHE.searcher.queryResultCache.hitratio":             float64(0.67),
		"CACHE.searcher.queryResultCache.evictions":            float64(0),
		"CACHE.searcher.queryResultCache.cumulative_lookups":   float64(3),
		"CACHE.searcher.queryResultCache.cumulative_hitratio":  float64(0.67),
		"CACHE.searcher.queryResultCache.warmupTime":           float64(0),
		"CACHE.searcher.queryResultCache.inserts":              float64(1),
		"CACHE.searcher.queryResultCache.cumulative_inserts":   float64(1),
		"CACHE.searcher.queryResultCache.cumulative_hits":      float64(2),
		"INDEX.sizeInBytes":                                    float64(18622286),
		"REPLICATION./replication.replicationEnabled":          bool(true),
	}

	solrCol02MetricsExpected = Metrics{
		"CACHE.searcher.queryResultCache.lookups":              float64(3),
		"CACHE.searcher.queryResultCache.hits":                 float64(2),
		"CACHE.searcher.queryResultCache.cumulative_evictions": float64(0),
		"CACHE.searcher.queryResultCache.size":                 float64(1),
		"CACHE.searcher.queryResultCache.hitratio":             float64(0.67),
		"CACHE.searcher.queryResultCache.evictions":            float64(0),
		"CACHE.searcher.queryResultCache.cumulative_lookups":   float64(3),
		"CACHE.searcher.queryResultCache.cumulative_hitratio":  float64(0.67),
		"CACHE.searcher.queryResultCache.warmupTime":           float64(0),
		"CACHE.searcher.queryResultCache.inserts":              float64(1),
		"CACHE.searcher.queryResultCache.cumulative_inserts":   float64(1),
		"CACHE.searcher.queryResultCache.cumulative_hits":      float64(2),
		"INDEX.sizeInBytes":                                    float64(169412206),
		"REPLICATION./replication.replicationEnabled":          bool(true),
	}

	solrCol03MetricsExpected = Metrics{
		"CACHE.searcher.queryResultCache.lookups":              float64(0),
		"CACHE.searcher.queryResultCache.hits":                 float64(0),
		"CACHE.searcher.queryResultCache.cumulative_evictions": float64(0),
		"CACHE.searcher.queryResultCache.size":                 float64(0),
		"CACHE.searcher.queryResultCache.hitratio":             float64(0),
		"CACHE.searcher.queryResultCache.evictions":            float64(0),
		"CACHE.searcher.queryResultCache.cumulative_lookups":   float64(0),
		"CACHE.searcher.queryResultCache.cumulative_hitratio":  float64(0),
		"CACHE.searcher.queryResultCache.warmupTime":           float64(0),
		"CACHE.searcher.queryResultCache.inserts":              float64(0),
		"CACHE.searcher.queryResultCache.cumulative_inserts":   float64(0),
		"CACHE.searcher.queryResultCache.cumulative_hits":      float64(0),
		"INDEX.sizeInBytes":                                    float64(259680378),
		"REPLICATION./replication.replicationEnabled":          bool(true),
	}

	solrCol04MetricsExpected = Metrics{
		"CACHE.searcher.queryResultCache.lookups":              float64(0),
		"CACHE.searcher.queryResultCache.hits":                 float64(0),
		"CACHE.searcher.queryResultCache.cumulative_evictions": float64(0),
		"CACHE.searcher.queryResultCache.size":                 float64(7),
		"CACHE.searcher.queryResultCache.hitratio":             float64(0),
		"CACHE.searcher.queryResultCache.evictions":            float64(0),
		"CACHE.searcher.queryResultCache.cumulative_lookups":   float64(0),
		"CACHE.searcher.queryResultCache.cumulative_hitratio":  float64(0),
		"CACHE.searcher.queryResultCache.warmupTime":           float64(0),
		"CACHE.searcher.queryResultCache.inserts":              float64(7),
		"CACHE.searcher.queryResultCache.cumulative_inserts":   float64(0),
		"CACHE.searcher.queryResultCache.cumulative_hits":      float64(0),
		"INDEX.sizeInBytes":                                    float64(8189000),
		"REPLICATION./replication.replicationEnabled":          bool(true),
	}

	solrJettyMetricsExpected = Metrics{
		"org.eclipse.jetty.server.handler.DefaultHandler.2xx-responses:count": float64(393179),
		"org.eclipse.jetty.server.handler.DefaultHandler.get-requests:count":  float64(399787),
	}

	solrJVMMetricsExpected = Metrics{
		"buffers.direct.MemoryUsed": float64(1034909),
		"memory.heap.init":          float64(64424509440),
	}

	solrNodeMetricsExpected = Metrics{
		"CONTAINER.fs.totalSpace":  float64(3348006576128),
		"CONTAINER.fs.usableSpace": float64(3097767038976),
	}
)
