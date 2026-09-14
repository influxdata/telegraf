package opensearch_query

import (
	"encoding/json"
	"time"
)

type query struct {
	Size         int                `json:"size"`
	Aggregations aggregationRequest `json:"aggregations"`
	Query        any                `json:"query,omitempty"`
}

type boolQuery struct {
	FilterQueryString string
	TimestampField    string
	TimeRangeFrom     time.Time
	TimeRangeTo       time.Time
	DateFieldFormat   string
}

// MarshalJSON customizes the JSON marshaling for boolQuery.
func (b *boolQuery) MarshalJSON() ([]byte, error) {
	// Construct range
	dateTimeRange := map[string]any{
		"from":          b.TimeRangeFrom,
		"to":            b.TimeRangeTo,
		"include_lower": true,
		"include_upper": true,
	}
	if b.DateFieldFormat != "" {
		dateTimeRange["format"] = b.DateFieldFormat
	}
	rangeFilter := map[string]map[string]any{"range": {b.TimestampField: dateTimeRange}}

	// Construct Filter
	queryFilter := map[string]map[string]any{
		"query_string": {"query": b.FilterQueryString},
	}

	// Construct boolean query
	bq := map[string]map[string]any{"bool": {"filter": []any{rangeFilter, queryFilter}}}

	return json.Marshal(&bq)
}
