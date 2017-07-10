// +build windows

package win_services

import (
	"log"
	"math/rand"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var sampleConfig = `
  ## This plugin returns by default service state and startup mode
  ## See the README file for more examples.
  ## Uncomment examples below or write your own as you see fit. If the system
  ## being polled for data does not have the Object at startup of the Telegraf
  ## agent, it will not be gathered.
  ## Settings:

  # Names of services to monitor
  Services = [
    "Server"
  ]
  Measurement = "win_services"
  # CustomTagName=Group
  # CustomTagValue=alpha
`

var description = "Input plugin to report Windows services info: name, state, startup mode, hostname"

type Win_Services struct {
	Services    []string
	Measurement     string
	CustomTagName	string
	CustomTagValue  string

	configParsed bool
}

type service struct {
	ServiceName		string
	State      		int
	StartUpMode     int
}


func (m *Win_Services) Description() string {
	return description
}

func (m *Win_Services) SampleConfig() string {
	return sampleConfig
}

func (m *Win_Services) ParseConfig() error {
	log.Printf("win_services: parse config: %v\n", *m)
	return nil
}

func (m *Win_Services) Gather(acc telegraf.Accumulator) error {
	// Parse the config once
	if !m.configParsed {
		err := m.ParseConfig()
		m.configParsed = true
		if err != nil {
			return err
		}
	}

	fields := make(map[string]interface{})
	tags := make(map[string]string)
	tags["service"] = "Server";
	fields["state"] = rand.Int()%3;
	fields["startupMode"] = 3;

	measurement := m.Measurement
	if measurement == "" {
		measurement = "win_services"
	}
	acc.AddFields(measurement, fields, tags)

	return nil
}

func init() {
	log.Println("win_services: init")
	inputs.Add("win_services", func() telegraf.Input { return &Win_Services{} })
}
