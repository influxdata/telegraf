package elasticsearch_query

import (
	"context"
	"fmt"
	"strings"
	"time"

	elastic "gopkg.in/olivere/elastic.v5"
)

func (e *ElasticsearchQuery) runAggregationQuery(ctx context.Context, aggregation Aggregation, aggregationQueryList []aggregationQueryData) (*elastic.SearchResult, error) {

	now := time.Now().UTC()
	from := now.Sub(aggregation.QueryPeriod.Duration)
	filterQuery := aggregation.FilterQuery

	if filterQuery == "" {
		filterQuery = "*"
	}

	query := elastic.NewBoolQuery()
	query = query.Filter(elastic.NewQueryStringQuery(filterQuery))
	query = query.Filter(elastic.NewRangeQuery(aggregation.DateField).From(from).To(now))

	search := e.ESClient.Search().
		Index(aggregation.Index).
		Query(query).
		Size(0)

	// add only parent elastic.Aggregations to the search request, all the rest are subaggregations of these
	for _, v := range aggregationQueryList {
		if v.isParent && v.aggregation != nil {
			search.Aggregation(v.aggKey.name, v.aggregation)
		}
	}
	searchResult, err := search.Do(ctx)

	if err != nil {
		if searchResult != nil {
			return searchResult, fmt.Errorf("%s - %s", searchResult.Error.Type, searchResult.Error.Reason)
		}
	}

	return searchResult, err

}

// GetMetricFields function returns a map of fields and field types on Elasticsearch that matches field.MetricFields
func (e *ElasticsearchQuery) getMetricFields(ctx context.Context, aggregation Aggregation) (map[string]string, error) {
	var ftype string
	mapMetricFields := make(map[string]string)

	if e.ESClient == nil {
		err := e.connectToES()
		if err != nil {
			return nil, err
		}
	}

	// TODO: check if it is possible to improve this
	for _, metricField := range aggregation.MetricFields {

		resp, err := e.ESClient.GetFieldMapping().Index(aggregation.Index).Field(metricField).Do(ctx)

		if err != nil {
			return mapMetricFields, err
		}

		for _, index := range resp {
			for _, _type := range index.(map[string]interface{})["mappings"].(map[string]interface{}) {
				for _, field := range _type.(map[string]interface{}) {
					fname := field.(map[string]interface{})["full_name"].(string)
					for _, fieldType := range field.(map[string]interface{})["mapping"].(map[string]interface{}) {
						ftype = fieldType.(map[string]interface{})["type"].(string)
					}
					mapMetricFields[fname] = ftype
				}
			}
		}
	}

	return mapMetricFields, nil
}

func (e *ElasticsearchQuery) buildAggregationQuery(mapMetricFields map[string]string, aggregation Aggregation) ([]aggregationQueryData, error) {

	aggregationQueryList := make([]aggregationQueryData, (len(aggregation.Tags) + len(aggregation.MetricFields)))

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
	// TODO:
	// case "percentile":
	// 	agg = elastic.NewPercentilesAggregation().Field(aggfield)
	default:
		return agg, fmt.Errorf("Aggregation function %s not supported", function)
	}

	return agg, nil
}

func (e *ElasticsearchQuery) getTermsAggregation(aggMeasurementName string, aggTerm string, includeMissing bool, missingTagValue string, subAggList []aggregationQueryData) []aggregationQueryData {

	var agg = elastic.NewTermsAggregation()

	if includeMissing && missingTagValue != "" {
		agg.Missing(missingTagValue)
	}

	// add each previous parent aggregations as subaggregations of this terms aggregation
	for key, aggMap := range subAggList {
		if aggMap.isParent {
			agg.Field(aggTerm).SubAggregation(aggMap.name, aggMap.aggregation).Size(1000)
			// update subaggregation map with parent information
			subAggList[key].isParent = false
		} else {
			agg.Field(aggTerm).Size(1000)
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
