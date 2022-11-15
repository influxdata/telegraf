package solr

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
