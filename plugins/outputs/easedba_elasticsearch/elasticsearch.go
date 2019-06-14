package easedbaelasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
	elastic "gopkg.in/olivere/elastic.v5"
)

var (
	// keys: hostname, localIP
	globalTagsPool = make(map[string]string)
)

func init() {
	hostName, _ := os.Hostname()
	globalTagsPool["hostname"] = hostName

	addrs, _ := net.InterfaceAddrs()
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			ipstr := ipnet.IP.String()
			// FIXME: Not accurate.
			if strings.HasPrefix(ipstr, "10.") ||
				strings.HasPrefix(ipstr, "192.") ||
				strings.HasPrefix(ipstr, "172.") {
				globalTagsPool["localIP"] = ipstr
				return
			}
		}
	}
}

type EaseDBAElasticsearch struct {
	Debug               bool
	URLs                []string `toml:"urls"`
	IndexName           string
	Type                string
	DefaultTagValue     string
	TagKeys             []string
	Username            string
	Password            string
	EnableSniffer       bool
	Timeout             internal.Duration
	HealthCheckInterval internal.Duration
	ManageTemplate      bool
	TemplateName        string
	OverwriteTemplate   bool
	tls.ClientConfig

	Filters MetricFilters

	Client *elastic.Client
}

type MetricFilters struct {
	once sync.Once

	GlobalTags []string
	//  	    metricName
	Metrics map[string]*MetricFilter
}

type MetricFilter struct {
	Fields     []string
	Tags       []string
	CustomTags map[string]string

	allTags   map[string]struct{}
	allFields map[string]struct{}
}

func (mf MetricFilters) init() {
	for _, filter := range mf.Metrics {
		allTags := make(map[string]struct{})
		for _, tagKey := range mf.GlobalTags {
			allTags[tagKey] = struct{}{}
		}
		for _, tagKey := range filter.Tags {
			allTags[tagKey] = struct{}{}
		}

		allFields := make(map[string]struct{})
		for _, field := range filter.Fields {
			allFields[field] = struct{}{}
		}

		filter.allTags, filter.allFields = allTags, allFields
	}
}

func (mf MetricFilters) filter(metrics []telegraf.Metric) []telegraf.Metric {
	mf.once.Do(mf.init)

	newMetrics := make([]telegraf.Metric, 0)
	for _, metric := range metrics {
		filter, exists := mf.Metrics[metric.Name()]
		if !exists {
			continue
		}
		newMetrics = append(newMetrics, metric)

		var tagKeysToDelete []string
		for _, tag := range metric.TagList() {
			if _, exists := filter.allTags[tag.Key]; !exists {
				tagKeysToDelete = append(tagKeysToDelete, tag.Key)
			}
		}
		for _, tagKey := range tagKeysToDelete {
			metric.RemoveTag(tagKey)
		}

		for tagKey := range filter.allTags {
			if metric.HasTag(tagKey) {
				continue
			}
			if tagValue, exists := globalTagsPool[tagKey]; exists {
				metric.AddTag(tagKey, tagValue)
			}
		}

		for k, v := range filter.CustomTags {
			metric.AddTag(k, v)
		}

		var fieldKeysToDelete []string
		for _, field := range metric.FieldList() {
			if _, exists := filter.allFields[field.Key]; exists {
				continue
			}
			fieldKeysToDelete = append(fieldKeysToDelete, field.Key)
		}
		for _, fieldKey := range fieldKeysToDelete {
			metric.RemoveField(fieldKey)
		}
	}

	return newMetrics
}

func (a *EaseDBAElasticsearch) doLog(metric map[string]interface{}) {
	if a.Debug {
		buff, err := json.Marshal(metric)
		if err != nil {
			log.Printf("marshal %#v failed: %v", metric, err)
		}
		log.Printf("%s\n", buff)
	}

	for _, value := range metric {
		v, ok := value.(uint64)
		if !ok {
			continue
		}
		if v > math.MaxInt64 {
			log.Printf("%+v got value larger than %d", metric, v)
		}
	}
}

