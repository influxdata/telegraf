//go:generate ../../../tools/readme_config_includer/generator
package opensearch_query

import (
	"context"
	"crypto/tls"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	influxtls "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
)

//go:embed sample.conf
var sampleConfig string

// OpensearchQuery struct
type OpensearchQuery struct {
	URLs                []string        `toml:"urls"`
	Username            config.Secret   `toml:"username"`
	Password            config.Secret   `toml:"password"`
	EnableSniffer       bool            `toml:"enable_sniffer"`
	Timeout             config.Duration `toml:"timeout"`
	HealthCheckInterval config.Duration `toml:"health_check_interval"`
	Aggregations        []osAggregation `toml:"aggregation"`

	Log telegraf.Logger `toml:"-"`

	influxtls.ClientConfig
	osClient *opensearch.Client
}

// osAggregation struct
type osAggregation struct {
	Index             string          `toml:"index"`
	MeasurementName   string          `toml:"measurement_name"`
	DateField         string          `toml:"date_field"`
	DateFieldFormat   string          `toml:"date_field_custom_format"`
	QueryPeriod       config.Duration `toml:"query_period"`
	FilterQuery       string          `toml:"filter_query"`
	MetricFields      []string        `toml:"metric_fields"`
	MetricFunction    string          `toml:"metric_function"`
	Tags              []string        `toml:"tags"`
	IncludeMissingTag bool            `toml:"include_missing_tag"`
	MissingTagValue   string          `toml:"missing_tag_value"`
	mapMetricFields   map[string]string

	aggregation AggregationRequest
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

func (*OpensearchQuery) SampleConfig() string {
	return sampleConfig
}

// Init the plugin.
func (o *OpensearchQuery) Init() error {
	if o.URLs == nil {
		return fmt.Errorf("no urls defined")
	}

	err := o.newClient()
	if err != nil {
		o.Log.Errorf("error creating OpenSearch client: %w", err)
	}

	for i, agg := range o.Aggregations {
		if agg.MeasurementName == "" {
			return fmt.Errorf("field 'measurement_name' is not set")
		}
		if agg.DateField == "" {
			return fmt.Errorf("field 'date_field' is not set")
		}
		err = o.initAggregation(agg, i)
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *OpensearchQuery) initAggregation(agg osAggregation, i int) (err error) {
	for _, metricField := range agg.MetricFields {
		if _, ok := agg.mapMetricFields[metricField]; !ok {
			return fmt.Errorf("metric field '%s' not found on index '%s'", metricField, agg.Index)
		}
	}

	err = agg.buildAggregationQuery()
	if err != nil {
		return fmt.Errorf("error building aggregation: %w", err)
	}

	o.Aggregations[i] = agg
	return nil
}

func (o *OpensearchQuery) newClient() error {
	username, err := o.Username.Get()
	if err != nil {
		return fmt.Errorf("getting username failed: %v", err)
	}
	defer config.ReleaseSecret(username)
	password, err := o.Password.Get()
	if err != nil {
		return fmt.Errorf("getting password failed: %v", err)
	}
	defer config.ReleaseSecret(password)

	clientConfig := opensearch.Config{
		Addresses: o.URLs,
		Username:  string(username),
		Password:  string(password),
	}

	if o.InsecureSkipVerify {
		clientConfig.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	client, err := opensearch.NewClient(clientConfig)
	o.osClient = client

	return err
}

// Gather writes the results of the queries from OpenSearch to the Accumulator.
func (o *OpensearchQuery) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup

	for _, agg := range o.Aggregations {
		wg.Add(1)
		go func(agg osAggregation) {
			defer wg.Done()
			err := o.osAggregationQuery(acc, agg)
			if err != nil {
				acc.AddError(fmt.Errorf("opensearch query aggregation %s: %s ", agg.MeasurementName, err))
			}
		}(agg)
	}

	wg.Wait()
	return nil
}

func (o *OpensearchQuery) osAggregationQuery(acc telegraf.Accumulator, aggregation osAggregation) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(o.Timeout))
	defer cancel()

	searchResult, err := o.runAggregationQuery(ctx, aggregation)
	if err != nil {
		return err
	}

	return searchResult.GetMetrics(acc, aggregation.MeasurementName)
}

func init() {
	inputs.Add("opensearch_query", func() telegraf.Input {
		return &OpensearchQuery{
			Timeout:             config.Duration(time.Second * 5),
			HealthCheckInterval: config.Duration(time.Second * 10),
		}
	})
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
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

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
	if response.StatusCode != 200 {
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
