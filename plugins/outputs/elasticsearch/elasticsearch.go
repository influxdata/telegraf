package elasticsearch

import (
	"fmt"
	"time"
	"os"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
  	"gopkg.in/olivere/elastic.v2"
)

type Elasticsearch struct {
	ServerHost    string
	IndexName     string
	Client 	      *elastic.Client
}

var sampleConfig = `
  server_host = "http://10.10.10.10:19200" # required.
  index_name = "twitter" #required
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
                elastic.SetSniff(false),
                elastic.SetHealthcheckInterval(30*time.Second),
                elastic.SetURL(a.ServerHost),
        )

  	if err != nil {
    		// Handle error
    		panic(err)
  	}

	a.Client = client

	return nil
}

func (a *Elasticsearch) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	for _, metric := range metrics {

	        m := make(map[string]interface{})
		m["created"] = time.Now().UnixNano() / 1000000
        	m["version"] = "1.1"
       		m["timestamp"] = metric.UnixNano() / 1000000
        	m["short_message"] = " "
        	m["name"] = metric.Name()

        	if host, ok := metric.Tags()["host"]; ok {
                	m["host"] = host
        	} else {
                	host, err := os.Hostname()
                	if err != nil {
				panic(err)
                	}
                	m["host"] = host
        	}

        	for key, value := range metric.Tags() {
                	if key != "host" {
                        	m["_"+key] = value
                	}
        	}

        	for key, value := range metric.Fields() {
                	m["_"+key] = value
        	}

		_, err := a.Client.Index().
                	Index(a.IndexName).
                	Type("stats2").
			BodyJson(m).
                	Do()

        	if err != nil {
                	// Handle error
                	panic(err)
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
	return nil
}

func init() {
	outputs.Add("elasticsearch", func() telegraf.Output {
		return &Elasticsearch{}
	})
}