var sampleConfig = `
  ## The debug flag specify whether to log print metrics.
  debug = false
  ## The full HTTP endpoint URL for your Elasticsearch instance
  ## Multiple urls can be specified as part of the same cluster,
  ## this means that only ONE of the urls will be written to each interval.
  urls = [ "http://node1.es.example.com:9200" ] # required.
  ## Elasticsearch client timeout, defaults to "5s" if not set.
  timeout = "5s"
  ## Set to true to ask Elasticsearch a list of all cluster nodes,
  ## thus it is not necessary to list all nodes in the urls config option.
  enable_sniffer = false
  ## Set the interval to check if the Elasticsearch nodes are available
  ## Setting to "0s" will disable the health check (not recommended in production)
  health_check_interval = "10s"
  ## HTTP basic authentication details (eg. when using Shield)
  # username = "telegraf"
  # password = "mypassword"

  ## Index Config
  ## The target index for metrics (Elasticsearch will create if it not exists).
  ## You can use the date specifiers below to create indexes per time frame.
  ## The metric timestamp will be used to decide the destination index name
  # %Y - year (2016)
  # %y - last two digits of year (00..99)
  # %m - month (01..12)
  # %d - day of month (e.g., 01)
  # %H - hour (00..23)
  # %V - week of the year (ISO week) (01..53)
  # %t - name of metric (cpu,mem,net,disk..)
  ## Additionally, you can specify a tag name using the notation {{tag_name}}
  ## which will be used as part of the index name. If the tag does not exist,
  ## the default tag value will be used.
  # index_name = "telegraf-{{host}}-%Y.%m.%d"
  # default_tag_value = "none"
  index_name = "telegraf-%Y.%m.%d" # required.

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Template Config
  ## Set to true if you want telegraf to manage its index template.
  ## If enabled it will create a recommended index template for telegraf indexes
  manage_template = true
  ## The template name used for telegraf indexes
  template_name = "telegraf"
  ## Set to true if you want telegraf to overwrite an existing template
  overwrite_template = false
`

func (a *EaseDBAElasticsearch) Connect() error {
	if a.URLs == nil || a.IndexName == "" {
		return fmt.Errorf("Elasticsearch urls or index_name is not defined")
	}

	ctx, cancel := context.WithTimeout(context.Background(), a.Timeout.Duration)
	defer cancel()

	var clientOptions []elastic.ClientOptionFunc

	tlsCfg, err := a.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}
	tr := &http.Transport{
		TLSClientConfig: tlsCfg,
	}

	httpclient := &http.Client{
		Transport: tr,
		Timeout:   a.Timeout.Duration,
	}

	clientOptions = append(clientOptions,
		elastic.SetHttpClient(httpclient),
		elastic.SetSniff(a.EnableSniffer),
		elastic.SetURL(a.URLs...),
		elastic.SetHealthcheckInterval(a.HealthCheckInterval.Duration),
	)

	if a.Username != "" && a.Password != "" {
		clientOptions = append(clientOptions,
			elastic.SetBasicAuth(a.Username, a.Password),
		)
	}

	if a.HealthCheckInterval.Duration == 0 {
		clientOptions = append(clientOptions,
			elastic.SetHealthcheck(false),
		)
		log.Printf("D! Elasticsearch output: disabling health check")
	}

	client, err := elastic.NewClient(clientOptions...)

	if err != nil {
		return err
	}

	// check for ES version on first node
	esVersion, err := client.ElasticsearchVersion(a.URLs[0])

	if err != nil {
		return fmt.Errorf("Elasticsearch version check failed: %s", err)
	}

	// quit if ES version is not supported
	i, err := strconv.Atoi(strings.Split(esVersion, ".")[0])
	if err != nil || i < 5 {
		return fmt.Errorf("Elasticsearch version not supported: %s", esVersion)
	}

	log.Println("I! Elasticsearch version: " + esVersion)

	a.Client = client

	if a.ManageTemplate {
		err := a.manageTemplate(ctx)
		if err != nil {
			return err
		}
	}

	a.IndexName, a.TagKeys = a.GetTagKeys(a.IndexName)

	return nil
}

func (a *EaseDBAElasticsearch) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	metrics = a.Filters.filter(metrics)
	if len(metrics) == 0 {
		return nil
	}

	bulkRequest := a.Client.Bulk()

	for _, metric := range metrics {
		var name = metric.Name()

		// index name has to be re-evaluated each time for telegraf
		// to send the metric to the correct time-based index
		indexName := a.GetIndexName(a.IndexName, name, metric.Time(), a.TagKeys, metric.Tags())

		m := make(map[string]interface{})

		m["timestamp"] = metric.Time()
		m["measurement_name"] = name
		m["tags"] = metric.Tags()
		m["fields"] = metric.Fields()

		a.doLog(m)

		bulkRequest.Add(elastic.NewBulkIndexRequest().
			Index(indexName).
			Type(a.Type).
			Doc(m))

	}

	ctx, cancel := context.WithTimeout(context.Background(), a.Timeout.Duration)
	defer cancel()

	res, err := bulkRequest.Do(ctx)

	if err != nil {
		return fmt.Errorf("Error sending bulk request to Elasticsearch: %s", err)
	}

	if res.Errors {
		for id, err := range res.Failed() {
			log.Printf("E! Elasticsearch indexing failure, id: %d, error: %s, caused by: %s, %s", id, err.Error.Reason, err.Error.CausedBy["reason"], err.Error.CausedBy["type"])
		}
		return fmt.Errorf("W! Elasticsearch failed to index %d metrics", len(res.Failed()))
	}

	return nil

}

