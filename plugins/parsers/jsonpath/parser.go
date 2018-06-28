package jsonpath

import (
	"encoding/json"
	"log"
	"reflect"
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/jsonpath"
)

type JSONPath struct {
	MetricName string
	TagPath    map[string]string
	FieldPath  map[string]string
}

func (j *JSONPath) Parse(buf []byte) ([]telegraf.Metric, error) {
	tags := make(map[string]string)
	fields := make(map[string]interface{})
	metrics := make([]telegraf.Metric, 0)

	//turn buf into json interface
	var jsonData interface{}
	json.Unmarshal(buf, &jsonData)
	log.Printf("unmarshaled jsonData: %v", jsonData)

	for k, v := range j.TagPath {
		c, err := jsonpath.JsonPathLookup(jsonData, v)
		if err != nil {
			log.Printf("E! Could not find JSON Path: %v", v)
		}
		cType := reflect.TypeOf(c)

		//if path returns multiple values, split each into a different metric
		if cType.Kind() == reflect.Array {
			log.Printf("E! Multiple return values for path: %v", v)
			continue
		}

		switch ct := c.(type) {
		case string:
			tags[k] = ct
		case bool:
			tags[k] = strconv.FormatBool(ct)
		case float64:
			tags[k] = strconv.FormatFloat(ct, 'f', -1, 64)
		default:
			log.Printf("E! [parsers.json] Unrecognized type %T", ct)
		}
	}

	for k, v := range j.FieldPath {
		c, err := jsonpath.JsonPathLookup(jsonData, v)
		if err != nil {
			log.Printf("E! Could not find JSON Path: %v", v)
			continue
		}

		cType := reflect.TypeOf(c)

		//if path returns multiple values, split each into a different metric
		if cType.Kind() == reflect.Array {
			log.Printf("E! Multiple return values for path: %v", v)
			continue
		}
		fields[k] = c
	}

	m, _ := metric.New(j.MetricName, tags, fields, time.Now())
	metrics = append(metrics, m)
	return metrics, nil
}
