package gjson

import (
	"log"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/tidwall/gjson"
)

type JSONPath struct {
	MetricName  string
	TagPath     map[string]string
	FloatPath   map[string]string
	IntPath     map[string]string
	StrPath     map[string]string
	BoolPath    map[string]string
	DefaultTags map[string]string
}

func (j *JSONPath) Parse(buf []byte) ([]telegraf.Metric, error) {
	tags := make(map[string]string)
	for k, v := range j.DefaultTags {
		tags[k] = v
	}
	fields := make(map[string]interface{})
	metrics := make([]telegraf.Metric, 0)

	for k, v := range j.TagPath {
		c := gjson.GetBytes(buf, v)
		tags[k] = c.String()
	}

	for k, v := range j.FloatPath {
		c := gjson.GetBytes(buf, v)
		fields[k] = c.Float()
	}

	for k, v := range j.IntPath {
		c := gjson.GetBytes(buf, v)
		fields[k] = c.Int()
	}

	for k, v := range j.BoolPath {
		c := gjson.GetBytes(buf, v)
		if c.String() == "true" {
			fields[k] = true
		} else if c.String() == "false" {
			fields[k] = false
		} else {
			log.Printf("E! Cannot decode: %v as bool", c.String())
		}
	}

	for k, v := range j.StrPath {
		c := gjson.GetBytes(buf, v)
		fields[k] = c.String()
	}

	m, err := metric.New(j.MetricName, tags, fields, time.Now())
	if err != nil {
		return nil, err
	}
	metrics = append(metrics, m)
	return metrics, nil
}

func (j *JSONPath) ParseLine(str string) (telegraf.Metric, error) {
	m, err := j.Parse([]byte(str))
	return m[0], err
}

func (j *JSONPath) SetDefaultTags(tags map[string]string) {
	j.DefaultTags = tags
}
