package multifile

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"path"
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type MultiFile struct {
	BaseDir   string
	FailEarly bool
	Files     []File `toml:"file"`

	initialized bool
}

type File struct {
	Name       string `toml:"file"`
	Dest       string
	Conversion string
}

const sampleConfig = `
  ## Base directory where telegraf will look for files.
  ## Omit this option to use absolute paths.
  base_dir = "/sys/bus/i2c/devices/1-0076/iio:device0"

  ## If true, Telegraf discard all data when a single file can't be read.
  ## Else, Telegraf omits the field generated from this file.
  # fail_early = true

  ## Files to parse each interval.
  [[inputs.multifile.file]]
    file = "in_pressure_input"
    dest = "pressure"
    conversion = "float"
  [[inputs.multifile.file]]
    file = "in_temp_input"
    dest = "temperature"
    conversion = "float(3)"
  [[inputs.multifile.file]]
    file = "in_humidityrelative_input"
    dest = "humidityrelative"
    conversion = "float(3)"
`

// SampleConfig returns the default configuration of the Input
func (m *MultiFile) SampleConfig() string {
	return sampleConfig
}

func (m *MultiFile) Description() string {
	return "Aggregates the contents of multiple files into a single point"
}

func (m *MultiFile) init() {
	if m.initialized {
		return
	}

	for i, file := range m.Files {
		if m.BaseDir != "" {
			m.Files[i].Name = path.Join(m.BaseDir, file.Name)
		}
		if file.Dest == "" {
			m.Files[i].Dest = path.Base(file.Name)
		}
	}

	m.initialized = true
}

func (m *MultiFile) Gather(acc telegraf.Accumulator) error {
	m.init()
	now := time.Now()
	fields := make(map[string]interface{})
	tags := make(map[string]string)

	for _, file := range m.Files {
		fileContents, err := ioutil.ReadFile(file.Name)

		if err != nil {
			if m.FailEarly {
				return err
			}
			continue
		}

		vStr := string(bytes.TrimSpace(bytes.Trim(fileContents, "\x00")))

		if file.Conversion == "tag" {
			tags[file.Dest] = vStr
			continue
		}

		var value interface{}

		var d int = 0
		if _, errfmt := fmt.Sscanf(file.Conversion, "float(%d)", &d); errfmt == nil || file.Conversion == "float" {
			var v float64
			v, err = strconv.ParseFloat(vStr, 64)
			value = v / math.Pow10(d)
		}

		if file.Conversion == "int" {
			value, err = strconv.ParseInt(vStr, 10, 64)
		}

		if file.Conversion == "string" || file.Conversion == "" {
			value = vStr
		}

		if file.Conversion == "bool" {
			value, err = strconv.ParseBool(vStr)
		}

		if err != nil {
			if m.FailEarly {
				return err
			}
			continue
		}

		if value == nil {
			return errors.New(fmt.Sprintf("invalid conversion %v", file.Conversion))
		}

		fields[file.Dest] = value
	}

	acc.AddGauge("multifile", fields, tags, now)
	return nil
}

func init() {
	inputs.Add("multifile", func() telegraf.Input {
		return &MultiFile{
			FailEarly: true,
		}
	})
}
