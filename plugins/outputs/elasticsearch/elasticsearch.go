package elasticsearch

import (
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"gopkg.in/olivere/elastic.v3"
	"os"
	"strconv"
	"strings"
	"time"
)

type Elasticsearch struct {
	ServerHost       string
	IndexName        string
	EnableSniffer    bool
	HealthCheck      bool
	NumberOfShards   int
	NumberOfReplicas int
	Client           *elastic.Client
	Version          string
}

var sampleConfig = `
  server_host = "http://10.10.10.10:9200" # required.
  index_name = "test" # required.
  # regex allowed on index_name: 
  # %Y - year (2016)
  # %y - last two digits of year (00..99)
  # %m - month (01..12)
  # %d - day of month (e.g., 01)
  # %H - hour (00..23)
  enable_sniffer = false
  health_check = false
  number_of_shards = 1
  number_of_replicas = 0
`

func (a *Elasticsearch) Connect() error {
	if a.ServerHost == "" || a.IndexName == "" {
		return fmt.Errorf("FAILED server_host and index_name are required fields for Elasticsearch output")
	}

	// Check if index's name has a prefix
	if strings.HasPrefix(a.IndexName, "%") {
		return fmt.Errorf("FAILED  Elasticsearch index's name must start with a prefix. \n")
	}

	client, err := elastic.NewClient(
		elastic.SetHealthcheck(a.HealthCheck),
		elastic.SetSniff(a.EnableSniffer),
		elastic.SetHealthcheckInterval(30*time.Second),
		elastic.SetURL(a.ServerHost),
	)

	if err != nil {
		return fmt.Errorf("FAILED to connect to Elasticsearch host %s : %s\n", a.ServerHost, err)
	}

	a.Client = client

	version, errVersion := a.Client.ElasticsearchVersion(a.ServerHost)

	if errVersion != nil {
		return fmt.Errorf("FAILED to get Elasticsearch version : %s\n", errVersion)
	}

	a.Version = version

	templateName := a.IndexName

	if strings.Contains(a.IndexName, "%") {
		// Template's name its Index's name without date patterns
		templateName = a.IndexName[0:strings.Index(a.IndexName, "%")]

		year := strconv.Itoa(time.Now().Year())
		a.IndexName = strings.Replace(a.IndexName, "%Y", year, -1)
		a.IndexName = strings.Replace(a.IndexName, "%y", year[len(year)-2:], -1)
		a.IndexName = strings.Replace(a.IndexName, "%m", strconv.Itoa(int(time.Now().Month())), -1)
		a.IndexName = strings.Replace(a.IndexName, "%d", strconv.Itoa(time.Now().Day()), -1)
		a.IndexName = strings.Replace(a.IndexName, "%H", strconv.Itoa(time.Now().Hour()), -1)
	}

	exists, errExists := a.Client.IndexExists(a.IndexName).Do()

	if errExists != nil {
		return fmt.Errorf("FAILED to check if Elasticsearch index %s exists : %s\n", a.IndexName, errExists)
	}

	if !exists {
		// First create a template for the new index

		// The [string] type is removed in 5.0
		typeHostandUnknow := "text"

		if strings.HasPrefix(a.Version, "2.") {
			typeHostandUnknow = "string"
		}

		tmpl := fmt.Sprintf(`{
			"template":"%s*",
    			"settings" : {
        			"number_of_shards" : %s,
				"number_of_replicas" : %s
    			},
    			"mappings" : {
        			"_default_" : {
					"_all": {
        					"enabled": false
      					},
            				"properties" : {
                				"created" : { "type" : "date" },
						"host":{"type":"%s"}
            				},
	    				"dynamic_templates": [
                			{ "unknowfields": {
                      				"match": "*", 
                      				"match_mapping_type": "unknow",
                      				"mapping": {
                          				"type":"%s"
                      				}
                			}}
            				]
        			}
    			}
		}`, templateName, strconv.Itoa(a.NumberOfShards), strconv.Itoa(a.NumberOfReplicas), typeHostandUnknow, typeHostandUnknow)

		_, errCreateTemplate := a.Client.IndexPutTemplate(templateName).BodyString(tmpl).Do()

		if errCreateTemplate != nil {
			return fmt.Errorf("FAILED to create Elasticsearch index template %s : %s\n", templateName, errCreateTemplate)
		}

		// Now create the new index
		_, errCreateIndex := a.Client.CreateIndex(a.IndexName).Do()

		if errCreateIndex != nil {
			return fmt.Errorf("FAILED to create Elasticsearch index %s : %s\n", a.IndexName, errCreateIndex)
		}

	}

	return nil
}

func (a *Elasticsearch) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	for _, metric := range metrics {

		m := make(map[string]interface{})
		m["created"] = metric.Time()

		if host, ok := metric.Tags()["host"]; ok {
			m["host"] = host
		} else {
			host, err := os.Hostname()
			if err != nil {
				return fmt.Errorf("FAILED Obtaining host to Elasticsearch output : %s\n", err)
			}
			m["host"] = host
		}

		// Elasticsearch 2.x does not support this dots-to-object transformation
		// and so dots in field names are not allowed in versions 2.X.
		// In this case, dots will be replaced with "_"

		for key, value := range metric.Tags() {
			if key != "host" {
				if strings.HasPrefix(a.Version, "2.") {
					m[strings.Replace(key, ".", "_", -1)] = value
				} else {
					m[key] = value
				}
			}
		}

		for key, value := range metric.Fields() {
			if strings.HasPrefix(a.Version, "2.") {
				m[strings.Replace(key, ".", "_", -1)] = value
			} else {
				m[key] = value
			}
		}

		_, errMessage := a.Client.Index().
			Index(a.IndexName).
			Type(metric.Name()).
			BodyJson(m).
			Do()

		if errMessage != nil {
			return fmt.Errorf("FAILED to send Elasticsearch message to index %s : %s\n", a.IndexName, errMessage)
		}

	}

	return nil
}

func (a *Elasticsearch) WriteOneMessage(metric telegraf.Metric) (string, error) {

	m := make(map[string]interface{})
	m["created"] = metric.Time()

	if host, ok := metric.Tags()["host"]; ok {
		m["host"] = host
	}

	// Elasticsearch 2.x does not support this dots-to-object transformation
	// and so dots in field names are not allowed in versions 2.X.
	// In this case, dots will be replaced with "_"

	for key, value := range metric.Tags() {
		if key != "host" {
			if strings.HasPrefix(a.Version, "2.") {
				m[strings.Replace(key, ".", "_", -1)] = value
			} else {
				m[key] = value
			}
		}
	}

	for key, value := range metric.Fields() {
		if strings.HasPrefix(a.Version, "2.") {
			m[strings.Replace(key, ".", "_", -1)] = value
		} else {
			m[key] = value
		}
	}

	put1, errMessage := a.Client.Index().
		Index(a.IndexName).
		Type(metric.Name()).
		BodyJson(m).
		Do()

	if errMessage != nil {
		return "", fmt.Errorf("FAILED to send Elasticsearch message to index %s : %s\n", a.IndexName, errMessage)
	}

	return put1.Id, nil

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
			EnableSniffer:    false,
			HealthCheck:      false,
			NumberOfShards:   1,
			NumberOfReplicas: 0,
		}
	})
}
