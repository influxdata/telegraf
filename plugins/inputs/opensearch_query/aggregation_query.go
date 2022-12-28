package opensearch_query

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
)

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

func (o *OpensearchQuery) runAggregationQuery(ctx context.Context, aggregation osAggregation) (*AggregationResponse, error) {
	now := time.Now().UTC()
	from := now.Add(time.Duration(-aggregation.QueryPeriod))
	filterQuery := aggregation.FilterQuery
	if filterQuery == "" {
		filterQuery = "*"
	}

	aq := &Query{
		Size:         0,
		Aggregations: aggregation.aggregation,
		Query:        nil,
	}

	boolQuery := &BoolQuery{
		FilterQueryString: filterQuery,
		TimestampField:    aggregation.DateField,
		TimeRangeFrom:     from,
		TimeRangeTo:       now,
		DateFieldFormat:   aggregation.DateFieldFormat,
	}

	aq.Query = boolQuery
	req, err := json.Marshal(aq)

	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %s", err)
	}
	o.Log.Debugf("{\"body\": %s}", string(req))

	searchRequest := &opensearchapi.SearchRequest{
		Body:    strings.NewReader(string(req)),
		Index:   []string{aggregation.Index},
		Timeout: time.Duration(o.Timeout),
	}

	resp, err := searchRequest.Do(ctx, o.osClient)
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("Opensearch SearchRequest failure: [%d] %s", resp.StatusCode, resp.Status())
	}
	defer resp.Body.Close()

	var searchResult AggregationResponse

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
	var agg AggregationRequest
	agg = &MetricAggregationRequest{}

	// create one aggregation per metric field found & function defined for numeric fields
	for k, v := range aggregation.mapMetricFields {
		switch v {
		case "long", "float", "integer", "short", "double", "scaled_float":
		default:
			continue
		}

		err := agg.AddAggregation(strings.ReplaceAll(k, ".", "_")+"_"+aggregation.MetricFunction, aggregation.MetricFunction, k)
		if err != nil {
			return err
		}
	}

	// create a terms aggregation per tag
	for _, term := range aggregation.Tags {
		bucket := &BucketAggregationRequest{}
		name := strings.ReplaceAll(term, ".", "_")
		err := bucket.AddAggregation(name, "terms", term)
		if err != nil {
			return err
		}
		_ = bucket.BucketSize(name, 1000)
		if aggregation.IncludeMissingTag && aggregation.MissingTagValue != "" {
			bucket.Missing(name, aggregation.MissingTagValue)
		}

		bucket.AddNestedAggregation(name, agg)

		agg = bucket
	}

	aggregation.aggregation = agg

	return nil
}
