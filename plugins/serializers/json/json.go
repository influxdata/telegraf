package json

import (
	ejson "encoding/json"
	"time"

	"github.com/influxdata/telegraf"
)

type JsonSerializer struct {
}

func (s *JsonSerializer) Serialize(metric telegraf.Metric, output_precision ...string) ([]byte, error) {
	// if no duration is specified, or if the output_precision value passed
	// in represents a duration that is not greater than zero, then default
	// to an output resolution for our timestamp values of 1 second (timestamps
	// will be returned to the nearest second)
	default_val := "1s"
	var output_resolution time.Duration
	if len(output_precision) > 0 && len(output_precision[0]) > 0 {
		parsed_val, err := time.ParseDuration(output_precision[0])
		if err != nil {
			return nil, err
		}
		if parsed_val > 0.0 {
			output_resolution = parsed_val
		} else {
			output_resolution, _ = time.ParseDuration(default_val)
		}
	} else {
		output_resolution, _ = time.ParseDuration(default_val)
	}
	m := make(map[string]interface{})
	m["tags"] = metric.Tags()
	m["fields"] = metric.Fields()
	m["name"] = metric.Name()
	m["timestamp"] = metric.UnixNano() / output_resolution.Nanoseconds()
	serialized, err := ejson.Marshal(m)
	if err != nil {
		return []byte{}, err
	}
	serialized = append(serialized, '\n')

	return serialized, nil
}
