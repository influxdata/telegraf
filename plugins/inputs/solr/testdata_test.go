package solr

var solrAdminMainCoreStatusExpected = map[string]interface{}{
	"num_docs":      int64(168943425),
	"max_docs":      int64(169562700),
	"deleted_docs":  int64(619275),
	"size_in_bytes": int64(247497521642),
}

var solrAdminCore1StatusExpected = map[string]interface{}{
	"num_docs":      int64(7517488),
	"max_docs":      int64(7620303),
	"deleted_docs":  int64(102815),
	"size_in_bytes": int64(1784635686),
}

var solrCoreExpected = map[string]interface{}{
	"num_docs":     int64(168962621),
	"max_docs":     int64(169647870),
	"deleted_docs": int64(685249),
}

var solrQueryHandlerExpected = map[string]interface{}{
	"15min_rate_reqs_per_second": float64(0),
	"5min_rate_reqs_per_second":  float64(0),
	"75th_pc_request_time":       float64(0),
	"95th_pc_request_time":       float64(0),
	"999th_pc_request_time":      float64(0),
	"99th_pc_request_time":       float64(0),
	"avg_requests_per_second":    float64(0),
	"avg_time_per_request":       float64(0),
	"errors":                     int64(0),
	"handler_start":              int64(1482259270810),
	"median_request_time":        float64(0),
	"requests":                   int64(0),
	"timeouts":                   int64(0),
	"total_time":                 float64(0),
}

var solrUpdateHandlerExpected = map[string]interface{}{
	"adds":                        int64(0),
	"autocommit_max_docs":         int64(500),
	"autocommit_max_time":         int64(900),
	"autocommits":                 int64(0),
	"commits":                     int64(0),
	"cumulative_adds":             int64(0),
	"cumulative_deletes_by_id":    int64(0),
	"cumulative_deletes_by_query": int64(0),
	"cumulative_errors":           int64(0),
	"deletes_by_id":               int64(0),
	"deletes_by_query":            int64(0),
	"docs_pending":                int64(0),
	"errors":                      int64(0),
	"expunge_deletes":             int64(0),
	"optimizes":                   int64(0),
	"rollbacks":                   int64(0),
	"soft_autocommits":            int64(0),
}

var solrCacheExpected = map[string]interface{}{
	"cumulative_evictions": int64(0),
	"cumulative_hitratio":  float64(0),
	"cumulative_hits":      int64(55),
	"cumulative_inserts":   int64(14),
	"cumulative_lookups":   int64(69),
	"evictions":            int64(0),
	"hitratio":             float64(0.01),
	"hits":                 int64(0),
	"inserts":              int64(0),
	"lookups":              int64(0),
	"size":                 int64(0),
	"warmup_time":          int64(0),
}
