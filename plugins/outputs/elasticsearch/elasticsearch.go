package elasticsearch

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	"crypto/sha256"

	"github.com/elastic/go-sysinfo"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/olivere/elastic/v7"
)

type Elasticsearch struct {
	URLs                []string `toml:"urls"`
	IndexName           string
	IsDataStream        bool `toml:"datastream"`
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
	ForceDocumentId     bool
	MajorReleaseNumber  int
	ElasticCommonSchema bool `toml:"use_ecs"`
	tls.ClientConfig

	Client *elastic.Client
}

var sampleConfig = `
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
  ## HTTP basic authentication details
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
  ## If set to true a unique ID hash will be sent as sha256(concat(timestamp,measurement,series-hash)) string
  ## it will enable data resend and update metric points avoiding duplicated metrics with diferent id's
  force_document_id = false

  ## Enable Elastic Common Schema format
  ## Set to true if you want to format output document with ECS standard
  use_ecs = false
  
  ## Use datastream
  ## Enable to save documents as datastream
  datastream = false
`

const telegrafTemplate = `
{
	{{ if (lt .Version 6) }}
	"template": "{{.TemplatePattern}}",
	{{ else }}
	"index_patterns" : [ "{{.TemplatePattern}}" ],
	{{ end }}
	"settings": {
		"index": {
			"refresh_interval": "10s",
			"mapping.total_fields.limit": 5000,
			"auto_expand_replicas" : "0-1",
			"codec" : "best_compression"
		}
	},
	"mappings" : {
		{{ if (lt .Version 7) }}
		"metrics" : {
			{{ if (lt .Version 6) }}
			"_all": { "enabled": false },
			{{ end }}
		{{ end }}
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
		{{ if (lt .Version 7) }}
		}
		{{ end }}
	}
}`

type templatePart struct {
	TemplatePattern string
	Version         int
}

func (a *Elasticsearch) Connect() error {
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
	majorReleaseNumber, err := strconv.Atoi(strings.Split(esVersion, ".")[0])
	if err != nil || majorReleaseNumber < 5 {
		return fmt.Errorf("Elasticsearch version not supported: %s", esVersion)
	}

	log.Println("I! Elasticsearch version: " + esVersion)

	a.Client = client
	a.MajorReleaseNumber = majorReleaseNumber

	if a.ManageTemplate {
		err := a.manageTemplate(ctx)
		if err != nil {
			return err
		}
	}

	a.IndexName, a.TagKeys = a.GetTagKeys(a.IndexName)

	return nil
}

// GetPointID generates a unique ID for a Metric Point
func GetPointID(m telegraf.Metric) string {

	var buffer bytes.Buffer
	//Timestamp(ns),measurement name and Series Hash for compute the final SHA256 based hash ID

	buffer.WriteString(strconv.FormatInt(m.Time().Local().UnixNano(), 10))
	buffer.WriteString(m.Name())
	buffer.WriteString(strconv.FormatUint(m.HashID(), 10))

	return fmt.Sprintf("%x", sha256.Sum256(buffer.Bytes()))
}

func (a *Elasticsearch) Write(metrics []telegraf.Metric) error {
	var ecs_agent, ecs_host, ecs_event map[string]interface{}
	if len(metrics) == 0 {
		return nil
	}

	bulkRequest := a.Client.Bulk()

	if a.ElasticCommonSchema {
		ecs_host = getEcsHostInfo()
		ecs_agent = getEcsAgentInfo()
	}

	for _, metric := range metrics {
		var name = metric.Name()

		if a.ElasticCommonSchema {
			ecs_event = getEcsEventInfo(name)
		}

		// index name has to be re-evaluated each time for telegraf
		// to send the metric to the correct time-based index
		indexName := a.GetIndexName(a.IndexName, metric.Time(), a.TagKeys, metric.Tags(), metric.Name())

		m := make(map[string]interface{})

		m["@timestamp"] = metric.Time()
		if !a.ElasticCommonSchema {
			m["measurement_name"] = name
			m["tag"] = metric.Tags()
			m[name] = metric.Fields()
		} else {
			t := make(map[string]interface{})
			t["tags"] = metric.Tags()
			t["metrics"] = map[string]interface{}{name: metric.Fields()}
			m["telegraf"] = t
			m["metricset"] = map[string]interface{}{"name": name}
			m["event"] = ecs_event
			m["host"] = ecs_host
			m["agent"] = ecs_agent
			m["ecs"] = map[string]interface{}{"version": "1.7.0"}
		}

		br := elastic.NewBulkIndexRequest().Index(indexName).Doc(m)

		if a.IsDataStream {
			br.OpType("create")
		}

		if a.ForceDocumentId {
			id := GetPointID(metric)
			br.Id(id)
		}

		if a.MajorReleaseNumber <= 6 {
			br.Type("metrics")
		}

		bulkRequest.Add(br)

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
			break
		}
		return fmt.Errorf("W! Elasticsearch failed to index %d metrics", len(res.Failed()))
	}

	return nil

}

