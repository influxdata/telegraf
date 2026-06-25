package elasticsearch_query

import (
	"fmt"
	"strings"

	elastic5 "gopkg.in/olivere/elastic.v5"

	"github.com/influxdata/telegraf"
)

type queryData struct {
	measurement string
	name        string
	function    string
	field       string
	isParent    bool
	aggregation elastic5.Aggregation
}

func (aggregation *aggregation) buildAggregationQuery() error {
	// Create one aggregation per metric field found or function defined for
	// numeric fields
	aggregation.aggregationQueryList = make([]queryData, 0, len(aggregation.mapMetricFields)+len(aggregation.Tags))
	for k, v := range aggregation.mapMetricFields {
		switch v {
		case "long", "float", "integer", "short", "double", "scaled_float":
		default:
			continue
		}

		var agg elastic5.Aggregation
		switch aggregation.MetricFunction {
		case "avg":
			agg = elastic5.NewAvgAggregation().Field(k)
		case "sum":
			agg = elastic5.NewSumAggregation().Field(k)
		case "min":
			agg = elastic5.NewMinAggregation().Field(k)
		case "max":
			agg = elastic5.NewMaxAggregation().Field(k)
		default:
			return fmt.Errorf("aggregation function %q not supported", aggregation.MetricFunction)
		}

		query := queryData{
			measurement: aggregation.MeasurementName,
			function:    aggregation.MetricFunction,
			field:       k,
			name:        strings.ReplaceAll(k, ".", "_") + "_" + aggregation.MetricFunction,
			isParent:    true,
			aggregation: agg,
		}
		aggregation.aggregationQueryList = append(aggregation.aggregationQueryList, query)
	}

	// Create a terms aggregation per tag
	for _, term := range aggregation.Tags {
		agg := elastic5.NewTermsAggregation()
		if aggregation.IncludeMissingTag && aggregation.MissingTagValue != "" {
			agg.Missing(aggregation.MissingTagValue)
		}
		agg.Field(term).Size(1000)

		// add each previous parent aggregations as subaggregations of this terms aggregation
		for key, aggMap := range aggregation.aggregationQueryList {
			if !aggMap.isParent {
				continue
			}

			agg.Field(term).SubAggregation(aggMap.name, aggMap.aggregation).Size(1000)

			// Update subaggregation map with parent information
			aggregation.aggregationQueryList[key].isParent = false
		}

		query := queryData{
			measurement: aggregation.MeasurementName,
			function:    "terms",
			field:       term,
			name:        strings.ReplaceAll(term, ".", "_"),
			isParent:    true,
			aggregation: agg,
		}
		aggregation.aggregationQueryList = append(aggregation.aggregationQueryList, query)
	}

	return nil
}

type resultMetric struct {
	name   string
	fields map[string]interface{}
	tags   map[string]string
}

func (m *resultMetric) recurseResponse(acc telegraf.Accumulator, nameFunction map[string]string, response elastic5.Aggregations) error {
	names := make([]string, 0, len(response))
	for k := range response {
		if k != "key" && k != "doc_count" {
			names = append(names, k)
		}
	}
	if len(names) == 0 {
		// We've reached a single bucket or response without aggregation, i.e.
		// we've reached a leaf node. Add the accumulated metric and reset it
		if len(m.fields) > 0 {
			acc.AddFields(m.name, m.fields, m.tags)
			m.fields = make(map[string]interface{})
		}
		return nil
	}

	// Metrics aggregations response can contain multiple field values, so we
	// iterate over them
	for _, name := range names {
		function, found := nameFunction[name]
		if !found {
			return fmt.Errorf("child aggregation function %q not found %v", name, nameFunction)
		}

		// Execute the aggregation function
		var result interface{}
		switch function {
		case "avg":
			result, _ = response.Avg(name)
		case "sum":
			result, _ = response.Sum(name)
		case "min":
			result, _ = response.Min(name)
		case "max":
			result, _ = response.Max(name)
		case "terms":
			result, _ = response.Terms(name)
		default:
			return fmt.Errorf("aggregation %q not supported", function)
		}

		switch r := result.(type) {
		case *elastic5.AggregationBucketKeyItems:
			// We've found a terms aggregation, iterate over the buckets and try
			// to retrieve the inner aggregation values
			for _, bucket := range r.Buckets {
				s, ok := bucket.Key.(string)
				if !ok {
					return fmt.Errorf("bucket key is not a string (%s, %s)", name, function)
				}
				m.tags[name] = s
				m.fields["doc_count"] = bucket.DocCount

				// We need to recurse down through the buckets, as it may
				// contain another terms aggregation
				if err := m.recurseResponse(acc, nameFunction, bucket.Aggregations); err != nil {
					return err
				}
				delete(m.tags, name)
			}
		case *elastic5.AggregationValueMetric:
			if r.Value != nil {
				m.fields[name] = *r.Value
			} else {
				m.fields[name] = float64(0)
			}
		default:
			return fmt.Errorf("aggregation type %T not supported", r)
		}
	}

	// If there are fields here it comes from a metrics aggregation without a
	// parent terms aggregation
	if len(m.fields) > 0 {
		acc.AddFields(m.name, m.fields, m.tags)
		m.fields = make(map[string]interface{})
	}

	return nil
}
