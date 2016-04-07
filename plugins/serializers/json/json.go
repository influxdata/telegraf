package json

import (
	ejson "encoding/json"

	"github.com/influxdata/telegraf"
)

type JsonSerializer struct {
}

func (s *JsonSerializer) Serialize(metric telegraf.Metric) ([]string, error) {
	out := []string{}

	m := make(map[string]interface{})
	m["tags"] = metric.Tags()
	m["fields"] = metric.Fields()
	m["name"] = metric.Name()
	m["timestamp"] = metric.UnixNano() / 1000000000
	serialized, err := ejson.Marshal(m)
	if err != nil {
		return []string{}, err
	}
	out = append(out, string(serialized))

	return out, nil
}
