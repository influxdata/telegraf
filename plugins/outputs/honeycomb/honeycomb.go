package honeycomb

import (
	"errors"
	"fmt"
	"github.com/honeycombio/libhoney-go"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

type Honeycomb struct {
	APIKey      string   `toml:"apiKey"`
	Dataset     string   `toml:"dataset"`
	APIHost     string   `toml:"apiHost"`
	SpecialTags []string `toml:"specialTags"`
}

var sampleConfig = `
  ## Honeycomb authentication token
  apiKey = "API_KEY"

  ## Dataset name in Honeycomb to send data to
  dataset = "my-dataset"  

  ## Special tags that will not get prefixed by the measurement name
  ## This should be set if you specified global tags, and it should include the list of all global tags + host
  ## Default value for this list is: host
  #specialTags = ["host"]
  
  ## Optional: the hostname for the Honeycomb API server
  #apiHost = "https://api.honeycomb.io/""
`

func (h *Honeycomb) Connect() error {

	if h.APIKey == "" {
		return errors.New("Honeycomb apiKey can not be empty")
	}

	if h.Dataset == "" {
		return errors.New("Honeycomb dataset can not be empty")
	}

	err := libhoney.Init(libhoney.Config{
		APIKey:       h.APIKey,
		Dataset:      h.Dataset,
		APIHost:      h.APIHost,
		MaxBatchSize: 500,
	})
	if err != nil {
		return fmt.Errorf("Honeycomb Init error: %s", err.Error())
	}

	return nil
}

func (h *Honeycomb) Write(metrics []telegraf.Metric) error {

	for _, m := range metrics {

		// BuildEvent from metric
		ev, err := h.BuildEvent(m)
		if err != nil {
			return fmt.Errorf("Honeycomb event creation error: %s", err.Error())
		}

		// send event
		if err = ev.Send(); err != nil {
			return fmt.Errorf("Honeycomb Send error: %s", err.Error())
		}
	}

	libhoney.Flush()

	return nil
}

func (h *Honeycomb) BuildEvent(m telegraf.Metric) (*libhoney.Event, error) {
	// define data payload for event
	data := make(map[string]interface{})

	// add each field and value prefixed by metric / measurement name to data payload
	for _, f := range m.FieldList() {
		data[m.Name()+"."+f.Key] = f.Value
	}

	// add each tag and value to data payload
	for _, t := range m.TagList() {
		// don't add the prefix to special tags
		prefixTag := true
		for _, st := range h.SpecialTags {
			if t.Key == st {
				prefixTag = false
				break
			}
		}
		if prefixTag {
			data[m.Name()+"."+t.Key] = t.Value
		} else {
			data[t.Key] = t.Value
		}
	}

	// create event, set timestamp and payload
	ev := libhoney.NewEvent()
	ev.Timestamp = m.Time()
	if err := ev.Add(data); err != nil {
		return nil, err
	}

	return ev, nil
}

func (h *Honeycomb) SampleConfig() string {
	return sampleConfig
}

func (h *Honeycomb) Description() string {
	return "Send telegraf metrics to Honeycomb.io"
}

func (h *Honeycomb) Close() error {
	libhoney.Close()
	return nil
}

func init() {
	outputs.Add("honeycomb", func() telegraf.Output {
		return &Honeycomb{
			SpecialTags: []string{"host"},
		}
	})
}
