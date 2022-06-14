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
	Metric    string            `json:"metric_type"`
	Resource  string            `json:"resource"`
	Node      string            `json:"node"`
	Value     interface{}       `json:"value"`
	Timestamp int64             `json:"timestamp"`
	CiMapping map[string]string `json:"ci2metric_id"`
	Source    string            `json:"source"`
}

type OIMetrics []OIMetric

func NewSerializer() (*serializer, error) {
	s := &serializer{}
	return s, nil
}

func (s *serializer) Serialize(metric telegraf.Metric) (out []byte, err error) {
	serialized, err := s.createObject(metric)
	if err != nil {
		return []byte{}, err
	}
	return serialized, nil
}

func (s *serializer) SerializeBatch(metrics []telegraf.Metric) (out []byte, err error) {
	objects := make([]byte, 0)
	for _, metric := range metrics {
		m, err := s.createObject(metric)
		if err != nil {
			return nil, fmt.Errorf("dropping invalid metric: %s", metric.Name())
		} else if m != nil {
			objects = append(objects, m...)
		}
	}
	replaced := bytes.Replace(objects, []byte("]["), []byte(","), -1)
	return replaced, nil
}

func (s *serializer) createObject(metric telegraf.Metric) ([]byte, error) {
	/*  ServiceNow Operational Intelligence supports an array of JSON objects.
	** Following elements accepted in the request body:
		 ** metric_type: 	The name of the metric
		 ** resource:   	Information about the resource for which metric data is being collected. In the example below, C:\ is the resource for which metric data is collected
		 ** node:       	IP, FQDN, name of the CI, or host
		 ** value:      	Value of the metric
		 ** timestamp: 		Epoch timestamp of the metric in milliseconds
		 ** ci2metric_id:	List of key-value pairs to identify the CI.
		 ** source:			Data source monitoring the metric type
	*/
	var allmetrics OIMetrics
	var oimetric OIMetric

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
	oimetric.Timestamp = metric.Time().UnixNano() / int64(time.Millisecond)

	// Loop of fields value pair and build datapoint for each of them
	for _, field := range metric.FieldList() {
		if !verifyValue(field.Value) {
			// Ignore String
			continue
		}

		if field.Key == "" {
			// Ignore Empty Key
			continue
		}

		oimetric.Metric = field.Key
		oimetric.Value = field.Value

		if oimetric.Node != "" {
			cimapping := map[string]string{}
			cimapping["node"] = oimetric.Node
			oimetric.CiMapping = cimapping
		}

		allmetrics = append(allmetrics, oimetric)
	}

	metricsJSON, err := json.Marshal(allmetrics)

	return metricsJSON, err
}

func verifyValue(v interface{}) bool {
	_, ok := v.(string)
	return !ok
}
