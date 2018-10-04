package nowmetric

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/influxdata/telegraf"
)

type serializer struct {
	TimestampUnits time.Duration
}

const METRICFMT string = "{ \"metric_type\": \"%s\", \"resource\": \"%s\", \"node\": \"%s\", \"value\": %v, \"timestamp\": %d, \"ci2metric_id\": { \"node\": \"%s\" }, \"source\": \"Telegraf\" }"

// field 1, resourcename, hostname, field 2, timestamp, hostname
//
func NewSerializer(timestampUnits time.Duration) (*serializer, error) {
	s := &serializer{
		TimestampUnits: truncateDuration(timestampUnits),
	}
	return s, nil
}

func (s *serializer) Serialize(metric telegraf.Metric) ([]byte, error) {
	serialized := s.createObject(metric)
	if serialized == nil {
		return []byte{}, nil
	}
	return serialized, nil
}

func (s *serializer) SerializeBatch(metrics []telegraf.Metric) ([]byte, error) {

	objects := make([]byte, 0)
	for _, metric := range metrics {
		m := s.createObject(metric)
		objects = append(objects, m...)
	}

	if objects == nil {
		return []byte{}, nil
	}
	return objects, nil

}

func (s *serializer) createObject(metric telegraf.Metric) []byte {
	var payload string
	var resourcename string
	var hostname string
	var utime int64

	// Process Tags to extract node & resource name info
	for _, tag := range metric.TagList() {
		key := tag.Key
		value := tag.Value

		if key == "" || value == "" {
			continue
		}

		if key == "objectname" {
			resourcename = value
		}
		if key == "host" {
			hostname = value
		}
	}

	// Format timestamp to UNIX epoch
	utime = (metric.Time().UnixNano() / int64(s.TimestampUnits)) * 1000

	nbdatapoint := 0
	// Loop of fields value pair and build datapoint for each of them
	for _, field := range metric.FieldList() {
		// field.Key, field.Value
		metrictype := field.Key
		metricvalue := field.Value

		if !verifyValue(metricvalue) {
			// Ignore String
			continue
		}

		if metrictype == "" || metricvalue == "" {
			continue
		}
		// Params : metrictype, resourcename, hostname, metricvalue, utime, hostname
		if nbdatapoint >= 1 {
			payload = payload + ",\n"
		}
		payload = payload + fmt.Sprintf(METRICFMT, metrictype, resourcename, hostname, metricvalue, utime, hostname)
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
