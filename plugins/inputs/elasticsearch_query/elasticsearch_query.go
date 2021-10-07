package elasticsearch_query

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	elastic5 "gopkg.in/olivere/elastic.v5"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const sampleConfig = `
  ## The full HTTP endpoint URL for your Elasticsearch instance
  ## Multiple urls can be specified as part of the same cluster,
  ## this means that only ONE of the urls will be written to each interval.
  urls = [ "http://node1.es.example.com:9200" ] # required.

  ## Elasticsearch client timeout, defaults to "5s".
  # timeout = "5s"

  ## Set to true to ask Elasticsearch a list of all cluster nodes,
  ## thus it is not necessary to list all nodes in the urls config option
  # enable_sniffer = false

  ## Set the interval to check if the Elasticsearch nodes are available
  ## This option is only used if enable_sniffer is also set (0s to disable it)
  # health_check_interval = "10s"

  ## HTTP basic authentication details (eg. when using x-pack)
  # username = "telegraf"
  # password = "mypassword"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  [[inputs.elasticsearch_query.aggregation]]
    ## measurement name for the results of the aggregation query
    measurement_name = "measurement"

    ## Elasticsearch indexes to query (accept wildcards).
    index = "index-*"

    ## The date/time field in the Elasticsearch index (mandatory).
    date_field = "@timestamp"

    ## If the field used for the date/time field in Elasticsearch is also using
    ## a custom date/time format it may be required to provide the format to
    ## correctly parse the field.
    ##
    ## If using one of the built in elasticsearch formats this is not required.
    # date_field_custom_format = ""

    ## Time window to query (eg. "1m" to query documents from last minute).
    ## Normally should be set to same as collection interval
    query_period = "1m"

    ## Lucene query to filter results
    # filter_query = "*"

    ## Fields to aggregate values (must be numeric fields)
    # metric_fields = ["metric"]

    ## Aggregation function to use on the metric fields
    ## Must be set if 'metric_fields' is set
    ## Valid values are: avg, sum, min, max, sum
    # metric_function = "avg"

    ## Fields to be used as tags
    ## Must be text, non-analyzed fields. Metric aggregations are performed per tag
    # tags = ["field.keyword", "field2.keyword"]

    ## Set to true to not ignore documents when the tag(s) above are missing
    # include_missing_tag = false

    ## String value of the tag when the tag does not exist
    ## Used when include_missing_tag is true
    # missing_tag_value = "null"
`

// ElasticsearchQuery struct
type ElasticsearchQuery struct {
	URLs                []string        `toml:"urls"`
	Username            string          `toml:"username"`
	Password            string          `toml:"password"`
	EnableSniffer       bool            `toml:"enable_sniffer"`
	Timeout             config.Duration `toml:"timeout"`
	HealthCheckInterval config.Duration `toml:"health_check_interval"`
	Aggregations        []esAggregation `toml:"aggregation"`

	Log telegraf.Logger `toml:"-"`

	tls.ClientConfig
	httpclient *http.Client
	esClient   *elastic5.Client
}

// esAggregation struct
type esAggregation struct {
	Index                string          `toml:"index"`
	MeasurementName      string          `toml:"measurement_name"`
	DateField            string          `toml:"date_field"`
	DateFieldFormat      string          `toml:"date_field_custom_format"`
	QueryPeriod          config.Duration `toml:"query_period"`
	FilterQuery          string          `toml:"filter_query"`
	MetricFields         []string        `toml:"metric_fields"`
	MetricFunction       string          `toml:"metric_function"`
	Tags                 []string        `toml:"tags"`
	IncludeMissingTag    bool            `toml:"include_missing_tag"`
	MissingTagValue      string          `toml:"missing_tag_value"`
	mapMetricFields      map[string]string
	aggregationQueryList []aggregationQueryData
}

// SampleConfig returns sample configuration for this plugin.
func (e *ElasticsearchQuery) SampleConfig() string {
	return sampleConfig
}

// Description returns the plugin description.
func (e *ElasticsearchQuery) Description() string {
	return `Derive metrics from aggregating Elasticsearch query results`
}

// Init the plugin.
func (e *ElasticsearchQuery) Init() error {
	if e.URLs == nil {
		return fmt.Errorf("elasticsearch urls is not defined")
	}

	err := e.connectToES()
	if err != nil {
		e.Log.Errorf("E! error connecting to elasticsearch: %s", err)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(e.Timeout))
	defer cancel()

	for i, agg := range e.Aggregations {
		if agg.MeasurementName == "" {
			return fmt.Errorf("field 'measurement_name' is not set")
		}
		if agg.DateField == "" {
			return fmt.Errorf("field 'date_field' is not set")
		}
		err = e.initAggregation(ctx, agg, i)
		if err != nil {
			e.Log.Errorf("%s", err)
			return nil
		}
	}
	return nil
}

