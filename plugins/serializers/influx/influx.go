package influx

import (
	"github.com/influxdata/telegraf"
)

type InfluxSerializer struct {
}

func (s *InfluxSerializer) Serialize(m telegraf.Metric, output_precision ...string) ([]byte, error) {
	return m.Serialize(), nil
}
