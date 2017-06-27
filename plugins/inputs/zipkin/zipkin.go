package zipkin

import "github.com/influxdata/telegraf"

type Zipkin struct {
	//Add configuration fields here later
	Field bool
}

const sampleConfig = `
  ##
  # field = value
`

func (z Zipkin) Description() string {
	return "Allows for the collection of zipkin tracing spans for storage in influxdb"
}

func (z Zipkin) SampleConfig() string {
	return sampleConfig
}

func (z *Zipkin) Gather(acc telegraf.Accumulator) {
	if z.Field {
		acc.AddFields("state", map[string]interface{}{"value": "true"}, nil)
	} else {
		acc.AddFields("state", map[string]interface{}{"value": "false"}, nil)
	}
}
