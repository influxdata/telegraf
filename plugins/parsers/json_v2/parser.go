package json_v2

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/tidwall/gjson"
)

type TimeFunc func() time.Time

type Parser struct {
	Configs     []Config
	DefaultTags map[string]string
	Log         telegraf.Logger
	TimeFunc    func() time.Time
}

type Config struct {
	MetricSelection    string `toml:"metric_selection"`
	MetricName         string `toml:"metric_name"`
	UniformCollections []UniformCollection
	ObjectFields       []ObjectSelection
}

type UniformCollection struct {
	Query     string `toml:"query"`      // REQUIRED
	Name      string `toml:"name"`       // OPTIONAL
	ValueType string `toml:"value_type"` // OPTIONAL
	SetType   string `toml:"set_type"`   // OPTIONAL
}

type ObjectSelection struct {
	Query        string            `toml:"query"`          // REQUIRED
	NameMap      map[string]string `toml:"name_map"`       // OPTIONAL
	ValueTypeMap map[string]string `toml:"value_type_map"` // OPTIONAL
	// TODO: Add include_list and ignore_list
	// TODO: Add tag_list
}

// One field is required, set to true if one is added
var FieldExists = false

type MetricNode struct {
	RootFieldName string
	DesiredType   string // Can be "int", "bool", "string"
	SetType       string // Can either be "field" or "tag"
	// TODO: Make this a slice, array's in an object could make this expand
	Metric telegraf.Metric
	gjson.Result
}

func (p *Parser) Parse(input []byte) ([]telegraf.Metric, error) {
	// This is used to set a default time in the unit tests
	if p.TimeFunc == nil {
		p.TimeFunc = time.Now
	}

	// Only valid JSON is supported
	if !gjson.Valid(string(input)) {
		return nil, fmt.Errorf("Invalid JSON provided, unable to parse")
	}

	var t []telegraf.Metric

	for _, config := range p.Configs {
		// Process all `uniform_collection` configurations
		uniformCollection, err := p.processUniformCollections(config.MetricName, config.UniformCollections, input)
		if err != nil {
			return nil, err
		}
		if len(uniformCollection) != 0 {
			t = append(t, uniformCollection...)
		}

		objectMetrics, err := p.processObjectSelections(config.MetricName, config.ObjectFields, input)
		if err != nil {
			return nil, err
		}
		if len(objectMetrics) != 0 {
			t = append(t, objectMetrics...)
		}

		//TODO: Should object metrics and uniform metrics be merged?!?!?
	}

	if !FieldExists {
		return nil, fmt.Errorf("No field configured for the metrics")
	}

	return t, nil
}

func (p *Parser) processUniformCollections(metricName string, uniformCollection []UniformCollection, input []byte) ([]telegraf.Metric, error) {
	// For each uniform_collection configuration, get all the metric data returned from the query
	// Keep the metric data per field separate so all results from each query can be combined
	var metrics [][]telegraf.Metric
	for _, c := range uniformCollection {
		result := gjson.GetBytes(input, c.Query)

		if result.IsObject() {
			p.Log.Debugf("Found object in the uniform collection query: %s, ignoring it please use object_selection to gather metrics from objects", c.Query)
			continue
		}

		var setType string
		if c.SetType == "tag" {
			setType = "tag"
		} else {
			FieldExists = true
			setType = "field"
		}

		fieldName := c.Name
		// Default to the last query word, should be the upper key name
		// TODO: figure out what to do with special characters, probably ok to remove any special characters?
		if fieldName == "" {
			s := strings.Split(c.Query, ".")
			fieldName = s[len(s)-1]
		}

		// Store result into a MetricNode to keep metadata together
		mNode := MetricNode{
			RootFieldName: fieldName,
			DesiredType:   c.ValueType,
			SetType:       setType,
			Metric: metric.New(
				metricName,
				map[string]string{},
				map[string]interface{}{},
				p.TimeFunc(),
			),
			Result: result,
		}
		// Expand all array's and nested arrays into separate metrics
		nodes, err := p.expandArray(mNode, metricName, false)
		if err != nil {
			return nil, err
		}

		var m []telegraf.Metric
		for _, n := range nodes {
			m = append(m, n.Metric)
		}
		metrics = append(metrics, m)
	}

	for i := 1; i < len(metrics); i++ {
		metrics[i] = cartersianProduct(metrics[i-1], metrics[i])
	}

	return metrics[len(metrics)-1], nil
}

