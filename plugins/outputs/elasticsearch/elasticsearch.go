package elasticsearch

import (
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"gopkg.in/olivere/elastic.v2"
	"os"
	"strings"
	"time"
)

type Elasticsearch struct {
	ServerHost    string
	IndexName     string
	EnableSniffer bool
	Client        *elastic.Client
	Version       string
}

var sampleConfig = `
  server_host = "http://10.10.10.10:9200" # required.
  index_name = "test" # required.
  enable_sniffer = false
`

func (a *Elasticsearch) Connect() error {
	if a.ServerHost == "" || a.IndexName == "" {
		return fmt.Errorf("FAILED server_host and index_name are required fields for Elasticsearch output")
	}

	client, err := elastic.NewClient(
		elastic.SetHealthcheck(true),
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
		return "",fmt.Errorf("FAILED to send Elasticsearch message to index %s : %s\n", a.IndexName, errMessage)
	}

	return put1.Id,nil

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
			EnableSniffer: false,
		}
	})
}
