package elasticsearch_query

import (
	"context"
	"fmt"
	"strings"
	"time"

	elastic "gopkg.in/olivere/elastic.v5"
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
	aggregation elastic.Aggregation
}

func (e *ElasticsearchQuery) runAggregationQuery(ctx context.Context, aggregation esAggregation) (*elastic.SearchResult, error) {

	now := time.Now().UTC()
	from := now.Add(aggregation.QueryPeriod.Duration * -1)
	filterQuery := aggregation.FilterQuery

	if filterQuery == "" {
		filterQuery = "*"
	}

	query := elastic.NewBoolQuery()
	query = query.Filter(elastic.NewQueryStringQuery(filterQuery))
	query = query.Filter(elastic.NewRangeQuery(aggregation.DateField).From(from).To(now))

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

// GetMetricFields function returns a map of fields and field types on Elasticsearch that matches field.MetricFields
func (e *ElasticsearchQuery) getMetricFields(ctx context.Context, aggregation esAggregation) (map[string]string, error) {
	mapMetricFields := make(map[string]string)

	if e.esClient == nil {
		err := e.connectToES()
		if err != nil {
			return nil, err
		}
	}

	for _, metricField := range aggregation.MetricFields {

		resp, err := e.esClient.GetFieldMapping().Index(aggregation.Index).Field(metricField).Do(ctx)

		if err != nil {
			return mapMetricFields, fmt.Errorf("error retrieving field mappings for %s: %s", aggregation.Index, err.Error())
		}

		for _, index := range resp {
			if mappings, ok := index.(map[string]interface{})["mappings"]; ok {
				if types, ok := mappings.(map[string]interface{}); ok {
					for _, _type := range types {
						if field, ok := _type.(map[string]interface{}); ok {
							for _, _field := range field {
								if fullname, ok := _field.(map[string]interface{})["full_name"]; ok {
									if fname, ok := fullname.(string); ok {
										if mapping, ok := _field.(map[string]interface{})["mapping"]; ok {
											if fieldType, ok := mapping.(map[string]interface{}); ok {
												for _, fieldType := range fieldType {
													if t, ok := fieldType.(map[string]interface{})["type"]; ok {
														if ftype, ok := t.(string); ok {
															mapMetricFields[fname] = ftype
														}
													}
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return mapMetricFields, nil
}

func (e *ElasticsearchQuery) buildAggregationQuery(mapMetricFields map[string]string, aggregation esAggregation) ([]aggregationQueryData, error) {

	var aggregationQueryList []aggregationQueryData

	// create one aggregation per metric field found & function defined for numeric fields
	for k, v := range mapMetricFields {
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

		agg, err := e.getFunctionAggregation(aggregation.MetricFunction, k)
		if err != nil {
			return nil, err
		}

		aggregationQuery := aggregationQueryData{
			aggKey: aggKey{
				measurement: aggregation.MeasurementName,
				function:    aggregation.MetricFunction,
				field:       k,
				name:        strings.Replace(k, ".", "_", -1) + "_" + aggregation.MetricFunction},
			isParent:    true,
			aggregation: agg,
		}

		aggregationQueryList = append(aggregationQueryList, aggregationQuery)

	}

	for _, term := range aggregation.Tags {
		aggregationQueryList = e.getTermsAggregation(aggregation.MeasurementName, term, aggregation.IncludeMissingTag, aggregation.MissingTagValue, aggregationQueryList)
	}

	return aggregationQueryList, nil
}

func (e *ElasticsearchQuery) getFunctionAggregation(function string, aggfield string) (elastic.Aggregation, error) {

	var agg elastic.Aggregation

	switch function {
	case "avg":
		agg = elastic.NewAvgAggregation().Field(aggfield)
	case "sum":
		agg = elastic.NewSumAggregation().Field(aggfield)
	case "min":
		agg = elastic.NewMinAggregation().Field(aggfield)
	case "max":
		agg = elastic.NewMaxAggregation().Field(aggfield)
	default:
		return agg, fmt.Errorf("aggregation function %s not supported", function)
	}

	return agg, nil
}

func (e *ElasticsearchQuery) getTermsAggregation(aggMeasurementName string, aggTerm string, includeMissing bool, missingTagValue string, subAggList []aggregationQueryData) []aggregationQueryData {

	var agg = elastic.NewTermsAggregation()

	if includeMissing && missingTagValue != "" {
		agg.Missing(missingTagValue)
	}

	agg.Field(aggTerm).Size(1000)

	// add each previous parent aggregations as subaggregations of this terms aggregation
	for key, aggMap := range subAggList {
		if aggMap.isParent {
			agg.Field(aggTerm).SubAggregation(aggMap.name, aggMap.aggregation).Size(1000)
			// update subaggregation map with parent information
			subAggList[key].isParent = false
		}
	}

	aggregationQuery := aggregationQueryData{
		aggKey: aggKey{
			measurement: aggMeasurementName,
			function:    "terms",
			field:       aggTerm,
			name:        strings.Replace(aggTerm, ".", "_", -1)},
		isParent:    true,
		aggregation: agg,
	}

	subAggList = append(subAggList, aggregationQuery)

	return subAggList
}
