package printer

import (
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
)

type Printer struct {
	serializer serializers.Serializer
}

func (p *Printer) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range in {
		octets, err := p.serializer.Serialize(metric)
		if err != nil {
			continue
		}
		fmt.Printf("%s", octets)
	}
	return in
}

func init() {
	processors.Add("printer", func() telegraf.Processor {
		return &Printer{
			serializer: influx.NewSerializer(),
		}
	})
}