func (e *ElasticsearchQuery) initAggregation(ctx context.Context, agg esAggregation, i int) (err error) {
	// retrieve field mapping and build queries only once
	agg.mapMetricFields, err = e.getMetricFields(ctx, agg)
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
		return err
	}

	e.Aggregations[i] = agg
	return nil
}

func (e *ElasticsearchQuery) connectToES() error {
	var clientOptions []elastic5.ClientOptionFunc

	if e.esClient != nil {
		if e.esClient.IsRunning() {
			return nil
		}
	}

	if e.httpclient == nil {
		httpclient, err := e.createHTTPClient()
		if err != nil {
			return err
		}
		e.httpclient = httpclient
	}

	clientOptions = append(clientOptions,
		elastic5.SetHttpClient(e.httpclient),
		elastic5.SetSniff(e.EnableSniffer),
		elastic5.SetURL(e.URLs...),
		elastic5.SetHealthcheckInterval(time.Duration(e.HealthCheckInterval)),
	)

	if e.Username != "" {
		clientOptions = append(clientOptions, elastic5.SetBasicAuth(e.Username, e.Password))
	}

	if time.Duration(e.HealthCheckInterval) == 0 {
		clientOptions = append(clientOptions, elastic5.SetHealthcheck(false))
	}

	client, err := elastic5.NewClient(clientOptions...)
	if err != nil {
		return err
	}

	// check for ES version on first node
	esVersion, err := client.ElasticsearchVersion(e.URLs[0])
	if err != nil {
		return fmt.Errorf("elasticsearch version check failed: %s", err)
	}

	esVersionSplit := strings.Split(esVersion, ".")

	// quit if ES version is not supported
	if len(esVersionSplit) == 0 {
		return fmt.Errorf("elasticsearch version check failed")
	}

	i, err := strconv.Atoi(esVersionSplit[0])
	if err != nil || i < 5 || i > 6 {
		return fmt.Errorf("elasticsearch version %s not supported (currently supported versions are 5.x and 6.x)", esVersion)
	}

	e.esClient = client
	return nil
}

// Gather writes the results of the queries from Elasticsearch to the Accumulator.
func (e *ElasticsearchQuery) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup

	err := e.connectToES()
	if err != nil {
		return err
	}

	for i, agg := range e.Aggregations {
		wg.Add(1)
		go func(agg esAggregation, i int) {
			defer wg.Done()
			err := e.esAggregationQuery(acc, agg, i)
			if err != nil {
				acc.AddError(fmt.Errorf("elasticsearch query aggregation %s: %s ", agg.MeasurementName, err.Error()))
			}
		}(agg, i)
	}

	wg.Wait()
	return nil
}

func (e *ElasticsearchQuery) createHTTPClient() (*http.Client, error) {
	tlsCfg, err := e.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}
	tr := &http.Transport{
		ResponseHeaderTimeout: time.Duration(e.Timeout),
		TLSClientConfig:       tlsCfg,
	}
	httpclient := &http.Client{
		Transport: tr,
		Timeout:   time.Duration(e.Timeout),
	}

	return httpclient, nil
}

func (e *ElasticsearchQuery) esAggregationQuery(acc telegraf.Accumulator, aggregation esAggregation, i int) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(e.Timeout))
	defer cancel()

	// try to init the aggregation query if it is not done already
	if aggregation.aggregationQueryList == nil {
		err := e.initAggregation(ctx, aggregation, i)
		if err != nil {
			return err
		}
		aggregation = e.Aggregations[i]
	}

	searchResult, err := e.runAggregationQuery(ctx, aggregation)
	if err != nil {
		return err
	}

	if searchResult.Aggregations == nil {
		parseSimpleResult(acc, aggregation.MeasurementName, searchResult)
		return nil
	}

	return parseAggregationResult(acc, aggregation.aggregationQueryList, searchResult)
}

func init() {
	inputs.Add("elasticsearch_query", func() telegraf.Input {
		return &ElasticsearchQuery{
			Timeout:             config.Duration(time.Second * 5),
			HealthCheckInterval: config.Duration(time.Second * 10),
		}
	})
}
