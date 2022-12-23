package opensearch_query

import "fmt"

type BucketAggregation map[string]*aggregationFunction

func (b BucketAggregation) AddAggregation(name, aggType, field string) error {
	switch aggType { // TODO: Use categorization function
	case "terms":
	default:
		return fmt.Errorf("aggregation function '%s' not supported", aggType)
	}

	b[name] = &aggregationFunction{
		aggType: aggType,
		field:   field,
	}

	return nil
}

func (b BucketAggregation) AddNestedAggregation(name string, a Aggregation) {
	b[name].nested = a
}

func (b BucketAggregation) BucketSize(name string, size int) error {
	if size <= 0 {
		return fmt.Errorf("invalid size; must be integer value > 0")
	}

	if _, ok := b[name]; !ok {
		return fmt.Errorf("aggregation %s not found", name)
	}

	b[name].Size(size)

	return nil
}
