package extr

import (
	"encoding/json"
	"github.com/influxdata/telegraf"
	"math"
	"time"
)

type serializer struct {
	TimestampUnits  time.Duration
	JsonBatchFields bool
}

func NewSerializer(timestampUnits time.Duration) (*serializer, error) {
	s := &serializer{
		TimestampUnits: truncateDuration(timestampUnits),
	}
	return s, nil
}

func (s *serializer) Serialize(metric telegraf.Metric) ([]byte, error) {

	m := s.createObject(metric)
	serialized, err := json.Marshal(m)
	if err != nil {
		return []byte{}, err
	}
	serialized = append(serialized, '\n')

	return serialized, nil
}

// Combine metrics whose Name, Tags, and TimeStamp match into a single batched metric
// where the fields are grouped into an array
// i.e.
// METRIC INPUT
// StatsCpu,node=NODE1  cpu=0,min=20,max=30,avg=25,interval=1,samplePeriod=10 1556813561098000000
// StatsCpu,node=NODE1  cpu=1,min=31,max=42,avg=76,interval=1,samplePeriod=10 1556813561098000000
// StatsCpu,node=NODE1  cpu=2,min=22,max=52,avg=11,interval=1,samplePeriod=10 1556813561098000000
// EventInterfaceStatus,node=NODE2  ifIndex="1001",port="1:1",adminStatus=1,operStatus=1 1556813561098000000
// EventInterfaceStatus,node=NODE2  ifIndex="1002",port="1:2",adminStatus=0,operStatus=0 1556813561098000000
//
// JSON OUTPUT
// [
//  {"fields":
//    [
//      {"avg":25,"cpu":0,"interval":1,"max":30,"min":20,"samplePeriod":10},
//      {"avg":76,"cpu":1,"interval":1,"max":42,"min":31,"samplePeriod":10},
//      {"avg":11,"cpu":2,"interval":1,"max":52,"min":22,"samplePeriod":10}
//    ],
//   "name":"StatsCpu",
//   "tags":{"node":"NODE1"},
//   "timestamp":1556813561},
//  {"fields":
//    [
//      {"adminStatus":1,"ifIndex":"1001","operStatus":1,"port":"1:1"},
//      {"adminStatus":0,"ifIndex":"1002","operStatus":0,"port":"1:2"}
//    ],
//   "name":"EventInterfaceStatus",
//   "tags":{"node":"NODE2"},
//   "timestamp":1556813561}
// ]
func (s *serializer) SerializeBatch(metrics []telegraf.Metric) ([]byte, error) {

	var serialized []byte
	var err error

	// Metric object
	object := make(map[string]interface{}, 4)

	// Array of metric objects
	objects := make([]interface{}, 0)

	for _, metric := range metrics {

		if object["fields"] == nil {
			// First batched metric
			object = s.createObject(metric)

		} else if s.metricMatch(object, metric) {
			// This metric's name, tags, and timestamp matches first metric.

			f := createField(metric)

			fieldArray := object["fields"].([]interface{})

			// Append field to fieldArray and re-assign new array to current metric object
			object["fields"] = append(fieldArray, f)

		} else {
			// This metric does not match first metric parameters.  Done with batched metric

			// Append the newly batched metric object to objects slice
			objects = append(objects, object)

			// Create a new batched metric with current metric object
			object = s.createObject(metric)
		}
	}

	fieldArray := object["fields"].([]interface{})
	if len(fieldArray) > 0 {

		objects = append(objects, object)

		serialized, err = json.Marshal(objects)
		if err != nil {
			return []byte{}, err
		}
	}

	return serialized, nil
}

func (s *serializer) metricMatch(newMetric map[string]interface{}, metric telegraf.Metric) bool {

	tags := newMetric["tags"].(map[string]string)
	name := newMetric["name"]
	timestamp := newMetric["timestamp"]

	if name != metric.Name() {
		return false
	}

	if timestamp != metric.Time().UnixNano()/int64(s.TimestampUnits) {
		return false
	}

	for _, tag := range metric.TagList() {
		if tags[tag.Key] != tag.Value {
			return false
		}
	}

	return true
}

func (s *serializer) createObject(metric telegraf.Metric) map[string]interface{} {
	m := make(map[string]interface{}, 4)
	fieldArray := make([]interface{}, 0)

	tags := make(map[string]string, len(metric.TagList()))
	for _, tag := range metric.TagList() {
		tags[tag.Key] = tag.Value
	}
	m["tags"] = tags

	// Create a field array
	f := createField(metric)
	fieldArray = append(fieldArray, f)
	m["fields"] = fieldArray

	m["name"] = metric.Name()
	m["timestamp"] = metric.Time().UnixNano() / int64(s.TimestampUnits)
	return m
}

func createField(metric telegraf.Metric) map[string]interface{} {

	f := make(map[string]interface{}, len(metric.FieldList()))

	for _, field := range metric.FieldList() {
		switch fv := field.Value.(type) {
		case float64:
			// JSON does not support these special values
			if math.IsNaN(fv) || math.IsInf(fv, 0) {
				continue
			}
		}
		f[field.Key] = field.Value
	}

	return f
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
