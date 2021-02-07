package elasticsearch_query

import elastic "gopkg.in/olivere/elastic.v5"

func (e *ElasticsearchQuery) parseAggregationResult(aggregationQueryList *[]aggregationQueryData, searchResult *elastic.SearchResult) error {

	var measurements = make(map[string][]*aggKey)

	for _, aggregationQuery := range *aggregationQueryList {
		a := aggKey{function: aggregationQuery.function, field: aggregationQuery.field, name: aggregationQuery.name}

		if m := measurements[aggregationQuery.measurement]; m != nil {
			measurements[aggregationQuery.measurement] = append(m, &a)
		} else {
			measurements[aggregationQuery.measurement] = []*aggKey{&a}
		}
	}

	// recurse over aggregation results
	for measurementName, aggKey := range measurements {
		fields := make(map[string]interface{})
		tags := make(map[string]string)

		err := e.recurseResponse(aggKey, searchResult.Aggregations, measurementName, fields, tags)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *ElasticsearchQuery) parseSimpleResult(measurement string, searchResult *elastic.SearchResult) error {

	fields := make(map[string]interface{})
	tags := make(map[string]string)

	fields["doc_count"] = searchResult.Hits.TotalHits

	e.acc.AddFields(measurement, fields, tags)
	return nil
}

func (e *ElasticsearchQuery) recurseResponse(aggKeys []*aggKey, aggs elastic.Aggregations, measurementName string, fields map[string]interface{}, tags map[string]string) error {

	for _, aggKey := range aggKeys {
		resp := e.getResponseAggregation(aggKey.function, aggKey.name, aggs)

		if resp != nil {

			switch resp := resp.(type) {
			case *elastic.AggregationValueMetric:

				// we've found a metric aggregation, add to field map
				fields[aggKey.name] = float64(0)

				if resp.Value != nil {
					fields[aggKey.name] = *resp.Value
				}

			case *elastic.AggregationBucketKeyItems:
				// we've found a terms aggregation, iterate over the buckets and try to retrieve the inner aggregation values
				for item := range resp.Buckets {
					bucket := resp.Buckets[item]

					fields["doc_count"] = bucket.DocCount
					tags[aggKey.name] = bucket.Key.(string)

					err := e.recurseResponse(aggKeys, bucket.Aggregations, measurementName, fields, tags)
					if err != nil {
						return err
					}
				}
				return nil

			case *elastic.AggregationPercentilesMetric:
				// TODO
				return nil
			}
		}
	}

	if len(fields) > 0 {
		e.acc.AddFields(measurementName, fields, tags)
	}

	return nil
}

func (e *ElasticsearchQuery) getResponseAggregation(function string, aggName string, aggs elastic.Aggregations) interface{} {
	var agg interface{}
	var found bool

	switch function {
	case "avg":
		agg, found = aggs.Avg(aggName)
	case "sum":
		agg, found = aggs.Sum(aggName)
	case "min":
		agg, found = aggs.Min(aggName)
	case "max":
		agg, found = aggs.Max(aggName)
	case "terms":
		agg, found = aggs.Terms(aggName)
		// TODO:
		// case "percentile":
		// 	agg, found = aggs.Percentiles(aggName)
	}

	if found {
		return agg
	}

	return nil
}
