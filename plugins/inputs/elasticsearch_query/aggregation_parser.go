package elasticsearch_query

import (
	"fmt"

	"github.com/influxdata/telegraf"
	elastic5 "gopkg.in/olivere/elastic.v5"
)

type resultMetric struct {
	name   string
	fields map[string]interface{}
	tags   map[string]string
}

func parseSimpleResult(acc telegraf.Accumulator, measurement string, searchResult *elastic5.SearchResult) {
	fields := make(map[string]interface{})
	tags := make(map[string]string)

	fields["doc_count"] = searchResult.Hits.TotalHits

	acc.AddFields(measurement, fields, tags)
}

func parseAggregationResult(acc telegraf.Accumulator, aggregationQueryList []aggregationQueryData, searchResult *elastic5.SearchResult) error {
	measurements := map[string]map[string]string{}

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

	// recurse over query aggregation results per measurement
	for measurement, aggNameFunction := range measurements {
		var m resultMetric

		m.fields = make(map[string]interface{})
		m.tags = make(map[string]string)
		m.name = measurement

		_, err := recurseResponse(acc, aggNameFunction, searchResult.Aggregations, m)
		if err != nil {
			return err
		}
	}
	return nil
}

func recurseResponse(acc telegraf.Accumulator, aggNameFunction map[string]string, bucketResponse elastic5.Aggregations, m resultMetric) (resultMetric, error) {
	var err error

	aggNames := getAggNames(bucketResponse)
	if len(aggNames) == 0 {
		// we've reached a single bucket or response without aggregation, nothing here
		return m, nil
	}

	// metrics aggregations response can contain multiple field values, so we iterate over them
	for _, aggName := range aggNames {
		aggFunction, found := aggNameFunction[aggName]
		if !found {
			return m, fmt.Errorf("child aggregation function '%s' not found %v", aggName, aggNameFunction)
		}

		resp := getResponseAggregation(aggFunction, aggName, bucketResponse)
		if resp == nil {
			return m, fmt.Errorf("child aggregation '%s' not found", aggName)
		}

		switch resp := resp.(type) {
		case *elastic5.AggregationBucketKeyItems:
			// we've found a terms aggregation, iterate over the buckets and try to retrieve the inner aggregation values
			for _, bucket := range resp.Buckets {
				var s string
				var ok bool
				m.fields["doc_count"] = bucket.DocCount
				if s, ok = bucket.Key.(string); !ok {
					return m, fmt.Errorf("bucket key is not a string (%s, %s)", aggName, aggFunction)
				}
				m.tags[aggName] = s

				// we need to recurse down through the buckets, as it may contain another terms aggregation
				m, err = recurseResponse(acc, aggNameFunction, bucket.Aggregations, m)
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

		case *elastic5.AggregationValueMetric:
			if resp.Value != nil {
				m.fields[aggName] = *resp.Value
			} else {
				m.fields[aggName] = float64(0)
			}

		default:
			return m, fmt.Errorf("aggregation type %T not supported", resp)
		}
	}

	// if there are fields here it comes from a metrics aggregation without a parent terms aggregation
	if len(m.fields) > 0 {
		acc.AddFields(m.name, m.fields, m.tags)
		m.fields = make(map[string]interface{})
	}
	return m, nil
}

func getResponseAggregation(function string, aggName string, aggs elastic5.Aggregations) (agg interface{}) {
	switch function {
	case "avg":
		agg, _ = aggs.Avg(aggName)
	case "sum":
		agg, _ = aggs.Sum(aggName)
	case "min":
		agg, _ = aggs.Min(aggName)
	case "max":
		agg, _ = aggs.Max(aggName)
	case "terms":
		agg, _ = aggs.Terms(aggName)
	}

	return agg
}

// getAggNames returns the aggregation names from a response aggregation
func getAggNames(agg elastic5.Aggregations) (aggs []string) {
	for k := range agg {
		if (k != "key") && (k != "doc_count") {
			aggs = append(aggs, k)
		}
	}

	return aggs
}
