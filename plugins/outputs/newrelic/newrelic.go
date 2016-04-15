package newrelic

import (
	"bytes"
	"fmt"
  "time"
	"net/http"
	// "io"

  "encoding/json"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/internal"
)

type NewRelic struct {
	ApiKey string
	GuidBase string
	Timeout internal.Duration

  LastWrite time.Time
	client *http.Client
}

var sampleConfig = `
  ## Your NewRelic Api Key
  api_key = "XXXX" # required

	## Guid base
  ## This allows you have unique GUIDs for your installation. This will generate
	## a separate "plugin" GUID for each of the inputs that you use.
	##
	## see https://docs.newrelic.com/docs/plugins/plugin-developer-resources/planning-your-plugin/parts-plugin#guid
	##
	## This setting will allow you to "fork" your "plugins", and have your own
	## dashboards and settings for them.
	## The default behaviour is that the original author of the plugin sets up
	## all the dashboards; other users cannot modify them.
	## As it is very hard to provide useful defaults for all possible setup, we
	## instead allow you to make your "own plugin" and modify the dashboards.
	##
	## The drawback is that your GUID must be unique, and that you must setup
	## your own dashboards for everything.
	##
	## TODO: The default for this should be
	##       a "proper" GUID that is maintained to have reasonable default
	# guid_base = 'my.domain.something.something' # TODO must still be implemented

	## Metric Type TODO - Not yet implemented
	##
	## Can either be "Component" or "Custom"
	##
	## Component metrics are the default for plugins. They make the metrics
	## available even to free accounts, but with the restrictions mentioned above.
	##
	## Custom metrics don't show up as plugins. They are freely usable in custom
	## dashboards, but you need to have a paid subscription to see the data.
	##
	## Default is "Component"
	# metric_type = "Custom"
`
func (nr *NewRelic) Connect() error {
	if nr.ApiKey == "" {
		return fmt.Errorf("apikey is a required field for newrelic output")
	}

	if nr.GuidBase == "" {
		nr.GuidBase = "com.influxdata.demo-newrelic-agent"
	}
	nr.client = &http.Client{
		Timeout: nr.Timeout.Duration,
	}
  nr.LastWrite = time.Now()
	return nil
}

func (nr *NewRelic) Close() error {
	return nil
}

func (nr *NewRelic) SampleConfig() string {
	return sampleConfig
}

func (nr *NewRelic) Description() string {
	return "Send telegraf metrics to NewRelic"
}

func (nr *NewRelic) PostPluginData(jsonData []byte) error {
	req, reqErr := http.NewRequest("POST", "https://platform-api.newrelic.com/platform/v1/metrics", bytes.NewBuffer(jsonData))
	if reqErr != nil { return reqErr }
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-License-Key", nr.ApiKey)

	resp, respErr := nr.client.Do(req)
	if respErr != nil { return respErr }
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 209 {
		return fmt.Errorf("received bad status code, %d\n", resp.StatusCode)
	}

	return nil
}

func (nr *NewRelic) SendDataPage(dataPage interface{}) error {
		cmpJson, err := json.Marshal(dataPage)
		if err != nil { return err }
		fmt.Println("Sending " + string(cmpJson) + " <")
		return nr.PostPluginData(cmpJson)
}

func (nr *NewRelic) Write(metrics []telegraf.Metric) error {
	data := NewRelicData{LastWrite: nr.LastWrite, Hosts: make(map[string][]NewRelicComponent)}
	data.AddMetrics(metrics)

	for _, dataPage := range(data.DataSets()) {
		nr.SendDataPage(dataPage)
	}
  nr.LastWrite = time.Now()
	return nil
}


func init() {
    outputs.Add("newrelic", func() telegraf.Output { return &NewRelic{} })
}
