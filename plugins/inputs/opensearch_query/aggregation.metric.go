package opensearch_query

import "fmt"

type MetricAggregationRequest map[string]*aggregationFunction

func (m MetricAggregationRequest) AddAggregation(name, aggType, field string) error {
	if t, _ := getAggregationFunctionType(aggType); t != "metric" {
		return fmt.Errorf("aggregation function %q not supported", aggType)
	}

	m[name] = &aggregationFunction{
		aggType: aggType,
		field:   field,
	}

	return nil
}
