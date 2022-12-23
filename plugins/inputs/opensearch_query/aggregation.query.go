package opensearch_query

type AggregationQuery struct {
	Size         int         `json:"size"`
	Aggregations Aggregation `json:"aggregations"`
	Query        interface{} `json:"query,omitempty"`
}
