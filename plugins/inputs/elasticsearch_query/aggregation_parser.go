package elasticsearch_query

import (
	"fmt"

	"github.com/influxdata/telegraf"
	elastic "gopkg.in/olivere/elastic.v5"
)

type resultMetric struct {
	name   string
	fields map[string]interface{}
	tags   map[string]string
}

func (e *ElasticsearchQuery) parseSimpleResult(measurement string, searchResult *elastic.SearchResult, acc telegraf.Accumulator) error {

	fields := make(map[string]interface{})
	tags := make(map[string]string)

	fields["doc_count"] = searchResult.Hits.TotalHits

	acc.AddFields(measurement, fields, tags)
	return nil
}

func (e *ElasticsearchQuery) parseAggregationResult(aggregationQueryList []aggregationQueryData, searchResult *elastic.SearchResult, acc telegraf.Accumulator) error {
	var measurements = map[string]map[string]string{}

	// organize the aggregation query data by measurement
	for _, aggregationQuery := range aggregationQueryList {

		if measurements[aggregationQuery.measurement] == nil {
			measurements[aggregationQuery.measurement] = map[string]string{
				aggregationQuery.name: aggregationQuery.function,
			}
		} else {
			t := measurements[aggregationQuery.measurement]
			t[aggregationQuery.name] = aggregationQuery.function
			measurements[aggregationQuery.measurement] = t
		}
	}

	// recurse over aggregation results per measurement
	for measurement, aggNameFunction := range measurements {
		var m resultMetric

		m.fields = make(map[string]interface{})
		m.tags = make(map[string]string)
		m.name = measurement

		m, err := e.recurseResponse(aggNameFunction, searchResult.Aggregations, m, acc)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *ElasticsearchQuery) recurseResponse(aggKeys map[string]string, bucketResult elastic.Aggregations, m resultMetric, acc telegraf.Accumulator) (resultMetric, error) {
	var err error

	aggName, found := getAggName(bucketResult)
	if !found {
		// we've reached a single bucket without aggregation, nothing here
		return m, nil
	}

	aggFunction, found := aggKeys[aggName]
	if !found {
		return m, fmt.Errorf("child aggregation function '%s' not found %v", aggName, aggKeys)
	}

	resp := e.getResponseAggregation(aggFunction, aggName, bucketResult)
	if resp == nil {
		return m, fmt.Errorf("child aggregation '%s' not found", aggName)
	}

	switch resp := resp.(type) {
	case *elastic.AggregationBucketKeyItems:
		// we've found a terms aggregation, iterate over the buckets and try to retrieve the inner aggregation values
		for _, bucket := range resp.Buckets {
			m.fields["doc_count"] = bucket.DocCount
			if s, ok := bucket.Key.(string); ok {
				m.tags[aggName] = s
			} else {
				return m, fmt.Errorf("bucket key is not a string")
			}

			// we need to recurse down through the buckets, as it may contain another terms aggregation
			m, err = e.recurseResponse(aggKeys, bucket.Aggregations, m, acc)
			if err != nil {
				return m, err
			}

			// if there are fields present after finishing the bucket, it is a complete metric
			// store it and clean the fields to start a new metric
			if len(m.fields) > 0 {
				acc.AddFields(m.name, m.fields, m.tags)
				m.fields = make(map[string]interface{})
			}

			// after finishing the bucket, remove its tag from the tags map
			delete(m.tags, aggName)
		}

	case *elastic.AggregationValueMetric:
		if resp.Value != nil {
			m.fields[aggName] = *resp.Value
		} else {
			m.fields[aggName] = float64(0)
		}

	default:
		return m, fmt.Errorf("aggregation type returned not supported")
	}

	return m, nil
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
	}

	if found {
		return agg
	}

	return nil
}

func getAggName(agg elastic.Aggregations) (string, bool) {
	for k := range agg {
		if (k != "key") && (k != "doc_count") {
			return k, true
		}
	}

	return "", false
}
