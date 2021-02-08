package elasticsearch_query

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	elastic "gopkg.in/olivere/elastic.v5"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const description = `Queries Elasticseach`
const sampleConfig = `
  ## The full HTTP endpoint URL for your Elasticsearch instance
  ## Multiple urls can be specified as part of the same cluster,
  ## this means that only ONE of the urls will be written to each interval.
  urls = [ "http://node1.es.example.com:9200" ] # required.
  ## Elasticsearch client timeout, defaults to "5s" if not set.
  timeout = "5s"
  ## Set to true to ask Elasticsearch a list of all cluster nodes,
  ## thus it is not necessary to list all nodes in the urls config option
  enable_sniffer = false
  ## Set the interval to check if the Elasticsearch nodes are available
  ## Setting to "0s" will disable the health check (not recommended in production)
  health_check_interval = "10s"
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
  measurement_name = "measurement"
  index = "index-*"
  date_field = "@timestamp"
  ## Time window to query (eg. "1m" to query documents from last minute).
  ## Normally should be set to same as collection interval
  query_period = "1m"

  ## Optional parameters:
  ## Lucene query to filter results
  filter_query = "*"
  ## Fields to aggregate values (must be numeric fields)
  metric_fields = ["metric"]
  ## Aggregation function to use on the metric fields
  ## Valid values are: avg, sum, min, max, sum
  metric_function = "avg"
  ## Fields to be used as tags
  ## Must be non-analyzed fields, aggregations are performed per tag
  tags = ["field.keyword", "field2.keyword"]
  ## Set to true to not ignore documents when the tag(s) above are missing
  include_missing_tag = true
  ## String value of the tag when the tag does not exist
  ## Used when include_missing_tag is true
  missing_tag_value = "null"
`

type ElasticsearchQuery struct {
	URLs                []string `toml:"urls"`
	Username            string
	Password            string
	EnableSniffer       bool
	Tracelog            bool
	Timeout             internal.Duration
	HealthCheckInterval internal.Duration
	Aggregations        []Aggregation `toml:"aggregation"`
	tls.ClientConfig
	httpclient *http.Client
	ESClient   *elastic.Client
	acc        telegraf.Accumulator
}

type Aggregation struct {
	Index             string
	MeasurementName   string
	FilterQuery       string
	QueryPeriod       internal.Duration
	MetricFields      []string `toml:"metric_fields"`
	DateField         string
	Tags              []string `toml:"tags"`
	IncludeMissingTag bool
	MissingTagValue   string
	MetricFunction    string
}

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

func (e *ElasticsearchQuery) SampleConfig() string {
	return sampleConfig
}

func (e *ElasticsearchQuery) Description() string {
	return description
}

func (e *ElasticsearchQuery) init() error {
	if e.URLs == nil {
		return fmt.Errorf("Elasticsearch urls is not defined")
	}

	return e.connectToES()
}

func (e *ElasticsearchQuery) connectToES() error {

	var clientOptions []elastic.ClientOptionFunc

	if e.httpclient == nil {
		httpclient, err := e.createHttpClient()
		if err != nil {
			return err
		}
		e.httpclient = httpclient
	}

	clientOptions = append(clientOptions,
		elastic.SetHttpClient(e.httpclient),
		elastic.SetSniff(e.EnableSniffer),
		elastic.SetURL(e.URLs...),
		elastic.SetHealthcheckInterval(e.HealthCheckInterval.Duration),
	)

	if e.Username != "" {
		clientOptions = append(clientOptions,
			elastic.SetBasicAuth(e.Username, e.Password),
		)
	}

	if e.HealthCheckInterval.Duration == 0 {
		clientOptions = append(clientOptions,
			elastic.SetHealthcheck(false),
		)
	}

	if e.Tracelog {
		clientOptions = append(clientOptions,
			elastic.SetTraceLog(log.New(os.Stdout, "", log.LstdFlags)),
		)
	}

	client, err := elastic.NewClient(clientOptions...)
	if err != nil {
		return err
	}

	// check for ES version on first node
	esVersion, err := client.ElasticsearchVersion(e.URLs[0])

	if err != nil {
		return fmt.Errorf("Elasticsearch query version check failed: %s", err)
	}

	// quit if ES version is not supported
	i, err := strconv.Atoi(strings.Split(esVersion, ".")[0])
	if err != nil || i < 5 {
		return fmt.Errorf("Elasticsearch query: ES version not supported: %s", esVersion)
	}

	e.ESClient = client
	return nil
}

func (e *ElasticsearchQuery) Gather(acc telegraf.Accumulator) error {
	if err := e.init(); err != nil {
		return err
	}

	e.acc = acc

	var wg sync.WaitGroup

	for _, agg := range e.Aggregations {
		wg.Add(1)
		go func(agg Aggregation) {
			defer wg.Done()
			err := e.esAggregationQuery(agg)
			if err != nil {
				acc.AddError(fmt.Errorf("Elasticsearch query aggregation %s: %s ", agg.MeasurementName, err.Error()))
			}
		}(agg)
	}

	wg.Wait()
	return nil
}

func (e *ElasticsearchQuery) createHttpClient() (*http.Client, error) {
	tlsCfg, err := e.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}
	tr := &http.Transport{
		ResponseHeaderTimeout: e.Timeout.Duration,
		TLSClientConfig:       tlsCfg,
	}
	httpclient := &http.Client{
		Transport: tr,
		Timeout:   e.Timeout.Duration,
	}

	return httpclient, nil
}

func (e *ElasticsearchQuery) esAggregationQuery(aggregation Aggregation) error {

	ctx, cancel := context.WithTimeout(context.Background(), e.Timeout.Duration)
	defer cancel()

	if aggregation.DateField == "" {
		return fmt.Errorf("Field 'date_field' not set")
	}

	mapMetricFields, err := e.getMetricFields(ctx, aggregation)
	if err != nil {
		return err
	}

	aggregationQueryList, err := e.buildAggregationQuery(mapMetricFields, aggregation)
	if err != nil {
		return err
	}

	searchResult, err := e.runAggregationQuery(ctx, aggregation, aggregationQueryList)
	if err != nil {
		return err
	}

	if searchResult.Aggregations != nil {
		return e.parseAggregationResult(&aggregationQueryList, searchResult)
	}
	return e.parseSimpleResult(aggregation.MeasurementName, searchResult)
}

func init() {
	inputs.Add("elasticsearch_query", func() telegraf.Input {
		return &ElasticsearchQuery{
			Timeout:             internal.Duration{Duration: time.Second * 5},
			HealthCheckInterval: internal.Duration{Duration: time.Second * 10},
		}
	})
}
