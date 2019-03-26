package splunkmetric

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/influxdata/telegraf"
)

type serializer struct {
	HecRouting bool
}

func NewSerializer(splunkmetric_hec_routing bool) (*serializer, error) {
	s := &serializer{
		HecRouting: splunkmetric_hec_routing,
	}
	return s, nil
}

func (s *serializer) Serialize(metric telegraf.Metric) ([]byte, error) {

	m, err := s.createObject(metric)
	if err != nil {
		return nil, fmt.Errorf("D! [serializer.splunkmetric] Dropping invalid metric: %s", metric.Name())
	}

	return m, nil
}

func (s *serializer) SerializeBatch(metrics []telegraf.Metric) ([]byte, error) {

	var serialized []byte

	for _, metric := range metrics {
		m, err := s.createObject(metric)
		if err != nil {
			return nil, fmt.Errorf("D! [serializer.splunkmetric] Dropping invalid metric: %s", metric.Name())
		} else if m != nil {
			serialized = append(serialized, m...)
		}
	}

	return serialized, nil
}

func (s *serializer) createObject(metric telegraf.Metric) (metricGroup []byte, err error) {

	/*  Splunk supports one metric json object, and does _not_ support an array of JSON objects.
	     ** Splunk has the following required names for the metric store:
		 ** metric_name: The name of the metric
		 ** _value:      The value for the metric
		 ** time:       The timestamp for the metric
		 ** All other index fields become dimensions.
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
	var metricJson []byte

	for _, field := range metric.FieldList() {

		value, valid := verifyValue(field.Value)

		if !valid {
			log.Printf("D! Can not parse value: %v for key: %v", field.Value, field.Key)
			continue
		}

		obj := map[string]interface{}{}
		obj["metric_name"] = metric.Name() + "." + field.Key
		obj["_value"] = value

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

		switch s.HecRouting {
		case true:
			// Output the data as a fields array and host,index,time,source overrides for the HEC.
			metricJson, err = json.Marshal(dataGroup)
		default:
			// Just output the data and the time, useful for file based outuputs
			dataGroup.Fields["time"] = dataGroup.Time
			metricJson, err = json.Marshal(dataGroup.Fields)
		}

		metricGroup = append(metricGroup, metricJson...)

		if err != nil {
			return nil, err
		}
	}

	return metricGroup, nil
}

func verifyValue(v interface{}) (value interface{}, valid bool) {
	switch v.(type) {
	case string:
		valid = false
		value = v
	case bool:
		if v == bool(true) {
			// Store 1 for a "true" value
			valid = true
			value = 1
		} else {
			// Otherwise store 0
			valid = true
			value = 0
		}
	default:
		valid = true
		value = v
	}
	return value, valid
}
