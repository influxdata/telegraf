package elasticsearch

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"gopkg.in/olivere/elastic.v5"
)

type Elasticsearch struct {
	URLs                []string `toml:"urls"`
	IndexName           string
	Username            string
	Password            string
	EnableSniffer       bool
	HealthCheckInterval int
	ManageTemplate      bool
	TemplateName        string
	OverwriteTemplate   bool
	Client              *elastic.Client
}

var sampleConfig = `
  ## The full HTTP endpoint URL for your Elasticsearch instance
  ## Multiple urls can be specified as part of the same cluster,
  ## this means that only ONE of the urls will be written to each interval.
  urls = [ "http://node1.es.example.com:9200" ] # required.
  ## Set to true to ask Elasticsearch a list of all cluster nodes,
  ## thus it is not necessary to list all nodes in the urls config option
  enable_sniffer = true
  ## Set the interval to check if the nodes are available, in seconds
  ## Setting to 0 will disable the health check (not recommended in production)
  health_check_interval = 10
  ## HTTP basic authentication details (eg. when using Shield)
  # username = "telegraf"
  # password = "mypassword"

  # Index Config
  ## The target index for metrics (Elasticsearch will create if it not exists).
  ## You can use the date specifiers below to create indexes per time frame.
  ## The metric timestamp will be used to decide the destination index name
  # %Y - year (2016)
  # %y - last two digits of year (00..99)
  # %m - month (01..12)
  # %d - day of month (e.g., 01)
  # %H - hour (00..23)
  index_name = "telegraf-%Y.%m.%d" # required.

  ## Template Config
  ## Set to true if you want telegraf to manage its index template.
  ## If enabled it will create a recommended index template for telegraf indexes
  manage_template = true
  ## The template name used for telegraf indexes
  template_name = "telegraf"
  ## Set to true if you want to overwrite an existing template
  overwrite_template = false
`

func (a *Elasticsearch) Connect() error {
	if a.URLs == nil || a.IndexName == "" {
		return fmt.Errorf("Elasticsearch urls or index_name is not defined")
	}

	ctx := context.Background()

	var clientOptions []elastic.ClientOptionFunc

	clientOptions = append(clientOptions,
		elastic.SetSniff(a.EnableSniffer),
		elastic.SetURL(a.URLs...),
	)

	if a.Username != "" && a.Password != "" {
		clientOptions = append(clientOptions,
			elastic.SetBasicAuth(a.Username, a.Password),
		)
	}

	if a.HealthCheckInterval > 0 {
		clientOptions = append(clientOptions,
			elastic.SetHealthcheckInterval(time.Duration(a.HealthCheckInterval)*time.Second),
		)
	}

	client, err := elastic.NewClient(clientOptions...)

	if err != nil {
		return fmt.Errorf("Elasticsearch connection failed: %s", err)
	}

	// check for version on first node
	esVersion, err := client.ElasticsearchVersion(a.URLs[0])

	if err != nil {
		return fmt.Errorf("Elasticsearch version check failed: %s", err)
	}

	// warn about ES version
	if i, err := strconv.Atoi(strings.Split(esVersion, ".")[0]); err == nil {
		if i < 5 {
			log.Println("W! Elasticsearch version not supported: " + esVersion)
		} else {
			log.Println("I! Elasticsearch version: " + esVersion)
		}
	}

	a.Client = client

	if a.ManageTemplate {
		if a.TemplateName == "" {
			return fmt.Errorf("Elasticsearch template_name configuration not defined")
		}

		templateExists, errExists := a.Client.IndexTemplateExists(a.TemplateName).Do(ctx)

		if errExists != nil {
			return fmt.Errorf("Elasticsearch template check failed, template name: %s, error: %s", a.TemplateName, errExists)
		}

		if (a.OverwriteTemplate) || (!templateExists) {
			// Create or update the template
			tmpl := fmt.Sprintf(`
			{ "template":"%s*",
				"mappings" : {
					"_default_" : {
						"_all": { "enabled": false	},
						"properties" : {
							"@timestamp" : { "type" : "date" },
							"input_plugin" : { "type" : "keyword" }
						},
						"dynamic_templates": [{
							"tag": {
								"path_match": "tag.*",
								"mapping": {
									"ignore_above": 512,
									"type": "keyword"
								},
								"match_mapping_type": "string"
							}
						}]
					}
				}
			}`, a.IndexName[0:strings.Index(a.IndexName, "%")])

			_, errCreateTemplate := a.Client.IndexPutTemplate(a.TemplateName).BodyString(tmpl).Do(ctx)

			if errCreateTemplate != nil {
				return fmt.Errorf("Elasticsearch failed to create index template %s : %s", a.TemplateName, errCreateTemplate)
			}

			log.Printf("D! Elasticsearch template %s created or updated\n", a.TemplateName)

		} else {

			log.Println("D! Found existing Elasticsearch template. Skipping template management")

		}
	}

	return nil
}

func (a *Elasticsearch) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	ctx := context.Background()
	bulkRequest := a.Client.Bulk()

	for _, metric := range metrics {
		var name = metric.Name()

		// index name has to be re-evaluated each time for telegraf
		// to send the metric to the correct time-based index
		indexName := a.GetIndexName(a.IndexName, metric.Time())

		m := make(map[string]interface{})
		mName := make(map[string]interface{})
		mTag := make(map[string]interface{})

		m["@timestamp"] = metric.Time()
		m["input_plugin"] = name

		for key, value := range metric.Tags() {
			mTag[key] = value
		}

		for key, value := range metric.Fields() {
			mName[key] = value
		}

		m["tag"] = mTag
		m[name] = mName

		bulkRequest.Add(elastic.NewBulkIndexRequest().
			Index(indexName).
			Type("metrics").
			Doc(m))

	}

	_, err := bulkRequest.Do(ctx)

	if err != nil {
		return fmt.Errorf("Error sending bulk request to Elasticsearch: %s", err)
	}

	return nil

}

func (a *Elasticsearch) GetIndexName(indexName string, eventTime time.Time) string {
	if strings.Contains(indexName, "%") {
		var dateReplacer = strings.NewReplacer(
			"%Y", eventTime.Format("2006"),
			"%y", eventTime.Format("06"),
			"%m", eventTime.Format("01"),
			"%d", eventTime.Format("02"),
			"%H", eventTime.Format("15"),
		)

		indexName = dateReplacer.Replace(indexName)
	}

	return indexName

}

func (a *Elasticsearch) SampleConfig() string {
	return sampleConfig
}

func (a *Elasticsearch) Description() string {
	return "Configuration for Elasticsearch to send metrics to."
}

func (a *Elasticsearch) Close() error {
	a.Client = nil
	return nil
}

func init() {
	outputs.Add("elasticsearch", func() telegraf.Output {
		return &Elasticsearch{}
	})
}
