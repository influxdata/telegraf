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
	Separator     string
	Client        *elastic.Client
	Version       string
}

var sampleConfig = `
  server_host = "http://10.10.10.10:19200" # required.
  index_name = "test" # required.
  enable_sniffer = false
  delimiter = "_"
`

type TimeSeries struct {
	Series []*Metric `json:"series"`
}

type Metric struct {
	Metric string   `json:"metric"`
	Points [1]Point `json:"metrics"`
}

type Point [2]float64

func (a *Elasticsearch) Connect() error {
	if a.ServerHost == "" || a.IndexName == "" {
		return fmt.Errorf("server_host and index_name are required fields for elasticsearch output")
	}

	client, err := elastic.NewClient(
		elastic.SetHealthcheck(true),
		elastic.SetSniff(a.EnableSniffer),
		elastic.SetHealthcheckInterval(30*time.Second),
		elastic.SetURL(a.ServerHost),
	)

	if err != nil {
		return fmt.Errorf("FAILED to connect to elasticsearch host %s : %s\n", a.ServerHost, err)
	}

	a.Client = client

	version, errVersion := a.Client.ElasticsearchVersion(a.ServerHost)

	if errVersion != nil {
		return fmt.Errorf("FAILED to get elasticsearch version : %s\n", errVersion)
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
		m["created"] = time.Now()

		if host, ok := metric.Tags()["host"]; ok {
			m["host"] = host
		} else {
			host, err := os.Hostname()
			if err != nil {
				panic(err)
			}
			m["host"] = host
		}

		// Earlier versions of EL doesnt accept '.' in field name
		if len(a.Separator) > 1 {
			return fmt.Errorf("FAILED Separator exceed one character : %s\n", a.Separator)
		}

		for key, value := range metric.Tags() {
			if key != "host" {
				if strings.HasPrefix(a.Version, "2.") {
					m[strings.Replace(key, ".", a.Separator, -1)] = value
				} else {
					m[key] = value
				}
			}
		}

		for key, value := range metric.Fields() {
			if strings.HasPrefix(a.Version, "2.") {
				m[strings.Replace(key, ".", a.Separator, -1)] = value
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
			return fmt.Errorf("FAILED to send elasticsearch message to index %s : %s\n", a.IndexName, errMessage)
		}

	}

	return nil
}

func (a *Elasticsearch) SampleConfig() string {
	return sampleConfig
}

func (a *Elasticsearch) Description() string {
	return "Configuration for Elasticsearch to send metrics to."
}

func buildMetrics(m telegraf.Metric) (map[string]Point, error) {
	ms := make(map[string]Point)
	for k, v := range m.Fields() {
		var p Point
		if err := p.setValue(v); err != nil {
			return ms, fmt.Errorf("unable to extract value from Fields, %s", err.Error())
		}
		p[0] = float64(m.Time().Unix())
		ms[k] = p
	}
	return ms, nil
}

func (p *Point) setValue(v interface{}) error {
	switch d := v.(type) {
	case int:
		p[1] = float64(int(d))
	case int32:
		p[1] = float64(int32(d))
	case int64:
		p[1] = float64(int64(d))
	case float32:
		p[1] = float64(d)
	case float64:
		p[1] = float64(d)
	default:
		return fmt.Errorf("undeterminable type")
	}
	return nil
}

func (a *Elasticsearch) Close() error {
	a.Client = nil
	return nil
}

func init() {
	outputs.Add("elasticsearch", func() telegraf.Output {
		return &Elasticsearch{
			EnableSniffer: false,
			Separator:     "_",
		}
	})
}