func cartersianProduct(a, b []telegraf.Metric) []telegraf.Metric {
	p := make([]telegraf.Metric, len(a)*len(b))
	i := 0
	for _, a := range a {
		for _, b := range b {
			m := a.Copy()
			m = mergeMetric(b, m)
			p[i] = m
			i++
		}
	}

	return p
}

func mergeMetric(a telegraf.Metric, m telegraf.Metric) telegraf.Metric {
	for _, f := range a.FieldList() {
		m.AddField(f.Key, f.Value)
	}
	for _, t := range a.TagList() {
		m.AddTag(t.Key, t.Value)
	}

	return m
}

// expandArray will recursively create a new MetricNode for each element in a JSON array
func (p *Parser) expandArray(result MetricNode, metricName string, combineObject bool) ([]MetricNode, error) {
	var results []MetricNode

	if result.IsObject() {
		return nil, fmt.Errorf("encountered object")
	}

	if result.IsArray() {
		var err error
		result.ForEach(func(_, val gjson.Result) bool {
			// TODO: implement `ignore_objects` config key to ignore this error
			// TODO: combineObjects calls this if it has an array of objects, need to handle this case
			if val.IsObject() {
				if combineObject {
					// TODO: call combine object
				} else {
					p.Log.Debugf("Found object in the uniform collection query ignoring it please use object_selection to gather metrics from objects")
				}
				return true
			}

			m := metric.New(
				metricName,
				map[string]string{},
				map[string]interface{}{},
				p.TimeFunc(),
			)

			if result.SetType == "field" {
				for _, f := range result.Metric.FieldList() {
					m.AddField(f.Key, f.Value)
				}
			} else {
				for _, f := range result.Metric.TagList() {
					m.AddTag(f.Key, f.Value)
				}
			}

			if val.IsArray() {
				n := MetricNode{
					SetType:       result.SetType,
					RootFieldName: result.RootFieldName,
					Metric:        m,
					Result:        val,
				}
				r, err := p.expandArray(n, metricName, combineObject)
				if err != nil {
					return false
				}
				results = append(results, r...)
			} else {
				if result.SetType == "field" {
					v, err := p.convertType(val.Value(), result.DesiredType, result.RootFieldName)
					if err != nil {
						return false
					}

					m.AddField(result.RootFieldName, v)

					for _, f := range result.Metric.FieldList() {
						m.AddField(f.Key, f.Value)
					}
				} else {
					v, err := p.convertType(val.Value(), "string", result.RootFieldName)
					if err != nil {
						return false
					}
					m.AddTag(result.RootFieldName, v.(string))

					for _, f := range result.Metric.TagList() {
						m.AddTag(f.Key, f.Value)
					}
				}

				n := MetricNode{
					RootFieldName: result.RootFieldName,
					Metric:        m,
					Result:        val,
				}
				results = append(results, n)
			}
			return true
		})
		if err != nil {
			return nil, err
		}
	} else {
		if result.Exists() {
			if result.SetType == "field" {
				v, err := p.convertType(result.Value(), result.DesiredType, result.RootFieldName)
				if err != nil {
					return nil, err
				}
				result.Metric.AddField(result.RootFieldName, v)
			} else {
				v, err := p.convertType(result.Value(), "string", result.RootFieldName)
				if err != nil {
					return nil, err
				}
				result.Metric.AddTag(result.RootFieldName, v.(string))
			}
			results = append(results, result)
		}
	}

	return results, nil
}

