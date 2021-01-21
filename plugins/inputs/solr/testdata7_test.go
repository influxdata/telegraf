package solr

const mBeansSolr7Response = `
{
   "responseHeader":{
      "status":0,
      "QTime":2
   },
   "solr-mbeans":[
      "CORE",
      {

      },
      "QUERYHANDLER",
      {

      },
      "UPDATEHANDLER",
      {

      },
      "CACHE",
      {
         "documentCache":{
            "class":"org.apache.solr.search.LRUCache",
            "description":"LRU Cache(maxSize=16384, initialSize=4096)",
            "stats":{
               "CACHE.searcher.documentCache.evictions": 141485,
               "CACHE.searcher.documentCache.cumulative_lookups": 265132,
               "CACHE.searcher.documentCache.hitratio": 0.44,
               "CACHE.searcher.documentCache.size": 8192,
               "CACHE.searcher.documentCache.cumulative_hitratio": 0.42,
               "CACHE.searcher.documentCache.lookups": 1234,
               "CACHE.searcher.documentCache.warmupTime": 1,
               "CACHE.searcher.documentCache.inserts": 987,
               "CACHE.searcher.documentCache.hits": 1111,
               "CACHE.searcher.documentCache.cumulative_hits": 115364,
               "CACHE.searcher.documentCache.cumulative_inserts": 149768,
               "CACHE.searcher.documentCache.cumulative_evictions": 141486
            }
         }
      }
   ]
}
`

var solr7CacheExpected = map[string]interface{}{
	"evictions":            int64(141485),
	"cumulative_evictions": int64(141486),
	"cumulative_hitratio":  float64(0.42),
	"cumulative_hits":      int64(115364),
	"cumulative_inserts":   int64(149768),
	"cumulative_lookups":   int64(265132),
	"hitratio":             float64(0.44),
	"hits":                 int64(1111),
	"inserts":              int64(987),
	"lookups":              int64(1234),
	"size":                 int64(8192),
	"warmup_time":          int64(1),
}
