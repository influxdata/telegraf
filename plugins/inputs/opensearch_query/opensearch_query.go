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

	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	influxtls "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
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
		o.Log.Errorf("Error creating OpenSearch client: %v", err)
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
			return fmt.Errorf("metric field %q not found on index %q", metricField, agg.Index)
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
		return fmt.Errorf("getting username failed: %w", err)
	}
	password, err := o.Password.Get()
	if err != nil {
		config.ReleaseSecret(username)
		return fmt.Errorf("getting password failed: %w", err)
	}

	clientConfig := opensearch.Config{
		Addresses: o.URLs,
		Username:  string(username),
		Password:  string(password),
	}
	config.ReleaseSecret(username)
	config.ReleaseSecret(password)

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
				acc.AddError(fmt.Errorf("opensearch query aggregation %q: %w ", agg.MeasurementName, err))
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
		return nil, fmt.Errorf("opensearch SearchRequest failure: [%d] %s", resp.StatusCode, resp.Status())
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