func (p *Parser) processObjectSelections(metricName string, objectSelections []ObjectSelection, input []byte) ([]telegraf.Metric, error) {
	var t []telegraf.Metric
	for _, field := range objectSelections {
		result := gjson.GetBytes(input, field.Query)

		// TODO: Figoure out how to handle root fieldname, will be blank
		// Default to the last query word, should be the upper key name
		// TODO: figure out what to do with special characters, probably ok to remove any special characters?
		s := strings.Split(field.Query, ".")
		fieldName := s[len(s)-1]

		rootObject := MetricNode{
			RootFieldName: fieldName,
			Metric: metric.New(
				metricName,
				map[string]string{},
				map[string]interface{}{},
				p.TimeFunc(),
			),
			Result: result,
		}
		metrics, err := p.combineObject(rootObject, field.NameMap, field.ValueTypeMap)
		if err != nil {
			return nil, err
		}
		for _, m := range metrics {
			t = append(t, m.Metric)
		}
	}

	return t, nil
}

func (p *Parser) combineObject(result MetricNode, nameMap map[string]string, typeMap map[string]string) ([]MetricNode, error) {

	var metrics []MetricNode
	result.ForEach(func(key, val gjson.Result) bool {
		// Update key with user configuration
		fieldName := key.String()
		if fieldName != "" {
			if newName, ok := nameMap[fieldName]; ok {
				fieldName = newName
			}
		} else {
			fieldName = result.RootFieldName
		}

		if val.IsArray() {
			arrayNode := MetricNode{
				RootFieldName: key.String(),
				Metric:        result.Metric,
				Result:        val,
			}

			// TODO: This will fail if its an array of objects
			m, err := p.expandArray(arrayNode, result.Metric.Name(), true)
			if err != nil {
				return false
			}
			// TODO: THIS IS WRONG, should be added to result metric slice
			metrics = append(metrics, m...)
		} else if val.IsObject() {
			arrayNode := MetricNode{
				RootFieldName: key.String(),
				Metric:        result.Metric,
				Result:        val,
			}
			_, err := p.combineObject(arrayNode, nameMap, typeMap)
			if err != nil {
				return false
			}
		} else {
			fieldValue := val.Value()
			if desiredType, ok := typeMap[key.String()]; ok {
				var err error
				// TODO: Return this error
				fieldValue, err = p.convertType(val.Value(), desiredType, key.String())
				if err != nil {
					return false
				}
			}

			result.Metric.AddField(fieldName, fieldValue)
		}

		return true
	})

	metrics = append(metrics, result)

	return metrics, nil
}

func (p *Parser) ParseLine(line string) (telegraf.Metric, error) {
	return nil, fmt.Errorf("ParseLine is designed for parsing influx line protocol, therefore not implemented for parsing JSON")
}

func (p *Parser) SetDefaultTags(tags map[string]string) {

}

func (p *Parser) SetTimeFunc(fn TimeFunc) {
	p.TimeFunc = fn
}

// convertType will convert the value parsed from the input JSON to the specified type in the config
func (p *Parser) convertType(input interface{}, desiredType string, name string) (interface{}, error) {
	switch inputType := input.(type) {
	case string:
		if desiredType != "string" {
			switch desiredType {
			case "int":
				r, err := strconv.Atoi(inputType)
				if err != nil {
					return nil, fmt.Errorf("Unable to convert field '%s' to type int: %v", name, err)
				}
				return r, nil
			case "float":
				r, err := strconv.ParseFloat(inputType, 64)
				if err != nil {
					return nil, fmt.Errorf("Unable to convert field '%s' to type float: %v", name, err)
				}
				return r, nil
			case "bool":
				r, err := strconv.ParseBool(inputType)
				if err != nil {
					return nil, fmt.Errorf("Unable to convert field '%s' to type bool: %v", name, err)
				}
				return r, nil
			}
		}
	case bool:
		if desiredType != "bool" {
			switch desiredType {
			case "string":
				return strconv.FormatBool(inputType), nil
			case "int":
				if inputType {
					return int64(1), nil
				}

				return int64(0), nil
			}
		}
	case float64:
		if desiredType != "float" {
			switch desiredType {
			case "string":
				return fmt.Sprint(inputType), nil
			case "int":
				return int64(inputType), nil
			case "bool":
				if inputType == 0 {
					return false, nil
				} else if inputType == 1 {
					return true, nil
				} else {
					return nil, fmt.Errorf("Unable to convert field '%s' to type bool", name)
				}
			}
		}
	default:
		return nil, fmt.Errorf("unknown format '%T' for field  '%s'", inputType, name)
	}

	return input, nil
}
