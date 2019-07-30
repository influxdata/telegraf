package customjson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"

	"github.com/influxdata/telegraf"
	"github.com/jmespath/go-jmespath"
)

type serializer struct {
	JmespathExpression string
	TagsPrefix         string
}

func NewSerializer(jmespath_expression string, tags_prefix string) (*serializer, error) {
	s := &serializer{
		JmespathExpression: jmespath_expression,
		TagsPrefix:         tags_prefix,
	}
	return s, nil
}

func (s *serializer) Serialize(metric telegraf.Metric) ([]byte, error) {

	m, err := s.createObject(metric)
	if err != nil {
		return nil, fmt.Errorf("D! [serializer.customjson] Dropping invalid metric: %s", metric.Name())
	}

	return m, nil
}

func (s *serializer) SerializeBatch(metrics []telegraf.Metric) ([]byte, error) {

	var serialized []byte

	for _, metric := range metrics {
		m, err := s.createObject(metric)
		if err != nil {
			return nil, fmt.Errorf("D! [serializer.customjson] Dropping invalid metric: %s", metric.Name())
		} else if m != nil {
			serialized = append(serialized, m...)
		}
	}

	return serialized, nil
}

func (s *serializer) createObject(metric telegraf.Metric) (metricGroup []byte, err error) {

	/*  All fields index become dimensions and all tags index can be prefixed by tags_prefix config input and located on json root.
	    ** Default customjson format contains the following fields:
		** metric_family: The name of the metric
		** metric_name:   The name of the fields dimension
		** metric_value:  The value of the fields dimension
		** *:             The name and values of the tags
		** timestamp:     The timestamp for the metric
	*/

	// Build output result
	dataGroup := map[string]interface{}{}
	var metricJson []byte

	if s.JmespathExpression != "" {
		jmespath.MustCompile(s.JmespathExpression)
	}

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

		// Build fields parameter
		dataGroup["metric_name"] = field.Key
		dataGroup["metric_value"] = fieldValue

		// Build tags parameter
		if s.JmespathExpression != "" {
			metricJson, err = buildJmespathTagsParameter(metric, dataGroup, s)
		} else {
			metricJson, err = buildDefaultTagsParameter(metric, dataGroup, s)
		}

		// Output the data as a fields array.
		metricJson = append(metricJson, '\n')

		metricGroup = append(metricGroup, metricJson...)

		if err != nil {
			return nil, err
		}
	}

	return metricGroup, nil
}

func buildJmespathTagsParameter(metric telegraf.Metric, dataGroup map[string]interface{}, s *serializer) (metricJson []byte, err error) {
	var jmespathBuffer bytes.Buffer
	jmespathExpressionLen := len(s.JmespathExpression)
	jmespathBuffer.WriteString(s.JmespathExpression[:jmespathExpressionLen-1])
	for n, t := range metric.Tags() {
		jmespathBuffer.WriteString(",")
		if s.TagsPrefix != "" {
			jmespathBuffer.WriteString(s.TagsPrefix)
			jmespathBuffer.WriteString("_")
		}
		jmespathBuffer.WriteString(n)
		jmespathBuffer.WriteString(":")
		jmespathBuffer.WriteString("'")
		jmespathBuffer.WriteString(t)
		jmespathBuffer.WriteString("'")
	}
	jmespathBuffer.WriteString("}")

	jmespathDataGroup, err := jmespath.Search(jmespathBuffer.String(), dataGroup)
	if err != nil {
		return nil, err
	}
	return json.Marshal(jmespathDataGroup)
}

func buildDefaultTagsParameter(metric telegraf.Metric, dataGroup map[string]interface{}, s *serializer) (metricJson []byte, err error) {
	for n, t := range metric.Tags() {
		var tagsBuffer bytes.Buffer
		if s.TagsPrefix != "" {
			tagsBuffer.WriteString(s.TagsPrefix)
			tagsBuffer.WriteString("_")
		}
		tagsBuffer.WriteString(n)
		tags_name := tagsBuffer.String()
		dataGroup[tags_name] = t
	}
	return json.Marshal(dataGroup)
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
