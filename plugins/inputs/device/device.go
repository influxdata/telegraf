package device

import (
	"fmt"
	"io/ioutil"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/globpath"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
)

type Device struct {
	Type  string `toml:"type"`
	Devices  []string `toml:"devices"`

	DeviceTypes map[string]DeviceDescriptor
}

type DeviceDescriptor struct {
	Fields []DeviceField
}

type DeviceField struct {
	Path string
	Name string
	Type DeviceFieldType
	Factor float
}

type DeviceFieldType int

const (
	DeviceFieldType_Integer DeviceFieldType = iota + 1
	DeviceFieldType_Float DeviceFieldType
	DeviceFieldType_String DeviceFieldType
)

const sampleConfig = `
  # device type, only specified once
  type = "bme280"
  # type = "ina219"

  ## Devices to read each interval.
  devices = ["/sys/bus/i2c/devices/1-0076/iio:device0"]
`

// SampleConfig returns the default configuration of the Input
func (f *Device) SampleConfig() string {
	return sampleConfig
}

func (f *Device) Description() string {
	return "Collect metrics from devices not supported by lm_sensors"
}

func (f *Device) Gather(acc telegraf.Accumulator) error {
	err := f.refreshFilePaths()
	if err != nil {
		return err
	}
	for _, k := range f.filenames {
		metrics, err := f.readMetric(k)
		if err != nil {
			return err
		}

		for _, m := range metrics {
			acc.AddFields(m.Name(), m.Fields(), m.Tags(), m.Time())
		}
	}
	return nil
}

func init() {
	types := map[string]DeviceDescriptor{
		"bme280": DeviceDescriptor{
			Fields: []DeviceField{
				DeviceField{
					Name: "humidity_relative",
					Path: "in_humidityrelative_input",
					Type: DeviceFieldType_Float,
					Factor: 1/1000
				},
				DeviceField{
					Name: "temperature",
					Path: "in_temp_input",
					Type: DeviceFieldType_Float,
					Factor: 1/1000
				},
				DeviceField{
					Name: "pressure",
					Path: "in_pressure_input",
					Type: DeviceFieldType_Float,
					Factor: 1
				}
			}
		}
	}

	inputs.Add("device", func() telegraf.Input {
		return &Device{DeviceTypes: types}
	})
}
