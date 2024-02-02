package solr

import "encoding/json"

// AdminCoresStatus is an exported type that
// contains a response with information about Solr cores.
type AdminCoresStatus struct {
	Status map[string]struct {
		Index struct {
			SizeInBytes int64 `json:"sizeInBytes"`
			NumDocs     int64 `json:"numDocs"`
			MaxDoc      int64 `json:"maxDoc"`
			DeletedDocs int64 `json:"deletedDocs"`
		} `json:"index"`
	} `json:"status"`
}

// MBeansData is an exported type that
// contains a response from Solr with metrics
type MBeansData struct {
	Headers    ResponseHeader    `json:"responseHeader"`
	SolrMbeans []json.RawMessage `json:"solr-mbeans"`
}

// ResponseHeader is an exported type that
// contains a response metrics: QTime and Status
type ResponseHeader struct {
	QTime  int64 `json:"QTime"`
	Status int64 `json:"status"`
}

// Core is an exported type that
// contains Core metrics
type Core struct {
	Stats struct {
		DeletedDocs int64 `json:"deletedDocs"`
		MaxDoc      int64 `json:"maxDoc"`
		NumDocs     int64 `json:"numDocs"`
	} `json:"stats"`
}

// QueryHandler is an exported type that
// contains query handler metrics
type QueryHandler struct {
	Stats interface{} `json:"stats"`
}

// UpdateHandler is an exported type that
// contains update handler metrics
type UpdateHandler struct {
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

// Cache is an exported type that
// contains cache metrics
type Cache struct {
	Stats map[string]interface{} `json:"stats"`
}
