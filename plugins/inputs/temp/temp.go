package temp

import (
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/system"
)

type Temperature struct {
	ps system.PS
}

func (t *Temperature) Description() string {
	return "Read metrics about temperature"
}

const sampleConfig = ""

func (t *Temperature) SampleConfig() string {
	return sampleConfig
}

func (t *Temperature) Gather(acc telegraf.Accumulator) error {
	temps, err := t.ps.Temperature()
	if err != nil {
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
