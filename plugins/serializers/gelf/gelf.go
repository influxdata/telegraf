package gelf

import (
	ejson "encoding/json"
	"fmt"
	"github.com/influxdata/telegraf"
	"os"
)

type GelfSerializer struct {
}

func (s *GelfSerializer) Serialize(metric telegraf.Metric) ([]string, error) {
	out := []string{}

	m := make(map[string]interface{})
	m["version"] = "1.1"
	m["timestamp"] = metric.UnixNano() / 1000000000
	m["short_message"] = " "
	m["name"] = metric.Name()

	if host, ok := metric.Tags()["host"]; ok {
		m["host"] = host
	} else {
		host, err := os.Hostname()
		if err != nil {
			return []string{}, err
		}
		m["host"] = host
	}

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
