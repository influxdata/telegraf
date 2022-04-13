package beat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	jsonparser "github.com/influxdata/telegraf/plugins/parsers/json"
)

const suffixInfo = "/"
const suffixStats = "/stats"

type Info struct {
	Beat     string `json:"beat"`
	Hostname string `json:"hostname"`
	Name     string `json:"name"`
	UUID     string `json:"uuid"`
	Version  string `json:"version"`
}

type Stats struct {
	Beat     map[string]interface{} `json:"beat"`
	FileBeat interface{}            `json:"filebeat"`
	Libbeat  interface{}            `json:"libbeat"`
	System   interface{}            `json:"system"`
}

type Beat struct {
	URL string `toml:"url"`

	Includes []string `toml:"include"`

	Username   string            `toml:"username"`
	Password   string            `toml:"password"`
	Method     string            `toml:"method"`
	Headers    map[string]string `toml:"headers"`
	HostHeader string            `toml:"host_header"`
	Timeout    config.Duration   `toml:"timeout"`

	tls.ClientConfig
	client *http.Client
}

func NewBeat() *Beat {
	return &Beat{
		URL:      "http://127.0.0.1:5066",
		Includes: []string{"beat", "libbeat", "filebeat"},
		Method:   "GET",
		Headers:  make(map[string]string),
		Timeout:  config.Duration(time.Second * 5),
	}
}

func (beat *Beat) Init() error {
	availableStats := []string{"beat", "libbeat", "system", "filebeat"}

	var err error
	beat.client, err = beat.createHTTPClient()

	if err != nil {
		return err
	}

	err = choice.CheckSlice(beat.Includes, availableStats)
	if err != nil {
		return err
	}

	return nil
}

// createHTTPClient create a clients to access API
func (beat *Beat) createHTTPClient() (*http.Client, error) {
	tlsConfig, err := beat.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: time.Duration(beat.Timeout),
	}

	return client, nil
}

// gatherJSONData query the data source and parse the response JSON
func (beat *Beat) gatherJSONData(address string, value interface{}) error {
	request, err := http.NewRequest(beat.Method, address, nil)
	if err != nil {
		return err
	}

	if beat.Username != "" {
		request.SetBasicAuth(beat.Username, beat.Password)
	}
	for k, v := range beat.Headers {
		request.Header.Add(k, v)
	}
	if beat.HostHeader != "" {
		request.Host = beat.HostHeader
	}

	response, err := beat.client.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	return json.NewDecoder(response.Body).Decode(value)
}

func (beat *Beat) Gather(accumulator telegraf.Accumulator) error {
	beatStats := &Stats{}
	beatInfo := &Info{}

	infoURL, err := url.Parse(beat.URL + suffixInfo)
	if err != nil {
		return err
	}
	statsURL, err := url.Parse(beat.URL + suffixStats)
	if err != nil {
		return err
	}

	err = beat.gatherJSONData(infoURL.String(), beatInfo)
	if err != nil {
		return err
	}
	tags := map[string]string{
		"beat_beat":    beatInfo.Beat,
		"beat_id":      beatInfo.UUID,
		"beat_name":    beatInfo.Name,
		"beat_host":    beatInfo.Hostname,
		"beat_version": beatInfo.Version,
	}

	err = beat.gatherJSONData(statsURL.String(), beatStats)
	if err != nil {
		return err
	}

	for _, name := range beat.Includes {
		var stats interface{}
		var metric string

		switch name {
		case "beat":
			stats = beatStats.Beat
			metric = "beat"
		case "filebeat":
			stats = beatStats.FileBeat
			metric = "beat_filebeat"
		case "system":
			stats = beatStats.System
			metric = "beat_system"
		case "libbeat":
			stats = beatStats.Libbeat
			metric = "beat_libbeat"
		default:
			return fmt.Errorf("unknown stats-type %q", name)
		}
		flattener := jsonparser.JSONFlattener{}
		err := flattener.FullFlattenJSON("", stats, true, true)
		if err != nil {
			return err
		}
		accumulator.AddFields(metric, flattener.Fields, tags)
	}

	return nil
}

func init() {
	inputs.Add("beat", func() telegraf.Input {
		return NewBeat()
	})
}
