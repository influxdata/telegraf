package flattenjson

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/influxdata/telegraf"
)

type serializer struct {
}

func NewSerializer() (*serializer, error) {
	s := &serializer{}
	return s, nil
}

func (s *serializer) Serialize(metric telegraf.Metric) ([]byte, error) {

	m, err := s.createObject(metric)
	if err != nil {
		return nil, fmt.Errorf("D! [serializer.flattenjson] Dropping invalid metric: %s", metric.Name())
	}

	return m, nil
}

func (s *serializer) SerializeBatch(metrics []telegraf.Metric) ([]byte, error) {

	var serialized []byte

	for _, metric := range metrics {
		m, err := s.createObject(metric)
		if err != nil {
			return nil, fmt.Errorf("D! [serializer.flattenjson] Dropping invalid metric: %s", metric.Name())
		} else if m != nil {
			serialized = append(serialized, m...)
		}
	}

	return serialized, nil
}

func (s *serializer) createObject(metric telegraf.Metric) (metricGroup []byte, err error) {

	/*  All fields index become dimensions and all tags index are prefixed by 'tags_' and located on json root.
	    ** Flattenjson contains the following fields:
		** metric_family: The name of the metric
		** metric_name:   The name of the fields dimension
		** metric_value:  The value of the fields dimension
		** tags_*:	The name and values of the tags
		** timestamp:     The timestamp for the metric
	*/

	// Build output result
	dataGroup := map[string]interface{}{}
	var metricJson []byte

	for _, field := range metric.FieldList() {

		fieldValue, valid := verifyValue(field.Value)

		if !valid {
			log.Printf("D! Can not parse value: %v for key: %v", field.Value, field.Key)
			continue
		}

		// Build root parameter
		dataGroup["metric_family"] = metric.Name()
		// Convert ns to float milliseconds since epoch.
		dataGroup["timestamp"] = float64(metric.Time().UnixNano()) / float64(1000000)

		// Build tags parameter
		for n, t := range metric.Tags() {
			dataGroup["tags_" + n] = t
		}

		// Build fields parameter
		dataGroup["metric_name"] = field.Key
		dataGroup["metric_value"] = fieldValue

		// Output the data as a fields array.
		metricJson, err = json.Marshal(dataGroup)
		metricJson = append(metricJson, '\n')

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
