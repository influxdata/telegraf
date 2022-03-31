package multifile

import (
	"bytes"
	"fmt"
	"math"
	"os"
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
		fileContents, err := os.ReadFile(file.Name)

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

		var d int
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
			return fmt.Errorf("invalid conversion %v", file.Conversion)
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
