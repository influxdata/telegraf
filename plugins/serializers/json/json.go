package json

import (
	ejson "encoding/json"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
)

type JsonSerializer struct {
	TimestampUnits internal.Duration
}

func (s *JsonSerializer) Serialize(metric telegraf.Metric) ([]byte, error) {
	m := make(map[string]interface{})
	m["tags"] = metric.Tags()
	m["fields"] = metric.Fields()
	m["name"] = metric.Name()
	m["timestamp"] = metric.UnixNano() / s.TimestampUnits.Duration.Nanoseconds()
	serialized, err := ejson.Marshal(m)
	if err != nil {
		return []byte{}, err
	}
	serialized = append(serialized, '\n')

	return serialized, nil
}
