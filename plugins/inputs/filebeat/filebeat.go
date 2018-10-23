package filebeat

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
  ## An URL from which to read Filebeat-formatted JSON
  ## Default is "http://127.0.0.1:5066".
  url = "http://127.0.0.1:5066"

  ## Enable collection of the generic Beat stats
  collect_beat_stats = true

  ## Enable the collection if Libbeat stats
  collect_libbeat_stats = true

  ## Enable the collection of OS level stats
  collect_system_stats = false

  ## Enable the collection of Filebeat stats
  collect_filebeat_stats = true

  ## HTTP method
  # method = "GET"

  ## Optional HTTP headers
  # headers = {"X-Special-Header" = "Special-Value"}

  ## Override HTTP "Host" header
  # host_header = "logstash.example.com"

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

const description = "Read metrics exposed by Filebeat"

const suffixInfo = "/"
const suffixStats = "/stats"

type FileBeatInfo struct {
	Beat     string `json:"beat"`
	Hostname string `json:"hostname"`
	Name     string `json:"name"`
	UUID     string `json:"uuid"`
	Version  string `json:"version"`
}

type FileBeatStats struct {
	Beat     map[string]interface{} `json:"beat"`
	Filebeat interface{}            `json:"filebeat"`
	Libbeat  interface{}            `json:"libbeat"`
	System   interface{}            `json:"system"`
}

type Filebeat struct {
	URL string `toml:"url"`

	CollectBeatStats     bool `toml:"collect_beat_stats"`
	CollectLibbeatStats  bool `toml:"collect_libbeat_stats"`
	CollectSystemStats   bool `toml:"collect_system_stats"`
	CollectFilebeatStats bool `toml:"collect_filebeat_stats"`

	Username   string            `toml:"username"`
	Password   string            `toml:"password"`
	Method     string            `toml:"method"`
	Headers    map[string]string `toml:"headers"`
	HostHeader string            `toml:"host_header"`
	Timeout    internal.Duration `toml:"timeout"`

	tls.ClientConfig
	client *http.Client
}

func NewFilebeat() *Filebeat {
	return &Filebeat{
		URL:                  "http://127.0.0.1:5066",
		CollectBeatStats:     true,
		CollectLibbeatStats:  true,
		CollectSystemStats:   true,
		CollectFilebeatStats: true,
		Method:               "GET",
		Headers:              make(map[string]string),
		HostHeader:           "",
		Timeout:              internal.Duration{Duration: time.Second * 5},
	}
}

func (filebeat *Filebeat) Description() string {
	return description
}

func (filebeat *Filebeat) SampleConfig() string {
	return sampleConfig
}

// createHttpClient create a clients to access API
func (filebeat *Filebeat) createHttpClient() (*http.Client, error) {
	tlsConfig, err := filebeat.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: filebeat.Timeout.Duration,
	}

	return client, nil
}

// gatherJsonData query the data source and parse the response JSON
func (filebeat *Filebeat) gatherJsonData(url string, value interface{}) error {

	var method string
	if filebeat.Method != "" {
		method = filebeat.Method
	} else {
		method = "GET"
	}

	request, err := http.NewRequest(method, url, nil)
	if err != nil {
		return err
	}

	if (filebeat.Username != "") || (filebeat.Password != "") {
		request.SetBasicAuth(filebeat.Username, filebeat.Password)
	}
	for header, value := range filebeat.Headers {
		request.Header.Add(header, value)
	}
	if filebeat.HostHeader != "" {
		request.Host = filebeat.HostHeader
	}

	response, err := filebeat.client.Do(request)
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

func (filebeat *Filebeat) gatherInfoTags(url string) (map[string]string, error) {
	fileBeatInfo := &FileBeatInfo{}

	err := filebeat.gatherJsonData(url, fileBeatInfo)
	if err != nil {
		return nil, err
	}

	tags := map[string]string{
		"beat_id":      fileBeatInfo.UUID,
		"beat_name":    fileBeatInfo.Name,
		"beat_host":    fileBeatInfo.Hostname,
		"beat_version": fileBeatInfo.Version,
	}

	return tags, nil
}

func (filebeat *Filebeat) gatherStats(accumulator telegraf.Accumulator) error {
	fileBeatStats := &FileBeatStats{}

	infoUrl, err := url.Parse(filebeat.URL + suffixInfo)
	if err != nil {
		return err
	}
	statsUrl, err := url.Parse(filebeat.URL + suffixStats)
	if err != nil {
		return err
	}

	tags, err := filebeat.gatherInfoTags(infoUrl.String())
	if err != nil {
		return err
	}

	err = filebeat.gatherJsonData(statsUrl.String(), fileBeatStats)
	if err != nil {
		return err
	}

	if filebeat.CollectBeatStats {
		flattenerBeat := jsonparser.JSONFlattener{}
		err := flattenerBeat.FlattenJSON("", fileBeatStats.Beat)
		if err != nil {
			return err
		}
		accumulator.AddFields("filebeat_beat", flattenerBeat.Fields, tags)
	}

	if filebeat.CollectFilebeatStats {
		flattenerFilebeat := jsonparser.JSONFlattener{}
		err := flattenerFilebeat.FlattenJSON("", fileBeatStats.Filebeat)
		if err != nil {
			return err
		}
		accumulator.AddFields("filebeat", flattenerFilebeat.Fields, tags)
	}

	if filebeat.CollectLibbeatStats {
		flattenerLibbeat := jsonparser.JSONFlattener{}
		err := flattenerLibbeat.FlattenJSON("", fileBeatStats.Libbeat)
		if err != nil {
			return err
		}
		accumulator.AddFields("filebeat_libbeat", flattenerLibbeat.Fields, tags)
	}

	if filebeat.CollectSystemStats {
		flattenerSystem := jsonparser.JSONFlattener{}
		err := flattenerSystem.FlattenJSON("", fileBeatStats.System)
		if err != nil {
			return err
		}
		accumulator.AddFields("filebeat_system", flattenerSystem.Fields, tags)
	}

	return nil
}

func (filebeat *Filebeat) Gather(accumulator telegraf.Accumulator) error {
	if filebeat.client == nil {
		client, err := filebeat.createHttpClient()

		if err != nil {
			return err
		}
		filebeat.client = client
	}

	err := filebeat.gatherStats(accumulator)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	inputs.Add("filebeat", func() telegraf.Input {
		return NewFilebeat()
	})
}
