package gelf

import (
	ejson "encoding/json"

	"github.com/influxdata/telegraf"
)

type GelfSerializer struct {
}

func (s *GelfSerializer) Serialize(metric telegraf.Metric) ([]string, error) {
	out := []string{}

	m := make(map[string]interface{})
	m["version"] = "1.1"
	m["host"] = metric.Tags()["host"]
	m["timestamp"] = metric.UnixNano()
	//m["fields"] = metric.Fields()
	m["name"] = metric.Name()
	serialized, err := ejson.Marshal(m)
	if err != nil {
		return []string{}, err
	}
	out = append(out, string(serialized))

	return out, nil
}
