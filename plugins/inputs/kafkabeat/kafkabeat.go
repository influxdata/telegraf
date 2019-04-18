package kafkabeat

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"

	jsonparser "github.com/influxdata/telegraf/plugins/parsers/json"
)

const sampleConfig = `
  ## An URL from which to read Kafkabeat-formatted JSON
  ## Default is "http://127.0.0.1:5066".
  url = "http://127.0.0.1:5066"

  ## Enable collection of the generic Beat stats
  collect_beat_stats = true

  ## Enable the collection if Libbeat stats
  collect_libbeat_stats = true

  ## Enable the collection of OS level stats
  collect_system_stats = true

  ## HTTP method
  # method = "GET"

  ## Optional HTTP headers
  # headers = {"X-Special-Header" = "Special-Value"}

  ## Override HTTP "Host" header
  # host_header = "kafkabeat.example.com"

  ## Timeout for HTTP requests
  timeout = "5s"

  ## Optional HTTP Basic Auth credentials
  # username = "username"
  # password = "pa$$word"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
`

const description = "Read metrics exposed by Kafkabeat"

const suffixInfo = "/"
const suffixStats = "/stats"

type KafkaBeatInfo struct {
	Beat     string `json:"beat"`
	Hostname string `json:"hostname"`
	Name     string `json:"name"`
	UUID     string `json:"uuid"`
	Version  string `json:"version"`
}

type KafkabeatStats struct {
	Beat     map[string]interface{} `json:"beat"`
	Kafkabeat interface{}           `json:"kafkabeat"`
	Libbeat  interface{}            `json:"libbeat"`
	System   interface{}            `json:"system"`
}

type Kafkabeat struct {
	URL string `toml:"url"`

	CollectBeatStats      bool `toml:"collect_beat_stats"`
	CollectLibbeatStats   bool `toml:"collect_libbeat_stats"`
	CollectSystemStats    bool `toml:"collect_system_stats"`

	Username   string            `toml:"username"`
	Password   string            `toml:"password"`
	Method     string            `toml:"method"`
	Headers    map[string]string `toml:"headers"`
	HostHeader string            `toml:"host_header"`
	Timeout    internal.Duration `toml:"timeout"`

	tls.ClientConfig
	client *http.Client
}

func NewKafkabeat() *Kafkabeat {
	return &Kafkabeat{
		URL:                  "http://127.0.0.1:5066",
		CollectBeatStats:     true,
		CollectLibbeatStats:  true,
		CollectSystemStats:   true,
		Method:               "GET",
		Headers:              make(map[string]string),
		HostHeader:           "",
		Timeout:              internal.Duration{Duration: time.Second * 5},
	}
}

func (kafkabeat *Kafkabeat) Description() string {
	return description
}

func (kafkabeat *Kafkabeat) SampleConfig() string {
	return sampleConfig
}

// createHttpClient create a clients to access API
func (kafkabeat *Kafkabeat) createHttpClient() (*http.Client, error) {
	tlsConfig, err := kafkabeat.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: kafkabeat.Timeout.Duration,
	}

	return client, nil
}

// gatherJsonData query the data source and parse the response JSON
func (kafkabeat *Kafkabeat) gatherJsonData(url string, value interface{}) error {

	var method string
	if kafkabeat.Method != "" {
		method = kafkabeat.Method
	} else {
		method = "GET"
	}

	request, err := http.NewRequest(method, url, nil)
	if err != nil {
		return err
	}

	if (kafkabeat.Username != "") || (kafkabeat.Password != "") {
		request.SetBasicAuth(kafkabeat.Username, kafkabeat.Password)
	}
	for header, value := range kafkabeat.Headers {
		request.Header.Add(header, value)
	}
	if kafkabeat.HostHeader != "" {
		request.Host = kafkabeat.HostHeader
	}

	response, err := kafkabeat.client.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	err = json.NewDecoder(response.Body).Decode(value)
	if err != nil {
		return err
	}

	return nil
}

func (kafkabeat *Kafkabeat) gatherInfoTags(url string) (map[string]string, error) {
	KafkaBeatInfo := &KafkaBeatInfo{}

	err := kafkabeat.gatherJsonData(url, KafkaBeatInfo)
	if err != nil {
		return nil, err
	}

	tags := map[string]string{
		"beat_id":      KafkaBeatInfo.UUID,
		"beat_name":    KafkaBeatInfo.Name,
		"beat_host":    KafkaBeatInfo.Hostname,
		"beat_version": KafkaBeatInfo.Version,
	}

	return tags, nil
}

func (kafkabeat *Kafkabeat) gatherStats(accumulator telegraf.Accumulator) error {
	kafkabeatStats := &KafkabeatStats{}

	infoUrl, err := url.Parse(kafkabeat.URL + suffixInfo)
	if err != nil {
		return err
	}
	statsUrl, err := url.Parse(kafkabeat.URL + suffixStats)
	if err != nil {
		return err
	}

	tags, err := kafkabeat.gatherInfoTags(infoUrl.String())
	if err != nil {
		return err
	}

	err = kafkabeat.gatherJsonData(statsUrl.String(), kafkabeatStats)
	if err != nil {
		return err
	}

	if kafkabeat.CollectBeatStats {
		flattenerBeat := jsonparser.JSONFlattener{}
		err := flattenerBeat.FlattenJSON("", kafkabeatStats.Beat)
		if err != nil {
			return err
		}
		accumulator.AddFields("kafkabeat_beat", flattenerBeat.Fields, tags)
	}

	if kafkabeat.CollectLibbeatStats {
		flattenerLibbeat := jsonparser.JSONFlattener{}
		err := flattenerLibbeat.FlattenJSON("", kafkabeatStats.Libbeat)
		if err != nil {
			return err
		}
		accumulator.AddFields("kafkabeat_libbeat", flattenerLibbeat.Fields, tags)
	}

	if kafkabeat.CollectSystemStats {
		flattenerSystem := jsonparser.JSONFlattener{}
		err := flattenerSystem.FlattenJSON("", kafkabeatStats.System)
		if err != nil {
			return err
		}
		accumulator.AddFields("kafkabeat_system", flattenerSystem.Fields, tags)
	}

	return nil
}

func (kafkabeat *Kafkabeat) Gather(accumulator telegraf.Accumulator) error {
	if kafkabeat.client == nil {
		client, err := kafkabeat.createHttpClient()

		if err != nil {
			return err
		}
		kafkabeat.client = client
	}

	err := kafkabeat.gatherStats(accumulator)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	inputs.Add("kafkabeat", func() telegraf.Input {
		return NewKafkabeat()
	})
}
