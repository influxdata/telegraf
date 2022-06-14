//go:generate ../../../tools/readme_config_includer/generator
package temp

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/system"
)

// DO NOT REMOVE THE NEXT TWO LINES! This is required to embed the sampleConfig data.
//go:embed sample.conf
var sampleConfig string

type Temperature struct {
	Scheme string `toml:"output_scheme"`

	ps system.PS
}

func (*Temperature) SampleConfig() string {
	return sampleConfig
}

func (t *Temperature) Init() error {
	if !choice.Contains(t.Scheme, []string{"", "measurement", "field"}) {
		return fmt.Errorf("invalid output_scheme %q", t.Scheme)
	}
	return nil
}

func (t *Temperature) Gather(acc telegraf.Accumulator) error {
	temps, err := t.ps.Temperature()
	if err != nil {
		if strings.Contains(err.Error(), "not implemented yet") {
			return fmt.Errorf("plugin is not supported on this platform: %v", err)
		}
		return fmt.Errorf("error getting temperatures info: %s", err)
	}
	for _, temp := range temps {
		switch t.Scheme {
		case "", "measurement":
			acc.AddFields(
				"temp",
				map[string]interface{}{"temp": temp.Critical},
				map[string]string{"sensor": temp.SensorKey + "_crit"},
			)
			acc.AddFields(
				"temp",
				map[string]interface{}{"temp": temp.High},
				map[string]string{"sensor": temp.SensorKey + "_max"},
			)
			acc.AddFields(
				"temp",
				map[string]interface{}{"temp": temp.Temperature},
				map[string]string{"sensor": temp.SensorKey + "_input"},
			)
		case "field":
			acc.AddFields(
				"temp",
				map[string]interface{}{
					"crit": temp.Critical,
					"high": temp.High,
					"temp": temp.Temperature,
				},
				map[string]string{"sensor": temp.SensorKey},
			)
		}
	}
	return nil
}

func init() {
	inputs.Add("temp", func() telegraf.Input {
		return &Temperature{ps: system.NewSystemPS()}
	})
}