func (a *Elasticsearch) manageTemplate(ctx context.Context) error {
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
		tp := templatePart{
			TemplatePattern: templatePattern + "*",
			Version:         a.MajorReleaseNumber,
		}

		t := template.Must(template.New("template").Parse(telegrafTemplate))
		var tmpl bytes.Buffer

		t.Execute(&tmpl, tp)
		_, errCreateTemplate := a.Client.IndexPutTemplate(a.TemplateName).BodyString(tmpl.String()).Do(ctx)

		if errCreateTemplate != nil {
			return fmt.Errorf("Elasticsearch failed to create index template %s : %s", a.TemplateName, errCreateTemplate)
		}

		log.Printf("D! Elasticsearch template %s created or updated\n", a.TemplateName)

	} else {

		log.Println("D! Found existing Elasticsearch template. Skipping template management")

	}
	return nil
}

func (a *Elasticsearch) GetTagKeys(indexName string) (string, []string) {

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

func (a *Elasticsearch) GetIndexName(indexName string, eventTime time.Time, tagKeys []string, metricTags map[string]string, metricName string) string {
	if strings.Contains(indexName, "%") {
		var dateReplacer = strings.NewReplacer(
			"%Y", eventTime.UTC().Format("2006"),
			"%y", eventTime.UTC().Format("06"),
			"%m", eventTime.UTC().Format("01"),
			"%d", eventTime.UTC().Format("02"),
			"%H", eventTime.UTC().Format("15"),
			"%V", getISOWeek(eventTime.UTC()),
		)

		indexName = dateReplacer.Replace(indexName)
	}

	if strings.Contains(indexName, "metric_name") {
		var metricNameReplacer = strings.NewReplacer(
			"metric_name", metricName,
		)
		indexName = metricNameReplacer.Replace(indexName)
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

func getEcsEventInfo(metric_name string) map[string]interface{} {
	i := make(map[string]interface{})
	i["dataset"] = metric_name + ".metrics"
	i["module"] = metric_name

	return i
}

func getEcsHostInfo() map[string]interface{} {
	var ipList []string
	var hwList []string

	h, _ := sysinfo.Host()
	info := h.Info()
	os := make(map[string]interface{})
	i := make(map[string]interface{})
	os["platform"] = info.OS.Platform
	os["version"] = info.OS.Version
	os["family"] = info.OS.Family
	os["name"] = info.OS.Name
	os["kernel"] = info.KernelVersion

	if info.OS.Codename != "" {
		os["codename"] = info.OS.Codename
	}
	if info.OS.Build != "" {
		os["build"] = info.OS.Build
	}
	i["name"] = info.Hostname
	i["architecture"] = info.Architecture
	i["os"] = os
	if info.UniqueID != "" {
		i["id"] = info.UniqueID
	}
	if info.Containerized != nil {
		i["containerized"] = *info.Containerized
	}

	ifaces, _ := net.Interfaces()
	for _, in := range ifaces {
		if in.Flags&net.FlagLoopback == net.FlagLoopback {
			continue
		}
		if in.HardwareAddr != nil {
			hwList = append(hwList, in.HardwareAddr.String())
		}
		addrs, err := in.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			default:
				continue
			}
			ipList = append(ipList, ip.String())
		}
	}
	i["ip"] = ipList
	i["mac"] = hwList

	return i
}

func getEcsAgentInfo() map[string]interface{} {
	hostname, _ := os.Hostname()
	i := make(map[string]interface{})
	i["type"] = "telegraf"
	i["version"] = internal.Version()
	i["host"] = hostname
	i["name"] = hostname
	i["build"] = map[string]interface{}{"original": internal.ProductToken()}
	return i
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
		return &Elasticsearch{
			Timeout:             internal.Duration{Duration: time.Second * 5},
			HealthCheckInterval: internal.Duration{Duration: time.Second * 10},
		}
	})
}