func (a *EaseDBAElasticsearch) manageTemplate(ctx context.Context) error {
	if a.TemplateName == "" {
		return fmt.Errorf("Elasticsearch template_name configuration not defined")
	}

	templateExists, errExists := a.Client.IndexTemplateExists(a.TemplateName).Do(ctx)

	if errExists != nil {
		return fmt.Errorf("Elasticsearch template check failed, template name: %s, error: %s", a.TemplateName, errExists)
	}

	templatePattern := a.IndexName

	if strings.Contains(templatePattern, "%") {
		templatePattern = templatePattern[0:strings.Index(templatePattern, "%")]
	}

	if strings.Contains(templatePattern, "{{") {
		templatePattern = templatePattern[0:strings.Index(templatePattern, "{{")]
	}

	if templatePattern == "" {
		return fmt.Errorf("Template cannot be created for dynamic index names without an index prefix")
	}

	if (a.OverwriteTemplate) || (!templateExists) || (templatePattern != "") {
		// Create or update the template
		tmpl := fmt.Sprintf(`
			{
				"template":"%s",
				"settings": {
					"index": {
						"refresh_interval": "10s",
						"mapping.total_fields.limit": 5000
					}
				},
				"mappings" : {
					"_default_" : {
						"_all": { "enabled": false	  },
						"properties" : {
							"@timestamp" : { "type" : "date" },
							"measurement_name" : { "type" : "keyword" }
						},
						"dynamic_templates": [
							{
								"tags": {
									"match_mapping_type": "string",
									"path_match": "tag.*",
									"mapping": {
										"ignore_above": 512,
										"type": "keyword"
									}
								}
							},
							{
								"metrics_long": {
									"match_mapping_type": "long",
									"mapping": {
										"type": "float",
										"index": false
									}
								}
							},
							{
								"metrics_double": {
									"match_mapping_type": "double",
									"mapping": {
										"type": "float",
										"index": false
									}
								}
							},
							{
								"text_fields": {
									"match": "*",
									"mapping": {
										"norms": false
									}
								}
							}
						]
					}
				}
			}`, templatePattern+"*")
		_, errCreateTemplate := a.Client.IndexPutTemplate(a.TemplateName).BodyString(tmpl).Do(ctx)

		if errCreateTemplate != nil {
			return fmt.Errorf("Elasticsearch failed to create index template %s : %s", a.TemplateName, errCreateTemplate)
		}

		log.Printf("D! Elasticsearch template %s created or updated\n", a.TemplateName)

	} else {

		log.Println("D! Found existing Elasticsearch template. Skipping template management")

	}
	return nil
}

func (a *EaseDBAElasticsearch) GetTagKeys(indexName string) (string, []string) {

	tagKeys := []string{}
	startTag := strings.Index(indexName, "{{")

	for startTag >= 0 {
		endTag := strings.Index(indexName, "}}")

		if endTag < 0 {
			startTag = -1

		} else {
			tagName := indexName[startTag+2 : endTag]

			var tagReplacer = strings.NewReplacer(
				"{{"+tagName+"}}", "%s",
			)

			indexName = tagReplacer.Replace(indexName)
			tagKeys = append(tagKeys, (strings.TrimSpace(tagName)))

			startTag = strings.Index(indexName, "{{")
		}
	}

	return indexName, tagKeys
}

func (a *EaseDBAElasticsearch) GetIndexName(indexName, metricName string, eventTime time.Time, tagKeys []string, metricTags map[string]string) string {
	if strings.Contains(indexName, "%") {
		var dateReplacer = strings.NewReplacer(
			"%Y", eventTime.UTC().Format("2006"),
			"%y", eventTime.UTC().Format("06"),
			"%m", eventTime.UTC().Format("01"),
			"%d", eventTime.UTC().Format("02"),
			"%H", eventTime.UTC().Format("15"),
			"%V", getISOWeek(eventTime.UTC()),
			"%t", metricName,
		)

		indexName = dateReplacer.Replace(indexName)
	}

	tagValues := []interface{}{}

	for _, key := range tagKeys {
		if value, ok := metricTags[key]; ok {
			tagValues = append(tagValues, value)
		} else {
			log.Printf("D! Tag '%s' not found, using '%s' on index name instead\n", key, a.DefaultTagValue)
			tagValues = append(tagValues, a.DefaultTagValue)
		}
	}

	return fmt.Sprintf(indexName, tagValues...)

}

func getISOWeek(eventTime time.Time) string {
	_, week := eventTime.ISOWeek()
	return strconv.Itoa(week)
}

func (a *EaseDBAElasticsearch) SampleConfig() string {
	return sampleConfig
}

func (a *EaseDBAElasticsearch) Description() string {
	return "Configuration for Elasticsearch to send metrics to."
}

func (a *EaseDBAElasticsearch) Close() error {
	a.Client = nil
	return nil
}

func init() {
	outputs.Add("easedba_elasticsearch", func() telegraf.Output {
		return &EaseDBAElasticsearch{
			Type:                "_doc",
			Timeout:             internal.Duration{Duration: time.Second * 5},
			HealthCheckInterval: internal.Duration{Duration: time.Second * 10},
		}
	})
}
