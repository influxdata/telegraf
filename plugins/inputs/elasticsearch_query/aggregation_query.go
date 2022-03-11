package elasticsearch_query

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	elastic5 "gopkg.in/olivere/elastic.v5"
)

type aggKey struct {
	measurement string
	name        string
	function    string
	field       string
}

type aggregationQueryData struct {
	aggKey
	isParent    bool
	aggregation elastic5.Aggregation
}

func (e *ElasticsearchQuery) runAggregationQuery(ctx context.Context, aggregation esAggregation) (*elastic5.SearchResult, error) {
	now := time.Now().UTC()
	from := now.Add(time.Duration(-aggregation.QueryPeriod))
	filterQuery := aggregation.FilterQuery
	if filterQuery == "" {
		filterQuery = "*"
	}

	query := elastic5.NewBoolQuery()
	query = query.Filter(elastic5.NewQueryStringQuery(filterQuery))
	query = query.Filter(elastic5.NewRangeQuery(aggregation.DateField).From(from).To(now).Format(aggregation.DateFieldFormat))

	src, err := query.Source()
	if err != nil {
		return nil, fmt.Errorf("failed to get query source - %v", err)
	}
	data, err := json.Marshal(src)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response - %v", err)
	}
	e.Log.Debugf("{\"query\": %s}", string(data))

	search := e.esClient.Search().Index(aggregation.Index).Query(query).Size(0)

	// add only parent elastic.Aggregations to the search request, all the rest are subaggregations of these
	for _, v := range aggregation.aggregationQueryList {
		if v.isParent && v.aggregation != nil {
			search.Aggregation(v.aggKey.name, v.aggregation)
		}
	}

	searchResult, err := search.Do(ctx)
	if err != nil && searchResult != nil {
		return searchResult, fmt.Errorf("%s - %s", searchResult.Error.Type, searchResult.Error.Reason)
	}

	return searchResult, err
}

// getMetricFields function returns a map of fields and field types on Elasticsearch that matches field.MetricFields
func (e *ElasticsearchQuery) getMetricFields(ctx context.Context, aggregation esAggregation) (map[string]string, error) {
	mapMetricFields := make(map[string]string)

	for _, metricField := range aggregation.MetricFields {
		resp, err := e.esClient.GetFieldMapping().Index(aggregation.Index).Field(metricField).Do(ctx)
		if err != nil {
			return mapMetricFields, fmt.Errorf("error retrieving field mappings for %s: %s", aggregation.Index, err.Error())
		}

		for _, index := range resp {
			var ok bool
			var mappings interface{}
			if mappings, ok = index.(map[string]interface{})["mappings"]; !ok {
				return nil, fmt.Errorf("assertion error, wrong type (expected map[string]interface{}, got %T)", index)
			}

			var types map[string]interface{}
			if types, ok = mappings.(map[string]interface{}); !ok {
				return nil, fmt.Errorf("assertion error, wrong type (expected map[string]interface{}, got %T)", mappings)
			}

			var fields map[string]interface{}
			for _, _type := range types {
				if fields, ok = _type.(map[string]interface{}); !ok {
					return nil, fmt.Errorf("assertion error, wrong type (expected map[string]interface{}, got %T)", _type)
				}

				var field map[string]interface{}
				for _, _field := range fields {
					if field, ok = _field.(map[string]interface{}); !ok {
						return nil, fmt.Errorf("assertion error, wrong type (expected map[string]interface{}, got %T)", _field)
					}

					fullname := field["full_name"]
					mapping := field["mapping"]

					var fname string
					if fname, ok = fullname.(string); !ok {
						return nil, fmt.Errorf("assertion error, wrong type (expected string, got %T)", fullname)
					}

					var fieldTypes map[string]interface{}
					if fieldTypes, ok = mapping.(map[string]interface{}); !ok {
						return nil, fmt.Errorf("assertion error, wrong type (expected map[string]interface{}, got %T)", mapping)
					}

					var fieldType interface{}
					for _, _fieldType := range fieldTypes {
						if fieldType, ok = _fieldType.(map[string]interface{})["type"]; !ok {
							return nil, fmt.Errorf("assertion error, wrong type (expected map[string]interface{}, got %T)", _fieldType)
						}

						var ftype string
						if ftype, ok = fieldType.(string); !ok {
							return nil, fmt.Errorf("assertion error, wrong type (expected string, got %T)", fieldType)
						}
						mapMetricFields[fname] = ftype
					}
				}
			}
		}
	}

	return mapMetricFields, nil
}

func (aggregation *esAggregation) buildAggregationQuery() error {
	// create one aggregation per metric field found & function defined for numeric fields
	for k, v := range aggregation.mapMetricFields {
		switch v {
		case "long":
		case "float":
		case "integer":
		case "short":
		case "double":
		case "scaled_float":
		default:
			continue
		}

		agg, err := getFunctionAggregation(aggregation.MetricFunction, k)
		if err != nil {
			return err
		}

		aggregationQuery := aggregationQueryData{
			aggKey: aggKey{
				measurement: aggregation.MeasurementName,
				function:    aggregation.MetricFunction,
				field:       k,
				name:        strings.Replace(k, ".", "_", -1) + "_" + aggregation.MetricFunction,
			},
			isParent:    true,
			aggregation: agg,
		}

		aggregation.aggregationQueryList = append(aggregation.aggregationQueryList, aggregationQuery)
	}

	// create a terms aggregation per tag
	for _, term := range aggregation.Tags {
		agg := elastic5.NewTermsAggregation()
		if aggregation.IncludeMissingTag && aggregation.MissingTagValue != "" {
			agg.Missing(aggregation.MissingTagValue)
		}

		agg.Field(term).Size(1000)

		// add each previous parent aggregations as subaggregations of this terms aggregation
		for key, aggMap := range aggregation.aggregationQueryList {
			if aggMap.isParent {
				agg.Field(term).SubAggregation(aggMap.name, aggMap.aggregation).Size(1000)
				// update subaggregation map with parent information
				aggregation.aggregationQueryList[key].isParent = false
			}
		}

		aggregationQuery := aggregationQueryData{
			aggKey: aggKey{
				measurement: aggregation.MeasurementName,
				function:    "terms",
				field:       term,
				name:        strings.Replace(term, ".", "_", -1),
			},
			isParent:    true,
			aggregation: agg,
		}

		aggregation.aggregationQueryList = append(aggregation.aggregationQueryList, aggregationQuery)
	}

	return nil
}

func getFunctionAggregation(function string, aggfield string) (elastic5.Aggregation, error) {
	var agg elastic5.Aggregation

	switch function {
	case "avg":
		agg = elastic5.NewAvgAggregation().Field(aggfield)
	case "sum":
		agg = elastic5.NewSumAggregation().Field(aggfield)
	case "min":
		agg = elastic5.NewMinAggregation().Field(aggfield)
	case "max":
		agg = elastic5.NewMaxAggregation().Field(aggfield)
	default:
		return nil, fmt.Errorf("aggregation function '%s' not supported", function)
	}

	return agg, nil
}
