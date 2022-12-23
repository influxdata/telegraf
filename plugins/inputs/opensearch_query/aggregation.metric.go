package opensearch_query

import "fmt"

type MetricAggregation map[string]*aggregationFunction

func (m MetricAggregation) AddAggregation(name, aggType, field string) error {
	if t, _ := getAggregationFunctionType(aggType); t != "metric" {
		return fmt.Errorf("aggregation function '%s' not supported", aggType)
	}

	m[name] = &aggregationFunction{
		aggType: aggType,
		field:   field,
	}

	return nil
}
