//go:generate ../../../tools/readme_config_includer/generator
package temp

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/system"
)

//go:embed sample.conf
var sampleConfig string

type Temperature struct {
	ps system.PS
}

func (*Temperature) SampleConfig() string {
	return sampleConfig
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
		tags := map[string]string{
			"sensor": temp.SensorKey,
		}
		fields := map[string]interface{}{
			"temp": temp.Temperature,
		}
		acc.AddFields("temp", fields, tags)
	}
	return nil
}

func init() {
	inputs.Add("temp", func() telegraf.Input {
		return &Temperature{ps: system.NewSystemPS()}
	})
}
