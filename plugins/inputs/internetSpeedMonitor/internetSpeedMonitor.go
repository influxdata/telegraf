package internetSpeedMonitor

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// SpeedMonitor is used to store configuration values.
type SpeedMonitor struct {
	EnableFileDownload bool            `toml:"enableFileDownload"`
	Measurement        string          `toml:"measurement"`
	Log                telegraf.Logger `toml:"-"`
}

func NewSpeedMonitor() telegraf.Input {
	return &SpeedMonitor{
		EnableFileDownload: false,
		Measurement:        "internet_speed",
	}
}

var SpeedMonitorConfig = `
## Sets if runs file download test
## Default: false  
enableFileDownload = false
	
## Sets measurement name
## Default: internet_speed
measurement = "internet_speed"  
`

// Description returns information about the plugin.
func (speedMonitor *SpeedMonitor) Description() string {
	return "Monitors internet speed in the network"
}

// SampleConfig displays configuration instructions.
func (speedMonitor *SpeedMonitor) SampleConfig() string {
	return SpeedMonitorConfig
}

func (speedMonitor *SpeedMonitor) Gather(acc telegraf.Accumulator) error {

	enableFileDownload := speedMonitor.EnableFileDownload
	measurement := speedMonitor.Measurement
	log := speedMonitor.Log
	c := make(chan InternetSpeedMonitor)
	go testInternetSpeed(c, enableFileDownload, log)
	results := <-c

	if results.Error != nil {
		return results.Error
	} else {
		fields := make(map[string]interface{})
		fields["download"] = results.Data.Download
		fields["upload"] = results.Data.Upload
		fields["latency"] = results.Data.Latency

		tags := make(map[string]string)

		acc.AddFields(measurement, fields, tags)
		return nil
	}
}
func init() {
	inputs.Add("internetSpeedMonitor", func() telegraf.Input {
		return &SpeedMonitor{
			EnableFileDownload: false,
			Measurement:        "internet_speed",
		}
	})
}
