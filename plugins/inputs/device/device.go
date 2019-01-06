package device

import (
	"bytes"
	"io/ioutil"
	"path"
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Device struct {
	Type    string   `toml:"type"`
	Devices []string `toml:"devices"`

	devicetypes map[string]DeviceDescriptor
}

type DeviceDescriptor struct {
	Fields []DeviceField
}

type DeviceField struct {
	Path   string
	Name   string
	Type   DeviceFieldType
	Factor float64
}

type DeviceFieldType int

const (
	DeviceFieldType_Integer DeviceFieldType = iota + 1
	DeviceFieldType_Float
	DeviceFieldType_String
	DeviceFieldType_Bool
)

const sampleConfig = `
  # device type, only specified once
  type = "bme280"
  # type = "ina219"

  # list of device paths
  devices = ["/sys/bus/i2c/devices/1-0076/iio:device0"]
`

// SampleConfig returns the default configuration of the Input
func (d *Device) SampleConfig() string {
	return sampleConfig
}

func (d *Device) Description() string {
	return "Collect metrics from devices not supported by lm_sensors"
}

func (d *Device) Gather(acc telegraf.Accumulator) error {
	devicetype := d.devicetypes[d.Type]
	now := time.Now()

	for _, devpath := range d.Devices {
		tags := map[string]string{
			"device": devpath,
		}

		fields := make(map[string]interface{})

		for _, field := range devicetype.Fields {
			fileContents, err := ioutil.ReadFile(path.Join(devpath, field.Path))

			if err != nil {
				return err
			}

			vStr := string(bytes.TrimSpace(bytes.Trim(fileContents, "\x00")))

			var value interface{}

			switch field.Type {
			case DeviceFieldType_Integer:
				value, err = strconv.Atoi(vStr)
			case DeviceFieldType_Float:
				value, err = strconv.ParseFloat(vStr, 64)
			case DeviceFieldType_String:
				value = vStr
			case DeviceFieldType_Bool:
				value, err = strconv.ParseBool(vStr)
			}

			if err != nil {
				return err
			}

			if field.Type == DeviceFieldType_Float {
				value = value.(float64) * field.Factor
			}

			fields[field.Name] = value
		}

		acc.AddGauge("device", fields, tags, now)
	}
	return nil
}

func init() {
	types := map[string]DeviceDescriptor{
		"bme280": DeviceDescriptor{
			Fields: []DeviceField{
				DeviceField{
					Name:   "humidity_relative",
					Path:   "in_humidityrelative_input",
					Type:   DeviceFieldType_Float,
					Factor: 1e-3,
				},
				DeviceField{
					Name:   "temperature",
					Path:   "in_temp_input",
					Type:   DeviceFieldType_Float,
					Factor: 1e-3,
				},
				DeviceField{
					Name:   "pressure",
					Path:   "in_pressure_input",
					Type:   DeviceFieldType_Float,
					Factor: 1.,
				},
			},
		},
		"ina219": DeviceDescriptor{
			Fields: []DeviceField{
				DeviceField{
					Name:   "current",
					Path:   "curr1_input",
					Type:   DeviceFieldType_Float,
					Factor: 1e-4,
				},
				DeviceField{
					Name:   "voltage",
					Path:   "in1_input",
					Type:   DeviceFieldType_Float,
					Factor: 1e-3,
				},
				DeviceField{
					Name:   "power",
					Path:   "power1_input",
					Type:   DeviceFieldType_Float,
					Factor: 1e-7,
				},
			},
		},
	}

	inputs.Add("device", func() telegraf.Input {
		return &Device{devicetypes: types}
	})
}
