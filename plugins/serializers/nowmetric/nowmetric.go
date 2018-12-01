package nowmetric

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/influxdata/telegraf"
)

type serializer struct {
	TimestampUnits time.Duration
}

/*
Example for the JSON generated and pushed to the MID
{
	"metric_type":"cpu_usage_system",
	"resource":"",
	"node":"ASGARD",
	"value": 0.89,
	"timestamp":1487365430,
	"ci2metric_id":{"node":"ASGARD"},
	"source":"Telegraf"
}
*/

type OIMetric struct {
	Metric    string                 `json:"metric_type"`
	Resource  string                 `json:"resource"`
	Node      string                 `json:"node"`
	Value     interface{}            `json:"value"`
	Timestamp int64                  `json:"timestamp"`
	CiMapping map[string]interface{} `json:"ci2metric_id"`
	Source    string                 `json:"source"`
}

func NewSerializer(timestampUnits time.Duration) (*serializer, error) {
	s := &serializer{
		TimestampUnits: truncateDuration(timestampUnits),
	}
	return s, nil
}

func (s *serializer) Serialize(metric telegraf.Metric) (out []byte, err error) {
	serialized := s.createObject(metric, err)
	if serialized == nil {
		return []byte{}, nil
	}
	return serialized, nil
}

func (s *serializer) SerializeBatch(metrics []telegraf.Metric) (out []byte, err error) {
	objects := make([]byte, 0)
	for _, metric := range metrics {
		m := s.createObject(metric, err)
		objects = append(objects, m...)
	}

	if objects == nil {
		return []byte{}, nil
	}
	return objects, nil
}

func (s *serializer) createObject(metric telegraf.Metric, err error) []byte {
	var payload []byte
	var oimetric OIMetric
	var metricJson []byte

	oimetric.Source = "Telegraf"

	// Process Tags to extract node & resource name info
	for _, tag := range metric.TagList() {
		if tag.Key == "" || tag.Value == "" {
			continue
		}

		if tag.Key == "objectname" {
			oimetric.Resource = tag.Value
		}

		if tag.Key == "host" {
			oimetric.Node = tag.Value
		}
	}

	// Format timestamp to UNIX epoch
	oimetric.Timestamp = (metric.Time().UnixNano() / int64(s.TimestampUnits)) * 1000

	nbdatapoint := 0
	// Loop of fields value pair and build datapoint for each of them
	for _, field := range metric.FieldList() {
		if !verifyValue(field.Value) {
			// Ignore String
			continue
		}

		if field.Key == "" || field.Value == "" {
			continue
		}

		if nbdatapoint >= 1 {
			payload = append(payload, ",\n"...)
		}
		oimetric.Metric = field.Key
		oimetric.Value = field.Value

		cimapping := map[string]interface{}{}
		cimapping["node"] = oimetric.Node
		oimetric.CiMapping = cimapping

		metricJson, err = json.Marshal(oimetric)
		payload = append(payload, metricJson...)

		nbdatapoint++
	}

	raw := make([]byte, 0)
	raw = append(raw, "[ "...)
	raw = append(raw, payload...)
	raw = append(raw, " ]\n"...)

	out := new(bytes.Buffer)
	json.HTMLEscape(out, raw)

	return out.Bytes()
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

func verifyValue(v interface{}) bool {
	switch v.(type) {
	case string:
		return false
	}
	return true
}
