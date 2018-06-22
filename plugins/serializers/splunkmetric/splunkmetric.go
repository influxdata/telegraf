package splunkmetric

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/influxdata/telegraf"
)

type serializer struct {
	TimestampUnits time.Duration
}

func NewSerializer(timestampUnits time.Duration) (*serializer, error) {
	s := &serializer{
		TimestampUnits: truncateDuration(timestampUnits),
	}
	return s, nil
}

func (s *serializer) Serialize(metric telegraf.Metric) ([]byte, error) {
	var serialized string

	m, err := s.createObject(metric)
	if err == nil {
		serialized = m + "\n"
	}

	return []byte(serialized), nil
}

func (s *serializer) SerializeBatch(metrics []telegraf.Metric) ([]byte, error) {

	var serialized string

	var objects []string

	for _, metric := range metrics {
		m, err := s.createObject(metric)
		if err == nil {
			objects = append(objects, m)
		}
	}

	for _, m := range objects {
		serialized = serialized + m + "\n"
	}

	return []byte(serialized), nil
}

func (s *serializer) createObject(metric telegraf.Metric) (metricString string, err error) {

	/* Splunk supports one metric per line and has the following required names:
	 ** metric_name: The name of the metric
	 ** _value:      The value for the metric
	 ** _time:       The timestamp for the metric
	 ** All other index fields become deminsions.
	 */
	type HECTimeSeries struct {
		Time   float64                `json:"time"`
		Event  string                 `json:"event"`
		Host   string                 `json:"host,omitempty"`
		Index  string                 `json:"index,omitempty"`
		Source string                 `json:"source,omitempty"`
		Fields map[string]interface{} `json:"fields"`
	}

	dataGroup := HECTimeSeries{}

	for k, v := range metric.Fields() {

		if !verifyValue(v) {
			err = errors.New("can not parse value")
			return "", err
		}

		obj := map[string]interface{}{}
		obj["metric_name"] = metric.Name() + "." + k
		obj["_value"] = v

		dataGroup.Event = "metric"
		dataGroup.Time = float64(metric.Time().UnixNano() / int64(s.TimestampUnits))
		dataGroup.Fields = obj

		// Break tags out into key(n)=value(t) pairs
		for n, t := range metric.Tags() {
			if n == "host" {
				dataGroup.Host = t
			} else if n == "index" {
				dataGroup.Index = t
			} else if n == "source" {
				dataGroup.Source = t
			} else {
				dataGroup.Fields[n] = t
			}
		}
		dataGroup.Fields["metric_name"] = metric.Name() + "." + k
		dataGroup.Fields["_value"] = v
	}

	metricJson, err := json.Marshal(dataGroup)

	if err != nil {
		return "", err
	}

	metricString = string(metricJson)
	return metricString, nil
}

func verifyValue(v interface{}) bool {
	switch v.(type) {
	case string:
		return false
	}
	return true
}

func truncateDuration(units time.Duration) time.Duration {
	// Default precision is 1s
	if units <= 0 {
		return time.Second
	}

	// Search for the power of ten less than the duration
	d := time.Nanosecond
	for {
		if d*10 > units {
			return d
		}
		d = d * 10
	}
}
