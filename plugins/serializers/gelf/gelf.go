package gelf

import (
	ejson "encoding/json"
	"fmt"

	"github.com/influxdata/telegraf"
)

type GelfSerializer struct {
}

func (s *GelfSerializer) Serialize(metric telegraf.Metric) ([]string, error) {
	out := []string{}

	m := make(map[string]interface{})
	m["version"] = "1.1"
	m["host"] = metric.Tags()["host"]
	m["timestamp"] = metric.UnixNano() / 1000000000
	m["short_message"] = "x"
	m["name"] = metric.Name()

	for key, value := range metric.Fields() {
		nkey := fmt.Sprintf("_%s", key)
		m[nkey] = value
	}

	serialized, err := ejson.Marshal(m)
	if err != nil {
		return []string{}, err
	}
	out = append(out, string(serialized))

	return out, nil
}
