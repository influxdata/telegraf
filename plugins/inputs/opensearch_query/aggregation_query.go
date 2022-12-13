package opensearch_query

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
	"github.com/pkg/errors"
	"strings"
	"time"

	"github.com/olivere/elastic/v7"
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

type mapping map[string]fieldIndex

type fieldIndex struct {
	Mappings map[string]fieldMapping `json:"mappings"`
}

type fieldMapping struct {
	FullName string               `json:"full_name"`
	Mapping  map[string]fieldType `json:"mapping"`
}

type fieldType struct {
	Type string `json:"type"`
}

func (o *OpensearchQuery) runAggregationQuery(ctx context.Context, aggregation osAggregation) (*elastic.SearchResult, error) {
	now := time.Now().UTC()
	from := now.Add(time.Duration(-aggregation.QueryPeriod))
	filterQuery := aggregation.FilterQuery
	if filterQuery == "" {
		filterQuery = "*"
	}

	query := elastic.NewBoolQuery()
	query = query.Filter(elastic.NewQueryStringQuery(filterQuery))
	query = query.Filter(elastic.NewRangeQuery(aggregation.DateField).From(from).To(now).Format(aggregation.DateFieldFormat))

	src, err := query.Source()
	if err != nil {
		return nil, fmt.Errorf("failed to get query source - %v", err)
	}
	data, err := json.Marshal(src)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response - %v", err)
	}
	o.Log.Debugf("{\"query\": %s}", string(data))

	searchSource := elastic.NewSearchSource().Query(query).Size(0)
	// add only parent elastic.Aggregations to the resp request, all the rest are subaggregations of these
	for _, v := range aggregation.aggregationQueryList {
		if v.isParent && v.aggregation != nil {
			searchSource.Aggregation(v.aggKey.name, v.aggregation)
		}
	}

	ss, err := searchSource.Source()
	if err != nil {
		return nil, fmt.Errorf("failed to get query source - %v", err)
	}
	s, err := json.Marshal(ss)
	if err != nil {
		return nil, err
	}

	req := strings.NewReader(string(s))

	searchRequest := &opensearchapi.SearchRequest{
		Body:  req,
		Index: []string{aggregation.Index},
		//Timeout: time.Duration(o.Timeout),
	}

	resp, err := searchRequest.Do(ctx, o.osClient)
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, errors.Errorf("Opensearch SearchRequest failure: [%d] %s", resp.StatusCode, resp.Status())
	}
	defer resp.Body.Close()

	var searchResult elastic.SearchResult

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&searchResult)
	if err != nil {
		return nil, err
	}

	return &searchResult, nil
}

// getMetricFields function returns a map of fields and field types on Elasticsearch that matches field.MetricFields
func (o *OpensearchQuery) getMetricFields(ctx context.Context, aggregation osAggregation) (map[string]string, error) {
	mapMetricFields := make(map[string]string)
	fieldMappingRequest := opensearchapi.IndicesGetFieldMappingRequest{
		Index:  []string{aggregation.Index},
		Fields: aggregation.MetricFields,
	}

	response, err := fieldMappingRequest.Do(ctx, o.osClient)
	if err != nil {
		return nil, err
	}

	// Bad request; move on
	if response.StatusCode == 400 {
		return mapMetricFields, nil
	}

	var m mapping
	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&m)
	if err != nil {
		return nil, err
	}

	for _, mm := range m {
		for _, f := range aggregation.MetricFields {
			if _, ok := mm.Mappings[f]; ok {
				mapMetricFields[f] = mm.Mappings[f].Mapping[f].Type
			}
		}
	}

	return mapMetricFields, nil
}

func (aggregation *osAggregation) buildAggregationQuery() error {
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
				name:        strings.ReplaceAll(k, ".", "_") + "_" + aggregation.MetricFunction,
			},
			isParent:    true,
			aggregation: agg,
		}

		aggregation.aggregationQueryList = append(aggregation.aggregationQueryList, aggregationQuery)
	}

	// create a terms aggregation per tag
	for _, term := range aggregation.Tags {
		agg := elastic.NewTermsAggregation()
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
				name:        strings.ReplaceAll(term, ".", "_"),
			},
			isParent:    true,
			aggregation: agg,
		}

		aggregation.aggregationQueryList = append(aggregation.aggregationQueryList, aggregationQuery)
	}

	return nil
}

func getFunctionAggregation(function string, aggfield string) (elastic.Aggregation, error) {
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
		return nil, fmt.Errorf("aggregation function '%s' not supported", function)
	}

	return agg, nil
}
