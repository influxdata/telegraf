package splunkmetric

import (
	"encoding/json"
//	"errors"
	"log"

	"github.com/influxdata/telegraf"
)

type serializer struct {
	HecRouting bool
}

func NewSerializer(hec_routing bool) (*serializer, error) {
	s := &serializer{
		HecRouting: hec_routing,
	}
	return s, nil
}

func (s *serializer) Serialize(metric telegraf.Metric) ([]byte, error) {

	m, err := s.createObject(metric)
	if err != nil {
        log.Printf("E! [serializer.splunkmetric] Dropping invalid metric: %v [%v]", metric, m)
		return []byte(""), err
	}

	return m, nil
}

func (s *serializer) SerializeBatch(metrics []telegraf.Metric) ([]byte, error) {

	var serialized []byte

	for _, metric := range metrics {
		m, err := s.createObject(metric)
		if err != nil {
            log.Printf("E! [serializer.splunkmetric] Dropping invalid metric: %v [%v]", metric, m)
		} else {
			serialized = append(serialized, m...)
		}
	}

	return serialized, nil
}

func (s *serializer) createObject(metric telegraf.Metric) (metricJson []byte, err error) {

	/*  Splunk supports one metric json object, and does _not_ support an array of JSON objects.
	     ** Splunk has the following required names for the metric store:
		 ** metric_name: The name of the metric
		 ** _value:      The value for the metric
		 ** time:       The timestamp for the metric
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
            log.Printf("E! Can not parse value: %v for key: %v",v,k)
		}

		obj := map[string]interface{}{}
		obj["metric_name"] = metric.Name() + "." + k
		obj["_value"] = v

		dataGroup.Event = "metric"
		// Convert ns to float seconds since epoch.
		dataGroup.Time = float64(metric.Time().UnixNano()) / float64(1000000000)
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

	switch s.HecRouting {
	case true:
		// Output the data as a fields array and host,index,time,source overrides for the HEC.
		metricJson, err = json.Marshal(dataGroup)
	default:
		// Just output the data and the time, useful for file based outuputs
		dataGroup.Fields["time"] = dataGroup.Time
		metricJson, err = json.Marshal(dataGroup.Fields)
	}

	if err != nil {
		return []byte(""), err
	}

	return metricJson, nil
}

func verifyValue(v interface{}) bool {
	switch v.(type) {
	case string:
		return false
	}
	return true
}
