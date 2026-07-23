//go:generate ../../../tools/readme_config_includer/generator
package elasticsearch_query

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"maps"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	common_http "github.com/influxdata/telegraf/plugins/common/http"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type client interface {
	close()
	getFieldMapping(context.Context, string, string) (map[string]interface{}, error)
	query(context.Context, *aggregation) (interface{}, int64, error)
}

//go:embed sample.conf
var sampleConfig string

type ElasticsearchQuery struct {
	URLs                []string        `toml:"urls"`
	Username            string          `toml:"username"`
	Password            string          `toml:"password"`
	EnableSniffer       bool            `toml:"enable_sniffer"`
	HealthCheckInterval config.Duration `toml:"health_check_interval"`
	Aggregations        []aggregation   `toml:"aggregation"`
	Log                 telegraf.Logger `toml:"-"`
	common_http.HTTPClientConfig

	client client
}

type aggregation struct {
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

	mapMetricFields map[string]string
	measurements    map[string]map[string]string
	queries         interface{} // prepared once and reused across collections
}

func (*ElasticsearchQuery) SampleConfig() string {
	return sampleConfig
}

func (e *ElasticsearchQuery) Init() error {
	if e.URLs == nil {
		return errors.New("no urls defined")
	}

	for i := range e.Aggregations {
		agg := &e.Aggregations[i]

		if agg.MeasurementName == "" {
			return errors.New("field 'measurement_name' is not set")
		}
		if agg.DateField == "" {
			return errors.New("field 'date_field' is not set")
		}
		if agg.FilterQuery == "" {
			agg.FilterQuery = "*"
		}
	}

	return nil
}

func (e *ElasticsearchQuery) Start(telegraf.Accumulator) error {
	// Make sure the HTTP client exists
	httpClient, err := e.HTTPClientConfig.CreateClient(context.Background(), e.Log)
	if err != nil {
		return fmt.Errorf("creating HTTP client failed: %w", err)
	}

	cfg := clientConfig{
		urls:              e.URLs,
		username:          e.Username,
		password:          e.Password,
		enableSniffer:     e.EnableSniffer,
		discoveryInterval: time.Duration(e.HealthCheckInterval),
		httpClient:        httpClient,
		log:               e.Log,
	}

	version, major, err := cfg.probeVersion(context.Background())
	if err != nil {
		httpClient.CloseIdleConnections()
		return err
	}

	var c client
	switch major {
	case 5:
		if cfg.enableSniffer {
			e.Log.Warn("'enable_sniffer' is not supported for ElasticSearch 5.x and will be ignored")
		}
		c, err = newClientV5(cfg)
	case 6:
		c, err = newClientV6(cfg)
	default:
		httpClient.CloseIdleConnections()
		return fmt.Errorf("server version %q not supported (currently supported versions are 5.x and 6.x)", version)
	}
	if err != nil {
		return err
	}
	e.client = c

	// Setup the aggregations, this needs to be done in Start as it will require
	// API calls to the ElasticSearch endpoint and can thus not happen in Init
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(e.Timeout))
	defer cancel()

	for i := range e.Aggregations {
		agg := &e.Aggregations[i]
		if err := e.initAggregation(ctx, agg); err != nil {
			e.Stop()
			return fmt.Errorf("initializing aggregation %q failed: %w", agg.MeasurementName, err)
		}
	}

	return nil
}

func (e *ElasticsearchQuery) Stop() {
	if e.client != nil {
		e.client.close()
	}
}

// Gather writes the results of the queries from Elasticsearch to the Accumulator.
func (e *ElasticsearchQuery) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup
	for i := range e.Aggregations {
		wg.Add(1)
		go func(agg *aggregation) {
			defer wg.Done()
			if err := e.gatherAggregation(acc, agg); err != nil {
				acc.AddError(fmt.Errorf("querying aggregation %q failed: %w", agg.MeasurementName, err))
			}
		}(&e.Aggregations[i])
	}
	wg.Wait()

	return nil
}

func (e *ElasticsearchQuery) initAggregation(ctx context.Context, agg *aggregation) error {
	// retrieve field mapping and build queries only once
	agg.mapMetricFields = make(map[string]string, len(agg.MetricFields))
	for _, f := range agg.MetricFields {
		response, err := e.client.getFieldMapping(ctx, agg.Index, f)
		if err != nil {
			return fmt.Errorf("retrieving index %q field mappings for %q failed: %w", agg.Index, f, err)
		}

		fields, err := getMetricField(response)
		if err != nil {
			return fmt.Errorf("not possible to retrieve field %q: %w", f, err)
		}
		maps.Copy(agg.mapMetricFields, fields)
	}

	for _, metricField := range agg.MetricFields {
		if _, ok := agg.mapMetricFields[metricField]; !ok {
			return fmt.Errorf("metric field %q not found on index %q", metricField, agg.Index)
		}
	}

	if err := agg.buildQueries(); err != nil {
		return fmt.Errorf("building aggregation query failed: %w", err)
	}

	return nil
}

func (e *ElasticsearchQuery) gatherAggregation(acc telegraf.Accumulator, aggregation *aggregation) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(e.Timeout))
	defer cancel()

	result, hits, err := e.client.query(ctx, aggregation)
	if err != nil {
		return fmt.Errorf("running query failed: %w", err)
	}

	// Handle simple non-aggregated results
	if result == nil {
		fields := map[string]interface{}{
			"doc_count": hits,
		}
		tags := make(map[string]string)
		acc.AddFields(aggregation.MeasurementName, fields, tags)
		return nil
	}

	// Aggregate results that support aggregation
	for measurement, aggNameFunction := range aggregation.measurements {
		if err := aggregate(acc, measurement, aggNameFunction, result); err != nil {
			return fmt.Errorf("recursing response failed: %w", err)
		}
	}

	return nil
}

func getMetricField(response map[string]interface{}) (map[string]string, error) {
	mapMetricFields := make(map[string]string, len(response))
	for _, index := range response {
		idx, ok := index.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected type %T for index", index)
		}
		mappings, found := idx["mappings"]
		if !found {
			return nil, errors.New("no mapping found in index")
		}

		types, ok := mappings.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected type %T for mappings", mappings)
		}

		for _, t := range types {
			fields, ok := t.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("unexpected type %T for types", t)
			}

			for _, f := range fields {
				field, ok := f.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("unexpected type %T for field", f)
				}

				fullname, ok := field["full_name"].(string)
				if !ok {
					return nil, fmt.Errorf("unexpected type %T for full_name field", field["full_name"])
				}

				mapping, ok := field["mapping"].(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("unexpected type %T for mapping field", field["mapping"])
				}

				for _, fm := range mapping {
					fieldType, ok := fm.(map[string]interface{})
					if !ok {
						return nil, fmt.Errorf("unexpected type %T for field", fm)
					}

					ftype, ok := fieldType["type"].(string)
					if !ok {
						return nil, fmt.Errorf("unexpected type %T for field type", fieldType["type"])
					}
					mapMetricFields[fullname] = ftype
				}
			}
		}
	}

	return mapMetricFields, nil
}

func init() {
	inputs.Add("elasticsearch_query", func() telegraf.Input {
		return &ElasticsearchQuery{
			HealthCheckInterval: config.Duration(15 * time.Minute),
			HTTPClientConfig: common_http.HTTPClientConfig{
				Timeout: config.Duration(5 * time.Second),
				TransportConfig: common_http.TransportConfig{
					ResponseHeaderTimeout: config.Duration(5 * time.Second),
				},
			},
		}
	})
}
