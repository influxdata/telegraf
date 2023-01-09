package opensearch_query

import "fmt"

type BucketAggregationRequest map[string]*aggregationFunction

func (b BucketAggregationRequest) AddAggregation(name, aggType, field string) error {
	switch aggType {
	case "terms":
	default:
		return fmt.Errorf("aggregation function %q not supported", aggType)
	}

	b[name] = &aggregationFunction{
		aggType: aggType,
		field:   field,
	}

	return nil
}

func (b BucketAggregationRequest) AddNestedAggregation(name string, a AggregationRequest) {
	b[name].nested = a
}

func (b BucketAggregationRequest) BucketSize(name string, size int) error {
	if size <= 0 {
		return fmt.Errorf("invalid size; must be integer value > 0")
	}

	if _, ok := b[name]; !ok {
		return fmt.Errorf("aggregation %q not found", name)
	}

	b[name].Size(size)

	return nil
}

func (b BucketAggregationRequest) Missing(name, missing string) {
	b[name].Missing(missing)
}
