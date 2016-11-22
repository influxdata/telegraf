package influx

import (
	"github.com/influxdata/telegraf"
)

type InfluxSerializer struct {
}

func (s *InfluxSerializer) Serialize(metric telegraf.Metric) ([]byte, error) {
	return metric.Serialize(), nil
}
