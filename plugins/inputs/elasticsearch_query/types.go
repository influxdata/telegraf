package elasticsearch_query

import "encoding/json"

type serverVersion struct {
	Number string `json:"number"`
}

type serverInfo struct {
	Version serverVersion `json:"version"`
}

type apiErrorDetails struct {
	Type   string `json:"type"`
	Reason string `json:"reason"`
}

type apiErrorResponse struct {
	Error apiErrorDetails `json:"error"`
}

type searchHits struct {
	Total json.RawMessage `json:"total"`
}

type searchResponse struct {
	Hits         searchHits                 `json:"hits"`
	Aggregations map[string]json.RawMessage `json:"aggregations"`
}

type totalHits struct {
	Value int64 `json:"value"`
}

type aggregationValue struct {
	Value *float64 `json:"value"`
}

type aggregationBuckets struct {
	Buckets []map[string]json.RawMessage `json:"buckets"`
}
