package solr

import "encoding/json"

// adminCoresStatus is an exported type that contains a response with information about Solr cores.
type adminCoresStatus struct {
	Status map[string]struct {
		Index struct {
			SizeInBytes int64 `json:"sizeInBytes"`
			NumDocs     int64 `json:"numDocs"`
			MaxDoc      int64 `json:"maxDoc"`
			DeletedDocs int64 `json:"deletedDocs"`
		} `json:"index"`
	} `json:"status"`
}

// mBeansData is an exported type that contains a response from Solr with metrics
type mBeansData struct {
	Headers    responseHeader    `json:"responseHeader"`
	SolrMbeans []json.RawMessage `json:"solr-mbeans"`
}

// responseHeader is an exported type that contains a response metrics: QTime and Status
type responseHeader struct {
	QTime  int64 `json:"QTime"`
	Status int64 `json:"status"`
}

// core is an exported type that contains Core metrics
type core struct {
	Stats struct {
		DeletedDocs int64 `json:"deletedDocs"`
		MaxDoc      int64 `json:"maxDoc"`
		NumDocs     int64 `json:"numDocs"`
	} `json:"stats"`
}

// queryHandler is an exported type that contains query handler metrics
type queryHandler struct {
	Stats interface{} `json:"stats"`
}

// updateHandler is an exported type that contains update handler metrics
type updateHandler struct {
	Stats struct {
		Adds                     int64  `json:"adds"`
		AutocommitMaxDocs        int64  `json:"autocommit maxDocs"`
		AutocommitMaxTime        string `json:"autocommit maxTime"`
		Autocommits              int64  `json:"autocommits"`
		Commits                  int64  `json:"commits"`
		CumulativeAdds           int64  `json:"cumulative_adds"`
		CumulativeDeletesByID    int64  `json:"cumulative_deletesById"`
		CumulativeDeletesByQuery int64  `json:"cumulative_deletesByQuery"`
		CumulativeErrors         int64  `json:"cumulative_errors"`
		DeletesByID              int64  `json:"deletesById"`
		DeletesByQuery           int64  `json:"deletesByQuery"`
		DocsPending              int64  `json:"docsPending"`
		Errors                   int64  `json:"errors"`
		ExpungeDeletes           int64  `json:"expungeDeletes"`
		Optimizes                int64  `json:"optimizes"`
		Rollbacks                int64  `json:"rollbacks"`
		SoftAutocommits          int64  `json:"soft autocommits"`
	} `json:"stats"`
}

// cache is an exported type that contains cache metrics
type cache struct {
	Stats map[string]interface{} `json:"stats"`
}
