package influx

import (
	"github.com/influxdata/telegraf/plugins"
)

type InfluxSerializer struct {
}

func (s *InfluxSerializer) Serialize(m plugins.Metric) ([]byte, error) {
	return m.Serialize(), nil
}
