package value

import (
	"fmt"
	"github.com/influxdata/telegraf"
)

type ValueSerializer struct {
}

func (s *ValueSerializer) Serialize(metric telegraf.Metric) ([]byte, error) {
	result := fmt.Sprintf("%v", metric.Fields()["value"])
	return []byte(result), nil
}
