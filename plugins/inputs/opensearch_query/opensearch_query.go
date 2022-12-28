//go:generate ../../../tools/readme_config_includer/generator
package opensearch_query

import (
	"context"
	"crypto/tls"
	_ "embed"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	influxtls "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/opensearch-project/opensearch-go/v2"
)

//go:embed sample.conf
var sampleConfig string

// OpensearchQuery struct
type OpensearchQuery struct {
	URLs                []string        `toml:"urls"`
	Username            string          `toml:"username"`
	Password            string          `toml:"password"`
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
		return fmt.Errorf("opensearch urls is not defined")
	}

	err := o.connectToOpensearch()
	if err != nil {
		return fmt.Errorf("error connecting to opensearch: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(o.Timeout))
	defer cancel()

	for i, agg := range o.Aggregations {
		if agg.MeasurementName == "" {
			return fmt.Errorf("field 'measurement_name' is not set")
		}
		if agg.DateField == "" {
			return fmt.Errorf("field 'date_field' is not set")
		}
		err = o.initAggregation(ctx, agg, i)
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *OpensearchQuery) initAggregation(ctx context.Context, agg osAggregation, i int) (err error) {
	// retrieve field mapping and build queries only once
	agg.mapMetricFields, err = o.getMetricFields(ctx, agg)
	if err != nil {
		return fmt.Errorf("not possible to retrieve fields: %v", err.Error())
	}

	for _, metricField := range agg.MetricFields {
		if _, ok := agg.mapMetricFields[metricField]; !ok {
			return fmt.Errorf("metric field '%s' not found on index '%s'", metricField, agg.Index)
		}
	}

	err = agg.buildAggregationQuery()
	if err != nil {
		return fmt.Errorf("error building aggregation: %s", err)
	}

	o.Aggregations[i] = agg
	return nil
}

func (o *OpensearchQuery) connectToOpensearch() error {
	var client *opensearch.Client
	var transport *http.Transport

	if o.InsecureSkipVerify {
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	clientConfig := opensearch.Config{
		Addresses: o.URLs,
		Username:  o.Username,
		Password:  o.Password,
	}

	if transport != nil {
		clientConfig.Transport = transport
	}

	client, err := opensearch.NewClient(clientConfig)
	o.osClient = client

	return err
}

// Gather writes the results of the queries from OpenSearch to the Accumulator.
func (o *OpensearchQuery) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup

	for i, agg := range o.Aggregations {
		wg.Add(1)
		go func(agg osAggregation, i int) {
			defer wg.Done()
			err := o.osAggregationQuery(acc, agg)
			if err != nil {
				acc.AddError(fmt.Errorf("opensearch query aggregation %s: %s ", agg.MeasurementName, err.Error()))
			}
		}(agg, i)
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
