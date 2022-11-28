package solr

var solr3CoreExpected = map[string]interface{}{
	"num_docs":     int64(117166),
	"max_docs":     int64(117305),
	"deleted_docs": int64(0),
}

var solr3QueryHandlerExpected = map[string]interface{}{
	"15min_rate_reqs_per_second": float64(0),
	"5min_rate_reqs_per_second":  float64(0),
	"75th_pc_request_time":       float64(0),
	"95th_pc_request_time":       float64(0),
	"999th_pc_request_time":      float64(0),
	"99th_pc_request_time":       float64(0),
	"avg_requests_per_second":    float64(0),
	"avg_time_per_request":       float64(0),
	"errors":                     int64(0),
	"handler_start":              int64(1516083353156),
	"median_request_time":        float64(0),
	"requests":                   int64(0),
	"timeouts":                   int64(0),
	"total_time":                 float64(0),
}

var solr3UpdateHandlerExpected = map[string]interface{}{
	"adds":                        int64(0),
	"autocommit_max_docs":         int64(0),
	"autocommit_max_time":         int64(0),
	"autocommits":                 int64(0),
	"commits":                     int64(3220),
	"cumulative_adds":             int64(354209),
	"cumulative_deletes_by_id":    int64(0),
	"cumulative_deletes_by_query": int64(3),
	"cumulative_errors":           int64(0),
	"deletes_by_id":               int64(0),
	"deletes_by_query":            int64(0),
	"docs_pending":                int64(0),
	"errors":                      int64(0),
	"expunge_deletes":             int64(0),
	"optimizes":                   int64(3),
	"rollbacks":                   int64(0),
	"soft_autocommits":            int64(0),
}

var solr3CacheExpected = map[string]interface{}{
	"cumulative_evictions": int64(0),
	"cumulative_hitratio":  float64(1.00),
	"cumulative_hits":      int64(4041),
	"cumulative_inserts":   int64(2828),
	"cumulative_lookups":   int64(4041),
	"evictions":            int64(0),
	"hitratio":             float64(1.00),
	"hits":                 int64(2),
	"inserts":              int64(2),
	"lookups":              int64(2),
	"size":                 int64(2),
	"warmup_time":          int64(0),
}
