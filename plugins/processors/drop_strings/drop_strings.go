package drop_strings

import (
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/processors"
)

type DropStrings struct {
}

var sampleConfig = `
`

func (d *DropStrings) SampleConfig() string {
	return sampleConfig
}

func (d *DropStrings) Description() string {
	return "Drops all metrics of type string that pass through this filter."
}

func (d *DropStrings) Apply(in ...telegraf.Metric) []telegraf.Metric {
	out := make([]telegraf.Metric, 0, len(in))
	for _, source := range in {
		target, error := dropStrings(source)
		if error == nil {
			out = append(out, target)
		}
	}
	return out
}

func dropStrings(source telegraf.Metric) (telegraf.Metric, error) {
	inFields := source.Fields()
	outFields := make(map[string]interface{}, len(inFields))
	changed := false
	for key, value := range inFields {
		if _, drop := value.(string); !drop {
			outFields[key] = value
		} else {
			changed = true
		}
	}
	if !changed {
		return source, nil
	}
	if changed && len(outFields) > 0 {
		return metric.New(source.Name(), source.Tags(), outFields, source.Time(), source.Type())
	}
	return nil, fmt.Errorf("No more fields in metric '%s'", source.Name())
}

func init() {
	processors.Add("drop_strings", func() telegraf.Processor {
		return &DropStrings{}
	})
}
