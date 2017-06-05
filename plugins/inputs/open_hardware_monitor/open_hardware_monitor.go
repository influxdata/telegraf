// +build windows

package open_hardware_monitor

import (
	"fmt"
	"github.com/StackExchange/wmi"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"strings"
)

type OpenHardwareMonitorConfig struct {
	SensorsType []string
	Parent      []string
}

type OpenHardwareMonitorData struct {
	Name       string
	SensorType string
	Parent     string
	Value      float32
}

func (p *OpenHardwareMonitorConfig) Description() string {
	return "Get sensors data from Open Hardware Monitor via WMI"
}

const sampleConfig = `
	## Sensors to query ( if not given then all is queried )
	SensorsType = ["Temperature", "Fan", "Voltage"] # optional
	
	## Which hardware should be available
	Parent = ["_intelcpu_0"]  # optional
`

func (p *OpenHardwareMonitorConfig) SampleConfig() string {
	return sampleConfig
}

func (p *OpenHardwareMonitorConfig) CreateQuery() (string, error) {
	query := "SELECT * FROM SENSOR"
	if len(p.SensorsType) != 0 {
		query += " WHERE "
		var sensors []string
		for _, sensor := range p.SensorsType {
			sensors = append(sensors, fmt.Sprint("SensorType='", sensor, "'"))
		}
		query += strings.Join(sensors, " OR ")
	}
	return query, nil
}

func (p *OpenHardwareMonitorConfig) QueryData(query string) ([]OpenHardwareMonitorData, error) {
	var dst []OpenHardwareMonitorData
	err := wmi.QueryNamespace(query, &dst, "root/OpenHardwareMonitor")

	// Replace all spaces
	replace := map[string]string{
		" ": "_",
		"/": "_",
	}
	for i := range dst {
		for key, value := range replace {
			dst[i].Name = strings.Replace(strings.Trim(dst[i].Name, key), key, value, -1)
			dst[i].SensorType = strings.Replace(strings.Trim(dst[i].SensorType, key), key, value, -1)
			dst[i].Parent = strings.Replace(strings.Trim(dst[i].Parent, key), key, value, -1)
		}
	}

	return dst, err
}

func contains(s []string, e string) bool {
	if len(s) == 0 {
		return true
	}
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func (p *OpenHardwareMonitorConfig) Gather(acc telegraf.Accumulator) error {
	var dst []OpenHardwareMonitorData
	query, err := p.CreateQuery()
	if err == nil {
		dst, err = p.QueryData(query)
		if err != nil {
			acc.AddError(err)
		}
		for _, sensorData := range dst {
			if contains(p.Parent, sensorData.Parent) {
				tags := map[string]string{
					"name":   sensorData.Name,
					"parent": sensorData.Parent,
				}
				fields := map[string]interface{}{sensorData.SensorType: sensorData.Value}
				acc.AddFields("ohm", fields, tags)
			}
		}
	} else {
		acc.AddError(err)
	}
	return nil
}

func init() {
	inputs.Add("open_hardware_monitor", func() telegraf.Input {
		return &OpenHardwareMonitorConfig{}
	})
}
