package sumologic

import (
	"bytes"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
	"log"
	"net/http"
)

type SumoLogic struct {
	Prefix       string
	Template     string
	Timeout      internal.Duration
	CollectorUrl string
	client       *http.Client
}

var sampleConfig = `
  ## Prefix metrics name
  prefix = ""
  ## Sumo Logic output template
  template = "host.tags.measurement.field"
  ## Connection timeout.
  # timeout = "5s"
  ## SumoLogic Collector Url
  CollectorUrl = "collector url" # required.
`

func (s *SumoLogic) Connect() error {

	if s.CollectorUrl == "" {
		return fmt.Errorf("SumoLogic collector url is a required field for sumologic output")
	}

	s.client = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
		Timeout: s.Timeout.Duration,
	}
	return nil
}

func (s *SumoLogic) Close() error {
	return nil
}

func (s *SumoLogic) SampleConfig() string {
	return sampleConfig
}

func (s *SumoLogic) Description() string {
	return "Configuration for SumoLogic server to send metrics to"
}

type requestParameters struct {
	URL           string
	RequestMethod string
	RequestBody   []byte
}

func prepareHttpRequest(params requestParameters) (*http.Request, error) {
	req, err := http.NewRequest(params.RequestMethod, params.URL, bytes.NewBuffer(params.RequestBody))
	if err != nil {
		return req, fmt.Errorf("Error creating the HTTP request: %s", err.Error())
	}
	req.Header.Add("Content-Type", "application/vnd.sumologic.graphite")
	req.Header.Add("X-Sumo-Client", "telegraf")

	if req.Header.Get("User-Agent") == "" {
		req.Header.Add("User-Agent", "Mozilla/5.0 (compatible; Sumo/1.0; +https://sumologic.com)")
	}
	return req, nil
}

func (s *SumoLogic) Write(metrics []telegraf.Metric) error {

	var batch []byte
	sd, err := serializers.NewGraphiteSerializer(s.Prefix, s.Template)
	if err != nil {
		return err
	}

	for _, metric := range metrics {
		buf, err := sd.Serialize(metric)
		if err != nil {
			log.Printf("E! Error serializing some metrics to graphite: %s", err.Error())
		}
		batch = append(batch, buf...)
	}

	requestParameters := requestParameters{
		URL:           s.CollectorUrl,
		RequestMethod: "POST",
		RequestBody:   batch,
	}

	req, err := prepareHttpRequest(requestParameters)
	if err != nil {
		return fmt.Errorf("Error creating the HTTP request: %s\n", err.Error())
	}
	response, err := s.client.Do(req)

	if err != nil {
		return fmt.Errorf("error posting metrics to sumologic server, %s\n", err.Error())
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode > 209 {
		return fmt.Errorf("Received bad status code from server, %d\n", response.StatusCode)
	}

	return nil
}

func init() {
	outputs.Add("sumologic", func() telegraf.Output {
		return &SumoLogic{}
	})
}
