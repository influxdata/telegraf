package opensearch_query

import (
	"encoding/json"
	"fmt"
)

type Aggregation interface {
	AddAggregation(string, string, string) error
}

type NestedAggregation interface {
	Nested(string, Aggregation)
	Missing(string)
	Size(int)
}

type aggregationFunction struct {
	aggType string
	field   string
	size    int

	nested Aggregation
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
	agg[a.aggType] = field

	if a.nested != nil {
		agg["aggregations"] = a.nested
	}
	return json.Marshal(agg)
}

func (a *aggregationFunction) Size(size int) {
	a.size = size
}

func getAggregationFunctionType(field string) (string, error) {
	switch field {
	case "avg", "sum", "min", "max", "cardinality", "value_count":
		return "metric", nil
	case "terms":
		return "bucket", nil
	default:
		return "", fmt.Errorf("invalid aggregation function %s", field)
	}
}
