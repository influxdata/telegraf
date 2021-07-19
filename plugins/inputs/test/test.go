package test

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
)

// taken ffrom file.go as an example but needs to be reworked
type Test struct {
	Metrics []string `toml:"metrics"`
	parser  parsers.Parser
}

const sampleConfig = `
	[[inputs.test]]
	## Metrics to parse each interval.
	metrics = [
		'weather,state=ny temperature=81.3',
		'weather,state=ca temperature=75.1'
	  ]

	## Data format to consume.
	## Each data format has its own unique set of configuration options, read
	## more about them here:
	## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
	data_format = "influx"
	`

// SampleConfig returns the default configuration of the Input
func (t *Test) SampleConfig() string {
	return sampleConfig
}

func (t *Test) Description() string {
	return "Parse a passed metric each interval"
}

func (t *Test) Init() error {
	var err error
	return err
}

func (t *Test) Gather(acc telegraf.Accumulator) error {
	for _, raw := range t.Metrics {
		metric, err := t.parser.Parse([]byte(raw))
		if err != nil {
			return err
		}

		for _, m := range metric {
			acc.AddMetric(m)
		}
	}
	return nil
}

func (t *Test) SetParser(p parsers.Parser) {
	t.parser = p
}

func init() {
	inputs.Add("test", func() telegraf.Input {
		return &Test{}
	})
}
