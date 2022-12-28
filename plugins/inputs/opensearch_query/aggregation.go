package opensearch_query

import (
	"encoding/json"
	"fmt"
)

type AggregationRequest interface {
	AddAggregation(string, string, string) error
}

type NestedAggregation interface {
	Nested(string, AggregationRequest)
	Missing(string)
	Size(int)
}

type aggregationFunction struct {
	aggType string
	field   string
	size    int
	missing string

	nested AggregationRequest
}

func (a *aggregationFunction) MarshalJSON() ([]byte, error) {
	agg := make(map[string]interface{})
	field := map[string]interface{}{"field": a.field}
	if t, _ := getAggregationFunctionType(a.aggType); t == "bucket" {
		// We'll use the default size of 10 if it hasn't been set; size == 0 is illegal in a bucket aggregation
		if a.size == 0 {
			a.size = 10
		}
		field["size"] = a.size
	}
	if a.missing != "" {
		field["missing"] = a.missing
	}

	agg[a.aggType] = field

	if a.nested != nil {
		agg["aggregations"] = a.nested
	}
	return json.Marshal(agg)
}

func (a *aggregationFunction) Size(size int) {
	a.size = size
}

func (a *aggregationFunction) Missing(missing string) {
	a.missing = missing
}

func getAggregationFunctionType(field string) (string, error) {
	switch field {
	case "avg", "sum", "min", "max", "value_count", "stats", "extended_stats", "percentiles":
		return "metric", nil
	case "terms":
		return "bucket", nil
	default:
		return "", fmt.Errorf("invalid aggregation function %s", field)
	}
}
